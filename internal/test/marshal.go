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

//go:build unit || integration

package test

import (
	"encoding/json/v2"
	"fmt"

	"github.com/goccy/go-yaml"
)

// Unmarshal implements the format.Unmarshaler interface.
//
// Supported formats: json | yaml
func Unmarshal(format string, data []byte, v any) error {
	switch format {
	case "json":
		if unmarshaler, ok := v.(json.Unmarshaler); ok {
			return unmarshaler.UnmarshalJSON(data)
		}
		return json.Unmarshal(data, v)
	case "yaml":
		if unmarshaler, ok := v.(yaml.BytesUnmarshaler); ok {
			return unmarshaler.UnmarshalYAML(data)
		}
		return yaml.Unmarshal(data, v)
	default:
		return fmt.Errorf("unsupported format: %q", format)
	}
}
