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

// Package github provides a github-based lookup type.
package github

import (
	"strconv"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	forgetypes "github.com/release-argus/Argus/service/latest_version/types/forge/api_type"
)

func TestData_String(t *testing.T) {
	tests := []struct {
		name       string
		githubData *Data
		want       string
	}{
		{
			name:       "nil",
			githubData: nil,
			want:       "",
		},
		{
			name:       "empty",
			githubData: &Data{},
			want:       "{}",
		},
		{
			name: "filled",
			githubData: &Data{
				etag: "argus",
				releases: []forgetypes.Release{
					{URL: "https://example.com/1.2.3"},
					{URL: "https://example.com/3.2.1", PreRelease: true},
				},
			},
			want: `
				{
					"etag": "argus",
					"releases": [
						{"url": "https://example.com/1.2.3", "prerelease": false},
						{"url": "https://example.com/3.2.1", "prerelease": true}
					]
				}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.want = test.TrimJSON(tc.want)

			got := tc.githubData.String()

			if got != tc.want {
				t.Errorf("Data.String() mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestData_TagFallback(t *testing.T) {
	gd := &Data{}
	tests := []bool{
		true,
		false,
		true,
		false,
		true,
	}

	for i, tc := range tests {
		gd.SetTagFallback()

		got := gd.TagFallback()

		if got != tc {
			t.Errorf("toggle %d: Data.TagFallback() mismatch\ngot:  %t\nwant: %t", i, got, tc)
		}
	}
}

func TestData_ETag(t *testing.T) {
	testData := &Data{}

	got := testData.ETag()

	want := ""
	if got != want {
		t.Errorf("fresh: Data.ETag() mismatch\ngot:  %q\nwant: %q", got, want)
	}

	newETag := "foo"
	testData.SetETag(newETag)

	got = testData.ETag()
	want = newETag
	if got != want {
		t.Errorf("after SetETag: Data.ETag() mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestData_PerPage(t *testing.T) {
	for _, foundOnPage := range []int{1, 2, 10} {
		t.Run(strconv.Itoa(foundOnPage), func(t *testing.T) {
			t.Parallel()

			data := &Data{}

			data.SetPerPage(foundOnPage)

			want := foundOnPage * defaultPerPage
			if got := data.PerPage(); got != want {
				t.Errorf("after SetPerPage: Data.PerPage() mismatch\ngot:  %d\nwant: %d", got, want)
			}

			data.ResetPerPage()

			if got, want := data.PerPage(), 0; got != want {
				t.Errorf(
					"after ResetPerPage: Data.PerPage() mismatch\ngot:  %d\nwant: %d",
					got, want,
				)
			}
		})
	}
}

func TestData_Releases(t *testing.T) {
	testData := &Data{}

	got := testData.Releases()

	if testErr := test.AssertSlicesEqualFunc(
		t,
		got,
		[]forgetypes.Release(nil),
		releaseEqual,
		"Data.Releases()",
	); testErr != nil {
		t.Fatalf("fresh: %v", testErr)
	}

	newReleases := []forgetypes.Release{
		{TagName: "foo"},
		{TagName: "bar"},
	}
	testData.SetReleases(newReleases)

	got = testData.Releases()
	if testErr := test.AssertSlicesEqualFunc(
		t,
		got,
		newReleases,
		releaseEqual,
		"Data.Releases()",
	); testErr != nil {
		t.Fatalf("after SetReleases: %v", testErr)
	}
}

func TestData_HasReleases(t *testing.T) {
	tests := []struct {
		name string
		gd   *Data
		want bool
	}{
		{
			name: "no releases",
			gd:   &Data{},
			want: false,
		},
		{
			name: "1 release",
			gd: &Data{
				releases: []forgetypes.Release{
					{TagName: "foo"},
				},
			},
			want: true,
		},
		{
			name: "multiple releases",
			gd: &Data{
				releases: []forgetypes.Release{
					{TagName: "foo"},
					{TagName: "bar"},
				},
			},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.gd.hasReleases()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestData_Copy(t *testing.T) {
	tests := []struct {
		name string
		gd   *Data
	}{
		{
			name: "empty",
			gd:   &Data{},
		},
		{
			name: "filled",
			gd: &Data{
				etag: "foo",
				releases: []forgetypes.Release{
					{TagName: "bar"},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.gd.Copy()

			if got.etag != tc.gd.etag {
				t.Errorf("Data.etag mismatch\ngot:  %q\nwant: %q", got.etag, tc.gd.etag)
			}
			if testErr := test.AssertSlicesEqualFunc(
				t,
				got.releases,
				tc.gd.releases,
				releaseEqual,
				"Data.releases",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestData_CopyFrom(t *testing.T) {
	tests := []struct {
		name  string
		fresh *Data
		gd    *Data
	}{
		{
			name: "empty",
			gd:   &Data{},
		},
		{
			name: "filled",
			gd: &Data{
				etag: "foo",
				releases: []forgetypes.Release{
					{TagName: "bar"},
				},
			},
		},
		{
			name: "filled with data to overwrite",
			fresh: &Data{
				etag: "fizz",
				releases: []forgetypes.Release{
					{TagName: "bang"},
				},
			},
			gd: &Data{
				etag: "foo",
				releases: []forgetypes.Release{
					{TagName: "bar"},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.fresh == nil {
				tc.fresh = &Data{}
			}

			tc.fresh.CopyFrom(tc.gd)

			if tc.fresh.etag != tc.gd.etag {
				t.Errorf("Data.etag mismatch\ngot:  %q\nwant: %q", tc.fresh.etag, tc.gd.etag)
			}
			if testErr := test.AssertSlicesEqualFunc(
				t,
				tc.fresh.releases,
				tc.gd.releases,
				releaseEqual,
				"Data.releases",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestGetEmptyListETag(t *testing.T) {
	tests := []struct {
		name string
		etag string
	}{
		{
			name: "weak",
			etag: `W/"abc"`,
		},
		{
			name: "strong",
			etag: `"abc"`,
		},
		{
			name: "empty",
			etag: "",
		},
	}
	hadETag := getEmptyListETag()
	t.Cleanup(func() { setEmptyListETag(hadETag) })

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: shares the empty-list ETag.

			setEmptyListETag(tc.etag)

			if got := getEmptyListETag(); got != tc.etag {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.etag)
			}
		})
	}
}

func TestNewData(t *testing.T) {
	startingEmptyListETag := getEmptyListETag()

	tests := []struct {
		name     string
		etag     string
		releases *[]forgetypes.Release
		want     *Data
	}{
		{
			name:     "no ETag or releases",
			etag:     "",
			releases: nil,
			want: &Data{
				etag: startingEmptyListETag,
			},
		},
		{
			name:     "ETag but no releases",
			etag:     "foo",
			releases: nil,
			want: &Data{
				etag: "foo",
			},
		},
		{
			name: "no ETag but have releases",
			etag: "",
			releases: &[]forgetypes.Release{
				{TagName: "bar"},
			},
			want: &Data{
				etag: startingEmptyListETag,
				releases: []forgetypes.Release{
					{TagName: "bar"},
				},
			},
		},
		{
			name: "ETag and releases",
			etag: "zing",
			releases: &[]forgetypes.Release{
				{TagName: "zap"},
			},
			want: &Data{
				etag: "zing",
				releases: []forgetypes.Release{
					{TagName: "zap"},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := newData(tc.etag, tc.releases)

			if got.etag != tc.want.etag {
				t.Errorf("Data.etag mismatch\ngot:  %q\nwant: %q", got.etag, tc.want.etag)
			}
			if testErr := test.AssertSlicesEqualFunc(
				t,
				got.releases,
				tc.want.releases,
				releaseEqual,
				"Data.releases",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}
