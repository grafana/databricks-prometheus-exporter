# Databricks Mixin

A set of Grafana dashboards and Prometheus alerts for monitoring Databricks.

The mixin follows the new monitoring [mixin design pattern](https://monitoring.mixins.dev/), using [jsonnet](https://jsonnet.org/) and [grafonnet](https://github.com/grafana/grafonnet) to generate dashboards and alerts.

## Overview

This mixin provides comprehensive monitoring for Databricks with three main personas in mind:

1. **FinOps Persona** - Cost & Billing monitoring
2. **SRE/Platform Persona** - Jobs & Pipelines reliability
3. **Analytics/BI Persona** - SQL Warehouse performance

## Dashboards

The mixin includes three main dashboards:

### 1. Databricks Overview
Executive summary dashboard showing:
- Yesterday's cost and DoD delta
- DBUs consumption
- Global reliability metrics (jobs, pipelines, SQL)
- Cost decomposition by SKU and workspace
- Resource usage heatmap

### 2. Databricks Workloads & SQL
Deep dive into jobs and pipelines:
- Job/Pipeline runs and success rates
- Duration percentiles (p50, p95)
- Failure rates by workspace
- Retries vs failures
- Duration regression analysis
- Status breakdowns

### 3. Databricks SQL/BI Deep Dive
SQL warehouse performance and monitoring:
- Query load (1h, 24h)
- Error rates
- Query latency (p50, p95)
- Concurrency metrics
- Top slow queries
- Top erroring workspaces
- DoD changes analysis

## Metrics

The mixin expects the following metrics from the Databricks Prometheus exporter:

### Billing & Cost
- `databricks_billing_dbus_total` - Daily DBUs consumed
- `databricks_billing_cost_estimate_usd` - Daily cost estimates
- `databricks_price_change_events` - SKU price changes

### Jobs
- `databricks_job_runs_total` - Total job runs
- `databricks_job_run_duration_seconds` - Job duration (summary)
- `databricks_job_run_status_total` - Job status counts
- `databricks_task_retries_total` - Task retry counts

### Pipelines
- `databricks_pipeline_runs_total` - Total pipeline runs
- `databricks_pipeline_run_duration_seconds` - Pipeline duration (summary)
- `databricks_pipeline_run_status_total` - Pipeline status counts
- `databricks_pipeline_retry_events_total` - Pipeline retry events
- `databricks_pipeline_freshness_lag_seconds` - Data freshness lag

### SQL Queries
- `databricks_queries_total` - Total queries executed
- `databricks_query_duration_seconds` - Query duration (summary)
- `databricks_query_errors_total` - Failed queries
- `databricks_queries_running` - Concurrent queries

## Alerts

The mixin includes 16 alerts across three personas:

### Finance Persona (4 alerts)
- Spend spike warnings (25% DoD, 50% DoD)
- No billing data warnings (2h, 4h)

### SRE/Platform Persona (8 alerts)
- Job failure rate (10%, 20%)
- Pipeline failure rate (10%, 20%)
- Job duration regression (30%, 60% vs 7-day median)
- Pipeline duration regression (30%, 60% vs 7-day median)

### Analytics/BI Persona (4 alerts)
- SQL query error rate (5%, 10%)
- SQL query latency regression (30%, 60% vs 7-day median)

## Installation

1. Install dependencies:
```bash
cd mixin
jb install
```

2. Build dashboards:
```bash
make build
```

This will generate dashboard JSON files in the `dashboards_out/` directory.

3. Generate Prometheus alerts:
```bash
make prometheus_alerts.yaml
```

## Configuration

The mixin can be configured via `config.libsonnet`. Key configuration options include:

- `filteringSelector` - Prometheus label selector for filtering metrics
- `dashboardRefresh` - Dashboard refresh interval (default: 30m)
- `dashboardPeriod` - Default time range (default: now-7d)
- Alert thresholds for all 16 alerts

## Development

To modify the mixin:

1. Edit the appropriate files:
   - `signals/` - Metric definitions
   - `panels.libsonnet` - Panel definitions
   - `rows.libsonnet` - Row layouts
   - `dashboards.libsonnet` - Dashboard definitions
   - `alerts.libsonnet` - Alert rules

2. Format code:
```bash
make fmt
```

3. Rebuild:
```bash
make build
```

## References

- [Design Document](https://docs.google.com/document/d/1xOIaA4lS-XKW30C8CldoMbP65CampAhDQnwd3shvBjg/edit)
- [Databricks System Tables](https://docs.databricks.com/admin/system-tables/)
- [Monitoring Mixins](https://monitoring.mixins.dev/)
