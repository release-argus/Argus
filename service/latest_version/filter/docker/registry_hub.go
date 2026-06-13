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
	"bytes"
	"errors"
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
	// hubTokenAddress is the Docker Hub token endpoint.
	hubTokenAddress = "https://registry.hub.docker.com/v2/auth/token"
	// hubQueryURL is Docker Hub query endpoint for image:tag queries.
	hubQueryURL = "https://registry.hub.docker.com/v2/repositories/%s/tags/%s"
)

// hubTokenRequest is the request body sent to Docker Hub to obtain an access token.
type hubTokenRequest struct {
	Identifier string `json:"identifier"`
	Secret     string `json:"secret"`
}

// hubTokenResponse is the response body for a Docker Hub access token request.
type hubTokenResponse struct {
	Token string `json:"access_token"`
}

// ####################
// # REGISTRY | TYPES #
// ####################

// HubRegistryDefaults holds defaults for queries on Docker Hub registries.
type HubRegistryDefaults struct {
	CommonRegistryDefaults `json:",inline" yaml:",inline"`
}

// HubRegistry holds data for queries on a Docker Hub registry.
type HubRegistry struct {
	CommonRegistry `json:",inline" yaml:",inline"`
}

// #######################
// # REGISTRY | DECODING #
// #######################

