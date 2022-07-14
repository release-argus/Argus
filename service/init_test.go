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
	"fmt"
	"testing"
	"time"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
)

func TestServiceGetAccessToken(t *testing.T) {
	// GIVEN a Service
	service := testServiceGitHub()

	// WHEN GetAccessToken is called on it
	got := service.GetAccessToken()

	// THEN AccessToken is returned
	want := *service.AccessToken
	if got != want {
		t.Errorf("Got %q, want %q",
			got, want)
	}
}

func TestServiceGetActive(t *testing.T) {
	// GIVEN a Service
	service := testServiceGitHub()

	// WHEN GetActive is called on it
	got := service.GetActive()

	// THEN Active is returned
	want := *service.HardDefaults.Active
	if got != want {
		t.Errorf("Got %t, want %t",
			got, want)
	}
}

func TestServiceGetAllowInvalidCerts(t *testing.T) {
	// GIVEN a Service
	service := testServiceGitHub()

	// WHEN GetAllowInvalidCerts is called on it
	got := service.GetAllowInvalidCerts()

	// THEN AllowInvalidCerts is returned
	want := *service.AllowInvalidCerts
	if got != want {
		t.Errorf("Got %t, want %t",
			got, want)
	}
}

func TestServiceGetAutoApprove(t *testing.T) {
	// GIVEN a Service
	service := testServiceGitHub()

	// WHEN GetAutoApprove is called on it
	got := service.GetAutoApprove()

	// THEN AutoApprove is returned
	want := *service.AutoApprove
	if got != want {
		t.Errorf("Got %t, want %t",
			got, want)
	}
}

func TestServiceGetIconURLWithNoIcon(t *testing.T) {
	// GIVEN a Service with nil Icon
	service := testServiceGitHub()
	service.Icon = ""

	// WHEN GetIconURL is called on it
	got := service.GetIconURL()

	// THEN an empty string is returned
	want := ""
	if got != want {
		t.Errorf("Got %q, want %q",
			got, want)
	}
}

func TestServiceGetIconURLWithEmojiIcon(t *testing.T) {
	// GIVEN a Service with nil Icon
	service := testServiceGitHub()
	service.Icon = "argus"

	// WHEN GetIconURL is called on it
	got := service.GetIconURL()

	// THEN an empty string is returned
	want := ""
	if got != want {
		t.Errorf("Got %q, want %q",
			got, want)
	}
}

func TestServiceGetIconURLWithWebIcon(t *testing.T) {
	// GIVEN a Service with nil Icon
	service := testServiceGitHub()
	service.Icon = "https://example.com/icon.png"

	// WHEN GetIconURL is called on it
	got := service.GetIconURL()

	// THEN an empty string is returned
	want := service.Icon
	if got != want {
		t.Errorf("Got %q, want %q",
			got, want)
	}
}

func TestServiceGetIconURLWithNotifyIcon(t *testing.T) {
	// GIVEN a Service with nil Icon
	service := testServiceGitHub()
	service.Icon = ""
	notify := shoutrrr.Shoutrrr{
		Params: map[string]string{
			"icon": "https://example.com/icon.png",
		},
		Main:         &shoutrrr.Shoutrrr{},
		Defaults:     &shoutrrr.Shoutrrr{},
		HardDefaults: &shoutrrr.Shoutrrr{},
	}
	service.Notify = &shoutrrr.Slice{
		"test": &notify,
	}

	// WHEN GetIconURL is called on it
	got := service.GetIconURL()

	// THEN an empty string is returned
	want := (*service.Notify)["test"].GetSelfParam("icon")
	if got != want {
		t.Errorf("Got %q, want %q",
			got, want)
	}
}

func TestServiceGetIgnoreMisses(t *testing.T) {
	// GIVEN a Service
	service := testServiceGitHub()

	// WHEN GetIgnoreMisses is called on it
	got := service.GetIgnoreMisses()

	// THEN IgnoreMisses is returned
	want := service.IgnoreMisses
	if got == nil || *got != *want {
		g := "nil"
		if got != nil {
			g = fmt.Sprint(*got)
		}
		t.Errorf("Got %s, want %t",
			g, *want)
	}
}

func TestServiceGetInterval(t *testing.T) {
	// GIVEN a Service
	service := testServiceGitHub()

	// WHEN GetInterval is called on it
	got := service.GetInterval()

	// THEN Interval is returned
	want := service.Interval
	if got != *want {
		t.Errorf("Got %s, want %s",
			got, *want)
	}
}

func TestServiceGetIntervalPointer(t *testing.T) {
	// GIVEN a Service
	service := testServiceGitHub()

	// WHEN GetIntervalPointer is called on it
	got := service.GetIntervalPointer()

	// THEN a pointer to service.Interval is returned
	want := service.Interval
	if got != want {
		t.Errorf("Got %v, want %v",
			got, *want)
	}
}

