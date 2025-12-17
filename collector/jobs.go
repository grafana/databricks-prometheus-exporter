package collector

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// JobsCollector collects job-related metrics from Databricks.
type JobsCollector struct {
	logger *slog.Logger
	db     *sql.DB
	ctx    context.Context
	config *Config

	metrics *MetricDescriptors
}

// NewJobsCollector creates a new JobsCollector.
func NewJobsCollector(ctx context.Context, db *sql.DB, metrics *MetricDescriptors, config *Config, logger *slog.Logger) *JobsCollector {
	return &JobsCollector{
		logger:  logger,
		db:      db,
		metrics: metrics,
		ctx:     ctx,
		config:  config,
	}
}

// Describe sends the descriptors of each metric over the provided channel.
func (c *JobsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.metrics.JobRunsTotal
	ch <- c.metrics.JobRunStatusTotal
	ch <- c.metrics.JobRunDurationSeconds
	ch <- c.metrics.TaskRetriesTotal
	ch <- c.metrics.JobSLAMissTotal
}

// Collect fetches metrics from Databricks and sends them to Prometheus.
func (c *JobsCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	c.logger.Debug("Collecting job metrics")

	if err := c.collectJobRuns(ch); err != nil {
		c.logger.Error("Failed to collect job runs", "err", err)
	}

	if err := c.collectJobRunStatus(ch); err != nil {
		c.logger.Error("Failed to collect job run status", "err", err)
	}

	if err := c.collectJobRunDuration(ch); err != nil {
		c.logger.Error("Failed to collect job run duration", "err", err)
	}

	if err := c.collectTaskRetries(ch); err != nil {
		c.logger.Error("Failed to collect task retries", "err", err)
	}

	if err := c.collectJobSLAMiss(ch); err != nil {
		c.logger.Error("Failed to collect job SLA misses", "err", err)
	}

	c.logger.Debug("Finished collecting job metrics", "duration_seconds", time.Since(start).Seconds())
}

