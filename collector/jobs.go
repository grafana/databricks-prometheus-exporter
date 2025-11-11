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

// JobsCollector collects job-related metrics from Databricks.
type JobsCollector struct {
	logger log.Logger
	db     *sql.DB

	// Metric descriptors
	metrics *MetricDescriptors
}

// NewJobsCollector creates a new JobsCollector.
func NewJobsCollector(logger log.Logger, db *sql.DB, metrics *MetricDescriptors) *JobsCollector {
	return &JobsCollector{
		logger:  logger,
		db:      db,
		metrics: metrics,
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
	level.Debug(c.logger).Log("msg", "Collecting job metrics")

	// Collect each metric, but continue on errors
	if err := c.collectJobRuns(ch); err != nil {
		level.Error(c.logger).Log("msg", "Failed to collect job runs", "err", err)
	}

	if err := c.collectJobRunStatus(ch); err != nil {
		level.Error(c.logger).Log("msg", "Failed to collect job run status", "err", err)
	}

	if err := c.collectJobRunDuration(ch); err != nil {
		level.Error(c.logger).Log("msg", "Failed to collect job run duration", "err", err)
	}

	if err := c.collectTaskRetries(ch); err != nil {
		level.Error(c.logger).Log("msg", "Failed to collect task retries", "err", err)
	}

	if err := c.collectJobSLAMiss(ch); err != nil {
		level.Error(c.logger).Log("msg", "Failed to collect job SLA misses", "err", err)
	}

	level.Debug(c.logger).Log("msg", "Finished collecting job metrics", "duration_seconds", time.Since(start).Seconds())
}

// collectJobRuns collects the total number of job runs per job.
func (c *JobsCollector) collectJobRuns(ch chan<- prometheus.Metric) error {
	rows, err := c.db.Query(jobRunsQuery)
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
				prometheus.CounterValue,
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
	rows, err := c.db.Query(jobRunStatusQuery)
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
				prometheus.CounterValue,
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
	rows, err := c.db.Query(jobRunDurationQuery)
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

		// Emit p50
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

		// Emit p95
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

		// Emit p99
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
	rows, err := c.db.Query(taskRetriesQuery)
	if err != nil {
		return fmt.Errorf("failed to execute task retries query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var workspaceID, jobID, jobName, taskKey sql.NullString
		var retries sql.NullFloat64

		if err := rows.Scan(&workspaceID, &jobID, &jobName, &taskKey, &retries); err != nil {
			return fmt.Errorf("failed to scan task retries row: %w", err)
		}

		if retries.Valid {
			ch <- prometheus.MustNewConstMetric(
				c.metrics.TaskRetriesTotal,
				prometheus.CounterValue,
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
	rows, err := c.db.Query(jobSLAMissQuery)
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
				prometheus.CounterValue,
				count.Float64,
				workspaceID.String,
				jobID.String,
				jobName.String,
			)
		}
	}

	return rows.Err()
}
