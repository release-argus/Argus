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

package latestver

import (
	"os"
	"testing"

	svcstatus "github.com/release-argus/Argus/service/status"
)

func TestLookup_GetAccessToken(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		env         map[string]string
		root        *string
		dfault      *string
		hardDefault *string
		wantString  string
	}{
		"root overrides all": {
			wantString:  "this",
			root:        stringPtr("this"),
			dfault:      stringPtr("not_this"),
			hardDefault: stringPtr("not_this")},
		"default overrides hardDefault": {
			wantString:  "this",
			dfault:      stringPtr("this"),
			hardDefault: stringPtr("not_this")},
		"hardDefault is last resort": {
			wantString:  "this",
			hardDefault: stringPtr("this")},
		"env var is used": {
			wantString: "this",
			env:        map[string]string{"TESTLOOKUP_LV_GETACCESSTOKEN_ONE": "this"},
			root:       stringPtr("${TESTLOOKUP_LV_GETACCESSTOKEN_ONE}"),
		},
		"env var partial is used": {
			wantString: "this",
			env:        map[string]string{"TESTLOOKUP_LV_GETACCESSTOKEN_TWO": "th"},
			root:       stringPtr("${TESTLOOKUP_LV_GETACCESSTOKEN_TWO}is"),
		},
		"empty env var is used": {
			wantString: "this",
			env:        map[string]string{"TESTLOOKUP_LV_GETACCESSTOKEN_THREE": ""},
			root:       stringPtr("th${TESTLOOKUP_LV_GETACCESSTOKEN_THREE}is"),
		},
		"undefined env var is used": {
			wantString: "${TESTLOOKUP_LV_GETACCESSTOKEN_UNSET}",
			root:       stringPtr("${TESTLOOKUP_LV_GETACCESSTOKEN_UNSET}"),
			dfault:     stringPtr("this"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}
			lookup := testLookup(false, false)
			lookup.AccessToken = tc.root
			lookup.Defaults.AccessToken = tc.dfault
			lookup.HardDefaults.AccessToken = tc.hardDefault

			// WHEN GetAccessToken is called
			got := lookup.GetAccessToken()

			// THEN the function returns the correct result
			if got == nil {
				t.Errorf("want: %q, got:  %v",
					tc.wantString, got)
			} else if *got != tc.wantString {
				t.Errorf("want: %q, got:  %q",
					tc.wantString, *got)
			}
		})
	}
}

func TestLookup_GetAllowInvalidCerts(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		root        *bool
		dfault      *bool
		hardDefault *bool
		wantBool    bool
	}{
		"root overrides all": {
			wantBool:    true,
			root:        boolPtr(true),
			dfault:      boolPtr(false),
			hardDefault: boolPtr(false)},
		"default overrides hardDefault": {
			wantBool:    true,
			dfault:      boolPtr(true),
			hardDefault: boolPtr(false)},
		"hardDefault is last resort": {
			wantBool:    true,
			hardDefault: boolPtr(true)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(false, false)
			lookup.AllowInvalidCerts = tc.root
			lookup.Defaults.AllowInvalidCerts = tc.dfault
			lookup.HardDefaults.AllowInvalidCerts = tc.hardDefault

			// WHEN GetAllowInvalidCerts is called
			got := lookup.GetAllowInvalidCerts()

			// THEN the function returns the correct result
			if got != tc.wantBool {
				t.Errorf("want: %t, got:  %t",
					tc.wantBool, got)
			}
		})
	}
}

