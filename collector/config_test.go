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
	"testing"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		expectedErr error
	}{
		{
			name: "valid configuration",
			config: Config{
				ServerHostname: "dbc-abc123-def456.cloud.databricks.com",
				HTTPPath:       "/sql/1.0/warehouses/abc123",
				ClientID:       "test-client-id",
				ClientSecret:   "test-client-secret",
				Catalog:        "system",
				Schema:         "billing",
			},
			expectError: false,
		},
		{
			name: "missing server hostname",
			config: Config{
				HTTPPath:     "/sql/1.0/warehouses/abc123",
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				Catalog:      "system",
				Schema:       "billing",
			},
			expectError: true,
			expectedErr: errNoServerHostname,
		},
		{
			name: "missing http path",
			config: Config{
				ServerHostname: "dbc-abc123-def456.cloud.databricks.com",
				ClientID:       "test-client-id",
				ClientSecret:   "test-client-secret",
				Catalog:        "system",
				Schema:         "billing",
			},
			expectError: true,
			expectedErr: errNoHTTPPath,
		},
		{
			name: "missing client id",
			config: Config{
				ServerHostname: "dbc-abc123-def456.cloud.databricks.com",
				HTTPPath:       "/sql/1.0/warehouses/abc123",
				ClientSecret:   "test-client-secret",
				Catalog:        "system",
				Schema:         "billing",
			},
			expectError: true,
			expectedErr: errNoClientID,
		},
		{
			name: "missing client secret",
			config: Config{
				ServerHostname: "dbc-abc123-def456.cloud.databricks.com",
				HTTPPath:       "/sql/1.0/warehouses/abc123",
				ClientID:       "test-client-id",
				Catalog:        "system",
				Schema:         "billing",
			},
			expectError: true,
			expectedErr: errNoClientSecret,
		},
		{
			name: "missing catalog",
			config: Config{
				ServerHostname: "dbc-abc123-def456.cloud.databricks.com",
				HTTPPath:       "/sql/1.0/warehouses/abc123",
				ClientID:       "test-client-id",
				ClientSecret:   "test-client-secret",
				Schema:         "billing",
			},
			expectError: true,
			expectedErr: errNoCatalog,
		},
		{
			name: "missing schema",
			config: Config{
				ServerHostname: "dbc-abc123-def456.cloud.databricks.com",
				HTTPPath:       "/sql/1.0/warehouses/abc123",
				ClientID:       "test-client-id",
				ClientSecret:   "test-client-secret",
				Catalog:        "system",
			},
			expectError: true,
			expectedErr: errNoSchema,
		},
		{
			name: "empty server hostname",
			config: Config{
				ServerHostname: "",
				HTTPPath:       "/sql/1.0/warehouses/abc123",
				ClientID:       "test-client-id",
				ClientSecret:   "test-client-secret",
				Catalog:        "system",
				Schema:         "billing",
			},
			expectError: true,
			expectedErr: errNoServerHostname,
		},
		{
			name: "empty http path",
			config: Config{
				ServerHostname: "dbc-abc123-def456.cloud.databricks.com",
				HTTPPath:       "",
				ClientID:       "test-client-id",
				ClientSecret:   "test-client-secret",
				Catalog:        "system",
				Schema:         "billing",
			},
			expectError: true,
			expectedErr: errNoHTTPPath,
		},
		{
			name: "empty client id",
			config: Config{
				ServerHostname: "dbc-abc123-def456.cloud.databricks.com",
				HTTPPath:       "/sql/1.0/warehouses/abc123",
				ClientID:       "",
				ClientSecret:   "test-client-secret",
				Catalog:        "system",
				Schema:         "billing",
			},
			expectError: true,
			expectedErr: errNoClientID,
		},
		{
			name: "empty client secret",
			config: Config{
				ServerHostname: "dbc-abc123-def456.cloud.databricks.com",
				HTTPPath:       "/sql/1.0/warehouses/abc123",
				ClientID:       "test-client-id",
				ClientSecret:   "",
				Catalog:        "system",
				Schema:         "billing",
			},
			expectError: true,
			expectedErr: errNoClientSecret,
		},
		{
			name: "all fields empty",
			config: Config{
				ServerHostname: "",
				HTTPPath:       "",
				ClientID:       "",
				ClientSecret:   "",
				Catalog:        "",
				Schema:         "",
			},
			expectError: true,
			expectedErr: errNoServerHostname,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if tt.expectedErr != nil && err != tt.expectedErr {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestConfigValidationOrder(t *testing.T) {
	// Test that validation checks fields in a specific order
	// This ensures consistent error messages for users
	config := Config{}

	err := config.Validate()
	if err != errNoServerHostname {
		t.Errorf("expected first validation error to be errNoServerHostname, got %v", err)
	}

	config.ServerHostname = "test.databricks.com"
	err = config.Validate()
	if err != errNoHTTPPath {
		t.Errorf("expected second validation error to be errNoHTTPPath, got %v", err)
	}

	config.HTTPPath = "/sql/1.0/warehouses/test"
	err = config.Validate()
	if err != errNoClientID {
		t.Errorf("expected third validation error to be errNoClientID, got %v", err)
	}

	config.ClientID = "test-id"
	err = config.Validate()
	if err != errNoClientSecret {
		t.Errorf("expected fourth validation error to be errNoClientSecret, got %v", err)
	}

	config.ClientSecret = "test-secret"
	err = config.Validate()
	if err != errNoCatalog {
		t.Errorf("expected fifth validation error to be errNoCatalog, got %v", err)
	}

	config.Catalog = "system"
	err = config.Validate()
	if err != errNoSchema {
		t.Errorf("expected sixth validation error to be errNoSchema, got %v", err)
	}

	config.Schema = "billing"
	err = config.Validate()
	if err != nil {
		t.Errorf("expected no error after all fields set, got %v", err)
	}
}