// collectJobRuns collects the total number of job runs per job.
func (c *JobsCollector) collectJobRuns(ch chan<- prometheus.Metric) error {
	lookback := c.config.JobsLookback
	if lookback == 0 {
		lookback = DefaultJobsLookback
	}
	query := BuildJobRunsQuery(lookback)
	rows, err := c.db.QueryContext(c.ctx, query)
	if err != nil {
		return fmt.Errorf("failed to execute job runs query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var workspaceID, jobID, jobName sql.NullString
		var count sql.NullFloat64

		if err := rows.Scan(&workspaceID, &jobID, &jobName, &count); err != nil {
			return fmt.Errorf("failed to scan job runs row: %w", err)
		}

		if count.Valid {
			ch <- prometheus.MustNewConstMetric(
				c.metrics.JobRunsTotal,
				prometheus.GaugeValue,
				count.Float64,
				workspaceID.String,
				jobID.String,
				jobName.String,
			)
		}
	}

	return rows.Err()
}

// collectJobRunStatus collects job run counts by status per job.
func (c *JobsCollector) collectJobRunStatus(ch chan<- prometheus.Metric) error {
	lookback := c.config.JobsLookback
	if lookback == 0 {
		lookback = DefaultJobsLookback
	}
	query := BuildJobRunStatusQuery(lookback)
	rows, err := c.db.QueryContext(c.ctx, query)
	if err != nil {
		return fmt.Errorf("failed to execute job run status query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var workspaceID, jobID, jobName, status sql.NullString
		var count sql.NullFloat64

		if err := rows.Scan(&workspaceID, &jobID, &jobName, &status, &count); err != nil {
			return fmt.Errorf("failed to scan job run status row: %w", err)
		}

		if count.Valid {
			ch <- prometheus.MustNewConstMetric(
				c.metrics.JobRunStatusTotal,
				prometheus.GaugeValue,
				count.Float64,
				workspaceID.String,
				jobID.String,
				jobName.String,
				status.String,
			)
		}
	}

	return rows.Err()
}

// collectJobRunDuration collects job run duration quantiles per job.
func (c *JobsCollector) collectJobRunDuration(ch chan<- prometheus.Metric) error {
	lookback := c.config.JobsLookback
	if lookback == 0 {
		lookback = DefaultJobsLookback
	}
	query := BuildJobRunDurationQuery(lookback)
	rows, err := c.db.QueryContext(c.ctx, query)
	if err != nil {
		return fmt.Errorf("failed to execute job run duration query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var workspaceID, jobID, jobName sql.NullString
		var p50, p95, p99 sql.NullFloat64

		if err := rows.Scan(&workspaceID, &jobID, &jobName, &p50, &p95, &p99); err != nil {
			return fmt.Errorf("failed to scan job run duration row: %w", err)
		}

		if p50.Valid {
			ch <- prometheus.MustNewConstMetric(
				c.metrics.JobRunDurationSeconds,
				prometheus.GaugeValue,
				p50.Float64,
				workspaceID.String,
				jobID.String,
				jobName.String,
				"0.50",
			)
		}
		if p95.Valid {
			ch <- prometheus.MustNewConstMetric(
				c.metrics.JobRunDurationSeconds,
				prometheus.GaugeValue,
				p95.Float64,
				workspaceID.String,
				jobID.String,
				jobName.String,
				"0.95",
			)
		}
		if p99.Valid {
			ch <- prometheus.MustNewConstMetric(
				c.metrics.JobRunDurationSeconds,
				prometheus.GaugeValue,
				p99.Float64,
				workspaceID.String,
				jobID.String,
				jobName.String,
				"0.99",
			)
		}
	}

	return rows.Err()
}

// collectTaskRetries collects the total number of task retries per job and task.
func (c *JobsCollector) collectTaskRetries(ch chan<- prometheus.Metric) error {
	lookback := c.config.JobsLookback
	if lookback == 0 {
		lookback = DefaultJobsLookback
	}
	query := BuildTaskRetriesQuery(lookback)
	rows, err := c.db.QueryContext(c.ctx, query)
	if err != nil {
		return fmt.Errorf("failed to execute task retries query: %w", err)
	}
	defer rows.Close()

	// Skip task retries if disabled (high cardinality due to task_key)
	if !c.config.CollectTaskRetries {
		return nil
	}

	for rows.Next() {
		var workspaceID, jobID, jobName, taskKey sql.NullString
		var retries sql.NullFloat64

		if err := rows.Scan(&workspaceID, &jobID, &jobName, &taskKey, &retries); err != nil {
			return fmt.Errorf("failed to scan task retries row: %w", err)
		}

		if retries.Valid {
			ch <- prometheus.MustNewConstMetric(
				c.metrics.TaskRetriesTotal,
				prometheus.GaugeValue,
				retries.Float64,
				workspaceID.String,
				jobID.String,
				jobName.String,
				taskKey.String,
			)
		}
	}

	return rows.Err()
}

// collectJobSLAMiss collects the number of jobs that missed their SLA per job.
func (c *JobsCollector) collectJobSLAMiss(ch chan<- prometheus.Metric) error {
	lookback := c.config.JobsLookback
	if lookback == 0 {
		lookback = DefaultJobsLookback
	}
	slaThreshold := c.config.SLAThresholdSeconds
	if slaThreshold == 0 {
		slaThreshold = DefaultSLAThresholdSeconds
	}
	query := BuildJobSLAMissQuery(lookback, slaThreshold)
	rows, err := c.db.QueryContext(c.ctx, query)
	if err != nil {
		return fmt.Errorf("failed to execute job SLA miss query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var workspaceID, jobID, jobName sql.NullString
		var count sql.NullFloat64

		if err := rows.Scan(&workspaceID, &jobID, &jobName, &count); err != nil {
			return fmt.Errorf("failed to scan job SLA miss row: %w", err)
		}

		if count.Valid {
			ch <- prometheus.MustNewConstMetric(
				c.metrics.JobSLAMissTotal,
				prometheus.GaugeValue,
				count.Float64,
				workspaceID.String,
				jobID.String,
				jobName.String,
			)
		}
	}

	return rows.Err()
}
