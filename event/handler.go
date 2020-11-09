package event

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/image"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/log.go/log"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

//go:generate moq -out mock/s3_reader.go -pkg mock . S3Reader
//go:generate moq -out mock/s3_writer.go -pkg mock . S3Writer
//go:generate moq -out mock/vault.go -pkg mock . VaultClient
//go:generate moq -out mock/image_api.go -pkg mock . ImageAPIClient

const (
	vaultKey       = "key" // is the key under each vault secret that contains the PSK needed to encrypt/decrypt files in S3
	importingState = "importing"
	importedState  = "imported"
	variantId      = "original"
	variantType    = "png"
)

//ErrVaultFilenameEmpty is an error returned when trying to obtain a PSK for an empty file name
var ErrVaultFilenameEmpty = errors.New("vault filename required but was empty")

// ImageUploadedHandler ...
type ImageUploadedHandler struct {
	AuthToken          string
	S3Upload           S3Reader
	S3Private          S3Writer
	VaultCli           VaultClient
	VaultPath          string
	ImageCli           ImageAPIClient
	DownloadServiceURL string
}

// S3Writer defines the required methods from dp-s3 to interact with a particular bucket of AWS S3
type S3Writer interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	Session() *session.Session
	BucketName() string
	Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
	UploadWithPSK(input *s3manager.UploadInput, psk []byte) (*s3manager.UploadOutput, error)
}

// S3Reader defines the required methods from dp-s3 to read data to an AWS S3 Bucket
type S3Reader interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	Session() *session.Session
	BucketName() string
	Get(key string) (io.ReadCloser, *int64, error)
	GetWithPSK(key string, psk []byte) (io.ReadCloser, *int64, error)
}

// VaultClient defines the required methods from dp-vault client
type VaultClient interface {
	ReadKey(path, key string) (string, error)
	WriteKey(path, key, value string) error
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}

// ImageAPIClient defines the required methods from image API client
type ImageAPIClient interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	PostDownloadVariant(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, imageID string, data image.NewImageDownload) (image.ImageDownload, error)
	PutDownloadVariant(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, imageID, variant string, data image.ImageDownload) (image.ImageDownload, error)
}

// Handle takes a single event. It reads the PSK from Vault, uses it to decrypt the encrypted file
// from the private S3 bucket, and writes it to the public static bucket without using the vault psk for encryption.
func (h *ImageUploadedHandler) Handle(ctx context.Context, event *ImageUploaded) (err error) {
	uploadBucket := h.S3Upload.BucketName()
	privateBucket := h.S3Private.BucketName()
	logData := log.Data{
		"event":          event,
		"upload_bucket":  uploadBucket,
		"private_bucket": privateBucket,
		"vault_path":     h.VaultPath,
	}
	log.Event(ctx, "event handler called", log.INFO, logData)

	startTime := time.Now().UTC()
	uploadPath := event.Path

	var uploadPsk []byte
	if h.VaultCli != nil {
		uploadPsk, err = h.getVaultKeyForFile(uploadPath)
		if err != nil {
			log.Event(ctx, "error reading key from vault", log.ERROR, log.Error(err), logData)
			return err
		}
	}

	reader, err := h.getS3Reader(ctx, uploadPath, uploadPsk)
	if err != nil {
		log.Event(ctx, "error getting s3 object reader", log.ERROR, log.Error(err), logData)
		return
	}
	defer reader.Close()

	// POST /images/{id}/downloads
	imageDownload, err := h.ImageCli.PostDownloadVariant(ctx, "", h.AuthToken, "", event.ImageID, image.NewImageDownload{
		Id:            variantId,
		Height:        nil,
		Type:          variantType,
		Width:         nil,
		State:         importingState,
		ImportStarted: &startTime,
	})
	if err != nil {
		log.Event(ctx, "error posting image variant to API", log.ERROR, log.Error(err), logData)
		return
	}
	logData["imageDownload"] = &imageDownload
	log.Event(ctx, "posted image download", log.INFO, logData)

	// Variant S3 key 'images/{id}/{variantId}'
	variantPath := path.Join("images", event.ImageID, variantId)
	logData["variant_path"] = variantPath

	var variantPSK []byte
	if h.VaultCli != nil {
		log.Event(ctx, "writing new key to vault", log.INFO, logData)
		variantPSK, err = h.createVaultKeyForFile(variantPath)
		if err != nil {
			log.Event(ctx, "error writing key to vault", log.ERROR, log.Error(err), logData)
			return
		}
	}

	log.Event(ctx, "uploading private file to s3", log.INFO, logData)
	err = h.uploadToS3(ctx, variantPath, variantPSK, reader)
	if err != nil {
		log.Event(ctx, "error uploading to s3", log.ERROR, log.Error(err), logData)
		return
	}
	endTime := time.Now().UTC()

	// PUT /images/{id}/downloads/{variantId})
	imageDownload.State = importedState
	imageDownload.ImportCompleted = &endTime
	fileName := event.Filename
	if fileName == "" {
		fileName = path.Base(uploadPath)
	}
	imageDownload.Href = fmt.Sprintf("%s/%s/%s", h.DownloadServiceURL, variantPath, fileName)
	imageDownload, err = h.ImageCli.PutDownloadVariant(ctx, "", h.AuthToken, "", event.ImageID, imageDownload.Id, imageDownload)
	if err != nil {
		log.Event(ctx, "error putting image variant to API", log.ERROR, log.Error(err), logData)
		return
	}
	log.Event(ctx, "put image download", log.INFO, logData)
	log.Event(ctx, "event successfully handled", log.INFO, logData)
	return nil
}

