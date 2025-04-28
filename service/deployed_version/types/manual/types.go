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
	"errors"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/service/deployed_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

// Lookup is a web-based lookup type.
type Lookup struct {
	base.Lookup `json:",inline" yaml:",inline"` // Base struct for a Lookup.

	mutex   sync.RWMutex // Lock for the Lookup.
	Version string       `json:"version,omitempty" yaml:"version,omitempty"` // OPTIONAL: Version to initialise with/set to.
}

// New returns a new Lookup from a string in a given format (json/yaml).
func New(
	configFormat string, // "json" | "yaml"
	configData any, // []byte | string | *yaml.Node | json.RawMessage.
	options *opt.Options,
	status *status.Status,
	defaults, hardDefaults *base.Defaults,
) (*Lookup, error) {
	lookup := &Lookup{}

	// Unmarshal.
	if err := util.UnmarshalConfig(configFormat, configData, lookup); err != nil {
		errStr := util.FormatUnmarshalError(configFormat, err)
		errStr = strings.ReplaceAll(errStr, "\n", "\n  ")
		return nil, errors.New("failed to unmarshal manual.Lookup:\n  " + errStr)
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
	return l.unmarshal(func(v interface{}) error {
		return json.Unmarshal(data, v)
	})
}

// UnmarshalYAML will unmarshal the Lookup.
func (l *Lookup) UnmarshalYAML(value *yaml.Node) error {
	return l.unmarshal(func(v interface{}) error {
		return value.Decode(v)
	})
}

// unmarshal will unmarshal the Lookup using the provided unmarshal function.
func (l *Lookup) unmarshal(unmarshalFunc func(interface{}) error) error {
	// Alias to avoid recursion.
	type Alias Lookup
	aux := &struct {
		*Alias `json:",inline" yaml:",inline"`
	}{
		Alias: (*Alias)(l),
	}

	// Lock the mutex if it's an existing Lookup.
	if l != nil {
		l.mutex.Lock()
		defer l.mutex.Unlock()
	}

	// Unmarshal using the provided function.
	if err := unmarshalFunc(aux); err != nil {
		return errors.New(strings.Replace(err.Error(), ".Alias", "", 1))
	}
	l.Type = "manual"
	return nil
}
