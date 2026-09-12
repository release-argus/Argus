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
	"strconv"
	"strings"
)

// ErrInvalidValue is returned when a value outside the Allowed values is given for the Key.
type ErrInvalidValue struct {
	Key     string
	Value   string
	Allowed []string
}

// Error implements the [error] interface.
//
// Output formats:
//
// With value:
//
//	KEY: "VALUE" <invalid> (supported values = ['A', 'B', 'C'])
//
// Without value (required):
//
//	KEY: <required> (supported values = ['A', 'B', 'C'])
func (e ErrInvalidValue) Error() string {
	valueMsg := "<required>"
	if e.Value != "" {
		valueMsg = fmt.Sprintf("%q <invalid>", e.Value)
	}

	return invalidMessage(e.Key, valueMsg, e.Allowed)
}

// ErrInvalidValues is returned when values outside the Allowed values are given
// for the list-valued Key.
type ErrInvalidValues struct {
	Key     string
	Values  []string
	Allowed []string
}

// maxNamedValues caps how many values one error names, so a large body cannot
// amplify into a far larger error response.
const maxNamedValues = 5

// Error implements the [error] interface.
//
// Output formats:
//
// With values:
//
//	KEY: "X", "Y" <invalid> (supported values = ['A', 'B', 'C'])
//
// With more than [maxNamedValues] values:
//
//	KEY: "V", "W", "X", "Y", "Z", and 2 more <invalid> (supported values = ['A', 'B', 'C'])
//
// Without values (required):
//
//	KEY: <required> (supported values = ['A', 'B', 'C'])
func (e ErrInvalidValues) Error() string {
	valueMsg := "<required>"
	if named := min(len(e.Values), maxNamedValues); named > 0 {
		parts := make([]string, 0, named+1)
		for _, value := range e.Values[:named] {
			parts = append(parts, strconv.Quote(value))
		}
		if extra := len(e.Values) - named; extra > 0 {
			parts = append(parts, fmt.Sprintf("and %d more", extra))
		}
		valueMsg = strings.Join(parts, ", ") + " <invalid>"
	}

	return invalidMessage(e.Key, valueMsg, e.Allowed)
}

// invalidMessage renders the shared grammar of the invalid-value errors.
func invalidMessage(key, valueMsg string, allowed []string) string {
	return fmt.Sprintf(
		"%s: %s (supported values = ['%s'])",
		key, valueMsg, strings.Join(allowed, "', '"),
	)
}
