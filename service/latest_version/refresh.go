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

// Package latestver provides the latest_version lookup service to for a service.
package latestver

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

// Refresh the Lookup with the provided overrides.
//
//	Returns: version, announceUpdate, err.
func Refresh(
	lookup Lookup,
	overrides *string,
	semanticVersioning *string, // nil, "true", "false", "null" (unchanged, true, false, default).
) (string, bool, error) {
	if lookup == nil {
		return "", false, errors.New("lookup is nil")
	}

	logFrom := logutil.LogFrom{Primary: "latest_version/refresh", Secondary: lookup.GetServiceID()}

	// Whether this new semantic_version resolves differently than the current one.
	semanticVerDiff := semanticVersioning != nil && (
	// semantic_versioning explicitly null, and the default resolves to a different value.
	(*semanticVersioning == "null" && lookup.GetOptions().GetSemanticVersioning() != *util.FirstNonNilPtr(
		lookup.GetDefaults().Options.SemanticVersioning,
		lookup.GetHardDefaults().Options.SemanticVersioning)) ||
		// semantic_versioning now resolves to a different value than the default.
		(*semanticVersioning != "null" && *semanticVersioning == "true" != lookup.GetOptions().GetSemanticVersioning()))
	// Whether we need to create a new Lookup.
	usingOverrides := overrides != nil || semanticVerDiff

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
			return "", false, err
		}
	}

	// Log the lookup in use.
	if logutil.Log.IsLevel("DEBUG") {
		logutil.Log.Debug(
			fmt.Sprintf("Refreshing with:\n%q",
				lookup.String(lookup, "")),
			logFrom, true)
	}

	hadVersion := lookup.GetStatus().LatestVersion()
	// Query the lookup.
	_, err := newLookup.Query(!usingOverrides, logFrom)
	if err != nil {
		return "", false, err //nolint:wrapcheck
	}

	version := newLookup.GetStatus().LatestVersion()
	lookup.Inherit(newLookup)

	// Announce the update? (if not using overrides, and the version changed).
	announceUpdate := !usingOverrides && version != hadVersion

	return version, announceUpdate, nil
}

// LookupTypeExtractor is a struct that extracts the type from a JSON string.
type LookupTypeExtractor struct {
	Type string `json:"type"`
}

// applyOverridesJSON applies the JSON overrides to the Lookup.
func applyOverridesJSON(
	lookup Lookup,
	overrides *string,
	semanticVerDiff bool,
	semanticVersioning *string, // nil, "true", "false", "null" (unchanged, true, false, default).
) (Lookup, error) {
	var newLookup Lookup
	var extractedOverrides *LookupTypeExtractor
	if overrides != nil {
		if err := json.Unmarshal([]byte(*overrides), &extractedOverrides); err != nil {
			errStr := util.FormatUnmarshalError("json", err)
			errStr = strings.ReplaceAll(errStr, "\n", "\n  ")
			return nil, errors.New("failed to unmarshal latestver.Lookup:\n  " + errStr)
		}
	}
	// Copy the existing Lookup.
	if overrides == nil || (extractedOverrides.Type == "" || extractedOverrides.Type == lookup.GetType()) {
		newLookup = Copy(lookup)
	} else {
		// Convert to the new type.
		var err error
		if newLookup, err = ChangeType(extractedOverrides.Type, lookup, ""); err != nil {
			return nil, err
		}
		newRequire := newLookup.GetRequire()
		if newRequire != nil {
			newRequire.Inherit(lookup.GetRequire())
		}
	}

	// Apply the new semantic_versioning JSON value.
	if semanticVerDiff {
		semanticVersioningRoot := util.CopyPointer(lookup.GetOptions().SemanticVersioning)
		// Apply the new semantic_versioning JSON value.
		if err := json.Unmarshal([]byte(*semanticVersioning), &semanticVersioningRoot); err != nil {
			errStr := util.FormatUnmarshalError("json", err)
			errStr = strings.ReplaceAll(errStr, "\n", "\n  ")
			return nil, errors.New("failed to unmarshal latestver.Lookup.Options.SemanticVersioning:\n  " + errStr)
		}
		newLookup.GetOptions().SemanticVersioning = semanticVersioningRoot
	}

	// Apply the overrides.
	if overrides != nil {
		if err := json.Unmarshal([]byte(*overrides), &newLookup); err != nil {
			errStr := util.FormatUnmarshalError("json", err)
			errStr = strings.ReplaceAll(errStr, "\n", "\n  ")
			return nil, errors.New("failed to unmarshal latestver.Lookup:\n  " + errStr)
		}
		if strings.Contains(*overrides, `"docker":`) {
			require := newLookup.GetRequire()
			if require.Docker != nil {
				require.Docker.ClearQueryToken()
				require.Inherit(lookup.GetRequire())
			}
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
