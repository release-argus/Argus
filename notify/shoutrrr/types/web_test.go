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

package types

import (
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

func TestSafeFieldsOf(t *testing.T) {
	tests := []struct {
		name  string
		sType string
		want  SafeFields
	}{
		{
			name:  "unknown type/none",
			sType: "",
			want:  SafeFields{},
		},
		{
			name:  "known type/its own",
			sType: "pushover",
			want: SafeFields{
				Params: []string{"expire", "priority", "retry", "timestamp", "title"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := SafeFieldsOf(tc.sType)

			if testErr := test.AssertSlicesEqual(
				t, got.URLFields, tc.want.URLFields, "SafeFields.URLFields",
			); testErr != nil {
				t.Error(testErr)
			}
			if testErr := test.AssertSlicesEqual(
				t, got.URLFieldMaps, tc.want.URLFieldMaps, "SafeFields.URLFieldMaps",
			); testErr != nil {
				t.Error(testErr)
			}
			if testErr := test.AssertSlicesEqual(
				t, got.Params, tc.want.Params, "SafeFields.Params",
			); testErr != nil {
				t.Error(testErr)
			}
			if testErr := test.AssertSlicesEqual(
				t, got.ParamMaps, tc.want.ParamMaps, "SafeFields.ParamMaps",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestSafeFields_Censor(t *testing.T) {
	safe := SafeFields{
		URLFields:    []string{"host"},
		URLFieldMaps: []string{"headers"},
		Params:       []string{"title"},
		ParamMaps:    []string{"details"},
	}

	tests := []struct {
		name                      string
		safe                      SafeFields
		urlFields, params         map[string]string
		wantURLFields, wantParams map[string]string
	}{
		{
			name:          "nil",
			safe:          safe,
			wantURLFields: map[string]string{},
			wantParams:    map[string]string{},
		},
		{
			name:          "safe kept/unlisted censored/empty kept",
			safe:          safe,
			urlFields:     map[string]string{"host": "example.com", "token": "TOKEN", "port": ""},
			params:        map[string]string{"title": "TITLE", "key": "KEY"},
			wantURLFields: map[string]string{"host": "example.com", "token": util.SecretValue, "port": ""},
			wantParams:    map[string]string{"title": "TITLE", "key": util.SecretValue},
		},
		{
			name:          "map censored by value only where listed",
			safe:          safe,
			urlFields:     map[string]string{"headers": `{"A":"b"}`, "query_vars": `{"c":"d"}`},
			params:        map[string]string{"details": `{"A":"b"}`, "headers": `{"A":"b"}`},
			wantURLFields: map[string]string{"headers": `{"A":"<secret>"}`, "query_vars": util.SecretValue},
			wantParams:    map[string]string{"details": `{"A":"<secret>"}`, "headers": util.SecretValue},
		},
		{
			name:          "no safe fields/every value censored",
			safe:          SafeFields{},
			urlFields:     map[string]string{"host": "example.com", "headers": `{"A":"b"}`},
			params:        map[string]string{"title": "TITLE"},
			wantURLFields: map[string]string{"host": util.SecretValue, "headers": util.SecretValue},
			wantParams:    map[string]string{"title": util.SecretValue},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotURLFields := tc.safe.CensorURLFields(tc.urlFields)
			gotParams := tc.safe.CensorParams(tc.params)

			if testErr := test.AssertMapEqual(
				t, gotURLFields, tc.wantURLFields, "SafeFields.CensorURLFields()",
			); testErr != nil {
				t.Error(testErr)
			}
			if testErr := test.AssertMapEqual(
				t, gotParams, tc.wantParams, "SafeFields.CensorParams()",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestCensorMapValues(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty",
			input: "",
			want:  "",
		},
		{
			name:  "empty object",
			input: `{}`,
			want:  `{}`,
		},
		{
			name:  "single key",
			input: `{"Authorization":"Bearer x"}`,
			want:  `{"Authorization":"<secret>"}`,
		},
		{
			name:  "multiple keys/sorted",
			input: `{"X-Foo":"y","Authorization":"Bearer x"}`,
			want:  `{"Authorization":"<secret>","X-Foo":"<secret>"}`,
		},
		{
			name:  "empty value is masked",
			input: `{"X-Foo":""}`,
			want:  `{"X-Foo":"<secret>"}`,
		},
		{
			name:  "invalid/not JSON",
			input: `Bearer x`,
			want:  util.SecretValue,
		},
		{
			name:  "invalid/non-string value",
			input: `{"X-Foo":1}`,
			want:  util.SecretValue,
		},
		{
			name:  "invalid/array",
			input: `["a","b"]`,
			want:  util.SecretValue,
		},
		{
			name:  "already masked",
			input: util.SecretValue,
			want:  util.SecretValue,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := CensorMapValues(tc.input)

			if got != tc.want {
				t.Errorf(
					"CensorMapValues(value=%q) mismatch\ngot:  %q\nwant: %q",
					tc.input, got, tc.want,
				)
			}
		})
	}
}

func TestRestoreMapValues(t *testing.T) {
	tests := []struct {
		name            string
		original, value string
		want            string
	}{
		{
			name:     "no original",
			original: "",
			value:    `{"X-Foo":"<secret>"}`,
			want:     `{"X-Foo":"<secret>"}`,
		},
		{
			name:     "whole value masked/restored",
			original: `{"X-Foo":"y"}`,
			value:    util.SecretValue,
			want:     `{"X-Foo":"y"}`,
		},
		{
			name:     "masked values restored by key",
			original: `{"Authorization":"Bearer x","X-Foo":"y"}`,
			value:    `{"Authorization":"<secret>","X-Foo":"<secret>"}`,
			want:     `{"Authorization":"Bearer x","X-Foo":"y"}`,
		},
		{
			name:     "new value kept",
			original: `{"Authorization":"Bearer x"}`,
			value:    `{"Authorization":"<secret>","X-Foo":"new"}`,
			want:     `{"Authorization":"Bearer x","X-Foo":"new"}`,
		},
		{
			name:     "changed value kept",
			original: `{"Authorization":"Bearer x"}`,
			value:    `{"Authorization":"Bearer z"}`,
			want:     `{"Authorization":"Bearer z"}`,
		},
		{
			name:     "removed key stays removed",
			original: `{"Authorization":"Bearer x","X-Foo":"y"}`,
			value:    `{"Authorization":"<secret>"}`,
			want:     `{"Authorization":"Bearer x"}`,
		},
		{
			name:     "renamed key does not take the old value",
			original: `{"X-Foo":"y"}`,
			value:    `{"X-Bar":"<secret>"}`,
			want:     `{"X-Bar":"<secret>"}`,
		},
		{
			name:     "masked key absent from original stays masked",
			original: `{"X-Foo":"y"}`,
			value:    `{"X-Foo":"<secret>","X-Bar":"<secret>"}`,
			want:     `{"X-Bar":"<secret>","X-Foo":"y"}`,
		},
		{
			name:     "nothing masked/unchanged",
			original: `{"X-Foo":"y"}`,
			value:    `{ "X-Foo": "z" }`,
			want:     `{ "X-Foo": "z" }`,
		},
		{
			name:     "invalid/value not JSON",
			original: `{"X-Foo":"y"}`,
			value:    `not json`,
			want:     `not json`,
		},
		{
			name:     "invalid/original not JSON",
			original: `not json`,
			value:    `{"X-Foo":"<secret>"}`,
			want:     `{"X-Foo":"<secret>"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := RestoreMapValues(tc.original, tc.value)

			if got != tc.want {
				t.Errorf(
					"RestoreMapValues(original=%q, value=%q) mismatch\ngot:  %q\nwant: %q",
					tc.original, tc.value, got, tc.want,
				)
			}
		})
	}
}

func TestRestoreSecrets(t *testing.T) {
	tests := []struct {
		name                   string
		original, target, want map[string]string
	}{
		{
			name:     "nil target",
			original: map[string]string{"headers": `{"A":"b"}`, "token": "TOKEN"},
			target:   nil,
			want:     nil,
		},
		{
			name:     "whole value restored",
			original: map[string]string{"token": "TOKEN"},
			target:   map[string]string{"token": util.SecretValue},
			want:     map[string]string{"token": "TOKEN"},
		},
		{
			name:     "any key restored",
			original: map[string]string{"made_up": "VALUE"},
			target:   map[string]string{"made_up": util.SecretValue},
			want:     map[string]string{"made_up": "VALUE"},
		},
		{
			name:     "map values restored by key",
			original: map[string]string{"headers": `{"A":"b","C":"d"}`},
			target:   map[string]string{"headers": `{"A":"<secret>","E":"f"}`},
			want:     map[string]string{"headers": `{"A":"b","E":"f"}`},
		},
		{
			name:     "no original value stays masked",
			original: map[string]string{},
			target:   map[string]string{"token": util.SecretValue},
			want:     map[string]string{"token": util.SecretValue},
		},
		{
			name:     "changed value kept",
			original: map[string]string{"token": "TOKEN"},
			target:   map[string]string{"token": "NEW"},
			want:     map[string]string{"token": "NEW"},
		},
		{
			name:     "field absent from target is not added",
			original: map[string]string{"query_vars": `{"c":"d"}`},
			target:   map[string]string{},
			want:     map[string]string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			RestoreSecrets(tc.original, tc.target)

			if testErr := test.AssertMapEqual(t, tc.target, tc.want, ""); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}
