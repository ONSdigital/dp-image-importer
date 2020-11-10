dp-image-importer
================

ONS service that imports uploaded images and adds them to a private bucket

### Getting started

* Run `make debug`

The service runs in the background consuming messages from Kafka. The messages are produced by the 
[image API](https://github.com/ONSdigital/dp-image-api) and an example image can be created using the helper script,
`make produce`.

### Dependencies

* Requires running…
  * [kafka & kafka](https://github.com/ONSdigital/dp/blob/master/guides/INSTALLING.md#prerequisites)
  * [zebedee](https://github.com/ONSdigital/zebedee)
  * [dp-image-api](https://github.com/ONSdigital/dp-image-api)
  * [AWS S3 access](https://github.com/ONSdigital/dp/blob/master/guides/AWS_CREDENTIALS.md)
* No further dependencies other than those defined in `go.mod`

### Configuration

| Environment variable         | Default                           | Description
| ---------------------------- | --------------------------------- | -----------
| BIND_ADDR                    | :24800                            | The host and port to bind to
| SERVICE_AUTH_TOKEN           | -                                 | The service token for this app
| ENCRYPTION_DISABLED          | false                             | Determines whether vault is used and whether files are encrypted on S3
| AWS_REGION                   | eu-west-1                         | The AWS region
| GRACEFUL_SHUTDOWN_TIMEOUT    | 5s                                | The graceful shutdown timeout in seconds (`time.Duration` format)
| HEALTHCHECK_INTERVAL         | 30s                               | Time between self-healthchecks (`time.Duration` format)
| HEALTHCHECK_CRITICAL_TIMEOUT | 90s                               | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format)
| IMAGE_API_URL                | http://localhost:24700            | The image api url
| KAFKA_ADDR                   | "localhost:9092"                  | The address of Kafka (accepts list)
| IMAGE_UPLOADED_GROUP         | dp-image-importer                 | The consumer group this application to consume ImageUploaded messages
| IMAGE_UPLOADED_TOPIC         | image-uploaded                    | The name of the topic to consume messages from
| S3_PRIVATE_BUCKET_NAME       | csv-exported                      | Name of the S3 bucket used to store generated images
| S3_UPLOADED_BUCKET_NAME      | dp-frontend-florence-file-uploads | Name of the S3 bucket used to read original images from
| VAULT_TOKEN                  | -                                 | Vault token required for the client to talk to vault. (Use `make debug` to create a vault token)
| VAULT_ADDR                   | http://localhost:8200             | The vault address
| VAULT_PATH                   | secret/shared/psk                 | The path where the psks will be stored in vault
| DOWNLOAD_SERVICE_URL         | http://localhost:23600            | The public address of the download service

### Healthcheck

 The `/health` endpoint returns the current status of the service. Dependent services are health checked on an interval defined by the `HEALTHCHECK_INTERVAL` environment variable.

 On a development machine a request to the health check endpoint can be made by:

 `curl localhost:24800/health`

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2020, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.

