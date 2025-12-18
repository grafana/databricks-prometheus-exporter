package collector

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// BillingCollector collects billing and cost metrics from Databricks System Tables.
type BillingCollector struct {
	db      *sql.DB
	metrics *MetricDescriptors
	logger  *slog.Logger
	ctx     context.Context
	config  *Config
}

// NewBillingCollector creates a new billing metrics collector.
func NewBillingCollector(ctx context.Context, db *sql.DB, metrics *MetricDescriptors, config *Config, logger *slog.Logger) *BillingCollector {
	return &BillingCollector{
		logger:  logger,
		db:      db,
		metrics: metrics,
		ctx:     ctx,
		config:  config,
	}
}

// Describe sends the descriptors of each metric over the provided channel.
func (c *BillingCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.metrics.BillingDBUs
	ch <- c.metrics.BillingCostEstimateUSD
	ch <- c.metrics.PriceChangeEvents
	ch <- c.metrics.BillingScrapeErrors
	ch <- c.metrics.ScrapeStatus
}

// Collect retrieves and emits all billing metrics.
// Queries run in parallel to reduce total scrape time (cost estimate query can take ~100s).
func (c *BillingCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	c.logger.Debug("Collecting billing metrics")

	var hasError atomic.Bool
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		if err := c.collectBillingDBUs(ch); err != nil {
			c.logger.Error("Failed to collect billing DBUs", "err", err)
			c.emitError(ch, "billing_dbus")
			hasError.Store(true)
		}
	}()

	go func() {
		defer wg.Done()
		if err := c.collectBillingCost(ch); err != nil {
			c.logger.Error("Failed to collect billing cost estimates", "err", err)
			c.emitError(ch, "billing_cost")
			hasError.Store(true)
		}
	}()

	go func() {
		defer wg.Done()
		if err := c.collectPriceChangeEvents(ch); err != nil {
			c.logger.Error("Failed to collect price change events", "err", err)
			c.emitError(ch, "price_changes")
			hasError.Store(true)
		}
	}()

	wg.Wait()

	// Emit scrape status
	status := 1.0
	if hasError.Load() {
		status = 0.0
	}
	ch <- prometheus.MustNewConstMetric(c.metrics.ScrapeStatus, prometheus.GaugeValue, status, "billing")

	c.logger.Debug("Finished collecting billing metrics", "duration_seconds", time.Since(start).Seconds())
}

// collectBillingDBUs retrieves total DBU consumption per workspace and SKU.
func (c *BillingCollector) collectBillingDBUs(ch chan<- prometheus.Metric) error {
	c.logger.Debug("Querying billing DBUs")

	lookback := c.config.BillingLookback
	if lookback == 0 {
		lookback = DefaultBillingLookback
	}
	query := BuildBillingDBUsQuery(lookback)
	rows, err := c.db.QueryContext(c.ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query billing DBUs: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var workspaceID, skuName sql.NullString
		var dbusTotal float64

		if err := rows.Scan(&workspaceID, &skuName, &dbusTotal); err != nil {
			c.logger.Error("Failed to scan billing DBUs row", "err", err)
			continue
		}

		// Skip rows with NULL workspace_id or sku_name (invalid data)
		if !workspaceID.Valid || !skuName.Valid {
			c.logger.Debug("Skipping billing DBU row with NULL workspace_id or sku_name")
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.metrics.BillingDBUs,
			prometheus.GaugeValue,
			dbusTotal,
			workspaceID.String,
			skuName.String,
		)
		count++
	}

	c.logger.Debug("Collected billing DBUs", "count", count)
	return rows.Err()
}

// collectBillingCost retrieves total cost estimates by joining usage with prices.
func (c *BillingCollector) collectBillingCost(ch chan<- prometheus.Metric) error {
	c.logger.Debug("Querying billing cost estimates")

	lookback := c.config.BillingLookback
	if lookback == 0 {
		lookback = DefaultBillingLookback
	}
	query := BuildBillingCostEstimateQuery(lookback)
	rows, err := c.db.QueryContext(c.ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query billing cost: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var workspaceID, skuName sql.NullString
		var costEstimateUSD float64

		if err := rows.Scan(&workspaceID, &skuName, &costEstimateUSD); err != nil {
			c.logger.Error("Failed to scan billing cost row", "err", err)
			continue
		}

		// Skip rows with NULL workspace_id or sku_name (invalid data)
		if !workspaceID.Valid || !skuName.Valid {
			c.logger.Debug("Skipping billing cost row with NULL workspace_id or sku_name")
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

	c.logger.Debug("Collected billing cost estimates", "count", count)
	return rows.Err()
}

// collectPriceChangeEvents tracks price changes from the list_prices table.
func (c *BillingCollector) collectPriceChangeEvents(ch chan<- prometheus.Metric) error {
	c.logger.Debug("Querying price change events")

	lookback := c.config.BillingLookback
	if lookback == 0 {
		lookback = DefaultBillingLookback
	}
	query := BuildPriceChangeEventsQuery(lookback)
	rows, err := c.db.QueryContext(c.ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query price changes: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var skuName sql.NullString
		var priceChangeCount float64

		if err := rows.Scan(&skuName, &priceChangeCount); err != nil {
			c.logger.Error("Failed to scan price change row", "err", err)
			continue
		}

		// Skip rows with NULL sku_name (invalid data)
		if !skuName.Valid {
			c.logger.Debug("Skipping price change row with NULL sku_name")
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.metrics.PriceChangeEvents,
			prometheus.GaugeValue,
			priceChangeCount,
			skuName.String,
		)
		count++
	}

	c.logger.Debug("Collected price change events", "count", count)
	return rows.Err()
}

// emitError emits a billing scrape error metric for the given stage.
func (c *BillingCollector) emitError(ch chan<- prometheus.Metric, stage string) {
	ch <- prometheus.MustNewConstMetric(
		c.metrics.BillingScrapeErrors,
		prometheus.GaugeValue,
		1,
		stage,
	)
}
