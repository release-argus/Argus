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
	"maps"
	"net"
	"net/http"
	"slices"
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
var (
	testInstanceHost        = test.ValidCertHTTPS + "/forgejo"
	testInstanceHostBadCert = test.InvalidCertHTTPS + "/forgejo"
)

const (
	repoHappy          = "argus/releases-happy"        // Releases and tags, in version order.
	repoOutOfOrder     = "argus/releases-out-of-order" // A lower version published most recently.
	repoTagsOnly       = "argus/argus-mirror"          // A pull mirror, with tags but no releases.
	repoNothing        = "argus/no-tags-no-releases"   // Empty of both.
	repoReleasesOff    = "argus/releases-disabled"     // Releases unit disabled; tags remain.
	repoPreReleaseOnly = "argus/prerelease-only"       // Every release flagged pre-release.
	repoPrivate        = "argus/private-releases"      // Repository guarded by an access token.
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
		t.Skipf("%s did not answer: %v", target.name, err)
	}
}

// skipIfThrottled is [skipIfUnavailable] for a response read before its status was mapped.
func skipIfThrottled(t *testing.T, target integrationTarget, resp *http.Response) {
	t.Helper()

	if target.thirdParty && resp != nil && isRateLimited(resp) {
		t.Skipf("%s rate-limited (%d)", target.name, resp.StatusCode)
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

func TestLookup_Query__integration(t *testing.T) {
	tests := []struct {
		name            string
		host            string // Empty queries every integration instance.
		repo            string // Empty queries the instance's own repo.
		lookupOverrides string
		defaults        map[string]HostDefaults
		useToken        bool // Give the queried host the instance's read-only token.
		wantPreRelease  bool
		errRegex        string
	}{
		{
			name:     "repository with releases reports a semantic version",
			errRegex: `^$`,
		},
		{
			name:     "not found/repository that does not exist names both causes",
			repo:     "release-argus/no-such-repository",
			errRegex: `^repository not found, or its releases and tags are disabled$`,
		},
		{
			name:     "tag fallback/pull mirror carries git references but no releases",
			host:     testInstanceHost,
			repo:     repoTagsOnly,
			errRegex: `^$`,
		},
		{
			name:     "tag fallback/releases disabled",
			host:     testInstanceHost,
			repo:     repoReleasesOff,
			errRegex: `^$`,
		},
		{
			name: "releases disabled/with the fallback suppressed, the error names both causes",
			host: testInstanceHost,
			repo: repoReleasesOff,
			// A content filter rules tags out, as they carry no release assets.
			lookupOverrides: "require:\n  regex_content: 'xxxx'",
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
			errRegex: `^$`,
		},
		{
			name:     "certificate trust/no host allows them",
			host:     testInstanceHostBadCert,
			repo:     repoHappy,
			errRegex: `certificate|x509`,
		},
		{
			name:     "private repository/the host's token grants access",
			host:     testInstanceHost,
			repo:     repoPrivate,
			useToken: true,
			errRegex: `^$`,
		},
		{
			name:     "private repository/no token gives the ambiguous not-found",
			host:     testInstanceHost,
			repo:     repoPrivate,
			errRegex: `repository not found, or its releases and tags are disabled`,
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

					defaults := tc.defaults
					if tc.useToken {
						defaults = map[string]HostDefaults{
							target.host: {
								URL:         target.host,
								AccessToken: test.ForgejoToken(t),
							},
						}
					}
					setHostDefaults(lookup, defaults, nil)

					_, err := lookup.Query(false, logx.LogFrom{Primary: t.Name()})
					skipIfUnavailable(t, target, err)
					e := errfmt.FormatError(err)
					if !util.RegexCheck(tc.errRegex, e) {
						t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
					}
					if err != nil {
						return
					}

					version := lookup.Status.LatestVersion()
					if version == "" {
						t.Fatal("Status.LatestVersion() mismatch\ngot:  \"\"\nwant: non-empty")
					}

					v, err := semver.NewVersion(version)
					if err != nil {
						t.Fatalf("semver.NewVersion(%q) failed: %v", version, err)
					}

					if got := v.Prerelease() != ""; got != tc.wantPreRelease {
						t.Errorf(
							"semver.NewVersion(%q).Prerelease() non-empty mismatch\n"+
								"got:  %t\nwant: %t",
							version, got, tc.wantPreRelease,
						)
					}
				})
			}
		})
	}
}

