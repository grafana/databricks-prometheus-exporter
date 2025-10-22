# databricks-prometheus-exporter

Exports [Databricks](https://databricks.com) billing and usage statistics via HTTP for Prometheus consumption.

## Overview

This exporter connects to a Databricks SQL Warehouse and queries the `system.billing.account_prices` table to retrieve billing information. The metrics are exposed in Prometheus format, allowing you to monitor and analyze your Databricks usage and costs.

## Configuration

### Command line flags

The exporter may be configured through its command line flags:

```
  -h, --help                          Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address=:9976 ...  Addresses on which to expose metrics and web interface. Repeatable for multiple addresses.
      --web.config.file=""            Path to configuration file that can enable TLS or authentication.
      --web.telemetry-path="/metrics" Path under which to expose metrics.
      --server-hostname=HOSTNAME      The Databricks workspace hostname (e.g., dbc-abc123-def456.cloud.databricks.com).
      --http-path=HTTP-PATH           The HTTP path of the SQL warehouse (e.g., /sql/1.0/warehouses/abc123def456).
      --client-id=CLIENT-ID           The OAuth2 Client ID (Application ID) for Service Principal authentication.
      --client-secret=CLIENT-SECRET   The OAuth2 Client Secret for Service Principal authentication.
      --catalog="system"              The catalog to use when querying.
      --schema="billing"              The schema to use when querying.
      --version                       Show application version.
      --log.level=info                Only log messages with the given severity or above. One of: [debug, info, warn, error]
      --log.format=logfmt             Output format of log messages. One of: [logfmt, json]
```

Example usage:

```sh
./databricks-exporter \
  --server-hostname=<your-host-name>.cloud.databricks.com \
  --http-path=/sql/1.0/warehouses/abc123def456 \
  --client-id=<your-client-or-app-ID>> \
  --client-secret=<your-client-secret-here>
```

### Environment Variables

Alternatively, the exporter may be configured using environment variables:

| Name                                           | Description                                                                                      |
| ---------------------------------------------- | ------------------------------------------------------------------------------------------------ |
| DATABRICKS_EXPORTER_SERVER_HOSTNAME            | The Databricks workspace hostname (e.g., dbc-abc123-def456.cloud.databricks.com).              |
| DATABRICKS_EXPORTER_HTTP_PATH                  | The HTTP path of the SQL warehouse (e.g., /sql/1.0/warehouses/abc123def456).                   |
| DATABRICKS_EXPORTER_CLIENT_ID                  | The OAuth2 Client ID (Application ID) for Service Principal authentication.                     |
| DATABRICKS_EXPORTER_CLIENT_SECRET              | The OAuth2 Client Secret for Service Principal authentication.                                  |
| DATABRICKS_EXPORTER_CATALOG                    | The catalog to use when querying (default: system).                                             |
| DATABRICKS_EXPORTER_SCHEMA                     | The schema to use when querying (default: billing).                                             |
| DATABRICKS_EXPORTER_WEB_TELEMETRY_PATH         | Path under which to expose metrics (default: /metrics).                                         |

Example usage:

```sh
export DATABRICKS_EXPORTER_SERVER_HOSTNAME="dbc-abc123-def456.cloud.databricks.com"
export DATABRICKS_EXPORTER_HTTP_PATH="/sql/1.0/warehouses/abc123def456"
export DATABRICKS_EXPORTER_CLIENT_ID="4a8adace-cdf5-4489-b9c2-2b6f9dd7682f"
export DATABRICKS_EXPORTER_CLIENT_SECRET="your-client-secret-here"

./databricks-exporter
```

## Authentication

### Service Principal OAuth2 Authentication

The exporter uses OAuth2 Machine-to-Machine (M2M) authentication with Databricks Service Principals, following Databricks' recommended security practices.

#### Prerequisites

1. **Unity Catalog Enabled**: Your Databricks workspace must have Unity Catalog enabled to access the `system.billing.account_prices` table.

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

The Service Principal needs:
- Access to the Databricks workspace
- Permission to use the SQL Warehouse
- Read access to `system.billing.account_prices` table (requires account-level permissions)

### Getting Required Configuration Values

#### Server Hostname
- Found in your Databricks workspace URL
- Example: `dbc-abc123-def456.cloud.databricks.com`
- Remove the `https://` prefix

#### HTTP Path
1. Go to **SQL Warehouses** in your Databricks workspace
2. Select your SQL Warehouse
3. Click **Connection Details** tab
4. Copy the **HTTP Path** (format: `/sql/1.0/warehouses/<warehouse-id>`)

#### Client ID and Client Secret
1. Go to **Settings** → **Admin Console** → **Service Principals**
2. Find your Service Principal
3. The **Application ID** is your **Client ID**
4. Generate and copy the **Client Secret**

## Metrics

The exporter provides the following metrics:

### `databricks_billing_account_price`

Account pricing information from the `system.billing.account_prices` table, aggregated over the last 7 days.

**Labels:**
- `account_id`: Databricks account identifier
- `sku_name`: SKU/product name
- `cloud`: Cloud provider (AWS, Azure, GCP)
- `currency_code`: Currency code (e.g., USD, EUR)
- `usage_unit`: Unit of measurement for usage

**Type:** Gauge

**Value:** Average pricing for the SKU over the collection period

### `databricks_up`

Metric indicating the status of the exporter collection.

**Type:** Gauge

**Values:**
- `1`: Connection to Databricks was successful, and all available metrics were collected
- `0`: The exporter failed to collect one or more metrics due to connection issues

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
  -e DATABRICKS_EXPORTER_HTTP_PATH="/sql/1.0/warehouses/abc123def456" \
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
- Verify the HTTP Path is correct (should start with `/sql/1.0/warehouses/`)
- Check that the Server Hostname doesn't include `https://`

### No Billing Data

If `databricks_up` is 1 but no billing metrics appear:
- Verify Unity Catalog is enabled on your workspace
- Ensure the Service Principal has account-level permissions to read billing data
- Check that data exists in `system.billing.account_prices` table

### Debug Logging

Enable debug logging for more detailed information:

```sh
./databricks-exporter --log.level=debug [other flags...]
```

## License

Licensed under the Apache License, Version 2.0. See LICENSE for details.
