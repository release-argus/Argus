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

//go:build integration

package github

import (
	"maps"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/httpx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/types/forge"
	forgetypes "github.com/release-argus/Argus/service/latest_version/types/forge/api_type"
)

// The tests below assert the shape of the GitHub API, so an API change fails here rather
// than in the unit tests.

// getAPI returns the response and body of a GET against the live API.
func getAPI(t *testing.T, path string, header http.Header) (*http.Response, string) {
	t.Helper()

	request, err := http.NewRequest(http.MethodGet, defaultAPIRoot+path, nil)
	if err != nil {
		t.Fatalf("http.NewRequest(%q) failed: %v", path, err)
	}
	for key, values := range header {
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}
	if request.Header.Get("Authorization") == "" {
		request.Header.Set("Authorization", "Bearer "+test.GitHubToken(t))
	}

	response, err := httpx.Client.Do(request)
	if err != nil {
		test.SkipUnlessRequired(t, "GitHub unreachable for %q: %v", path, err)
	}
	t.Cleanup(func() { _ = response.Body.Close() })

	body, err := httpx.ReadBody(response.Body)
	if err != nil {
		t.Fatalf("httpx.ReadBody(%q) failed: %v", path, err)
	}

	return response, string(body)
}

// getList returns the decoded body of a GET against a live list endpoint.
func getList(t *testing.T, path string) []forgetypes.Release {
	t.Helper()

	_, body := getAPI(t, path, nil)
	list, err := forge.UnmarshalReleases([]byte(body))
	if err != nil {
		t.Fatalf("%s: forge.UnmarshalReleases() failed: %v\nbody: %q", path, err, body)
	}

	return list
}

func TestAPIContract__emptyListIsTwoBytes(t *testing.T) {
	t.Parallel()

	// release-argus/.github publishes no tags.
	_, body := getAPI(t, "/repos/release-argus/.github/tags", nil)

	// Lookup.handleStatusOK matches an empty list on exactly these two bytes.
	if got, want := body, "[]"; got != want {
		t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestAPIContract__tagsCarryANameNotATagName(t *testing.T) {
	t.Parallel()

	tags := getList(t, "/repos/release-argus/Argus/tags?per_page=5")
	if len(tags) == 0 {
		t.Fatal("tags length mismatch\ngot:  0\nwant: > 0")
	}

	// forge.FilterReleases falls back to name.
	for i, tag := range tags {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			if tag.Name == "" {
				t.Error("Release.Name mismatch\ngot:  \"\"\nwant: non-empty")
			}
		})
	}
}

func TestAPIContract__latestReleaseHasMatchingTag(t *testing.T) {
	t.Parallel()

	releases := getList(t, "/repos/release-argus/Argus/releases?per_page=1")
	tags := getList(t, "/repos/release-argus/Argus/tags?per_page=100")
	if gotReleases, gotTags := len(releases), len(tags); gotReleases == 0 || gotTags == 0 {
		t.Fatalf(
			"releases, tags length mismatch\ngot:  %d, %d\nwant: > 0, > 0",
			gotReleases, gotTags,
		)
	}

	latest := releases[0]
	if latest.TagName == "" {
		t.Fatalf("Release.TagName mismatch\ngot:  \"\" (name=%q)\nwant: non-empty", latest.Name)
	}

	tagNames := make(map[string]bool, len(tags))
	for _, tag := range tags {
		tagNames[tag.Name] = true
	}
	if !tagNames[latest.TagName] {
		t.Errorf(
			"tagNames[%q] mismatch\ngot:  absent (keys=%q)\nwant: present",
			latest.TagName, slices.Sorted(maps.Keys(tagNames)),
		)
	}
}

func TestAPIContract__listCarriesAnETag(t *testing.T) {
	t.Parallel()

	response, _ := getAPI(t, "/repos/release-argus/.github/tags", nil)

	// Lookup sends this ETag to make its next request conditional.
	if response.Header.Get("ETag") == "" {
		t.Error("Response.Header[ETag] mismatch\ngot:  \"\"\nwant: non-empty")
	}
}

func TestAPIContract__matchingETagGives304(t *testing.T) {
	t.Parallel()

	response, _ := getAPI(t, "/repos/release-argus/.github/tags", nil)
	etag := response.Header.Get("ETag")
	if etag == "" {
		test.SkipUnlessRequired(t, "no ETag to make a conditional request with")
	}

	header := http.Header{}
	header.Set("If-None-Match", etag)
	conditional, body := getAPI(t, "/repos/release-argus/.github/tags", header)

	wantStatusCode := http.StatusNotModified
	if got := conditional.StatusCode; got != wantStatusCode {
		t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, body, wantStatusCode)
	}
	if body != "" {
		t.Errorf("body mismatch\ngot:  %q\nwant: \"\"", body)
	}
}

func TestAPIContract__nextPageIsNamedInALinkHeader(t *testing.T) {
	t.Parallel()

	response, _ := getAPI(t, "/repos/release-argus/Argus/releases?per_page=1", nil)

	// forge.NextPage reads the next page from this header.
	link := response.Header.Get("Link")
	if want := `rel="next"`; !strings.Contains(link, want) {
		t.Errorf("Response.Header[Link] mismatch\ngot:  %q\nwant: %q", link, want)
	}
}

func TestAPIContract__unknownRepositoryGives404(t *testing.T) {
	t.Parallel()

	response, body := getAPI(t, "/repos/release-argus/no-such-repository-exists/releases", nil)

	wantStatusCode := http.StatusNotFound
	if got := response.StatusCode; got != wantStatusCode {
		t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, body, wantStatusCode)
	}
	if got, want := body, `"Not Found"`; !strings.Contains(got, want) {
		t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestAPIContract__badCredentialGives401(t *testing.T) {
	t.Parallel()

	response, body := getAPI(
		t, "/repos/release-argus/Argus/releases",
		http.Header{"Authorization": {"Bearer not-a-real-token"}},
	)

	if got, want := response.StatusCode, http.StatusUnauthorized; got != want {
		t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, body, want)
	}
	// Lookup.handleStatusUnauthorized reads this to report an invalid token.
	if got, want := body, "Bad credentials"; !strings.Contains(got, want) {
		t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, want)
	}
}
