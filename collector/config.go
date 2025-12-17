package collector

import (
	"errors"
	"time"
)

// Time constants
const (
	SecondsPerHour = 60 * 60 // 3600 seconds
)

// Default values for configuration options.
const (
	DefaultQueryTimeout        = 5 * time.Minute
	DefaultBillingLookback     = 24 * time.Hour
	DefaultJobsLookback        = 2 * time.Hour
	DefaultPipelinesLookback   = 2 * time.Hour
	DefaultQueriesLookback     = 1 * time.Hour
	DefaultSLAThresholdSeconds = SecondsPerHour
	DefaultTableCheckInterval  = 10 // Number of scrapes between table availability checks
)

// Config holds the configuration for the Databricks exporter.
type Config struct {
	// Exporter metadata
	Version string // Exporter version for info metric

	// Connection settings
	ServerHostname    string
	WarehouseHTTPPath string
	ClientID          string
	ClientSecret      string

	// Query settings
	QueryTimeout time.Duration // Timeout for individual database queries

	// Lookback windows for different metric domains
	BillingLookback   time.Duration // How far back to look for billing data
	JobsLookback      time.Duration // How far back to look for job runs
	PipelinesLookback time.Duration // How far back to look for pipeline runs
	QueriesLookback   time.Duration // How far back to look for SQL warehouse queries

	// SLA settings
	SLAThresholdSeconds int // Duration threshold (in seconds) for SLA miss detection

	// Cardinality controls
	CollectTaskRetries bool // Collect task retry metrics (high cardinality due to task_key)

	// Table availability settings
	TableCheckInterval int // Number of scrapes between table availability checks (for optional tables like pipelines)
}

var (
	errNoServerHostname    = errors.New("server_hostname must be specified")
	errNoWarehouseHTTPPath = errors.New("warehouse_http_path must be specified")
	errNoClientID          = errors.New("client_id must be specified")
	errNoClientSecret      = errors.New("client_secret must be specified")
)

// DefaultConfig returns a Config with all default values set.
// Useful for tests that don't need specific config values.
func DefaultConfig() *Config {
	return &Config{
		Version:             "unknown", // Set by main.go from build info
		QueryTimeout:        DefaultQueryTimeout,
		BillingLookback:     DefaultBillingLookback,
		JobsLookback:        DefaultJobsLookback,
		PipelinesLookback:   DefaultPipelinesLookback,
		QueriesLookback:     DefaultQueriesLookback,
		SLAThresholdSeconds: DefaultSLAThresholdSeconds,
		CollectTaskRetries:  false,
		TableCheckInterval:  DefaultTableCheckInterval,
	}
}

func (c Config) Validate() error {
	if c.ServerHostname == "" {
		return errNoServerHostname
	}

	if c.WarehouseHTTPPath == "" {
		return errNoWarehouseHTTPPath
	}

	if c.ClientID == "" {
		return errNoClientID
	}

	if c.ClientSecret == "" {
		return errNoClientSecret
	}

	return nil
}
