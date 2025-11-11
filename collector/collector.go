// Copyright 2025 Grafana Labs
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
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

	// Common labels
	labelAccountID   = "account_id"
	labelWorkspaceID = "workspace_id"
	labelSKUName     = "sku_name"
	labelCloud       = "cloud"
	labelUsageUnit   = "usage_unit"
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
	config *Config
	logger log.Logger
	// For mocking
	openDatabase func(*Config) (*sql.DB, error)

	// Metric descriptors
	metrics *MetricDescriptors
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

// Describe returns all metric descriptions of the collector by emitting them down the provided channel.
// It implements prometheus.Collector.
func (c *Collector) Describe(descs chan<- *prometheus.Desc) {
	// Describe all metrics from the MetricDescriptors
	c.metrics.Describe(descs)
}

// Collect collects all metrics for this collector, and emits them through the provided channel.
// It implements prometheus.Collector.
func (c *Collector) Collect(metrics chan<- prometheus.Metric) {
	level.Debug(c.logger).Log("msg", "Collecting metrics.")

	// Open a new connection to the database each time; This makes the connection more robust to transient failures
	db, err := c.openDatabase(c.config)
	if err != nil {
		level.Error(c.logger).Log("msg", "Failed to connect to Databricks.", "err", err)
		// Emit up metric here, to indicate connection failed.
		metrics <- prometheus.MustNewConstMetric(c.metrics.Up, prometheus.GaugeValue, 0)
		return
	}
	defer db.Close()

	// Emit up=1 immediately after successful connection
	// This ensures the metric is always emitted even if subsequent collection hangs/times out
	metrics <- prometheus.MustNewConstMetric(c.metrics.Up, prometheus.GaugeValue, 1)
	level.Debug(c.logger).Log("msg", "Database connection successful, emitted up=1")

	// Initialize specialized collectors with the database connection
	billingCollector := NewBillingCollector(c.logger, db, c.metrics)
	jobsCollector := NewJobsCollector(c.logger, db, c.metrics)
	pipelinesCollector := NewPipelinesCollector(c.logger, db, c.metrics)
	sqlWarehouseCollector := NewSQLWarehouseCollector(c.logger, db, c.metrics)

	start := time.Now()

	// Create a WaitGroup to run collectors in parallel
	// This significantly reduces total collection time
	var wg sync.WaitGroup

	// Collect billing metrics in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		billingCollector.Collect(metrics)
	}()

	// Collect jobs metrics in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		jobsCollector.Collect(metrics)
	}()

	// Collect pipelines metrics in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		pipelinesCollector.Collect(metrics)
	}()

	// Collect SQL warehouse metrics in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		sqlWarehouseCollector.Collect(metrics)
	}()

	// Wait for all collectors to finish
	wg.Wait()

	level.Debug(c.logger).Log("msg", "Finished collecting metrics", "duration_seconds", time.Since(start).Seconds())
}
