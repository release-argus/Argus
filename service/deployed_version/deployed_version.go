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

package deployed_version

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/web/metrics"
)

// GetAllowInvalidCerts returns whether invalid HTTPS certs are allowed.
func (l *Lookup) GetAllowInvalidCerts() bool {
	return *utils.GetFirstNonNilPtr(l.AllowInvalidCerts, l.Defaults.AllowInvalidCerts, l.HardDefaults.AllowInvalidCerts)
}

// Track the deployed version (DeployedVersion) of the `parent`.
func (l *Lookup) Track() {
	if l == nil {
		return
	}
	logFrom := utils.LogFrom{Primary: *l.Status.ServiceID}

	// Track forever.
	for {
		deployedVersion, err := l.Query(logFrom)
		// If new release found by ^ query.
		if err == nil {
			metrics.IncreasePrometheusCounterWithIDAndResult(metrics.DeployedVersionQueryMetric, *l.Status.ServiceID, "SUCCESS")
			metrics.SetPrometheusGaugeWithID(metrics.DeployedVersionQueryLiveness, *l.Status.ServiceID, 1)
			if deployedVersion != l.Status.DeployedVersion {
				// Announce the updated deployment
				l.Status.SetDeployedVersion(deployedVersion)

				// If this new deployedVersion isn't LatestVersion
				// Check that it's not a later version than LatestVersion
				if deployedVersion != l.Status.LatestVersion && l.Options.GetSemanticVersioning() && l.Status.LatestVersion != "" {
					//#nosec G104 -- Disregard as deployedVersion will always be semantic if GetSemanticVersioning
					//nolint:errcheck // ^
					deployedVersionSV, _ := semver.NewVersion(deployedVersion)
					//#nosec G104 -- Disregard as LatestVersion will always be semantic if GetSemanticVersioning
					//nolint:errcheck // ^
					latestVersionSV, _ := semver.NewVersion(l.Status.LatestVersion)

					// Update LatestVersion to DeployedVersion if it's newer
					if latestVersionSV.LessThan(*deployedVersionSV) {
						l.Status.SetLatestVersion(l.Status.DeployedVersion)
						l.Status.LatestVersionTimestamp = l.Status.DeployedVersionTimestamp
						l.Status.AnnounceQueryNewVersion()
					}
				} else if l.Status.LatestVersion == "" {
					l.Status.SetLatestVersion(l.Status.DeployedVersion)
					l.Status.LatestVersionTimestamp = l.Status.DeployedVersionTimestamp
					l.Status.AnnounceQueryNewVersion()
				}

				// Announce version change to WebSocket clients.
				jLog.Info(
					fmt.Sprintf("Updated to %q", deployedVersion),
					logFrom,
					true)
				l.Status.AnnounceUpdate()
			}
		} else {
			metrics.IncreasePrometheusCounterWithIDAndResult(metrics.DeployedVersionQueryMetric, *l.Status.ServiceID, "FAIL")
			metrics.SetPrometheusGaugeWithID(metrics.DeployedVersionQueryLiveness, *l.Status.ServiceID, 0)
		}
		// Sleep interval between queries.
		time.Sleep(l.Options.GetIntervalDuration())
	}
}

// Query the deployed version (DeployedVersion) of the Service.
func (l *Lookup) Query(logFrom utils.LogFrom) (string, error) {
	rawBody, err := l.httpRequest(logFrom)
	if err != nil {
		return "", err
	}

	var version string
	if l.JSON != "" {
		jsonKeys := strings.Split(l.JSON, ".")
		var queriedJSON map[string]interface{}
		err := json.Unmarshal(rawBody, &queriedJSON)
		if err != nil {
			err := fmt.Errorf("failed to unmarshal the following from %q into json:%s",
				l.URL,
				string(rawBody))
			jLog.Error(err, logFrom, true)
			return "", err
		}

		// birds := result["birds"].(map[string]interface{})
		for k := range jsonKeys {
			if queriedJSON[jsonKeys[k]] == nil {
				err := fmt.Errorf("%q could not be found in the following JSON. Failed at %q:\n%s",
					l.JSON,
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

	if l.Regex != "" {
		re := regexp.MustCompile(l.Regex)
		texts := re.FindStringSubmatch(version)
		index := 1

		if len(texts) == 0 {
			err := fmt.Errorf("%q regex didn't return any matches in %q",
				l.Regex,
				version)
			jLog.Warn(err, logFrom, true)
			return "", err
		} else if len(texts) == 1 {
			// no capture group in regex
			index = 0
		}

		version = texts[index]
	}

	if l.Options.GetSemanticVersioning() {
		_, err = semver.NewVersion(version)
		if err != nil {
			err = fmt.Errorf("failed converting %q to a semantic version. If all versions are in this style, consider adding json/regex to get the version into the style of 'MAJOR.MINOR.PATCH' (https://semver.org/), or disabling semantic versioning (globally with defaults.service.semantic_versioning or just for this service with the semantic_versioning var)",
				version)
			jLog.Error(err, logFrom, true)
			return "", err
		}
	}

	return version, nil
}

func (l *Lookup) httpRequest(logFrom utils.LogFrom) (rawBody []byte, err error) {
	// HTTPS insecure skip verify.
	customTransport := &http.Transport{}
	if l.GetAllowInvalidCerts() {
		customTransport = http.DefaultTransport.(*http.Transport).Clone()
		//#nosec G402 -- explicitly wanted InsecureSkipVerify
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	req, err := http.NewRequest(http.MethodGet, l.URL, nil)
	if err != nil {
		jLog.Error(err, logFrom, true)
		return
	}

	// Set headers
	for _, header := range l.Headers {
		req.Header.Set(header.Key, header.Value)
	}

	// Basic auth
	if l.BasicAuth != nil {
		req.SetBasicAuth((*l.BasicAuth).Username, (*l.BasicAuth).Password)
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
	rawBody, err = io.ReadAll(resp.Body)
	jLog.Error(err, logFrom, err != nil)
	return
}