// Get an S3 reader
func (h *ImageUploadedHandler) getS3Reader(ctx context.Context, path string, psk []byte) (reader io.ReadCloser, err error) {
	if psk != nil {
		// Decrypt image from upload bucket using PSK obtained from Vault
		reader, _, err = h.S3Upload.GetWithPSK(path, psk)
		if err != nil {
			return
		}
	} else {
		// Get image from upload bucket
		reader, _, err = h.S3Upload.Get(path)
		if err != nil {
			return
		}
	}
	return
}

// Upload to S3 from a reader
func (h *ImageUploadedHandler) uploadToS3(ctx context.Context, path string, psk []byte, reader io.Reader) error {
	privateBucket := h.S3Private.BucketName()
	uploadInput := &s3manager.UploadInput{
		Body:   reader,
		Bucket: &privateBucket,
		Key:    &path,
	}
	if psk != nil {
		// Upload file to private bucket
		_, err := h.S3Private.UploadWithPSK(uploadInput, psk)
		if err != nil {
			return err
		}
	} else {
		// Upload file to private bucket
		_, err := h.S3Private.Upload(uploadInput)
		if err != nil {
			return err
		}
	}
	return nil
}

// getVaultKeyForFile reads the encryption key from Vault for the provided path
func (h *ImageUploadedHandler) getVaultKeyForFile(keyPath string) ([]byte, error) {
	if len(keyPath) == 0 {
		return nil, ErrVaultFilenameEmpty
	}

	vp := path.Join(h.VaultPath, keyPath)
	pskStr, err := h.VaultCli.ReadKey(vp, vaultKey)
	if err != nil {
		return nil, err
	}

	psk, err := hex.DecodeString(pskStr)
	if err != nil {
		return nil, err
	}

	return psk, nil
}

// getVaultKeyForFile creates a new encryption key and stores it in Vault for the provided path
func (h *ImageUploadedHandler) createVaultKeyForFile(keyPath string) ([]byte, error) {
	psk := make([]byte, 16)
	rand.Read(psk)

	vaultPath := path.Join(h.VaultPath, keyPath)
	if err := h.VaultCli.WriteKey(vaultPath, vaultKey, hex.EncodeToString(psk)); err != nil {
		return nil, err
	}

	return psk, nil
}
