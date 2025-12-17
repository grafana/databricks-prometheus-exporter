package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/databricks-prometheus-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
)

var (
	webConfig         = webflag.AddFlags(kingpin.CommandLine, ":9976")
	metricPath        = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").Envar("DATABRICKS_EXPORTER_WEB_TELEMETRY_PATH").String()
	serverHostname    = kingpin.Flag("server-hostname", "The Databricks workspace hostname (e.g., dbc-abc123-def456.cloud.databricks.com).").Envar("DATABRICKS_EXPORTER_SERVER_HOSTNAME").Required().String()
	warehouseHTTPPath = kingpin.Flag("warehouse-http-path", "The HTTP path of the SQL Warehouse (e.g., /sql/1.0/warehouses/abc123def456).").Envar("DATABRICKS_EXPORTER_WAREHOUSE_HTTP_PATH").Required().String()
	clientID          = kingpin.Flag("client-id", "The OAuth2 Client ID (Application ID) for Service Principal authentication.").Envar("DATABRICKS_EXPORTER_CLIENT_ID").Required().String()
	clientSecret      = kingpin.Flag("client-secret", "The OAuth2 Client Secret for Service Principal authentication.").Envar("DATABRICKS_EXPORTER_CLIENT_SECRET").Required().String()

	// Query settings
	queryTimeout = kingpin.Flag("query-timeout", "Timeout for database queries.").Default("5m").Envar("DATABRICKS_EXPORTER_QUERY_TIMEOUT").Duration()

	// Lookback windows
	billingLookback   = kingpin.Flag("billing-lookback", "How far back to look for billing data.").Default("24h").Envar("DATABRICKS_EXPORTER_BILLING_LOOKBACK").Duration()
	jobsLookback      = kingpin.Flag("jobs-lookback", "How far back to look for job runs.").Default("2h").Envar("DATABRICKS_EXPORTER_JOBS_LOOKBACK").Duration()
	pipelinesLookback = kingpin.Flag("pipelines-lookback", "How far back to look for pipeline runs.").Default("2h").Envar("DATABRICKS_EXPORTER_PIPELINES_LOOKBACK").Duration()
	queriesLookback   = kingpin.Flag("queries-lookback", "How far back to look for SQL warehouse queries.").Default("1h").Envar("DATABRICKS_EXPORTER_QUERIES_LOOKBACK").Duration()

	// SLA settings (default matches collector.DefaultSLAThresholdSeconds)
	slaThreshold = kingpin.Flag("sla-threshold", "Duration threshold (in seconds) for job SLA miss detection.").Default("3600").Envar("DATABRICKS_EXPORTER_SLA_THRESHOLD").Int()

	// Cardinality controls
	collectTaskRetries = kingpin.Flag("collect-task-retries", "Collect task retry metrics (high cardinality due to task_key label).").Default("false").Envar("DATABRICKS_EXPORTER_COLLECT_TASK_RETRIES").Bool()
)

const (
	// The name of the exporter.
	exporterName    = "databricks_exporter"
	landingPageHTML = `<html>
<head><title>Databricks exporter</title></head>
	<body>
		<h1>Databricks exporter</h1>
		<p><a href='%s'>Metrics</a></p>
	</body>
</html>`
)

func main() {
	kingpin.Version(version.Print(exporterName))

	promlogConfig := &promlog.Config{}

	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promlog.New(promlogConfig)

	// Construct the collector, using the flags for configuration
	c := &collector.Config{
		ServerHostname:    *serverHostname,
		WarehouseHTTPPath: *warehouseHTTPPath,
		ClientID:          *clientID,
		ClientSecret:      *clientSecret,
		QueryTimeout:      *queryTimeout,

		// Lookback windows
		BillingLookback:   *billingLookback,
		JobsLookback:      *jobsLookback,
		PipelinesLookback: *pipelinesLookback,
		QueriesLookback:   *queriesLookback,

		// SLA settings
		SLAThresholdSeconds: *slaThreshold,

		// Cardinality controls
		CollectTaskRetries: *collectTaskRetries,
	}

	if err := c.Validate(); err != nil {
		level.Error(logger).Log("msg", "Configuration is invalid.", "err", err)
		os.Exit(1)
	}

	// Add component prefix to logger for better log correlation
	collectorLogger := log.With(logger, "component", "databricks-exporter")
	col := collector.NewCollector(collectorLogger, c)

	// Register collector with prometheus client library
	prometheus.MustRegister(col)

	serveMetrics(logger)
}

func serveMetrics(logger log.Logger) {
	landingPage := []byte(fmt.Sprintf(landingPageHTML, *metricPath))

	http.Handle(*metricPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		if _, err := w.Write(landingPage); err != nil {
			level.Error(logger).Log("msg", "Failed to write landing page", "err", err)
		}
	})

	srv := &http.Server{}
	slogger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	if err := web.ListenAndServe(srv, webConfig, slogger); err != nil {
		level.Error(logger).Log("msg", "Error running HTTP server", "err", err)
		os.Exit(1)
	}
}
