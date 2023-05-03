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
	"strings"
	"testing"

	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	apitype "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
)

func TestService_String(t *testing.T) {
	tests := map[string]struct {
		svc  *Service
		want string
	}{
		"nil": {
			svc:  nil,
			want: "",
		},
		"empty": {
			svc:  &Service{},
			want: "{}",
		},
		"all fields defined": {
			svc: &Service{
				Comment: "svc for blah",
				Options: opt.Options{
					Active: boolPtr(false)},
				LatestVersion: latestver.Lookup{
					URL: "release-argus/Argus"},
				DeployedVersionLookup: &deployedver.Lookup{
					URL: "https://valid.release-argus.io/plain"},
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "", nil, nil,
						"discord",
						&map[string]string{
							"token": "bar"},
						nil, nil, nil)},
				Command: command.Slice{
					{"ls", "-la"}},
				WebHook: webhook.Slice{
					"foo": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil,
						"github",
						"https://example.com",
						nil, nil, nil)},
				Dashboard: *NewDashboardOptions(
					boolPtr(true), "", "", "",
					nil, nil),
				Defaults: &ServiceDefaults{
					Options: *opt.NewDefaults(
						"", boolPtr(false))},
				HardDefaults: &ServiceDefaults{
					Options: *opt.NewDefaults(
						"", boolPtr(false))}},
			want: `
comment: svc for blah
options:
  active: false
latest_version:
  url: release-argus/Argus
deployed_version:
  url: https://valid.release-argus.io/plain
notify:
  foo:
    type: discord
    url_fields:
      token: bar
command:
  - - ls
    - -la
webhook:
  foo:
    type: github
    url: https://example.com
dashboard:
  auto_approve: true`,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			prefixes := []string{"", " ", "  ", "    ", "- "}
			for _, prefix := range prefixes {
				want := strings.TrimPrefix(tc.want, "\n")
				if want != "" {
					if want != "{}" {
						want = prefix + strings.ReplaceAll(want, "\n", "\n"+prefix)
					}
					want += "\n"
				}

				// WHEN the Service is stringified with String
				got := tc.svc.String(prefix)

				// THEN the result is as expected
				if got != want {
					t.Errorf("(prefix=%q) got:\n%q\nwant:\n%q",
						prefix, got, want)
				}
			}
		})
	}
}

