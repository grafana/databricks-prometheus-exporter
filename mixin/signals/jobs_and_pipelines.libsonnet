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
    jobRunP95Aggregate: {
      name: 'Job run p95 duration (aggregate)',
      nameShort: 'Job p95 (all)',
      description: 'Aggregate p95 job run duration across all jobs (max p95).',
      type: 'raw',
      unit: 's',
      sources: {
        prometheus: {
          expr: 'max(databricks_job_run_duration_seconds{%(queriesSelector)s, quantile="0.95"})',
          legendCustomTemplate: 'Jobs p95',
        },
      },
    },

    jobStatusBreakdown: {
      name: 'Job status breakdown',
      nameShort: 'Job status',
      description: 'Job status counts by job name (adapts to dashboard time range).',
      type: 'raw',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'increase(databricks_job_run_status_total{%(queriesSelector)s}[$__interval:] offset $__interval)',
          legendCustomTemplate: '{{job_name}} - {{status}}',
        },
      },
    },

    jobFailures: {
      name: 'Job failures',
      nameShort: 'Job failures',
      description: 'Number of failed job runs (adapts to dashboard time range).',
      type: 'raw',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'increase(databricks_job_run_status_total{%(queriesSelector)s, status=~"FAILED|ERROR"}[$__interval:] offset $__interval)',
          legendCustomTemplate: '{{workspace_id}}',
        },
      },
    },

    jobSuccessRate: {
      name: 'Job success rate',
      nameShort: 'Success %',
      description: 'Percentage of successful jobs (adapts to dashboard time range).',
      type: 'raw',
      unit: 'percentunit',
      sources: {
        prometheus: {
          expr: |||
            sum (increase(databricks_job_run_status_total{%(queriesSelector)s, status="SUCCEEDED"}[$__interval:] offset $__interval))
            /
            sum (increase(databricks_job_run_status_total{%(queriesSelector)s}[$__interval:] offset $__interval))
          ||| % {
            queriesSelector: '%(queriesSelector)s',
          },
          legendCustomTemplate: 'Success rate',
        },
      },
    },

    pipelineRunP95Aggregate: {
      name: 'Pipeline run p95 duration (aggregate)',
      nameShort: 'Pipeline p95 (all)',
      description: 'Aggregate p95 pipeline run duration across all pipelines (max p95).',
      type: 'raw',
      unit: 's',
      sources: {
        prometheus: {
          expr: 'max(databricks_pipeline_run_duration_seconds{%(queriesSelector)s, quantile="0.95"})',
          legendCustomTemplate: 'Pipelines p95',
        },
      },
    },

    pipelineStatusBreakdown: {
      name: 'Pipeline status breakdown',
      nameShort: 'Pipeline status',
      description: 'Pipeline status counts by pipeline name (adapts to dashboard time range).',
      type: 'raw',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'increase(databricks_pipeline_run_status_total{%(queriesSelector)s}[$__interval:] offset $__interval)',
          legendCustomTemplate: '{{pipeline_name}} - {{status}}',
        },
      },
    },

    pipelineFailures: {
      name: 'Pipeline failures',
      nameShort: 'Pipeline failures',
      description: 'Number of failed pipeline runs (adapts to dashboard time range).',
      type: 'raw',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'increase(databricks_pipeline_run_status_total{%(queriesSelector)s, status="FAILED"}[$__interval:] offset $__interval)',
          legendCustomTemplate: '{{workspace_id}}',
        },
      },
    },

    pipelineSuccessRate: {
      name: 'Pipeline success rate',
      nameShort: 'Success %',
      description: 'Percentage of successful pipelines (adapts to dashboard time range).',
      type: 'raw',
      unit: 'percentunit',
      sources: {
        prometheus: {
          expr: |||
            sum (increase(databricks_pipeline_run_status_total{%(queriesSelector)s, status="COMPLETED"}[$__interval:] offset $__interval))
            /
            sum (increase(databricks_pipeline_run_status_total{%(queriesSelector)s}[$__interval:] offset $__interval))
          ||| % {
            queriesSelector: '%(queriesSelector)s',
          },
          legendCustomTemplate: 'Success rate',
        },
      },
    },

    jobsThroughput: {
      name: 'Jobs throughput',
      nameShort: 'Jobs/min',
      description: 'Job runs per interval (adapts to dashboard time range).',
      type: 'raw',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'sum by (' + aggregationLabels + ') (increase(databricks_job_runs_total{%(queriesSelector)s}[$__interval:] offset $__interval))',
          legendCustomTemplate: 'Jobs',
        },
      },
    },

    pipelinesThroughput: {
      name: 'Pipelines throughput',
      nameShort: 'Pipelines/min',
      description: 'Pipeline runs per interval (adapts to dashboard time range).',
      type: 'raw',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'sum by (' + aggregationLabels + ') (increase(databricks_pipeline_runs_total{%(queriesSelector)s}[$__interval:] offset $__interval))',
          legendCustomTemplate: 'Pipelines',
        },
      },
    },

    jobFailureRate: {
      name: 'Job failure rate',
      nameShort: 'Job failures',
      description: 'Job failure rate by workspace (adapts to dashboard time range).',
      type: 'raw',
      unit: 'percentunit',
      sources: {
        prometheus: {
          expr: |||
            sum by (workspace_id) (increase(databricks_job_run_status_total{%(queriesSelector)s, status=~"FAILED|ERROR"}[$__interval:] offset $__interval))
            /
            sum by (workspace_id) (increase(databricks_job_run_status_total{%(queriesSelector)s}[$__interval:] offset $__interval))
          ||| % { queriesSelector: '%(queriesSelector)s' },
          legendCustomTemplate: '{{workspace_id}} - Jobs',
        },
      },
    },

    pipelineFailureRate: {
      name: 'Pipeline failure rate',
      nameShort: 'Pipeline failures',
      description: 'Pipeline failure rate by workspace (adapts to dashboard time range).',
      type: 'raw',
      unit: 'percentunit',
      sources: {
        prometheus: {
          expr: |||
            sum by (workspace_id) (increase(databricks_pipeline_run_status_total{%(queriesSelector)s, status="FAILED"}[$__interval:] offset $__interval))
            /
            sum by (workspace_id) (increase(databricks_pipeline_run_status_total{%(queriesSelector)s}[$__interval:] offset $__interval))
          ||| % { queriesSelector: '%(queriesSelector)s' },
          legendCustomTemplate: '{{workspace_id}} - Pipelines',
        },
      },
    },

    taskRetries: {
      name: 'Task retries',
      nameShort: 'Task retries',
      description: 'Number of task retries (adapts to dashboard time range).',
      type: 'raw',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'increase(databricks_task_retries_total{%(queriesSelector)s}[$__interval:] offset $__interval)',
          legendCustomTemplate: '{{workspace_id}} - Task retries',
        },
      },
    },

    pipelineRetries: {
      name: 'Pipeline retries',
      nameShort: 'Pipeline retries',
      description: 'Number of pipeline retries (adapts to dashboard time range).',
      type: 'raw',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'increase(databricks_pipeline_retry_events_total{%(queriesSelector)s}[$__interval:] offset $__interval)',
          legendCustomTemplate: '{{workspace_id}} - Pipeline retries',
        },
      },
    },

    jobFailuresByName: {
      name: 'Job failures by job name',
      nameShort: 'Job failures',
      description: 'Number of failed job runs by job name (adapts to dashboard time range).',
      type: 'raw',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'increase(databricks_job_run_status_total{%(queriesSelector)s, status=~"ERROR|FAILED|CANCELLED"}[$__interval:] offset $__interval)',
          legendCustomTemplate: '{{job_name}}',
        },
      },
    },

    jobDurationByName: {
      name: 'Job duration by job name',
      nameShort: 'Job duration',
      description: 'Job p95 duration by job name.',
      type: 'raw',
      unit: 's',
      sources: {
        prometheus: {
          expr: 'databricks_job_run_duration_seconds{%(queriesSelector)s, quantile="0.95"}',
          legendCustomTemplate: '{{job_name}} (p95)',
        },
      },
    },

    pipelineFailuresByName: {
      name: 'Pipeline failures by pipeline name',
      nameShort: 'Pipeline failures',
      description: 'Number of failed pipeline runs by pipeline name (adapts to dashboard time range).',
      type: 'raw',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'increase(databricks_pipeline_run_status_total{%(queriesSelector)s, status="FAILED"}[$__interval:] offset $__interval)',
          legendCustomTemplate: '{{pipeline_name}}',
        },
      },
    },

    pipelineDurationByName: {
      name: 'Pipeline duration by pipeline name',
      nameShort: 'Pipeline duration',
      description: 'Pipeline p95 duration by pipeline name.',
      type: 'raw',
      unit: 's',
      sources: {
        prometheus: {
          expr: 'databricks_pipeline_run_duration_seconds{%(queriesSelector)s, quantile="0.95"}',
          legendCustomTemplate: '{{pipeline_name}} (p95)',
        },
      },
    },

    pipelineFreshnessByName: {
      name: 'Pipeline freshness lag by pipeline name',
      nameShort: 'Freshness lag',
      description: 'Data freshness lag by pipeline name.',
      type: 'gauge',
      unit: 's',
      sources: {
        prometheus: {
          expr: 'databricks_pipeline_freshness_lag_seconds{%(queriesSelector)s}',
          legendCustomTemplate: '{{pipeline_name}}',
        },
      },
    },

    // DEDICATED SIGNALS FOR SPECIFIC PANELS - DO NOT REUSE
    // These are created to avoid breaking panels repeatedly due to sparse counter issues

    topJobsByRunsTableSignal: {
      name: 'Top jobs by runs (table)',
      nameShort: 'Top jobs',
      description: 'Jobs ranked by total run count (for table display only).',
      type: 'raw',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'sum by (workspace_id, job_id, job_name) (databricks_job_runs_total{%(queriesSelector)s})',
          legendCustomTemplate: '{{job_name}}',
        },
      },
    },

    taskRetriesByJobTableSignal: {
      name: 'Task retries by job (table)',
      nameShort: 'Task retries',
      description: 'Task retries aggregated by job (for table display only).',
      type: 'raw',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'sum by (job_id, job_name, task_key) (databricks_task_retries_total{%(queriesSelector)s})',
          legendCustomTemplate: '{{job_name}} / {{task_key}}',
        },
      },
    },

    jobRunsByNameChartSignal: {
      name: 'Job runs by name (chart)',
      nameShort: 'Job runs',
      description: 'Job runs over time by job name - adapts to dashboard time range (for chart display only).',
      type: 'raw',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'increase(databricks_job_runs_total{%(queriesSelector)s}[$__interval:] offset $__interval)',
          legendCustomTemplate: '{{job_name}}',
        },
      },
    },

    topPipelinesByRunsTableSignal: {
      name: 'Top pipelines by runs (table)',
      nameShort: 'Top pipelines',
      description: 'Pipelines ranked by total run count (for table display only).',
      type: 'raw',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'sum by (workspace_id, pipeline_id, pipeline_name) (databricks_pipeline_runs_total{%(queriesSelector)s})',
          legendCustomTemplate: '{{pipeline_name}}',
        },
      },
    },

    pipelineRunsByNameChartSignal: {
      name: 'Pipeline runs by name (chart)',
      nameShort: 'Pipeline runs',
      description: 'Pipeline runs over time by pipeline name - adapts to dashboard time range (for chart display only).',
      type: 'raw',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'increase(databricks_pipeline_runs_total{%(queriesSelector)s}[$__interval:] offset $__interval)',
          legendCustomTemplate: '{{pipeline_name}}',
        },
      },
    },
  },
}
