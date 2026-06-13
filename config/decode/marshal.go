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

// Package decode provides shared JSON and YAML marshaling helpers and config error types.
package decode

import (
	"bytes"
	"encoding/json/v2"
	"fmt"

	"github.com/goccy/go-yaml"
)

// nullBytes is the byte representation of a JSON null literal.
var nullBytes = []byte("null")

// IsNull reports whether data represents a JSON null value (case-insensitive, ignoring surrounding whitespace).
func IsNull(data []byte) bool {
	return bytes.EqualFold(bytes.TrimSpace(data), nullBytes)
}

// UnsupportedFormatError represents an error for an unsupported [Marshal]/[Unmarshal] format.
type UnsupportedFormatError struct {
	Format string
}

// Error implements the [error] interface.
func (e *UnsupportedFormatError) Error() string {
	return fmt.Sprintf("unsupported format: %q", e.Format)
}

// Unmarshal decodes data in the given format into v.
//
// Supported formats: json | yaml.
func Unmarshal(format string, data []byte, v any) error {
	if len(data) == 0 && format == "json" {
		data = []byte("{}")
	}

	switch format {
	case "json":
		if unmarshaler, ok := v.(json.Unmarshaler); ok {
			return unmarshaler.UnmarshalJSON(data) //nolint:wrapcheck
		}
		return json.Unmarshal(data, v) //nolint:wrapcheck
	case "yaml":
		if unmarshaler, ok := v.(yaml.BytesUnmarshaler); ok {
			return unmarshaler.UnmarshalYAML(data) //nolint:wrapcheck
		}
		if len(data) == 0 {
			return nil
		}
		return yaml.Unmarshal(data, v) //nolint:wrapcheck
	default:
		return &UnsupportedFormatError{Format: format} //nolint:wrapcheck
	}
}

// Marshal encodes v in the given format.
func Marshal(format string, m any) ([]byte, error) {
	switch format {
	case "json":
		return json.Marshal(m, JSONMarshalOpts...) //nolint:wrapcheck
	case "yaml":
		return yaml.MarshalWithOptions(m, YAMLMarshalOpts...) //nolint:wrapcheck
	default:
		return nil, &UnsupportedFormatError{Format: format} //nolint:wrapcheck
	}
}