func TestService_Summary(t *testing.T) {
	// GIVEN a Service
	tests := map[string]struct {
		svc                      *Service
		approvedVersion          string
		deployedVersion          string
		deployedVersionTimestamp string
		latestVersion            string
		latestVersionTimestamp   string
		lastQueried              string
		want                     apitype.ServiceSummary
	}{
		"empty": {
			svc: &Service{},
			want: apitype.ServiceSummary{
				Type:                     stringPtr(""),
				Icon:                     stringPtr(""),
				IconLinkTo:               stringPtr(""),
				HasDeployedVersionLookup: boolPtr(false),
				Command:                  intPtr(0),
				WebHook:                  intPtr(0),
				Status:                   &apitype.Status{}},
		},
		"only id": {
			svc: &Service{
				ID: "foo"},
			want: apitype.ServiceSummary{
				ID:                       "foo",
				Type:                     stringPtr(""),
				Icon:                     stringPtr(""),
				IconLinkTo:               stringPtr(""),
				HasDeployedVersionLookup: boolPtr(false),
				Command:                  intPtr(0),
				WebHook:                  intPtr(0),
				Status:                   &apitype.Status{}},
		},
		"only options.active": {
			svc: &Service{
				Options: opt.Options{
					Active: boolPtr(false)}},
			want: apitype.ServiceSummary{
				Active:                   boolPtr(false),
				Type:                     stringPtr(""),
				Icon:                     stringPtr(""),
				IconLinkTo:               stringPtr(""),
				HasDeployedVersionLookup: boolPtr(false),
				Command:                  intPtr(0),
				WebHook:                  intPtr(0),
				Status:                   &apitype.Status{}},
		},
		"only latest_version.type": {
			svc: &Service{
				LatestVersion: latestver.Lookup{
					Type: "github"}},
			want: apitype.ServiceSummary{
				Type:                     stringPtr("github"),
				Icon:                     stringPtr(""),
				IconLinkTo:               stringPtr(""),
				HasDeployedVersionLookup: boolPtr(false),
				Command:                  intPtr(0),
				WebHook:                  intPtr(0),
				Status:                   &apitype.Status{}},
		},
		"only dashboard.icon, and it's a url": {
			svc: &Service{
				Dashboard: DashboardOptions{
					Icon: "https://example.com/icon.png"}},
			want: apitype.ServiceSummary{
				Type:                     stringPtr(""),
				Icon:                     stringPtr("https://example.com/icon.png"),
				IconLinkTo:               stringPtr(""),
				HasDeployedVersionLookup: boolPtr(false),
				Command:                  intPtr(0),
				WebHook:                  intPtr(0),
				Status:                   &apitype.Status{}},
		},
		"only dashboard.icon, and it's not a url": {
			svc: &Service{
				Dashboard: DashboardOptions{
					Icon: "smile"}},
			want: apitype.ServiceSummary{
				Type:                     stringPtr(""),
				Icon:                     stringPtr(""),
				IconLinkTo:               stringPtr(""),
				HasDeployedVersionLookup: boolPtr(false),
				Command:                  intPtr(0),
				WebHook:                  intPtr(0),
				Status:                   &apitype.Status{}},
		},
		"only dashboard.icon, from notify": {
			svc: &Service{
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "", nil,
						&map[string]string{
							"icon": "https://example.com/notify.png"},
						"", nil,
						shoutrrr.NewDefaults(
							"", nil, nil, nil),
						shoutrrr.NewDefaults(
							"", nil, nil, nil),
						shoutrrr.NewDefaults(
							"", nil, nil, nil))}},
			want: apitype.ServiceSummary{
				Type:                     stringPtr(""),
				Icon:                     stringPtr("https://example.com/notify.png"),
				IconLinkTo:               stringPtr(""),
				HasDeployedVersionLookup: boolPtr(false),
				Command:                  intPtr(0),
				WebHook:                  intPtr(0),
				Status:                   &apitype.Status{}},
		},
		"only dashboard.icon, dashboard overrides notify": {
			svc: &Service{
				Dashboard: DashboardOptions{
					Icon: "https://example.com/icon.png"},
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "", nil,
						&map[string]string{
							"icon": "https://example.com/notify.png"},
						"", nil,
						shoutrrr.NewDefaults(
							"", nil, nil, nil),
						shoutrrr.NewDefaults(
							"", nil, nil, nil),
						shoutrrr.NewDefaults(
							"", nil, nil, nil))}},
			want: apitype.ServiceSummary{
				Type:                     stringPtr(""),
				Icon:                     stringPtr("https://example.com/icon.png"),
				IconLinkTo:               stringPtr(""),
				HasDeployedVersionLookup: boolPtr(false),
				Command:                  intPtr(0),
				WebHook:                  intPtr(0),
				Status:                   &apitype.Status{}},
		},
		"only dashboard.icon_link_to": {
			svc: &Service{
				Dashboard: DashboardOptions{
					IconLinkTo: "https://example.com"}},
			want: apitype.ServiceSummary{
				Type:                     stringPtr(""),
				Icon:                     stringPtr(""),
				IconLinkTo:               stringPtr("https://example.com"),
				HasDeployedVersionLookup: boolPtr(false),
				Command:                  intPtr(0),
				WebHook:                  intPtr(0),
				Status:                   &apitype.Status{}},
		},
		"only deployed_version": {
			svc: &Service{
				DeployedVersionLookup: &deployedver.Lookup{}},
			want: apitype.ServiceSummary{
				Type:                     stringPtr(""),
				Icon:                     stringPtr(""),
				IconLinkTo:               stringPtr(""),
				HasDeployedVersionLookup: boolPtr(true),
				Command:                  intPtr(0),
				WebHook:                  intPtr(0),
				Status:                   &apitype.Status{}},
		},
		"no commands": {
			svc: &Service{
				Command: command.Slice{}},
			want: apitype.ServiceSummary{
				Type:                     stringPtr(""),
				Icon:                     stringPtr(""),
				IconLinkTo:               stringPtr(""),
				HasDeployedVersionLookup: boolPtr(false),
				Command:                  intPtr(0),
				WebHook:                  intPtr(0),
				Status:                   &apitype.Status{}},
		},
		"3 commands": {
			svc: &Service{
				Command: command.Slice{
					{"ls", "-la"},
					{"true"},
					{"false"}}},
			want: apitype.ServiceSummary{
				Type:                     stringPtr(""),
				Icon:                     stringPtr(""),
				IconLinkTo:               stringPtr(""),
				HasDeployedVersionLookup: boolPtr(false),
				Command:                  intPtr(3),
				WebHook:                  intPtr(0),
				Status:                   &apitype.Status{}},
		},
		"0 webhooks": {
			svc: &Service{
				WebHook: webhook.Slice{}},
			want: apitype.ServiceSummary{
				Type:                     stringPtr(""),
				Icon:                     stringPtr(""),
				IconLinkTo:               stringPtr(""),
				HasDeployedVersionLookup: boolPtr(false),
				Command:                  intPtr(0),
				WebHook:                  intPtr(0),
				Status:                   &apitype.Status{}},
		},
		"3 webhooks": {
			svc: &Service{
				WebHook: webhook.Slice{
					"bish": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil,
						"github",
						"", nil, nil, nil),
					"bash": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil,
						"github",
						"", nil, nil, nil),
					"bosh": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil,
						"gitlab",
						"", nil, nil, nil)}},
			want: apitype.ServiceSummary{
				Type:                     stringPtr(""),
				Icon:                     stringPtr(""),
				IconLinkTo:               stringPtr(""),
				HasDeployedVersionLookup: boolPtr(false),
				Command:                  intPtr(0),
				WebHook:                  intPtr(3),
				Status:                   &apitype.Status{}},
		},
		"only status": {
			svc: &Service{
				Status: svcstatus.Status{}},
			approvedVersion:          "1",
			deployedVersion:          "2",
			deployedVersionTimestamp: "2-",
			latestVersion:            "3",
			latestVersionTimestamp:   "3-",
			lastQueried:              "4",
			want: apitype.ServiceSummary{
				Type:                     stringPtr(""),
				Icon:                     stringPtr(""),
				IconLinkTo:               stringPtr(""),
				HasDeployedVersionLookup: boolPtr(false),
				Command:                  intPtr(0),
				WebHook:                  intPtr(0),
				Status: &apitype.Status{
					ApprovedVersion:          "1",
					DeployedVersion:          "2",
					DeployedVersionTimestamp: "2-",
					LatestVersion:            "3",
					LatestVersionTimestamp:   "3-",
					LastQueried:              "4"}},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// status
			tc.svc.Status.Init(
				len(tc.svc.Notify), len(tc.svc.Command), len(tc.svc.WebHook),
				&tc.svc.ID,
				&tc.svc.Dashboard.WebURL)
			if tc.approvedVersion != "" {
				tc.svc.Status.SetApprovedVersion(tc.approvedVersion, false)
				tc.svc.Status.SetDeployedVersion(tc.deployedVersion, false)
				tc.svc.Status.SetDeployedVersionTimestamp(tc.deployedVersionTimestamp)
				tc.svc.Status.SetLatestVersion(tc.latestVersion, false)
				tc.svc.Status.SetLatestVersionTimestamp(tc.latestVersionTimestamp)
				tc.svc.Status.SetLastQueried(tc.lastQueried)
			}

			// WHEN the Service is converted to a ServiceSummary
			got := tc.svc.Summary()

			// THEN the result is as expected
			if got.String() != tc.want.String() {
				t.Errorf("got:\n%q\nwant:\n%q",
					got.String(), tc.want.String())
			}
		})
	}
}
