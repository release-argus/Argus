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

package docker

import (
	"fmt"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/util/polymorphic"
)

// Decode creates and returns a new [Registry] from format-encoded data.
func Decode(
	format string,
	data []byte,
	defaults *Defaults,
) (Registry, error) {
	// Remove.
	if decode.IsNull(data) {
		return nil, nil
	}
	if data == nil {
		data = []byte("{}")
	}

	// Create.
	fieldInheritable, err := polymorphic.Instantiate(
		format, data,
		nil,
		defaults.GetType(),
		RegistryMapInheritable,
	)
	if err != nil {
		return nil, &decode.KeyFieldError{
			Key: "docker",
			Err: err,
		}
	}

	// Assert back to Registry.
	field, ok := fieldInheritable.(Registry)
	if !ok {
		err := fmt.Errorf("expected Registry, got %T", field)
		return nil, &decode.KeyFieldError{
			Key: "docker",
			Err: err,
		}
	}

	field.SetDefaults(field.GetType(), defaults)
	if err := field.DecodeSelf(format, data); err != nil {
		return nil, &decode.KeyFieldError{
			Key: "docker",
			Err: err,
		}
	}

	return field, nil
}

// ApplyOverrides applies format-encoded overrides to target.
// If the target is nil, a new [Registry] is created.
func ApplyOverrides(
	format string,
	data []byte,
	target Registry,
	defaults *Defaults,
) (Registry, error) {
	// No overrides.
	if len(data) == 0 {
		return target, nil
	}
	// Remove.
	if decode.IsNull(data) {
		return nil, nil
	}

	var (
		previousAuth   RegistryAuth
		previousDetail ContainerDetail
	)
	// Clear query token, and store previous state to inherit it from if query-token would retrieve the same token.
	if target != nil {
		target.GetAuth().SetQueryToken("", time.Time{})
		previousAuth = target.GetAuth().Copy()
		previousDetail = target.Detail()
	}

	// Apply overrides.
	newTarget, err := polymorphic.ApplyOverrides(
		format, data,
		target,
		defaults.GetType(),
		RegistryMapInheritable,
	)
	if err != nil {
		return nil, &decode.KeyFieldError{
			Key: "docker",
			Err: err,
		}
	}

	// Assert back to Registry.
	field, ok := newTarget.(Registry)
	if !ok {
		return nil, &decode.KeyFieldError{
			Key: "docker",
			Err: fmt.Errorf("expected Registry, got %T", field),
		}
	}
	field.SetDefaults(field.GetType(), defaults)
	field.GetAuth().Inherit(
		previousAuth,
		field.Detail(),
		previousDetail,
	)

	return field, nil
}