func TestServiceGetIntervalDuration(t *testing.T) {
	// GIVEN a Service
	service := testServiceGitHub()

	// WHEN GetIntervalDuration is called
	got := service.GetIntervalDuration()

	// THEN the function returns Interval as a time.Duration
	want, _ := time.ParseDuration(service.GetInterval())
	if got != want {
		t.Errorf("Want %s, got %s",
			want, got)
	}
}

func TestServiceGetUsePreRelease(t *testing.T) {
	// GIVEN a Service
	service := testServiceGitHub()

	// WHEN GetUsePreRelease is called on it
	got := service.GetUsePreRelease()

	// THEN UsePreRelease is returned
	want := *service.UsePreRelease
	if got != want {
		t.Errorf("Got %t, want %t",
			got, want)
	}
}

func TestServiceGetSemanticVersioning(t *testing.T) {
	// GIVEN a Service
	service := testServiceGitHub()

	// WHEN GetSemanticVersioning is called on it
	got := service.GetSemanticVersioning()

	// THEN SemanticVersioning is returned
	want := *service.SemanticVersioning
	if got != want {
		t.Errorf("Got %t, want %t",
			got, want)
	}
}

func TestServiceGetRegexContent(t *testing.T) {
	// GIVEN a Service
	service := testServiceGitHub()

	// WHEN GetRegexContent is called on it
	got := service.GetRegexContent()

	// THEN RegexContent is returned
	want := *service.RegexContent
	if got == nil || *got != want {
		g := "nil"
		if got != nil {
			g = *got
		}
		t.Errorf("Got %q, want %q",
			g, want)
	}
}

func TestServiceGetRegexVersion(t *testing.T) {
	// GIVEN a Service
	service := testServiceGitHub()

	// WHEN GetRegexVersion is called on it
	got := service.GetRegexVersion()

	// THEN RegexVersion is returned
	want := *service.RegexVersion
	if got == nil || *got != want {
		g := "nil"
		if got != nil {
			g = *got
		}
		t.Errorf("Got %q, want %q",
			g, want)
	}
}

func TestServiceGetServiceURLWithLatestVersionAndNoIgnoreWebURL(t *testing.T) {
	// GIVEN a Service that's got a LatestVersion
	service := testServiceGitHub()

	// WHEN GetServiceURL is called on it
	got := service.GetServiceURL(false)

	// THEN WebURL is returned
	want := service.GetWebURL()
	if got != want {
		t.Errorf("Got %s, want %s",
			got, want)
	}
}

func TestServiceGetServiceURLWithNoLatestVersionAndNoIgnoreWebURLAndVersionNotInWebURL(t *testing.T) {
	// GIVEN a Service that's not got a LatestVersion
	service := testServiceGitHub()
	service.Status.LatestVersion = ""

	// WHEN GetServiceURL is called on it
	got := service.GetServiceURL(false)

	// THEN WebURL is returned
	want := service.GetWebURL()
	if got != want {
		t.Errorf("Got %s, want %s",
			got, want)
	}
}

func TestServiceGetServiceURLWithNoLatestVersionAndNoIgnoreWebURLAndVersionInWebURL(t *testing.T) {
	// GIVEN a Service that's not got a LatestVersion
	service := testServiceGitHub()
	service.Status.LatestVersion = ""
	*service.WebURL = "https://example.com/{{ version }}"

	// WHEN GetServiceURL is called on it
	got := service.GetServiceURL(false)

	// THEN formatted GitHub URL is returned
	want := fmt.Sprintf("https://github.com/%s", *service.URL)
	if got != want {
		t.Errorf("Got %s, want %s",
			got, want)
	}
}

func TestServiceGetServiceURLWithGitHubOwnerRepo(t *testing.T) {
	// GIVEN a Service with type GitHub
	service := testServiceGitHub()

	// WHEN GetServiceURL is called on it
	got := service.GetServiceURL(true)

	// THEN the formatted GitHub URL is returned
	want := fmt.Sprintf("https://github.com/%s", *service.URL)
	if got != want {
		t.Errorf("Got %s, want %s",
			got, want)
	}
}

func TestServiceGetServiceURLWithGitHubFullPath(t *testing.T) {
	// GIVEN a Service with type GitHub
	service := testServiceGitHub()
	*service.URL = "https://github.com/foo/bar"

	// WHEN GetServiceURL is called on it
	got := service.GetServiceURL(true)

	// THEN URL is returned
	want := *service.URL
	if got != want {
		t.Errorf("Got %s, want %s",
			got, want)
	}
}

