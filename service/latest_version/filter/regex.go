// Copyright [2024] [Argus]
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

// Package filter provides filtering for latest_version queries.
package filter

import (
	"fmt"
	"strings"

	github_types "github.com/release-argus/Argus/service/latest_version/types/github/api_type"
	"github.com/release-argus/Argus/util"
)

// RegexCheckVersion returns whether `version` matches the regex.
func (r *Require) RegexCheckVersion(
	version string,
	logFrom util.LogFrom,
) error {
	if r == nil {
		return nil
	}

	// Check the version grabbed satisfies the specified regex (if there is any).
	if r.RegexVersion == "" {
		return nil
	}
	regexMatch := util.RegexCheck(r.RegexVersion, version)
	if !regexMatch {
		err := fmt.Errorf("regex %q not matched on version %q",
			r.RegexVersion, version)
		r.Status.RegexMissVersion()
		jLog.Info(err, logFrom, r.Status.RegexMissesVersion() == 1)
		return err
	}

	return nil
}

func (r *Require) regexCheckString(
	version string,
	logFrom util.LogFrom,
	searchArea ...string,
) bool {
	for _, text := range searchArea {
		regexMatch := util.RegexCheckWithVersion(r.RegexContent, text, version)
		if jLog.IsLevel("DEBUG") {
			jLog.Debug(
				fmt.Sprintf("%q RegexContent on %q, match=%t",
					r.RegexContent, text, regexMatch),
				logFrom, true)
		}
		if regexMatch {
			return true
		}
	}
	return false
}

// regexCheckContentFail simply returns an error for the RegexCheckContent* functions.
func (r *Require) regexCheckContentFail(version string, logFrom util.LogFrom) error {
	// Escape all dots in the version.
	regexStr := util.TemplateString(r.RegexContent,
		util.ServiceInfo{
			LatestVersion: strings.ReplaceAll(version, ".", `\.`)})
	r.Status.RegexMissContent()
	err := fmt.Errorf(
		"regex %q not matched on content for version %q",
		regexStr, version)
	jLog.Info(err, logFrom, r.Status.RegexMissesContent() == 1)
	return err
}

// RegexCheckContent of body with version.
func (r *Require) RegexCheckContent(
	version string,
	body string,
	logFrom util.LogFrom,
) error {
	if r == nil || r.RegexContent == "" {
		return nil
	}

	// Create a list to search as `github` service types we'll only
	// search asset `name`, and `browser_download_url`.
	if match := r.regexCheckString(version, logFrom, body); match {
		return nil
	}

	return r.regexCheckContentFail(version, logFrom)
}

// RegexCheckContentGitHub checks the content of the GitHub release assets.
// for a RegexContent match.
//
// Returns the date of release.
func (r *Require) RegexCheckContentGitHub(
	version string,
	assets []github_types.Asset,
	logFrom util.LogFrom,
) (string, error) {
	if r == nil || r.RegexContent == "" {
		return "", nil
	}

	for _, asset := range assets {
		match := r.regexCheckString(version, logFrom,
			asset.Name, asset.BrowserDownloadURL)
		if match {
			// Copy asset date as release date.
			return asset.CreatedAt, nil
		}
	}

	return "", r.regexCheckContentFail(version, logFrom)
}
