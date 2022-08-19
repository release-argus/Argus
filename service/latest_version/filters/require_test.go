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
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

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
					t.Fatalf("JLog shouldn't have been initialised to the one we called Init with when Require is %v",
						tc.req)
				}
				// THEN the Require is still nil
				if tc.req != nil {
					t.Fatal("Init with a nil require shouldn't inititalise it")
				}
			} else {
				if jLog != newJLog {
					t.Fatal("JLog should have been initialised to the one we called Init with")
				}
				// THEN the status is given to the Require
				if tc.req.Status != &status {
					t.Fatalf("Status should be the address of the var given to it %v, not %v",
						&status, tc.req.Status)
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
		"only command":       {require: &Require{Command: []string{"bash", "update.sh"}}, lines: 2},
		"only docker":        {require: &Require{Docker: &DockerCheck{Type: "ghcr"}}, lines: 3},
		"full require":       {require: &Require{RegexContent: "content", RegexVersion: "version", Command: []string{"bash", "update.sh"}, Docker: &DockerCheck{Type: "ghcr"}}, lines: 6},
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
			out, _ := io.ReadAll(r)
			os.Stdout = stdout
			got := strings.Count(string(out), "\n")
			if got != tc.lines {
				t.Errorf("Print should have given %d lines, but gave %d\n%s",
					tc.lines, got, out)
			}
		})
	}
}

func TestRequireCheckValues(t *testing.T) {
	// GIVEN a Require
	tests := map[string]struct {
		require  *Require
		errRegex []string
	}{
		"nil":                            {require: nil, errRegex: []string{"^$"}},
		"valid regex_content regex":      {require: &Require{RegexContent: "[0-9]"}, errRegex: []string{`^$`}},
		"invalid regex_content regex":    {require: &Require{RegexContent: "[0-"}, errRegex: []string{`^require:$`, `^  regex_content: .* <invalid>.*RegEx`}},
		"valid regex_content template":   {require: &Require{RegexContent: `{% if  version $}.linux-amd64{% endif %}`}, errRegex: []string{`^$`}},
		"invalid regex_content template": {require: &Require{RegexContent: "{% if  version }.linux-amd64"}, errRegex: []string{`^require:$`, `^  regex_content: .* <invalid>.*templating`}},
		"valid regex_version":            {require: &Require{RegexVersion: "[0-9]"}, errRegex: []string{`^$`}},
		"invalid regex_version":          {require: &Require{RegexVersion: "[0-"}, errRegex: []string{`^require:$`, `^  regex_version: .* <invalid>`}},
		"valid command":                  {require: &Require{Command: []string{"bash", "update.sh", "{{ version }}"}}, errRegex: []string{`^$`}},
		"invalid command":                {require: &Require{Command: []string{"{{ version }"}}, errRegex: []string{`^require:$`, `^  command: .* <invalid>.*templating`}},
		"valid docker":                   {require: &Require{Docker: &DockerCheck{Type: "ghcr", Image: "release-argus/argus", Tag: "{{ version }}"}}, errRegex: []string{`^$`}},
		"invalid docker":                 {require: &Require{Docker: &DockerCheck{Type: "foo"}}, errRegex: []string{`^require:$`, `^  docker:$`, `^    type: .* <invalid>`}},
		"all possible errors": {require: &Require{RegexContent: "[0-", RegexVersion: "[0-", Docker: &DockerCheck{Type: "foo"}},
			errRegex: []string{`^require:$`, `^  regex_content: .* <invalid>`, `^  regex_version: .* <invalid>`, `^  docker:$`, `^    type: .* <invalid>`}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// WHEN CheckValues is called on it
			err := tc.require.CheckValues("")

			// THEN err is expected
			e := utils.ErrorToString(err)
			lines := strings.Split(e, `\`)
			for i := range tc.errRegex {
				re := regexp.MustCompile(tc.errRegex[i])
				found := false
				for j := range lines {
					match := re.MatchString(lines[j])
					if match {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("want match for: %q\ngot:  %q",
						tc.errRegex[i], strings.ReplaceAll(e, `\`, "\n"))
				}
			}
		})
	}
}
