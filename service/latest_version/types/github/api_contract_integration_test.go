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
	"net/http"
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/httpx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/types/forge"
	forgetypes "github.com/release-argus/Argus/service/latest_version/types/forge/api_type"
)

// These assert the shape of the GitHub API, so that an API change is reported here.

// getAPI returns the response and body of a GET against the live API.
func getAPI(t *testing.T, path string, header http.Header) (*http.Response, string) {
	t.Helper()

	request, err := http.NewRequest(http.MethodGet, defaultAPIRoot+path, nil)
	if err != nil {
		t.Fatalf(
			"%s\ncould not build a request for %q: %v",
			packageName, path, err,
		)
	}
	for key, values := range header {
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}
	request.Header.Set("Authorization", "Bearer "+test.GitHubToken(t))

	response, err := httpx.Client.Do(request)
	if err != nil {
		t.Skipf(
			"%s\nGitHub unreachable for %q: %v",
			packageName, path, err,
		)
	}
	t.Cleanup(func() { _ = response.Body.Close() })

	body, err := httpx.ReadBody(response.Body)
	if err != nil {
		t.Fatalf(
			"%s\ncould not read the body of %q: %v",
			packageName, path, err,
		)
	}

	return response, string(body)
}

// getList returns the decoded body of a GET against a live list endpoint.
func getList(t *testing.T, path string) []forgetypes.Release {
	t.Helper()

	_, body := getAPI(t, path, nil)
	list, err := forge.UnmarshalReleases([]byte(body))
	if err != nil {
		t.Fatalf(
			"%s\n%q no longer decodes as a release list: %v\ngot:  %q",
			packageName, path, err, body,
		)
	}

	return list
}

func TestAPIContract__emptyListIsTwoBytes(t *testing.T) {
	// GIVEN: a repository that publishes no tags.
	// WHEN: its tags are listed.
	_, body := getAPI(t, "/repos/release-argus/.github/tags", nil)

	// THEN: the empty list is exactly "[]", which handleStatusOK matches on.
	if got, want := body, "[]"; got != want {
		t.Errorf(
			"%s\nan empty list is no longer exactly %q\ngot:  %q\nwant: %q",
			packageName, want, got, want,
		)
	}
}

func TestAPIContract__tagsCarryANameNotATagName(t *testing.T) {
	// GIVEN: a repository that publishes tags.
	// WHEN: its tags are listed.
	tags := getList(t, "/repos/release-argus/Argus/tags?per_page=5")
	if len(tags) == 0 {
		t.Fatalf(
			"%s\nthe tags list is empty - this repository no longer exercises the /tags fallback",
			packageName,
		)
	}

	// THEN: every entry names its tag in `name`, which forge.FilterReleases falls back to.
	for _, tag := range tags {
		if tag.Name == "" {
			t.Errorf(
				"%s\na tag no longer names itself in `name`\ngot:  tag_name=%q",
				packageName, tag.TagName,
			)
		}
	}
}

func TestAPIContract__theLatestReleaseHasAMatchingTag(t *testing.T) {
	// GIVEN: a repository that publishes both releases and tags.
	// WHEN: both lists are read.
	releases := getList(t, "/repos/release-argus/Argus/releases?per_page=1")
	tags := getList(t, "/repos/release-argus/Argus/tags?per_page=100")
	if gotReleases, gotTags := len(releases), len(tags); gotReleases == 0 || gotTags == 0 {
		t.Fatalf(
			"%s\nexpected both endpoints to hold versions\ngot:  %d releases, %d tags",
			packageName, gotReleases, gotTags,
		)
	}

	// THEN: the newest release names its tag in `tag_name`.
	latest := releases[0]
	if latest.TagName == "" {
		t.Fatalf(
			"%s\na release no longer names its tag in `tag_name`\ngot:  name=%q",
			packageName, latest.Name,
		)
	}

	// AND: a tag of that name exists, so the /tags fallback reaches the same versions.
	tagNames := make(map[string]bool, len(tags))
	for _, tag := range tags {
		tagNames[tag.Name] = true
	}
	if !tagNames[latest.TagName] {
		t.Errorf(
			"%s\nthe newest release %q has no matching tag in the newest %d\ngot tags: %v",
			packageName, latest.TagName, len(tags), tagNames,
		)
	}
}

