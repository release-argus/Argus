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
	"strings"
)

// InvalidTypeError is returned when a value outside the Allowed values is given for the Key.
type InvalidTypeError struct {
	Key     string
	Value   string
	Allowed []string
}

// Error implements the [error] interface.
func (e InvalidTypeError) Error() string {
	valueMsg := "<required>"
	if e.Value != "" {
		valueMsg = fmt.Sprintf("%q <invalid>", e.Value)
	}

	return fmt.Sprintf(
		"%s: %s (supported values = ['%s'])",
		e.Key, valueMsg, strings.Join(e.Allowed, "', '"),
	)
}
