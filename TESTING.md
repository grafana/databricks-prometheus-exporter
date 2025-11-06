# Testing

This document describes the test suite for the Databricks Prometheus Exporter.

## Running Tests

### Run all tests
```sh
go test ./...
```

### Run tests with verbose output
```sh
go test ./... -v
```

### Run tests with coverage
```sh
go test ./... -cover
```

### Run tests for a specific package
```sh
go test ./collector/... -v
```

## Test Coverage

Current test coverage:
- **collector package**: 55.1% of statements

## Test Structure

### Config Tests (`collector/config_test.go`)

Tests for configuration validation:
- **TestConfigValidate**: Comprehensive validation tests for all config fields
  - Valid configuration
  - Missing required fields (server hostname, HTTP path, client ID, client secret, catalog, schema)
  - Empty string validation
- **TestConfigValidationOrder**: Ensures validation errors occur in a predictable order

Total: 13 test cases

### Collector Tests (`collector/collector_test.go`)

Tests for the Prometheus collector implementation:
- **TestNewCollector**: Verifies collector initialization
- **TestCollectorDescribe**: Tests metric descriptor registration
- **TestCollectorCollect_DatabaseConnectionFailure**: Validates behavior when database connection fails
- **TestCollectorMetricNames**: Verifies correct metric names are registered
- **TestCollectorLabels**: Tests label constant definitions
- **TestNamespaceConstant**: Validates the namespace constant
- **TestOpenDatabricksDatabase_ValidatesConnection**: Tests database connector creation

Total: 7 test cases

### Query Tests (`collector/query_test.go`)

Tests for SQL query validation:
- **TestBillingMetricQuery**: Verifies query contains expected SQL keywords
- **TestBillingMetricQuery_TimeFilter**: Validates time filtering logic (7-day lookback)
- **TestBillingMetricQuery_Aggregation**: Tests AVG aggregation and GROUP BY clauses
- **TestBillingMetricQuery_SelectColumns**: Ensures all required columns are selected
- **TestBillingMetricQuery_ValidSQL**: Basic SQL syntax validation
- **TestBillingMetricQuery_TableReference**: Verifies correct table reference
- **TestBillingMetricQuery_TimeWindow**: Validates 7-day time window

Total: 7 test cases

## Test Philosophy

The test suite focuses on:
1. **Configuration Validation**: Ensuring invalid configurations are caught early
2. **Metric Registration**: Verifying metrics are properly defined and registered
3. **Error Handling**: Testing failure scenarios (connection failures, invalid configs)
4. **SQL Query Integrity**: Validating the billing metrics query structure

## Mocking

The tests use dependency injection to mock database connections:
- The `Collector` struct's `openDatabase` function can be replaced for testing
- This allows testing collector behavior without actual Databricks credentials

## Future Improvements

Potential areas for expanded test coverage:
- Integration tests with a test Databricks workspace
- End-to-end tests for the HTTP server
- Performance/load testing for metric collection
- Tests for the main command-line interface
