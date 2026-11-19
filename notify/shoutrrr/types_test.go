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
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/nicholas-fedor/shoutrrr/pkg/format"
	"github.com/nicholas-fedor/shoutrrr/pkg/router"

	"github.com/release-argus/Argus/internal/test"
	shoutrrr_types "github.com/release-argus/Argus/notify/shoutrrr/types"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestSupportedTypes_SafeFields(t *testing.T) {
	// A type missing from SafeFieldsByType shows none of its values, so the lists must match.
	want := slices.Sorted(slices.Values(SupportedTypes))
	got := util.SortedKeys(shoutrrr_types.SafeFieldsByType)

	if testErr := test.AssertSlicesEqual(t, got, want, "SafeFieldsByType keys"); testErr != nil {
		t.Error(testErr)
	}
}

var secretParams = map[string][]string{
	"bark":      {"url"},
	"join":      {"devices"},
	"lark":      {"secret"},
	"notifiarr": {"apikey"},
	"ntfy":      {"attach", "click", "token"},
	"pagerduty": {"details"},
	"pushover":  {"devices", "encryptionkey", "key"},
	"signal":    {"apikey", "password", "token"},
	"smtp":      {"clientsecret", "refreshtoken"},
	"teams":     {"host"},
	"twilio":    {"apikey"},
	"wecom":     {"key"},
	"xmpp":      {"roompassword"},
}

func TestSafeFieldsByType_ShoutrrrParams(t *testing.T) {
	var serviceRouter router.ServiceRouter
	for _, sType := range SupportedTypes {
		// The shoutrrr type sends a raw URL, so has no Shoutrrr service of its own.
		if sType == "shoutrrr" {
			continue
		}

		t.Run(sType, func(t *testing.T) {
			t.Parallel()

			service, err := serviceRouter.NewService(sType)
			if err != nil {
				t.Fatalf("ServiceRouter.NewService() failed: %v", err)
			}
			resolver := format.NewPropKeyResolver(format.GetServiceConfig(service))
			safeFields := shoutrrr_types.SafeFieldsOf(sType)
			safe := slices.Concat(safeFields.Params, safeFields.ParamMaps)

			queryFields := resolver.QueryFields()
			// Each param a Shoutrrr bump adds must be classified before its value is shown.
			for _, key := range queryFields {
				isSafe := slices.Contains(safe, key)
				isSecret := slices.Contains(secretParams[sType], key)
				if isSafe == isSecret {
					t.Errorf(
						"param %q classification mismatch\n"+
							"got:  safe=%t, secret=%t\nwant: one of safe, secret",
						key, isSafe, isSecret,
					)
				}
			}
			for _, key := range secretParams[sType] {
				if !slices.Contains(queryFields, key) {
					t.Errorf(
						"PropKeyResolver.QueryFields() %q mismatch\ngot:  absent\nwant: present",
						key,
					)
				}
			}
		})
	}
}

// ############
// # DECODING #
// ############

func TestShoutrrrs_MarshalJSON(t *testing.T) {
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

			data, err := tc.shoutrrrs.MarshalJSON()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if dataStr := string(data); dataStr != tc.wantStr {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", dataStr, tc.wantStr)
			}
		})
	}
}

