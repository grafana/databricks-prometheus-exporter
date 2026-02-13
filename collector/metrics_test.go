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
			name:   "BillingDBUs",
			desc:   metrics.BillingDBUs,
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
			name:   "BillingScrapeErrors",
			desc:   metrics.BillingScrapeErrors,
			labels: []string{labelStage},
		},
		// Jobs metrics
		{
			name:   "JobRuns",
			desc:   metrics.JobRuns,
			labels: []string{labelWorkspaceID, labelJobID, labelJobName},
		},
		{
			name:   "JobRunStatus",
			desc:   metrics.JobRunStatus,
			labels: []string{labelWorkspaceID, labelJobID, labelJobName, labelStatus},
		},
		{
			name:   "JobRunDurationSeconds",
			desc:   metrics.JobRunDurationSeconds,
			labels: []string{labelWorkspaceID, labelJobID, labelJobName, labelQuantile},
		},
		{
			name:   "TaskRetries",
			desc:   metrics.TaskRetries,
			labels: []string{labelWorkspaceID, labelJobID, labelJobName, labelTaskKey},
		},
		{
			name:   "JobSLAMiss",
			desc:   metrics.JobSLAMiss,
			labels: []string{labelWorkspaceID, labelJobID, labelJobName},
		},
		// Pipelines metrics
		{
			name:   "PipelineRuns",
			desc:   metrics.PipelineRuns,
			labels: []string{labelWorkspaceID, labelPipelineID, labelPipelineName},
		},
		{
			name:   "PipelineRunStatus",
			desc:   metrics.PipelineRunStatus,
			labels: []string{labelWorkspaceID, labelPipelineID, labelPipelineName, labelStatus},
		},
		{
			name:   "PipelineRunDurationSeconds",
			desc:   metrics.PipelineRunDurationSeconds,
			labels: []string{labelWorkspaceID, labelPipelineID, labelPipelineName, labelQuantile},
		},
		{
			name:   "PipelineRetryEvents",
			desc:   metrics.PipelineRetryEvents,
			labels: []string{labelWorkspaceID, labelPipelineID, labelPipelineName},
		},
		{
			name:   "PipelineFreshnessLagSeconds",
			desc:   metrics.PipelineFreshnessLagSeconds,
			labels: []string{labelWorkspaceID, labelPipelineID, labelPipelineName},
		},
		// SQL Warehouse metrics
		{
			name:   "Queries",
			desc:   metrics.Queries,
			labels: []string{labelWorkspaceID, labelWarehouseID},
		},
		{
			name:   "QueryDurationSeconds",
			desc:   metrics.QueryDurationSeconds,
			labels: []string{labelWorkspaceID, labelWarehouseID, labelQuantile},
		},
		{
			name:   "QueryErrors",
			desc:   metrics.QueryErrors,
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

	// We expect 21 metrics:
	// - 4 billing metrics
	// - 5 jobs metrics
	// - 5 pipelines metrics
	// - 4 SQL warehouse metrics
	// - 3 health metrics (exporter_up, scrape_status, exporter_info)
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
		{"BillingDBUs", metrics.BillingDBUs},
		{"BillingCostEstimateUSD", metrics.BillingCostEstimateUSD},
		{"PriceChangeEvents", metrics.PriceChangeEvents},
		{"BillingScrapeErrors", metrics.BillingScrapeErrors},
		{"JobRuns", metrics.JobRuns},
		{"JobRunStatus", metrics.JobRunStatus},
		{"JobRunDurationSeconds", metrics.JobRunDurationSeconds},
		{"TaskRetries", metrics.TaskRetries},
		{"JobSLAMiss", metrics.JobSLAMiss},
		{"PipelineRuns", metrics.PipelineRuns},
		{"PipelineRunStatus", metrics.PipelineRunStatus},
		{"PipelineRunDurationSeconds", metrics.PipelineRunDurationSeconds},
		{"PipelineRetryEvents", metrics.PipelineRetryEvents},
		{"PipelineFreshnessLagSeconds", metrics.PipelineFreshnessLagSeconds},
		{"Queries", metrics.Queries},
		{"QueryDurationSeconds", metrics.QueryDurationSeconds},
		{"QueryErrors", metrics.QueryErrors},
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
