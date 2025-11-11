local g = import './g.libsonnet';

{
  new(this): {
    local link = g.dashboard.link,
    local dashboards = this.grafana.dashboards,
    local prefix = this.config.dashboardNamePrefix,
    local uid = this.config.uid,
    local vars = this.grafana.variables,

    databricksOverview:
      link.link.new('Overview', '/d/' + uid + '_overview/databricks-overview')
      + link.link.options.withKeepTime(true)
      + link.link.options.withIncludeVars(true),

    databricksJobsPipelines:
      link.link.new('Jobs & Pipelines', '/d/' + uid + '_jobs_pipelines/databricks-jobs-and-pipelines')
      + link.link.options.withKeepTime(true)
      + link.link.options.withIncludeVars(true),

    databricksWarehousesQueries:
      link.link.new('Warehouses & Queries', '/d/' + uid + '_warehouses_queries/databricks-warehouses-and-queries')
      + link.link.options.withKeepTime(true)
      + link.link.options.withIncludeVars(true),
  },
}
