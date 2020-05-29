package service

import (
	"context"
	"github.com/ONSdigital/dp-image-importer/api"
	"github.com/ONSdigital/dp-image-importer/config"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Service contains all the configs, server and clients to run the Image API
type Service struct {
	config      *config.Config
	server      HTTPServer
	router      *mux.Router
	api         *api.API
	serviceList *ExternalServiceList
	healthCheck HealthChecker
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

	// Setup the API
	a := api.Setup(ctx, r)

	// Get HealthCheck
	hc, err := serviceList.GetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		log.Event(ctx, "could not instantiate healthcheck", log.FATAL, log.Error(err))
		return nil, err
	}
	if err := registerCheckers(ctx, hc); err != nil {
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
		api:         a,
		serviceList: serviceList,
		healthCheck: hc,
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

		// stop any incoming requests before closing any outbound connections
		if err := svc.server.Shutdown(ctx); err != nil {
			log.Event(ctx, "failed to shutdown http server", log.Error(err), log.ERROR)
			hasShutdownError = true
		}

		// close API
		if err := svc.api.Close(ctx); err != nil {
			log.Event(ctx, "error closing API", log.Error(err), log.ERROR)
			hasShutdownError = true
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

func registerCheckers(ctx context.Context, hc HealthChecker) (err error) {
	// TODO ADD HEALTH CHECKS HERE
	return
}