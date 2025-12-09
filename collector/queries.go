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
	"fmt"
	"time"
)

// SQL queries for Databricks System Tables
// See: https://docs.databricks.com/aws/en/admin/system-tables

// durationToSQLInterval converts a Go duration to a Databricks SQL INTERVAL string.
// Examples: 24h -> "24 HOURS", 2h -> "2 HOURS", 30m -> "30 MINUTES"
func durationToSQLInterval(d time.Duration) string {
	hours := int(d.Hours())
	if hours >= 24 && hours%24 == 0 {
		days := hours / 24
		return fmt.Sprintf("%d DAY", days)
	}
	if hours > 0 {
		return fmt.Sprintf("%d HOURS", hours)
	}
	minutes := int(d.Minutes())
	return fmt.Sprintf("%d MINUTES", minutes)
}

// ===== Billing & Cost Queries =====

// BuildBillingDBUsQuery returns the query for DBU consumption with configurable lookback.
func BuildBillingDBUsQuery(lookback time.Duration) string {
	interval := durationToSQLInterval(lookback)
	return fmt.Sprintf(`
		SELECT 
			workspace_id,
			sku_name,
			SUM(usage_quantity) as dbus_total
		FROM system.billing.usage
		WHERE usage_date >= current_date() - INTERVAL %s
			AND workspace_id IS NOT NULL
			AND sku_name IS NOT NULL
		GROUP BY workspace_id, sku_name
		ORDER BY workspace_id, sku_name
	`, interval)
}

// BuildBillingCostEstimateQuery returns the query for cost estimates with configurable lookback.
// Uses current prices only (price_end_time IS NULL) to avoid expensive temporal JOINs.
func BuildBillingCostEstimateQuery(lookback time.Duration) string {
	interval := durationToSQLInterval(lookback)
	return fmt.Sprintf(`
		WITH current_prices AS (
			SELECT DISTINCT sku_name, cloud, pricing.default as unit_price
			FROM system.billing.list_prices
			WHERE price_end_time IS NULL
		)
		SELECT 
			u.workspace_id,
			u.sku_name,
			SUM(u.usage_quantity * COALESCE(p.unit_price, 0)) as cost_estimate_usd
		FROM system.billing.usage u
		LEFT JOIN current_prices p ON u.sku_name = p.sku_name AND u.cloud = p.cloud
		WHERE u.usage_date >= current_date() - INTERVAL %s
			AND u.workspace_id IS NOT NULL
			AND u.sku_name IS NOT NULL
		GROUP BY u.workspace_id, u.sku_name
		ORDER BY u.workspace_id, u.sku_name
	`, interval)
}

// BuildPriceChangeEventsQuery returns the query for price change events with configurable lookback.
func BuildPriceChangeEventsQuery(lookback time.Duration) string {
	interval := durationToSQLInterval(lookback)
	return fmt.Sprintf(`
		SELECT 
			sku_name,
			COUNT(*) as price_change_count
		FROM system.billing.list_prices
		WHERE price_start_time >= current_timestamp() - INTERVAL %s
			AND sku_name IS NOT NULL
		GROUP BY sku_name
		HAVING COUNT(*) > 1
		ORDER BY price_change_count DESC
	`, interval)
}

// Legacy constants for backwards compatibility (deprecated, use Build* functions)
const (
	billingDBUsQuery = `
		SELECT 
			workspace_id,
			sku_name,
			SUM(usage_quantity) as dbus_total
		FROM system.billing.usage
		WHERE usage_date >= current_date() - INTERVAL 1 DAY
			AND workspace_id IS NOT NULL
			AND sku_name IS NOT NULL
		GROUP BY workspace_id, sku_name
		ORDER BY workspace_id, sku_name
	`

	billingCostEstimateQuery = `
		WITH current_prices AS (
			SELECT DISTINCT sku_name, cloud, pricing.default as unit_price
			FROM system.billing.list_prices
			WHERE price_end_time IS NULL
		)
		SELECT 
			u.workspace_id,
			u.sku_name,
			SUM(u.usage_quantity * COALESCE(p.unit_price, 0)) as cost_estimate_usd
		FROM system.billing.usage u
		LEFT JOIN current_prices p ON u.sku_name = p.sku_name AND u.cloud = p.cloud
		WHERE u.usage_date >= current_date() - INTERVAL 1 DAY
			AND u.workspace_id IS NOT NULL
			AND u.sku_name IS NOT NULL
		GROUP BY u.workspace_id, u.sku_name
		ORDER BY u.workspace_id, u.sku_name
	`

	priceChangeEventsQuery = `
		SELECT 
			sku_name,
			COUNT(*) as price_change_count
		FROM system.billing.list_prices
		WHERE price_start_time >= current_timestamp() - INTERVAL 1 DAY
			AND sku_name IS NOT NULL
		GROUP BY sku_name
		HAVING COUNT(*) > 1
		ORDER BY price_change_count DESC
	`
)

