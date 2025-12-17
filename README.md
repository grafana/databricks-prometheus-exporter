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

## System tables used

Please refer to the [Systems table reference](docs/databricks-system-tables.md) for more information on which system tables are queried by the exporter.

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
