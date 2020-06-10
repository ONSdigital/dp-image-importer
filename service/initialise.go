package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-image-importer/api"
	"github.com/ONSdigital/dp-image-importer/config"

	"github.com/ONSdigital/dp-api-clients-go/image"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	dps3 "github.com/ONSdigital/dp-s3"
	dpvault "github.com/ONSdigital/dp-vault"
)

// ExternalServiceList holds the initialiser and initialisation state of external services.
type ExternalServiceList struct {
	Vault       bool
	S3Private   bool
	S3Uploaded  bool
	ImageAPI    bool
	HealthCheck bool
	Init        Initialiser
}

// NewServiceList creates a new service list with the provided initialiser
func NewServiceList(initialiser Initialiser) *ExternalServiceList {
	return &ExternalServiceList{
		Vault:       false,
		S3Private:   false,
		S3Uploaded:  false,
		ImageAPI:    false,
		HealthCheck: false,
		Init:        initialiser,
	}
}

// Init implements the Initialiser interface to initialise dependencies
type Init struct{}

// GetHTTPServer creates an http server and sets the Server flag to true
func (e *ExternalServiceList) GetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := e.Init.DoGetHTTPServer(bindAddr, router)
	return s
}

// GetVault creates a Vault client and sets the Vault flag to true
func (e *ExternalServiceList) GetVault(ctx context.Context, cfg *config.Config) (api.VaultClienter, error) {
	vault, err := e.Init.DoGetVault(ctx, cfg)
	if err != nil {
		return nil, err
	}
	e.Vault = true
	return vault, nil
}

// GetS3Private creates a S3 client and sets the S3Private flag to true
func (e *ExternalServiceList) GetS3Private(ctx context.Context, cfg *config.Config) (api.S3Clienter, error) {
	s3, err := e.Init.DoGetS3Private(ctx, cfg)
	if err != nil {
		return nil, err
	}
	e.S3Private = true
	return s3, nil
}

// GetS3Uploaded creates a S3 client and sets the S3Uploaded flag to true
func (e *ExternalServiceList) GetS3Uploaded(ctx context.Context, cfg *config.Config) (api.S3Clienter, error) {
	s3, err := e.Init.DoGetS3Uploaded(ctx, cfg)
	if err != nil {
		return nil, err
	}
	e.S3Uploaded = true
	return s3, nil
}

// GetImageAPI creates an ImageAPI client and sets the ImageAPI flag to true
func (e *ExternalServiceList) GetImageAPI(ctx context.Context, cfg *config.Config) api.ImageAPIClienter {
	imageAPI := e.Init.DoGetImageAPI(ctx, cfg)
	e.ImageAPI = true
	return imageAPI
}

// GetHealthCheck creates a healthcheck with versionInfo and sets teh HealthCheck flag to true
func (e *ExternalServiceList) GetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	hc, err := e.Init.DoGetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	e.HealthCheck = true
	return hc, nil
}

// DoGetHTTPServer creates an HTTP Server with the provided bind address and router
func (e *Init) DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := dphttp.NewServer(bindAddr, router)
	s.HandleOSSignals = false
	return s
}

// DoGetVault returns a VaultClient
func (e *Init) DoGetVault(ctx context.Context, cfg *config.Config) (api.VaultClienter, error) {
	vault, err := dpvault.CreateClient(cfg.VaultToken, cfg.VaultAddress, 3)
	if err != nil {
		return nil, err
	}
	return vault, nil
}

// DoGetS3Private returns a S3Client
func (e *Init) DoGetS3Private(ctx context.Context, cfg *config.Config) (api.S3Clienter, error) {
	vault, err := dps3.NewClient(cfg.AwsRegion, cfg.S3PrivateBucketName, true)
	if err != nil {
		return nil, err
	}
	return vault, nil
}

// DoGetS3Uploaded returns a S3Client
func (e *Init) DoGetS3Uploaded(ctx context.Context, cfg *config.Config) (api.S3Clienter, error) {
	vault, err := dps3.NewClient(cfg.AwsRegion, cfg.S3UploadedBucketName, true)
	if err != nil {
		return nil, err
	}
	return vault, nil
}

// DoGetImageAPI returns an Image API client
func (e *Init) DoGetImageAPI(ctx context.Context, cfg *config.Config) api.ImageAPIClienter {
	return image.NewAPIClient(cfg.ImageAPIURL)
}

// DoGetHealthCheck creates a healthcheck with versionInfo
func (e *Init) DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return &hc, nil
}
