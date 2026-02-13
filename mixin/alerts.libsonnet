{
  new(this): {
    groups: [
      {
        name: 'DatabricksAlerts',
        rules: [
          {
            alert: 'DatabricksWarnSpendSpike',
            expr: |||
              (
                sum by (job, workspace_id) (databricks_billing_cost_estimate_usd_sliding{} offset 1d)
                - sum by (job, workspace_id) (databricks_billing_cost_estimate_usd_sliding{} offset 2d)
              )
              / sum by (job, workspace_id) (databricks_billing_cost_estimate_usd_sliding{} offset 2d)
              > (%(alertsSpendSpikeWarning)s / 100)
            ||| % this.config,
            'for': '5m',
            labels: {
              severity: 'warning',
            },
            annotations: {
              summary: 'Databricks spend increased significantly day-over-day.',
              description:
                ('Spending on workspace {{$labels.workspace_id}} increased by {{ printf "%%.0f" $value }}%%, ' +
                 'which is above the warning threshold of %(alertsSpendSpikeWarning)s%%. Check cost drivers.') % this.config,
            },
          },
          {
            alert: 'DatabricksCriticalSpendSpike',
            expr: |||
              (
                sum by (job, workspace_id) (databricks_billing_cost_estimate_usd_sliding{} offset 1d)
                - sum by (job, workspace_id) (databricks_billing_cost_estimate_usd_sliding{} offset 2d)
              )
              / sum by (job, workspace_id) (databricks_billing_cost_estimate_usd_sliding{} offset 2d)
              > (%(alertsSpendSpikeCritical)s / 100)
            ||| % this.config,
            'for': '5m',
            labels: {
              severity: 'critical',
            },
            annotations: {
              summary: 'Databricks spend spiked critically day-over-day.',
              description:
                ('Spending on workspace {{$labels.workspace_id}} spiked by {{ printf "%%.0f" $value }}%%, ' +
                 'which is above the critical threshold of %(alertsSpendSpikeCritical)s%%. Immediate investigation required.') % this.config,
            },
          },
          {
            alert: 'DatabricksWarnNoBillingData',
            expr: |||
              (max_over_time(databricks_billing_dbus_sliding{}[6h]) > 0)
              and
              (increase(databricks_billing_dbus_sliding{}[%(alertsNoBillingDataWarningLookback)s]) == 0)
            ||| % this.config,
            'for': '5m',
            labels: {
              severity: 'warning',
            },
            annotations: {
              summary: 'No billing data received from Databricks.',
              description:
                ('No billing data has been received for workspace {{$labels.workspace_id}} in the last %(alertsNoBillingDataWarningLookback)s. ' +
                 'Check SQL Warehouse status and system tables availability.') % this.config,
            },
          },
          {
            alert: 'DatabricksCriticalNoBillingData',
            expr: |||
              (max_over_time(databricks_billing_dbus_sliding{}[6h]) > 0)
              and
              (increase(databricks_billing_dbus_sliding{}[%(alertsNoBillingDataCriticalLookback)s]) == 0)
            ||| % this.config,
            'for': '5m',
            labels: {
              severity: 'critical',
            },
            annotations: {
              summary: 'Critical: No billing data received from Databricks.',
              description:
                ('No billing data has been received for workspace {{$labels.workspace_id}} in the last %(alertsNoBillingDataCriticalLookback)s. ' +
                 'Immediate investigation required - check SQL Warehouse and exporter status.') % this.config,
            },
          },

          // SRE / Platform Persona Alerts (Jobs)
          {
            alert: 'DatabricksWarnJobFailureRate',
            expr: |||
              (
                sum by (job, workspace_id) (increase(databricks_job_run_status_sliding{status="FAILED"}[1h]))
                / sum by (job, workspace_id) (increase(databricks_job_run_status_sliding{}[1h]))
              ) > (%(alertsJobFailureRateWarning)s / 100)
            ||| % this.config,
            'for': '5m',
            labels: {
              severity: 'warning',
            },
            annotations: {
              summary: 'High job failure rate detected.',
              description:
                ('Job failure rate on workspace {{$labels.workspace_id}} is {{ printf "%%.1f" $value }}%%, ' +
                 'which exceeds the warning threshold of %(alertsJobFailureRateWarning)s%%.') % this.config,
            },
          },
          {
            alert: 'DatabricksCriticalJobFailureRate',
            expr: |||
              (
                sum by (job, workspace_id) (increase(databricks_job_run_status_sliding{status="FAILED"}[2h]))
                / sum by (job, workspace_id) (increase(databricks_job_run_status_sliding{}[2h]))
              ) > (%(alertsJobFailureRateCritical)s / 100)
            ||| % this.config,
            'for': '5m',
            labels: {
              severity: 'critical',
            },
            annotations: {
              summary: 'Critical job failure rate detected.',
              description:
                ('Job failure rate on workspace {{$labels.workspace_id}} is {{ printf "%%.1f" $value }}%%, ' +
                 'which exceeds the critical threshold of %(alertsJobFailureRateCritical)s%%. Investigate immediately.') % this.config,
            },
          },
          {
            alert: 'DatabricksWarnJobDurationRegression',
            expr: |||
              (
                databricks_job_run_duration_seconds_sliding{quantile="0.95"}
                / quantile_over_time(0.5, databricks_job_run_duration_seconds_sliding{quantile="0.95"}[7d])
              ) - 1 > (%(alertsJobDurationRegressionWarning)s / 100)
            ||| % this.config,
            'for': '5m',
            labels: {
              severity: 'warning',
            },
            annotations: {
              summary: 'Job duration p95 regression detected.',
              description:
                ('Job p95 duration on workspace {{$labels.workspace_id}} increased by {{ printf "%%.0f" $value }}%% ' +
                 'compared to 7-day median, exceeding warning threshold of %(alertsJobDurationRegressionWarning)s%%.') % this.config,
            },
          },
          {
            alert: 'DatabricksCriticalJobDurationRegression',
            expr: |||
              (
                databricks_job_run_duration_seconds_sliding{quantile="0.95"}
                / quantile_over_time(0.5, databricks_job_run_duration_seconds_sliding{quantile="0.95"}[7d])
              ) - 1 > (%(alertsJobDurationRegressionCritical)s / 100)
            ||| % this.config,
            'for': '5m',
            labels: {
              severity: 'critical',
            },
            annotations: {
              summary: 'Critical job duration p95 regression detected.',
              description:
                ('Job p95 duration on workspace {{$labels.workspace_id}} increased by {{ printf "%%.0f" $value }}%% ' +
                 'compared to 7-day median, exceeding critical threshold of %(alertsJobDurationRegressionCritical)s%%.') % this.config,
            },
          },

          // SRE / Platform Persona Alerts (Pipelines)
          {
            alert: 'DatabricksWarnPipelineFailureRate',
            expr: |||
              (
                sum by (job, workspace_id) (increase(databricks_pipeline_run_status_sliding{status="FAILED"}[1h]))
                / sum by (job, workspace_id) (increase(databricks_pipeline_run_status_sliding{}[1h]))
              ) > (%(alertsPipelineFailureRateWarning)s / 100)
            ||| % this.config,
            'for': '5m',
            labels: {
              severity: 'warning',
            },
            annotations: {
              summary: 'High pipeline failure rate detected.',
              description:
                ('Pipeline failure rate on workspace {{$labels.workspace_id}} is {{ printf "%%.1f" $value }}%%, ' +
                 'which exceeds the warning threshold of %(alertsPipelineFailureRateWarning)s%%.') % this.config,
            },
          },
          {
            alert: 'DatabricksCriticalPipelineFailureRate',
            expr: |||
              (
                sum by (job, workspace_id) (increase(databricks_pipeline_run_status_sliding{status="FAILED"}[1h]))
                / sum by (job, workspace_id) (increase(databricks_pipeline_run_status_sliding{}[1h]))
              ) > (%(alertsPipelineFailureRateCritical)s / 100)
            ||| % this.config,
            'for': '5m',
            labels: {
              severity: 'critical',
            },
            annotations: {
              summary: 'Critical pipeline failure rate detected.',
              description:
                ('Pipeline failure rate on workspace {{$labels.workspace_id}} is {{ printf "%%.1f" $value }}%%, ' +
                 'which exceeds the critical threshold of %(alertsPipelineFailureRateCritical)s%%. Investigate immediately.') % this.config,
            },
          },
          {
            alert: 'DatabricksWarnPipelineDurationRegression',
            expr: |||
              (
                databricks_pipeline_run_duration_seconds_sliding{quantile="0.95"}
                / quantile_over_time(0.5, databricks_pipeline_run_duration_seconds_sliding{quantile="0.95"}[7d])
              ) - 1 > (%(alertsPipelineDurationRegressionWarning)s / 100)
            ||| % this.config,
            'for': '5m',
            labels: {
              severity: 'warning',
            },
            annotations: {
              summary: 'Pipeline duration p95 regression detected.',
              description:
                ('Pipeline p95 duration on workspace {{$labels.workspace_id}} increased by {{ printf "%%.0f" $value }}%% ' +
                 'compared to 7-day median, exceeding warning threshold of %(alertsPipelineDurationRegressionWarning)s%%.') % this.config,
            },
          },
          {
            alert: 'DatabricksCritPipelineDurationHigh',
            expr: |||
              (
                databricks_pipeline_run_duration_seconds_sliding{quantile="0.95"}
                / quantile_over_time(0.5, databricks_pipeline_run_duration_seconds_sliding{quantile="0.95"}[7d])
              ) - 1 > (%(alertsPipelineDurationRegressionCritical)s / 100)
            ||| % this.config,
            'for': '5m',
            labels: {
              severity: 'critical',
            },
            annotations: {
              summary: 'Critical pipeline duration p95 regression detected.',
              description:
                ('Pipeline p95 duration on workspace {{$labels.workspace_id}} increased by {{ printf "%%.0f" $value }}%% ' +
                 'compared to 7-day median, exceeding critical threshold of %(alertsPipelineDurationRegressionCritical)s%%.') % this.config,
            },
          },

          // Analytics/BI Persona Alerts (SQL Warehouse)
          {
            alert: 'DatabricksWarnSqlQueryErrorRate',
            expr: |||
              (
                sum by (job, workspace_id) (rate(databricks_query_errors_sliding{}[30m]))
                / sum by (job, workspace_id) (rate(databricks_queries_sliding{}[30m]))
              ) > (%(alertsSqlQueryErrorRateWarning)s / 100)
            ||| % this.config,
            'for': '1h',
            labels: {
              severity: 'warning',
            },
            annotations: {
              summary: 'High SQL query error rate detected.',
              description:
                ('SQL query error rate on workspace {{$labels.workspace_id}} is {{ printf "%%.1f" $value }}%%, ' +
                 'which exceeds the warning threshold of %(alertsSqlQueryErrorRateWarning)s%%.') % this.config,
            },
          },
          {
            alert: 'DatabricksCriticalSqlQueryErrorRate',
            expr: |||
              (
                sum by (job, workspace_id) (rate(databricks_query_errors_sliding{}[30m]))
                / sum by (job, workspace_id) (rate(databricks_queries_sliding{}[30m]))
              ) > (%(alertsSqlQueryErrorRateCritical)s / 100)
            ||| % this.config,
            'for': '1h',
            labels: {
              severity: 'critical',
            },
            annotations: {
              summary: 'Critical SQL query error rate detected.',
              description:
                ('SQL query error rate on workspace {{$labels.workspace_id}} is {{ printf "%%.1f" $value }}%%, ' +
                 'which exceeds the critical threshold of %(alertsSqlQueryErrorRateCritical)s%%. Investigate immediately.') % this.config,
            },
          },
          {
            alert: 'DatabricksWarnSqlQueryLatencyRegression',
            expr: |||
              (
                databricks_query_duration_seconds_sliding{quantile="0.95"}
                / quantile_over_time(0.5, databricks_query_duration_seconds_sliding{quantile="0.95"}[7d])
              ) - 1 > (%(alertsSqlQueryLatencyRegressionWarning)s / 100)
            ||| % this.config,
            'for': '5m',
            labels: {
              severity: 'warning',
            },
            annotations: {
              summary: 'SQL query latency p95 regression detected.',
              description:
                ('SQL query p95 latency on workspace {{$labels.workspace_id}} increased by {{ printf "%%.0f" $value }}%% ' +
                 'compared to 7-day median, exceeding warning threshold of %(alertsSqlQueryLatencyRegressionWarning)s%%.') % this.config,
            },
          },
          {
            alert: 'DatabricksCritQueryLatencyHigh',
            expr: |||
              (
                databricks_query_duration_seconds_sliding{quantile="0.95"}
                / quantile_over_time(0.5, databricks_query_duration_seconds_sliding{quantile="0.95"}[7d])
              ) - 1 > (%(alertsSqlQueryLatencyRegressionCritical)s / 100)
            ||| % this.config,
            'for': '5m',
            labels: {
              severity: 'critical',
            },
            annotations: {
              summary: 'Critical SQL query latency p95 regression detected.',
              description:
                ('SQL query p95 latency on workspace {{$labels.workspace_id}} increased by {{ printf "%%.0f" $value }}%% ' +
                 'compared to 7-day median, exceeding critical threshold of %(alertsSqlQueryLatencyRegressionCritical)s%%.') % this.config,
            },
          },
        ],
      },
    ],
  },
}
