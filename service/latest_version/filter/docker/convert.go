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
)

// Deprecated: oldDockerRegistryDefaults captures defaults for a Docker registry (<=0.29.4).
type oldDockerRegistryDefaults struct {
	Username string `json:"username" yaml:"username"`
	Token    string `json:"token" yaml:"token"`
}

// IsZero implements the yaml.IsZeroer interface.
func (o *oldDockerRegistryDefaults) IsZero() bool {
	return o == nil || (o.Username == "" && o.Token == "")
}

// Deprecated: oldDockerDefaults captures all old Docker registry defaults (<=0.29.4).
type oldDockerDefaults struct {
	RegistryGHCR *oldDockerRegistryDefaults `json:"ghcr,omitempty" yaml:"ghcr,omitempty"`
	RegistryHub  *oldDockerRegistryDefaults `json:"hub,omitempty" yaml:"hub,omitempty"`
	RegistryQuay *oldDockerRegistryDefaults `json:"quay,omitempty" yaml:"quay,omitempty"`
}

// IsZero implements the yaml.IsZeroer interface.
func (o *oldDockerDefaults) IsZero() bool {
	return o == nil || (o.RegistryGHCR.IsZero() && o.RegistryHub.IsZero() && o.RegistryQuay.IsZero())
}

// convertOldDefaults parses the old Docker defaults and converts them to the new format.
func convertOldDefaults(format string, data []byte, newFormat *Defaults) {
	oldData := oldDockerDefaults{}
	err := decode.Unmarshal(format, data, &oldData)
	if err != nil || oldData.IsZero() {
		return
	}

	// GitHub Packages only supports authentication using a personal access token (classic).
	oldGHCR := oldData.RegistryGHCR
	newAuthGHCR := newFormat.Registry.GHCR.Auth.(*GHCRAuthDefaults)
	if !oldGHCR.IsZero() {
		newAuthGHCR.Token = oldGHCR.Token
	}

	// Docker Hub: Personal access token.
	oldHub := oldData.RegistryHub
	newAuthHub := newFormat.Registry.Hub.Auth.(*HubAuthDefaults)
	if !oldHub.IsZero() {
		newAuthHub.Username = oldHub.Username
		newAuthHub.Token = oldHub.Token
	}

	// Quay: Access Token - currently deprecated and should not be used. Robot Accounts are their replacement.
	oldQuay := oldData.RegistryQuay
	newAuthQuay := newFormat.Registry.Quay.Auth.(*QuayAuthDefaults)
	if !oldQuay.IsZero() {
		newAuthQuay.Token = oldQuay.Token
	}
}
