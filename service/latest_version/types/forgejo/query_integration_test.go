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
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/Masterminds/semver/v3"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/types/forge"
	forgetypes "github.com/release-argus/Argus/service/latest_version/types/forge/api_type"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

// The project's test instance.
const (
	testInstanceHost        = "https://valid.release-argus.io/forgejo"
	testInstanceHostBadCert = "https://invalid.release-argus.io/forgejo"

	repoHappy          = "argus/releases-happy"        // Releases and tags, in version order.
	repoOutOfOrder     = "argus/releases-out-of-order" // A lower version published most recently.
	repoTagsOnly       = "argus/argus-mirror"          // A pull mirror - git references, no releases.
	repoNothing        = "argus/no-tags-no-releases"   // Empty of both.
	repoReleasesOff    = "argus/releases-disabled"     // Releases unit disabled; tags remain.
	repoPreReleaseOnly = "argus/prerelease-only"       // Every release flagged pre-release.
)

// Public instances to query.
var integrationInstances = []struct {
	name       string
	host       string
	repo       string
	thirdParty bool // May be throttled or unreachable, so a limited case skips.
}{
	{
		name:       "codeberg.org",
		host:       "https://codeberg.org",
		repo:       "forgejo/forgejo",
		thirdParty: true,
	},
	{
		name:       "gitea.com",
		host:       "https://gitea.com",
		repo:       "gitea/tea",
		thirdParty: true,
	},
	{
		name: testInstanceHost,
		host: testInstanceHost,
		repo: repoHappy,
	},
}

// integrationTarget is one live instance a case runs against.
type integrationTarget struct {
	name       string
	host       string
	repo       string
	thirdParty bool
}

