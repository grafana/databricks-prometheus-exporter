package collector

import (
	"strings"
	"testing"
	"time"
)

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

// ===== Billing Query Builder Tests =====

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

// ===== Jobs Query Builder Tests =====

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

func TestBuildJobRunStatusQuery(t *testing.T) {
	query := BuildJobRunStatusQuery(2 * time.Hour)
	if !strings.Contains(query, "INTERVAL 2 HOURS") {
		t.Error("Query should contain INTERVAL 2 HOURS")
	}
	if !strings.Contains(query, "result_state") {
		t.Error("Query should select result_state")
	}
	if !strings.Contains(query, "system.lakeflow.job_run_timeline") {
		t.Error("Query should reference system.lakeflow.job_run_timeline")
	}
}

func TestBuildJobRunDurationQuery(t *testing.T) {
	query := BuildJobRunDurationQuery(2 * time.Hour)
	if !strings.Contains(query, "INTERVAL 2 HOURS") {
		t.Error("Query should contain INTERVAL 2 HOURS")
	}
	if !strings.Contains(query, "percentile_approx") {
		t.Error("Query should use percentile_approx for quantiles")
	}
	// Verify quantile levels
	for _, q := range []string{"0.5", "0.95", "0.99"} {
		if !strings.Contains(query, q) {
			t.Errorf("Query should calculate %s quantile", q)
		}
	}
}

func TestBuildTaskRetriesQuery(t *testing.T) {
	query := BuildTaskRetriesQuery(2 * time.Hour)
	if !strings.Contains(query, "INTERVAL 2 HOURS") {
		t.Error("Query should contain INTERVAL 2 HOURS")
	}
	if !strings.Contains(query, "system.lakeflow.job_task_run_timeline") {
		t.Error("Query should reference system.lakeflow.job_task_run_timeline")
	}
	if !strings.Contains(query, "retry_count") {
		t.Error("Query should calculate retry_count")
	}
}

func TestBuildJobSLAMissQuery(t *testing.T) {
	query := BuildJobSLAMissQuery(2*time.Hour, 3600)
	if !strings.Contains(query, "INTERVAL 2 HOURS") {
		t.Error("Query should contain INTERVAL 2 HOURS")
	}
	if !strings.Contains(query, "3600") {
		t.Error("Query should use SLA threshold of 3600 seconds")
	}
	if !strings.Contains(query, "sla_miss_count") {
		t.Error("Query should calculate sla_miss_count")
	}
}

// ===== Pipelines Query Builder Tests =====

func TestBuildPipelineRunsQuery(t *testing.T) {
	query := BuildPipelineRunsQuery(2 * time.Hour)
	if !strings.Contains(query, "INTERVAL 2 HOURS") {
		t.Error("Query should contain INTERVAL 2 HOURS")
	}
	if !strings.Contains(query, "system.lakeflow.pipeline_update_timeline") {
		t.Error("Query should reference system.lakeflow.pipeline_update_timeline")
	}
}

func TestBuildPipelineRunStatusQuery(t *testing.T) {
	query := BuildPipelineRunStatusQuery(2 * time.Hour)
	if !strings.Contains(query, "INTERVAL 2 HOURS") {
		t.Error("Query should contain INTERVAL 2 HOURS")
	}
	if !strings.Contains(query, "result_state") {
		t.Error("Query should select result_state")
	}
}

func TestBuildPipelineRunDurationQuery(t *testing.T) {
	query := BuildPipelineRunDurationQuery(2 * time.Hour)
	if !strings.Contains(query, "INTERVAL 2 HOURS") {
		t.Error("Query should contain INTERVAL 2 HOURS")
	}
	if !strings.Contains(query, "percentile_approx") {
		t.Error("Query should use percentile_approx for quantiles")
	}
}

func TestBuildPipelineRetryEventsQuery(t *testing.T) {
	query := BuildPipelineRetryEventsQuery(2 * time.Hour)
	if !strings.Contains(query, "INTERVAL 2 HOURS") {
		t.Error("Query should contain INTERVAL 2 HOURS")
	}
	if !strings.Contains(query, "retry_count") {
		t.Error("Query should calculate retry_count")
	}
}