func TestServiceGetIconURLWithServiceIcon(t *testing.T) {
	// GIVEN a Service with an Icon
	service := testServiceGitHub()
	service.Icon = "argus"

	// WHEN GetIconURL is called on it
	got := service.GetIconURL()

	// THEN Icon is not returned as it's not a URL
	want := ""
	if got != want {
		t.Errorf("Got %q, want %q",
			got, want)
	}
}

func TestServiceGetIconURLWithNoServiceIconButWithNotifyIcon(t *testing.T) {
	// GIVEN a Service with an Icon
	service := testServiceGitHub()
	emptyShoutrrr := shoutrrr.Shoutrrr{
		Params:       map[string]string{},
		Main:         &shoutrrr.Shoutrrr{Params: map[string]string{}},
		Defaults:     &shoutrrr.Shoutrrr{Params: map[string]string{}},
		HardDefaults: &shoutrrr.Shoutrrr{Params: map[string]string{}},
	}
	service.Notify = &shoutrrr.Slice{
		"a": &emptyShoutrrr,
		"b": &emptyShoutrrr,
		"c": &emptyShoutrrr,
	}
	want := "https://release-argus.io"
	(*service.Notify)["b"].SetParam("icon", want)

	// WHEN GetIconURL is called on it
	got := service.GetIconURL()

	// THEN Icon is returned
	if got != want {
		t.Errorf("Got %q, want %q",
			got, want)
	}
}

func TestServiceGetIconURLWithNoServiceIconButWithNotifyIconEmoji(t *testing.T) {
	// GIVEN a Service with an emoji Icon
	service := testServiceGitHub()
	emptyShoutrrr := shoutrrr.Shoutrrr{
		Params:       map[string]string{},
		Main:         &shoutrrr.Shoutrrr{Params: map[string]string{}},
		Defaults:     &shoutrrr.Shoutrrr{Params: map[string]string{}},
		HardDefaults: &shoutrrr.Shoutrrr{Params: map[string]string{}},
	}
	service.Notify = &shoutrrr.Slice{
		"a": &emptyShoutrrr,
		"b": &emptyShoutrrr,
		"c": &emptyShoutrrr,
	}
	icon := "argus"
	(*service.Notify)["b"].SetParam("icon", icon)

	// WHEN GetIconURL is called on it
	got := service.GetIconURL()

	// THEN no Icon is returned as it's no a URL
	want := ""
	if got != want {
		t.Errorf("Got %q, want %q",
			got, want)
	}
}

func TestServiceGetURLWithTypeURL(t *testing.T) {
	// GIVEN a Service of type URL
	serviceType := "url"
	serviceURL := "https://release-argus.io"

	// WHEN GetURL is called on it
	got := GetURL(serviceURL, serviceType)

	// THEN the url will be returned unmodified
	if got != serviceURL {
		t.Errorf("Got %q, want %q",
			got, serviceURL)
	}
}

func TestServiceGetURLWithTypeGitHubAndShortPath(t *testing.T) {
	// GIVEN a Service of type GitHub and owner/repo as the URL
	serviceType := "github"
	serviceURL := "release-argus.io/Argus"

	// WHEN GetURL is called on it
	got := GetURL(serviceURL, serviceType)

	// THEN the GitHub API URL to serviceURL will be returned
	want := fmt.Sprintf("https://api.github.com/repos/%s/releases", serviceURL)
	if got != want {
		t.Errorf("Got %q, want %q",
			got, want)
	}
}

func TestServiceGetURLWithTypeGitHubAndFullPath(t *testing.T) {
	// GIVEN a Service of type GitHub and the full path as the URL
	serviceType := "github"
	serviceURL := "https://api.github.com/repos/foo/bar/releases"

	// WHEN GetURL is called on it
	got := GetURL(serviceURL, serviceType)

	// THEN the url will be returned unmodified
	if got != serviceURL {
		t.Errorf("Got %q, want %q",
			got, serviceURL)
	}
}

func TestServiceGetWebURLWithNoWebURL(t *testing.T) {
	// GIVEN a Service with no WebURL
	service := testServiceGitHub()
	service.WebURL = nil

	// WHEN GetWebURL is called on it
	got := service.GetWebURL()

	// THEN an empty string is returned
	want := ""
	if got != want {
		t.Errorf("Got %q, want %q",
			got, want)
	}
}

func TestServiceGetWebURLWithTemplatedWebURL(t *testing.T) {
	// GIVEN a Service with no WebURL
	service := testServiceGitHub()
	*service.WebURL = "foo{% if 'a' == 'a' %}{{ version }}{% endif %}"

	// WHEN GetWebURL is called on it
	got := service.GetWebURL()

	// THEN an empty string is returned
	want := "foo" + service.Status.LatestVersion
	if got != want {
		t.Errorf("Got %q, want %q",
			got, want)
	}
}

func TestServiceInitMetrics(t *testing.T) {
	// GIVEN a Service
	service := testServiceGitHub()

	// WHEN initMetrics is called on it
	service.initMetrics()

	// THEN the function doesn't hang
}