// isUnavailable reports whether err is the instance declining to answer rather than
// an answer to assert against: throttling ([handleRateLimited]), or a timeout.
func isUnavailable(err error) bool {
	if err == nil {
		return false
	}
	if strings.Contains(err.Error(), "rate limit reached") {
		return true
	}

	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

// skipIfUnavailable skips a case a third-party instance did not answer, as the project's
// own instance carries the same assertion.
func skipIfUnavailable(t *testing.T, target integrationTarget, err error) {
	t.Helper()

	if target.thirdParty && isUnavailable(err) {
		t.Skipf(
			"%s\n%s did not answer: %v",
			packageName, target.name, err,
		)
	}
}

// unavailable is [skipIfUnavailable] for a case with further targets to assert against.
func unavailable(t *testing.T, target integrationTarget, err error) bool {
	t.Helper()

	if target.thirdParty && isUnavailable(err) {
		t.Logf(
			"%s\n%s did not answer, moving on: %v",
			packageName, target.name, err,
		)
		return true
	}

	return false
}

// skipIfThrottled is [skipIfUnavailable] for a response read before its status was mapped.
func skipIfThrottled(t *testing.T, target integrationTarget, resp *http.Response) {
	t.Helper()

	if target.thirdParty && resp != nil && isRateLimited(resp) {
		t.Skipf(
			"%s\n%s rate-limited (%d)",
			packageName, target.name, resp.StatusCode,
		)
	}
}

// integrationTargets resolves the instances a case runs against: the host it names,
// or every entry of [integrationInstances] when it names none.
func integrationTargets(host, repo string) []integrationTarget {
	if host != "" {
		return []integrationTarget{{name: host, host: host, repo: repo}}
	}

	targets := make([]integrationTarget, 0, len(integrationInstances))
	for _, instance := range integrationInstances {
		targets = append(targets, integrationTarget{
			name:       instance.name,
			host:       instance.host,
			repo:       util.FirstNonDefault(repo, instance.repo),
			thirdParty: instance.thirdParty,
		})
	}

	return targets
}

func TestLookup_Query__Integration(t *testing.T) {
	// GIVEN: a repository on a live instance, and the outcome it must produce.
	tests := []struct {
		name            string
		host            string // host to query. Empty = all integration instances.
		repo            string // repo to query. Empty = instance's own.
		lookupOverrides string
		defaults        map[string]HostDefaults
		wantVersion     bool
		wantPreRelease  bool
		errRegex        string
	}{
		{
			name:        "repository with releases reports a semantic version",
			wantVersion: true,
			errRegex:    `^$`,
		},
		{
			name:     "not found/repository that does not exist names both causes",
			repo:     "release-argus/no-such-repository",
			errRegex: `^repository not found, or its releases and tags are disabled$`,
		},
		{
			name:        "tag fallback/pull mirror carries git references but no releases",
			host:        testInstanceHost,
			repo:        repoTagsOnly,
			wantVersion: true,
			errRegex:    `^$`,
		},
		{
			name:        "tag fallback/releases disabled",
			host:        testInstanceHost,
			repo:        repoReleasesOff,
			wantVersion: true,
			errRegex:    `^$`,
		},
		{
			name: "releases disabled/with the fallback suppressed, the error names both causes",
			host: testInstanceHost,
			repo: repoReleasesOff,
			// Pre-releases rule tags out.
			lookupOverrides: "use_prerelease: true",
			errRegex:        `^repository not found, or its releases are disabled$`,
		},
		{
			name:     "nothing to find/no releases nor tags doesn't give missing-repository claim",
			host:     testInstanceHost,
			repo:     repoNothing,
			errRegex: `^no releases or tags were found$`,
		},
		{
			name:     "pre-releases/excluded by default",
			host:     testInstanceHost,
			repo:     repoPreReleaseOnly,
			errRegex: `^no releases were found matching the url_commands on page 1 of the API response$`,
		},
		{
			name:            "pre-releases/opted into per service",
			host:            testInstanceHost,
			repo:            repoPreReleaseOnly,
			lookupOverrides: "use_prerelease: true",
			wantVersion:     true,
			wantPreRelease:  true,
			errRegex:        `^$`,
		},
		{
			name: "certificate trust/queried host allows invalid certificates",
			host: testInstanceHostBadCert,
			repo: repoHappy,
			defaults: map[string]HostDefaults{
				testInstanceHostBadCert: {
					AllowInvalidCerts: new(true),
				},
			},
			wantVersion: true,
			errRegex:    `^$`,
		},
		{
			name:     "certificate trust/no host allows them",
			host:     testInstanceHostBadCert,
			repo:     repoHappy,
			errRegex: `certificate|x509`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for _, target := range integrationTargets(tc.host, tc.repo) {
				t.Run(target.name, func(t *testing.T) {
					t.Parallel()

					lookup := testLookup(t, test.TrimYAML(`
						host: `+target.host+`
						url: `+target.repo+`
						`+tc.lookupOverrides))
					setHostDefaults(lookup, tc.defaults, nil)

					// WHEN: it is queried.
					_, err := lookup.Query(false, logx.LogFrom{Primary: t.Name()})
					skipIfUnavailable(t, target, err)

					prefix := fmt.Sprintf(
						"%s\nLookup{host=%q, url=%q}.Query()",
						packageName, target.host, target.repo,
					)

					// THEN: any error is as expected.
					e := errfmt.FormatError(err)
					if !util.RegexCheck(tc.errRegex, e) {
						t.Fatalf(
							"%s error mismatch\ngot:  %q\nwant: %q",
							prefix, e, tc.errRegex,
						)
					}

					// AND: a version is reported only when one was expected.
					version := lookup.Status.LatestVersion()
					if tc.wantVersion != (version != "") {
						t.Fatalf(
							"%s version mismatch\ngot:  %q\nwant a version: %t",
							prefix, version, tc.wantVersion,
						)
					}
					if !tc.wantVersion {
						return
					}

					// AND: it parses as a semantic version.
					parsed, err := semver.NewVersion(version)
					if err != nil {
						t.Fatalf(
							"%s version %q is not semantic: %v",
							prefix, version, err,
						)
					}

					// AND: it is a pre-release only where one was asked for.
					if tc.wantPreRelease && parsed.Prerelease() == "" {
						t.Fatalf(
							"%s reported %q, which is not a pre-release",
							prefix, version,
						)
					}
				})
			}
		})
	}
}

