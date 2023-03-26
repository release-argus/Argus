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
	serviceType := "url"
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
	service := Service{
		Type:               serviceType,
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
	service.Status.SaveChannel = &saveChannel

	// WHEN Convert is called
	service.Convert()

	// THEN the vars are converted correctly
	// Active -> Options.Active
	if service.Options.Active != &active {
		t.Errorf("Type not pushed to option.Active correctly\nwant: %t\ngot:  %v",
			active, service.Options.Active)
	}
	// Interval -> Options.Interval
	if service.Options.Interval != interval {
		t.Errorf("Interval not pushed to option.Interval correctly\nwant: %q\ngot:  %v",
			interval, service.Options.Interval)
	}
	// SemanticVersioning -> Options.SemanticVersioning
	if *service.Options.SemanticVersioning != semanticVersioning {
		t.Errorf("SemanticVersioning not pushed to option.SemanticVersioning correctly\nwant: %t\ngot:  %v",
			semanticVersioning, service.Options.SemanticVersioning)
	}
	// Type -> LatestVersion.Type
	if service.LatestVersion.Type != serviceType {
		t.Errorf("Type not pushed to LatestVersion.Type correctly\nwant: %q\ngot:  %q",
			serviceType, service.LatestVersion.Type)
	}
	// URL -> LatestVersion.URL
	if service.LatestVersion.URL != url {
		t.Errorf("URL not pushed to LatestVersion.URL correctly\nwant: %q\ngot:  %q",
			url, service.LatestVersion.URL)
	}
	// AllowInvalidCerts -> LatestVersion.AllowInvalidCerts
	if service.LatestVersion.AllowInvalidCerts != &allowInvalidCerts {
		t.Errorf("AllowInvalidCerts not pushed to LatestVersion.AllowInvalidCerts correctly\nwant: %t\ngot:  %v",
			allowInvalidCerts, service.LatestVersion.AllowInvalidCerts)
	}
	// AccessToken -> LatestVersion.AccessToken
	if service.LatestVersion.AccessToken != &accessToken {
		t.Errorf("AccessToken not pushed to LatestVersion.AccessToken correctly\nwant: %q\ngot:  %q",
			accessToken, util.EvalNilPtr(service.LatestVersion.AccessToken, "nil"))
	}
	// UsePreRelease -> LatestVersion.UsePreRelease
	if service.LatestVersion.UsePreRelease != &usePreRelease {
		t.Errorf("UsePreRelease not pushed to LatestVersion.UsePreRelease correctly\nwant: %t\ngot:  %v",
			usePreRelease, service.LatestVersion.UsePreRelease)
	}
	// URLCommands -> LatestVersion.URLCommands
	if len(service.LatestVersion.URLCommands) != len(urlCommands) {
		t.Errorf("URLCommands not pushed to LatestVersion.URLCommands correctly\nwant: %v\ngot:  %v",
			urlCommands, service.LatestVersion.URLCommands)
	}
	// AutoApprove -> Dashboard.AutoApprove
	if service.Dashboard.AutoApprove != &autoApprove {
		t.Errorf("AutoApprove not pushed to Dashboard.AutoApprove correctly\nwant: %t\ngot:  %v",
			autoApprove, service.Dashboard.AutoApprove)
	}
	// Icon -> Dashboard.Icon
	if service.Dashboard.Icon != icon {
		t.Errorf("Icon not pushed to Dashboard.Icon correctly\nwant: %q\ngot:  %q",
			icon, service.Dashboard.Icon)
	}
	// IconLinkTo -> Dashboard.IconLinkTo
	if service.Dashboard.IconLinkTo != iconLinkTo {
		t.Errorf("IconLinkTo not pushed to Dashboard.IconLinkTo correctly\nwant: %q\ngot:  %q",
			iconLinkTo, service.Dashboard.IconLinkTo)
	}
	// WebURL -> Dashboard.WebURL
	if service.Dashboard.WebURL != webURL {
		t.Errorf("WebURL not pushed to Dashboard.WebURL correctly\nwant: %q\ngot:  %q",
			webURL, service.Dashboard.WebURL)
	}
	// Should have sent a message to the save channel
	if len(saveChannel) != 1 {
		t.Fatalf("Service was converted so message should have been sent to the save channel")
	}
}

func TestService_String(t *testing.T) {
	tests := map[string]struct {
		service *Service
		want    string
	}{
		"nil": {
			service: nil,
			want:    "<nil>"},
		"empty": {
			service: &Service{},
			want:    "{}\n"},
		"all fields defined": {
			service: &Service{
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
				Status: svcstatus.Status{
					LatestVersion: "1.2.3"},
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
`},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the Service is stringified with String
			got := tc.service.String()

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
		service Service
		want    apitype.ServiceSummary
	}{
		"empty": {
			service: Service{},
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
			service: Service{
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
			service: Service{
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
			service: Service{
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
			service: Service{
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
			service: Service{
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
			service: Service{
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
			service: Service{
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
			service: Service{
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
			service: Service{
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
			service: Service{
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
			service: Service{
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
			service: Service{
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
			service: Service{
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
			service: Service{
				Status: svcstatus.Status{
					ApprovedVersion:          "1",
					DeployedVersion:          "2",
					DeployedVersionTimestamp: "2-",
					LatestVersion:            "3",
					LatestVersionTimestamp:   "3-",
					LastQueried:              "4"}},
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

			// WHEN the Service is converted to a ServiceSummary
			got := tc.service.Summary()

			// THEN the result is as expected
			if got.String() != tc.want.String() {
				t.Errorf("got:\n%q\nwant:\n%q",
					got.String(), tc.want.String())
			}
		})
	}

}
