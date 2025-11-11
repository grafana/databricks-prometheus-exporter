function(this) {
  local legendCustomTemplate = '{{instance}} - {{workspace_id}}',
  local aggregationLabels = std.join(',', this.groupLabels + this.instanceLabels),
  filteringSelector: this.filteringSelector,
  groupLabels: this.groupLabels,
  instanceLabels: this.instanceLabels,
  enableLokiLogs: false,
  legendCustomTemplate: legendCustomTemplate,
  aggLevel: 'none',
  aggFunction: 'avg',
  alertsInterval: '5m',
  discoveryMetric: {
    prometheus: 'databricks_billing_dbus_total',
  },
  signals: {
    billingDbusTotal: {
      name: 'Billing DBUs total',
      nameShort: 'DBUs',
      description: 'Databricks units (DBUs) consumed per workspace × SKU (24h rolling window from exporter).',
      type: 'gauge',
      unit: 'none',  // would be 'dbu' but no custom unit available
      sources: {
        prometheus: {
          expr: 'databricks_billing_dbus_total{%(queriesSelector)s}',
          legendCustomTemplate: '{{workspace_id}} - {{sku_name}}',
        },
      },
    },

    billingCostEstimateUsd: {
      name: 'Billing cost estimate USD',
      nameShort: 'Cost ($)',
      description: 'Daily list-price cost estimate (DBUs × list price with effective windows).',
      type: 'gauge',
      unit: 'currencyUSD',
      sources: {
        prometheus: {
          expr: 'databricks_billing_cost_estimate_usd{%(queriesSelector)s}',
          legendCustomTemplate: '{{workspace_id}} - {{sku_name}}',
        },
      },
    },

    yesterdayCost: {
      name: 'Total cost (24h window)',
      nameShort: 'Cost $',
      description: 'Total cost estimate for the past 24 hours (rolling 1-day window from exporter).',
      type: 'raw',
      unit: 'currencyUSD',
      sources: {
        prometheus: {
          expr: 'sum(databricks_billing_cost_estimate_usd{%(queriesSelector)s})',
          legendCustomTemplate: 'Cost',
        },
      },
    },

    dodCostDelta: {
      name: 'Cost change (vs 24h ago)',
      nameShort: '24h Δ%',
      description: 'Percentage change comparing current 24h window to the 24h window from yesterday. Positive = costs increased.',
      type: 'raw',
      unit: 'percentunit',
      sources: {
        prometheus: {
          expr: |||
            (
              sum(databricks_billing_cost_estimate_usd{%(queriesSelector)s})
              - 
              sum(databricks_billing_cost_estimate_usd{%(queriesSelector)s} offset 24h)
            )
            /
            sum(databricks_billing_cost_estimate_usd{%(queriesSelector)s} offset 24h)
          ||| % {
            queriesSelector: '%(queriesSelector)s',
          },
          legendCustomTemplate: '24h change',
        },
      },
    },

    totalDbusConsumed: {
      name: 'Total DBUs (24h window)',
      nameShort: 'DBUs',
      description: 'Total DBUs consumed over the past 24 hours (rolling 1-day window from SQL query).',
      type: 'raw',
      unit: 'none',  // would be 'dbu' but no custom unit available
      sources: {
        prometheus: {
          expr: 'sum(databricks_billing_dbus_total{%(queriesSelector)s})',
          legendCustomTemplate: 'DBUs',
        },
      },
    },

    costBySku: {
      name: 'Cost by SKU',
      nameShort: 'Cost by SKU',
      description: 'Cost breakdown by SKU over time (24h rolling window from exporter).',
      type: 'gauge',
      unit: 'currencyUSD',
      sources: {
        prometheus: {
          expr: 'sum by (sku_name) (databricks_billing_cost_estimate_usd{%(queriesSelector)s})',
          legendCustomTemplate: '{{sku_name}}',
        },
      },
    },

    costPerDbuBySku: {
      name: 'Cost per DBU by SKU',
      nameShort: 'Cost/DBU',
      description: 'Cost efficiency per DBU by SKU - shows how much each DBU costs for each SKU (24h rolling window).',
      type: 'gauge',
      unit: 'currencyUSD',
      sources: {
        prometheus: {
          expr: |||
            sum by (sku_name) (databricks_billing_cost_estimate_usd{%(queriesSelector)s})
            /
            sum by (sku_name) (databricks_billing_dbus_total{%(queriesSelector)s})
          |||,
          legendCustomTemplate: '{{sku_name}}',
        },
      },
    },

    topWorkspacesByCost: {
      name: 'Top workspaces by cost (24h window)',
      nameShort: 'Top workspaces',
      description: 'Workspaces with highest costs over the past 24 hours (rolling 1-day window from exporter).',
      type: 'gauge',
      unit: 'currencyUSD',
      sources: {
        prometheus: {
          expr: 'sum by (workspace_id) (databricks_billing_cost_estimate_usd{%(queriesSelector)s})',
          legendCustomTemplate: '{{workspace_id}}',
        },
      },
    },

    topSkusByCost: {
      name: 'Top SKUs by cost (24h window)',
      nameShort: 'Top SKUs',
      description: 'SKUs with highest costs over the past 24 hours (rolling 1-day window from exporter).',
      type: 'gauge',
      unit: 'currencyUSD',
      sources: {
        prometheus: {
          expr: 'sum by (sku_name) (databricks_billing_cost_estimate_usd{%(queriesSelector)s})',
          legendCustomTemplate: '{{sku_name}}',
        },
      },
    },
  },
}
