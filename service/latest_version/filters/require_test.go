// Copyright [2022] [Argus]
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

//go:built unit

package filters

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"

	github_types "github.com/release-argus/Argus/service/latest_version/api_types"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
)

func TestRequireInit(t *testing.T) {
	// GIVEN a Require, JLog and a Status
	tests := map[string]struct {
		req   *Require
		lines int
	}{
		"nil require":     {req: nil},
		"non-nil require": {req: &Require{}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			status := service_status.Status{DeployedVersion: "1.2.3"}
			newJLog := utils.NewJLog("WARN", false)

			// WHEN Init is called with it
			tc.req.Init(newJLog, &status)

			// THEN the global JLog is set to its address
			if tc.req == nil {
				if jLog == newJLog {
					t.Fatalf("%s:\nJLog shouldn't have been initialised to the one we called Init with when Require is %v",
						name, tc.req)
				}
				// THEN the Require is still nil
				if tc.req != nil {
					t.Fatalf("%s:\nInit with a nil require shouldn't inititalise it",
						name)
				}
			} else {
				if jLog != newJLog {
					t.Fatalf("%s:\nJLog should have been initialised to the one we called Init with",
						name)
				}
				// THEN the status is given to the Require
				if tc.req.Status != &status {
					t.Fatalf("%s:\nStatus should be the address of the var given to it %v, not %v",
						name, &status, tc.req.Status)
				}
			}
		})
	}
}

func TestRequirePrint(t *testing.T) {
	// GIVEN a Require
	tests := map[string]struct {
		require *Require
		lines   int
	}{
		"nil require":        {require: nil, lines: 0},
		"only regex_content": {require: &Require{RegexContent: "content"}, lines: 2},
		"only regex_version": {require: &Require{RegexVersion: "version"}, lines: 2},
		"full require":       {require: &Require{RegexContent: "content", RegexVersion: "version"}, lines: 3},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN Print is called on it
			tc.require.Print("")

			// THEN the expected number of lines are printed
			w.Close()
			out, _ := ioutil.ReadAll(r)
			os.Stdout = stdout
			got := strings.Count(string(out), "\n")
			if got != tc.lines {
				t.Errorf("%s:\nPrint should have given %d lines, but gave %d\n%s",
					name, tc.lines, got, out)
			}
		})
	}
}

func TestRequireRegexCheckVersion(t *testing.T) {
	// GIVEN a Require
	jLog := utils.NewJLog("WARN", false)
	tests := map[string]struct {
		require  *Require
		errRegex string
	}{
		"nil require":         {require: nil, errRegex: "^$"},
		"empty regex_version": {require: &Require{}, errRegex: "^$"},
		"match":               {require: &Require{RegexVersion: "^[0-9.]+-beta$"}, errRegex: "^$"},
		"no match":            {require: &Require{RegexVersion: "^[0-9.]+$"}, errRegex: "regex not matched on version"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.require != nil {
				tc.require.Status = &service_status.Status{}
			}

			// WHEN RegexCheckVersion is called on it
			err := tc.require.RegexCheckVersion("0.1.1-beta", jLog, utils.LogFrom{})

			// THEN the err is what we expect
			e := utils.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("%s:\nwant match for %q\nnot: %q",
					name, tc.errRegex, e)
			}
		})
	}
}

func TestRequireRegexCheckContent(t *testing.T) {
	// GIVEN a Require
	jLog := utils.NewJLog("WARN", false)
	tests := map[string]struct {
		require  *Require
		body     interface{}
		errRegex string
	}{
		"nil require":         {require: nil, errRegex: "^$"},
		"empty regex_content": {require: &Require{}, errRegex: "^$"},
		"invalid body": {require: &Require{RegexContent: `argus-[0-9.]+.linux-amd64`}, errRegex: "invalid body type",
			body: 123},
		"string body match": {require: &Require{RegexContent: `argus-[0-9.]+.linux-amd64`}, errRegex: "^$",
			body: `darwin amd64 - argus-1.2.3.darwin-amd64, linux amd64 - argus-1.2.3.linux-amd64, windows amd64 - argus-1.2.3.windows-amd64,`},
		"string body no match": {require: &Require{RegexContent: `argus-[0-9.]+.linux-amd64`}, errRegex: "regex .* not matched on content",
			body: `darwin amd64 - argus-1.2.3.darwin-amd64, linux arm64 - argus-1.2.3.linux-arm64, windows amd64 - argus-1.2.3.windows-amd64,`},
		"github api body match": {require: &Require{RegexContent: `argus-[0-9.]+.linux-amd64`}, errRegex: "^$",
			body: []github_types.Asset{{Name: "argus-1.2.3.darwin-amd64"}, {Name: "argus-1.2.3.linux-amd64"}, {Name: "argus-1.2.3.windows-amd64"}}},
		"github api body no match": {require: &Require{RegexContent: `argus-[0-9.]+.linux-amd64`}, errRegex: "regex .* not matched on content",
			body: []github_types.Asset{{Name: "argus-1.2.3.darwin-amd64"}, {Name: "argus-1.2.3.linux-arm64"}, {Name: "argus-1.2.3.windows-amd64"}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.require != nil {
				tc.require.Status = &service_status.Status{}
			}

			// WHEN RegexCheckContent is called on it
			err := tc.require.RegexCheckContent("0.1.1-beta", tc.body, jLog, utils.LogFrom{})

			// THEN the err is what we expect
			e := utils.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("%s:\nwant match for %q\nnot: %q",
					name, tc.errRegex, e)
			}
		})
	}
}

func TestRequireCheckValues(t *testing.T) {
	// GIVEN a Require
	tests := map[string]struct {
		require  *Require
		errRegex string
	}{
		"nil":                   {require: nil, errRegex: "^$"},
		"invalid regex_content": {require: &Require{RegexContent: "[0-"}, errRegex: "regex_content: .* <invalid>"},
		"invalid regex_version": {require: &Require{RegexVersion: "[0-"}, errRegex: "regex_version: .* <invalid>"},
		"all possible errors":   {require: &Require{RegexContent: "[0-", RegexVersion: "[0-"}, errRegex: `regex_content: .* <invalid>.*\s *regex_version: .* <invalid>`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// WHEN CheckValues is called on it
			err := tc.require.CheckValues("")

			// THEN err is expected
			e := utils.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("%s:\nwant match for %q\nnot: %q",
					name, tc.errRegex, e)
			}
		})
	}
}
