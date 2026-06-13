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
	"fmt"
	"net/http"
	"testing"
)

func TestConfigureInsecureTransport(t *testing.T) {
	// GIVEN: a transport.
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

			// WHEN: ConfigureInsecureTransport is called on it.
			configureInsecureTransport(tc.transport)

			prefix := fmt.Sprintf(
				"%s\nconfigureInsecureTransport(%v)",
				packageName, tc.transport,
			)

			// THEN: the transport correctly configured to allow insecure connections.
			cfg := tc.transport.TLSClientConfig
			if cfg == nil {
				t.Fatalf("%s Transport.TLSClientConfig is nil", prefix)
			}
			if !cfg.InsecureSkipVerify {
				t.Fatalf("%s Transport.TLSClientConfig.InsecureSkipVerify is false", prefix)
			}
			if tc.checkMinVersion && cfg.MinVersion != tc.wantMinVersion {
				t.Fatalf(
					"%s MinVersion mismatch (TLSClientConfig was replaced)\ngot:  %d\nwant: %d",
					prefix, cfg.MinVersion, tc.wantMinVersion,
				)
			}
		})
	}
}

func TestInsecureTransport_init(t *testing.T) {
	if InsecureTransport.TLSClientConfig == nil {
		t.Fatalf("%s\nInsecureTransport.TLSClientConfig is nil after init", packageName)
	}
	if !InsecureTransport.TLSClientConfig.InsecureSkipVerify {
		t.Fatalf("%s\nInsecureTransport.TLSClientConfig.InsecureSkipVerify is false after init", packageName)
	}
}
