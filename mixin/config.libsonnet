{
  local this = self,
  filteringSelector: '',
  groupLabels: ['job', 'workspace_id'],
  uid: 'databricks',
  instanceLabels: ['instance'],
  enableLokiLogs: false,

  // dashboard config
  dashboardTags: [this.uid + '-mixin'],
  dashboardNamePrefix: 'Databricks',
  dashboardPeriod: 'now-1h',
  dashboardTimezone: 'default',
  dashboardRefresh: '30s',
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
    jobsAndPipelines: (import './signals/jobs_and_pipelines.libsonnet')(this),
    warehousesAndQueries: (import './signals/warehouses_and_queries.libsonnet')(this),
  },
}
