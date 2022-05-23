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

package service

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/release-argus/Argus/utils"
)

// GetAllowInvalidCerts returns whether invalid HTTPS certs are allowed.
func (d *DeployedVersionLookup) GetAllowInvalidCerts() bool {
	return *utils.GetFirstNonNilPtr(d.AllowInvalidCerts, d.Defaults.AllowInvalidCerts, d.HardDefaults.AllowInvalidCerts)
}

// Track the deployed version (DeployedVersion) of the `parent`.
func (d *DeployedVersionLookup) Track(parent *Service) {
	if d == nil {
		return
	}
	logFrom := utils.LogFrom{Primary: *parent.ID}

	// Track forever.
	for {
		deployedVersion, err := d.Query(logFrom, parent.GetSemanticVersioning())
		// If new release found by ^ query.
		if err == nil && deployedVersion != parent.Status.DeployedVersion {
			// If this new deployedVersion isn't LatestVersion
			// Check that it's not a later version than LatestVersion
			if deployedVersion != parent.Status.LatestVersion && parent.GetSemanticVersioning() {
				//#nosec G104 -- Disregard as deployedVersion will always be semantic if GetSemanticVersioning
				//nolint:errcheck // ^
				deployedVersionSV, _ := semver.NewVersion(deployedVersion)
				//#nosec G104 -- Disregard as LatestVersion will always be semantic if GetSemanticVersioning
				//nolint:errcheck // ^
				latestVersionSV, _ := semver.NewVersion(parent.Status.LatestVersion)

				// Update LatestVersion to DeployedVersion if it's newer
				if deployedVersionSV.LessThan(*latestVersionSV) {
					parent.Status.LatestVersion = parent.Status.DeployedVersion
					parent.Status.LatestVersionTimestamp = parent.Status.DeployedVersionTimestamp
					parent.AnnounceQueryNewVersion()
				}
			}
			// Announce the updated deployment
			parent.Status.SetDeployedVersion(deployedVersion)

			// Announce version change to WebSocket clients.
			jLog.Info(
				fmt.Sprintf("Updated to %q", deployedVersion),
				logFrom,
				true)
			parent.AnnounceUpdate()
			if parent.SaveChannel != nil {
				*parent.SaveChannel <- true
			}
		}
		// Sleep interval between queries.
		time.Sleep(parent.GetIntervalDuration())
	}
}

// Query the deployed version (DeployedVersion) of the Service.
func (d *DeployedVersionLookup) Query(logFrom utils.LogFrom, semanticVersioning bool) (string, error) {
	rawBody, err := d.httpRequest(logFrom)
	if err != nil {
		return "", err
	}

	var version string
	if d.JSON != "" {
		jsonKeys := strings.Split(d.JSON, ".")
		var queriedJSON map[string]interface{}
		err := json.Unmarshal(rawBody, &queriedJSON)
		if err != nil {
			err := fmt.Errorf("Failed to unmarshal the following from %q into JSON:%s",
				d.URL,
				string(rawBody))
			jLog.Error(err, logFrom, true)
			return "", err
		}

		// birds := result["birds"].(map[string]interface{})
		for k := range jsonKeys {
			if queriedJSON[jsonKeys[k]] == nil {
				err := fmt.Errorf("%q could not be found in the following JSON. Failed at %q:\n%s",
					d.JSON,
					jsonKeys[k],
					string(rawBody))
				jLog.Warn(err, logFrom, true)
				return "", err
			}
			switch queriedJSON[jsonKeys[k]].(type) {
			case string, int, float32, float64:
				version = fmt.Sprintf("%v", queriedJSON[jsonKeys[k]])
			case map[string]interface{}:
				queriedJSON = queriedJSON[jsonKeys[k]].(map[string]interface{})
			}
		}
	} else {
		// Use the whole body if not parsing as JSON.
		version = string(rawBody)
	}

	if d.Regex != "" {
		re := regexp.MustCompile(d.Regex)
		texts := re.FindStringSubmatch(version)

		if len(texts) < 2 {
			err := fmt.Errorf("%q RegEx didn't return any matches in %q",
				d.Regex,
				version)
			jLog.Warn(err, logFrom, true)
			return "", err
		}

		version = texts[1]
	}

	if semanticVersioning {
		_, err = semver.NewVersion(version)
		if err != nil {
			err = fmt.Errorf("failed converting %q to a semantic version. If all versions are in this style, consider adding json/regex to get the version into the style of 'MAJOR.MINOR.PATCH' (https://semver.org/), or disabling semantic versioning (globally with defaults.service.semantic_versioning or just for this service with the semantic_versioning var)", version)
			jLog.Error(err, logFrom, true)
			return "", err
		}
	}

	return version, nil
}

