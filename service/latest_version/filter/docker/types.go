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
	"github.com/release-argus/Argus/util/polymorphic"
)

// PossibleTypes for the docker Lookup.
var PossibleTypes = []string{
	"ghcr",
	"hub",
	"quay",
}

// RegistryMap maps a registry type to a Registry constructor.
var RegistryMap = map[string]func() Registry{
	"ghcr": func() Registry {
		return &GHCRRegistry{
			CommonRegistry: CommonRegistry{
				Auth: &GHCRAuth{},
			},
		}
	},
	"hub": func() Registry {
		return &HubRegistry{
			CommonRegistry: CommonRegistry{
				Auth: &HubAuth{},
			},
		}
	},
	"quay": func() Registry {
		return &QuayRegistry{
			CommonRegistry: CommonRegistry{Auth: &QuayAuth{}},
		}
	},
}

// RegistryMapInheritable is [RegistryMap] wrapped for polymorphic inheritance decoding.
var RegistryMapInheritable = polymorphic.ToInheritableMap(RegistryMap)

// RegistryDefaultsMap maps a registry type to a RegistryDefaults constructor.
var RegistryDefaultsMap = map[string]func() RegistryDefaults{
	"ghcr": func() RegistryDefaults {
		return &GHCRRegistryDefaults{
			CommonRegistryDefaults: CommonRegistryDefaults{
				Auth: &GHCRAuthDefaults{},
			},
		}
	},
	"hub": func() RegistryDefaults {
		return &HubRegistryDefaults{
			CommonRegistryDefaults: CommonRegistryDefaults{
				Auth: &HubAuthDefaults{},
			},
		}
	},
	"quay": func() RegistryDefaults {
		return &QuayRegistryDefaults{
			CommonRegistryDefaults: CommonRegistryDefaults{
				Auth: &QuayAuthDefaults{},
			},
		}
	},
}
