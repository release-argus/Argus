// Copyright [2023] [Argus]
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

package latestver

import (
	"fmt"

	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

// applyOverrides to the Lookup and return that new Lookup.
func (l *Lookup) applyOverrides(
	accessToken *string,
	allowInvalidCerts *string,
	require *string,
	semanticVersioning *string,
	typeStr *string,
	url *string,
	urlCommands *string,
	usePreRelease *string,
	serviceID *string,
	logFrom *util.LogFrom,
) (*Lookup, error) {
	// Use the provided overrides, or the defaults.
	// access_token
	useAccessToken := util.FirstNonNilPtr(accessToken, l.AccessToken)
	// allow_invalid_certs
	useAllowInvalidCerts := l.AllowInvalidCerts
	if allowInvalidCerts != nil {
		useAllowInvalidCerts = util.StringToBoolPtr(*allowInvalidCerts)
	}
	// require
	useRequire, errRequire := filter.RequireFromStr(
		require,
		l.Require,
		logFrom)
	if errRequire != nil {
		//nolint:wrapcheck // Don't wrap error.
		return nil, errRequire
	}
	// semantic_versioning
	useSemanticVersioning := l.Options.SemanticVersioning
	if semanticVersioning != nil {
		useSemanticVersioning = util.StringToBoolPtr(*semanticVersioning)
	}
	// type
	useType := util.ValueOrDefault(typeStr, l.Type)
	if useType == "" {
		useType = "github"
	}
	// url
	useURL := util.ValueOrDefault(url, l.URL)
	// url_commands
	useURLCommands, errURLCommands := filter.URLCommandsFromStr(
		urlCommands,
		&l.URLCommands,
		logFrom)
	if errURLCommands != nil {
		//nolint:wrapcheck // Don't wrap error.
		return nil, errURLCommands
	}
	// use_pre_release
	useUsePreRelease := l.UsePreRelease
	if usePreRelease != nil {
		useUsePreRelease = util.StringToBoolPtr(*usePreRelease)
	}

	// Create a new lookup with the overrides.
	lookup := New(
		useAccessToken,
		useAllowInvalidCerts,
		nil, // GitHubData
		opt.New(
			nil, "", useSemanticVersioning,
			nil, nil),
		useRequire,
		nil, // Status
		useType,
		useURL,
		useURLCommands,
		useUsePreRelease,
		l.Defaults,
		l.HardDefaults)
	lookup.Status = &svcstatus.Status{
		ServiceID: serviceID}
	lookup.Options.Defaults = l.Options.Defaults
	lookup.Options.HardDefaults = l.Options.HardDefaults
	lookup.Status.Init(
		0, 0, 0,
		serviceID,
		nil)
	lookup.Status.SetLatestVersion(l.Status.LatestVersion(), false)

	if lookup.Type == "github" {
		// Use the current ETag/releases
		// (if ETag is the same, won't count towards API limit)
		if l.Type == "github" && l.GitHubData != nil {
			releases := l.GitHubData.Releases()
			lookup.GitHubData = NewGitHubData(
				l.GitHubData.ETag(),
				&releases)

			// Type changed to github (or new service)
		} else {
			lookup.GitHubData = &GitHubData{}
		}
	}

	if err := lookup.CheckValues(""); err != nil {
		jLog.Error(err, *logFrom, true)
		return nil, fmt.Errorf("values failed validity check:\n%w", err)
	}

	return lookup, nil
}

// Refresh queries the Service source with the provided overrides and returns:
//
// `version` - found from this query
//
// `annoounceUpdate` - Whether that version is new and should be announced (if no overrides provided),
//
// `err` - any errs encountered
func (l *Lookup) Refresh(
	accessToken *string,
	allowInvalidCerts *string,
	require *string,
	semanticVersioning *string,
	typeStr *string,
	url *string,
	urlCommands *string,
	usePreRelease *string,
) (version string, announceUpdate bool, err error) {
	serviceID := *l.Status.ServiceID
	logFrom := util.LogFrom{Primary: "latest_version/refresh", Secondary: serviceID}

	var lookup *Lookup
	lookup, err = l.applyOverrides(
		accessToken,
		allowInvalidCerts,
		require,
		semanticVersioning,
		typeStr,
		url,
		urlCommands,
		usePreRelease,
		&serviceID,
		&logFrom)
	if err != nil {
		return
	}

	// Log the lookup being used if debug.
	if jLog.IsLevel("DEBUG") {
		jLog.Debug(
			fmt.Sprintf("Refreshing with:\n%v", lookup),
			logFrom, true)
	}

	// Whether overrides were provided or not, we can update the status if not.
	overrides := require != nil ||
		semanticVersioning != nil ||
		url != nil ||
		urlCommands != nil ||
		usePreRelease != nil

	// Query the lookup.
	_, err = lookup.Query(!overrides, &logFrom)
	if err != nil {
		return
	}
	version = lookup.Status.LatestVersion()
	announceUpdate = l.updateFromRefresh(lookup, overrides)
	return
}

// updateFromRefresh updates the current Lookup with the values from a Query on this
// new Lookup if the values should retrieve the same data.
//
// `changingOverrides` - whether the overrides provided to the Refresh method would change the Query.
//
// Returns whether a new version was found and should be announced.
func (l *Lookup) updateFromRefresh(newLookup *Lookup, changingOverrides bool) (announceUpdate bool) {
	// Querying the same GitHub repo and the ETag has changed
	if l.Type == "github" && newLookup.Type == "github" &&
		l.URL == newLookup.URL &&
		l.GitHubData != nil &&
		l.GitHubData.ETag() != newLookup.GitHubData.ETag() {
		// Update the ETag and releases
		l.GitHubData.Set(newLookup.GitHubData.ETag(), newLookup.GitHubData.Releases())
	}

	// Copy the Docker queryToken if it's what the current would fetch
	if l.Require != nil && l.Require.Docker != nil &&
		newLookup.Require != nil && newLookup.Require.Docker != nil &&
		l.Require.Docker.Type == newLookup.Require.Docker.Type &&
		l.Require.Docker.Token == newLookup.Require.Docker.Token &&
		l.Require.Docker.Username == newLookup.Require.Docker.Username &&
		l.Require.Docker.Image == newLookup.Require.Docker.Image {
		queryToken, validUntil := newLookup.Require.Docker.CopyQueryToken()
		l.Require.Docker.SetQueryToken(
			&newLookup.Require.Docker.Token,
			&queryToken, &validUntil)
	}

	// If overrides that may change a successful query were provided
	if changingOverrides {
		return
	}

	// Update the last queried time.
	l.Status.SetLastQueried(newLookup.Status.LastQueried())
	// Update the latest version if it has changed.
	newLatestVersion := newLookup.Status.LatestVersion()
	if newLatestVersion != l.Status.LatestVersion() {
		announceUpdate = true
		l.Status.SetLatestVersion(newLatestVersion, true)
	}
	return
}
