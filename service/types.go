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

package service

import (
	command "github.com/release-argus/Argus/commands"
	db_types "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/webhook"
)

var (
	jLog *utils.JLog
)

// Slice is a slice mapping of Service.
type Slice map[string]*Service

// Service is a source to be serviceed and provides everything needed to extract
// the latest version from the URL provided.
type Service struct {
	ID                    string                 `yaml:"-"`                          // service_name
	Type                  string                 `yaml:"-"`                          // service_name
	Comment               *string                `yaml:"comment,omitempty"`          // Comment on the Service
	Options               Options                `yaml:"options,omitempty"`          // Options to give the Service
	LatestVersion         LatestVersion          `yaml:"latest_version,omitempty"`   // Vars to getting the latest version of the Service
	CommandController     *command.Controller    `yaml:"-"`                          // The controller for the OS Commands that tracks fails and has the announce channel
	Command               *command.Slice         `yaml:"command,omitempty"`          // OS Commands to run on new release
	WebHook               *webhook.Slice         `yaml:"webhook,omitempty"`          // Service-specific WebHook vars
	Notify                *shoutrrr.Slice        `yaml:"notify,omitempty"`           // Service-specific Shoutrrr vars
	DeployedVersionLookup *DeployedVersionLookup `yaml:"deployed_version,omitempty"` // Var to scrape the Service's current deployed version
	Dashboard             DashboardOptions       `yaml:"dashboard,omitempty"`        // Options for the dashboard
	Status                *service_status.Status `yaml:"-"`                          // Track the Status of this source (version and regex misses)
	HardDefaults          *Service               `yaml:"-"`                          // Hardcoded default values
	Defaults              *Service               `yaml:"-"`                          // Default values
	Announce              *chan []byte           `yaml:"-"`                          // Announce to the WebSocket
	DatabaseChannel       *chan db_types.Message `yaml:"-"`                          // Channel for broadcasts to the Database
	SaveChannel           *chan bool             `yaml:"-"`                          // Channel for triggering a save of the config

	// TODO: Deprecate
	OldStatus *service_status.OldStatus `yaml:"status,omitempty"` // For moving version info to argus.db
}

type LatestVersion interface {
	GetAccessToken() string // GitHub
	GetSelfAllowInvalidCerts() bool
	GetAllowInvalidCerts() bool
	GetFriendlyURL() string
	GetLookupURL() string
	GetVersion([]byte, utils.LogFrom) (string, error)
	GetType() *string
	GetURL() string
	GetRequire() LatestVersionRequireOptions
	GetURLCommands() *URLCommandSlice

	URLCommandsCheckValues(string) error
	Print(string)
}

// DeployedVersionLookup of the service.
type DeployedVersionLookup struct {
	URL               string                 `yaml:"url,omitempty"`                 // URL to query.
	AllowInvalidCerts *bool                  `json:"allow_invalid_certs,omitempty"` // default - false = Disallows invalid HTTPS certificates
	BasicAuth         *BasicAuth             `yaml:"basic_auth,omitempty"`          // Basic Auth for the HTTP(S) request.
	Headers           []Header               `yaml:"headers,omitempty"`             // Headers for the HTTP(S) request.
	JSON              string                 `yaml:"json,omitempty"`                // JSON key to use e.g. version_current.
	Regex             string                 `yaml:"regex,omitempty"`               // Regex to get the DeployedVersion
	HardDefaults      *DeployedVersionLookup `yaml:"-"`                             // Hardcoded default values.
	Defaults          *DeployedVersionLookup `yaml:"-"`                             // Default values.
	options           *Options               `yaml:"-"`                             // Options for the lookups
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
