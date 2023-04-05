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

package deployedver

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
	"github.com/release-argus/Argus/util"
	metric "github.com/release-argus/Argus/web/metrics"
)

// Track the deployed version (DeployedVersion) of the `parent`.
func (l *Lookup) Track() {
	if l == nil {
		return
	}
	logFrom := util.LogFrom{Primary: *l.Status.ServiceID}

	// Track forever.
	for {
		// If we're deleting this Service, stop tracking it.
		if l.Status.Deleting {
			return
		}

		// Query the deployed version.
		deployedVersion, _ := l.Query(true, &logFrom)
		// If new release found by ^ query.
		l.HandleNewVersion(deployedVersion, true)
		// Sleep interval between queries.
		time.Sleep(l.Options.GetIntervalDuration())
	}
}

// query the deployed version (DeployedVersion) of the Service.
func (l *Lookup) query(logFrom *util.LogFrom) (string, error) {
	rawBody, err := l.httpRequest(logFrom)
	if err != nil {
		return "", err
	}

	var version string
	// If JSON is provided, use it to extract the version.
	if l.JSON != "" {
		jsonKeys := strings.Split(l.JSON, ".")
		var queriedJSON map[string]interface{}
		err := json.Unmarshal(rawBody, &queriedJSON)
		// If the JSON is invalid, return an error.
		if err != nil {
			err := fmt.Errorf("failed to unmarshal the following from %q into json:%s",
				l.URL, string(rawBody))
			jLog.Error(err, *logFrom, true)
			return "", err
		}

		// Iterate through the keys.
		for k := range jsonKeys {
			// If the key doesn't exist, return an error.
			if queriedJSON[jsonKeys[k]] == nil {
				err := fmt.Errorf("%q could not be found in the following JSON. Failed at %q:\n%s",
					l.JSON, jsonKeys[k], string(rawBody))
				jLog.Warn(err, *logFrom, true)
				return "", err
			}

			switch v := queriedJSON[jsonKeys[k]].(type) {
			// If the key is a string, int, float32, or float64, return it.
			case string, int, float32, float64:
				version = fmt.Sprint(queriedJSON[jsonKeys[k]])
				// If the key is a map, set it as the queriedJSON.
			case map[string]interface{}:
				queriedJSON = v
			}
		}
	} else {
		// Use the whole body if not parsing as JSON.
		version = string(rawBody)
	}

	// If a regex is provided, use it to extract the version.
	if l.Regex != "" {
		re := regexp.MustCompile(l.Regex)
		texts := re.FindStringSubmatch(version)
		index := 1

		if len(texts) == 0 {
			err := fmt.Errorf("regex %q didn't return any matches on %q",
				l.Regex, version)
			jLog.Warn(err, *logFrom, true)
			return "", err
		} else if len(texts) == 1 {
			// no capture group in regex
			index = 0
		}

		version = texts[index]
	}

	// If semantic versioning is enabled, check that the version is in the correct format.
	if l.Options.GetSemanticVersioning() {
		_, err = semver.NewVersion(version)
		if err != nil {
			err = fmt.Errorf("failed converting %q to a semantic version. If all "+
				"versions are in this style, consider adding json/regex to get the version into the "+
				"style of 'MAJOR.MINOR.PATCH' (https://semver.org/), or disabling semantic versioning "+
				"(globally with defaults.service.semantic_versioning or just for this service with the semantic_versioning var)",
				version)
			jLog.Error(err, *logFrom, true)
			return "", err
		}
	}

	return version, nil
}

// Query the deployed version (DeployedVersion) of the Service.
func (l *Lookup) Query(metrics bool, logFrom *util.LogFrom) (version string, err error) {
	version, err = l.query(logFrom)

	if metrics {
		l.queryMetrics(err == nil)
	}

	return
}

