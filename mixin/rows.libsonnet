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
            panels.yesterdayDbusStat { gridPos: { w: 4, h: 4 } },
            panels.jobsSuccessRateStat { gridPos: { w: 4, h: 4 } },
            panels.pipelinesSuccessRateStat { gridPos: { w: 4, h: 4 } },
            panels.sqlErrorRateStat { gridPos: { w: 4, h: 4 } },
          ]
        ),

      overviewCharts:
        g.panel.row.new('Cost vs usage trends')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.costVsDbusChart { gridPos: { w: 12, h: 8 } },
            panels.globalReliabilityChart { gridPos: { w: 12, h: 8 } },
          ]
        ),

      overviewDecomposition:
        g.panel.row.new('Cost decomposition')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.whyCostsChangedTable { gridPos: { w: 8, h: 6 } },
            panels.topWorkspacesByCostTable { gridPos: { w: 8, h: 6 } },
            panels.topSkusByCostTable { gridPos: { w: 8, h: 6 } },
          ]
        ),

      overviewHeatmap:
        g.panel.row.new('Resource usage heatmap')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.dbusByWorkspaceHeatmap { gridPos: { w: 24, h: 8 } },
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

      // Workloads Dashboard Rows
      workloadsStatistics:
        g.panel.row.new('Workloads overview')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.jobRuns24hStat { gridPos: { w: 4, h: 4 } },
            panels.pipelineRuns24hStat { gridPos: { w: 4, h: 4 } },
            panels.jobSuccessRate24hStat { gridPos: { w: 4, h: 4 } },
            panels.pipelineSuccessRate24hStat { gridPos: { w: 4, h: 4 } },
            panels.jobP95DurationStat { gridPos: { w: 4, h: 4 } },
            panels.pipelineP95DurationStat { gridPos: { w: 4, h: 4 } },
          ]
        ),

      workloadsThroughput:
        g.panel.row.new('Throughput & duration')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.runsThroughputChart { gridPos: { w: 12, h: 8 } },
            panels.durationP50P95Chart { gridPos: { w: 12, h: 8 } },
          ]
        ),

      workloadsReliability:
        g.panel.row.new('Reliability & stability')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.failureRateByWorkspaceChart { gridPos: { w: 8, h: 8 } },
            panels.retriesVsFailuresChart { gridPos: { w: 8, h: 8 } },
            panels.durationRegressionChart { gridPos: { w: 8, h: 8 } },
          ]
        ),

      workloadsFailures:
        g.panel.row.new('Failure analysis')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.topFailingJobsTable { gridPos: { w: 24, h: 8 } },
          ]
        ),

      workloadsStatusBreakdown:
        g.panel.row.new('Status breakdown')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.jobsStatusBreakdownChart { gridPos: { w: 12, h: 8 } },
            panels.pipelinesStatusBreakdownChart { gridPos: { w: 12, h: 8 } },
          ]
        ),

      // SQL/BI Dashboard Rows
      sqlbiStatistics:
        g.panel.row.new('SQL warehouse overview')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.queriesTotal1hStat { gridPos: { w: 3, h: 4 } },
            panels.queriesTotal24hStat { gridPos: { w: 3, h: 4 } },
            panels.queryErrorRate1hStat { gridPos: { w: 3, h: 4 } },
            panels.queryP95LatencyStat { gridPos: { w: 3, h: 4 } },
            panels.concurrencyCurrentStat { gridPos: { w: 3, h: 4 } },
            panels.concurrencyBaseline7dStat { gridPos: { w: 3, h: 4 } },
            panels.dodQueriesDiffStat { gridPos: { w: 3, h: 4 } },
            panels.dodP95LatencyDiffStat { gridPos: { w: 3, h: 4 } },
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

      sqlbiTopQueries:
        g.panel.row.new('Top slow queries & errors')
        + g.panel.row.withCollapsed(false)
        + g.panel.row.withPanels(
          [
            panels.topSlowQueriesTable { gridPos: { w: 12, h: 8 } },
            panels.topErroringWorkspacesTable { gridPos: { w: 12, h: 8 } },
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
    },
}