func (d *DeployedVersionLookup) httpRequest(logFrom utils.LogFrom) (rawBody []byte, err error) {
	customTransport := &http.Transport{}
	// HTTPS insecure skip verify.
	if d.GetAllowInvalidCerts() {
		customTransport = http.DefaultTransport.(*http.Transport).Clone()
		//#nosec G402 -- explicitly wanted InsecureSkipVerify
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	req, err := http.NewRequest(http.MethodGet, d.URL, nil)
	if err != nil {
		jLog.Error(err, logFrom, true)
		return
	}

	// Set headers
	for _, header := range d.Headers {
		req.Header.Set(header.Key, header.Value)
	}

	// Basic auth
	if d.BasicAuth != nil {
		req.SetBasicAuth((*d.BasicAuth).Username, (*d.BasicAuth).Password)
	}

	client := &http.Client{Transport: customTransport}
	resp, err := client.Do(req)
	if err != nil {
		// Don't crash on invalid certs.
		if strings.Contains(err.Error(), "x509") {
			err = fmt.Errorf("x509 (Cert invalid)")
			jLog.Warn(err, logFrom, true)
			return
		}
		jLog.Error(err, logFrom, true)
		return
	}

	// Read the response body.
	rawBody, err = ioutil.ReadAll(resp.Body)
	jLog.Error(err, logFrom, err != nil)
	return
}

// Print will print the DeployedVersionLookup.
func (d *DeployedVersionLookup) Print(prefix string) {
	if d == nil {
		return
	}
	fmt.Printf("%sdeployed_version:\n", prefix)
	prefix += "  "

	utils.PrintlnIfNotDefault(d.URL, fmt.Sprintf("%surl: %s", prefix, d.URL))
	utils.PrintlnIfNotNil(d.AllowInvalidCerts, fmt.Sprintf("%sallow_invalid_certs: %t", prefix, utils.DefaultIfNil(d.AllowInvalidCerts)))
	if d.BasicAuth != nil {
		fmt.Printf("%sbasic_auth:\n", prefix)
		fmt.Printf("%s  username: %s\n", prefix, d.BasicAuth.Username)
		fmt.Printf("%s  password: <secret>\n", prefix)
	}
	if d.Headers != nil {
		fmt.Printf("%sheaders:\n", prefix)
		for _, header := range d.Headers {
			fmt.Printf("%s  - key: %s\n", prefix, header.Key)
			fmt.Printf("%s    value: <secret>\n", prefix)
		}
	}
	utils.PrintlnIfNotDefault(d.JSON, fmt.Sprintf("%sjson: %s", prefix, d.URL))
	utils.PrintlnIfNotDefault(d.Regex, fmt.Sprintf("%sregex: %s", prefix, d.URL))
}

// CheckValues of the DeployedVersionLookup.
func (d *DeployedVersionLookup) CheckValues(prefix string) (errs error) {
	if d == nil {
		return
	}

	// URL
	if d.URL == "" && d.Defaults != nil {
		errs = fmt.Errorf("%s%s  url: <missing> (URL to get the deployed_version is required)\\", utils.ErrorToString(errs), prefix)
	}

	// RegEx
	_, err := regexp.Compile(d.Regex)
	if err != nil {
		errs = fmt.Errorf("%s%s  regex: <invalid> %q (Invalid RegEx)\\", utils.ErrorToString(errs), prefix, d.Regex)
	}

	if errs != nil {
		errs = fmt.Errorf("%sdeployed_version:\\%w", prefix, errs)
	}
	return
}
