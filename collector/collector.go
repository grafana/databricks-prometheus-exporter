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

	dbsql "github.com/databricks/databricks-sql-go"
	"github.com/databricks/databricks-sql-go/auth/oauth/m2m"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "databricks"

	labelAccountID   = "account_id"
	labelWorkspaceID = "workspace_id"
	labelSKUName     = "sku_name"
	labelCloud       = "cloud"
	labelUsageUnit   = "usage_unit"
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
		dbsql.WithHTTPPath(config.HTTPPath),
		dbsql.WithPort(443),
		dbsql.WithAuthenticator(authenticator),
		dbsql.WithInitialNamespace(config.Catalog, config.Schema),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create connector: %w", err)
	}

	return sql.OpenDB(connector), nil
}

// Collector is a prometheus.Collector that retrieves billing and usage metrics for a Databricks account.
type Collector struct {
	config *Config
	logger log.Logger
	// For mocking
	openDatabase func(*Config) (*sql.DB, error)

	usageQuantity *prometheus.Desc
	up            *prometheus.Desc
}

// NewCollector creates a new collector from a given config.
// The config is assumed to be valid.
func NewCollector(logger log.Logger, c *Config) *Collector {
	return &Collector{
		config:       c,
		logger:       logger,
		openDatabase: openDatabricksDatabase,
		usageQuantity: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "billing", "usage_quantity"),
			"Usage quantity information from system.billing.usage table, aggregated over the last 7 days.",
			[]string{labelAccountID, labelWorkspaceID, labelSKUName, labelCloud, labelUsageUnit},
			nil,
		),
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Metric indicating the status of the exporter collection. 1 indicates that the connection to Databricks was successful, and all available metrics were collected. "+
				"0 indicates that the exporter failed to collect 1 or more metrics, due to an inability to connect to Databricks.",
			nil,
			nil,
		),
	}
}

// Describe returns all metric descriptions of the collector by emitting them down the provided channel.
// It implements prometheus.Collector.
func (c *Collector) Describe(descs chan<- *prometheus.Desc) {
	descs <- c.usageQuantity
	descs <- c.up
}

// Collect collects all metrics for this collector, and emits them through the provided channel.
// It implements prometheus.Collector.
func (c *Collector) Collect(metrics chan<- prometheus.Metric) {
	level.Debug(c.logger).Log("msg", "Collecting metrics.")

	var up float64 = 1
	// Open a new connection to the database each time; This makes the connection more robust to transient failures
	db, err := c.openDatabase(c.config)
	if err != nil {
		level.Error(c.logger).Log("msg", "Failed to connect to Databricks.", "err", err)
		// Emit up metric here, to indicate connection failed.
		metrics <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0)
		return
	}
	defer db.Close()

	if err := c.collectBillingMetrics(db, metrics); err != nil {
		level.Error(c.logger).Log("msg", "Failed to collect billing metrics.", "err", err)
		up = 0
	}

	metrics <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, up)
	level.Debug(c.logger).Log("msg", "Finished collecting metrics.")
}

func (c *Collector) collectBillingMetrics(db *sql.DB, metrics chan<- prometheus.Metric) error {
	level.Debug(c.logger).Log("msg", "Collecting billing metrics.")
	rows, err := db.Query(billingMetricQuery)
	level.Debug(c.logger).Log("msg", "Done querying billing metrics.")
	if err != nil {
		return fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var accountID, workspaceID, skuName, cloud, usageUnit sql.NullString
		var usageQuantity sql.NullFloat64

		if err := rows.Scan(&accountID, &workspaceID, &skuName, &cloud, &usageUnit, &usageQuantity); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		if usageQuantity.Valid {
			metrics <- prometheus.MustNewConstMetric(
				c.usageQuantity,
				prometheus.GaugeValue,
				usageQuantity.Float64,
				accountID.String,
				workspaceID.String,
				skuName.String,
				cloud.String,
				usageUnit.String,
			)
		}
	}

	level.Debug(c.logger).Log("msg", "Finished collecting billing metrics.")
	return rows.Err()
}
