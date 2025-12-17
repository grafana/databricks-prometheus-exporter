package collector

import (
	"strings"
	"testing"
)

func TestQueries_NotEmpty(t *testing.T) {
	queries := []struct {
		name  string
		query string
	}{
		{"billingDBUsQuery", billingDBUsQuery},
		{"billingCostEstimateQuery", billingCostEstimateQuery},
		{"priceChangeEventsQuery", priceChangeEventsQuery},
		{"jobRunsQuery", jobRunsQuery},
		{"jobRunStatusQuery", jobRunStatusQuery},
		{"jobRunDurationQuery", jobRunDurationQuery},
		{"taskRetriesQuery", taskRetriesQuery},
		{"jobSLAMissQuery", jobSLAMissQuery},
		{"pipelineRunsQuery", pipelineRunsQuery},
		{"pipelineRunStatusQuery", pipelineRunStatusQuery},
		{"pipelineRunDurationQuery", pipelineRunDurationQuery},
		{"pipelineRetryEventsQuery", pipelineRetryEventsQuery},
		{"pipelineFreshnessLagQuery", pipelineFreshnessLagQuery},
		{"queriesQuery", queriesQuery},
		{"queryErrorsQuery", queryErrorsQuery},
		{"queryDurationQuery", queryDurationQuery},
		{"queriesRunningQuery", queriesRunningQuery},
	}

	for _, tt := range queries {
		t.Run(tt.name, func(t *testing.T) {
			if tt.query == "" {
				t.Errorf("%s is empty", tt.name)
			}
			if len(strings.TrimSpace(tt.query)) == 0 {
				t.Errorf("%s contains only whitespace", tt.name)
			}
		})
	}
}

func TestQueries_ContainSelect(t *testing.T) {
	queries := []struct {
		name  string
		query string
	}{
		{"billingDBUsQuery", billingDBUsQuery},
		{"billingCostEstimateQuery", billingCostEstimateQuery},
		{"priceChangeEventsQuery", priceChangeEventsQuery},
		{"jobRunsQuery", jobRunsQuery},
		{"jobRunStatusQuery", jobRunStatusQuery},
		{"jobRunDurationQuery", jobRunDurationQuery},
		{"taskRetriesQuery", taskRetriesQuery},
		{"jobSLAMissQuery", jobSLAMissQuery},
		{"pipelineRunsQuery", pipelineRunsQuery},
		{"pipelineRunStatusQuery", pipelineRunStatusQuery},
		{"pipelineRunDurationQuery", pipelineRunDurationQuery},
		{"pipelineRetryEventsQuery", pipelineRetryEventsQuery},
		{"pipelineFreshnessLagQuery", pipelineFreshnessLagQuery},
		{"queriesQuery", queriesQuery},
		{"queryErrorsQuery", queryErrorsQuery},
		{"queryDurationQuery", queryDurationQuery},
		{"queriesRunningQuery", queriesRunningQuery},
	}

	for _, tt := range queries {
		t.Run(tt.name, func(t *testing.T) {
			upper := strings.ToUpper(tt.query)
			if !strings.Contains(upper, "SELECT") {
				t.Errorf("%s does not contain SELECT statement", tt.name)
			}
		})
	}
}

