package collector

import "github.com/prometheus/client_golang/prometheus"

// MetricDescriptors holds all Prometheus metric descriptors for the Databricks exporter.
type MetricDescriptors struct {
	// Billing & Cost Metrics (FinOps)
	BillingDBUs            *prometheus.Desc
	BillingCostEstimateUSD *prometheus.Desc
	PriceChangeEvents      *prometheus.Desc
	BillingScrapeErrors    *prometheus.Desc

	// Jobs Metrics (SRE/Platform)
	JobRuns               *prometheus.Desc
	JobRunStatus          *prometheus.Desc
	JobRunDurationSeconds *prometheus.Desc
	TaskRetries           *prometheus.Desc
	JobSLAMiss            *prometheus.Desc

	// Pipelines Metrics (SRE/Platform)
	PipelineRuns                *prometheus.Desc
	PipelineRunStatus           *prometheus.Desc
	PipelineRunDurationSeconds  *prometheus.Desc
	PipelineRetryEvents         *prometheus.Desc
	PipelineFreshnessLagSeconds *prometheus.Desc

	// SQL Warehouse Metrics (Analytics/BI)
	Queries              *prometheus.Desc
	QueryDurationSeconds *prometheus.Desc
	QueryErrors          *prometheus.Desc
	QueriesRunning       *prometheus.Desc

	// Exporter health
	ExporterUp *prometheus.Desc

	// Scrape status (per-query health)
	ScrapeStatus *prometheus.Desc

	// Exporter info (version and configuration)
	ExporterInfo *prometheus.Desc
}

