BIN_DIR     := bin
BINARY_NAME := databricks-exporter
GO          := go
GOFLAGS     :=
pkgs        := ./...

.PHONY: all
all: check build

.PHONY: check
check: vet lint fmt test-exporter-unit

.PHONY: vet
vet:
	@echo ">> running go vet"
	$(GO) vet $(GOFLAGS) $(pkgs)

.PHONY: lint
lint:
	@echo ">> running golangci-lint"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run $(pkgs); \
	else \
		echo "golangci-lint not installed, skipping (install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)"; \
	fi

.PHONY: fmt
fmt:
	@echo ">> checking code formatting"
	@fmtRes=$$(gofmt -l $$(find . -name '*.go' -not -path './vendor/*' -not -path './mixin/*')); \
	if [ -n "$${fmtRes}" ]; then \
		echo "gofmt found issues in:"; echo "$${fmtRes}"; \
		echo "Run 'make fmt-fix' to fix"; \
		exit 1; \
	fi

.PHONY: fmt-fix
fmt-fix:
	$(GO) fmt $(pkgs)

.PHONY: test
test: test-exporter-unit test-exporter-e2e

.PHONY: test-exporter-unit
test-exporter-unit:
	@echo ">> running unit tests"
	$(GO) test -race $(GOFLAGS) $(pkgs)

.PHONY: test-exporter-e2e
test-exporter-e2e:
	@echo ">> running e2e tests"
	$(GO) test -tags=integration -v -timeout 10m ./collector/...

.PHONY: test-coverage
test-coverage:
	@echo ">> running tests with coverage"
	$(GO) test -race -coverprofile=coverage.out $(GOFLAGS) $(pkgs)
	$(GO) tool cover -html=coverage.out -o coverage.html

.PHONY: build
build:
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/databricks-exporter/

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)
	rm -f coverage.out coverage.html
