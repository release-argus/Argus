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
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

// Query queries the source
// and returns whether a new release was found, updating LatestVersion if so.
//
// Parameters:
//
//	metrics: if true, set Prometheus metrics based on the query.
func (l *Lookup) Query(metrics bool, logFrom logutil.LogFrom) (bool, error) {
	isNewVersion, err := l.query(logFrom)

	if metrics {
		l.QueryMetrics(l, err)
	}

	return isNewVersion, err
}

// Query queries the source
// and returns whether a new release was found, updating LatestVersion if so.
func (l *Lookup) query(logFrom logutil.LogFrom) (bool, error) {
	body, err := l.httpRequest(logFrom)
	if err != nil {
		return false, err
	}

	version, err := l.getVersion(string(body), logFrom)
	if err != nil {
		logutil.Log.Error(err, logFrom, true)
		return false, err
	}

	l.Status.SetLastQueried("")

	// If this version differs (new?).
	if previousVersion := l.Status.LatestVersion(); version != previousVersion {
		// Verify Semantic Versioning (if enabled).
		if l.Options.GetSemanticVersioning() {
			if err := l.VerifySemanticVersioning(version, previousVersion, logFrom); err != nil {
				return false, err //nolint:wrapcheck
			}
		}

		return l.HandleNewVersion(version, "", logFrom) //nolint:wrapcheck
	}

	// Announce `LastQueried`.
	l.Status.AnnounceQuery()
	// No version change.
	return false, nil
}

// httpRequest makes a HTTP GET request to the URL and returns the body.
func (l *Lookup) httpRequest(logFrom logutil.LogFrom) ([]byte, error) {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	// HTTPS insecure skip verify.
	if l.allowInvalidCerts() {
		//#nosec G402 -- explicitly wanted InsecureSkipVerify
		customTransport.TLSClientConfig.InsecureSkipVerify = true
	}

	// Create the request.
	req, err := http.NewRequest(http.MethodGet, l.URL, nil)
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

	// Read the response body.
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 50<<20)) // Limit to 50 MiB.
	logutil.Log.Error(err, logFrom, err != nil)
	return body, err //nolint:wrapcheck
}

// getVersion returns the latest version from `body` that matches the URLCommands, and Regex requirements.
func (l *Lookup) getVersion(body string, logFrom logutil.LogFrom) (string, error) {
	filteredVersions, err := l.URLCommands.GetVersions(body, logFrom)
	if err != nil {
		return "", fmt.Errorf("no releases were found matching the url_commands\n%w", err)
	}
	if len(filteredVersions) == 0 {
		return "", errors.New("no releases were found matching the url_commands")
	}

	// Sort versions if Semantic Versioning enabled.
	if l.Options.GetSemanticVersioning() {
		sort.Slice(filteredVersions, func(i, j int) bool {
			vI, errI := semver.NewVersion(filteredVersions[i])
			vJ, errJ := semver.NewVersion(filteredVersions[j])

			if errI != nil || errJ != nil {
				return false
			}
			return vI.GreaterThan(vJ)
		})
	}

	// Check all releases for the one meeting the requirements.
	var firstErr error
	for _, version := range filteredVersions {
		if err := l.versionMeetsRequirements(version, body, logFrom); err == nil {
			return version, nil
		} else if firstErr == nil {
			firstErr = err
		}
	}

	return "", fmt.Errorf("no releases were found matching the require fields\n%w", firstErr)
}

// versionMeetsRequirements checks whether `version` meets the requirements of the Lookup.
func (l *Lookup) versionMeetsRequirements(version, body string, logFrom logutil.LogFrom) error {
	// No `Require` filters.
	if l.Require == nil {
		return nil
	}

	// Check all `Require` filters for this version.
	// Version RegEx.
	if err := l.Require.RegexCheckVersion(version, logFrom); err != nil {
		return err //nolint:wrapcheck
	}

	// Content RegEx (on response body).
	if err := l.Require.RegexCheckContent(version, body, logFrom); err != nil {
		return err //nolint:wrapcheck
	}

	// If the Command didn't return successfully.
	if err := l.Require.ExecCommand(version, logFrom); err != nil {
		return err //nolint:wrapcheck
	}

	// If the Docker tag doesn't exist.
	if err := l.Require.DockerTagCheck(version); err != nil {
		errStr := err.Error()
		if strings.HasSuffix(errStr, "\n") {
			err = errors.New(strings.TrimSuffix(errStr, "\n"))
		}
		logutil.Log.Warn(err, logFrom, true)
		return err
		// Docker image:tag does exist.
	} else if l.Require.Docker != nil {
		logutil.Log.Info(
			fmt.Sprintf(`found %s container "%s:%s"`,
				l.Require.Docker.GetType(), l.Require.Docker.Image, l.Require.Docker.GetTag(version)),
			logFrom, true)
	}

	return nil
}
