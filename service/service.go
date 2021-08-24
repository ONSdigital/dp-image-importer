package service

import (
	"context"

	"github.com/ONSdigital/dp-image-importer/config"
	"github.com/ONSdigital/dp-image-importer/event"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/log.go/log"
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
	log.Event(ctx, "running service", log.INFO)

	// Read config
	cfg, err := config.Get()
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve service configuration")
	}
	log.Event(ctx, "got service configuration", log.Data{"config": cfg}, log.INFO)

	// Get HTTP Server with collectionID checkHeader middleware
	r := mux.NewRouter()
	s := serviceList.GetHTTPServer(cfg.BindAddr, r)

	// Get Vault client
	vault, err := serviceList.GetVault(ctx, cfg)
	if err != nil {
		log.Event(ctx, "failed to initialise Vault client", log.FATAL, log.Error(err))
		return nil, err
	}

	// Get S3 Clients
	s3Uploaded, s3Private, err := serviceList.GetS3Clients(cfg)
	if err != nil {
		log.Event(ctx, "could not instantiate S3 clients", log.FATAL, log.Error(err))
		return nil, err
	}

	// Get Image API Client
	imageAPI := serviceList.GetImageAPI(ctx, cfg)

	// Get Kafka consumer
	consumer, err := serviceList.GetKafkaConsumer(ctx, cfg)
	if err != nil {
		log.Event(ctx, "failed to initialise kafka consumer", log.FATAL, log.Error(err))
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
		log.Event(ctx, "could not instantiate healthcheck", log.FATAL, log.Error(err))
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
	log.Event(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout}, log.INFO)
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
				log.Event(ctx, "error stopping kafka consumer listener", log.ERROR, log.Error(err))
				hasShutdownError = true
			}
			log.Event(ctx, "stopped kafka consumer listener", log.INFO)
		}

		// stop any incoming requests before closing any outbound connections
		if err := svc.server.Shutdown(ctx); err != nil {
			log.Event(ctx, "failed to shutdown http server", log.Error(err), log.ERROR)
			hasShutdownError = true
		}

		// If kafka consumer exists, close it.
		if svc.serviceList.KafkaConsumer {
			if err := svc.consumer.Close(ctx); err != nil {
				log.Event(ctx, "error closing kafka consumer", log.Error(err), log.ERROR)
				hasShutdownError = true
			}
			log.Event(ctx, "closed kafka consumer", log.INFO)
		}

		if !hasShutdownError {
			gracefulShutdown = true
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	if !gracefulShutdown {
		err := errors.New("failed to shutdown gracefully")
		log.Event(ctx, "failed to shutdown gracefully ", log.ERROR, log.Error(err))
		return err
	}

	log.Event(ctx, "graceful shutdown was successful", log.INFO)
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
			log.Event(ctx, "error adding check for vault", log.ERROR, log.Error(err))
		}
	}

	if err := hc.AddCheck("S3 private bucket", s3Private.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "error adding check for s3Private private bucket", log.ERROR, log.Error(err))
	}

	if err := hc.AddCheck("S3 uploaded bucket", s3Uploaded.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "error adding check for s3Private uploaded bucket", log.ERROR, log.Error(err))
	}

	if err := hc.AddCheck("Image API client", imageAPI.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "error adding check for Image API", log.ERROR, log.Error(err))
	}

	if err := hc.AddCheck("Kafka consumer", consumer.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "error adding check for Image API", log.ERROR, log.Error(err))
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}
