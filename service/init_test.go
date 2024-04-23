// Copyright [2023] [Argus]
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
	test_webhook "github.com/release-argus/Argus/webhook/test"
)

func TestService_ServiceInfo(t *testing.T) {
	// GIVEN a Service
	svc := testService("TestServiceInfo", "url")
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

	// When ServiceInfo is called on it
	got := svc.ServiceInfo()
	want := util.ServiceInfo{
		ID:            id,
		URL:           url,
		WebURL:        webURL,
		LatestVersion: latestVersion,
	}

	// THEN we get the correct ServiceInfo
	if *got != want {
		t.Errorf("ServiceInfo didn't get the correct data\nwant: %#v\ngot:  %#v",
			want, got)
	}
}

func TestService_IconURL(t *testing.T) {
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
			notify: shoutrrr.Slice{"test": {
				Main:         &shoutrrr.ShoutrrrDefaults{},
				Defaults:     &shoutrrr.ShoutrrrDefaults{},
				HardDefaults: &shoutrrr.ShoutrrrDefaults{},
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
			notify: shoutrrr.Slice{"test": shoutrrr.New(
				nil, "", nil,
				&map[string]string{
					"icon": "https://example.com/icon.png"},
				"", nil,
				&shoutrrr.ShoutrrrDefaults{},
				&shoutrrr.ShoutrrrDefaults{},
				&shoutrrr.ShoutrrrDefaults{})},
		},
		"notify icon takes precedence over emoji": {
			want:          "https://example.com/icon.png",
			dashboardIcon: ":smile:",
			notify: shoutrrr.Slice{"test": shoutrrr.New(
				nil, "", nil,
				&map[string]string{
					"icon": "https://example.com/icon.png"},
				"", nil,
				&shoutrrr.ShoutrrrDefaults{},
				&shoutrrr.ShoutrrrDefaults{},
				&shoutrrr.ShoutrrrDefaults{})},
		},
		"dashboard icon takes precedence over notify icon": {
			want:          "https://root.com/icon.png",
			dashboardIcon: "https://root.com/icon.png",
			notify: shoutrrr.Slice{"test": shoutrrr.New(
				nil, "", nil,
				&map[string]string{
					"icon": "https://example.com/icon.png"},
				"", nil,
				&shoutrrr.ShoutrrrDefaults{},
				&shoutrrr.ShoutrrrDefaults{},
				&shoutrrr.ShoutrrrDefaults{})},
		},
	}

	for name, tc := range tests {
		svc := testService(name, "github")

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc.Dashboard.Icon = tc.dashboardIcon
			svc.Notify = tc.notify

			// WHEN IconURL is called
			got := svc.IconURL()

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
	tests := map[string]struct {
		svc      *Service
		defaults *Defaults
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
					"test": shoutrrr.New(
						nil, "", nil, nil,
						"discord",
						nil, nil, nil, nil)},
				Command: command.Slice{
					{"ls"}},
				WebHook: webhook.Slice{
					"test": test_webhook.WebHook(false, false, false)}},
		},
		"service with notifys from defaults": {
			svc: &Service{
				ID: "Init",
				LatestVersion: latestver.Lookup{
					Type: "github", URL: "release-argus/Argus"}},
			defaults: &Defaults{
				Notify: map[string]struct{}{
					"foo": {}}},
		},
		"service with notifys not from defaults": {
			svc: &Service{
				ID: "Init",
				LatestVersion: latestver.Lookup{
					Type: "github", URL: "release-argus/Argus"},
				Notify: shoutrrr.Slice{
					"test": &shoutrrr.Shoutrrr{}}},
			defaults: &Defaults{
				Notify: map[string]struct{}{
					"foo": {}}},
		},
		"service with commands from defaults": {
			svc: &Service{
				ID: "Init",
				LatestVersion: latestver.Lookup{
					Type: "github", URL: "release-argus/Argus"}},
			defaults: &Defaults{
				Command: command.Slice{
					{"ls"}}},
		},
		"service with commands not from defaults": {
			svc: &Service{
				ID: "Init",
				LatestVersion: latestver.Lookup{
					Type: "github", URL: "release-argus/Argus"},
				Command: command.Slice{
					{"test"}}},
			defaults: &Defaults{
				Command: command.Slice{
					{"ls"}}},
		},
		"service with webhooks from defaults": {
			svc: &Service{
				ID: "Init",
				LatestVersion: latestver.Lookup{
					Type: "github", URL: "release-argus/Argus"}},
			defaults: &Defaults{
				WebHook: map[string]struct{}{
					"bar": {}}},
		},
		"service with webhooks not from defaults": {
			svc: &Service{
				ID: "Init",
				LatestVersion: latestver.Lookup{
					Type: "github", URL: "release-argus/Argus"},
				WebHook: webhook.Slice{
					"test": test_webhook.WebHook(false, false, false)}},
			defaults: &Defaults{
				WebHook: map[string]struct{}{
					"bar": {}}},
		},
		"service with webhooks/commands from defaults and notify overriden": {
			svc: &Service{
				ID: "Init",
				LatestVersion: latestver.Lookup{
					Type: "github", URL: "release-argus/Argus"},
				Notify: shoutrrr.Slice{
					"test": &shoutrrr.Shoutrrr{}}},
			defaults: &Defaults{
				Notify: map[string]struct{}{
					"foo": {}},
				Command: command.Slice{
					{"ls"}},
				WebHook: map[string]struct{}{
					"bar": {}}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			if tc.defaults == nil {
				tc.defaults = &Defaults{}
			}
			var hardDefaults Defaults
			tc.svc.ID = name
			hadNotify := util.SortedKeys(tc.svc.Notify)
			hadWebHook := util.SortedKeys(tc.svc.WebHook)
			hadCommand := make(command.Slice, len(tc.svc.Command))
			copy(hadCommand, tc.svc.Command)

			// WHEN Init is called on it
			tc.svc.Init(
				tc.defaults, &hardDefaults,
				&shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{},
				&webhook.SliceDefaults{}, &webhook.WebHookDefaults{}, &webhook.WebHookDefaults{})

			// THEN pointers to those vars are handed out to the Lookup
			// defaults
			if tc.svc.Defaults != tc.defaults {
				t.Errorf("Defaults were not handed to the Lookup correctly\n want: %v\ngot:  %v",
					tc.defaults, tc.svc.Defaults)
			}
			// dashboard.defaults
			if tc.svc.Dashboard.Defaults != &tc.defaults.Dashboard {
				t.Errorf("Dashboard defaults were not handed to the Lookup correctly\n want: %v\ngot:  %v",
					&tc.defaults.Dashboard, tc.svc.Dashboard.Defaults)
			}
			// option.defaults
			if tc.svc.Options.Defaults != &tc.defaults.Options {
				t.Errorf("Options defaults were not handed to the Lookup correctly\n want: %v\ngot:  %v",
					&tc.defaults.Options, tc.svc.Options.Defaults)
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
			// Notifys shouldn't be overriden if non-empty
			if len(hadNotify) != 0 && len(tc.svc.Notify) != len(hadNotify) {
				t.Fatalf("Notify length changed\n want: %d (%v)\ngot:  %d (%v)",
					len(hadNotify), hadNotify, len(tc.svc.Notify), util.SortedKeys(tc.svc.Notify))
			}
			wantNotify := hadNotify
			if len(hadNotify) == 0 && tc.defaults != nil {
				wantNotify = make([]string, len(tc.defaults.Notify))
				wantNotify = util.SortedKeys(tc.defaults.Notify)
			}
			for _, i := range wantNotify {
				if tc.svc.Notify[i] == nil {
					t.Errorf("Notify[%s] was nil", i)
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
			// Command shouldn't be overriden if non-empty
			if len(hadCommand) != 0 && len(tc.svc.Command) != len(hadCommand) {
				t.Fatalf("Command length changed\n want: %d (%v)\ngot: %d (%v)",
					len(hadCommand), hadCommand, len(tc.svc.Command), tc.svc.Command)
			}
			wantCommand := hadCommand
			if len(hadCommand) == 0 && tc.defaults != nil {
				wantCommand = make(command.Slice, len(tc.defaults.Command))
				wantCommand = tc.defaults.Command
			}
			for i := range wantCommand {
				if tc.svc.Command[i].String() != wantCommand[i].String() {
					t.Errorf("Command[%d] changed\n want: %q\ngot:  %q",
						i, wantCommand[i].String(), tc.svc.Command[i].String())
				}
			}
			// WebHook
			if len(tc.svc.WebHook) != 0 {
				for i := range tc.svc.WebHook {
					if tc.svc.WebHook[i].Main == nil {
						t.Error("WebHook init didn't initialise the Main")
					}
				}
			}
			// WebHook shouldn't be overriden if non-empty
			if len(hadWebHook) != 0 && len(tc.svc.WebHook) != len(hadWebHook) {
				t.Fatalf("WebHook length changed\n want: %d (%v)\ngot: %d (%v)",
					len(hadWebHook), hadWebHook, len(tc.svc.WebHook), util.SortedKeys(tc.svc.WebHook))
			}
			wantWebHook := hadWebHook
			if len(hadWebHook) == 0 && tc.defaults != nil {
				wantWebHook = make([]string, len(tc.defaults.WebHook))
				wantWebHook = util.SortedKeys(tc.defaults.WebHook)
			}
			for _, i := range wantWebHook {
				if tc.svc.WebHook[i] == nil {
					t.Errorf("hadWebHook[%s] was nil", i)
				}
			}
		})
	}
}
