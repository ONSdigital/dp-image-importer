BINPATH ?= build

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)

LDFLAGS = -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"

VAULT_ADDR=${VAULT_ADDR:-'http://127.0.0.1:8200'}

# The following variables are used to generate a vault token for the app. The reason for declaring variables, is that
# its difficult to move the token code in a Makefile action. Doing so makes the Makefile more difficult to
# read and starts introduction if/else statements.
VAULT_POLICY=${VAULT_POLICY:-"$(shell vault policy write -address=$(VAULT_ADDR) read-write-psk policy.hcl)"}
TOKEN_INFO=${TOKEN_INFO":-$(shell vault token create -address=$(VAULT_ADDR) -policy=read-write-psk -period=24h -display-name=dp-image-importer)"}
APP_TOKEN=${APP_TOKEN:-"$(shell echo $(TOKEN_INFO) | awk '{print $$6}')"}

.PHONY: all
all: audit test build

.PHONY: audit
audit:
	go list -m all | nancy sleuth

.PHONY: build
build:
	go build -tags 'production' $(LDFLAGS) -o $(BINPATH)/dp-image-importer

.PHONY: debug
debug:
	go build -tags 'debug' $(LDFLAGS) -o $(BINPATH)/dp-image-importer
	VAULT_TOKEN=$(APP_TOKEN) VAULT_ADDR=$(VAULT_ADDR) HUMAN_LOG=1 DEBUG=1 $(BINPATH)/dp-image-importer

debug-run:
	HUMAN_LOG=1 go run -tags 'debug' $(LDFLAGS) main.go

.PHONY: test
test:
	go test -race -cover ./...

.PHONY: produce
produce:
	HUMAN_LOG=1 go run cmd/producer/main.go

.PHONY: convey
convey:
	goconvey ./...

