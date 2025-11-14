local g = import './g.libsonnet';
local commonlib = import 'common-lib/common/main.libsonnet';

{
  new(this)::
    {
      local signals = this.signals,

      // Overview Dashboard Panels - Row 1 Statistics
      yesterdayCostStat:
        commonlib.panels.generic.stat.base.new(
          "Yesterday's cost ($)",
          targets=[signals.overview.yesterdayCost.asTarget()],
          description='Total cost for yesterday (D-1).'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.options.withTextMode('value')
        + g.panel.stat.standardOptions.withUnit('currencyUSD')
        + g.panel.stat.standardOptions.withDecimals(2),

      dodCostDeltaStat:
        commonlib.panels.generic.stat.base.new(
          'DoD cost delta %',
          targets=[signals.overview.dodCostDelta.asTarget()],
          description='Day-over-day cost delta percentage (D-1 vs D-2).'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.options.withTextMode('value')
        + g.panel.stat.standardOptions.withUnit('percentunit')
        + g.panel.stat.standardOptions.withDecimals(1)
        + g.panel.stat.standardOptions.thresholds.withSteps([
          g.panel.stat.thresholdStep.withColor('green'),
          g.panel.stat.thresholdStep.withColor('yellow') + g.panel.stat.thresholdStep.withValue(0.15),
          g.panel.stat.thresholdStep.withColor('red') + g.panel.stat.thresholdStep.withValue(0.25),
        ]),

      yesterdayDbusStat:
        commonlib.panels.generic.stat.base.new(
          'DBUs (D-1)',
          targets=[signals.overview.yesterdayDbus.asTarget()],
          description='Total DBUs consumed yesterday.'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.options.withTextMode('value')
        + g.panel.stat.standardOptions.withUnit('dbu'),

      jobsSuccessRateStat:
        g.panel.gauge.new('Jobs success % (D-1)')
        + g.panel.gauge.queryOptions.withTargets([signals.workloads.jobSuccessRate.asTarget()])
        + g.panel.gauge.panelOptions.withDescription('Percentage of successful jobs in the last 24h.')
        + g.panel.gauge.standardOptions.withUnit('percentunit')
        + g.panel.gauge.standardOptions.withMin(0)
        + g.panel.gauge.standardOptions.withMax(1)
        + g.panel.gauge.standardOptions.thresholds.withSteps([
          g.panel.gauge.thresholdStep.withColor('red'),
          g.panel.gauge.thresholdStep.withColor('yellow') + g.panel.gauge.thresholdStep.withValue(0.8),
          g.panel.gauge.thresholdStep.withColor('green') + g.panel.gauge.thresholdStep.withValue(0.9),
        ])
        + g.panel.gauge.options.withShowThresholdLabels(false)
        + g.panel.gauge.options.withShowThresholdMarkers(true),

      pipelinesSuccessRateStat:
        g.panel.gauge.new('Pipelines success % (D-1)')
        + g.panel.gauge.queryOptions.withTargets([signals.workloads.pipelineSuccessRate.asTarget()])
        + g.panel.gauge.panelOptions.withDescription('Percentage of successful pipelines in the last 24h.')
        + g.panel.gauge.standardOptions.withUnit('percentunit')
        + g.panel.gauge.standardOptions.withMin(0)
        + g.panel.gauge.standardOptions.withMax(1)
        + g.panel.gauge.standardOptions.thresholds.withSteps([
          g.panel.gauge.thresholdStep.withColor('red'),
          g.panel.gauge.thresholdStep.withColor('yellow') + g.panel.gauge.thresholdStep.withValue(0.8),
          g.panel.gauge.thresholdStep.withColor('green') + g.panel.gauge.thresholdStep.withValue(0.9),
        ])
        + g.panel.gauge.options.withShowThresholdLabels(false)
        + g.panel.gauge.options.withShowThresholdMarkers(true),

      sqlErrorRateStat:
        g.panel.gauge.new('SQL error % (D-1)')
        + g.panel.gauge.queryOptions.withTargets([signals.sqlbi.queryErrorRate1h.asTarget()])
        + g.panel.gauge.panelOptions.withDescription('SQL query error percentage in the last hour.')
        + g.panel.gauge.standardOptions.withUnit('percentunit')
        + g.panel.gauge.standardOptions.withMin(0)
        + g.panel.gauge.standardOptions.withMax(0.2)
        + g.panel.gauge.standardOptions.thresholds.withSteps([
          g.panel.gauge.thresholdStep.withColor('green'),
          g.panel.gauge.thresholdStep.withColor('yellow') + g.panel.gauge.thresholdStep.withValue(0.05),
          g.panel.gauge.thresholdStep.withColor('red') + g.panel.gauge.thresholdStep.withValue(0.1),
        ])
        + g.panel.gauge.options.withShowThresholdLabels(false)
        + g.panel.gauge.options.withShowThresholdMarkers(true),

      // Overview Dashboard Panels - Row 2
      costVsDbusChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Cost vs DBUs by SKU',
          targets=[
            signals.overview.costBySku.asTarget(),
          ],
          description='Cost breakdown by SKU over time (7d).'
        )
        + g.panel.timeSeries.standardOptions.withUnit('currencyUSD')
        + g.panel.timeSeries.options.legend.withShowLegend(true)
        + g.panel.timeSeries.options.legend.withPlacement('bottom'),

      globalReliabilityChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Global reliability',
          targets=[
            signals.workloads.jobSuccessRate.asTarget(),
            signals.workloads.pipelineSuccessRate.asTarget(),
            signals.sqlbi.queryErrorRate.asTarget(),
          ],
          description='Jobs/pipelines success % and SQL error % overlay.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('percentunit'),

      // Overview Dashboard Panels - Row 3
      whyCostsChangedTable:
        commonlib.panels.generic.table.base.new(
          'Why costs changed',
          targets=[
            signals.overview.billingCostEstimateUsd.asTableTarget(),
          ],
          description='Usage delta vs Price delta per SKU (D-1 vs D-2).'
        )
        + g.panel.table.standardOptions.withOverridesMixin([
          g.panel.table.fieldOverride.byName.new('Cost')
          + g.panel.table.fieldOverride.byName.withProperty('unit', 'currencyUSD')
          + g.panel.table.fieldOverride.byName.withProperty('decimals', 2),
        ]),

      topWorkspacesByCostTable:
        commonlib.panels.generic.table.base.new(
          'Top workspaces by cost',
          targets=[
            signals.overview.topWorkspacesByCost.asTableTarget(),
          ],
          description='Workspaces with highest costs (yesterday).'
        )
        + g.panel.table.standardOptions.withOverridesMixin([
          g.panel.table.fieldOverride.byName.new('Cost')
          + g.panel.table.fieldOverride.byName.withProperty('unit', 'currencyUSD')
          + g.panel.table.fieldOverride.byName.withProperty('decimals', 2),
        ])
        + g.panel.table.queryOptions.withTransformations([
          { id: 'sortBy', options: { sort: [{ field: 'Value', desc: true }] } },
          { id: 'limit', options: { limitField: 10 } },
        ]),

      topSkusByCostTable:
        commonlib.panels.generic.table.base.new(
          'Top SKUs by cost',
          targets=[
            signals.overview.topSkusByCost.asTableTarget(),
          ],
          description='SKUs with highest costs (yesterday).'
        )
        + g.panel.table.standardOptions.withOverridesMixin([
          g.panel.table.fieldOverride.byName.new('Cost')
          + g.panel.table.fieldOverride.byName.withProperty('unit', 'currencyUSD')
          + g.panel.table.fieldOverride.byName.withProperty('decimals', 2),
        ])
        + g.panel.table.queryOptions.withTransformations([
          { id: 'sortBy', options: { sort: [{ field: 'Value', desc: true }] } },
          { id: 'limit', options: { limitField: 10 } },
        ]),

      // Overview Dashboard Panels - Row 4
      dbusByWorkspaceHeatmap:
        g.panel.heatmap.new('DBUs by workspace × SKU')
        + g.panel.heatmap.queryOptions.withTargets([signals.overview.dbusByWorkspace.asTarget()])
        + g.panel.heatmap.panelOptions.withDescription('DBUs consumed per workspace (heatmap).')
        + g.panel.heatmap.standardOptions.withUnit('dbu'),

      // Overview Dashboard Panels - Row 5
      jobsFailureTrend:
        commonlib.panels.generic.timeSeries.base.new(
          'Jobs failure trend',
          targets=[signals.workloads.jobFailureRate.asTarget()],
          description='Job failure rate over time.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('percentunit'),

      pipelinesFailureTrend:
        commonlib.panels.generic.timeSeries.base.new(
          'Pipelines failure trend',
          targets=[signals.workloads.pipelineFailureRate.asTarget()],
          description='Pipeline failure rate over time.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('percentunit'),

      sqlErrorTrend:
        commonlib.panels.generic.timeSeries.base.new(
          'SQL error trend',
          targets=[signals.sqlbi.queryErrorRate.asTarget()],
          description='SQL query error rate over time.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('percentunit'),

      // Workloads Dashboard Panels - Row 1 Statistics
      jobRuns24hStat:
        commonlib.panels.generic.stat.base.new(
          'Job runs (24h)',
          targets=[signals.workloads.jobsThroughput.asTarget()],
          description='Job runs in the last 24 hours.'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('runs'),

      pipelineRuns24hStat:
        commonlib.panels.generic.stat.base.new(
          'Pipeline runs (24h)',
          targets=[signals.workloads.pipelinesThroughput.asTarget()],
          description='Pipeline runs in the last 24 hours.'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('runs'),

      jobSuccessRate24hStat:
        commonlib.panels.generic.stat.base.new(
          'Job success % (24h)',
          targets=[signals.workloads.jobSuccessRate.asTarget()],
          description='Percentage of successful jobs.'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('percentunit'),

      pipelineSuccessRate24hStat:
        commonlib.panels.generic.stat.base.new(
          'Pipeline success % (24h)',
          targets=[signals.workloads.pipelineSuccessRate.asTarget()],
          description='Percentage of successful pipelines.'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('percentunit'),

      jobP95DurationStat:
        commonlib.panels.generic.stat.base.new(
          'Job p95 duration (min)',
          targets=[signals.workloads.jobsP95Duration.asTarget()],
          description='Current p95 job duration.'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('s'),

      pipelineP95DurationStat:
        commonlib.panels.generic.stat.base.new(
          'Pipeline p95 duration (min)',
          targets=[signals.workloads.pipelinesP95Duration.asTarget()],
          description='Current p95 pipeline duration.'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('s'),

      // Workloads Dashboard Panels - Row 2
      runsThroughputChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Runs throughput (jobs & pipelines)',
          targets=[
            signals.workloads.jobsThroughput.asTarget(),
            signals.workloads.pipelinesThroughput.asTarget(),
          ],
          description='Job and pipeline run rates.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('runs'),

      durationP50P95Chart:
        commonlib.panels.generic.timeSeries.base.new(
          'p50/p95 duration (jobs & pipelines)',
          targets=[
            signals.workloads.jobRunDuration.asTarget(),
            signals.workloads.pipelineRunDuration.asTarget(),
          ],
          description='Duration percentiles for jobs and pipelines.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('s'),

      // Workloads Dashboard Panels - Row 3
      failureRateByWorkspaceChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Failure rate by workspace',
          targets=[
            signals.workloads.jobFailureRate.asTarget(),
            signals.workloads.pipelineFailureRate.asTarget(),
          ],
          description='Failure rate breakdown by workspace.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('percentunit'),

      retriesVsFailuresChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Retries vs failures',
          targets=[
            signals.workloads.taskRetriesTotal.asTarget(),
            signals.workloads.pipelineRetryEventsTotal.asTarget(),
          ],
          description='Retries and failures over time.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('short'),

      durationRegressionChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Duration regression vs 7-day median',
          targets=[
            signals.workloads.jobsP95Duration.asTarget(),
            signals.workloads.pipelinesP95Duration.asTarget(),
          ],
          description='Duration regression showing +30/+60% thresholds.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('s'),

      // Workloads Dashboard Panels - Row 4
      topFailingJobsTable:
        commonlib.panels.generic.table.base.new(
          'Top failing jobs/pipelines (24h)',
          targets=[
            signals.workloads.jobRunStatusTotal.asTableTarget(),
          ],
          description='Jobs and pipelines with most failures.'
        )
        + g.panel.table.queryOptions.withTransformations([
          { id: 'filterByValue', options: { filters: [{ fieldName: 'status', config: { id: 'equal', options: { value: 'FAILED' } } }] } },
          { id: 'sortBy', options: { sort: [{ field: 'Value', desc: true }] } },
          { id: 'limit', options: { limitField: 10 } },
        ]),

      // Workloads Dashboard Panels - Row 5
      jobsStatusBreakdownChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Jobs status breakdown',
          targets=[signals.workloads.jobRunStatusTotal.asTarget()],
          description='Job status: success/failed/canceled.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('runs')
        + g.panel.timeSeries.fieldConfig.defaults.custom.stacking.withMode('normal'),

      pipelinesStatusBreakdownChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Pipelines status breakdown',
          targets=[signals.workloads.pipelineRunStatusTotal.asTarget()],
          description='Pipeline status: success/failed.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('runs')
        + g.panel.timeSeries.fieldConfig.defaults.custom.stacking.withMode('normal'),

      // SQL/BI Dashboard Panels - Row 1 Statistics
      queriesTotal1hStat:
        commonlib.panels.generic.stat.base.new(
          'Queries total (1h)',
          targets=[signals.sqlbi.queryLoad1h.asTarget()],
          description='Total queries in the last hour.'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('queries'),

      queriesTotal24hStat:
        commonlib.panels.generic.stat.base.new(
          'Queries total (24h)',
          targets=[signals.sqlbi.queryLoad24h.asTarget()],
          description='Total queries in the last 24 hours.'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('queries'),

      queryErrorRate1hStat:
        commonlib.panels.generic.stat.base.new(
          'Query error % (1h)',
          targets=[signals.sqlbi.queryErrorRate1h.asTarget()],
          description='Error percentage in the last hour.'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('percentunit'),

      queryP95LatencyStat:
        commonlib.panels.generic.stat.base.new(
          'p95 latency (s)',
          targets=[signals.sqlbi.queryP95Latency.asTarget()],
          description='Current p95 query latency.'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('s'),

      concurrencyCurrentStat:
        commonlib.panels.generic.stat.base.new(
          'Concurrency (now)',
          targets=[signals.sqlbi.concurrencyCurrent.asTarget()],
          description='Current concurrent queries.'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('queries'),

      concurrencyBaseline7dStat:
        commonlib.panels.generic.stat.base.new(
          'Concurrency (7d p95)',
          targets=[signals.sqlbi.concurrencyBaseline.asTarget()],
          description='Baseline concurrency (7d p95).'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('queries'),

      dodQueriesDiffStat:
        commonlib.panels.generic.stat.base.new(
          'DoD queries diff %',
          targets=[signals.sqlbi.dodQueriesDelta.asTarget()],
          description='Day-over-day query count change.'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('percentunit'),

      dodP95LatencyDiffStat:
        commonlib.panels.generic.stat.base.new(
          'DoD p95 latency diff %',
          targets=[signals.sqlbi.dodP95LatencyDelta.asTarget()],
          description='Day-over-day p95 latency change.'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('percentunit'),

      // SQL/BI Dashboard Panels - Row 2
      queryLoadChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Query load (rate)',
          targets=[signals.sqlbi.queryRate.asTarget()],
          description='Query execution rate over time.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('qps'),

      latencyP50P95Chart:
        commonlib.panels.generic.timeSeries.base.new(
          'Latency p50/p95',
          targets=[
            signals.sqlbi.queryP50Latency.asTarget(),
            signals.sqlbi.queryP95Latency.asTarget(),
          ],
          description='Query latency percentiles with +30/+60% thresholds.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('s'),

      // SQL/BI Dashboard Panels - Row 3
      sqlErrorRateChart:
        commonlib.panels.generic.timeSeries.base.new(
          'SQL error rate',
          targets=[signals.sqlbi.queryErrorRate.asTarget()],
          description='SQL error rate with warn/crit bands.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('percentunit'),

      concurrencyVsLatencyChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Concurrency vs latency',
          targets=[
            signals.sqlbi.concurrencyCurrent.asTarget(),
            signals.sqlbi.queryP95Latency.asTarget(),
          ],
          description='Overlay showing saturation effect.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('short'),

      queriesByWorkspaceChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Queries by workspace',
          targets=[signals.sqlbi.queriesByWorkspace.asTarget()],
          description='Query volume by workspace.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('queries'),

      // SQL/BI Dashboard Panels - Row 4
      topSlowQueriesTable:
        commonlib.panels.generic.table.base.new(
          'Top slow queries by workspace',
          targets=[
            signals.sqlbi.queryDuration.asTableTarget(),
          ],
          description='Slowest queries normalized by workspace.'
        )
        + g.panel.table.standardOptions.withOverridesMixin([
          g.panel.table.fieldOverride.byName.new('Duration')
          + g.panel.table.fieldOverride.byName.withProperty('unit', 's')
          + g.panel.table.fieldOverride.byName.withProperty('decimals', 2),
        ])
        + g.panel.table.queryOptions.withTransformations([
          { id: 'sortBy', options: { sort: [{ field: 'Value', desc: true }] } },
          { id: 'limit', options: { limitField: 10 } },
        ]),

      topErroringWorkspacesTable:
        commonlib.panels.generic.table.base.new(
          'Top erroring workspaces',
          targets=[
            signals.sqlbi.queryErrorsTotal.asTableTarget(),
          ],
          description='Workspaces with most errors.'
        )
        + g.panel.table.queryOptions.withTransformations([
          { id: 'sortBy', options: { sort: [{ field: 'Value', desc: true }] } },
          { id: 'limit', options: { limitField: 10 } },
        ]),

      // SQL/BI Dashboard Panels - Row 5
      queryLatencyDistribution:
        g.panel.histogram.new('Query latency distribution')
        + g.panel.histogram.queryOptions.withTargets([signals.sqlbi.queryDuration.asTarget()])
        + g.panel.histogram.panelOptions.withDescription('Query latency distribution histogram.')
        + g.panel.histogram.standardOptions.withUnit('s'),

      medianLatencyVs7dChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Median latency vs 7-day',
          targets=[
            signals.sqlbi.queryP50Latency.asTarget(),
          ],
          description='Median latency ratio line.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('s'),

      dodChangesChart:
        commonlib.panels.generic.timeSeries.base.new(
          'DoD changes (queries, error %, p95)',
          targets=[
            signals.sqlbi.dodQueriesDelta.asTarget(),
            signals.sqlbi.queryErrorRate1h.asTarget(),
            signals.sqlbi.dodP95LatencyDelta.asTarget(),
          ],
          description='Multiple lines showing daily changes.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('percentunit'),
    },
}

