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

	"github.com/release-argus/Argus/service/latest_version/filters"
	"github.com/release-argus/Argus/utils"
)

func TestConvert(t *testing.T) {
	// GIVEN a Service with the old var style
	serviceType := "url"
	active := true
	interval := "10s"
	semanticVersioning := true
	url := "https://release-argus.io"
	allowInvalidCerts := true
	accessToken := "foo"
	usePreRelease := true
	urlCommands := filters.URLCommandSlice{{Type: "regex", Regex: stringPtr("foo")}}
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
	if service.Options.Active != &active {
		t.Errorf("Type not pushed to Options.Active correctly\nwant: %t\ngot:  %v",
			active, service.Options.Active)
	}
	if service.Options.Interval != interval {
		t.Errorf("Interval not pushed to Options.Interval correctly\nwant: %q\ngot:  %v",
			interval, service.Options.Interval)
	}
	if *service.Options.SemanticVersioning != semanticVersioning {
		t.Errorf("SemanticVersioning not pushed to Options.SemanticVersioning correctly\nwant: %t\ngot:  %v",
			semanticVersioning, service.Options.SemanticVersioning)
	}
	if service.LatestVersion.Type != serviceType {
		t.Errorf("Type not pushed to LatestVersion.Type correctly\nwant: %q\ngot:  %q",
			serviceType, service.LatestVersion.Type)
	}
	if service.LatestVersion.URL != url {
		t.Errorf("URL not pushed to LatestVersion.URL correctly\nwant: %q\ngot:  %q",
			url, service.LatestVersion.URL)
	}
	if service.LatestVersion.AllowInvalidCerts != &allowInvalidCerts {
		t.Errorf("AllowInvalidCerts not pushed to LatestVersion.AllowInvalidCerts correctly\nwant: %t\ngot:  %v",
			allowInvalidCerts, service.LatestVersion.AllowInvalidCerts)
	}
	if service.LatestVersion.AccessToken != &accessToken {
		t.Errorf("AccessToken not pushed to LatestVersion.AccessToken correctly\nwant: %q\ngot:  %q",
			accessToken, utils.EvalNilPtr(service.LatestVersion.AccessToken, "nil"))
	}
	if service.LatestVersion.UsePreRelease != &usePreRelease {
		t.Errorf("UsePreRelease not pushed to LatestVersion.UsePreRelease correctly\nwant: %t\ngot:  %v",
			usePreRelease, service.LatestVersion.UsePreRelease)
	}
	if len(service.LatestVersion.URLCommands) != len(urlCommands) {
		t.Errorf("URLCommands not pushed to LatestVersion.URLCommands correctly\nwant: %v\ngot:  %v",
			urlCommands, service.LatestVersion.URLCommands)
	}
	if service.Dashboard.AutoApprove != &autoApprove {
		t.Errorf("AutoApprove not pushed to Dashboard.AutoApprove correctly\nwant: %t\ngot:  %v",
			autoApprove, service.Dashboard.AutoApprove)
	}
	if service.Dashboard.Icon != icon {
		t.Errorf("Icon not pushed to Dashboard.Icon correctly\nwant: %q\ngot:  %q",
			icon, service.Dashboard.Icon)
	}
	if service.Dashboard.IconLinkTo != iconLinkTo {
		t.Errorf("IconLinkTo not pushed to Dashboard.IconLinkTo correctly\nwant: %q\ngot:  %q",
			iconLinkTo, service.Dashboard.IconLinkTo)
	}
	if service.Dashboard.WebURL != webURL {
		t.Errorf("WebURL not pushed to Dashboard.WebURL correctly\nwant: %q\ngot:  %q",
			webURL, service.Dashboard.WebURL)
	}
	if len(saveChannel) != 1 {
		t.Fatalf("Service was converted so message should have been sent to the save channel")
	}
}
