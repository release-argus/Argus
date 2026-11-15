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
	"io"
	"net/http"
	"testing"
)

func TestHeaderAuthServer(t *testing.T) {
	hookURL := HeaderAuthServer(t, "1.2.3")
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{
			name:   "pass header",
			header: LookupWithHeaderAuth.HeaderValuePass,
			want:   "1.2.3",
		},
		{
			name:   "fail header",
			header: LookupWithHeaderAuth.HeaderValueFail,
			want:   "Hook rules were not satisfied.",
		},
		{
			name: "no header",
			want: "Hook rules were not satisfied.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, hookURL, nil)
			if err != nil {
				t.Fatalf("http.NewRequestWithContext() failed: %v", err)
			}
			if tc.header != "" {
				req.Header.Set(LookupWithHeaderAuth.HeaderKey, tc.header)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("http.Client.Do() failed: %v", err)
			}
			t.Cleanup(func() { _ = resp.Body.Close() })

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}
			if got := string(body); got != tc.want {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}
