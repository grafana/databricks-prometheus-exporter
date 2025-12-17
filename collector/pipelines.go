package collector

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// PipelinesCollector collects pipeline-related metrics from Databricks.
type PipelinesCollector struct {
	logger *slog.Logger
	db     *sql.DB
	ctx    context.Context
	config *Config

	// Metric descriptors
	metrics *MetricDescriptors

	// Table availability tracking
	mu                   sync.RWMutex
	tableAvailable       *bool     // nil = unknown, true = available, false = unavailable
	tableLastChecked     time.Time // when we last checked
	tableCheckCounter    int       // number of scrapes since last check
	tableUnavailableOnce sync.Once // ensures we only log unavailability once
}

// NewPipelinesCollector creates a new PipelinesCollector.
func NewPipelinesCollector(ctx context.Context, db *sql.DB, metrics *MetricDescriptors, config *Config, logger *slog.Logger) *PipelinesCollector {
	return &PipelinesCollector{
		logger:  logger,
		db:      db,
		metrics: metrics,
		ctx:     ctx,
		config:  config,
	}
}

// Describe sends the descriptors of each metric over the provided channel.
func (c *PipelinesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.metrics.PipelineRunsTotal
	ch <- c.metrics.PipelineRunStatusTotal
	ch <- c.metrics.PipelineRunDurationSeconds
	ch <- c.metrics.PipelineRetryEventsTotal
	ch <- c.metrics.PipelineFreshnessLagSeconds
}

// Collect fetches metrics from Databricks and sends them to Prometheus.
func (c *PipelinesCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	c.logger.Debug("Collecting pipeline metrics")

	// Check if we should verify table availability
	if c.shouldCheckTable() {
		c.checkTableAvailability()
	}

	// Skip collection if table is known to be unavailable
	if !c.isTableAvailable() {
		c.logger.Debug("Skipping pipeline metrics collection - table unavailable")
		return
	}

	// Collect each metric, but continue on errors
	if err := c.collectPipelineRuns(ch); err != nil {
		c.handleCollectionError("pipeline runs", err)
	}

	if err := c.collectPipelineRunStatus(ch); err != nil {
		c.handleCollectionError("pipeline run status", err)
	}

	if err := c.collectPipelineRunDuration(ch); err != nil {
		c.handleCollectionError("pipeline run duration", err)
	}

	if err := c.collectPipelineRetryEvents(ch); err != nil {
		c.handleCollectionError("pipeline retry events", err)
	}

	if err := c.collectPipelineFreshnessLag(ch); err != nil {
		c.handleCollectionError("pipeline freshness lag", err)
	}

	c.logger.Debug("Finished collecting pipeline metrics", "duration_seconds", time.Since(start).Seconds())
}

// shouldCheckTable determines if we should check table availability.
// Returns true if:
// - We've never checked before (tableAvailable is nil)
// - Enough scrapes have passed since last check
func (c *PipelinesCollector) shouldCheckTable() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Always check if we haven't checked yet
	if c.tableAvailable == nil {
		return true
	}

	// If table was unavailable, check periodically
	if !*c.tableAvailable && c.tableCheckCounter >= c.config.TableCheckInterval {
		return true
	}

	return false
}

// isTableAvailable returns whether the table is available.
// Increments check counter for periodic retries.
func (c *PipelinesCollector) isTableAvailable() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.tableCheckCounter++

	// If we don't know yet, assume unavailable
	if c.tableAvailable == nil {
		return false
	}

	return *c.tableAvailable
}

