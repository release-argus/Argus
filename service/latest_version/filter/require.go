// Copyright [2024] [Argus]
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

// Package filter provides filtering for latest_version queries.
package filter

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

var (
	jLog *util.JLog
)

// LogInit for this package.
func LogInit(log *util.JLog) {
	jLog = log
	command.LogInit(log)
}

// RequireDefaults are the default values for the Require struct.
// It contains configuration defaults for validating version requirements.
type RequireDefaults struct {
	Docker DockerCheckDefaults `yaml:"docker" json:"docker"` // Docker image tag requirements.
}

// Default sets this RequireDefaults to the default values.
func (r *RequireDefaults) Default() {
	r.Docker.Default()
}

// CheckValues validates the fields of the RequireDefaults struct.
func (r *RequireDefaults) CheckValues(prefix string) error {
	if dockerErr := r.Docker.CheckValues(prefix + "  "); dockerErr != nil {
		return fmt.Errorf("%sdocker:\n%w",
			prefix, dockerErr)
	}

	return nil
}

// Require defines validation requirements that must be met for a version to be considered valid.
type Require struct {
	Status       *status.Status  `yaml:"-" json:"-"`                                             // Service Status.
	RegexContent string          `yaml:"regex_content,omitempty" json:"regex_content,omitempty"` // "abc-[a-z]+-{{ version }}_amd64.deb" This regex must exist in the body of the URL to trigger new version actions.
	RegexVersion string          `yaml:"regex_version,omitempty" json:"regex_version,omitempty"` // "v*[0-9.]+" The version found must match this release to trigger new version actions.
	Command      command.Command `yaml:"command,omitempty" json:"command,omitempty"`             // Require Command to pass.
	Docker       *DockerCheck    `yaml:"docker,omitempty" json:"docker,omitempty"`               // Docker image tag requirements.
}

// String returns a string representation of the Require.
func (r *Require) String() string {
	if r == nil {
		return ""
	}
	return util.ToYAMLString(r, "")
}

// Init will give the Require struct the Service's Status and defaults.
func (r *Require) Init(
	status *status.Status,
	defaults *RequireDefaults,
) {
	if r == nil {
		return
	}

	r.Status = status

	if r.Docker != nil {
		r.Docker.Defaults = &defaults.Docker
		r.removeUnusedRequireDocker()
	}
}

// CheckValues validates the fields of the Require struct.
func (r *Require) CheckValues(prefix string) error {
	if r == nil {
		return nil
	}

	var errs []error
	// Content RegEx.
	if r.RegexContent != "" {
		if !util.CheckTemplate(r.RegexContent) {
			errs = append(errs,
				fmt.Errorf("%sregex_content: %q <invalid> (didn't pass templating)",
					prefix, r.RegexContent))
		} else {
			_, err := regexp.Compile(r.RegexContent)
			if err != nil {
				errs = append(errs,
					fmt.Errorf("%sregex_content: %q <invalid> (Invalid RegEx)",
						prefix, r.RegexContent))
			}
		}
	}

	// Version RegEx.
	if r.RegexVersion != "" {
		if _, err := regexp.Compile(r.RegexVersion); err != nil {
			errs = append(errs,
				fmt.Errorf("%sregex_version: %q <invalid> (Invalid RegEx)",
					prefix, r.RegexVersion))
		}
	}

	for _, cmd := range r.Command {
		if !util.CheckTemplate(cmd) {
			errs = append(errs,
				fmt.Errorf("%scommand: %v (%q) <invalid> (didn't pass templating)",
					prefix, r.Command, cmd))
			break
		}
	}

	util.AppendCheckError(&errs, prefix, "docker", r.Docker.CheckValues(prefix+"  "))

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// Inherit will copy the Docker queryToken if it is what the provider would fetch.
func (r *Require) Inherit(from *Require) {
	// If the Docker token is for the same image.
	if r != nil && r.Docker != nil &&
		from != nil && from.Docker != nil &&
		r.Docker.Type == from.Docker.Type &&
		r.Docker.Image == from.Docker.Image &&
		// with the r.Docker referencing the from.Docker token.
		r.Docker.Username == from.Docker.Username &&
		((r.Docker.Token == util.SecretValue && from.Docker.Token != "") ||
			r.Docker.Token == from.Docker.Token) {
		// Copy it.
		r.Docker.Token = from.Docker.Token
		queryToken, validUntil := from.Docker.CopyQueryToken()
		r.Docker.SetQueryToken(
			from.Docker.Token,
			queryToken, validUntil)
	}
}

// removeUnusedRequireDocker will nil the Docker requirement if there is no image or tag.
func (r *Require) removeUnusedRequireDocker() {
	if r == nil || r.Docker == nil {
		return
	}

	// Remove the Docker requirement if there's no image or tag.
	if r.Docker.Image == "" || r.Docker.Tag == "" {
		r.Docker = nil
	}
}
