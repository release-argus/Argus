// Copyright [2024] [Argus]
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

func TestShoutrrrDefaults(t *testing.T) {
	// GIVEN the failing and self-signed certificate flags
	tests := map[string]struct {
		failing        bool
		selfSignedCert bool
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
			wantHost := "valid.release-argus.io"
			if tc.selfSignedCert {
				wantHost = "invalid.release-argus.io"
			}

			// WHEN ShoutrrrDefaults is called
			got := ShoutrrrDefaults(tc.failing, tc.selfSignedCert)

			// THEN the token should be as expected
			if got.URLFields["token"] != wantToken {
				t.Errorf("expected %q but got %q",
					wantToken, got.URLFields["token"])
			}
			// AND the host should be as expected
			if wantHost != got.URLFields["host"] {
				t.Errorf("expected %q but got %q",
					wantHost, got.URLFields["host"])
			}
		})
	}
}

func TestShoutrrr(t *testing.T) {
	// GIVEN the failing and self-signed certificate flags
	tests := map[string]struct {
		failing        bool
		selfSignedCert bool
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
			wantHost := "valid.release-argus.io"
			if tc.selfSignedCert {
				wantHost = "invalid.release-argus.io"
			}

			// WHEN Shoutrrr is called
			got := Shoutrrr(tc.failing, tc.selfSignedCert)

			// THEN the token should be as expected
			if wantToken != got.URLFields["token"] {
				t.Errorf("expected %q but got %q",
					wantToken, got.URLFields["token"])
			}
			// AND the host should be as expected
			if wantHost != got.URLFields["host"] {
				t.Errorf("expected %q but got %q",
					wantHost, got.URLFields["host"])
			}
			// AND the maps should be initialised
			if got.Options == nil {
				t.Error("Options map not initialised")
			}
			if got.URLFields == nil {
				t.Error("URLFields map not initialised")
			}
			if got.Params == nil {
				t.Error("Params map not initialised")
			}
			// AND the defaults should be set
			if got.Main == nil {
				t.Error("Main not set")
			}
			if got.Defaults == nil {
				t.Error("Defaults not set")
			}
			if got.HardDefaults == nil {
				t.Error("HardDefaults not set")
			}
			// AND the fails are initialised and set
			if got.ServiceStatus == nil || got.Failed == nil {
				if got.ServiceStatus == nil {
					t.Error("ServiceStatus not set")
				} else {
					t.Error("Failed not set")
				}
			}
		})
	}
}
