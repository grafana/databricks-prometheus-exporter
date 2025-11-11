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
				ServerHostname:    "dbc-abc123-def456.cloud.databricks.com",
				WarehouseHTTPPath: "/sql/1.0/warehouses/abc123",
				ClientID:          "test-client-id",
				ClientSecret:      "test-client-secret",
			},
			expectError: false,
		},
		{
			name: "missing server hostname",
			config: Config{
				WarehouseHTTPPath: "/sql/1.0/warehouses/abc123",
				ClientID:          "test-client-id",
				ClientSecret:      "test-client-secret",
			},
			expectError: true,
			expectedErr: errNoServerHostname,
		},
		{
			name: "missing warehouse http path",
			config: Config{
				ServerHostname: "dbc-abc123-def456.cloud.databricks.com",
				ClientID:       "test-client-id",
				ClientSecret:   "test-client-secret",
			},
			expectError: true,
			expectedErr: errNoWarehouseHTTPPath,
		},
		{
			name: "missing client id",
			config: Config{
				ServerHostname:    "dbc-abc123-def456.cloud.databricks.com",
				WarehouseHTTPPath: "/sql/1.0/warehouses/abc123",
				ClientSecret:      "test-client-secret",
			},
			expectError: true,
			expectedErr: errNoClientID,
		},
		{
			name: "missing client secret",
			config: Config{
				ServerHostname:    "dbc-abc123-def456.cloud.databricks.com",
				WarehouseHTTPPath: "/sql/1.0/warehouses/abc123",
				ClientID:          "test-client-id",
			},
			expectError: true,
			expectedErr: errNoClientSecret,
		},
		{
			name: "empty server hostname",
			config: Config{
				ServerHostname:    "",
				WarehouseHTTPPath: "/sql/1.0/warehouses/abc123",
				ClientID:          "test-client-id",
				ClientSecret:      "test-client-secret",
			},
			expectError: true,
			expectedErr: errNoServerHostname,
		},
		{
			name: "empty warehouse http path",
			config: Config{
				ServerHostname:    "dbc-abc123-def456.cloud.databricks.com",
				WarehouseHTTPPath: "",
				ClientID:          "test-client-id",
				ClientSecret:      "test-client-secret",
			},
			expectError: true,
			expectedErr: errNoWarehouseHTTPPath,
		},
		{
			name: "empty client id",
			config: Config{
				ServerHostname:    "dbc-abc123-def456.cloud.databricks.com",
				WarehouseHTTPPath: "/sql/1.0/warehouses/abc123",
				ClientID:          "",
				ClientSecret:      "test-client-secret",
			},
			expectError: true,
			expectedErr: errNoClientID,
		},
		{
			name: "empty client secret",
			config: Config{
				ServerHostname:    "dbc-abc123-def456.cloud.databricks.com",
				WarehouseHTTPPath: "/sql/1.0/warehouses/abc123",
				ClientID:          "test-client-id",
				ClientSecret:      "",
			},
			expectError: true,
			expectedErr: errNoClientSecret,
		},
		{
			name: "all fields empty",
			config: Config{
				ServerHostname:    "",
				WarehouseHTTPPath: "",
				ClientID:          "",
				ClientSecret:      "",
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
	if err != errNoWarehouseHTTPPath {
		t.Errorf("expected second validation error to be errNoWarehouseHTTPPath, got %v", err)
	}

	config.WarehouseHTTPPath = "/sql/1.0/warehouses/test"
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
	if err != nil {
		t.Errorf("expected no error after all fields set, got %v", err)
	}
}
