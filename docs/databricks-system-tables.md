# System tables reference

The exporter queries several Databricks System Tables to collect metrics. These tables contain operational metadata about your Databricks workloads.

## `system.billing.usage`

Contains usage records for all Databricks services. Each row represents a usage event with details about consumption, timestamps, and associated metadata.

**Key columns:**
- `workspace_id` - ID of the workspace where usage occurred
- `sku_name` - The service or product being consumed
- `usage_quantity` - Amount of usage (typically in DBUs)
- `usage_date` - Date of the usage record
- `cloud` - Cloud provider (AWS, Azure, or GCP)
- `usage_metadata` - Structured metadata including cluster IDs, job IDs, warehouse IDs, etc.

## `system.billing.list_prices`

Contains pricing information for Databricks SKUs. Tracks price changes over time with effective date ranges.

**Key columns:**
- `sku_name` - The service or product name
- `cloud` - Cloud provider
- `pricing` - Structured pricing data (default, promotional, effective_list)
- `price_start_time` - When this price became effective
- `price_end_time` - When this price stopped being effective (NULL if current)
- `currency_code` - Currency for the price

## `system.lakeflow.job_run_timeline`

Tracks Databricks job executions at the run level. Each row represents a time period within a job run.

**Key columns:**
- `workspace_id` - ID of the workspace
- `job_id` - ID of the job definition
- `run_id` - ID of this specific run
- `result_state` - Outcome (SUCCEEDED, FAILED, CANCELED, etc.)
- `period_start_time` - Start time for this period
- `period_end_time` - End time for this period
- `run_type` - Type of run (JOB_RUN, WORKFLOW_RUN, etc.)

## `system.lakeflow.job_task_run_timeline`

Tracks individual task executions within jobs. Tasks can retry independently, making this useful for tracking retry behavior.

**Key columns:**
- `workspace_id` - ID of the workspace
- `job_id` - ID of the parent job
- `job_run_id` - ID of the parent job run
- `run_id` - ID of this specific task run
- `task_key` - Identifier for the task within the job
- `result_state` - Outcome of the task
- `period_start_time` - Start time
- `period_end_time` - End time

## `system.lakeflow.pipeline_update_timeline`

Tracks Delta Live Tables (DLT) pipeline update executions. Records both successful and failed pipeline runs.

> **⚠️ Permissions Note:** This table exists but may require explicit `SELECT` permission for your Service Principal. If the Service Principal lacks permission, you'll see "TABLE_OR_VIEW_NOT_FOUND" errors (even though the table exists). Grant the Service Principal `SELECT` permission on this table, and the exporter will automatically detect it and resume collection. See [Troubleshooting](#pipeline-metrics-not-available-table_or_view_not_found) for details.

**Key columns:**
- `workspace_id` - ID of the workspace
- `pipeline_id` - ID of the pipeline
- `update_id` - ID of this update
- `request_id` - Request ID (multiple requests for same update indicate retries)
- `result_state` - Outcome of the update
- `period_start_time` - Start time
- `period_end_time` - End time
- `update_type` - Type of update (FULL_REFRESH, INCREMENTAL, etc.)

## `system.query.history`

Contains execution history for all SQL queries run in the workspace. Includes queries from SQL warehouses, notebooks, and serverless compute.

**Key columns:**
- `workspace_id` - ID of the workspace
- `statement_id` - Unique identifier for the statement execution
- `execution_status` - Status (FINISHED, FAILED, CANCELED)
- `start_time` - When execution started
- `end_time` - When execution finished
- `total_duration_ms` - Total execution time in milliseconds
- `error_message` - Error details if the query failed
- `query_source` - Structured data about what triggered the query

## Metrics

The exporter collects metrics across four categories: billing, jobs, pipelines, and SQL queries.

## Billing and cost metrics

These metrics help with FinOps and cost tracking.

### `databricks_billing_dbus_total`
Sliding window DBU consumption per workspace and SKU (default: last 24 hours).

- **Source table:** `system.billing.usage`
- **Type:** Gauge (sliding window count that can decrease as the window moves)
- **Labels:** `workspace_id`, `sku_name`

### `databricks_billing_cost_estimate_usd`
Estimated cost in USD calculated by joining usage with pricing data (sliding window, default: last 24 hours).

- **Source tables:** `system.billing.usage`, `system.billing.list_prices`
- **Type:** Gauge (sliding window value that can decrease as the window moves)
- **Labels:** `workspace_id`, `sku_name`

### `databricks_price_change_events`
Count of price changes per SKU within the billing lookback window (default: last 24 hours). Useful for attributing cost changes to pricing vs. usage increases.

- **Source table:** `system.billing.list_prices`
- **Type:** Gauge (sliding window count that can decrease as the window moves)
- **Labels:** `sku_name`

## Job metrics

These metrics track Databricks job executions (sliding window, default: last 2 hours).

### `databricks_job_runs_total`
Job runs per workspace and job within the lookback window.

- **Source table:** `system.lakeflow.job_run_timeline`
- **Type:** Gauge (sliding window count that can decrease as the window moves)
- **Labels:** `workspace_id`, `job_id`, `job_name`

