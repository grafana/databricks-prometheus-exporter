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
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// BillingCollector collects billing and cost metrics from Databricks System Tables.
type BillingCollector struct {
	db      *sql.DB
	metrics *MetricDescriptors
	logger  log.Logger
}

// NewBillingCollector creates a new billing metrics collector.
func NewBillingCollector(logger log.Logger, db *sql.DB, metrics *MetricDescriptors) *BillingCollector {
	return &BillingCollector{
		logger:  logger,
		db:      db,
		metrics: metrics,
	}
}

// Collect retrieves and emits all billing metrics.
func (c *BillingCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	level.Debug(c.logger).Log("msg", "Collecting billing metrics")

	// Collect DBUs Total
	if err := c.collectBillingDBUs(ch); err != nil {
		level.Error(c.logger).Log("msg", "Failed to collect billing DBUs", "err", err)
		c.emitError(ch, "billing_dbus")
		// Continue collecting other metrics
	}

	// Collect Cost Estimates
	if err := c.collectBillingCost(ch); err != nil {
		level.Error(c.logger).Log("msg", "Failed to collect billing cost estimates", "err", err)
		c.emitError(ch, "billing_cost")
		// Continue collecting other metrics
	}

	// Collect Price Change Events
	if err := c.collectPriceChangeEvents(ch); err != nil {
		level.Error(c.logger).Log("msg", "Failed to collect price change events", "err", err)
		c.emitError(ch, "price_changes")
		// Continue collecting other metrics
	}

	level.Debug(c.logger).Log("msg", "Finished collecting billing metrics", "duration_seconds", time.Since(start).Seconds())
}

// collectBillingDBUs retrieves total DBU consumption per workspace and SKU.
func (c *BillingCollector) collectBillingDBUs(ch chan<- prometheus.Metric) error {
	level.Debug(c.logger).Log("msg", "Querying billing DBUs")

	rows, err := c.db.Query(billingDBUsQuery)
	if err != nil {
		return fmt.Errorf("failed to query billing DBUs: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var workspaceID, skuName sql.NullString
		var dbusTotal float64

		if err := rows.Scan(&workspaceID, &skuName, &dbusTotal); err != nil {
			level.Error(c.logger).Log("msg", "Failed to scan billing DBUs row", "err", err)
			continue
		}

		// Skip rows with NULL workspace_id or sku_name (invalid data)
		if !workspaceID.Valid || !skuName.Valid {
			level.Debug(c.logger).Log("msg", "Skipping billing DBU row with NULL workspace_id or sku_name")
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.metrics.BillingDBUsTotal,
			prometheus.GaugeValue,
			dbusTotal,
			workspaceID.String,
			skuName.String,
		)
		count++
	}

	level.Debug(c.logger).Log("msg", "Collected billing DBUs", "count", count)
	return rows.Err()
}

// collectBillingCost retrieves total cost estimates by joining usage with prices.
func (c *BillingCollector) collectBillingCost(ch chan<- prometheus.Metric) error {
	level.Debug(c.logger).Log("msg", "Querying billing cost estimates")

	rows, err := c.db.Query(billingCostEstimateQuery)
	if err != nil {
		return fmt.Errorf("failed to query billing cost: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var workspaceID, skuName sql.NullString
		var costEstimateUSD float64

		if err := rows.Scan(&workspaceID, &skuName, &costEstimateUSD); err != nil {
			level.Error(c.logger).Log("msg", "Failed to scan billing cost row", "err", err)
			continue
		}

		// Skip rows with NULL workspace_id or sku_name (invalid data)
		if !workspaceID.Valid || !skuName.Valid {
			level.Debug(c.logger).Log("msg", "Skipping billing cost row with NULL workspace_id or sku_name")
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.metrics.BillingCostEstimateUSD,
			prometheus.GaugeValue,
			costEstimateUSD,
			workspaceID.String,
			skuName.String,
		)
		count++
	}

	level.Debug(c.logger).Log("msg", "Collected billing cost estimates", "count", count)
	return rows.Err()
}

// collectPriceChangeEvents tracks price changes from the list_prices table.
func (c *BillingCollector) collectPriceChangeEvents(ch chan<- prometheus.Metric) error {
	level.Debug(c.logger).Log("msg", "Querying price change events")

	rows, err := c.db.Query(priceChangeEventsQuery)
	if err != nil {
		return fmt.Errorf("failed to query price changes: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var skuName sql.NullString
		var priceChangeCount float64

		if err := rows.Scan(&skuName, &priceChangeCount); err != nil {
			level.Error(c.logger).Log("msg", "Failed to scan price change row", "err", err)
			continue
		}

		// Skip rows with NULL sku_name (invalid data)
		if !skuName.Valid {
			level.Debug(c.logger).Log("msg", "Skipping price change row with NULL sku_name")
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.metrics.PriceChangeEvents,
			prometheus.CounterValue,
			priceChangeCount,
			skuName.String,
		)
		count++
	}

	level.Debug(c.logger).Log("msg", "Collected price change events", "count", count)
	return rows.Err()
}

// emitError emits a billing export error metric for the given stage.
func (c *BillingCollector) emitError(ch chan<- prometheus.Metric, stage string) {
	ch <- prometheus.MustNewConstMetric(
		c.metrics.BillingExportErrorsTotal,
		prometheus.CounterValue,
		1,
		stage,
	)
}