// checkTableAvailability checks if the pipeline_update_timeline table exists.
func (c *PipelinesCollector) checkTableAvailability() {
	c.mu.Lock()
	c.tableLastChecked = time.Now()
	c.tableCheckCounter = 0
	c.mu.Unlock()

	// Try a simple query to check if the table exists
	query := "SELECT 1 FROM system.lakeflow.pipeline_update_timeline LIMIT 1"
	rows, err := c.db.QueryContext(c.ctx, query)
	if rows != nil {
		rows.Close()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err != nil {
		// Check if it's a "table not found" error
		if strings.Contains(strings.ToLower(err.Error()), "table_or_view_not_found") ||
			strings.Contains(strings.ToLower(err.Error()), "cannot be found") {

			available := false
			wasAvailable := c.tableAvailable != nil && *c.tableAvailable

			c.tableAvailable = &available

			// Log once when we first discover the table is unavailable
			if c.tableAvailable != nil && !wasAvailable {
				c.tableUnavailableOnce.Do(func() {
					c.logger.Warn("Pipeline metrics table not available - pipeline metrics will not be collected",
						"table", "system.lakeflow.pipeline_update_timeline",
						"note", "This is expected in some Databricks environments. All other metrics will continue to work normally.",
						"suggestion", "Contact Databricks Support to enable this table, or see documentation for more information.",
					)
				})
			}

			c.logger.Debug("Verified table is unavailable",
				"table", "system.lakeflow.pipeline_update_timeline",
				"will_retry_in_scrapes", c.config.TableCheckInterval,
			)
		} else {
			// Some other error - log it and assume unavailable for now
			c.logger.Debug("Error checking table availability",
				"table", "system.lakeflow.pipeline_update_timeline",
				"err", err,
			)
			available := false
			c.tableAvailable = &available
		}
	} else {
		available := true
		wasUnavailable := c.tableAvailable != nil && !*c.tableAvailable

		c.tableAvailable = &available

		// Log when table becomes available (was previously unavailable)
		if wasUnavailable {
			c.logger.Info("Pipeline metrics table is now available - resuming pipeline metrics collection",
				"table", "system.lakeflow.pipeline_update_timeline",
			)
		} else {
			c.logger.Debug("Verified table is available",
				"table", "system.lakeflow.pipeline_update_timeline",
			)
		}
	}
}

// handleCollectionError handles errors during metric collection.
// If it's a table-not-found error, marks the table as unavailable.
// Otherwise, logs the error.
func (c *PipelinesCollector) handleCollectionError(metricName string, err error) {
	errStr := strings.ToLower(err.Error())

	// Check if this is a table-not-found error
	if strings.Contains(errStr, "table_or_view_not_found") ||
		strings.Contains(errStr, "cannot be found") {

		c.mu.Lock()
		available := false
		c.tableAvailable = &available
		c.mu.Unlock()

		c.logger.Debug("Table became unavailable during collection",
			"metric", metricName,
		)
	} else {
		// Some other error - log it
		c.logger.Error("Failed to collect pipeline metric",
			"metric", metricName,
			"err", err,
		)
	}
}

// collectPipelineRuns collects the total number of pipeline runs per pipeline.
func (c *PipelinesCollector) collectPipelineRuns(ch chan<- prometheus.Metric) error {
	lookback := c.config.PipelinesLookback
	if lookback == 0 {
		lookback = DefaultPipelinesLookback
	}
	query := BuildPipelineRunsQuery(lookback)
	rows, err := c.db.QueryContext(c.ctx, query)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline runs query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var workspaceID, pipelineID, pipelineName sql.NullString
		var count sql.NullFloat64

		if err := rows.Scan(&workspaceID, &pipelineID, &pipelineName, &count); err != nil {
			return fmt.Errorf("failed to scan pipeline runs row: %w", err)
		}

		if count.Valid {
			ch <- prometheus.MustNewConstMetric(
				c.metrics.PipelineRunsTotal,
				prometheus.GaugeValue,
				count.Float64,
				workspaceID.String,
				pipelineID.String,
				pipelineName.String,
			)
		}
	}

	return rows.Err()
}

// collectPipelineRunStatus collects pipeline run counts by status per pipeline.
func (c *PipelinesCollector) collectPipelineRunStatus(ch chan<- prometheus.Metric) error {
	lookback := c.config.PipelinesLookback
	if lookback == 0 {
		lookback = DefaultPipelinesLookback
	}
	query := BuildPipelineRunStatusQuery(lookback)
	rows, err := c.db.QueryContext(c.ctx, query)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline run status query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var workspaceID, pipelineID, pipelineName, status sql.NullString
		var count sql.NullFloat64

		if err := rows.Scan(&workspaceID, &pipelineID, &pipelineName, &status, &count); err != nil {
			return fmt.Errorf("failed to scan pipeline run status row: %w", err)
		}

		if count.Valid {
			ch <- prometheus.MustNewConstMetric(
				c.metrics.PipelineRunStatusTotal,
				prometheus.GaugeValue,
				count.Float64,
				workspaceID.String,
				pipelineID.String,
				pipelineName.String,
				status.String,
			)
		}
	}

	return rows.Err()
}

