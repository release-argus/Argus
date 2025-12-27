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
	"github.com/release-argus/Argus/webhook"
)

var packageName = "webhook_test"

func TestWebHook(t *testing.T) {
	// GIVEN the failing, self-signed certificate, and custom headers flags.
	tests := map[string]struct {
		failing, selfSignedCert, headers bool
		expectedURL, expectedSecret      string
		expectedHeaders                  *webhook.Headers
	}{
		"passing, signed, no custom headers": {
			failing:         false,
			selfSignedCert:  false,
			headers:         false,
			expectedURL:     test.LookupGitHub["url_valid"],
			expectedSecret:  test.LookupGitHub["secret_pass"],
			expectedHeaders: nil,
		},
		"passing, signed, with custom headers": {
			failing:        false,
			selfSignedCert: false,
			headers:        true,
			expectedURL:    test.LookupWithHeaderAuth["url_valid"],
			expectedSecret: test.LookupGitHub["secret_pass"],
			expectedHeaders: &webhook.Headers{
				{Key: test.LookupWithHeaderAuth["header_key"], Value: test.LookupWithHeaderAuth["header_value_pass"]}},
		},
		"passing, self-signed, no custom headers": {
			failing:         false,
			selfSignedCert:  true,
			headers:         false,
			expectedURL:     test.LookupGitHub["url_invalid"],
			expectedSecret:  test.LookupGitHub["secret_pass"],
			expectedHeaders: nil,
		},
		"passing, self-signed, with custom headers": {
			failing:        false,
			selfSignedCert: true,
			headers:        true,
			expectedURL:    test.LookupWithHeaderAuth["url_invalid"],
			expectedSecret: test.LookupGitHub["secret_pass"],
			expectedHeaders: &webhook.Headers{
				{Key: test.LookupWithHeaderAuth["header_key"], Value: test.LookupWithHeaderAuth["header_value_pass"]}},
		},
		"failing, signed, no custom headers": {
			failing:         true,
			selfSignedCert:  false,
			headers:         false,
			expectedURL:     test.LookupGitHub["url_valid"],
			expectedSecret:  test.LookupGitHub["secret_fail"],
			expectedHeaders: nil,
		},
		"failing, signed, with custom headers": {
			failing:        true,
			selfSignedCert: false,
			headers:        true,
			expectedURL:    test.LookupWithHeaderAuth["url_valid"],
			expectedSecret: test.LookupGitHub["secret_fail"],
			expectedHeaders: &webhook.Headers{
				{Key: test.LookupWithHeaderAuth["header_key"], Value: test.LookupWithHeaderAuth["header_value_fail"]}},
		},
		"failing, self-signed, no custom headers": {
			failing:         true,
			selfSignedCert:  true,
			headers:         false,
			expectedURL:     test.LookupGitHub["url_invalid"],
			expectedSecret:  test.LookupGitHub["secret_fail"],
			expectedHeaders: nil,
		},
		"failing, self-signed, with custom headers": {
			failing:        true,
			selfSignedCert: true,
			headers:        true,
			expectedURL:    test.LookupWithHeaderAuth["url_invalid"],
			expectedSecret: test.LookupGitHub["secret_fail"],
			expectedHeaders: &webhook.Headers{
				{Key: test.LookupWithHeaderAuth["header_key"], Value: test.LookupWithHeaderAuth["header_value_fail"]}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN WebHook is called.
			got := WebHook(
				tc.failing,
				tc.selfSignedCert,
				tc.headers)

			// THEN the URL should be as expected.
			if got.URL != tc.expectedURL {
				t.Errorf("%s\nURL mismatch\nwant: %q\ngot:  %q",
					packageName, tc.expectedURL, got.URL)
			}

			// AND the secret should be as expected.
			if got.Secret != tc.expectedSecret {
				t.Errorf("%s\nSecret mismatch\nwant: %q\ngot:  %q",
					packageName, tc.expectedSecret, got.Secret)
			}

			// AND the custom headers should be as expected.
			if tc.expectedHeaders == nil {
				if got.Headers != nil {
					t.Errorf("%s\nHeaders mismatch\nwant: nil\ngot:  %+v",
						packageName, got.Headers)
				}
			} else {
				if got.Headers == nil {
					t.Errorf("%s\nHeaders mismatch\nwant: %+v\ngot:  nil",
						packageName, tc.expectedHeaders)
				} else {
					// Lengths differ.
					if len(*got.Headers) != len(*tc.expectedHeaders) {
						t.Errorf("%s\nHeaders length mismatch\nwant: %d\ngot:  %d",
							packageName, len(*tc.expectedHeaders), len(*got.Headers))
					} else {
						// Check each header.
						for i := range *tc.expectedHeaders {
							if (*tc.expectedHeaders)[i].Key != (*got.Headers)[i].Key ||
								(*tc.expectedHeaders)[i].Value != (*got.Headers)[i].Value {
								t.Errorf("%s\nHeaders mismatch\nwant: %v (%+v)\ngot:  %v (%+v)",
									packageName,
									(*tc.expectedHeaders)[i], *tc.expectedHeaders,
									(*got.Headers)[i], *got.Headers)
								break
							}
						}
					}
				}
			}

			// AND the ID should be set.
			wantID := "test"
			if got.ID != wantID {
				t.Errorf("%s\nID mismatch\nwant: %q\ngot:  %q",
					packageName, wantID, got.ID)
			}

			// AND the ServiceStatus should be initialised.
			if got.ServiceStatus == nil {
				t.Errorf("%s\nServiceStatus not initialised",
					packageName)
			}

			// AND the Fails should be set.
			if got.ServiceStatus == nil || got.Failed == nil {
				if got.Failed == nil {
					t.Errorf("%s\nServiceStatus.Failed not set",
						packageName)
				} else {
					t.Errorf("%s\nServiceStatus not set",
						packageName)
				}
			}

			// AND the DesiredStatusCode should be set.
			wantDesiredStatusCode := 0
			if got.DesiredStatusCode == nil || *got.DesiredStatusCode != uint16(wantDesiredStatusCode) {
				if got.DesiredStatusCode == nil {
					t.Errorf("%s\nDesiredStatusCode not set",
						packageName)
				} else {
					t.Errorf("%s\nDesiredStatusCode mismatch\nwant: %d\ngot:  %d",
						packageName, wantDesiredStatusCode, *got.DesiredStatusCode)
				}
			}

			// AND the MaxTries should be set.
			wantMaxTries := 1
			if got.MaxTries == nil || *got.MaxTries != uint8(wantMaxTries) {
				if got.MaxTries == nil {
					t.Errorf("%s\nMaxTries not set",
						packageName)
				} else {
					t.Errorf("%s\nMaxTries mismatch\nwant: %d\ngot:  %d",
						packageName, wantMaxTries, *got.MaxTries)
				}
			}

			// AND the Delay should be set.
			wantDelay := "0s"
			if got.Delay != wantDelay {
				if got.Delay == "" {
					t.Errorf("%s\nDelay not set",
						packageName)
				} else {
					t.Errorf("%s\nDelay mismatch\nwant: %q\ngot:  %q",
						packageName, wantDelay, got.Delay)
				}
			}

			// AND the Main should be set.
			if got.Main == nil {
				t.Errorf("%s\nMain not set",
					packageName)
			}

			// AND the Defaults should be set.
			if got.Defaults == nil {
				t.Errorf("%s\nDefaults not set",
					packageName)
			}

			// AND the HardDefaults should be set.
			if got.HardDefaults == nil {
				t.Errorf("%s\nHardDefaults not set",
					packageName)
			}

			// AND the URL should be modified if selfSignedCert is true.
			if tc.selfSignedCert {
				if got.URL != tc.expectedURL {
					t.Errorf("%s\nSelfSignedCert: url mismatch\nwant: %q\ngot:  %q",
						packageName, tc.expectedURL, got.URL)
				}
			}

			// AND the Secret should be modified if failing is true.
			if tc.failing {
				expectedSecret := test.LookupGitHub["secret_fail"]
				if got.Secret != expectedSecret {
					t.Errorf("%s\nFailing webhook, secret mismatch\nwant: %q\ngot:  %q",
						packageName, expectedSecret, got.Secret)
				}
			}

			// AND the URL should be modified and headers should be set if headers is true.
			if tc.headers {
				if got.URL != tc.expectedURL {
					t.Errorf("%s\nHeaders, url mismatch\nwant: %q\ngot:  %q",
						packageName, tc.expectedURL, got.URL)
				}
			}
		})
	}
}
