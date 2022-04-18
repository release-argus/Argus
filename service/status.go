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
	"time"

	"github.com/hymenaios-io/Hymenaios/notifiers/gotify"
	"github.com/hymenaios-io/Hymenaios/notifiers/slack"
	"github.com/hymenaios-io/Hymenaios/utils"
	"github.com/hymenaios-io/Hymenaios/webhook"
)

// Status is the current state of the Service element (version and regex misses).
type Status struct {
	ApprovedVersion         *string      `yaml:"approved_version,omitempty"`          // The version that's been approved
	CurrentVersion          *string      `yaml:"current_version,omitempty"`           // Track the current version of the service from the last successful WebHook.
	CurrentVersionTimestamp *string      `yaml:"current_version_timestamp,omitempty"` // UTC timestamp of CurrentVersion being changed.
	LatestVersion           *string      `yaml:"latest_version,omitempty"`            // Latest version found from query().
	LatestVersionTimestamp  *string      `yaml:"latest_version_timestamp,omitempty"`  // UTC timestamp of LatestVersion being changed.
	LastQueried             *string      `yaml:"-"`                                   // UTC timestamp that version was last queried/checked.
	RegexMissesContent      uint         `yaml:"-"`                                   // Counter for the number of regex misses on URL content.
	RegexMissesVersion      uint         `yaml:"-"`                                   // Counter for the number of regex misses on version.
	Fails                   *StatusFails `yaml:"-"`                                   // Track the Gotify/Slack/WebHook fails
}

// StatusFails keeps track of whether any of the notifications failed on the last version change.
type StatusFails struct {
	Gotify  *[]bool `yaml:"-"` // Track whether any of the Slice failed.
	Slack   *[]bool `yaml:"-"` // Track whether any of the Slice failed.
	WebHook *[]bool `yaml:"-"` // Track whether any of the WebHookSlice failed.
}

// Init initialises the Status vars when more than the default value is needed.
func (s *Status) Init(
	gotifies *gotify.Slice,
	slacks *slack.Slice,
	webhooks *webhook.Slice,
) {
	s.Fails = new(StatusFails)
	if gotifies != nil {
		gotifyFails := make([]bool, len(*gotifies))
		s.Fails.Gotify = &gotifyFails
	}
	if slacks != nil {
		slackFails := make([]bool, len(*slacks))
		s.Fails.Slack = &slackFails
	}
	if webhooks != nil {
		webhookFails := make([]bool, len(*webhooks))
		s.Fails.WebHook = &webhookFails
	}
}

// SetCurrentVersion will set CurrentVersion as well as CurrentVersionTimestamp.
func (s *Status) SetCurrentVersion(version string) {
	if s.CurrentVersion == nil {
		initString0 := ""
		s.CurrentVersion = &initString0
		initString1 := ""
		s.CurrentVersionTimestamp = &initString1
	}

	*s.CurrentVersion = version
	*s.CurrentVersionTimestamp = time.Now().UTC().Format(time.RFC3339)
}

// SetLastQueried will update LastQueried to now.
func (s *Status) SetLastQueried() {
	if s.LastQueried == nil {
		init := ""
		s.LastQueried = &init
	}

	*s.LastQueried = time.Now().UTC().Format(time.RFC3339)
}

// SetLatestVersion will set LatestVersion to `version` and LatestVersionTimestamp to s.LastQueried.
func (s *Status) SetLatestVersion(version string) {
	if s.LatestVersion == nil {
		initString0 := ""
		s.LatestVersion = &initString0
		initString1 := ""
		s.LatestVersionTimestamp = &initString1
	}

	*s.LatestVersion = version
	*s.LatestVersionTimestamp = *s.LastQueried
}

// Print will print the Status.
func (s *Status) Print(prefix string) {
	utils.PrintlnIfNotNil(s.ApprovedVersion, fmt.Sprintf("%sapproved_version: %s", prefix, utils.DefaultIfNil(s.ApprovedVersion)))
	utils.PrintlnIfNotNil(s.CurrentVersion, fmt.Sprintf("%scurrent_version: %s", prefix, utils.DefaultIfNil(s.CurrentVersion)))
	utils.PrintlnIfNotNil(s.CurrentVersionTimestamp, fmt.Sprintf("%scurrent_version_timestamp: %q", prefix, utils.DefaultIfNil(s.CurrentVersionTimestamp)))
	utils.PrintlnIfNotNil(s.LatestVersion, fmt.Sprintf("%slatest_version: %s", prefix, utils.DefaultIfNil(s.LatestVersion)))
	utils.PrintlnIfNotNil(s.LatestVersionTimestamp, fmt.Sprintf("%slatest_version_timestamp: %q", prefix, utils.DefaultIfNil(s.LatestVersionTimestamp)))
}