func TestShoutrrrs_UnmarshalJSON(t *testing.T) {
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
			name: "duplicate names rejected",
			data: test.TrimJSON(`[
				{"name": "dupe", "type": "slack"},
				{"name": "dupe", "type": "gotify"}
			]`),
			errRegex: test.TrimYAML(`
				^notify\[1\]:
					name: "dupe" <invalid> \(must be unique\)$`,
			),
		},
		{
			name: "missing name rejected",
			data: test.TrimJSON(`[
				{"name": "a", "type": "slack"},
				{"type": "gotify"}
			]`),
			errRegex: test.TrimYAML(`
				^notify\[1\]:
					name: <required>$`,
			),
		},
		{
			name: "empty name rejected",
			data: `[{"name": "", "type": "slack"}]`,
			errRegex: test.TrimYAML(`
				^notify\[0\]:
					name: <required>$`,
			),
		},
		{
			name:     "invalid JSON",
			data:     `[`,
			errRegex: `^jsontext: unexpected EOF after offset 1$`,
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

			var s Shoutrrrs
			err := s.UnmarshalJSON([]byte(tc.data))
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			if gotLen, wantLen := len(s), len(tc.wantKeys); gotLen != wantLen {
				t.Fatalf(
					"len(Shoutrrrs) mismatch\ngot:  %d %q\nwant: %d %q",
					gotLen, util.SortedKeys(s),
					wantLen, util.SortedKeys(tc.wantKeys),
				)
			}
			for id, wantType := range tc.wantKeys {
				got, ok := s[id]
				if !ok {
					t.Errorf("Shoutrrrs[%q] mismatch\ngot:  absent\nwant: present", id)
					continue
				}
				if got == nil {
					t.Fatalf("Shoutrrrs[%q] mismatch\ngot:  nil\nwant: non-nil", id)
				}
				if got.Type != wantType {
					t.Errorf(
						"Shoutrrrs[%q].Type mismatch\ngot:  %q\nwant: %q",
						id, got.Type, wantType,
					)
				}
				if got.ID != id {
					t.Errorf("Shoutrrrs[%q].ID mismatch\ngot:  %q\nwant: %q", id, got.ID, id)
				}
			}
		})
	}
}

func testShoutrrrForCopy(svcStatus *status.Status, id string) *Shoutrrr {
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
	shoutrrr.Failed.Set(id, new(false))

	return shoutrrr
}

func assertFailsShoutrrrState(
	got *status.FailsShoutrrr,
	wantFails map[string]*bool,
	target string,
) error {
	var errs []error
	for _, key := range util.SortedKeys(wantFails) {
		gotStr := test.StringifyPtr(got.Get(key))
		wantStr := test.StringifyPtr(wantFails[key])
		if gotStr != wantStr {
			errs = append(errs, fmt.Errorf(
				"%s.Failed.Get(%q) mismatch\ngot:  %s\nwant: %s",
				target, key,
				gotStr, wantStr,
			))
		}
	}
	return errors.Join(errs...)
}

