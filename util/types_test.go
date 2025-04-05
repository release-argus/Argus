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

//go:build unit

package util

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/release-argus/Argus/test"
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
	// Alias to avoid recursion.
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
	// Alias to avoid recursion.
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
		config   interface{}
		errRegex string
	}

	// GIVEN different config formats and data.
	tests := map[string]struct {
		configFormat string
		configData   interface{}
		wants        wants
	}{
		"JSON, Invalid": {
			configFormat: "json",
			configData:   `{"field1": "value1`,
			wants: wants{
				config:   Config{},
				errRegex: `^unexpected end of JSON input$`},
		},
		"JSON, Valid": {
			configFormat: "json",
			configData:   `{"field1": "value1", "field2": 2}`,
			wants: wants{
				config:   Config{Field1: "value1", Field2: 2},
				errRegex: `^$`},
		},
		"JSON, Invalid CustomUnmarshal": {
			configFormat: "json",
			configData:   `{"field1": []}`,
			wants: wants{
				config:   ConfigCustom{},
				errRegex: `^json: cannot unmarshal array into Go struct field (\.Alias)?\.field1 of type string$`},
		},
		"JSON, valid CustomUnmarshal": {
			configFormat: "json",
			configData:   `{"field1": "value1", "field2": 2}`,
			wants: wants{
				config:   ConfigCustom{Field1: "value1", Field2: 2},
				errRegex: `^$`},
		},
		"YAML, Invalid": {
			configFormat: "yaml",
			configData:   "field1: [value1]",
			wants: wants{
				config: Config{},
				errRegex: test.TrimYAML(`
					^yaml: unmarshal errors:
						line 1: cannot unmarshal !!seq into string$`)},
		},
		"YAML, valid": {
			configFormat: "yaml",
			configData:   "field1: value1\nfield2: 2",
			wants: wants{
				config:   Config{Field1: "value1", Field2: 2},
				errRegex: `^$`},
		},
		"YAML, Invalid CustomUnmarshal - tabs": {
			configFormat: "yaml",
			configData:   `	field1: value1`,
			wants: wants{
				config:   ConfigCustom{},
				errRegex: `^yaml: found character that cannot start any token$`},
		},
		"YAML, Invalid CustomUnmarshal": {
			configFormat: "yaml",
			configData:   "field1: []",
			wants: wants{
				config: ConfigCustom{},
				errRegex: test.TrimYAML(`
					^yaml: unmarshal errors:
						line 1: cannot unmarshal !!seq into string$`)},
		},
		"YAML, valid CustomUnmarshal": {
			configFormat: "yaml",
			configData:   "field1: value1\nfield2: 2",
			wants: wants{
				config:   ConfigCustom{Field1: "value1", Field2: 2},
				errRegex: `^$`},
		},
		"unsupported config format": {
			configFormat: "xml",
			configData:   `<config><field1>value1</field1><field2>2</field2></config>`,
			wants: wants{
				config:   Config{},
				errRegex: `^unsupported configFormat: xml$`},
		},
		"unsupported config data type": {
			configFormat: "json",
			configData:   12345,
			wants: wants{
				config:   Config{},
				errRegex: `^unsupported configData type: int$`},
		},
		"YAML, *yaml.Node": {
			configFormat: "yaml",
			configData: func() *yaml.Node {
				var node yaml.Node
				_ = yaml.Unmarshal([]byte("field1: value1\nfield2: 2"), &node)
				return &node
			}(),
			wants: wants{
				config:   Config{Field1: "value1", Field2: 2},
				errRegex: `^$`},
		},
		"YAML, []byte": {
			configFormat: "yaml",
			configData:   []byte("field1: value1\nfield2: 2"),
			wants: wants{
				config:   Config{Field1: "value1", Field2: 2},
				errRegex: `^$`},
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

			// WHEN UnmarshalConfig is called.
			err := UnmarshalConfig(tc.configFormat, tc.configData, got)

			// THEN an error is returned if expected.
			e := ErrorToString(err)
			if !RegexCheck(tc.wants.errRegex, e) {
				t.Errorf("%s\nerror mismatch:\nwant: %q\ngot:  %q",
					packageName, tc.wants.errRegex, e)
			}
			if err != nil {
				return
			}
			// AND the config is unmarshalled as expected.
			gotStr := ToYAMLString(got, "")
			wantStr := ToYAMLString(tc.wants.config, "")
			if gotStr != wantStr {
				t.Fatalf("%s\nstringified mismatch\nwant: %v\ngot:  %v",
					packageName, wantStr, gotStr)
			}
			// AND the custom Unmarshal is called when the struct implements it.
			if _, ok := tc.wants.config.(ConfigCustom); ok {
				wantField3 := "custom"
				if got.(*ConfigCustom).Field3 != wantField3 {
					t.Errorf("%s\nstruct-specific Unmarshal not called for field3\nwant: %q\ngot:  %q",
						packageName, wantField3, got.(*ConfigCustom).Field3)
				}
			}
		})
	}
}

func TestFormatUnmarshalError(t *testing.T) {
	// GIVEN data to unmarshal in a given format, or an error from an unmarshal.
	tests := map[string]struct {
		format string
		data   string
		err    string
		want   string
	}{
		"yaml, invalid type": {
			format: "yaml",
			data:   "field1: [invalid]",
			want:   "line 1: cannot unmarshal !!seq into string",
		},
		"yaml, missing field": {
			format: "yaml",
			data:   "field2: 2",
			want:   "",
		},
		"yaml, invalid syntax": {
			format: "yaml",
			data:   "field1: value1\nfield2: 2\ninvalid",
			want:   "line 3: could not find expected ':'",
		},
		"json, invalid type": {
			format: "json",
			data:   `{"field1": ["invalid"]}`,
			want:   "cannot unmarshal array into Go struct field Config.field1 of type string",
		},
		"json, missing field": {
			format: "json",
			data:   `{"field2": 2}`,
			want:   "",
		},
		"json, invalid syntax": {
			format: "json",
			data:   `{"field1": "value1"`,
			want:   "unexpected end of JSON input",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var config Config
			var err error
			if tc.data != "" {
				switch tc.format {
				case "yaml":
					err = yaml.Unmarshal([]byte(tc.data), &config)
				case "json":
					err = json.Unmarshal([]byte(tc.data), &config)
				default:
					t.Fatalf("%s - unsupported format: %s",
						packageName, tc.format)
				}
			} else {
				err = errors.New(tc.err)
			}

			// WHEN the error is formatted.
			result := FormatUnmarshalError(tc.format, err)

			// THEN the error message is returned as expected.
			if result != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, result)
			}
		})
	}
}