// ===== Jobs Query Builders =====

// BuildJobRunsQuery returns the query for job run counts with configurable lookback.
func BuildJobRunsQuery(lookback time.Duration) string {
	interval := durationToSQLInterval(lookback)
	return fmt.Sprintf(`
		SELECT 
			t.workspace_id,
			t.job_id,
			COALESCE(j.name, 'unknown') as job_name,
			COUNT(*) as run_count
		FROM system.lakeflow.job_run_timeline t
		LEFT JOIN (
			SELECT workspace_id, job_id, name
			FROM system.lakeflow.jobs
			WHERE delete_time IS NULL
			QUALIFY ROW_NUMBER() OVER (PARTITION BY workspace_id, job_id ORDER BY change_time DESC) = 1
		) j ON t.workspace_id = j.workspace_id AND t.job_id = j.job_id
		WHERE t.period_start_time >= current_timestamp() - INTERVAL %s
		GROUP BY t.workspace_id, t.job_id, j.name
	`, interval)
}

// BuildJobRunStatusQuery returns the query for job status counts with configurable lookback.
func BuildJobRunStatusQuery(lookback time.Duration) string {
	interval := durationToSQLInterval(lookback)
	return fmt.Sprintf(`
		SELECT 
			t.workspace_id,
			t.job_id,
			COALESCE(j.name, 'unknown') as job_name,
			t.result_state as status,
			COUNT(*) as run_count
		FROM system.lakeflow.job_run_timeline t
		LEFT JOIN (
			SELECT workspace_id, job_id, name
			FROM system.lakeflow.jobs
			WHERE delete_time IS NULL
			QUALIFY ROW_NUMBER() OVER (PARTITION BY workspace_id, job_id ORDER BY change_time DESC) = 1
		) j ON t.workspace_id = j.workspace_id AND t.job_id = j.job_id
		WHERE t.period_start_time >= current_timestamp() - INTERVAL %s
			AND t.result_state IS NOT NULL
		GROUP BY t.workspace_id, t.job_id, j.name, t.result_state
	`, interval)
}

// BuildJobRunDurationQuery returns the query for job duration quantiles with configurable lookback.
func BuildJobRunDurationQuery(lookback time.Duration) string {
	interval := durationToSQLInterval(lookback)
	return fmt.Sprintf(`
		SELECT 
			workspace_id,
			job_id,
			job_name,
			percentile_approx(duration_seconds, 0.5) as p50,
			percentile_approx(duration_seconds, 0.95) as p95,
			percentile_approx(duration_seconds, 0.99) as p99
		FROM (
			SELECT 
				t.workspace_id,
				t.job_id,
				COALESCE(j.name, 'unknown') as job_name,
				unix_timestamp(t.period_end_time) - unix_timestamp(t.period_start_time) as duration_seconds
			FROM system.lakeflow.job_run_timeline t
			LEFT JOIN (
				SELECT workspace_id, job_id, name
				FROM system.lakeflow.jobs
				WHERE delete_time IS NULL
				QUALIFY ROW_NUMBER() OVER (PARTITION BY workspace_id, job_id ORDER BY change_time DESC) = 1
			) j ON t.workspace_id = j.workspace_id AND t.job_id = j.job_id
			WHERE t.period_start_time >= current_timestamp() - INTERVAL %s
				AND t.period_end_time IS NOT NULL
				AND t.period_end_time > t.period_start_time
		)
		GROUP BY workspace_id, job_id, job_name
	`, interval)
}

// BuildTaskRetriesQuery returns the query for task retries with configurable lookback.
func BuildTaskRetriesQuery(lookback time.Duration) string {
	interval := durationToSQLInterval(lookback)
	return fmt.Sprintf(`
		SELECT 
			t.workspace_id,
			t.job_id,
			COALESCE(j.name, 'unknown') as job_name,
			t.task_key,
			COUNT(*) - COUNT(DISTINCT CONCAT(t.job_run_id, '-', t.task_key)) as retry_count
		FROM system.lakeflow.job_task_run_timeline t
		LEFT JOIN (
			SELECT workspace_id, job_id, name
			FROM system.lakeflow.jobs
			WHERE delete_time IS NULL
			QUALIFY ROW_NUMBER() OVER (PARTITION BY workspace_id, job_id ORDER BY change_time DESC) = 1
		) j ON t.workspace_id = j.workspace_id AND t.job_id = j.job_id
		WHERE t.period_start_time >= current_timestamp() - INTERVAL %s
			AND t.job_run_id IS NOT NULL
		GROUP BY t.workspace_id, t.job_id, j.name, t.task_key
		HAVING COUNT(*) > COUNT(DISTINCT CONCAT(t.job_run_id, '-', t.task_key))
	`, interval)
}

