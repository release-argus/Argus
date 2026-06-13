// Copyright [2026] [Argus]
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
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/service/latest_version/filter/docker"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

// Require defines validation requirements that must be met for a version to be considered valid.
type Require struct {
	Status       *status.Status  `json:"-" yaml:"-"`                                             // Service Status.
	RegexContent string          `json:"regex_content,omitempty" yaml:"regex_content,omitempty"` // "abc-[a-z]+-{{ version }}_amd64.deb" This regex must exist in the body of the URL to trigger new version actions.
	RegexVersion string          `json:"regex_version,omitempty" yaml:"regex_version,omitempty"` // "v*[0-9.]+" The version found must match this release to trigger new version actions.
	Command      command.Command `json:"command,omitempty" yaml:"command,omitempty"`             // Require Command to pass.
	Docker       docker.Registry `json:"docker,omitempty" yaml:"docker,omitempty"`               // Docker image tag requirements.

	defaults *RequireDefaults // Defaults for Require.
}

// IsZero implements the yaml.IsZeroer interface.
func (r *Require) IsZero() bool {
	return r == nil || (r.RegexContent == "" && r.RegexVersion == "" && len(r.Command) == 0 &&
		(r.Docker == nil || r.Docker.IsZero()))
}

// RequireDecode is an unmarshal-only helper for [Require].
type RequireDecode struct {
	RegexContent string          `json:"regex_content,omitempty" yaml:"regex_content,omitempty"` // "abc-[a-z]+-{{ version }}_amd64.deb" This regex must exist in the body of the URL to trigger new version actions.
	RegexVersion string          `json:"regex_version,omitempty" yaml:"regex_version,omitempty"` // "v*[0-9.]+" The version found must match this release to trigger new version actions.
	Command      command.Command `json:"command,omitempty" yaml:"command,omitempty"`             // Require Command to pass.
}

// String returns a string representation of the receiver.
func (r *Require) String(prefix string) string {
	if r == nil {
		return ""
	}
	return decode.ToYAMLString(r, prefix)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// Use [Decode] for a full unmarshal.
func (r *Require) UnmarshalJSON(data []byte) error {
	return r.unmarshal("json", data)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
// Use [Decode] for a full unmarshal.
func (r *Require) UnmarshalYAML(data []byte) error {
	return r.unmarshal("yaml", data)
}

// unmarshal implements the format.Unmarshaler interface.
func (r *Require) unmarshal(format string, data []byte) error {
	if len(data) == 0 {
		return nil
	}

	aux := RequireDecode{
		RegexContent: r.RegexContent,
		RegexVersion: r.RegexVersion,
		Command:      r.Command,
	}

	// Unmarshal.
	if err := decode.Unmarshal(format, data, &aux); err != nil {
		return err //nolint:wrapcheck
	}
	r.RegexContent = aux.RegexContent
	r.RegexVersion = aux.RegexVersion
	r.Command = aux.Command

	return nil
}

// Copy returns a deep copy of the receiver with the given status.
func (r *Require) Copy(status *status.Status) *Require {
	if r == nil {
		return nil
	}

	var requireDocker docker.Registry
	if r.Docker != nil {
		requireDocker = r.Docker.Copy()
	}

	return &Require{
		Status:       status,
		RegexContent: r.RegexContent,
		RegexVersion: r.RegexVersion,
		Command:      r.Command.Copy(),
		Docker:       requireDocker,
		defaults:     r.defaults,
	}
}

// DockerTagCheck verifies that Tag exists for Image.
func (r *Require) DockerTagCheck(
	version string,
) error {
	if r == nil || r.Docker == nil {
		return nil
	}

	return r.Docker.Check(version) //nolint:wrapcheck
}

// Init will assign the status and defaults to the receiver.
func (r *Require) Init(
	status *status.Status,
	defaults *RequireDefaults,
) {
	if r == nil {
		return
	}

	r.Status = status

	r.defaults = defaults
	if r.Docker != nil {
		r.removeUnusedRequireDocker()
	}
}

// CheckValues validates the fields of the receiver.
func (r *Require) CheckValues() error {
	if r == nil {
		return nil
	}

	var errs []error
	// Content RegEx.
	if r.RegexContent != "" {
		if !util.CheckTemplate(r.RegexContent) {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "regex_content",
					Value:       r.RegexContent,
					Description: "didn't pass templating",
				},
			)
		} else {
			_, err := regexp.Compile(r.RegexContent)
			if err != nil {
				errs = append(
					errs,
					&decode.FieldError{
						Key:         "regex_content",
						Value:       r.RegexContent,
						Description: "invalid RegEx",
					},
				)
			}
		}
	}

	// Version RegEx.
	if r.RegexVersion != "" {
		if _, err := regexp.Compile(r.RegexVersion); err != nil {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "regex_version",
					Value:       r.RegexVersion,
					Description: "invalid RegEx",
				},
			)
		}
	}

	for _, cmd := range r.Command {
		if !util.CheckTemplate(cmd) {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "command",
					Value:       r.Command.String(),
					Description: fmt.Sprintf("(%q) didn't pass templating", cmd),
				},
			)
			break
		}
	}

	if r.Docker != nil {
		// Clear Docker if no image:tag (with defaults), or no image / tag (without defaults).
		if (r.Docker.GetImage() == "" && r.Docker.GetTag() == "") ||
			(r.Docker.GetImageSelf() == "" && r.Docker.GetTagSelf() == "") {
			r.Docker = nil
		} else if err := r.Docker.CheckValues(); err != nil {
			errs = append(
				errs,
				&decode.KeyFieldError{
					Key: "docker",
					Err: err,
				},
			)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// Inherit will copy the Docker queryToken if it is what the provider would fetch.
func (r *Require) Inherit(from *Require) {
	if r != nil && r.Docker != nil &&
		from != nil && from.Docker != nil {
		r.Docker.Inherit(from.Docker)
	}
}

// removeUnusedRequireDocker will nil the Docker requirement if there is no image or tag.
func (r *Require) removeUnusedRequireDocker() {
	if r == nil || r.Docker == nil {
		return
	}

	// Remove the Docker requirement if there's no image or tag.
	if r.Docker.GetImage() == "" || r.Docker.GetTag() == "" {
		r.Docker = nil
	}
}
