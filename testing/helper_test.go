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

// Package testing provides utilities for CLI-based testing.
package testing

import (
	"os"
	"testing"
)

func TestRunAndExit(t *testing.T) {
	type args struct {
		ok   bool
		flag string
	}

	tests := []struct {
		name     string
		args     args
		wantExit bool
		wantCode int
	}{
		{
			name:     "empty flag",
			args:     args{ok: true, flag: ""},
			wantExit: false,
		},
		{
			name:     "ok=true, flag set",
			args:     args{ok: true, flag: "set"},
			wantExit: true,
			wantCode: 0,
		},
		{
			name:     "ok=false, flag set",
			args:     args{ok: false, flag: "set"},
			wantExit: true,
			wantCode: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: swaps exitFunc.

			flag := tc.args.flag
			var exited bool
			var code int

			exitFunc = func(c int) {
				exited = true
				code = c
			}
			defer func() { exitFunc = os.Exit }()

			RunAndExit(tc.args.ok, &flag)

			if exited != tc.wantExit {
				t.Fatalf("exited mismatch\ngot:  %t\nwant: %t", exited, tc.wantExit)
			}

			if exited && code != tc.wantCode {
				t.Fatalf("exit code mismatch\ngot:  %d\nwant: %d", code, tc.wantCode)
			}
		})
	}
}
