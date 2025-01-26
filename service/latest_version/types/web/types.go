// Copyright [2025] [Argus]
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

// Package web provides a web-based lookup type.
package web

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

// Lookup is a web-based lookup type.
type Lookup struct {
	base.Lookup `yaml:",inline" json:",inline"` // Base struct for a Lookup.

	AllowInvalidCerts *bool `yaml:"allow_invalid_certs,omitempty" json:"allow_invalid_certs,omitempty"` // Allow invalid SSL certificates.
}

// New returns a new Lookup from a string in a given format (json/yaml).
func New(
	configFormat string,
	configData interface{}, // []byte | string | *yaml.Node.
	options *opt.Options,
	status *status.Status,
	defaults, hardDefaults *base.Defaults,
) (*Lookup, error) {
	lookup := &Lookup{}

	// Unmarshal.
	if err := util.UnmarshalConfig(configFormat, configData, lookup); err != nil {
		return nil, fmt.Errorf("failed to unmarshal web.Lookup:\n%w", err)
	}

	lookup.Init(options,
		status,
		defaults, hardDefaults)

	return lookup, nil
}

// UnmarshalJSON will unmarshal the Lookup.
func (l *Lookup) UnmarshalJSON(data []byte) error {
	// Alias to avoid recursion.
	type Alias Lookup
	aux := &struct {
		*Alias `json:",inline"`
	}{Alias: (*Alias)(l)}

	// Unmarshal.
	if err := json.Unmarshal(data, aux); err != nil {
		return err //nolint:wrapcheck
	}
	l.Type = "url"

	return nil
}

// UnmarshalYAML will unmarshal the Lookup.
func (l *Lookup) UnmarshalYAML(value *yaml.Node) error {
	// Alias to avoid recursion.
	type Alias Lookup
	aux := &struct {
		*Alias `yaml:",inline"`
	}{
		Alias: (*Alias)(l),
	}

	// Decode the YAML node into the struct.
	if err := value.Decode(aux); err != nil {
		return err //nolint:wrapcheck
	}

	l.Type = "url"
	return nil
}
