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
	"errors"
)

type Config struct {
	ServerHostname    string
	WarehouseHTTPPath string
	ClientID          string
	ClientSecret      string
}

var (
	errNoServerHostname    = errors.New("server_hostname must be specified")
	errNoWarehouseHTTPPath = errors.New("warehouse_http_path must be specified")
	errNoClientID          = errors.New("client_id must be specified")
	errNoClientSecret      = errors.New("client_secret must be specified")
)

func (c Config) Validate() error {
	if c.ServerHostname == "" {
		return errNoServerHostname
	}

	if c.WarehouseHTTPPath == "" {
		return errNoWarehouseHTTPPath
	}

	if c.ClientID == "" {
		return errNoClientID
	}

	if c.ClientSecret == "" {
		return errNoClientSecret
	}

	return nil
}
