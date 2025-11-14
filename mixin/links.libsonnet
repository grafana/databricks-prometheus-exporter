local g = import './g.libsonnet';

{
  new(this): {
    local link = g.dashboard.link,
    local dashboards = this.grafana.dashboards,
    local prefix = this.config.dashboardNamePrefix,
    local uid = this.config.uid,
    local vars = this.grafana.variables,

    databricksOverview:
      link.link.new('Databricks overview', '/d/' + uid + '_overview/databricks-overview')
      + link.link.options.withKeepTime(true)
      + link.link.options.withIncludeVars(true),

    databricksWorkloads:
      link.link.new('Workloads & SQL', '/d/' + uid + '_workloads/databricks-workloads-sql')
      + link.link.options.withKeepTime(true)
      + link.link.options.withIncludeVars(true),

    databricksSqlBi:
      link.link.new('SQL/BI deep dive', '/d/' + uid + '_sqlbi/databricks-sql-bi-deep-dive')
      + link.link.options.withKeepTime(true)
      + link.link.options.withIncludeVars(true),
  },
}

