# databricks-prometheus-exporter

Exports [Databricks](https://databricks.com) metrics via HTTP for Prometheus consumption.

## Overview

This exporter connects to a Databricks SQL Warehouse and queries Databricks System Tables to collect metrics about billing, job runs, pipeline executions, and SQL query performance. The metrics are exposed in Prometheus format, making it easy to monitor and analyze your Databricks workloads.

## Configuration

### Command-line flags

The exporter may be configured through its command line flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--web.listen-address` | `:9976` | Addresses on which to expose metrics and web interface. Repeatable for multiple addresses. |
| `--web.config.file` | `""` | Path to configuration file that can enable TLS or authentication. |
| `--web.telemetry-path` | `/metrics` | Path under which to expose metrics. |
| `--server-hostname` | *required* | The Databricks workspace hostname (e.g., `dbc-abc123.cloud.databricks.com`). |
| `--warehouse-http-path` | *required* | The HTTP path of the SQL Warehouse (e.g., `/sql/1.0/warehouses/abc123`). |
| `--client-id` | *required* | The OAuth2 Client ID (Application ID) for Service Principal authentication. |
| `--client-secret` | *required* | The OAuth2 Client Secret for Service Principal authentication. |
| `--query-timeout` | `5m` | Timeout for database queries. |
| `--billing-lookback` | `24h` | How far back to look for billing data. See [Lookback Windows](#lookback-windows). |
| `--jobs-lookback` | `4h` | How far back to look for job runs. See [Lookback Windows](#lookback-windows). |
| `--pipelines-lookback` | `4h` | How far back to look for pipeline runs. See [Lookback Windows](#lookback-windows). |
| `--queries-lookback` | `2h` | How far back to look for SQL warehouse queries. See [Lookback Windows](#lookback-windows). |
| `--sla-threshold` | `3600` | Duration threshold (in seconds) for job SLA miss detection. |
| `--collect-task-retries` | `false` | Collect task retry metrics (high cardinality due to `task_key` label). |
| `--table-check-interval` | `10` | Number of scrapes between table availability checks (for optional tables like pipelines). |
| `--log.level` | `info` | Only log messages with the given severity or above. One of: `debug`, `info`, `warn`, `error`. |
| `--log.format` | `logfmt` | Output format of log messages. One of: `logfmt`, `json`. |

Example usage:

```sh
./databricks-exporter \
  --server-hostname=<your-host-name>.cloud.databricks.com \
  --warehouse-http-path=/sql/1.0/warehouses/abc123def456 \
  --client-id=<your-client-or-app-ID> \
  --client-secret=<your-client-secret-here>
```

### Environment variables

Alternatively, the exporter may be configured using environment variables:

| Name | Description |
|------|-------------|
| `DATABRICKS_EXPORTER_SERVER_HOSTNAME` | The Databricks workspace hostname. |
| `DATABRICKS_EXPORTER_WAREHOUSE_HTTP_PATH` | The HTTP path of the SQL Warehouse. |
| `DATABRICKS_EXPORTER_CLIENT_ID` | The OAuth2 Client ID for Service Principal authentication. |
| `DATABRICKS_EXPORTER_CLIENT_SECRET` | The OAuth2 Client Secret for Service Principal authentication. |
| `DATABRICKS_EXPORTER_WEB_TELEMETRY_PATH` | Path under which to expose metrics. |
| `DATABRICKS_EXPORTER_QUERY_TIMEOUT` | Timeout for database queries. |
| `DATABRICKS_EXPORTER_BILLING_LOOKBACK` | How far back to look for billing data. |
| `DATABRICKS_EXPORTER_JOBS_LOOKBACK` | How far back to look for job runs. |
| `DATABRICKS_EXPORTER_PIPELINES_LOOKBACK` | How far back to look for pipeline runs. |
| `DATABRICKS_EXPORTER_QUERIES_LOOKBACK` | How far back to look for SQL warehouse queries. |
| `DATABRICKS_EXPORTER_SLA_THRESHOLD` | Duration threshold (in seconds) for job SLA miss detection. |
| `DATABRICKS_EXPORTER_COLLECT_TASK_RETRIES` | Collect task retry metrics (set to `true` to enable). |
| `DATABRICKS_EXPORTER_TABLE_CHECK_INTERVAL` | Number of scrapes between table availability checks. |

Example usage:

```sh
export DATABRICKS_EXPORTER_SERVER_HOSTNAME="dbc-abc123-def456.cloud.databricks.com"
export DATABRICKS_EXPORTER_WAREHOUSE_HTTP_PATH="/sql/1.0/warehouses/abc123def456"
export DATABRICKS_EXPORTER_CLIENT_ID="4a8adace-cdf5-4489-b9c2-2b6f9dd7682f"
export DATABRICKS_EXPORTER_CLIENT_SECRET="your-client-secret-here"

./databricks-exporter
```

## Authentication

### Service principal OAuth2 authentication

The exporter uses OAuth2 Machine-to-Machine (M2M) authentication with Databricks Service Principals, following Databricks' recommended security practices.

#### Prerequisites

1. **Unity Catalog Enabled**: Your Databricks workspace must have Unity Catalog enabled to access System Tables.

2. **Service Principal**: Create a Service Principal in your Databricks workspace with appropriate permissions.

3. **SQL Warehouse**: Have a running SQL Warehouse (or one configured to auto-start) for executing queries.

#### Setting up a service principal

1. Log into your Databricks workspace
2. Go to **Settings** → **Admin Console** → **Service Principals**
3. Click **Add Service Principal**
4. Note the **Application ID** (this is your Client ID)
5. Click **Generate Secret** under OAuth Secrets
6. Copy and securely store the **Client Secret** (you won't see it again!)

#### Required permissions

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

### Getting required configuration values

#### Server hostname
- Found in your Databricks workspace URL
- Example: `dbc-abc123-def456.cloud.databricks.com`
- Remove the `https://` prefix

#### Warehouse HTTP path
1. Go to **SQL Warehouses** in your Databricks workspace
2. Select your SQL Warehouse
3. Click **Connection Details** tab
4. Copy the **HTTP Path** (format: `/sql/1.0/warehouses/<warehouse-id>`)

#### Client ID and client secret
1. Go to **Settings** → **Admin Console** → **Service Principals**
2. Find your Service Principal
3. The **Application ID** is your **Client ID**
4. Generate and copy the **Client Secret**

## Metrics and system tables

- **[Metrics Reference](docs/metrics-reference.md)** — Complete list of exported metrics with descriptions, labels, and types
- **[System Tables Reference](docs/databricks-system-tables.md)** — Databricks system tables queried by the exporter

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

## Prometheus configuration

Add the exporter as a scrape target in your Prometheus configuration:

```yaml
scrape_configs:
  - job_name: 'databricks'
    scrape_interval: 10m
    scrape_timeout: 9m
    static_configs:
      - targets: ['localhost:9976']
```

### Scrape interval requirements

| Setting | Minimum | Maximum | Recommended |
|---------|---------|---------|-------------|
| `scrape_interval` | 10m | 30m | 10m |
| `scrape_timeout` | 9m | 29m | 9m |

**Why these constraints?**

- **Minimum 10 minutes**: The exporter queries Databricks System Tables which can take 30-120 seconds depending on data volume. Scraping more frequently wastes resources and may cause overlapping scrapes.

- **Maximum 30 minutes**: The mixin dashboards use `last_over_time(...[30m:])` to bridge gaps between scrapes. Intervals longer than 30 minutes will cause gaps in dashboard visualizations.

- **Timeout < Interval**: Always set `scrape_timeout` less than `scrape_interval` to prevent overlapping scrapes. A 1-minute buffer (e.g., 9m timeout for 10m interval) is recommended.

### Grafana Alloy configuration

If using Grafana Alloy instead of Prometheus:

```alloy
prometheus.scrape "databricks" {
  targets = [{
    __address__ = "localhost:9976",
  }]

  forward_to      = [prometheus.remote_write.default.receiver]
  scrape_interval = "10m"
  scrape_timeout  = "9m"
}
```

### Lookback windows

The exporter uses **sliding window queries** to collect metrics from Databricks System Tables. Each scrape queries data from `now - lookback` to `now`, meaning:

- Metrics represent counts/aggregates over the lookback window
- Values can decrease as older data "slides out" of the window
- The lookback must be long enough to ensure data continuity between scrapes

#### Default lookback windows

| Domain | Lookback | Rationale |
|--------|----------|-----------|
| Billing | 24h | Databricks billing data has 24-48h lag; daily granularity |
| Jobs | 4h | Covers multiple scrape intervals with safety margin |
| Pipelines | 4h | Covers multiple scrape intervals with safety margin |
| Queries | 2h | SQL queries are more frequent; shorter window sufficient |

#### Lookback vs scrape interval relationship

The lookback window must be **significantly larger** than the scrape interval to prevent data loss:

```
Minimum safe lookback = scrape_interval × 4
```

| Scrape Interval | Minimum Lookback | Recommended Lookback |
|-----------------|------------------|----------------------|
| 10m | 40m | 2h+ |
| 15m | 1h | 2h+ |
| 30m | 2h | 4h+ |

**Why this matters:**

1. **Data continuity**: If a job completes at time T, it appears in scrapes from T to T+lookback. With a 10m scrape interval and 4h lookback, that job appears in ~24 consecutive scrapes.

2. **Missed scrapes**: If a scrape fails or is delayed, the next scrape still captures the data because the lookback window overlaps.

3. **Dashboard accuracy**: The mixin dashboards use `last_over_time([30m:])` to bridge gaps. This assumes data points exist within each 30-minute window.

#### Customizing lookback windows

You can adjust lookback windows based on your workload patterns:

```sh
# For high-frequency job environments (many short jobs)
./databricks-exporter --jobs-lookback=2h --pipelines-lookback=2h

# For low-frequency batch environments (few long jobs)
./databricks-exporter --jobs-lookback=8h --pipelines-lookback=8h
```

**Trade-offs:**
- Longer lookback = more data per scrape = slower queries = higher Databricks costs
- Shorter lookback = risk of missing data if scrapes are delayed

## Troubleshooting

### Authentication errors (401 Unauthorized)

If you see authentication errors:
- Verify your Client ID (Application ID) is correct
- Ensure the Client Secret hasn't expired
- Check that the Service Principal has access to the workspace

### Credit exhaustion (400 Bad Request)

Error: `Sorry, cannot run the resource because you've exhausted your available credits`

This means your Databricks account has run out of credits. Add a payment method or credits to your account to continue.

### Connection errors

- Ensure the SQL Warehouse is running (or set to auto-start)
- Verify the Warehouse HTTP Path is correct (should start with `/sql/1.0/warehouses/`)
- Check that the Server Hostname doesn't include `https://`

### No metrics appearing

If `databricks_exporter_up` is 1 but some metrics don't appear:
- Some metrics only appear when there's data to report (e.g., no retries means no retry metric)
- Billing metrics require recent usage data (check the last 30 days)
- Job and pipeline metrics require recent executions (check the last 24 hours)
- Query metrics require recent SQL activity (check the last hour)
- Verify Unity Catalog is enabled on your workspace
- Ensure the Service Principal has permissions to read all System Tables

### Pipeline metrics not available (TABLE_OR_VIEW_NOT_FOUND)

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

### Debug logging

Enable debug logging for more detailed information:

```sh
./databricks-exporter --log.level=debug [other flags...]
```

## License

Licensed under the Apache License, Version 2.0. See LICENSE for details.
