package api

import (
	"context"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out mock/vault.go -pkg mock . VaultClienter

type VaultClienter interface {
	Read(path string) (map[string]interface{}, error)
	Write(path string, data map[string]interface{}) error
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}
