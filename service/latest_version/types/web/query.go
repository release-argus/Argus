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
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/release-argus/Argus/internal/httpx"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/util"
)

// Query fetches the URL, sets Prometheus metrics if requested, and returns whether a new version was found.
func (l *Lookup) Query(metrics bool, logFrom logx.LogFrom) (bool, error) {
	isNewVersion, err := l.query(logFrom)

	if metrics {
		l.QueryMetrics(l, err)
	}

	return isNewVersion, err
}

// query fetches the URL and returns whether a new version was found, updating LatestVersion if so.
func (l *Lookup) query(logFrom logx.LogFrom) (bool, error) {
	body, err := l.httpRequest(logFrom)
	if err != nil {
		return false, err
	}

	version, err := l.getVersion(string(body), logFrom)
	if err != nil {
		return false, err
	}

	l.Status.SetLastQueried("")

	// If this version differs (new?).
	if previousVersion := l.Status.LatestVersion(); version != previousVersion {
		return l.HandleNewVersion(version, "", logFrom) //nolint:wrapcheck
	}

	// Announce `LastQueried`.
	l.Status.AnnounceQuery()
	// No version change.
	return false, nil
}

// httpRequest makes a HTTP GET request to the URL and returns the body.
func (l *Lookup) httpRequest(logFrom logx.LogFrom) ([]byte, error) {
	client := httpx.Client
	// HTTPS insecure skip verify.
	if l.allowInvalidCerts() {
		client = httpx.InsecureClient
	}

	// Create the request.
	req, err := http.NewRequest(http.MethodGet, l.URL, nil)
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

	// Read the response body.
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 50<<20)) // Limit to 50 MiB.
	logx.Error(err, logFrom, err != nil)
	return body, err //nolint:wrapcheck
}

// getVersion returns the latest version from `body` that matches the URLCommands, and Regex requirements.
func (l *Lookup) getVersion(body string, logFrom logx.LogFrom) (string, error) {
	filteredVersions, err := l.URLCommands.GetVersions(body, logFrom)
	if err != nil {
		err := fmt.Errorf("no releases were found matching the url_commands %w", err)
		logx.Error(err, logFrom, true)
		return "", err
	}
	if len(filteredVersions) == 0 {
		err := errors.New("no releases were found matching the url_commands")
		logx.Error(err, logFrom, true)
		return "", err
	}

	// Sort versions if Semantic Versioning enabled.
	if l.Options.GetSemanticVersioning() {
		// Sorting won't catch non-semver if there's only one version.
		if len(filteredVersions) == 1 {
			if _, err := l.Options.VerifySemanticVersioning(filteredVersions[0], logFrom); err != nil {
				logx.Warn(err, logFrom, true)
				return "", err //nolint:wrapcheck
			}
		} else {
			sort.Slice(filteredVersions, func(i, j int) bool {
				return versionSortsBefore(filteredVersions[i], filteredVersions[j])
			})
		}
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

	err = fmt.Errorf("no releases were found matching the require fields %w", firstErr)
	logx.Error(err, logFrom, true)
	return "", err
}

// versionSortsBefore reports whether version `a` should sort ahead of version `b` when ordering
// candidates from newest to oldest.
//
// A version that parses as semver always sorts ahead of one that does not; when neither parses,
// a fixed lexical order is used so the result never depends on the order
// candidates were found in.
func versionSortsBefore(a, b string) bool {
	semverA, errA := semver.NewVersion(a)
	semverB, errB := semver.NewVersion(b)

	switch {
	case errA == nil && errB == nil:
		return semverA.GreaterThan(semverB)
	case errA == nil: // a parses, b does not - a sorts first.
		return true
	case errB == nil: // b parses, a does not - b sorts first.
		return false
	default: // Neither parses - fall back to lexical order.
		return strings.Compare(a, b) < 0
	}
}

// versionMeetsRequirements returns an error if version does not satisfy all Require filters.
func (l *Lookup) versionMeetsRequirements(version, body string, logFrom logx.LogFrom) error {
	// No `Require` filters.
	if l.Require == nil {
		return nil
	}

	// Check all `Require` filters for this version.
	// ---

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
		logx.Warn(err, logFrom, true)
		return err //nolint:wrapcheck
		// Docker image:tag does exist.
	} else if l.Require.Docker != nil {
		logx.Info(
			fmt.Sprintf(
				`found %s container "%s:%s"`,
				l.Require.Docker.GetType(), l.Require.Docker.GetImage(), l.Require.Docker.GetTagForVersion(version),
			),
			logFrom,
			true,
		)
	}

	return nil
}