func TestShoutrrrs_Copy(t *testing.T) {
	tests := []struct {
		name string
		// Number of entries in the map.
		wantLen int
		// Key to reassign in the original map.
		reassignOriginalKey string
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

			origStatus := newTestStatus(tc.name, 2)
			copyStatus := newTestStatus(tc.name, 2)

			var orig Shoutrrrs
			switch tc.wantLen {
			case 0:
				orig = Shoutrrrs{}
			case 1:
				orig = Shoutrrrs{
					"foo": testShoutrrrForCopy(origStatus, "foo"),
				}
			case 2:
				orig = Shoutrrrs{
					"foo": testShoutrrrForCopy(origStatus, "foo"),
					"bar": testShoutrrrForCopy(origStatus, "bar"),
				}
			}

			got := orig.Copy(copyStatus)

			if gotLen, wantLen := len(got), tc.wantLen; gotLen != wantLen {
				t.Fatalf(
					"after Copy: len(copy) mismatch\ngot:  %d %q\nwant: %d",
					gotLen, util.SortedKeys(got), wantLen,
				)
			}

			for key, want := range orig {
				t.Run(key, func(t *testing.T) {
					// Serial: shares orig.

					gotEntry := got[key]
					if gotEntry == nil {
						t.Fatalf("after Copy: copy[%q] mismatch\ngot:  nil\nwant: non-nil", key)
					}
					if gotEntry == want {
						t.Fatalf(
							"after Copy: copy aliases the original\ngot:  %p\nwant: NOT %p",
							gotEntry, want,
						)
					}
					if gotEntry.ServiceStatus != copyStatus {
						t.Errorf(
							"after Copy: copy[%q].ServiceStatus mismatch\ngot:  %p\nwant: %p",
							key, gotEntry.ServiceStatus, copyStatus,
						)
					}
					if testErr := test.AssertMapEqual(
						t,
						gotEntry.Options, want.Options,
						fmt.Sprintf("copy[%q].Options", key),
					); testErr != nil {
						t.Errorf("after Copy: %v", testErr)
					}
					if testErr := test.AssertMapEqual(
						t,
						gotEntry.URLFields, want.URLFields,
						fmt.Sprintf("copy[%q].URLFields", key),
					); testErr != nil {
						t.Errorf("after Copy: %v", testErr)
					}
					if testErr := test.AssertMapEqual(
						t,
						gotEntry.Params, want.Params,
						fmt.Sprintf("copy[%q].Params", key),
					); testErr != nil {
						t.Errorf("after Copy: %v", testErr)
					}

					if tc.reassignOriginalKey != "" {
						reassignTo := testShoutrrrForCopy(origStatus, "replacement")
						orig[tc.reassignOriginalKey] = reassignTo
						if _, ok := got[tc.reassignOriginalKey]; !ok {
							t.Fatalf(
								"reassign: copy[%q] mismatch\ngot:  absent\nwant: present",
								tc.reassignOriginalKey,
							)
						}
						if got[tc.reassignOriginalKey] == reassignTo {
							t.Fatalf(
								"reassign: copy[%q] mismatch\ngot:  %p\nwant: NOT %p",
								tc.reassignOriginalKey, got[tc.reassignOriginalKey], reassignTo,
							)
						}
					}

					if tc.mutateOriginalOptionsKey != "" || tc.mutateOriginalURLFieldsKey != "" || tc.mutateOriginalParamsKey != "" {
						originalVal := gotEntry.Options[tc.mutateOriginalOptionsKey]
						want.Options[tc.mutateOriginalOptionsKey] = tc.mutateOriginalOptionsVal
						if g := gotEntry.Options[tc.mutateOriginalOptionsKey]; g != originalVal {
							t.Fatalf(
								"mutate original: copy[%q].Options mismatch\ngot:  %q\nwant: %q",
								key, g, originalVal,
							)
						}
						originalVal = gotEntry.URLFields[tc.mutateOriginalURLFieldsKey]
						want.URLFields[tc.mutateOriginalURLFieldsKey] = tc.mutateOriginalURLFieldsVal
						if g := gotEntry.URLFields[tc.mutateOriginalURLFieldsKey]; g != originalVal {
							t.Fatalf(
								"mutate original: copy[%q].URLFields mismatch\ngot:  %q\nwant: %q",
								key, g, originalVal,
							)
						}
						originalVal = gotEntry.Params[tc.mutateOriginalParamsKey]
						want.Params[tc.mutateOriginalParamsKey] = tc.mutateOriginalParamsVal
						if g := gotEntry.Params[tc.mutateOriginalParamsKey]; g != originalVal {
							t.Fatalf(
								"mutate original: copy[%q].Params mismatch\ngot:  %q\nwant: %q",
								key, g, originalVal,
							)
						}
					}

					if tc.mutateCopyOptionsKey != "" || tc.mutateCopyURLFieldsKey != "" || tc.mutateCopyParamsKey != "" {
						originalVal := want.Options[tc.mutateCopyOptionsKey]
						gotEntry.Options[tc.mutateCopyOptionsKey] = tc.mutateCopyOptionsVal
						if g := want.Options[tc.mutateCopyOptionsKey]; g != originalVal {
							t.Fatalf(
								"mutate copy: Shoutrrrs[%q].Options mismatch\ngot:  %q\nwant: %q",
								key, g, originalVal,
							)
						}
						originalVal = want.URLFields[tc.mutateCopyURLFieldsKey]
						gotEntry.URLFields[tc.mutateCopyURLFieldsKey] = tc.mutateCopyURLFieldsVal
						if g := want.URLFields[tc.mutateCopyURLFieldsKey]; g != originalVal {
							t.Fatalf(
								"mutate copy: Shoutrrrs[%q].URLFields mismatch\ngot:  %q\nwant: %q",
								key, g, originalVal,
							)
						}
						originalVal = want.Params[tc.mutateCopyParamsKey]
						gotEntry.Params[tc.mutateCopyParamsKey] = tc.mutateCopyParamsVal
						if g := want.Params[tc.mutateCopyParamsKey]; g != originalVal {
							t.Fatalf(
								"mutate copy: Shoutrrrs[%q].Params mismatch\ngot:  %q\nwant: %q",
								key, g, originalVal,
							)
						}
					}
				})
			}
		})
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
			mutateOriginalFails:     map[string]*bool{"notify": new(true)},
			mutateCopyFails:         map[string]*bool{"notify": nil},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			origStatus := newTestStatus(tc.name, 2)
			copyStatus := newTestStatus(tc.name, 2)

			var orig *Shoutrrr
			if !tc.wantNil {
				orig = testShoutrrrForCopy(origStatus, "notify")
			}
			var wantOrigOptions, wantOrigURLFields, wantOrigParams map[string]string
			if orig != nil {
				wantOrigOptions = util.CopyMap(orig.Options)
				wantOrigURLFields = util.CopyMap(orig.URLFields)
				wantOrigParams = util.CopyMap(orig.Params)
			}

			got := orig.Copy(copyStatus)

			if tc.wantNil {
				if got != nil {
					t.Fatal("after Copy: copy mismatch\ngot:  non-nil\nwant: nil")
				}
				return
			}
			if got == nil {
				t.Fatal("after Copy: copy mismatch\ngot:  nil\nwant: non-nil")
			}

			fieldTests := []test.FieldAssertion{
				{Name: "ID", Got: got.ID, Want: orig.ID, Mode: test.CompareEqual},
				{Name: "Type", Got: got.Type, Want: orig.Type, Mode: test.CompareEqual},
				{Name: "Main", Got: got.Main, Want: orig.Main, Mode: test.CompareSamePointer},
				{Name: "Defaults", Got: got.Defaults, Want: orig.Defaults, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: got.HardDefaults, Want: orig.HardDefaults, Mode: test.CompareSamePointer},
				{Name: "ServiceStatus", Got: got.ServiceStatus, Want: copyStatus, Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, fieldTests, "copy"); testErr != nil {
				t.Fatalf("after Copy: %v", testErr)
			}

			if testErr := test.AssertMapEqual(
				t,
				got.Options,
				orig.Options,
				"copy.Options",
			); testErr != nil {
				t.Errorf("after Copy: %v", testErr)
			}
			if testErr := test.AssertMapEqual(
				t,
				got.URLFields,
				orig.URLFields,
				"copy.URLFields",
			); testErr != nil {
				t.Errorf("after Copy: %v", testErr)
			}
			if testErr := test.AssertMapEqual(
				t,
				got.Params,
				orig.Params,
				"copy.Params",
			); testErr != nil {
				t.Errorf("after Copy: %v", testErr)
			}
			if got.Failed == orig.Failed {
				t.Errorf(
					"after Copy: Shoutrrr.Failed: copy aliases the original\n"+
						"got:  %p\nwant: NOT %p",
					got.Failed, orig.Failed,
				)
			}
			if testErr := assertFailsShoutrrrState(
				got.Failed,
				map[string]*bool{"notify": new(false)},
				"copy",
			); testErr != nil {
				t.Errorf("after Copy: %v", testErr)
			}

			wantMutatedOrigOptions := *orig.Options.Copy()
			wantMutatedOrigURLFields := *orig.URLFields.Copy()
			wantMutatedOrigParams := *orig.Params.Copy()
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

			if testErr := test.AssertMapEqual(
				t,
				got.Options,
				wantOrigOptions,
				"copy.Options",
			); testErr != nil {
				t.Errorf("mutate original: %v", testErr)
			}
			if testErr := test.AssertMapEqual(
				t,
				got.URLFields,
				wantOrigURLFields,
				"copy.URLFields",
			); testErr != nil {
				t.Errorf("mutate original: %v", testErr)
			}
			if testErr := test.AssertMapEqual(
				t,
				got.Params,
				wantOrigParams,
				"copy.Params",
			); testErr != nil {
				t.Errorf("mutate original: %v", testErr)
			}
			if testErr := assertFailsShoutrrrState(
				got.Failed,
				map[string]*bool{"notify": new(false)},
				"copy",
			); testErr != nil {
				t.Errorf("mutate original: %v", testErr)
			}

			wantCopyOptions := *got.Options.Copy()
			wantCopyURLFields := *got.URLFields.Copy()
			wantCopyParams := *got.Params.Copy()
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

			if testErr := test.AssertMapEqual(
				t,
				got.Options,
				wantCopyOptions,
				"copy.Options",
			); testErr != nil {
				t.Errorf("mutate copy: %v", testErr)
			}
			if testErr := test.AssertMapEqual(
				t,
				got.URLFields,
				wantCopyURLFields,
				"copy.URLFields",
			); testErr != nil {
				t.Errorf("mutate copy: %v", testErr)
			}
			if testErr := test.AssertMapEqual(
				t,
				got.Params,
				wantCopyParams,
				"copy.Params",
			); testErr != nil {
				t.Errorf("mutate copy: %v", testErr)
			}
			if testErr := assertFailsShoutrrrState(
				got.Failed,
				tc.mutateCopyFails,
				"copy",
			); testErr != nil {
				t.Errorf("mutate copy: %v", testErr)
			}

			if testErr := test.AssertMapEqual(
				t,
				orig.Options,
				wantMutatedOrigOptions,
				"Shoutrrr.Options",
			); testErr != nil {
				t.Errorf("mutate copy: %v", testErr)
			}
			if testErr := test.AssertMapEqual(
				t,
				orig.URLFields,
				wantMutatedOrigURLFields,
				"Shoutrrr.URLFields",
			); testErr != nil {
				t.Errorf("mutate copy: %v", testErr)
			}
			if testErr := test.AssertMapEqual(
				t,
				orig.Params,
				wantMutatedOrigParams,
				"Shoutrrr.Params",
			); testErr != nil {
				t.Errorf("mutate copy: %v", testErr)
			}
			if testErr := assertFailsShoutrrrState(
				orig.Failed,
				tc.mutateOriginalFails,
				"Shoutrrr",
			); testErr != nil {
				t.Errorf("mutate copy: %v", testErr)
			}
		})
	}
}