func TestLookup_Query__Integration_HighestVersionWins(t *testing.T) {
	// GIVEN: a repository on a live instance.
	tests := []struct {
		name               string
		host               string
		repo               string
		releasesOutOfOrder bool
	}{
		{
			name:               "releases published in version order",
			host:               "",
			repo:               "",
			releasesOutOfOrder: false,
		},
		{
			name:               "releases published out of version order",
			host:               testInstanceHost,
			repo:               repoOutOfOrder,
			releasesOutOfOrder: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for _, target := range integrationTargets(tc.host, tc.repo) {
				t.Run(target.name, func(t *testing.T) {
					t.Parallel()

					logFrom := logx.LogFrom{Primary: t.Name()}
					lookup := testLookup(t, test.TrimYAML(`
						host: `+target.host+`
						url: `+target.repo))

					prefix := fmt.Sprintf(
						"%s\nLookup{host: %q, url: %q}.Query()",
						packageName, target.host, target.repo,
					)

					// WHEN: it is queried.
					_, err := lookup.Query(false, logFrom)
					skipIfUnavailable(t, target, err)
					if err != nil {
						t.Fatalf("%s unexpected error: %v", prefix, err)
					}
					reported, err := semver.NewVersion(lookup.Status.LatestVersion())
					if err != nil {
						t.Fatalf(
							"%s version %q is not semantic: %v",
							prefix, lookup.Status.LatestVersion(), err,
						)
					}

					// THEN: no release on the first page sorts above it.
					body, _, err := lookup.httpRequest(endpointReleases, 1, logFrom)
					skipIfUnavailable(t, target, err)
					if err != nil {
						t.Fatalf("%s reading page 1: %v", prefix, err)
					}
					releases := test.Must(t, func() ([]forgetypes.Release, error) {
						return forge.UnmarshalReleases(body)
					})
					compared := 0
					for _, release := range releases {
						if release.PreRelease {
							continue
						}
						version, err := semver.NewVersion(
							util.FirstNonDefault(release.TagName, release.Name))
						if err != nil {
							continue
						}

						compared++
						if version.GreaterThan(reported) {
							t.Fatalf(
								"%s reported %q, but %q is higher",
								prefix, reported, version,
							)
						}
					}
					if compared == 0 {
						t.Fatalf("%s no release on page 1 was comparable, so nothing was asserted", prefix)
					}

					// AND: the forge's own latest-release endpoint never answers higher -
					// and answers strictly lower where the repository is out of order.
					latestBody, _, err := lookup.httpRequest(endpointReleases+"/latest", 1, logFrom)
					skipIfUnavailable(t, target, err)
					if err != nil {
						t.Fatalf(
							"%s reading the latest-release endpoint: %v",
							prefix, err,
						)
					}
					var latest forgetypes.Release
					if err := decode.Unmarshal("json", latestBody, &latest); err != nil {
						t.Fatalf(
							"%s decoding the latest release: %v",
							prefix, err,
						)
					}
					endpointVersion, err := semver.NewVersion(latest.TagName)
					if err != nil {
						t.Fatalf(
							"%s the latest-release endpoint gave %q, which is not semantic: %v",
							prefix, latest.TagName, err,
						)
					}

					if endpointVersion.GreaterThan(reported) {
						t.Fatalf(
							"%s reported %q, below the latest-release endpoint's %q",
							prefix, reported, endpointVersion,
						)
					}
					if tc.releasesOutOfOrder && !reported.GreaterThan(endpointVersion) {
						t.Fatalf(
							"%s reported %q and the latest-release endpoint gave %q - this"+
								" repository no longer demonstrates the ordering problem",
							prefix, reported, endpointVersion,
						)
					}
				})
			}
		})
	}
}

