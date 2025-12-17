package collector

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/promslog"
)

func TestNewSQLWarehouseCollector(t *testing.T) {
	logger := promslog.NewNopLogger()
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	metrics := NewMetricDescriptors()
	collector := NewSQLWarehouseCollector(context.Background(), db, metrics, DefaultConfig(), logger)

	if collector == nil {
		t.Fatal("expected collector to be created, got nil")
	}

	if collector.logger == nil {
		t.Error("collector logger should not be nil")
	}

	if collector.db == nil {
		t.Error("collector db should not be nil")
	}

	if collector.metrics == nil {
		t.Error("collector metrics should not be nil")
	}
}

func TestSQLWarehouseCollector_Describe(t *testing.T) {
	logger := promslog.NewNopLogger()
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	metrics := NewMetricDescriptors()
	collector := NewSQLWarehouseCollector(context.Background(), db, metrics, DefaultConfig(), logger)

	descCh := make(chan *prometheus.Desc, 10)
	go func() {
		collector.Describe(descCh)
		close(descCh)
	}()

	descriptions := make([]*prometheus.Desc, 0)
	for desc := range descCh {
		descriptions = append(descriptions, desc)
	}

	expectedCount := 5 // QueriesTotal, QueryDurationSeconds, QueryErrorsTotal, QueriesRunning, ScrapeStatus
	if len(descriptions) != expectedCount {
		t.Errorf("expected %d metric descriptions, got %d", expectedCount, len(descriptions))
	}
}

func TestSQLWarehouseCollector_CollectQueries(t *testing.T) {
	logger := promslog.NewNopLogger()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	// Mock query result
	rows := sqlmock.NewRows([]string{"workspace_id", "warehouse_id", "query_count"}).
		AddRow("123456789", "wh1", 5000.0).
		AddRow("987654321", "wh2", 3500.0)

	mock.ExpectQuery("SELECT(.+)FROM system.query.history").WillReturnRows(rows)

	metrics := NewMetricDescriptors()
	collector := NewSQLWarehouseCollector(context.Background(), db, metrics, DefaultConfig(), logger)

	// Create a registry and register the collector
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Gather metrics
	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	// Verify the queries_total metric was collected
	found := false
	for _, mf := range metricFamilies {
		if *mf.Name == "databricks_queries_total" {
			found = true
			if len(mf.Metric) != 2 {
				t.Errorf("expected 2 metrics, got %d", len(mf.Metric))
			}
		}
	}

	if !found {
		t.Error("expected databricks_queries_total metric not found")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestSQLWarehouseCollector_CollectQueryErrors(t *testing.T) {
	logger := promslog.NewNopLogger()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	// Mock all queries to prevent errors
	mock.ExpectQuery("SELECT(.+)FROM system.query.history").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "warehouse_id", "query_count"}))

	rows := sqlmock.NewRows([]string{"workspace_id", "warehouse_id", "error_count"}).
		AddRow("123456789", "wh1", 25.0).
		AddRow("987654321", "wh2", 12.0)

	mock.ExpectQuery("SELECT(.+)FROM system.query.history").WillReturnRows(rows)

	mock.ExpectQuery("SELECT(.+)FROM system.query.history").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "warehouse_id", "p50", "p95", "p99"}))
	mock.ExpectQuery("SELECT(.+)FROM system.query.history").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "warehouse_id", "max_concurrent"}))

	metrics := NewMetricDescriptors()
	collector := NewSQLWarehouseCollector(context.Background(), db, metrics, DefaultConfig(), logger)

	// Create a registry and register the collector
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Gather metrics
	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	// Verify the query_errors_total metric was collected
	found := false
	for _, mf := range metricFamilies {
		if *mf.Name == "databricks_query_errors_total" {
			found = true
			if len(mf.Metric) != 2 {
				t.Errorf("expected 2 metrics, got %d", len(mf.Metric))
			}
		}
	}

	if !found {
		t.Error("expected databricks_query_errors_total metric not found")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestSQLWarehouseCollector_CollectQueryDuration(t *testing.T) {
	logger := promslog.NewNopLogger()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	// Mock all queries
	mock.ExpectQuery("SELECT(.+)FROM system.query.history").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "warehouse_id", "query_count"}))
	mock.ExpectQuery("SELECT(.+)FROM system.query.history").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "warehouse_id", "error_count"}))

	rows := sqlmock.NewRows([]string{"workspace_id", "warehouse_id", "p50", "p95", "p99"}).
		AddRow("123456789", "wh1", 2.5, 15.8, 45.3)

	mock.ExpectQuery("SELECT(.+)FROM system.query.history").WillReturnRows(rows)

	mock.ExpectQuery("SELECT(.+)FROM system.query.history").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "warehouse_id", "max_concurrent"}))

	metrics := NewMetricDescriptors()
	collector := NewSQLWarehouseCollector(context.Background(), db, metrics, DefaultConfig(), logger)

	// Create a registry and register the collector
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Gather metrics
	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	// Verify the query_duration_seconds metric was collected
	found := false
	for _, mf := range metricFamilies {
		if *mf.Name == "databricks_query_duration_seconds" {
			found = true
			if len(mf.Metric) != 3 {
				t.Errorf("expected 3 metrics (p50, p95, p99), got %d", len(mf.Metric))
			}
		}
	}

	if !found {
		t.Error("expected databricks_query_duration_seconds metric not found")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestSQLWarehouseCollector_CollectWithError(t *testing.T) {
	logger := promslog.NewNopLogger()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	// Simulate a query error
	mock.ExpectQuery("SELECT(.+)FROM system.query.history").
		WillReturnError(errors.New("database connection lost"))

	// Mock remaining queries as empty
	mock.ExpectQuery("SELECT(.+)FROM system.query.history").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "warehouse_id", "error_count"}))
	mock.ExpectQuery("SELECT(.+)FROM system.query.history").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "warehouse_id", "p50", "p95", "p99"}))
	mock.ExpectQuery("SELECT(.+)FROM system.query.history").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "warehouse_id", "max_concurrent"}))

	metrics := NewMetricDescriptors()
	collector := NewSQLWarehouseCollector(context.Background(), db, metrics, DefaultConfig(), logger)

	ch := make(chan prometheus.Metric, 10)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	// Drain the channel
	count := 0
	for range ch {
		count++
	}

	// Should still collect other metrics despite the error
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestSQLWarehouseCollector_CollectQueriesRunning(t *testing.T) {
	logger := promslog.NewNopLogger()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	// Mock all queries
	mock.ExpectQuery("SELECT(.+)FROM system.query.history").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "warehouse_id", "query_count"}))
	mock.ExpectQuery("SELECT(.+)FROM system.query.history").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "warehouse_id", "error_count"}))
	mock.ExpectQuery("SELECT(.+)FROM system.query.history").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "warehouse_id", "p50", "p95", "p99"}))

	rows := sqlmock.NewRows([]string{"workspace_id", "warehouse_id", "max_concurrent"}).
		AddRow("123456789", "wh1", 15.0).
		AddRow("987654321", "wh2", 8.0)

	mock.ExpectQuery("SELECT(.+)FROM system.query.history").WillReturnRows(rows)

	metrics := NewMetricDescriptors()
	collector := NewSQLWarehouseCollector(context.Background(), db, metrics, DefaultConfig(), logger)

	// Use testutil to count metrics
	count := testutil.CollectAndCount(collector)

	// Should have metrics from all collectors
	if count < 2 {
		t.Errorf("expected at least 2 running query metrics, got %d", count)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
