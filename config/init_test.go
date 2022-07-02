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
	"fmt"
	"testing"

	"github.com/release-argus/Argus/utils"
)

func testLoad(fileOverride string) Config {
	var (
		config     Config
		configFile string = "../test/config_test.yml"
	)
	if fileOverride != "" {
		configFile = fileOverride
	}

	flags := make(map[string]bool)
	config.Load(configFile, &flags, &utils.JLog{})

	return config
}

func TestLoad(t *testing.T) {
	// GIVEN Load is ran on a config.yml
	config := testLoad("")

	// WHEN the Default Service Interval is looked at
	got := config.Defaults.Service.Interval

	// THEN it matches the config.yml
	want := "123s"
	if !(want == *got) {
		t.Errorf(`config.Defaults.Service.Interval = %v, want %s`, *got, want)
	}
}

func TestLoadNotify(t *testing.T) {
	// GIVEN Load is ran on a config.yml
	config := testLoad("")

	// WHEN the Default Notify Param.title is looked at for slack
	got := config.Defaults.Notify["slack"].GetSelfParam("title")

	// THEN it matches the config.yml
	want := "defaultTitle"
	if !(want == got) {
		fmt.Println(config.Defaults.Notify["slack"])
		t.Errorf(`config.Defaults.Notify["slack"].Params.Title = %q, want %q`, got, want)
	}
}

func TestLoadWebHook(t *testing.T) {
	// GIVEN Load is ran on a config.yml
	config := testLoad("")

	// WHEN the Default WebHook Delay is looked at
	got := config.Defaults.WebHook.Delay

	// THEN it matches the config.yml
	want := "2s"
	if !(want == *got) {
		t.Errorf(`config.Defaults.WebHook.Delay = %s, want %s`, *got, want)
	}
}

func TestLoadDefaultsService(t *testing.T) {
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
	got := config.Service["WantDefaults"].GetSemanticVersioning()
	if got != want {
		t.Errorf(`config.Service['WantDefaults'].SemanticVersioning = %v. GetSemanticVersion gave %t, want %t`,
			got, *config.Service["WantDefaults"].SemanticVersioning, want)
	}
}

// func TestLoadDefaultsService(t *testing.T) {
// 	// GIVEN config to Load
// 	var (
// 		config     Config
// 		configFile string = "../test/config_test.yml"
// 	)
// 	wantBool = false
// 	gotBool := config.Service["WantDefaults"].GetSemanticVersioning()
// 	if !(wantBool == gotBool) {
// 		fmt.Printf("service: %v\n", config.Service["WantDefaults"].SemanticVersioning)
// 		fmt.Printf("default: %v\n", config.Service["WantDefaults"].Defaults.SemanticVersioning)
// 		fmt.Printf("hardDefaults: %v\n", config.Service["WantDefaults"].HardDefaults.SemanticVersioning)

// 		t.Errorf(`post-setDefaults config.Defaults.Service.SemanticVersioning = %v, want match for %t`, gotBool, wantBool)
// 	}

// 	wantString = "0.0.0.0"
// 	gotString := config.Settings.GetWebListenHost()
// 	if !(wantString == gotString) {
// 		t.Errorf(`post-setDefaults config.Settings.Web.ListenHost = %v, want match for %s`, gotString, wantString)
// 	}
// }
