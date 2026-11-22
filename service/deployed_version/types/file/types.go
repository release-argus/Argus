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

// Package file provides a local-file deployed-version lookup.
package file

import (
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/service/deployed_version/types/base"
	"github.com/release-argus/Argus/service/status"
)

// Type is the lookup type identifier for file-based deployed version lookups.
var Type = "file"

// Lookup reads the deployed version from a local file.
type Lookup struct {
	base.Lookup `json:",embed" yaml:",inline"`

	Path          string `json:"path,omitzero" yaml:"path,omitzero"`                     // Absolute path to a file allowed by settings.data.file_read_paths.
	Regex         string `json:"regex,omitzero" yaml:"regex,omitzero"`                   // OPTIONAL: regex to extract the version.
	RegexTemplate string `json:"regex_template,omitzero" yaml:"regex_template,omitzero"` // OPTIONAL: template to apply to the RegEx match.
}

// Copy returns a deep copy of the receiver.
func (l *Lookup) Copy(svcStatus *status.Status) base.Interface {
	if l == nil {
		return nil
	}

	return &Lookup{
		Lookup:        *l.Lookup.Clone(svcStatus), //nolint:staticcheck
		Path:          l.Path,
		Regex:         l.Regex,
		RegexTemplate: l.RegexTemplate,
	}
}

// String returns a string representation of the receiver.
func (l *Lookup) String(prefix string) string {
	return decode.ToYAMLString(l, prefix)
}

// GetType returns the type of the receiver.
func (l *Lookup) GetType() string {
	return Type
}
