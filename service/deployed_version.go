// Copyright [2022] [Hymenaios]
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

	"github.com/hymenaios-io/Hymenaios/utils"
)

// GetAllowInvalidCerts returns whether invalid HTTPS certs are allowed.
func (c *DeployedVersionLookup) GetAllowInvalidCerts() bool {
	return *utils.GetFirstNonNilPtr(c.AllowInvalidCerts, c.Defaults.AllowInvalidCerts, c.HardDefaults.AllowInvalidCerts)
}

// Track the deployed version (CurrentVersion) of the `parent`.
func (c *DeployedVersionLookup) Track(parent *Service) {
	if c == nil {
		return
	}
	logFrom := utils.LogFrom{Primary: *parent.ID}

	// Track forever.
	for {
		// If new release found by this query.
		currentVersion, err := c.Query(logFrom)

		if err == nil && currentVersion != utils.DefaultIfNil(parent.Status.CurrentVersion) {
			parent.Status.SetCurrentVersion(currentVersion)

			// Announce version change to WebSocket clients
			jLog.Info(
				fmt.Sprintf("Updated to %q", currentVersion),
				logFrom,
				true)
			parent.AnnounceUpdate()
			if parent.SaveChannel != nil {
				*parent.SaveChannel <- true
			}
		}
		// Sleep interval between checks.
		time.Sleep(parent.GetIntervalDuration())
	}
}

// Query the deployed version (CurrentVersion) of the Service.
func (c *DeployedVersionLookup) Query(logFrom utils.LogFrom) (string, error) {
	customTransport := &http.Transport{}
	// HTTPS insecure skip verify.
	if c.GetAllowInvalidCerts() {
		customTransport = http.DefaultTransport.(*http.Transport).Clone()
		//#nosec G402 -- explicitly wanted InsecureSkipVerify
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	req, err := http.NewRequest(http.MethodGet, c.URL, nil)
	if err != nil {
		jLog.Error(err, logFrom, true)
		return "", err
	}

	// Set headers
	for _, header := range c.Headers {
		req.Header.Set(header.Key, header.Value)
	}

	// Basic auth
	if c.BasicAuth != nil {
		req.SetBasicAuth((*c.BasicAuth).Username, (*c.BasicAuth).Password)
	}

	client := &http.Client{Transport: customTransport}
	resp, err := client.Do(req)

	if err != nil {
		// Don't crash on invalid certs.
		if strings.Contains(err.Error(), "x509") {
			err = fmt.Errorf("x509 (Cert invalid)")
			jLog.Warn(err, logFrom, true)
			return "", err
		}
		jLog.Error(err, logFrom, true)
		return "", err
	}

	// Read the response body.
	rawBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		jLog.Error(err, logFrom, true)
		return "", err
	}

	var version string
	if c.JSON != "" {
		jsonKeys := strings.Split(c.JSON, ".")
		var queriedJSON map[string]interface{}
		err := json.Unmarshal(rawBody, &queriedJSON)
		if err != nil {
			err := fmt.Errorf("Failed to unmarshal the following from %q into JSON:%s",
				c.URL,
				string(rawBody))
			jLog.Error(err, logFrom, true)
			return "", err
		}

		// birds := result["birds"].(map[string]interface{})
		for k := range jsonKeys {
			if queriedJSON[jsonKeys[k]] == nil {
				err := fmt.Errorf("%q could not be found in the following JSON. Failed at %q:\n%s",
					c.JSON,
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

	if c.Regex != "" {
		re := regexp.MustCompile(c.Regex)
		texts := re.FindStringSubmatch(version)

		if len(texts) < 2 {
			err := fmt.Errorf("%q RegEx didn't return any matches in %q",
				c.Regex,
				version)
			jLog.Warn(err, logFrom, true)
			return "", err
		}

		version = texts[1]
	}

	return version, nil
}

// Print will print the DeployedVersionLookup.
func (c *DeployedVersionLookup) Print(prefix string) {
	if c == nil {
		return
	}
	fmt.Printf("%sdeployed_version:\n", prefix)
	prefix += "  "

	utils.PrintlnIfNotDefault(c.URL, fmt.Sprintf("%surl: %s", prefix, c.URL))
	utils.PrintlnIfNotNil(c.AllowInvalidCerts, fmt.Sprintf("%sallow_invalid_certs: %t", prefix, utils.DefaultIfNil(c.AllowInvalidCerts)))
	if c.BasicAuth != nil {
		fmt.Printf("%sbasic_auth:\n", prefix)
		fmt.Printf("%s  username: %s\n", prefix, c.BasicAuth.Username)
		fmt.Printf("%s  password: <secret>\n", prefix)
	}
	if c.Headers != nil {
		fmt.Printf("%sheaders:\n", prefix)
		for _, header := range c.Headers {
			fmt.Printf("%s  - key: %s\n", prefix, header.Key)
			fmt.Printf("%s    value: <secret>\n", prefix)
		}
	}
	utils.PrintlnIfNotDefault(c.JSON, fmt.Sprintf("%sjson: %s", prefix, c.URL))
	utils.PrintlnIfNotDefault(c.Regex, fmt.Sprintf("%sregex: %s", prefix, c.URL))
}

// CheckValues of the DeployedVersionLookup.
func (c *DeployedVersionLookup) CheckValues(prefix string) (errs error) {
	if c == nil {
		return
	}

	// URL
	if c.URL == "" && c.Defaults != nil {
		errs = fmt.Errorf("%s%s  url: <missing> (URL to get the current_version is required)\\", utils.ErrorToString(errs), prefix)
	}

	// RegEx
	_, err := regexp.Compile(c.Regex)
	if err != nil {
		errs = fmt.Errorf("%s%s  regex: <invalid> %q (Invalid RegEx)\\", utils.ErrorToString(errs), prefix, c.Regex)
	}

	if errs != nil {
		errs = fmt.Errorf("%sdeployed_version:\\%w", prefix, errs)
	}
	return
}
