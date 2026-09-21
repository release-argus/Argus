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
	// Scheme defaults to HTTPS, and the port defaults by scheme.
	address := util.EvalEnvVars(l.Host)
	if !strings.Contains(address, "://") {
		address = "https://" + address
	}

	parsed, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf(
			"invalid host %q: %w",
			l.Host, err,
		)
	}

	return &url.URL{
		Scheme: parsed.Scheme,
		Host:   parsed.Host,
		Path:   parsed.Path,
	}, nil
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
