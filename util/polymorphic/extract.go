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
	"fmt"

	"github.com/release-argus/Argus/config/decode"
)

type dataExtract map[string]any

// marshalExtractSubtree serialises an extracted subtree (overridable for tests).
var marshalExtractSubtree = decode.Marshal

// ExtractError reports a failure while extracting a configuration subtree by key.
type ExtractError struct {
	Key string
	Err error
}

// Error implements the [error] interface.
func (e *ExtractError) Error() string {
	return fmt.Sprintf(
		"extract %q: %v",
		e.Key, e.Err,
	)
}

// Unwrap returns the underlying error.
func (e *ExtractError) Unwrap() error {
	return e.Err
}

// Extract extracts a subtree from the provided configuration data by key and serialises it in the specified format.
func Extract(
	format string,
	data []byte,
	key string,
) ([]byte, error) {
	// No data.
	if len(data) == 0 {
		return nil, nil
	}

	// Unmarshal into a map.
	var m dataExtract
	if err := decode.Unmarshal(format, data, &m); err != nil {
		return nil, &ExtractError{
			Key: key,
			Err: err,
		}
	}

	// Extract the subtree.
	if n, ok := m[key]; ok {
		b, err := marshalExtractSubtree(format, n)
		if err != nil {
			return nil, &decode.KeyFieldError{
				Key: key,
				Err: err,
			}
		}

		return b, nil
	}

	return nil, nil
}