func TestAPIContract__listCarriesAnETag(t *testing.T) {
	// WHEN: any list is requested.
	response, _ := getAPI(t, "/repos/release-argus/.github/tags", nil)

	// THEN: it carries an ETag to make the next request conditional with.
	etag := response.Header.Get("ETag")
	if etag == "" {
		t.Errorf(
			"%s\na list no longer carries an ETag\nwant: a non-empty ETag header",
			packageName,
		)
	}
	if strings.HasPrefix(etag, `W/`) {
		t.Logf(
			"%s\nthe ETag is now a weak validator (%q) - the 'W/' strip is load-bearing",
			packageName, etag,
		)
	}
}

func TestAPIContract__matchingETagGives304(t *testing.T) {
	// GIVEN: the ETag of a list already held.
	response, _ := getAPI(t, "/repos/release-argus/.github/tags", nil)
	etag := response.Header.Get("ETag")
	if etag == "" {
		t.Skipf("%s\nno ETag to make a conditional request with", packageName)
	}

	// WHEN: it is asked for again, conditionally with this ETag.
	header := http.Header{}
	header.Set("If-None-Match", etag)
	conditional, body := getAPI(t, "/repos/release-argus/.github/tags", header)

	// THEN: nothing is sent back, and it costs no rate limit.
	wantStatusCode := http.StatusNotModified
	if got := conditional.StatusCode; got != wantStatusCode {
		t.Errorf(
			"%s\nstatus code mismatch on matching If-None-Match\ngot:  %d\nwant: %d",
			packageName, got, wantStatusCode,
		)
	}
	if body != "" {
		t.Errorf(
			"%s\na %d now carries a body\ngot:  %q\nwant: empty",
			packageName, wantStatusCode, body,
		)
	}
}

func TestAPIContract__nextPageIsNamedInALinkHeader(t *testing.T) {
	// GIVEN: a repository with more releases than one page holds.
	// WHEN: the first page is requested.
	response, _ := getAPI(t, "/repos/release-argus/Argus/releases?per_page=1", nil)

	// THEN: the next page is advertised the way forge.NextPage reads it.
	link := response.Header.Get("Link")
	if want := `rel="next"`; !strings.Contains(link, want) {
		t.Errorf(
			"%s\nthe next page is no longer named in a Link header\ngot:  %q\nwant: %q entry",
			packageName, link, want,
		)
	}
}

func TestAPIContract__unknownRepositoryGives404(t *testing.T) {
	// WHEN: a repository that cannot be seen is requested.
	response, body := getAPI(t, "/repos/release-argus/no-such-repository-exists/releases", nil)

	// THEN: it is a 404 naming the cause.
	wantStatusCode := http.StatusNotFound
	if got := response.StatusCode; got != wantStatusCode {
		t.Errorf(
			"%s\nstatus code mismatch on unknown repository\ngot:  %d\nwant: %d",
			packageName, got, wantStatusCode,
		)
	}
	if got, want := body, `"Not Found"`; !strings.Contains(got, want) {
		t.Errorf(
			"%s\nthe %d body no longer names an unknown repository as expected\ngot:  %q\nwant: %q",
			packageName, wantStatusCode,
			got, want,
		)
	}
}

func TestAPIContract__badCredentialGives401(t *testing.T) {
	// GIVEN: a token the API will not accept.
	request, err := http.NewRequest(
		http.MethodGet, defaultAPIRoot+"/repos/release-argus/Argus/releases", nil)
	if err != nil {
		t.Fatalf("%s\ncould not build a request: %v", packageName, err)
	}
	request.Header.Set("Authorization", "Bearer not-a-real-token")

	// WHEN: it is sent.
	response, err := httpx.Client.Do(request)
	if err != nil {
		t.Skipf("%s\nGitHub unreachable: %v", packageName, err)
	}
	t.Cleanup(func() { _ = response.Body.Close() })
	body, _ := httpx.ReadBody(response.Body)

	// THEN: it is a 401 naming the cause, which handleStatusUnauthorized reads.
	wantStatusCode := http.StatusUnauthorized
	if got := response.StatusCode; got != wantStatusCode {
		t.Errorf(
			"%s\nstatus code mismatch on request with bad credential\ngot:  %d\nwant: %d",
			packageName, got, wantStatusCode,
		)
	}
	if got, want := string(body), "Bad credentials"; !strings.Contains(got, want) {
		t.Errorf(
			"%s\nthe %d body no longer names bad repository credentials as expected\ngot:  %q\nwant: %q",
			packageName, wantStatusCode,
			got, want,
		)
	}
}