func TestLookup_Query__integration__highestVersionWins(t *testing.T) {
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
					lookup := testLookup(
						t,
						test.TrimYAML(`
							host: `+target.host+`
							url: `+target.repo,
						),
					)

					_, err := lookup.Query(false, logFrom)
					skipIfUnavailable(t, target, err)
					if err != nil {
						t.Fatalf("Lookup.Query() failed: %v", err)
					}
					reported, err := semver.NewVersion(lookup.Status.LatestVersion())
					if err != nil {
						t.Fatalf(
							"semver.NewVersion(%q) failed: %v",
							lookup.Status.LatestVersion(), err,
						)
					}

					body, _, err := lookup.httpRequest(endpointReleases, 1, logFrom)
					skipIfUnavailable(t, target, err)
					if err != nil {
						t.Fatalf("releases: Lookup.httpRequest() failed: %v", err)
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
								"Status.LatestVersion() mismatch\ngot:  %q\nwant: >= %q",
								reported, version,
							)
						}
					}
					if compared == 0 {
						t.Fatal("page 1 comparable release count mismatch\ngot:  0\nwant: > 0")
					}

					latestBody, _, err := lookup.httpRequest(endpointReleases+"/latest", 1, logFrom)
					skipIfUnavailable(t, target, err)
					if err != nil {
						t.Fatalf("latest: Lookup.httpRequest() failed: %v", err)
					}
					var latest forgetypes.Release
					if err := decode.Unmarshal("json", latestBody, &latest); err != nil {
						t.Fatalf("decode.Unmarshal() failed: %v", err)
					}
					endpointVersion, err := semver.NewVersion(latest.TagName)
					if err != nil {
						t.Fatalf("semver.NewVersion(%q) failed: %v", latest.TagName, err)
					}

					if endpointVersion.GreaterThan(reported) {
						t.Fatalf(
							"Status.LatestVersion() mismatch\n"+
								"got:  %q\nwant: >= %q (/releases/latest)",
							reported, endpointVersion,
						)
					}
					// Failing here means the fixture repo no longer shows the ordering problem.
					if tc.releasesOutOfOrder && !reported.GreaterThan(endpointVersion) {
						t.Fatalf(
							"Status.LatestVersion() mismatch\n"+
								"got:  %q\nwant: > %q (/releases/latest)",
							reported, endpointVersion,
						)
					}
				})
			}
		})
	}
}

func TestLookup_Query__integration__releasesEndpoint(t *testing.T) {
	for _, target := range integrationTargets("", "") {
		t.Run(target.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, test.TrimYAML(`
				host: `+target.host+`
				url: `+target.repo))
			logFrom := logx.LogFrom{Primary: t.Name()}

			request, err := lookup.createRequest(endpointReleases, 1, logFrom)
			if err != nil {
				t.Fatalf("Lookup.createRequest() failed: %v", err)
			}
			response, releasesJSON, err := lookup.getResponse(request, logFrom)
			skipIfUnavailable(t, target, err)
			skipIfThrottled(t, target, response)
			if err != nil {
				t.Fatalf("Lookup.getResponse() failed: %v", err)
			}

			// If an ETag appears, conditional requests may be worth adding.
			if etag, expected := response.Header.Get("ETag"), ""; etag != expected {
				t.Fatalf("Response.Header[ETag] mismatch\ngot:  %q\nwant: %q", etag, expected)
			}

			tagsJSON, _, err := lookup.httpRequest(endpointTags, 1, logFrom)
			skipIfUnavailable(t, target, err)
			if err != nil {
				t.Fatalf("Lookup.httpRequest(%q) failed: %v", endpointTags, err)
			}

			releases := test.Must(t, func() ([]forgetypes.Release, error) { return forge.UnmarshalReleases(releasesJSON) })
			releases = lookup.filterReleases(releases, logFrom)
			tags := test.Must(t, func() ([]forgetypes.Release, error) { return forge.UnmarshalReleases(tagsJSON) })
			tags = lookup.filterReleases(tags, logFrom)
			if gotReleases, gotTags := len(releases), len(tags); gotReleases == 0 || gotTags == 0 {
				t.Fatalf(
					"releases, tags length mismatch\ngot:  %d, %d\nwant: > 0, > 0",
					gotReleases, gotTags,
				)
			}

			tagNames := make(map[string]bool, len(tags))
			for _, tag := range tags {
				tagNames[tag.SemanticVersion.String()] = true
			}
			// Only the latest is checked, since not every tag must have a release.
			if !tagNames[releases[0].SemanticVersion.String()] {
				t.Fatalf(
					"tagNames[%q] mismatch\ngot:  absent (keys=%q)\nwant: present",
					releases[0].SemanticVersion.String(), slices.Sorted(maps.Keys(tagNames)),
				)
			}
		})
	}
}

