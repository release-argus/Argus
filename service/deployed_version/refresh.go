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
	"strings"

	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

// Refresh the Lookup with the provided overrides.
//
//	Returns: version, err.
func Refresh(
	lookup Lookup,
	previousType string,
	overrides string,
	semanticVersioning *string, // nil, "true", "false", "null" (unchanged, true, false, default).
) (string, error) {
	if lookup == nil {
		return "", errors.New("lookup is nil")
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
	usingOverrides := (overrides != "" && previousType != "manual") || semanticVerDiff

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
	} else if previousType == "manual" && lookup.GetType() == "manual" &&
		overrides != "" {
		if err := json.Unmarshal([]byte(overrides), &lookup); err != nil {
			errStr := util.FormatUnmarshalError("json", err)
			errStr = strings.ReplaceAll(errStr, "\n", "\n  ")
			return "", errors.New("failed to unmarshal deployedver.Lookup:\n  " + errStr)
		}
	}

	// Log the lookup in use.
	if logutil.Log.IsLevel("DEBUG") {
		logutil.Log.Debug(
			fmt.Sprintf("Refreshing with:\n%q",
				lookup.String(lookup, "")),
			logFrom, true)
	}

	// Query the lookup.
	err := newLookup.Query(!usingOverrides, logFrom)
	if err != nil {
		return "", err //nolint: wrapcheck
	}

	return newLookup.GetStatus().DeployedVersion(), nil
}

// LookupTypeExtractor is a struct that extracts the type from a JSON string.
type LookupTypeExtractor struct {
	Type string `json:"type"`
}

// applyOverridesJSON applies the JSON overrides to the Lookup.
func applyOverridesJSON(
	lookup Lookup,
	overrides string,
	semanticVerDiff bool,
	semanticVersioning *string, // nil, "true", "false", "null" (unchanged, true, false, default).
) (Lookup, error) {
	var newLookup Lookup
	var extractedOverrides *LookupTypeExtractor
	if overrides != "" {
		if err := json.Unmarshal([]byte(overrides), &extractedOverrides); err != nil {
			errStr := util.FormatUnmarshalError("json", err)
			errStr = strings.ReplaceAll(errStr, "\n", "\n  ")
			return nil, errors.New("failed to unmarshal deployedver.Lookup:\n  " + errStr)
		}
	}
	// Copy the existing Lookup.
	if overrides == "" || (extractedOverrides.Type == "" || extractedOverrides.Type == lookup.GetType()) {
		newLookup = Copy(lookup)
	} else {
		// Convert to the new type.
		var err error
		newLookup, err = New(
			extractedOverrides.Type,
			"yaml", "",
			lookup.GetOptions(),
			lookup.GetStatus(),
			lookup.GetDefaults(), lookup.GetHardDefaults())
		if err != nil {
			return nil, err
		}
	}

	// Apply the new semantic_versioning JSON value.
	if semanticVerDiff {
		semanticVersioningRoot := util.CopyPointer(lookup.GetOptions().SemanticVersioning)
		// Apply the new semantic_versioning JSON value.
		if err := json.Unmarshal([]byte(*semanticVersioning), &semanticVersioningRoot); err != nil {
			errStr := util.FormatUnmarshalError("json", err)
			errStr = strings.ReplaceAll(errStr, "\n", "\n  ")
			return nil, errors.New("failed to unmarshal deployedver.Lookup.Options.SemanticVersioning:\n  " + errStr)
		}
		newLookup.GetOptions().SemanticVersioning = semanticVersioningRoot
	}

	// Apply the overrides.
	if overrides != "" {
		if err := json.Unmarshal([]byte(overrides), &newLookup); err != nil {
			errStr := util.FormatUnmarshalError("json", err)
			errStr = strings.ReplaceAll(errStr, "\n", "\n  ")
			return nil, errors.New("failed to unmarshal deployedver.Lookup:\n  " + errStr)
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
