// Copyright [2025] [Argus]
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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	net_url "net/url"
	"strings"
	"sync"
	"time"

	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/util"
)

var dockerCheckTypes = []string{
	"hub", "quay", "ghcr"}

// DockerCheckRegistryBase is the base for checking a Docker registry for an image:tag.
type DockerCheckRegistryBase struct {
	Token string `yaml:"token,omitempty" json:"token,omitempty"` // Token to get the token for the queries.

	mutex      sync.RWMutex // Mutex for the token.
	queryToken string       // Token for queries.
	validUntil time.Time    // Time until the token needs to be renewed.
}

// String returns a string representation of the DockerCheckRegistryBase.
func (d *DockerCheckRegistryBase) String(prefix string) string {
	return util.ToYAMLString(d, prefix)
}

// DockerCheckGHCR contains the credentials to get a token for queries on GHCR.
type DockerCheckGHCR struct {
	DockerCheckRegistryBase `yaml:",inline" json:",inline"`
}

// DockerCheckHub contains the credentials to get a token for queries on DockerHub.
type DockerCheckHub struct {
	DockerCheckRegistryBase `yaml:",inline" json:",inline"`
	Username                string `yaml:"username,omitempty" json:"username,omitempty"` // Username to get a new token.
}

// String returns a string representation of the DockerCheckHub.
func (d *DockerCheckHub) String(prefix string) string {
	if d == nil {
		return ""
	}
	return util.ToYAMLString(d, prefix)
}

// DockerCheckQuay contains the credentials to get a token for queries on Quay.
type DockerCheckQuay struct {
	DockerCheckRegistryBase `yaml:",inline" json:",inline"`
}

// DockerCheckDefaults are the default values for DockerCheck.
type DockerCheckDefaults struct {
	Type string `yaml:"type,omitempty" json:"type,omitempty"` // Type of the Docker registry.

	RegistryGHCR *DockerCheckGHCR `yaml:"ghcr,omitempty" json:"ghcr,omitempty"` // Default GHCR Token.
	RegistryHub  *DockerCheckHub  `yaml:"hub,omitempty" json:"hub,omitempty"`   // Default DockerHub Username/Token.
	RegistryQuay *DockerCheckQuay `yaml:"quay,omitempty" json:"quay,omitempty"` // Default Quay Token.

	defaults *DockerCheckDefaults // Defaults to fall back on.
}

// NewDockerCheckDefaults returns a new DockerCheckDefaults with the given values.
func NewDockerCheckDefaults(
	dType string,
	tokenGHCR string,
	tokenHub, usernameHub string,
	tokenQuay string,
	defaults *DockerCheckDefaults,
) *DockerCheckDefaults {
	dockerCheckDefaults := &DockerCheckDefaults{
		Type:     dType,
		defaults: defaults}
	if tokenGHCR != "" {
		dockerCheckDefaults.RegistryGHCR = &DockerCheckGHCR{
			DockerCheckRegistryBase: DockerCheckRegistryBase{
				Token: tokenGHCR}}
	}
	if tokenHub != "" || usernameHub != "" {
		dockerCheckDefaults.RegistryHub = &DockerCheckHub{
			DockerCheckRegistryBase: DockerCheckRegistryBase{
				Token: tokenHub},
			Username: usernameHub}
	}
	if tokenQuay != "" {
		dockerCheckDefaults.RegistryQuay = &DockerCheckQuay{
			DockerCheckRegistryBase: DockerCheckRegistryBase{
				Token: tokenQuay}}
	}
	return dockerCheckDefaults
}

// Default sets this DockerCheckDefaults to the default values.
func (d *DockerCheckDefaults) Default() {
	d.Type = "hub"
}

func (d *DockerCheckDefaults) String(prefix string) string {
	if d == nil {
		return ""
	}

	var builder strings.Builder
	if d.Type != "" {
		builder.WriteString(fmt.Sprintf("%stype: %s\n",
			prefix, d.Type))
	}

	if d.RegistryGHCR != nil {
		registryGHCRStr := d.RegistryGHCR.String(prefix + "    ")
		if registryGHCRStr != "" {
			builder.WriteString(fmt.Sprintf("%sghcr:\n%s",
				prefix, registryGHCRStr))
		}
	}
	registryHubStr := d.RegistryHub.String(prefix + "    ")
	if registryHubStr != "" {
		builder.WriteString(fmt.Sprintf("%shub:\n%s",
			prefix, registryHubStr))
	}
	if d.RegistryQuay != nil {
		registryQuayStr := d.RegistryQuay.String(prefix + "    ")
		if registryQuayStr != "" {
			builder.WriteString(fmt.Sprintf("%squay:\n%s",
				prefix, registryQuayStr))
		}
	}

	return builder.String()
}

