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

package filters

import (
	"fmt"
	"regexp"

	github_types "github.com/release-argus/Argus/service/latest_version/api_types"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
)

// Require for version to be considered valid.
type Require struct {
	Status       *service_status.Status `yaml:"-"`                       // Service Status
	RegexContent string                 `yaml:"regex_content,omitempty"` // "abc-[a-z]+-{{ version }}_amd64.deb" This regex must exist in the body of the URL to trigger new version actions
	RegexVersion string                 `yaml:"regex_version,omitempty"` // "v*[0-9.]+" The version found must match this release to trigger new version actions
}

// RegexCheckVersion
func (r *Require) RegexCheckVersion(
	version string,
	log *utils.JLog,
	logFrom utils.LogFrom,
) error {
	if r == nil {
		return nil
	}

	// Check that the version grabbed satisfies the specified regex (if there is any).
	if r.RegexVersion == "" {
		return nil
	}
	regexMatch := utils.RegexCheck(r.RegexVersion, version)
	if !regexMatch {
		err := fmt.Errorf("regex not matched on version %q",
			version)
		r.Status.RegexMissesVersion++
		log.Info(err, logFrom, r.Status.RegexMissesVersion == 1)
		return err
	}

	return nil
}

// RegexCheckContent of body with version
func (r *Require) RegexCheckContent(
	version string,
	body interface{},
	log *utils.JLog,
	logFrom utils.LogFrom,
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
		regexMatch := utils.RegexCheckWithParams(r.RegexContent, searchArea[i], version)
		log.Debug(
			fmt.Sprintf("%q RegexContent on %q, match=%t", r.RegexContent, searchArea[i], regexMatch),
			logFrom,
			true)
		if !regexMatch {
			// if we're on the last asset
			if i == len(searchArea)-1 {
				err := fmt.Errorf(
					"regex %q not matched on content for version %q",
					utils.TemplateString(r.RegexContent, utils.ServiceInfo{LatestVersion: version}),
					version,
				)
				r.Status.RegexMissesContent++
				log.Info(err, logFrom, r.Status.RegexMissesContent == 1)
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

// Print will print the Require.
func (r *Require) Print(prefix string) {
	if r == nil {
		return
	}
	utils.PrintlnIfNotDefault(r.RegexContent, fmt.Sprintf("%s  regex_content: %q", prefix, r.RegexContent))
	utils.PrintlnIfNotDefault(r.RegexVersion, fmt.Sprintf("%s  regex_version: %q", prefix, r.RegexVersion))
}

// CheckValues of the Require options.
func (r *Require) CheckValues(prefix string) (errs error) {
	if r == nil {
		return
	}

	// Content RegEx
	if r.RegexContent != "" {
		_, err := regexp.Compile(r.RegexContent)
		if err != nil {
			errs = fmt.Errorf("%s%s  regex_content: %q <invalid> (Invalid RegEx)\\",
				utils.ErrorToString(errs), prefix, r.RegexContent)
		}
	}

	// Version RegEx
	if r.RegexVersion != "" {
		_, err := regexp.Compile(r.RegexVersion)
		if err != nil {
			errs = fmt.Errorf("%s%s  regex_version: %q <invalid> (Invalid RegEx)\\",
				utils.ErrorToString(errs), prefix, r.RegexVersion)
		}
	}

	return
}
