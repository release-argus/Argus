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
	"fmt"
	"time"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/webhook"
)

// Status is the current state of the Service element (version and regex misses).
type Status struct {
	ApprovedVersion         string       `yaml:"approved_version,omitempty"`          // The version that's been approved
	CurrentVersion          string       `yaml:"current_version,omitempty"`           // Track the current version of the service from the last successful WebHook.
	CurrentVersionTimestamp string       `yaml:"current_version_timestamp,omitempty"` // UTC timestamp of CurrentVersion being changed.
	LatestVersion           string       `yaml:"latest_version,omitempty"`            // Latest version found from query().
	LatestVersionTimestamp  string       `yaml:"latest_version_timestamp,omitempty"`  // UTC timestamp of LatestVersion being changed.
	LastQueried             string       `yaml:"-"`                                   // UTC timestamp that version was last queried/checked.
	RegexMissesContent      uint         `yaml:"-"`                                   // Counter for the number of regex misses on URL content.
	RegexMissesVersion      uint         `yaml:"-"`                                   // Counter for the number of regex misses on version.
	Fails                   *StatusFails `yaml:"-"`                                   // Track the Shoutrrr/WebHook fails
}

// StatusFails keeps track of whether any of the notifications failed on the last version change.
type StatusFails struct {
	Shoutrrr *[]bool `yaml:"-"` // Track whether any of the Slice failed.
	WebHook  *[]bool `yaml:"-"` // Track whether any of the WebHookSlice failed.
}

// Init initialises the Status vars when more than the default value is needed.
func (s *Status) Init(
	shoutrrrs *shoutrrr.Slice,
	webhooks *webhook.Slice,
) {
	s.Fails = new(StatusFails)
	if shoutrrrs != nil {
		shoutrrrFails := make([]bool, len(*shoutrrrs))
		s.Fails.Shoutrrr = &shoutrrrFails
	}
	if webhooks != nil {
		webhookFails := make([]bool, len(*webhooks))
		s.Fails.WebHook = &webhookFails
	}
}

// SetCurrentVersion will set CurrentVersion as well as CurrentVersionTimestamp.
func (s *Status) SetCurrentVersion(version string) {
	s.CurrentVersion = version
	s.CurrentVersionTimestamp = time.Now().UTC().Format(time.RFC3339)
}

// SetLastQueried will update LastQueried to now.
func (s *Status) SetLastQueried() {
	s.LastQueried = time.Now().UTC().Format(time.RFC3339)
}

// SetLatestVersion will set LatestVersion to `version` and LatestVersionTimestamp to s.LastQueried.
func (s *Status) SetLatestVersion(version string) {
	s.LatestVersion = version
	s.LatestVersionTimestamp = s.LastQueried
}

// Print will print the Status.
func (s *Status) Print(prefix string) {
	utils.PrintlnIfNotDefault(s.ApprovedVersion, fmt.Sprintf("%sapproved_version: %s", prefix, s.ApprovedVersion))
	utils.PrintlnIfNotDefault(s.CurrentVersion, fmt.Sprintf("%scurrent_version: %s", prefix, s.CurrentVersion))
	utils.PrintlnIfNotDefault(s.CurrentVersionTimestamp, fmt.Sprintf("%scurrent_version_timestamp: %q", prefix, s.CurrentVersionTimestamp))
	utils.PrintlnIfNotDefault(s.LatestVersion, fmt.Sprintf("%slatest_version: %s", prefix, s.LatestVersion))
	utils.PrintlnIfNotDefault(s.LatestVersionTimestamp, fmt.Sprintf("%slatest_version_timestamp: %q", prefix, s.LatestVersionTimestamp))
}
