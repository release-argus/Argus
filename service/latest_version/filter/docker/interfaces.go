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

// Package docker provides Docker registry require filters for latest_version lookups.
package docker

import (
	"net/http"
	"time"

	"github.com/goccy/go-yaml"

	"github.com/release-argus/Argus/util/polymorphic"
)

// RegistryAuthDefaults is the shared auth configuration surface for a registry
// checker.
type RegistryAuthDefaults interface {
	yaml.IsZeroer

	// GetTokenSelf returns the configured credential token on this auth defaults node.
	GetTokenSelf() string

	// GetQueryTokenSelf returns the cached query token and its expiry on this node.
	GetQueryTokenSelf() (string, time.Time)

	// SetQueryToken stores a cached query token and expiry on this node.
	SetQueryToken(queryToken string, validUntil time.Time)

	// Defaults returns the next link in the auth defaults chain.
	Defaults() RegistryAuthDefaults

	// SetDefaults assigns defaults to the receiver.
	SetDefaults(defaults RegistryAuthDefaults)
}

// RegistryAuth is the runtime auth behavior for a registry checker, including
// token resolution and validation.
type RegistryAuth interface {
	RegistryAuthDefaults

	// GetQueryToken returns the bearer token to use for registry API requests.
	GetQueryToken(detail ContainerDetail) (string, error)

	// Inherit copies the token data from src to dst if the ContainerDetail's match.
	Inherit(from RegistryAuth, src, dst ContainerDetail)

	// Copy returns a deep copy of the receiver.
	Copy() RegistryAuth

	// CheckValues validates the receiver.
	CheckValues() error
}

// RegistryDefaults defines the methods that defaults for a Docker registry checker must implement.
type RegistryDefaults interface {
	yaml.IsZeroer

	// GetType returns the fixed registry kind (e.g. "hub", "ghcr", "quay").
	GetType() string

	GetContainerDetail() *ContainerDetail
	GetAuth() RegistryAuthDefaults
	GetImage() string
	GetTag() string

	// Defaults returns the next link in the registry defaults chain.
	Defaults() RegistryDefaults

	// SetDefaults applies defaults to the receiver.
	// receiver -> rDefaults -> defaultDetail -> hardDefaultDetail.
	SetDefaults(rDefaults RegistryDefaults, defaultDetail, hardDefaultDetail *ContainerDetail)

	// String returns a YAML string representation of the receiver.
	String(prefix string) string
}

// Registry defines the methods that a Docker registry checker must implement.
type Registry interface {
	polymorphic.Inheritable
	yaml.IsZeroer

	// Copy returns a deep copy of the receiver.
	Copy() Registry

	// CheckValues validates the fields of the receiver.
	CheckValues() error

	// GetAuth returns the auth configured on this registry.
	GetAuth() RegistryAuth

	// String returns a YAML string representation of this registry.
	String(prefix string) string

	// Detail returns the resolved image and tag used for registry queries.
	Detail() ContainerDetail

	// Check queries the registry for the tag templated from version.
	Check(version string) error

	// Inherit copies query token state from when auth credentials match.
	Inherit(from Registry)

	// GetTypeSelf returns the configured type field; GetType (from polymorphic.Inheritable)
	// returns the fixed registry kind (e.g. "hub", "ghcr", "quay").
	GetTypeSelf() string
	GetImage() string
	GetImageSelf() string
	GetTag() string
	GetTagSelf() string

	// GetTagForVersion returns the tag to search for, templated with version.
	GetTagForVersion(version string) string
	newRequest(tag string) (*http.Request, error)

	parseBody(tag string, resp *http.Response) error

	// Defaults returns the next link in the registry defaults chain.
	Defaults() RegistryDefaults

	// SetDefaults applies the dType registry defaults to the receiver.
	SetDefaults(dType string, defaults *Defaults)
}
