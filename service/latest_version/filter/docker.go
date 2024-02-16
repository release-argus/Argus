// Copyright [2023] [Argus]
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

package filter

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	net_url "net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/release-argus/Argus/util"
)

var dockerCheckTypes = []string{
	"hub", "quay", "ghcr"}

// DockerCheckRegistryBase is the base for checking a Docker registry for an image:tag.
type DockerCheckRegistryBase struct {
	Token string `yaml:"token,omitempty" json:"token,omitempty"` // Token to get the token for the queries

	queryToken string       // Token for queries
	validUntil time.Time    // Time until the token needs to be renewed
	mutex      sync.RWMutex // Mutex for the token
}

func (d *DockerCheckRegistryBase) String(prefix string) string {
	return util.ToYAMLString(d, prefix)
}

// DoockerCheckGHCR contains the information to get a token for queries on GHCR.
type DockerCheckGHCR struct {
	DockerCheckRegistryBase `yaml:",inline" json:",inline"`
}
type DockerCheckHub struct {
	DockerCheckRegistryBase `yaml:",inline" json:",inline"`
	Username                string `yaml:"username,omitempty" json:"username,omitempty"` // Username to get a new token
}

func (d *DockerCheckHub) String(prefix string) (str string) {
	if d != nil {
		str = util.ToYAMLString(d, prefix)
	}
	return
}

type DockerCheckQuay struct {
	DockerCheckRegistryBase `yaml:",inline" json:",inline"`
}

// DockerCheckDefaults are the default values for DockerCheck.
type DockerCheckDefaults struct {
	Type string `yaml:"type,omitempty" json:"type,omitempty"` // Type of the Docker registry

	RegistryGHCR *DockerCheckGHCR `yaml:"ghcr,omitempty" json:"ghcr,omitempty"` // Default GHCR Token
	RegistryHub  *DockerCheckHub  `yaml:"hub,omitempty" json:"hub,omitempty"`   // Default DockerHub Username/Token
	RegistryQuay *DockerCheckQuay `yaml:"quay,omitempty" json:"quay,omitempty"` // Defautlt Quay Token

	defaults *DockerCheckDefaults `yaml:"-" json:"-"` // Defaults to fall back on
}

// New DockerCheckDefaults.
func NewDockerCheckDefaults(
	dType string,
	tokenGHCR string,
	tokenHub string,
	usernameHub string,
	tokenQuay string,
	defaults *DockerCheckDefaults,
) (dflts *DockerCheckDefaults) {
	dflts = &DockerCheckDefaults{
		Type:     dType,
		defaults: defaults}
	if tokenGHCR != "" {
		dflts.RegistryGHCR = &DockerCheckGHCR{
			DockerCheckRegistryBase: DockerCheckRegistryBase{
				Token: tokenGHCR}}
	}
	if tokenHub != "" || usernameHub != "" {
		dflts.RegistryHub = &DockerCheckHub{
			DockerCheckRegistryBase: DockerCheckRegistryBase{
				Token: tokenHub},
			Username: usernameHub}
	}
	if tokenQuay != "" {
		dflts.RegistryQuay = &DockerCheckQuay{
			DockerCheckRegistryBase: DockerCheckRegistryBase{
				Token: tokenQuay}}
	}
	return
}

func (d *DockerCheckDefaults) String(prefix string) (str string) {
	if d == nil {
		return
	}

	if d.Type != "" {
		str += fmt.Sprintf("%stype: %s\n",
			prefix, d.Type)
	}

	if d.RegistryGHCR != nil {
		registryGHCRStr := d.RegistryGHCR.String(prefix + "    ")
		if registryGHCRStr != "" {
			str += fmt.Sprintf("%sghcr:\n%s",
				prefix, registryGHCRStr)
		}
	}
	registryHubStr := d.RegistryHub.String(prefix + "    ")
	if registryHubStr != "" {
		str += fmt.Sprintf("%shub:\n%s",
			prefix, registryHubStr)
	}
	if d.RegistryQuay != nil {
		registryQuayStr := d.RegistryQuay.String(prefix + "    ")
		if registryQuayStr != "" {
			str += fmt.Sprintf("%squay:\n%s",
				prefix, registryQuayStr)
		}
	}

	return
}

// CheckValues of the DockerCheckDefaults.
func (d *DockerCheckDefaults) CheckValues(prefix string) (errs error) {
	if d == nil {
		return
	}

	if d.Type != "" && !util.Contains(dockerCheckTypes, d.Type) {
		errs = fmt.Errorf("%s%stype: %q <invalid> (supported types = [%v])\\",
			util.ErrorToString(errs), prefix, d.Type, strings.Join(dockerCheckTypes, ","))
	}

	return
}

