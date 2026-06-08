BIN_DIR     := bin
BINARY_NAME := databricks-exporter
GO          := go
GOFLAGS     :=
pkgs        := ./...

.PHONY: all
all: check build

.PHONY: check
check: vet lint fmt security-check test-exporter-unit

.PHONY: vet
vet:
	@echo ">> running go vet"
	$(GO) vet $(GOFLAGS) $(pkgs)

.PHONY: lint
lint: fmt vet
	@echo ">> running golangci-lint"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run $(pkgs); \
	else \
		echo "golangci-lint not installed, skipping (install: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2)"; \
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

FIRST_GOPATH  := $(firstword $(subst :, ,$(shell go env GOPATH)))
GOVULNCHECK    = $(FIRST_GOPATH)/bin/govulncheck

.PHONY: vuln-check
vuln-check:
	@echo ">> Running govulncheck..."
	@command -v $(GOVULNCHECK) >/dev/null 2>&1 || { echo "govulncheck not installed. Install: go install golang.org/x/vuln/cmd/govulncheck@0782b76014f15f24e22a438f30f308df42899ba1 # v1.3.0"; exit 1; }
	$(GOVULNCHECK) ./...
	@echo ">> govulncheck passed!"

.PHONY: gosec-check
gosec-check:
	@echo ">> Running gosec via golangci-lint..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. Install: https://golangci-lint.run/docs/welcome/install/"; exit 1; }
	golangci-lint run --enable-only gosec $(pkgs)
	@echo ">> Security checks passed!"

.PHONY: security-check
security-check: vuln-check gosec-check

###
### Jsonnet
###

# Check if .github/workflows/*.yml need to be updated
# when changing the install-ci-deps target.
.PHONY: install-ci-deps
install-ci-deps:
	go install github.com/google/go-jsonnet/cmd/jsonnet@v0.20.0
	go install github.com/google/go-jsonnet/cmd/jsonnetfmt@v0.20.0
	go install github.com/google/go-jsonnet/cmd/jsonnet-lint@v0.20.0
	go install github.com/monitoring-mixins/mixtool/cmd/mixtool@ea35232b9d85b4cd7943b481c6f90fd94f1ec0ca # main @ 2026-05-04 (no tagged releases)
	go install github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb@v0.5.1
	go install github.com/grafana/grizzly/cmd/grr@v0.7.1 # v0.7.1
