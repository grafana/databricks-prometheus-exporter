local g = import './g.libsonnet';

{
  new(this):
    {
      local panels = this.grafana.panels,

      // Overview Dashboard Rows
      overviewStatistics:
        g.panel.row.new('Overview')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.yesterdayCostStat { gridPos: { w: 4, h: 4 } },
            panels.dodCostDeltaStat { gridPos: { w: 4, h: 4 } },
            panels.totalDbusConsumedStat { gridPos: { w: 4, h: 4 } },
            panels.jobsSuccessRateStat { gridPos: { w: 4, h: 4 } },
            panels.pipelinesSuccessRateStat { gridPos: { w: 4, h: 4 } },
            panels.sqlErrorRateStat { gridPos: { w: 4, h: 4 } },
          ]
        ),

      overviewCharts:
        g.panel.row.new('Cost & reliability trends')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.costBySkuChart { gridPos: { w: 8, h: 8 } },
            panels.costPerDbuBySkuChart { gridPos: { w: 8, h: 8 } },
            panels.globalReliabilityChart { gridPos: { w: 8, h: 8 } },
          ]
        ),

      overviewDecomposition:
        g.panel.row.new('Cost decomposition')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.topWorkspacesByCostTable { gridPos: { w: 8, h: 6 } },
            panels.topSkusByCostTable { gridPos: { w: 8, h: 6 } },
            panels.topDbusByWorkspaceTable { gridPos: { w: 8, h: 6 } },
          ]
        ),

      overviewTrends:
        g.panel.row.new('Reliability trends')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.jobsFailureTrend { gridPos: { w: 8, h: 6 } },
            panels.pipelinesFailureTrend { gridPos: { w: 8, h: 6 } },
            panels.sqlErrorTrend { gridPos: { w: 8, h: 6 } },
          ]
        ),

      // Workloads (Jobs & Pipelines) Dashboard Rows
      workloadsStatistics:
        g.panel.row.new('Workloads overview')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.jobRunsStat { gridPos: { w: 4, h: 4 } },
            panels.pipelineRunsStat { gridPos: { w: 4, h: 4 } },
            panels.jobSuccessRateStat { gridPos: { w: 4, h: 4 } },
            panels.pipelineSuccessRateStat { gridPos: { w: 4, h: 4 } },
            panels.jobP95DurationStat { gridPos: { w: 4, h: 4 } },
            panels.pipelineP95DurationStat { gridPos: { w: 4, h: 4 } },
          ]
        ),

      workloadsThroughput:
        g.panel.row.new('Throughput & duration')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.runsThroughputChart { gridPos: { w: 8, h: 8 } },
            panels.jobP95DurationChart { gridPos: { w: 8, h: 8 } },
            panels.pipelineP95DurationChart { gridPos: { w: 8, h: 8 } },
          ]
        ),

      workloadsReliability:
        g.panel.row.new('Reliability & stability')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.failureRateByWorkspaceChart { gridPos: { w: 12, h: 8 } },
            panels.retriesVsFailuresChart { gridPos: { w: 12, h: 8 } },
          ]
        ),

      workloadsStatusBreakdown:
        g.panel.row.new('Status breakdown')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.jobsStatusBreakdownChart { gridPos: { w: 12, h: 10 } },
            panels.pipelinesStatusBreakdownChart { gridPos: { w: 12, h: 10 } },
          ]
        ),

      workloadsJobDrilldown:
        g.panel.row.new('Jobs drill-down (by job name)')
        + g.panel.row.withCollapsed(true)
        + g.panel.row.withPanels(
          [
            panels.topJobsByRunsTable { gridPos: { w: 8, h: 8 } },
            panels.topJobsByDurationTable { gridPos: { w: 8, h: 8 } },
            panels.topJobsByFailuresTable { gridPos: { w: 8, h: 8 } },
            panels.jobRunsByNameChart { gridPos: { w: 12, h: 10 } },
            panels.jobDurationByNameChart { gridPos: { w: 12, h: 10 } },
          ]
        ),

      workloadsPipelineDrilldown:
        g.panel.row.new('Pipelines drill-down (by pipeline name)')
        + g.panel.row.withCollapsed(true)
        + g.panel.row.withPanels(
          [
            panels.pipelineRunsByNameChart { gridPos: { w: 24, h: 10 } },
            panels.topPipelinesByRunsTable { gridPos: { w: 8, h: 8 } },
            panels.topPipelinesByDurationTable { gridPos: { w: 8, h: 8 } },
            panels.topPipelinesByFailuresTable { gridPos: { w: 8, h: 8 } },
            panels.pipelineFreshnessByNameTable { gridPos: { w: 10, h: 10 } },
            panels.pipelineDurationByNameChart { gridPos: { w: 14, h: 10 } },
          ]
        ),

      // SQL (Warehouses & Queries) Dashboard Rows
      sqlbiStatistics:
        g.panel.row.new('SQL warehouse overview')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.queriesTotalStat { gridPos: { w: 3, h: 4 } },
            panels.queryErrorRateStat { gridPos: { w: 3, h: 4 } },
            panels.queryP95LatencyStat { gridPos: { w: 3, h: 4 } },
            panels.concurrencyCurrentStat { gridPos: { w: 3, h: 4 } },
            panels.concurrencyBaseline7dStat { gridPos: { w: 4, h: 4 } },
            panels.dodQueriesDiffStat { gridPos: { w: 4, h: 4 } },
            panels.dodP95LatencyDiffStat { gridPos: { w: 4, h: 4 } },
          ]
        ),

      sqlbiLoadAndLatency:
        g.panel.row.new('Load & latency')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.queryLoadChart { gridPos: { w: 12, h: 8 } },
            panels.latencyP50P95Chart { gridPos: { w: 12, h: 8 } },
          ]
        ),

      sqlbiErrorsAndConcurrency:
        g.panel.row.new('Errors, concurrency & workspace distribution')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.sqlErrorRateChart { gridPos: { w: 8, h: 8 } },
            panels.concurrencyVsLatencyChart { gridPos: { w: 8, h: 8 } },
            panels.queriesByWorkspaceChart { gridPos: { w: 8, h: 8 } },
          ]
        ),

      sqlbiTopWarehouses:
        g.panel.row.new('Top warehouses')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.topWarehousesByQueriesTable { gridPos: { w: 8, h: 8 } },
            panels.topWarehousesByErrorsTable { gridPos: { w: 8, h: 8 } },
            panels.topWarehousesByLatencyTable { gridPos: { w: 8, h: 8 } },
          ]
        ),

      sqlbiDistribution:
        g.panel.row.new('Distribution & trends')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.queryLatencyDistribution { gridPos: { w: 8, h: 8 } },
            panels.medianLatencyVs7dChart { gridPos: { w: 8, h: 8 } },
            panels.dodChangesChart { gridPos: { w: 8, h: 8 } },
          ]
        ),

      sqlbiWarehouseDrilldown:
        g.panel.row.new('Performance by warehouse')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.queriesByWarehouseChart { gridPos: { w: 6, h: 8 } },
            panels.queryErrorsByWarehouseChart { gridPos: { w: 6, h: 8 } },
            panels.queryLatencyByWarehouseChart { gridPos: { w: 6, h: 8 } },
            panels.concurrencyByWarehouseChart { gridPos: { w: 6, h: 8 } },
          ]
        ),
    },
}
