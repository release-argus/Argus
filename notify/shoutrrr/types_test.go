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

package shoutrrr

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

// ############
// # DECODING #
// ############

func TestShoutrrrs_MarshalJSON(t *testing.T) {
	// GIVEN: various Shoutrrrs states to marshal.
	tests := []struct {
		name      string
		shoutrrrs *Shoutrrrs
		wantStr   string
	}{
		{
			name:      "nil map -> null",
			shoutrrrs: nil,
			wantStr:   "null",
		},
		{
			name:      "empty map -> empty array",
			shoutrrrs: &Shoutrrrs{},
			wantStr:   "[]",
		},
		{
			name: "two items",
			shoutrrrs: func() *Shoutrrrs {
				m := Shoutrrrs{
					"a": New(
						nil,
						"a",
						"slack",
						nil,
						nil,
						nil,
						nil,
						nil,
						nil,
					),
					"b": New(
						nil,
						"b",
						"gotify",
						nil,
						nil,
						nil,
						nil,
						nil,
						nil,
					),
				}
				return &m
			}(),
			wantStr: test.TrimJSON(`[
				{"type": "slack", "name": "a"},
				{"type": "gotify", "name": "b"}
			]`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: marshaling the Shoutrrrs.
			data, err := tc.shoutrrrs.MarshalJSON()
			if err != nil {
				t.Fatalf(
					"%s\nShoutrrrs MarshalJSON() returned unexpected error: %v",
					packageName, err,
				)
			}

			// THEN: the result matches the expected JSON.
			if dataStr := string(data); dataStr != tc.wantStr {
				t.Errorf(
					"%s\nShoutrrrs MarshalJSON() mismatch\ngot:  %q\nwant: %q",
					packageName, dataStr, tc.wantStr,
				)
			}
		})
	}
}

