package api

import (
	"context"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out mock/vault.go -pkg mock . VaultClienter
//go:generate moq -out mock/s3.go -pkg mock . S3Clienter
//go:generate moq -out mock/image.go -pkg mock . ImageAPIClienter
//go:generate moq -out mock/kafka.go -pkg mock . KafkaConsumer

type VaultClienter interface {
	Read(path string) (map[string]interface{}, error)
	Write(path string, data map[string]interface{}) error
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}

type S3Clienter interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}

type ImageAPIClienter interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}

type KafkaConsumer interface {
	StopListeningToConsumer(ctx context.Context) (err error)
	Close(ctx context.Context) (err error)
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}
