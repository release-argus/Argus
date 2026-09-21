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

// Package forgejo provides a Forgejo-based lookup type.
package forgejo

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/release-argus/Argus/util"
)

// apiPageSize is the page size requested from the forge, at its MAX_RESPONSE_ITEMS default.
const apiPageSize = 50

// parseHost resolves the Host of the receiver to the root URL of the instance.
func (l *Lookup) parseHost() (*url.URL, error) {
	parsed, err := parseInstanceURL(util.EvalEnvVars(l.resolveHost()))
	if err != nil {
		return nil, fmt.Errorf(
			"invalid host %q: %w",
			l.Host, err,
		)
	}

	return parsed, nil
}

// parseInstanceURL parses an instance address to the root URL of the instance.
func parseInstanceURL(address string) (*url.URL, error) {
	if !strings.Contains(address, "://") {
		address = "https://" + address
	}

	parsed, err := url.Parse(address)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	return &url.URL{
		Scheme: parsed.Scheme,
		Host:   parsed.Host,
		Path:   strings.TrimRight(parsed.Path, "/"),
	}, nil
}

// urlProblem returns what to tell the user about an instance URL, or an empty
// string when it is usable.
func urlProblem(address string) string {
	parsed, err := parseInstanceURL(address)
	if err != nil {
		return "not a valid URL"
	}

	switch {
	case parsed.Scheme != "http" && parsed.Scheme != "https":
		return "scheme must be http or https"
	case parsed.Hostname() == "":
		return "no hostname"
	}

	return ""
}

// canonicalHost returns the comparison form of a host spelling, so that every
// spelling of one instance resolves to the same key.
func canonicalHost(host string) string {
	address := strings.TrimSpace(util.EvalEnvVars(host))
	if address == "" {
		return ""
	}

	if !strings.Contains(address, "://") {
		address = "https://" + address
	}

	parsed, err := url.Parse(address)
	if err != nil {
		return strings.ToLower(address)
	}

	scheme := strings.ToLower(parsed.Scheme)
	hostname := strings.ToLower(parsed.Host)
	if port := parsed.Port(); port != "" && isDefaultPort(scheme, port) {
		hostname = strings.TrimSuffix(hostname, ":"+port)
	}

	return fmt.Sprintf(
		"%s://%s%s",
		scheme, hostname, strings.TrimRight(parsed.Path, "/"),
	)
}

// isDefaultPort reports whether port is the one scheme implies.
func isDefaultPort(scheme, port string) bool {
	switch port {
	case "443":
		return scheme == "https"
	case "80":
		return scheme == "http"
	}
	return false
}

// apiURL returns the instance's API URL for `endpoint`, requesting `page`.
func (l *Lookup) apiURL(endpoint string, page int) (string, error) {
	host, err := l.parseHost()
	if err != nil {
		return "", err
	}

	query := url.Values{}
	query.Set("limit", strconv.Itoa(apiPageSize))
	if page > 1 {
		query.Set("page", strconv.Itoa(page))
	}

	return (&url.URL{
		Scheme:   host.Scheme,
		Host:     host.Host,
		Path:     fmt.Sprintf("%s/api/v1/repos/%s/%s", host.Path, l.URL, endpoint),
		RawQuery: query.Encode(),
	}).String(), nil
}
