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
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/httpx"
	"github.com/release-argus/Argus/util"
)

// #############
// # CONSTANTS #
// #############

var (
	// ghcrTokenAddress is the GHCR token endpoint.
	ghcrTokenAddress = "https://ghcr.io/token?scope=repository:%s:pull"
	// ghcrQueryURL is the GHCR query endpoint for image:tag queries.
	ghcrQueryURL = "https://ghcr.io/v2/%s/manifests/%s"
)

// ghcrTokenResponse is the response body for a GHCR access token request.
type ghcrTokenResponse struct {
	Token string `json:"token"`
}

// ####################
// # REGISTRY | TYPES #
// ####################

// GHCRRegistryDefaults holds defaults for queries on GHCR registries.
type GHCRRegistryDefaults struct {
	CommonRegistryDefaults `json:",inline" yaml:",inline"`
}

// GHCRRegistry holds data for queries on a GHCR registry.
type GHCRRegistry struct {
	CommonRegistry `json:",inline" yaml:",inline"`
}

// #######################
// # REGISTRY | DECODING #
// #######################

// UnmarshalJSON implements the json.Unmarshaler interface.
// Use [DecodeDefaults] for a complete GHCRRegistryDefaults.
func (r *GHCRRegistryDefaults) UnmarshalJSON(data []byte) error {
	return r.unmarshal("json", data)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
// Use [DecodeDefaults] for a complete GHCRRegistryDefaults.
func (r *GHCRRegistryDefaults) UnmarshalYAML(data []byte) error {
	return r.unmarshal("yaml", data)
}

// unmarshal implements the format.Unmarshaler interface.
func (r *GHCRRegistryDefaults) unmarshal(format string, data []byte) error {
	// Alias to avoid recursion.
	type Alias GHCRRegistryDefaults
	aux := (*Alias)(r)

	if r.Auth == nil {
		r.Auth = &GHCRAuthDefaults{}
	}
	// CommonRegistryDefaults.
	if len(data) != 0 {
		if err := decode.Unmarshal(format, data, aux); err != nil {
			return err //nolint:wrapcheck
		}
	}

	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// Use [Decode] for a complete GHCRRegistry.
func (r *GHCRRegistry) UnmarshalJSON(data []byte) error {
	return r.unmarshal("json", data)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
// Use [Decode] for a complete GHCRRegistry.
func (r *GHCRRegistry) UnmarshalYAML(data []byte) error {
	return r.unmarshal("yaml", data)
}

// unmarshal implements the format.Unmarshaler interface.
func (r *GHCRRegistry) unmarshal(format string, data []byte) error {
	// Alias to avoid recursion.
	type Alias GHCRRegistry
	aux := (*Alias)(r)

	if r.Auth == nil {
		r.Auth = &GHCRAuth{}
	}
	// CommonRegistry.
	if err := decode.Unmarshal(format, data, aux); err != nil {
		return err //nolint:wrapcheck
	}

	return nil
}

// DecodeSelf decodes the format-encoded data into the receiver.
func (r *GHCRRegistry) DecodeSelf(format string, data []byte) error {
	if err := decode.Unmarshal(format, data, r); err != nil {
		return err //nolint:wrapcheck
	}
	return nil
}

// ApplyOverrides applies format-encoded overrides to the receiver.
func (r *GHCRRegistry) ApplyOverrides(format string, data []byte) error {
	return r.DecodeSelf(format, data)
}

// ####################
// # REGISTRY | STATE #
// ####################

// IsZero implements the yaml.IsZeroer interface.
func (r *GHCRRegistryDefaults) IsZero() bool {
	if r == nil {
		return true
	}

	return r.Image == "" &&
		r.Tag == "" &&
		(r.Auth == nil || r.Auth.IsZero())
}

// IsZero implements the yaml.IsZeroer interface.
func (r *GHCRRegistry) IsZero() bool {
	if r == nil {
		return true
	}
	return r.CommonRegistry.IsZero()
}

// Copy returns a deep copy of the receiver.
func (r *GHCRRegistry) Copy() Registry {
	if r == nil {
		return nil
	}

	return &GHCRRegistry{
		CommonRegistry: *r.CommonRegistry.Clone(), //nolint:staticcheck,
	}
}

// ########################
// # REGISTRY | STRINGIFY #
// ########################

// String returns a string representation of the receiver.
func (r *GHCRRegistryDefaults) String(prefix string) string {
	if r == nil {
		return ""
	}
	return decode.ToYAMLString(r, prefix)
}

// String returns a string representation of the receiver.
func (r *GHCRRegistry) String(prefix string) string {
	if r == nil {
		return ""
	}
	return decode.ToYAMLString(r, prefix)
}

// #######################
// # REGISTRY | METADATA #
// #######################

// GetType returns the registry type identifier.
func (r *GHCRRegistryDefaults) GetType() string {
	return "ghcr"
}

// GetType returns the registry type identifier.
func (r *GHCRRegistry) GetType() string {
	return "ghcr"
}

// #########################
// # REGISTRY | OPERATIONS #
// #########################

// newRequest returns a HTTP GET request to query whether the given tag exists for the receiver's image.
func (r *GHCRRegistry) newRequest(tag string) (*http.Request, error) {
	url := fmt.Sprintf(
		ghcrQueryURL, r.GetImage(), tag,
	)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	req.Header.Set("Accept", "application/vnd.oci.image.index.v1+json")
	return req, nil
}

// Check queries the GHCR registry for the image:tag.
func (r *GHCRRegistry) Check(version string) error {
	return check(version, r)
}

// ################
// # AUTH | TYPES #
// ################

// GHCRAuthDefaults holds authentication defaults for GHCR.
type GHCRAuthDefaults struct {
	Token      string       `json:"token,omitempty" yaml:"token,omitempty"` // Personal access token used to obtain GHCR query tokens.
	mu         sync.RWMutex // Protects query token cache state.
	queryToken string       // Cached GHCR bearer token used for registry queries.
	validUntil time.Time    // Expiry time for the cached bearer token.

	// defaults form a fallback chain:
	//
	// instance -> provider defaults -> global defaults
	//
	// Values are resolved from most specific to least specific.
	defaults *GHCRAuthDefaults
}

// GHCRAuth holds authentication state for GHCR.
type GHCRAuth struct {
	GHCRAuthDefaults `json:",inline" yaml:",inline"`

	sf singleflight.Group // Deduplicate refreshes.
}

// ###################
// # AUTH | DECODING #
// ###################

// UnmarshalJSON implements the json.Unmarshaler interface.
// Use [DecodeDefaults] for a complete GHCRAuthDefaults.
func (d *GHCRAuthDefaults) UnmarshalJSON(data []byte) error {
	return d.unmarshal("json", data)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
// Use [DecodeDefaults] for a complete GHCRAuthDefaults.
func (d *GHCRAuthDefaults) UnmarshalYAML(data []byte) error {
	return d.unmarshal("yaml", data)
}

// unmarshal implements the format.Unmarshaler interface.
func (d *GHCRAuthDefaults) unmarshal(format string, data []byte) error {
	// Alias to avoid recursion.
	type Alias GHCRAuthDefaults
	aux := (*Alias)(d)

	// Unmarshal this format.
	if err := decode.Unmarshal(format, data, aux); err != nil {
		return err //nolint:wrapcheck
	}

	// Base64-encode the token if not already.
	if d.Token != "" {
		if strings.HasPrefix(d.Token, "ghp_") {
			d.queryToken = base64.StdEncoding.EncodeToString([]byte(d.Token))
		} else {
			d.queryToken = d.Token
		}
		d.validUntil = time.Now().Add(24 * time.Hour)
	}

	return nil
}

// ################
// # AUTH | STATE #
// ################

// IsZero implements the yaml.IsZeroer interface.
func (d *GHCRAuthDefaults) IsZero() bool {
	if d == nil {
		return true
	}

	return d.Token == ""
}

// Clone returns a deep copy of the receiver.
func (a *GHCRAuth) Clone() *GHCRAuth {
	if a == nil {
		return nil
	}

	return &GHCRAuth{
		GHCRAuthDefaults: GHCRAuthDefaults{
			Token:      a.Token,
			queryToken: a.queryToken,
			validUntil: a.validUntil,
			defaults:   a.defaults,
		},
	}
}

// Copy returns a deep copy of the receiver as a [RegistryAuth].
func (a *GHCRAuth) Copy() RegistryAuth {
	return a.Clone()
}

// ####################
// # AUTH | STRINGIFY #
// ####################

// String returns a YAML string representation of the receiver.
func (d *GHCRAuthDefaults) String(prefix string) string {
	return decode.ToYAMLString(d, prefix)
}

// ###################
// # AUTH | DEFAULTS #
// ###################

// Defaults returns the next link in the auth defaults chain.
func (d *GHCRAuthDefaults) Defaults() RegistryAuthDefaults {
	if d.defaults == nil {
		return nil
	}
	return d.defaults
}

// SetDefaults assigns defaults to the receiver.
func (d *GHCRAuthDefaults) SetDefaults(defaults RegistryAuthDefaults) {
	if ghcrDefaults, ok := defaults.(*GHCRAuthDefaults); ok {
		d.defaults = ghcrDefaults
	}
}

// #####################
// # AUTH | VALIDATION #
// #####################

// CheckValues validates the fields of the receiver.
func (a *GHCRAuth) CheckValues() error {
	return nil
}

// ######################
// # AUTH | CREDENTIALS #
// ######################

// GetTokenSelf returns the GHCR token configured on the receiver.
func (d *GHCRAuthDefaults) GetTokenSelf() string {
	return d.Token
}

// GetToken returns the GHCR token resolved from the receiver and its defaults chain.
func (a *GHCRAuth) GetToken() string {
	for auth := &a.GHCRAuthDefaults; auth != nil; auth = auth.defaults {
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
func (d *GHCRAuthDefaults) GetQueryTokenSelf() (string, time.Time) {
	d.mu.RLock()
	queryToken, validUntil := d.queryToken, d.validUntil
	d.mu.RUnlock()

	if isUsable(queryToken, validUntil) {
		return queryToken, validUntil
	}
	if d.Token == "" {
		return "", time.Time{}
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// Update expiry.
	d.validUntil = time.Now().Add(24 * time.Hour)
	return d.queryToken, d.validUntil
}

// GetQueryToken returns a cached repo-specific query token if available, otherwise refreshes it.
func (a *GHCRAuth) GetQueryToken(detail ContainerDetail) (string, error) {
	for auth := &a.GHCRAuthDefaults; auth != nil; auth = auth.defaults {
		// Only use the cached token if still usable.
		if queryToken, _ := auth.GetQueryTokenSelf(); queryToken != "" {
			return queryToken, nil
		}
	}

	// Deduplicate refreshes.
	v, err, _ := a.sf.Do("refresh-token", func() (any, error) {
		return a.refreshQueryToken(detail)
	})
	if err != nil {
		return "", err //nolint:wrapcheck
	}

	return v.(string), nil
}

// SetQueryToken stores the cached query token and expiry time, propagating to defaults when applicable
// (Ignores defaults since query tokens are repo-specific).
func (d *GHCRAuthDefaults) SetQueryToken(qT string, until time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.queryToken = qT
	d.validUntil = until
}

// refreshQueryToken retrieves a new query token for the given container details from GHCR using configured credentials.
func (a *GHCRAuth) refreshQueryToken(detail ContainerDetail) (string, error) {
	// Double-check whether the token is usable.
	a.mu.RLock()
	if isUsable(a.queryToken, a.validUntil) {
		token := a.queryToken
		a.mu.RUnlock()
		return token, nil
	}
	a.mu.RUnlock()

	address := fmt.Sprintf(ghcrTokenAddress, detail.Image)
	// Do the request.
	resp, err := httpx.Client.Get(address)
	if err != nil {
		return "", fmt.Errorf("ghcr token refresh fail: %w", err)
	}
	defer resp.Body.Close()

	// Read the token.
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ghcr token request failed: %s", body)
	}
	var tokenJSON ghcrTokenResponse
	if err := decode.Unmarshal("json", body, &tokenJSON); err != nil {
		return "", fmt.Errorf("failed to parse ghcr token response: %w", err)
	}
	// 30-second lifetime.
	validUntil := time.Now().UTC().Add(30 * time.Second)

	a.SetQueryToken(tokenJSON.Token, validUntil)
	return tokenJSON.Token, nil
}

// ######################
// # AUTH | INHERITANCE #
// ######################

// Inherit copies token data from another [GHCRAuth].
//
// - GHCR query tokens are repo-scoped, so we also require the same image.
func (a *GHCRAuth) Inherit(from RegistryAuth, srcDetail, dstDetail ContainerDetail) {
	o, ok := from.(*GHCRAuth)
	if !ok || srcDetail.Image != dstDetail.Image {
		return
	}
	// New token. Don't copy existing token data.
	if a.Token != util.SecretValue && a.GetToken() != o.GetToken() {
		return
	}

	// Copy token data.
	if a.Token == util.SecretValue {
		a.Token = o.Token
	}
	a.queryToken = o.queryToken
	a.validUntil = o.validUntil
}
