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
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	ghtypes "github.com/release-argus/Argus/service/latest_version/types/github/api_type"
)

var emptyListETagTestMu = sync.Mutex{}

func TestSetEmptyListETag(t *testing.T) {
	// GIVEN: emptyListETag is set to the incorrect value.
	emptyListETagTestMu.Lock()
	t.Cleanup(emptyListETagTestMu.Unlock)
	incorrectValue := "foo"
	setEmptyListETag(incorrectValue)

	// WHEN: SetEmptyListETag is called.
	SetEmptyListETag(test.GitHubToken(t))

	prefix := fmt.Sprintf("%s\nSetEmptyListETag(val)", packageName)

	// THEN: the emptyListETag is set.
	got := getEmptyListETag()
	if incorrectValue == got {
		t.Errorf(
			"%s didn't change emptyListETag from getEmptyListETag()\ngot:  %q\nwant: %q",
			prefix, got, emptyListETag,
		)
	}
	if got != initialEmptyListETag {
		t.Errorf(
			"%s changed empty list ETag incorrectly\ngot:  %q\nwant: %q",
			prefix, got, initialEmptyListETag,
		)
	}
}

func TestGetEmptyListETag(t *testing.T) {
	// GIVEN: emptyListETag exists.
	emptyListETagTestMu.Lock()
	t.Cleanup(emptyListETagTestMu.Unlock)
	emptyListETagMu.RLock()
	t.Cleanup(emptyListETagMu.RUnlock)

	// WHEN: getEmptyListETag is called.
	got := getEmptyListETag()

	// THEN: the emptyListETag is returned.
	if emptyListETag != got {
		t.Errorf(
			"%s\ngetEmptyListETag() mismatch\ngot:  %q\nwant: %q",
			packageName, got, emptyListETag,
		)
	}
}

