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

const (
	// billingMetricQuery retrieves billing and usage information from the system.billing.usage table
	// This query aggregates usage data by account, workspace, SKU, cloud provider, and usage unit
	billingMetricQuery = `
		SELECT 
			account_id,
			workspace_id,
			sku_name,
			cloud,
			usage_unit,
			SUM(usage_quantity) as usage_quantity
		FROM system.billing.usage
		WHERE usage_date >= date_sub(current_date(), 7)
		GROUP BY account_id, workspace_id, sku_name, cloud, usage_unit
	`
)
