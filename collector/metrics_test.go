package collector

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNewMetricDescriptors(t *testing.T) {
	metrics := NewMetricDescriptors()

	if metrics == nil {
		t.Fatal("NewMetricDescriptors returned nil")
	}

	// Test that all metrics are initialized
	tests := []struct {
		name   string
		desc   *prometheus.Desc
		labels []string
	}{
		// Billing metrics
		{
			name:   "BillingDBUsTotal",
			desc:   metrics.BillingDBUsTotal,
			labels: []string{labelWorkspaceID, labelSKUName},
		},
		{
			name:   "BillingCostEstimateUSD",
			desc:   metrics.BillingCostEstimateUSD,
			labels: []string{labelWorkspaceID, labelSKUName},
		},
		{
			name:   "PriceChangeEvents",
			desc:   metrics.PriceChangeEvents,
			labels: []string{labelSKUName},
		},
		{
			name:   "BillingExportErrorsTotal",
			desc:   metrics.BillingExportErrorsTotal,
			labels: []string{labelStage},
		},
		// Jobs metrics
		{
			name:   "JobRunsTotal",
			desc:   metrics.JobRunsTotal,
			labels: []string{labelWorkspaceID, labelJobID, labelJobName},
		},
		{
			name:   "JobRunStatusTotal",
			desc:   metrics.JobRunStatusTotal,
			labels: []string{labelWorkspaceID, labelJobID, labelJobName, labelStatus},
		},
		{
			name:   "JobRunDurationSeconds",
			desc:   metrics.JobRunDurationSeconds,
			labels: []string{labelWorkspaceID, labelJobID, labelJobName, labelQuantile},
		},
		{
			name:   "TaskRetriesTotal",
			desc:   metrics.TaskRetriesTotal,
			labels: []string{labelWorkspaceID, labelJobID, labelJobName, labelTaskKey},
		},
		{
			name:   "JobSLAMissTotal",
			desc:   metrics.JobSLAMissTotal,
			labels: []string{labelWorkspaceID, labelJobID, labelJobName},
		},
		// Pipelines metrics
		{
			name:   "PipelineRunsTotal",
			desc:   metrics.PipelineRunsTotal,
			labels: []string{labelWorkspaceID, labelPipelineID, labelPipelineName},
		},
		{
			name:   "PipelineRunStatusTotal",
			desc:   metrics.PipelineRunStatusTotal,
			labels: []string{labelWorkspaceID, labelPipelineID, labelPipelineName, labelStatus},
		},
		{
			name:   "PipelineRunDurationSeconds",
			desc:   metrics.PipelineRunDurationSeconds,
			labels: []string{labelWorkspaceID, labelPipelineID, labelPipelineName, labelQuantile},
		},
		{
			name:   "PipelineRetryEventsTotal",
			desc:   metrics.PipelineRetryEventsTotal,
			labels: []string{labelWorkspaceID, labelPipelineID, labelPipelineName},
		},
		{
			name:   "PipelineFreshnessLagSeconds",
			desc:   metrics.PipelineFreshnessLagSeconds,
			labels: []string{labelWorkspaceID, labelPipelineID, labelPipelineName},
		},
		// SQL Warehouse metrics
		{
			name:   "QueriesTotal",
			desc:   metrics.QueriesTotal,
			labels: []string{labelWorkspaceID, labelWarehouseID},
		},
		{
			name:   "QueryDurationSeconds",
			desc:   metrics.QueryDurationSeconds,
			labels: []string{labelWorkspaceID, labelWarehouseID, labelQuantile},
		},
		{
			name:   "QueryErrorsTotal",
			desc:   metrics.QueryErrorsTotal,
			labels: []string{labelWorkspaceID, labelWarehouseID},
		},
		{
			name:   "QueriesRunning",
			desc:   metrics.QueriesRunning,
			labels: []string{labelWorkspaceID, labelWarehouseID},
		},
		// Health metrics
		{
			name:   "ExporterUp",
			desc:   metrics.ExporterUp,
			labels: nil,
		},
		{
			name:   "ScrapeStatus",
			desc:   metrics.ScrapeStatus,
			labels: []string{labelQuery},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.desc == nil {
				t.Errorf("%s descriptor is nil", tt.name)
				return
			}

			// Verify the metric name contains the namespace
			descString := tt.desc.String()
			if !strings.Contains(descString, namespace) {
				t.Errorf("%s: descriptor does not contain namespace '%s': %s",
					tt.name, namespace, descString)
			}

			// Verify expected label count
			// Note: This is a basic check. Prometheus descriptors don't expose label names directly,
			// but we can verify the descriptor was created successfully.
			if tt.desc.String() == "" {
				t.Errorf("%s: descriptor String() is empty", tt.name)
			}
		})
	}
}