func TestLookup_ServiceURL(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		serviceType   string
		url           string
		webURL        string
		ignoreWebURL  bool
		latestVersion string
		want          string
	}{
		"github - want repo url address": {
			want:         "https://github.com/release-argus/Argus",
			serviceType:  "github",
			url:          "release-argus/Argus",
			webURL:       "foo",
			ignoreWebURL: true,
		},
		"github - want web_url address": {
			want:         "foo",
			serviceType:  "github",
			url:          "release-argus/Argus",
			webURL:       "foo",
			ignoreWebURL: false,
		},
		"github - want web_url address with version templating": {
			want:          "foo/1.2.3",
			serviceType:   "github",
			url:           "release-argus/Argus",
			webURL:        "foo/{{ version }}",
			latestVersion: "1.2.3",
			ignoreWebURL:  false,
		},
		"github - want web_url address with version templating, but have no latest_version": {
			want:          "https://github.com/release-argus/Argus",
			serviceType:   "github",
			url:           "release-argus/Argus",
			webURL:        "foo/{{ version }}",
			latestVersion: "",
			ignoreWebURL:  false,
		},
		"url - want query url": {
			want:         "https://release-argus.io",
			serviceType:  "url",
			url:          "https://release-argus.io",
			webURL:       "foo",
			ignoreWebURL: true,
		},
		"url - want web_url address": {
			want:         "foo",
			serviceType:  "url",
			url:          "https://release-argus.io",
			webURL:       "foo",
			ignoreWebURL: false,
		},
		"url - want web_url address with version templating": {
			want:          "foo/1.2.3",
			serviceType:   "url",
			url:           "https://release-argus.io",
			webURL:        "foo/{{ version }}",
			latestVersion: "1.2.3", ignoreWebURL: false,
		},
		"url - want web_url address with version templating, but have no latest_version": {
			want:          "https://release-argus.io",
			serviceType:   "url",
			url:           "https://release-argus.io",
			webURL:        "foo/{{ version }}",
			latestVersion: "",
			ignoreWebURL:  false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			status := svcstatus.Status{}
			status.Init(
				0, 0, 0,
				&name,
				stringPtr("http://example.com"))
			status.SetLatestVersion(tc.latestVersion, false)
			status.WebURL = &tc.webURL
			lookup := Lookup{Type: tc.serviceType, URL: tc.url, Status: &status}

			// WHEN GetAllowInvalidCerts is called
			got := lookup.ServiceURL(tc.ignoreWebURL)

			// THEN the function returns the correct result
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestLookup_GetUsePreRelease(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		root        *bool
		dfault      *bool
		hardDefault *bool
		wantBool    bool
	}{
		"root overrides all": {
			wantBool:    true,
			root:        boolPtr(true),
			dfault:      boolPtr(false),
			hardDefault: boolPtr(false)},
		"default overrides hardDefault": {
			wantBool:    true,
			dfault:      boolPtr(true),
			hardDefault: boolPtr(false)},
		"hardDefault is last resort": {
			wantBool:    true,
			hardDefault: boolPtr(true)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(false, false)
			lookup.UsePreRelease = tc.root
			lookup.Defaults.UsePreRelease = tc.dfault
			lookup.HardDefaults.UsePreRelease = tc.hardDefault

			// WHEN GetUsePreRelease is called
			got := lookup.GetUsePreRelease()

			// THEN the function returns the correct result
			if got != tc.wantBool {
				t.Errorf("want: %t\ngot:  %t",
					tc.wantBool, got)
			}
		})
	}
}

func TestLookup_GetURL(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		env         map[string]string
		urlType     bool
		url         string
		tagFallback bool
		want        string
	}{
		"type=url": {
			urlType: true,
			url:     "https://release-argus.io",
			want:    "https://release-argus.io",
		},
		"type=github": {
			url:  "release-argus/Argus",
			want: "https://api.github.com/repos/release-argus/Argus/releases",
		},
		"type=github, tagFallback": {
			url:         "release-argus/Argus",
			tagFallback: true,
			want:        "https://api.github.com/repos/release-argus/Argus/tags",
		},
		"env var is used": {
			env:     map[string]string{"TESTLOOKUP_LV_GETURL_ONE": "https://release-argus.io"},
			urlType: true,
			url:     "${TESTLOOKUP_LV_GETURL_ONE}",
			want:    "https://release-argus.io",
		},
		"env var partial is used": {
			env:     map[string]string{"TESTLOOKUP_LV_GETURL_TWO": "release-argus"},
			urlType: true,
			url:     "https://${TESTLOOKUP_LV_GETURL_TWO}.io",
			want:    "https://release-argus.io",
		},
		"empty env var is used": {
			env:     map[string]string{"TESTLOOKUP_LV_GETURL_THREE": ""},
			urlType: true,
			url:     "https://${TESTLOOKUP_LV_GETURL_THREE}.io",
			want:    "https://.io",
		},
		"undefined env var is used": {
			urlType: true,
			url:     "${TESTLOOKUP_LV_GETURL_UNSET}",
			want:    "${TESTLOOKUP_LV_GETURL_UNSET}",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}
			lookup := testLookup(tc.urlType, false)
			lookup.URL = tc.url
			if !tc.urlType {
				lookup.GitHubData.tagFallback = tc.tagFallback
			}

			// WHEN GetURL is called
			got := lookup.GetURL()

			// THEN the function returns the correct result
			if got != tc.want {
				t.Errorf("\nwant: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}
