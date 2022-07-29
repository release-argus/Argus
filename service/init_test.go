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

package service

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/web/metrics"
)

func TestGetServiceInfo(t *testing.T) {
	// GIVEN a Service
	service := testServiceURL()
	id := "test_id"
	service.ID = id
	url := "https://test_url.com"
	service.LatestVersion.URL = url
	webURL := "https://test_webURL.com"
	service.Dashboard.WebURL = webURL
	latestVersion := "latest.version"
	service.Status.LatestVersion = latestVersion
	time.Sleep(10 * time.Millisecond)
	time.Sleep(time.Second)

	// When GetServiceInfo is called on it
	got := service.GetServiceInfo()
	want := utils.ServiceInfo{
		ID:            id,
		URL:           url,
		WebURL:        webURL,
		LatestVersion: latestVersion,
	}

	// THEN we get the correct ServiceInfo
	if got != want {
		t.Errorf("GetServiceInfo didn't get the correct data\nwant: %#v\ngot:  %#v",
			want, got)
	}
}

func TestServiceGetIconURL(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		icon   string
		want   string
		notify shoutrrr.Slice
	}{
		"no icon": {want: "", icon: ""},
		"no icon anywhere": {want: "", notify: shoutrrr.Slice{"test": &shoutrrr.Shoutrrr{
			Main:         &shoutrrr.Shoutrrr{},
			Defaults:     &shoutrrr.Shoutrrr{},
			HardDefaults: &shoutrrr.Shoutrrr{},
		}}},
		"emoji icon": {want: "", icon: ":smile:"},
		"web icon":   {want: "https://example.com/icon.png", icon: "https://example.com/icon.png"},
		"notify icon only": {want: "https://example.com/icon.png", notify: shoutrrr.Slice{"test": &shoutrrr.Shoutrrr{
			Params: map[string]string{
				"icon": "https://example.com/icon.png",
			},
			Main:         &shoutrrr.Shoutrrr{},
			Defaults:     &shoutrrr.Shoutrrr{},
			HardDefaults: &shoutrrr.Shoutrrr{},
		}}},
		"notify icon takes precedence over emoji": {want: "https://example.com/icon.png", icon: ":smile:",
			notify: shoutrrr.Slice{"test": &shoutrrr.Shoutrrr{
				Params: map[string]string{
					"icon": "https://example.com/icon.png",
				},
				Main:         &shoutrrr.Shoutrrr{},
				Defaults:     &shoutrrr.Shoutrrr{},
				HardDefaults: &shoutrrr.Shoutrrr{},
			}}},
		"dashboard icon takes precedence over notify icon": {want: "https://root.com/icon.png", icon: "https://root.com/icon.png",
			notify: shoutrrr.Slice{"test": &shoutrrr.Shoutrrr{
				Params: map[string]string{
					"icon": "https://example.com/icon.png",
				},
				Main:         &shoutrrr.Shoutrrr{},
				Defaults:     &shoutrrr.Shoutrrr{},
				HardDefaults: &shoutrrr.Shoutrrr{},
			}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			service := testServiceGitHub()
			service.Dashboard.Icon = tc.icon
			service.Notify = tc.notify

			// WHEN GetIconURL is called
			got := service.GetIconURL()

			// THEN the function returns the correct result
			if got != tc.want {
				t.Errorf("%s:\nwant: %q\ngot:  %q",
					name, tc.want, got)
			}
		})
	}
}

func TestInit(t *testing.T) {
	// GIVEN a Service
	service := testServiceURL()
	log := utils.NewJLog("WARN", false)
	var defaults Service
	var hardDefaults Service

	// WHEN Init is called on it
	hadC := testutil.CollectAndCount(metrics.LatestVersionQueryMetric)
	service.Init(log, &defaults, &hardDefaults)

	// THEN pointers to those vars are handed out to the Lookup
	// log
	if jLog != log {
		t.Errorf("JLog was not initialised from the Init\n want: %v\ngot:  %v",
			log, jLog)
	}
	// defaults
	if service.Defaults != &defaults {
		t.Errorf("Defaults were not handed to the Lookup correctly\n want: %v\ngot:  %v",
			&defaults, service.Defaults)
	}
	// dashboard.defaults
	if service.Dashboard.Defaults != &defaults.Dashboard {
		t.Errorf("Dashboard defaults were not handed to the Lookup correctly\n want: %v\ngot:  %v",
			&defaults.Dashboard, service.Dashboard.Defaults)
	}
	// options.defaults
	if service.Options.Defaults != &defaults.Options {
		t.Errorf("Options defaults were not handed to the Lookup correctly\n want: %v\ngot:  %v",
			&defaults.Options, service.Options.Defaults)
	}
	// hardDefaults
	if service.HardDefaults != &hardDefaults {
		t.Errorf("HardDefaults were not handed to the Lookup correctly\n want: %v\ngot:  %v",
			&hardDefaults, service.HardDefaults)
	}
	// dashboard.hardDefaults
	if service.Dashboard.HardDefaults != &hardDefaults.Dashboard {
		t.Errorf("Dashboard hardDefaults were not handed to the Lookup correctly\n want: %v\ngot:  %v",
			&hardDefaults.Dashboard, service.Dashboard.HardDefaults)
	}
	// options.hardDefaults
	if service.Options.HardDefaults != &hardDefaults.Options {
		t.Errorf("Options hardDefaults were not handed to the Lookup correctly\n want: %v\ngot:  %v",
			&hardDefaults.Options, service.Options.HardDefaults)
	}
	// initMetrics - counters
	gotC := testutil.CollectAndCount(metrics.LatestVersionQueryMetric)
	wantC := 2
	if (gotC - hadC) != wantC {
		t.Errorf("%d Counter metrics's were initialised, expecting %d",
			(gotC - hadC), wantC)
	}
}
