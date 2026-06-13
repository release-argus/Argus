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

package decode

import (
	"fmt"
	"strings"
)

// FieldError represents an error associated with a specific field.
type FieldError struct {
	Key         string
	Value       string
	Description string
}

// Error implements the [error] interface.
//
// Output formats:
//
// With value:
//
//	KEY: "VALUE" <invalid>
//	KEY: "VALUE" <invalid> (DESCRIPTION)
//
// Without value (required):
//
//	KEY: <required>
//	KEY: <required> (DESCRIPTION)
func (e *FieldError) Error() string {
	var builder strings.Builder

	// Key.
	builder.WriteString(e.Key)
	builder.WriteString(": ")

	// Value.
	if e.Value == "" {
		builder.WriteString("<required>")
	} else {
		fmt.Fprintf(&builder, "%q", e.Value)
		builder.WriteString(" <invalid>")
	}

	// Description.
	if e.Description != "" {
		builder.WriteString(" (")
		builder.WriteString(e.Description)
		builder.WriteString(")")
	}

	return builder.String()
}
