package service_test

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-image-importer/config"
	"github.com/ONSdigital/dp-image-importer/event"
	eventMock "github.com/ONSdigital/dp-image-importer/event/mock"
	"github.com/ONSdigital/dp-image-importer/service"
	serviceMock "github.com/ONSdigital/dp-image-importer/service/mock"
	kafka "github.com/ONSdigital/dp-kafka"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx           = context.Background()
	testBuildTime = "BuildTime"
	testGitCommit = "GitCommit"
	testVersion   = "Version"
)

var (
	errVault         = errors.New("vault error")
	errS3Private     = errors.New("S3 private error")
	errKafkaConsumer = errors.New("Kafka consumer error")
	errHealthcheck   = errors.New("healthCheck error")
)

var funcDoGetVaultErr = func(ctx context.Context, cfg *config.Config) (event.VaultClient, error) {
	return nil, errVault
}

var funcDoS3PrivateErr = func(awsRegion, bucketName string, encryptionEnabled bool) (event.S3Writer, error) {
	return nil, errS3Private
}

var funcDoGetKafkaConsumerErr = func(ctx context.Context, cfg *config.Config) (service.KafkaConsumer, error) {
	return nil, errKafkaConsumer
}

var funcDoGetHealthcheckErr = func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
	return nil, errHealthcheck
}

var funcDoGetHTTPServerNil = func(bindAddr string, router http.Handler) service.HTTPServer {
	return nil
}