func TestShoutrrrs_UnmarshalJSON(t *testing.T) {
	// GIVEN: various JSON inputs to unmarshal into Shoutrrrs.
	tests := []struct {
		name     string
		data     string
		errRegex string
		wantKeys map[string]string
	}{
		{
			name: "valid array with two items",
			data: test.TrimJSON(`[
				{"name": "a", "type": "slack"},
				{"name": "b", "type": "gotify"}
			]`),
			errRegex: `^$`,
			wantKeys: map[string]string{
				"a": "slack",
				"b": "gotify",
			},
		},
		{
			name:     "empty array becomes empty map",
			data:     `[]`,
			errRegex: `^$`,
			wantKeys: map[string]string{},
		},
		{
			name:     "null becomes empty map",
			data:     `null`,
			errRegex: `^$`,
			wantKeys: map[string]string{},
		},
		{
			name: "duplicate ids - last wins",
			data: test.TrimJSON(`[
				{"name": "dupe", "type": "slack"},
				{"name": "dupe", "type": "gotify"}
			]`),
			errRegex: `^$`,
			wantKeys: map[string]string{
				"dupe": "gotify",
			},
		},
		{
			name:     "invalid JSON",
			data:     `{`,
			errRegex: `.+`,
		},
		{
			name:     "wrong shape (object instead of array)",
			data:     `{"name": "a", "type": "slack"}`,
			errRegex: `json: .*unmarshal .+$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: unmarshaling JSON into a Shoutrrrs.
			var s Shoutrrrs
			err := s.UnmarshalJSON([]byte(tc.data))

			prefix := fmt.Sprintf(
				"%s\nShoutrrrs.UnmarshalJSON(%q)",
				packageName, tc.data,
			)

			// THEN: errors produced match the regex.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			if e != "" {
				return
			}

			// AND: map keys and types are as expected.
			if gotLen, wantLen := len(s), len(tc.wantKeys); gotLen != wantLen {
				t.Fatalf(
					"%s length mismatch\ngot:  %d\nwant: %d",
					prefix, gotLen, wantLen,
				)
			}
			for id, wantType := range tc.wantKeys {
				got, ok := s[id]
				if !ok {
					t.Errorf(
						"%s missing key %q",
						prefix, id,
					)
				}
				if got == nil {
					t.Fatalf(
						"%s value for key %q is nil",
						prefix, id,
					)
				}
				if got.Type != wantType {
					t.Errorf(
						"%s .Type mismatch for %q\n got:  %q\nwant: %q",
						prefix, id,
						got.Type, wantType,
					)
				}
				if got.ID != id {
					t.Errorf(
						"%s .ID mismatch for key %q\n got:  %q\nwant: %q",
						packageName, id,
						got.ID, id,
					)
				}
			}
		})
	}
}

// #########
// # STATE #
// #########

func TestShoutrrrsDefaults_IsZero(t *testing.T) {
	// GIVEN: ShoutrrrsDefaults.
	tests := []struct {
		name              string
		shoutrrrsDefaults *ShoutrrrsDefaults
		want              bool
	}{
		{
			name:              "empty/no elements",
			shoutrrrsDefaults: &ShoutrrrsDefaults{},
			want:              true,
		},
		{
			name: "empty/one element",
			shoutrrrsDefaults: &ShoutrrrsDefaults{
				"foo": &Defaults{},
			},
			want: true,
		},
		{
			name: "non-empty/one element",
			shoutrrrsDefaults: &ShoutrrrsDefaults{
				"foo": NewDefaults(
					"discord",
					nil, nil, nil,
				),
			},
			want: false,
		},
		{
			name: "non-empty/multiple elements",
			shoutrrrsDefaults: &ShoutrrrsDefaults{
				"foo": NewDefaults(
					"discord",
					nil, nil, nil,
				),
				"bar": NewDefaults(
					"gotify",
					nil, nil, nil,
				),
			},
			want: false,
		},
		{
			name: "mixed",
			shoutrrrsDefaults: &ShoutrrrsDefaults{
				"foo": &Defaults{},
				"bar": NewDefaults(
					"gotify",
					nil, nil, nil,
				),
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.shoutrrrsDefaults.IsZero()

			// THEN: the result is expected.
			if got != tc.want {
				t.Errorf(
					"%s\nShoutrrrsDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestDefaults_IsZero(t *testing.T) {
	// GIVEN: Defaults.
	tests := []struct {
		name     string
		defaults *Defaults
		want     bool
	}{
		{
			name:     "nil",
			defaults: nil,
			want:     true,
		},
		{
			name:     "empty",
			defaults: &Defaults{},
			want:     true,
		},
		{
			name: "non-empty/Type",
			defaults: &Defaults{
				Base: Base{
					Type: "discord",
				},
			},
			want: false,
		},
		{
			name: "non-empty/Options",
			defaults: &Defaults{
				Base: Base{
					Options: map[string]string{
						"delay": "1h",
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty/URLFields",
			defaults: &Defaults{
				Base: Base{
					URLFields: map[string]string{
						"webhookid": "456",
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty/Params",
			defaults: &Defaults{
				Base: Base{
					Params: map[string]string{
						"title": "argus",
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty/all",
			defaults: &Defaults{
				Base: Base{
					Type: "discord",
					Options: map[string]string{
						"delay": "1h",
					},
					URLFields: map[string]string{
						"webhookid": "456",
					},
					Params: map[string]string{
						"title": "argus",
					},
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.defaults.IsZero()

			// THEN: the result is expected.
			if got != tc.want {
				t.Errorf(
					"%s\nDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestShoutrrrs_IsZero(t *testing.T) {
	// GIVEN: Shoutrrrs.
	tests := []struct {
		name      string
		shoutrrrs *Shoutrrrs
		want      bool
	}{
		{
			name:      "nil",
			shoutrrrs: nil,
			want:      true,
		},
		{
			name:      "empty/0 elements",
			shoutrrrs: &Shoutrrrs{},
			want:      true,
		},
		{
			name: "non-empty/1 element",
			shoutrrrs: &Shoutrrrs{
				"foo": New(nil, "", "discord", nil, nil, nil, nil, nil, nil),
			},
			want: false,
		},
		{
			name: "empty/1 element",
			shoutrrrs: &Shoutrrrs{
				"bop": &Shoutrrr{},
			},
			want: false,
		},
		{
			name: "non-empty/multiple elements",
			shoutrrrs: &Shoutrrrs{
				"foo": New(nil, "", "discord", nil, nil, nil, nil, nil, nil),
				"bar": New(nil, "", "gotify", nil, nil, nil, nil, nil, nil),
			},
			want: false,
		},
		{
			name: "mixed",
			shoutrrrs: &Shoutrrrs{
				"foo": New(nil, "", "discord", nil, nil, nil, nil, nil, nil),
				"bar": New(nil, "", "gotify", nil, nil, nil, nil, nil, nil),
				"baz": New(nil, "", "slack", nil, nil, nil, nil, nil, nil),
				"bop": &Shoutrrr{},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.shoutrrrs.IsZero()

			// THEN: the result is expected.
			if got != tc.want {
				t.Errorf(
					"%s\nShoutrrrs.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestShoutrrr_IsDefault(t *testing.T) {
	// GIVEN: a Shoutrrr.
	tests := []struct {
		name     string
		shoutrrr *Shoutrrr
		want     bool
	}{
		{
			name:     "empty",
			shoutrrr: &Shoutrrr{},
			want:     true,
		},
		{
			name: "initialised empty maps",
			shoutrrr: New(
				nil, "foo",
				"discord",
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
			),
			want: true,
		},
		{
			name: "non-empty Type",
			shoutrrr: &Shoutrrr{
				Base: Base{
					Type: "discord",
				},
			},
			want: true,
		},
		{
			name: "non-empty Options",
			shoutrrr: &Shoutrrr{
				Base: Base{
					Options: MapStringStringOmitNull{
						"delay": "1h",
					},
				},
			},
			want: false,
		},
		{
			name: "url fields only",
			shoutrrr: &Shoutrrr{
				Base: Base{
					URLFields: MapStringStringOmitNull{
						"webhookid": "456",
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty Params",
			shoutrrr: &Shoutrrr{
				Base: Base{
					Params: MapStringStringOmitNull{
						"title": "argus",
					},
				},
			},
			want: false,
		},
		{
			name: "filled",
			shoutrrr: &Shoutrrr{
				Base: Base{
					Type: "discord",
					Options: MapStringStringOmitNull{
						"delay": "1h",
					},
					URLFields: MapStringStringOmitNull{
						"webhookid": "456",
					},
					Params: MapStringStringOmitNull{
						"title": "argus",
					},
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsDefault is called on it.
			got := tc.shoutrrr.IsDefault()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nShoutrrr.IsDefault() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func testShoutrrrForCopy(t *testing.T, svcStatus *status.Status, id string) *Shoutrrr {
	t.Helper()

	main := NewDefaults(
		"discord",
		map[string]string{"main": "1"},
		map[string]string{"main": "1"},
		map[string]string{"main": "1"},
	)
	defaults := NewDefaults(
		"discord",
		map[string]string{"def": "2"},
		map[string]string{"def": "2"},
		map[string]string{"def": "2"},
	)
	hardDefaults := NewDefaults(
		"discord",
		map[string]string{"hard": "3"},
		map[string]string{"hard": "3"},
		map[string]string{"hard": "3"},
	)

	shoutrrr := New(
		&svcStatus.Fails.Shoutrrr,
		id,
		"discord",
		map[string]string{"opt": "1"},
		map[string]string{"host": "example.com"},
		map[string]string{"title": "argus"},
		main,
		defaults,
		hardDefaults,
	)
	shoutrrr.ServiceStatus = svcStatus
	shoutrrr.Failed.Set(id, test.Ptr(false))

	return shoutrrr
}

func assertFailsShoutrrrState(
	t *testing.T,
	got *status.FailsShoutrrr,
	wantFails map[string]*bool,
	prefix, target string,
) {
	t.Helper()

	for key, wantFail := range wantFails {
		gotFail := got.Get(key)
		gotStr := test.StringifyPtr(gotFail)
		wantStr := test.StringifyPtr(wantFail)
		if gotStr != wantStr {
			t.Errorf(
				"%s %s[%q] mismatch\ngot:  %s\nwant: %s",
				prefix, target, key,
				gotStr, wantStr,
			)
		}
	}
}

func TestShoutrrr_Copy(t *testing.T) {
	tests := []struct {
		name     string
		shoutrrr *Shoutrrr
		// If true, the receiver is nil.
		wantNil bool
		// Mutations to verify that Copy doesn't alias the underlying maps.
		mutateOriginalOptions, mutateOriginalURLFields, mutateOriginalParams map[string]string
		mutateCopyOptions, mutateCopyURLFields, mutateCopyParams             map[string]string
		mutateOriginalFails                                                  map[string]*bool
		mutateCopyFails                                                      map[string]*bool
	}{
		{
			name:    "nil receiver",
			wantNil: true,
		},
		{
			name:                    "copies fields and deep-copies maps and fails",
			mutateOriginalOptions:   map[string]string{"opt": "mutated"},
			mutateCopyOptions:       map[string]string{"opt": "copy-mutated"},
			mutateOriginalURLFields: map[string]string{"host": "mutated"},
			mutateCopyURLFields:     map[string]string{"host": "copy-mutated"},
			mutateOriginalParams:    map[string]string{"title": "mutated"},
			mutateCopyParams:        map[string]string{"title": "copy-mutated"},
			mutateOriginalFails:     map[string]*bool{"notify": test.Ptr(true)},
			mutateCopyFails:         map[string]*bool{"notify": test.Ptr(false)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// GIVEN: a Status.
			origStatus, _ := statustest.New("yaml", nil)
			origStatus.Fails.Shoutrrr.Init(2)
			copyStatus, _ := statustest.New("yaml", nil)
			copyStatus.Fails.Shoutrrr.Init(2)

			// AND: a Shoutrrr.
			var orig *Shoutrrr
			if !tc.wantNil {
				orig = testShoutrrrForCopy(t, origStatus, "notify")
			}
			var wantOrigOptions, wantOrigURLFields, wantOrigParams map[string]string
			if orig != nil {
				wantOrigOptions = util.CopyMap(orig.Options)
				wantOrigURLFields = util.CopyMap(orig.URLFields)
				wantOrigParams = util.CopyMap(orig.Params)
			}

			// WHEN: Copy is called.
			got := orig.Copy(copyStatus)

			prefix := fmt.Sprintf("%s\nShoutrrr.Copy()", packageName)

			// THEN: nil handling.
			if tc.wantNil {
				if got != nil {
					t.Fatalf("%s got %v want nil", prefix, got)
				}
				return
			}
			if got == nil {
				t.Fatalf("%s got nil want non-nil", prefix)
			}

			// AND: scalar fields and shared defaults pointers match.
			fieldTests := []test.FieldAssertion{
				{Name: "ID", Got: got.ID, Want: orig.ID, Mode: test.CompareEqual},
				{Name: "Type", Got: got.Type, Want: orig.Type, Mode: test.CompareEqual},
				{Name: "Main", Got: got.Main, Want: orig.Main, Mode: test.CompareSamePointer},
				{Name: "Defaults", Got: got.Defaults, Want: orig.Defaults, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: got.HardDefaults, Want: orig.HardDefaults, Mode: test.CompareSamePointer},
				{Name: "ServiceStatus", Got: got.ServiceStatus, Want: copyStatus, Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Shoutrrr"); err != nil {
				t.Fatal(err)
			}

			// AND: maps and fails are deep-copied.
			if testErr := test.AssertMapEqual(
				t,
				got.Options,
				orig.Options,
				prefix,
				"Options",
			); testErr != nil {
				t.Error(testErr)
			}
			if testErr := test.AssertMapEqual(
				t,
				got.URLFields,
				orig.URLFields,
				prefix,
				"URLFields",
			); testErr != nil {
				t.Error(testErr)
			}
			if testErr := test.AssertMapEqual(
				t,
				got.Params,
				orig.Params,
				prefix,
				"Params",
			); testErr != nil {
				t.Error(testErr)
			}
			if got.Failed == orig.Failed {
				t.Errorf("%s Failed should be a distinct copy", prefix)
			}
			assertFailsShoutrrrState(
				t,
				got.Failed,
				map[string]*bool{"notify": test.Ptr(false)},
				prefix,
				"Failed",
			)

			// WHEN: the original is mutated.
			var wantMutatedOrigOptions, wantMutatedOrigURLFields, wantMutatedOrigParams MapStringStringOmitNull
			if got != nil {
				wantMutatedOrigOptions = *got.Options.Copy()
				wantMutatedOrigURLFields = *got.URLFields.Copy()
				wantMutatedOrigParams = *got.Params.Copy()
			}
			for k, v := range tc.mutateOriginalOptions {
				orig.Options[k] = v
				wantMutatedOrigOptions[k] = v
			}
			for k, v := range tc.mutateOriginalURLFields {
				orig.URLFields[k] = v
				wantMutatedOrigURLFields[k] = v
			}
			for k, v := range tc.mutateOriginalParams {
				orig.Params[k] = v
				wantMutatedOrigParams[k] = v
			}
			for k, v := range tc.mutateOriginalFails {
				orig.Failed.Set(k, v)
			}

			// THEN: the copy is unchanged.
			if testErr := test.AssertMapEqual(
				t,
				got.Options,
				wantOrigOptions,
				fmt.Sprintf("%s after mutating original", prefix),
				"Options",
			); testErr != nil {
				t.Error(testErr)
			}
			if testErr := test.AssertMapEqual(
				t,
				got.URLFields,
				wantOrigURLFields,
				fmt.Sprintf("%s after mutating original", prefix),
				"URLFields",
			); testErr != nil {
				t.Error(testErr)
			}
			if testErr := test.AssertMapEqual(
				t,
				got.Params,
				wantOrigParams,
				fmt.Sprintf("%s after mutating original", prefix),
				"Params",
			); testErr != nil {
				t.Error(testErr)
			}
			assertFailsShoutrrrState(
				t,
				got.Failed,
				map[string]*bool{"notify": test.Ptr(false)},
				fmt.Sprintf("%s after mutating original", prefix),
				"Failed",
			)

			// WHEN: the copy is mutated.
			var wantCopyOptions, wantCopyURLFields, wantCopyParams MapStringStringOmitNull
			if got != nil {
				wantCopyOptions = *got.Options.Copy()
				wantCopyURLFields = *got.URLFields.Copy()
				wantCopyParams = *got.Params.Copy()
			}
			for k, v := range tc.mutateCopyOptions {
				got.Options[k] = v
				wantCopyOptions[k] = v
			}
			for k, v := range tc.mutateCopyURLFields {
				got.URLFields[k] = v
				wantCopyURLFields[k] = v
			}
			for k, v := range tc.mutateCopyParams {
				got.Params[k] = v
				wantCopyParams[k] = v
			}
			for k, v := range tc.mutateCopyFails {
				got.Failed.Set(k, v)
			}

			// THEN: the copy reflects mutations.
			//   Options.
			if testErr := test.AssertMapEqual(
				t,
				got.Options,
				wantCopyOptions,
				fmt.Sprintf("%s after mutating copy", prefix),
				"Options",
			); testErr != nil {
				t.Error(testErr)
			}
			//   URLFields.
			if testErr := test.AssertMapEqual(
				t,
				got.URLFields,
				wantCopyURLFields,
				fmt.Sprintf("%s after mutating copy", prefix),
				"URLFields",
			); testErr != nil {
				t.Error(testErr)
			}
			//   Params.
			if testErr := test.AssertMapEqual(
				t,
				got.Params,
				wantCopyParams,
				fmt.Sprintf("%s after mutating copy", prefix),
				"Params",
			); testErr != nil {
				t.Error(testErr)
			}
			//   Failed.
			assertFailsShoutrrrState(
				t,
				got.Failed,
				tc.mutateCopyFails,
				fmt.Sprintf("%s after mutating copy", prefix),
				"Failed",
			)

			// AND: the original does not reflect copy mutations.
			//   Options.
			if testErr := test.AssertMapEqual(
				t,
				orig.Options,
				wantMutatedOrigOptions,
				fmt.Sprintf("%s after mutating copy", prefix),
				"original Options",
			); testErr != nil {
				t.Error(testErr)
			}
			//   URLFields.
			if testErr := test.AssertMapEqual(
				t,
				orig.URLFields,
				wantMutatedOrigURLFields,
				fmt.Sprintf("%s after mutating copy", prefix),
				"original URLFields",
			); testErr != nil {
				t.Error(testErr)
			}
			//   Params.
			if testErr := test.AssertMapEqual(
				t,
				orig.Params,
				wantMutatedOrigParams,
				fmt.Sprintf("%s after mutating copy", prefix),
				"original Params",
			); testErr != nil {
				t.Error(testErr)
			}
			//   Failed.
			assertFailsShoutrrrState(
				t,
				orig.Failed,
				tc.mutateOriginalFails,
				fmt.Sprintf("%s after mutating copy", prefix),
				"original fails",
			)
		})
	}
}

func TestShoutrrrs_Copy(t *testing.T) {
	tests := []struct {
		name      string
		shoutrrrs Shoutrrrs
		// Number of entries in the map.
		wantLen int
		// Key to reassign in the original map.
		reassignOriginalKey string
		reassignTo          *Shoutrrr
		// Mutations to verify that Copy doesn't alias the underlying maps.
		mutateOriginalOptionsKey, mutateOriginalURLFieldsKey, mutateOriginalParamsKey string
		mutateOriginalOptionsVal, mutateOriginalURLFieldsVal, mutateOriginalParamsVal string
		mutateCopyOptionsKey, mutateCopyURLFieldsKey, mutateCopyParamsKey             string
		mutateCopyOptionsVal, mutateCopyURLFieldsVal, mutateCopyParamsVal             string
	}{
		{
			name:    "empty map",
			wantLen: 0,
		},
		{
			name:    "copies each entry",
			wantLen: 2,

			mutateOriginalOptionsKey:   "foo",
			mutateOriginalURLFieldsKey: "foo",
			mutateOriginalParamsKey:    "foo",
			mutateOriginalOptionsVal:   "mutated",
			mutateOriginalURLFieldsVal: "mutated",
			mutateOriginalParamsVal:    "mutated",

			mutateCopyOptionsKey:   "bar",
			mutateCopyURLFieldsKey: "bar",
			mutateCopyParamsKey:    "bar",
			mutateCopyOptionsVal:   "copy-mutated",
			mutateCopyURLFieldsVal: "copy-mutated",
			mutateCopyParamsVal:    "copy-mutated",
		},
		{
			name:                "reassigning map entry does not affect copy",
			wantLen:             1,
			reassignOriginalKey: "foo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			origStatus, _ := statustest.New("yaml", nil)
			origStatus.Fails.Shoutrrr.Init(2)
			copyStatus, _ := statustest.New("yaml", nil)
			copyStatus.Fails.Shoutrrr.Init(2)

			var orig Shoutrrrs
			switch tc.wantLen {
			case 0:
				orig = Shoutrrrs{}
			case 1:
				orig = Shoutrrrs{
					"foo": testShoutrrrForCopy(t, origStatus, "foo"),
				}
			case 2:
				orig = Shoutrrrs{
					"foo": testShoutrrrForCopy(t, origStatus, "foo"),
					"bar": testShoutrrrForCopy(t, origStatus, "bar"),
				}
			}

			// WHEN: Copy is called.
			got := orig.Copy(copyStatus)

			prefix := fmt.Sprintf("%s\nShoutrrrs.Copy()", packageName)

			// THEN: length matches.
			if gotLen, wantLen := len(got), tc.wantLen; gotLen != wantLen {
				t.Fatalf(
					"%s length mismatch\ngot:  %d\nwant: %d",
					prefix, gotLen, wantLen,
				)
			}

			// AND: each entry is a distinct copy.
			for key, want := range orig {
				gotEntry := got[key]
				if gotEntry == nil {
					t.Fatalf("%s missing copied entry %q", prefix, key)
				}
				if gotEntry == want {
					t.Fatalf("%s entry %q should be a distinct copy\ngot:  %p\nwant: %p",
						prefix, key,
						gotEntry, want,
					)
				}

				if gotEntry.ServiceStatus != copyStatus {
					t.Errorf("%s entry %q should use provided Status\ngot:  %p\nwant: %p",
						prefix, key,
						gotEntry.ServiceStatus, copyStatus,
					)

					fieldTests := []test.FieldAssertion{
						{Name: "ServiceStatus", Got: gotEntry.ServiceStatus, Want: copyStatus, Mode: test.CompareSamePointer},
						{Name: "Options", Got: gotEntry.Options, Want: want.Options, Mode: test.CompareSamePointer},
						{Name: "URLFields", Got: gotEntry.URLFields, Want: want.URLFields, Mode: test.CompareSamePointer},
						{Name: "Params", Got: gotEntry.Params, Want: want.Params, Mode: test.CompareSamePointer},
						{Name: "Failed", Got: gotEntry.Failed, Want: want.Failed, Mode: test.CompareSamePointer},
					}
					if err := test.AssertFields(t, fieldTests, prefix, "Shoutrrr"); err != nil {
						t.Fatal(err)
					}
				}
				// Options.
				if testErr := test.AssertMapEqual(
					t,
					gotEntry.Options,
					want.Options,
					prefix,
					fmt.Sprintf("%q options", key),
				); testErr != nil {
					t.Error(testErr)
				}
				// URLFields.
				if testErr := test.AssertMapEqual(
					t,
					gotEntry.URLFields,
					want.URLFields,
					prefix,
					fmt.Sprintf("%q URLFields", key),
				); testErr != nil {
					t.Error(testErr)
				}
				// Params.
				if testErr := test.AssertMapEqual(
					t,
					gotEntry.Params,
					want.Params,
					prefix,
					fmt.Sprintf("%q Params", key),
				); testErr != nil {
					t.Error(testErr)
				}

				// AND: reassigning map entry does not affect copy.
				//   Options.
				if tc.reassignOriginalKey != "" {
					reassignTo := tc.reassignTo
					if reassignTo == nil {
						reassignTo = testShoutrrrForCopy(t, origStatus, "replacement")
					}
					orig[tc.reassignOriginalKey] = reassignTo
					if _, ok := got[tc.reassignOriginalKey]; !ok {
						t.Fatalf("%s missing entry %q after reassigning original", prefix, tc.reassignOriginalKey)
					}
					if got[tc.reassignOriginalKey] == reassignTo {
						t.Fatalf(
							"%s entry %q should not point at the replacement value",
							prefix,
							tc.reassignOriginalKey,
						)
					}
				}

				// AND: mutating original map entries does not affect copy.
				if tc.mutateOriginalOptionsKey != "" || tc.mutateOriginalURLFieldsKey != "" || tc.mutateOriginalParamsKey != "" {
					// Options.
					originalVal := gotEntry.Options[tc.mutateOriginalOptionsKey]
					want.Options[tc.mutateOriginalOptionsKey] = tc.mutateOriginalOptionsVal
					if g := gotEntry.Options[tc.mutateOriginalOptionsKey]; g != originalVal {
						t.Fatalf("%s Options mismatch after mutating original [%q]=%q\ngot:  %q\nwant: %q",
							prefix, tc.mutateOriginalOptionsKey, tc.mutateOriginalOptionsVal,
							g, originalVal,
						)
					}
					// URLFields.
					originalVal = gotEntry.URLFields[tc.mutateOriginalURLFieldsKey]
					want.URLFields[tc.mutateOriginalURLFieldsKey] = tc.mutateOriginalURLFieldsVal
					if g := gotEntry.URLFields[tc.mutateOriginalURLFieldsKey]; g != originalVal {
						t.Fatalf("%s URLFields mismatch after mutating original [%q]=%q\ngot:  %q\nwant: %q",
							prefix, tc.mutateOriginalURLFieldsKey, tc.mutateOriginalURLFieldsVal,
							g, originalVal,
						)
					}
					// Params.
					originalVal = gotEntry.Params[tc.mutateOriginalParamsKey]
					want.Params[tc.mutateOriginalParamsKey] = tc.mutateOriginalParamsVal
					if g := gotEntry.Params[tc.mutateOriginalParamsKey]; g != originalVal {
						t.Fatalf("%s Params mismatch after mutating original [%q]=%q\ngot:  %q\nwant: %q",
							prefix, tc.mutateOriginalParamsKey, tc.mutateOriginalParamsVal,
							g, originalVal,
						)
					}
				}

				// AND: mutating copy map entries does not affect original.
				if tc.mutateCopyOptionsKey != "" || tc.mutateCopyURLFieldsKey != "" || tc.mutateCopyParamsKey != "" {
					// Options.
					originalVal := want.Options[tc.mutateCopyOptionsKey]
					gotEntry.Options[tc.mutateCopyOptionsKey] = tc.mutateCopyOptionsVal
					if g := want.Options[tc.mutateCopyOptionsKey]; g != originalVal {
						t.Fatalf("%s Options mismatch after mutating copy\ngot:  %q\nwant: %q",
							prefix, g, originalVal,
						)
					}
					// URLFields.
					originalVal = want.URLFields[tc.mutateCopyURLFieldsKey]
					gotEntry.URLFields[tc.mutateCopyURLFieldsKey] = tc.mutateCopyURLFieldsVal
					if g := want.URLFields[tc.mutateCopyURLFieldsKey]; g != originalVal {
						t.Fatalf("%s URLFields mismatch after mutating copy\ngot:  %q\nwant: %q",
							prefix, g, originalVal,
						)
					}
					// Params.
					originalVal = want.Params[tc.mutateCopyParamsKey]
					gotEntry.Params[tc.mutateCopyParamsKey] = tc.mutateCopyParamsVal
					if g := want.Params[tc.mutateCopyParamsKey]; g != originalVal {
						t.Fatalf("%s Params mismatch after mutating copy\ngot:  %q\nwant: %q",
							prefix, g, originalVal,
						)
					}
				}
			}
		})
	}
}

// #############
// # STRINGIFY #
// #############

func TestShoutrrrs_String(t *testing.T) {
	// GIVEN: Shoutrrrs.
	tests := []struct {
		name      string
		shoutrrrs *Shoutrrrs
		want      string
	}{
		{
			name:      "nil",
			shoutrrrs: nil,
			want:      "",
		},
		{
			name:      "empty",
			shoutrrrs: &Shoutrrrs{},
			want:      "{}\n",
		},
		{
			name: "one element",
			shoutrrrs: &Shoutrrrs{
				"foo": New(
					nil,
					"", "discord",
					nil, nil, nil,
					nil, nil, nil,
				),
			},
			want: test.TrimYAML(`
				foo:
					type: discord
			`),
		},
		{
			name: "multiple elements",
			shoutrrrs: &Shoutrrrs{
				"foo": New(
					nil,
					"", "discord",
					nil, nil, nil,
					nil, nil, nil,
				),
				"bar": New(
					nil,
					"", "gotify",
					nil, nil, nil,
					nil, nil, nil,
				),
			},
			want: test.TrimYAML(`
				bar:
					type: gotify
				foo:
					type: discord
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				packageName,
				tc.shoutrrrs.String,
				tc.want,
			)
		})
	}
}

