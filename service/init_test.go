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

	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

func TestService_GetServiceInfo(t *testing.T) {
	// GIVEN a Service
	svc := testServiceURL("TestGetServiceInfo")
	id := "test_id"
	svc.ID = id
	url := "https://test_url.com"
	svc.LatestVersion.URL = url
	webURL := "https://test_webURL.com"
	svc.Dashboard.WebURL = webURL
	latestVersion := "latest.version"
	svc.Status.SetLatestVersion(latestVersion, false)
	time.Sleep(10 * time.Millisecond)
	time.Sleep(time.Second)

	// When GetServiceInfo is called on it
	got := svc.GetServiceInfo()
	want := util.ServiceInfo{
		ID:            id,
		URL:           url,
		WebURL:        webURL,
		LatestVersion: latestVersion,
	}

	// THEN we get the correct ServiceInfo
	if *got != want {
		t.Errorf("GetServiceInfo didn't get the correct data\nwant: %#v\ngot:  %#v",
			want, got)
	}
}

func TestService_GetIconURL(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		dashboardIcon string
		want          string
		notify        shoutrrr.Slice
	}{
		"no dashboard.icon": {
			want:          "",
			dashboardIcon: "",
		},
		"no icon anywhere": {
			want:          "",
			dashboardIcon: "",
			notify: shoutrrr.Slice{"test": &shoutrrr.Shoutrrr{
				Main:         &shoutrrr.Shoutrrr{},
				Defaults:     &shoutrrr.Shoutrrr{},
				HardDefaults: &shoutrrr.Shoutrrr{},
			}},
		},
		"emoji icon": {
			want:          "",
			dashboardIcon: ":smile:",
		},
		"web icon": {
			want:          "https://example.com/icon.png",
			dashboardIcon: "https://example.com/icon.png",
		},
		"notify icon only": {
			want: "https://example.com/icon.png",
			notify: shoutrrr.Slice{"test": {
				Params: map[string]string{
					"icon": "https://example.com/icon.png",
				},
				Main:         &shoutrrr.Shoutrrr{},
				Defaults:     &shoutrrr.Shoutrrr{},
				HardDefaults: &shoutrrr.Shoutrrr{},
			}},
		},
		"notify icon takes precedence over emoji": {
			want:          "https://example.com/icon.png",
			dashboardIcon: ":smile:",
			notify: shoutrrr.Slice{"test": {
				Params: map[string]string{
					"icon": "https://example.com/icon.png",
				},
				Main:         &shoutrrr.Shoutrrr{},
				Defaults:     &shoutrrr.Shoutrrr{},
				HardDefaults: &shoutrrr.Shoutrrr{},
			}},
		},
		"dashboard icon takes precedence over notify icon": {
			want:          "https://root.com/icon.png",
			dashboardIcon: "https://root.com/icon.png",
			notify: shoutrrr.Slice{"test": {
				Params: map[string]string{
					"icon": "https://example.com/icon.png",
				},
				Main:         &shoutrrr.Shoutrrr{},
				Defaults:     &shoutrrr.Shoutrrr{},
				HardDefaults: &shoutrrr.Shoutrrr{},
			}},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		svc := testServiceGitHub(name)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc.Dashboard.Icon = tc.dashboardIcon
			svc.Notify = tc.notify

			// WHEN GetIconURL is called
			got := svc.GetIconURL()

			// THEN the function returns the correct result
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestService_Init(t *testing.T) {
	// GIVEN a Service
	testLogging()
	tests := map[string]struct {
		svc *Service
	}{
		"bare service": {
			svc: &Service{
				ID: "Init",
				LatestVersion: latestver.Lookup{
					Type: "github", URL: "release-argus/Argus"}},
		},
		"service with notify, command and webhook": {
			svc: &Service{
				ID: "Init",
				LatestVersion: latestver.Lookup{
					Type: "github", URL: "release-argus/Argus"},
				Notify: shoutrrr.Slice{
					"test": &shoutrrr.Shoutrrr{Type: "discord"}},
				Command: command.Slice{
					{"ls"}},
				WebHook: webhook.Slice{
					"test": testWebHook(false)}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			var defaults Service
			var hardDefaults Service
			tc.svc.ID = name

			// WHEN Init is called on it
			tc.svc.Init(
				&defaults, &hardDefaults,
				&shoutrrr.Slice{}, &shoutrrr.Slice{}, &shoutrrr.Slice{},
				&webhook.Slice{}, &webhook.WebHook{}, &webhook.WebHook{})

			// THEN pointers to those vars are handed out to the Lookup
			// defaults
			if tc.svc.Defaults != &defaults {
				t.Errorf("Defaults were not handed to the Lookup correctly\n want: %v\ngot:  %v",
					&defaults, tc.svc.Defaults)
			}
			// dashboard.defaults
			if tc.svc.Dashboard.Defaults != &defaults.Dashboard {
				t.Errorf("Dashboard defaults were not handed to the Lookup correctly\n want: %v\ngot:  %v",
					&defaults.Dashboard, tc.svc.Dashboard.Defaults)
			}
			// option.defaults
			if tc.svc.Options.Defaults != &defaults.Options {
				t.Errorf("Options defaults were not handed to the Lookup correctly\n want: %v\ngot:  %v",
					&defaults.Options, tc.svc.Options.Defaults)
			}
			// hardDefaults
			if tc.svc.HardDefaults != &hardDefaults {
				t.Errorf("HardDefaults were not handed to the Lookup correctly\n want: %v\ngot:  %v",
					&hardDefaults, tc.svc.HardDefaults)
			}
			// dashboard.hardDefaults
			if tc.svc.Dashboard.HardDefaults != &hardDefaults.Dashboard {
				t.Errorf("Dashboard hardDefaults were not handed to the Lookup correctly\n want: %v\ngot:  %v",
					&hardDefaults.Dashboard, tc.svc.Dashboard.HardDefaults)
			}
			// option.hardDefaults
			if tc.svc.Options.HardDefaults != &hardDefaults.Options {
				t.Errorf("Options hardDefaults were not handed to the Lookup correctly\n want: %v\ngot:  %v",
					&hardDefaults.Options, tc.svc.Options.HardDefaults)
			}
			// Notify
			if len(tc.svc.Notify) != 0 {
				for i := range tc.svc.Notify {
					if tc.svc.Notify[i].Main == nil {
						t.Error("Notify init didn't initialise the Main")
					}
				}
			}
			// Command
			if len(tc.svc.Command) != 0 {
				if tc.svc.CommandController == nil {
					t.Errorf("CommandController is still nil with %v Commands present",
						tc.svc.Command)
				}
			} else if tc.svc.CommandController != nil {
				t.Errorf("CommandController should be nil with %v Commands present",
					tc.svc.Command)
			}
			// WebHook
			if len(tc.svc.WebHook) != 0 {
				for i := range tc.svc.WebHook {
					if tc.svc.WebHook[i].Main == nil {
						t.Error("WebHook init didn't initialise the Main")
					}
				}
			}
		})
	}
}
