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
    // Jobs Metrics
    jobRunsTotal: {
      name: 'Job runs total',
      nameShort: 'Job runs',
      description: 'Number of Lakeflow Jobs runs (per workspace/day).',
      type: 'counter',
      unit: 'runs',
      sources: {
        prometheus: {
          expr: 'databricks_job_runs_total{%(queriesSelector)s}',
          legendCustomTemplate: '{{workspace_id}}',
        },
      },
    },

    jobRunDuration: {
      name: 'Job run duration',
      nameShort: 'Job duration',
      description: 'Job run duration (p50/p95 summaries derived from start/end).',
      type: 'raw',
      unit: 's',
      sources: {
        prometheus: {
          expr: 'databricks_job_run_duration_seconds{%(queriesSelector)s}',
          legendCustomTemplate: '{{workspace_id}} - p{{quantile}}',
        },
      },
    },

    jobRunStatusTotal: {
      name: 'Job run status total',
      nameShort: 'Job status',
      description: 'Job status counts (SUCCEEDED/FAILED/CANCELED).',
      type: 'counter',
      unit: 'runs',
      sources: {
        prometheus: {
          expr: 'databricks_job_run_status_total{%(queriesSelector)s}',
          legendCustomTemplate: '{{workspace_id}} - {{status}}',
        },
      },
    },

    taskRetriesTotal: {
      name: 'Task retries total',
      nameShort: 'Task retries',
      description: 'Number of retries across job tasks (stability signal).',
      type: 'counter',
      unit: 'retries',
      sources: {
        prometheus: {
          expr: 'databricks_task_retries_total{%(queriesSelector)s}',
          legendCustomTemplate: '{{workspace_id}}',
        },
      },
    },

    jobSlaMissTotal: {
      name: 'Job SLA miss total',
      nameShort: 'SLA misses',
      description: 'Runs exceeding configured SLA (computed in exporter).',
      type: 'counter',
      unit: 'breaches',
      sources: {
        prometheus: {
          expr: 'databricks_job_sla_miss_total{%(queriesSelector)s}',
          legendCustomTemplate: '{{workspace_id}}',
        },
      },
    },

    // Job success rate (24h)
    jobSuccessRate: {
      name: 'Job success rate',
      nameShort: 'Success %',
      description: 'Percentage of successful jobs in the last 24h.',
      type: 'raw',
      unit: 'percentunit',
      sources: {
        prometheus: {
          expr: |||
            sum by (%(aggregationLabels)s) (increase(databricks_job_run_status_total{%(queriesSelector)s, status="SUCCEEDED"}[24h]))
            /
            sum by (%(aggregationLabels)s) (increase(databricks_job_run_status_total{%(queriesSelector)s}[24h]))
          ||| % {
            aggregationLabels: aggregationLabels,
            queriesSelector: '%(queriesSelector)s',
          },
          legendCustomTemplate: 'Success rate',
        },
      },
    },

    // Jobs p95 duration current
    jobsP95Duration: {
      name: 'Jobs p95 duration',
      nameShort: 'p95 duration',
      description: 'Current p95 job duration.',
      type: 'raw',
      unit: 's',
      sources: {
        prometheus: {
          expr: 'databricks_job_run_duration_seconds{%(queriesSelector)s, quantile="0.95"}',
          legendCustomTemplate: 'p95',
        },
      },
    },

    // Pipeline Metrics
    pipelineRunsTotal: {
      name: 'Pipeline runs total',
      nameShort: 'Pipeline runs',
      description: 'DLT / Lakeflow Pipelines executions (per day).',
      type: 'counter',
      unit: 'runs',
      sources: {
        prometheus: {
          expr: 'databricks_pipeline_runs_total{%(queriesSelector)s}',
          legendCustomTemplate: '{{workspace_id}}',
        },
      },
    },

    pipelineRunDuration: {
      name: 'Pipeline run duration',
      nameShort: 'Pipeline duration',
      description: 'Pipeline run duration (p50/p95).',
      type: 'raw',
      unit: 's',
      sources: {
        prometheus: {
          expr: 'databricks_pipeline_run_duration_seconds{%(queriesSelector)s}',
          legendCustomTemplate: '{{workspace_id}} - p{{quantile}}',
        },
      },
    },

    pipelineRunStatusTotal: {
      name: 'Pipeline run status total',
      nameShort: 'Pipeline status',
      description: 'Pipeline run status counts (SUCCESS/FAILED...).',
      type: 'counter',
      unit: 'runs',
      sources: {
        prometheus: {
          expr: 'databricks_pipeline_run_status_total{%(queriesSelector)s}',
          legendCustomTemplate: '{{workspace_id}} - {{status}}',
        },
      },
    },

    pipelineRetryEventsTotal: {
      name: 'Pipeline retry events total',
      nameShort: 'Retry events',
      description: 'Retry/backoff events within pipeline updates.',
      type: 'counter',
      unit: 'events',
      sources: {
        prometheus: {
          expr: 'databricks_pipeline_retry_events_total{%(queriesSelector)s}',
          legendCustomTemplate: '{{workspace_id}}',
        },
      },
    },

    pipelineFreshnessLag: {
      name: 'Pipeline freshness lag',
      nameShort: 'Freshness lag',
      description: 'Data freshness lag vs target watermark (derived).',
      type: 'gauge',
      unit: 's',
      sources: {
        prometheus: {
          expr: 'databricks_pipeline_freshness_lag_seconds{%(queriesSelector)s}',
          legendCustomTemplate: '{{workspace_id}}',
        },
      },
    },

    // Pipeline success rate (24h)
    pipelineSuccessRate: {
      name: 'Pipeline success rate',
      nameShort: 'Success %',
      description: 'Percentage of successful pipelines in the last 24h.',
      type: 'raw',
      unit: 'percentunit',
      sources: {
        prometheus: {
          expr: |||
            sum by (%(aggregationLabels)s) (increase(databricks_pipeline_run_status_total{%(queriesSelector)s, status="SUCCESS"}[24h]))
            /
            sum by (%(aggregationLabels)s) (increase(databricks_pipeline_run_status_total{%(queriesSelector)s}[24h]))
          ||| % {
            aggregationLabels: aggregationLabels,
            queriesSelector: '%(queriesSelector)s',
          },
          legendCustomTemplate: 'Success rate',
        },
      },
    },

    // Pipelines p95 duration current
    pipelinesP95Duration: {
      name: 'Pipelines p95 duration',
      nameShort: 'p95 duration',
      description: 'Current p95 pipeline duration.',
      type: 'raw',
      unit: 's',
      sources: {
        prometheus: {
          expr: 'databricks_pipeline_run_duration_seconds{%(queriesSelector)s, quantile="0.95"}',
          legendCustomTemplate: 'p95',
        },
      },
    },

    // Job/Pipeline throughput
    jobsThroughput: {
      name: 'Jobs throughput',
      nameShort: 'Jobs/min',
      description: 'Job runs throughput over time.',
      type: 'raw',
      unit: 'runs',
      sources: {
        prometheus: {
          expr: 'sum by (' + aggregationLabels + ') (rate(databricks_job_runs_total{%(queriesSelector)s}[5m]))',
          legendCustomTemplate: 'Jobs',
        },
      },
    },

    pipelinesThroughput: {
      name: 'Pipelines throughput',
      nameShort: 'Pipelines/min',
      description: 'Pipeline runs throughput over time.',
      type: 'raw',
      unit: 'runs',
      sources: {
        prometheus: {
          expr: 'sum by (' + aggregationLabels + ') (rate(databricks_pipeline_runs_total{%(queriesSelector)s}[5m]))',
          legendCustomTemplate: 'Pipelines',
        },
      },
    },

    // Failure rates
    jobFailureRate: {
      name: 'Job failure rate',
      nameShort: 'Job failures',
      description: 'Job failure rate by workspace.',
      type: 'raw',
      unit: 'percentunit',
      sources: {
        prometheus: {
          expr: |||
            sum by (workspace_id) (rate(databricks_job_run_status_total{%(queriesSelector)s, status="FAILED"}[1h]))
            /
            sum by (workspace_id) (rate(databricks_job_run_status_total{%(queriesSelector)s}[1h]))
          ||| % { queriesSelector: '%(queriesSelector)s' },
          legendCustomTemplate: '{{workspace_id}}',
        },
      },
    },

    pipelineFailureRate: {
      name: 'Pipeline failure rate',
      nameShort: 'Pipeline failures',
      description: 'Pipeline failure rate by workspace.',
      type: 'raw',
      unit: 'percentunit',
      sources: {
        prometheus: {
          expr: |||
            sum by (workspace_id) (rate(databricks_pipeline_run_status_total{%(queriesSelector)s, status="FAILED"}[1h]))
            /
            sum by (workspace_id) (rate(databricks_pipeline_run_status_total{%(queriesSelector)s}[1h]))
          ||| % { queriesSelector: '%(queriesSelector)s' },
          legendCustomTemplate: '{{workspace_id}}',
        },
      },
    },
  },
}

