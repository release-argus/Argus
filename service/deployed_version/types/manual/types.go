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

// Package manual provides a manually set version lookup.
package manual

import (
	"encoding/json"
	"fmt"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/service/deployed_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

// Lookup is a web-based lookup type.
type Lookup struct {
	base.Lookup `yaml:",inline" json:",inline"` // Base struct for a Lookup.

	mutex   sync.RWMutex // Lock for the Lookup.
	Version string       `yaml:"version,omitempty" json:"version,omitempty"` // OPTIONAL: Version to initialise with/set to.
}

// New returns a new Lookup from a string in a given format (json/yaml).
func New(
	configFormat string,
	configData any, // []byte | string | *yaml.Node | json.RawMessage.
	options *opt.Options,
	status *status.Status,
	defaults, hardDefaults *base.Defaults,
) (*Lookup, error) {
	lookup := &Lookup{}

	// Unmarshal.
	if err := util.UnmarshalConfig(configFormat, configData, lookup); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manual.Lookup:\n%w", err)
	}

	lookup.Init(
		options,
		status,
		defaults, hardDefaults)

	if lookup.Status == nil || lookup.Options == nil {
		return lookup, nil
	}

	// Transfer the Version to the Status.
	err := lookup.CheckValues("")

	return lookup, err
}

// UnmarshalJSON will unmarshal the Lookup.
func (l *Lookup) UnmarshalJSON(data []byte) error {
	// Alias to avoid recursion.
	type Alias Lookup
	aux := &struct {
		*Alias `json:",inline"`
	}{Alias: (*Alias)(l)}

	// Lock the mutex if it's an existing Lookup.
	if l != nil {
		l.mutex.Lock()
		defer l.mutex.Unlock()
	}

	// Unmarshal.
	if err := json.Unmarshal(data, aux); err != nil {
		return err //nolint:wrapcheck
	}
	l.Type = "manual"

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

	l.Type = "manual"
	return nil
}