func TestBuildPipelineFreshnessLagQuery(t *testing.T) {
	query := BuildPipelineFreshnessLagQuery(2 * time.Hour)
	if !strings.Contains(query, "INTERVAL 2 HOURS") {
		t.Error("Query should contain INTERVAL 2 HOURS")
	}
	if !strings.Contains(query, "freshness_lag_seconds") {
		t.Error("Query should calculate freshness_lag_seconds")
	}
	if !strings.Contains(query, "COMPLETED") {
		t.Error("Query should filter for completed pipelines")
	}
}

// ===== SQL Warehouse Query Builder Tests =====

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

func TestBuildQueryErrorsQuery(t *testing.T) {
	query := BuildQueryErrorsQuery(1 * time.Hour)
	if !strings.Contains(query, "INTERVAL 1 HOUR") {
		t.Error("Query should contain INTERVAL 1 HOUR")
	}
	if !strings.Contains(query, "error_message IS NOT NULL") {
		t.Error("Query should filter for error_message IS NOT NULL")
	}
}

func TestBuildQueryDurationQuery(t *testing.T) {
	query := BuildQueryDurationQuery(1 * time.Hour)
	if !strings.Contains(query, "INTERVAL 1 HOUR") {
		t.Error("Query should contain INTERVAL 1 HOUR")
	}
	if !strings.Contains(query, "percentile_approx") {
		t.Error("Query should use percentile_approx for quantiles")
	}
	if !strings.Contains(query, "total_duration_ms") {
		t.Error("Query should use total_duration_ms field")
	}
}

func TestBuildQueriesRunningQuery(t *testing.T) {
	query := BuildQueriesRunningQuery(1 * time.Hour)
	if !strings.Contains(query, "INTERVAL 1 HOUR") {
		t.Error("Query should contain INTERVAL 1 HOUR")
	}
	if !strings.Contains(query, "max_concurrent") {
		t.Error("Query should calculate max_concurrent")
	}
}

// ===== Query Properties Tests =====

