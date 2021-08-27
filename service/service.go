package service

import (
	"context"

	"github.com/ONSdigital/dp-image-importer/config"
	"github.com/ONSdigital/dp-image-importer/event"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Service contains all the configs, server and clients to run the Image API
type Service struct {
	config      *config.Config
	server      HTTPServer
	router      *mux.Router
	serviceList *ExternalServiceList
	healthCheck HealthChecker
	vault       event.VaultClient
	consumer    kafka.IConsumerGroup
}

// Run the service
func Run(ctx context.Context, serviceList *ExternalServiceList, buildTime, gitCommit, version string, svcErrors chan error) (*Service, error) {
	log.Info(ctx, "running service")

	// Read config
	cfg, err := config.Get()
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve service configuration")
	}
	log.Info(ctx, "got service configuration", log.Data{"config": cfg})

	// Get HTTP Server with collectionID checkHeader middleware
	r := mux.NewRouter()
	s := serviceList.GetHTTPServer(cfg.BindAddr, r)

	// Get Vault client
	vault, err := serviceList.GetVault(ctx, cfg)
	if err != nil {
		log.Fatal(ctx, "failed to initialise Vault client", err)
		return nil, err
	}

	// Get S3 Clients
	s3Uploaded, s3Private, err := serviceList.GetS3Clients(cfg)
	if err != nil {
		log.Fatal(ctx, "could not instantiate S3 clients", err)
		return nil, err
	}

	// Get Image API Client
	imageAPI := serviceList.GetImageAPI(ctx, cfg)

	// Get Kafka consumer
	consumer, err := serviceList.GetKafkaConsumer(ctx, cfg)
	if err != nil {
		log.Fatal(ctx, "failed to initialise kafka consumer", err)
		return nil, err
	}

	// Event Handler for Kafka Consumer with the created S3 Clients and Vault
	event.Consume(ctx, consumer, &event.ImageUploadedHandler{
		AuthToken:          cfg.ServiceAuthToken,
		S3Private:          s3Private,
		S3Upload:           s3Uploaded,
		VaultCli:           vault,
		VaultPath:          cfg.VaultPath,
		ImageCli:           imageAPI,
		DownloadServiceURL: cfg.DownloadServiceURL,
	}, cfg.KafkaConsumerWorkers)

	// Get HealthCheck
	hc, err := serviceList.GetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		log.Fatal(ctx, "could not instantiate healthcheck", err)
		return nil, err
	}

	if err := registerCheckers(ctx, hc, vault, s3Private, s3Uploaded, imageAPI, consumer); err != nil {
		return nil, errors.Wrap(err, "unable to register checkers")
	}

	r.StrictSlash(true).Path("/health").HandlerFunc(hc.Handler)
	hc.Start(ctx)

	// Run the http server in a new go-routine
	go func() {
		if err := s.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()

	return &Service{
		config:      cfg,
		server:      s,
		router:      r,
		serviceList: serviceList,
		healthCheck: hc,
		vault:       vault,
		consumer:    consumer,
	}, nil
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.config.GracefulShutdownTimeout
	log.Info(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout})
	ctx, cancel := context.WithTimeout(ctx, timeout)

	// track shutown gracefully closes up
	var gracefulShutdown bool

	go func() {
		defer cancel()
		var hasShutdownError bool

		// stop healthcheck, as it depends on everything else
		if svc.serviceList.HealthCheck {
			svc.healthCheck.Stop()
		}

		// If kafka consumer exists, stop listening to it. (Will close later)
		if svc.serviceList.KafkaConsumer {
			if err := svc.consumer.StopListeningToConsumer(ctx); err != nil {
				log.Error(ctx, "error stopping kafka consumer listener", err)
				hasShutdownError = true
			}
			log.Info(ctx, "stopped kafka consumer listener")
		}

		// stop any incoming requests before closing any outbound connections
		if err := svc.server.Shutdown(ctx); err != nil {
			log.Error(ctx, "failed to shutdown http server", err)
			hasShutdownError = true
		}

		// If kafka consumer exists, close it.
		if svc.serviceList.KafkaConsumer {
			if err := svc.consumer.Close(ctx); err != nil {
				log.Error(ctx, "error closing kafka consumer", err)
				hasShutdownError = true
			}
			log.Info(ctx, "closed kafka consumer")
		}

		if !hasShutdownError {
			gracefulShutdown = true
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	if !gracefulShutdown {
		err := errors.New("failed to shutdown gracefully")
		log.Error(ctx, "failed to shutdown gracefully ", err)
		return err
	}

	log.Info(ctx, "graceful shutdown was successful")
	return nil
}

func registerCheckers(ctx context.Context,
	hc HealthChecker,
	vault event.VaultClient,
	s3Private event.S3Writer,
	s3Uploaded event.S3Reader,
	imageAPI event.ImageAPIClient,
	consumer kafka.IConsumerGroup) (err error) {

	hasErrors := false

	if vault != nil {
		if err = hc.AddCheck("Vault client", vault.Checker); err != nil {
			hasErrors = true
			log.Error(ctx, "error adding check for vault", err)
		}
	}

	if err := hc.AddCheck("S3 private bucket", s3Private.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for s3Private private bucket", err)
	}

	if err := hc.AddCheck("S3 uploaded bucket", s3Uploaded.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for s3Private uploaded bucket", err)
	}

	if err := hc.AddCheck("Image API client", imageAPI.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for Image API", err)
	}

	if err := hc.AddCheck("Kafka consumer", consumer.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for Image API", err)
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}