// collectPipelineRunDuration collects pipeline run duration quantiles per pipeline.
func (c *PipelinesCollector) collectPipelineRunDuration(ch chan<- prometheus.Metric) error {
	lookback := c.config.PipelinesLookback
	if lookback == 0 {
		lookback = DefaultPipelinesLookback
	}
	query := BuildPipelineRunDurationQuery(lookback)
	rows, err := c.db.QueryContext(c.ctx, query)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline run duration query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var workspaceID, pipelineID, pipelineName sql.NullString
		var p50, p95, p99 sql.NullFloat64

		if err := rows.Scan(&workspaceID, &pipelineID, &pipelineName, &p50, &p95, &p99); err != nil {
			return fmt.Errorf("failed to scan pipeline run duration row: %w", err)
		}

		if p50.Valid {
			ch <- prometheus.MustNewConstMetric(
				c.metrics.PipelineRunDurationSeconds,
				prometheus.GaugeValue,
				p50.Float64,
				workspaceID.String,
				pipelineID.String,
				pipelineName.String,
				"0.50",
			)
		}
		if p95.Valid {
			ch <- prometheus.MustNewConstMetric(
				c.metrics.PipelineRunDurationSeconds,
				prometheus.GaugeValue,
				p95.Float64,
				workspaceID.String,
				pipelineID.String,
				pipelineName.String,
				"0.95",
			)
		}
		if p99.Valid {
			ch <- prometheus.MustNewConstMetric(
				c.metrics.PipelineRunDurationSeconds,
				prometheus.GaugeValue,
				p99.Float64,
				workspaceID.String,
				pipelineID.String,
				pipelineName.String,
				"0.99",
			)
		}
	}

	return rows.Err()
}

// collectPipelineRetryEvents collects the total number of pipeline retry events per pipeline.
func (c *PipelinesCollector) collectPipelineRetryEvents(ch chan<- prometheus.Metric) error {
	lookback := c.config.PipelinesLookback
	if lookback == 0 {
		lookback = DefaultPipelinesLookback
	}
	query := BuildPipelineRetryEventsQuery(lookback)
	rows, err := c.db.QueryContext(c.ctx, query)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline retry events query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var workspaceID, pipelineID, pipelineName sql.NullString
		var retries sql.NullFloat64

		if err := rows.Scan(&workspaceID, &pipelineID, &pipelineName, &retries); err != nil {
			return fmt.Errorf("failed to scan pipeline retry events row: %w", err)
		}

		if retries.Valid {
			ch <- prometheus.MustNewConstMetric(
				c.metrics.PipelineRetryEventsTotal,
				prometheus.GaugeValue,
				retries.Float64,
				workspaceID.String,
				pipelineID.String,
				pipelineName.String,
			)
		}
	}

	return rows.Err()
}

// collectPipelineFreshnessLag collects pipeline data freshness lag per pipeline.
func (c *PipelinesCollector) collectPipelineFreshnessLag(ch chan<- prometheus.Metric) error {
	lookback := c.config.PipelinesLookback
	if lookback == 0 {
		lookback = DefaultPipelinesLookback
	}
	query := BuildPipelineFreshnessLagQuery(lookback)
	rows, err := c.db.QueryContext(c.ctx, query)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline freshness lag query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var workspaceID, pipelineID, pipelineName sql.NullString
		var lagSeconds sql.NullFloat64

		if err := rows.Scan(&workspaceID, &pipelineID, &pipelineName, &lagSeconds); err != nil {
			return fmt.Errorf("failed to scan pipeline freshness lag row: %w", err)
		}

		if lagSeconds.Valid {
			ch <- prometheus.MustNewConstMetric(
				c.metrics.PipelineFreshnessLagSeconds,
				prometheus.GaugeValue,
				lagSeconds.Float64,
				workspaceID.String,
				pipelineID.String,
				pipelineName.String,
			)
		}
	}

	return rows.Err()
}
