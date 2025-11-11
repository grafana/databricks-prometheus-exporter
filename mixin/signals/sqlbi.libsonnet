function(this) {
  local legendCustomTemplate = '{{instance}} - {{workspace_id}}',
  local aggregationLabels = std.join(',', this.groupLabels + this.instanceLabels),
  filteringSelector: this.filteringSelector,
  groupLabels: this.groupLabels,
  instanceLabels: this.instanceLabels,
  legendCustomTemplate: legendCustomTemplate,
  aggLevel: 'none',
  aggFunction: 'avg',
  signals: {
    // SQL Query Metrics
    queriesTotal: {
      name: 'Queries total',
      nameShort: 'Queries',
      description: 'SQL queries executed (warehouse & serverless).',
      type: 'counter',
      unit: 'queries',
      sources: {
        prometheus: {
          expr: 'databricks_queries_total{%(queriesSelector)s}',
          legendCustomTemplate: '{{workspace_id}}',
        },
      },
    },

    queryDuration: {
      name: 'Query duration',
      nameShort: 'Query latency',
      description: 'Query latency (p50/p95 derived from history).',
      type: 'raw',
      unit: 's',
      sources: {
        prometheus: {
          expr: 'databricks_query_duration_seconds{%(queriesSelector)s}',
          legendCustomTemplate: '{{workspace_id}} - p{{quantile}}',
        },
      },
    },

    queryErrorsTotal: {
      name: 'Query errors total',
      nameShort: 'Query errors',
      description: 'Failed queries (count).',
      type: 'counter',
      unit: 'queries',
      sources: {
        prometheus: {
          expr: 'databricks_query_errors_total{%(queriesSelector)s}',
          legendCustomTemplate: '{{workspace_id}}',
        },
      },
    },

    queriesRunning: {
      name: 'Queries running',
      nameShort: 'Concurrent queries',
      description: 'Concurrent/running queries (derived from overlapping intervals).',
      type: 'gauge',
      unit: 'queries',
      sources: {
        prometheus: {
          expr: 'databricks_queries_running{%(queriesSelector)s}',
          legendCustomTemplate: '{{workspace_id}}',
        },
      },
    },

    // Query load rate (1h)
    queryLoad1h: {
      name: 'Query load (1h)',
      nameShort: 'Queries (1h)',
      description: 'Total queries executed in the last hour.',
      type: 'raw',
      unit: 'queries',
      sources: {
        prometheus: {
          expr: 'sum by (' + aggregationLabels + ') (increase(databricks_queries_total{%(queriesSelector)s}[1h]))',
          legendCustomTemplate: 'Last hour',
        },
      },
    },

    // Query load rate (24h)
    queryLoad24h: {
      name: 'Query load (24h)',
      nameShort: 'Queries (24h)',
      description: 'Total queries executed in the last 24 hours.',
      type: 'raw',
      unit: 'queries',
      sources: {
        prometheus: {
          expr: 'sum by (' + aggregationLabels + ') (increase(databricks_queries_total{%(queriesSelector)s}[24h]))',
          legendCustomTemplate: 'Last 24h',
        },
      },
    },

    // Query error rate (1h)
    queryErrorRate1h: {
      name: 'Query error rate (1h)',
      nameShort: 'Error % (1h)',
      description: 'Query error percentage in the last hour.',
      type: 'raw',
      unit: 'percentunit',
      sources: {
        prometheus: {
          expr: |||
            sum by (%(aggregationLabels)s) (increase(databricks_query_errors_total{%(queriesSelector)s}[1h]))
            /
            sum by (%(aggregationLabels)s) (increase(databricks_queries_total{%(queriesSelector)s}[1h]))
          ||| % {
            aggregationLabels: aggregationLabels,
            queriesSelector: '%(queriesSelector)s',
          },
          legendCustomTemplate: 'Error rate',
        },
      },
    },

    // Query p95 latency current
    queryP95Latency: {
      name: 'Query p95 latency',
      nameShort: 'p95 latency',
      description: 'Current p95 query latency.',
      type: 'raw',
      unit: 's',
      sources: {
        prometheus: {
          expr: 'databricks_query_duration_seconds{%(queriesSelector)s, quantile="0.95"}',
          legendCustomTemplate: 'p95',
        },
      },
    },

    // Query p50 latency current
    queryP50Latency: {
      name: 'Query p50 latency',
      nameShort: 'p50 latency',
      description: 'Current p50 query latency.',
      type: 'raw',
      unit: 's',
      sources: {
        prometheus: {
          expr: 'databricks_query_duration_seconds{%(queriesSelector)s, quantile="0.50"}',
          legendCustomTemplate: 'p50',
        },
      },
    },

    // Concurrency current
    concurrencyCurrent: {
      name: 'Concurrency (current)',
      nameShort: 'Concurrency now',
      description: 'Current number of concurrent queries.',
      type: 'raw',
      unit: 'queries',
      sources: {
        prometheus: {
          expr: 'databricks_queries_running{%(queriesSelector)s}',
          legendCustomTemplate: 'Now',
        },
      },
    },

    // Concurrency baseline (7d p95)
    concurrencyBaseline: {
      name: 'Concurrency baseline (7d p95)',
      nameShort: 'Concurrency 7d p95',
      description: 'Baseline concurrency over 7 days (p95).',
      type: 'raw',
      unit: 'queries',
      sources: {
        prometheus: {
          expr: 'quantile_over_time(0.95, databricks_queries_running{%(queriesSelector)s}[7d])',
          legendCustomTemplate: '7d baseline',
        },
      },
    },

    // DoD queries delta
    dodQueriesDelta: {
      name: 'DoD queries delta',
      nameShort: 'DoD Queries Δ%',
      description: 'Day-over-day change in query count (D-1 vs D-2).',
      type: 'raw',
      unit: 'percentunit',
      sources: {
        prometheus: {
          expr: |||
            (
              sum by (%(aggregationLabels)s) (increase(databricks_queries_total{%(queriesSelector)s}[24h] offset 1d))
              -
              sum by (%(aggregationLabels)s) (increase(databricks_queries_total{%(queriesSelector)s}[24h] offset 2d))
            )
            /
            sum by (%(aggregationLabels)s) (increase(databricks_queries_total{%(queriesSelector)s}[24h] offset 2d))
          ||| % {
            aggregationLabels: aggregationLabels,
            queriesSelector: '%(queriesSelector)s',
          },
          legendCustomTemplate: 'DoD',
        },
      },
    },

    // DoD p95 latency delta
    dodP95LatencyDelta: {
      name: 'DoD p95 latency delta',
      nameShort: 'DoD p95 Δ%',
      description: 'Day-over-day change in p95 query latency (D-1 vs D-2).',
      type: 'raw',
      unit: 'percentunit',
      sources: {
        prometheus: {
          expr: |||
            (
              avg_over_time(databricks_query_duration_seconds{%(queriesSelector)s, quantile="0.95"}[1d] offset 1d)
              -
              avg_over_time(databricks_query_duration_seconds{%(queriesSelector)s, quantile="0.95"}[1d] offset 2d)
            )
            /
            avg_over_time(databricks_query_duration_seconds{%(queriesSelector)s, quantile="0.95"}[1d] offset 2d)
          ||| % { queriesSelector: '%(queriesSelector)s' },
          legendCustomTemplate: 'DoD p95',
        },
      },
    },

    // Query rate
    queryRate: {
      name: 'Query rate',
      nameShort: 'Query rate',
      description: 'Query execution rate over time.',
      type: 'raw',
      unit: 'qps',
      sources: {
        prometheus: {
          expr: 'sum by (' + aggregationLabels + ') (rate(databricks_queries_total{%(queriesSelector)s}[5m]))',
          legendCustomTemplate: 'Queries/sec',
        },
      },
    },

    // Query error rate over time
    queryErrorRate: {
      name: 'Query error rate',
      nameShort: 'Error rate',
      description: 'SQL error rate over time.',
      type: 'raw',
      unit: 'percentunit',
      sources: {
        prometheus: {
          expr: |||
            sum by (%(aggregationLabels)s) (rate(databricks_query_errors_total{%(queriesSelector)s}[30m]))
            /
            sum by (%(aggregationLabels)s) (rate(databricks_queries_total{%(queriesSelector)s}[30m]))
          ||| % {
            aggregationLabels: aggregationLabels,
            queriesSelector: '%(queriesSelector)s',
          },
          legendCustomTemplate: 'Error rate',
        },
      },
    },

    // Queries by workspace
    queriesByWorkspace: {
      name: 'Queries by workspace',
      nameShort: 'By workspace',
      description: 'Query volume breakdown by workspace.',
      type: 'raw',
      unit: 'queries',
      sources: {
        prometheus: {
          expr: 'sum by (workspace_id) (increase(databricks_queries_total{%(queriesSelector)s}[1h]))',
          legendCustomTemplate: '{{workspace_id}}',
        },
      },
    },
  },
}

