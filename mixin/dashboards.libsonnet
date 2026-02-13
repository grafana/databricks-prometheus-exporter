local g = import './g.libsonnet';
local commonlib = import 'common-lib/common/main.libsonnet';

{
  local root = self,
  new(this)::
    local prefix = this.config.dashboardNamePrefix;
    local links = this.grafana.links;
    local tags = this.config.dashboardTags;
    local uid = g.util.string.slugify(this.config.uid);
    local vars = this.grafana.variables;
    local annotations = this.grafana.annotations;
    local refresh = this.config.dashboardRefresh;
    local period = this.config.dashboardPeriod;
    local timezone = this.config.dashboardTimezone;
    {
      'databricks-overview.json':
        g.dashboard.new(prefix + ' Overview')
        + g.dashboard.withDescription('Executive summary dashboard showing costs, billing, and overall reliability metrics for Databricks.')
        + g.dashboard.withPanels(
          g.util.panel.resolveCollapsedFlagOnRows(
            g.util.grid.wrapPanels(
              [
                this.grafana.rows.overviewStatistics,
                this.grafana.rows.overviewCharts,
                this.grafana.rows.overviewDecomposition,
                this.grafana.rows.overviewTrends,
              ]
            )
          )
        ) + root.applyCommon(
          vars.multiInstance,
          uid + '_overview',
          tags,
          links { databricksOverview:: {} },
          annotations,
          timezone,
          refresh,
          period,
        ),

      'databricks-jobs-and-pipelines.json':
        g.dashboard.new(prefix + ' Jobs & Pipelines')
        + g.dashboard.withDescription('Deep dive into jobs and pipelines reliability, throughput, and performance metrics.')
        + g.dashboard.withPanels(
          g.util.panel.resolveCollapsedFlagOnRows(
            g.util.grid.wrapPanels(
              [
                this.grafana.rows.workloadsStatistics,
                this.grafana.rows.workloadsThroughput,
                this.grafana.rows.workloadsReliability,
                this.grafana.rows.workloadsStatusBreakdown,
                this.grafana.rows.workloadsJobDrilldown,
                this.grafana.rows.workloadsPipelineDrilldown,
              ]
            )
          )
        ) + root.applyCommon(
          vars.multiInstance,
          uid + '_jobs_pipelines',
          tags,
          links { databricksJobsPipelines:: {} },
          annotations,
          timezone,
          refresh,
          period,
        ),

      'databricks-warehouses-and-queries.json':
        g.dashboard.new(prefix + ' Warehouses & Queries')
        + g.dashboard.withDescription('Comprehensive view of SQL warehouse performance, query latency, errors, and concurrency.')
        + g.dashboard.withPanels(
          g.util.panel.resolveCollapsedFlagOnRows(
            g.util.grid.wrapPanels(
              [
                this.grafana.rows.sqlbiStatistics,
                this.grafana.rows.sqlbiLoadAndLatency,
                this.grafana.rows.sqlbiErrorsAndConcurrency,
                this.grafana.rows.sqlbiTopWarehouses,
                this.grafana.rows.sqlbiDistribution,
                this.grafana.rows.sqlbiWarehouseDrilldown,
              ]
            )
          )
        ) + root.applyCommon(
          vars.multiInstance,
          uid + '_warehouses_queries',
          tags,
          links { databricksWarehousesQueries:: {} },
          annotations,
          timezone,
          refresh,
          period,
        ),
    },
  applyCommon(vars, uid, tags, links, annotations, timezone, refresh, period):
    g.dashboard.withTags(tags)
    + g.dashboard.withUid(uid)
    + g.dashboard.withLinks(std.objectValues(links))
    + g.dashboard.withTimezone(timezone)
    + g.dashboard.withRefresh(refresh)
    + g.dashboard.time.withFrom(period)
    + g.dashboard.withVariables(vars)
    + g.dashboard.withAnnotations(std.objectValues(annotations)),

}