// queryMetrics sets the Prometheus metrics for the DeployedVersion query.
func (l *Lookup) queryMetrics(successfulQuery bool) {
	if successfulQuery {
		metric.IncreasePrometheusCounter(metric.DeployedVersionQueryMetric,
			*l.Status.ServiceID,
			"",
			"",
			"SUCCESS")
		metric.SetPrometheusGauge(metric.DeployedVersionQueryLiveness,
			*l.Status.ServiceID,
			1)
	} else {
		metric.IncreasePrometheusCounter(metric.DeployedVersionQueryMetric,
			*l.Status.ServiceID,
			"",
			"",
			"FAIL")
		metric.SetPrometheusGauge(metric.DeployedVersionQueryLiveness,
			*l.Status.ServiceID,
			0)
	}
}

// HandleNewVersion performs a check for whether this `version` is new, and if so,
// checks whether this is later than LatestVersion and announces and updates `Status` accordingly.
func (l *Lookup) HandleNewVersion(version string, writeToDB bool) {
	// If the new version is the same as what we had, do nothing.
	if version == "" || version == l.Status.GetDeployedVersion() {
		return
	}

	// Set the new Deployed version.
	l.Status.SetDeployedVersion(version, writeToDB)

	// If this new version isn't LatestVersion
	// Check that it's not a later version than LatestVersion
	latestVersion := l.Status.GetLatestVersion()
	if latestVersion == "" {
		l.Status.SetLatestVersion(l.Status.GetDeployedVersion(), writeToDB)
		l.Status.SetLatestVersionTimestamp(l.Status.GetDeployedVersionTimestamp())
		l.Status.AnnounceQueryNewVersion()
	} else if version != latestVersion &&
		l.Options.GetSemanticVersioning() {
		//#nosec G104 -- Disregard as deployedVersion will always be semantic if GetSemanticVersioning
		deployedVersionSV, _ := semver.NewVersion(version)
		//#nosec G104 -- Disregard as LatestVersion will always be semantic if GetSemanticVersioning
		latestVersionSV, _ := semver.NewVersion(latestVersion)

		// Update LatestVersion to DeployedVersion if it's newer
		if latestVersionSV.LessThan(*deployedVersionSV) {
			l.Status.SetLatestVersion(l.Status.GetDeployedVersion(), writeToDB)
			l.Status.SetLatestVersionTimestamp(l.Status.GetDeployedVersionTimestamp())
			l.Status.AnnounceQueryNewVersion()
		}
	}

	// Announce version change to WebSocket clients.
	jLog.Info(
		fmt.Sprintf("Updated to %q", version),
		util.LogFrom{Primary: *l.Status.ServiceID},
		true)
	l.Status.AnnounceUpdate()
}

func (l *Lookup) httpRequest(logFrom *util.LogFrom) (rawBody []byte, err error) {
	// HTTPS insecure skip verify.
	customTransport := &http.Transport{}
	if l.GetAllowInvalidCerts() {
		customTransport = http.DefaultTransport.(*http.Transport).Clone()
		//#nosec G402 -- explicitly wanted InsecureSkipVerify
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	req, err := http.NewRequest(http.MethodGet, l.URL, nil)
	if err != nil {
		jLog.Error(err, *logFrom, true)
		return
	}

	// Set headers
	req.Header.Set("Connection", "close")
	for _, header := range l.Headers {
		req.Header.Set(header.Key, header.Value)
	}

	// Basic auth
	if l.BasicAuth != nil {
		req.SetBasicAuth(l.BasicAuth.Username, l.BasicAuth.Password)
	}

	client := &http.Client{Transport: customTransport}
	resp, err := client.Do(req)
	if err != nil {
		// Don't crash on invalid certs.
		if strings.Contains(err.Error(), "x509") {
			err = fmt.Errorf("x509 (certificate invalid)")
			jLog.Warn(err, *logFrom, true)
			return
		}
		jLog.Error(err, *logFrom, true)
		return
	}

	// Read the response body.
	defer resp.Body.Close()
	rawBody, err = io.ReadAll(resp.Body)
	jLog.Error(err, *logFrom, err != nil)
	return
}