func TestAllQueriesContainSelect(t *testing.T) {
	lookback := 2 * time.Hour
	billingLookback := 24 * time.Hour

	queries := []struct {
		name  string
		query string
	}{
		{"BuildBillingDBUsQuery", BuildBillingDBUsQuery(billingLookback)},
		{"BuildBillingCostEstimateQuery", BuildBillingCostEstimateQuery(billingLookback)},
		{"BuildPriceChangeEventsQuery", BuildPriceChangeEventsQuery(billingLookback)},
		{"BuildJobRunsQuery", BuildJobRunsQuery(lookback)},
		{"BuildJobRunStatusQuery", BuildJobRunStatusQuery(lookback)},
		{"BuildJobRunDurationQuery", BuildJobRunDurationQuery(lookback)},
		{"BuildTaskRetriesQuery", BuildTaskRetriesQuery(lookback)},
		{"BuildJobSLAMissQuery", BuildJobSLAMissQuery(lookback, 3600)},
		{"BuildPipelineRunsQuery", BuildPipelineRunsQuery(lookback)},
		{"BuildPipelineRunStatusQuery", BuildPipelineRunStatusQuery(lookback)},
		{"BuildPipelineRunDurationQuery", BuildPipelineRunDurationQuery(lookback)},
		{"BuildPipelineRetryEventsQuery", BuildPipelineRetryEventsQuery(lookback)},
		{"BuildPipelineFreshnessLagQuery", BuildPipelineFreshnessLagQuery(lookback)},
		{"BuildQueriesQuery", BuildQueriesQuery(lookback)},
		{"BuildQueryErrorsQuery", BuildQueryErrorsQuery(lookback)},
		{"BuildQueryDurationQuery", BuildQueryDurationQuery(lookback)},
		{"BuildQueriesRunningQuery", BuildQueriesRunningQuery(lookback)},
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

func TestAllQueriesContainTimeFilter(t *testing.T) {
	lookback := 2 * time.Hour
	billingLookback := 24 * time.Hour

	queries := []struct {
		name  string
		query string
	}{
		{"BuildBillingDBUsQuery", BuildBillingDBUsQuery(billingLookback)},
		{"BuildBillingCostEstimateQuery", BuildBillingCostEstimateQuery(billingLookback)},
		{"BuildPriceChangeEventsQuery", BuildPriceChangeEventsQuery(billingLookback)},
		{"BuildJobRunsQuery", BuildJobRunsQuery(lookback)},
		{"BuildJobRunStatusQuery", BuildJobRunStatusQuery(lookback)},
		{"BuildJobRunDurationQuery", BuildJobRunDurationQuery(lookback)},
		{"BuildTaskRetriesQuery", BuildTaskRetriesQuery(lookback)},
		{"BuildJobSLAMissQuery", BuildJobSLAMissQuery(lookback, 3600)},
		{"BuildPipelineRunsQuery", BuildPipelineRunsQuery(lookback)},
		{"BuildPipelineRunStatusQuery", BuildPipelineRunStatusQuery(lookback)},
		{"BuildPipelineRunDurationQuery", BuildPipelineRunDurationQuery(lookback)},
		{"BuildPipelineRetryEventsQuery", BuildPipelineRetryEventsQuery(lookback)},
		{"BuildPipelineFreshnessLagQuery", BuildPipelineFreshnessLagQuery(lookback)},
		{"BuildQueriesQuery", BuildQueriesQuery(lookback)},
		{"BuildQueryErrorsQuery", BuildQueryErrorsQuery(lookback)},
		{"BuildQueryDurationQuery", BuildQueryDurationQuery(lookback)},
		{"BuildQueriesRunningQuery", BuildQueriesRunningQuery(lookback)},
	}

	for _, tt := range queries {
		t.Run(tt.name, func(t *testing.T) {
			upper := strings.ToUpper(tt.query)
			hasTimeFilter := strings.Contains(upper, "INTERVAL") ||
				strings.Contains(upper, "CURRENT_DATE") ||
				strings.Contains(upper, "CURRENT_TIMESTAMP")

			if !hasTimeFilter {
				t.Errorf("%s does not contain time-based filter", tt.name)
			}
		})
	}
}

func TestAllQueriesContainAggregation(t *testing.T) {
	lookback := 2 * time.Hour
	billingLookback := 24 * time.Hour

	queries := []struct {
		name  string
		query string
	}{
		{"BuildBillingDBUsQuery", BuildBillingDBUsQuery(billingLookback)},
		{"BuildBillingCostEstimateQuery", BuildBillingCostEstimateQuery(billingLookback)},
		{"BuildPriceChangeEventsQuery", BuildPriceChangeEventsQuery(billingLookback)},
		{"BuildJobRunsQuery", BuildJobRunsQuery(lookback)},
		{"BuildJobRunStatusQuery", BuildJobRunStatusQuery(lookback)},
		{"BuildJobRunDurationQuery", BuildJobRunDurationQuery(lookback)},
		{"BuildTaskRetriesQuery", BuildTaskRetriesQuery(lookback)},
		{"BuildJobSLAMissQuery", BuildJobSLAMissQuery(lookback, 3600)},
		{"BuildPipelineRunsQuery", BuildPipelineRunsQuery(lookback)},
		{"BuildPipelineRunStatusQuery", BuildPipelineRunStatusQuery(lookback)},
		{"BuildPipelineRunDurationQuery", BuildPipelineRunDurationQuery(lookback)},
		{"BuildPipelineRetryEventsQuery", BuildPipelineRetryEventsQuery(lookback)},
		{"BuildPipelineFreshnessLagQuery", BuildPipelineFreshnessLagQuery(lookback)},
		{"BuildQueriesQuery", BuildQueriesQuery(lookback)},
		{"BuildQueryErrorsQuery", BuildQueryErrorsQuery(lookback)},
		{"BuildQueryDurationQuery", BuildQueryDurationQuery(lookback)},
		{"BuildQueriesRunningQuery", BuildQueriesRunningQuery(lookback)},
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

func TestQueriesContainCorrectSystemTable(t *testing.T) {
	lookback := 2 * time.Hour
	billingLookback := 24 * time.Hour

	tests := []struct {
		name      string
		query     string
		tableName string
	}{
		{"BuildBillingDBUsQuery", BuildBillingDBUsQuery(billingLookback), "system.billing.usage"},
		{"BuildBillingCostEstimateQuery", BuildBillingCostEstimateQuery(billingLookback), "system.billing.usage"},
		{"BuildPriceChangeEventsQuery", BuildPriceChangeEventsQuery(billingLookback), "system.billing.list_prices"},
		{"BuildJobRunsQuery", BuildJobRunsQuery(lookback), "system.lakeflow.job_run_timeline"},
		{"BuildJobRunStatusQuery", BuildJobRunStatusQuery(lookback), "system.lakeflow.job_run_timeline"},
		{"BuildJobRunDurationQuery", BuildJobRunDurationQuery(lookback), "system.lakeflow.job_run_timeline"},
		{"BuildTaskRetriesQuery", BuildTaskRetriesQuery(lookback), "system.lakeflow.job_task_run_timeline"},
		{"BuildJobSLAMissQuery", BuildJobSLAMissQuery(lookback, 3600), "system.lakeflow.job_run_timeline"},
		{"BuildPipelineRunsQuery", BuildPipelineRunsQuery(lookback), "system.lakeflow.pipeline_update_timeline"},
		{"BuildPipelineRunStatusQuery", BuildPipelineRunStatusQuery(lookback), "system.lakeflow.pipeline_update_timeline"},
		{"BuildPipelineRunDurationQuery", BuildPipelineRunDurationQuery(lookback), "system.lakeflow.pipeline_update_timeline"},
		{"BuildPipelineRetryEventsQuery", BuildPipelineRetryEventsQuery(lookback), "system.lakeflow.pipeline_update_timeline"},
		{"BuildPipelineFreshnessLagQuery", BuildPipelineFreshnessLagQuery(lookback), "system.lakeflow.pipeline_update_timeline"},
		{"BuildQueriesQuery", BuildQueriesQuery(lookback), "system.query.history"},
		{"BuildQueryErrorsQuery", BuildQueryErrorsQuery(lookback), "system.query.history"},
		{"BuildQueryDurationQuery", BuildQueryDurationQuery(lookback), "system.query.history"},
		{"BuildQueriesRunningQuery", BuildQueriesRunningQuery(lookback), "system.query.history"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(tt.query, tt.tableName) {
				t.Errorf("%s does not reference expected table %s", tt.name, tt.tableName)
			}
		})
	}
}

func TestQueriesContainWorkspaceID(t *testing.T) {
	lookback := 2 * time.Hour
	billingLookback := 24 * time.Hour

	queries := []struct {
		name          string
		query         string
		shouldContain bool
	}{
		{"BuildBillingDBUsQuery", BuildBillingDBUsQuery(billingLookback), true},
		{"BuildBillingCostEstimateQuery", BuildBillingCostEstimateQuery(billingLookback), true},
		{"BuildPriceChangeEventsQuery", BuildPriceChangeEventsQuery(billingLookback), false}, // No workspace_id
		{"BuildJobRunsQuery", BuildJobRunsQuery(lookback), true},
		{"BuildJobRunStatusQuery", BuildJobRunStatusQuery(lookback), true},
		{"BuildJobRunDurationQuery", BuildJobRunDurationQuery(lookback), true},
		{"BuildTaskRetriesQuery", BuildTaskRetriesQuery(lookback), true},
		{"BuildJobSLAMissQuery", BuildJobSLAMissQuery(lookback, 3600), true},
		{"BuildPipelineRunsQuery", BuildPipelineRunsQuery(lookback), true},
		{"BuildPipelineRunStatusQuery", BuildPipelineRunStatusQuery(lookback), true},
		{"BuildPipelineRunDurationQuery", BuildPipelineRunDurationQuery(lookback), true},
		{"BuildPipelineRetryEventsQuery", BuildPipelineRetryEventsQuery(lookback), true},
		{"BuildPipelineFreshnessLagQuery", BuildPipelineFreshnessLagQuery(lookback), true},
		{"BuildQueriesQuery", BuildQueriesQuery(lookback), true},
		{"BuildQueryErrorsQuery", BuildQueryErrorsQuery(lookback), true},
		{"BuildQueryDurationQuery", BuildQueryDurationQuery(lookback), true},
		{"BuildQueriesRunningQuery", BuildQueriesRunningQuery(lookback), true},
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
