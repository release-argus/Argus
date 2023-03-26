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

package svcstatus

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/util"
)

// Status is the current state of the Service element (version and regex misses).
type Status struct {
	ApprovedVersion          string `yaml:"-" json:"-"` // The version that's been approved
	DeployedVersion          string `yaml:"-" json:"-"` // Track the deployed version of the service from the last successful WebHook.
	DeployedVersionTimestamp string `yaml:"-" json:"-"` // UTC timestamp of DeployedVersion being changed.
	LatestVersion            string `yaml:"-" json:"-"` // Latest version found from query().
	LatestVersionTimestamp   string `yaml:"-" json:"-"` // UTC timestamp of LatestVersion being changed.
	LastQueried              string `yaml:"-" json:"-"` // UTC timestamp that version was last queried/checked.
	RegexMissesContent       uint   `yaml:"-" json:"-"` // Counter for the number of regex misses on URL content.
	RegexMissesVersion       uint   `yaml:"-" json:"-"` // Counter for the number of regex misses on version.
	Fails                    Fails  `yaml:"-" json:"-"` // Track the Notify/WebHook fails
	Deleting                 bool   `yaml:"-" json:"-"` // Flag to indicate the service is being deleted

	// Announces
	AnnounceChannel *chan []byte         `yaml:"-" json:"-"` // Announce to the WebSocket
	DatabaseChannel *chan dbtype.Message `yaml:"-" json:"-"` // Channel for broadcasts to the Database
	SaveChannel     *chan bool           `yaml:"-" json:"-"` // Channel for triggering a save of the config
	ServiceID       *string              `yaml:"-" json:"-"` // ID of the Service
	WebURL          *string              `yaml:"-" json:"-"` // Web URL of the Service
}

// String returns a string representation of the Status.
func (s *Status) String() string {
	fields := []util.Field{
		{Name: "approved_version", Value: s.ApprovedVersion},
		{Name: "deployed_version", Value: s.DeployedVersion},
		{Name: "deployed_version_timestamp", Value: s.DeployedVersionTimestamp},
		{Name: "latest_version", Value: s.LatestVersion},
		{Name: "latest_version_timestamp", Value: s.LatestVersionTimestamp},
		{Name: "last_queried", Value: s.LastQueried},
		{Name: "regex_misses_content", Value: s.RegexMissesContent},
		{Name: "regex_misses_version", Value: s.RegexMissesVersion},
		{Name: "fails", Value: s.Fails},
	}

	var buf bytes.Buffer
	for _, f := range fields {
		switch v := f.Value.(type) {
		case string:
			if v != "" {
				fmt.Fprint(&buf, f.Name, ": ", v, ", ")
			}
		case uint:
			if v != 0 {
				fmt.Fprint(&buf, f.Name, ": ", v, ", ")
			}
		case Fails:
			if fails := v.String(); fails != "" {
				fmt.Fprint(&buf, f.Name, ": {", fails, "}, ")
			}
		}
	}

	return strings.TrimSuffix(buf.String(), ", ")
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
	Shoutrrr map[string]*bool `yaml:"-" json:"-"` // Shoutrrr unsent/fail/pass.
	Command  []*bool          `yaml:"-" json:"-"` // Command unsent/fail/pass.
	WebHook  map[string]*bool `yaml:"-" json:"-"` // WebHook unsent/fail/pass.
}

// String returns a string representation of the Fails.
func (s *Fails) String() string {
	fields := []util.Field{
		{Name: "shoutrrr", Value: s.Shoutrrr},
		{Name: "command", Value: s.Command},
		{Name: "webhook", Value: s.WebHook},
	}

	var buf bytes.Buffer
	for _, f := range fields {
		switch v := f.Value.(type) {
		case map[string]*bool:
			if len(v) == 0 {
				continue
			}
			// Check for fails in the map.
			hasFail := false
			for i := range v {
				if util.DefaultIfNil(v[i]) {
					hasFail = true
					break
				}
			}
			// If there are no fails, skip this field.
			if !hasFail {
				continue
			}

			fmt.Fprint(&buf, f.Name, ": {")

			// Create a slice of keys and sort them
			keys := make([]string, 0, len(v))
			for k := range v {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			// Iterate over the sorted keys
			for _, key := range keys {
				if util.DefaultIfNil(v[key]) {
					fmt.Fprint(&buf, key, ": ", *v[key], ", ")
				}
			}
			// Remove the trailing ", "
			buf.Truncate(buf.Len() - 2)
			fmt.Fprint(&buf, "}, ")
		case []*bool:
			if len(v) > 0 {
				// Check for fails in the list.
				hasFail := false
				for i := range v {
					if util.DefaultIfNil(v[i]) {
						hasFail = true
						break
					}
				}
				// If there are no fails, skip this field.
				if !hasFail {
					continue
				}

				fmt.Fprint(&buf, f.Name, ": [")
				for i, v := range v {
					if util.DefaultIfNil(v) {
						fmt.Fprint(&buf, i, ": ", *v, ", ")
					}
				}
				// Remove the trailing ", "
				buf.Truncate(buf.Len() - 2)
				fmt.Fprint(&buf, "], ")
			}
		}
	}

	return strings.TrimSuffix(buf.String(), ", ")
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

	if s.DatabaseChannel != nil {
		*s.DatabaseChannel <- dbtype.Message{
			ServiceID: *s.ServiceID,
			Cells: []dbtype.Cell{
				{Column: "deployed_version", Value: s.DeployedVersion},
				{Column: "deployed_version_timestamp", Value: s.DeployedVersionTimestamp},
			},
		}
	}
}

// SetLatestVersion will set LatestVersion to `version` and LatestVersionTimestamp to s.LastQueried.
func (s *Status) SetLatestVersion(version string) {
	s.LatestVersion = version
	s.LatestVersionTimestamp = s.LastQueried

	// Clear the fail status of WebHooks/Commands
	s.Fails.resetFails()

	if s.DatabaseChannel != nil {
		*s.DatabaseChannel <- dbtype.Message{
			ServiceID: *s.ServiceID,
			Cells: []dbtype.Cell{
				{Column: "latest_version", Value: s.LatestVersion},
				{Column: "latest_version_timestamp", Value: s.LatestVersionTimestamp},
			},
		}
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
	if util.DefaultIfNil(s.WebURL) == "" {
		return ""
	}

	return util.TemplateString(*s.WebURL, util.ServiceInfo{LatestVersion: s.LatestVersion})
}

// Print will print the Status.
func (s *Status) Print(prefix string) {
	util.PrintlnIfNotDefault(s.ApprovedVersion,
		fmt.Sprintf("%sapproved_version: %s", prefix, s.ApprovedVersion))
	util.PrintlnIfNotDefault(s.DeployedVersion,
		fmt.Sprintf("%sdeployed_version: %s", prefix, s.DeployedVersion))
	util.PrintlnIfNotDefault(s.DeployedVersionTimestamp,
		fmt.Sprintf("%sdeployed_version_timestamp: %q", prefix, s.DeployedVersionTimestamp))
	util.PrintlnIfNotDefault(s.LatestVersion,
		fmt.Sprintf("%slatest_version: %s", prefix, s.LatestVersion))
	util.PrintlnIfNotDefault(s.LatestVersionTimestamp,
		fmt.Sprintf("%slatest_version_timestamp: %q", prefix, s.LatestVersionTimestamp))
}
