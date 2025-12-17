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

		BillingDBUsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "billing", "dbus_total"),
			"Databricks Units (DBUs) consumed per workspace and SKU (sliding window, configurable via --billing-lookback, default: 24h).",
			[]string{labelWorkspaceID, labelSKUName},
			nil,
		),

		BillingCostEstimateUSD: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "billing", "cost_estimate_usd"),
			"List-price cost estimate (DBUs × list price) per workspace and SKU (sliding window, configurable via --billing-lookback, default: 24h).",
			[]string{labelWorkspaceID, labelSKUName},
			nil,
		),

		PriceChangeEvents: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "price_change_events"),
			"Pricing changes for a SKU (sliding window, configurable via --billing-lookback, default: 24h).",
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
			"Lakeflow Jobs runs per workspace and job (sliding window, configurable via --jobs-lookback, default: 2h).",
			[]string{labelWorkspaceID, labelJobID, labelJobName},
			nil,
		),

		JobRunStatusTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "job_run_status_total"),
			"Job status counts (SUCCEEDED/FAILED/CANCELED) per workspace and job (sliding window, configurable via --jobs-lookback, default: 2h).",
			[]string{labelWorkspaceID, labelJobID, labelJobName, labelStatus},
			nil,
		),

		JobRunDurationSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "job_run_duration_seconds"),
			"Job run duration quantiles (p50/p95/p99) per workspace and job (sliding window, configurable via --jobs-lookback, default: 2h).",
			[]string{labelWorkspaceID, labelJobID, labelJobName, labelQuantile},
			nil,
		),

		TaskRetriesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "task_retries_total"),
			"Retries across job tasks per workspace, job, and task key (sliding window, configurable via --jobs-lookback, default: 2h).",
			[]string{labelWorkspaceID, labelJobID, labelJobName, labelTaskKey},
			nil,
		),

		JobSLAMissTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "job_sla_miss_total"),
			"Job runs exceeding SLA threshold (configurable via --sla-threshold) per workspace and job (sliding window, configurable via --jobs-lookback, default: 2h).",
			[]string{labelWorkspaceID, labelJobID, labelJobName},
			nil,
		),

		// ===== Pipelines Metrics (SRE/Platform) =====

		PipelineRunsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pipeline_runs_total"),
			"DLT / Lakeflow Pipelines executions per workspace and pipeline (sliding window, configurable via --pipelines-lookback, default: 2h).",
			[]string{labelWorkspaceID, labelPipelineID, labelPipelineName},
			nil,
		),

		PipelineRunStatusTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pipeline_run_status_total"),
			"Pipeline run status counts (COMPLETED/FAILED) per workspace and pipeline (sliding window, configurable via --pipelines-lookback, default: 2h).",
			[]string{labelWorkspaceID, labelPipelineID, labelPipelineName, labelStatus},
			nil,
		),

		PipelineRunDurationSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pipeline_run_duration_seconds"),
			"Pipeline run duration quantiles (p50/p95/p99) per workspace and pipeline (sliding window, configurable via --pipelines-lookback, default: 2h).",
			[]string{labelWorkspaceID, labelPipelineID, labelPipelineName, labelQuantile},
			nil,
		),

		PipelineRetryEventsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pipeline_retry_events_total"),
			"Retry/backoff events within pipeline updates per workspace and pipeline (sliding window, configurable via --pipelines-lookback, default: 2h).",
			[]string{labelWorkspaceID, labelPipelineID, labelPipelineName},
			nil,
		),

		PipelineFreshnessLagSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pipeline_freshness_lag_seconds"),
			"Data freshness lag vs target watermark per workspace and pipeline (point-in-time, derived from latest pipeline runs within lookback window).",
			[]string{labelWorkspaceID, labelPipelineID, labelPipelineName},
			nil,
		),

		// ===== SQL Warehouse Metrics (Analytics/BI) =====

		QueriesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "queries_total"),
			"SQL queries executed (warehouse & serverless) per workspace and warehouse (sliding window, configurable via --queries-lookback, default: 1h).",
			[]string{labelWorkspaceID, labelWarehouseID},
			nil,
		),

		QueryDurationSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "query_duration_seconds"),
			"Query latency quantiles (p50/p95/p99) per workspace and warehouse (sliding window, configurable via --queries-lookback, default: 1h).",
			[]string{labelWorkspaceID, labelWarehouseID, labelQuantile},
			nil,
		),

		QueryErrorsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "query_errors_total"),
			"Failed queries per workspace and warehouse (sliding window, configurable via --queries-lookback, default: 1h).",
			[]string{labelWorkspaceID, labelWarehouseID},
			nil,
		),

		QueriesRunning: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "queries_running"),
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
			[]string{labelQuery, labelStatus},
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
	ch <- m.ExporterUp
	ch <- m.ScrapeStatus
	ch <- m.ExporterInfo
}
