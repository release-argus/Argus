// Copyright [2024] [Argus]
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

// Package deployedver provides the deployed_version lookup.
package deployedver

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/release-argus/Argus/util"
)

// Refresh creates a new Lookup instance (if overrides are provided), and queries the Lookup for the deployed version
// and returns that version.
func (l *Lookup) Refresh(
	serviceID *string,
	overrides *string,
	semanticVersioning *string,
) (string, error) {
	if l == nil {
		return "", fmt.Errorf("lookup is nil")
	}
	logFrom := util.LogFrom{Primary: "deployed_version/refresh", Secondary: *serviceID}

	// Whether this new semantic_version resolves differently than the current one.
	semanticVerDiff := semanticVersioning != nil && (
	// semantic_versioning is explicitly null, and the default resolves to a different value.
	(*semanticVersioning == "null" && l.Options.GetSemanticVersioning() == *util.FirstNonNilPtr(
		l.Defaults.Options.SemanticVersioning, l.HardDefaults.Options.SemanticVersioning)) ||
		// semantic_versioning now resolves to a different value than the default.
		(*semanticVersioning == "true") != l.Options.GetSemanticVersioning())
	// Whether we need to create a new Lookup.
	usingOverrides := overrides != nil || semanticVerDiff

	lookup := l
	// Create a new lookup if we don't have one, or overrides were provided.
	if usingOverrides {
		var err error
		lookup, err = applyOverridesJSON(
			l,
			overrides,
			semanticVerDiff,
			semanticVersioning)
		if err != nil {
			return "", err
		}
	}

	// Log the lookup in use.
	if jLog.IsLevel("DEBUG") {
		jLog.Debug(
			fmt.Sprintf("Refreshing with:\n%q", lookup.String("")),
			logFrom, true)
	}

	// Query the lookup.
	version, err := lookup.Query(!usingOverrides, logFrom)
	if err != nil {
		return "", err
	}

	// Update the deployed version if it has changed.
	if version != l.Status.DeployedVersion() &&
		// and no overrides that may change a successful query were provided.
		!usingOverrides {
		l.HandleNewVersion(version, true)
	}

	return version, nil
}

// applyOverridesJSON applies JSON-based overrides and semantic versioning changes to a copy of the Lookup object
// and returns that copy.
//
// Note: The semanticVersioning parameter can be (nil, "true", "false", or "null")
// to indicate (unchanged, true, false, or default) values respectively.
func applyOverridesJSON(
	lookup *Lookup,
	overrides *string,
	semanticVerDiff bool,
	semanticVersioning *string, // nil, "true", "false", "null" (unchanged, "true", "false", default).
) (*Lookup, error) {
	// Copy the existing lookup.
	lookup = Copy(lookup)

	// Apply the new semantic_versioning json value.
	if semanticVerDiff {
		var newSemanticVersioning *bool
		// Apply the new semantic_versioning json value.
		if err := json.Unmarshal([]byte(*semanticVersioning), &newSemanticVersioning); err != nil {
			return nil, fmt.Errorf("failed to unmarshal semantic_versioning: %w", err)
		}
		lookup.Options.SemanticVersioning = newSemanticVersioning
	}

	// Apply the overrides.
	if overrides != nil {
		if err := json.Unmarshal([]byte(*overrides), &lookup); err != nil {
			return nil, fmt.Errorf("failed to unmarshal deployed_version: %w", err)
		}
	}

	// Check the overrides.
	errs := lookup.CheckValues("")
	if errs != nil {
		return nil, errors.Join(errs)
	}
	return lookup, nil
}
