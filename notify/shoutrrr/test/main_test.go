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

//go:build unit || integration

package test

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

var packageName = "shoutrrrtest"

func TestShoutrrr(t *testing.T) {
	// GIVEN: the failing and self-signed certificate flags.
	tests := []struct {
		name                    string
		failing, selfSignedCert bool
	}{
		{
			name:           "passing, signed",
			failing:        false,
			selfSignedCert: false,
		},
		{
			name:           "passing, self-signed",
			failing:        false,
			selfSignedCert: true,
		},
		{
			name:           "failing, signed",
			failing:        true,
			selfSignedCert: false,
		},
		{
			name:           "failing, self-signed",
			failing:        true,
			selfSignedCert: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			wantToken := test.ShoutrrrGotifyToken()
			if tc.failing {
				wantToken = "invalid"
			}
			wantHost := test.ValidCertNoProtocol
			if tc.selfSignedCert {
				wantHost = test.InvalidCertNoProtocol
			}

			// WHEN: Shoutrrr is called.
			got := Shoutrrr(t, tc.failing, tc.selfSignedCert)

			prefix := fmt.Sprintf(
				"%s\nTestShoutrrr(failing: %t, selfSigned: %t)",
				packageName, tc.failing, tc.selfSignedCert,
			)

			// THEN: the token should be as expected.
			key := "token"
			if got.URLFields[key] != wantToken {
				t.Errorf(
					"%s got a mismatch on url_fields[%q]\ngot:  %q\nwant: %q",
					prefix, key,
					got.URLFields["token"], wantToken,
				)
			}

			// AND: the host should be as expected.
			key = "host"
			if got.URLFields[key] != wantHost {
				t.Errorf(
					"%s got a mismatch on url_fields[%q]\ngot:  %q\nwant: %q",
					prefix, key,
					got.URLFields["host"], wantHost,
				)
			}

			// AND: the maps should be initialised.
			if got.Options == nil {
				t.Errorf("%s Options not initialised. got nil", prefix)
			}
			if got.URLFields == nil {
				t.Errorf("%s URLFields not initialised. got nil", prefix)
			}
			if got.Params == nil {
				t.Errorf("%s Params not initialised. got nil", prefix)
			}

			// AND: the defaults should be set.
			if got.Main == nil {
				t.Errorf("%s Main not set", prefix)
			}
			if got.Defaults == nil {
				t.Errorf("%s Defaults not set", prefix)
			}
			if got.HardDefaults == nil {
				t.Errorf("%s HardDefaults not set", prefix)
			}

			// AND: the fails are initialised and set.
			if got.ServiceStatus == nil || got.Failed == nil {
				if got.ServiceStatus == nil {
					t.Errorf("%s ServiceStatus not set", prefix)
				} else {
					t.Errorf("%s ServiceStatus.Failed not set", prefix)
				}
			}
		})
	}
}