func TestRun(t *testing.T) {

	Convey("Having a set of mocked dependencies", t, func() {

		vaultMock := &eventMock.VaultClientMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		s3PrivateMock := &eventMock.S3WriterMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
			SessionFunc: func() *session.Session { return nil },
		}

		s3UploadedMock := &eventMock.S3ReaderMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		imageAPIMock := &eventMock.ImageAPIClientMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		consumerMock := &serviceMock.KafkaConsumerMock{
			CheckerFunc:  func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
			ChannelsFunc: func() *kafka.ConsumerGroupChannels { return &kafka.ConsumerGroupChannels{} },
		}

		hcMock := &serviceMock.HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(ctx context.Context) {},
		}

		serverWg := &sync.WaitGroup{}
		serverMock := &serviceMock.HTTPServerMock{
			ListenAndServeFunc: func() error {
				serverWg.Done()
				return nil
			},
		}

		funcDoGetVaultOk := func(ctx context.Context, cfg *config.Config) (event.VaultClient, error) {
			return vaultMock, nil
		}

		funcDoGetS3PrivateOk := func(awsRegion, bucketName string, encryptionEnabled bool) (event.S3Writer, error) {
			return s3PrivateMock, nil
		}

		funcDoGetS3UploadedOk := func(bucketName string, encryptionEnabled bool, s *session.Session) event.S3Reader {
			return s3UploadedMock
		}

		funcDoGetImageAPIOk := func(ctx context.Context, cfg *config.Config) event.ImageAPIClient {
			return imageAPIMock
		}

		funcDoGetKafkaConsumerOk := func(ctx context.Context, cfg *config.Config) (service.KafkaConsumer, error) {
			return consumerMock, nil
		}

		funcDoGetHealthcheckOk := func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
			return hcMock, nil
		}

		funcDoGetHTTPServer := func(bindAddr string, router http.Handler) service.HTTPServer {
			return serverMock
		}

		Convey("Given that initialising vault returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:          funcDoGetHTTPServerNil,
				DoGetVaultFunc:               funcDoGetVaultErr,
				DoGetS3ClientFunc:            funcDoGetS3PrivateOk,
				DoGetS3ClientWithSessionFunc: funcDoGetS3UploadedOk,
				DoGetImageAPIFunc:            funcDoGetImageAPIOk,
				DoGetKafkaConsumerFunc:       funcDoGetKafkaConsumerOk,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set", func() {
				So(err, ShouldResemble, errVault)
				So(svcList.Vault, ShouldBeFalse)
				So(svcList.S3Private, ShouldBeFalse)
				So(svcList.S3Uploaded, ShouldBeFalse)
				So(svcList.ImageAPI, ShouldBeFalse)
				So(svcList.KafkaConsumer, ShouldBeFalse)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that initialising s3 private bucket returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:          funcDoGetHTTPServerNil,
				DoGetVaultFunc:               funcDoGetVaultOk,
				DoGetS3ClientFunc:            funcDoS3PrivateErr,
				DoGetS3ClientWithSessionFunc: funcDoGetS3UploadedOk,
				DoGetImageAPIFunc:            funcDoGetImageAPIOk,
				DoGetKafkaConsumerFunc:       funcDoGetKafkaConsumerOk,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set", func() {
				So(err, ShouldResemble, errS3Private)
				So(svcList.Vault, ShouldBeTrue)
				So(svcList.S3Private, ShouldBeFalse)
				So(svcList.S3Uploaded, ShouldBeFalse)
				So(svcList.ImageAPI, ShouldBeFalse)
				So(svcList.KafkaConsumer, ShouldBeFalse)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that initialising Kafka consumer returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:          funcDoGetHTTPServerNil,
				DoGetVaultFunc:               funcDoGetVaultOk,
				DoGetS3ClientFunc:            funcDoGetS3PrivateOk,
				DoGetS3ClientWithSessionFunc: funcDoGetS3UploadedOk,
				DoGetImageAPIFunc:            funcDoGetImageAPIOk,
				DoGetKafkaConsumerFunc:       funcDoGetKafkaConsumerErr,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set", func() {
				So(err, ShouldResemble, errKafkaConsumer)
				So(svcList.Vault, ShouldBeTrue)
				So(svcList.S3Private, ShouldBeTrue)
				So(svcList.S3Uploaded, ShouldBeTrue)
				So(svcList.ImageAPI, ShouldBeTrue)
				So(svcList.KafkaConsumer, ShouldBeFalse)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that initialising healthcheck returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:          funcDoGetHTTPServerNil,
				DoGetVaultFunc:               funcDoGetVaultOk,
				DoGetHealthCheckFunc:         funcDoGetHealthcheckErr,
				DoGetS3ClientFunc:            funcDoGetS3PrivateOk,
				DoGetS3ClientWithSessionFunc: funcDoGetS3UploadedOk,
				DoGetImageAPIFunc:            funcDoGetImageAPIOk,
				DoGetKafkaConsumerFunc:       funcDoGetKafkaConsumerOk,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set", func() {
				So(err, ShouldResemble, errHealthcheck)
				So(svcList.Vault, ShouldBeTrue)
				So(svcList.S3Private, ShouldBeTrue)
				So(svcList.S3Uploaded, ShouldBeTrue)
				So(svcList.ImageAPI, ShouldBeTrue)
				So(svcList.KafkaConsumer, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that all dependencies are successfully initialised", func() {

			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:          funcDoGetHTTPServer,
				DoGetVaultFunc:               funcDoGetVaultOk,
				DoGetHealthCheckFunc:         funcDoGetHealthcheckOk,
				DoGetS3ClientFunc:            funcDoGetS3PrivateOk,
				DoGetS3ClientWithSessionFunc: funcDoGetS3UploadedOk,
				DoGetImageAPIFunc:            funcDoGetImageAPIOk,
				DoGetKafkaConsumerFunc:       funcDoGetKafkaConsumerOk,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			serverWg.Add(1)
			_, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run succeeds and all the flags are set", func() {
				So(err, ShouldBeNil)
				So(svcList.Vault, ShouldBeTrue)
				So(svcList.S3Private, ShouldBeTrue)
				So(svcList.S3Uploaded, ShouldBeTrue)
				So(svcList.ImageAPI, ShouldBeTrue)
				So(svcList.KafkaConsumer, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeTrue)
			})

			Convey("The checkers are registered and the healthcheck and http server started", func() {
				So(hcMock.AddCheckCalls(), ShouldHaveLength, 5)
				So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "Vault client")
				So(hcMock.AddCheckCalls()[1].Name, ShouldResemble, "S3 private bucket")
				So(hcMock.AddCheckCalls()[2].Name, ShouldResemble, "S3 uploaded bucket")
				So(hcMock.AddCheckCalls()[3].Name, ShouldResemble, "Image API client")
				So(hcMock.AddCheckCalls()[4].Name, ShouldResemble, "Kafka consumer")
				So(initMock.DoGetHTTPServerCalls(), ShouldHaveLength, 1)
				So(initMock.DoGetHTTPServerCalls()[0].BindAddr, ShouldEqual, "localhost:24800")
				So(hcMock.StartCalls(), ShouldHaveLength, 1)
				serverWg.Wait() // Wait for HTTP server go-routine to finish
				So(serverMock.ListenAndServeCalls(), ShouldHaveLength, 1)
			})
		})

		Convey("Given that Checkers cannot be registered", func() {

			errAddheckFail := errors.New("Error(s) registering checkers for healthcheck")
			hcMockAddFail := &serviceMock.HealthCheckerMock{
				AddCheckFunc: func(name string, checker healthcheck.Checker) error { return errAddheckFail },
				StartFunc:    func(ctx context.Context) {},
			}

			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc: funcDoGetHTTPServerNil,
				DoGetVaultFunc:      funcDoGetVaultOk,
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
					return hcMockAddFail, nil
				},
				DoGetS3ClientFunc:            funcDoGetS3PrivateOk,
				DoGetS3ClientWithSessionFunc: funcDoGetS3UploadedOk,
				DoGetImageAPIFunc:            funcDoGetImageAPIOk,
				DoGetKafkaConsumerFunc:       funcDoGetKafkaConsumerOk,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails, but all checks try to register", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, fmt.Sprintf("unable to register checkers: %s", errAddheckFail.Error()))
				So(svcList.Vault, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeTrue)
				So(svcList.S3Private, ShouldBeTrue)
				So(svcList.S3Uploaded, ShouldBeTrue)
				So(svcList.ImageAPI, ShouldBeTrue)
				So(svcList.KafkaConsumer, ShouldBeTrue)
				So(hcMockAddFail.AddCheckCalls(), ShouldHaveLength, 5)
				So(hcMockAddFail.AddCheckCalls()[0].Name, ShouldResemble, "Vault client")
				So(hcMockAddFail.AddCheckCalls()[1].Name, ShouldResemble, "S3 private bucket")
				So(hcMockAddFail.AddCheckCalls()[2].Name, ShouldResemble, "S3 uploaded bucket")
				So(hcMockAddFail.AddCheckCalls()[3].Name, ShouldResemble, "Image API client")
				So(hcMockAddFail.AddCheckCalls()[4].Name, ShouldResemble, "Kafka consumer")
			})
		})
	})
}

