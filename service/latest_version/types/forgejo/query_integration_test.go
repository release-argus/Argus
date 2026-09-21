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

// Package forgejo provides a Forgejo-based lookup type.
package forgejo

import (
	"testing"

	"github.com/Masterminds/semver/v3"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/types/forge"
	forgetypes "github.com/release-argus/Argus/service/latest_version/types/forge/api_type"
)

// Public instances to query.
var integrationInstances = []struct {
	name string
	host string
	repo string
}{
	{
		name: "codeberg.org",
		host: "https://codeberg.org",
		repo: "forgejo/forgejo",
	},
	{
		name: "gitea.com",
		host: "https://gitea.com",
		repo: "gitea/tea",
	},
}

func TestLookup_Query__Integration(t *testing.T) {
	// GIVEN: a repository on a public instance, queried unauthenticated.
	for _, instance := range integrationInstances {
		t.Run(instance.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, test.TrimYAML(`
				host: `+instance.host+`
				url: `+instance.repo))

			// WHEN: it is queried.
			if _, err := lookup.Query(false, logx.LogFrom{Primary: t.Name()}); err != nil {
				t.Fatalf(
					"%s\nLookup.Query() unexpected error: %v",
					packageName, err,
				)
			}

			// THEN: a version is reported.
			version := lookup.Status.LatestVersion()
			if version == "" {
				t.Fatalf("%s\nLookup.Query() reported no version", packageName)
			}

			// AND: it parses as a semantic version.
			if _, err := semver.NewVersion(version); err != nil {
				t.Fatalf(
					"%s\nLookup.Query() version %q is not semantic: %v",
					packageName, version, err,
				)
			}
		})
	}
}

func TestLookup_Query__Integration_HighestVersionWins(t *testing.T) {
	// GIVEN: a repository whose releases are not published in version order.
	for _, instance := range integrationInstances {
		t.Run(instance.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, test.TrimYAML(`
				host: `+instance.host+`
				url: `+instance.repo))
			logFrom := logx.LogFrom{Primary: t.Name()}

			if _, err := lookup.Query(false, logFrom); err != nil {
				t.Fatalf(
					"%s\nLookup.Query() unexpected error: %v",
					packageName, err,
				)
			}
			lv := lookup.Status.LatestVersion()
			reported, err := semver.NewVersion(lv)
			if err != nil {
				t.Fatalf(
					"%s\nLookup.Query() version %q is not semantic: %v",
					packageName, lv, err,
				)
			}

			// WHEN: every release on the first page is read.
			body, _, err := lookup.httpRequest(endpointReleases, 1, logFrom)
			if err != nil {
				t.Fatalf(
					"%s\nLookup.httpRequest(%q) unexpected error: %v",
					packageName, endpointReleases, err,
				)
			}
			releases := test.Must(t, func() ([]forgetypes.Release, error) { return (forge.UnmarshalReleases(body)) })
			if len(releases) == 0 {
				t.Fatalf(
					"%s\nthe first page of %s held no usable releases",
					packageName, endpointReleases,
				)
			}

			// THEN: no release on it sorts above the one reported - the highest semantic
			// version wins, not the most recently published.
			for _, release := range releases {
				if release.SemanticVersion.GreaterThan(reported) {
					t.Fatalf(
						"%s\nLookup.Query() reported %q, but %q is higher",
						packageName, reported, release.SemanticVersion,
					)
				}
			}
		})
	}
}

func TestLookup_Query__Integration_ReleasesAndTagsAgree(t *testing.T) {
	// GIVEN: a repository on a public instance.
	for _, instance := range integrationInstances {
		t.Run(instance.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, test.TrimYAML(`
				host: `+instance.host+`
				url: `+instance.repo))
			logFrom := logx.LogFrom{Primary: t.Name()}

			// WHEN: both endpoints are read.
			releasesBody, _, err := lookup.httpRequest(endpointReleases, 1, logFrom)
			if err != nil {
				t.Fatalf(
					"%s\nLookup.httpRequest(%q) unexpected error: %v",
					packageName, endpointReleases, err,
				)
			}
			tagsBody, _, err := lookup.httpRequest(endpointTags, 1, logFrom)
			if err != nil {
				t.Fatalf(
					"%s\nLookup.httpRequest(%q) unexpected error: %v",
					packageName, endpointTags, err,
				)
			}

			releases := test.Must(t, func() ([]forgetypes.Release, error) { return (forge.UnmarshalReleases(releasesBody)) })
			releases = lookup.filterReleases(releases, logFrom)
			tags := test.Must(t, func() ([]forgetypes.Release, error) { return (forge.UnmarshalReleases(tagsBody)) })
			tags = lookup.filterReleases(tags, logFrom)
			if gotReleases, gotTags := len(releases), len(tags); gotReleases == 0 || gotTags == 0 {
				t.Fatalf(
					"%s\nexpected both endpoints to hold versions\ngot:  %d releases, %d tags",
					packageName, gotReleases, gotTags,
				)
			}

			// THEN: every release has a tag of the same name
			// (only check latest since not every tag must have a release).
			tagNames := make(map[string]bool, len(tags))
			for _, tag := range tags {
				tagNames[tag.SemanticVersion.String()] = true
			}
			if !tagNames[releases[0].SemanticVersion.String()] {
				t.Fatalf(
					"%s\nthe highest release %q has no matching tag\ngot tags: %v",
					packageName, releases[0].SemanticVersion, tagNames,
				)
			}
		})
	}
}

