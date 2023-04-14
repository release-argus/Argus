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
	"strings"
	"testing"

	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
)

func TestService_Convert(t *testing.T) {
	// GIVEN a Service with the old var style
	svcType := "url"
	active := true
	interval := "10s"
	semanticVersioning := true
	url := "https://release-argus.io"
	allowInvalidCerts := true
	accessToken := "foo"
	usePreRelease := true
	urlCommands := filter.URLCommandSlice{{Type: "regex", Regex: stringPtr("foo")}}
	autoApprove := true
	icon := "https://github.com/release-argus/Argus/raw/master/web/ui/static/favicon.svg"
	iconLinkTo := "https://release-argus.io/demo"
	webURL := "https://release-argus.io/docs"
	svc := Service{
		Type:               svcType,
		Active:             &active,
		Interval:           &interval,
		SemanticVersioning: &semanticVersioning,
		URL:                &url,
		AllowInvalidCerts:  &allowInvalidCerts,
		AccessToken:        &accessToken,
		UsePreRelease:      &usePreRelease,
		URLCommands:        &urlCommands,
		AutoApprove:        &autoApprove,
		Icon:               &icon,
		IconLinkTo:         &iconLinkTo,
		WebURL:             &webURL,
	}
	saveChannel := make(chan bool, 5)
	svc.Status.SaveChannel = &saveChannel

	// WHEN Convert is called
	svc.Convert()

	// THEN the vars are converted correctly
	// Active -> Options.Active
	if svc.Options.Active != &active {
		t.Errorf("Type not pushed to option.Active correctly\nwant: %t\ngot:  %v",
			active, svc.Options.Active)
	}
	// Interval -> Options.Interval
	if svc.Options.Interval != interval {
		t.Errorf("Interval not pushed to option.Interval correctly\nwant: %q\ngot:  %v",
			interval, svc.Options.Interval)
	}
	// SemanticVersioning -> Options.SemanticVersioning
	if *svc.Options.SemanticVersioning != semanticVersioning {
		t.Errorf("SemanticVersioning not pushed to option.SemanticVersioning correctly\nwant: %t\ngot:  %v",
			semanticVersioning, svc.Options.SemanticVersioning)
	}
	// Type -> LatestVersion.Type
	if svc.LatestVersion.Type != svcType {
		t.Errorf("Type not pushed to LatestVersion.Type correctly\nwant: %q\ngot:  %q",
			svcType, svc.LatestVersion.Type)
	}
	// URL -> LatestVersion.URL
	if svc.LatestVersion.URL != url {
		t.Errorf("URL not pushed to LatestVersion.URL correctly\nwant: %q\ngot:  %q",
			url, svc.LatestVersion.URL)
	}
	// AllowInvalidCerts -> LatestVersion.AllowInvalidCerts
	if svc.LatestVersion.AllowInvalidCerts != &allowInvalidCerts {
		t.Errorf("AllowInvalidCerts not pushed to LatestVersion.AllowInvalidCerts correctly\nwant: %t\ngot:  %v",
			allowInvalidCerts, svc.LatestVersion.AllowInvalidCerts)
	}
	// AccessToken -> LatestVersion.AccessToken
	if svc.LatestVersion.AccessToken != &accessToken {
		t.Errorf("AccessToken not pushed to LatestVersion.AccessToken correctly\nwant: %q\ngot:  %q",
			accessToken, util.EvalNilPtr(svc.LatestVersion.AccessToken, "nil"))
	}
	// UsePreRelease -> LatestVersion.UsePreRelease
	if svc.LatestVersion.UsePreRelease != &usePreRelease {
		t.Errorf("UsePreRelease not pushed to LatestVersion.UsePreRelease correctly\nwant: %t\ngot:  %v",
			usePreRelease, svc.LatestVersion.UsePreRelease)
	}
	// URLCommands -> LatestVersion.URLCommands
	if len(svc.LatestVersion.URLCommands) != len(urlCommands) {
		t.Errorf("URLCommands not pushed to LatestVersion.URLCommands correctly\nwant: %v\ngot:  %v",
			urlCommands, svc.LatestVersion.URLCommands)
	}
	// AutoApprove -> Dashboard.AutoApprove
	if svc.Dashboard.AutoApprove != &autoApprove {
		t.Errorf("AutoApprove not pushed to Dashboard.AutoApprove correctly\nwant: %t\ngot:  %v",
			autoApprove, svc.Dashboard.AutoApprove)
	}
	// Icon -> Dashboard.Icon
	if svc.Dashboard.Icon != icon {
		t.Errorf("Icon not pushed to Dashboard.Icon correctly\nwant: %q\ngot:  %q",
			icon, svc.Dashboard.Icon)
	}
	// IconLinkTo -> Dashboard.IconLinkTo
	if svc.Dashboard.IconLinkTo != iconLinkTo {
		t.Errorf("IconLinkTo not pushed to Dashboard.IconLinkTo correctly\nwant: %q\ngot:  %q",
			iconLinkTo, svc.Dashboard.IconLinkTo)
	}
	// WebURL -> Dashboard.WebURL
	if svc.Dashboard.WebURL != webURL {
		t.Errorf("WebURL not pushed to Dashboard.WebURL correctly\nwant: %q\ngot:  %q",
			webURL, svc.Dashboard.WebURL)
	}
	// Should have sent a message to the save channel
	if len(saveChannel) != 1 {
		t.Fatalf("Service was converted so message should have been sent to the save channel")
	}
}

func TestService_String(t *testing.T) {
	tests := map[string]struct {
		svc  *Service
		want string
	}{
		"nil": {
			svc:  nil,
			want: "<nil>",
		},
		"empty": {
			svc:  &Service{},
			want: "{}\n",
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
					"foo": {
						Type:      "discord",
						URLFields: map[string]string{"token": "bar"}}},
				Command: command.Slice{
					{"ls", "-la"}},
				WebHook: webhook.Slice{
					"foo": {
						Type: "github",
						URL:  "https://example.com"}},
				Dashboard: DashboardOptions{
					AutoApprove: boolPtr(true)},
				Defaults: &Service{
					Options: opt.Options{
						SemanticVersioning: boolPtr(false)}},
				HardDefaults: &Service{
					Options: opt.Options{
						SemanticVersioning: boolPtr(false)}},
			},
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
    auto_approve: true
`,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the Service is stringified with String
			got := tc.svc.String()

			// THEN the result is as expected
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
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
					"foo": {
						Params: map[string]string{
							"icon": "https://example.com/notify.png"},
						Main: &shoutrrr.Shoutrrr{
							Params: map[string]string{}},
						Defaults: &shoutrrr.Shoutrrr{
							Params: map[string]string{}},
						HardDefaults: &shoutrrr.Shoutrrr{
							Params: map[string]string{}}}}},
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
					"foo": {
						Params: map[string]string{
							"icon": "https://example.com/notify.png"},
						Main: &shoutrrr.Shoutrrr{
							Params: map[string]string{}},
						Defaults: &shoutrrr.Shoutrrr{
							Params: map[string]string{}},
						HardDefaults: &shoutrrr.Shoutrrr{
							Params: map[string]string{}}}}},
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
					"bish": {Type: "github"},
					"bash": {Type: "github"},
					"bosh": {Type: "github"}}},
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
