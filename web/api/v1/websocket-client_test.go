// Copyright [2023] [Argus]
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

package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetIP(t *testing.T) {
	// GIVEN a request
	tests := map[string]struct {
		headers    map[string]string
		remoteAddr string
		want       string
	}{
		"CF-Connecting-Ip": {
			want: "1.1.1.1",
			headers: map[string]string{
				"CF-Connecting-IP": "1.1.1.1",
				"X-REAL-IP":        "2.2.2.2",
				"X-FORWARDED-FOR":  "3.3.3.3"},
			remoteAddr: "4.4.4.4:123"},
		"X-Real-Ip": {
			want: "2.2.2.2",
			headers: map[string]string{
				"X-REAL-IP":       "2.2.2.2",
				"X-FORWARDED-FOR": "3.3.3.3"},
			remoteAddr: "4.4.4.4:123"},
		"X-Forwarded-For": {
			headers: map[string]string{
				"X-FORWARDED-FOR": "3.3.3.3"},
			remoteAddr: "4.4.4.4:123",
			want:       "3.3.3.3"},
		"RemoteAddr": {
			want:       "4.4.4.4",
			remoteAddr: "4.4.4.4:123"},
		"Invalid RemoteAddr (SplitHostPort fail)": {
			want:       "",
			remoteAddr: "1111"},
		"Invalid RemoteAddr (ParseIP fail)": {
			want:       "",
			remoteAddr: "1111:123"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/approvals", nil)
			for header, val := range tc.headers {
				req.Header.Set(header, val)
			}
			req.RemoteAddr = tc.remoteAddr

			// WHEN getIP is called on this request
			got := getIP(req)

			// THEN the function returns the correct result
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %v",
					tc.want, got)
			}
		})
	}
}
