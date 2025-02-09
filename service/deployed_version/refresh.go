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

// Package deployedver provides the deployed_version lookup service to for a service.
package deployedver

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

// Refresh the Lookup with the provided overrides.
//
//	Returns: version, err.
func Refresh(
	lookup Lookup,
	overrides string,
	semanticVersioning *string, // nil, "true", "false", "null" (unchanged, true, false, default).
) (string, error) {
	if lookup == nil {
		return "", errors.New("lookup is nil")
	}

	logFrom := logutil.LogFrom{Primary: "latest_version/refresh", Secondary: *lookup.GetStatus().ServiceID}

	// Whether this new semantic_version resolves differently than the current one.
	semanticVerDiff := semanticVersioning != nil && (
	// semantic_versioning explicitly null, and the default resolves to a different value.
	(*semanticVersioning == "null" && lookup.GetOptions().GetSemanticVersioning() != *util.FirstNonNilPtr(
		lookup.GetDefaults().Options.SemanticVersioning,
		lookup.GetHardDefaults().Options.SemanticVersioning)) ||
		// semantic_versioning now resolves to a different value than the default.
		(*semanticVersioning != "null" && *semanticVersioning == "true" != lookup.GetOptions().GetSemanticVersioning()))
	// Whether we need to create a new Lookup.
	usingOverrides := overrides != "" || semanticVerDiff

	newLookup := lookup
	// Create a new Lookup if overrides provided.
	if usingOverrides {
		var err error
		newLookup, err = applyOverridesJSON(
			lookup,
			overrides,
			semanticVerDiff,
			semanticVersioning)
		if err != nil {
			return "", err
		}
	}

	// Log the lookup in use.
	if logutil.Log.IsLevel("DEBUG") {
		logutil.Log.Debug(
			fmt.Sprintf("Refreshing with:\n%q", lookup.String(lookup, "")),
			logFrom, true)
	}

	// Query the lookup.
	err := newLookup.Query(!usingOverrides, logFrom)
	if err != nil {
		return "", err //nolint: wrapcheck
	}

	return newLookup.GetStatus().DeployedVersion(), nil
}

// applyOverridesJSON applies the JSON overrides to the Lookup.
func applyOverridesJSON(
	lookup Lookup,
	overrides string,
	semanticVerDiff bool,
	semanticVersioning *string, // nil, "true", "false", "null" (unchanged, true, false, default).
) (Lookup, error) {
	// Copy the existing Lookup.
	newLookup := Copy(lookup)

	// Apply the new semantic_versioning JSON value.
	if semanticVerDiff {
		semanticVersioningRoot := util.CopyPointer(lookup.GetOptions().SemanticVersioning)
		// Apply the new semantic_versioning JSON value.
		if err := json.Unmarshal([]byte(*semanticVersioning), &semanticVersioningRoot); err != nil {
			return nil, fmt.Errorf("failed to unmarshal deployedver.Lookup.SemanticVersioning: %w", err)
		}
		newLookup.GetOptions().SemanticVersioning = semanticVersioningRoot
	}

	// Apply the overrides.
	if overrides != "" {
		if err := json.Unmarshal([]byte(overrides), &newLookup); err != nil {
			return nil, fmt.Errorf("failed to unmarshal deployedver.Lookup: %w", err)
		}
		newLookup.Init(
			newLookup.GetOptions(),
			newLookup.GetStatus(),
			newLookup.GetDefaults(), newLookup.GetHardDefaults())
	}

	// Check the overrides.
	err := newLookup.CheckValues("")
	return newLookup, err //nolint:wrapcheck
}