func TestNewData(t *testing.T) {
	emptyListETagTestMu.Lock()
	t.Cleanup(emptyListETagTestMu.Unlock)
	startingEmptyListETag := getEmptyListETag()
	// GIVEN: a Data is wanted with/without an eTag/releases.
	tests := []struct {
		name     string
		eTag     string
		releases *[]ghtypes.Release
		want     *Data
	}{
		{
			name:     "no eTag or releases",
			eTag:     "",
			releases: nil,
			want: &Data{
				eTag:     startingEmptyListETag,
				releases: []ghtypes.Release{},
			},
		},
		{
			name:     "eTag but no releases",
			eTag:     "foo",
			releases: nil,
			want: &Data{
				eTag:     "foo",
				releases: []ghtypes.Release{},
			},
		},
		{
			name: "no eTag but releases",
			eTag: "",
			releases: &[]ghtypes.Release{
				{TagName: "bar"},
			},
			want: &Data{
				eTag: startingEmptyListETag,
				releases: []ghtypes.Release{
					{TagName: "bar"},
				},
			},
		},
		{
			name: "eTag and releases",
			eTag: "zing",
			releases: &[]ghtypes.Release{
				{TagName: "zap"},
			},
			want: &Data{
				eTag: "zing",
				releases: []ghtypes.Release{
					{TagName: "zap"},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: newData is called.
			got := newData(tc.eTag, tc.releases)

			prefix := fmt.Sprintf(
				"%s\nnewData(eTag=%q, releases=%v)",
				packageName, tc.eTag, tc.releases,
			)

			// THEN: the correct Data is returned.
			if got.eTag != tc.want.eTag {
				t.Errorf(
					"%s eTag mismatch\ngot %q\nwant: %q",
					prefix, got.eTag, tc.want.eTag,
				)
			}
			if err := test.AssertSlicesEqualFunc(
				t,
				got.releases,
				tc.want.releases,
				func(a, b ghtypes.Release) bool { return a.String() == b.String() },
				prefix,
				"Data.releases",
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestData_String(t *testing.T) {
	// GIVEN: a Data.
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
				eTag: "argus",
				releases: []ghtypes.Release{
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

			// WHEN: the Data is stringified with String.
			got := tc.githubData.String()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nData.String() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestData_TagFallback(t *testing.T) {
	// GIVEN: a Data.
	gd := &Data{}
	tests := []bool{
		true,
		false,
		true,
		false,
		true,
	}

	if gd.tagFallback != false {
		t.Fatalf("%s\nData.tagFallback wasn't set to false initially", packageName)
	}

	for _, tc := range tests {
		gd.SetTagFallback()

		// WHEN: TagFallback is called.
		got := gd.TagFallback()

		// THEN: the correct value is returned.
		if got != tc {
			t.Errorf(
				"%s\nData.TagFallback() mismatch\ngot:  %t\nwant: %t",
				packageName, got, tc,
			)
		}
	}
}

func TestData_ETag(t *testing.T) {
	// GIVEN: a Data.
	testData := &Data{}

	// WHEN: ETag is called.
	got := testData.ETag()

	// THEN: the releases are returned.
	want := testData.eTag
	if got != want {
		t.Errorf(
			"%s\nfresh Data.ETag() mismatch\ngot:  %q\nwant: %q",
			packageName, got, want,
		)
	}

	// WHEN: the releases are changed.
	newETag := "foo"
	testData.SetETag(newETag)

	// THEN: the new releases can be fetched.
	got = testData.ETag()
	want = newETag
	if got != want {
		t.Errorf(
			"%s\nData.ETag() mismatch after SetETag(%q)\ngot:  %q\nwant: %q",
			packageName, newETag,
			got, want,
		)
	}
}

func TestData_PerPage(t *testing.T) {
	for range 10 {
		// GIVEN: a Data instance.
		data := &Data{}

		// WHEN: SetPerPage is called with a new value.
		foundOnPage := rand.Intn(10) + 1
		data.SetPerPage(foundOnPage)

		// THEN: the PerPage field is updated correctly.
		want := foundOnPage * defaultPerPage
		if got := data.PerPage(); got != want {
			t.Errorf(
				"%s\nData.SetPerPage(%d) failed to update .PerPage()\ngot:  %d\nwant: %d",
				packageName, foundOnPage,
				got, want,
			)
		}

		// WHEN: ResetPerPage is called.
		data.ResetPerPage()

		// THEN: the PerPage field is reset to 0.
		if got := data.PerPage(); got != 0 {
			t.Errorf(
				"%s\nData.PerPage() mismatch after ResetPerPage()\ngot:  %d\nwant: %d",
				packageName, got, 0,
			)
		}
	}
}

func TestData_Releases(t *testing.T) {
	// GIVEN: a Data.
	testData := &Data{}

	// WHEN: Releases is called.
	got := testData.Releases()

	// THEN: the releases are returned.
	want := testData.releases
	match := len(got) == len(want)
	if match {
		for i, release := range got {
			if release.String() != want[i].String() {
				match = false
				break
			}
		}
	}
	if !match {
		t.Errorf(
			"%s\nfresh Data.Releases() mismatch\ngot:  %v\nwant: %v",
			packageName, got, want,
		)
	}

	// WHEN: the releases are changed.
	newReleases := []ghtypes.Release{
		{TagName: "foo"},
		{TagName: "bar"},
	}
	testData.SetReleases(newReleases)

	// THEN: the new releases can be fetched.
	got = testData.Releases()
	if err := test.AssertSlicesEqualFunc(
		t,
		got,
		newReleases,
		func(a, b ghtypes.Release) bool { return a.String() == b.String() },
		fmt.Sprintf("%s\nData.Releases", packageName),
		"",
	); err != nil {
		t.Fatal(err)
	}
}

func TestData_HasReleases(t *testing.T) {
	// GIVEN: a Data that may/may not have releases.
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
				releases: []ghtypes.Release{
					{TagName: "foo"},
				},
			},
			want: true,
		},
		{
			name: "multiple releases",
			gd: &Data{
				releases: []ghtypes.Release{
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

			// WHEN: hasReleases is called on it.
			got := tc.gd.hasReleases()

			// THEN: the correct value is returned.
			if got != tc.want {
				t.Errorf(
					"%s\nData.hasReleases() mismatch\ngot:  %v\nwant: %v",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestData_Copy(t *testing.T) {
	// GIVEN: a Data to copy from.
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
				eTag: "foo",
				releases: []ghtypes.Release{
					{TagName: "bar"},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Copy is called.
			got := tc.gd.Copy()

			// THEN: the correct Data is returned.
			if got.eTag != tc.gd.eTag {
				t.Errorf(
					"%s\nData.Copy() .eTag mismatch\ngot:  %q\nwant: %q",
					packageName, got.eTag, tc.gd.eTag,
				)
			}
			if err := test.AssertSlicesEqualFunc(
				t,
				got.releases,
				tc.gd.releases,
				func(a, b ghtypes.Release) bool { return a.String() == b.String() },
				fmt.Sprintf("%s\nData.Copy()", packageName),
				".releases",
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestData_CopyFrom(t *testing.T) {
	// GIVEN: a fresh Data and a Data to copy from.
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
				eTag: "foo",
				releases: []ghtypes.Release{
					{TagName: "bar"},
				},
			},
		},
		{
			name: "filled with data to overwrite",
			fresh: &Data{
				eTag: "fizz",
				releases: []ghtypes.Release{
					{TagName: "bang"},
				},
			},
			gd: &Data{
				eTag: "foo",
				releases: []ghtypes.Release{
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

			// WHEN: CopyFrom is called.
			tc.fresh.CopyFrom(tc.gd)

			// THEN: the correct Data is returned.
			if tc.fresh.eTag != tc.gd.eTag {
				t.Errorf(
					"%s\nData.CopyFrom() .eTag mismatch\ngot:  %q\nwant: %q",
					packageName, tc.fresh.eTag, tc.gd.eTag,
				)
			}
			if err := test.AssertSlicesEqualFunc(
				t,
				tc.fresh.releases,
				tc.gd.releases,
				func(a, b ghtypes.Release) bool { return a.String() == b.String() },
				fmt.Sprintf(
					"%s\nCopyFrom(%v)",
					packageName, tc.gd,
				),
				".releases",
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}
