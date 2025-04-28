// Copyright [2025] [Argus]
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

	"github.com/Masterminds/semver/v3"

	"github.com/release-argus/Argus/test"
)

var packageName = "latestver_github_types"

func TestRelease_String(t *testing.T) {
	tests := map[string]struct {
		release                  *Release
		release_semantic_version string
		want                     string
	}{
		"nil": {
			release: nil,
			want:    ""},
		"empty": {
			release: &Release{},
			want:    `{"prerelease": false}`},
		"only assets": {
			release: &Release{Assets: []Asset{
				{ID: 1, Name: "test", URL: "https://example.com", BrowserDownloadURL: "https://example.com/download"},
				{ID: 2, Name: "test2"}}},
			want: `
				{
					"prerelease": false,
					"assets": [
						{"url": "https://example.com", "id": 1, "name": "test", "browser_download_url": "https://example.com/download"},
						{"id": 2, "name": "test2"}
					]
				}`},
		"all fields defined": {
			release: &Release{
				URL:        "https://example.com",
				AssetsURL:  "https://example.com/assets",
				TagName:    "v1.2.3",
				PreRelease: true,
				Assets: []Asset{
					{ID: 1, Name: "test", URL: "https://example.com", BrowserDownloadURL: "https://example.com/download"}}},
			release_semantic_version: "1.2.3",
			want: `
				{
					"url": "https://example.com",
					"assets_url": "https://example.com/assets",
					"tag_name": "v1.2.3",
					"prerelease": true,
					"assets": [
						{"url": "https://example.com", "id": 1, "name": "test", "browser_download_url": "https://example.com/download"}
					]
				}`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.want = test.TrimJSON(tc.want)
			if tc.release_semantic_version != "" {
				tc.release.SemanticVersion, _ = semver.NewVersion(tc.release_semantic_version)
			}

			// WHEN the Release is stringified with String.
			got := tc.release.String()

			// THEN the result is as expected.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestAsset_String(t *testing.T) {
	tests := map[string]struct {
		asset *Asset
		want  string
	}{
		"nil": {
			asset: nil,
			want:  ""},
		"empty": {
			asset: &Asset{},
			want:  `{"id": 0}`},
		"all fields defined": {
			asset: &Asset{
				ID: 1, Name: "test", URL: "https://example.com", BrowserDownloadURL: "https://example.com/download"},
			want: `
				{
					"url": "https://example.com",
					"id": 1,
					"name": "test",
					"browser_download_url": "https://example.com/download"
				}`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.want = test.TrimJSON(tc.want)

			// WHEN the Asset is stringified with String.
			got := tc.asset.String()

			// THEN the result is as expected.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}
