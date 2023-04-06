// Copyright [2022] [Argus]
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
	useAccessToken := util.GetFirstNonNilPtr(accessToken, l.AccessToken)
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
	useType := util.GetValue(typeStr, l.Type)
	if useType == "" {
		useType = "github"
	}
	// url
	useURL := util.GetValue(url, l.URL)
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
	lookup := Lookup{
		Type:              useType,
		URL:               useURL,
		AccessToken:       useAccessToken,
		AllowInvalidCerts: useAllowInvalidCerts,
		UsePreRelease:     useUsePreRelease,
		URLCommands:       *useURLCommands,
		Require:           useRequire,
		Options: &opt.Options{
			SemanticVersioning: useSemanticVersioning,
			Defaults:           l.Options.Defaults,
			HardDefaults:       l.Options.HardDefaults,
		},
		Status: &svcstatus.Status{
			ServiceID: serviceID,
		},
		Defaults:     l.Defaults,
		HardDefaults: l.HardDefaults,
	}
	lookup.Status.Init(
		0, 0, 0,
		serviceID,
		nil,
	)
	lookup.Status.SetLatestVersion(l.Status.GetLatestVersion(), false)

	if lookup.Type == "github" {
		// Use the current ETag/releases
		// (if ETag is the same, won't count towards API limit)
		if l.Type == "github" {
			lookup.GitHubData = &GitHubData{
				ETag: l.GitHubData.ETag,
			}
			lookup.GitHubData.ETag = l.GitHubData.ETag
			lookup.GitHubData.Releases = l.GitHubData.Releases

			// Type changed to github (or new service)
		} else {
			lookup.GitHubData = &GitHubData{}
		}
	}

	// require
	if lookup.Require != nil {
		lookup.Require.Status = lookup.Status
	}

	if err := lookup.CheckValues(""); err != nil {
		jLog.Error(err, *logFrom, true)
		return nil, fmt.Errorf("values failed validity check:\n%w", err)
	}

	return &lookup, nil
}

// Refresh queries the Service source with the provided overrides,
// returning the `version` found from this query as well, as whether
// that new version should be announced (no overrides provided),
// and any errors encountered
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
		usePreRelease != nil ||
		url != nil ||
		urlCommands != nil

	// Query the lookup.
	_, err = lookup.Query(!overrides, &logFrom)
	if err != nil {
		return
	}
	version = lookup.Status.GetLatestVersion()

	// Querying the same GitHub repo
	if url == nil &&
		lookup.Type == "github" && l.Type == "github" &&
		lookup.GitHubData.ETag != l.GitHubData.ETag {
		// Update the ETag and releases
		l.GitHubData.ETag = lookup.GitHubData.ETag
		l.GitHubData.Releases = lookup.GitHubData.Releases
	}

	// If no overrides that may change a successful query were provided
	// then we can update the Status.
	if !overrides {
		// Update the last queried time.
		l.Status.SetLastQueried(lookup.Status.GetLastQueried())
		// Update the latest version if it has changed.
		mewLatestVersion := lookup.Status.GetLatestVersion()
		if mewLatestVersion != l.Status.GetLatestVersion() {
			announceUpdate = true
			l.Status.SetLatestVersion(mewLatestVersion, true)
		}
	}

	return
}