func TestQueries_ContainSystemTable(t *testing.T) {
	// Verify each query references the correct system table
	tests := []struct {
		name      string
		query     string
		tableName string
	}{
		{"billingDBUsQuery", billingDBUsQuery, "system.billing.usage"},
		{"billingCostEstimateQuery", billingCostEstimateQuery, "system.billing.usage"},
		{"priceChangeEventsQuery", priceChangeEventsQuery, "system.billing.list_prices"},
		{"jobRunsQuery", jobRunsQuery, "system.lakeflow.job_run_timeline"},
		{"jobRunStatusQuery", jobRunStatusQuery, "system.lakeflow.job_run_timeline"},
		{"jobRunDurationQuery", jobRunDurationQuery, "system.lakeflow.job_run_timeline"},
		{"taskRetriesQuery", taskRetriesQuery, "system.lakeflow.job_task_run_timeline"},
		{"jobSLAMissQuery", jobSLAMissQuery, "system.lakeflow.job_run_timeline"},
		{"pipelineRunsQuery", pipelineRunsQuery, "system.lakeflow.pipeline_update_timeline"},
		{"pipelineRunStatusQuery", pipelineRunStatusQuery, "system.lakeflow.pipeline_update_timeline"},
		{"pipelineRunDurationQuery", pipelineRunDurationQuery, "system.lakeflow.pipeline_update_timeline"},
		{"pipelineRetryEventsQuery", pipelineRetryEventsQuery, "system.lakeflow.pipeline_update_timeline"},
		{"pipelineFreshnessLagQuery", pipelineFreshnessLagQuery, "system.lakeflow.pipeline_update_timeline"},
		{"queriesQuery", queriesQuery, "system.query.history"},
		{"queryErrorsQuery", queryErrorsQuery, "system.query.history"},
		{"queryDurationQuery", queryDurationQuery, "system.query.history"},
		{"queriesRunningQuery", queriesRunningQuery, "system.query.history"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(tt.query, tt.tableName) {
				t.Errorf("%s does not reference expected table %s", tt.name, tt.tableName)
			}
		})
	}
}

func TestQueries_ContainWorkspaceID(t *testing.T) {
	// Most queries should filter or group by workspace_id
	// Exceptions: priceChangeEventsQuery (aggregates across all workspaces)
	queries := []struct {
		name          string
		query         string
		shouldContain bool
	}{
		{"billingDBUsQuery", billingDBUsQuery, true},
		{"billingCostEstimateQuery", billingCostEstimateQuery, true},
		{"priceChangeEventsQuery", priceChangeEventsQuery, false}, // No workspace_id
		{"jobRunsQuery", jobRunsQuery, true},
		{"jobRunStatusQuery", jobRunStatusQuery, true},
		{"jobRunDurationQuery", jobRunDurationQuery, true},
		{"taskRetriesQuery", taskRetriesQuery, true},
		{"jobSLAMissQuery", jobSLAMissQuery, true},
		{"pipelineRunsQuery", pipelineRunsQuery, true},
		{"pipelineRunStatusQuery", pipelineRunStatusQuery, true},
		{"pipelineRunDurationQuery", pipelineRunDurationQuery, true},
		{"pipelineRetryEventsQuery", pipelineRetryEventsQuery, true},
		{"pipelineFreshnessLagQuery", pipelineFreshnessLagQuery, true},
		{"queriesQuery", queriesQuery, true},
		{"queryErrorsQuery", queryErrorsQuery, true},
		{"queryDurationQuery", queryDurationQuery, true},
		{"queriesRunningQuery", queriesRunningQuery, true},
	}

	for _, tt := range queries {
		t.Run(tt.name, func(t *testing.T) {
			containsWorkspaceID := strings.Contains(tt.query, "workspace_id")
			if tt.shouldContain && !containsWorkspaceID {
				t.Errorf("%s should contain workspace_id but doesn't", tt.name)
			}
			if !tt.shouldContain && containsWorkspaceID {
				t.Errorf("%s should not contain workspace_id but does", tt.name)
			}
		})
	}
}

func TestQueries_ContainTimeFilter(t *testing.T) {
	// All queries should have a time-based WHERE clause for performance
	queries := []struct {
		name  string
		query string
	}{
		{"billingDBUsQuery", billingDBUsQuery},
		{"billingCostEstimateQuery", billingCostEstimateQuery},
		{"priceChangeEventsQuery", priceChangeEventsQuery},
		{"jobRunsQuery", jobRunsQuery},
		{"jobRunStatusQuery", jobRunStatusQuery},
		{"jobRunDurationQuery", jobRunDurationQuery},
		{"taskRetriesQuery", taskRetriesQuery},
		{"jobSLAMissQuery", jobSLAMissQuery},
		{"pipelineRunsQuery", pipelineRunsQuery},
		{"pipelineRunStatusQuery", pipelineRunStatusQuery},
		{"pipelineRunDurationQuery", pipelineRunDurationQuery},
		{"pipelineRetryEventsQuery", pipelineRetryEventsQuery},
		{"pipelineFreshnessLagQuery", pipelineFreshnessLagQuery},
		{"queriesQuery", queriesQuery},
		{"queryErrorsQuery", queryErrorsQuery},
		{"queryDurationQuery", queryDurationQuery},
		{"queriesRunningQuery", queriesRunningQuery},
	}

	for _, tt := range queries {
		t.Run(tt.name, func(t *testing.T) {
			upper := strings.ToUpper(tt.query)
			// Check for time filters (INTERVAL, current_date, current_timestamp)
			hasTimeFilter := strings.Contains(upper, "INTERVAL") ||
				strings.Contains(upper, "CURRENT_DATE") ||
				strings.Contains(upper, "CURRENT_TIMESTAMP")

			if !hasTimeFilter {
				t.Errorf("%s does not contain time-based filter", tt.name)
			}
		})
	}
}

