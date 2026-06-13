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
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/util"
)

// #############
// # CONSTANTS #
// #############

// quayQueryURL is the Quay query endpoint for image:tag queries.
var quayQueryURL = "https://quay.io/api/v1/repository/%s/tag/?onlyActiveTags=true&specificTag=%s"

// ####################
// # REGISTRY | TYPES #
// ####################

// QuayRegistryDefaults holds defaults for queries on Quay registries.
type QuayRegistryDefaults struct {
	CommonRegistryDefaults `json:",inline" yaml:",inline"`
}

// QuayRegistry holds data for queries on a Quay registry.
type QuayRegistry struct {
	CommonRegistry `json:",inline" yaml:",inline"`
}

// #######################
// # REGISTRY | DECODING #
// #######################

// UnmarshalJSON implements the json.Unmarshaler interface.
// Use [DecodeDefaults] for a full unmarshal.
func (r *QuayRegistryDefaults) UnmarshalJSON(data []byte) error {
	return r.unmarshal("json", data)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
// Use [DecodeDefaults] for a full unmarshal.
func (r *QuayRegistryDefaults) UnmarshalYAML(data []byte) error {
	return r.unmarshal("yaml", data)
}

// unmarshal implements the format.Unmarshaler interface.
func (r *QuayRegistryDefaults) unmarshal(format string, data []byte) error {
	// Alias to avoid recursion.
	type Alias QuayRegistryDefaults
	aux := (*Alias)(r)

	// CommonRegistryDefaults.
	if r.Auth == nil {
		r.Auth = &QuayAuthDefaults{}
	}
	if len(data) != 0 {
		if err := decode.Unmarshal(format, data, aux); err != nil {
			return err //nolint:wrapcheck
		}
	}

	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// Use [Decode] for a full unmarshal.
func (r *QuayRegistry) UnmarshalJSON(data []byte) error {
	return r.unmarshal("json", data)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
// Use [Decode] for a full unmarshal.
func (r *QuayRegistry) UnmarshalYAML(data []byte) error {
	return r.unmarshal("yaml", data)
}

// unmarshal implements the format.Unmarshaler interface.
func (r *QuayRegistry) unmarshal(format string, data []byte) error {
	// Alias to avoid recursion.
	type Alias QuayRegistry
	aux := (*Alias)(r)

	if r.Auth == nil {
		r.Auth = &QuayAuth{}
	}
	// CommonRegistry.
	if err := decode.Unmarshal(format, data, aux); err != nil {
		return err //nolint:wrapcheck
	}

	return nil
}

// DecodeSelf decodes the format-encoded data into the receiver.
func (r *QuayRegistry) DecodeSelf(format string, data []byte) error {
	if err := decode.Unmarshal(format, data, r); err != nil {
		return err //nolint:wrapcheck
	}
	return nil
}

// ApplyOverrides applies format-encoded overrides to the receiver.
func (r *QuayRegistry) ApplyOverrides(format string, data []byte) error {
	return r.DecodeSelf(format, data)
}

// ####################
// # REGISTRY | STATE #
// ####################

// IsZero implements the yaml.IsZeroer interface.
func (r *QuayRegistryDefaults) IsZero() bool {
	if r == nil {
		return true
	}

	return r.Image == "" &&
		r.Tag == "" &&
		(r.Auth == nil || r.Auth.IsZero())
}

// IsZero implements the yaml.IsZeroer interface.
func (r *QuayRegistry) IsZero() bool {
	if r == nil {
		return true
	}
	return r.CommonRegistry.IsZero()
}

// Copy returns a deep copy of the receiver.
func (r *QuayRegistry) Copy() Registry {
	if r == nil {
		return nil
	}

	return &QuayRegistry{
		CommonRegistry: *r.CommonRegistry.Clone(), //nolint:staticcheck
	}
}

// ########################
// # REGISTRY | STRINGIFY #
// ########################

// String returns a string representation of the receiver.
func (r *QuayRegistryDefaults) String(prefix string) string {
	if r == nil {
		return ""
	}
	return decode.ToYAMLString(r, prefix)
}

// String returns a string representation of the receiver.
func (r *QuayRegistry) String(prefix string) string {
	if r == nil {
		return ""
	}
	return decode.ToYAMLString(r, prefix)
}

// #######################
// # REGISTRY | METADATA #
// #######################

// GetType returns the registry type identifier.
func (r *QuayRegistryDefaults) GetType() string {
	return "quay"
}

// GetType returns the registry type identifier.
func (r *QuayRegistry) GetType() string {
	return "quay"
}

// #########################
// # REGISTRY | OPERATIONS #
// #########################

// newRequest returns a HTTP GET request to query whether the given tag exists for the receiver's image.
func (r *QuayRegistry) newRequest(tag string) (*http.Request, error) {
	url := fmt.Sprintf(
		quayQueryURL,
		r.GetImage(), tag,
	)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	return req, nil
}

// parseBody parses the HTTP response returned by newRequest for the given tag.
func (r *QuayRegistry) parseBody(tag string, resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"%s:%s %w",
			r.GetImage(), tag, errors.New(string(body)),
		)
	}

	// Quay will give a 200 even when the tag does not exist.
	if strings.Contains(string(body), `"tags": []`) {
		return TagNotFoundError{Image: r.GetImage(), Tag: tag}
	}

	return nil
}

// Check queries the Quay registry for the image:tag.
func (r *QuayRegistry) Check(version string) error {
	return check(version, r)
}

// ################
// # AUTH | TYPES #
// ################

// QuayAuthDefaults holds authentication defaults for Quay.
type QuayAuthDefaults struct {
	Token string `json:"token,omitempty" yaml:"token,omitempty"` // Token for registry queries.

	// defaults form a fallback chain:
	//
	// instance -> provider defaults -> global defaults
	//
	// Values are resolved from most specific to least specific.
	defaults *QuayAuthDefaults
}

// QuayAuth holds authentication state for Quay.
type QuayAuth struct {
	QuayAuthDefaults `json:",inline" yaml:",inline"`
}

// ################
// # AUTH | STATE #
// ################

// IsZero implements the yaml.IsZeroer interface.
func (d *QuayAuthDefaults) IsZero() bool {
	if d == nil {
		return true
	}

	return d.Token == ""
}

// Clone returns a deep copy of the receiver.
func (a *QuayAuth) Clone() *QuayAuth {
	if a == nil {
		return nil
	}

	return &QuayAuth{
		QuayAuthDefaults: QuayAuthDefaults{
			Token:    a.Token,
			defaults: a.defaults,
		},
	}
}

// Copy returns a deep copy of the receiver as a [RegistryAuth].
func (a *QuayAuth) Copy() RegistryAuth {
	return a.Clone()
}

// ####################
// # AUTH | STRINGIFY #
// ####################

// String returns a YAML string representation of the receiver.
func (d *QuayAuthDefaults) String(prefix string) string {
	return decode.ToYAMLString(d, prefix)
}

// ###################
// # AUTH | DEFAULTS #
// ###################

// Defaults returns the next link in the auth defaults chain.
func (d *QuayAuthDefaults) Defaults() RegistryAuthDefaults {
	if d.defaults == nil {
		return nil
	}

	return d.defaults
}

// SetDefaults assigns defaults to the receiver.
func (d *QuayAuthDefaults) SetDefaults(defaults RegistryAuthDefaults) {
	if quayDefaults, ok := defaults.(*QuayAuthDefaults); ok {
		d.defaults = quayDefaults
	}
}

// #####################
// # AUTH | VALIDATION #
// #####################

// CheckValues validates the fields of the receiver.
func (a *QuayAuth) CheckValues() error {
	return nil
}

// ######################
// # AUTH | CREDENTIALS #
// ######################

// GetTokenSelf returns the Quay token configured on the receiver.
func (d *QuayAuthDefaults) GetTokenSelf() string {
	return util.EvalEnvVars(d.Token)
}

// GetToken returns the Quay token resolved from the receiver and its defaults chain.
func (a *QuayAuth) GetToken() string {
	for auth := &a.QuayAuthDefaults; auth != nil; auth = auth.defaults {
		if t := util.EvalEnvVars(auth.Token); t != "" {
			return t
		}
	}
	return ""
}

// #######################
// # AUTH | QUERY TOKENS #
// #######################

// GetQueryTokenSelf returns the cached query token and its expiry time stored on the receiver.
func (d *QuayAuthDefaults) GetQueryTokenSelf() (string, time.Time) {
	return d.GetTokenSelf(), time.Time{}
}

// GetQueryToken returns a cached query token if available, otherwise refreshes it.
func (a *QuayAuth) GetQueryToken(ContainerDetail) (string, error) {
	return a.GetToken(), nil
}

// SetQueryToken implements [RegistryAuth]
// (Token-only registries do not use query tokens).
func (d *QuayAuthDefaults) SetQueryToken(_ string, _ time.Time) {
	// no-op: Token-only registries do not cache query tokens.
}

// ######################
// # AUTH | INHERITANCE #
// ######################

// Inherit copies token data from another [QuayAuth].
//
// - Quay tokens fully identify; image/tag does not matter here.
func (a *QuayAuth) Inherit(from RegistryAuth, srcDetail, dstDetail ContainerDetail) {
	o, ok := from.(*QuayAuth)
	if !ok {
		return
	}

	// Copy token data.
	if a.Token == util.SecretValue {
		a.Token = o.Token
	}
}
