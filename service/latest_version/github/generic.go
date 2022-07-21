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

package github

import (
	"fmt"
	"strings"

	url_command "github.com/release-argus/Argus/service/url_commands"
	"github.com/release-argus/Argus/utils"
	api_types "github.com/release-argus/Argus/web/api/types"
)

func (l LatestVersion) CheckValues(prefix string) (errs error) {
	if l.URL == "" {
		errs = fmt.Errorf("%s%srepo: <required> e.g. 'release-argus/Argus'",
			utils.ErrorToString(errs), prefix)
	}
	if strings.Count(l.URL, "/") > 1 {
		parts := strings.Split(l.URL, "/")
		l.URL = strings.Join(parts[len(parts)-2:], "/")
	}

	if requireErrs := l.Options.CheckValues(prefix + "  "); requireErrs != nil {
		errs = fmt.Errorf("%s  require:\\%w",
			prefix, requireErrs)
	}
	if urlCommandErrs := l.URLCommands.CheckValues(prefix + "  "); urlCommandErrs != nil {
		errs = fmt.Errorf("%s  url_commands:\\%w",
			prefix, urlCommandErrs)
	}

	if errs != nil {
		errs = fmt.Errorf("%slatest_version:\\%w",
			prefix, errs)
	}

	return
}

func (l *LatestVersion) GetAccessToken() string {
	return utils.GetFirstNonDefault(l.AccessToken, l.Defaults.AccessToken, l.HardDefaults.AccessToken)
}

func (l *LatestVersion) GetAllowInvalidCerts() *bool {
	return utils.GetFirstNonNilPtr(l.AllowInvalidCerts, l.Defaults.AllowInvalidCerts, l.HardDefaults.AllowInvalidCerts)
}

func (l *LatestVersion) GetURL() *string {
	return &l.URL
}

func (l LatestVersion) GetType() string {
	return "github"
}

func (l LatestVersion) GetURLCommands() *url_command.Slice {
	return l.URLCommands
}

func (l LatestVersion) ConvertToAPIType() *api_types.LatestVersion {
	return &api_types.LatestVersion{
		URL:               l.URL,
		AccessToken:       utils.ValueIfNotDefault(l.AccessToken, "<secret>"),
		AllowInvalidCerts: l.AllowInvalidCerts,
		UsePreRelease:     l.UsePreRelease,
		Require: &api_types.LatestVersionRequire{
			RegexContent: l.Require.RegexContent,
			RegexVersion: l.Require.RegexVersion,
		},
	}
}
