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

package filter

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	command "github.com/release-argus/Argus/commands"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"gopkg.in/yaml.v3"
)

var (
	jLog *util.JLog
)

// LogInit for this package.
func LogInit(log *util.JLog) {
	jLog = log
	command.LogInit(log)
}

// Require for version to be considered valid.
type Require struct {
	Status       *svcstatus.Status `yaml:"-" json:"-"`                                             // Service Status
	RegexContent string            `yaml:"regex_content,omitempty" json:"regex_content,omitempty"` // "abc-[a-z]+-{{ version }}_amd64.deb" This regex must exist in the body of the URL to trigger new version actions
	RegexVersion string            `yaml:"regex_version,omitempty" json:"regex_version,omitempty"` // "v*[0-9.]+" The version found must match this release to trigger new version actions
	Command      command.Command   `yaml:"command,omitempty" json:"command,omitempty"`             // Require Command to pass
	Docker       *DockerCheck      `yaml:"docker,omitempty" json:"docker,omitempty"`               // Docker image tag requirements
}

// String returns a string representation of the Require.
func (r *Require) String() string {
	if r == nil {
		return "<nil>"
	}
	yamlBytes, _ := yaml.Marshal(r)
	return string(yamlBytes)
}

// Init will give the filter package the Service's Status.
func (r *Require) Init(status *svcstatus.Status) {
	if r == nil {
		return
	}

	r.Status = status
}

// Print the Require.
func (r *Require) Print(prefix string) {
	if r == nil {
		return
	}
	fmt.Printf("%srequire:\n", prefix)
	util.PrintlnIfNotDefault(r.RegexContent,
		fmt.Sprintf("%s  regex_content: %q", prefix, r.RegexContent))
	util.PrintlnIfNotDefault(r.RegexVersion,
		fmt.Sprintf("%s  regex_version: %q", prefix, r.RegexVersion))
	if len(r.Command) != 0 {
		fmt.Printf("%s  command: %s\n", prefix, r.Command.FormattedString())
	}
	r.Docker.Print(prefix + "  ")
}

// CheckValues of the Require option.
func (r *Require) CheckValues(prefix string) (errs error) {
	if r == nil {
		return
	}

	// Content RegEx
	if r.RegexContent != "" {
		if !util.CheckTemplate(r.RegexContent) {
			errs = fmt.Errorf("%s%s  regex_content: %q <invalid> (didn't pass templating)\\",
				util.ErrorToString(errs), prefix, r.RegexContent)
		} else {
			_, err := regexp.Compile(r.RegexContent)
			if err != nil {
				errs = fmt.Errorf("%s%s  regex_content: %q <invalid> (Invalid RegEx)\\",
					util.ErrorToString(errs), prefix, r.RegexContent)
			}
		}
	}

	// Version RegEx
	if r.RegexVersion != "" {
		_, err := regexp.Compile(r.RegexVersion)
		if err != nil {
			errs = fmt.Errorf("%s%s  regex_version: %q <invalid> (Invalid RegEx)\\",
				util.ErrorToString(errs), prefix, r.RegexVersion)
		}
	}

	for i := range r.Command {
		if !util.CheckTemplate(r.Command[i]) {
			errs = fmt.Errorf("%s%s  command: %v (%q) <invalid> (didn't pass templating)\\",
				util.ErrorToString(errs), prefix, r.Command, r.Command[i])
			break
		}
	}

	if err := r.Docker.CheckValues(prefix + "    "); err != nil {
		errs = fmt.Errorf("%s%s  docker:\\%w",
			util.ErrorToString(errs), prefix, err)
	}

	if errs != nil {
		errs = fmt.Errorf("%srequire:\\%s",
			prefix, util.ErrorToString(errs))
	}
	return
}

// ReqyureFromStr will convert a JSON string to a Require.
func RequireFromStr(jsonStr *string, previous *Require, logFrom *util.LogFrom) (*Require, error) {
	// jsonStr == nil when it hasn't been changed, so just use previous
	if jsonStr == nil || *jsonStr == "{}" {
		return previous, nil
	}

	var require Require
	dec := json.NewDecoder(strings.NewReader(*jsonStr))
	dec.DisallowUnknownFields()
	// Ignore the JSON if it failed to unmarshal
	if err := dec.Decode(&require); err != nil {
		jLog.Error(
			fmt.Sprintf("Failed converting JSON - %q\n%s",
				*jsonStr, util.ErrorToString(err)),
			*logFrom, true)
		return nil, fmt.Errorf("require - %w", err)
	}

	// Default the params to the previous values
	if previous != nil {
		// Get JSON keys so we know which have been changed
		jsonKeys := util.GetKeysFromJSON(*jsonStr)

		if !util.Contains(jsonKeys, "regex_content") {
			require.RegexContent = previous.RegexContent
		}
		if !util.Contains(jsonKeys, "regex_version") {
			require.RegexVersion = previous.RegexVersion
		}

		if !util.Contains(jsonKeys, "command") {
			require.Command = previous.Command
		}

		// Default the Docker params
		if previous.Docker != nil {
			if util.Contains(jsonKeys, "docker") {
				// Default params that haven't been changed
				if !util.Contains(jsonKeys, "docker.type") {
					require.Docker.Type = previous.Docker.Type
				}
				if !util.Contains(jsonKeys, "docker.image") {
					require.Docker.Image = previous.Docker.Image
				}
				if !util.Contains(jsonKeys, "docker.tag") {
					require.Docker.Tag = previous.Docker.Tag
				}
				if !util.Contains(jsonKeys, "docker.username") {
					require.Docker.Username = previous.Docker.Username
				}
				if !util.Contains(jsonKeys, "docker.token") {
					require.Docker.Token = previous.Docker.Token
				}
				// Haven't changed any Docker params
			} else {
				require.Docker = previous.Docker
			}
		}
	}
	if require.Docker != nil {
		// nil Docker if only Type
		if require.Docker.Image == "" && require.Docker.Tag == "" {
			require.Docker = nil

			// Default Docker Type
		} else if require.Docker.Type == "" {
			require.Docker.Type = "hub"
		}
	}

	// Check validity
	err := require.CheckValues("")
	if err != nil {
		return nil, err
	}

	return &require, err
}