// #########
// # STATE #
// #########

func TestShoutrrrsDefaults_IsZero(t *testing.T) {
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

			got := tc.shoutrrrsDefaults.IsZero()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestDefaults_IsZero(t *testing.T) {
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
				Type: "discord",
			},
			want: false,
		},
		{
			name: "non-empty/Options",
			defaults: &Defaults{
				Options: map[string]string{
					"delay": "1h",
				},
			},
			want: false,
		},
		{
			name: "non-empty/URLFields",
			defaults: &Defaults{
				URLFields: map[string]string{
					"webhookid": "456",
				},
			},
			want: false,
		},
		{
			name: "non-empty/Params",
			defaults: &Defaults{
				Params: map[string]string{
					"title": "argus",
				},
			},
			want: false,
		},
		{
			name: "non-empty/all",
			defaults: &Defaults{
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
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.defaults.IsZero()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestShoutrrrs_IsZero(t *testing.T) {
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
			want: false, // IsZero counts entries, not their contents.
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

			got := tc.shoutrrrs.IsZero()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestShoutrrr_IsDefault(t *testing.T) {
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
				Type: "discord",
			},
			want: true, // IsDefault ignores Type.
		},
		{
			name: "non-empty Options",
			shoutrrr: &Shoutrrr{
				Options: MapStringStringOmitNull{
					"delay": "1h",
				},
			},
			want: false,
		},
		{
			name: "url fields only",
			shoutrrr: &Shoutrrr{
				URLFields: MapStringStringOmitNull{
					"webhookid": "456",
				},
			},
			want: false,
		},
		{
			name: "non-empty Params",
			shoutrrr: &Shoutrrr{
				Params: MapStringStringOmitNull{
					"title": "argus",
				},
			},
			want: false,
		},
		{
			name: "filled",
			shoutrrr: &Shoutrrr{
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
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.shoutrrr.IsDefault()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

// #############
// # STRINGIFY #
// #############

func TestShoutrrrs_String(t *testing.T) {
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
				"Shoutrrrs.String",
				tc.shoutrrrs.String,
				tc.want,
			)
		})
	}
}

func TestShoutrrr_String(t *testing.T) {
	tests := []struct {
		name     string
		shoutrrr *Shoutrrr
		want     string
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
			name: "filled",
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

			test.AssertStringWithPrefixes(
				t,
				"Shoutrrr.String",
				tc.shoutrrr.String,
				tc.want,
			)
		})
	}
}

func TestShoutrrrsDefaults_String(t *testing.T) {
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
				"ShoutrrrsDefaults.String",
				tc.shoutrrrsDefaults.String,
				tc.want,
			)
		})
	}
}

func TestDefaults_String(t *testing.T) {
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
				"Defaults.String",
				tc.shoutrrr.String,
				tc.want,
			)
		})
	}
}
