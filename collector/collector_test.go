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
	"database/sql"
	"errors"
	"testing"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

func TestNewCollector(t *testing.T) {
	logger := log.NewNopLogger()
	config := &Config{
		ServerHostname:    "test.databricks.com",
		WarehouseHTTPPath: "/sql/1.0/warehouses/test",
		ClientID:          "test-id",
		ClientSecret:      "test-secret",
	}

	collector := NewCollector(logger, config)

	if collector == nil {
		t.Fatal("expected collector to be created, got nil")
	}

	if collector.config != config {
		t.Error("collector config does not match provided config")
	}

	if collector.logger == nil {
		t.Error("collector logger should not be nil")
	}

	if collector.openDatabase == nil {
		t.Error("collector openDatabase function should not be nil")
	}

	if collector.metrics == nil {
		t.Error("collector metrics should not be nil")
	}
}

func TestCollectorDescribe(t *testing.T) {
	logger := log.NewNopLogger()
	config := &Config{
		ServerHostname:    "test.databricks.com",
		WarehouseHTTPPath: "/sql/1.0/warehouses/test",
		ClientID:          "test-id",
		ClientSecret:      "test-secret",
	}

	collector := NewCollector(logger, config)

	descCh := make(chan *prometheus.Desc, 10)
	go func() {
		collector.Describe(descCh)
		close(descCh)
	}()

	descriptions := make([]*prometheus.Desc, 0)
	for desc := range descCh {
		descriptions = append(descriptions, desc)
	}

	// Should have all 20 metrics (18 main + 2 future billing metrics + 1 up metric)
	expectedCount := 20
	if len(descriptions) != expectedCount {
		t.Errorf("expected %d metric descriptions, got %d", expectedCount, len(descriptions))
	}
}

func TestCollectorCollect_DatabaseConnectionFailure(t *testing.T) {
	logger := log.NewNopLogger()
	config := &Config{
		ServerHostname:    "test.databricks.com",
		WarehouseHTTPPath: "/sql/1.0/warehouses/test",
		ClientID:          "test-id",
		ClientSecret:      "test-secret",
	}

	collector := NewCollector(logger, config)

	// Mock the openDatabase function to return an error
	collector.openDatabase = func(c *Config) (*sql.DB, error) {
		return nil, errors.New("connection failed")
	}

	// Create a registry and gather metrics
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	// Find the up metric
	var upValue float64
	found := false
	for _, mf := range metricFamilies {
		if *mf.Name == "databricks_up" {
			if len(mf.Metric) > 0 {
				upValue = *mf.Metric[0].Gauge.Value
				found = true
			}
			break
		}
	}

	if !found {
		t.Fatal("up metric not found")
	}

	if upValue != 0 {
		t.Errorf("expected up metric to be 0, got %f", upValue)
	}
}

func TestCollectorMetricNames(t *testing.T) {
	logger := log.NewNopLogger()
	config := &Config{
		ServerHostname:    "test.databricks.com",
		WarehouseHTTPPath: "/sql/1.0/warehouses/test",
		ClientID:          "test-id",
		ClientSecret:      "test-secret",
	}

	collector := NewCollector(logger, config)

	// Mock the openDatabase function to prevent actual connection attempts
	collector.openDatabase = func(c *Config) (*sql.DB, error) {
		return nil, errors.New("mocked error - no actual connection")
	}

	// Check that metric descriptors have correct names
	expectedUpName := "databricks_up"

	// Create a temporary registry to extract metric info
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	foundUp := false

	for _, mf := range metricFamilies {
		if *mf.Name == expectedUpName {
			foundUp = true
		}
	}

	if !foundUp {
		t.Errorf("expected to find metric %s", expectedUpName)
	}

	// Note: databricks_billing_account_price will only appear if data is successfully collected
	// Since we're mocking a connection failure, we only expect the 'up' metric
}

func TestCollectorLabels(t *testing.T) {
	expectedLabels := []string{
		labelAccountID,
		labelWorkspaceID,
		labelSKUName,
		labelCloud,
		labelUsageUnit,
	}

	// Verify labels are correctly defined
	for _, label := range expectedLabels {
		if label == "" {
			t.Errorf("label should not be empty")
		}
	}

	// Verify expected label values
	if labelAccountID != "account_id" {
		t.Errorf("expected labelAccountID to be 'account_id', got '%s'", labelAccountID)
	}
	if labelWorkspaceID != "workspace_id" {
		t.Errorf("expected labelWorkspaceID to be 'workspace_id', got '%s'", labelWorkspaceID)
	}
	if labelSKUName != "sku_name" {
		t.Errorf("expected labelSKUName to be 'sku_name', got '%s'", labelSKUName)
	}
	if labelCloud != "cloud" {
		t.Errorf("expected labelCloud to be 'cloud', got '%s'", labelCloud)
	}
	if labelUsageUnit != "usage_unit" {
		t.Errorf("expected labelUsageUnit to be 'usage_unit', got '%s'", labelUsageUnit)
	}
}

func TestNamespaceConstant(t *testing.T) {
	if namespace != "databricks" {
		t.Errorf("expected namespace to be 'databricks', got '%s'", namespace)
	}
}

func TestOpenDatabricksDatabase_ValidatesConnection(t *testing.T) {
	// Test that openDatabricksDatabase creates a connector
	// We can't test actual connection without valid credentials,
	// but we can verify the function signature and basic behavior
	config := &Config{
		ServerHostname:    "test.cloud.databricks.com",
		WarehouseHTTPPath: "/sql/1.0/warehouses/test123",
		ClientID:          "test-client-id",
		ClientSecret:      "test-secret",
	}

	// This will attempt to create a connector but should return without error
	// The actual connection/authentication will fail when the DB is used
	db, err := openDatabricksDatabase(config)

	// We expect no error during connector creation (errors happen on use)
	if err != nil {
		t.Errorf("unexpected error creating connector: %v", err)
	}

	if db != nil {
		db.Close()
	}
}
