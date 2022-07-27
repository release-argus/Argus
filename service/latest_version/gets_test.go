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

package latest_version

import (
	"testing"

	service_status "github.com/release-argus/Argus/service/status"
)

func TestGetAccessToken(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		accessTokenRoot        *string
		accessTokenDefault     *string
		accessTokenHardDefault *string
		wantString             string
	}{
		"root overrides all": {wantString: "this", accessTokenRoot: stringPtr("this"),
			accessTokenDefault: stringPtr("not_this"), accessTokenHardDefault: stringPtr("not_this")},
		"default overrides hardDefault": {wantString: "this", accessTokenRoot: nil,
			accessTokenDefault: stringPtr("this"), accessTokenHardDefault: stringPtr("not_this")},
		"hardDefault is last resort": {wantString: "this", accessTokenRoot: nil, accessTokenDefault: nil,
			accessTokenHardDefault: stringPtr("this")},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			lookup := testLookupGitHub()
			lookup.AccessToken = tc.accessTokenRoot
			(*lookup.Defaults).AccessToken = tc.accessTokenDefault
			(*lookup.HardDefaults).AccessToken = tc.accessTokenHardDefault

			// WHEN GetAccessToken is called
			got := lookup.GetAccessToken()

			// THEN the function returns the correct result
			if got == nil {
				t.Errorf("%s:\nwant: %q\ngot:  %v",
					name, tc.wantString, got)
			} else if *got != tc.wantString {
				t.Errorf("%s:\nwant: %q\ngot:  %q",
					name, tc.wantString, *got)
			}
		})
	}
}

func TestGetAllowInvalidCerts(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		allowInvalidCertsRoot        *bool
		allowInvalidCertsDefault     *bool
		allowInvalidCertsHardDefault *bool
		wantBool                     bool
	}{
		"root overrides all": {wantBool: true, allowInvalidCertsRoot: boolPtr(true),
			allowInvalidCertsDefault: boolPtr(false), allowInvalidCertsHardDefault: boolPtr(false)},
		"default overrides hardDefault": {wantBool: true, allowInvalidCertsRoot: nil,
			allowInvalidCertsDefault: boolPtr(true), allowInvalidCertsHardDefault: boolPtr(false)},
		"hardDefault is last resort": {wantBool: true, allowInvalidCertsRoot: nil, allowInvalidCertsDefault: nil,
			allowInvalidCertsHardDefault: boolPtr(true)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			lookup := testLookupGitHub()
			lookup.AllowInvalidCerts = tc.allowInvalidCertsRoot
			(*lookup.Defaults).AllowInvalidCerts = tc.allowInvalidCertsDefault
			(*lookup.HardDefaults).AllowInvalidCerts = tc.allowInvalidCertsHardDefault

			// WHEN GetAllowInvalidCerts is called
			got := lookup.GetAllowInvalidCerts()

			// THEN the function returns the correct result
			if got != tc.wantBool {
				t.Errorf("%s:\nwant: %t\ngot:  %t",
					name, tc.wantBool, got)
			}
		})
	}
}

func TestGetServiceURL(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		serviceType   string
		url           string
		webURL        string
		ignoreWebURL  bool
		latestVersion string
		want          string
	}{
		"github - want repo url address": {want: "https://github.com/release-argus/Argus",
			serviceType: "github", url: "release-argus/Argus", webURL: "foo", ignoreWebURL: true},
		"github - want web_url address": {want: "foo",
			serviceType: "github", url: "release-argus/Argus", webURL: "foo", ignoreWebURL: false},
		"github - want web_url address with version templating": {want: "foo/1.2.3",
			serviceType: "github", url: "release-argus/Argus", webURL: "foo/{{ version }}", latestVersion: "1.2.3", ignoreWebURL: false},
		"github - want web_url address with version templating, but have no latest_version": {want: "https://github.com/release-argus/Argus",
			serviceType: "github", url: "release-argus/Argus", webURL: "foo/{{ version }}", latestVersion: "", ignoreWebURL: false},
		"url - want query url": {want: "https://release-argus.io",
			serviceType: "url", url: "https://release-argus.io", webURL: "foo", ignoreWebURL: true},
		"url - want web_url address": {want: "foo",
			serviceType: "url", url: "https://release-argus.io", webURL: "foo", ignoreWebURL: false},
		"url - want web_url address with version templating": {want: "foo/1.2.3",
			serviceType: "url", url: "https://release-argus.io", webURL: "foo/{{ version }}", latestVersion: "1.2.3", ignoreWebURL: false},
		"url - want web_url address with version templating, but have no latest_version": {want: "https://release-argus.io",
			serviceType: "url", url: "https://release-argus.io", webURL: "foo/{{ version }}", latestVersion: "", ignoreWebURL: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			webURL := &tc.webURL
			status := service_status.Status{LatestVersion: tc.latestVersion, WebURL: &webURL}
			lookup := Lookup{Type: tc.serviceType, URL: tc.url, status: &status}

			// WHEN GetAllowInvalidCerts is called
			got := lookup.GetServiceURL(tc.ignoreWebURL)

			// THEN the function returns the correct result
			if got != tc.want {
				t.Errorf("%s:\nwant: %q\ngot:  %q",
					name, tc.want, got)
			}
		})
	}
}

func TestGetUsePreRelease(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		usePreReleaseRoot        *bool
		usePreReleaseDefault     *bool
		usePreReleaseHardDefault *bool
		wantBool                 bool
	}{
		"root overrides all": {wantBool: true, usePreReleaseRoot: boolPtr(true),
			usePreReleaseDefault: boolPtr(false), usePreReleaseHardDefault: boolPtr(false)},
		"default overrides hardDefault": {wantBool: true, usePreReleaseRoot: nil,
			usePreReleaseDefault: boolPtr(true), usePreReleaseHardDefault: boolPtr(false)},
		"hardDefault is last resort": {wantBool: true, usePreReleaseRoot: nil, usePreReleaseDefault: nil,
			usePreReleaseHardDefault: boolPtr(true)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			lookup := testLookupGitHub()
			lookup.UsePreRelease = tc.usePreReleaseRoot
			(*lookup.Defaults).UsePreRelease = tc.usePreReleaseDefault
			(*lookup.HardDefaults).UsePreRelease = tc.usePreReleaseHardDefault

			// WHEN GetUsePreRelease is called
			got := lookup.GetUsePreRelease()

			// THEN the function returns the correct result
			if got != tc.wantBool {
				t.Errorf("%s:\nwant: %t\ngot:  %t",
					name, tc.wantBool, got)
			}
		})
	}
}
