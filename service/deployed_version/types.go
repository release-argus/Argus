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

package deployed_version

import (
	"github.com/release-argus/Argus/service/options"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
)

var (
	jLog *utils.JLog
)

// Lookup the deployed version of the service.
type Lookup struct {
	URL               string                 `yaml:"url,omitempty"`                 // URL to query.
	AllowInvalidCerts *bool                  `yaml:"allow_invalid_certs,omitempty"` // default - false = Disallows invalid HTTPS certificates
	BasicAuth         *BasicAuth             `yaml:"basic_auth,omitempty"`          // Basic Auth for the HTTP(S) request.
	Headers           []Header               `yaml:"headers,omitempty"`             // Headers for the HTTP(S) request.
	JSON              string                 `yaml:"json,omitempty"`                // JSON key to use e.g. version_current.
	Regex             string                 `yaml:"regex,omitempty"`               // Regex to get the DeployedVersion
	Options           *options.Options       `yaml:"-"`                             // Options for the lookups
	Status            *service_status.Status `yaml:"-"`                             // Service Status
	HardDefaults      *Lookup                `yaml:"-"`                             // Hardcoded default values.
	Defaults          *Lookup                `yaml:"-"`                             // Default values.
}

// BasicAuth to use on the HTTP(s) request.
type BasicAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Header to use in the HTTP request.
type Header struct {
	Key   string `yaml:"key"`   // Header key, e.g. X-Sig
	Value string `yaml:"value"` // Value to give the key
}