// BuildJobSLAMissQuery returns the query for SLA misses with configurable lookback and threshold.
func BuildJobSLAMissQuery(lookback time.Duration, slaThresholdSeconds int) string {
	interval := durationToSQLInterval(lookback)
	return fmt.Sprintf(`
		SELECT 
			workspace_id,
			job_id,
			job_name,
			COUNT(*) as sla_miss_count
		FROM (
			SELECT 
				t.workspace_id,
				t.job_id,
				COALESCE(j.name, 'unknown') as job_name,
				unix_timestamp(t.period_end_time) - unix_timestamp(t.period_start_time) as duration_seconds
			FROM system.lakeflow.job_run_timeline t
			LEFT JOIN (
				SELECT workspace_id, job_id, name
				FROM system.lakeflow.jobs
				WHERE delete_time IS NULL
				QUALIFY ROW_NUMBER() OVER (PARTITION BY workspace_id, job_id ORDER BY change_time DESC) = 1
			) j ON t.workspace_id = j.workspace_id AND t.job_id = j.job_id
			WHERE t.period_start_time >= current_timestamp() - INTERVAL %s
				AND t.period_end_time IS NOT NULL
				AND t.period_end_time > t.period_start_time
		)
		WHERE duration_seconds > %d
		GROUP BY workspace_id, job_id, job_name
	`, interval, slaThresholdSeconds)
}

// ===== Pipelines Query Builders =====

// BuildPipelineRunsQuery returns the query for pipeline run counts with configurable lookback.
func BuildPipelineRunsQuery(lookback time.Duration) string {
	interval := durationToSQLInterval(lookback)
	return fmt.Sprintf(`
		SELECT 
			t.workspace_id,
			t.pipeline_id,
			COALESCE(p.name, 'unknown') as pipeline_name,
			COUNT(*) as run_count
		FROM system.lakeflow.pipeline_update_timeline t
		LEFT JOIN (
			SELECT workspace_id, pipeline_id, name
			FROM system.lakeflow.pipelines
			WHERE delete_time IS NULL
			QUALIFY ROW_NUMBER() OVER (PARTITION BY workspace_id, pipeline_id ORDER BY change_time DESC) = 1
		) p ON t.workspace_id = p.workspace_id AND t.pipeline_id = p.pipeline_id
		WHERE t.period_start_time >= current_timestamp() - INTERVAL %s
		GROUP BY t.workspace_id, t.pipeline_id, p.name
	`, interval)
}

// BuildPipelineRunStatusQuery returns the query for pipeline status counts with configurable lookback.
func BuildPipelineRunStatusQuery(lookback time.Duration) string {
	interval := durationToSQLInterval(lookback)
	return fmt.Sprintf(`
		SELECT 
			t.workspace_id,
			t.pipeline_id,
			COALESCE(p.name, 'unknown') as pipeline_name,
			t.result_state as status,
			COUNT(*) as run_count
		FROM system.lakeflow.pipeline_update_timeline t
		LEFT JOIN (
			SELECT workspace_id, pipeline_id, name
			FROM system.lakeflow.pipelines
			WHERE delete_time IS NULL
			QUALIFY ROW_NUMBER() OVER (PARTITION BY workspace_id, pipeline_id ORDER BY change_time DESC) = 1
		) p ON t.workspace_id = p.workspace_id AND t.pipeline_id = p.pipeline_id
		WHERE t.period_start_time >= current_timestamp() - INTERVAL %s
			AND t.result_state IS NOT NULL
		GROUP BY t.workspace_id, t.pipeline_id, p.name, t.result_state
	`, interval)
}

