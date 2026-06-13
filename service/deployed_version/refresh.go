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

// Package deployedver provides the deployed_version lookup service to for a service.
package deployedver

import (
	"errors"
	"fmt"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/service/deployed_version/types/base"
	dvmanual "github.com/release-argus/Argus/service/deployed_version/types/manual"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/util"
)

// Refresh the Lookup with the provided overrides.
//
//	Returns: version, decode.
func Refresh(
	lookup Lookup,
	previousType string,
	overrides []byte,
	semanticVersioning *string, // nil, "true", "false", "null" (unchanged, true, false, default).
	secretRefs *shared.VSecretRef,
) (string, error) {
	if lookup == nil {
		return "", errors.New("lookup is nil")
	}

	logFrom := logx.LogFrom{Primary: "deployed_version/refresh", Secondary: lookup.GetServiceID()}

	// Whether this new `semantic_version` resolves differently than the current one.
	semanticVerDiff := semanticVersioning != nil && (
	// semantic_versioning explicitly null, and the default resolves to a different value.
	(*semanticVersioning == "null" && lookup.GetOptions().GetSemanticVersioning() != *util.FirstNonNilPtr(
		lookup.GetDefaults().Options.SemanticVersioning,
		lookup.GetHardDefaults().Options.SemanticVersioning)) ||
		// semantic_versioning now resolves to a different value than the default.
		(*semanticVersioning != "null" && *semanticVersioning == "true" != lookup.GetOptions().GetSemanticVersioning()))
	// Create a new lookup if overrides provided, or `semantic_versioning` changed.
	usingOverrides := overrides != nil || semanticVerDiff

	// Create a new Lookup if overrides provided.
	newLookup, err := applyOverridesJSON(
		lookup,
		overrides,
		semanticVerDiff,
		semanticVersioning,
	)
	if err != nil {
		return "", err
	}
	if overrides != nil {
		newLookup.InheritSecrets(lookup, secretRefs)
		// Remove channels from Status when using overrides (if not manual).
		if newLookup.GetType() != dvmanual.Type && previousType != dvmanual.Type {
			newLookup.SetStatus(lookup.GetStatus().Copy(false))
		}
		// Check values as they may have changed.
		if err := newLookup.CheckValues(); err != nil {
			return "", err //nolint:wrapcheck
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

	// Query the lookup.
	if err := newLookup.Query(!usingOverrides, logFrom); err != nil {
		return "", err //nolint:wrapcheck
	}

	return newLookup.GetStatus().DeployedVersion(), nil
}

// applyOverridesJSON applies the JSON overrides to the Lookup.
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
