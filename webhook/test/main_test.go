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
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

func TestWebhook(t *testing.T) {
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
			expectedURL:     test.WebhookGitHub.URLValid,
			expectedSecret:  test.WebhookGitHub.SecretPass,
			expectedHeaders: nil,
		},
		{
			name:           "passing, signed, with custom headers",
			failing:        false,
			selfSignedCert: false,
			headers:        true,
			expectedURL:    test.LookupWithHeaderAuth.URLValid,
			expectedSecret: test.WebhookGitHub.SecretPass,
			expectedHeaders: &webhook.Headers{
				{
					Key:   test.LookupWithHeaderAuth.HeaderKey,
					Value: test.LookupWithHeaderAuth.HeaderValuePass,
				},
			},
		},
		{
			name:            "passing, self-signed, no custom headers",
			failing:         false,
			selfSignedCert:  true,
			headers:         false,
			expectedURL:     test.WebhookGitHub.URLInvalid,
			expectedSecret:  test.WebhookGitHub.SecretPass,
			expectedHeaders: nil,
		},
		{
			name:           "passing, self-signed, with custom headers",
			failing:        false,
			selfSignedCert: true,
			headers:        true,
			expectedURL:    test.LookupWithHeaderAuth.URLInvalid,
			expectedSecret: test.WebhookGitHub.SecretPass,
			expectedHeaders: &webhook.Headers{
				{
					Key:   test.LookupWithHeaderAuth.HeaderKey,
					Value: test.LookupWithHeaderAuth.HeaderValuePass,
				},
			},
		},
		{
			name:            "failing, signed, no custom headers",
			failing:         true,
			selfSignedCert:  false,
			headers:         false,
			expectedURL:     test.WebhookGitHub.URLValid,
			expectedSecret:  test.WebhookGitHub.SecretFail,
			expectedHeaders: nil,
		},
		{
			name:           "failing, signed, with custom headers",
			failing:        true,
			selfSignedCert: false,
			headers:        true,
			expectedURL:    test.LookupWithHeaderAuth.URLValid,
			expectedSecret: test.WebhookGitHub.SecretFail,
			expectedHeaders: &webhook.Headers{
				{
					Key:   test.LookupWithHeaderAuth.HeaderKey,
					Value: test.LookupWithHeaderAuth.HeaderValueFail,
				},
			},
		},
		{
			name:            "failing, self-signed, no custom headers",
			failing:         true,
			selfSignedCert:  true,
			headers:         false,
			expectedURL:     test.WebhookGitHub.URLInvalid,
			expectedSecret:  test.WebhookGitHub.SecretFail,
			expectedHeaders: nil,
		},
		{
			name:           "failing, self-signed, with custom headers",
			failing:        true,
			selfSignedCert: true,
			headers:        true,
			expectedURL:    test.LookupWithHeaderAuth.URLInvalid,
			expectedSecret: test.WebhookGitHub.SecretFail,
			expectedHeaders: &webhook.Headers{
				{
					Key:   test.LookupWithHeaderAuth.HeaderKey,
					Value: test.LookupWithHeaderAuth.HeaderValueFail,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := Webhook(
				t,
				tc.failing,
				tc.selfSignedCert,
				tc.headers,
			)

			if testErr := test.AssertSlicesEqual(
				t,
				result.Headers,
				util.DerefOrZero(tc.expectedHeaders),
				"Webhook.Headers",
			); testErr != nil {
				t.Error(testErr)
			}

			fieldTests := []test.FieldAssertion{
				{Name: "URL", Got: result.URL, Want: tc.expectedURL, Mode: test.CompareEqual},
				{Name: "Secret", Got: result.Secret, Want: tc.expectedSecret, Mode: test.CompareEqual},
				{Name: "ID", Got: result.ID, Want: "test", Mode: test.CompareEqual},
				{Name: "Delay", Got: result.Delay, Want: "0s", Mode: test.CompareEqual},
				{Name: "DesiredStatusCode", Got: result.DesiredStatusCode, Mode: test.CompareNotNil},
				{Name: "*DesiredStatusCode", Got: util.DerefOrZero(result.DesiredStatusCode), Want: uint16(0), Mode: test.CompareEqual},
				{Name: "MaxTries", Got: result.MaxTries, Mode: test.CompareNotNil},
				{Name: "*MaxTries", Got: util.DerefOrZero(result.MaxTries), Want: uint8(1), Mode: test.CompareEqual},
				{Name: "Main", Got: result.Main, Mode: test.CompareNotNil},
				{Name: "Defaults", Got: result.Defaults, Mode: test.CompareNotNil},
				{Name: "HardDefaults", Got: result.HardDefaults, Mode: test.CompareNotNil},
				{Name: "ServiceStatus", Got: result.ServiceStatus, Mode: test.CompareNotNil},
				{Name: "Failed", Got: result.Failed, Mode: test.CompareNotNil},
			}
			if testErr := test.AssertFields(t, fieldTests, "Webhook"); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}
