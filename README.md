# dp-image-importer

ONS service that imports uploaded images and adds them to a private bucket

## Getting started

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
| KAFKA_ADDR                   | `localhost:9092`                  | The address of (TLS-ready) Kafka brokers (comma-separated values)
| KAFKA_VERSION                | `1.0.2`                           | The version of (TLS-ready) Kafka
| KAFKA_SEC_PROTO              | _unset_            (only `TLS`)   | if set to `TLS`, kafka connections will use TLS
| KAFKA_SEC_CLIENT_KEY         | _unset_                           | PEM [2] for the client key (optional, used for client auth) [1]
| KAFKA_SEC_CLIENT_CERT        | _unset_                           | PEM [2] for the client certificate (optional, used for client auth) [1]
| KAFKA_SEC_CA_CERTS           | _unset_                           | PEM [2] of CA cert chain if using private CA for the server cert [1]
| KAFKA_SEC_SKIP_VERIFY        | false                             | ignore server certificate issues if set to `true` [1]
| KAFKA_CONSUMER_WORKERS       | 1                                 | The maximum number of parallel kafka consumers
| IMAGE_UPLOADED_GROUP         | dp-image-importer                 | The consumer group this application to consume ImageUploaded messages
| IMAGE_UPLOADED_TOPIC         | image-uploaded                    | The name of the topic to consume messages from
| S3_PRIVATE_BUCKET_NAME       | csv-exported                      | Name of the S3 bucket used to store generated images
| S3_UPLOADED_BUCKET_NAME      | dp-frontend-florence-file-uploads | Name of the S3 bucket used to read original images from
| VAULT_TOKEN                  | -                                 | Vault token required for the client to talk to vault. (Use `make debug` to create a vault token)
| VAULT_ADDR                   | http://localhost:8200             | The vault address
| VAULT_PATH                   | secret/shared/psk                 | The path where the psks will be stored in vault
| DOWNLOAD_SERVICE_URL         | http://localhost:23600            | The public address of the download service

Notes:

1. Ignored unless using TLS (i.e. `KAFKA_SEC_PROTO` has a value enabling TLS)

2. PEM values are identified as those starting with `-----BEGIN`
    and can use `\n` (sic) instead of newlines (they will be converted to newlines before use).
    Any other value will be treated as a path to the given PEM file.

### Healthcheck

 The `/health` endpoint returns the current status of the service. Dependent services are health checked on an interval defined by the `HEALTHCHECK_INTERVAL` environment variable.

 On a development machine a request to the health check endpoint can be made by:

 `curl localhost:24800/health`

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2021, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.

