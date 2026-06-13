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

//go:build unit

// Package base provides the base struct for deployed_version lookups.
package base

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/service/status"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
)

// ########
// # INIT #
// ########

func TestLookup_Init(t *testing.T) {
	// GIVEN: a Lookup and its dependencies.
	options := &opt.Options{}
	svcStatus := &status.Status{}
	dvCfg := plainDefaultsConfig(t)
	l := &Lookup{}

	// WHEN: Init is called.
	l.Init(options, svcStatus, dvCfg)

	prefix := fmt.Sprintf(
		"%s\nLookup.Init(options=%v, status=%v, defaults=%v)",
		packageName, options, &svcStatus, dvCfg,
	)

	// THEN: pointers to those vars are handed out to the Lookup.
	fieldTests := []test.FieldAssertion{
		{Name: "Options", Got: l.Options, Want: options, Mode: test.CompareSamePointer},
		{Name: "Status", Got: l.Status, Want: svcStatus, Mode: test.CompareSamePointer},
		{Name: "Defaults", Got: l.Defaults, Want: dvCfg.Soft, Mode: test.CompareSamePointer},
		{Name: "HardDefaults", Got: l.HardDefaults, Want: dvCfg.Hard, Mode: test.CompareSamePointer},
	}
	if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
		t.Fatal(err)
	}
}

// #############
// # ACCESSORS #
// #############

func TestLookup_GetServiceID(t *testing.T) {
	// GIVEN: a Lookup with a Status containing a ServiceID.
	serviceID := "foo"
	l := &testLookup{
		Lookup: Lookup{
			Status: &status.Status{
				ServiceInfo: serviceinfo.ServiceInfo{
					ID: serviceID,
				},
			},
		},
	}

	// WHEN: GetService is called.
	got := l.GetServiceID()

	// THEN: the ServiceID is returned.
	if got != serviceID {
		t.Errorf(
			"%s\nLookup.GetServiceID() value mismatch\ngot:  %q\nwant: %q",
			packageName, got, serviceID,
		)
	}
}

func TestLookup_GetOptions(t *testing.T) {
	// GIVEN: a Lookup with Options.
	options := &opt.Options{}
	l := &testLookup{
		Lookup: Lookup{
			Options: options,
		},
	}

	// WHEN: GetOptions is called.
	got := l.GetOptions()

	// THEN: the Options are returned.
	if got != options {
		t.Errorf(
			"%s\nLookup.GetOptions() pointer mismatch\ngot:  %p\nwant: %p",
			packageName, options, got,
		)
	}
}

func TestLookup_GetStatus(t *testing.T) {
	// GIVEN: a Lookup with Status.
	svcStatus := &status.Status{}
	l := &testLookup{
		Lookup: Lookup{
			Status: svcStatus,
		},
	}

	// WHEN: GetStatus is called.
	got := l.GetStatus()

	// THEN: the Status is returned.
	if got != svcStatus {
		t.Errorf(
			"%s\nLookup.GetStatus() pointer mismatch\ngot:  %p\nwant: %p",
			packageName, svcStatus, got,
		)
	}
}

func TestLookup_SetStatus(t *testing.T) {
	// GIVEN: a Lookup and a Status.
	l := &Lookup{}
	svcStatus := &status.Status{}

	// WHEN: SetStatus is called.
	l.SetStatus(svcStatus)

	// THEN: the Status is set.
	if l.Status != svcStatus {
		t.Errorf(
			"%s\nLookup.SetStatus(%p) pointer mismatch\ngot:  %v\nwant: %v",
			packageName, svcStatus,
			l.Status, svcStatus,
		)
	}

	// ---

	// GIVEN: a new Status.
	svcStatus = &status.Status{}

	// WHEN: SetStatus is called.
	l.SetStatus(svcStatus)

	// THEN: the Status is set.
	if l.Status != svcStatus {
		t.Errorf(
			"%s\nLookup.SetStatus(%p) pointer mismatch\ngot:  %v\nwant: %v",
			packageName, svcStatus,
			l.Status, svcStatus,
		)
	}
}

func TestLookup_GetDefaults(t *testing.T) {
	// GIVEN: a Lookup with Defaults.
	defaults := &Defaults{}
	l := &testLookup{
		Lookup: Lookup{Defaults: defaults},
	}

	// WHEN: GetDefaults is called.
	got := l.GetDefaults()

	// THEN: the Defaults are returned.
	if got != defaults {
		t.Errorf(
			"%s\nLookup.GetDefaults() pointer mismatch\ngot:  %v\nwant: %v",
			packageName, defaults, got,
		)
	}
}

func TestLookup_GetHardDefaults(t *testing.T) {
	// GIVEN: a Lookup with HardDefaults.
	hardDefaults := &Defaults{}
	l := &testLookup{
		Lookup: Lookup{HardDefaults: hardDefaults},
	}

	// WHEN: GetHardDefaults is called.
	got := l.GetHardDefaults()

	// THEN: the HardDefaults are returned.
	if got != hardDefaults {
		t.Errorf(
			"%s\nLookup.GetHardDefaults() pointer mismatch\ngot:  %v\nwant: %v",
			packageName, hardDefaults, got,
		)
	}
}

// #############
// # INTERFACE #
// #############

func TestLookup_CheckValues(t *testing.T) {
	// GIVEN: a Lookup.
	tests := []struct {
		name     string
		data     string
		errRegex string
	}{
		{
			name:     "no URL",
			data:     "type: url",
			errRegex: `^$`,
		},
		{
			name: "have URL",
			data: test.TrimYAML(`
				type: url
				url: owner/repo
			`),
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := &testLookup{}
			// Apply the YAML.
			if err := decode.Unmarshal("yaml", []byte(tc.data), input); err != nil {
				t.Fatalf(
					"%s\nLookup failed unmarshaling YAML: %v",
					packageName, err,
				)
			}

			_ = test.AssertCheckValuesWithError(
				t,
				packageName,
				tc.errRegex,
				input.CheckValues,
			)
		})
	}
}

func TestLookup_Query(t *testing.T) {
	// GIVEN: a Lookup.
	l := &testLookup{
		Lookup: Lookup{
			Type: "test",
		},
	}

	// WHEN: Query is called.
	gotErr := l.Query(true, logx.LogFrom{})

	// THEN: the function returns an error as it is not implemented.
	if gotErr == nil {
		t.Errorf(
			"%s\nLookup.Query(), unexpected nil error",
			packageName,
		)
	}
}

func TestLookup_InheritSecrets(t *testing.T) {
	// GIVEN: a Lookup and another Lookup to inherit secrets from.
	otherLookup := &testLookup{
		Lookup: Lookup{
			Type: "other",
		},
	}
	secretRefs := &shared.VSecretRef{
		Headers: []shared.OldIntIndex{
			{OldIndex: test.Ptr(0)},
		},
	}
	l := &testLookup{
		Lookup: Lookup{
			Type: "test",
		},
	}
	strBefore := decode.ToYAMLString(l, "")

	// WHEN: InheritSecrets is called.
	l.InheritSecrets(otherLookup, secretRefs)

	// THEN: no secrets are inherited as the function does nothing.
	// As the function does nothing, we just ensure it doesn't panic or error.
	if strAfter := decode.ToYAMLString(l, ""); strBefore != strAfter {
		t.Errorf(
			"%s\nLookup.InheritSecrets(), unexpected change\nbefore: %q\nafter:  %q",
			packageName, strBefore, strBefore,
		)
	}
}