// NewMetricDescriptors creates and returns all metric descriptors for the Databricks exporter.
func NewMetricDescriptors() *MetricDescriptors {
	return &MetricDescriptors{
		// ===== Billing & Cost Metrics (FinOps) =====

		BillingDBUs: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "billing", "dbus_sliding"),
			"Databricks Units (DBUs) consumed per workspace and SKU. "+
				"Note: Databricks billing data has 24-48h lag from actual usage. "+
				"Sliding window configurable via --billing-lookback (default: 24h).",
			[]string{labelWorkspaceID, labelSKUName},
			nil,
		),

		BillingCostEstimateUSD: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "billing", "cost_estimate_usd_sliding"),
			"List-price cost estimate (DBUs × list price) per workspace and SKU. "+
				"Note: Databricks billing data has 24-48h lag from actual usage. "+
				"Sliding window configurable via --billing-lookback (default: 24h).",
			[]string{labelWorkspaceID, labelSKUName},
			nil,
		),

		PriceChangeEvents: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "price_change_events_sliding"),
			"Pricing changes for a SKU. "+
				"Note: Databricks billing data has 24-48h lag. "+
				"Sliding window configurable via --billing-lookback (default: 24h).",
			[]string{labelSKUName},
			nil,
		),

		BillingScrapeErrors: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "billing", "scrape_errors"),
			"Billing scrape errors by stage (1 if error occurred this scrape, 0 otherwise).",
			[]string{labelStage},
			nil,
		),

		// ===== Jobs Metrics (SRE/Platform) =====

		JobRuns: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "job_runs_sliding"),
			"Lakeflow Jobs runs per workspace and job (sliding window, configurable via --jobs-lookback, default: 3h).",
			[]string{labelWorkspaceID, labelJobID, labelJobName},
			nil,
		),

		JobRunStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "job_run_status_sliding"),
			"Job status counts (SUCCEEDED/FAILED/CANCELED) per workspace and job (sliding window, configurable via --jobs-lookback, default: 3h).",
			[]string{labelWorkspaceID, labelJobID, labelJobName, labelStatus},
			nil,
		),

		JobRunDurationSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "job_run_duration_seconds_sliding"),
			"Job run duration quantiles (p50/p95/p99) per workspace and job (sliding window, configurable via --jobs-lookback, default: 3h).",
			[]string{labelWorkspaceID, labelJobID, labelJobName, labelQuantile},
			nil,
		),

		TaskRetries: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "task_retries_sliding"),
			"Retries across job tasks per workspace, job, and task key (sliding window, configurable via --jobs-lookback, default: 3h).",
			[]string{labelWorkspaceID, labelJobID, labelJobName, labelTaskKey},
			nil,
		),

		JobSLAMiss: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "job_sla_miss_sliding"),
			"Job runs exceeding SLA threshold (configurable via --sla-threshold) per workspace and job (sliding window, configurable via --jobs-lookback, default: 3h).",
			[]string{labelWorkspaceID, labelJobID, labelJobName},
			nil,
		),

		// ===== Pipelines Metrics (SRE/Platform) =====

		PipelineRuns: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pipeline_runs_sliding"),
			"DLT / Lakeflow Pipelines executions per workspace and pipeline (sliding window, configurable via --pipelines-lookback, default: 3h).",
			[]string{labelWorkspaceID, labelPipelineID, labelPipelineName},
			nil,
		),

		PipelineRunStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pipeline_run_status_sliding"),
			"Pipeline run status counts (COMPLETED/FAILED) per workspace and pipeline (sliding window, configurable via --pipelines-lookback, default: 3h).",
			[]string{labelWorkspaceID, labelPipelineID, labelPipelineName, labelStatus},
			nil,
		),

		PipelineRunDurationSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pipeline_run_duration_seconds_sliding"),
			"Pipeline run duration quantiles (p50/p95/p99) per workspace and pipeline (sliding window, configurable via --pipelines-lookback, default: 3h).",
			[]string{labelWorkspaceID, labelPipelineID, labelPipelineName, labelQuantile},
			nil,
		),

		PipelineRetryEvents: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pipeline_retry_events_sliding"),
			"Retry/backoff events within pipeline updates per workspace and pipeline (sliding window, configurable via --pipelines-lookback, default: 3h).",
			[]string{labelWorkspaceID, labelPipelineID, labelPipelineName},
			nil,
		),

		PipelineFreshnessLagSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pipeline_freshness_lag_seconds_sliding"),
			"Data freshness lag vs target watermark per workspace and pipeline (point-in-time, derived from latest pipeline runs within lookback window).",
			[]string{labelWorkspaceID, labelPipelineID, labelPipelineName},
			nil,
		),
		// ===== SQL Warehouse Metrics (Analytics/BI) =====

		Queries: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "queries_sliding"),
			"SQL queries executed (warehouse & serverless) per workspace and warehouse (sliding window, configurable via --queries-lookback, default: 2h).",
			[]string{labelWorkspaceID, labelWarehouseID},
			nil,
		),

		QueryDurationSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "query_duration_seconds_sliding"),
			"Query latency quantiles (p50/p95/p99) per workspace and warehouse (sliding window, configurable via --queries-lookback, default: 2h).",
			[]string{labelWorkspaceID, labelWarehouseID, labelQuantile},
			nil,
		),

		QueryErrors: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "query_errors_sliding"),
			"Failed queries per workspace and warehouse (sliding window, configurable via --queries-lookback, default: 2h).",
			[]string{labelWorkspaceID, labelWarehouseID},
			nil,
		),

		QueriesRunning: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "queries_running_sliding"),
			"Concurrent/running queries per workspace and warehouse (derived from overlapping intervals within lookback window).",
			[]string{labelWorkspaceID, labelWarehouseID},
			nil,
		),

		// ===== Exporter Health =====

		ExporterUp: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "exporter_up"),
			"Whether the exporter successfully connected to Databricks. "+
				"1 = connection established, 0 = connection failed. "+
				"Note: Individual query failures do not affect this metric.",
			nil,
			nil,
		),

		ScrapeStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "scrape_status"),
			"Status of individual scrape queries. "+
				"1 = success, 0 = failure (timeout, error, or table unavailable).",
			[]string{labelQuery},
			nil,
		),

		ExporterInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "exporter_info"),
			"Build and configuration information for the exporter.",
			[]string{"version", "billing_window", "jobs_window", "pipelines_window", "queries_window"},
			nil,
		),
	}
}

// Describe sends all metric descriptors to the provided channel.
// This implements the prometheus.Collector interface.
func (m *MetricDescriptors) Describe(ch chan<- *prometheus.Desc) {
	// Billing & Cost
	ch <- m.BillingDBUs
	ch <- m.BillingCostEstimateUSD
	ch <- m.PriceChangeEvents
	ch <- m.BillingScrapeErrors

	// Jobs
	ch <- m.JobRuns
	ch <- m.JobRunStatus
	ch <- m.JobRunDurationSeconds
	ch <- m.TaskRetries
	ch <- m.JobSLAMiss

	// Pipelines
	ch <- m.PipelineRuns
	ch <- m.PipelineRunStatus
	ch <- m.PipelineRunDurationSeconds
	ch <- m.PipelineRetryEvents
	ch <- m.PipelineFreshnessLagSeconds

	// SQL Warehouse
	ch <- m.Queries
	ch <- m.QueryDurationSeconds
	ch <- m.QueryErrors
	ch <- m.QueriesRunning

	// Health
	ch <- m.ExporterUp
	ch <- m.ScrapeStatus
	ch <- m.ExporterInfo
}
