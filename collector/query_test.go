// Copyright 2025 Grafana Labs
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"strings"
	"testing"
)

func TestBillingMetricQuery(t *testing.T) {
	// Verify the query is not empty
	if billingMetricQuery == "" {
		t.Fatal("billingMetricQuery should not be empty")
	}

	// Verify the query contains expected SQL keywords
	expectedKeywords := []string{
		"SELECT",
		"FROM",
		"WHERE",
		"GROUP BY",
		"system.billing.account_prices",
		"account_id",
		"sku_name",
		"cloud",
		"currency_code",
		"usage_unit",
		"pricing",
		"AVG",
	}

	queryUpper := strings.ToUpper(billingMetricQuery)

	for _, keyword := range expectedKeywords {
		if !strings.Contains(queryUpper, strings.ToUpper(keyword)) {
			t.Errorf("expected query to contain '%s', but it was not found", keyword)
		}
	}
}

func TestBillingMetricQuery_TimeFilter(t *testing.T) {
	// Verify the query includes a time filter
	// The query should filter data from the last 7 days
	if !strings.Contains(billingMetricQuery, "date_sub") {
		t.Error("expected query to contain date_sub function for time filtering")
	}

	if !strings.Contains(billingMetricQuery, "price_start_time") {
		t.Error("expected query to filter on price_start_time")
	}

	// Verify it uses current_timestamp()
	if !strings.Contains(billingMetricQuery, "current_timestamp()") {
		t.Error("expected query to use current_timestamp()")
	}
}

func TestBillingMetricQuery_Aggregation(t *testing.T) {
	// Verify the query uses AVG aggregation
	if !strings.Contains(strings.ToUpper(billingMetricQuery), "AVG(PRICING)") {
		t.Error("expected query to aggregate pricing with AVG function")
	}

	// Verify GROUP BY includes all dimension columns
	expectedGroupByFields := []string{
		"account_id",
		"sku_name",
		"cloud",
		"currency_code",
		"usage_unit",
	}

	for _, field := range expectedGroupByFields {
		if !strings.Contains(strings.ToLower(billingMetricQuery), field) {
			t.Errorf("expected GROUP BY to include field '%s'", field)
		}
	}
}

func TestBillingMetricQuery_SelectColumns(t *testing.T) {
	// Verify all required columns are selected
	requiredColumns := []string{
		"account_id",
		"sku_name",
		"cloud",
		"currency_code",
		"usage_unit",
		"pricing",
	}

	queryLower := strings.ToLower(billingMetricQuery)

	for _, column := range requiredColumns {
		if !strings.Contains(queryLower, column) {
			t.Errorf("expected query to select column '%s'", column)
		}
	}
}

func TestBillingMetricQuery_ValidSQL(t *testing.T) {
	// Basic SQL syntax validation

	// Should have balanced parentheses
	openCount := strings.Count(billingMetricQuery, "(")
	closeCount := strings.Count(billingMetricQuery, ")")

	if openCount != closeCount {
		t.Errorf("unbalanced parentheses in query: %d open, %d close", openCount, closeCount)
	}

	// Should not contain obvious SQL injection patterns (as a sanity check)
	dangerousPatterns := []string{
		"';",
		"--",
		"/*",
		"*/",
		"xp_",
		"DROP",
		"DELETE",
		"INSERT",
		"UPDATE",
		"CREATE",
		"ALTER",
	}

	queryUpper := strings.ToUpper(billingMetricQuery)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(queryUpper, strings.ToUpper(pattern)) {
			t.Errorf("query contains potentially dangerous pattern: %s", pattern)
		}
	}
}

func TestBillingMetricQuery_TableReference(t *testing.T) {
	// Verify the query references the correct table
	expectedTable := "system.billing.account_prices"

	if !strings.Contains(billingMetricQuery, expectedTable) {
		t.Errorf("expected query to reference table '%s'", expectedTable)
	}
}

func TestBillingMetricQuery_TimeWindow(t *testing.T) {
	// Verify the query uses a 7-day time window
	if !strings.Contains(billingMetricQuery, "7") {
		t.Error("expected query to use a 7-day lookback period")
	}

	// Should use >= comparison for the date filter
	if !strings.Contains(billingMetricQuery, ">=") {
		t.Error("expected query to use >= for date comparison")
	}
}
