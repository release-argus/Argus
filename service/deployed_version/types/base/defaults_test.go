// Copyright [2025] [Argus]
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
	"net/http"
	"testing"

	"github.com/release-argus/Argus/util"
)

func TestDefaults_Default(t *testing.T) {
	// GIVEN Defaults.
	defaults := Defaults{}

	// WHEN Default is called.
	defaults.Default()

	// THEN it should set the defaults.
	if defaults.AllowInvalidCerts == nil {
		t.Errorf("%s\nAllowInvalidCerts not set, got %v",
			packageName, defaults.AllowInvalidCerts)
	}
}

func TestDefaults_CheckValues(t *testing.T) {
	// GIVEN Defaults.
	tests := map[string]struct {
		method     string
		prefix     string
		wantErr    string
		wantMethod string
	}{
		"empty method - no error": {
			method:     "",
			wantErr:    `^$`,
			wantMethod: "",
		},
		"valid lowercase method - uppercased and ok": {
			method:     "post",
			wantErr:    `^$`,
			wantMethod: http.MethodPost,
		},
		"valid uppercase method - unchanged and ok": {
			method:     "GET",
			wantErr:    `^$`,
			wantMethod: http.MethodGet,
		},
		"unsupported method - with error prefix": {
			method:     http.MethodDelete,
			prefix:     "root: ",
			wantErr:    `^root: method: "` + http.MethodDelete + `" <invalid> .*` + http.MethodGet + `.*$`,
			wantMethod: http.MethodDelete,
		},
		"invalid method - no prefix": {
			method:     "foo",
			wantErr:    `^method: "FOO" <invalid> .*` + http.MethodPost + `.*$`,
			wantMethod: "FOO",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			d := &Defaults{Method: tc.method}

			// WHEN CheckValues is called.
			err := d.CheckValues(tc.prefix)

			// THEN the error matches expectation.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.wantErr, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wantErr, e)
			}

			// AND Method is uppercased/unchanged as expected.
			if d.Method != tc.wantMethod {
				t.Errorf("%s\nMethod mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wantMethod, d.Method)
			}
		})
	}
}