// CheckValues validates the fields of the DockerCheckDefaults struct.
func (d *DockerCheckDefaults) CheckValues(prefix string) error {
	if d == nil {
		return nil
	}

	if d.Type != "" && !util.Contains(dockerCheckTypes, d.Type) {
		return fmt.Errorf("%stype: %q <invalid> (supported types = [%v])",
			prefix, d.Type, strings.Join(dockerCheckTypes, ","))
	}

	return nil
}

// DockerCheck will verify that Tag exists for Image.
type DockerCheck struct {
	Type                    string `yaml:"type,omitempty" json:"type,omitempty"`         // Type of the Docker registry.
	Username                string `yaml:"username,omitempty" json:"username,omitempty"` // Username to get a new token.
	DockerCheckRegistryBase `yaml:",inline" json:",inline"`

	Image string `yaml:"image,omitempty" json:"image,omitempty"` // Image to check.
	Tag   string `yaml:"tag,omitempty" json:"tag,omitempty"`     // Tag to check for.

	Defaults *DockerCheckDefaults `yaml:"-" json:"-"` // Default values for DockerCheck.
}

// NewDockerCheck returns a new DockerCheck with the given values.
func NewDockerCheck(
	dType string,
	image, tag string,
	username, token string,
	queryToken string,
	validUntil time.Time,
	defaults *DockerCheckDefaults,
) *DockerCheck {
	return &DockerCheck{
		DockerCheckRegistryBase: DockerCheckRegistryBase{
			Token:      token,
			queryToken: queryToken,
			validUntil: validUntil},
		Username: username,
		Type:     dType,
		Image:    image,
		Tag:      tag,
		Defaults: defaults}
}

// String returns a string representation of the DockerCheck.
func (d *DockerCheck) String(prefix string) string {
	if d == nil {
		return ""
	}
	return util.ToYAMLString(d, prefix)
}

// DockerTagCheck verifies that Tag exists for Image.
func (r *Require) DockerTagCheck(
	version string,
) error {
	if r == nil || r.Docker == nil {
		return nil
	}
	var url string
	tag := r.Docker.GetTag(version)
	queryToken, err := r.Docker.getQueryToken()
	if err != nil {
		return fmt.Errorf("%s:%s - %w",
			r.Docker.Image, tag, err)
	}

	req, _ := http.NewRequest(http.MethodGet, "", nil)
	switch r.Docker.GetType() {
	case "hub":
		url = fmt.Sprintf("https://registry.hub.docker.com/v2/repositories/%s/tags/%s",
			r.Docker.Image, tag)
	case "ghcr":
		url = fmt.Sprintf("https://ghcr.io/v2/%s/manifests/%s",
			r.Docker.Image, tag)
		req.Header.Set("Accept", "application/vnd.oci.image.index.v1+json")
	case "quay":
		url = fmt.Sprintf("https://quay.io/api/v1/repository/%s/tag/?onlyActiveTags=true&specificTag=%s",
			r.Docker.Image, tag)
	}
	//#nosec G104 -- URL verified in CheckValues.
	//nolint:errcheck // ^
	parsedURL, _ := net_url.Parse(url)
	req.URL = parsedURL
	if queryToken != "" {
		req.Header.Set("Authorization", "Bearer "+queryToken)
	}
	req.Header.Set("Connection", "close")

	// Do the request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s:%s - %w",
			r.Docker.Image, tag, err)
	}

	// Parse the body.
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s:%s - %s",
			r.Docker.Image, tag, string(body))
	}
	// Quay will give a 200 even when the tag does not exist.
	if r.Docker.GetType() == "quay" && strings.Contains(string(body), `"tags": []`) {
		return fmt.Errorf("%s:%s - tag not found",
			r.Docker.Image, tag)
	}

	return nil
}

