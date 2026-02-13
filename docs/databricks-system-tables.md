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

For the complete list of metrics exported from these system tables, see the **[Metrics Reference](metrics-reference.md)**.
