package event_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/image"
	"github.com/ONSdigital/dp-image-importer/event"
	"github.com/ONSdigital/dp-image-importer/event/mock"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testVaultPath            = "/vault/path/for/testing"
	testVaultUploadFilePath  = "/vault/path/for/testing/1234-uploadpng"
	testVaultPrivateFilePath = "/vault/path/for/testing/images/123/original"
	testAuthToken            = "auth-123"
	testDownloadURL          = "http://some.download.server"
)

var (
	testPrivateBucket        = "privateBucket"
	testUploadedBucket       = "uploadedBucket"
	testPrivatePath          = "images/123/original"
	testSize           int64 = 1234
	fileBytes                = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	testFileContent          = ioutil.NopCloser(bytes.NewReader(fileBytes))
	encodedPSK               = "48656C6C6F20576F726C64"
	errVault                 = errors.New("vault error")
	errS3Private             = errors.New("s3Private error")
	errS3Uploaded            = errors.New("s3Uploaded error")
	errImageAPI              = errors.New("imageAPI error")

	testImportStarted   = time.Date(2020, time.April, 26, 8, 5, 52, 0, time.UTC)
	testImportCompleted = time.Date(2020, time.April, 26, 8, 7, 32, 0, time.UTC)

	testCreatedDownload = image.ImageDownload{
		Id:            "original",
		State:         "importing",
		ImportStarted: &testImportStarted,
	}
	testImportedDownload = image.ImageDownload{
		Id:              "original",
		State:           "imported",
		ImportStarted:   &testImportStarted,
		ImportCompleted: &testImportCompleted,
	}
)

var testEventNoFilename = event.ImageUploaded{
	ImageID: "123",
	Path:    "1234-uploadpng",
}

var testEventNoPath = event.ImageUploaded{
	ImageID:  "123",
	Filename: "Pathless.png",
}