// CheckValues validates the fields of the DockerCheck struct.
func (d *DockerCheck) CheckValues(prefix string) error {
	if d == nil {
		return nil
	}

	var errs []error
	if d.Type != "" {
		if !util.Contains(dockerCheckTypes, d.Type) {
			errs = append(errs, fmt.Errorf("%stype: %q <invalid> (supported types = [%v])",
				prefix, d.Type, strings.Join(dockerCheckTypes, ",")))
		}
	} else if d.GetType() == "" {
		errs = append(errs, fmt.Errorf("%stype: <required> (supported types = %v)",
			prefix, strings.Join(dockerCheckTypes, ",")))
	}

	// Image
	switch {
	case d.Image == "":
		errs = append(errs, errors.New(prefix+
			"image: <required> (image to check tags for)"))
		// Invalid image.
	case !util.RegexCheck(`^[\w\-\.\/]+$`, d.Image):
		errs = append(errs, fmt.Errorf("%simage: %q <invalid> (non-ASCII)",
			prefix, d.Image))
		// e.g. prometheus = library/prometheus on the docker hub api.
	case d.Type == "hub" && strings.Count(d.Image, "/") == 0:
		d.Image = "library/" + d.Image
	}

	// Tag
	switch {
	case d.Tag == "":
		errs = append(errs, errors.New(prefix+
			"tag: <required> (tag of image to check for existence)"))
	case !util.CheckTemplate(d.Tag):
		errs = append(errs, fmt.Errorf("%stag: %q <invalid> (didn't pass templating)",
			prefix, d.Tag))
	default:
		if _, err := net_url.Parse("https://example.com/" + d.Tag); err != nil {
			errs = append(errs, fmt.Errorf("%stag: %q <invalid> (invalid for URL formatting)",
				prefix, d.Tag))
		}
	}

	if err := d.checkToken(); err != nil {
		errs = append(errs, fmt.Errorf("%s%w",
			prefix, err))
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// checkToken verifies that a Token exists for registries that require one.
func (d *DockerCheck) checkToken() error {
	if d == nil {
		return nil
	}

	switch d.GetType() {
	case "hub":
		username := d.getUsername()
		token := d.getToken()
		// require token if username defined or vice versa.
		if username != "" && token == "" {
			return fmt.Errorf("token: <required> (token for %s)",
				username)
		} else if username == "" && token != "" {
			return errors.New("username: <required> (token is for who?)")
		}
	case "quay", "ghcr":
		// Token not required.
	}

	return nil
}

// GetTag to search for on Image.
func (d *DockerCheck) GetTag(version string) string {
	return util.TemplateString(d.Tag, serviceinfo.ServiceInfo{LatestVersion: version})
}

// GetType of the DockerCheckDefaults.
func (d *DockerCheckDefaults) GetType() string {
	if d == nil {
		return ""
	}

	if d.Type != "" {
		return d.Type
	}
	// Recurse into defaults.
	return d.defaults.GetType()
}

// GetType of the DockerCheck.
func (d *DockerCheck) GetType() string {
	if d == nil {
		return ""
	}

	if d.Type != "" {
		return d.Type
	}
	return d.Defaults.GetType()
}

// getToken returns the token as is from this struct/defaults/hardDefaults.
func (d *DockerCheck) getToken() string {
	if d == nil {
		return ""
	}

	// Have a Token, return it.
	if token := util.EvalEnvVars(d.Token); token != "" {
		return token
	}

	return util.EvalEnvVars(d.Defaults.getToken(d.GetType()))
}

// getToken returns the token as is.
func (d *DockerCheckDefaults) getToken(dType string) string {
	if d == nil {
		return ""
	}

	// Use the type specific token.
	var token string
	switch dType {
	case "ghcr":
		if d.RegistryGHCR != nil {
			token = util.EvalEnvVars(d.RegistryGHCR.Token)
		}
	case "hub":
		if d.RegistryHub != nil {
			token = util.EvalEnvVars(d.RegistryHub.Token)
		}
	case "quay":
		if d.RegistryQuay != nil {
			token = util.EvalEnvVars(d.RegistryQuay.Token)
		}
	}

	// Return token if found.
	if token != "" {
		return token
	}
	// Recurse into defaults.
	return util.EvalEnvVars(d.defaults.getToken(dType))
}

// getValidToken looks for an existing queryToken on this registry type that is currently valid.
//
// Empty string if no valid queryToken is found.
func (d *DockerCheck) getValidToken() string {
	if d == nil {
		return ""
	}

	d.mutex.RLock()
	defer d.mutex.RUnlock()

	// Have a QueryToken, and it is valid for at least 2s.
	if d.queryToken != "" && d.validUntil.After(time.Now().Add(2*time.Second).UTC()) {
		return d.queryToken
	}

	queryToken, validUntil := d.Defaults.getQueryToken(d.GetType())
	// Have a queryToken, and it is valid for at least 2s.
	if queryToken != "" && validUntil.After(time.Now().Add(2*time.Second).UTC()) {
		return queryToken
	}

	// No valid queryToken found.
	return ""
}

// getQueryToken recurses into itself to find a query token for the type.
//
//	Returns the queryToken, and the time it is valid until.
func (d *DockerCheckDefaults) getQueryToken(dType string) (string, time.Time) {
	if d == nil {
		return "", time.Time{}
	}

	// Helper function to retrieve query token and validity.
	getQueryToken := func(registry *DockerCheckRegistryBase) (string, time.Time) {
		registry.mutex.RLock()
		defer registry.mutex.RUnlock()
		return registry.queryToken, registry.validUntil
	}

	// Use the type specific queryToken.
	switch dType {
	case "ghcr":
		if d.RegistryGHCR != nil {
			if token, valid := getQueryToken(&d.RegistryGHCR.DockerCheckRegistryBase); token != "" {
				return token, valid
			}
		}
	case "hub":
		if d.RegistryHub != nil {
			if token, valid := getQueryToken(&d.RegistryHub.DockerCheckRegistryBase); token != "" {
				return token, valid
			}
		}
	case "quay":
		if d.RegistryQuay != nil {
			if token, valid := getQueryToken(&d.RegistryQuay.DockerCheckRegistryBase); token != "" {
				return token, valid
			}
		}
	}

	// Recurse into defaults.
	return d.defaults.getQueryToken(dType)
}

// getQueryToken for API queries.
func (d *DockerCheck) getQueryToken() (string, error) {
	dType := d.GetType()
	if queryToken := d.getValidToken(); queryToken != "" {
		return queryToken, nil
	}

	// No valid queryToken found, get a new one from the Token.
	token := d.getToken()
	switch dType {
	case "hub":
		// Anonymous query.
		if token == "" {
			d.mutex.Lock()
			d.validUntil = time.Now().AddDate(1, 0, 0)
			d.mutex.Unlock()
			// Refresh token.
		} else if err := d.refreshDockerHubToken(); err != nil {
			return "", err
		}
	case "ghcr":
		if token != "" {
			queryToken := token
			// Base64 encode the token if not already.
			if strings.HasPrefix(token, "ghp_") {
				queryToken = base64.StdEncoding.EncodeToString([]byte(token))
			}
			validUntil := time.Now().AddDate(1, 0, 0)
			d.SetQueryToken(token, queryToken, validUntil)
			// Get a NOOP token for public images.
		} else if err := d.refreshGHCRToken(); err != nil {
			return "", err
		}
	case "quay":
		validUntil := time.Now().AddDate(1, 0, 0)
		d.SetQueryToken(token, token, validUntil)
	}

	// Get the refreshed token.
	return d.getValidToken(), nil
}

// getUsername for the given type.
func (d *DockerCheckDefaults) getUsername() string {
	if d == nil {
		return ""
	}

	if d.RegistryHub != nil {
		if username := util.EvalEnvVars(d.RegistryHub.Username); username != "" {
			return username
		}
	}

	return util.EvalEnvVars(d.defaults.getUsername())
}

// getUsername for the given type.
func (d *DockerCheck) getUsername() string {
	if d == nil {
		return ""
	}

	if username := util.EvalEnvVars(d.Username); username != "" {
		return username
	}
	return d.Defaults.getUsername()
}

// CopyQueryToken will return a copy of the queryToken along with the validUntil time.
func (d *DockerCheck) CopyQueryToken() (string, time.Time) {
	if d == nil {
		return "", time.Time{}
	}
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	return d.queryToken, d.validUntil
}

// SetDefaults to fall back to.
func (d *DockerCheckDefaults) SetDefaults(defaults *DockerCheckDefaults) {
	if d == nil {
		return
	}

	d.defaults = defaults
}

// setQueryToken and validUntil for the given type with this token.
func (d *DockerCheckDefaults) setQueryToken(dType, token, queryToken string, validUntil time.Time) {
	if d == nil {
		return
	}

	// Helper function to set the query token, and valid period if the token matches.
	setQueryToken := func(registry *DockerCheckRegistryBase, token, queryToken string, validUntil time.Time) bool {
		if token != registry.Token {
			return false
		}

		registry.mutex.RLock()
		defer registry.mutex.RUnlock()
		registry.queryToken = queryToken
		registry.validUntil = validUntil
		return true
	}

	// If it came from the typed struct, update that.
	switch dType {
	case "ghcr":
		// Ignore NOOP tokens as they are repo/image specific.
		if token == "" {
			return
		}
		if d.RegistryGHCR != nil {
			if didSet := setQueryToken(&d.RegistryGHCR.DockerCheckRegistryBase, token, queryToken, validUntil); didSet {
				return
			}
		}
	case "hub":
		if d.RegistryHub != nil {
			if didSet := setQueryToken(&d.RegistryHub.DockerCheckRegistryBase, token, queryToken, validUntil); didSet {
				return
			}
		}
	case "quay":
		if d.RegistryQuay != nil {
			if didSet := setQueryToken(&d.RegistryQuay.DockerCheckRegistryBase, token, queryToken, validUntil); didSet {
				return
			}
		}
	}

	d.defaults.setQueryToken(dType, token, queryToken, validUntil)
}

// SetQueryToken sets the queryToken, and its validUntil for DockerCheck's type at the given token.
func (d *DockerCheck) SetQueryToken(token, queryToken string, validUntil time.Time) {
	if d == nil {
		return
	}
	// Give the Token/ValidUntil to the main struct.
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.queryToken = queryToken
	d.validUntil = validUntil

	// If the queryToken came direct from the main struct, and the token it came from was not empty, we are done.
	if d.Token == token && d.Token != "" {
		return
	}

	dType := d.GetType()
	d.Defaults.setQueryToken(dType, token, queryToken, validUntil)
}

// ClearQueryToken for the DockerCheck.
func (d *DockerCheck) ClearQueryToken() {
	if d == nil {
		return
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.queryToken = ""
}

// refreshDockerHubToken for the Image.
func (d *DockerCheck) refreshDockerHubToken() error {
	token := d.getToken()
	// No Token found.
	if token == "" {
		return nil
	}

	// Get the http.Request.
	url := "https://registry.hub.docker.com/v2/users/login"
	reqBody := net_url.Values{}
	reqBody.Set("username", d.getUsername())
	reqBody.Set("password", token)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(reqBody.Encode()))
	if err != nil {
		return fmt.Errorf("dockerHub login request, creation failed: %w", err)
	}
	req.Header.Set("Connection", "close")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// Do the request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("dockerHub login fail: %w", err)
	}

	// Parse the body.
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refreshing the DockerHub token failed - %s", body)
	}
	type hubJSON struct {
		Token string `json:"token"`
	}
	var tokenJSON hubJSON
	err = json.Unmarshal(body, &tokenJSON)

	queryToken := tokenJSON.Token
	validUntil := time.Now().UTC().Add(5 * time.Minute)
	// Give the Token/ValidUntil to this struct,
	// and to the source of the Token.
	d.SetQueryToken(token, queryToken, validUntil)
	//nolint:wrapcheck
	return err
}

// refreshGHCRToken for the image.
func (d *DockerCheck) refreshGHCRToken() error {
	url := fmt.Sprintf("https://ghcr.io/token?scope=repository:%s:pull", d.Image)
	//#nosec G107 -- url built from the image.
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("ghcr token refresh fail: %w", err)
	}
	defer resp.Body.Close()

	// Read the token.
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ghcr token request failed: %s", body)
	}
	type ghcrJSON struct {
		Token string `json:"token"`
	}
	var tokenJSON ghcrJSON
	err = json.Unmarshal(body, &tokenJSON)

	queryToken := tokenJSON.Token
	validUntil := time.Now().UTC().Add(5 * time.Minute)
	// Give the Token/ValidUntil to this struct,
	// and to the source of the Token.
	d.SetQueryToken(d.Token, queryToken, validUntil)
	//nolint:wrapcheck
	return err
}
