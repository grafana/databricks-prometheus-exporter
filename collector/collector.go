package collector

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	dbsql "github.com/databricks/databricks-sql-go"
	"github.com/databricks/databricks-sql-go/auth/oauth/m2m"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "databricks"

	// Common labels used in metrics
	labelWorkspaceID = "workspace_id"
	labelSKUName     = "sku_name"
	labelStatus      = "status"
	labelStage       = "stage"
	labelQuantile    = "quantile"

	// Resource identification labels
	labelJobID        = "job_id"
	labelJobName      = "job_name"
	labelPipelineID   = "pipeline_id"
	labelPipelineName = "pipeline_name"
	labelTaskKey      = "task_key"
	labelWarehouseID  = "warehouse_id"
)

// openDatabricksDatabase opens a connection to a Databricks SQL Warehouse using OAuth2 M2M authentication.
func openDatabricksDatabase(config *Config) (*sql.DB, error) {
	// Create OAuth M2M authenticator with Service Principal credentials
	authenticator := m2m.NewAuthenticator(
		config.ClientID,
		config.ClientSecret,
		config.ServerHostname,
	)

	// Create connector with OAuth authentication
	connector, err := dbsql.NewConnector(
		dbsql.WithServerHostname(config.ServerHostname),
		dbsql.WithHTTPPath(config.WarehouseHTTPPath),
		dbsql.WithPort(443),
		dbsql.WithAuthenticator(authenticator),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create connector: %w", err)
	}

	db := sql.OpenDB(connector)

	// Configure connection pool for better resilience
	db.SetMaxOpenConns(10)                 // Limit concurrent connections to avoid overwhelming Databricks
	db.SetMaxIdleConns(5)                  // Keep some connections warm
	db.SetConnMaxLifetime(5 * time.Minute) // Recycle connections every 5 minutes
	db.SetConnMaxIdleTime(1 * time.Minute) // Close idle connections after 1 minute

	return db, nil
}

// Collector is a prometheus.Collector that retrieves all metrics for a Databricks account.
// It orchestrates multiple specialized collectors for different metric categories.
type Collector struct {
	config       *Config
	logger       log.Logger
	openDatabase func(*Config) (*sql.DB, error) // For mocking
	metrics      *MetricDescriptors

	// Persistent connection pool - reused across scrapes
	db   *sql.DB
	dbMu sync.RWMutex
}

// NewCollector creates a new collector from a given config.
// The config is assumed to be valid.
func NewCollector(logger log.Logger, c *Config) *Collector {
	metrics := NewMetricDescriptors()

	return &Collector{
		config:       c,
		logger:       logger,
		openDatabase: openDatabricksDatabase,
		metrics:      metrics,
	}
}

// getDB returns a healthy database connection, creating one if needed.
// It tests the connection with Ping() and recreates if unhealthy.
func (c *Collector) getDB() (*sql.DB, error) {
	c.dbMu.RLock()
	db := c.db
	c.dbMu.RUnlock()

	// Test existing connection
	if db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err == nil {
			return db, nil
		}
		level.Warn(c.logger).Log("msg", "Existing connection unhealthy, reconnecting", "err", "ping failed")
	}

	// Need to create or recreate connection
	c.dbMu.Lock()
	defer c.dbMu.Unlock()

	// Double-check after acquiring write lock
	if c.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := c.db.PingContext(ctx); err == nil {
			return c.db, nil
		}
		// Close unhealthy connection
		c.db.Close()
		c.db = nil
	}

	// Create new connection
	db, err := c.openDatabase(c.config)
	if err != nil {
		return nil, err
	}

	c.db = db
	level.Debug(c.logger).Log("msg", "Created new database connection pool")
	return db, nil
}

// Describe implements prometheus.Collector.
func (c *Collector) Describe(descs chan<- *prometheus.Desc) {
	c.metrics.Describe(descs)
}

// Collect collects all metrics for this collector, and emits them through the provided channel.
// It implements prometheus.Collector.
func (c *Collector) Collect(metrics chan<- prometheus.Metric) {
	level.Debug(c.logger).Log("msg", "Collecting metrics.")

	// Get a healthy connection from the pool (creates one if needed)
	db, err := c.getDB()
	if err != nil {
		level.Error(c.logger).Log("msg", "Failed to connect to Databricks.", "err", err)
		metrics <- prometheus.MustNewConstMetric(c.metrics.Up, prometheus.GaugeValue, 0)
		return
	}
	// Don't close - connection is reused across scrapes

	// Emit up=1 early so it's always reported even if collection hangs
	metrics <- prometheus.MustNewConstMetric(c.metrics.Up, prometheus.GaugeValue, 1)
	level.Debug(c.logger).Log("msg", "Database connection healthy, emitted up=1")

	queryTimeout := c.config.QueryTimeout
	if queryTimeout == 0 {
		queryTimeout = DefaultQueryTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	billingCollector := NewBillingCollector(c.logger, db, c.metrics, ctx, c.config)
	jobsCollector := NewJobsCollector(c.logger, db, c.metrics, ctx, c.config)
	pipelinesCollector := NewPipelinesCollector(c.logger, db, c.metrics, ctx, c.config)
	sqlWarehouseCollector := NewSQLWarehouseCollector(c.logger, db, c.metrics, ctx, c.config)

	start := time.Now()

	// Run collectors in parallel to reduce total scrape time
	var wg sync.WaitGroup
	wg.Add(4)
	go func() { defer wg.Done(); billingCollector.Collect(metrics) }()
	go func() { defer wg.Done(); jobsCollector.Collect(metrics) }()
	go func() { defer wg.Done(); pipelinesCollector.Collect(metrics) }()
	go func() { defer wg.Done(); sqlWarehouseCollector.Collect(metrics) }()
	wg.Wait()

	level.Debug(c.logger).Log("msg", "Finished collecting metrics", "duration_seconds", time.Since(start).Seconds())
}
