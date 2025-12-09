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
    queryDuration: {
      name: 'Query duration (p95)',
      nameShort: 'Query p95 duration',
      description: 'Query duration (p95 by warehouse).',
      type: 'gauge',
      unit: 's',
      sources: {
        prometheus: {
          expr: 'max by (warehouse_id) (databricks_query_duration_seconds{%(queriesSelector)s, quantile="0.95"})',
          legendCustomTemplate: 'Warehouse - {{warehouse_id}}',
        },
      },
    },

    queryErrors: {
      name: 'Query errors',
      nameShort: 'Errors',
      description: 'Query errors (adapts to dashboard time range).',
      type: 'raw',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'sum by (workspace_id) (increase(databricks_query_errors_total{%(queriesSelector)s}[$__interval:] offset $__interval))',
          legendCustomTemplate: '{{workspace_id}}',
        },
      },
    },

    queryLoad: {
      name: 'Query load',
      nameShort: 'Queries',
      description: 'Total queries executed (adapts to dashboard time range).',
      type: 'raw',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'sum by (' + aggregationLabels + ') (increase(databricks_queries_total{%(queriesSelector)s}[$__interval:] offset $__interval))',
          legendCustomTemplate: 'Queries',
        },
      },
    },

    queryErrorRateAggregate: {
      name: 'Query error rate (aggregate)',
      nameShort: 'Error %',
      description: 'Aggregate query error percentage across all workspaces (adapts to dashboard time range).',
      type: 'raw',
      unit: 'percentunit',
      sources: {
        prometheus: {
          expr: |||
            sum (increase(databricks_query_errors_total{%(queriesSelector)s}[$__interval:] offset $__interval))
            /
            sum (increase(databricks_queries_total{%(queriesSelector)s}[$__interval:] offset $__interval))
          ||| % {
            queriesSelector: '%(queriesSelector)s',
          },
          legendCustomTemplate: 'Error rate %%',
        },
      },
    },

    queryP95Latency: {
      name: 'Query p95 latency (max)',
      nameShort: 'p95 latency (max)',
      description: 'Maximum p95 query latency across all warehouses.',
      type: 'gauge',
      unit: 's',
      sources: {
        prometheus: {
          expr: 'max(databricks_query_duration_seconds{%(queriesSelector)s, quantile="0.95"})',
          legendCustomTemplate: 'p95 latency',
        },
      },
    },

    queryP50Latency: {
      name: 'Query p50 latency (max)',
      nameShort: 'p50 latency (max)',
      description: 'Maximum p50 query latency across all warehouses.',
      type: 'gauge',
      unit: 's',
      sources: {
        prometheus: {
          expr: 'max(databricks_query_duration_seconds{%(queriesSelector)s, quantile="0.50"})',
          legendCustomTemplate: 'Current p50',
        },
      },
    },

    queryP50Latency7dBaseline: {
      name: 'Query p50 latency (7d baseline)',
      nameShort: 'p50 7d baseline',
      description: '7-day rolling median baseline for p50 query latency.',
      type: 'raw',
      unit: 's',
      sources: {
        prometheus: {
          expr: 'max(quantile_over_time(0.5, databricks_query_duration_seconds{%(queriesSelector)s, quantile="0.50"}[7d]))',
          legendCustomTemplate: '7 days median',
        },
      },
    },

    concurrencyCurrent: {
      name: 'Total queries running (max)',
      nameShort: 'Total queries running (max)',
      description: 'Maximum number of concurrent queries across all warehouses.',
      type: 'gauge',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'max(databricks_queries_running{%(queriesSelector)s})',
          legendCustomTemplate: 'Total queries running',
        },
      },
    },

    concurrencyBaseline: {
      name: 'Max concurrent queries (7d p95)',
      nameShort: 'Concurrency 7d p95',
      description: 'Maximum baseline concurrency over 7 days (p95).',
      type: 'raw',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'max(quantile_over_time(0.95, databricks_queries_running{%(queriesSelector)s}[7d]))',
          legendCustomTemplate: '7d baseline',
        },
      },
    },

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

    dodP95LatencyDelta: {
      name: 'DoD p95 latency delta (max)',
      nameShort: 'DoD p95 Δ%',
      description: 'Maximum day-over-day change in p95 query latency across all warehouses (D-1 vs D-2).',
      type: 'raw',
      unit: 'percentunit',
      sources: {
        prometheus: {
          expr: |||
            max(
              (
                avg_over_time(databricks_query_duration_seconds{%(queriesSelector)s, quantile="0.95"}[1d] offset 1d)
                -
                avg_over_time(databricks_query_duration_seconds{%(queriesSelector)s, quantile="0.95"}[1d] offset 2d)
              )
              /
              avg_over_time(databricks_query_duration_seconds{%(queriesSelector)s, quantile="0.95"}[1d] offset 2d)
            )
          ||| % { queriesSelector: '%(queriesSelector)s' },
          legendCustomTemplate: 'DoD p95 (max)',
        },
      },
    },

    queryErrorRate: {
      name: 'Query error rate',
      nameShort: 'Error rate',
      description: 'SQL error percentage (adapts to dashboard time range).',
      type: 'raw',
      unit: 'percentunit',
      sources: {
        prometheus: {
          expr: |||
            sum by (%(aggregationLabels)s) (increase(databricks_query_errors_total{%(queriesSelector)s}[$__interval:] offset $__interval))
            /
            sum by (%(aggregationLabels)s) (increase(databricks_queries_total{%(queriesSelector)s}[$__interval:] offset $__interval))
          ||| % {
            aggregationLabels: aggregationLabels,
            queriesSelector: '%(queriesSelector)s',
          },
          legendCustomTemplate: '{{workspace_id}} - Error rate %%',
        },
      },
    },

    queriesByWorkspace: {
      name: 'Queries by workspace',
      nameShort: 'By workspace',
      description: 'Query volume breakdown by workspace (adapts to dashboard time range).',
      type: 'raw',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'sum by (workspace_id) (increase(databricks_queries_total{%(queriesSelector)s}[$__interval:] offset $__interval))',
          legendCustomTemplate: '{{workspace_id}}',
        },
      },
    },

    // Detailed drill-down signals by warehouse_id
    topWarehousesByQueries: {
      name: 'Top warehouses by queries',
      nameShort: 'Top by queries',
      description: 'Warehouses ranked by total query volume (sliding window gauge).',
      type: 'gauge',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'sum by (workspace_id, warehouse_id) (databricks_queries_total{%(queriesSelector)s})',
          legendCustomTemplate: '{{warehouse_id}}',
        },
      },
    },

    topWarehousesByErrors: {
      name: 'Top warehouses by errors',
      nameShort: 'Top by errors',
      description: 'Warehouses ranked by total error count (sliding window gauge).',
      type: 'gauge',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'sum by (workspace_id, warehouse_id) (databricks_query_errors_total{%(queriesSelector)s})',
          legendCustomTemplate: '{{warehouse_id}}',
        },
      },
    },

    topWarehousesByLatency: {
      name: 'Top warehouses by p95 latency',
      nameShort: 'Top by latency',
      description: 'Warehouses ranked by p95 query latency (current gauge value).',
      type: 'gauge',
      unit: 's',
      sources: {
        prometheus: {
          expr: 'max by (warehouse_id) (databricks_query_duration_seconds{%(queriesSelector)s, quantile="0.95"})',
          legendCustomTemplate: '{{warehouse_id}}',
        },
      },
    },

    queriesByWarehouse: {
      name: 'Queries by warehouse (time series)',
      nameShort: 'By warehouse',
      description: 'Query volume over time by warehouse (adapts to dashboard time range).',
      type: 'raw',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'increase(databricks_queries_total{%(queriesSelector)s}[$__interval:] offset $__interval)',
          legendCustomTemplate: '{{warehouse_id}}',
        },
      },
    },

    queryErrorsByWarehouse: {
      name: 'Query errors by warehouse (time series)',
      nameShort: 'Errors by warehouse',
      description: 'Query errors over time by warehouse (adapts to dashboard time range).',
      type: 'raw',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'increase(databricks_query_errors_total{%(queriesSelector)s}[$__interval:] offset $__interval)',
          legendCustomTemplate: '{{warehouse_id}}',
        },
      },
    },

    queryLatencyByWarehouse: {
      name: 'Query latency by warehouse (time series)',
      nameShort: 'Latency by warehouse',
      description: 'Query p95 latency over time by warehouse.',
      type: 'gauge',
      unit: 's',
      sources: {
        prometheus: {
          expr: 'databricks_query_duration_seconds{%(queriesSelector)s, quantile="0.95"}',
          legendCustomTemplate: '{{warehouse_id}}',
        },
      },
    },

    concurrencyByWarehouse: {
      name: 'Concurrency by warehouse',
      nameShort: 'Concurrency by warehouse',
      description: 'Concurrent queries by warehouse.',
      type: 'gauge',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'databricks_queries_running{%(queriesSelector)s}',
          legendCustomTemplate: '{{warehouse_id}}',
        },
      },
    },
  },
}
