# Metrics reference

The exporter collects metrics across four categories: billing, jobs, pipelines, and SQL queries. All workload metrics use the `_sliding` suffix to indicate they represent sliding window aggregations.

> **Note:** Most metrics are Gauges because they represent sliding window counts that can decrease as the window moves forward. See [Lookback Windows](../README.md#lookback-windows) for details.

## Quick reference

| Category | Metric | Labels | Description |
|----------|--------|--------|-------------|
| Billing | `databricks_billing_dbus_sliding` | `workspace_id`, `sku_name` | DBUs consumed (24h window) |
| Billing | `databricks_billing_cost_estimate_usd_sliding` | `workspace_id`, `sku_name` | Estimated cost in USD (24h window) |
| Billing | `databricks_price_change_events_sliding` | `sku_name` | Price changes per SKU (24h window) |
| Jobs | `databricks_job_runs_sliding` | `workspace_id`, `job_id`, `job_name` | Job runs count |
| Jobs | `databricks_job_run_status_sliding` | `workspace_id`, `job_id`, `job_name`, `status` | Job runs by status |
| Jobs | `databricks_job_run_duration_seconds_sliding` | `workspace_id`, `job_id`, `job_name`, `quantile` | Job duration quantiles |
| Jobs | `databricks_task_retries_sliding` | `workspace_id`, `job_id`, `job_name`, `task_key` | Task retry counts |
| Jobs | `databricks_job_sla_miss_sliding` | `workspace_id`, `job_id`, `job_name` | Jobs exceeding SLA threshold |
| Pipelines | `databricks_pipeline_runs_sliding` | `workspace_id`, `pipeline_id`, `pipeline_name` | Pipeline runs count |
| Pipelines | `databricks_pipeline_run_status_sliding` | `workspace_id`, `pipeline_id`, `pipeline_name`, `status` | Pipeline runs by status |
| Pipelines | `databricks_pipeline_run_duration_seconds_sliding` | `workspace_id`, `pipeline_id`, `pipeline_name`, `quantile` | Pipeline duration quantiles |
| Pipelines | `databricks_pipeline_retry_events_sliding` | `workspace_id`, `pipeline_id`, `pipeline_name` | Pipeline retry events |
| Pipelines | `databricks_pipeline_freshness_lag_seconds_sliding` | `workspace_id`, `pipeline_id`, `pipeline_name` | Data freshness lag |
| Queries | `databricks_queries_sliding` | `workspace_id`, `warehouse_id` | SQL queries executed |
| Queries | `databricks_query_errors_sliding` | `workspace_id`, `warehouse_id` | Failed SQL queries |
| Queries | `databricks_query_duration_seconds_sliding` | `workspace_id`, `warehouse_id`, `quantile` | Query duration quantiles |
| Queries | `databricks_queries_running_sliding` | `workspace_id`, `warehouse_id` | Concurrent queries estimate |
| Health | `databricks_exporter_up` | — | Exporter connectivity (1=up, 0=down) |
| Health | `databricks_scrape_status` | `query` | Per-query scrape status |
| Health | `databricks_exporter_info` | `version`, `*_window` | Build and config info |

All metrics also include standard Prometheus labels `job` and `instance` for scrape identification.

---

## Billing and cost metrics

These metrics help with FinOps and cost tracking. Data has 24-48h lag from actual usage.

### `databricks_billing_dbus_sliding`

Sliding window DBU consumption per workspace and SKU (default: last 24 hours).

- **Source table:** `system.billing.usage`
- **Type:** Gauge (sliding window count that can decrease as the window moves)
- **Labels:** `workspace_id`, `sku_name`

### `databricks_billing_cost_estimate_usd_sliding`

Estimated cost in USD calculated by joining usage with pricing data (sliding window, default: last 24 hours).

- **Source tables:** `system.billing.usage`, `system.billing.list_prices`
- **Type:** Gauge (sliding window value that can decrease as the window moves)
- **Labels:** `workspace_id`, `sku_name`

### `databricks_price_change_events_sliding`

Count of price changes per SKU within the billing lookback window (default: last 24 hours). Useful for attributing cost changes to pricing vs. usage increases.

- **Source table:** `system.billing.list_prices`
- **Type:** Gauge (sliding window count that can decrease as the window moves)
- **Labels:** `sku_name`

---

## Job metrics

These metrics track Databricks job executions (sliding window, default: last 4 hours).

### `databricks_job_runs_sliding`

Job runs per workspace and job within the lookback window.

- **Source table:** `system.lakeflow.job_run_timeline`
- **Type:** Gauge (sliding window count that can decrease as the window moves)
- **Labels:** `workspace_id`, `job_id`, `job_name`

### `databricks_job_run_status_sliding`

Job run counts broken down by result state.

- **Source table:** `system.lakeflow.job_run_timeline`
- **Type:** Gauge (sliding window count that can decrease as the window moves)
- **Labels:** `workspace_id`, `job_id`, `job_name`, `status`
- **Status values:** `SUCCEEDED`, `FAILED`, `CANCELED`, `TIMED_OUT`, etc.

### `databricks_job_run_duration_seconds_sliding`

Job run duration quantiles (p50, p95, p99).

- **Source table:** `system.lakeflow.job_run_timeline`
- **Type:** Gauge
- **Labels:** `workspace_id`, `job_id`, `job_name`, `quantile`
- **Quantile values:** `0.50`, `0.95`, `0.99`

### `databricks_task_retries_sliding`

Count of task retry attempts within the lookback window.

- **Source table:** `system.lakeflow.job_task_run_timeline`
- **Type:** Gauge (sliding window count that can decrease as the window moves)
- **Labels:** `workspace_id`, `job_id`, `job_name`, `task_key`
- **Note:** Disabled by default due to high cardinality. Enable with `--collect-task-retries`.

### `databricks_job_sla_miss_sliding`

Jobs that exceeded the SLA threshold (default: 1 hour) within the lookback window.

- **Source table:** `system.lakeflow.job_run_timeline`
- **Type:** Gauge (sliding window count that can decrease as the window moves)
- **Labels:** `workspace_id`, `job_id`, `job_name`

---

## Pipeline metrics

These metrics track Delta Live Tables (DLT) pipeline executions (sliding window, default: last 4 hours).

> **⚠️ Permissions Note:** Pipeline metrics require `SELECT` permission on `system.lakeflow.pipeline_update_timeline`. See [Troubleshooting](../README.md#pipeline-metrics-not-available-table_or_view_not_found).

### `databricks_pipeline_runs_sliding`

Pipeline update runs per workspace and pipeline within the lookback window.

- **Source table:** `system.lakeflow.pipeline_update_timeline`
- **Type:** Gauge (sliding window count that can decrease as the window moves)
- **Labels:** `workspace_id`, `pipeline_id`, `pipeline_name`

### `databricks_pipeline_run_status_sliding`

Pipeline run counts broken down by result state.

- **Source table:** `system.lakeflow.pipeline_update_timeline`
- **Type:** Gauge (sliding window count that can decrease as the window moves)
- **Labels:** `workspace_id`, `pipeline_id`, `pipeline_name`, `status`
- **Status values:** `COMPLETED`, `FAILED`, `CANCELED`, etc.

### `databricks_pipeline_run_duration_seconds_sliding`

Pipeline run duration quantiles (p50, p95, p99).

- **Source table:** `system.lakeflow.pipeline_update_timeline`
- **Type:** Gauge
- **Labels:** `workspace_id`, `pipeline_id`, `pipeline_name`, `quantile`
- **Quantile values:** `0.50`, `0.95`, `0.99`

### `databricks_pipeline_retry_events_sliding`

Pipeline retry events within the lookback window.

- **Source table:** `system.lakeflow.pipeline_update_timeline`
- **Type:** Gauge (sliding window count that can decrease as the window moves)
- **Labels:** `workspace_id`, `pipeline_id`, `pipeline_name`

### `databricks_pipeline_freshness_lag_seconds_sliding`

Average time lag between pipeline completion and current time.

- **Source table:** `system.lakeflow.pipeline_update_timeline`
- **Type:** Gauge
- **Labels:** `workspace_id`, `pipeline_id`, `pipeline_name`

---

## SQL query metrics

These metrics track SQL query performance across warehouses and serverless compute (sliding window, default: last 2 hours).

### `databricks_queries_sliding`

SQL queries executed per workspace and warehouse within the lookback window.

- **Source table:** `system.query.history`
- **Type:** Gauge (sliding window count that can decrease as the window moves)
- **Labels:** `workspace_id`, `warehouse_id`

### `databricks_query_errors_sliding`

Failed queries per workspace and warehouse within the lookback window.

- **Source table:** `system.query.history`
- **Type:** Gauge (sliding window count that can decrease as the window moves)
- **Labels:** `workspace_id`, `warehouse_id`

### `databricks_query_duration_seconds_sliding`

Query duration quantiles (p50, p95, p99) in seconds.

- **Source table:** `system.query.history`
- **Type:** Gauge
- **Labels:** `workspace_id`, `warehouse_id`, `quantile`
- **Quantile values:** `0.50`, `0.95`, `0.99`

### `databricks_queries_running_sliding`

Estimated count of concurrent queries (derived from overlapping execution intervals).

- **Source table:** `system.query.history`
- **Type:** Gauge
- **Labels:** `workspace_id`, `warehouse_id`

---

## System and health metrics

These metrics are **not** sliding window metrics — they reflect point-in-time exporter state.

### `databricks_exporter_up`

Indicates whether the exporter successfully connected to Databricks.

- **Type:** Gauge
- **Values:**
  - `1` - Exporter successfully established database connection
  - `0` - Exporter failed to establish database connection
- **Note:** This metric indicates exporter health, not Databricks availability. Individual query failures are logged but do not affect this metric.

### `databricks_scrape_status`

Status of individual scrape queries. Provides granular visibility into which system table queries succeeded or failed during each scrape.

- **Type:** Gauge
- **Labels:** `query` (e.g., `billing`, `jobs`, `pipelines`, `queries`)
- **Values:**
  - `1` - Query completed successfully
  - `0` - Query failed (timeout, error, or table unavailable)

### `databricks_exporter_info`

Build and configuration information for the exporter. Useful for tracking deployed versions and configured lookback windows across instances.

- **Type:** Gauge (always 1)
- **Labels:** `version`, `billing_window`, `jobs_window`, `pipelines_window`, `queries_window`

### `databricks_billing_scrape_errors`

Count of errors encountered during billing data collection.

- **Type:** Counter
- **Labels:** `workspace_id`

