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
	"strconv"
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func TestLookup_GetType(t *testing.T) {
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

			got := l.GetType()

			wantType := "url"

			if got != wantType {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, wantType)
			}
		})
	}
}

func TestLookup_AllowInvalidCerts(t *testing.T) {
	tests := []struct {
		name                                      string
		rootValue, defaultValue, hardDefaultValue *bool
		want                                      bool
	}{
		{
			name:             "true/root overrides all",
			want:             true,
			rootValue:        new(true),
			defaultValue:     new(false),
			hardDefaultValue: new(false),
		},
		{
			name:             "true/default overrides hardDefault",
			want:             true,
			defaultValue:     new(true),
			hardDefaultValue: new(false),
		},
		{
			name:             "true/hardDefault is last resort",
			want:             true,
			hardDefaultValue: new(true),
		},
		{
			name:             "false/root overrides all",
			want:             false,
			rootValue:        new(false),
			defaultValue:     new(true),
			hardDefaultValue: new(true),
		},
		{
			name:             "false/default overrides hardDefault",
			want:             false,
			defaultValue:     new(false),
			hardDefaultValue: new(true),
		},
		{
			name:             "false/hardDefault is last resort",
			want:             false,
			hardDefaultValue: new(false),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t)
			lookup.AllowInvalidCerts = tc.rootValue
			lookup.Defaults.AllowInvalidCerts = tc.defaultValue
			lookup.HardDefaults.AllowInvalidCerts = tc.hardDefaultValue

			got := lookup.allowInvalidCerts()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestLookup_Body(t *testing.T) {
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

			lookup := testLookup(t)
			lookup.Body = tc.body

			got := lookup.body()

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("mismatch\ngot:  %s\nwant: %s", readerString(got), readerString(tc.want))
			}
		})
	}
}

func TestLookup_Method(t *testing.T) {
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

			lookup := testLookup(t)
			lookup.Method = tc.rootValue
			lookup.Defaults.Method = tc.defaultValue
			lookup.HardDefaults.Method = tc.hardDefaultValue

			got := lookup.method()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestLookup_URL(t *testing.T) {
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

			test.IsolateEnv(t, tc.env)

			lookup := testLookup(t)
			lookup.URL = tc.url

			got := lookup.url()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

// readerString returns the quoted contents of r, or nil when r is nil.
func readerString(r io.Reader) string {
	if r == nil {
		return "nil"
	}
	b, _ := io.ReadAll(r)
	return strconv.Quote(string(b))
}
