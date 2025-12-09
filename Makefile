# Databricks Prometheus Exporter Makefile

# Docker settings
DOCKER_ARCHS       ?= amd64 armv7 arm64
DOCKER_IMAGE_NAME  ?= databricks-exporter

# Build settings
BIN_DIR            := bin
BINARY_NAME        := databricks-exporter
GO                 := go
GOFLAGS            :=
pkgs               := ./...

# Include common Prometheus Makefile targets
include Makefile.common

.PHONY: all
all: check build

# ============================================================================
# Quality checks
# ============================================================================

.PHONY: check
check: vet lint fmt test ## Run all quality checks (vet, lint, fmt, test)

.PHONY: vet
vet: ## Run go vet
	@echo ">> running go vet"
	$(GO) vet $(GOFLAGS) $(pkgs)

.PHONY: lint
lint: ## Run golangci-lint (includes deadcode analysis)
	@echo ">> running golangci-lint"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run $(pkgs); \
	else \
		echo "golangci-lint not installed, skipping (install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)"; \
	fi

.PHONY: fmt
fmt: ## Run gofmt and check for formatting issues
	@echo ">> checking code formatting"
	@fmtRes=$$(gofmt -l $$(find . -name '*.go' -not -path './vendor/*' -not -path './mixin/*')); \
	if [ -n "$${fmtRes}" ]; then \
		echo "gofmt found issues in:"; echo "$${fmtRes}"; \
		echo "Run 'make fmt-fix' to fix"; \
		exit 1; \
	fi
	@echo ">> formatting OK"

.PHONY: fmt-fix
fmt-fix: ## Auto-fix formatting issues
	@echo ">> fixing code formatting"
	$(GO) fmt $(pkgs)

.PHONY: test
test: ## Run all tests
	@echo ">> running tests"
	$(GO) test -race $(GOFLAGS) $(pkgs)

.PHONY: test-short
test-short: ## Run short tests only
	@echo ">> running short tests"
	$(GO) test -short $(GOFLAGS) $(pkgs)

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo ">> running tests with coverage"
	$(GO) test -race -coverprofile=coverage.out $(GOFLAGS) $(pkgs)
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo ">> coverage report: coverage.html"

# ============================================================================
# Build
# ============================================================================

.PHONY: build
build: $(BIN_DIR)/$(BINARY_NAME) ## Build the exporter binary

$(BIN_DIR)/$(BINARY_NAME):
	@echo ">> building $(BINARY_NAME)"
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/databricks-exporter/

.PHONY: build-all
build-all: promu ## Build for all platforms using promu
	@echo ">> building binaries for all platforms"
	$(PROMU) crossbuild

.PHONY: clean
clean: ## Remove build artifacts
	@echo ">> cleaning build artifacts"
	rm -rf $(BIN_DIR)
	rm -f coverage.out coverage.html

# ============================================================================
# Integration tests (require Databricks credentials)
# ============================================================================

.PHONY: test-integration
test-integration: ## Run integration tests (requires Databricks credentials in env)
	@echo ">> running integration tests"
	$(GO) test -tags=integration -v -timeout 10m ./collector/...

.PHONY: test-integration-fast
test-integration-fast: ## Run integration tests with shorter timeout
	@echo ">> running integration tests (fast)"
	$(GO) test -tags=integration -v -timeout 3m ./collector/... -run TestIntegration_RealDatabricks

# ============================================================================
# Help
# ============================================================================

.PHONY: help
help: ## Show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*##' Makefile | sed 's/:.*##/:/' | awk -F: '{printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
