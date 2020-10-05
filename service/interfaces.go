package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-image-importer/config"
	kafka "github.com/ONSdigital/dp-kafka"
	"github.com/aws/aws-sdk-go/aws/session"
)

//go:generate moq -out mock/initialiser.go -pkg mock . Initialiser
//go:generate moq -out mock/server.go -pkg mock . HTTPServer
//go:generate moq -out mock/healthCheck.go -pkg mock . HealthChecker

//go:generate moq -out mock/vault.go -pkg mock . VaultClienter
//go:generate moq -out mock/s3.go -pkg mock . S3Clienter
//go:generate moq -out mock/image.go -pkg mock . ImageAPIClienter
//go:generate moq -out mock/kafka.go -pkg mock . KafkaConsumer

// Initialiser defines the methods to initialise external services
type Initialiser interface {
	DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer
	DoGetVault(ctx context.Context, cfg *config.Config) (VaultClienter, error)
	DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error)
	DoGetS3Client(awsRegion, bucketName string, encryptionEnabled bool) (S3Clienter, error)
	DoGetS3ClientWithSession(bucketName string, encryptionEnabled bool, s *session.Session) S3Clienter
	DoGetImageAPI(ctx context.Context, cfg *config.Config) ImageAPIClienter
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

type VaultClienter interface {
	Read(path string) (map[string]interface{}, error)
	Write(path string, data map[string]interface{}) error
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}

// S3Clienter defines the required methods from S3 client
type S3Clienter interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	Session() *session.Session
}

type ImageAPIClienter interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}

type KafkaConsumer interface {
	StopListeningToConsumer(ctx context.Context) (err error)
	Close(ctx context.Context) (err error)
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	Channels() *kafka.ConsumerGroupChannels
	Release()
}
