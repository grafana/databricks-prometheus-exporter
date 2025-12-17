package collector

import (
	"strings"
	"testing"
	"time"
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

// TestDurationToSQLInterval tests the duration to SQL interval conversion function.
// Databricks SQL accepts both singular and plural forms (e.g., "1 HOUR" and "1 HOURS"),
// but we use grammatically correct forms for clarity.
func TestDurationToSQLInterval(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		// Days - singular and plural
		{"1 day", 24 * time.Hour, "1 DAY"},
		{"2 days", 48 * time.Hour, "2 DAYS"},
		{"3 days", 72 * time.Hour, "3 DAYS"},
		{"7 days", 7 * 24 * time.Hour, "7 DAYS"},
		{"30 days", 30 * 24 * time.Hour, "30 DAYS"},
		{"90 days", 90 * 24 * time.Hour, "90 DAYS"},

		// Hours - singular and plural
		{"1 hour", 1 * time.Hour, "1 HOUR"},
		{"2 hours", 2 * time.Hour, "2 HOURS"},
		{"12 hours", 12 * time.Hour, "12 HOURS"},
		{"23 hours", 23 * time.Hour, "23 HOURS"},

		// Edge case: 25 hours should be hours, not days
		{"25 hours", 25 * time.Hour, "25 HOURS"},

		// Minutes - singular and plural
		{"1 minute", 1 * time.Minute, "1 MINUTE"},
		{"30 minutes", 30 * time.Minute, "30 MINUTES"},
		{"59 minutes", 59 * time.Minute, "59 MINUTES"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := durationToSQLInterval(tt.duration)
			if result != tt.expected {
				t.Errorf("durationToSQLInterval(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestQueries_UseConfiguredWindows(t *testing.T) {
	// Verify legacy queries use the expected time windows
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

// ===== Build* Function Tests =====

func TestBuildBillingDBUsQuery(t *testing.T) {
	tests := []struct {
		name           string
		lookback       time.Duration
		expectedWindow string
	}{
		{"1 day lookback", 24 * time.Hour, "INTERVAL 1 DAY"},
		{"2 days lookback", 48 * time.Hour, "INTERVAL 2 DAYS"},
		{"7 days lookback", 7 * 24 * time.Hour, "INTERVAL 7 DAYS"},
		{"30 days lookback", 30 * 24 * time.Hour, "INTERVAL 30 DAYS"},
		// Edge case: 25 hours should use hours, not days
		{"25 hours lookback", 25 * time.Hour, "INTERVAL 25 HOURS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := BuildBillingDBUsQuery(tt.lookback)
			if !strings.Contains(query, tt.expectedWindow) {
				t.Errorf("BuildBillingDBUsQuery(%v) should contain %q", tt.lookback, tt.expectedWindow)
			}
			// Verify required elements
			if !strings.Contains(query, "system.billing.usage") {
				t.Error("Query should reference system.billing.usage")
			}
			if !strings.Contains(query, "workspace_id") {
				t.Error("Query should select workspace_id")
			}
			if !strings.Contains(query, "sku_name") {
				t.Error("Query should select sku_name")
			}
		})
	}
}

func TestBuildBillingCostEstimateQuery(t *testing.T) {
	tests := []struct {
		name           string
		lookback       time.Duration
		expectedWindow string
	}{
		{"1 day lookback", 24 * time.Hour, "INTERVAL 1 DAY"},
		{"7 days lookback", 7 * 24 * time.Hour, "INTERVAL 7 DAYS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := BuildBillingCostEstimateQuery(tt.lookback)
			if !strings.Contains(query, tt.expectedWindow) {
				t.Errorf("BuildBillingCostEstimateQuery(%v) should contain %q", tt.lookback, tt.expectedWindow)
			}
			// Verify joins pricing data
			if !strings.Contains(query, "system.billing.list_prices") {
				t.Error("Query should reference system.billing.list_prices")
			}
		})
	}
}

func TestBuildPriceChangeEventsQuery(t *testing.T) {
	tests := []struct {
		name           string
		lookback       time.Duration
		expectedWindow string
	}{
		{"1 day lookback", 24 * time.Hour, "INTERVAL 1 DAY"},
		{"90 days lookback", 90 * 24 * time.Hour, "INTERVAL 90 DAYS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := BuildPriceChangeEventsQuery(tt.lookback)
			if !strings.Contains(query, tt.expectedWindow) {
				t.Errorf("BuildPriceChangeEventsQuery(%v) should contain %q", tt.lookback, tt.expectedWindow)
			}
			if !strings.Contains(query, "price_start_time") {
				t.Error("Query should filter by price_start_time")
			}
		})
	}
}

func TestBuildJobRunsQuery(t *testing.T) {
	tests := []struct {
		name           string
		lookback       time.Duration
		expectedWindow string
	}{
		{"1 hour lookback", 1 * time.Hour, "INTERVAL 1 HOUR"},
		{"2 hours lookback", 2 * time.Hour, "INTERVAL 2 HOURS"},
		{"24 hours lookback", 24 * time.Hour, "INTERVAL 1 DAY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := BuildJobRunsQuery(tt.lookback)
			if !strings.Contains(query, tt.expectedWindow) {
				t.Errorf("BuildJobRunsQuery(%v) should contain %q", tt.lookback, tt.expectedWindow)
			}
			if !strings.Contains(query, "system.lakeflow.job_run_timeline") {
				t.Error("Query should reference system.lakeflow.job_run_timeline")
			}
		})
	}
}

func TestBuildQueriesQuery(t *testing.T) {
	tests := []struct {
		name           string
		lookback       time.Duration
		expectedWindow string
	}{
		{"1 hour lookback", 1 * time.Hour, "INTERVAL 1 HOUR"},
		{"2 hours lookback", 2 * time.Hour, "INTERVAL 2 HOURS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := BuildQueriesQuery(tt.lookback)
			if !strings.Contains(query, tt.expectedWindow) {
				t.Errorf("BuildQueriesQuery(%v) should contain %q", tt.lookback, tt.expectedWindow)
			}
			if !strings.Contains(query, "system.query.history") {
				t.Error("Query should reference system.query.history")
			}
		})
	}
}
