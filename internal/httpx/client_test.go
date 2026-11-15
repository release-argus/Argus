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

package httpx

import (
	"crypto/tls"
	"net/http"
	"testing"
)

func TestInit(t *testing.T) {
	if InsecureTransport.TLSClientConfig == nil {
		t.Fatal("InsecureTransport.TLSClientConfig mismatch\ngot:  nil\nwant: non-nil")
	}
	if !InsecureTransport.TLSClientConfig.InsecureSkipVerify {
		t.Fatal(
			"InsecureTransport.TLSClientConfig.InsecureSkipVerify mismatch\n" +
				"got:  false\nwant: true",
		)
	}
}

func TestConfigureInsecureTransport(t *testing.T) {
	tests := []struct {
		name            string
		transport       *http.Transport
		wantMinVersion  uint16
		checkMinVersion bool
	}{
		{
			name:      "nil TLSClientConfig",
			transport: &http.Transport{},
		},
		{
			name: "existing TLSClientConfig",
			transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
			},
			wantMinVersion:  tls.VersionTLS12,
			checkMinVersion: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			configureInsecureTransport(tc.transport)

			cfg := tc.transport.TLSClientConfig
			if cfg == nil {
				t.Fatal("Transport.TLSClientConfig mismatch\ngot:  nil\nwant: non-nil")
			}
			if !cfg.InsecureSkipVerify {
				t.Fatal(
					"Transport.TLSClientConfig.InsecureSkipVerify mismatch\n" +
						"got:  false\nwant: true",
				)
			}
			if tc.checkMinVersion && cfg.MinVersion != tc.wantMinVersion {
				t.Fatalf(
					"Transport.TLSClientConfig.MinVersion mismatch\ngot:  %s\nwant: %s",
					tls.VersionName(cfg.MinVersion), tls.VersionName(tc.wantMinVersion),
				)
			}
		})
	}
}
