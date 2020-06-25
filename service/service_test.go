package service_test

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-image-importer/api"
	apiMock "github.com/ONSdigital/dp-image-importer/api/mock"
	"github.com/ONSdigital/dp-image-importer/config"
	"github.com/ONSdigital/dp-image-importer/service"
	"github.com/ONSdigital/dp-image-importer/service/mock"
	serviceMock "github.com/ONSdigital/dp-image-importer/service/mock"
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
	errS3Uploaded    = errors.New("S3 uploaded error")
	errKafkaConsumer = errors.New("Kafka consumer error")
	errHealthcheck   = errors.New("healthCheck error")
)

var funcDoGetVaultErr = func(ctx context.Context, cfg *config.Config) (api.VaultClienter, error) {
	return nil, errVault
}

var funcDoS3PrivateErr = func(ctx context.Context, cfg *config.Config) (api.S3Clienter, error) {
	return nil, errS3Private
}

var funcDoS3UploadedErr = func(ctx context.Context, cfg *config.Config) (api.S3Clienter, error) {
	return nil, errS3Uploaded
}

var funcDoGetKafkaConsumerErr = func(ctx context.Context, cfg *config.Config) (api.KafkaConsumer, error) {
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

		vaultMock := &apiMock.VaultClienterMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		s3PrivateMock := &apiMock.S3ClienterMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		s3UploadedMock := &apiMock.S3ClienterMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		imageAPIMock := &apiMock.ImageAPIClienterMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		consumerMock := &apiMock.KafkaConsumerMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
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

		funcDoGetVaultOk := func(ctx context.Context, cfg *config.Config) (api.VaultClienter, error) {
			return vaultMock, nil
		}

		funcDoGetS3PrivateOk := func(ctx context.Context, cfg *config.Config) (api.S3Clienter, error) {
			return s3PrivateMock, nil
		}

		funcDoGetS3UploadedOk := func(ctx context.Context, cfg *config.Config) (api.S3Clienter, error) {
			return s3UploadedMock, nil
		}

		funcDoGetImageAPIOk := func(ctx context.Context, cfg *config.Config) api.ImageAPIClienter {
			return imageAPIMock
		}

		funcDoGetKafkaConsumerOk := func(ctx context.Context, cfg *config.Config) (api.KafkaConsumer, error) {
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
				DoGetHTTPServerFunc:    funcDoGetHTTPServerNil,
				DoGetVaultFunc:         funcDoGetVaultErr,
				DoGetS3PrivateFunc:     funcDoGetS3PrivateOk,
				DoGetS3UploadedFunc:    funcDoGetS3UploadedOk,
				DoGetImageAPIFunc:      funcDoGetImageAPIOk,
				DoGetKafkaConsumerFunc: funcDoGetKafkaConsumerOk,
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

		Convey("Given that initialising s3 uploaded bucket returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:    funcDoGetHTTPServerNil,
				DoGetVaultFunc:         funcDoGetVaultOk,
				DoGetS3PrivateFunc:     funcDoGetS3PrivateOk,
				DoGetS3UploadedFunc:    funcDoS3UploadedErr,
				DoGetImageAPIFunc:      funcDoGetImageAPIOk,
				DoGetKafkaConsumerFunc: funcDoGetKafkaConsumerOk,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set", func() {
				So(err, ShouldResemble, errS3Uploaded)
				So(svcList.Vault, ShouldBeTrue)
				So(svcList.S3Private, ShouldBeTrue)
				So(svcList.S3Uploaded, ShouldBeFalse)
				So(svcList.ImageAPI, ShouldBeFalse)
				So(svcList.KafkaConsumer, ShouldBeFalse)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that initialising s3 private bucket returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:    funcDoGetHTTPServerNil,
				DoGetVaultFunc:         funcDoGetVaultOk,
				DoGetS3PrivateFunc:     funcDoS3PrivateErr,
				DoGetS3UploadedFunc:    funcDoGetS3UploadedOk,
				DoGetImageAPIFunc:      funcDoGetImageAPIOk,
				DoGetKafkaConsumerFunc: funcDoGetKafkaConsumerOk,
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
				DoGetHTTPServerFunc:    funcDoGetHTTPServerNil,
				DoGetVaultFunc:         funcDoGetVaultOk,
				DoGetS3PrivateFunc:     funcDoGetS3PrivateOk,
				DoGetS3UploadedFunc:    funcDoGetS3UploadedOk,
				DoGetImageAPIFunc:      funcDoGetImageAPIOk,
				DoGetKafkaConsumerFunc: funcDoGetKafkaConsumerErr,
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
				DoGetHTTPServerFunc:    funcDoGetHTTPServerNil,
				DoGetVaultFunc:         funcDoGetVaultOk,
				DoGetHealthCheckFunc:   funcDoGetHealthcheckErr,
				DoGetS3PrivateFunc:     funcDoGetS3PrivateOk,
				DoGetS3UploadedFunc:    funcDoGetS3UploadedOk,
				DoGetImageAPIFunc:      funcDoGetImageAPIOk,
				DoGetKafkaConsumerFunc: funcDoGetKafkaConsumerOk,
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
				DoGetHTTPServerFunc:    funcDoGetHTTPServer,
				DoGetVaultFunc:         funcDoGetVaultOk,
				DoGetHealthCheckFunc:   funcDoGetHealthcheckOk,
				DoGetS3PrivateFunc:     funcDoGetS3PrivateOk,
				DoGetS3UploadedFunc:    funcDoGetS3UploadedOk,
				DoGetImageAPIFunc:      funcDoGetImageAPIOk,
				DoGetKafkaConsumerFunc: funcDoGetKafkaConsumerOk,
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
				So(len(hcMock.AddCheckCalls()), ShouldEqual, 5)
				So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "Vault client")
				So(hcMock.AddCheckCalls()[1].Name, ShouldResemble, "S3 private bucket")
				So(hcMock.AddCheckCalls()[2].Name, ShouldResemble, "S3 uploaded bucket")
				So(hcMock.AddCheckCalls()[3].Name, ShouldResemble, "Image API client")
				So(hcMock.AddCheckCalls()[4].Name, ShouldResemble, "Kafka consumer")
				So(len(initMock.DoGetHTTPServerCalls()), ShouldEqual, 1)
				So(initMock.DoGetHTTPServerCalls()[0].BindAddr, ShouldEqual, "localhost:24800")
				So(len(hcMock.StartCalls()), ShouldEqual, 1)
				serverWg.Wait() // Wait for HTTP server go-routine to finish
				So(len(serverMock.ListenAndServeCalls()), ShouldEqual, 1)
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
				DoGetS3PrivateFunc:     funcDoGetS3PrivateOk,
				DoGetS3UploadedFunc:    funcDoGetS3UploadedOk,
				DoGetImageAPIFunc:      funcDoGetImageAPIOk,
				DoGetKafkaConsumerFunc: funcDoGetKafkaConsumerOk,
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
				So(len(hcMockAddFail.AddCheckCalls()), ShouldEqual, 5)
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

		vaultMock := &apiMock.VaultClienterMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		s3PrivateMock := &apiMock.S3ClienterMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		s3UploadedMock := &apiMock.S3ClienterMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		imageAPIMock := &apiMock.ImageAPIClienterMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		consumerMock := &apiMock.KafkaConsumerMock{
			StopListeningToConsumerFunc: func(ctx context.Context) error { return nil },
			CloseFunc:                   func(ctx context.Context) error { return nil },
			CheckerFunc:                 func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		// healthcheck Stop does not depend on any other service being closed/stopped
		hcMock := &serviceMock.HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(ctx context.Context) {},
			StopFunc:     func() { hcStopped = true },
		}

		// server Shutdown will fail if healthcheck is not stopped
		serverMock := &mock.HTTPServerMock{
			ListenAndServeFunc: func() error { return nil },
			ShutdownFunc: func(ctx context.Context) error {
				if !hcStopped {
					return errors.New("Server stopped before healthcheck")
				}
				return nil
			},
		}

		Convey("Closing the service results in all the dependencies being closed in the expected order", func() {

			initMock := &mock.InitialiserMock{
				DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer { return serverMock },
				DoGetVaultFunc:      func(ctx context.Context, cfg *config.Config) (api.VaultClienter, error) { return vaultMock, nil },
				DoGetS3PrivateFunc:  func(ctx context.Context, cfg *config.Config) (api.S3Clienter, error) { return s3PrivateMock, nil },
				DoGetS3UploadedFunc: func(ctx context.Context, cfg *config.Config) (api.S3Clienter, error) { return s3UploadedMock, nil },
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
					return hcMock, nil
				},
				DoGetImageAPIFunc:      func(ctx context.Context, cfg *config.Config) api.ImageAPIClienter { return imageAPIMock },
				DoGetKafkaConsumerFunc: func(ctx context.Context, cfg *config.Config) (api.KafkaConsumer, error) { return consumerMock, nil },
			}

			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)
			So(err, ShouldBeNil)

			err = svc.Close(context.Background())
			So(err, ShouldBeNil)
			So(len(consumerMock.StopListeningToConsumerCalls()), ShouldEqual, 1)
			So(len(hcMock.StopCalls()), ShouldEqual, 1)
			So(len(consumerMock.CloseCalls()), ShouldEqual, 1)
			So(len(serverMock.ShutdownCalls()), ShouldEqual, 1)
		})

		Convey("If services fail to stop, the Close operation tries to close all dependencies and returns an error", func() {

			failingserverMock := &mock.HTTPServerMock{
				ListenAndServeFunc: func() error { return nil },
				ShutdownFunc: func(ctx context.Context) error {
					return errors.New("Failed to stop http server")
				},
			}

			initMock := &mock.InitialiserMock{
				DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer { return failingserverMock },
				DoGetVaultFunc:      func(ctx context.Context, cfg *config.Config) (api.VaultClienter, error) { return vaultMock, nil },
				DoGetS3PrivateFunc:  func(ctx context.Context, cfg *config.Config) (api.S3Clienter, error) { return s3PrivateMock, nil },
				DoGetS3UploadedFunc: func(ctx context.Context, cfg *config.Config) (api.S3Clienter, error) { return s3UploadedMock, nil },
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
					return hcMock, nil
				},
				DoGetImageAPIFunc:      func(ctx context.Context, cfg *config.Config) api.ImageAPIClienter { return imageAPIMock },
				DoGetKafkaConsumerFunc: func(ctx context.Context, cfg *config.Config) (api.KafkaConsumer, error) { return consumerMock, nil },
			}

			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)
			So(err, ShouldBeNil)

			err = svc.Close(context.Background())
			So(err, ShouldNotBeNil)
			So(len(hcMock.StopCalls()), ShouldEqual, 1)
			So(len(failingserverMock.ShutdownCalls()), ShouldEqual, 1)
		})
	})
}