func TestServiceInitWithNoStatus(t *testing.T) {
	// GIVEN a Service with nil Status
	service := testServiceGitHub()
	service.Status = nil

	// WHEN Init is called on it
	service.Init(nil, &Service{}, &Service{})

	// THEN Status is no longer nil
	if service.Status == nil {
		t.Errorf("Status shouldn't be %v, it should have been initialised",
			service.Status)
	}
}

func TestServiceInitWithNoLatestVersion(t *testing.T) {
	// GIVEN a Service with no LatestVersion
	service := testServiceGitHub()
	service.Status.LatestVersion = ""

	// WHEN Init is called on it
	service.Init(nil, &Service{}, &Service{})

	// THEN LatestVersion is now DeployedVersion
	got := service.Status.LatestVersion
	want := service.Status.DeployedVersion
	if got != want {
		t.Errorf("LatestVersion should have been defaulted to %q, not %q",
			want, got)
	}
}

func TestServiceInitWithNoDeployedVersion(t *testing.T) {
	// GIVEN a Service with no DeployedVersion
	service := testServiceGitHub()
	service.Status.LatestVersion = ""

	// WHEN Init is called on it
	service.Init(nil, &Service{}, &Service{})

	// THEN DeployedVersion is now LatestVersion
	got := service.Status.DeployedVersion
	want := service.Status.LatestVersion
	if got != want {
		t.Errorf("DeployedVersion should have been defaulted to %q, not %q",
			want, got)
	}
}

func TestServiceInitWithLatestVersionApprovedAndDeployed(t *testing.T) {
	// GIVEN a Service with DeployedVersion == LatestVersion
	service := testServiceGitHub()
	service.Status.ApprovedVersion = service.Status.LatestVersion
	service.Status.DeployedVersion = service.Status.ApprovedVersion

	// WHEN Init is called on it
	service.Init(nil, &Service{}, &Service{})

	// THEN ApprovedVersion is reset
	got := service.Status.ApprovedVersion
	want := ""
	if got != want {
		t.Errorf("Latest==Deployed==Approved, so ApprovedVersion should have been reset to %q, not %q",
			want, got)
	}
}

func TestServiceInitHandsOutDefaults(t *testing.T) {
	// GIVEN a Service with nil Defaults
	service := testServiceGitHub()
	service.Defaults = nil
	id := "test"
	defaults := Service{
		ID: &id,
	}

	// WHEN Init is called on it
	service.Init(nil, &defaults, &Service{})

	// THEN ApprovedVersion is reset
	got := service.Defaults
	if got != &defaults {
		t.Errorf("Service should've been given %v Defaults, not %v",
			defaults, got)
	}
}

func TestServiceInitHandsOutHardDefaults(t *testing.T) {
	// GIVEN a Service with nil HardDefaults
	service := testServiceGitHub()
	service.HardDefaults = nil
	id := "test"
	defaults := Service{
		ID: &id,
	}

	// WHEN Init is called on it
	service.Init(nil, &Service{}, &defaults)

	// THEN ApprovedVersion is reset
	got := service.HardDefaults
	if got != &defaults {
		t.Errorf("Service should've been given %v HardDefaults, not %v",
			defaults, got)
	}
}

func TestServiceInitWithDeployedVersionLookupHandsOutDefaults(t *testing.T) {
	// GIVEN a Service with DeployedVersionLookup
	service := testServiceGitHub()
	service.DeployedVersionLookup = &DeployedVersionLookup{}
	defaults := DeployedVersionLookup{
		Regex: "test",
	}

	// WHEN Init is called on it
	service.Init(nil, &Service{DeployedVersionLookup: &defaults}, &Service{DeployedVersionLookup: &DeployedVersionLookup{}})

	// THEN ApprovedVersion is reset
	got := service.DeployedVersionLookup.Defaults
	if got != &defaults {
		t.Errorf("DeployedVersionLookup should've been given %v Defaults, not %v",
			defaults, got)
	}
}

func TestServiceInitWithDeployedVersionLookupHandsOutHardDefaults(t *testing.T) {
	// GIVEN a Service with DeployedVersionLookup
	service := testServiceGitHub()
	service.DeployedVersionLookup = &DeployedVersionLookup{}
	defaults := DeployedVersionLookup{
		Regex: "test",
	}

	// WHEN Init is called on it
	service.Init(nil, &Service{DeployedVersionLookup: &DeployedVersionLookup{}}, &Service{DeployedVersionLookup: &defaults})

	// THEN ApprovedVersion is reset
	got := service.DeployedVersionLookup.HardDefaults
	if got != &defaults {
		t.Errorf("DeployedVersionLookup should've been given %v Defaults, not %v",
			defaults, got)
	}
}
