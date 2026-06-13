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

package filter

import (
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/service/latest_version/filter/docker"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util/polymorphic"
)

// Decode creates and returns a new [Require] from format-encoded data.
func Decode(
	format string,
	data []byte,
	svcStatus *status.Status,
	defaults *RequireDefaults,
) (*Require, error) {
	// Remove.
	if decode.IsNull(data) {
		return nil, nil
	}

	field := Require{
		Status:   svcStatus,
		defaults: defaults,
	}

	// Polymorphic fields.
	//   Docker.
	dockerRaw, err := polymorphic.Extract(format, data, "docker")
	if err != nil {
		return nil, &decode.KeyFieldError{
			Key: "require",
			Err: err,
		}
	}
	field.Docker, err = docker.Decode(
		format, dockerRaw,
		&defaults.Docker,
	)
	if err != nil {
		return nil, &decode.KeyFieldError{
			Key: "require",
			Err: err,
		}
	}

	// Static fields.
	if err := decode.Unmarshal(format, data, &field); err != nil {
		return nil, &decode.KeyFieldError{
			Key: "require",
			Err: err,
		}
	}

	return &field, nil
}

// ApplyOverrides applies format-encoded overrides to a [Require] object.
// If the target is nil, a new [Require] is created.
func (r *Require) ApplyOverrides(
	format string,
	data []byte,
	svcStatus *status.Status,
	defaults *RequireDefaults,
) (*Require, error) {
	// No overrides.
	if len(data) == 0 {
		return r, nil
	}
	// Remove.
	if decode.IsNull(data) {
		return nil, nil
	}
	// New.
	if r == nil {
		return Decode(
			format, data,
			svcStatus,
			defaults,
		)
	}

	// Polymorphic fields.
	dockerRaw, err := polymorphic.Extract(format, data, "docker")
	if err != nil {
		return nil, &decode.KeyFieldError{
			Key: "require",
			Err: err,
		}
	}
	if dockerRaw != nil {
		var err error
		// Docker creation/overrides.
		r.Docker, err = docker.ApplyOverrides(
			format, dockerRaw,
			r.Docker,
			&defaults.Docker,
		)
		if err != nil {
			return nil, &decode.KeyFieldError{
				Key: "require",
				Err: err,
			}
		}
	}

	// Static fields.
	if err := decode.Unmarshal(format, data, r); err != nil {
		return nil, &decode.KeyFieldError{
			Key: "require",
			Err: err,
		}
	}

	return r, nil
}

// DecodeDefaults creates and returns new [RequireDefaults] from format-encoded data.
func DecodeDefaults(format string, data []byte) (*RequireDefaults, error) {
	// Remove.
	if decode.IsNull(data) {
		return nil, nil
	}

	var field RequireDefaults

	dockerRaw, err := polymorphic.Extract(format, data, "docker")
	if err != nil {
		return nil, &decode.KeyFieldError{
			Key: "require",
			Err: err,
		}
	}
	if len(dockerRaw) != 0 {
		dockerDefaults, err := docker.DecodeDefaults(format, dockerRaw, nil)
		if err != nil {
			return nil, &decode.KeyFieldError{
				Key: "require",
				Err: err,
			}
		}
		field.Docker = *dockerDefaults
	}

	return &field, nil
}
