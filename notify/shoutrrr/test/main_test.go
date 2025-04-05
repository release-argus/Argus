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

//go:build unit || integration

package test

import (
	"testing"

	"github.com/release-argus/Argus/test"
)

var packageName = "shoutrrr_test"

func TestDefaults(t *testing.T) {
	// GIVEN the failing and self-signed certificate flags.
	tests := map[string]struct {
		failing, selfSignedCert bool
	}{
		"passing, signed":      {failing: false, selfSignedCert: false},
		"passing, self-signed": {failing: false, selfSignedCert: true},
		"failing, signed":      {failing: true, selfSignedCert: false},
		"failing, self-signed": {failing: true, selfSignedCert: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wantToken := test.ShoutrrrGotifyToken()
			if tc.failing {
				wantToken = "invalid"
			}
			wantHost := test.ValidCertNoProtocol
			if tc.selfSignedCert {
				wantHost = test.InvalidCertNoProtocol
			}

			// WHEN Defaults is called.
			got := Defaults(tc.failing, tc.selfSignedCert)

			// THEN the token should be as expected.
			key := "token"
			if got.URLFields[key] != wantToken {
				t.Errorf("%s\nmismatch on url_fields[%q]\nwant: %q\ngot:  %q",
					packageName, key, wantToken, got.URLFields["token"])
			}
			// AND the host should be as expected.
			key = "host"
			if got.URLFields[key] != wantHost {
				t.Errorf("%s\nmismatch on url_fields[%q]\nwant: %q\ngot:  %q",
					packageName, key, wantHost, got.URLFields["host"])
			}
		})
	}
}

func TestShoutrrr(t *testing.T) {
	// GIVEN the failing and self-signed certificate flags.
	tests := map[string]struct {
		failing, selfSignedCert bool
	}{
		"passing, signed":      {failing: false, selfSignedCert: false},
		"passing, self-signed": {failing: false, selfSignedCert: true},
		"failing, signed":      {failing: true, selfSignedCert: false},
		"failing, self-signed": {failing: true, selfSignedCert: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wantToken := test.ShoutrrrGotifyToken()
			if tc.failing {
				wantToken = "invalid"
			}
			wantHost := test.ValidCertNoProtocol
			if tc.selfSignedCert {
				wantHost = test.InvalidCertNoProtocol
			}

			// WHEN Shoutrrr is called.
			got := Shoutrrr(tc.failing, tc.selfSignedCert)

			// THEN the token should be as expected.
			key := "token"
			if got.URLFields[key] != wantToken {
				t.Errorf("%s\nmismatch on url_fields[%q]\nwant: %q\ngot:  %q",
					packageName, key, wantToken, got.URLFields["token"])
			}
			// AND the host should be as expected.
			key = "host"
			if got.URLFields[key] != wantHost {
				t.Errorf("%s\nmismatch on url_fields[%q]\nwant: %q\ngot:  %q",
					packageName, key, wantHost, got.URLFields["host"])
			}
			// AND the maps should be initialised.
			if got.Options == nil {
				t.Errorf("%s\nOptions not initialised",
					packageName)
			}
			if got.URLFields == nil {
				t.Errorf("%s\nURLFields not initialised",
					packageName)
			}
			if got.Params == nil {
				t.Errorf("%s\nParams not initialised",
					packageName)
			}
			// AND the defaults should be set.
			if got.Main == nil {
				t.Errorf("%s\nMain not set",
					packageName)
			}
			if got.Defaults == nil {
				t.Errorf("%s\nDefaults not set",
					packageName)
			}
			if got.HardDefaults == nil {
				t.Errorf("%s\nHardDefaults not set",
					packageName)
			}
			// AND the fails are initialised and set.
			if got.ServiceStatus == nil || got.Failed == nil {
				if got.ServiceStatus == nil {
					t.Errorf("%s\nServiceStatus not set",
						packageName)
				} else {
					t.Errorf("%s\nServiceStatus.Failed not set",
						packageName)
				}
			}
		})
	}
}