// BuildPipelineRunDurationQuery returns the query for pipeline duration quantiles with configurable lookback.
func BuildPipelineRunDurationQuery(lookback time.Duration) string {
	interval := durationToSQLInterval(lookback)
	return fmt.Sprintf(`
		SELECT 
			workspace_id,
			pipeline_id,
			pipeline_name,
			percentile_approx(duration_seconds, 0.5) as p50,
			percentile_approx(duration_seconds, 0.95) as p95,
			percentile_approx(duration_seconds, 0.99) as p99
		FROM (
			SELECT 
				t.workspace_id,
				t.pipeline_id,
				COALESCE(p.name, 'unknown') as pipeline_name,
				unix_timestamp(t.period_end_time) - unix_timestamp(t.period_start_time) as duration_seconds
			FROM system.lakeflow.pipeline_update_timeline t
			LEFT JOIN (
				SELECT workspace_id, pipeline_id, name
				FROM system.lakeflow.pipelines
				WHERE delete_time IS NULL
				QUALIFY ROW_NUMBER() OVER (PARTITION BY workspace_id, pipeline_id ORDER BY change_time DESC) = 1
			) p ON t.workspace_id = p.workspace_id AND t.pipeline_id = p.pipeline_id
			WHERE t.period_start_time >= current_timestamp() - INTERVAL %s
				AND t.period_end_time IS NOT NULL
				AND t.period_end_time > t.period_start_time
		)
		GROUP BY workspace_id, pipeline_id, pipeline_name
	`, interval)
}

// BuildPipelineRetryEventsQuery returns the query for pipeline retry events with configurable lookback.
func BuildPipelineRetryEventsQuery(lookback time.Duration) string {
	interval := durationToSQLInterval(lookback)
	return fmt.Sprintf(`
		SELECT 
			t.workspace_id,
			t.pipeline_id,
			COALESCE(p.name, 'unknown') as pipeline_name,
			COUNT(*) - COUNT(DISTINCT t.update_id) as retry_count
		FROM system.lakeflow.pipeline_update_timeline t
		LEFT JOIN (
			SELECT workspace_id, pipeline_id, name
			FROM system.lakeflow.pipelines
			WHERE delete_time IS NULL
			QUALIFY ROW_NUMBER() OVER (PARTITION BY workspace_id, pipeline_id ORDER BY change_time DESC) = 1
		) p ON t.workspace_id = p.workspace_id AND t.pipeline_id = p.pipeline_id
		WHERE t.period_start_time >= current_timestamp() - INTERVAL %s
		GROUP BY t.workspace_id, t.pipeline_id, p.name
		HAVING COUNT(*) > COUNT(DISTINCT t.update_id)
	`, interval)
}

// BuildPipelineFreshnessLagQuery returns the query for pipeline freshness lag with configurable lookback.
func BuildPipelineFreshnessLagQuery(lookback time.Duration) string {
	interval := durationToSQLInterval(lookback)
	return fmt.Sprintf(`
		SELECT 
			t.workspace_id,
			t.pipeline_id,
			COALESCE(p.name, 'unknown') as pipeline_name,
			AVG(unix_timestamp(current_timestamp()) - unix_timestamp(t.period_end_time)) as freshness_lag_seconds
		FROM system.lakeflow.pipeline_update_timeline t
		LEFT JOIN (
			SELECT workspace_id, pipeline_id, name
			FROM system.lakeflow.pipelines
			WHERE delete_time IS NULL
			QUALIFY ROW_NUMBER() OVER (PARTITION BY workspace_id, pipeline_id ORDER BY change_time DESC) = 1
		) p ON t.workspace_id = p.workspace_id AND t.pipeline_id = p.pipeline_id
		WHERE t.period_start_time >= current_timestamp() - INTERVAL %s
			AND t.period_end_time IS NOT NULL
			AND t.result_state = 'COMPLETED'
		GROUP BY t.workspace_id, t.pipeline_id, p.name
	`, interval)
}

// ===== SQL Warehouse Query Builders =====

// BuildQueriesQuery returns the query for SQL query counts with configurable lookback.
func BuildQueriesQuery(lookback time.Duration) string {
	interval := durationToSQLInterval(lookback)
	return fmt.Sprintf(`
		SELECT 
			workspace_id,
			COALESCE(compute.warehouse_id, 'unknown') as warehouse_id,
			COUNT(*) as query_count
		FROM system.query.history
		WHERE start_time >= current_timestamp() - INTERVAL %s
		GROUP BY workspace_id, compute.warehouse_id
	`, interval)
}

// BuildQueryErrorsQuery returns the query for SQL query errors with configurable lookback.
func BuildQueryErrorsQuery(lookback time.Duration) string {
	interval := durationToSQLInterval(lookback)
	return fmt.Sprintf(`
		SELECT 
			workspace_id,
			COALESCE(compute.warehouse_id, 'unknown') as warehouse_id,
			COUNT(*) as error_count
		FROM system.query.history
		WHERE start_time >= current_timestamp() - INTERVAL %s
			AND error_message IS NOT NULL
		GROUP BY workspace_id, compute.warehouse_id
	`, interval)
}