### `databricks_job_run_status_total`
Job run counts broken down by result state.

- **Source table:** `system.lakeflow.job_run_timeline`
- **Type:** Gauge (sliding window count that can decrease as the window moves)
- **Labels:** `workspace_id`, `job_id`, `job_name`, `status`

### `databricks_job_run_duration_seconds`
Job run duration quantiles (p50, p95, p99).

- **Source table:** `system.lakeflow.job_run_timeline`
- **Type:** Gauge
- **Labels:** `workspace_id`, `job_id`, `job_name`, `quantile`

### `databricks_task_retries_total`
Count of task retry attempts within the lookback window.

- **Source table:** `system.lakeflow.job_task_run_timeline`
- **Type:** Gauge (sliding window count that can decrease as the window moves)
- **Labels:** `workspace_id`, `job_id`, `job_name`, `task_key`

### `databricks_job_sla_miss_total`
Jobs that exceeded the SLA threshold (default: 1 hour) within the lookback window.

- **Source table:** `system.lakeflow.job_run_timeline`
- **Type:** Gauge (sliding window count that can decrease as the window moves)
- **Labels:** `workspace_id`, `job_id`, `job_name`

## Pipeline metrics

These metrics track Delta Live Tables (DLT) pipeline executions (sliding window, default: last 2 hours).

### `databricks_pipeline_runs_total`
Pipeline update runs per workspace and pipeline within the lookback window.

- **Source table:** `system.lakeflow.pipeline_update_timeline`
- **Type:** Gauge (sliding window count that can decrease as the window moves)
- **Labels:** `workspace_id`, `pipeline_id`, `pipeline_name`

### `databricks_pipeline_run_status_total`
Pipeline run counts broken down by result state.

- **Source table:** `system.lakeflow.pipeline_update_timeline`
- **Type:** Gauge (sliding window count that can decrease as the window moves)
- **Labels:** `workspace_id`, `pipeline_id`, `pipeline_name`, `status`

### `databricks_pipeline_run_duration_seconds`
Pipeline run duration quantiles (p50, p95, p99).

- **Source table:** `system.lakeflow.pipeline_update_timeline`
- **Type:** Gauge
- **Labels:** `workspace_id`, `pipeline_id`, `pipeline_name`, `quantile`

### `databricks_pipeline_retry_events_total`
Pipeline retry events within the lookback window.

- **Source table:** `system.lakeflow.pipeline_update_timeline`
- **Type:** Gauge (sliding window count that can decrease as the window moves)
- **Labels:** `workspace_id`, `pipeline_id`, `pipeline_name`

### `databricks_pipeline_freshness_lag_seconds`
Average time lag between pipeline completion and current time.

- **Source table:** `system.lakeflow.pipeline_update_timeline`
- **Type:** Gauge
- **Labels:** `workspace_id`, `pipeline_id`, `pipeline_name`

## SQL query metrics

These metrics track SQL query performance across warehouses and serverless compute (sliding window, default: last 1 hour).

### `databricks_queries_total`
SQL queries executed per workspace and warehouse within the lookback window.

- **Source table:** `system.query.history`
- **Type:** Gauge (sliding window count that can decrease as the window moves)
- **Labels:** `workspace_id`, `warehouse_id`

### `databricks_query_errors_total`
Failed queries per workspace and warehouse within the lookback window.

- **Source table:** `system.query.history`
- **Type:** Gauge (sliding window count that can decrease as the window moves)
- **Labels:** `workspace_id`, `warehouse_id`

### `databricks_query_duration_seconds`
Query duration quantiles (p50, p95, p99) in seconds.

- **Source table:** `system.query.history`
- **Type:** Gauge
- **Labels:** `workspace_id`, `warehouse_id`, `quantile`

### `databricks_queries_running`
Estimated count of concurrent queries (derived from overlapping execution intervals).

- **Source table:** `system.query.history`
- **Type:** Gauge
- **Labels:** `workspace_id`, `warehouse_id`

## System and health metrics

### `databricks_exporter_up`
Indicates whether the exporter successfully connected to Databricks. Note: This metric indicates exporter health, not Databricks availability. Individual query failures are logged but do not affect this metric.

- **Type:** Gauge
- **Values:**
  - `1` - Exporter successfully established database connection
  - `0` - Exporter failed to establish database connection

### `databricks_scrape_status`
Status of individual scrape queries. Provides granular visibility into which system table queries succeeded or failed during each scrape.

- **Type:** Gauge
- **Labels:** `query` (e.g., `billing_dbus`, `job_runs`, `queries`), `status` (`success` or `error`)
- **Values:**
  - `1` - Query completed successfully
  - `0` - Query failed (timeout, error, or table unavailable)

### `databricks_exporter_info`
Build and configuration information for the exporter. Useful for tracking deployed versions and configured lookback windows across instances.

- **Type:** Gauge (always 1)
- **Labels:** `version`, `billing_window`, `jobs_window`, `pipelines_window`, `queries_window`
