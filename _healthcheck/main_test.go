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
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

var packageName = "healthcheck"

func TestRun(t *testing.T) {
	// GIVEN: various arguments.
	tests := []struct {
		name        string
		args        []string
		stderrRegex string
	}{
		{
			name:        "missing arguments",
			args:        []string{},
			stderrRegex: `^expected URL as command-line argument$`,
		},
		{
			name: "invalid URL",
			args: []string{"http://invalid-url"},
			stderrRegex: test.TrimYAML(`
				^error:
					Get "http://invalid-url":
						dial tcp:
							lookup .* no such host$`,
			),
		},
		{
			name:        "valid URL",
			args:        []string{"https://www.google.com"},
			stderrRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: run is called with those arguments.
			err := run(tc.args)

			// THEN: it errors when expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.stderrRegex, e) {
				t.Errorf(
					"%s\nrun() stderr mismatch\ngot:  %q:\nwant: %q",
					packageName, e, tc.stderrRegex,
				)
			}
		})
	}
}
