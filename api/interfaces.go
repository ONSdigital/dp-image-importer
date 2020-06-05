package api

import (
	"context"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out mock/vault.go -pkg mock . VaultClienter
//go:generate moq -out mock/s3.go -pkg mock . S3Clienter

type VaultClienter interface {
	Read(path string) (map[string]interface{}, error)
	Write(path string, data map[string]interface{}) error
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}

type S3Clienter interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}
