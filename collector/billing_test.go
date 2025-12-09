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
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBillingCollector_Initialization(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err, "failed to create mock db")
	defer db.Close()

	metrics := NewMetricDescriptors()
	logger := log.NewNopLogger()

	collector := NewBillingCollector(logger, db, metrics, context.Background(), DefaultConfig())

	require.NotNil(t, collector, "NewBillingCollector returned nil")
	assert.Equal(t, db, collector.db, "db not set correctly")
	assert.Equal(t, metrics, collector.metrics, "metrics not set correctly")
}

func TestBillingCollector_CollectBillingDBUs(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "failed to create mock db")
	defer db.Close()

	// Set up mock expectations
	rows := sqlmock.NewRows([]string{"workspace_id", "sku_name", "dbus_total"}).
		AddRow("87654321", "STANDARD_ALL_PURPOSE_COMPUTE", 125.5).
		AddRow("87654321", "PREMIUM_JOBS_COMPUTE", 450.25).
		AddRow("87654322", "STANDARD_ALL_PURPOSE_COMPUTE", 89.75)

	mock.ExpectQuery("SELECT (.+) FROM system.billing.usage").
		WillReturnRows(rows)

	metrics := NewMetricDescriptors()
	logger := log.NewNopLogger()
	collector := NewBillingCollector(logger, db, metrics, context.Background(), DefaultConfig())

	// Collect metrics
	ch := make(chan prometheus.Metric, 10)
	err = collector.collectBillingDBUs(ch)
	close(ch)

	require.NoError(t, err, "collectBillingDBUs failed")

	// Verify metrics
	count := 0
	for m := range ch {
		count++

		// Extract metric details
		pb := &dto.Metric{}
		require.NoError(t, m.Write(pb), "failed to write metric")

		// Verify it's a gauge
		require.NotNil(t, pb.Gauge, "expected gauge metric")

		// Verify labels
		labels := make(map[string]string)
		for _, lp := range pb.Label {
			labels[lp.GetName()] = lp.GetValue()
		}

		assert.Contains(t, labels, "workspace_id", "missing workspace_id label")
		assert.Contains(t, labels, "sku_name", "missing sku_name label")

		// Verify value
		assert.Greater(t, pb.Gauge.GetValue(), float64(0), "expected positive value")
	}

	assert.Equal(t, 3, count, "expected 3 metrics")

	// Verify all expectations were met
	require.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
}

