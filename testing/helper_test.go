// Copyright [2025] [Argus]
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
	// GIVEN various combinations of ok and flag.
	type args struct {
		ok   bool
		flag string
	}

	tests := map[string]struct {
		args     args
		wantExit bool
		wantCode int
	}{
		"empty flag": {
			args:     args{ok: true, flag: ""},
			wantExit: false,
		},
		"ok=true, flag set": {
			args:     args{ok: true, flag: "set"},
			wantExit: true,
			wantCode: 0,
		},
		"ok=false, flag set": {
			args:     args{ok: false, flag: "set"},
			wantExit: true,
			wantCode: 1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're sharing exitFunc.

			flag := tc.args.flag
			var exited bool
			var code int

			// Override exitFunc to capture the exit instead of actually exiting.
			exitFunc = func(c int) {
				exited = true
				code = c
			}
			defer func() { exitFunc = os.Exit }()

			// WHEN ExitIfTestFails is called.
			RunAndExit(tc.args.ok, &flag)

			// THEN it should exit as expected.
			if exited != tc.wantExit {
				t.Fatalf("%s\nwant: exit=%v\ngot:  exit=%v",
					packageName, tc.wantExit, exited)
			}
			// AND with the expected code.
			if exited && code != tc.wantCode {
				t.Fatalf("%s\nwant: exit=%d\ngot:  exit=%d",
					packageName, tc.wantCode, code)
			}
		})
	}
}
