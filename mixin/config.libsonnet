{
  local this = self,
  filteringSelector: 'job="integrations/databricks"',
  groupLabels: ['job', 'workspace_id'],
  uid: 'databricks',
  instanceLabels: ['instance'],
  enableLokiLogs: false,

  // dashboard config
  dashboardTags: [this.uid + '-mixin'],
  dashboardNamePrefix: 'Databricks',
  dashboardPeriod: 'now-7d',
  dashboardTimezone: 'default',
  dashboardRefresh: '30m',
  metricsSource: 'prometheus',

  // for alerts - Finance Persona
  alertsSpendSpikeWarning: '25',  // % DoD increase
  alertsSpendSpikeCritical: '50',  // % DoD increase
  alertsNoBillingDataWarningLookback: '2h',
  alertsNoBillingDataCriticalLookback: '4h',

  // for alerts - SRE / Platform Persona (Jobs & Pipelines)
  alertsJobFailureRateWarning: '10',  // %
  alertsJobFailureRateCritical: '20',  // %
  alertsPipelineFailureRateWarning: '10',  // %
  alertsPipelineFailureRateCritical: '20',  // %
  alertsJobDurationRegressionWarning: '30',  // % vs 7-day median
  alertsJobDurationRegressionCritical: '60',  // % vs 7-day median
  alertsPipelineDurationRegressionWarning: '30',  // % vs 7-day median
  alertsPipelineDurationRegressionCritical: '60',  // % vs 7-day median

  // for alerts - Analytics/BI Persona (SQL Warehouse)
  alertsSqlQueryErrorRateWarning: '5',  // %
  alertsSqlQueryErrorRateCritical: '10',  // %
  alertsSqlQueryLatencyRegressionWarning: '30',  // % vs 7-day median
  alertsSqlQueryLatencyRegressionCritical: '60',  // % vs 7-day median

  signals+: {
    overview: (import './signals/overview.libsonnet')(this),
    workloads: (import './signals/workloads.libsonnet')(this),
    sqlbi: (import './signals/sqlbi.libsonnet')(this),
  },
}