func TestBillingCollector_CollectBillingCost(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	// Set up mock expectations
	rows := sqlmock.NewRows([]string{"workspace_id", "sku_name", "cost_estimate_usd"}).
		AddRow("87654321", "STANDARD_ALL_PURPOSE_COMPUTE", 69.025).
		AddRow("87654321", "PREMIUM_JOBS_COMPUTE", 337.6875)

	mock.ExpectQuery("SELECT (.+) FROM system.billing.usage u").
		WillReturnRows(rows)

	metrics := NewMetricDescriptors()
	logger := log.NewNopLogger()
	collector := NewBillingCollector(logger, db, metrics, context.Background(), DefaultConfig())

	// Collect metrics
	ch := make(chan prometheus.Metric, 10)
	err = collector.collectBillingCost(ch)
	close(ch)

	if err != nil {
		t.Fatalf("collectBillingCost failed: %v", err)
	}

	// Verify metrics
	count := 0
	for m := range ch {
		count++

		pb := &dto.Metric{}
		if err := m.Write(pb); err != nil {
			t.Fatalf("failed to write metric: %v", err)
		}

		if pb.Gauge == nil {
			t.Error("expected gauge metric")
			continue
		}

		// Verify value is positive (cost estimate)
		if pb.Gauge.GetValue() <= 0 {
			t.Errorf("expected positive cost, got %f", pb.Gauge.GetValue())
		}
	}

	if count != 2 {
		t.Errorf("expected 2 metrics, got %d", count)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestBillingCollector_CollectPriceChangeEvents(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	// Set up mock expectations
	rows := sqlmock.NewRows([]string{"sku_name", "price_change_count"}).
		AddRow("STANDARD_ALL_PURPOSE_COMPUTE", 2.0).
		AddRow("PREMIUM_JOBS_COMPUTE", 2.0)

	mock.ExpectQuery("SELECT (.+) FROM system.billing.list_prices").
		WillReturnRows(rows)

	metrics := NewMetricDescriptors()
	logger := log.NewNopLogger()
	collector := NewBillingCollector(logger, db, metrics, context.Background(), DefaultConfig())

	// Collect metrics
	ch := make(chan prometheus.Metric, 10)
	err = collector.collectPriceChangeEvents(ch)
	close(ch)

	if err != nil {
		t.Fatalf("collectPriceChangeEvents failed: %v", err)
	}

	// Verify metrics
	count := 0
	for m := range ch {
		count++

		pb := &dto.Metric{}
		if err := m.Write(pb); err != nil {
			t.Fatalf("failed to write metric: %v", err)
		}

		// Verify it's a gauge (sliding window metric)
		if pb.Gauge == nil {
			t.Error("expected gauge metric")
			continue
		}

		// Verify labels
		labels := make(map[string]string)
		for _, lp := range pb.Label {
			labels[lp.GetName()] = lp.GetValue()
		}

		if _, ok := labels["sku_name"]; !ok {
			t.Error("missing sku_name label")
		}
	}

	if count != 2 {
		t.Errorf("expected 2 metrics, got %d", count)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestBillingCollector_CollectWithError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "failed to create mock db")
	defer db.Close()

	// Simulate query error
	mock.ExpectQuery("SELECT (.+) FROM system.billing.usage").
		WillReturnError(sql.ErrConnDone)

	metrics := NewMetricDescriptors()
	logger := log.NewNopLogger()
	collector := NewBillingCollector(logger, db, metrics, context.Background(), DefaultConfig())

	// Collect metrics
	ch := make(chan prometheus.Metric, 10)
	err = collector.collectBillingDBUs(ch)
	close(ch)

	require.Error(t, err, "expected error, got nil")

	// Verify no metrics were emitted
	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 0, count, "expected 0 metrics on error")
}

func TestBillingCollector_CollectEmitsErrorMetric(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	metrics := NewMetricDescriptors()
	logger := log.NewNopLogger()
	collector := NewBillingCollector(logger, db, metrics, context.Background(), DefaultConfig())

	// Test error emission
	ch := make(chan prometheus.Metric, 1)
	collector.emitError(ch, "test_stage")
	close(ch)

	count := 0
	for m := range ch {
		count++

		pb := &dto.Metric{}
		if err := m.Write(pb); err != nil {
			t.Fatalf("failed to write metric: %v", err)
		}

		// Verify it's a counter
		if pb.Counter == nil {
			t.Error("expected counter metric")
			continue
		}

		// Verify stage label
		labels := make(map[string]string)
		for _, lp := range pb.Label {
			labels[lp.GetName()] = lp.GetValue()
		}

		if stage, ok := labels["stage"]; !ok {
			t.Error("missing stage label")
		} else if stage != "test_stage" {
			t.Errorf("expected stage=test_stage, got stage=%s", stage)
		}

		// Verify value is 1
		if pb.Counter.GetValue() != 1 {
			t.Errorf("expected error count of 1, got %f", pb.Counter.GetValue())
		}
	}

	if count != 1 {
		t.Errorf("expected 1 error metric, got %d", count)
	}
}

func TestBillingCollector_CollectContinuesOnPartialFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	// First query fails
	mock.ExpectQuery("SELECT (.+) FROM system.billing.usage").
		WillReturnError(sql.ErrConnDone)

	// Second query succeeds
	costRows := sqlmock.NewRows([]string{"workspace_id", "sku_name", "cost_estimate_usd"}).
		AddRow("87654321", "STANDARD_ALL_PURPOSE_COMPUTE", 69.025)
	mock.ExpectQuery("SELECT (.+) FROM system.billing.usage u").
		WillReturnRows(costRows)

	// Third query succeeds
	priceRows := sqlmock.NewRows([]string{"sku_name", "price_change_count"}).
		AddRow("STANDARD_ALL_PURPOSE_COMPUTE", 2.0)
	mock.ExpectQuery("SELECT (.+) FROM system.billing.list_prices").
		WillReturnRows(priceRows)

	metrics := NewMetricDescriptors()
	logger := log.NewNopLogger()
	collector := NewBillingCollector(logger, db, metrics, context.Background(), DefaultConfig())

	// Collect all metrics - should continue despite first failure
	ch := make(chan prometheus.Metric, 20)
	collector.Collect(ch)
	close(ch)

	// Count metrics (should have cost + price + error metrics)
	count := 0
	errorCount := 0
	for m := range ch {
		count++

		pb := &dto.Metric{}
		if err := m.Write(pb); err != nil {
			continue
		}

		// Check for error metrics
		labels := make(map[string]string)
		for _, lp := range pb.Label {
			labels[lp.GetName()] = lp.GetValue()
		}

		if _, ok := labels["stage"]; ok {
			errorCount++
		}
	}

	// Should have at least 2 data metrics + 1 error metric
	if count < 3 {
		t.Errorf("expected at least 3 metrics (2 data + 1 error), got %d", count)
	}

	if errorCount == 0 {
		t.Error("expected at least one error metric")
	}
}
