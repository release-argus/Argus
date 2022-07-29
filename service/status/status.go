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

package service_status

import (
	"fmt"
	"time"

	db_types "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/utils"
)

// Status is the current state of the Service element (version and regex misses).
type Status struct {
	ApprovedVersion          string `yaml:"-"` // The version that's been approved
	DeployedVersion          string `yaml:"-"` // Track the deployed version of the service from the last successful WebHook.
	DeployedVersionTimestamp string `yaml:"-"` // UTC timestamp of DeployedVersion being changed.
	LatestVersion            string `yaml:"-"` // Latest version found from query().
	LatestVersionTimestamp   string `yaml:"-"` // UTC timestamp of LatestVersion being changed.
	LastQueried              string `yaml:"-"` // UTC timestamp that version was last queried/checked.
	RegexMissesContent       uint   `yaml:"-"` // Counter for the number of regex misses on URL content.
	RegexMissesVersion       uint   `yaml:"-"` // Counter for the number of regex misses on version.
	Fails                    Fails  `yaml:"-"` // Track the Notify/WebHook fails

	// Announces
	AnnounceChannel *chan []byte           `yaml:"-"` // Announce to the WebSocket
	DatabaseChannel *chan db_types.Message `yaml:"-"` // Channel for broadcasts to the Database
	SaveChannel     *chan bool             `yaml:"-"` // Channel for triggering a save of the config
	ServiceID       *string                `yaml:"-"` // ID of the Service
	WebURL          *string                `yaml:"-"` // Web URL of the Service
}

// TODO: Deprecate
// OldStatus is for handling config.yml's containing data that now belongs in argus.db
type OldStatus struct {
	ApprovedVersion          string `yaml:"approved_version,omitempty"`           // The version that's been approved
	DeployedVersion          string `yaml:"deployed_version,omitempty"`           // Track the deployed version of the service from the last successful WebHook.
	DeployedVersionTimestamp string `yaml:"deployed_version_timestamp,omitempty"` // UTC timestamp of DeployedVersion being changed.
	LatestVersion            string `yaml:"latest_version,omitempty"`             // Latest version found from query().
	LatestVersionTimestamp   string `yaml:"latest_version_timestamp,omitempty"`   // UTC timestamp of LatestVersion being changed.
}

// Fails keeps track of whether any of the notifications failed on the last version change.
type Fails struct {
	Shoutrrr map[string]*bool `yaml:"-"` // Shoutrrr unsent/fail/pass.
	WebHook  map[string]*bool `yaml:"-"` // WebHook unsent/fail/pass.
	Command  []*bool          `yaml:"-"` // Command unsent/fail/pass.
}

// Init initialises the Status vars when more than the default value is needed.
func (s *Status) Init(
	shoutrrrs int,
	commands int,
	webhooks int,
	serviceID *string,
	webURL *string,
) {
	s.Fails.Shoutrrr = make(map[string]*bool, shoutrrrs)
	s.Fails.Command = make([]*bool, commands)
	s.Fails.WebHook = make(map[string]*bool, webhooks)

	s.ServiceID = serviceID
	s.WebURL = webURL
}

// SetLastQueried will update LastQueried to now.
func (s *Status) SetLastQueried() {
	s.LastQueried = time.Now().UTC().Format(time.RFC3339)
}

// SetDeployedVersion will set DeployedVersion as well as DeployedVersionTimestamp.
func (s *Status) SetDeployedVersion(version string) {
	s.DeployedVersion = version
	s.DeployedVersionTimestamp = time.Now().UTC().Format(time.RFC3339)
	// Ignore ApprovedVersion if we're on it
	if version == s.ApprovedVersion {
		s.ApprovedVersion = ""
	}

	// Clear the fail status of WebHooks/Commands
	s.Fails.resetFails()

	*s.DatabaseChannel <- db_types.Message{
		ServiceID: *s.ServiceID,
		Cells: []db_types.Cell{
			{Column: "deployed_version", Value: s.DeployedVersion},
			{Column: "deployed_version_timestamp", Value: s.DeployedVersionTimestamp},
		},
	}
}

// SetLatestVersion will set LatestVersion to `version` and LatestVersionTimestamp to s.LastQueried.
func (s *Status) SetLatestVersion(version string) {
	s.LatestVersion = version
	s.LatestVersionTimestamp = s.LastQueried

	// Clear the fail status of WebHooks/Commands
	s.Fails.resetFails()

	*s.DatabaseChannel <- db_types.Message{
		ServiceID: *s.ServiceID,
		Cells: []db_types.Cell{
			{Column: "latest_version", Value: s.LatestVersion},
			{Column: "latest_version_timestamp", Value: s.LatestVersionTimestamp},
		},
	}
}

// ResetFails of the Status.Fails
func (f *Fails) resetFails() {
	for i := range f.Shoutrrr {
		f.Shoutrrr[i] = nil
	}
	for i := range f.Command {
		f.Command[i] = nil
	}
	for i := range f.WebHook {
		f.WebHook[i] = nil
	}
}

// GetWebURL returns the Web URL.
func (s *Status) GetWebURL() string {
	if utils.DefaultIfNil(s.WebURL) == "" {
		return ""
	}

	return utils.TemplateString(*s.WebURL, utils.ServiceInfo{LatestVersion: s.LatestVersion})
}

// Print will print the Status.
func (s *Status) Print(prefix string) {
	utils.PrintlnIfNotDefault(s.ApprovedVersion, fmt.Sprintf("%sapproved_version: %s", prefix, s.ApprovedVersion))
	utils.PrintlnIfNotDefault(s.DeployedVersion, fmt.Sprintf("%sdeployed_version: %s", prefix, s.DeployedVersion))
	utils.PrintlnIfNotDefault(s.DeployedVersionTimestamp, fmt.Sprintf("%sdeployed_version_timestamp: %q", prefix, s.DeployedVersionTimestamp))
	utils.PrintlnIfNotDefault(s.LatestVersion, fmt.Sprintf("%slatest_version: %s", prefix, s.LatestVersion))
	utils.PrintlnIfNotDefault(s.LatestVersionTimestamp, fmt.Sprintf("%slatest_version_timestamp: %q", prefix, s.LatestVersionTimestamp))
}
