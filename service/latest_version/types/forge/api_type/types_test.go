// Copyright [2026] [Argus]
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

package types

import (
	"testing"

	"github.com/Masterminds/semver/v3"

	"github.com/release-argus/Argus/internal/test"
)

var packageName = "latestver_forgetypes"

func TestRelease_String(t *testing.T) {
	// GIVEN: a Release.
	tests := []struct {
		name    string
		release *Release
		want    string
	}{
		{
			name:    "nil",
			release: nil,
			want:    "",
		},
		{
			name:    "empty",
			release: &Release{},
			want:    `{"prerelease": false}`,
		},
		{
			name: "only assets",
			release: &Release{
				Assets: []Asset{
					{
						ID:                 1,
						Name:               "test",
						URL:                "https://example.com",
						BrowserDownloadURL: "https://example.com/download",
					},
					{
						ID:   2,
						Name: "test2",
					},
				},
			},
			want: `
				{
					"prerelease": false,
					"assets": [
						{
							"url": "https://example.com",
							"id": 1,
							"name": "test",
							"browser_download_url": "https://example.com/download"
						},
						{
							"id": 2,
							"name": "test2"
						}
					]
				}`,
		},
		{
			name: "filled",
			release: &Release{
				URL:        "https://example.com",
				AssetsURL:  "https://example.com/assets",
				TagName:    "v1.2.3",
				PreRelease: true,
				Assets: []Asset{
					{
						ID:                 1,
						Name:               "test",
						URL:                "https://example.com",
						BrowserDownloadURL: "https://example.com/download",
					},
				},
				SemanticVersion: semver.MustParse("1.2.3"),
			},
			want: `
				{
					"url": "https://example.com",
					"assets_url": "https://example.com/assets",
					"tag_name": "v1.2.3",
					"prerelease": true,
					"assets": [
						{
							"url": "https://example.com",
							"id": 1,
							"name": "test",
							"browser_download_url": "https://example.com/download"
						}
					]
				}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: the Release is stringified with String.
			got := tc.release.String()

			// THEN: the result is as expected.
			tc.want = test.TrimJSON(tc.want)
			if got != tc.want {
				t.Errorf(
					"%s\nRelease.String() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestAsset_String(t *testing.T) {
	// GIVEN: an Asset.
	tests := []struct {
		name  string
		asset *Asset
		want  string
	}{
		{
			name:  "nil",
			asset: nil,
			want:  "",
		},
		{
			name:  "empty",
			asset: &Asset{},
			want:  `{"id": 0}`,
		},
		{
			name: "filled",
			asset: &Asset{
				ID:                 1,
				Name:               "test",
				URL:                "https://example.com",
				BrowserDownloadURL: "https://example.com/download",
			},
			want: `
				{
					"url": "https://example.com",
					"id": 1,
					"name": "test",
					"browser_download_url": "https://example.com/download"
				}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: the Asset is stringified with String.
			got := tc.asset.String()

			// THEN: the result is as expected.
			tc.want = test.TrimJSON(tc.want)
			if got != tc.want {
				t.Errorf(
					"%s\nAsset.String() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}
