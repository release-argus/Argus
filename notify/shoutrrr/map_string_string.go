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

package shoutrrr

import (
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/util"
)

// MapStringStringOmitNull is a map[string]string that ignores null values when unmarshaling.
// Any key with a null value in YAML/JSON will be omitted from the map.
type MapStringStringOmitNull map[string]string

// UnmarshalYAML implements yaml.Unmarshaler for MapStringStringOmitNull.
//
// It decodes a YAML mapping into the map, ignoring keys whose values are null.
// Null values are omitted rather than converted to empty strings.
func (m *MapStringStringOmitNull) UnmarshalYAML(data []byte) error {
	return m.unmarshal("yaml", data)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
//
// It decodes a JSON object into the map, ignoring keys whose values are null.
// Null values are omitted rather than converted to empty strings.
func (m *MapStringStringOmitNull) UnmarshalJSON(data []byte) error {
	return m.unmarshal("json", data)
}

// unmarshal implements the format.Unmarshaler interface.
func (m *MapStringStringOmitNull) unmarshal(format string, data []byte) error {
	temp := map[string]*string{}
	if err := decode.Unmarshal(format, data, &temp); err != nil {
		// nolint:wrapcheck
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

// Copy returns a deep copy of the receiver.
func (m MapStringStringOmitNull) Copy() *MapStringStringOmitNull {
	field := MapStringStringOmitNull(util.CopyMap(m))
	return &field
}
