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

// Package util provides utility functions for the Argus project.
package util

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

// MapStringStringOmitNull is a map[string]string that ignores null values when unmarshalling.
// Any key with a null value in YAML/JSON will be omitted from the map.
type MapStringStringOmitNull map[string]string

// UnmarshalYAML implements the YAML unmarshaler for MapStringStringOmitNull.
// It decodes the YAML into a temporary map of *string, and only keeps
// keys whose value is non-nil. This ensures null values are not turned into "".
func (m *MapStringStringOmitNull) UnmarshalYAML(value *yaml.Node) error {
	temp := map[string]*string{}
	if err := value.Decode(&temp); err != nil {
		return err
	}

	res := make(map[string]string, len(temp))
	for k, v := range temp {
		if v != nil && *v != "" {
			res[k] = *v
		}
	}
	*m = res
	return nil
}

// UnmarshalJSON implements the json unmarshaler for MapStringStringOmitNull.
// It decodes the JSON into a temporary map of *string, and only keeps
// keys whose value is non-nil. This prevents null values from becoming "".
func (m *MapStringStringOmitNull) UnmarshalJSON(data []byte) error {
	temp := map[string]*string{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	res := make(map[string]string, len(temp))
	for k, v := range temp {
		if v != nil {
			res[k] = *v
		}
	}
	*m = res
	return nil
}
