// Copyright [2024] [Argus]
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

	"github.com/release-argus/Argus/util"
)

func TestConfig_Load(t *testing.T) {
	// GIVEN Load is ran on a config
	file := "TestConfig_Load.yml"
	testYAML_config_test(file, t)
	config := testLoad(file, t)

	// WHEN the vars loaded are inspected
	tests := map[string]struct {
		got, want string
	}{
		"Defaults.Service.Interval": {
			got:  config.Defaults.Service.Options.Interval,
			want: "123s"},
		"Notify.discord.username": {
			got:  config.Defaults.Notify["slack"].GetParam("title"),
			want: "defaultTitle"},
		"WebHook.Delay": {
			got:  config.Defaults.WebHook.Delay,
			want: "2s"},
		"EmptyServiceIsDeleted": {
			got:  config.Service["EmptyService"].String(""),
			want: ""},
	}

	// THEN they match the config file
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			if tc.got != tc.want {
				t.Errorf("invalid %s:\nwant: %s\ngot:  %s",
					name, tc.want, tc.got)
			}
		})
	}
}

func TestConfig_LoadDeleteNil(t *testing.T) {
	// GIVEN config to Load
	var (
		config     Config
		configFile func(path string, t *testing.T) = testYAML_SomeNilServices
	)
	flags := make(map[string]bool)
	file := "TestConfig_LoadDeleteNil.yml"
	configFile(file, t)

	// WHEN Load is called on it
	config.Load(file, &flags, &util.JLog{})

	// THEN Services that are nil are deleted
	for name, service := range config.Service {
		if service == nil {
			t.Errorf("Service %q is nil",
				name)
		}
	}
	if len(config.Service) != 2 {
		t.Errorf("config.Service has %d entries, want 2",
			len(config.Service))
	}
}

func TestConfig_LoadDefaults(t *testing.T) {
	// GIVEN config to Load
	var (
		config     Config
		configFile func(path string, t *testing.T) = testYAML_config_test
	)
	flags := make(map[string]bool)
	file := "TestConfig_LoadDefaults.yml"
	configFile(file, t)

	// WHEN Load is called on it
	config.Load(file, &flags, &util.JLog{})

	// THEN the defaults are assigned correctly to Services
	want := false
	got := config.Service["WantDefaults"].Options.GetSemanticVersioning()
	if got != want {
		t.Errorf(`config.Service['WantDefaults'].Options.SemanticVersioning = %v. GetSemanticVersion gave %t, want %t`,
			got, *config.Service["WantDefaults"].Options.SemanticVersioning, want)
	}
}
