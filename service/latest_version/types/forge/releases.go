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

// Package forge provides the release handling shared by git forge lookup types.
package forge

import (
	"net/url"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/Masterminds/semver/v3"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/service/latest_version/filter"
	forgetypes "github.com/release-argus/Argus/service/latest_version/types/forge/api_type"
	"github.com/release-argus/Argus/util"
)

// UnmarshalReleases validates that the response body conforms to the JSON formatting.
func UnmarshalReleases(body []byte) ([]forgetypes.Release, error) {
	var releases []forgetypes.Release
	if err := decode.Unmarshal("json", body, &releases); err != nil {
		return nil, err //nolint:wrapcheck
	}

	return releases, nil
}

// ownerRepoSegment matches a segment a forge API can be addressed with.
var ownerRepoSegment = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// IsOwnerRepo reports whether path is an "owner/repo" a forge API can be addressed
// with.
func IsOwnerRepo(path string) bool {
	owner, repo, _ := strings.Cut(path, "/")

	return path == strings.TrimSpace(path) &&
		isRepoSegment(owner) && isRepoSegment(repo)
}

// isRepoSegment reports whether segment names an owner or a repository.
func isRepoSegment(segment string) bool {
	return segment != "." && segment != ".." &&
		ownerRepoSegment.MatchString(segment)
}

// EscapedOwnerRepo returns path with each segment escaped for an API path, so a
// crafted url cannot address another endpoint.
func EscapedOwnerRepo(path string) string {
	owner, repo, _ := strings.Cut(path, "/")

	return escapeSegment(owner) + "/" + escapeSegment(repo)
}

// escapeSegment escapes one path segment.
func escapeSegment(segment string) string {
	if segment == "." || segment == ".." {
		return strings.ReplaceAll(segment, ".", "%2E")
	}

	return url.PathEscape(segment)
}

// maxErrorBody bounds the host-supplied body quoted in an error.
const maxErrorBody = 256

// BodyExcerpt returns a short single-line excerpt of a host-supplied body.
func BodyExcerpt(body []byte) string {
	if len(body) > maxErrorBody*4 {
		body = body[:maxErrorBody*4]
	}

	printable := strings.Map(func(r rune) rune {
		switch {
		case unicode.IsSpace(r):
			return ' '
		case !unicode.IsPrint(r):
			return -1
		}
		return r
	}, string(body))

	excerpt := strings.Join(strings.Fields(printable), " ")
	if excerpt == "" {
		return ""
	}
	if len(excerpt) > maxErrorBody {
		excerpt = strings.ToValidUTF8(excerpt[:maxErrorBody], "") + "..."
	}

	return "\n" + excerpt
}

// MarkPreReleaseTags flags every release whose tag name carries a semantic-version
// pre-release label.
//
// A tag that is not semver-shaped ("nightly", "1.2.3rc1") cannot be recognised and
// stays a stable release.
func MarkPreReleaseTags(releases []forgetypes.Release) {
	for i := range releases {
		tag := util.FirstNonDefault(releases[i].TagName, releases[i].Name)
		if semVer, err := semver.NewVersion(tag); err == nil && semVer.Prerelease() != "" {
			releases[i].PreRelease = true
		}
	}
}

// FilterOptions are the settings [FilterReleases] filters against.
type FilterOptions struct {
	URLCommands        filter.URLCommands
	SemanticVersioning bool
	UsePreReleases     bool
}

// FilterReleases filters releases based on the following:
//   - opts.URLCommands.
//   - Non-semantic versions (if opts.SemanticVersioning).
//   - Pre-releases (if not opts.UsePreReleases).
//
// Returns the filtered list, sorted in descending order (if opts.SemanticVersioning).
func FilterReleases(
	releases []forgetypes.Release,
	opts FilterOptions,
	logFrom logx.LogFrom,
) []forgetypes.Release {
	filteredReleases := make([]forgetypes.Release, 0, len(releases))

	for _, release := range releases {
		if release.PreRelease && !opts.UsePreReleases {
			continue
		}

		tag := util.FirstNonDefault(release.TagName, release.Name)
		tagName, err := opts.URLCommands.Run(tag, logFrom)
		if err != nil || len(tagName) == 0 {
			continue
		}

		release.TagName = tagName[0]

		if opts.SemanticVersioning {
			semVer, err := semver.NewVersion(tagName[0])
			if err != nil {
				continue
			}
			release.SemanticVersion = semVer
		}

		filteredReleases = append(filteredReleases, release)
	}

	if opts.SemanticVersioning {
		sort.Slice(filteredReleases, func(i, j int) bool {
			return releaseSortsBefore(filteredReleases[i], filteredReleases[j])
		})
	}

	return filteredReleases
}

// releaseSortsBefore reports whether release `a` should sort ahead of release `b` when ordering
// releases from newest to oldest.
//
// Both releases must have a parsed SemanticVersion.
func releaseSortsBefore(a, b forgetypes.Release) bool {
	return a.SemanticVersion.GreaterThan(b.SemanticVersion)
}
