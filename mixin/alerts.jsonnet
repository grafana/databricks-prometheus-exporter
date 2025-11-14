// Helper file to extract just the Prometheus alerts from the mixin
local mixin = import './mixin.libsonnet';

// Extract alerts and format as YAML-compatible structure
{
  groups: [mixin.prometheusAlerts],
}

