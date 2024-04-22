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

	"github.com/release-argus/Argus/webhook"
)

func TestWebHook(t *testing.T) {
	// GIVEN the failing, self-signed certificate, and custom headers flags
	tests := map[string]struct {
		failing         bool
		selfSignedCert  bool
		customHeaders   bool
		expectedURL     string
		expectedSecret  string
		expectedHeaders *webhook.Headers
	}{
		"passing, signed, no custom headers": {
			failing:         false,
			selfSignedCert:  false,
			customHeaders:   false,
			expectedURL:     "https://valid.release-argus.io/hooks/github-style",
			expectedSecret:  "argus",
			expectedHeaders: nil,
		},
		"passing, signed, with custom headers": {
			failing:         false,
			selfSignedCert:  false,
			customHeaders:   true,
			expectedURL:     "https://valid.release-argus.io/hooks/single-header",
			expectedSecret:  "argus",
			expectedHeaders: &webhook.Headers{{Key: "X-Test", Value: "secret"}},
		},
		"passing, self-signed, no custom headers": {
			failing:         false,
			selfSignedCert:  true,
			customHeaders:   false,
			expectedURL:     "https://invalid.release-argus.io/hooks/github-style",
			expectedSecret:  "argus",
			expectedHeaders: nil,
		},
		"passing, self-signed, with custom headers": {
			failing:         false,
			selfSignedCert:  true,
			customHeaders:   true,
			expectedURL:     "https://invalid.release-argus.io/hooks/single-header",
			expectedSecret:  "argus",
			expectedHeaders: &webhook.Headers{{Key: "X-Test", Value: "secret"}},
		},
		"failing, signed, no custom headers": {
			failing:         true,
			selfSignedCert:  false,
			customHeaders:   false,
			expectedURL:     "https://valid.release-argus.io/hooks/github-style",
			expectedSecret:  "invalid",
			expectedHeaders: nil,
		},
		"failing, signed, with custom headers": {
			failing:         true,
			selfSignedCert:  false,
			customHeaders:   true,
			expectedURL:     "https://valid.release-argus.io/hooks/single-header",
			expectedSecret:  "invalid",
			expectedHeaders: &webhook.Headers{{Key: "X-Test", Value: "invalid"}},
		},
		"failing, self-signed, no custom headers": {
			failing:         true,
			selfSignedCert:  true,
			customHeaders:   false,
			expectedURL:     "https://invalid.release-argus.io/hooks/github-style",
			expectedSecret:  "invalid",
			expectedHeaders: nil,
		},
		"failing, self-signed, with custom headers": {
			failing:         true,
			selfSignedCert:  true,
			customHeaders:   true,
			expectedURL:     "https://invalid.release-argus.io/hooks/single-header",
			expectedSecret:  "invalid",
			expectedHeaders: &webhook.Headers{{Key: "X-Test", Value: "invalid"}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN WebHook is called
			got := WebHook(tc.failing, tc.selfSignedCert, tc.customHeaders)

			// THEN the URL should be as expected
			if got.URL != tc.expectedURL {
				t.Errorf("URL: expected %q but got %q",
					tc.expectedURL, got.URL)
			}

			// AND the secret should be as expected
			if got.Secret != tc.expectedSecret {
				t.Errorf("Secret: expected %q but got %q",
					tc.expectedSecret, got.Secret)
			}

			// AND the custom headers should be as expected
			if tc.expectedHeaders == nil {
				if got.CustomHeaders != nil {
					t.Errorf("CustomHeaders: expected nil but got %+v",
						got.CustomHeaders)
				}
			} else {
				if got.CustomHeaders == nil {
					t.Errorf("CustomHeaders: expected %+v but got nil",
						tc.expectedHeaders)
				} else {
					// Length differ
					if len(*got.CustomHeaders) != len(*tc.expectedHeaders) {
						t.Errorf("CustomHeaders: length differs, expected %+v but got %+v",
							tc.expectedHeaders, got.CustomHeaders)
					} else {
						// Check each header
						for i := range *tc.expectedHeaders {
							if (*tc.expectedHeaders)[i].Key != (*got.CustomHeaders)[i].Key ||
								(*tc.expectedHeaders)[i].Value != (*got.CustomHeaders)[i].Value {
								t.Errorf("CustomHeaders: expected %+v but got %+v",
									tc.expectedHeaders, got.CustomHeaders)
								break
							}
						}
					}
				}
			}

			// AND the ID should be set
			if got.ID != "test" {
				t.Errorf("ID: expected %q but got %q",
					"test", got.ID)
			}

			// AND the ServiceStatus should be initialized
			if got.ServiceStatus == nil {
				t.Error("ServiceStatus not initialized")
			}

			// AND the Fails should be set
			if got.ServiceStatus == nil || got.Failed == nil {
				if got.Failed == nil {
					t.Error("Failed not set")
				} else {
					t.Error("ServiceStatus not set")
				}
			}

			// AND the desiredStatusCode should be set
			if got.DesiredStatusCode == nil || *got.DesiredStatusCode != 0 {
				if got.DesiredStatusCode == nil {
					t.Error("DesiredStatusCode not set")
				} else {
					t.Errorf("DesiredStatusCode: expected %d but got %d",
						0, *got.DesiredStatusCode)
				}
			}

			// AND the MaxTries should be set
			if got.MaxTries == nil || *got.MaxTries != 1 {
				if got.MaxTries == nil {
					t.Error("MaxTries not set")
				} else {
					t.Errorf("MaxTries: expected %d but got %d",
						1, *got.MaxTries)
				}
			}

			// AND the Delay should be set
			wantDelay := "0s"
			if got.Delay == "" || got.Delay != wantDelay {
				if got.Delay == "" {
					t.Error("Delay not set")
				} else {
					t.Errorf("Delay:expected %q but got %q",
						wantDelay, got.Delay)
				}
			}

			// AND the Main should be set
			if got.Main == nil {
				t.Error("Main not set")
			}

			// AND the Defaults should be set
			if got.Defaults == nil {
				t.Error("Defaults not set")
			}

			// AND the HardDefaults should be set
			if got.HardDefaults == nil {
				t.Error("HardDefaults not set")
			}

			// AND the URL should be modified if selfSignedCert is true
			if tc.selfSignedCert {
				if got.URL != tc.expectedURL {
					t.Errorf("SelfSignedCert: url, expected %q but got %q",
						tc.expectedURL, got.URL)
				}
			}

			// AND the Secret should be modified if failing is true
			if tc.failing {
				expectedSecret := "invalid"
				if got.Secret != expectedSecret {
					t.Errorf("Failing: url, expected %q but got %q",
						expectedSecret, got.Secret)
				}
			}

			// AND the URL should be modified and custom headers should be set if customHeaders is true
			if tc.customHeaders {
				if got.URL != tc.expectedURL {
					t.Errorf("CustomHeaders: url, expected %q but got %q",
						tc.expectedURL, got.URL)
				}
			}
		})
	}
}
