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
	options := &opt.Options{}
	svcStatus := &status.Status{}
	dvCfg := plainDefaultsConfig()
	l := &Lookup{}

	l.Init(options, svcStatus, dvCfg)

	fieldTests := []test.FieldAssertion{
		{Name: "Options", Got: l.Options, Want: options, Mode: test.CompareSamePointer},
		{Name: "Status", Got: l.Status, Want: svcStatus, Mode: test.CompareSamePointer},
		{Name: "Defaults", Got: l.Defaults, Want: dvCfg.Soft, Mode: test.CompareSamePointer},
		{Name: "HardDefaults", Got: l.HardDefaults, Want: dvCfg.Hard, Mode: test.CompareSamePointer},
	}
	if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
		t.Fatal(testErr)
	}
}

// #############
// # ACCESSORS #
// #############

func TestLookup_GetServiceID(t *testing.T) {
	serviceID := "foo"
	l := &testLookup{
		Status: &status.Status{
			ServiceInfo: serviceinfo.ServiceInfo{
				ID: serviceID,
			},
		},
	}

	got := l.GetServiceID()

	if got != serviceID {
		t.Errorf("mismatch\ngot:  %q\nwant: %q", got, serviceID)
	}
}

func TestLookup_GetOptions(t *testing.T) {
	options := &opt.Options{}
	l := &testLookup{
		Options: options,
	}

	got := l.GetOptions()

	if got != options {
		t.Errorf("mismatch\ngot:  %p\nwant: %p", got, options)
	}
}

func TestLookup_GetStatus(t *testing.T) {
	svcStatus := &status.Status{}
	l := &testLookup{
		Status: svcStatus,
	}

	got := l.GetStatus()

	if got != svcStatus {
		t.Errorf("mismatch\ngot:  %p\nwant: %p", got, svcStatus)
	}
}

func TestLookup_SetStatus(t *testing.T) {
	l := &Lookup{}
	svcStatus := &status.Status{}

	l.SetStatus(svcStatus)

	if l.Status != svcStatus {
		t.Errorf("first SetStatus: Lookup.Status mismatch\ngot:  %p\nwant: %p", l.Status, svcStatus)
	}

	// ---

	svcStatus = &status.Status{}

	l.SetStatus(svcStatus)

	if l.Status != svcStatus {
		t.Errorf(
			"second SetStatus: Lookup.Status mismatch\ngot:  %p\nwant: %p",
			l.Status, svcStatus,
		)
	}
}

func TestLookup_GetDefaults(t *testing.T) {
	defaults := &Defaults{}
	l := &testLookup{
		Defaults: defaults,
	}

	got := l.GetDefaults()

	if got != defaults {
		t.Errorf("mismatch\ngot:  %p\nwant: %p", got, defaults)
	}
}

func TestLookup_GetHardDefaults(t *testing.T) {
	hardDefaults := &Defaults{}
	l := &testLookup{
		HardDefaults: hardDefaults,
	}

	got := l.GetHardDefaults()

	if got != hardDefaults {
		t.Errorf("mismatch\ngot:  %p\nwant: %p", got, hardDefaults)
	}
}

// #############
// # INTERFACE #
// #############

func TestLookup_CheckValues(t *testing.T) {
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
			if err := decode.Unmarshal("yaml", []byte(tc.data), input); err != nil {
				t.Fatalf("decode.Unmarshal() failed: %v", err)
			}

			_ = test.AssertCheckValuesWithError(
				t,
				tc.errRegex,
				input.CheckValues,
			)
		})
	}
}

func TestLookup_Query(t *testing.T) {
	l := &testLookup{
		Type: "test",
	}

	gotErr := l.Query(true, logx.LogFrom{})

	if gotErr == nil {
		t.Error("error mismatch\ngot:  nil\nwant: non-nil")
	}
}

func TestLookup_InheritSecrets(t *testing.T) {
	otherLookup := &testLookup{
		Type: "other",
	}
	secretRefs := &shared.VSecretRef{
		Headers: []shared.OldIntIndex{
			{OldIndex: new(0)},
		},
	}
	l := &testLookup{
		Type: "test",
	}
	strBefore := decode.ToYAMLString(l, "")

	l.InheritSecrets(otherLookup, secretRefs)

	if strAfter := decode.ToYAMLString(l, ""); strBefore != strAfter {
		t.Errorf("Lookup mismatch\ngot:  %q\nwant: %q", strAfter, strBefore)
	}
}