// BuildQueryDurationQuery returns the query for SQL query duration quantiles with configurable lookback.
func BuildQueryDurationQuery(lookback time.Duration) string {
	interval := durationToSQLInterval(lookback)
	return fmt.Sprintf(`
		SELECT 
			workspace_id,
			warehouse_id,
			percentile_approx(duration_seconds, 0.5) as p50,
			percentile_approx(duration_seconds, 0.95) as p95,
			percentile_approx(duration_seconds, 0.99) as p99
		FROM (
			SELECT 
				workspace_id,
				COALESCE(compute.warehouse_id, 'unknown') as warehouse_id,
				total_duration_ms / 1000.0 as duration_seconds
			FROM system.query.history
			WHERE start_time >= current_timestamp() - INTERVAL %s
				AND total_duration_ms IS NOT NULL
				AND total_duration_ms > 0
		)
		GROUP BY workspace_id, warehouse_id
	`, interval)
}

// BuildQueriesRunningQuery returns the query for concurrent queries estimate with configurable lookback.
func BuildQueriesRunningQuery(lookback time.Duration) string {
	interval := durationToSQLInterval(lookback)
	return fmt.Sprintf(`
		SELECT 
			workspace_id,
			warehouse_id,
			MAX(concurrent_count) as max_concurrent
		FROM (
			SELECT 
				q1.workspace_id,
				COALESCE(q1.compute.warehouse_id, 'unknown') as warehouse_id,
				q1.start_time as time_point,
				COUNT(*) as concurrent_count
			FROM system.query.history q1
			JOIN system.query.history q2
				ON q1.workspace_id = q2.workspace_id
				AND COALESCE(q1.compute.warehouse_id, 'unknown') = COALESCE(q2.compute.warehouse_id, 'unknown')
				AND q2.start_time <= q1.start_time
				AND (q2.end_time >= q1.start_time OR q2.end_time IS NULL)
			WHERE q1.start_time >= current_timestamp() - INTERVAL %s
			GROUP BY q1.workspace_id, q1.compute.warehouse_id, q1.start_time
		)
		GROUP BY workspace_id, warehouse_id
	`, interval)
}

