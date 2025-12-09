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

import "github.com/prometheus/client_golang/prometheus"

// MetricDescriptors holds all Prometheus metric descriptors for the Databricks exporter.
type MetricDescriptors struct {
	// Billing & Cost Metrics (FinOps)
	BillingDBUsTotal         *prometheus.Desc
	BillingCostEstimateUSD   *prometheus.Desc
	PriceChangeEvents        *prometheus.Desc
	BillingExportErrorsTotal *prometheus.Desc

	// Jobs Metrics (SRE/Platform)
	JobRunsTotal          *prometheus.Desc
	JobRunStatusTotal     *prometheus.Desc
	JobRunDurationSeconds *prometheus.Desc
	TaskRetriesTotal      *prometheus.Desc
	JobSLAMissTotal       *prometheus.Desc

	// Pipelines Metrics (SRE/Platform)
	PipelineRunsTotal           *prometheus.Desc
	PipelineRunStatusTotal      *prometheus.Desc
	PipelineRunDurationSeconds  *prometheus.Desc
	PipelineRetryEventsTotal    *prometheus.Desc
	PipelineFreshnessLagSeconds *prometheus.Desc

	// SQL Warehouse Metrics (Analytics/BI)
	QueriesTotal         *prometheus.Desc
	QueryDurationSeconds *prometheus.Desc
	QueryErrorsTotal     *prometheus.Desc
	QueriesRunning       *prometheus.Desc

	// Exporter health
	Up *prometheus.Desc
}

// NewMetricDescriptors creates and returns all metric descriptors for the Databricks exporter.
func NewMetricDescriptors() *MetricDescriptors {
	return &MetricDescriptors{
		// ===== Billing & Cost Metrics (FinOps) =====

		BillingDBUsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "billing", "dbus_total"),
			"Daily Databricks Units (DBUs) consumed per workspace and SKU.",
			[]string{labelWorkspaceID, labelSKUName},
			nil,
		),

		BillingCostEstimateUSD: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "billing", "cost_estimate_usd"),
			"Daily list-price cost estimate (DBUs × list price) per workspace and SKU.",
			[]string{labelWorkspaceID, labelSKUName},
			nil,
		),

		PriceChangeEvents: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "price_change_events"),
			"Count of pricing changes for a SKU from historical list prices.",
			[]string{labelSKUName},
			nil,
		),

		BillingExportErrorsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "billing", "export_errors_total"),
			"Exporter error count segmented by stage (sql, publish, etc.).",
			[]string{labelStage},
			nil,
		),

		// ===== Jobs Metrics (SRE/Platform) =====

		JobRunsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "job_runs_total"),
			"Number of Lakeflow Jobs runs per workspace and job.",
			[]string{labelWorkspaceID, labelJobID, labelJobName},
			nil,
		),

		JobRunStatusTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "job_run_status_total"),
			"Job status counts (SUCCEEDED/FAILED/CANCELED) per workspace and job.",
			[]string{labelWorkspaceID, labelJobID, labelJobName, labelStatus},
			nil,
		),

		JobRunDurationSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "job_run_duration_seconds"),
			"Job run duration quantiles (p50/p95/p99) per workspace and job.",
			[]string{labelWorkspaceID, labelJobID, labelJobName, labelQuantile},
			nil,
		),

		TaskRetriesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "task_retries_total"),
			"Number of retries across job tasks per workspace, job, and task key.",
			[]string{labelWorkspaceID, labelJobID, labelJobName, labelTaskKey},
			nil,
		),

		JobSLAMissTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "job_sla_miss_total"),
			"Job runs exceeding configured SLA threshold per workspace and job.",
			[]string{labelWorkspaceID, labelJobID, labelJobName},
			nil,
		),

		// ===== Pipelines Metrics (SRE/Platform) =====

		PipelineRunsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pipeline_runs_total"),
			"DLT / Lakeflow Pipelines executions per workspace and pipeline.",
			[]string{labelWorkspaceID, labelPipelineID, labelPipelineName},
			nil,
		),

		PipelineRunStatusTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pipeline_run_status_total"),
			"Pipeline run status counts (COMPLETED/FAILED) per workspace and pipeline.",
			[]string{labelWorkspaceID, labelPipelineID, labelPipelineName, labelStatus},
			nil,
		),

		PipelineRunDurationSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pipeline_run_duration_seconds"),
			"Pipeline run duration quantiles (p50/p95/p99) per workspace and pipeline.",
			[]string{labelWorkspaceID, labelPipelineID, labelPipelineName, labelQuantile},
			nil,
		),

		PipelineRetryEventsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pipeline_retry_events_total"),
			"Retry/backoff events within pipeline updates per workspace and pipeline.",
			[]string{labelWorkspaceID, labelPipelineID, labelPipelineName},
			nil,
		),

		PipelineFreshnessLagSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pipeline_freshness_lag_seconds"),
			"Data freshness lag vs target watermark per workspace and pipeline.",
			[]string{labelWorkspaceID, labelPipelineID, labelPipelineName},
			nil,
		),

		// ===== SQL Warehouse Metrics (Analytics/BI) =====

		QueriesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "queries_total"),
			"SQL queries executed (warehouse & serverless) per workspace and warehouse.",
			[]string{labelWorkspaceID, labelWarehouseID},
			nil,
		),

		QueryDurationSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "query_duration_seconds"),
			"Query latency quantiles (p50/p95/p99) per workspace and warehouse.",
			[]string{labelWorkspaceID, labelWarehouseID, labelQuantile},
			nil,
		),

		QueryErrorsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "query_errors_total"),
			"Failed queries count per workspace and warehouse.",
			[]string{labelWorkspaceID, labelWarehouseID},
			nil,
		),

		QueriesRunning: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "queries_running"),
			"Concurrent/running queries per workspace and warehouse (derived from overlapping intervals).",
			[]string{labelWorkspaceID, labelWarehouseID},
			nil,
		),

		// ===== Exporter Health =====

		Up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Metric indicating whether the exporter successfully connected to Databricks. "+
				"1 indicates successful database connection. "+
				"0 indicates failure to establish database connection. "+
				"Note: This metric is emitted early in the scrape cycle; individual query failures are logged but do not affect this metric.",
			nil,
			nil,
		),
	}
}

// Describe sends all metric descriptors to the provided channel.
// This implements the prometheus.Collector interface.
func (m *MetricDescriptors) Describe(ch chan<- *prometheus.Desc) {
	// Billing & Cost
	ch <- m.BillingDBUsTotal
	ch <- m.BillingCostEstimateUSD
	ch <- m.PriceChangeEvents
	ch <- m.BillingExportErrorsTotal

	// Jobs
	ch <- m.JobRunsTotal
	ch <- m.JobRunStatusTotal
	ch <- m.JobRunDurationSeconds
	ch <- m.TaskRetriesTotal
	ch <- m.JobSLAMissTotal

	// Pipelines
	ch <- m.PipelineRunsTotal
	ch <- m.PipelineRunStatusTotal
	ch <- m.PipelineRunDurationSeconds
	ch <- m.PipelineRetryEventsTotal
	ch <- m.PipelineFreshnessLagSeconds

	// SQL Warehouse
	ch <- m.QueriesTotal
	ch <- m.QueryDurationSeconds
	ch <- m.QueryErrorsTotal
	ch <- m.QueriesRunning

	// Health
	ch <- m.Up
}