func TestLookup_Query__Integration_ReleasesEndpoint(t *testing.T) {
	// GIVEN: a repository on a public instance.
	for _, target := range integrationTargets("", "") {
		t.Run(target.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, test.TrimYAML(`
				host: `+target.host+`
				url: `+target.repo))
			logFrom := logx.LogFrom{Primary: t.Name()}

			// WHEN: the releases endpoint is read.
			request, err := lookup.createRequest(endpointReleases, 1, logFrom)
			if err != nil {
				t.Fatalf(
					"%s\nLookup.createRequest() unexpected error: %v",
					packageName, err,
				)
			}
			response, releasesBody, err := lookup.getResponse(request, logFrom)
			skipIfUnavailable(t, target, err)
			skipIfThrottled(t, target, response)
			if err != nil {
				t.Fatalf(
					"%s\nLookup.getResponse() unexpected error: %v",
					packageName, err,
				)
			}

			// THEN: the response carries no ETag.
			if eTag := response.Header.Get("ETag"); eTag != "" {
				t.Fatalf(
					"%s\n%s issued an ETag %q - conditional requests may now be worth having",
					packageName, target.host, eTag,
				)
			}

			// AND: the tags endpoint holds versions too.
			tagsBody, _, err := lookup.httpRequest(endpointTags, 1, logFrom)
			skipIfUnavailable(t, target, err)
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

			// AND: every release has a tag of the same name
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

func TestLookup_Query__Integration_PaginatesPastThePageSizeCap(t *testing.T) {
	// GIVEN: the public instances, at least one of which holds more releases than a page can.
	logFrom := logx.LogFrom{Primary: t.Name()}
	var paginated bool

	for _, target := range integrationTargets("", "") {
		lookup := testLookup(t, test.TrimYAML(`
				host: `+target.host+`
				url: `+target.repo))

		// WHEN: the first page is requested.
		body, nextPage, err := lookup.httpRequest(endpointReleases, 1, logFrom)
		if unavailable(t, target, err) {
			continue
		}
		if err != nil {
			t.Fatalf(
				"%s\nLookup.httpRequest(%q) on %s unexpected error: %v",
				packageName, endpointReleases, target.host, err,
			)
		}

		// THEN: the target never serves more than limit in one response.
		releases := test.Must(t, func() ([]forgetypes.Release, error) { return (forge.UnmarshalReleases(body)) })
		if got := len(releases); got > apiPageSize {
			t.Fatalf(
				"%s\n%s served %d releases on one page, above the cap of %d",
				packageName, target.host, got, apiPageSize,
			)
		}

		// Not enough releases to fill a page - nothing to paginate.
		if len(releases) < apiPageSize {
			continue
		}

		// A page can be exactly full with nothing after it. Confirm rather than assume.
		if nextPage < 2 {
			beyond, _, err := lookup.httpRequest(endpointReleases, 2, logFrom)
			if unavailable(t, target, err) {
				continue
			}
			if err != nil {
				t.Fatalf(
					"%s\nLookup.httpRequest(%q, page 2) on %s unexpected error: %v",
					packageName, endpointReleases, target.host, err,
				)
			}
			if !isEmptyJSONList(beyond) {
				t.Fatalf(
					"%s\n%s advertised no next page, but page 2 of %s holds releases\ngot:  %q",
					packageName, target.host, target.repo, string(beyond),
				)
			}

			continue
		}

		// AND: where a next page is advertised, it holds different releases.
		nextBody, _, err := lookup.httpRequest(endpointReleases, nextPage, logFrom)
		if unavailable(t, target, err) {
			continue
		}
		if err != nil {
			t.Fatalf(
				"%s\nLookup.httpRequest(%q, page %d) on %s unexpected error: %v",
				packageName, endpointReleases, nextPage, target.host, err,
			)
		}
		nextReleases := test.Must(t, func() ([]forgetypes.Release, error) { return (forge.UnmarshalReleases(nextBody)) })
		if len(nextReleases) == 0 {
			t.Fatalf(
				"%s\npage %d of %s held no releases",
				packageName, nextPage, target.repo,
			)
		}
		if nextReleases[0].TagName == releases[0].TagName {
			t.Fatalf(
				"%s\npage %d of %s repeated page 1 - the page parameter had no effect",
				packageName, nextPage, target.repo,
			)
		}

		paginated = true
	}

	if !paginated {
		t.Skipf(
			"%s\nno target held more than %d releases - pagination went untested",
			packageName, apiPageSize,
		)
	}
}
