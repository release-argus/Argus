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

package filter

import (
	"fmt"

	github_types "github.com/release-argus/Argus/service/latest_version/api_type"
	"github.com/release-argus/Argus/util"
)

// RegexCheckVersion
func (r *Require) RegexCheckVersion(
	version string,
	logFrom *util.LogFrom,
) error {
	if r == nil {
		return nil
	}

	// Check that the version grabbed satisfies the specified regex (if there is any).
	if r.RegexVersion == "" {
		return nil
	}
	regexMatch := util.RegexCheck(r.RegexVersion, version)
	if !regexMatch {
		err := fmt.Errorf("regex not matched on version %q",
			version)
		r.Status.RegexMissesVersion++
		jLog.Info(err, *logFrom, r.Status.RegexMissesVersion == 1)
		return err
	}

	return nil
}

// RegexCheckContent of body with version
func (r *Require) RegexCheckContent(
	version string,
	body interface{},
	logFrom *util.LogFrom,
) error {
	if r == nil {
		return nil
	}

	// Check for a regex match in the body if one is desired.
	if r.RegexContent == "" {
		return nil
	}
	// Create a list to search as `github` service types we'll only
	// search asset `name` and `browser_download_url`
	var searchArea []string
	switch v := body.(type) {
	case string:
		searchArea = []string{body.(string)}
	case []github_types.Asset:
		for i := range body.([]github_types.Asset) {
			searchArea = append(searchArea,
				body.([]github_types.Asset)[i].Name,
				body.([]github_types.Asset)[i].BrowserDownloadURL,
			)
		}
	default:
		return fmt.Errorf("invalid body type %T",
			v)
	}

	for i := range searchArea {
		regexMatch := util.RegexCheckWithParams(r.RegexContent, searchArea[i], version)
		if jLog.IsLevel("DEBUG") {
			jLog.Debug(
				fmt.Sprintf("%q RegexContent on %q, match=%t", r.RegexContent, searchArea[i], regexMatch),
				*logFrom,
				true)
		}
		if !regexMatch {
			// if we're on the last asset
			if i == len(searchArea)-1 {
				err := fmt.Errorf(
					"regex %q not matched on content for version %q",
					util.TemplateString(r.RegexContent, util.ServiceInfo{LatestVersion: version}),
					version,
				)
				r.Status.RegexMissesContent++
				jLog.Info(err, *logFrom, r.Status.RegexMissesContent == 1)
				return err
			}
			// continue searching the other assets
			continue
		}
		// regex matched
		break
	}

	return nil
}
