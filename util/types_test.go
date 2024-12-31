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

//go:build unit

package util

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Field1 string `json:"field1" yaml:"field1"`
	Field2 int    `json:"field2" yaml:"field2"`
}
type ConfigCustom struct {
	Field1 string `json:"field1" yaml:"field1"`
	Field2 int    `json:"field2" yaml:"field2"`
	Field3 string `json:"-" yaml:"-"`
}

func (c *ConfigCustom) UnmarshalYAML(value *yaml.Node) error {
	// Alias to avoid recursion
	type Alias ConfigCustom
	aux := &struct {
		*Alias `yaml:",inline"`
	}{
		Alias: (*Alias)(c),
	}
	if err := value.Decode(aux); err != nil {
		return err //nolint:wrapcheck
	}
	c.Field3 = "custom"
	return nil
}
func (c *ConfigCustom) UnmarshalJSON(data []byte) error {
	// Alias to avoid recursion
	type Alias ConfigCustom
	aux := &struct {
		*Alias `json:",inline"`
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err //nolint:wrapcheck
	}
	c.Field3 = "custom"
	return nil
}

func TestUnmarshalConfig(t *testing.T) {
	type wants struct {
		config interface{}
		err    bool
	}

	// GIVEN different config formats and data
	tests := map[string]struct {
		configFormat string
		configData   interface{}
		wants        wants
	}{
		"JSON, Invalid": {
			configFormat: "json",
			configData:   `{"field1": "value1`,
			wants: wants{
				config: Config{},
				err:    true},
		},
		"JSON, Valid": {
			configFormat: "json",
			configData:   `{"field1": "value1", "field2": 2}`,
			wants: wants{
				config: Config{Field1: "value1", Field2: 2},
				err:    false},
		},
		"JSON, Invalid CustomUnmarshal": {
			configFormat: "json",
			configData:   `{"field1": []}`,
			wants: wants{
				config: ConfigCustom{},
				err:    true},
		},
		"JSON, valid CustomUnmarshal": {
			configFormat: "json",
			configData:   `{"field1": "value1", "field2": 2}`,
			wants: wants{
				config: ConfigCustom{Field1: "value1", Field2: 2},
				err:    false},
		},
		"YAML, Invalid": {
			configFormat: "yaml",
			configData:   "field1: [value1]",
			wants: wants{
				config: Config{},
				err:    true},
		},
		"YAML, valid": {
			configFormat: "yaml",
			configData:   "field1: value1\nfield2: 2",
			wants: wants{
				config: Config{Field1: "value1", Field2: 2},
				err:    false},
		},
		"YAML, Invalid CustomUnmarshal - tabs": {
			configFormat: "yaml",
			configData:   `	field1: value1`,
			wants: wants{
				config: ConfigCustom{},
				err:    true},
		},
		"YAML, Invalid CustomUnmarshal": {
			configFormat: "yaml",
			configData:   "field1: []",
			wants: wants{
				config: ConfigCustom{},
				err:    true},
		},
		"YAML, valid CustomUnmarshal": {
			configFormat: "yaml",
			configData:   "field1: value1\nfield2: 2",
			wants: wants{
				config: ConfigCustom{Field1: "value1", Field2: 2},
				err:    false},
		},
		"unsupported config format": {
			configFormat: "xml",
			configData:   `<config><field1>value1</field1><field2>2</field2></config>`,
			wants: wants{
				config: Config{},
				err:    true},
		},
		"unsupported config data type": {
			configFormat: "json",
			configData:   12345,
			wants: wants{
				config: Config{},
				err:    true},
		},
		"YAML, *yaml.Node": {
			configFormat: "yaml",
			configData: func() *yaml.Node {
				var node yaml.Node
				_ = yaml.Unmarshal([]byte("field1: value1\nfield2: 2"), &node)
				return &node
			}(),
			wants: wants{
				config: Config{Field1: "value1", Field2: 2},
				err:    false},
		},
		"YAML, []byte": {
			configFormat: "yaml",
			configData:   []byte("field1: value1\nfield2: 2"),
			wants: wants{
				config: Config{Field1: "value1", Field2: 2},
				err:    false},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var got interface{}
			if _, ok := tc.wants.config.(Config); ok {
				got = &Config{}
			} else {
				got = &ConfigCustom{}
			}

			// WHEN UnmarshalConfig is called
			err := UnmarshalConfig(tc.configFormat, tc.configData, got)

			// THEN an error is returned if expected
			if (err != nil) != tc.wants.err {
				t.Errorf("UnmarshalConfig() error = %v, wantErr %v",
					err, tc.wants.err)
				return
			}
			// AND the config is unmarshalled as expected
			gotStr := ToYAMLString(got, "")
			wantStr := ToYAMLString(tc.wants.config, "")
			if !tc.wants.err && gotStr != wantStr {
				t.Fatalf("UnmarshalConfig() = %v, want %v",
					gotStr, wantStr)
			}
			// AND the custom Unmarshal is called when the struct implements it
			if _, ok := tc.wants.config.(ConfigCustom); ok {
				if got.(*ConfigCustom).Field3 != "custom" {
					if tc.wants.err {
						return // errored, so ignore the field3 check
					}
					t.Errorf("UnmarshalConfig() didn't call the struct-specific Unmarshal\nfield3 = %v, want %v",
						got.(*ConfigCustom).Field3, "custom")
				}
			}
		})
	}
}
