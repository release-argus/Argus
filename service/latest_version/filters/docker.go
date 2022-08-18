// Copyright [2022] [Argus]
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

package filters

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/release-argus/Argus/utils"
)

type DockerCheck struct {
	Type       string    `yaml:"type"`               // Where to check, e.g. hub (DockerHub), GHCR, Quay
	Image      string    `yaml:"image"`              // Image to check
	Tag        string    `yaml:"tag"`                // Tag to check for
	Username   string    `yaml:"username,omitempty"` // Username to get a new token
	Token      string    `yaml:"token,omitempty"`    // Token to get the token for the queries
	token      string    `yaml:"token,omitempty"`    // Token to use for the queries
	validUntil time.Time `yaml:"-"`                  // Time this token is valud until
}

// DockerTagCheck
func (r *Require) DockerTagCheck(
	version string,
) error {
	if r == nil || r.Docker == nil {
		return nil
	}
	var url string
	tag := r.Docker.GetTag(version)
	var req *http.Request
	token, err := r.Docker.GetToken()
	if err != nil {
		return fmt.Errorf("%s:%s - %s",
			r.Docker.Image, tag, err)
	}
	switch r.Docker.Type {
	case "hub":
		url = fmt.Sprintf("https://registry.hub.docker.com/v2/repositories/%s/tags/%s",
			r.Docker.Image, tag)
		req, _ = http.NewRequest(http.MethodGet, url, nil)
	case "ghcr":
		url = fmt.Sprintf("https://ghcr.io/v2/%s/manifests/%s",
			r.Docker.Image, tag)
		req, _ = http.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("Authorization", "Bearer "+token)
	case "quay":
		url = fmt.Sprintf("https://quay.io/api/v1/repository/%s/tag/?onlyActiveTags=true&specificTag=%s",
			r.Docker.Image, tag)
		req, _ = http.NewRequest(http.MethodGet, url, nil)
	}
	// Do the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s:%s - %s",
			r.Docker.Image, tag, err)
	}
	defer resp.Body.Close()
	// Parse the body
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s:%s - %s",
			r.Docker.Image, tag, string(body))
	}
	// Quay will give a 200 even when the tag doesn't exist
	if r.Docker.Type == "quay" && strings.Contains(string(body), `"tags": []`) {
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

	validTypes := []string{"hub", "quay", "ghcr"}
	if !utils.Contains(validTypes, d.Type) {
		errs = fmt.Errorf("%s%stype: %q <invalid> (should be hub/quay/ghcr)\\",
			utils.ErrorToString(errs), prefix, d.Type)
	}

	if d.Image == "" {
		errs = fmt.Errorf("%s%simage: <required> (image to check tags for)",
			utils.ErrorToString(errs), prefix)
	} else {
		regex := regexp.MustCompile(`^[\w\-\/]+$`)
		if !regex.MatchString(d.Image) {
			errs = fmt.Errorf("%s%simage: %q <invalid> (non-ASCII)\\",
				utils.ErrorToString(errs), prefix, d.Image)
		}
	}

	if d.Tag == "" {
		errs = fmt.Errorf("%s%stag: <required> (tag to check for existence)",
			utils.ErrorToString(errs), prefix)
	} else if !utils.CheckTemplate(d.Tag) {
		errs = fmt.Errorf("%s%stag: %q <invalid> (didn't pass templating)\\",
			utils.ErrorToString(errs), prefix, d.Tag)
	}

	return
}

func (d *DockerCheck) CheckToken() (err error) {
	if d == nil {
		return
	}

	switch d.Type {
	case "hub":
		_, err = d.GetToken()

	case "quay":
		if d.Token == "" {
			return fmt.Errorf("token: <required>")
		}
	case "ghcr":
	}

	if err != nil {
		err = fmt.Errorf("token: %q <invalid> (%s)",
			d.Token, err)
	}

	return
}

// GetTag to search for on Image
func (d *DockerCheck) GetTag(version string) string {
	return utils.TemplateString(d.Tag, utils.ServiceInfo{LatestVersion: version})
}

func (d *DockerCheck) GetToken() (string, error) {
	if time.Now().UTC().Before(d.validUntil) {
		return d.token, nil
	}
	var err error
	switch d.Type {
	case "hub":
		if err = d.refreshDockerHubToken(); err != nil {
			return "", err
		}
	case "ghcr":
		if d.Token != "" {
			if strings.HasPrefix(d.Token, "ghp_") {
				d.token = base64.StdEncoding.EncodeToString([]byte(d.Token))
			}
			d.validUntil = time.Now().AddDate(1, 0, 0)
		}
		// Get a NOOP token for public images by default
		if err = d.refreshGHCRToken(); err != nil {
			return "", err
		}
	case "quay":
		// create an org -> application
		d.token = d.Token
		d.validUntil = time.Now().AddDate(1, 0, 0)
	}

	return d.token, err
}

func (d *DockerCheck) refreshDockerHubToken() (err error) {
	// Get the http.Request
	url := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", d.Image)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if d.Username != "" {
		req.Header.Set("Authorization", "Basic "+utils.BasicAuth(d.Username, d.Token))
	}
	// Do the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		jLog.Error(err, utils.LogFrom{Primary: "docker-hub", Secondary: d.Image}, true)
		return
	}
	defer resp.Body.Close()
	// Parse the body
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(string(body))
	}
	type quayJSON struct {
		Token     string    `json:"token"`
		ExpiresIn int       `json:"expires_in"`
		IssuedAt  time.Time `json:"issued_at"`
	}
	var tokenJSON quayJSON
	err = json.Unmarshal(body, &tokenJSON)
	d.token = tokenJSON.Token
	d.validUntil = tokenJSON.IssuedAt.Add(time.Duration(tokenJSON.ExpiresIn) * time.Second)
	return err
}

func (d *DockerCheck) refreshGHCRToken() (err error) {
	url := fmt.Sprintf("https://ghcr.io/token?scope=repository:%s:pull", d.Image)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read the token
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(string(body))
	}
	type ghcrJSON struct {
		Token string `json:"token"`
	}
	var tokenJSON ghcrJSON
	err = json.Unmarshal(body, &tokenJSON)
	d.token = tokenJSON.Token
	return err
}
