DOCKER_ARCHS ?= amd64 armv7 arm64
DOCKER_IMAGE_NAME ?= databricks-exporter

ALL_SRC := $(shell find . -name '*.go' -o -name 'Dockerfile*' -type f | sort)

all:: vet common-all

.PHONY: test-integration
test-integration: ## Run integration tests (requires Databricks credentials in env)
	@echo "Running integration tests against live Databricks instance..."
	@go test -tags=integration -v -timeout 10m ./collector/...

.PHONY: test-integration-fast
test-integration-fast: ## Run integration tests with shorter timeout
	@echo "Running integration tests (fast)..."
	@go test -tags=integration -v -timeout 3m ./collector/... -run TestIntegration_RealDatabricks

.PHONY: bench-integration
bench-integration: ## Run integration benchmarks
	@echo "Running integration benchmarks..."
	@go test -tags=integration -v -bench=. -benchmem -timeout 10m ./collector/...

include Makefile.common

