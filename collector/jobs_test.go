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

func TestNewJobsCollector(t *testing.T) {
	logger := promslog.NewNopLogger()
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	metrics := NewMetricDescriptors()
	collector := NewJobsCollector(context.Background(), db, metrics, DefaultConfig(), logger)

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

func TestJobsCollector_Describe(t *testing.T) {
	logger := promslog.NewNopLogger()
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	metrics := NewMetricDescriptors()
	collector := NewJobsCollector(context.Background(), db, metrics, DefaultConfig(), logger)

	descCh := make(chan *prometheus.Desc, 10)
	go func() {
		collector.Describe(descCh)
		close(descCh)
	}()

	descriptions := make([]*prometheus.Desc, 0)
	for desc := range descCh {
		descriptions = append(descriptions, desc)
	}

	expectedCount := 5 // JobRuns, JobRunStatus, JobRunDuration, TaskRetries, JobSLAMiss
	if len(descriptions) != expectedCount {
		t.Errorf("expected %d metric descriptions, got %d", expectedCount, len(descriptions))
	}
}

func TestJobsCollector_CollectJobRuns(t *testing.T) {
	logger := promslog.NewNopLogger()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	// Mock query result
	rows := sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "run_count"}).
		AddRow("123456789", "job1", "Test Job 1", 150.0).
		AddRow("987654321", "job2", "Test Job 2", 75.0)

	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_run_timeline").WillReturnRows(rows)

	metrics := NewMetricDescriptors()
	collector := NewJobsCollector(context.Background(), db, metrics, DefaultConfig(), logger)

	// Create a registry and register the collector
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Gather metrics
	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	// Verify the job_runs_total metric was collected
	found := false
	for _, mf := range metricFamilies {
		if *mf.Name == "databricks_job_runs_total" {
			found = true
			if len(mf.Metric) != 2 {
				t.Errorf("expected 2 metrics, got %d", len(mf.Metric))
			}
		}
	}

	if !found {
		t.Error("expected databricks_job_runs_total metric not found")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestJobsCollector_CollectJobRunStatus(t *testing.T) {
	logger := promslog.NewNopLogger()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	// Mock all queries to prevent errors
	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_run_timeline").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "run_count"}))

	rows := sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "status", "run_count"}).
		AddRow("123456789", "job1", "Test Job 1", "SUCCESS", 120.0).
		AddRow("123456789", "job1", "Test Job 1", "FAILED", 10.0).
		AddRow("987654321", "job2", "Test Job 2", "SUCCESS", 60.0)

	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_run_timeline").WillReturnRows(rows)

	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_run_timeline").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "p50", "p95", "p99"}))
	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_task_run_timeline").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "task_key", "retry_count"}))
	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_run_timeline").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "sla_miss_count"}))

	metrics := NewMetricDescriptors()
	collector := NewJobsCollector(context.Background(), db, metrics, DefaultConfig(), logger)

	// Create a registry and register the collector
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Gather metrics
	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	// Verify the job_run_status_total metric was collected
	found := false
	for _, mf := range metricFamilies {
		if *mf.Name == "databricks_job_run_status_total" {
			found = true
			if len(mf.Metric) != 3 {
				t.Errorf("expected 3 metrics, got %d", len(mf.Metric))
			}
		}
	}

	if !found {
		t.Error("expected databricks_job_run_status_total metric not found")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestJobsCollector_CollectJobRunDuration(t *testing.T) {
	logger := promslog.NewNopLogger()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	// Mock all queries
	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_run_timeline").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "run_count"}))
	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_run_timeline").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "status", "run_count"}))

	rows := sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "p50", "p95", "p99"}).
		AddRow("123456789", "job1", "Test Job 1", 300.5, 850.2, 1200.8)

	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_run_timeline").WillReturnRows(rows)

	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_task_run_timeline").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "task_key", "retry_count"}))
	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_run_timeline").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "sla_miss_count"}))

	metrics := NewMetricDescriptors()
	collector := NewJobsCollector(context.Background(), db, metrics, DefaultConfig(), logger)

	// Create a registry and register the collector
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Gather metrics
	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	// Verify the job_run_duration_seconds metric was collected
	found := false
	for _, mf := range metricFamilies {
		if *mf.Name == "databricks_job_run_duration_seconds" {
			found = true
			if len(mf.Metric) != 3 {
				t.Errorf("expected 3 metrics (p50, p95, p99), got %d", len(mf.Metric))
			}
		}
	}

	if !found {
		t.Error("expected databricks_job_run_duration_seconds metric not found")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestJobsCollector_CollectWithError(t *testing.T) {
	logger := promslog.NewNopLogger()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	// Simulate a query error
	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_run_timeline").
		WillReturnError(errors.New("database connection lost"))

	// Mock remaining queries as empty
	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_run_timeline").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "status", "run_count"}))
	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_run_timeline").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "p50", "p95", "p99"}))
	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_task_run_timeline").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "task_key", "retry_count"}))
	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_run_timeline").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "sla_miss_count"}))

	metrics := NewMetricDescriptors()
	collector := NewJobsCollector(context.Background(), db, metrics, DefaultConfig(), logger)

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

func TestJobsCollector_CollectTaskRetries(t *testing.T) {
	logger := promslog.NewNopLogger()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	// Mock all queries
	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_run_timeline").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "run_count"}))
	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_run_timeline").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "status", "run_count"}))
	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_run_timeline").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "p50", "p95", "p99"}))

	rows := sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "task_key", "retry_count"}).
		AddRow("123456789", "job1", "Test Job 1", "task1", 25.0).
		AddRow("987654321", "job2", "Test Job 2", "task2", 12.0)

	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_task_run_timeline").WillReturnRows(rows)

	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_run_timeline").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "sla_miss_count"}))

	metrics := NewMetricDescriptors()
	collector := NewJobsCollector(context.Background(), db, metrics, DefaultConfig(), logger)

	ch := make(chan prometheus.Metric, 10)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	// Drain the channel
	for range ch {
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestJobsCollector_CollectJobSLAMiss(t *testing.T) {
	logger := promslog.NewNopLogger()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	// Mock all queries
	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_run_timeline").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "run_count"}))
	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_run_timeline").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "status", "run_count"}))
	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_run_timeline").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "p50", "p95", "p99"}))
	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_task_run_timeline").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "task_key", "retry_count"}))

	rows := sqlmock.NewRows([]string{"workspace_id", "job_id", "job_name", "sla_miss_count"}).
		AddRow("123456789", "job1", "Test Job 1", 5.0).
		AddRow("987654321", "job2", "Test Job 2", 2.0)

	mock.ExpectQuery("SELECT(.+)FROM system.lakeflow.job_run_timeline").WillReturnRows(rows)

	metrics := NewMetricDescriptors()
	collector := NewJobsCollector(context.Background(), db, metrics, DefaultConfig(), logger)

	// Use testutil to count metrics
	count := testutil.CollectAndCount(collector)

	// Should have metrics from all collectors
	if count < 2 {
		t.Errorf("expected at least 2 SLA miss metrics, got %d", count)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
