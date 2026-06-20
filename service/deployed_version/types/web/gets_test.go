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

package web

import (
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func TestLookup_GetType(t *testing.T) {
	// GIVEN: a Lookup with a Type.
	tests := []struct {
		name  string
		lType string
	}{
		{name: "empty", lType: ""},
		{name: "test", lType: "test"},
		{name: "x", lType: "x"},
		{name: "y", lType: "y"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l := &Lookup{}
			l.Type = tc.lType

			// WHEN: GetType is called.
			got := l.GetType()

			wantType := "url"
			// THEN: the Type is returned.
			if got != wantType {
				t.Errorf(
					"%s\nLookup.GetType() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, wantType,
				)
			}
		})
	}
}

func TestLookup_AllowInvalidCerts(t *testing.T) {
	// GIVEN: a Lookup.
	tests := []struct {
		name                                      string
		rootValue, defaultValue, hardDefaultValue *bool
		want                                      bool
	}{
		{
			name:             "root overrides all",
			want:             true,
			rootValue:        test.Ptr(true),
			defaultValue:     test.Ptr(false),
			hardDefaultValue: test.Ptr(false),
		},
		{
			name:             "default overrides hardDefault",
			want:             true,
			defaultValue:     test.Ptr(true),
			hardDefaultValue: test.Ptr(false),
		},
		{
			name:             "hardDefault is last resort",
			want:             true,
			hardDefaultValue: test.Ptr(true),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, false)
			lookup.AllowInvalidCerts = tc.rootValue
			lookup.Defaults.AllowInvalidCerts = tc.defaultValue
			lookup.HardDefaults.AllowInvalidCerts = tc.hardDefaultValue

			// WHEN: allowInvalidCerts is called.
			got := lookup.allowInvalidCerts()

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Errorf(
					"%s\nLookup.allowInvalidCerts() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestLookup_Body(t *testing.T) {
	// GIVEN: a Lookup.
	tests := []struct {
		name string
		body string
		want io.Reader
	}{
		{
			name: "empty body",
			body: "",
			want: nil,
		},
		{
			name: "non-empty body",
			body: "test body",
			want: strings.NewReader("test body"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, false)
			lookup.Body = tc.body

			// WHEN: body is called.
			got := lookup.body()

			// THEN: the function returns the correct result.
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf(
					"%s\nLookup.body() value mismatch\ngot:  %v\nwant: %v",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestLookup_Method(t *testing.T) {
	// GIVEN: a Lookup.
	tests := []struct {
		name                                      string
		rootValue, defaultValue, hardDefaultValue string
		want                                      string
	}{
		{
			name:             "root overrides all",
			want:             http.MethodGet,
			rootValue:        http.MethodGet,
			defaultValue:     http.MethodPost,
			hardDefaultValue: http.MethodPost,
		},
		{
			name:             "default overrides hardDefault",
			want:             http.MethodPost,
			defaultValue:     http.MethodPost,
			hardDefaultValue: http.MethodGet,
		},
		{
			name:             "hardDefault is last resort",
			want:             http.MethodGet,
			hardDefaultValue: http.MethodGet,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, false)
			lookup.Method = tc.rootValue
			lookup.Defaults.Method = tc.defaultValue
			lookup.HardDefaults.Method = tc.hardDefaultValue

			// WHEN: method is called.
			got := lookup.method()

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Errorf(
					"%s\nLookup.method() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestLookup_URL(t *testing.T) {
	// GIVEN: a Lookup.
	tests := []struct {
		name string
		env  map[string]string
		url  string
		want string
	}{
		{
			name: "URL/plain",
			url:  "https://example.com",
			want: "https://example.com",
		},
		{
			name: "URL/from env",
			env: map[string]string{
				"TEST_LOOKUP__DV_GET_URL_ONE": "https://example.com",
			},
			url:  "${TEST_LOOKUP__DV_GET_URL_ONE}",
			want: "https://example.com",
		},
		{
			name: "URL/with env partial",
			env: map[string]string{
				"TEST_LOOKUP__DV_GET_URL_TWO": "example.com",
			},
			url:  "https://${TEST_LOOKUP__DV_GET_URL_TWO}",
			want: "https://example.com",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.SetEnv(t, tc.env)

			lookup := testLookup(t, false)
			lookup.URL = tc.url

			// WHEN: url is called.
			got := lookup.url()

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Errorf(
					"%s\nLookup.url() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}
