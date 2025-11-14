function(this) {
  local legendCustomTemplate = '{{instance}} - {{workspace_id}}',
  local aggregationLabels = std.join(',', this.groupLabels + this.instanceLabels),
  filteringSelector: this.filteringSelector,
  groupLabels: this.groupLabels,
  instanceLabels: this.instanceLabels,
  enableLokiLogs: false,  // Databricks does not gather logs via this exporter
  legendCustomTemplate: legendCustomTemplate,
  aggLevel: 'none',
  aggFunction: 'avg',
  alertsInterval: '5m',
  discoveryMetric: {
    prometheus: 'databricks_billing_dbus_total',
  },
  signals: {
    // Billing & Cost Metrics
    billingDbusTotal: {
      name: 'Billing DBUs total',
      nameShort: 'DBUs',
      description: 'Daily Databricks Units (DBUs) consumed per workspace × SKU.',
      type: 'gauge',
      unit: 'dbu',
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

    priceChangeEvents: {
      name: 'Price change events',
      nameShort: 'Price changes',
      description: 'Count of pricing changes for a SKU (from historical list prices).',
      type: 'counter',
      unit: 'events',
      sources: {
        prometheus: {
          expr: 'databricks_price_change_events{%(queriesSelector)s}',
          legendCustomTemplate: '{{sku_name}}',
        },
      },
    },

    billingExportErrors: {
      name: 'Billing export errors',
      nameShort: 'Export errors',
      description: 'Exporter error count segmented by stage (sql, csv, s3, publish…).',
      type: 'counter',
      unit: 'errors',
      sources: {
        prometheus: {
          expr: 'databricks_billing_export_errors_total{%(queriesSelector)s}',
          legendCustomTemplate: '{{stage}}',
        },
      },
    },

    // Yesterday's cost for stat panel
    yesterdayCost: {
      name: "Yesterday's cost",
      nameShort: 'Yesterday $',
      description: 'Total cost for yesterday (D-1).',
      type: 'raw',
      unit: 'currencyUSD',
      sources: {
        prometheus: {
          expr: 'sum by (' + aggregationLabels + ') (databricks_billing_cost_estimate_usd{%(queriesSelector)s} offset 1d)',
          legendCustomTemplate: 'Yesterday',
        },
      },
    },

    // DoD cost delta
    dodCostDelta: {
      name: 'DoD cost delta',
      nameShort: 'DoD Δ%',
      description: 'Day-over-day cost delta percentage (D-1 vs D-2).',
      type: 'raw',
      unit: 'percentunit',
      sources: {
        prometheus: {
          expr: |||
            (
              sum by (%(aggregationLabels)s) (databricks_billing_cost_estimate_usd{%(queriesSelector)s} offset 1d)
              - 
              sum by (%(aggregationLabels)s) (databricks_billing_cost_estimate_usd{%(queriesSelector)s} offset 2d)
            )
            /
            sum by (%(aggregationLabels)s) (databricks_billing_cost_estimate_usd{%(queriesSelector)s} offset 2d)
          ||| % {
            aggregationLabels: aggregationLabels,
            queriesSelector: '%(queriesSelector)s',
          },
          legendCustomTemplate: 'DoD',
        },
      },
    },

    // Yesterday's DBUs
    yesterdayDbus: {
      name: "Yesterday's DBUs",
      nameShort: 'DBUs (D-1)',
      description: 'Total DBUs consumed yesterday.',
      type: 'raw',
      unit: 'dbu',
      sources: {
        prometheus: {
          expr: 'sum by (' + aggregationLabels + ') (databricks_billing_dbus_total{%(queriesSelector)s} offset 1d)',
          legendCustomTemplate: 'DBUs',
        },
      },
    },

    // Cost by SKU over time
    costBySku: {
      name: 'Cost by SKU',
      nameShort: 'Cost by SKU',
      description: 'Cost breakdown by SKU over time.',
      type: 'raw',
      unit: 'currencyUSD',
      sources: {
        prometheus: {
          expr: 'sum by (' + aggregationLabels + ', sku_name) (databricks_billing_cost_estimate_usd{%(queriesSelector)s})',
          legendCustomTemplate: '{{sku_name}}',
        },
      },
    },

    // DBUs by SKU over time
    dbusBySku: {
      name: 'DBUs by SKU',
      nameShort: 'DBUs by SKU',
      description: 'DBUs breakdown by SKU over time.',
      type: 'raw',
      unit: 'dbu',
      sources: {
        prometheus: {
          expr: 'sum by (' + aggregationLabels + ', sku_name) (databricks_billing_dbus_total{%(queriesSelector)s})',
          legendCustomTemplate: '{{sku_name}}',
        },
      },
    },

    // Top workspaces by cost
    topWorkspacesByCost: {
      name: 'Top workspaces by cost',
      nameShort: 'Top workspaces',
      description: 'Workspaces with highest costs (yesterday).',
      type: 'gauge',
      unit: 'currencyUSD',
      sources: {
        prometheus: {
          expr: 'sum by (workspace_id) (databricks_billing_cost_estimate_usd{%(queriesSelector)s} offset 1d)',
          legendCustomTemplate: '{{workspace_id}}',
        },
      },
    },

    // Top SKUs by cost
    topSkusByCost: {
      name: 'Top SKUs by cost',
      nameShort: 'Top SKUs',
      description: 'SKUs with highest costs (yesterday).',
      type: 'gauge',
      unit: 'currencyUSD',
      sources: {
        prometheus: {
          expr: 'sum by (sku_name) (databricks_billing_cost_estimate_usd{%(queriesSelector)s} offset 1d)',
          legendCustomTemplate: '{{sku_name}}',
        },
      },
    },

    // DBUs by workspace
    dbusByWorkspace: {
      name: 'DBUs by workspace',
      nameShort: 'DBUs by workspace',
      description: 'DBUs consumed per workspace (for heatmap).',
      type: 'raw',
      unit: 'dbu',
      sources: {
        prometheus: {
          expr: 'sum by (workspace_id, sku_name) (databricks_billing_dbus_total{%(queriesSelector)s})',
          legendCustomTemplate: '{{workspace_id}} - {{sku_name}}',
        },
      },
    },
  },
}

