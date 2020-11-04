package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dp-image-importer
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	ServiceAuthToken           string        `envconfig:"SERVICE_AUTH_TOKEN"             json:"-"`
	EncryptionDisabled         bool          `envconfig:"ENCRYPTION_DISABLED"`
	AwsRegion                  string        `envconfig:"AWS_REGION"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	ImageAPIURL                string        `envconfig:"IMAGE_API_URL"`
	Brokers                    []string      `envconfig:"KAFKA_ADDR"                     json:"-"`
	ImageUploadedGroup         string        `envconfig:"IMAGE_UPLOADED_GROUP"`
	ImageUploadedTopic         string        `envconfig:"IMAGE_UPLOADED_TOPIC"`
	S3PrivateBucketName        string        `envconfig:"S3_PRIVATE_BUCKET_NAME"`
	S3UploadedBucketName       string        `envconfig:"S3_UPLOADED_BUCKET_NAME"`
	VaultToken                 string        `envconfig:"VAULT_TOKEN"                    json:"-"`
	VaultAddress               string        `envconfig:"VAULT_ADDR"`
	VaultPath                  string        `envconfig:"VAULT_PATH"`
	DownloadServiceURL         string        `envconfig:"DOWNLOAD_SERVICE_URL"`
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg := &Config{
		BindAddr:                   "localhost:24800",
		ServiceAuthToken:           "4424A9F2-B903-40F4-85F1-240107D1AFAF",
		EncryptionDisabled:         false,
		AwsRegion:                  "eu-west-1",
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		ImageAPIURL:                "http://localhost:24700",
		Brokers:                    []string{"localhost:9092"},
		ImageUploadedGroup:         "dp-image-importer",
		ImageUploadedTopic:         "image-uploaded",
		S3PrivateBucketName:        "csv-exported",
		S3UploadedBucketName:       "dp-frontend-florence-file-uploads",
		VaultPath:                  "secret/shared/psk",
		VaultAddress:               "http://localhost:8200",
		VaultToken:                 "",
		DownloadServiceURL:         "http://localhost:23600",
	}

	return cfg, envconfig.Process("", cfg)
}
