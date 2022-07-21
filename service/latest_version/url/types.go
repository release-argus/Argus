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

package url

import (
	"github.com/release-argus/Argus/service/latest_version/require"
	"github.com/release-argus/Argus/service/options"
	service_status "github.com/release-argus/Argus/service/status"
	url_command "github.com/release-argus/Argus/service/url_commands"
	"github.com/release-argus/Argus/utils"
)

var (
	jLog *utils.JLog
)

type LatestVersion struct {
	serviceID         *string                `yaml:"-"`                             // Service's ID
	Type              *string                `yaml:"type,omitempty"`                // "github"/"URL"
	URL               string                 `yaml:"url,omitempty"`                 // type:URL - "https://example.com", type:github - "owner/repo" or "https://github.com/owner/repo"
	AllowInvalidCerts *bool                  `json:"allow_invalid_certs,omitempty"` // default - false = Disallows invalid HTTPS certificates
	URLCommands       *url_command.Slice     `yaml:"url_commands,omitempty"`        // Commands to filter the release from the URL request
	Require           *require.Options       `yaml:"require,omitempty"`             // Options to require before a release is considered valid
	Status            *service_status.Status `yaml:"-"`                             // Service Status
	Options           *options.Options       `yaml:"-"`                             // Options
	Defaults          *LatestVersion         `yaml:"-"`                             // Defaults
	HardDefaults      *LatestVersion         `yaml:"-"`                             // Hard Defaults
}
