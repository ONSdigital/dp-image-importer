dp-image-importer
================

ONS service that imports uploaded images and adds them to a private bucket

### Getting started

* Run `make debug`

### Dependencies

* No further dependencies other than those defined in `go.mod`

### Configuration

| Environment variable         | Default                           | Description
| ---------------------------- | --------------------------------- | -----------
| BIND_ADDR                    | :24800                            | The host and port to bind to
| AWS_REGION                   | eu-west-1                         | The AWS region
| GRACEFUL_SHUTDOWN_TIMEOUT    | 5s                                | The graceful shutdown timeout in seconds (`time.Duration` format)
| HEALTHCHECK_INTERVAL         | 30s                               | Time between self-healthchecks (`time.Duration` format)
| HEALTHCHECK_CRITICAL_TIMEOUT | 90s                               | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format)
| IMAGE_API_URL                | http://localhost:24700            | The image api url
| S3_PRIVATE_BUCKET_NAME       | csv-exported                      | Name of the S3 bucket used to store generated images
| S3_UPLOADED_BUCKET_NAME      | dp-frontend-florence-file-uploads | Name of the S3 bucket used to read original images from
| VAULT_TOKEN                  | -                                 | Vault token required for the client to talk to vault. (Use `make debug` to create a vault token)
| VAULT_ADDR                   | http://localhost:8200             | The vault address
| VAULT_PATH                   | secret/shared/psk                 | The path where the psks will be stored in vault

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2020, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.

