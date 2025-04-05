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

package config

import (
	"testing"
)

func TestConfig_Load(t *testing.T) {
	// GIVEN Load is ran on a config.
	file := "TestConfig_Load.yml"
	testYAML_config_test(file, t)
	config := testLoad(file, t)

	// WHEN the vars loaded are inspected.
	tests := map[string]struct {
		want, got string
	}{
		"Defaults.Service.Interval": {
			want: "123s",
			got:  config.Defaults.Service.Options.Interval},
		"Notify.discord.username": {
			want: "defaultTitle",
			got:  config.Defaults.Notify["slack"].GetParam("title")},
		"WebHook.Delay": {
			want: "2s",
			got:  config.Defaults.WebHook.Delay},
		"EmptyServiceIsDeleted": {
			want: "",
			got:  config.Service["EmptyService"].String("")},
	}

	// THEN they match the config file.
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			if tc.got != tc.want {
				t.Errorf("%s\ninvalid %s:\nwant: %s\ngot:  %s",
					packageName, name, tc.want, tc.got)
			}
		})
	}
}

func TestConfig_LoadDeleteNil(t *testing.T) {
	// GIVEN config to Load.
	var (
		config     Config
		configFile func(path string, t *testing.T) = testYAML_SomeNilServices
	)
	flags := make(map[string]bool)
	file := "TestConfig_LoadDeleteNil.yml"
	configFile(file, t)

	// WHEN Load is called on it.
	config.Load(file, &flags)

	// THEN Services that are nil are deleted.
	for name, service := range config.Service {
		if service == nil {
			t.Errorf("%s\nService %q is nil",
				packageName, name)
		}
	}
	if len(config.Service) != 2 {
		t.Errorf("%s\nlength mismatch\nwant: 2\ngot:  %d",
			packageName, len(config.Service))
	}
}

func TestConfig_LoadDefaults(t *testing.T) {
	// GIVEN config to Load.
	var (
		config     Config
		configFile func(path string, t *testing.T) = testYAML_config_test
	)
	flags := make(map[string]bool)
	file := "TestConfig_LoadDefaults.yml"
	configFile(file, t)

	// WHEN Load is called on it.
	config.Load(file, &flags)

	// THEN the defaults are assigned correctly to Services.
	want := false
	got := config.Service["WantDefaults"].Options.GetSemanticVersioning()
	if got != want {
		t.Errorf(`%s\nSemanticVersioning = %v\nwant: %t\ngot:  %t`,
			packageName, *config.Service["WantDefaults"].Options.SemanticVersioning, want, got)
	}
}
