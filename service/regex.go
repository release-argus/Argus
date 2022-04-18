// Copyright [2022] [Hymenaios]
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

package service

import (
	"fmt"

	"github.com/hymenaios-io/Hymenaios/utils"
)

func (s *Service) regexCheckVersion(
	version string,
	logFrom utils.LogFrom,
) error {
	// Check that the version grabbed satisfies the specified regex (if there is any).
	wantedRegexVersion := s.GetRegexVersion()
	if s.RegexVersion != nil {
		regexMatch := utils.RegexCheck(*wantedRegexVersion, version)
		if !regexMatch {
			err := fmt.Errorf("regex not matched on version %q", version)
			s.Status.RegexMissesVersion++
			jLog.Info(err, logFrom, s.Status.RegexMissesVersion == 1)
			return err
		}
	}

	return nil
}

func (s *Service) regexCheckContent(
	version string,
	body interface{},
	logFrom utils.LogFrom,
) error {
	// Check for a regex match in the body if one is desired.
	wantedRegexContent := s.GetRegexContent()
	if wantedRegexContent != nil {
		// Create a list to search as `github` service types we'll only
		// search asset `name` and `browser_download_url`
		var searchArea []string
		switch v := body.(type) {
		case string:
			searchArea[0] = body.(string)
		case []GitHubAsset:
			for i := range body.([]GitHubAsset) {
				searchArea = append(searchArea,
					body.([]GitHubAsset)[i].Name,
					body.([]GitHubAsset)[i].BrowserDownloadURL,
				)
			}
		default:
			_ = v
		}

		for i := range searchArea {
			regexMatch := utils.RegexCheckWithParams(*wantedRegexContent, searchArea[i], version)
			jLog.Debug(
				fmt.Sprintf("%q RegexContent on %q, match=%t", *wantedRegexContent, searchArea[i], regexMatch),
				logFrom,
				true)
			if !regexMatch {
				if i == len(searchArea)-1 {
					err := fmt.Errorf(
						"regex %q not matched on content for version %q",
						utils.TemplateString(*wantedRegexContent, utils.ServiceInfo{LatestVersion: version}),
						version,
					)
					s.Status.RegexMissesContent++
					jLog.Info(err, logFrom, s.Status.RegexMissesContent == 1)
					return err
				}
				continue
			}
			break
		}
	}

	return nil
}