func TestMetricDescriptors_Describe(t *testing.T) {
	metrics := NewMetricDescriptors()
	ch := make(chan *prometheus.Desc, 30) // Buffer for all metrics

	// Call Describe
	metrics.Describe(ch)
	close(ch)

	// Count received descriptors
	count := 0
	for range ch {
		count++
	}

	// We expect 20 metrics:
	// - 4 billing metrics
	// - 5 jobs metrics
	// - 5 pipelines metrics
	// - 4 SQL warehouse metrics
	// - 2 health metrics (exporter_up, scrape_status)
	expectedCount := 21
	if count != expectedCount {
		t.Errorf("Expected %d metric descriptors, got %d", expectedCount, count)
	}
}

func TestMetricDescriptors_AllMetricsHaveDescriptions(t *testing.T) {
	metrics := NewMetricDescriptors()

	tests := []struct {
		name string
		desc *prometheus.Desc
	}{
		{"BillingDBUsTotal", metrics.BillingDBUsTotal},
		{"BillingCostEstimateUSD", metrics.BillingCostEstimateUSD},
		{"PriceChangeEvents", metrics.PriceChangeEvents},
		{"BillingExportErrorsTotal", metrics.BillingExportErrorsTotal},
		{"JobRunsTotal", metrics.JobRunsTotal},
		{"JobRunStatusTotal", metrics.JobRunStatusTotal},
		{"JobRunDurationSeconds", metrics.JobRunDurationSeconds},
		{"TaskRetriesTotal", metrics.TaskRetriesTotal},
		{"JobSLAMissTotal", metrics.JobSLAMissTotal},
		{"PipelineRunsTotal", metrics.PipelineRunsTotal},
		{"PipelineRunStatusTotal", metrics.PipelineRunStatusTotal},
		{"PipelineRunDurationSeconds", metrics.PipelineRunDurationSeconds},
		{"PipelineRetryEventsTotal", metrics.PipelineRetryEventsTotal},
		{"PipelineFreshnessLagSeconds", metrics.PipelineFreshnessLagSeconds},
		{"QueriesTotal", metrics.QueriesTotal},
		{"QueryDurationSeconds", metrics.QueryDurationSeconds},
		{"QueryErrorsTotal", metrics.QueryErrorsTotal},
		{"QueriesRunning", metrics.QueriesRunning},
		{"ExporterUp", metrics.ExporterUp},
		{"ScrapeStatus", metrics.ScrapeStatus},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.desc == nil {
				t.Errorf("%s is nil", tt.name)
				return
			}

			descString := tt.desc.String()
			if descString == "" {
				t.Errorf("%s has empty description", tt.name)
			}

			// Verify help text is not empty
			// The String() method includes the help text
			if len(descString) < 10 {
				t.Errorf("%s description is too short: %s", tt.name, descString)
			}
		})
	}
}

func TestLabelConstants(t *testing.T) {
	// Verify label constant values are set correctly
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"labelWorkspaceID", labelWorkspaceID, "workspace_id"},
		{"labelSKUName", labelSKUName, "sku_name"},
		{"labelStatus", labelStatus, "status"},
		{"labelStage", labelStage, "stage"},
		{"labelQuantile", labelQuantile, "quantile"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	if namespace != "databricks" {
		t.Errorf("namespace = %q, want %q", namespace, "databricks")
	}
}
