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

The exporter collects 18 metrics across four categories: billing, jobs, pipelines, and SQL queries.

## Billing and Cost Metrics

These metrics help with FinOps and cost tracking.

### `databricks_billing_dbus_total`
Daily DBU consumption per workspace and SKU over the last 30 days.

- **Source table:** `system.billing.usage`
- **Type:** Counter
- **Labels:** `workspace_id`, `sku_name`, `usage_date`

### `databricks_billing_cost_estimate_usd`
Estimated cost in USD calculated by joining usage with pricing data.

- **Source tables:** `system.billing.usage`, `system.billing.list_prices`
- **Type:** Gauge
- **Labels:** `workspace_id`, `sku_name`, `usage_date`

### `databricks_price_change_events`
Count of price changes per SKU over the last 90 days. Useful for attributing cost changes to pricing vs. usage increases.

- **Source table:** `system.billing.list_prices`
- **Type:** Counter
- **Labels:** `sku_name`

## Job Metrics

These metrics track Databricks job executions.

### `databricks_job_runs_total`
Total number of job runs in the last 24 hours.

- **Source table:** `system.lakeflow.job_run_timeline`
- **Type:** Counter
- **Labels:** `workspace_id`

### `databricks_job_run_status`
Job run counts broken down by result state.

- **Source table:** `system.lakeflow.job_run_timeline`
- **Type:** Gauge
- **Labels:** `workspace_id`, `status`

### `databricks_job_run_duration_seconds`
Job run duration quantiles (p50, p95, p99).

- **Source table:** `system.lakeflow.job_run_timeline`
- **Type:** Gauge
- **Labels:** `workspace_id`, `quantile`

### `databricks_task_retries_total`
Count of task retry attempts. Detected by finding the same task_key running multiple times within a job_run_id.

- **Source table:** `system.lakeflow.job_task_run_timeline`
- **Type:** Counter
- **Labels:** `workspace_id`

### `databricks_job_sla_miss_total`
Number of jobs that exceeded 1 hour (3600 seconds) in the last 24 hours.

- **Source table:** `system.lakeflow.job_run_timeline`
- **Type:** Counter
- **Labels:** `workspace_id`

## Pipeline Metrics

These metrics track Delta Live Tables pipeline executions.

### `databricks_pipeline_runs_total`
Total number of pipeline update runs in the last 24 hours.

- **Source table:** `system.lakeflow.pipeline_update_timeline`
- **Type:** Counter
- **Labels:** `workspace_id`

### `databricks_pipeline_run_status`
Pipeline run counts broken down by result state.

- **Source table:** `system.lakeflow.pipeline_update_timeline`
- **Type:** Gauge
- **Labels:** `workspace_id`, `status`

### `databricks_pipeline_run_duration_seconds`
Pipeline run duration quantiles (p50, p95, p99).

- **Source table:** `system.lakeflow.pipeline_update_timeline`
- **Type:** Gauge
- **Labels:** `workspace_id`, `quantile`

### `databricks_pipeline_retry_events_total`
Count of pipeline retry events. Detected when multiple request_ids exist for the same update_id.

- **Source table:** `system.lakeflow.pipeline_update_timeline`
- **Type:** Counter
- **Labels:** `workspace_id`

### `databricks_pipeline_freshness_lag_seconds`
Average time lag between pipeline completion and current time. Can be used to track data freshness.

- **Source table:** `system.lakeflow.pipeline_update_timeline`
- **Type:** Gauge
- **Labels:** `workspace_id`, `stage`

## SQL Query Metrics

These metrics track SQL query performance across warehouses and serverless compute.

### `databricks_queries_total`
Total number of SQL queries executed in the last hour.

- **Source table:** `system.query.history`
- **Type:** Counter
- **Labels:** `workspace_id`

### `databricks_query_errors_total`
Number of failed queries in the last hour.

- **Source table:** `system.query.history`
- **Type:** Counter
- **Labels:** `workspace_id`

### `databricks_query_duration_seconds`
Query duration quantiles (p50, p95, p99) in seconds.

- **Source table:** `system.query.history`
- **Type:** Gauge
- **Labels:** `workspace_id`, `quantile`

### `databricks_queries_running`
Estimated count of concurrent queries. Calculated by finding overlapping query execution intervals.

- **Source table:** `system.query.history`
- **Type:** Gauge
- **Labels:** `workspace_id`

## System Metrics

### `databricks_up`
Indicates whether the exporter successfully connected to Databricks and collected metrics.

- **Type:** Gauge
- **Values:**
  - `1` - Connection successful, metrics collected
  - `0` - Connection failed or unable to collect metrics