func TestShoutrrr_String(t *testing.T) {
	// GIVEN: a Shoutrrr.
	tests := []struct {
		name          string
		shoutrrr      *Shoutrrr
		latestVersion string
		want          string
	}{
		{
			name:     "nil",
			shoutrrr: nil,
			want:     "",
		},
		{
			name:     "empty",
			shoutrrr: &Shoutrrr{},
			want:     "{}\n",
		},
		{
			name:          "filled",
			latestVersion: "1.2.3",
			shoutrrr: New(
				nil,
				"foo", "discord",
				map[string]string{
					"delay": "1h",
				},
				map[string]string{
					"webhookid": "456",
				},
				map[string]string{
					"title": "argus",
				},
				NewDefaults(
					"", nil,
					map[string]string{
						"token": "bar",
					},
					nil,
				),
				NewDefaults(
					"",
					map[string]string{
						"delay": "2h",
					},
					nil, nil,
				),
				NewDefaults(
					"",
					map[string]string{
						"delay": "3h",
					},
					nil, nil,
				),
			),
			want: test.TrimYAML(`
				type: discord
				options:
					delay: 1h
				url_fields:
					webhookid: '456'
				params:
					title: argus
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.latestVersion != "" {
				if tc.shoutrrr.ServiceStatus == nil {
					tc.shoutrrr.ServiceStatus = status.New(
						nil, nil, nil,
						"",
						"", "",
						"", "",
						"",
						&dashboard.Options{},
					)
				}
				tc.shoutrrr.ServiceStatus.SetLatestVersion(tc.latestVersion, "", false)
			}

			test.AssertStringWithPrefixes(
				t,
				packageName,
				tc.shoutrrr.String,
				tc.want,
			)
		})
	}
}

func TestShoutrrrsDefaults_String(t *testing.T) {
	// GIVEN: Shoutrrrs.
	tests := []struct {
		name              string
		shoutrrrsDefaults *ShoutrrrsDefaults
		want              string
	}{
		{
			name:              "nil",
			shoutrrrsDefaults: nil,
			want:              "",
		},
		{
			name:              "empty",
			shoutrrrsDefaults: &ShoutrrrsDefaults{},
			want:              "{}\n",
		},
		{
			name: "one empty element",
			shoutrrrsDefaults: &ShoutrrrsDefaults{
				"foo": &Defaults{},
			},
			want: "foo: {}\n",
		},
		{
			name: "one non-empty element",
			shoutrrrsDefaults: &ShoutrrrsDefaults{
				"foo": NewDefaults(
					"discord",
					nil, nil, nil,
				),
			},
			want: test.TrimYAML(`
				foo:
					type: discord
			`),
		},
		{
			name: "multiple non-empty elements",
			shoutrrrsDefaults: &ShoutrrrsDefaults{
				"foo": NewDefaults(
					"discord",
					nil, nil, nil,
				),
				"bar": NewDefaults(
					"gotify",
					nil, nil, nil,
				),
			},
			want: test.TrimYAML(`
				bar:
					type: gotify
				foo:
					type: discord
			`),
		},
		{
			name: "multiple empty and non-empty elements",
			shoutrrrsDefaults: &ShoutrrrsDefaults{
				"foo": NewDefaults(
					"discord",
					nil, nil, nil,
				),
				"biz": &Defaults{},
				"bar": NewDefaults(
					"gotify",
					nil, nil, nil,
				),
			},
			want: test.TrimYAML(`
				bar:
					type: gotify
				biz: {}
				foo:
					type: discord
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				packageName,
				tc.shoutrrrsDefaults.String,
				tc.want,
			)
		})
	}
}

func TestDefaults_String(t *testing.T) {
	// GIVEN: Defaults.
	tests := []struct {
		name     string
		shoutrrr *Defaults
		want     string
	}{
		{
			name:     "nil",
			shoutrrr: nil,
			want:     "",
		},
		{
			name:     "empty",
			shoutrrr: &Defaults{},
			want:     "{}\n",
		},
		{
			name: "filled",
			shoutrrr: NewDefaults(
				"discord",
				map[string]string{
					"delay": "1h",
				},
				map[string]string{
					"webhookid": "456",
				},
				map[string]string{
					"title": "argus",
				},
			),
			want: test.TrimYAML(`
				type: discord
				options:
					delay: 1h
				url_fields:
					webhookid: '456'
				params:
					title: argus
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				packageName,
				tc.shoutrrr.String,
				tc.want,
			)
		})
	}
}