func TestClose(t *testing.T) {

	Convey("Having a correctly initialised service", t, func() {

		hcStopped := false

		vaultMock := &eventMock.VaultClientMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		s3PrivateMock := &eventMock.S3WriterMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
			SessionFunc: func() *session.Session { return nil },
		}

		s3UploadedMock := &eventMock.S3ReaderMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		imageAPIMock := &eventMock.ImageAPIClientMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		consumerMock := &serviceMock.KafkaConsumerMock{
			StopListeningToConsumerFunc: func(ctx context.Context) error { return nil },
			CloseFunc:                   func(ctx context.Context) error { return nil },
			CheckerFunc:                 func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
			ChannelsFunc:                func() *kafka.ConsumerGroupChannels { return &kafka.ConsumerGroupChannels{} },
		}

		// healthcheck Stop does not depend on any other service being closed/stopped
		hcMock := &serviceMock.HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(ctx context.Context) {},
			StopFunc:     func() { hcStopped = true },
		}

		// server Shutdown will fail if healthcheck is not stopped
		serverMock := &serviceMock.HTTPServerMock{
			ListenAndServeFunc: func() error { return nil },
			ShutdownFunc: func(ctx context.Context) error {
				if !hcStopped {
					return errors.New("Server stopped before healthcheck")
				}
				return nil
			},
		}

		Convey("Closing the service results in all the dependencies being closed in the expected order", func() {

			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer { return serverMock },
				DoGetVaultFunc:      func(ctx context.Context, cfg *config.Config) (event.VaultClient, error) { return vaultMock, nil },
				DoGetS3ClientFunc: func(awsRegion, bucketName string, encryptionEnabled bool) (event.S3Writer, error) {
					return s3PrivateMock, nil
				},
				DoGetS3ClientWithSessionFunc: func(bucketName string, encryptionEnabled bool, s *session.Session) event.S3Reader {
					return s3UploadedMock
				},
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
					return hcMock, nil
				},
				DoGetImageAPIFunc:      func(ctx context.Context, cfg *config.Config) event.ImageAPIClient { return imageAPIMock },
				DoGetKafkaConsumerFunc: func(ctx context.Context, cfg *config.Config) (service.KafkaConsumer, error) { return consumerMock, nil },
			}

			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)
			So(err, ShouldBeNil)

			err = svc.Close(context.Background())
			So(err, ShouldBeNil)
			So(consumerMock.StopListeningToConsumerCalls(), ShouldHaveLength, 1)
			So(hcMock.StopCalls(), ShouldHaveLength, 1)
			So(consumerMock.CloseCalls(), ShouldHaveLength, 1)
			So(serverMock.ShutdownCalls(), ShouldHaveLength, 1)
		})

		Convey("If services fail to stop, the Close operation tries to close all dependencies and returns an error", func() {

			failingserverMock := &serviceMock.HTTPServerMock{
				ListenAndServeFunc: func() error { return nil },
				ShutdownFunc: func(ctx context.Context) error {
					return errors.New("Failed to stop http server")
				},
			}

			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer { return failingserverMock },
				DoGetVaultFunc:      func(ctx context.Context, cfg *config.Config) (event.VaultClient, error) { return vaultMock, nil },
				DoGetS3ClientFunc: func(awsRegion, bucketName string, encryptionEnabled bool) (event.S3Writer, error) {
					return s3PrivateMock, nil
				},
				DoGetS3ClientWithSessionFunc: func(bucketName string, encryptionEnabled bool, s *session.Session) event.S3Reader {
					return s3UploadedMock
				},
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
					return hcMock, nil
				},
				DoGetImageAPIFunc:      func(ctx context.Context, cfg *config.Config) event.ImageAPIClient { return imageAPIMock },
				DoGetKafkaConsumerFunc: func(ctx context.Context, cfg *config.Config) (service.KafkaConsumer, error) { return consumerMock, nil },
			}

			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)
			So(err, ShouldBeNil)

			err = svc.Close(context.Background())
			So(err, ShouldNotBeNil)
			So(hcMock.StopCalls(), ShouldHaveLength, 1)
			So(failingserverMock.ShutdownCalls(), ShouldHaveLength, 1)
		})
	})
}
