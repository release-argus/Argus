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

package config

import (
	"fmt"
	"testing"
)

func TestLoad(t *testing.T) {
	var (
		config                Config
		configFile            string = "config_test.yml"
		wantServiceInterval   string = "123s"
		wantNotifyTitle       string = "defaultTitle"
		wantWebHookDelay      string = "2s"
		wantServiceServiceURL string = "release-argus/argus"
	)
	flags := make(map[string]bool)
	config.Load(configFile, &flags)

	// Service
	gotServiceInterval := config.Defaults.Service.Interval
	if !(wantServiceInterval == *gotServiceInterval) {
		t.Fatalf(`config.Defaults.Service.Interval = %v, want match for %s`, *gotServiceInterval, wantServiceInterval)
	}

	// Notify
	gotNotifyTitle := (*config.Defaults.Notify["slack"].Params)["title"]
	if !(wantNotifyTitle == gotNotifyTitle) {
		t.Fatalf(`config.Defaults.Notify.Params.Title = %s, want match for %s`, gotNotifyTitle, wantNotifyTitle)
	}

	// WebHook
	gotWebHookDelay := config.Defaults.WebHook.Delay
	if !(wantWebHookDelay == *gotWebHookDelay) {
		t.Fatalf(`config.Defaults.WebHook.Delay = %s, want match for %s`, *gotWebHookDelay, wantWebHookDelay)
	}

	// Service
	gotServiceServiceURL := (config.Service)["NoDefaults"].URL
	if !(wantServiceServiceURL == *gotServiceServiceURL) {
		t.Fatalf(`config.Service[0].Service[0].URL = %s, want match for %s`, *gotServiceServiceURL, wantServiceServiceURL)
	}
}
func TestSetDefaults(t *testing.T) {
	var (
		config      Config
		configFile  string = "config_test.yml"
		wantBool    bool
		wantBoolPtr *bool
		wantString  string
	)
	flags := make(map[string]bool)
	flags["log.timestamps"] = false
	config.Load(configFile, &flags)
	wantBoolPtr = nil
	gotBoolPtr := config.Service["WantDefaults"].SemanticVersioning
	if !(wantBoolPtr == gotBoolPtr) {
		t.Fatalf(`pre-setDefaults config.Defaults.Service.SemanticVersioning = %v, want match for %t`, *gotBoolPtr, *wantBoolPtr)
	}

	wantBool = false
	gotBool := config.Service["WantDefaults"].GetSemanticVersioning()
	if !(wantBool == gotBool) {
		fmt.Printf("service: %v\n", config.Service["WantDefaults"].SemanticVersioning)
		fmt.Printf("default: %v\n", config.Service["WantDefaults"].Defaults.SemanticVersioning)
		fmt.Printf("hardDefault: %v\n", config.Service["WantDefaults"].HardDefaults.SemanticVersioning)

		t.Fatalf(`post-setDefaults config.Defaults.Service.SemanticVersioning = %v, want match for %t`, gotBool, wantBool)
	}

	wantString = "0.0.0.0"
	gotString := config.Settings.GetWebListenHost()
	if !(wantString == gotString) {
		t.Fatalf(`post-setDefaults config.Settings.Web.ListenHost = %v, want match for %s`, gotString, wantString)
	}
}