// DockerCheck will verify that Tag exists for Image
type DockerCheck struct {
	Type                    string `yaml:"type,omitempty" json:"type,omitempty"`         // Type of the Docker registry
	Username                string `yaml:"username,omitempty" json:"username,omitempty"` // Username to get a new token
	DockerCheckRegistryBase `yaml:",inline" json:",inline"`

	Image string `yaml:"image,omitempty" json:"image,omitempty"` // Image to check
	Tag   string `yaml:"tag,omitempty" json:"tag,omitempty"`     // Tag to check for

	Defaults *DockerCheckDefaults `yaml:"-" json:"-"` // Default values for DockerCheck
}

// New DockerCheck.
func NewDockerCheck(
	dType string,
	image string,
	tag string,
	username string,
	token string,
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
func (d *DockerCheck) String(prefix string) (str string) {
	if d != nil {
		str = util.ToYAMLString(d, prefix)
	}
	return
}

// DockerTagCheck will verify that Tag exists for Image and return an error if not.
func (r *Require) DockerTagCheck(
	version string,
) error {
	if r == nil || r.Docker == nil {
		return nil
	}
	var url string
	tag := r.Docker.GetTag(version)
	var req *http.Request
	queryToken, err := r.Docker.getQueryToken()
	if err != nil {
		return fmt.Errorf("%s:%s - %w",
			r.Docker.Image, tag, err)
	}
	switch r.Docker.GetType() {
	case "hub":
		url = fmt.Sprintf("https://registry.hub.docker.com/v2/repositories/%s/tags/%s",
			r.Docker.Image, tag)
	case "ghcr":
		url = fmt.Sprintf("https://ghcr.io/v2/%s/manifests/%s",
			r.Docker.Image, tag)
	case "quay":
		url = fmt.Sprintf("https://quay.io/api/v1/repository/%s/tag/?onlyActiveTags=true&specificTag=%s",
			r.Docker.Image, tag)
	}
	req, _ = http.NewRequest(http.MethodGet, url, nil)
	if queryToken != "" {
		req.Header.Set("Authorization", "Bearer "+queryToken)
	}
	req.Header.Set("Connection", "close")

	// Do the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s:%s - %w",
			r.Docker.Image, tag, err)
	}

	// Parse the body
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s:%s - %s",
			r.Docker.Image, tag, string(body))
	}
	// Quay will give a 200 even when the tag doesn't exist
	if r.Docker.GetType() == "quay" && strings.Contains(string(body), `"tags": []`) {
		return fmt.Errorf("%s:%s - tag not found",
			r.Docker.Image, tag)
	}

	return nil
}

// CheckValues of the DockerCheck.
func (d *DockerCheck) CheckValues(prefix string) (errs error) {
	if d == nil {
		return
	}

	if d.Type != "" {
		if !util.Contains(dockerCheckTypes, d.Type) {
			errs = fmt.Errorf("%s%stype: %q <invalid> (supported types = [%v])\\",
				util.ErrorToString(errs), prefix, d.Type, strings.Join(dockerCheckTypes, ","))
		}
	} else if d.GetType() == "" {
		errs = fmt.Errorf("%s%stype: <required> (supported types = %v)\\",
			util.ErrorToString(errs), prefix, strings.Join(dockerCheckTypes, ","))
	}

	if d.Image == "" {
		errs = fmt.Errorf("%s%simage: <required> (image to check tags for)",
			util.ErrorToString(errs), prefix)
	} else {
		regex := regexp.MustCompile(`^[\w\-\.\/]+$`)
		// invalid image
		if !regex.MatchString(d.Image) {
			errs = fmt.Errorf("%s%simage: %q <invalid> (non-ASCII)\\",
				util.ErrorToString(errs), prefix, d.Image)
			// e.g. prometheus = library/prometheus on the docker hub api
		} else if d.Type == "hub" && strings.Count(d.Image, "/") == 0 {
			d.Image = fmt.Sprintf("library/%s", d.Image)
		}
	}

	if d.Tag == "" {
		errs = fmt.Errorf("%s%stag: <required> (tag of image to check for existence)",
			util.ErrorToString(errs), prefix)
	} else if !util.CheckTemplate(d.Tag) {
		errs = fmt.Errorf("%s%stag: %q <invalid> (didn't pass templating)\\",
			util.ErrorToString(errs), prefix, d.Tag)
	}

	if err := d.checkToken(); err != nil {
		errs = fmt.Errorf("%s%s%w\\",
			util.ErrorToString(errs), prefix, err)
	}

	return
}

