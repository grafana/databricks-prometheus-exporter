# databricks-prometheus-exporter

Exports [Databricks](https://databricks.com) metrics via HTTP for Prometheus consumption.

## Overview

This exporter connects to a Databricks SQL Warehouse and queries Databricks System Tables to collect metrics about billing, job runs, pipeline executions, and SQL query performance. The metrics are exposed in Prometheus format, making it easy to monitor and analyze your Databricks workloads.

## Configuration

### Command line flags

The exporter may be configured through its command line flags:

```
  -h, --help                              Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address=:9976 ...      Addresses on which to expose metrics and web interface. Repeatable for multiple addresses.
      --web.config.file=""                Path to configuration file that can enable TLS or authentication.
      --web.telemetry-path="/metrics"     Path under which to expose metrics.
      --server-hostname=HOSTNAME          The Databricks workspace hostname (e.g., dbc-abc123-def456.cloud.databricks.com).
      --warehouse-http-path=PATH          The HTTP path of the SQL Warehouse (e.g., /sql/1.0/warehouses/abc123def456).
      --client-id=CLIENT-ID               The OAuth2 Client ID (Application ID) for Service Principal authentication.
      --client-secret=CLIENT-SECRET       The OAuth2 Client Secret for Service Principal authentication.
      --version                           Show application version.
      --log.level=info                    Only log messages with the given severity or above. One of: [debug, info, warn, error]
      --log.format=logfmt                 Output format of log messages. One of: [logfmt, json]
```

Example usage:

```sh
./databricks-exporter \
  --server-hostname=<your-host-name>.cloud.databricks.com \
  --warehouse-http-path=/sql/1.0/warehouses/abc123def456 \
  --client-id=<your-client-or-app-ID> \
  --client-secret=<your-client-secret-here>
```

### Environment Variables

Alternatively, the exporter may be configured using environment variables:

| Name                                           | Description                                                                                      |
| ---------------------------------------------- | ------------------------------------------------------------------------------------------------ |
| DATABRICKS_EXPORTER_SERVER_HOSTNAME            | The Databricks workspace hostname (e.g., dbc-abc123-def456.cloud.databricks.com).              |
| DATABRICKS_EXPORTER_WAREHOUSE_HTTP_PATH        | The HTTP path of the SQL Warehouse (e.g., /sql/1.0/warehouses/abc123def456).                   |
| DATABRICKS_EXPORTER_CLIENT_ID                  | The OAuth2 Client ID (Application ID) for Service Principal authentication.                     |
| DATABRICKS_EXPORTER_CLIENT_SECRET              | The OAuth2 Client Secret for Service Principal authentication.                                  |
| DATABRICKS_EXPORTER_WEB_TELEMETRY_PATH         | Path under which to expose metrics (default: /metrics).                                         |

Example usage:

```sh
export DATABRICKS_EXPORTER_SERVER_HOSTNAME="dbc-abc123-def456.cloud.databricks.com"
export DATABRICKS_EXPORTER_WAREHOUSE_HTTP_PATH="/sql/1.0/warehouses/abc123def456"
export DATABRICKS_EXPORTER_CLIENT_ID="4a8adace-cdf5-4489-b9c2-2b6f9dd7682f"
export DATABRICKS_EXPORTER_CLIENT_SECRET="your-client-secret-here"

./databricks-exporter
```

## Authentication

### Service Principal OAuth2 Authentication

The exporter uses OAuth2 Machine-to-Machine (M2M) authentication with Databricks Service Principals, following Databricks' recommended security practices.

#### Prerequisites

1. **Unity Catalog Enabled**: Your Databricks workspace must have Unity Catalog enabled to access System Tables.

2. **Service Principal**: Create a Service Principal in your Databricks workspace with appropriate permissions.

3. **SQL Warehouse**: Have a running SQL Warehouse (or one configured to auto-start) for executing queries.

#### Setting up a Service Principal

1. Log into your Databricks workspace
2. Go to **Settings** → **Admin Console** → **Service Principals**
3. Click **Add Service Principal**
4. Note the **Application ID** (this is your Client ID)
5. Click **Generate Secret** under OAuth Secrets
6. Copy and securely store the **Client Secret** (you won't see it again!)

#### Required Permissions

The Service Principal needs access to the Databricks workspace, permission to use the SQL Warehouse, and appropriate Unity Catalog permissions on System Tables.

Run these SQL commands as a Databricks admin to grant the necessary permissions (replace `<service-principal-id>` with your Service Principal's Application ID):

```sql
GRANT MANAGE ON CATALOG system TO `<service-principal-id>`;

GRANT USE CATALOG ON CATALOG system TO `<service-principal-id>`;

GRANT USE SCHEMA ON SCHEMA system.billing TO `<service-principal-id>`;

GRANT SELECT ON SCHEMA system.billing TO `<service-principal-id>`;

GRANT USE SCHEMA ON SCHEMA system.query TO `<service-principal-id>`;

GRANT SELECT ON SCHEMA system.query TO `<service-principal-id>`;

GRANT USE SCHEMA ON SCHEMA system.lakeflow TO `<service-principal-id>`;

GRANT SELECT ON SCHEMA system.lakeflow TO `<service-principal-id>`;

GRANT SELECT ON TABLE system.lakeflow.pipeline_update_timeline TO `<service-principal-id>`;
```

These grants provide:
- Management and usage rights on the `system` catalog
- Schema-level `USE` and `SELECT` permissions on `system.billing`, `system.query`, and `system.lakeflow`
- Access to all tables within these schemas including:
  - `system.billing.usage`, `system.billing.list_prices`
  - `system.lakeflow.job_run_timeline`, `system.lakeflow.job_task_run_timeline`, `system.lakeflow.pipeline_update_timeline`
  - `system.query.history`

### Getting Required Configuration Values

#### Server Hostname
- Found in your Databricks workspace URL
- Example: `dbc-abc123-def456.cloud.databricks.com`
- Remove the `https://` prefix

#### Warehouse HTTP Path
1. Go to **SQL Warehouses** in your Databricks workspace
2. Select your SQL Warehouse
3. Click **Connection Details** tab
4. Copy the **HTTP Path** (format: `/sql/1.0/warehouses/<warehouse-id>`)

#### Client ID and Client Secret
1. Go to **Settings** → **Admin Console** → **Service Principals**
2. Find your Service Principal
3. The **Application ID** is your **Client ID**
4. Generate and copy the **Client Secret**

## System Tables

The exporter queries several Databricks System Tables to collect metrics. These tables contain operational metadata about your Databricks workloads.

### `system.billing.usage`

Contains usage records for all Databricks services. Each row represents a usage event with details about consumption, timestamps, and associated metadata.

**Key columns:**
- `workspace_id` - ID of the workspace where usage occurred
- `sku_name` - The service or product being consumed
- `usage_quantity` - Amount of usage (typically in DBUs)
- `usage_date` - Date of the usage record
- `cloud` - Cloud provider (AWS, Azure, or GCP)
- `usage_metadata` - Structured metadata including cluster IDs, job IDs, warehouse IDs, etc.

### `system.billing.list_prices`

Contains pricing information for Databricks SKUs. Tracks price changes over time with effective date ranges.

**Key columns:**
- `sku_name` - The service or product name
- `cloud` - Cloud provider
- `pricing` - Structured pricing data (default, promotional, effective_list)
- `price_start_time` - When this price became effective
- `price_end_time` - When this price stopped being effective (NULL if current)
- `currency_code` - Currency for the price

### `system.lakeflow.job_run_timeline`

Tracks Databricks job executions at the run level. Each row represents a time period within a job run.

**Key columns:**
- `workspace_id` - ID of the workspace
- `job_id` - ID of the job definition
- `run_id` - ID of this specific run
- `result_state` - Outcome (SUCCEEDED, FAILED, CANCELED, etc.)
- `period_start_time` - Start time for this period
- `period_end_time` - End time for this period
- `run_type` - Type of run (JOB_RUN, WORKFLOW_RUN, etc.)

### `system.lakeflow.job_task_run_timeline`

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

### `system.lakeflow.pipeline_update_timeline`

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

### `system.query.history`

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

### Billing and Cost Metrics

These metrics help with FinOps and cost tracking.

#### `databricks_billing_dbus_total`
Daily DBU consumption per workspace and SKU over the last 30 days.

- **Source table:** `system.billing.usage`
- **Type:** Counter
- **Labels:** `workspace_id`, `sku_name`, `usage_date`

#### `databricks_billing_cost_estimate_usd`
Estimated cost in USD calculated by joining usage with pricing data.

- **Source tables:** `system.billing.usage`, `system.billing.list_prices`
- **Type:** Gauge
- **Labels:** `workspace_id`, `sku_name`, `usage_date`

#### `databricks_price_change_events`
Count of price changes per SKU over the last 90 days. Useful for attributing cost changes to pricing vs. usage increases.

- **Source table:** `system.billing.list_prices`
- **Type:** Counter
- **Labels:** `sku_name`

### Job Metrics

These metrics track Databricks job executions.

#### `databricks_job_runs_total`
Total number of job runs in the last 24 hours.

- **Source table:** `system.lakeflow.job_run_timeline`
- **Type:** Counter
- **Labels:** `workspace_id`

#### `databricks_job_run_status`
Job run counts broken down by result state.

- **Source table:** `system.lakeflow.job_run_timeline`
- **Type:** Gauge
- **Labels:** `workspace_id`, `status`

#### `databricks_job_run_duration_seconds`
Job run duration quantiles (p50, p95, p99).

- **Source table:** `system.lakeflow.job_run_timeline`
- **Type:** Gauge
- **Labels:** `workspace_id`, `quantile`

#### `databricks_task_retries_total`
Count of task retry attempts. Detected by finding the same task_key running multiple times within a job_run_id.

- **Source table:** `system.lakeflow.job_task_run_timeline`
- **Type:** Counter
- **Labels:** `workspace_id`

#### `databricks_job_sla_miss_total`
Number of jobs that exceeded 1 hour (3600 seconds) in the last 24 hours.

- **Source table:** `system.lakeflow.job_run_timeline`
- **Type:** Counter
- **Labels:** `workspace_id`

### Pipeline Metrics

These metrics track Delta Live Tables pipeline executions.

#### `databricks_pipeline_runs_total`
Total number of pipeline update runs in the last 24 hours.

- **Source table:** `system.lakeflow.pipeline_update_timeline`
- **Type:** Counter
- **Labels:** `workspace_id`

#### `databricks_pipeline_run_status`
Pipeline run counts broken down by result state.

- **Source table:** `system.lakeflow.pipeline_update_timeline`
- **Type:** Gauge
- **Labels:** `workspace_id`, `status`

#### `databricks_pipeline_run_duration_seconds`
Pipeline run duration quantiles (p50, p95, p99).

- **Source table:** `system.lakeflow.pipeline_update_timeline`
- **Type:** Gauge
- **Labels:** `workspace_id`, `quantile`

#### `databricks_pipeline_retry_events_total`
Count of pipeline retry events. Detected when multiple request_ids exist for the same update_id.

- **Source table:** `system.lakeflow.pipeline_update_timeline`
- **Type:** Counter
- **Labels:** `workspace_id`

#### `databricks_pipeline_freshness_lag_seconds`
Average time lag between pipeline completion and current time. Can be used to track data freshness.

- **Source table:** `system.lakeflow.pipeline_update_timeline`
- **Type:** Gauge
- **Labels:** `workspace_id`, `stage`

### SQL Query Metrics

These metrics track SQL query performance across warehouses and serverless compute.

#### `databricks_queries_total`
Total number of SQL queries executed in the last hour.

- **Source table:** `system.query.history`
- **Type:** Counter
- **Labels:** `workspace_id`

#### `databricks_query_errors_total`
Number of failed queries in the last hour.

- **Source table:** `system.query.history`
- **Type:** Counter
- **Labels:** `workspace_id`

#### `databricks_query_duration_seconds`
Query duration quantiles (p50, p95, p99) in seconds.

- **Source table:** `system.query.history`
- **Type:** Gauge
- **Labels:** `workspace_id`, `quantile`

#### `databricks_queries_running`
Estimated count of concurrent queries. Calculated by finding overlapping query execution intervals.

- **Source table:** `system.query.history`
- **Type:** Gauge
- **Labels:** `workspace_id`

### System Metrics

#### `databricks_up`
Indicates whether the exporter successfully connected to Databricks and collected metrics.

- **Type:** Gauge
- **Values:**
  - `1` - Connection successful, metrics collected
  - `0` - Connection failed or unable to collect metrics

## Building

```sh
go build -o databricks-exporter ./cmd/databricks-exporter
```

## Running with Docker

Build the Docker image:

```sh
docker build -t databricks-exporter .
```

Run the container:

```sh
docker run -p 9976:9976 \
  -e DATABRICKS_EXPORTER_SERVER_HOSTNAME="dbc-abc123-def456.cloud.databricks.com" \
  -e DATABRICKS_EXPORTER_WAREHOUSE_HTTP_PATH="/sql/1.0/warehouses/abc123def456" \
  -e DATABRICKS_EXPORTER_CLIENT_ID="your-client-id" \
  -e DATABRICKS_EXPORTER_CLIENT_SECRET="your-client-secret" \
  databricks-exporter
```

## Prometheus Configuration

Add the exporter as a scrape target in your Prometheus configuration:

```yaml
scrape_configs:
  - job_name: 'databricks'
    static_configs:
      - targets: ['localhost:9976']
```

## Troubleshooting

### Authentication Errors (401 Unauthorized)

If you see authentication errors:
- Verify your Client ID (Application ID) is correct
- Ensure the Client Secret hasn't expired
- Check that the Service Principal has access to the workspace

### Credit Exhaustion (400 Bad Request)

Error: `Sorry, cannot run the resource because you've exhausted your available credits`

This means your Databricks account has run out of credits. Add a payment method or credits to your account to continue.

### Connection Errors

- Ensure the SQL Warehouse is running (or set to auto-start)
- Verify the Warehouse HTTP Path is correct (should start with `/sql/1.0/warehouses/`)
- Check that the Server Hostname doesn't include `https://`

### No Metrics Appearing

If `databricks_up` is 1 but some metrics don't appear:
- Some metrics only appear when there's data to report (e.g., no retries means no retry metric)
- Billing metrics require recent usage data (check the last 30 days)
- Job and pipeline metrics require recent executions (check the last 24 hours)
- Query metrics require recent SQL activity (check the last hour)
- Verify Unity Catalog is enabled on your workspace
- Ensure the Service Principal has permissions to read all System Tables

### Pipeline Metrics Not Available (TABLE_OR_VIEW_NOT_FOUND)

If you see errors like:
```
TABLE_OR_VIEW_NOT_FOUND: The table or view `system`.`lakeflow`.`pipeline_update_timeline` cannot be found
```

**Cause:** The `system.lakeflow.pipeline_update_timeline` table exists in Databricks, but the Service Principal likely doesn't have `SELECT` permission on it.

**Impact:** Pipeline metrics (`databricks_pipeline_*`) will not be collected, but all other metrics (jobs, billing, queries) will continue to work normally. The exporter will check table availability periodically and automatically resume collection if permissions are granted.

**Verification:** Run this query in your Databricks SQL Warehouse to check if the table exists:
```sql
SELECT COUNT(*) FROM system.lakeflow.pipeline_update_timeline LIMIT 1;
```

If you get a permissions error instead of "table not found", this confirms it's a permissions issue.

**Solutions:**
1. **Grant Permissions** - Ensure all required permissions are granted as described in the [Required Permissions](#required-permissions) section above
2. **Verify Schema Access** - Confirm the Service Principal has `USE SCHEMA` and `SELECT` on `system.lakeflow`
3. **Automatic Recovery** - Once permissions are granted, the exporter will automatically detect the table is available and resume collection within ~10 scrapes (typically ~10 minutes)

The exporter now handles this gracefully:
- Checks table availability at startup and periodically
- Logs a clear warning once (not every scrape)
- Automatically resumes collection when permissions are fixed
- Continues collecting all other metrics normally

For more information, see the [Known Limitations section in the mixin README](mixin/README.md#known-limitations).

### Debug Logging

Enable debug logging for more detailed information:

```sh
./databricks-exporter --log.level=debug [other flags...]
```

## License

Licensed under the Apache License, Version 2.0. See LICENSE for details.
