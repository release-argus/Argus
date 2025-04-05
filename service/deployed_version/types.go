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

// Package deployedver provides the deployed_version lookup service to for a service.
package deployedver

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/service/deployed_version/types/base"
	"github.com/release-argus/Argus/service/deployed_version/types/manual"
	"github.com/release-argus/Argus/service/deployed_version/types/web"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

// Lookup provides methods for retrieving the deployed version of a service.
type Lookup interface {
	base.Interface
}

// New returns a new Lookup.
func New(
	lType string, // "manual" | ("url"|"web")
	configFormat string, // "json" | "yaml"
	configData any, // []byte | string | *yaml.Node | json.RawMessage.
	options *opt.Options,
	status *status.Status,
	defaults, hardDefaults *base.Defaults,
) (base.Interface, error) {
	switch lType {
	case "url", "web":
		return web.New( //nolint:wrapcheck
			configFormat,
			configData,
			options,
			status,
			defaults,
			hardDefaults)
	case "manual":
		return manual.New( //nolint:wrapcheck
			configFormat,
			configData,
			options,
			status,
			defaults,
			hardDefaults)
	}

	// No/invalid type.
	errorMsg := "<required>"
	if lType != "" {
		errorMsg = fmt.Sprintf("%q <invalid>", lType)
	}
	return nil, fmt.Errorf("failed to unmarshal deployedver.Lookup:\n  type: %s (expected one of [%s])",
		errorMsg, strings.Join(PossibleTypes, ", "))
}

// Copy returns a copy of the Lookup.
func Copy(
	lookup Lookup,
) Lookup {
	if lookup == nil {
		return nil
	}

	// JSON of existing lookup.
	lookupJSON, _ := json.Marshal(lookup)

	svcStatus := status.Status{}
	svcStatus.Init(
		0, 0, 0,
		lookup.GetStatus().ServiceID, lookup.GetStatus().ServiceName,
		lookup.GetStatus().WebURL)

	// Create a new lookup.
	newLookup, _ := New(
		lookup.GetType(),
		"json", lookupJSON,
		lookup.GetOptions().Copy(),
		&svcStatus,
		lookup.GetDefaults(), lookup.GetHardDefaults())

	return newLookup
}

// IsEqual will return whether `this` lookup is the same as `other` (excluding status).
func IsEqual(this, other Lookup) bool {
	if other == nil || this == nil {
		// Equal if both are nil.
		return other == nil && this == nil
	}

	return this.GetOptions().String() == other.GetOptions().String() &&
		this.String(this, "") == other.String(other, "")
}

// UnmarshalJSON unmarshals a Lookup from JSON.
func UnmarshalJSON(data []byte) (Lookup, error) {
	return unmarshal(data, "json")
}

// UnmarshalYAML unmarshals a Lookup from YAML.
func UnmarshalYAML(data []byte) (Lookup, error) {
	return unmarshal(data, "yaml")
}

// unmarshal handles the unmarshalling of a Lookup.
//
// (dynamic typing).
func unmarshal(data []byte, format string) (Lookup, error) {
	baseErr := "failed to unmarshal deployedver.Lookup:"

	var temp struct {
		Type string `yaml:"type" json:"type"`
	}

	// Unmarshal into temp to extract the type.
	var err error
	switch format {
	case "json":
		err = json.Unmarshal(data, &temp)
	case "yaml":
		err = yaml.Unmarshal(data, &temp)
	default:
		return nil, fmt.Errorf("%s\n  unknown format: %q",
			baseErr, format)
	}
	if err != nil {
		errStr := util.FormatUnmarshalError(format, err)
		errStr = strings.ReplaceAll(errStr, "\n", "\n  ")
		return nil, fmt.Errorf("%s\n  %s",
			baseErr, errStr)
	}

	// -- Dynamic deployedVersion type --
	if temp.Type == "" {
		// TODO: Swap defaulting to an error
		temp.Type = "url"
	}
	// Supported type?
	if _, exists := ServiceMap[temp.Type]; !exists {
		return nil, fmt.Errorf("%s\n  type: %q <invalid> (expected one of [%s])",
			baseErr, temp.Type, strings.Join(PossibleTypes, ", "))
	}

	// New Lookup based on the type.
	lookup, err := New(temp.Type,
		format, data,
		nil,
		nil,
		nil, nil)
	if err != nil {
		return nil, err
	}

	return lookup, nil
}
