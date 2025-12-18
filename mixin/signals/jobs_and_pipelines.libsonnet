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
      type: 'gauge',
      unit: 's',
      sources: {
        prometheus: {
          expr: 'max(databricks_job_run_duration_seconds_sliding{%(queriesSelector)s, quantile="0.95"})',
          exprWrappers: [['last_over_time(', '[30m:])']],
          legendCustomTemplate: 'Jobs p95',
        },
      },
    },

    jobStatusBreakdown: {
      name: 'Job status breakdown',
      nameShort: 'Job status',
      description: 'Job status counts by job name.',
      type: 'gauge',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'databricks_job_run_status_sliding{%(queriesSelector)s}',
          exprWrappers: [['last_over_time(', '[30m:])']],
          legendCustomTemplate: '{{job_name}} - {{status}}',
        },
      },
    },

    jobFailures: {
      name: 'Job failures',
      nameShort: 'Job failures',
      description: 'Number of failed job runs.',
      type: 'gauge',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'databricks_job_run_status_sliding{%(queriesSelector)s, status=~"FAILED|ERROR"}',
          exprWrappers: [['last_over_time(', '[30m:])']],
          legendCustomTemplate: '{{workspace_id}}',
        },
      },
    },

    jobSuccessRate: {
      name: 'Job success rate',
      nameShort: 'Success %',
      description: 'Percentage of successful jobs.',
      type: 'raw',
      unit: 'percentunit',
      sources: {
        prometheus: {
          expr: |||
            last_over_time((
              sum(databricks_job_run_status_sliding{%(queriesSelector)s, status="SUCCEEDED"})
              /
              sum(databricks_job_run_status_sliding{%(queriesSelector)s})
            )[30m:])
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
      type: 'gauge',
      unit: 's',
      sources: {
        prometheus: {
          expr: 'max(databricks_pipeline_run_duration_seconds_sliding{%(queriesSelector)s, quantile="0.95"})',
          exprWrappers: [['last_over_time(', '[30m:])']],
          legendCustomTemplate: 'Pipelines p95',
        },
      },
    },

    pipelineStatusBreakdown: {
      name: 'Pipeline status breakdown',
      nameShort: 'Pipeline status',
      description: 'Pipeline status counts by pipeline name.',
      type: 'gauge',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'databricks_pipeline_run_status_sliding{%(queriesSelector)s}',
          exprWrappers: [['last_over_time(', '[30m:])']],
          legendCustomTemplate: '{{pipeline_name}} - {{status}}',
        },
      },
    },

    pipelineFailures: {
      name: 'Pipeline failures',
      nameShort: 'Pipeline failures',
      description: 'Number of failed pipeline runs.',
      type: 'gauge',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'databricks_pipeline_run_status_sliding{%(queriesSelector)s, status="FAILED"}',
          exprWrappers: [['last_over_time(', '[30m:])']],
          legendCustomTemplate: '{{workspace_id}}',
        },
      },
    },

    pipelineSuccessRate: {
      name: 'Pipeline success rate',
      nameShort: 'Success %',
      description: 'Percentage of successful pipelines.',
      type: 'raw',
      unit: 'percentunit',
      sources: {
        prometheus: {
          expr: |||
            last_over_time((
              sum(databricks_pipeline_run_status_sliding{%(queriesSelector)s, status="COMPLETED"})
              /
              sum(databricks_pipeline_run_status_sliding{%(queriesSelector)s})
            )[30m:])
          ||| % {
            queriesSelector: '%(queriesSelector)s',
          },
          legendCustomTemplate: 'Success rate',
        },
      },
    },

    jobsThroughput: {
      name: 'Jobs throughput',
      nameShort: 'Jobs',
      description: 'Total job runs in the sliding window.',
      type: 'gauge',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'sum by (' + aggregationLabels + ') (databricks_job_runs_sliding{%(queriesSelector)s})',
          exprWrappers: [['last_over_time(', '[30m:])']],
          legendCustomTemplate: 'Jobs',
        },
      },
    },

    pipelinesThroughput: {
      name: 'Pipelines throughput',
      nameShort: 'Pipelines',
      description: 'Total pipeline runs in the sliding window.',
      type: 'gauge',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'sum by (' + aggregationLabels + ') (databricks_pipeline_runs_sliding{%(queriesSelector)s})',
          exprWrappers: [['last_over_time(', '[30m:])']],
          legendCustomTemplate: 'Pipelines',
        },
      },
    },

    jobFailureRate: {
      name: 'Job failure rate',
      nameShort: 'Job failures',
      description: 'Job failure rate by workspace.',
      type: 'raw',
      unit: 'percentunit',
      sources: {
        prometheus: {
          expr: |||
            last_over_time((
              sum by (workspace_id) (databricks_job_run_status_sliding{%(queriesSelector)s, status=~"FAILED|ERROR"})
              /
              sum by (workspace_id) (databricks_job_run_status_sliding{%(queriesSelector)s})
            )[30m:])
          ||| % { queriesSelector: '%(queriesSelector)s' },
          legendCustomTemplate: '{{workspace_id}} - Jobs',
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
            last_over_time((
              sum by (workspace_id) (databricks_pipeline_run_status_sliding{%(queriesSelector)s, status="FAILED"})
              /
              sum by (workspace_id) (databricks_pipeline_run_status_sliding{%(queriesSelector)s})
            )[30m:])
          ||| % { queriesSelector: '%(queriesSelector)s' },
          legendCustomTemplate: '{{workspace_id}} - Pipelines',
        },
      },
    },

    taskRetries: {
      name: 'Task retries',
      nameShort: 'Task retries',
      description: 'Number of task retries.',
      type: 'gauge',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'databricks_task_retries_sliding{%(queriesSelector)s}',
          exprWrappers: [['last_over_time(', '[30m:])']],
          legendCustomTemplate: '{{workspace_id}} - Task retries',
        },
      },
    },

    pipelineRetries: {
      name: 'Pipeline retries',
      nameShort: 'Pipeline retries',
      description: 'Number of pipeline retries.',
      type: 'gauge',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'databricks_pipeline_retry_events_sliding{%(queriesSelector)s}',
          exprWrappers: [['last_over_time(', '[30m:])']],
          legendCustomTemplate: '{{workspace_id}} - Pipeline retries',
        },
      },
    },

    jobFailuresByName: {
      name: 'Job failures by job name',
      nameShort: 'Job failures',
      description: 'Number of failed job runs by job name.',
      type: 'gauge',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'databricks_job_run_status_sliding{%(queriesSelector)s, status=~"ERROR|FAILED|CANCELLED"}',
          exprWrappers: [['last_over_time(', '[30m:])']],
          legendCustomTemplate: '{{job_name}}',
        },
      },
    },

    jobDurationByName: {
      name: 'Job duration by job name',
      nameShort: 'Job duration',
      description: 'Job p95 duration by job name.',
      type: 'gauge',
      unit: 's',
      sources: {
        prometheus: {
          expr: 'databricks_job_run_duration_seconds_sliding{%(queriesSelector)s, quantile="0.95"}',
          exprWrappers: [['last_over_time(', '[30m:])']],
          legendCustomTemplate: '{{job_name}} (p95)',
        },
      },
    },

    pipelineFailuresByName: {
      name: 'Pipeline failures by pipeline name',
      nameShort: 'Pipeline failures',
      description: 'Number of failed pipeline runs by pipeline name.',
      type: 'gauge',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'databricks_pipeline_run_status_sliding{%(queriesSelector)s, status="FAILED"}',
          exprWrappers: [['last_over_time(', '[30m:])']],
          legendCustomTemplate: '{{pipeline_name}}',
        },
      },
    },

    pipelineDurationByName: {
      name: 'Pipeline duration by pipeline name',
      nameShort: 'Pipeline duration',
      description: 'Pipeline p95 duration by pipeline name.',
      type: 'gauge',
      unit: 's',
      sources: {
        prometheus: {
          expr: 'databricks_pipeline_run_duration_seconds_sliding{%(queriesSelector)s, quantile="0.95"}',
          exprWrappers: [['last_over_time(', '[30m:])']],
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
          expr: 'databricks_pipeline_freshness_lag_seconds_sliding{%(queriesSelector)s}',
          exprWrappers: [['last_over_time(', '[30m:])']],
          legendCustomTemplate: '{{pipeline_name}}',
        },
      },
    },

    // DEDICATED SIGNALS FOR SPECIFIC PANELS - DO NOT REUSE
    // These are created to avoid breaking panels repeatedly due to sparse counter issues

    topJobsByRunsTableSignal: {
      name: 'Top jobs by runs (table)',
      nameShort: 'Top jobs',
      description: 'Jobs ranked by total run count (sliding window gauge).',
      type: 'gauge',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'sum by (workspace_id, job_id, job_name) (databricks_job_runs_sliding{%(queriesSelector)s})',
          exprWrappers: [['last_over_time(', '[30m:])']],
          legendCustomTemplate: '{{job_name}}',
        },
      },
    },

    taskRetriesByJobTableSignal: {
      name: 'Task retries by job (table)',
      nameShort: 'Task retries',
      description: 'Task retries aggregated by job (sliding window gauge).',
      type: 'gauge',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'sum by (job_id, job_name, task_key) (databricks_task_retries_sliding{%(queriesSelector)s})',
          exprWrappers: [['last_over_time(', '[30m:])']],
          legendCustomTemplate: '{{job_name}} / {{task_key}}',
        },
      },
    },

    jobRunsByNameChartSignal: {
      name: 'Job runs by name (chart)',
      nameShort: 'Job runs',
      description: 'Job runs by job name.',
      type: 'gauge',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'databricks_job_runs_sliding{%(queriesSelector)s}',
          exprWrappers: [['last_over_time(', '[30m:])']],
          legendCustomTemplate: '{{job_name}}',
        },
      },
    },

    topPipelinesByRunsTableSignal: {
      name: 'Top pipelines by runs (table)',
      nameShort: 'Top pipelines',
      description: 'Pipelines ranked by total run count (sliding window gauge).',
      type: 'gauge',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'sum by (workspace_id, pipeline_id, pipeline_name) (databricks_pipeline_runs_sliding{%(queriesSelector)s})',
          exprWrappers: [['last_over_time(', '[30m:])']],
          legendCustomTemplate: '{{pipeline_name}}',
        },
      },
    },

    pipelineRunsByNameChartSignal: {
      name: 'Pipeline runs by name (chart)',
      nameShort: 'Pipeline runs',
      description: 'Pipeline runs by pipeline name.',
      type: 'gauge',
      unit: 'short',
      sources: {
        prometheus: {
          expr: 'databricks_pipeline_runs_sliding{%(queriesSelector)s}',
          exprWrappers: [['last_over_time(', '[30m:])']],
          legendCustomTemplate: '{{pipeline_name}}',
        },
      },
    },
  },
}