// UnmarshalJSON implements the json.Unmarshaler interface.
// Use [DecodeDefaults] for a full unmarshal.
func (r *HubRegistryDefaults) UnmarshalJSON(data []byte) error {
	return r.unmarshal("json", data)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
// Use [DecodeDefaults] for a full unmarshal.
func (r *HubRegistryDefaults) UnmarshalYAML(data []byte) error {
	return r.unmarshal("yaml", data)
}

// unmarshal implements the format.Unmarshaler interface.
func (r *HubRegistryDefaults) unmarshal(format string, data []byte) error {
	// Alias to avoid recursion.
	type Alias HubRegistryDefaults
	aux := (*Alias)(r)

	if r.Auth == nil {
		r.Auth = &HubAuthDefaults{}
	}
	// CommonRegistryDefaults.
	if err := decode.Unmarshal(format, data, aux); err != nil {
		return err //nolint:wrapcheck
	}

	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// Use [Decode] for a full unmarshal.
func (r *HubRegistry) UnmarshalJSON(data []byte) error {
	return r.unmarshal("json", data)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
// Use [Decode] for a full unmarshal.
func (r *HubRegistry) UnmarshalYAML(data []byte) error {
	return r.unmarshal("yaml", data)
}

// unmarshal implements the format.Unmarshaler interface.
func (r *HubRegistry) unmarshal(format string, data []byte) error {
	// Alias to avoid recursion.
	type Alias HubRegistry
	aux := (*Alias)(r)

	if r.Auth == nil {
		r.Auth = &HubAuth{}
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
func (r *HubRegistry) DecodeSelf(format string, data []byte) error {
	if err := decode.Unmarshal(format, data, r); err != nil {
		return err //nolint:wrapcheck
	}
	return nil
}

// ApplyOverrides applies format-encoded overrides to the receiver.
func (r *HubRegistry) ApplyOverrides(format string, data []byte) error {
	return r.DecodeSelf(format, data)
}

// ####################
// # REGISTRY | STATE #
// ####################

// IsZero implements the yaml.IsZeroer interface.
func (r *HubRegistryDefaults) IsZero() bool {
	if r == nil {
		return true
	}

	return r.Image == "" &&
		r.Tag == "" &&
		(r.Auth == nil || r.Auth.IsZero())
}

// IsZero implements the yaml.IsZeroer interface.
func (r *HubRegistry) IsZero() bool {
	if r == nil {
		return true
	}
	return r.CommonRegistry.IsZero()
}

// Copy returns a deep copy of the receiver.
func (r *HubRegistry) Copy() Registry {
	if r == nil {
		return nil
	}

	return &HubRegistry{
		CommonRegistry: *r.CommonRegistry.Clone(), //nolint:staticcheck,
	}
}

// ########################
// # REGISTRY | STRINGIFY #
// ########################

// String returns a string representation of the receiver.
func (r *HubRegistryDefaults) String(prefix string) string {
	if r == nil {
		return ""
	}
	return decode.ToYAMLString(r, prefix)
}

// String returns a string representation of the receiver.
func (r *HubRegistry) String(prefix string) string {
	if r == nil {
		return ""
	}
	return decode.ToYAMLString(r, prefix)
}

// #######################
// # REGISTRY | METADATA #
// #######################

// GetType returns the registry type identifier.
func (r *HubRegistryDefaults) GetType() string {
	return "hub"
}

// GetType returns the registry type identifier.
func (r *HubRegistry) GetType() string {
	return "hub"
}

// #########################
// # REGISTRY | VALIDATION #
// #########################

// CheckValues validates the fields of the receiver.
func (r *HubRegistry) CheckValues() error {
	if err := r.CommonRegistry.CheckValues(); err != nil {
		return err
	}

	// e.g. prometheus = library/prometheus on the docker hub api.
	if image := r.GetImage(); image != "" && strings.Count(image, "/") == 0 {
		r.Image = "library/" + image
	}

	return nil
}

// #########################
// # REGISTRY | OPERATIONS #
// #########################

// newRequest returns a HTTP GET request to query whether the given tag exists for the receiver's image.
func (r *HubRegistry) newRequest(tag string) (*http.Request, error) {
	url := fmt.Sprintf(
		hubQueryURL,
		r.GetImage(), tag,
	)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	return req, nil
}

// Check queries the Docker Hub registry for the image:tag.
func (r *HubRegistry) Check(version string) error {
	return check(version, r)
}

// ################
// # AUTH | TYPES #
// ################

// HubAuthDefaults holds authentication defaults for Docker Hub.
type HubAuthDefaults struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty"` // Username to get a new token for.
	Token    string `json:"token,omitempty" yaml:"token,omitempty"`       // Personal access token used to obtain Docker Hub query tokens.

	mu         sync.RWMutex // Protects query token cache state.
	queryToken string       // Cached Docker Hub bearer token used for registry queries.
	validUntil time.Time    // Expiry time for the cached bearer token.

	// defaults form a fallback chain:
	//
	// instance -> provider defaults -> global defaults
	//
	// Values are resolved from most specific to least specific.
	defaults *HubAuthDefaults
}

// HubAuth holds authentication values for queries on Docker Hub.
type HubAuth struct {
	HubAuthDefaults `json:",inline" yaml:",inline"`

	sf singleflight.Group // Deduplicate refreshes.
}

// ################
// # AUTH | STATE #
// ################

// IsZero implements the yaml.IsZeroer interface.
func (d *HubAuthDefaults) IsZero() bool {
	if d == nil {
		return true
	}

	return d.Username == "" &&
		d.Token == ""
}

// Clone returns a deep copy of the receiver.
func (d *HubAuth) Clone() *HubAuth {
	if d == nil {
		return nil
	}

	return &HubAuth{
		HubAuthDefaults: HubAuthDefaults{
			Username:   d.Username,
			Token:      d.Token,
			queryToken: d.queryToken,
			validUntil: d.validUntil,
			defaults:   d.defaults,
		},
	}
}

// Copy returns a deep copy of the receiver as a [RegistryAuth].
func (a *HubAuth) Copy() RegistryAuth {
	return a.Clone()
}

// ####################
// # AUTH | STRINGIFY #
// ####################

// String returns a YAML string representation of the receiver.
func (d *HubAuthDefaults) String(prefix string) string {
	return decode.ToYAMLString(d, prefix)
}

// ###################
// # AUTH | DEFAULTS #
// ###################

// Defaults returns the next link in the auth defaults chain.
func (d *HubAuthDefaults) Defaults() RegistryAuthDefaults {
	if d.defaults == nil {
		return nil
	}

	return d.defaults
}

// SetDefaults assigns defaults to the receiver.
func (d *HubAuthDefaults) SetDefaults(defaults RegistryAuthDefaults) {
	if hubDefaults, ok := defaults.(*HubAuthDefaults); ok {
		d.defaults = hubDefaults
	}
}

// #####################
// # AUTH | VALIDATION #
// #####################

// CheckValues validates the fields of the receiver.
func (a *HubAuth) CheckValues() error {
	if a == nil {
		return nil
	}
	var errs []error

	username := a.GetUsername()
	token := a.GetToken()
	// Require 'token' if 'username' defined, or vice versa.
	if (username == "") != (token == "") {
		if username == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "username",
					Description: "user for the token",
				},
			)
		} else {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "token",
					Description: "token for " + username,
				},
			)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// ######################
// # AUTH | CREDENTIALS #
// ######################

// GetUsernameSelf returns the Docker Hub username configured on the receiver.
func (d *HubAuthDefaults) GetUsernameSelf() string {
	return util.EvalEnvVars(d.Username)
}

// GetUsername returns the Docker Hub username resolved from the receiver and its defaults chain.
func (a *HubAuth) GetUsername() string {
	for auth := &a.HubAuthDefaults; auth != nil; auth = auth.defaults {
		if u := util.EvalEnvVars(auth.Username); u != "" {
			return u
		}
	}
	return ""
}

// GetTokenSelf returns the Docker Hub token configured on the receiver.
func (d *HubAuthDefaults) GetTokenSelf() string {
	return d.Token
}

// GetToken returns the Docker Hub token resolved from the receiver and its defaults chain.
func (a *HubAuth) GetToken() string {
	for auth := &a.HubAuthDefaults; auth != nil; auth = auth.defaults {
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
func (d *HubAuthDefaults) GetQueryTokenSelf() (string, time.Time) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// No Query Token.
	if d.queryToken == "" {
		return "", time.Time{}
	}

	// Usable Query Token.
	if isUsable(d.queryToken, d.validUntil) {
		return d.queryToken, d.validUntil
	}

	// Expired Query Token.
	return "", time.Time{}
}

// GetQueryToken returns a cached query token if available, otherwise refreshes it.
func (a *HubAuth) GetQueryToken(detail ContainerDetail) (string, error) {
	for auth := &a.HubAuthDefaults; auth != nil; auth = auth.defaults {
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

// SetQueryToken stores the cached query token and expiry time, propagating to defaults when applicable.
func (d *HubAuthDefaults) SetQueryToken(q string, until time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.queryToken = q
	d.validUntil = until

	if defaults, ok := d.Defaults().(*HubAuthDefaults); ok &&
		((d.Token == "" && d.Username == "") ||
			(d.Token == defaults.Token && d.Username == defaults.Username)) {
		defaults.SetQueryToken(q, until)
	}
}

// refreshQueryToken retrieves a new query token from Docker Hub using configured credentials.
func (a *HubAuth) refreshQueryToken(ContainerDetail) (string, error) {
	username := a.GetUsername()
	token := a.GetToken()
	// No Username/Token found.
	if username == "" || token == "" {
		return "", nil
	}

	reqBody := hubTokenRequest{
		Identifier: username,
		Secret:     token,
	}

	// Create the http.Request:
	// https://docs.docker.com/reference/api/hub/latest/#tag/authentication-api/operation/AuthCreateAccessToken.
	b, err := decode.Marshal("json", reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal docker-hub request body: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, hubTokenAddress, bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("create docker-hub token request: %w", err)
	}
	// req.Header.Set("Connection", "close")
	req.Header.Add("Content-Type", "application/json")

	// Do the request.
	resp, err := httpx.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("docker-hub token request failed: %w", err)
	}
	defer resp.Body.Close()

	// Parse the body.
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"docker-hub token request failed (status=%d): %s",
			resp.StatusCode, body,
		)
	}
	var tokenJSON hubTokenResponse
	if err := decode.Unmarshal("json", body, &tokenJSON); err != nil {
		return "", fmt.Errorf("failed to parse docker-hub token response: %w", err)
	}

	queryToken := tokenJSON.Token
	// 10-minute lifetime.
	validUntil := time.Now().UTC().Add(10 * time.Minute)

	a.SetQueryToken(tokenJSON.Token, validUntil)
	return queryToken, nil
}

// ######################
// # AUTH | INHERITANCE #
// ######################

// Inherit copies token data from another [HubAuth].
//
// - Docker Hub auth identity is username+token; image/tag does not matter here.
func (a *HubAuth) Inherit(from RegistryAuth, srcDetail, dstDetail ContainerDetail) {
	o, ok := from.(*HubAuth)
	if !ok || a.GetUsername() != o.GetUsername() {
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
