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
	github_types "github.com/release-argus/Argus/service/latest_version/api_type"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

var (
	jLog *util.JLog
)

type Lookup struct {
	Type              string                 `yaml:"type,omitempty"`                // "github"/"URL"
	URL               string                 `yaml:"url,omitempty"`                 // type:URL - "https://example.com", type:github - "owner/repo" or "https://github.com/owner/repo".
	AccessToken       *string                `yaml:"access_token,omitempty"`        // GitHub access token to use
	AllowInvalidCerts *bool                  `yaml:"allow_invalid_certs,omitempty"` // default - false = Disallows invalid HTTPS certificates
	UsePreRelease     *bool                  `yaml:"use_prerelease,omitempty"`      // Whether the prerelease tag should be used (prereleases are ignored by default)
	URLCommands       filter.URLCommandSlice `yaml:"url_commands,omitempty"`        // Commands to filter the release from the URL request
	Require           *filter.Require        `yaml:"require,omitempty"`             // Options to require before a release is considered valid
	Options           *opt.Options           `yaml:"-"`                             // Options
	GitHubData        *GitHubData            `yaml:"-"`                             // GitHub Conditional Request vars
	Status            *svcstatus.Status      `yaml:"-"`                             // Service Status
	Defaults          *Lookup                `yaml:"-"`                             // Defaults
	HardDefaults      *Lookup                `yaml:"-"`                             // Hard Defaults
}

// GitHubData is data needed in GitHub requests
type GitHubData struct {
	ETag     string                 // GitHub ETag for conditional requests https://docs.github.com/en/rest/overview/resources-in-the-rest-api#conditional-requestsl
	Releases []github_types.Release // Track the ETag releases until they're usable
}
