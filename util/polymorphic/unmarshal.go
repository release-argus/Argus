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

package polymorphic

import (
	"github.com/release-argus/Argus/config/decode"
)

// Unmarshal extracts key from format-encoded data and decodes the subtree into v.
func Unmarshal(
	format string,
	data []byte,
	key string,
	v any,
) error {
	if len(data) == 0 {
		return nil
	}

	// Extract the key value.
	keyVal, err := Extract(format, data, key)
	if err != nil {
		return err
	}
	// No value.
	if keyVal == nil {
		return nil
	}

	// Unmarshal the key value into the interface.
	if err := decode.Unmarshal(format, keyVal, v); err != nil {
		return &decode.KeyFieldError{
			Key: key,
			Err: err,
		}
	}

	return nil
}
