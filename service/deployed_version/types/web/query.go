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

// Package web provides a web-based lookup type.
package web

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

// Track the deployed version (DeployedVersion) of the `parent`.
func (l *Lookup) Track() {
	logFrom := logutil.LogFrom{Primary: l.GetServiceID()}

	// Track forever.
	for {
		// If we are deleting this Service, stop tracking it.
		if l.Status.Deleting() {
			return
		}

		// Query the deployed version.
		l.Query(true, logFrom) //nolint:errcheck

		// Sleep interval between queries.
		time.Sleep(l.Options.GetIntervalDuration())
	}
}

// Query queries the source,
// and returns whether a new release was found, and updates LatestVersion if so.
//
// Parameters:
//
//	metrics: if true, set Prometheus metrics based on the query.
func (l *Lookup) Query(metrics bool, logFrom logutil.LogFrom) error {
	err := l.query(metrics, logFrom)

	if metrics {
		l.QueryMetrics(l, err)
	}

	return err
}

// Query queries the source,
// and returns whether a new release was found, and updates LatestVersion if so.
func (l *Lookup) query(writeToDB bool, logFrom logutil.LogFrom) error {
	body, err := l.httpRequest(logFrom)
	if err != nil {
		return err
	}

	version, err := l.getVersion(body, logFrom)
	if err != nil {
		logutil.Log.Error(err, logFrom, true)
		return err
	}

	// Set the deployed version if it has changed.
	if previousVersion := l.Status.DeployedVersion(); version != previousVersion {
		l.HandleNewVersion(version, "", writeToDB, logFrom) //nolint:wrapcheck
	}

	return nil
}

// httpRequest makes a HTTP GET request to the URL, and returns the body.
func (l *Lookup) httpRequest(logFrom logutil.LogFrom) ([]byte, error) {
	customTransport := &http.Transport{}
	// HTTPS insecure skip verify.
	if l.allowInvalidCerts() {
		customTransport = http.DefaultTransport.(*http.Transport).Clone()
		//#nosec G402 -- explicitly wanted InsecureSkipVerify
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// Create the request.
	req, err := http.NewRequest(l.Method, l.url(), l.body())
	if err != nil {
		err = fmt.Errorf("failed creating http request for %q: %w",
			l.URL, err)
		logutil.Log.Error(err, logFrom, true)
		return nil, err
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
			err = errors.New("x509 (certificate invalid)")
			logutil.Log.Warn(err, logFrom, true)
			return nil, err
		}
		logutil.Log.Error(err, logFrom, true)
		return nil, err
	}

	// Ignore non-2XX responses.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("non-2XX response code: %d", resp.StatusCode)
		logutil.Log.Warn(err, logFrom, true)
		return nil, err
	}

	// Read the response body.
	defer resp.Body.Close()
	// If we're targeting a specific header, ignore the body.
	if l.TargetHeader != "" {
		if headerValue := resp.Header.Get(l.TargetHeader); headerValue != "" {
			return []byte(headerValue), nil
		}
		return nil, fmt.Errorf("target header %q not found", l.TargetHeader)
	}

	// Return the body.
	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // Limit to 10 MB.
	logutil.Log.Error(err, logFrom, err != nil)
	return body, err //nolint:wrapcheck
}

// getVersion returns the latest version from `body` that matches the URLCommands, and Regex requirements.
func (l *Lookup) getVersion(body []byte, logFrom logutil.LogFrom) (version string, err error) {
	// If JSON is provided, use it to extract the version.
	if l.JSON != "" {
		version, err = util.GetValueByKey(body, l.JSON, l.url())
		if err != nil {
			logutil.Log.Error(err, logFrom, true)
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
			logutil.Log.Warn(err, logFrom, true)
			return "", err
		}

		regexMatches := texts[0]
		version = util.RegexTemplate(regexMatches, l.RegexTemplate)
	}

	// If semantic versioning is enabled, check the version is in the correct format.
	if l.Options.GetSemanticVersioning() {
		if _, err := l.Options.VerifySemanticVersioning(version, logFrom); err != nil {
			return "", err //nolint:wrapcheck
		}
	}

	return version, nil
}