func TestLookup_Query__integration__paginatesPastThePageSizeCap(t *testing.T) {
	var paginated bool

	for _, target := range integrationTargets("", "") {
		t.Run(target.name, func(t *testing.T) {
			// Serial: shares paginated, which the parent reads after the loop.

			lookup := testLookup(t, "host: "+target.host+"\nurl: "+target.repo)
			logFrom := logx.LogFrom{Primary: t.Name()}

			body, nextPage, err := lookup.httpRequest(endpointReleases, 1, logFrom)
			skipIfUnavailable(t, target, err)
			if err != nil {
				t.Fatalf("Lookup.httpRequest(page=1) failed: %v", err)
			}

			releases := test.Must(t, func() ([]forgetypes.Release, error) { return forge.UnmarshalReleases(body) })
			if got := len(releases); got > apiPageSize {
				t.Fatalf(
					"page 1: len(forge.UnmarshalReleases()) mismatch\ngot:  %d\nwant: <= %d",
					got, apiPageSize,
				)
			}

			// Not enough releases to fill a page, so nothing to paginate.
			if len(releases) < apiPageSize {
				return
			}

			// A page can be exactly full with nothing after it, so check page 2 is empty.
			if nextPage < 2 {
				beyond, _, err := lookup.httpRequest(endpointReleases, 2, logFrom)
				skipIfUnavailable(t, target, err)
				if err != nil {
					t.Fatalf("Lookup.httpRequest(page=2) failed: %v", err)
				}
				if !isEmptyJSONList(beyond) {
					t.Fatalf("page 2: body mismatch\ngot:  %q\nwant: []", string(beyond))
				}

				return
			}

			nextBody, _, err := lookup.httpRequest(endpointReleases, nextPage, logFrom)
			skipIfUnavailable(t, target, err)
			if err != nil {
				t.Fatalf("Lookup.httpRequest(page=%d) failed: %v", nextPage, err)
			}
			nextReleases := test.Must(t, func() ([]forgetypes.Release, error) { return forge.UnmarshalReleases(nextBody) })
			if len(nextReleases) == 0 {
				t.Fatalf(
					"page %d: len(forge.UnmarshalReleases()) mismatch\ngot:  0\nwant: > 0",
					nextPage,
				)
			}
			if nextReleases[0].TagName == releases[0].TagName {
				t.Fatalf(
					"page %d: Release.TagName mismatch\ngot:  %q\nwant: NOT %q",
					nextPage, nextReleases[0].TagName, releases[0].TagName,
				)
			}

			paginated = true
		})
	}

	if !paginated {
		t.Skipf("no target held more than %d releases - pagination went untested", apiPageSize)
	}
}

func TestLookup_Query__integration__dockerTag(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		errRegex string
	}{
		{
			name:     "found",
			tag:      "{{ version }}",
			errRegex: `^$`,
		},
		{
			name:     "not found",
			tag:      "x{{ version }}",
			errRegex: `release-argus\/argus:x[0-9.]+ - .*tag not found`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, test.TrimYAML(`
				host: `+testInstanceHost+`
				url: `+repoTagsOnly+`
				require:
					docker:
						type: ghcr
						image: `+test.ArgusDockerGHCRRepo+`
						tag: "`+tc.tag+`"
						token: `+test.GitHubToken(t)+`
			`))

			_, err := lookup.Query(false, logx.LogFrom{Primary: t.Name()})
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}
