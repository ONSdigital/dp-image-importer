package config

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	Convey("Given an environment with no environment variables set", t, func() {
		os.Clearenv()
		cfg, err := Get()

		Convey("When the config values are retrieved", func() {

			Convey("Then there should be no error returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the values should be set to the expected defaults", func() {
				So(cfg.BindAddr, ShouldEqual, "localhost:24800")
				So(cfg.ServiceAuthToken, ShouldEqual, "4424A9F2-B903-40F4-85F1-240107D1AFAF")
				So(cfg.EncryptionDisabled, ShouldBeFalse)
				So(cfg.AwsRegion, ShouldEqual, "eu-west-1")
				So(cfg.GracefulShutdownTimeout, ShouldEqual, 5*time.Second)
				So(cfg.HealthCheckInterval, ShouldEqual, 30*time.Second)
				So(cfg.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
				So(cfg.ImageAPIURL, ShouldEqual, "http://localhost:24700")
				So(cfg.Brokers, ShouldHaveLength, 1)
				So(cfg.Brokers[0], ShouldEqual, "localhost:9092")
				So(cfg.KafkaVersion, ShouldEqual, "1.0.2")
				So(cfg.KafkaConsumerWorkers, ShouldEqual, 1)
				So(cfg.ImageUploadedGroup, ShouldEqual, "dp-image-importer")
				So(cfg.ImageUploadedTopic, ShouldEqual, "image-uploaded")
				So(cfg.S3PrivateBucketName, ShouldEqual, "csv-exported")
				So(cfg.S3UploadedBucketName, ShouldEqual, "dp-frontend-florence-file-uploads")
				So(cfg.VaultAddress, ShouldEqual, "http://localhost:8200")
				So(cfg.VaultPath, ShouldEqual, "secret/shared/psk")
				So(cfg.VaultToken, ShouldEqual, "")
				So(cfg.DownloadServiceURL, ShouldEqual, "http://localhost:23600")
			})

			Convey("Then a second call to config should return the same config", func() {
				newCfg, newErr := Get()
				So(newErr, ShouldBeNil)
				So(newCfg, ShouldResemble, cfg)
			})
		})
	})
}
