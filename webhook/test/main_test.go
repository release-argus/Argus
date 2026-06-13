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
	"github.com/release-argus/Argus/webhook"
)

var packageName = "whtest"

func TestWebHook(t *testing.T) {
	// GIVEN: the failing, self-signed certificate, and custom headers flags.
	tests := []struct {
		name                             string
		failing, selfSignedCert, headers bool
		expectedURL, expectedSecret      string
		expectedHeaders                  *webhook.Headers
	}{
		{
			name:            "passing, signed, no custom headers",
			failing:         false,
			selfSignedCert:  false,
			headers:         false,
			expectedURL:     test.WebHookGitHub["url_valid"],
			expectedSecret:  test.WebHookGitHub["secret_pass"],
			expectedHeaders: nil,
		},
		{
			name:           "passing, signed, with custom headers",
			failing:        false,
			selfSignedCert: false,
			headers:        true,
			expectedURL:    test.LookupWithHeaderAuth["url_valid"],
			expectedSecret: test.WebHookGitHub["secret_pass"],
			expectedHeaders: &webhook.Headers{
				{
					Key:   test.LookupWithHeaderAuth["header_key"],
					Value: test.LookupWithHeaderAuth["header_value_pass"],
				},
			},
		},
		{
			name:            "passing, self-signed, no custom headers",
			failing:         false,
			selfSignedCert:  true,
			headers:         false,
			expectedURL:     test.WebHookGitHub["url_invalid"],
			expectedSecret:  test.WebHookGitHub["secret_pass"],
			expectedHeaders: nil,
		},
		{
			name:           "passing, self-signed, with custom headers",
			failing:        false,
			selfSignedCert: true,
			headers:        true,
			expectedURL:    test.LookupWithHeaderAuth["url_invalid"],
			expectedSecret: test.WebHookGitHub["secret_pass"],
			expectedHeaders: &webhook.Headers{
				{
					Key:   test.LookupWithHeaderAuth["header_key"],
					Value: test.LookupWithHeaderAuth["header_value_pass"],
				},
			},
		},
		{
			name:            "failing, signed, no custom headers",
			failing:         true,
			selfSignedCert:  false,
			headers:         false,
			expectedURL:     test.WebHookGitHub["url_valid"],
			expectedSecret:  test.WebHookGitHub["secret_fail"],
			expectedHeaders: nil,
		},
		{
			name:           "failing, signed, with custom headers",
			failing:        true,
			selfSignedCert: false,
			headers:        true,
			expectedURL:    test.LookupWithHeaderAuth["url_valid"],
			expectedSecret: test.WebHookGitHub["secret_fail"],
			expectedHeaders: &webhook.Headers{
				{
					Key:   test.LookupWithHeaderAuth["header_key"],
					Value: test.LookupWithHeaderAuth["header_value_fail"],
				},
			},
		},
		{
			name:            "failing, self-signed, no custom headers",
			failing:         true,
			selfSignedCert:  true,
			headers:         false,
			expectedURL:     test.WebHookGitHub["url_invalid"],
			expectedSecret:  test.WebHookGitHub["secret_fail"],
			expectedHeaders: nil,
		},
		{
			name:           "failing, self-signed, with custom headers",
			failing:        true,
			selfSignedCert: true,
			headers:        true,
			expectedURL:    test.LookupWithHeaderAuth["url_invalid"],
			expectedSecret: test.WebHookGitHub["secret_fail"],
			expectedHeaders: &webhook.Headers{
				{
					Key:   test.LookupWithHeaderAuth["header_key"],
					Value: test.LookupWithHeaderAuth["header_value_fail"],
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: WebHook is called.
			result := WebHook(
				tc.failing,
				tc.selfSignedCert,
				tc.headers,
			)

			prefix := fmt.Sprintf(
				"%s\nWebHook(failed=%t, selfSigned=%t, headers=%t)",
				packageName, tc.failing, tc.selfSignedCert, tc.headers,
			)

			// THEN: the URL should be as expected.
			if result.URL != tc.expectedURL {
				t.Errorf(
					"%s URL mismatch\ngot:  %q\nwant: %q",
					prefix, result.URL, tc.expectedURL,
				)
			}

			// AND: the secret should be as expected.
			if result.Secret != tc.expectedSecret {
				t.Errorf(
					"%s Secret mismatch\ngot:  %q\nwant: %q",
					prefix, result.Secret, tc.expectedSecret,
				)
			}

			// AND: the custom headers should be as expected.
			if tc.expectedHeaders == nil {
				if result.Headers != nil {
					t.Errorf(
						"%s Headers mismatch\ngot:  %+v\nwant: nil",
						prefix, result.Headers,
					)
				}
			} else {
				if result.Headers == nil {
					t.Errorf(
						"%s Headers mismatch\ngot:  nil\nwant: %+v",
						prefix, tc.expectedHeaders,
					)
				} else {
					// Lengths differ.
					if gotLen, wantLen := len(result.Headers), len(*tc.expectedHeaders); gotLen != wantLen {
						t.Errorf(
							"%s Headers length mismatch\ngot:  %d\nwant: %d",
							prefix, gotLen, wantLen,
						)
					} else {
						// Check each header.
						for i := range *tc.expectedHeaders {
							got := (result.Headers)[i]
							want := (*tc.expectedHeaders)[i]
							if got.Key != want.Key || got.Value != want.Value {
								t.Errorf(
									"%s Headers mismatch\ngot:  %v (%+v)\nwant: %v (%+v)",
									prefix,
									got, result.Headers,
									want, *tc.expectedHeaders,
								)
								break
							}
						}
					}
				}
			}

			// AND: the ID should be set.
			wantID := "test"
			if result.ID != wantID {
				t.Errorf(
					"%s .ID mismatch\ngot:  %q\nwant: %q",
					prefix, result.ID, wantID,
				)
			}

			// AND: the ServiceStatus should be initialised.
			if result.ServiceStatus == nil {
				t.Errorf("%s ServiceStatus not initialised", prefix)
			}

			// AND: the Fails should be set.
			if result.ServiceStatus == nil || result.Failed == nil {
				if result.Failed == nil {
					t.Errorf("%s ServiceStatus.Failed not set", prefix)
				} else {
					t.Errorf("%s ServiceStatus not set", prefix)
				}
			}

			// AND: the DesiredStatusCode should be set.
			wantDesiredStatusCode := 0
			if result.DesiredStatusCode == nil || *result.DesiredStatusCode != uint16(wantDesiredStatusCode) {
				if result.DesiredStatusCode == nil {
					t.Errorf(
						"%s DesiredStatusCode not set", prefix,
					)
				} else {
					t.Errorf(
						"%s DesiredStatusCode mismatch\ngot:  %d\nwant: %d",
						prefix, *result.DesiredStatusCode, wantDesiredStatusCode,
					)
				}
			}

			// AND: the MaxTries should be set.
			wantMaxTries := 1
			if result.MaxTries == nil || *result.MaxTries != uint8(wantMaxTries) {
				if result.MaxTries == nil {
					t.Errorf("%s MaxTries not set", prefix)
				} else {
					t.Errorf(
						"%s MaxTries mismatch\ngot:  %d\nwant: %d",
						prefix, *result.MaxTries, wantMaxTries,
					)
				}
			}

			// AND: the Delay should be set.
			wantDelay := "0s"
			if result.Delay != wantDelay {
				if result.Delay == "" {
					t.Errorf("%s Delay not set", prefix)
				} else {
					t.Errorf(
						"%s Delay mismatch\ngot:  %q\nwant: %q",
						prefix, result.Delay, wantDelay,
					)
				}
			}

			// AND: the Main should be set.
			if result.Main == nil {
				t.Errorf("%s Main not set", prefix)
			}

			// AND: the Defaults should be set.
			if result.Defaults == nil {
				t.Errorf("%s Defaults not set", prefix)
			}

			// AND: the HardDefaults should be set.
			if result.HardDefaults == nil {
				t.Errorf("%s HardDefaults not set", prefix)
			}

			// AND: the URL should be modified if selfSignedCert is true.
			if tc.selfSignedCert {
				if result.URL != tc.expectedURL {
					t.Errorf(
						"%s URL mismatch when SelfSignedCert=true\ngot:  %q\nwant: %q",
						prefix, result.URL, tc.expectedURL,
					)
				}
			}

			// AND: the Secret should be modified if failing is true.
			if tc.failing {
				expectedSecret := test.WebHookGitHub["secret_fail"]
				if result.Secret != expectedSecret {
					t.Errorf(
						"%s Secret mismatch for failing WebHook\ngot:  %q\nwant: %q",
						prefix, result.Secret, expectedSecret,
					)
				}
			}

			// AND: the URL should be modified and headers should be set if headers is true.
			if tc.headers {
				if result.URL != tc.expectedURL {
					t.Errorf(
						"%s URL mismatch for WebHook with Headers\ngot:  %q\nwant: %q",
						packageName, result.URL, tc.expectedURL,
					)
				}
			}
		})
	}
}
