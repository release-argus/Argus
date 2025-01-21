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
	"fmt"

	"gopkg.in/yaml.v3"
)

// RawNode is a struct that holds a *yaml.Node.
type RawNode struct{ *yaml.Node }

// UnmarshalYAML handles the unmarshalling of a RawNode.
func (n *RawNode) UnmarshalYAML(node *yaml.Node) error {
	n.Node = node
	return nil
}

// UnmarshalConfig will unmarshal configuration data.
//
// Parameters:
//
//	configFormat: json/yaml
//	configData: []byte | string | *yaml.Node.
//	to: Pointer to unmarshal into.
func UnmarshalConfig(
	configFormat string,
	configData interface{}, // []byte | string | *yaml.Node.
	to interface{}, // struct pointer to unmarshal into.
) error {
	var rawData []byte
	switch v := configData.(type) {
	case []byte:
		rawData = v
	case string:
		rawData = []byte(v)
	case *yaml.Node:
		return v.Decode(to) //nolint:wrapcheck
	default:
		return fmt.Errorf("unsupported configData type: %T", configData)
	}

	// Unmarshal rawData based on configFormat.
	switch configFormat {
	case "json":
		return json.Unmarshal(rawData, to) //nolint:wrapcheck
	case "yaml":
		if unmarshaler, ok := to.(yaml.Unmarshaler); ok {
			// If `to` implements yaml.Unmarshaler, use its UnmarshalYAML method.
			var node yaml.Node
			if err := yaml.Unmarshal(rawData, &node); err != nil {
				return err //nolint:wrapcheck
			}
			return unmarshaler.UnmarshalYAML(&node) //nolint:wrapcheck
		}
		// For structs without custom UnmarshalYAML.
		return yaml.Unmarshal(rawData, to) //nolint:wrapcheck
	default:
		return fmt.Errorf("unsupported configFormat: %s", configFormat)
	}
}
