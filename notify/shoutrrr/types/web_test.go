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
	"slices"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

func TestSafeFieldsByType_ParamJSONMaps(t *testing.T) {
	for _, sType := range util.SortedKeys(SafeFieldsByType) {
		t.Run(sType, func(t *testing.T) {
			t.Parallel()

			for param := range SafeFieldsByType[sType].ParamJSONMaps {
				// A param outside ParamJSON is censored whole, so no marker reaches its sites.
				if _, ok := SafeFieldsByType[sType].ParamJSON[param]; !ok {
					t.Errorf(
						"SafeFields.ParamJSON[%q] mismatch\ngot:  absent\nwant: present",
						param,
					)
				}
			}
		})
	}
}

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
		{
			name:  "known type/JSON paths",
			sType: "gotify",
			want: SafeFields{
				URLFields: []string{"host", "path", "port"},
				Params: []string{
					"date", "disabletls", "insecureskipverify", "priority", "skiptlsverify", "title",
					"useheader",
				},
				ParamJSON:     map[string][]string{"extras": {"client::display.contentType"}},
				ParamJSONMaps: map[string][]string{"extras": {""}},
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
			if testErr := test.AssertMapEqualFunc(
				t, got.ParamJSON, tc.want.ParamJSON, slices.Equal, "SafeFields.ParamJSON",
			); testErr != nil {
				t.Error(testErr)
			}
			if testErr := test.AssertMapEqualFunc(
				t, got.ParamJSONMaps, tc.want.ParamJSONMaps, slices.Equal,
				"SafeFields.ParamJSONMaps",
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
		ParamJSON:    map[string][]string{"actions": {"[].label"}},
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
			name:          "safe fields/unlisted censored and empty kept",
			safe:          safe,
			urlFields:     map[string]string{"host": "example.com", "token": "TOKEN", "port": ""},
			params:        map[string]string{"title": "TITLE", "key": "KEY"},
			wantURLFields: map[string]string{"host": "example.com", "token": util.SecretValue, "port": ""},
			wantParams:    map[string]string{"title": "TITLE", "key": util.SecretValue},
		},
		{
			name:          "safe fields/map censored by value only where listed",
			safe:          safe,
			urlFields:     map[string]string{"headers": `{"A":"b"}`, "query_vars": `{"c":"d"}`},
			params:        map[string]string{"details": `{"A":"b"}`, "headers": `{"A":"b"}`},
			wantURLFields: map[string]string{"headers": `{"A":"` + util.SecretValue + `"}`, "query_vars": util.SecretValue},
			wantParams:    map[string]string{"details": `{"A":"` + util.SecretValue + `"}`, "headers": util.SecretValue},
		},
		{
			name:          "safe fields/JSON censored bar safe paths",
			safe:          safe,
			params:        map[string]string{"actions": `[{"label":"A","url":"https://a"}]`},
			wantURLFields: map[string]string{},
			wantParams:    map[string]string{"actions": `[{"label":"A","url":"<secret>"}]`},
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
			want:  `{"Authorization":"` + util.SecretValue + `"}`,
		},
		{
			name:  "multiple keys/sorted",
			input: `{"X-Foo":"y","Authorization":"Bearer x"}`,
			want:  `{"Authorization":"` + util.SecretValue + `","X-Foo":"` + util.SecretValue + `"}`,
		},
		{
			name:  "empty value is masked",
			input: `{"X-Foo":""}`,
			want:  `{"X-Foo":"` + util.SecretValue + `"}`,
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

func TestCensorJSON(t *testing.T) {
	ntfyPaths := []string{"[].action", "[].label"}

	tests := []struct {
		name      string
		value     string
		safePaths []string
		want      string
	}{
		{
			name:  "empty",
			value: "",
			want:  "",
		},
		{
			name:  "invalid/not JSON",
			value: "view, Open, https://a",
			want:  util.SecretValue,
		},
		{
			name:      "array/safe paths shown, rest censored",
			value:     `[{"action":"http","clear":true,"headers":{"Authorization":"Bearer x"},"label":"A","url":"https://a"}]`,
			safePaths: ntfyPaths,
			want:      `[{"action":"http","clear":true,"headers":{"Authorization":"<secret>"},"label":"A","url":"<secret>"}]`,
		},
		{
			name:      "array/safe key name deeper than its path censored",
			value:     `[{"headers":{"label":"x"}}]`,
			safePaths: ntfyPaths,
			want:      `[{"headers":{"label":"<secret>"}}]`,
		},
		{
			name:      "object/nested safe path shown",
			value:     `{"client::display":{"contentType":"text/plain"},"client::notification":{"click":{"url":"https://a"}}}`,
			safePaths: []string{"client::display.contentType"},
			want:      `{"client::display":{"contentType":"text/plain"},"client::notification":{"click":{"url":"<secret>"}}}`,
		},
		{
			name:  "scalars/strings and numbers censored, bools and null kept",
			value: `{"a":"x","b":1,"c":false,"d":null}`,
			want:  `{"a":"<secret>","b":"<secret>","c":false,"d":null}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := CensorJSON(tc.value, tc.safePaths)

			if got != tc.want {
				t.Errorf(
					"CensorJSON(value=%q) mismatch\ngot:  %q\nwant: %q",
					tc.value, got, tc.want,
				)
			}
		})
	}
}

func TestRestoreJSONValues(t *testing.T) {
	tests := []struct {
		name            string
		original, value string
		want            string
	}{
		{
			name:     "no original",
			original: "",
			value:    `{"X-Foo":"` + util.SecretValue + `"}`,
			want:     `{"X-Foo":"` + util.SecretValue + `"}`,
		},
		{
			name:     "no original/value left as written",
			original: "",
			value:    `{ "X-Foo": "` + util.SecretValue + `" }`,
			want:     `{ "X-Foo": "` + util.SecretValue + `" }`,
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
			value:    `{"Authorization":"` + util.SecretValue + `","X-Foo":"` + util.SecretValue + `"}`,
			want:     `{"Authorization":"Bearer x","X-Foo":"y"}`,
		},
		{
			name:     "new value kept",
			original: `{"Authorization":"Bearer x"}`,
			value:    `{"Authorization":"` + util.SecretValue + `","X-Foo":"new"}`,
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
			value:    `{"Authorization":"` + util.SecretValue + `"}`,
			want:     `{"Authorization":"Bearer x"}`,
		},
		{
			name:     "renamed key does not take the old value",
			original: `{"X-Foo":"y"}`,
			value:    `{"X-Bar":"` + util.SecretValue + `"}`,
			want:     `{"X-Bar":"` + util.SecretValue + `"}`,
		},
		{
			name:     "renamed key/object of old_key and value kept as written",
			original: `{"a": {""value": "a"}}`,
			value:    `{"x": {"old_key": "a", "value": "b"}}`,
			want:     `{"x": {"old_key": "a", "value": "b"}}`,
		},
		{
			name:     "renamed key/no secret moves to another key",
			original: `[{"label":"Open","url":"https://a"}]`,
			value:    `[{"label":{"old_key":"url","value":"<secret>"},"old_index":0,"url":"<secret>"}]`,
			want:     `[{"label":{"old_key":"url","value":"<secret>"},"url":"https://a"}]`,
		},
		{
			name:     "renamed key/no secret moves to another path",
			original: `{"client::display":{"contentType":"text/plain"},"client::notification":{"click":{"url":"https://a"}}}`,
			value:    `{"client::display":{"old_key":"client::notification","value":{"contentType":{"old_key":"click","value":"<secret>"}}}}`,
			want:     `{"client::display":{"old_key":"client::notification","value":{"contentType":{"old_key":"click","value":"<secret>"}}}}`,
		},
		{
			name:     "masked key absent from original stays masked",
			original: `{"X-Foo":"y"}`,
			value:    `{"X-Foo":"` + util.SecretValue + `","X-Bar":"` + util.SecretValue + `"}`,
			want:     `{"X-Bar":"` + util.SecretValue + `","X-Foo":"y"}`,
		},
		{
			name:     "nothing masked/unchanged",
			original: `{"X-Foo":"y"}`,
			value:    `{ "X-Foo": "z" }`,
			want:     `{ "X-Foo": "z" }`,
		},
		{
			name:     "nested object/restored by key path",
			original: `{"client::notification":{"click":{"url":"https://a"}}}`,
			value:    `{"client::notification":{"click":{"url":"<secret>"}}}`,
			want:     `{"client::notification":{"click":{"url":"https://a"}}}`,
		},
		{
			name:     "array/restored by old index, which is dropped",
			original: `[{"label":"A","url":"https://a"},{"label":"B","url":"https://b"}]`,
			value:    `[{"label":"A","old_index":0,"url":"<secret>"},{"label":"B","old_index":1,"url":"<secret>"}]`,
			want:     `[{"label":"A","url":"https://a"},{"label":"B","url":"https://b"}]`,
		},
		{
			name:     "array/earlier element deleted",
			original: `[{"label":"A","url":"https://a"},{"label":"B","url":"https://b"}]`,
			value:    `[{"label":"B","old_index":1,"url":"<secret>"}]`,
			want:     `[{"label":"B","url":"https://b"}]`,
		},
		{
			name:     "array/reordered",
			original: `[{"label":"A","url":"https://a"},{"label":"B","url":"https://b"}]`,
			value:    `[{"label":"B","old_index":1,"url":"<secret>"},{"label":"A","old_index":0,"url":"<secret>"}]`,
			want:     `[{"label":"B","url":"https://b"},{"label":"A","url":"https://a"}]`,
		},
		{
			name:     "array/label changed, secret still restored",
			original: `[{"label":"A","url":"https://a"},{"label":"B","url":"https://b"}]`,
			value:    `[{"label":"Z","old_index":0,"url":"<secret>"}]`,
			want:     `[{"label":"Z","url":"https://a"}]`,
		},
		{
			name:     "array/no old index stays masked",
			original: `[{"label":"A","url":"https://a"},{"label":"B","url":"https://b"}]`,
			value:    `[{"label":"A","url":"<secret>"}]`,
			want:     `[{"label":"A","url":"<secret>"}]`,
		},
		{
			name:     "array/invalid old index stays masked",
			original: `[{"label":"A","url":"https://a"},{"label":"B","url":"https://b"}]`,
			value:    `[{"old_index":2,"url":"<secret>"},{"old_index":-1,"url":"<secret>"},{"old_index":0.5,"url":"<secret>"},{"old_index":"0","url":"<secret>"}]`,
			want:     `[{"url":"<secret>"},{"url":"<secret>"},{"url":"<secret>"},{"url":"<secret>"}]`,
		},
		{
			name:     "array/old index dropped without original",
			original: "",
			value:    `[{"label":"A","old_index":null}]`,
			want:     `[{"label":"A"}]`,
		},
		{
			name:     "numbers/restored as written",
			original: `{"a":1.50,"b":12345678901234567890}`,
			value:    `{"a":"<secret>","b":"<secret>"}`,
			want:     `{"a":1.50,"b":12345678901234567890}`,
		},
		{
			name:     "numbers/shown kept as written",
			original: `{"a":"x"}`,
			value:    `{"a":"<secret>","b":12345678901234567890}`,
			want:     `{"a":"x","b":12345678901234567890}`,
		},
		{
			name:     "invalid/value not JSON",
			original: `{"X-Foo":"y"}`,
			value:    `not json`,
			want:     `not json`,
		},
		{
			name:     "invalid/value with a masked value not JSON",
			original: `{"a":"x"}`,
			value:    `{"a":"<secret>","b":-}`,
			want:     `{"a":"<secret>","b":-}`,
		},
		{
			name:     "invalid/original not JSON",
			original: `not json`,
			value:    `{"X-Foo":"` + util.SecretValue + `"}`,
			want:     `{"X-Foo":"` + util.SecretValue + `"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := RestoreJSONValues(tc.original, tc.value)

			if got != tc.want {
				t.Errorf(
					"RestoreJSONValues(original=%q, value=%q) mismatch\ngot:  %q\nwant: %q",
					tc.original, tc.value, got, tc.want,
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
			name:     "no old key/restored as JSON values",
			original: `{"A":"a"}`,
			value:    `{"A":"<secret>"}`,
			want:     `{"A":"a"}`,
		},
		{
			name:     "renamed key/restored from its old key, which is dropped",
			original: `{"X-Foo":"y"}`,
			value:    `{"X-Bar":{"old_key":"X-Foo","value":"<secret>"}}`,
			want:     `{"X-Bar":"y"}`,
		},
		{
			name:     "renamed key/swapped keys each take their own old value",
			original: `{"A":"a","B":"b"}`,
			value:    `{"A":{"old_key":"B","value":"<secret>"},"B":{"old_key":"A","value":"<secret>"}}`,
			want:     `{"A":"b","B":"a"}`,
		},
		{
			name:     "renamed key/old key absent from original stays masked",
			original: `{"X-Foo":"y"}`,
			value:    `{"X-Bar":{"old_key":"X-Baz","value":"<secret>"}}`,
			want:     `{"X-Bar":"<secret>"}`,
		},
		{
			name:     "renamed key/changed value kept",
			original: `{"X-Foo":"y"}`,
			value:    `{"X-Bar":{"old_key":"X-Foo","value":"new"}}`,
			want:     `{"X-Bar":"new"}`,
		},
		{
			name:     "renamed key/old key dropped without original",
			original: "",
			value:    `{"X-Bar":{"old_key":"X-Foo","value":"<secret>"}}`,
			want:     `{"X-Bar":"<secret>"}`,
		},
		{
			name:     "renamed key/object with other keys is not a reference",
			original: `{"a":{"old_key":"b","value":"v","x":"y"},"b":"bb"}`,
			value:    `{"a":{"old_key":"b","value":"<secret>","x":"y"}}`,
			want:     `{"a":{"old_key":"b","value":"v","x":"y"}}`,
		},
		{
			name:     "renamed key/object without value is not a reference",
			original: `{"a":"x"}`,
			value:    `{"b":{"old_key":"a","x":"<secret>"}}`,
			want:     `{"b":{"old_key":"a","x":"<secret>"}}`,
		},
		{
			name:     "renamed key/non-string old key is not a reference",
			original: `{"a":"x"}`,
			value:    `{"b":{"old_key":1,"value":"<secret>"}}`,
			want:     `{"b":{"old_key":1,"value":"<secret>"}}`,
		},
		{
			name:     "renamed key/non-string value is not a reference",
			original: `{"a":{"c":"x"}}`,
			value:    `{"b":{"old_key":"a","value":{"c":"<secret>"}}}`,
			want:     `{"b":{"old_key":"a","value":{"c":"<secret>"}}}`,
		},
		{
			name:     "renamed key/onto a non-rename-shaped object restored from its old key",
			original: `{"a":"x","b":{"old_key":"c","value":{"d":"e"}}}`,
			value:    `{"b":{"old_key":"a","value":"<secret>"}}`,
			want:     `{"b":"x"}`,
		},
		{
			name:     "not an object/restored as JSON values",
			original: `["x"]`,
			value:    `["old_key","<secret>"]`,
			want:     `["old_key","<secret>"]`,
		},
		{
			name:     "invalid/value not JSON",
			original: `{"a":"x"}`,
			value:    `old_key: <secret>`,
			want:     `old_key: <secret>`,
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

func TestSafeFields_RestoreURLFields(t *testing.T) {
	tests := []struct {
		name                   string
		sType                  string
		original, target, want map[string]string
	}{
		{
			name:     "nil target",
			sType:    "generic",
			original: map[string]string{"headers": `{"A":"b"}`, "token": "TOKEN"},
			target:   nil,
			want:     nil,
		},
		{
			name:     "whole value restored",
			sType:    "generic",
			original: map[string]string{"token": "TOKEN"},
			target:   map[string]string{"token": util.SecretValue},
			want:     map[string]string{"token": "TOKEN"},
		},
		{
			name:     "any key restored",
			sType:    "generic",
			original: map[string]string{"made_up": "VALUE"},
			target:   map[string]string{"made_up": util.SecretValue},
			want:     map[string]string{"made_up": "VALUE"},
		},
		{
			name:     "map values restored by key",
			sType:    "generic",
			original: map[string]string{"headers": `{"A":"b","C":"d"}`},
			target:   map[string]string{"headers": `{"A":"` + util.SecretValue + `","E":"f"}`},
			want:     map[string]string{"headers": `{"A":"b","E":"f"}`},
		},
		{
			name:     "map field/renamed key restored from its old key",
			sType:    "generic",
			original: map[string]string{"headers": `{"authorization":"Basic x"}`},
			target:   map[string]string{"headers": `{"Authorization":{"old_key":"authorization","value":"<secret>"}}`},
			want:     map[string]string{"headers": `{"Authorization":"Basic x"}`},
		},
		{
			name:     "map field of another type/renamed key not restored",
			sType:    "ntfy", // ntfy has no URL field maps.
			original: map[string]string{"headers": `{"authorization":"Basic x"}`},
			target:   map[string]string{"headers": `{"Authorization":{"old_key":"authorization","value":"<secret>"}}`},
			want:     map[string]string{"headers": `{"Authorization":{"old_key":"authorization","value":"<secret>"}}`},
		},
		{
			name:     "no original value stays masked",
			sType:    "generic",
			original: map[string]string{},
			target:   map[string]string{"token": util.SecretValue},
			want:     map[string]string{"token": util.SecretValue},
		},
		{
			name:     "changed value kept",
			sType:    "generic",
			original: map[string]string{"token": "TOKEN"},
			target:   map[string]string{"token": "NEW"},
			want:     map[string]string{"token": "NEW"},
		},
		{
			name:     "field absent from target is not added",
			sType:    "generic",
			original: map[string]string{"query_vars": `{"c":"d"}`},
			target:   map[string]string{},
			want:     map[string]string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			SafeFieldsOf(tc.sType).RestoreURLFields(tc.original, tc.target)

			if testErr := test.AssertMapEqual(t, tc.target, tc.want, ""); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestSafeFields_RestoreParams(t *testing.T) {
	tests := []struct {
		name                   string
		sType                  string
		original, target, want map[string]string
	}{
		{
			name:     "map field/renamed key restored from its old key",
			sType:    "opsgenie",
			original: map[string]string{"details": `{"key":"x"}`},
			target:   map[string]string{"details": `{"Key":{"old_key":"key","value":"<secret>"}}`},
			want:     map[string]string{"details": `{"Key":"x"}`},
		},
		{
			name:     "URL field map/renamed key not restored",
			sType:    "generic", // generic headers are a URL field, not a param.
			original: map[string]string{"headers": `{"authorization":"Basic x"}`},
			target:   map[string]string{"headers": `{"Authorization":{"old_key":"authorization","value":"<secret>"}}`},
			want:     map[string]string{"headers": `{"Authorization":{"old_key":"authorization","value":"<secret>"}}`},
		},
		{
			name:     "JSON field/renamed key does not move a secret to a shown path",
			sType:    "ntfy",
			original: map[string]string{"actions": `[{"label":"Open","url":"https://a"}]`},
			target: map[string]string{
				"actions": `[{"label":{"old_key":"url","value":"<secret>"},"old_index":0,"url":"<secret>"}]`,
			},
			want: map[string]string{
				"actions": `[{"label":{"old_key":"url","value":"<secret>"},"url":"https://a"}]`,
			},
		},
		{
			name:  "JSON map/renamed header restored from its old key",
			sType: "ntfy",
			original: map[string]string{
				"actions": `[{"headers":{"authorization":"Basic x"},"label":"A"},{"label":"B"}]`,
			},
			target: map[string]string{
				"actions": test.TrimJSON(`[
					{"label":"B","old_index":1},
					{"headers":{"Authorization":{"old_key":"authorization","value":"<secret>"}},"label":"A","old_index":0}
				]`),
			},
			want: map[string]string{
				"actions": test.TrimJSON(`[
					{"label":"B"},
					{"headers":{"Authorization":"Basic x"},"label":"A"}
				]`),
			},
		},
		{
			name:     "JSON map/renamed extra restored from its old key",
			sType:    "ntfy",
			original: map[string]string{"actions": `[{"extras":{"a":"x"},"label":"A"}]`},
			target: map[string]string{
				"actions": test.TrimJSON(`[
					{
						"extras":{"b":{"old_key":"a","value":"<secret>"}},
						"label":"A",
						"old_index":0
					}
				]`),
			},
			want: map[string]string{"actions": `[{"extras":{"b":"x"},"label":"A"}]`},
		},
		{
			name:     "JSON map/renamed extra of an action without old_index stays masked",
			sType:    "ntfy",
			original: map[string]string{"actions": `[{"extras":{"a":"x"},"label":"A"}]`},
			target: map[string]string{
				"actions": test.TrimJSON(`[
					{
						"extras":{"b":{"old_key":"a","value":"<secret>"}},
						"label":"A"
					}
				]`),
			},
			want: map[string]string{"actions": `[{"extras":{"b":"<secret>"},"label":"A"}]`},
		},
		{
			name:     "JSON map/renamed top-level key restored from its old key",
			sType:    "gotify",
			original: map[string]string{"extras": `{"my::ns":"x"}`},
			target:   map[string]string{"extras": `{"your::ns":{"old_key":"my::ns","value":"<secret>"}}`},
			want:     map[string]string{"extras": `{"your::ns":"x"}`},
		},
		{
			name:     "JSON map/renamed number restored from its old key",
			sType:    "gotify",
			original: map[string]string{"extras": `{"n":5}`},
			target:   map[string]string{"extras": `{"m":{"old_key":"n","value":"<secret>"}}`},
			want:     map[string]string{"extras": `{"m":5}`},
		},
		{
			name:     "JSON map/old key absent from original stays masked",
			sType:    "gotify",
			original: map[string]string{"extras": `{"my::ns":"x","your::ns":"y"}`},
			target:   map[string]string{"extras": `{"your::ns":{"old_key":"other::ns","value":"<secret>"}}`},
			want:     map[string]string{"extras": `{"your::ns":"<secret>"}`},
		},
		{
			name:  "JSON map/ntfy UI payload restored",
			sType: "ntfy",
			original: map[string]string{
				"actions": test.TrimJSON(`[
					{"action":"http","headers":{"authorization":"Basic x"},"label":"A","method":"POST","url":"https://a"},
					{"action":"broadcast","extras":{"a":"x"},"label":"B"}]`),
			},
			target: map[string]string{
				"actions": test.TrimJSON(`[
					{
						"clear":false,"label":"A","old_index":0,"action":"http","body":"",
						"headers":{"Authorization":{"old_key":"authorization","value":"<secret>"}},"method":"post","url":"<secret>"
					},
					{
						"clear":false,"label":"B","old_index":1,"action":"broadcast",
						"extras":{"b":{"old_key":"a","value":"<secret>"}},"intent":""
					}
				]`),
			},
			want: map[string]string{
				"actions": test.TrimJSON(`[
					{
						"action":"http","body":"","clear":false,
						"headers":{"Authorization":"Basic x"},"label":"A","method":"post","url":"https://a"},
					{
						"action":"broadcast","clear":false,
						"extras":{"b":"x"},"intent":"","label":"B"
					}
				]`),
			},
		},
		{
			name:     "JSON map/gotify UI payload restored",
			sType:    "gotify",
			original: map[string]string{"extras": `{"client::display":{"contentType":"text/plain"},"my::ns":"x"}`},
			target: map[string]string{
				"extras": `{"your::ns":{"old_key":"my::ns","value":"<secret>"},"client::display":{"contentType":"text/plain"}}`,
			},
			want: map[string]string{"extras": `{"client::display":{"contentType":"text/plain"},"your::ns":"x"}`},
		},
		{
			name:     "JSON map/renamed key not moved to a shown path",
			sType:    "gotify",
			original: map[string]string{"extras": `{"my::ns":"x"}`},
			target: map[string]string{
				"extras": `{"client::display.contentType":{"old_key":"my::ns","value":"<secret>"}}`,
			},
			want: map[string]string{"extras": `{"client::display.contentType":"<secret>"}`},
		},
		{
			name:     "JSON map/renamed key not moved below a shown path - dotted key",
			sType:    "gotify",
			original: map[string]string{"extras": `{"my::ns":"x"}`},
			target: map[string]string{
				"extras": `{"client::display.contentType.x":{"old_key":"my::ns","value":"<secret>"}}`,
			},
			want: map[string]string{"extras": `{"client::display.contentType.x":"<secret>"}`},
		},
		{
			name:     "JSON map/renamed key not moved below a shown path - [] key",
			sType:    "gotify",
			original: map[string]string{"extras": `{"my::ns":"x"}`},
			target: map[string]string{
				"extras": `{"client::display.contentType[]":{"old_key":"my::ns","value":"<secret>"}}`,
			},
			want: map[string]string{"extras": `{"client::display.contentType[]":"<secret>"}`},
		},
		{
			name:  "JSON map/renamed object value not restored, as it could hold a shown path",
			sType: "gotify",
			original: map[string]string{
				"extras": `{"client::notification":{"click":{"url":"https://a"}}}`,
			},
			target: map[string]string{
				"extras": `{"client::display":{"old_key":"client::notification","value":"<secret>"}}`,
			},
			want: map[string]string{"extras": `{"client::display":"<secret>"}`},
		},
		{
			name:     "JSON map/nested renamed marker kept as written",
			sType:    "gotify",
			original: map[string]string{"extras": `{"client::notification":{"click":{"url":"https://a"}}}`},
			target: map[string]string{
				"extras": `{"client::display":{"contentType":{"old_key":"click","value":"<secret>"}}}`,
			},
			want: map[string]string{
				"extras": `{"client::display":{"contentType":{"old_key":"click","value":"<secret>"}}}`,
			},
		},
		{
			name:     "JSON map/censored value shaped like a marker restored as a value",
			sType:    "gotify",
			original: map[string]string{"extras": `{"my::ns":{"old_key":"a","value":"b"}}`},
			target:   map[string]string{"extras": `{"my::ns":{"old_key":"<secret>","value":"<secret>"}}`},
			want:     map[string]string{"extras": `{"my::ns":{"old_key":"a","value":"b"}}`},
		},
		{
			name:     "JSON map/censored value shaped like a marker restored as a value - number old_key",
			sType:    "gotify",
			original: map[string]string{"extras": `{"my::ns":{"old_key":5,"value":"b"}}`},
			target:   map[string]string{"extras": `{"my::ns":{"old_key":"<secret>","value":"<secret>"}}`},
			want:     map[string]string{"extras": `{"my::ns":{"old_key":5,"value":"b"}}`},
		},
		{
			name:     "JSON map/censored value shaped like a marker restored as a value - number value",
			sType:    "ntfy",
			original: map[string]string{"actions": `[{"extras":{"a":{"old_key":"b","value":5}},"label":"A"}]`},
			target: map[string]string{
				"actions": `[{"extras":{"a":{"old_key":"<secret>","value":"<secret>"}},"label":"A","old_index":0}]`,
			},
			want: map[string]string{"actions": `[{"extras":{"a":{"old_key":"b","value":5}},"label":"A"}]`},
		},
		{
			name:     "plain param/marker kept as written",
			sType:    "opsgenie",
			original: map[string]string{"actions": `["a"]`},
			target:   map[string]string{"actions": `[{"old_key":"a","value":"<secret>"}]`},
			want:     map[string]string{"actions": `[{"old_key":"a","value":"<secret>"}]`},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			SafeFieldsOf(tc.sType).RestoreParams(tc.original, tc.target)

			if testErr := test.AssertMapEqual(t, tc.target, tc.want, ""); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}
