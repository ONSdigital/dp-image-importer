package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/image"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-image-importer/config"
	"github.com/ONSdigital/dp-image-importer/event"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	dps3 "github.com/ONSdigital/dp-s3/v3"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ExternalServiceList holds the initialiser and initialisation state of external services.
type ExternalServiceList struct {
	S3Private     bool
	S3Uploaded    bool
	ImageAPI      bool
	HealthCheck   bool
	KafkaConsumer bool
	Init          Initialiser
}

// NewServiceList creates a new service list with the provided initialiser
func NewServiceList(initialiser Initialiser) *ExternalServiceList {
	return &ExternalServiceList{
		S3Private:     false,
		S3Uploaded:    false,
		ImageAPI:      false,
		HealthCheck:   false,
		KafkaConsumer: false,
		Init:          initialiser,
	}
}

// Init implements the Initialiser interface to initialise dependencies
type Init struct{}

// GetHTTPServer creates an http server and sets the Server flag to true
func (e *ExternalServiceList) GetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := e.Init.DoGetHTTPServer(bindAddr, router)
	return s
}

// GetS3Clients returns S3 clients uploaded and private. They share the same AWS session.
func (e *ExternalServiceList) GetS3Clients(cfg *config.Config) (s3Uploaded event.S3Reader, s3Private event.S3Writer, err error) {
	ctx := context.Background()
	s3Private, err = e.Init.DoGetS3Client(ctx, cfg.AwsRegion, cfg.S3PrivateBucketName)
	if err != nil {
		return nil, nil, err
	}
	e.S3Private = true
	s3Uploaded, err = e.Init.DoGetS3Client(ctx, cfg.AwsRegion, cfg.S3UploadedBucketName)
	if err != nil {
		return nil, nil, err
	}
	e.S3Uploaded = true
	return
}

// GetImageAPI creates an ImageAPI client and sets the ImageAPI flag to true
func (e *ExternalServiceList) GetImageAPI(ctx context.Context, cfg *config.Config) event.ImageAPIClient {
	imageAPI := e.Init.DoGetImageAPI(ctx, cfg)
	e.ImageAPI = true
	return imageAPI
}

// GetKafkaConsumer creates a Kafka consumer and sets the consumer flag to true
func (e *ExternalServiceList) GetKafkaConsumer(ctx context.Context, cfg *config.Config) (kafka.IConsumerGroup, error) {
	consumer, err := e.Init.DoGetKafkaConsumer(ctx, cfg)
	if err != nil {
		return nil, err
	}
	e.KafkaConsumer = true
	return consumer, nil
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

// DoGetS3Client creates a new S3Client for the provided AWS region and bucket name.
func (e *Init) DoGetS3Client(ctx context.Context, awsRegion, bucketName string) (event.S3Writer, error) {
	cfg, _ := config.Get()

	var s3Client *dps3.Client

	if cfg.LocalS3URL != "" {
		config, err := awsConfig.LoadDefaultConfig(
			ctx, awsConfig.WithRegion(awsRegion),
			awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.LocalS3ID, cfg.LocalS3Secret, "")),
		)
		if err != nil {
			return nil, err
		}

		s3Client = dps3.NewClientWithConfig(bucketName, config, func(options *s3.Options) {
			options.BaseEndpoint = aws.String(cfg.LocalS3URL)
			options.UsePathStyle = true
		})

	} else {
		config, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithRegion(awsRegion))
		if err != nil {
			return nil, err
		}

		s3Client = dps3.NewClientWithConfig(bucketName, config)
	}

	return s3Client, nil
}

// DoGetS3ClientWithConfig creates a new S3Clienter (extension of S3Client with Upload operations)
// for the provided bucket name, using an existing AWS session
func (e *Init) DoGetS3ClientWithConfig(bucketName string, cfg aws.Config) event.S3Reader {
	return dps3.NewClientWithConfig(bucketName, cfg)
}

// DoGetImageAPI returns an Image API client
func (e *Init) DoGetImageAPI(ctx context.Context, cfg *config.Config) event.ImageAPIClient {
	return image.NewAPIClient(cfg.ImageAPIURL)
}

// DoGetKafkaConsumer returns a Kafka Consumer group
func (e *Init) DoGetKafkaConsumer(ctx context.Context, cfg *config.Config) (kafka.IConsumerGroup, error) {
	kafkaOffset := kafka.OffsetOldest

	cConfig := &kafka.ConsumerGroupConfig{
		Offset:       &kafkaOffset,
		KafkaVersion: &cfg.KafkaVersion,
	}
	if cfg.KafkaSecProtocol == "TLS" {
		cConfig.SecurityConfig = kafka.GetSecurityConfig(
			cfg.KafkaSecCACerts,
			cfg.KafkaSecClientCert,
			cfg.KafkaSecClientKey,
			cfg.KafkaSecSkipVerify,
		)
	}

	cgChannels := kafka.CreateConsumerGroupChannels(cfg.KafkaConsumerWorkers)

	return kafka.NewConsumerGroup(
		ctx,
		cfg.Brokers,
		cfg.ImageUploadedTopic,
		cfg.ImageUploadedGroup,
		cgChannels,
		cConfig,
	)
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
