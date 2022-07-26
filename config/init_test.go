// Copyright [2022] [Argus]
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

	"github.com/release-argus/Argus/utils"
)

func TestLoad(t *testing.T) {
	// GIVEN Load is ran on a config
	config := testLoad("")

	// WHEN the vars loaded are inspected
	tests := map[string]struct {
		got  *string
		want string
	}{
		"Defaults.Service.Interval": {
			got: config.Defaults.Service.Interval, want: "123s"},
		"Notify.discord.username": {
			got: stringPtr(config.Defaults.Notify["slack"].GetSelfParam("title")), want: "defaultTitle"},
		"WebHook.Delay": {
			got: config.Defaults.WebHook.Delay, want: "2s"},
	}

	// THEN they match the config file
	for name, tc := range tests {
		if utils.EvalNilPtr(tc.got, "") != tc.want {
			t.Errorf("invalid %s:\nwant: %s\ngot:  %s",
				name, tc.want, utils.EvalNilPtr(tc.got, ""))
		}
	}
}

func TestLoadDefaults(t *testing.T) {
	// GIVEN config to Load
	var (
		config     Config
		configFile string = "../test/config_test.yml"
	)
	flags := make(map[string]bool)

	// WHEN Load is called on it
	config.Load(configFile, &flags, &utils.JLog{})

	// THEN the defaults are assigned correctly to Services
	want := false
	got := config.Service["WantDefaults"].Options.GetSemanticVersioning()
	if got != want {
		t.Errorf(`config.Service['WantDefaults'].SemanticVersioning = %v. GetSemanticVersion gave %t, want %t`,
			got, *config.Service["WantDefaults"].SemanticVersioning, want)
	}
}
