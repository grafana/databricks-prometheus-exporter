package collector

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// SQLWarehouseCollector collects SQL warehouse-related metrics from Databricks.
type SQLWarehouseCollector struct {
	logger *slog.Logger
	db     *sql.DB
	ctx    context.Context
	config *Config

	// Metric descriptors
	metrics *MetricDescriptors
}

// NewSQLWarehouseCollector creates a new SQLWarehouseCollector.
func NewSQLWarehouseCollector(ctx context.Context, db *sql.DB, metrics *MetricDescriptors, config *Config, logger *slog.Logger) *SQLWarehouseCollector {
	return &SQLWarehouseCollector{
		logger:  logger,
		db:      db,
		metrics: metrics,
		ctx:     ctx,
		config:  config,
	}
}

// Describe sends the descriptors of each metric over the provided channel.
func (c *SQLWarehouseCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.metrics.QueriesTotal
	ch <- c.metrics.QueryDurationSeconds
	ch <- c.metrics.QueryErrorsTotal
	ch <- c.metrics.QueriesRunning
	ch <- c.metrics.ScrapeStatus
}

// Collect fetches metrics from Databricks and sends them to Prometheus.
func (c *SQLWarehouseCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	c.logger.Debug("Collecting SQL warehouse metrics")

	var hasError bool

	// Collect each metric, but continue on errors
	if err := c.collectQueries(ch); err != nil {
		c.logger.Error("Failed to collect queries", "err", err)
		hasError = true
	}

	if err := c.collectQueryErrors(ch); err != nil {
		c.logger.Error("Failed to collect query errors", "err", err)
		hasError = true
	}

	if err := c.collectQueryDuration(ch); err != nil {
		c.logger.Error("Failed to collect query duration", "err", err)
		hasError = true
	}

	if err := c.collectQueriesRunning(ch); err != nil {
		c.logger.Error("Failed to collect running queries", "err", err)
		hasError = true
	}

	// Emit scrape status
	status := 1.0
	if hasError {
		status = 0.0
	}
	ch <- prometheus.MustNewConstMetric(c.metrics.ScrapeStatus, prometheus.GaugeValue, status, "queries")

	c.logger.Debug("Finished collecting SQL warehouse metrics", "duration_seconds", time.Since(start).Seconds())
}

// collectQueries collects the total number of queries executed per warehouse.
func (c *SQLWarehouseCollector) collectQueries(ch chan<- prometheus.Metric) error {
	lookback := c.config.QueriesLookback
	if lookback == 0 {
		lookback = DefaultQueriesLookback
	}
	query := BuildQueriesQuery(lookback)
	rows, err := c.db.QueryContext(c.ctx, query)
	if err != nil {
		return fmt.Errorf("failed to execute queries query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var workspaceID, warehouseID sql.NullString
		var count sql.NullFloat64

		if err := rows.Scan(&workspaceID, &warehouseID, &count); err != nil {
			return fmt.Errorf("failed to scan queries row: %w", err)
		}

		if count.Valid {
			ch <- prometheus.MustNewConstMetric(
				c.metrics.QueriesTotal,
				prometheus.GaugeValue, // Gauge because this is a sliding window count that can decrease
				count.Float64,
				workspaceID.String,
				warehouseID.String,
			)
		}
	}

	return rows.Err()
}

// collectQueryErrors collects the total number of query errors per warehouse.
func (c *SQLWarehouseCollector) collectQueryErrors(ch chan<- prometheus.Metric) error {
	lookback := c.config.QueriesLookback
	if lookback == 0 {
		lookback = DefaultQueriesLookback
	}
	query := BuildQueryErrorsQuery(lookback)
	rows, err := c.db.QueryContext(c.ctx, query)
	if err != nil {
		return fmt.Errorf("failed to execute query errors query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var workspaceID, warehouseID sql.NullString
		var count sql.NullFloat64

		if err := rows.Scan(&workspaceID, &warehouseID, &count); err != nil {
			return fmt.Errorf("failed to scan query errors row: %w", err)
		}

		if count.Valid {
			ch <- prometheus.MustNewConstMetric(
				c.metrics.QueryErrorsTotal,
				prometheus.GaugeValue, // Gauge because this is a sliding window count that can decrease
				count.Float64,
				workspaceID.String,
				warehouseID.String,
			)
		}
	}

	return rows.Err()
}

// collectQueryDuration collects query duration quantiles per warehouse.
func (c *SQLWarehouseCollector) collectQueryDuration(ch chan<- prometheus.Metric) error {
	lookback := c.config.QueriesLookback
	if lookback == 0 {
		lookback = DefaultQueriesLookback
	}
	query := BuildQueryDurationQuery(lookback)
	rows, err := c.db.QueryContext(c.ctx, query)
	if err != nil {
		return fmt.Errorf("failed to execute query duration query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var workspaceID, warehouseID sql.NullString
		var p50, p95, p99 sql.NullFloat64

		if err := rows.Scan(&workspaceID, &warehouseID, &p50, &p95, &p99); err != nil {
			return fmt.Errorf("failed to scan query duration row: %w", err)
		}

		// Emit p50
		if p50.Valid {
			ch <- prometheus.MustNewConstMetric(
				c.metrics.QueryDurationSeconds,
				prometheus.GaugeValue,
				p50.Float64,
				workspaceID.String,
				warehouseID.String,
				"0.50",
			)
		}

		// Emit p95
		if p95.Valid {
			ch <- prometheus.MustNewConstMetric(
				c.metrics.QueryDurationSeconds,
				prometheus.GaugeValue,
				p95.Float64,
				workspaceID.String,
				warehouseID.String,
				"0.95",
			)
		}

		// Emit p99
		if p99.Valid {
			ch <- prometheus.MustNewConstMetric(
				c.metrics.QueryDurationSeconds,
				prometheus.GaugeValue,
				p99.Float64,
				workspaceID.String,
				warehouseID.String,
				"0.99",
			)
		}
	}

	return rows.Err()
}

// collectQueriesRunning collects the current number of running queries per warehouse.
func (c *SQLWarehouseCollector) collectQueriesRunning(ch chan<- prometheus.Metric) error {
	lookback := c.config.QueriesLookback
	if lookback == 0 {
		lookback = DefaultQueriesLookback
	}
	query := BuildQueriesRunningQuery(lookback)
	rows, err := c.db.QueryContext(c.ctx, query)
	if err != nil {
		return fmt.Errorf("failed to execute running queries query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var workspaceID, warehouseID sql.NullString
		var count sql.NullFloat64

		if err := rows.Scan(&workspaceID, &warehouseID, &count); err != nil {
			return fmt.Errorf("failed to scan running queries row: %w", err)
		}

		if count.Valid {
			ch <- prometheus.MustNewConstMetric(
				c.metrics.QueriesRunning,
				prometheus.GaugeValue,
				count.Float64,
				workspaceID.String,
				warehouseID.String,
			)
		}
	}

	return rows.Err()
}