func TestQueries_ContainAggregation(t *testing.T) {
	// Most queries should have aggregation (COUNT, SUM, AVG, etc.)
	// This is important for metric generation
	queries := []struct {
		name  string
		query string
	}{
		{"billingDBUsQuery", billingDBUsQuery},
		{"billingCostEstimateQuery", billingCostEstimateQuery},
		{"priceChangeEventsQuery", priceChangeEventsQuery},
		{"jobRunsQuery", jobRunsQuery},
		{"jobRunStatusQuery", jobRunStatusQuery},
		{"jobRunDurationQuery", jobRunDurationQuery},
		{"taskRetriesQuery", taskRetriesQuery},
		{"jobSLAMissQuery", jobSLAMissQuery},
		{"pipelineRunsQuery", pipelineRunsQuery},
		{"pipelineRunStatusQuery", pipelineRunStatusQuery},
		{"pipelineRunDurationQuery", pipelineRunDurationQuery},
		{"pipelineRetryEventsQuery", pipelineRetryEventsQuery},
		{"pipelineFreshnessLagQuery", pipelineFreshnessLagQuery},
		{"queriesQuery", queriesQuery},
		{"queryErrorsQuery", queryErrorsQuery},
		{"queryDurationQuery", queryDurationQuery},
		{"queriesRunningQuery", queriesRunningQuery},
	}

	for _, tt := range queries {
		t.Run(tt.name, func(t *testing.T) {
			upper := strings.ToUpper(tt.query)
			hasAggregation := strings.Contains(upper, "COUNT(") ||
				strings.Contains(upper, "SUM(") ||
				strings.Contains(upper, "AVG(") ||
				strings.Contains(upper, "MAX(") ||
				strings.Contains(upper, "MIN(") ||
				strings.Contains(upper, "PERCENTILE_APPROX(")

			if !hasAggregation {
				t.Errorf("%s does not contain aggregation function", tt.name)
			}
		})
	}
}

func TestTimeWindowConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"BillingWindow", BillingWindow, "1 DAY"},
		{"JobsWindow", JobsWindow, "2 HOURS"},
		{"QueriesWindow", QueriesWindow, "1 HOURS"},
		{"PricesWindow", PricesWindow, "1 DAY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestQueries_UseConfiguredWindows(t *testing.T) {
	// Verify queries use the time window constants where appropriate
	tests := []struct {
		queryName string
		query     string
		window    string
	}{
		{"billingDBUsQuery", billingDBUsQuery, "1 DAY"},
		{"billingCostEstimateQuery", billingCostEstimateQuery, "1 DAY"},
		{"priceChangeEventsQuery", priceChangeEventsQuery, "1 DAY"},
		{"jobRunsQuery", jobRunsQuery, "2 HOURS"},
		{"jobRunStatusQuery", jobRunStatusQuery, "2 HOURS"},
		{"jobRunDurationQuery", jobRunDurationQuery, "2 HOURS"},
		{"queriesQuery", queriesQuery, "1 HOUR"},
		{"queryErrorsQuery", queryErrorsQuery, "1 HOUR"},
	}

	for _, tt := range tests {
		t.Run(tt.queryName, func(t *testing.T) {
			if !strings.Contains(tt.query, tt.window) {
				t.Errorf("%s should contain time window %q", tt.queryName, tt.window)
			}
		})
	}
}
