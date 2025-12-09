local g = import './g.libsonnet';
local commonlib = import 'common-lib/common/main.libsonnet';

{
  new(this)::
    {
      local signals = this.signals,

      // Overview Dashboard Panels - Row 1 Statistics
      yesterdayCostStat:
        commonlib.panels.generic.stat.base.new(
          'Total cost (24h window)',
          targets=[signals.overview.yesterdayCost.asTarget()],
          description='Total cost estimate for the past 24 hours. This is a rolling 1-day window from exporter.'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.options.withTextMode('value')
        + g.panel.stat.standardOptions.withUnit('currencyUSD')
        + g.panel.stat.standardOptions.withDecimals(2),

      dodCostDeltaStat:
        commonlib.panels.generic.stat.base.new(
          'Cost change % (previous 24h)',
          targets=[signals.overview.dodCostDelta.asTarget()],
          description='Percentage change comparing current 24h window to the 24h window from yesterday. Green = low growth, Yellow = moderate growth (>15%), Red = high growth (>25%).'
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

      totalDbusConsumedStat:
        commonlib.panels.generic.stat.base.new(
          'Total DBUs (24h window)',
          targets=[signals.overview.totalDbusConsumed.asTarget()],
          description='Total DBUs consumed over the past 24 hours. This is a rolling 1-day window from the exporter.'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.options.withTextMode('value')
        + g.panel.stat.standardOptions.withUnit('short'),

      jobsSuccessRateStat:
        g.panel.gauge.new('Jobs success %')
        + g.panel.gauge.queryOptions.withTargets([signals.jobsAndPipelines.jobSuccessRate.asTarget()])
        + g.panel.gauge.panelOptions.withDescription('Percentage of successful jobs (adapts to dashboard time range).')
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
        g.panel.gauge.new('Pipelines success %')
        + g.panel.gauge.queryOptions.withTargets([signals.jobsAndPipelines.pipelineSuccessRate.asTarget()])
        + g.panel.gauge.panelOptions.withDescription('Percentage of successful pipelines (adapts to dashboard time range).')
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
        g.panel.gauge.new('SQL error %')
        + g.panel.gauge.queryOptions.withTargets([signals.warehousesAndQueries.queryErrorRateAggregate.asTarget()])
        + g.panel.gauge.panelOptions.withDescription('SQL query error percentage (adapts to dashboard time range).')
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
      costBySkuChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Cost by SKU',
          targets=[
            signals.overview.costBySku.asTarget(),
          ],
          description='Cost breakdown by SKU over time (24h rolling window from exporter).'
        )
        + g.panel.timeSeries.standardOptions.withUnit('currencyUSD')
        + g.panel.timeSeries.options.legend.withShowLegend(true)
        + g.panel.timeSeries.options.legend.withPlacement('bottom'),

      costPerDbuBySkuChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Cost per DBU by SKU',
          targets=[
            signals.overview.costPerDbuBySku.asTarget(),
          ],
          description='Cost efficiency metric - shows how much each DBU costs for each SKU. Lower is more cost-efficient (24h rolling window).'
        )
        + g.panel.timeSeries.standardOptions.withUnit('currencyUSD')
        + g.panel.timeSeries.options.legend.withShowLegend(true)
        + g.panel.timeSeries.options.legend.withPlacement('bottom'),

      globalReliabilityChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Global reliability',
          targets=[
            signals.jobsAndPipelines.jobSuccessRate.asTarget()
            + g.query.prometheus.withLegendFormat('Jobs success rate'),
            signals.jobsAndPipelines.pipelineSuccessRate.asTarget()
            + g.query.prometheus.withLegendFormat('Pipelines success rate'),
            signals.warehousesAndQueries.queryErrorRate.asTarget()
            + g.query.prometheus.withLegendFormat('Query errors rate'),
          ],
          description='Jobs and pipelines success rate (%) and SQL error rate (%) overlay.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('percentunit'),

      // Overview Dashboard Panels - Row 3
      topWorkspacesByCostTable:
        commonlib.panels.generic.table.base.new(
          'Top workspaces by cost',
          targets=[
            signals.overview.topWorkspacesByCost.asTableTarget(),
          ],
          description='Workspaces with highest costs (24h rolling window from exporter).'
        )
        + g.panel.table.standardOptions.withOverridesMixin([
          g.panel.table.fieldOverride.byName.new('Cost')
          + g.panel.table.fieldOverride.byName.withProperty('unit', 'currencyUSD')
          + g.panel.table.fieldOverride.byName.withProperty('decimals', 2),
        ])
        + g.panel.table.queryOptions.withTransformations([
          {
            id: 'organize',
            options: {
              excludeByName: {
                Time: true,
                job: true,
                instance: true,
              },
              indexByName: {
                workspace_id: 0,
                Value: 1,
              },
              renameByName: {
                workspace_id: 'Workspace ID',
                Value: 'Cost',
              },
            },
          },
          { id: 'sortBy', options: { sort: [{ field: 'Cost', desc: true }] } },
          { id: 'limit', options: { limitField: 10 } },
        ]),

      topSkusByCostTable:
        commonlib.panels.generic.table.base.new(
          'Top SKUs by cost',
          targets=[
            signals.overview.topSkusByCost.asTableTarget(),
          ],
          description='SKUs with highest costs (24h rolling window from exporter).'
        )
        + g.panel.table.standardOptions.withOverridesMixin([
          g.panel.table.fieldOverride.byName.new('Total Cost')
          + g.panel.table.fieldOverride.byName.withProperty('unit', 'currencyUSD')
          + g.panel.table.fieldOverride.byName.withProperty('decimals', 2),
        ])
        + g.panel.table.queryOptions.withTransformations([
          {
            id: 'organize',
            options: {
              excludeByName: {
                Time: true,
                job: true,
                instance: true,
              },
              indexByName: {
                workspace_id: 0,
                sku_name: 1,
                Value: 2,
              },
              renameByName: {
                workspace_id: 'Workspace ID',
                sku_name: 'SKU Name',
                Value: 'Total Cost',
              },
            },
          },
          { id: 'sortBy', options: { sort: [{ field: 'Total Cost', desc: true }] } },
          { id: 'limit', options: { limitField: 10 } },
        ]),

      // Overview Dashboard Panels - Row 4
      topDbusByWorkspaceTable:
        commonlib.panels.generic.table.base.new(
          'DBUs by workspace and SKU',
          targets=[signals.overview.billingDbusTotal.asTableTarget()],
          description='DBUs consumed broken down by workspace and SKU (24h rolling window from exporter).'
        )
        + g.panel.table.standardOptions.withOverrides([
          g.panel.table.fieldOverride.byName.new('DBUs')
          + g.panel.table.fieldOverride.byName.withPropertiesFromOptions(
            g.panel.table.standardOptions.withUnit('short')
          ),
        ])
        + g.panel.table.queryOptions.withTransformations([
          {
            id: 'organize',
            options: {
              excludeByName: {
                Time: true,
                job: true,
                instance: true,
                __name__: true,
              },
              indexByName: {
                workspace_id: 0,
                sku_name: 1,
                Value: 2,
              },
              renameByName: {
                workspace_id: 'Workspace ID',
                sku_name: 'SKU Name',
                Value: 'DBUs',
              },
            },
          },
          { id: 'sortBy', options: { sort: [{ field: 'DBUs', desc: true }] } },
          { id: 'limit', options: { limitField: 10 } },
        ]),

      // Overview Dashboard Panels - Row 5
      jobsFailureTrend:
        commonlib.panels.generic.timeSeries.base.new(
          'Jobs failure trend',
          targets=[signals.jobsAndPipelines.jobFailureRate.asTarget()],
          description='Job failure rate over time.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('percentunit')
        + g.panel.timeSeries.standardOptions.withNoValue('0'),

      pipelinesFailureTrend:
        commonlib.panels.generic.timeSeries.base.new(
          'Pipelines failure trend',
          targets=[signals.jobsAndPipelines.pipelineFailureRate.asTarget()],
          description='Pipeline failure rate over time.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('percentunit')
        + g.panel.timeSeries.standardOptions.withNoValue('0'),

      sqlErrorTrend:
        commonlib.panels.generic.timeSeries.base.new(
          'SQL error trend',
          targets=[signals.warehousesAndQueries.queryErrorRate.asTarget()],
          description='SQL query error rate over time.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('percentunit')
        + g.panel.timeSeries.standardOptions.withNoValue('0'),

      // Jobs & Pipelines Dashboard Panels - Row 1 Statistics
      jobRunsStat:
        commonlib.panels.generic.stat.base.new(
          'Job runs',
          targets=[signals.jobsAndPipelines.jobsThroughput.asTarget()],
          description='Job runs (adapts to dashboard time range).'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('short'),

      pipelineRunsStat:
        commonlib.panels.generic.stat.base.new(
          'Pipeline runs',
          targets=[signals.jobsAndPipelines.pipelinesThroughput.asTarget()],
          description='Pipeline runs (adapts to dashboard time range).'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('short'),

      jobSuccessRateStat:
        commonlib.panels.generic.stat.base.new(
          'Job success %',
          targets=[signals.jobsAndPipelines.jobSuccessRate.asTarget()],
          description='Percentage of successful jobs (adapts to dashboard time range).'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('percentunit'),

      pipelineSuccessRateStat:
        commonlib.panels.generic.stat.base.new(
          'Pipeline success %',
          targets=[signals.jobsAndPipelines.pipelineSuccessRate.asTarget()],
          description='Percentage of successful pipelines (adapts to dashboard time range).'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('percentunit'),

      jobP95DurationStat:
        commonlib.panels.generic.stat.base.new(
          'Job p95 duration (max)',
          targets=[signals.jobsAndPipelines.jobRunP95Aggregate.asTarget()],
          description='Maximum p95 job duration across all jobs.'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('s'),

      pipelineP95DurationStat:
        commonlib.panels.generic.stat.base.new(
          'Pipeline p95 duration (max)',
          targets=[signals.jobsAndPipelines.pipelineRunP95Aggregate.asTarget()],
          description='Maximum p95 pipeline duration across all pipelines.'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('s'),

      // Jobs & Pipelines Dashboard Panels - Row 2
      runsThroughputChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Total runs',
          targets=[
            signals.jobsAndPipelines.jobsThroughput.asTarget(),
            signals.jobsAndPipelines.pipelinesThroughput.asTarget(),
          ],
          description='Number of job and pipeline runs (adapts to dashboard time range).'
        )
        + g.panel.timeSeries.standardOptions.withUnit('short'),

      // Jobs & Pipelines Dashboard Panels - Row 3
      failureRateByWorkspaceChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Failure rate by workspace',
          targets=[
            signals.jobsAndPipelines.jobFailureRate.asTarget(),
            signals.jobsAndPipelines.pipelineFailureRate.asTarget(),
          ],
          description='Job and pipeline failure rates by workspace (adapts to dashboard time range).'
        )
        + g.panel.timeSeries.standardOptions.withUnit('percentunit'),

      retriesVsFailuresChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Retries',
          targets=[
            signals.jobsAndPipelines.taskRetries.asTarget(),
            signals.jobsAndPipelines.pipelineRetries.asTarget(),
          ],
          description='Number of task and pipeline retries (adapts to dashboard time range).'
        )
        + g.panel.timeSeries.standardOptions.withUnit('short')
        + g.panel.timeSeries.options.legend.withAsTable(true)
        + g.panel.timeSeries.options.legend.withPlacement('right'),

      jobP95DurationChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Job p95 duration',
          targets=[
            signals.jobsAndPipelines.jobRunP95Aggregate.asTarget(),
          ],
          description='Maximum p95 duration across all jobs over time.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('s')
        + g.panel.timeSeries.fieldConfig.defaults.custom.withShowPoints('never')
        + g.panel.timeSeries.options.legend.withDisplayMode('list')
        + g.panel.timeSeries.options.legend.withPlacement('bottom'),

      pipelineP95DurationChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Pipeline p95 duration',
          targets=[
            signals.jobsAndPipelines.pipelineRunP95Aggregate.asTarget(),
          ],
          description='Maximum p95 duration across all pipelines over time.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('s')
        + g.panel.timeSeries.fieldConfig.defaults.custom.withShowPoints('never')
        + g.panel.timeSeries.options.legend.withDisplayMode('list')
        + g.panel.timeSeries.options.legend.withPlacement('bottom'),

      // Jobs & Pipelines Dashboard Panels - Row 4
      jobsStatusBreakdownChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Jobs status breakdown',
          targets=[signals.jobsAndPipelines.jobStatusBreakdown.asTarget()],
          description='Job status breakdown showing counts (adapts to dashboard time range).'
        )
        + g.panel.timeSeries.standardOptions.withUnit('short')
        + g.panel.timeSeries.options.legend.withAsTable(true)
        + g.panel.timeSeries.options.legend.withPlacement('right')
        + g.panel.timeSeries.fieldConfig.defaults.custom.stacking.withMode('normal'),

      pipelinesStatusBreakdownChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Pipelines status breakdown',
          targets=[signals.jobsAndPipelines.pipelineStatusBreakdown.asTarget()],
          description='Pipeline status breakdown showing counts (adapts to dashboard time range).'
        )
        + g.panel.timeSeries.standardOptions.withUnit('short')
        + g.panel.timeSeries.options.legend.withAsTable(true)
        + g.panel.timeSeries.options.legend.withPlacement('right')
        + g.panel.timeSeries.fieldConfig.defaults.custom.stacking.withMode('normal'),

      // Warehouses & Queries Dashboard Panels - Row 1 Statistics
      queriesTotalStat:
        commonlib.panels.generic.stat.base.new(
          'Total queries',
          targets=[signals.warehousesAndQueries.queryLoad.asTarget()],
          description='Total queries executed (adapts to dashboard time range).'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('short'),

      queryErrorRateStat:
        commonlib.panels.generic.stat.base.new(
          'Query error %',
          targets=[signals.warehousesAndQueries.queryErrorRateAggregate.asTarget()],
          description='Error percentage (adapts to dashboard time range).'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('percentunit'),

      queryP95LatencyStat:
        commonlib.panels.generic.stat.base.new(
          'p95 latency (max)',
          targets=[signals.warehousesAndQueries.queryP95Latency.asTarget()],
          description='Maximum p95 query latency across all warehouses.'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('s'),

      concurrencyCurrentStat:
        commonlib.panels.generic.stat.base.new(
          'Max concurrent queries (now)',
          targets=[signals.warehousesAndQueries.concurrencyCurrent.asTarget()],
          description='Maximum concurrent queries across all warehouses.'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('short'),

      concurrencyBaseline7dStat:
        commonlib.panels.generic.stat.base.new(
          'Max concurrent queries (7d p95)',
          targets=[signals.warehousesAndQueries.concurrencyBaseline.asTarget()],
          description='Maximum baseline concurrency over 7 days (p95).'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('short'),

      dodQueriesDiffStat:
        commonlib.panels.generic.stat.base.new(
          'DoD queries diff %',
          targets=[signals.warehousesAndQueries.dodQueriesDelta.asTarget()],
          description='Day-over-day query count change.'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('percentunit'),

      dodP95LatencyDiffStat:
        commonlib.panels.generic.stat.base.new(
          'DoD p95 latency diff %',
          targets=[signals.warehousesAndQueries.dodP95LatencyDelta.asTarget()],
          description='Day-over-day p95 latency change.'
        )
        + g.panel.stat.options.withGraphMode('none')
        + g.panel.stat.standardOptions.withUnit('percentunit'),

      // SQL/BI Dashboard Panels - Row 2
      queryLoadChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Query load (rate)',
          targets=[signals.warehousesAndQueries.queryLoad.asTarget()],
          description='Query execution rate over time.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('short'),

      latencyP50P95Chart:
        commonlib.panels.generic.timeSeries.base.new(
          'Query latency: p50 vs. p95',
          targets=[
            signals.warehousesAndQueries.queryP50Latency.asTarget(),
            signals.warehousesAndQueries.queryP95Latency.asTarget(),
          ],
          description='Query latency: p50 vs. p95.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('s'),

      // SQL/BI Dashboard Panels - Row 3
      sqlErrorRateChart:
        commonlib.panels.generic.timeSeries.base.new(
          'SQL error rate',
          targets=[signals.warehousesAndQueries.queryErrorRate.asTarget()],
          description='SQL error rate with warn/crit bands.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('percentunit'),

      concurrencyVsLatencyChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Concurrency vs latency',
          targets=[
            signals.warehousesAndQueries.concurrencyCurrent.asTarget(),
            signals.warehousesAndQueries.queryP95Latency.asTarget(),
          ],
          description='Overlay showing saturation effect.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('short'),

      queriesByWorkspaceChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Queries by workspace',
          targets=[signals.warehousesAndQueries.queriesByWorkspace.asTarget()],
          description='Query volume by workspace.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('short'),

      // SQL/BI Dashboard Panels - Row 4
      queryLatencyDistribution:
        g.panel.histogram.new('Query latency distribution per warehouse')
        + g.panel.histogram.queryOptions.withTargets([signals.warehousesAndQueries.queryDuration.asTarget()])
        + g.panel.histogram.panelOptions.withDescription('Query latency distribution histogram per warehouse.')
        + g.panel.histogram.standardOptions.withUnit('s'),

      medianLatencyVs7dChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Query latency: current median vs. 7 days median',
          targets=[
            signals.warehousesAndQueries.queryP50Latency.asTarget(),
            signals.warehousesAndQueries.queryP50Latency7dBaseline.asTarget(),
          ],
          description='Compares current median (p50) query latency to a 7-day rolling median baseline. X-axis shows the selected time range. For each point in time, the baseline is calculated from the previous 7 days. When the current line rises above the baseline, queries are slower than the recent 7-day average.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('s'),

      dodChangesChart:
        commonlib.panels.generic.timeSeries.base.new(
          'DoD changes (queries, error %, p95)',
          targets=[
            signals.warehousesAndQueries.dodQueriesDelta.asTarget(),
            signals.warehousesAndQueries.queryErrorRateAggregate.asTarget(),
            signals.warehousesAndQueries.dodP95LatencyDelta.asTarget(),
          ],
          description='Multiple lines showing daily changes.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('percentunit'),

      // Detailed drill-down panels - Jobs
      topJobsByRunsTable:
        commonlib.panels.generic.table.base.new(
          'Top jobs by runs',
          targets=[signals.jobsAndPipelines.topJobsByRunsTableSignal.asTableTarget()],
          description='Jobs with most runs (all time).'
        )
        + g.panel.table.standardOptions.withOverrides([
          g.panel.table.fieldOverride.byName.new('Total Runs')
          + g.panel.table.fieldOverride.byName.withPropertiesFromOptions(
            g.panel.table.standardOptions.withUnit('short')
          ),
        ])
        + g.panel.table.queryOptions.withTransformations([
          {
            id: 'organize',
            options: {
              excludeByName: {
                __name__: true,
                instance: true,
                job: true,
                Time: true,
              },
              indexByName: {
                workspace_id: 0,
                job_id: 1,
                job_name: 2,
                Value: 3,
              },
              renameByName: {
                workspace_id: 'Workspace ID',
                job_id: 'Job ID',
                job_name: 'Job Name',
                Value: 'Total Runs',
              },
            },
          },
          { id: 'sortBy', options: { sort: [{ field: 'Total Runs', desc: true }] } },
          { id: 'limit', options: { limitField: 10 } },
        ]),

      topJobsByDurationTable:
        commonlib.panels.generic.table.base.new(
          'Top jobs by p95 duration',
          targets=[signals.jobsAndPipelines.jobDurationByName.asTableTarget()],
          description='Jobs with longest p95 duration.'
        )
        + g.panel.table.standardOptions.withOverrides([
          g.panel.table.fieldOverride.byName.new('p95 Duration (seconds)')
          + g.panel.table.fieldOverride.byName.withPropertiesFromOptions(
            g.panel.table.standardOptions.withUnit('s')
          ),
        ])
        + g.panel.table.queryOptions.withTransformations([
          {
            id: 'organize',
            options: {
              excludeByName: {
                Time: true,
                __name__: true,
                instance: true,
                job: true,
                quantile: true,
              },
              indexByName: {
                workspace_id: 0,
                job_id: 1,
                job_name: 2,
                Value: 3,
              },
              renameByName: {
                workspace_id: 'Workspace ID',
                job_id: 'Job ID',
                job_name: 'Job Name',
                Value: 'p95 Duration (seconds)',
              },
            },
          },
          { id: 'sortBy', options: { sort: [{ field: 'p95 Duration (seconds)', desc: true }] } },
          { id: 'limit', options: { limitField: 10 } },
        ]),

      topJobsByFailuresTable:
        commonlib.panels.generic.table.base.new(
          'Top jobs by failures',
          targets=[signals.jobsAndPipelines.jobFailuresByName.asTableTarget()],
          description='Jobs with most failures (adapts to dashboard time range).'
        )
        + g.panel.table.standardOptions.withOverrides([
          g.panel.table.fieldOverride.byName.new('Total Failures')
          + g.panel.table.fieldOverride.byName.withPropertiesFromOptions(
            g.panel.table.standardOptions.withUnit('short')
          ),
        ])
        + g.panel.table.queryOptions.withTransformations([
          {
            id: 'organize',
            options: {
              excludeByName: {
                Time: true,
                __name__: true,
                instance: true,
                job: true,
              },
              indexByName: {
                workspace_id: 0,
                job_id: 1,
                job_name: 2,
                status: 3,
                Value: 4,
              },
              renameByName: {
                workspace_id: 'Workspace ID',
                job_id: 'Job ID',
                job_name: 'Job Name',
                status: 'Status',
                Value: 'Total Failures',
              },
            },
          },
          { id: 'sortBy', options: { sort: [{ field: 'Total Failures', desc: true }] } },
          { id: 'limit', options: { limitField: 10 } },
        ]),

      jobRunsByNameChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Job runs by name',
          targets=[signals.jobsAndPipelines.jobRunsByNameChartSignal.asTarget()],
          description='Job run counts over time by job name (adapts to dashboard time range).'
        )
        + g.panel.timeSeries.standardOptions.withUnit('short')
        + g.panel.timeSeries.options.legend.withDisplayMode('table')
        + g.panel.timeSeries.options.legend.withPlacement('right'),

      jobDurationByNameChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Job p95 duration by name',
          targets=[signals.jobsAndPipelines.jobDurationByName.asTarget()],
          description='Job p95 duration over time by job name.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('s')
        + g.panel.timeSeries.options.legend.withDisplayMode('table')
        + g.panel.timeSeries.options.legend.withPlacement('right'),

      // Detailed drill-down panels - Pipelines
      topPipelinesByRunsTable:
        commonlib.panels.generic.table.base.new(
          'Top pipelines by runs',
          targets=[signals.jobsAndPipelines.topPipelinesByRunsTableSignal.asTableTarget()],
          description='Pipelines with most runs (all time).'
        )
        + g.panel.table.standardOptions.withOverrides([
          g.panel.table.fieldOverride.byName.new('Total Runs')
          + g.panel.table.fieldOverride.byName.withPropertiesFromOptions(
            g.panel.table.standardOptions.withUnit('short')
          ),
        ])
        + g.panel.table.queryOptions.withTransformations([
          {
            id: 'organize',
            options: {
              excludeByName: {
                Time: true,
                __name__: true,
                instance: true,
                job: true,
              },
              indexByName: {
                workspace_id: 0,
                pipeline_id: 1,
                pipeline_name: 2,
                Value: 3,
              },
              renameByName: {
                workspace_id: 'Workspace ID',
                pipeline_id: 'Pipeline ID',
                pipeline_name: 'Pipeline Name',
                Value: 'Total Runs',
              },
            },
          },
          { id: 'sortBy', options: { sort: [{ field: 'Total Runs', desc: true }] } },
          { id: 'limit', options: { limitField: 10 } },
        ]),

      topPipelinesByDurationTable:
        commonlib.panels.generic.table.base.new(
          'Top pipelines by p95 duration',
          targets=[signals.jobsAndPipelines.pipelineDurationByName.asTableTarget()],
          description='Pipelines with longest p95 duration.'
        )
        + g.panel.table.standardOptions.withOverrides([
          g.panel.table.fieldOverride.byName.new('p95 Duration (seconds)')
          + g.panel.table.fieldOverride.byName.withPropertiesFromOptions(
            g.panel.table.standardOptions.withUnit('s')
          ),
        ])
        + g.panel.table.queryOptions.withTransformations([
          {
            id: 'organize',
            options: {
              excludeByName: {
                Time: true,
                __name__: true,
                instance: true,
                job: true,
                quantile: true,
              },
              indexByName: {
                workspace_id: 0,
                pipeline_id: 1,
                pipeline_name: 2,
                Value: 3,
              },
              renameByName: {
                workspace_id: 'Workspace ID',
                pipeline_id: 'Pipeline ID',
                pipeline_name: 'Pipeline Name',
                Value: 'p95 Duration (seconds)',
              },
            },
          },
          { id: 'sortBy', options: { sort: [{ field: 'p95 Duration (seconds)', desc: true }] } },
          { id: 'limit', options: { limitField: 10 } },
        ]),

      topPipelinesByFailuresTable:
        commonlib.panels.generic.table.base.new(
          'Top pipelines by failures',
          targets=[signals.jobsAndPipelines.pipelineFailuresByName.asTableTarget()],
          description='Pipelines with most failures (adapts to dashboard time range).'
        )
        + g.panel.table.standardOptions.withOverrides([
          g.panel.table.fieldOverride.byName.new('Total Failures')
          + g.panel.table.fieldOverride.byName.withPropertiesFromOptions(
            g.panel.table.standardOptions.withUnit('short')
          ),
        ])
        + g.panel.table.queryOptions.withTransformations([
          {
            id: 'organize',
            options: {
              excludeByName: {
                Time: true,
                __name__: true,
                instance: true,
                job: true,
              },
              indexByName: {
                workspace_id: 0,
                pipeline_id: 1,
                pipeline_name: 2,
                status: 3,
                Value: 4,
              },
              renameByName: {
                workspace_id: 'Workspace ID',
                pipeline_id: 'Pipeline ID',
                pipeline_name: 'Pipeline Name',
                status: 'Status',
                Value: 'Total Failures',
              },
            },
          },
          { id: 'sortBy', options: { sort: [{ field: 'Total Failures', desc: true }] } },
          { id: 'limit', options: { limitField: 10 } },
        ]),

      pipelineFreshnessByNameTable:
        commonlib.panels.generic.table.base.new(
          'Pipeline freshness lag',
          targets=[signals.jobsAndPipelines.pipelineFreshnessByName.asTableTarget()],
          description='Data freshness lag by pipeline.'
        )
        + g.panel.table.standardOptions.withOverrides([
          g.panel.table.fieldOverride.byName.new('Freshness Lag (seconds)')
          + g.panel.table.fieldOverride.byName.withPropertiesFromOptions(
            g.panel.table.standardOptions.withUnit('s')
          ),
        ])
        + g.panel.table.queryOptions.withTransformations([
          {
            id: 'organize',
            options: {
              excludeByName: {
                Time: true,
                __name__: true,
                instance: true,
                job: true,
              },
              indexByName: {
                workspace_id: 0,
                pipeline_id: 1,
                pipeline_name: 2,
                Value: 3,
              },
              renameByName: {
                workspace_id: 'Workspace ID',
                pipeline_id: 'Pipeline ID',
                pipeline_name: 'Pipeline Name',
                Value: 'Freshness Lag (seconds)',
              },
            },
          },
          { id: 'sortBy', options: { sort: [{ field: 'Freshness Lag (seconds)', desc: true }] } },
          { id: 'limit', options: { limitField: 10 } },
        ]),

      pipelineRunsByNameChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Pipeline runs by name',
          targets=[signals.jobsAndPipelines.pipelineRunsByNameChartSignal.asTarget()],
          description='Pipeline run counts over time by pipeline name (adapts to dashboard time range).'
        )
        + g.panel.timeSeries.standardOptions.withUnit('short')
        + g.panel.timeSeries.options.legend.withDisplayMode('table')
        + g.panel.timeSeries.options.legend.withPlacement('right'),

      pipelineDurationByNameChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Pipeline p95 duration by name',
          targets=[signals.jobsAndPipelines.pipelineDurationByName.asTarget()],
          description='Pipeline p95 duration over time by pipeline name.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('s')
        + g.panel.timeSeries.options.legend.withDisplayMode('table')
        + g.panel.timeSeries.options.legend.withPlacement('right'),

      // Detailed drill-down panels - SQL Warehouses
      topWarehousesByQueriesTable:
        commonlib.panels.generic.table.base.new(
          'Top warehouses by queries',
          targets=[signals.warehousesAndQueries.topWarehousesByQueries.asTableTarget()],
          description='Warehouses ranked by total query volume.'
        )
        + g.panel.table.standardOptions.withOverrides([
          g.panel.table.fieldOverride.byName.new('Total Queries')
          + g.panel.table.fieldOverride.byName.withPropertiesFromOptions(
            g.panel.table.standardOptions.withUnit('short')
          ),
        ])
        + g.panel.table.queryOptions.withTransformations([
          {
            id: 'organize',
            options: {
              excludeByName: {
                __name__: true,
                Time: true,
                instance: true,
                job: true,
              },
              indexByName: {
                warehouse_id: 0,
                Value: 1,
              },
              renameByName: {
                warehouse_id: 'Warehouse ID',
                Value: 'Total Queries',
              },
            },
          },
          { id: 'sortBy', options: { sort: [{ field: 'Total Queries', desc: true }] } },
          { id: 'limit', options: { limitField: 10 } },
        ]),

      topWarehousesByErrorsTable:
        commonlib.panels.generic.table.base.new(
          'Top warehouses by errors',
          targets=[signals.warehousesAndQueries.topWarehousesByErrors.asTableTarget()],
          description='Warehouses ranked by total error count.'
        )
        + g.panel.table.standardOptions.withOverrides([
          g.panel.table.fieldOverride.byName.new('Total Errors')
          + g.panel.table.fieldOverride.byName.withPropertiesFromOptions(
            g.panel.table.standardOptions.withUnit('short')
          ),
        ])
        + g.panel.table.queryOptions.withTransformations([
          {
            id: 'organize',
            options: {
              excludeByName: {
                __name__: true,
                Time: true,
                instance: true,
                job: true,
              },
              indexByName: {
                warehouse_id: 0,
                Value: 1,
              },
              renameByName: {
                warehouse_id: 'Warehouse ID',
                Value: 'Total Errors',
              },
            },
          },
          { id: 'sortBy', options: { sort: [{ field: 'Total Errors', desc: true }] } },
          { id: 'limit', options: { limitField: 10 } },
        ]),

      topWarehousesByLatencyTable:
        commonlib.panels.generic.table.base.new(
          'Top warehouses by p95 latency',
          targets=[signals.warehousesAndQueries.topWarehousesByLatency.asTableTarget()],
          description='Warehouses ranked by highest p95 latency.'
        )
        + g.panel.table.standardOptions.withOverrides([
          g.panel.table.fieldOverride.byName.new('p95 Latency (seconds)')
          + g.panel.table.fieldOverride.byName.withPropertiesFromOptions(
            g.panel.table.standardOptions.withUnit('s')
          ),
        ])
        + g.panel.table.queryOptions.withTransformations([
          {
            id: 'organize',
            options: {
              excludeByName: {
                __name__: true,
                Time: true,
                instance: true,
                job: true,
              },
              indexByName: {
                warehouse_id: 0,
                Value: 1,
              },
              renameByName: {
                warehouse_id: 'Warehouse ID',
                Value: 'p95 Latency (seconds)',
              },
            },
          },
          { id: 'sortBy', options: { sort: [{ field: 'p95 Latency (seconds)', desc: true }] } },
          { id: 'limit', options: { limitField: 10 } },
        ]),

      queriesByWarehouseChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Queries by warehouse',
          targets=[signals.warehousesAndQueries.queriesByWarehouse.asTarget()],
          description='Query volume over time by warehouse (adapts to dashboard time range).'
        )
        + g.panel.timeSeries.standardOptions.withUnit('short'),

      queryErrorsByWarehouseChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Errors by warehouse',
          targets=[signals.warehousesAndQueries.queryErrorsByWarehouse.asTarget()],
          description='Query errors over time by warehouse (adapts to dashboard time range).'
        )
        + g.panel.timeSeries.standardOptions.withUnit('short'),

      queryLatencyByWarehouseChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Query p95 latency by warehouse',
          targets=[signals.warehousesAndQueries.queryLatencyByWarehouse.asTarget()],
          description='Query p95 latency over time by warehouse.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('s'),

      concurrencyByWarehouseChart:
        commonlib.panels.generic.timeSeries.base.new(
          'Concurrency by warehouse',
          targets=[signals.warehousesAndQueries.concurrencyByWarehouse.asTarget()],
          description='Concurrent queries by warehouse.'
        )
        + g.panel.timeSeries.standardOptions.withUnit('short'),
    },
}
