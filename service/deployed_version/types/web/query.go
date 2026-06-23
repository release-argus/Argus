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

// Package web provides a web-based lookup type.
package web

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/httpx"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/util"
)

// Track polls the URL at the configured interval, updating the deployed version on each query.
func (l *Lookup) Track() {
	logFrom := logx.LogFrom{Primary: l.GetServiceID()}

	// Track forever.
	for {
		// If we are deleting this Service, stop tracking it.
		if l.Status.Deleting() {
			return
		}

		// Query the deployed version.
		_ = l.Query(true, logFrom) //nolint:errcheck

		// Sleep interval between queries.
		time.Sleep(l.Options.GetIntervalDuration())
	}
}

// Query fetches the deployed version, sets Prometheus metrics if requested, and returns any error.
func (l *Lookup) Query(metrics bool, logFrom logx.LogFrom) error {
	err := l.query(metrics, logFrom)

	if metrics {
		l.QueryMetrics(l, err)
	}

	return err
}

// query fetches the deployed version URL and updates DeployedVersion if changed.
func (l *Lookup) query(writeToDB bool, logFrom logx.LogFrom) error {
	body, err := l.httpRequest(logFrom)
	if err != nil {
		return err
	}

	version, err := l.getVersion(body, logFrom)
	if err != nil {
		return err
	}

	// Set the deployed version if it has changed.
	l.HandleNewVersion(version, "", writeToDB, logFrom) //nolint:wrapcheck

	return nil
}

// httpRequest makes a HTTP GET request to the URL and returns the body.
func (l *Lookup) httpRequest(logFrom logx.LogFrom) ([]byte, error) {
	client := httpx.Client
	// HTTPS insecure skip verify.
	if l.allowInvalidCerts() {
		client = httpx.InsecureClient
	}

	// Create the request.
	req, err := http.NewRequest(l.method(), l.url(), l.body())
	if err != nil {
		err = fmt.Errorf(
			"failed creating http request for %q: %w",
			l.URL, err,
		)
		logx.Error(err, logFrom, true)
		return nil, err
	}

	// Set headers.
	for _, header := range l.Headers {
		req.Header.Set(
			util.EvalEnvVars(header.Key),
			util.EvalEnvVars(header.Value),
		)
	}
	// Basic auth.
	if l.BasicAuth != nil {
		req.SetBasicAuth(
			util.EvalEnvVars(l.BasicAuth.Username),
			util.EvalEnvVars(l.BasicAuth.Password),
		)
	}

	// Send the request.
	resp, err := client.Do(req)
	if err != nil {
		// Don't crash on invalid certs.
		if strings.Contains(err.Error(), "x509") {
			err = errors.New("x509 (certificate invalid)")
			logx.Warn(err, logFrom, true)
			return nil, err
		}
		logx.Error(err, logFrom, true)
		return nil, err
	}
	defer resp.Body.Close()

	if l.TargetHeader != "" {
		if headerValue := resp.Header.Get(l.TargetHeader); headerValue != "" {
			return []byte(headerValue), nil
		}
		var err error
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			err = fmt.Errorf("target header %q not found (status: %d)", l.TargetHeader, resp.StatusCode)
		} else {
			err = fmt.Errorf("target header %q not found", l.TargetHeader)
		}
		logx.Warn(err, logFrom, true)
		return nil, err
	}

	// Ignore non-2XX responses.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := fmt.Errorf("non-2XX response code: %d", resp.StatusCode)
		logx.Warn(err, logFrom, true)
		return nil, err
	}

	// Return the body.
	body, err := io.ReadAll(io.LimitReader(resp.Body, 50<<20)) // Limit to 50 MiB.
	logx.Error(err, logFrom, err != nil)
	return body, err //nolint:wrapcheck
}

// getVersion returns the version from `body` that matches the URLCommands, and Regex requirements.
func (l *Lookup) getVersion(body []byte, logFrom logx.LogFrom) (string, error) {
	var version string
	// If JSON is provided, use it to extract the version.
	if l.JSON != "" {
		var err error
		version, err = decode.GetValueByKey(body, l.JSON, l.url())
		if err != nil {
			logx.Error(err, logFrom, true)
			//nolint:wrapcheck
			return "", err
		}
	} else {
		// Use the entire body if not parsing as JSON.
		version = string(body)
	}

	if version == "" {
		err := fmt.Errorf(
			"no version found in %q",
			util.TruncateMessage(version, 100),
		)
		logx.Warn(err, logFrom, true)
		return "", err
	}

	// If a regex is provided, use it to extract the version.
	if l.Regex != "" {
		re := regexp.MustCompile(l.Regex)
		texts := re.FindAllStringSubmatch(version, 1)

		if len(texts) == 0 {
			err := fmt.Errorf(
				"regex %q didn't return any matches on %q",
				l.Regex, util.TruncateMessage(version, 100),
			)
			logx.Warn(err, logFrom, true)
			return "", err
		}

		regexMatches := texts[0]
		version = util.RegexTemplate(regexMatches, l.RegexTemplate)
	}

	// If semantic versioning is enabled, check the version is in the correct format.
	if l.Options.GetSemanticVersioning() {
		if _, err := l.Options.VerifySemanticVersioning(version, logFrom); err != nil {
			logx.Warn(err, logFrom, true)
			return "", err //nolint:wrapcheck
		}
	}

	return version, nil
}
