package event

import (
	"context"
	//"encoding/hex"
	"errors"
	"io"
	//"path"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/log.go/log"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

//go:generate moq -out mock/s3_reader.go -pkg mock . S3Reader
//go:generate moq -out mock/s3_writer.go -pkg mock . S3Writer
//go:generate moq -out mock/vault.go -pkg mock . VaultClient

// VaultKey is the key under the vault that contains the PSK needed to decrypt files from the encrypted private S3 bucket
const VaultKey = "key"

//ErrVaultFilenameEmpty is an error returned when trying to obtain a PSK for an empty file name
var ErrVaultFilenameEmpty = errors.New("vault filename required but was empty")

// ImageUploadedHandler ...
type ImageUploadedHandler struct {
	AuthToken string
	S3Upload  S3Reader
	S3Private S3Writer
	VaultCli  VaultClient
	VaultPath string
}

// S3Writer defines the required methods from dp-s3 to interact with a particular bucket of AWS S3
type S3Writer interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	Session() *session.Session
	BucketName() string
	GetWithPSK(key string, psk []byte) (io.ReadCloser, *int64, error)
}

// S3Reader defines the required methods from dp-s3 to read data to an AWS S3 Bucket
type S3Reader interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	Session() *session.Session
	BucketName() string
	Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

// VaultClient defines the required methods from dp-vault client
type VaultClient interface {
	ReadKey(path, key string) (string, error)
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}

// Handle takes a single event. It reads the PSK from Vault, uses it to decrypt the encrypted file
// from the private S3 bucket, and writes it to the public static bucket without using the vault psk for encryption.
func (h *ImageUploadedHandler) Handle(ctx context.Context, event *ImageUploaded) error {
	uploadBucket := h.S3Upload.BucketName()
	logData := log.Data{
		"event":          event,
		"upload_bucket":  uploadBucket,
		"private_bucket": h.S3Private.BucketName(),
		"vault_path":     h.VaultPath,
	}
	log.Event(ctx, "event handler called", log.INFO, logData)

	log.Event(ctx, "event successfully handled", log.INFO, logData)
	return nil
}
