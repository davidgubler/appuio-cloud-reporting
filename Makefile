# Set Shell to bash, otherwise some targets fail with dash/zsh etc.
SHELL := /bin/bash

# Disable built-in rules
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --no-builtin-variables
.SUFFIXES:
.SECONDARY:
.DEFAULT_GOAL := help

# General variables
include Makefile.vars.mk

# Following includes do not print warnings or error if files aren't found
# Optional Documentation module.
-include docs/antora-preview.mk docs/antora-build.mk
# Optional kind module
-include kind/kind.mk

.PHONY: help
help: ## Show this help
	@grep -E -h '\s##\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: build-bin build-docker ## All-in-one build

.PHONY: build-bin
build-bin: export CGO_ENABLED = 0
build-bin: fmt vet ## Build binary
	@go build -o $(BIN_FILENAME) github.com/appuio/appuio-cloud-reporting

.PHONY: build-docker
build-docker: build-bin ## Build docker image
	$(DOCKER_CMD) build -t $(CONTAINER_IMG) .

.PHONY: ensure-prometheus
ensure-prometheus:
	go run ./util/ensure_prometheus

.PHONY: test
test: ensure-prometheus
	docker rm -f test-migrations ||:
	docker run -d --name test-migrations -e POSTGRES_DB=test-migrations -e POSTGRES_USER=test-migrations -e POSTGRES_PASSWORD=test-migrations -p65432:5432 postgres:13-bullseye
	docker exec -t test-migrations sh -c 'until pg_isready; do sleep 1; done; sleep 1'
	go run ./cmd/migrate '-db-url=postgres://test-migrations:test-migrations@localhost:65432/test-migrations?sslmode=disable'
	go run ./cmd/migrate -seed '-db-url=postgres://test-migrations:test-migrations@localhost:65432/test-migrations?sslmode=disable'
	go test ./... -tags integration -args '-db-url=postgres://test-migrations:test-migrations@localhost:65432/test-migrations?sslmode=disable'
	go run ./cmd/testreport '-db-url=postgres://test-migrations:test-migrations@localhost:65432/test-migrations?sslmode=disable'
	docker rm -f test-migrations

.PHONY: fmt
fmt: ## Run 'go fmt' against code
	go fmt ./...

.PHONY: vet
vet: ## Run 'go vet' against code
	go vet ./...

.PHONY: lint
lint: fmt vet generate ## All-in-one linting
	@echo 'Check for uncommitted changes ...'
	git diff --exit-code

.PHONY: generate
generate: ## Generate additional code and artifacts
	@go generate ./...

.PHONY: clean
clean: ## Cleans local build artifacts
	rm -rf docs/node_modules $(docs_out_dir) dist .cache
