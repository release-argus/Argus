// Copyright [2025] [Argus]
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build unit

// Package base provides the base struct for deployed_version lookups.
package base

import (
	"testing"
)

func TestDefaults_Default(t *testing.T) {
	// GIVEN a Default.
	defaults := Defaults{}

	// WHEN Default is called.
	defaults.Default()

	// THEN it should set the defaults.
	if defaults.AllowInvalidCerts == nil {
		t.Errorf("AllowInvalidCerts not set, got %v",
			defaults.AllowInvalidCerts)
	}
}
