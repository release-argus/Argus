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
	"fmt"
	"regexp"

	command "github.com/release-argus/Argus/commands"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

var (
	jLog *util.JLog
)

// Require for version to be considered valid.
type Require struct {
	Status       *svcstatus.Status `yaml:"-"`                       // Service Status
	RegexContent string            `yaml:"regex_content,omitempty"` // "abc-[a-z]+-{{ version }}_amd64.deb" This regex must exist in the body of the URL to trigger new version actions
	RegexVersion string            `yaml:"regex_version,omitempty"` // "v*[0-9.]+" The version found must match this release to trigger new version actions
	Command      command.Command   `yaml:"command,omitempty"`       // Docker image tag requirements
	Docker       *DockerCheck      `yaml:"docker,omitempty"`        // Docker image tag requirements
}

// Init will give the filter package the log and the Service's Status.
func (r *Require) Init(log *util.JLog, status *svcstatus.Status) {
	if r == nil {
		return
	}

	if log != nil {
		jLog = log
	}

	r.Status = status
}

// Print the Require.
func (r *Require) Print(prefix string) {
	if r == nil {
		return
	}
	fmt.Printf("%srequire:\n", prefix)
	util.PrintlnIfNotDefault(r.RegexContent, fmt.Sprintf("%s  regex_content: %q", prefix, r.RegexContent))
	util.PrintlnIfNotDefault(r.RegexVersion, fmt.Sprintf("%s  regex_version: %q", prefix, r.RegexVersion))
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
