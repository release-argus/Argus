// Copyright [2024] [Argus]
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

// Package base provides the base struct for latest_version lookups.
package base

import (
	"strings"
	"testing"

	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestDefaults_Default(t *testing.T) {
	// GIVEN a LookupDefault
	defaults := Defaults{}

	// WHEN Default is called
	defaults.Default()

	// THEN it should set the defaults
	if defaults.AllowInvalidCerts == nil {
		t.Errorf("AllowInvalidCerts not set, got %v",
			defaults.AllowInvalidCerts)
	}
	if defaults.UsePreRelease == nil {
		t.Errorf("UsePreRelease not set, got %v",
			defaults.UsePreRelease)
	}
	// AND Require has been defaulted
	if defaults.Require.Docker.Type == "" {
		t.Errorf("Require.Docker.Type not set, got %v",
			defaults.Require.Docker.Type)
	}
}

func TestDefaults_CheckValues(t *testing.T) {
	type args struct {
		require         filter.RequireDefaults
		urlCommandSlice filter.URLCommandSlice
	}
	// GIVEN a LookupDefault
	tests := map[string]struct {
		args     args
		errRegex string
	}{
		"valid": {
			args: args{require: filter.RequireDefaults{
				Docker: *filter.NewDockerCheckDefaults(
					"ghcr", "", "", "", "", nil)}},
			errRegex: `^$`,
		},
		"invalid require": {
			errRegex: test.TrimYAML(`
				^require:
					docker:
						type: "[^"]+" <invalid>.*$`),
			args: args{require: filter.RequireDefaults{
				Docker: *filter.NewDockerCheckDefaults(
					"someType", "", "", "", "", nil)}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			defaults := Defaults{
				Require: tc.args.require}

			// WHEN CheckValues is called
			err := defaults.CheckValues("")

			// THEN it errors when expected
			e := util.ErrorToString(err)
			lines := strings.Split(e, "\n")
			wantLines := strings.Count(tc.errRegex, "\n")
			if wantLines > len(lines) {
				t.Fatalf("Defaults.CheckValues() want %d lines of error:\n%q\ngot %d lines:\n%v\nstdout: %q",
					wantLines, tc.errRegex, len(lines), lines, e)
				return
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("Defaults.CheckValues() error mismatch\nwant match for:\n%q\ngot:\n%q",
					tc.errRegex, e)
				return
			}
		})
	}
}
