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

// Package polymorphic provides helpers for decoding typed configuration variants.
package polymorphic

import (
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/util"
)

// Inheritable is a typed configuration value that can be decoded and overridden in place.
type Inheritable interface {
	GetType() string

	ApplyOverrides(string, []byte) error
	DecodeSelf(string, []byte) error
}

// ToInheritableMap converts a map of constructors returning concrete types implementing Inheritable
// into a map returning Inheritable.
func ToInheritableMap[T Inheritable](constructors map[string]func() T) map[string]func() Inheritable {
	result := make(map[string]func() Inheritable, len(constructors))
	for k, ctor := range constructors {
		result[k] = func() Inheritable {
			return ctor()
		}
	}
	return result
}

type typeProbe struct {
	Type string `json:"type" yaml:"type"`
}

// ResolveType determines the effective type for format-encoded data using an explicit type field,
// the previous value, or defaultType when neither is present.
func ResolveType(
	format string,
	data []byte,
	previous Inheritable,
	defaultType string,
) (string, error) {
	var probe typeProbe
	// Extract 'type'.
	if !decode.IsNull(data) && len(data) != 0 {
		if err := decode.Unmarshal(format, data, &probe); err != nil {
			return "", err //nolint:wrapcheck
		}
	}

	// Determine 'type' to use.
	var typ string
	switch {
	case probe.Type != "":
		typ = probe.Type
	case previous != nil:
		typ = previous.GetType()
	default:
		typ = defaultType
	}

	return util.EvalEnvVars(typ), nil
}

// Construct instantiates the Inheritable implementation registered for typ.
func Construct(
	typ string,
	constructors map[string]func() Inheritable,
) (Inheritable, error) {
	constructor, ok := constructors[typ]
	if !ok {
		return nil, &InvalidTypeError{
			Key:     "type",
			Value:   typ,
			Allowed: util.SortedKeys(constructors),
		}
	}
	return constructor(), nil
}

// Instantiate instantiates an object from format-encoded data and constructors with a default type.
func Instantiate(
	format string,
	data []byte,
	previous Inheritable,
	defaultType string,
	constructors map[string]func() Inheritable,
) (Inheritable, error) {
	// Look up the constructor for the effective type.
	typ, err := ResolveType(format, data, previous, defaultType)
	if err != nil {
		return nil, err
	}

	// Instantiate the implementation.
	field, err := Construct(typ, constructors)
	if err != nil {
		return nil, err
	}

	return field, nil
}

// ApplyOverrides applies format-encoded overrides to target.
// If the target is nil, a new [Inheritable] is created.
func ApplyOverrides(
	format string,
	data []byte,
	target Inheritable,
	defaultType string,
	constructors map[string]func() Inheritable,
) (Inheritable, error) {
	// No overrides.
	if len(data) == 0 {
		return target, nil
	}
	// Remove.
	if decode.IsNull(data) {
		return nil, nil
	}

	// Check whether the type has changed.
	newType, err := ResolveType(
		format, data,
		target,
		defaultType,
	)
	if err != nil {
		return nil, err
	}

	// If type changed, replace the target.
	if target == nil || newType != target.GetType() {
		target, err = Instantiate(
			format, data,
			nil,
			defaultType,
			constructors,
		)
		if err != nil {
			return nil, err
		}
	}

	// Otherwise, patch the existing target.
	if err := decode.Unmarshal(
		format, data,
		target,
	); err != nil {
		return nil, err //nolint:wrapcheck
	}

	return target, nil
}
