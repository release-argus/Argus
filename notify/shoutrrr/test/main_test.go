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
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func TestShoutrrr(t *testing.T) {
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

			wantToken := test.NotifyGotify.TokenPass
			if tc.failing {
				wantToken = test.NotifyGotify.TokenMalformed
			}
			wantHost := test.NotifyGotify.HostValid
			if tc.selfSignedCert {
				wantHost = test.NotifyGotify.HostInvalid
			}

			got := Shoutrrr(tc.failing, tc.selfSignedCert)

			fieldTests := []test.FieldAssertion{
				{Name: "URLFields", Got: got.URLFields, Mode: test.CompareNotNil},
				{Name: `URLFields["token"]`, Got: got.URLFields["token"], Want: wantToken, Mode: test.CompareEqual},
				{Name: `URLFields["host"]`, Got: got.URLFields["host"], Want: wantHost, Mode: test.CompareEqual},
				{Name: "Options", Got: got.Options, Mode: test.CompareNotNil},
				{Name: "Params", Got: got.Params, Mode: test.CompareNotNil},
				{Name: "Main", Got: got.Main, Mode: test.CompareNotNil},
				{Name: "Defaults", Got: got.Defaults, Mode: test.CompareNotNil},
				{Name: "HardDefaults", Got: got.HardDefaults, Mode: test.CompareNotNil},
				{Name: "ServiceStatus", Got: got.ServiceStatus, Mode: test.CompareNotNil},
				{Name: "Failed", Got: got.Failed, Mode: test.CompareNotNil},
			}
			if testErr := test.AssertFields(t, fieldTests, "Shoutrrr"); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}
