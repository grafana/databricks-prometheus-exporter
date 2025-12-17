# Databricks Mixin

A set of Grafana dashboards and prometheus-compatible alerts for monitoring Databricks.

The mixin follows the new monitoring [mixin design pattern](https://monitoring.mixins.dev/), using [jsonnet](https://jsonnet.org/) and [grafonnet](https://github.com/grafana/grafonnet) to generate dashboards and alerts.

## Overview

This mixin provides comprehensive monitoring for Databricks with three main personas in mind:

1. **FinOps Persona** - Cost & Billing monitoring
2. **SRE/Platform Persona** - Jobs & Pipelines reliability
3. **Analytics/BI Persona** - SQL Warehouse performance

### Key features

- **Detailed drill-down**: All metrics include detailed labels (`job_id`, `job_name`, `pipeline_id`, `pipeline_name`, `task_key`, `warehouse_id`) enabling deep analysis of specific workloads
- **Multi-level views**: From high-level overview to detailed per-job/pipeline/warehouse breakdowns
- **Metrics**: Signals covering billing, jobs, pipelines, SQL queries and SQL warehouses
- **Alerts**: Tiered warning/critical alerts across all three personas
- **Sparse data handling**: Queries optimized for infrequent updates and sliding window metrics
- **Grafana variables**: Pre-configured filtering by job, workspace, and instance

## Dashboards

The mixin includes three main dashboards:

### 1. Databricks overview
Executive summary dashboard showing:
- Total cost (30-day) and 24h growth %
- Total DBUs (30-day) consumption
- Global reliability metrics (jobs, pipelines, SQL success/error rates)
- Cost decomposition by SKU and workspace (tables)
- Failure trends over time

### 2. Databricks jobs and pipelines
Deep dive into jobs and pipelines with drill-down capabilities:
- **Overview**: Runs, success rates, p95 duration stats
- **Throughput & Duration**: Total runs and p95 duration time series for jobs and pipelines
- **Reliability & Stability**: Failure rates and retries vs failures
- **Status Breakdowns**: Jobs and pipelines status over time
- **Jobs Drill-down**: Top jobs by runs/failures/duration, task retries, and time series by job name
- **Pipelines Drill-down**: Top pipelines by runs/failures/duration, freshness lag, and time series by pipeline name

### 3. Databricks warehouses and queries
SQL warehouse performance and monitoring:
- **Overview**: Query totals (1h/24h), error rate, latency, concurrency metrics
- **Time Series**: Query load, latency (p50/p95), error rates, concurrency trends
- **Analysis**: Query volume by workspace
- **Top Warehouses**: Tables showing top warehouses by query count, errors, and latency
- **Warehouse Drill-down**: Queries, errors, latency, and concurrency by warehouse ID

## Metrics

The mixin expects the following metrics from the Databricks Prometheus exporter. Note: Most metrics are Gauges because they represent sliding window counts that can decrease as the window moves forward.

### Billing and cost
- `databricks_billing_dbus_total{workspace_id, sku_name}` - DBUs consumed (sliding window, default 24h)
- `databricks_billing_cost_estimate_usd{workspace_id, sku_name}` - Cost estimates (sliding window, default 24h)

### Jobs
- `databricks_job_runs_total{workspace_id, job_id, job_name}` - Job runs (sliding window, default 2h)
- `databricks_job_run_duration_seconds{workspace_id, job_id, job_name, quantile}` - Job duration quantiles (p50, p95, p99)
- `databricks_job_run_status_total{workspace_id, job_id, job_name, status}` - Job status counts
- `databricks_task_retries_total{workspace_id, job_id, job_name, task_key}` - Task retry counts

### Pipelines
- `databricks_pipeline_runs_total{workspace_id, pipeline_id, pipeline_name}` - Pipeline runs (sliding window, default 2h)
- `databricks_pipeline_run_duration_seconds{workspace_id, pipeline_id, pipeline_name, quantile}` - Pipeline duration quantiles (p50, p95, p99)
- `databricks_pipeline_run_status_total{workspace_id, pipeline_id, pipeline_name, status}` - Pipeline status counts
- `databricks_pipeline_retry_events_total{workspace_id, pipeline_id, pipeline_name}` - Pipeline retry events
- `databricks_pipeline_freshness_lag_seconds{workspace_id, pipeline_id, pipeline_name}` - Data freshness lag

### SQL queries and warehouses
- `databricks_queries_total{workspace_id, warehouse_id}` - Queries executed (sliding window, default 1h)
- `databricks_query_duration_seconds{workspace_id, warehouse_id, quantile}` - Query duration quantiles (p50, p95, p99)
- `databricks_query_errors_total{workspace_id, warehouse_id}` - Failed queries
- `databricks_queries_running{workspace_id, warehouse_id}` - Concurrent queries

### System and health
- `databricks_exporter_up` - Exporter connectivity (1=healthy, 0=failed)
- `databricks_scrape_status{query, status}` - Per-query scrape status (1=success, 0=failure)
- `databricks_exporter_info{version, billing_window, jobs_window, pipelines_window, queries_window}` - Build and configuration info

All metrics include standard Prometheus labels `job` and `instance` for scrape identification.

## Alerts

### FinOps persona alerts
- `DatabricksWarnSpendSpike` - 25% DoD cost increase
- `DatabricksCriticalSpendSpike` - 50% DoD cost increase
- `DatabricksWarnNoBillingData` - No billing data for 2 hours
- `DatabricksCriticalNoBillingData` - No billing data for 4 hours

### SRE/platform persona alerts
- `DatabricksWarnJobFailureRate` - Job failure rate > 10%
- `DatabricksCriticalJobFailureRate` - Job failure rate > 20%
- `DatabricksWarnJobDurationRegression` - Job duration 30% above 7-day median
- `DatabricksCriticalJobDurationRegression` - Job duration 60% above 7-day median
- `DatabricksWarnPipelineFailureRate` - Pipeline failure rate > 10%
- `DatabricksCriticalPipelineFailureRate` - Pipeline failure rate > 20%
- `DatabricksWarnPipelineDurationRegression` - Pipeline duration 30% above 7-day median
- `DatabricksCritPipelineDurationHigh` - Pipeline duration 60% above 7-day median

### Analytics/BI persona alerts
- `DatabricksWarnSqlQueryErrorRate` - SQL error rate > 5%
- `DatabricksCriticalSqlQueryErrorRate` - SQL error rate > 10%
- `DatabricksWarnSqlQueryLatencyRegression` - Query latency 30% above 7-day median
- `DatabricksCritQueryLatencyHigh` - Query latency 60% above 7-day median

## Installation

### Prerequisites
- [jsonnet-bundler](https://github.com/jsonnet-bundler/jsonnet-bundler) (`jb`)
- [jsonnet](https://github.com/google/jsonnet)
- [mixtool](https://github.com/monitoring-mixins/mixtool)

### Steps

1. Install dependencies:
```bash
cd mixin
jb install
```

2. Build dashboards:
```bash
make dashboards_out
```

This will generate dashboard JSON files in the `dashboards_out/` directory:
- `databricks-overview.json`
- `databricks-jobs-and-pipelines.json`
- `databricks-warehouses-and-queries.json`

3. Generate Prometheus alerts:
```bash
make prometheus_alerts.yaml
```

This will generate a `prometheus_alerts.yaml` file containing all alert rules.

## Configuration

The mixin can be configured via `config.libsonnet`. Key configuration options include:

- `filteringSelector` - Prometheus label selector (default: `''`)
- `groupLabels` - Labels to group by (default: `['job', 'workspace_id']`)
- `dashboardRefresh` - Dashboard refresh interval (default: `1m`)
- `dashboardPeriod` - Default time range (default: `now-7d`)
- `dashboardTimezone` - Dashboard timezone (default: `utc`)
- Dashboard-specific settings for overview, jobs & pipelines, and warehouses & queries
- Alert thresholds and evaluation intervals

## Development

To modify the mixin:

1. Edit the appropriate files:
   - `signals/overview.libsonnet` - Overview dashboard metric definitions
   - `signals/jobs_and_pipelines.libsonnet` - Jobs & Pipelines metric definitions
   - `signals/warehouses_and_queries.libsonnet` - Warehouses & Queries metric definitions
   - `panels.libsonnet` - Panel definitions
   - `rows.libsonnet` - Row layouts
   - `dashboards.libsonnet` - Dashboard definitions
   - `alerts.libsonnet` - Alert rules
   - `config.libsonnet` - Configuration options

2. Format code:
```bash
make fmt
```

3. Lint:
```bash
make lint
```

4. Rebuild dashboards:
```bash
make dashboards_out
```

5. Generate Prometheus alerts:
```bash
make prometheus_alerts.yaml
```

## Known limitations

### Pipeline metrics require permissions on `system.lakeflow.pipeline_update_timeline`

**Affected Metrics:**
- `databricks_pipeline_runs_total`
- `databricks_pipeline_run_status_total`
- `databricks_pipeline_run_duration_seconds`
- `databricks_pipeline_retry_events_total`
- `databricks_pipeline_freshness_lag_seconds`

**Issue:** The `system.lakeflow.pipeline_update_timeline` table exists but the Service Principal may not have `SELECT` permission on it.

**Symptoms:** You'll see errors in the exporter logs (but only logged once, not repeatedly):
```
TABLE_OR_VIEW_NOT_FOUND: The table or view `system`.`lakeflow`.`pipeline_update_timeline` cannot be found
```

**Impact:** 
- Pipeline metrics will not be collected until permissions are granted
- Pipeline-related panels in the "Jobs & Pipelines" dashboard will show no data
- Pipeline-related alerts will not fire
- Other metrics (jobs, billing, queries) will continue to work normally
- **Automatic Recovery**: Once permissions are granted, the exporter automatically detects availability and resumes collection

**Root Cause:**
The error message says "not found" but it's actually a **permissions issue**. The table exists in Databricks, but without proper permissions, it appears as if it doesn't exist.

**Verification:**
Run this query in your Databricks SQL Warehouse:
```sql
SELECT COUNT(*) FROM system.lakeflow.pipeline_update_timeline LIMIT 1;
```

- If you get "TABLE_OR_VIEW_NOT_FOUND" → Permissions issue
- If you get a count → Permissions are correct

**Solutions:**
1. **Grant Permissions**: Ensure all required Unity Catalog permissions are granted as described in the main [README - Required Permissions](../README.md#required-permissions) section
2. **Verify Schema Access**: Confirm the Service Principal has `USE SCHEMA` and `SELECT` permissions on `system.lakeflow`
3. **Wait for Auto-Recovery**: The exporter checks table availability periodically (every ~10 scrapes). Once permissions are granted, collection resumes automatically - no restart needed!

## References

- [Databricks System Tables](https://docs.databricks.com/admin/system-tables/)
- [Monitoring Mixins](https://monitoring.mixins.dev/)