// Legacy constants for backwards compatibility (deprecated, use Build* functions)
const (
	// ===== Jobs Queries (Lakeflow) =====

	// jobRunsQuery retrieves job execution counts per workspace and job.
	// Data from system.lakeflow.job_run_timeline over the last 24 hours.
	// JOINs with system.lakeflow.jobs to get job names.
	jobRunsQuery = `
		SELECT 
			t.workspace_id,
			t.job_id,
			COALESCE(j.name, 'unknown') as job_name,
			COUNT(*) as run_count
		FROM system.lakeflow.job_run_timeline t
		LEFT JOIN (
			SELECT workspace_id, job_id, name
			FROM system.lakeflow.jobs
			WHERE delete_time IS NULL
			QUALIFY ROW_NUMBER() OVER (PARTITION BY workspace_id, job_id ORDER BY change_time DESC) = 1
		) j ON t.workspace_id = j.workspace_id AND t.job_id = j.job_id
		WHERE t.period_start_time >= current_timestamp() - INTERVAL 2 HOURS
		GROUP BY t.workspace_id, t.job_id, j.name
	`

	// jobRunStatusQuery retrieves job status counts per workspace and job.
	// Breaks down by result_state (SUCCEEDED, FAILED, CANCELED, etc.)
	// JOINs with system.lakeflow.jobs to get job names.
	jobRunStatusQuery = `
		SELECT 
			t.workspace_id,
			t.job_id,
			COALESCE(j.name, 'unknown') as job_name,
			t.result_state as status,
			COUNT(*) as run_count
		FROM system.lakeflow.job_run_timeline t
		LEFT JOIN (
			SELECT workspace_id, job_id, name
			FROM system.lakeflow.jobs
			WHERE delete_time IS NULL
			QUALIFY ROW_NUMBER() OVER (PARTITION BY workspace_id, job_id ORDER BY change_time DESC) = 1
		) j ON t.workspace_id = j.workspace_id AND t.job_id = j.job_id
		WHERE t.period_start_time >= current_timestamp() - INTERVAL 2 HOURS
			AND t.result_state IS NOT NULL
		GROUP BY t.workspace_id, t.job_id, j.name, t.result_state
	`

	// jobRunDurationQuery calculates duration quantiles (p50, p95, p99) per job.
	// using Databricks' percentile_approx function for server-side aggregation.
	// This reduces memory footprint compared to client-side calculation.
	// JOINs with system.lakeflow.jobs to get job names.
	jobRunDurationQuery = `
		SELECT 
			workspace_id,
			job_id,
			job_name,
			percentile_approx(duration_seconds, 0.5) as p50,
			percentile_approx(duration_seconds, 0.95) as p95,
			percentile_approx(duration_seconds, 0.99) as p99
		FROM (
			SELECT 
				t.workspace_id,
				t.job_id,
				COALESCE(j.name, 'unknown') as job_name,
				unix_timestamp(t.period_end_time) - unix_timestamp(t.period_start_time) as duration_seconds
			FROM system.lakeflow.job_run_timeline t
			LEFT JOIN (
				SELECT workspace_id, job_id, name
				FROM system.lakeflow.jobs
				WHERE delete_time IS NULL
				QUALIFY ROW_NUMBER() OVER (PARTITION BY workspace_id, job_id ORDER BY change_time DESC) = 1
			) j ON t.workspace_id = j.workspace_id AND t.job_id = j.job_id
			WHERE t.period_start_time >= current_timestamp() - INTERVAL 2 HOURS
				AND t.period_end_time IS NOT NULL
				AND t.period_end_time > t.period_start_time
		)
		GROUP BY workspace_id, job_id, job_name
	`

	// taskRetriesQuery counts retries across job tasks from the task timeline table.
	// A task retry is detected when the same task_key within a job_run_id runs multiple times.
	// The difference between total task runs and unique (job_run_id, task_key) pairs = retry count.
	// JOINs with system.lakeflow.jobs to get job names.
	taskRetriesQuery = `
		SELECT 
			t.workspace_id,
			t.job_id,
			COALESCE(j.name, 'unknown') as job_name,
			t.task_key,
			COUNT(*) - COUNT(DISTINCT CONCAT(t.job_run_id, '-', t.task_key)) as retry_count
		FROM system.lakeflow.job_task_run_timeline t
		LEFT JOIN (
			SELECT workspace_id, job_id, name
			FROM system.lakeflow.jobs
			WHERE delete_time IS NULL
			QUALIFY ROW_NUMBER() OVER (PARTITION BY workspace_id, job_id ORDER BY change_time DESC) = 1
		) j ON t.workspace_id = j.workspace_id AND t.job_id = j.job_id
		WHERE t.period_start_time >= current_timestamp() - INTERVAL 2 HOURS
			AND t.job_run_id IS NOT NULL
		GROUP BY t.workspace_id, t.job_id, j.name, t.task_key
		HAVING COUNT(*) > COUNT(DISTINCT CONCAT(t.job_run_id, '-', t.task_key))
	`

	// jobSLAMissQuery identifies jobs exceeding a configured SLA threshold.
	// Currently hardcoded to 3600 seconds (1 hour) - should be configurable.
	// TODO: Make SLA threshold configurable per job or globally.
	// JOINs with system.lakeflow.jobs to get job names.
	jobSLAMissQuery = `
		SELECT 
			workspace_id,
			job_id,
			job_name,
			COUNT(*) as sla_miss_count
		FROM (
			SELECT 
				t.workspace_id,
				t.job_id,
				COALESCE(j.name, 'unknown') as job_name,
				unix_timestamp(t.period_end_time) - unix_timestamp(t.period_start_time) as duration_seconds
			FROM system.lakeflow.job_run_timeline t
			LEFT JOIN (
				SELECT workspace_id, job_id, name
				FROM system.lakeflow.jobs
				WHERE delete_time IS NULL
				QUALIFY ROW_NUMBER() OVER (PARTITION BY workspace_id, job_id ORDER BY change_time DESC) = 1
			) j ON t.workspace_id = j.workspace_id AND t.job_id = j.job_id
			WHERE t.period_start_time >= current_timestamp() - INTERVAL 2 HOURS
				AND t.period_end_time IS NOT NULL
				AND t.period_end_time > t.period_start_time
		)
		WHERE duration_seconds > 3600
		GROUP BY workspace_id, job_id, job_name
	`

	// ===== Pipelines Queries (Lakeflow) =====

	// pipelineRunsQuery retrieves pipeline execution counts per workspace and pipeline.
	// Data from system.lakeflow.pipeline_update_timeline over the last 24 hours.
	// JOINs with system.lakeflow.pipelines to get pipeline names.
	pipelineRunsQuery = `
		SELECT 
			t.workspace_id,
			t.pipeline_id,
			COALESCE(p.name, 'unknown') as pipeline_name,
			COUNT(*) as run_count
		FROM system.lakeflow.pipeline_update_timeline t
		LEFT JOIN (
			SELECT workspace_id, pipeline_id, name
			FROM system.lakeflow.pipelines
			WHERE delete_time IS NULL
			QUALIFY ROW_NUMBER() OVER (PARTITION BY workspace_id, pipeline_id ORDER BY change_time DESC) = 1
		) p ON t.workspace_id = p.workspace_id AND t.pipeline_id = p.pipeline_id
		WHERE t.period_start_time >= current_timestamp() - INTERVAL 2 HOURS
		GROUP BY t.workspace_id, t.pipeline_id, p.name
	`

	// pipelineRunStatusQuery retrieves pipeline status counts per workspace and pipeline.
	// Breaks down by result_state (COMPLETED, FAILED, etc.)
	// JOINs with system.lakeflow.pipelines to get pipeline names.
	pipelineRunStatusQuery = `
		SELECT 
			t.workspace_id,
			t.pipeline_id,
			COALESCE(p.name, 'unknown') as pipeline_name,
			t.result_state as status,
			COUNT(*) as run_count
		FROM system.lakeflow.pipeline_update_timeline t
		LEFT JOIN (
			SELECT workspace_id, pipeline_id, name
			FROM system.lakeflow.pipelines
			WHERE delete_time IS NULL
			QUALIFY ROW_NUMBER() OVER (PARTITION BY workspace_id, pipeline_id ORDER BY change_time DESC) = 1
		) p ON t.workspace_id = p.workspace_id AND t.pipeline_id = p.pipeline_id
		WHERE t.period_start_time >= current_timestamp() - INTERVAL 2 HOURS
			AND t.result_state IS NOT NULL
		GROUP BY t.workspace_id, t.pipeline_id, p.name, t.result_state
	`

	// pipelineRunDurationQuery calculates duration quantiles for pipelines.
	// Similar to job duration, uses server-side percentile calculation.
	// JOINs with system.lakeflow.pipelines to get pipeline names.
	pipelineRunDurationQuery = `
		SELECT 
			workspace_id,
			pipeline_id,
			pipeline_name,
			percentile_approx(duration_seconds, 0.5) as p50,
			percentile_approx(duration_seconds, 0.95) as p95,
			percentile_approx(duration_seconds, 0.99) as p99
		FROM (
			SELECT 
				t.workspace_id,
				t.pipeline_id,
				COALESCE(p.name, 'unknown') as pipeline_name,
				unix_timestamp(t.period_end_time) - unix_timestamp(t.period_start_time) as duration_seconds
			FROM system.lakeflow.pipeline_update_timeline t
			LEFT JOIN (
				SELECT workspace_id, pipeline_id, name
				FROM system.lakeflow.pipelines
				WHERE delete_time IS NULL
				QUALIFY ROW_NUMBER() OVER (PARTITION BY workspace_id, pipeline_id ORDER BY change_time DESC) = 1
			) p ON t.workspace_id = p.workspace_id AND t.pipeline_id = p.pipeline_id
			WHERE t.period_start_time >= current_timestamp() - INTERVAL 2 HOURS
				AND t.period_end_time IS NOT NULL
				AND t.period_end_time > t.period_start_time
		)
		GROUP BY workspace_id, pipeline_id, pipeline_name
	`

	// pipelineRetryEventsQuery counts retry events in pipeline updates.
	// Uses request_id to identify updates that had multiple retry attempts.
	// Multiple request_ids for the same update_id indicate retries/restarts.
	// JOINs with system.lakeflow.pipelines to get pipeline names.
	pipelineRetryEventsQuery = `
		SELECT 
			t.workspace_id,
			t.pipeline_id,
			COALESCE(p.name, 'unknown') as pipeline_name,
			COUNT(*) - COUNT(DISTINCT t.update_id) as retry_count
		FROM system.lakeflow.pipeline_update_timeline t
		LEFT JOIN (
			SELECT workspace_id, pipeline_id, name
			FROM system.lakeflow.pipelines
			WHERE delete_time IS NULL
			QUALIFY ROW_NUMBER() OVER (PARTITION BY workspace_id, pipeline_id ORDER BY change_time DESC) = 1
		) p ON t.workspace_id = p.workspace_id AND t.pipeline_id = p.pipeline_id
		WHERE t.period_start_time >= current_timestamp() - INTERVAL 2 HOURS
		GROUP BY t.workspace_id, t.pipeline_id, p.name
		HAVING COUNT(*) > COUNT(DISTINCT t.update_id)
	`

	// pipelineFreshnessLagQuery calculates data freshness lag per pipeline.
	// Measures average time since last completed pipeline run.
	// JOINs with system.lakeflow.pipelines to get pipeline names.
	pipelineFreshnessLagQuery = `
		SELECT 
			t.workspace_id,
			t.pipeline_id,
			COALESCE(p.name, 'unknown') as pipeline_name,
			AVG(unix_timestamp(current_timestamp()) - unix_timestamp(t.period_end_time)) as freshness_lag_seconds
		FROM system.lakeflow.pipeline_update_timeline t
		LEFT JOIN (
			SELECT workspace_id, pipeline_id, name
			FROM system.lakeflow.pipelines
			WHERE delete_time IS NULL
			QUALIFY ROW_NUMBER() OVER (PARTITION BY workspace_id, pipeline_id ORDER BY change_time DESC) = 1
		) p ON t.workspace_id = p.workspace_id AND t.pipeline_id = p.pipeline_id
		WHERE t.period_start_time >= current_timestamp() - INTERVAL 2 HOURS
			AND t.period_end_time IS NOT NULL
			AND t.result_state = 'COMPLETED'
		GROUP BY t.workspace_id, t.pipeline_id, p.name
	`

	// ===== SQL Warehouse Queries =====

	// queriesQuery retrieves total SQL query counts per workspace and warehouse.
	// Note: query.history can be high volume, limiting to last 1 hour.
	// Extracts warehouse_id from compute STRUCT.
	queriesQuery = `
		SELECT 
			workspace_id,
			COALESCE(compute.warehouse_id, 'unknown') as warehouse_id,
			COUNT(*) as query_count
		FROM system.query.history
		WHERE start_time >= current_timestamp() - INTERVAL 1 HOURS
		GROUP BY workspace_id, compute.warehouse_id
	`

	// queryErrorsQuery counts failed queries per workspace and warehouse.
	// A query is considered failed if error_message is not null.
	// Extracts warehouse_id from compute STRUCT.
	queryErrorsQuery = `
		SELECT 
			workspace_id,
			COALESCE(compute.warehouse_id, 'unknown') as warehouse_id,
			COUNT(*) as error_count
		FROM system.query.history
		WHERE start_time >= current_timestamp() - INTERVAL 1 HOURS
			AND error_message IS NOT NULL
		GROUP BY workspace_id, compute.warehouse_id
	`

	// queryDurationQuery calculates query latency quantiles per warehouse.
	// Uses total_duration_ms field from query.history.
	// Extracts warehouse_id from compute STRUCT.
	queryDurationQuery = `
		SELECT 
			workspace_id,
			warehouse_id,
			percentile_approx(duration_seconds, 0.5) as p50,
			percentile_approx(duration_seconds, 0.95) as p95,
			percentile_approx(duration_seconds, 0.99) as p99
		FROM (
			SELECT 
				workspace_id,
				COALESCE(compute.warehouse_id, 'unknown') as warehouse_id,
				total_duration_ms / 1000.0 as duration_seconds
			FROM system.query.history
			WHERE start_time >= current_timestamp() - INTERVAL 1 HOURS
				AND total_duration_ms IS NOT NULL
				AND total_duration_ms > 0
		)
		GROUP BY workspace_id, warehouse_id
	`

	// queriesRunningQuery estimates concurrent queries by finding overlapping intervals per warehouse.
	// This is an approximation - exact concurrency would require real-time polling.
	// Returns the maximum concurrent queries over the last hour.
	// Extracts warehouse_id from compute STRUCT.
	queriesRunningQuery = `
		SELECT 
			workspace_id,
			warehouse_id,
			MAX(concurrent_count) as max_concurrent
		FROM (
			SELECT 
				q1.workspace_id,
				COALESCE(q1.compute.warehouse_id, 'unknown') as warehouse_id,
				q1.start_time as time_point,
				COUNT(*) as concurrent_count
			FROM system.query.history q1
			JOIN system.query.history q2
				ON q1.workspace_id = q2.workspace_id
				AND COALESCE(q1.compute.warehouse_id, 'unknown') = COALESCE(q2.compute.warehouse_id, 'unknown')
				AND q2.start_time <= q1.start_time
				AND (q2.end_time >= q1.start_time OR q2.end_time IS NULL)
			WHERE q1.start_time >= current_timestamp() - INTERVAL 1 HOUR
			GROUP BY q1.workspace_id, q1.compute.warehouse_id, q1.start_time
		)
		GROUP BY workspace_id, warehouse_id
	`
)

// Time window constants for configurable lookback periods
const (
	PricesWindow  = "1 DAY"
	BillingWindow = "1 DAY"
	JobsWindow    = "2 HOURS"
	QueriesWindow = "1 HOURS"
)
