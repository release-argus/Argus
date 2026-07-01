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
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/polymorphic"
)

// #########
// # TYPES #
// #########

// Defaults are the default values for DockerCheck.
type Defaults struct {
	Type                    string                          `json:"type,omitempty" yaml:"type,omitempty"` // Type of the Docker registry.
	ContainerDetailDefaults `json:",inline" yaml:",inline"` // Default Tag template.
	Registry                RegistryDefaultsSet             `json:"registry,omitzero" yaml:"registry,omitzero"` // Registry-specific defaults.
	Defaults                *Defaults                       `json:"-" yaml:"-"`                                 // Defaults to fall back on.
}

// RegistryDefaultsSet holds per-registry default configuration.
type RegistryDefaultsSet struct {
	ECR  *ECRRegistryDefaults  `json:"ecr,omitzero" yaml:"ecr,omitzero"`   // Amazon ECR Public Gallery (anonymous: no serialisable config).
	GHCR *GHCRRegistryDefaults `json:"ghcr,omitzero" yaml:"ghcr,omitzero"` // GitHub Container Registry.
	Hub  *HubRegistryDefaults  `json:"hub,omitzero" yaml:"hub,omitzero"`   // Docker Hub.
	Quay *QuayRegistryDefaults `json:"quay,omitzero" yaml:"quay,omitzero"` // Quay.
}

// ############
// # DECODING #
// ############

// DecodeDefaults creates and returns new [Defaults] from format-encoded data.
func DecodeDefaults(
	format string,
	data []byte,
	defaults *Defaults,
) (*Defaults, error) {
	var field Defaults
	field.initRegistries()

	// Unmarshal static fields.
	if err := decode.Unmarshal(format, data, &field); err != nil {
		return nil, &decode.KeyFieldError{
			Key: "docker",
			Err: err,
		}
	}

	if field.Registry.IsZero() {
		// Deprecated: Convert '(ghcr|hub|quay)' -> 'registry.(ghcr|hub|quay)'.
		convertOldDefaults(format, data, &field)
	}

	// Defaults.
	field.SetDefaults(defaults)

	return &field, nil
}

// #########
// # STATE #
// #########

// IsZero implements the yaml.IsZeroer interface.
func (d *Defaults) IsZero() bool {
	return d.Type == "" &&
		d.ContainerDetailDefaults.IsZero() &&
		d.Registry.IsZero()
}

// IsZero implements the yaml.IsZeroer interface.
func (r *RegistryDefaultsSet) IsZero() bool {
	if r == nil {
		return true
	}

	return r.ECR.IsZero() &&
		r.GHCR.IsZero() &&
		r.Hub.IsZero() &&
		r.Quay.IsZero()
}

// #############
// # STRINGIFY #
// #############

// String returns a YAML string representation of the receiver.
func (d *Defaults) String(prefix string) string {
	return decode.ToYAMLString(d, prefix)
}

// ############
// # DEFAULTS #
// ############

// Default sets the values of the receiver to their default values.
func (d *Defaults) Default() {
	d.Type = "hub"
	d.initRegistries()
	d.ContainerDetailDefaults.Default()
}

// SetDefaults assigns defaults to the receiver.
func (d *Defaults) SetDefaults(defaults *Defaults) {
	d.initRegistries()
	if defaults == nil {
		return
	}
	d.Defaults = defaults
	d.ContainerDetailDefaults.Defaults = &defaults.ContainerDetailDefaults

	defaults.initRegistries()
	setRegistryDefaults(d.Registry.ECR, defaults.Registry.ECR)
	setRegistryDefaults(d.Registry.GHCR, defaults.Registry.GHCR)
	setRegistryDefaults(d.Registry.Hub, defaults.Registry.Hub)
	setRegistryDefaults(d.Registry.Quay, defaults.Registry.Quay)
}

// ##########
// # VALUES #
// ##########

// GetType returns the default registry type, falling back through the defaults chain.
func (d *Defaults) GetType() string {
	for dflt := d; dflt != nil; dflt = dflt.Defaults {
		if dflt.Type != "" {
			return dflt.Type
		}
	}
	return ""
}

// ##############
// # VALIDATION #
// ##############

// CheckValues validates the fields of the receiver.
func (d *Defaults) CheckValues() error {
	if d == nil {
		return nil
	}

	// Type.
	if d.Type != "" && !util.Contains(PossibleTypes, d.Type) {
		return polymorphic.InvalidTypeError{
			Key:     "type",
			Value:   d.Type,
			Allowed: PossibleTypes,
		}
	}

	return nil
}

// #############
// # UTILITIES #
// #############

// initRegistries ensures each registry-specific defaults slot is non-nil.
func (d *Defaults) initRegistries() {
	if d.Registry.ECR == nil {
		d.Registry.ECR = RegistryDefaultsMap["ecr"]().(*ECRRegistryDefaults)
	}
	if d.Registry.GHCR == nil {
		d.Registry.GHCR = RegistryDefaultsMap["ghcr"]().(*GHCRRegistryDefaults)
	}
	if d.Registry.Hub == nil {
		d.Registry.Hub = RegistryDefaultsMap["hub"]().(*HubRegistryDefaults)
	}
	if d.Registry.Quay == nil {
		d.Registry.Quay = RegistryDefaultsMap["quay"]().(*QuayRegistryDefaults)
	}
}

// getRegistryDefaults returns the defaults for dType from defaults, or nil if unset.
func getRegistryDefaults(dType string, defaults *Defaults) RegistryDefaults {
	switch dType {
	case "ecr":
		if defaults.Registry.ECR == nil {
			return nil
		}
		return defaults.Registry.ECR
	case "ghcr":
		if defaults.Registry.GHCR == nil {
			return nil
		}
		return defaults.Registry.GHCR
	case "hub":
		if defaults.Registry.Hub == nil {
			return nil
		}
		return defaults.Registry.Hub
	case "quay":
		if defaults.Registry.Quay == nil {
			return nil
		}
		return defaults.Registry.Quay
	}
	return nil
}

// setRegistryDefaults links a registry's auth defaults to its parent defaults.
func setRegistryDefaults(registry, defaultRegistry RegistryDefaults) {
	if registry == nil || defaultRegistry == nil {
		return
	}

	registry.GetAuth().SetDefaults(defaultRegistry.GetAuth())
}
