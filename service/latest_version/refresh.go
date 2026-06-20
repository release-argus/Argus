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

// Package latestver provides the latest_version lookup service to for a service.
package latestver

import (
	"errors"
	"fmt"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/util"
)

// Refresh queries the lookup with the provided overrides, returning the version, whether to announce an update, and any error.
func Refresh(
	lookup Lookup,
	overrides []byte,
	semanticVersioning *string, // nil, "true", "false", "null" (unchanged, true, false, default).
	secretRefs *shared.VSecretRef,
) (string, bool, error) {
	if lookup == nil {
		return "", false, errors.New("lookup is nil")
	}

	logFrom := logx.LogFrom{Primary: "latest_version/refresh", Secondary: lookup.GetServiceID()}

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

	// Create a new Lookup if overrides provided.
	newLookup, err := applyOverridesJSON(
		lookup,
		overrides,
		semanticVerDiff,
		semanticVersioning,
	)
	if err != nil {
		return "", false, err
	}
	if overrides != nil {
		newLookup.InheritSecrets(lookup, secretRefs)
		// Remove channels from Status when using overrides.
		newLookup.SetStatus(lookup.GetStatus().Copy(false))
		// Check values as they may have changed.
		if err := newLookup.CheckValues(); err != nil {
			return "", false, err //nolint:wrapcheck
		}
	}

	// Log the lookup in use.
	if logx.IsLevel("DEBUG") {
		logx.Debug(
			fmt.Sprintf("Refreshing with:\n%q", lookup.String("")),
			logFrom,
			true,
		)
	}

	hadVersion := lookup.GetStatus().LatestVersion()
	// Query the lookup.
	if _, err := newLookup.Query(!usingOverrides, logFrom); err != nil {
		return "", false, err //nolint:wrapcheck
	}

	version := newLookup.GetStatus().LatestVersion()
	// Inherit any updated 'Require' tokens.
	lookup.InheritSecrets(newLookup, secretRefs)

	// Announce the update? (if not using overrides, and the version changed).
	announceUpdate := !usingOverrides && version != hadVersion

	return version, announceUpdate, nil
}

// applyOverridesJSON applies the JSON overrides to the lookup.
func applyOverridesJSON(
	lookup Lookup,
	overrides []byte,
	semanticVerDiff bool,
	semanticVersioning *string, // nil, "true", "false", "null" (unchanged, true, false, default).
) (Lookup, error) {
	newLookup, err := ApplyOverrides(
		"yaml", overrides,
		lookup,
		lookup.GetOptions(),
		lookup.GetStatus(),
		base.DefaultsConfig{
			Soft: lookup.GetDefaults(),
			Hard: lookup.GetHardDefaults(),
		},
	)
	if err != nil {
		return nil, err
	}

	// Apply the new semantic_versioning JSON value.
	if semanticVerDiff {
		semanticVersioningRoot := util.ClonePtr(lookup.GetOptions().SemanticVersioning)
		// Apply the new semantic_versioning JSON value.
		if err := decode.Unmarshal("json", []byte(*semanticVersioning), &semanticVersioningRoot); err != nil {
			return nil, &decode.KeyFieldError{
				Key: "semantic_versioning",
				Err: err,
			}
		}
		newLookup.GetOptions().SemanticVersioning = semanticVersioningRoot
	}

	return newLookup, nil
}
