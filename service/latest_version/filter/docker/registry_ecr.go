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
	"io"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/httpx"
)

// #############
// # CONSTANTS #
// #############

var (
	// ecrTokenAddress is the Amazon ECR Public Gallery anonymous token endpoint.
	ecrTokenAddress = "https://public.ecr.aws/token/"
	// ecrQueryURL is the Amazon ECR Public Gallery query endpoint for image:tag queries.
	ecrQueryURL = "https://public.ecr.aws/v2/%s/manifests/%s"
)

// ecrTokenResponse is the response body for an Amazon ECR Public Gallery access token request.
type ecrTokenResponse struct {
	Token string `json:"token"`
}

// ####################
// # REGISTRY | TYPES #
// ####################

// ECRRegistryDefaults holds defaults for queries on Amazon ECR Public Gallery registries.
type ECRRegistryDefaults struct {
	CommonRegistryDefaults `json:",inline" yaml:",inline"`
}

// ECRRegistry holds data for queries on an Amazon ECR Public Gallery registry.
type ECRRegistry struct {
	CommonRegistry `json:",inline" yaml:",inline"`
}

// #######################
// # REGISTRY | DECODING #
// #######################

// UnmarshalJSON implements the json.Unmarshaler interface.
// Use [DecodeDefaults] for a complete ECRRegistryDefaults.
func (r *ECRRegistryDefaults) UnmarshalJSON(data []byte) error {
	return r.unmarshal("json", data)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
// Use [DecodeDefaults] for a complete ECRRegistryDefaults.
func (r *ECRRegistryDefaults) UnmarshalYAML(data []byte) error {
	return r.unmarshal("yaml", data)
}

// unmarshal implements the format.Unmarshaler interface.
func (r *ECRRegistryDefaults) unmarshal(format string, data []byte) error {
	// Alias to avoid recursion.
	type Alias ECRRegistryDefaults
	aux := (*Alias)(r)

	if r.Auth == nil {
		r.Auth = &ECRAuthDefaults{}
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
// Use [Decode] for a complete ECRRegistry.
func (r *ECRRegistry) UnmarshalJSON(data []byte) error {
	return r.unmarshal("json", data)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
// Use [Decode] for a complete ECRRegistry.
func (r *ECRRegistry) UnmarshalYAML(data []byte) error {
	return r.unmarshal("yaml", data)
}

// unmarshal implements the format.Unmarshaler interface.
func (r *ECRRegistry) unmarshal(format string, data []byte) error {
	// Alias to avoid recursion.
	type Alias ECRRegistry
	aux := (*Alias)(r)

	if r.Auth == nil {
		r.Auth = &ECRAuth{}
	}
	// CommonRegistry.
	if len(data) != 0 {
		if err := decode.Unmarshal(format, data, aux); err != nil {
			return err //nolint:wrapcheck
		}
	}

	return nil
}

// DecodeSelf decodes the format-encoded data into the receiver.
func (r *ECRRegistry) DecodeSelf(format string, data []byte) error {
	if err := decode.Unmarshal(format, data, r); err != nil {
		return err //nolint:wrapcheck
	}
	return nil
}

// ApplyOverrides applies format-encoded overrides to the receiver.
func (r *ECRRegistry) ApplyOverrides(format string, data []byte) error {
	return r.DecodeSelf(format, data)
}

// ####################
// # REGISTRY | STATE #
// ####################

// IsZero implements the yaml.IsZeroer interface.
func (r *ECRRegistryDefaults) IsZero() bool {
	if r == nil {
		return true
	}

	return r.Auth == nil || r.Auth.IsZero()
}

// IsZero implements the yaml.IsZeroer interface.
func (r *ECRRegistry) IsZero() bool {
	if r == nil {
		return true
	}
	return r.CommonRegistry.IsZero()
}

// Copy returns a deep copy of the receiver.
func (r *ECRRegistry) Copy() Registry {
	if r == nil {
		return nil
	}

	return &ECRRegistry{
		CommonRegistry: *r.CommonRegistry.Clone(), //nolint:staticcheck
	}
}

// ########################
// # REGISTRY | STRINGIFY #
// ########################

// String returns a string representation of the receiver.
func (r *ECRRegistryDefaults) String(prefix string) string {
	if r == nil {
		return ""
	}
	return decode.ToYAMLString(r, prefix)
}

// String returns a string representation of the receiver.
func (r *ECRRegistry) String(prefix string) string {
	if r == nil {
		return ""
	}
	return decode.ToYAMLString(r, prefix)
}

// #######################
// # REGISTRY | METADATA #
// #######################

// GetType returns the registry type identifier.
func (r *ECRRegistryDefaults) GetType() string {
	return "ecr"
}

// GetType returns the registry type identifier.
func (r *ECRRegistry) GetType() string {
	return "ecr"
}

// #########################
// # REGISTRY | OPERATIONS #
// #########################

// newRequest returns a HTTP GET request to query whether the given tag exists for the receiver's image.
func (r *ECRRegistry) newRequest(tag string) (*http.Request, error) {
	url := fmt.Sprintf(
		ecrQueryURL, r.GetImage(), tag,
	)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	req.Header.Set("Accept", "application/vnd.oci.image.index.v1+json")
	return req, nil
}

// Check queries the Amazon ECR Public Gallery registry for the image:tag.
func (r *ECRRegistry) Check(version string) error {
	return check(version, r)
}

// ################
// # AUTH | TYPES #
// ################

// ECRAuthDefaults holds authentication defaults for the Amazon ECR Public Gallery.
//
// The Public Gallery is anonymous, so there are no configurable credentials; this
// type only caches the anonymous bearer token used for registry queries.
type ECRAuthDefaults struct {
	mu         sync.RWMutex // Protects query token cache state.
	queryToken string       // Cached ECR bearer token used for registry queries.
	validUntil time.Time    // Expiry time for the cached bearer token.

	// defaults form a fallback chain:
	//
	// instance -> provider defaults -> global defaults
	//
	// Values are resolved from most specific to least specific.
	defaults *ECRAuthDefaults
}

// ECRAuth holds authentication state for the Amazon ECR Public Gallery.
type ECRAuth struct {
	ECRAuthDefaults `json:",inline" yaml:",inline"`

	sf singleflight.Group // Deduplicate refreshes.
}

// ###################
// # AUTH | DECODING #
// ###################

// UnmarshalJSON implements the json.Unmarshaler interface.
// Use [DecodeDefaults] for a complete ECRAuthDefaults.
func (d *ECRAuthDefaults) UnmarshalJSON(data []byte) error {
	return d.unmarshal("json", data)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
// Use [DecodeDefaults] for a complete ECRAuthDefaults.
func (d *ECRAuthDefaults) UnmarshalYAML(data []byte) error {
	return d.unmarshal("yaml", data)
}

// unmarshal implements the format.Unmarshaler interface.
//
// Amazon ECR Public Gallery has no configurable credential, so there is nothing to
// decode; this only validates that the auth block is a mapping (mirroring the
// other registries) and discards its contents.
func (d *ECRAuthDefaults) unmarshal(format string, data []byte) error {
	var discard map[string]any
	if err := decode.Unmarshal(format, data, &discard); err != nil {
		return err //nolint:wrapcheck
	}

	return nil
}

// ################
// # AUTH | STATE #
// ################

// IsZero implements the yaml.IsZeroer interface.
//
// ECR auth carries no configuration, so it is always zero.
func (d *ECRAuthDefaults) IsZero() bool {
	return true
}

// Clone returns a deep copy of the receiver.
func (a *ECRAuth) Clone() *ECRAuth {
	if a == nil {
		return nil
	}

	return &ECRAuth{
		ECRAuthDefaults: ECRAuthDefaults{
			queryToken: a.queryToken,
			validUntil: a.validUntil,
			defaults:   a.defaults,
		},
	}
}

// Copy returns a deep copy of the receiver as a [RegistryAuth].
func (a *ECRAuth) Copy() RegistryAuth {
	return a.Clone()
}

// ####################
// # AUTH | STRINGIFY #
// ####################

// String returns a YAML string representation of the receiver.
func (d *ECRAuthDefaults) String(prefix string) string {
	return decode.ToYAMLString(d, prefix)
}

// ###################
// # AUTH | DEFAULTS #
// ###################

// Defaults returns the next link in the auth defaults chain.
func (d *ECRAuthDefaults) Defaults() RegistryAuthDefaults {
	if d.defaults == nil {
		return nil
	}
	return d.defaults
}

// SetDefaults assigns defaults to the receiver.
func (d *ECRAuthDefaults) SetDefaults(defaults RegistryAuthDefaults) {
	if ecrDefaults, ok := defaults.(*ECRAuthDefaults); ok {
		d.defaults = ecrDefaults
	}
}

// #####################
// # AUTH | VALIDATION #
// #####################

// CheckValues validates the fields of the receiver.
func (a *ECRAuth) CheckValues() error {
	return nil
}

// ######################
// # AUTH | CREDENTIALS #
// ######################

// GetTokenSelf returns the configured credential token (Amazon ECR Public Gallery has none).
func (d *ECRAuthDefaults) GetTokenSelf() string {
	return ""
}

// #######################
// # AUTH | QUERY TOKENS #
// #######################

// GetQueryTokenSelf returns the cached query token and its expiry time stored on the receiver.
func (d *ECRAuthDefaults) GetQueryTokenSelf() (string, time.Time) {
	d.mu.RLock()
	queryToken, validUntil := d.queryToken, d.validUntil
	d.mu.RUnlock()

	if isUsable(queryToken, validUntil) {
		return queryToken, validUntil
	}
	return "", time.Time{}
}

// GetQueryToken returns a cached anonymous query token if available, otherwise refreshes it.
func (a *ECRAuth) GetQueryToken(detail ContainerDetail) (string, error) {
	for auth := &a.ECRAuthDefaults; auth != nil; auth = auth.defaults {
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

// SetQueryToken stores the cached query token and expiry time on the receiver.
func (d *ECRAuthDefaults) SetQueryToken(qT string, until time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.queryToken = qT
	d.validUntil = until
}

// refreshQueryToken retrieves a new anonymous query token from the Amazon ECR Public Gallery.
func (a *ECRAuth) refreshQueryToken(_ ContainerDetail) (string, error) {
	// Double-check whether the token is usable.
	a.mu.RLock()
	if isUsable(a.queryToken, a.validUntil) {
		token := a.queryToken
		a.mu.RUnlock()
		return token, nil
	}
	a.mu.RUnlock()

	// Do the request.
	resp, err := httpx.Client.Get(ecrTokenAddress)
	if err != nil {
		return "", fmt.Errorf("ecr token refresh fail: %w", err)
	}
	defer resp.Body.Close()

	// Read the token.
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ecr token request failed: %s", body)
	}
	var tokenJSON ecrTokenResponse
	if err := decode.Unmarshal("json", body, &tokenJSON); err != nil {
		return "", fmt.Errorf("failed to parse ecr token response: %w", err)
	}
	validUntil := time.Now().UTC().Add(12 * time.Hour).Add(-10 * time.Minute)

	a.SetQueryToken(tokenJSON.Token, validUntil)
	return tokenJSON.Token, nil
}

// ######################
// # AUTH | INHERITANCE #
// ######################

// Inherit copies token data from another [ECRAuth].
//
// - Amazon ECR Public Gallery tokens are anonymous and global; image/tag does not matter here.
func (a *ECRAuth) Inherit(from RegistryAuth, _, _ ContainerDetail) {
	o, ok := from.(*ECRAuth)
	if !ok {
		return
	}

	// Copy token data.
	a.queryToken = o.queryToken
	a.validUntil = o.validUntil
}
