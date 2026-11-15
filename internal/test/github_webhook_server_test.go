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

package test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/iotest"
)

func TestGitHubWebhookServer(t *testing.T) {
	server := GitHubWebhookServer(t)
	body := `{"ref":"refs/heads/main"}`
	sign := func(secret string) string {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(body))
		return "sha256=" + hex.EncodeToString(mac.Sum(nil))
	}
	tests := []struct {
		name       string
		signature  string
		wantStatus int
	}{
		{
			name:       "signed with the pass secret",
			signature:  sign(WebhookGitHub.SecretPass),
			wantStatus: http.StatusOK,
		},
		{
			name:       "signed with the fail secret",
			signature:  sign(WebhookGitHub.SecretFail),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "unsigned",
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequestWithContext(
				t.Context(), http.MethodPost, server.URL, strings.NewReader(body),
			)
			if err != nil {
				t.Fatalf("http.NewRequestWithContext() failed: %v", err)
			}
			if tc.signature != "" {
				req.Header.Set("X-Hub-Signature-256", tc.signature)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("http.Client.Do() failed: %v", err)
			}
			t.Cleanup(func() { _ = resp.Body.Close() })

			if got, want := resp.StatusCode, tc.wantStatus; got != want {
				t.Errorf("status code mismatch\ngot:  %d\nwant: %d", got, want)
			}
		})
	}
}

func TestGitHubWebhookServer__unreadableBody(t *testing.T) {
	server := GitHubWebhookServer(t)
	req := httptest.NewRequestWithContext(
		t.Context(), http.MethodPost, server.URL, iotest.ErrReader(errors.New("read failed")),
	)
	rec := httptest.NewRecorder()

	server.Config.Handler.ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusBadRequest; got != want {
		t.Errorf("status code mismatch\ngot:  %d\nwant: %d", got, want)
	}
}