func TestLookup_Query__Integration_NoETagIssued(t *testing.T) {
	// GIVEN: a repository on a public instance.
	for _, instance := range integrationInstances {
		t.Run(instance.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, test.TrimYAML(`
				host: `+instance.host+`
				url: `+instance.repo))
			logFrom := logx.LogFrom{Primary: t.Name()}

			request, err := lookup.createRequest(endpointReleases, 1, logFrom)
			if err != nil {
				t.Fatalf(
					"%s\nLookup.createRequest() unexpected error: %v",
					packageName, err,
				)
			}

			// WHEN: the releases endpoint is requested.
			response, _, err := getResponse(request, logFrom)
			if err != nil {
				t.Fatalf(
					"%s\ngetResponse() unexpected error: %v",
					packageName, err,
				)
			}

			// THEN: the response carries no ETag.
			if eTag := response.Header.Get("ETag"); eTag != "" {
				t.Fatalf(
					"%s\n%s issued an ETag %q - conditional requests may now be worth having",
					packageName, instance.host, eTag,
				)
			}
		})
	}
}

func TestLookup_Query__Integration_NotFound(t *testing.T) {
	// GIVEN: a repository that does not exist on a public instance.
	for _, instance := range integrationInstances {
		t.Run(instance.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, test.TrimYAML(`
				host: `+instance.host+`
				url: release-argus/no-such-repository`))

			// WHEN: it is queried.
			_, err := lookup.Query(false, logx.LogFrom{Primary: t.Name()})

			// THEN: the error names both possible causes.
			want := "repository not found, or its releases and tags are disabled"
			if err == nil || err.Error() != want {
				t.Fatalf(
					"%s\nLookup.Query() error mismatch\ngot:  %v\nwant: %q",
					packageName, err, want,
				)
			}
		})
	}
}

func TestLookup_Query__Integration_PaginatesPastThePageSizeCap(t *testing.T) {
	// GIVEN: the public instances, at least one of which holds more releases than a page can.
	logFrom := logx.LogFrom{Primary: t.Name()}
	var paginated bool

	for _, instance := range integrationInstances {
		lookup := testLookup(t, test.TrimYAML(`
				host: `+instance.host+`
				url: `+instance.repo))

		// WHEN: the first page is requested.
		body, nextPage, err := lookup.httpRequest(endpointReleases, 1, logFrom)
		if err != nil {
			t.Fatalf(
				"%s\nLookup.httpRequest(%q) on %s unexpected error: %v",
				packageName, endpointReleases, instance.host, err,
			)
		}

		// THEN: the instance never serves more than limit in one response.
		releases := test.Must(t, func() ([]forgetypes.Release, error) { return (forge.UnmarshalReleases(body)) })
		if got := len(releases); got > apiPageSize {
			t.Fatalf(
				"%s\n%s served %d releases on one page, above the cap of %d",
				packageName, instance.host, got, apiPageSize,
			)
		}

		// Not enough releases to fill a page - nothing to paginate.
		if len(releases) < apiPageSize {
			continue
		}

		// A page can be exactly full with nothing after it. Confirm rather than assume.
		if nextPage < 2 {
			beyond, _, err := lookup.httpRequest(endpointReleases, 2, logFrom)
			if err != nil {
				t.Fatalf(
					"%s\nLookup.httpRequest(%q, page 2) on %s unexpected error: %v",
					packageName, endpointReleases, instance.host, err,
				)
			}
			if !isEmptyJSONList(beyond) {
				t.Fatalf(
					"%s\n%s advertised no next page, but page 2 of %s holds releases\ngot:  %q",
					packageName, instance.host, instance.repo, string(beyond),
				)
			}

			continue
		}

		// AND: where a next page is advertised, it holds different releases.
		nextBody, _, err := lookup.httpRequest(endpointReleases, nextPage, logFrom)
		if err != nil {
			t.Fatalf(
				"%s\nLookup.httpRequest(%q, page %d) on %s unexpected error: %v",
				packageName, endpointReleases, nextPage, instance.host, err,
			)
		}
		nextReleases := test.Must(t, func() ([]forgetypes.Release, error) { return (forge.UnmarshalReleases(nextBody)) })
		if len(nextReleases) == 0 {
			t.Fatalf(
				"%s\npage %d of %s held no releases",
				packageName, nextPage, instance.repo,
			)
		}
		if nextReleases[0].TagName == releases[0].TagName {
			t.Fatalf(
				"%s\npage %d of %s repeated page 1 - the page parameter had no effect",
				packageName, nextPage, instance.repo,
			)
		}

		paginated = true
	}

	if !paginated {
		t.Skipf(
			"%s\nno instance held more than %d releases - pagination went untested",
			packageName, apiPageSize,
		)
	}
}
