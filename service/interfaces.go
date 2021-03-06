package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-image-importer/config"
	"github.com/ONSdigital/dp-image-importer/event"
	kafka "github.com/ONSdigital/dp-kafka"
	"github.com/aws/aws-sdk-go/aws/session"
)

//go:generate moq -out mock/initialiser.go -pkg mock . Initialiser
//go:generate moq -out mock/server.go -pkg mock . HTTPServer
//go:generate moq -out mock/healthCheck.go -pkg mock . HealthChecker

//go:generate moq -out mock/kafka.go -pkg mock . KafkaConsumer

// Initialiser defines the methods to initialise external services
type Initialiser interface {
	DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer
	DoGetVault(ctx context.Context, cfg *config.Config) (event.VaultClient, error)
	DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error)
	DoGetS3Client(awsRegion, bucketName string, encryptionEnabled bool) (event.S3Writer, error)
	DoGetS3ClientWithSession(bucketName string, encryptionEnabled bool, s *session.Session) event.S3Reader
	DoGetImageAPI(ctx context.Context, cfg *config.Config) event.ImageAPIClient
	DoGetKafkaConsumer(ctx context.Context, cfg *config.Config) (KafkaConsumer, error)
}

// HTTPServer defines the required methods from the HTTP server
type HTTPServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

// HealthChecker defines the required methods from Healthcheck
type HealthChecker interface {
	Handler(w http.ResponseWriter, req *http.Request)
	Start(ctx context.Context)
	Stop()
	AddCheck(name string, checker healthcheck.Checker) (err error)
}

type KafkaConsumer interface {
	StopListeningToConsumer(ctx context.Context) (err error)
	Close(ctx context.Context) (err error)
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	Channels() *kafka.ConsumerGroupChannels
	Release()
}

// EventConsumer defines the required methods from event Consumer
type EventConsumer interface {
	Consume(ctx context.Context, messageConsumer event.MessageConsumer, handler event.Handler)
	Close(ctx context.Context) (err error)
}
