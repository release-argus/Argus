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

	"github.com/release-argus/Argus/notifiers/gotify"
	"github.com/release-argus/Argus/notifiers/slack"
	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/webhook"
)

// Status is the current state of the Service element (version and regex misses).
type Status struct {
	ApprovedVersion          string       `yaml:"approved_version,omitempty"`           // The version that's been approved
	DeployedVersion          string       `yaml:"deployed_version,omitempty"`           // Track the deployed version of the service from the last successful WebHook.
	DeployedVersionTimestamp string       `yaml:"deployed_version_timestamp,omitempty"` // UTC timestamp of DeployedVersion being changed.
	LatestVersion            string       `yaml:"latest_version,omitempty"`             // Latest version found from query().
	LatestVersionTimestamp   string       `yaml:"latest_version_timestamp,omitempty"`   // UTC timestamp of LatestVersion being changed.
	LastQueried              string       `yaml:"-"`                                    // UTC timestamp that version was last queried/checked.
	RegexMissesContent       uint         `yaml:"-"`                                    // Counter for the number of regex misses on URL content.
	RegexMissesVersion       uint         `yaml:"-"`                                    // Counter for the number of regex misses on version.
	Fails                    *StatusFails `yaml:"-"`                                    // Track the Gotify/Slack/WebHook fails
	// TODO: Remove V
	CurrentVersion          string `yaml:"current_version,omitempty"`           // Track the current version of the service from the last successful WebHook.
	CurrentVersionTimestamp string `yaml:"current_version_timestamp,omitempty"` // UTC timestamp that the current version change was noticed.
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

// SetDeployedVersion will set DeployedVersion as well as DeployedVersionTimestamp.
func (s *Status) SetDeployedVersion(version string) {
	s.DeployedVersion = version
	s.DeployedVersionTimestamp = time.Now().UTC().Format(time.RFC3339)
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
	utils.PrintlnIfNotDefault(s.DeployedVersion, fmt.Sprintf("%sdeployed_version: %s", prefix, s.DeployedVersion))
	utils.PrintlnIfNotDefault(s.DeployedVersionTimestamp, fmt.Sprintf("%sdeployed_version_timestamp: %q", prefix, s.DeployedVersionTimestamp))
	utils.PrintlnIfNotDefault(s.LatestVersion, fmt.Sprintf("%slatest_version: %s", prefix, s.LatestVersion))
	utils.PrintlnIfNotDefault(s.LatestVersionTimestamp, fmt.Sprintf("%slatest_version_timestamp: %q", prefix, s.LatestVersionTimestamp))
}