// checkToken is provided for registries that require one.
func (d *DockerCheck) checkToken() (err error) {
	if d == nil {
		return
	}

	switch d.GetType() {
	case "hub":
		username := d.getUsername()
		token := d.getToken()
		// require token if username is defined or vice-versa
		if username != "" && token == "" {
			err = fmt.Errorf("token: <required> (token for %s)",
				username)
		} else if username == "" && token != "" {
			err = fmt.Errorf("username: <required> (token is for who?)")
		}
	case "quay":
	case "ghcr":
	}

	return
}

// GetTag to search for on Image
func (d *DockerCheck) GetTag(version string) string {
	return util.TemplateString(d.Tag, util.ServiceInfo{LatestVersion: version})
}

// GetType of the DockerCheckDefaults.
func (d *DockerCheckDefaults) GetType() string {
	if d == nil {
		return ""
	}

	if d.Type != "" {
		return d.Type
	}
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
func (d *DockerCheck) getToken() (token string) {
	if d == nil {
		return
	}

	// Have a Token, return it
	if token = util.EvalEnvVars(d.Token); token != "" {
		return
	}

	token = util.EvalEnvVars(d.Defaults.getToken(d.GetType()))
	return
}

// getToken returns the token as is
func (d *DockerCheckDefaults) getToken(dType string) (token string) {
	if d == nil {
		return
	}

	// Use the type specific token
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
	if token != "" {
		return
	}

	token = util.EvalEnvVars(d.defaults.getToken(dType))
	return
}

// getValidToken looks for an existing queryToken on this registry type that's currently valid.
//
// empty string if no valid queryToken is found.
func (d *DockerCheck) getValidToken(dType string) (queryToken string) {
	if d == nil {
		return
	}

	d.mutex.RLock()
	defer d.mutex.RUnlock()

	// Have a QueryToken and it's valid for atleast 2s
	if d.queryToken != "" && d.validUntil.After(time.Now().Add(2*time.Second).UTC()) {
		queryToken = d.queryToken
		return
	}

	var validUntil time.Time
	queryToken, validUntil = d.Defaults.getQueryToken(d.GetType())
	// Have a queryToken and it's valid for atleast 2s
	if queryToken != "" && validUntil.After(time.Now().Add(2*time.Second).UTC()) {
		return
	}

	// No valid queryToken found
	queryToken = ""
	return
}

// getQueryToken recurses into itself to find a query token for the type.
func (d *DockerCheckDefaults) getQueryToken(dType string) (queryToken string, validUntil time.Time) {
	if d == nil {
		return
	}

	// Use the type specific queryToken
	switch dType {
	case "ghcr":
		if d.RegistryGHCR != nil {
			d.RegistryGHCR.mutex.RLock()
			defer d.RegistryGHCR.mutex.RUnlock()
			queryToken = d.RegistryGHCR.queryToken
			validUntil = d.RegistryGHCR.validUntil
		}
	case "hub":
		if d.RegistryHub != nil {
			d.RegistryHub.mutex.RLock()
			defer d.RegistryHub.mutex.RUnlock()
			queryToken = d.RegistryHub.queryToken
			validUntil = d.RegistryHub.validUntil
		}
	case "quay":
		if d.RegistryQuay != nil {
			d.RegistryQuay.mutex.RLock()
			defer d.RegistryQuay.mutex.RUnlock()
			queryToken = d.RegistryQuay.queryToken
			validUntil = d.RegistryQuay.validUntil
		}
	}
	if queryToken != "" {
		return
	}

	// Recurse into defaults
	queryToken, validUntil = d.defaults.getQueryToken(dType)
	return
}

// getQueryToken for API queries.
func (d *DockerCheck) getQueryToken() (queryToken string, err error) {
	dType := d.GetType()
	queryToken = d.getValidToken(dType)
	if queryToken != "" {
		return
	}

	// No valid queryToken found, get a new one from the Token
	token := d.getToken()
	switch dType {
	case "hub":
		// Anonymous query
		if token == "" {
			d.mutex.Lock()
			d.validUntil = time.Now().AddDate(1, 0, 0)
			d.mutex.Unlock()
			// Refresh token
		} else if err = d.refreshDockerHubToken(); err != nil {
			return
		}
	case "ghcr":
		if token != "" {
			queryToken = token
			// Base64 encode the token if it's not already
			if strings.HasPrefix(token, "ghp_") {
				queryToken = base64.StdEncoding.EncodeToString([]byte(token))
			}
			validUntil := time.Now().AddDate(1, 0, 0)
			d.SetQueryToken(&token, &queryToken, &validUntil)
			// Get a NOOP token for public images
		} else if err = d.refreshGHCRToken(); err != nil {
			return
		}
	case "quay":
		queryToken = token
		validUntil := time.Now().AddDate(1, 0, 0)
		d.SetQueryToken(&token, &queryToken, &validUntil)
	}

	// Get the refreshed token
	queryToken = d.getValidToken(dType)
	return
}

// getUsername for the given type
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

// getUsername for the given type
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
func (d *DockerCheck) CopyQueryToken() (queryToken string, validUntil time.Time) {
	if d == nil {
		return
	}
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	queryToken = d.queryToken
	validUntil = d.validUntil
	return
}

// SetDefaults to fall back to.
func (d *DockerCheckDefaults) SetDefaults(defaults *DockerCheckDefaults) {
	if d == nil {
		return
	}

	d.defaults = defaults
}

// setQueryToken and validUntil for the given type with this token.
func (d *DockerCheckDefaults) setQueryToken(dType, token *string, queryToken *string, validUntil *time.Time) {
	if d == nil {
		return
	}

	// If it came from the typed struct, update that
	switch *dType {
	case "ghcr":
		if d.RegistryGHCR != nil {
			d.RegistryGHCR.mutex.Lock()
			defer d.RegistryGHCR.mutex.Unlock()
			// Ignore NOOP tokens as they're repo/image specific
			if *token != "" && d.RegistryGHCR.Token == *token {
				d.RegistryGHCR.queryToken = *queryToken
				d.RegistryGHCR.validUntil = *validUntil
				return
			}
		}
	case "hub":
		if d.RegistryHub != nil {
			d.RegistryHub.mutex.Lock()
			defer d.RegistryHub.mutex.Unlock()
			// queryToken is global to all repos/images
			if d.RegistryHub.Token == *token {
				d.RegistryHub.queryToken = *queryToken
				d.RegistryHub.validUntil = *validUntil
				return
			}
		}
	case "quay":
		if d.RegistryQuay != nil {
			d.RegistryQuay.mutex.Lock()
			defer d.RegistryQuay.mutex.Unlock()
			// queryToken is global to all repos/images
			if d.RegistryQuay.Token == *token {
				d.RegistryQuay.queryToken = *queryToken
				d.RegistryQuay.validUntil = *validUntil
				return
			}
		}
	}

	d.defaults.setQueryToken(dType, token, queryToken, validUntil)
}

// SetQueryToken and validUntil for this DockerCheck's type at the given token.
func (d *DockerCheck) SetQueryToken(token, queryToken *string, validUntil *time.Time) {
	if d == nil {
		return
	}
	// Give the Token/ValidUntil to the main struct
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.queryToken = *queryToken
	d.validUntil = *validUntil

	// If the queryToken came direct from the main struct and the token it came from wasn't empty, we're done
	if d.Token == *token && d.Token != "" {
		return
	}

	dType := d.GetType()
	d.Defaults.setQueryToken(&dType, token, queryToken, validUntil)
}

// refreshDockerHubToken for the Image
func (d *DockerCheck) refreshDockerHubToken() error {
	token := d.getToken()
	// No Token found
	if token == "" {
		return nil
	}

	// Get the http.Request
	url := "https://registry.hub.docker.com/v2/users/login"
	reqBody := net_url.Values{}
	reqBody.Set("username", d.getUsername())
	reqBody.Set("password", token)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(reqBody.Encode()))
	if err != nil {
		return fmt.Errorf("DockerHub login request, creation failed: %w", err)
	}
	req.Header.Set("Connection", "close")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// Do the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("DockerHub login fail: %w", err)
	}

	// Parse the body
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
	// Give the Token/ValidUntil to this struct
	// and to the source of the Token
	d.SetQueryToken(&token, &queryToken, &validUntil)
	//nolint:wrapcheck
	return err
}

// refreshGHCRToken for the image
func (d *DockerCheck) refreshGHCRToken() error {
	url := fmt.Sprintf("https://ghcr.io/token?scope=repository:%s:pull", d.Image)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("GHCR token refresh fail: %w", err)
	}
	defer resp.Body.Close()

	// Read the token
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GHCR Token request failed: %s", body)
	}
	type ghcrJSON struct {
		Token string `json:"token"`
	}
	var tokenJSON ghcrJSON
	err = json.Unmarshal(body, &tokenJSON)

	queryToken := tokenJSON.Token
	validUntil := time.Now().UTC().Add(5 * time.Minute)
	// Give the Token/ValidUntil to this struct
	// and to the source of the Token
	d.SetQueryToken(&d.Token, &queryToken, &validUntil)
	//nolint:wrapcheck
	return err
}
