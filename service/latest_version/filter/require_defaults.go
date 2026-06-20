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

// Package filter provides filtering for latest_version queries.
package filter

import (
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/service/latest_version/filter/docker"
)

// RequireDefaults are the default values for the Require struct.
// It contains configuration defaults for validating version requirements.
type RequireDefaults struct {
	Docker docker.Defaults `json:"docker,omitzero" yaml:"docker,omitzero"` // Docker image tag requirements.
}

// IsZero implements the yaml.IsZeroer interface.
func (d RequireDefaults) IsZero() bool {
	return d.Docker.IsZero()
}

// Default sets the values of the receiver to their default values.
func (r *RequireDefaults) Default() {
	r.Docker.Default()
}

// SetDefaults assigns defaults to the receiver.
func (r *RequireDefaults) SetDefaults(dflts *RequireDefaults) {
	r.Docker.SetDefaults(&dflts.Docker)
}

// CheckValues validates the fields of the receiver.
func (r *RequireDefaults) CheckValues() error {
	if dockerErr := r.Docker.CheckValues(); dockerErr != nil {
		return &decode.KeyFieldError{
			Key: "docker",
			Err: dockerErr,
		}
	}

	return nil
}
