local databrickslib = import './main.libsonnet';
local config = (import './config.libsonnet');
local util = import 'grafana-cloud-integration-utils/util.libsonnet';

local databricks =
  databrickslib.new()
  + databrickslib.withConfigMixin(
    {
      filteringSelector: config.filteringSelector,
      uid: config.uid,
      enableLokiLogs: false,
    }
  );

local optional_labels = {
  job+: {
    label: 'Job',
    allValue: '.+',
  },
  workspace_id+: {
    label: 'Workspace',
    allValue: '.*',
  },
  sku_name+: {
    label: 'SKU',
    allValue: '.*',
  },
  job_name+: {
    label: 'Job Name',
    allValue: '.*',
    multi: true,
  },
  pipeline_name+: {
    label: 'Pipeline Name',
    allValue: '.*',
    multi: true,
  },
  warehouse_id+: {
    label: 'Warehouse ID',
    allValue: '.*',
    multi: true,
  },
};

{
  grafanaDashboards+:: {
    [fname]:
      local dashboard = databricks.grafana.dashboards[fname];
      dashboard + util.patch_variables(dashboard, optional_labels)

    for fname in std.objectFields(databricks.grafana.dashboards)
  },

  prometheusAlerts+:: databricks.prometheus.alerts,
  prometheusRules+:: databricks.prometheus.recordingRules,
}
