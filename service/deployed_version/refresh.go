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
	"encoding/json"
	"fmt"

	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

// applyOverrides to the Lookup and return that new Lookup.
func (l *Lookup) applyOverrides(
	allowInvalidCerts *string,
	basicAuth *string,
	headers *string,
	json *string,
	regex *string,
	semanticVersioning *string,
	url *string,
	serviceID *string,
	logFrom *util.LogFrom,
) (*Lookup, error) {
	// Use the provided overrides, or the defaults.
	// allow_invalid_certs
	useAllowInvalidCerts := l.AllowInvalidCerts
	if allowInvalidCerts != nil {
		useAllowInvalidCerts = util.StringToBoolPtr(*allowInvalidCerts)
	}
	// basic_auth
	useBasicAuth := basicAuthFromString(
		basicAuth,
		l.BasicAuth,
		logFrom)
	// headers
	useHeaders := headersFromString(
		headers,
		&l.Headers,
		logFrom)
	// json
	useJSON := util.GetValue(json, l.JSON)
	// regex
	useRegex := util.GetValue(regex, l.Regex)
	// semantic_versioning
	useSemanticVersioning := l.Options.SemanticVersioning
	if semanticVersioning != nil {
		useSemanticVersioning = util.StringToBoolPtr(*semanticVersioning)
	}
	// url
	useURL := util.GetValue(url, l.URL)

	// Create a new lookup with the overrides.
	lookup := Lookup{
		URL:               useURL,
		AllowInvalidCerts: useAllowInvalidCerts,
		BasicAuth:         useBasicAuth,
		Headers:           *useHeaders,
		JSON:              useJSON,
		Regex:             useRegex,
		Options: &opt.Options{
			SemanticVersioning: useSemanticVersioning,
			Defaults:           l.Options.Defaults,
			HardDefaults:       l.Options.HardDefaults,
		},
		Status:       &svcstatus.Status{},
		Defaults:     l.Defaults,
		HardDefaults: l.HardDefaults,
	}
	if err := lookup.CheckValues(""); err != nil {
		jLog.Error(err, *logFrom, true)
		return nil, fmt.Errorf("values failed validity check:\n%w", err)
	}
	lookup.Status.Init(
		0, 0, 0,
		serviceID,
		nil,
	)
	return &lookup, nil
}

// Refresh (query) the Lookup with the provided overrides,
// returning the version found with this query
func (l *Lookup) Refresh(
	allowInvalidCerts *string,
	basicAuth *string,
	headers *string,
	json *string,
	regex *string,
	semanticVersioning *string,
	url *string,
) (version string, announceUpdate bool, err error) {
	serviceID := *l.Status.ServiceID
	logFrom := util.LogFrom{Primary: "deployed_version/refresh", Secondary: serviceID}

	var lookup *Lookup
	lookup, err = l.applyOverrides(
		allowInvalidCerts,
		basicAuth,
		headers,
		json,
		regex,
		semanticVersioning,
		url,
		&serviceID,
		&logFrom)
	if err != nil {
		return
	}

	// Log the lookup being used if debug.
	if jLog.IsLevel("DEBUG") {
		jLog.Debug(
			fmt.Sprintf("Refreshing with:\n%v", lookup),
			logFrom, true)
	}

	// Whether overrides were provided or not, we can update the status ig not.
	overrides := headers != nil ||
		json != nil ||
		regex != nil ||
		semanticVersioning != nil ||
		url != nil

	// Query the lookup.
	version, err = lookup.Query(!overrides, &logFrom)
	if err != nil {
		return
	}

	// Update the deployed version if it has changed.
	if version != l.Status.GetDeployedVersion() &&
		// and no overrides that may change a successful query were provided
		!overrides {
		announceUpdate = true
		l.Status.SetDeployedVersion(version, true)
		l.Status.AnnounceUpdate()
	}

	return
}

func basicAuthFromString(jsonStr *string, previous *BasicAuth, logFrom *util.LogFrom) *BasicAuth {
	// jsonStr == nil when it hasn't been changed, so return the previous
	if jsonStr == nil {
		return previous
	}

	basicAuth := &BasicAuth{}
	err := json.Unmarshal([]byte(*jsonStr), &basicAuth)
	// Ignore the JSON if it failed to unmarshal
	if err != nil {
		jLog.Error(fmt.Sprintf("Failed converting JSON - %q\n%s", *jsonStr, util.ErrorToString(err)),
			*logFrom, true)
		return previous
	}
	keys := util.GetKeysFromJSON(*jsonStr)

	// Had no previous, so can't use it as defaults
	if previous == nil {
		return basicAuth
	}

	// defaults
	if !util.Contains(keys, "username") {
		basicAuth.Username = previous.Username
	}
	if !util.Contains(keys, "password") {
		basicAuth.Password = previous.Password
	}

	return basicAuth
}

func headersFromString(jsonStr *string, previous *[]Header, logFrom *util.LogFrom) *[]Header {
	// jsonStr == nil when it hasn't been changed, so return the previous
	if jsonStr == nil {
		return previous
	}

	var headers []Header
	err := json.Unmarshal([]byte(*jsonStr), &headers)
	// Ignore the JSON if it failed to unmarshal
	if err != nil {
		jLog.Error(fmt.Sprintf("Failed converting JSON - %q\n%s", *jsonStr, util.ErrorToString(err)),
			*logFrom, true)
		return previous
	}

	return &headers
}