func TestImageUploadedHandler_Handle(t *testing.T) {

	Convey("Given S3 and Vault client mocks", t, func() {

		mockS3Private := &mock.S3WriterMock{
			BucketNameFunc: func() string {
				return testPrivateBucket
			},
		}
		mockS3Upload := &mock.S3ReaderMock{
			BucketNameFunc: func() string {
				return testUploadedBucket
			},
		}
		mockVault := &mock.VaultClientMock{
			ReadKeyFunc: func(path string, key string) (string, error) {
				return encodedPSK, nil
			},
			WriteKeyFunc: func(path string, key string, value string) error {
				return nil
			},
		}
		mockImageAPI := &mock.ImageAPIClientMock{
			PostDownloadVariantFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string, data image.NewImageDownload) (image.ImageDownload, error) {
				return testCreatedDownload, nil
			},
			PutDownloadVariantFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string, variant string, data image.ImageDownload) (image.ImageDownload, error) {
				return testImportedDownload, nil
			},
		}
		psk, err := hex.DecodeString(encodedPSK)
		So(err, ShouldBeNil)

		Convey("And a successful event handler, when Handle is triggered", func() {
			mockS3Upload.GetWithPSKFunc = func(key string, psk []byte) (io.ReadCloser, *int64, error) {
				return testFileContent, &testSize, nil
			}
			mockS3Private.UploadWithPSKFunc = func(input *s3manager.UploadInput, psk []byte) (*s3manager.UploadOutput, error) {
				return &s3manager.UploadOutput{}, nil
			}
			eventHandler := event.ImageUploadedHandler{
				AuthToken:          testAuthToken,
				S3Upload:           mockS3Upload,
				S3Private:          mockS3Private,
				VaultCli:           mockVault,
				VaultPath:          testVaultPath,
				ImageCli:           mockImageAPI,
				DownloadServiceURL: testDownloadURL,
			}
			err := eventHandler.Handle(testCtx, &testEvent)
			So(err, ShouldBeNil)

			Convey("An image download variant is posted to the image API", func() {
				So(len(mockImageAPI.PostDownloadVariantCalls()), ShouldEqual, 1)
				So(mockImageAPI.PostDownloadVariantCalls()[0].ImageID, ShouldEqual, testEvent.ImageID)
				So(mockImageAPI.PostDownloadVariantCalls()[0].ServiceAuthToken, ShouldResemble, testAuthToken)
				newImageData := mockImageAPI.PostDownloadVariantCalls()[0].Data
				So(newImageData, ShouldNotBeNil)
				So(newImageData.Id, ShouldEqual, "original")
				So(newImageData.State, ShouldEqual, "importing")
			})

			Convey("Encryption key is read from Vault with the expected path", func() {
				So(len(mockVault.ReadKeyCalls()), ShouldEqual, 1)
				So(mockVault.ReadKeyCalls()[0].Path, ShouldEqual, testVaultUploadFilePath)
				So(mockVault.ReadKeyCalls()[0].Key, ShouldEqual, "key")
			})

			Convey("The file is obtained from the private bucket and decrypted with the psk obtained from Vault", func() {
				So(len(mockS3Upload.GetWithPSKCalls()), ShouldEqual, 1)
				So(mockS3Upload.GetWithPSKCalls()[0].Key, ShouldEqual, testEvent.Path)
				So(mockS3Upload.GetWithPSKCalls()[0].Psk, ShouldResemble, psk)
			})

			Convey("Encryption key is written to Vault with the expected path", func() {
				So(len(mockVault.WriteKeyCalls()), ShouldEqual, 1)
				So(mockVault.WriteKeyCalls()[0].Path, ShouldEqual, testVaultPrivateFilePath)
				So(mockVault.WriteKeyCalls()[0].Key, ShouldEqual, "key")
			})

			Convey("The file is uploaded to the private bucket", func() {
				So(len(mockS3Private.UploadWithPSKCalls()), ShouldEqual, 1)
				So(mockS3Private.UploadWithPSKCalls()[0].Psk, ShouldHaveLength, 16)
				So(*mockS3Private.UploadWithPSKCalls()[0].Input, ShouldResemble, s3manager.UploadInput{
					Body:   testFileContent,
					Bucket: &testPrivateBucket,
					Key:    &testPrivatePath,
				})
			})

			Convey("The image download variant is put to the image API with a state of imported", func() {
				So(len(mockImageAPI.PutDownloadVariantCalls()), ShouldEqual, 1)
				So(mockImageAPI.PutDownloadVariantCalls()[0].ImageID, ShouldEqual, testEvent.ImageID)
				So(mockImageAPI.PutDownloadVariantCalls()[0].ServiceAuthToken, ShouldResemble, testAuthToken)
				newImageData := mockImageAPI.PutDownloadVariantCalls()[0].Data
				So(newImageData, ShouldNotBeNil)
				So(newImageData.Id, ShouldEqual, "original")
				So(newImageData.State, ShouldEqual, "imported")
				So(newImageData.ImportCompleted, ShouldNotBeNil)
				So(newImageData.Href, ShouldEqual, testDownloadURL+"/"+testPrivatePath+"/"+testEvent.Filename)
			})
		})

		Convey("And an event with no filename supplied, when Handle is triggered", func() {
			mockS3Upload.GetWithPSKFunc = func(key string, psk []byte) (io.ReadCloser, *int64, error) {
				return testFileContent, &testSize, nil
			}
			mockS3Private.UploadWithPSKFunc = func(input *s3manager.UploadInput, psk []byte) (*s3manager.UploadOutput, error) {
				return &s3manager.UploadOutput{}, nil
			}
			eventHandler := event.ImageUploadedHandler{
				AuthToken:          testAuthToken,
				S3Upload:           mockS3Upload,
				S3Private:          mockS3Private,
				VaultCli:           mockVault,
				VaultPath:          testVaultPath,
				ImageCli:           mockImageAPI,
				DownloadServiceURL: testDownloadURL,
			}
			err := eventHandler.Handle(testCtx, &testEventNoFilename)
			So(err, ShouldBeNil)

			Convey("The image download variant is put to the image API with a state of imported", func() {
				So(len(mockImageAPI.PostDownloadVariantCalls()), ShouldEqual, 1)
				So(mockImageAPI.PostDownloadVariantCalls()[0].ImageID, ShouldEqual, testEvent.ImageID)
				So(mockImageAPI.PostDownloadVariantCalls()[0].ServiceAuthToken, ShouldResemble, testAuthToken)
				newImageData := mockImageAPI.PutDownloadVariantCalls()[0].Data
				So(newImageData, ShouldNotBeNil)
				So(newImageData.Id, ShouldEqual, "original")
				So(newImageData.State, ShouldEqual, "imported")
				So(newImageData.Href, ShouldEqual, testDownloadURL+"/"+testPrivatePath+"/"+testEventNoFilename.Path)
			})
		})

		Convey("And a nil-vault event handler (developer env), when Handle is triggered", func() {
			mockS3Upload.GetFunc = func(key string) (io.ReadCloser, *int64, error) {
				return testFileContent, &testSize, nil
			}
			mockS3Private.UploadFunc = func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
				return &s3manager.UploadOutput{}, nil
			}
			eventHandler := event.ImageUploadedHandler{
				AuthToken:          testAuthToken,
				S3Upload:           mockS3Upload,
				S3Private:          mockS3Private,
				VaultCli:           nil,
				VaultPath:          "",
				ImageCli:           mockImageAPI,
				DownloadServiceURL: testDownloadURL,
			}
			err := eventHandler.Handle(testCtx, &testEvent)
			So(err, ShouldBeNil)

			Convey("An image download variant is posted to the image API", func() {
				So(len(mockImageAPI.PostDownloadVariantCalls()), ShouldEqual, 1)
				So(mockImageAPI.PostDownloadVariantCalls()[0].ImageID, ShouldEqual, testEvent.ImageID)
				So(mockImageAPI.PostDownloadVariantCalls()[0].ServiceAuthToken, ShouldResemble, testAuthToken)
				newImageData := mockImageAPI.PostDownloadVariantCalls()[0].Data
				So(newImageData, ShouldNotBeNil)
				So(newImageData.Id, ShouldEqual, "original")
				So(newImageData.State, ShouldEqual, "importing")
			})

			Convey("The file is obtained from the private bucket", func() {
				So(len(mockS3Upload.GetCalls()), ShouldEqual, 1)
				So(mockS3Upload.GetCalls()[0].Key, ShouldEqual, testEvent.Path)
			})

			Convey("The file is uploaded to the private bucket", func() {
				So(len(mockS3Private.UploadCalls()), ShouldEqual, 1)
				So(*mockS3Private.UploadCalls()[0].Input, ShouldResemble, s3manager.UploadInput{
					Body:   testFileContent,
					Bucket: &testPrivateBucket,
					Key:    &testPrivatePath,
				})
			})

			Convey("The image download variant is put to the image API with a state of imported", func() {
				So(len(mockImageAPI.PutDownloadVariantCalls()), ShouldEqual, 1)
				So(mockImageAPI.PutDownloadVariantCalls()[0].ImageID, ShouldEqual, testEvent.ImageID)
				So(mockImageAPI.PutDownloadVariantCalls()[0].ServiceAuthToken, ShouldResemble, testAuthToken)
				newImageData := mockImageAPI.PutDownloadVariantCalls()[0].Data
				So(newImageData, ShouldNotBeNil)
				So(newImageData.Id, ShouldEqual, "original")
				So(newImageData.State, ShouldEqual, "imported")
				So(newImageData.ImportCompleted, ShouldNotBeNil)
				So(newImageData.Href, ShouldEqual, testDownloadURL+"/"+testPrivatePath+"/"+testEvent.Filename)
			})
		})

		Convey("And an event handler with a failing vault client, when Handle is triggered", func() {
			mockVaultFail := &mock.VaultClientMock{
				ReadKeyFunc: func(path string, key string) (string, error) {
					return "", errVault
				},
			}
			eventHandler := event.ImageUploadedHandler{
				AuthToken: testAuthToken,
				S3Upload:  mockS3Upload,
				S3Private: mockS3Private,
				VaultCli:  mockVaultFail,
				VaultPath: testVaultPath,
			}
			err := eventHandler.Handle(testCtx, &testEvent)

			Convey("Vault ReadKey is called and the error is returned", func() {
				So(err, ShouldResemble, errVault)
				So(len(mockVaultFail.ReadKeyCalls()), ShouldEqual, 1)
			})
		})

		Convey("And an event handler with a vault client that returns an invalid psk, when Handle is triggered", func() {
			mockVaultFail := &mock.VaultClientMock{
				ReadKeyFunc: func(path string, key string) (string, error) {
					return "invalidPSK", nil
				},
			}
			eventHandler := event.ImageUploadedHandler{
				AuthToken:          testAuthToken,
				S3Upload:           mockS3Upload,
				S3Private:          mockS3Private,
				VaultCli:           mockVaultFail,
				VaultPath:          testVaultPath,
				ImageCli:           mockImageAPI,
				DownloadServiceURL: testDownloadURL,
			}
			err := eventHandler.Handle(testCtx, &testEvent)

			Convey("Vault ReadKey is called and the decoding error is returned", func() {
				So(err, ShouldNotBeNil)
				So(len(mockVaultFail.ReadKeyCalls()), ShouldEqual, 1)
			})
		})

		Convey("And an event handler with an S3Uploaded client that fails to obtain the source file, when Handle is triggered", func() {
			mockS3Upload.GetWithPSKFunc = func(key string, psk []byte) (io.ReadCloser, *int64, error) {
				return nil, nil, errS3Uploaded
			}
			eventHandler := event.ImageUploadedHandler{
				AuthToken:          testAuthToken,
				S3Upload:           mockS3Upload,
				S3Private:          mockS3Private,
				VaultCli:           mockVault,
				VaultPath:          testVaultPath,
				ImageCli:           mockImageAPI,
				DownloadServiceURL: testDownloadURL,
			}
			err := eventHandler.Handle(testCtx, &testEvent)

			Convey("S3Private is called and the same error is returned", func() {
				So(err, ShouldResemble, errS3Uploaded)
				So(len(mockVault.ReadKeyCalls()), ShouldEqual, 1)
				So(len(mockS3Upload.GetWithPSKCalls()), ShouldEqual, 1)
			})
		})

		Convey("And a nil-vault event handler (developer env) with an S3Uploaded client that fails to obtain the source file, when Handle is triggered", func() {
			mockS3Upload.GetFunc = func(key string) (io.ReadCloser, *int64, error) {
				return nil, nil, errS3Uploaded
			}
			eventHandler := event.ImageUploadedHandler{
				AuthToken:          testAuthToken,
				S3Upload:           mockS3Upload,
				S3Private:          mockS3Private,
				VaultCli:           nil,
				VaultPath:          "",
				ImageCli:           mockImageAPI,
				DownloadServiceURL: testDownloadURL,
			}
			err := eventHandler.Handle(testCtx, &testEvent)

			Convey("S3Private is called and the same error is returned", func() {
				So(err, ShouldResemble, errS3Uploaded)
				So(len(mockS3Upload.GetCalls()), ShouldEqual, 1)
			})
		})

		Convey("And an event handler with an image client that fails to create a new variant, when Handle is triggered", func() {
			mockS3Upload.GetWithPSKFunc = func(key string, psk []byte) (io.ReadCloser, *int64, error) {
				return testFileContent, &testSize, nil
			}
			mockImageAPIFail := &mock.ImageAPIClientMock{
				PostDownloadVariantFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string, data image.NewImageDownload) (image.ImageDownload, error) {
					return image.ImageDownload{}, errImageAPI
				},
			}
			eventHandler := event.ImageUploadedHandler{
				AuthToken:          testAuthToken,
				S3Upload:           mockS3Upload,
				S3Private:          mockS3Private,
				VaultCli:           mockVault,
				VaultPath:          testVaultPath,
				ImageCli:           mockImageAPIFail,
				DownloadServiceURL: testDownloadURL,
			}
			err := eventHandler.Handle(testCtx, &testEvent)

			Convey("ImageAPI.PostDownloadVariant is called and the error is returned", func() {
				So(err, ShouldNotBeNil)
				So(len(mockImageAPIFail.PostDownloadVariantCalls()), ShouldEqual, 1)
			})
		})

		Convey("And an event handler with a vault client that fails to write, when Handle is triggered", func() {
			mockS3Upload.GetWithPSKFunc = func(key string, psk []byte) (io.ReadCloser, *int64, error) {
				return testFileContent, &testSize, nil
			}
			mockVaultFail := &mock.VaultClientMock{
				ReadKeyFunc: mockVault.ReadKeyFunc,
				WriteKeyFunc: func(path string, key string, value string) error {
					return errVault
				},
			}
			eventHandler := event.ImageUploadedHandler{
				AuthToken:          testAuthToken,
				S3Upload:           mockS3Upload,
				S3Private:          mockS3Private,
				VaultCli:           mockVaultFail,
				VaultPath:          testVaultPath,
				ImageCli:           mockImageAPI,
				DownloadServiceURL: testDownloadURL,
			}
			err := eventHandler.Handle(testCtx, &testEvent)

			Convey("Vault ReadKey is called and the decoding error is returned", func() {
				So(err, ShouldNotBeNil)
				So(len(mockVaultFail.WriteKeyCalls()), ShouldEqual, 1)
			})
		})

		Convey("And an event handler with an S3Private client that fails to upload the file, when Handle is triggered", func() {
			mockS3Upload.GetWithPSKFunc = func(key string, psk []byte) (io.ReadCloser, *int64, error) {
				return testFileContent, &testSize, nil
			}
			mockS3Private.UploadWithPSKFunc = func(input *s3manager.UploadInput, psk []byte) (*s3manager.UploadOutput, error) {
				return nil, errS3Private
			}
			eventHandler := event.ImageUploadedHandler{
				AuthToken:          testAuthToken,
				S3Upload:           mockS3Upload,
				S3Private:          mockS3Private,
				VaultCli:           mockVault,
				VaultPath:          testVaultPath,
				ImageCli:           mockImageAPI,
				DownloadServiceURL: testDownloadURL,
			}
			err := eventHandler.Handle(testCtx, &testEvent)

			Convey("S3Private is called and the same error is returned", func() {
				So(err, ShouldResemble, errS3Private)
				So(len(mockVault.ReadKeyCalls()), ShouldEqual, 1)
				So(len(mockS3Upload.GetWithPSKCalls()), ShouldEqual, 1)
				So(len(mockS3Private.BucketNameCalls()), ShouldEqual, 2)
			})
		})

		Convey("And a nil-vault event handler (developer env) with an S3Private client that fails to upload the file, when Handle is triggered", func() {
			mockS3Upload.GetFunc = func(key string) (io.ReadCloser, *int64, error) {
				return testFileContent, &testSize, nil
			}
			mockS3Private.UploadFunc = func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
				return nil, errS3Private
			}
			eventHandler := event.ImageUploadedHandler{
				AuthToken:          testAuthToken,
				S3Upload:           mockS3Upload,
				S3Private:          mockS3Private,
				VaultCli:           nil,
				VaultPath:          "",
				ImageCli:           mockImageAPI,
				DownloadServiceURL: testDownloadURL,
			}
			err := eventHandler.Handle(testCtx, &testEvent)

			Convey("S3Private is called and the same error is returned", func() {
				So(err, ShouldResemble, errS3Private)
				So(len(mockS3Upload.GetCalls()), ShouldEqual, 1)
				So(len(mockS3Private.BucketNameCalls()), ShouldEqual, 2)
			})
		})

		Convey("And an event handler with an image client that fails to update a variant, when Handle is triggered", func() {
			mockS3Upload.GetWithPSKFunc = func(key string, psk []byte) (io.ReadCloser, *int64, error) {
				return testFileContent, &testSize, nil
			}
			mockS3Private.UploadWithPSKFunc = func(input *s3manager.UploadInput, psk []byte) (*s3manager.UploadOutput, error) {
				return &s3manager.UploadOutput{}, nil
			}
			mockImageAPIFail := &mock.ImageAPIClientMock{
				PostDownloadVariantFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string, data image.NewImageDownload) (image.ImageDownload, error) {
					return testCreatedDownload, nil
				},
				PutDownloadVariantFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string, variant string, data image.ImageDownload) (image.ImageDownload, error) {
					return image.ImageDownload{}, errImageAPI
				},
			}
			eventHandler := event.ImageUploadedHandler{
				AuthToken:          testAuthToken,
				S3Upload:           mockS3Upload,
				S3Private:          mockS3Private,
				VaultCli:           mockVault,
				VaultPath:          testVaultPath,
				ImageCli:           mockImageAPIFail,
				DownloadServiceURL: testDownloadURL,
			}
			err := eventHandler.Handle(testCtx, &testEvent)

			Convey("ImageAPI.PutDownloadVariant is called and the error is returned", func() {
				So(err, ShouldNotBeNil)
				So(len(mockImageAPIFail.PutDownloadVariantCalls()), ShouldEqual, 1)
			})
		})

		Convey("And an event with no path supplied, when Handle is triggered", func() {
			eventHandler := event.ImageUploadedHandler{
				AuthToken:          testAuthToken,
				S3Upload:           mockS3Upload,
				S3Private:          mockS3Private,
				VaultCli:           mockVault,
				VaultPath:          testVaultPath,
				ImageCli:           mockImageAPI,
				DownloadServiceURL: testDownloadURL,
			}
			err := eventHandler.Handle(testCtx, &testEventNoPath)

			Convey("The image download variant is put to the image API with a state of imported", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, "vault filename required but was empty")
			})
		})

	})

}
