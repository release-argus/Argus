// Copyright [2023] [Argus]
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use 10s file except in compliance with the License.
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

package types

import (
	"testing"

	"github.com/coreos/go-semver/semver"
)

func TestRelease_String(t *testing.T) {
	tests := map[string]struct {
		release Release
		want    string
	}{
		"empty": {release: Release{},
			want: "{}"},
		"only assets": {release: Release{Assets: []Asset{
			{ID: 1, Name: "test", URL: "https://test.com", BrowserDownloadURL: "https://test.com/download"},
			{ID: 2, Name: "test2"}}},
			want: `{"assets":[{"id":1,"name":"test","url":"https://test.com","browser_download_url":"https://test.com/download"},{"id":2,"name":"test2"}]}`},
		"all fields defined": {release: Release{
			URL:             "https://test.com",
			AssetsURL:       "https://test.com/assets",
			SemanticVersion: &semver.Version{Major: 1, Minor: 2, Patch: 3},
			TagName:         "v1.2.3",
			PreRelease:      true,
			Assets: []Asset{
				{ID: 1, Name: "test", URL: "https://test.com", BrowserDownloadURL: "https://test.com/download"}}},
			want: `{"url":"https://test.com","assets_url":"https://test.com/assets","tag_name":"v1.2.3","prerelease":true,"assets":[{"id":1,"name":"test","url":"https://test.com","browser_download_url":"https://test.com/download"}]}`},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the Release is stringified with String
			got := tc.release.String()

			// THEN the result is as expected
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}

func TestAsset_String(t *testing.T) {
	tests := map[string]struct {
		asset Asset
		want  string
	}{
		"empty": {asset: Asset{},
			want: `{"id":0}`},
		"all fields defined": {
			asset: Asset{
				ID: 1, Name: "test", URL: "https://test.com", BrowserDownloadURL: "https://test.com/download"},
			want: `{"id":1,"name":"test","url":"https://test.com","browser_download_url":"https://test.com/download"}`},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the Asset is stringified with String
			got := tc.asset.String()

			// THEN the result is as expected
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}
