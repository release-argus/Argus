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

package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestRun(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	t.Cleanup(server.Close)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() failed: %v", err)
	}
	closedAddr := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatalf("Listener.Close() failed: %v", err)
	}

	tests := []struct {
		name     string
		args     []string
		errRegex string
	}{
		{
			name:     "missing arguments",
			args:     []string{},
			errRegex: `^expected URL as command-line argument$`,
		},
		{
			name: "unreachable URL",
			args: []string{"http://" + closedAddr},
			errRegex: test.TrimYAML(`
				^error:
					Get "http://` + regexp.QuoteMeta(closedAddr) + `":
						dial tcp ` + regexp.QuoteMeta(closedAddr) + `:
							connect:
								connection refused$`,
			),
		},
		{
			name:     "valid URL",
			args:     []string{server.URL},
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: shares http.DefaultTransport, which run writes.

			err := run(tc.args)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}
