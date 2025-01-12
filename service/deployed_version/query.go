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

// Package deployedver provides the deployed_version lookup.
package deployedver

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/web/metric"
)

// Track the deployed version (DeployedVersion) of the `parent`.
func (l *Lookup) Track() {
	if l == nil {
		return
	}
	logFrom := util.LogFrom{Primary: *l.Status.ServiceID}

	// Track forever.
	for {
		// If we are deleting this Service, stop tracking it.
		if l.Status.Deleting() {
			return
		}

		// Query the deployed version.
		deployedVersion, _ := l.Query(true, logFrom)
		// If new release found by ^ query.
		l.HandleNewVersion(deployedVersion, true)
		// Sleep interval between queries.
		time.Sleep(l.Options.GetIntervalDuration())
	}
}

// query the deployed version (DeployedVersion) of the Service.
func (l *Lookup) query(logFrom util.LogFrom) (string, error) {
	body, err := l.httpRequest(logFrom)
	if err != nil {
		return "", err
	}

	var version string
	// If JSON is provided, use it to extract the version.
	if l.JSON != "" {
		version, err = util.GetValueByKey(body, l.JSON, l.GetURL())
		if err != nil {
			jLog.Error(err, logFrom, true)
			//nolint:wrapcheck
			return "", err
		}
	} else {
		// Use the entire body if not parsing as JSON.
		version = string(body)
	}

	// If a regex is provided, use it to extract the version.
	if l.Regex != "" {
		re := regexp.MustCompile(l.Regex)
		texts := re.FindAllStringSubmatch(version, 1)

		if len(texts) == 0 {
			err := fmt.Errorf("regex %q didn't return any matches on %q",
				l.Regex, util.TruncateMessage(version, 100))
			jLog.Warn(err, logFrom, true)
			return "", err
		}

		regexMatches := texts[0]
		version = util.RegexTemplate(regexMatches, l.RegexTemplate)
	}

	// If semantic versioning is enabled, check the version is in the correct format.
	if l.Options.GetSemanticVersioning() {
		_, err = semver.NewVersion(version)
		if err != nil {
			err = fmt.Errorf("failed converting %q to a semantic version. If all "+
				"versions are in this style, consider adding json/regex to get the version into the "+
				"style of 'MAJOR.MINOR.PATCH' (https://semver.org/), or disabling semantic versioning "+
				"(globally with defaults.service.semantic_versioning or just for this service with the semantic_versioning var)",
				version)
			jLog.Error(err, logFrom, true)
			return "", err
		}
	}

	return version, nil
}

// Query the deployed version (DeployedVersion) of the Service.
func (l *Lookup) Query(metrics bool, logFrom util.LogFrom) (string, error) {
	version, err := l.query(logFrom)

	if metrics {
		l.queryMetrics(err == nil)
	}

	return version, err
}

// queryMetrics sets the Prometheus metrics for the DeployedVersion query.
func (l *Lookup) queryMetrics(successfulQuery bool) {
	if successfulQuery {
		metric.IncPrometheusCounter(metric.DeployedVersionQueryResultTotal,
			*l.Status.ServiceID,
			"",
			"",
			"SUCCESS")
		metric.SetPrometheusGauge(metric.DeployedVersionQueryResultLast,
			*l.Status.ServiceID, "",
			1)
	} else {
		metric.IncPrometheusCounter(metric.DeployedVersionQueryResultTotal,
			*l.Status.ServiceID,
			"",
			"",
			"FAIL")
		metric.SetPrometheusGauge(metric.DeployedVersionQueryResultLast,
			*l.Status.ServiceID, "",
			0)
	}
}

// HandleNewVersion performs a check for whether this `version` is new, and if so,
// checks whether this is later than LatestVersion and announces and updates `Status` accordingly.
func (l *Lookup) HandleNewVersion(version string, writeToDB bool) {
	// If the new version is the same as what we had, do nothing.
	if version == "" || version == l.Status.DeployedVersion() {
		return
	}

	// Set the new Deployed version.
	l.Status.SetDeployedVersion(version, "", writeToDB)

	// If this new version is not LatestVersion,
	// check it is not a later version than LatestVersion.
	latestVersion := l.Status.LatestVersion()
	if latestVersion == "" {
		l.Status.SetLatestVersion(l.Status.DeployedVersion(), l.Status.DeployedVersionTimestamp(), writeToDB)
		l.Status.AnnounceQueryNewVersion()
	} else if version != latestVersion &&
		l.Options.GetSemanticVersioning() {
		//#nosec G104 -- Disregard as deployedVersion will always be semantic if GetSemanticVersioning.
		deployedVersionSV, _ := semver.NewVersion(version)
		//#nosec G104 -- Disregard as LatestVersion will always be semantic if GetSemanticVersioning.
		latestVersionSV, _ := semver.NewVersion(latestVersion)

		// Update LatestVersion to DeployedVersion if newer.
		if latestVersionSV.LessThan(deployedVersionSV) {
			l.Status.SetLatestVersion(l.Status.DeployedVersion(), l.Status.DeployedVersionTimestamp(), writeToDB)
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

// httpRequest sends an HTTP request to the URL and returns the response body.
func (l *Lookup) httpRequest(logFrom util.LogFrom) ([]byte, error) {
	// HTTPS insecure skip verify.
	customTransport := &http.Transport{}
	if l.GetAllowInvalidCerts() {
		customTransport = http.DefaultTransport.(*http.Transport).Clone()
		//#nosec G402 -- explicitly wanted InsecureSkipVerify.
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// Create the request.
	req, err := http.NewRequest(l.Method, l.GetURL(), l.GetBody())
	if err != nil {
		jLog.Error(err, logFrom, true)
		return nil, err //nolint:wrapcheck
	}
	// Set headers.
	req.Header.Set("Connection", "close")
	for _, header := range l.Headers {
		req.Header.Set(util.EvalEnvVars(header.Key), util.EvalEnvVars(header.Value))
	}
	// Basic auth.
	if l.BasicAuth != nil {
		req.SetBasicAuth(util.EvalEnvVars(l.BasicAuth.Username), util.EvalEnvVars(l.BasicAuth.Password))
	}

	// Send the request.
	client := &http.Client{Transport: customTransport}
	resp, err := client.Do(req)
	if err != nil {
		// Don't crash on invalid certs.
		if strings.Contains(err.Error(), "x509") {
			err = fmt.Errorf("x509 (certificate invalid)")
			jLog.Warn(err, logFrom, true)
			return nil, err
		}
		jLog.Error(err, logFrom, true)
		return nil, err
	}

	// Ignore non-2XX responses.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("non-2XX response code: %d", resp.StatusCode)
		jLog.Warn(err, logFrom, true)
		return nil, err
	}

	// Read the response body.
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	jLog.Error(err, logFrom, err != nil)
	return body, err //nolint:wrapcheck
}
