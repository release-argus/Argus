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

//go:build unit

package filter

import (
	"io"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	command "github.com/release-argus/Argus/commands"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

func TestRequire_Init(t *testing.T) {
	// GIVEN a Require, JLog and a Status
	tests := map[string]struct {
		req   *Require
		lines int
	}{
		"nil require": {
			req: nil},
		"non-nil require": {
			req: &Require{}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			status := svcstatus.Status{}
			status.Init(
				0, 0, 0,
				stringPtr("test"),
				stringPtr("http://example.com"))
			status.SetDeployedVersion("1.2.3", false)

			// WHEN Init is called with it
			tc.req.Init(&status)

			// THEN the global JLog is set to its address
			if tc.req == nil {
				// THEN the Require is still nil
				if tc.req != nil {
					t.Fatal("Init with a nil require shouldn't inititalise it")
				}
			} else {
				// THEN the status is given to the Require
				if tc.req.Status != &status {
					t.Fatalf("Status should be the address of the var given to it %v, not %v",
						&status, tc.req.Status)
				}
			}
		})
	}
}

func TestRequire_Print(t *testing.T) {
	// GIVEN a Require
	tests := map[string]struct {
		require *Require
		lines   int
	}{
		"nil require": {
			require: nil,
			lines:   0,
		},
		"only regex_content": {
			require: &Require{
				RegexContent: "content"},
			lines: 2,
		},
		"only regex_version": {
			require: &Require{
				RegexVersion: "version"},
			lines: 2,
		},
		"only command": {
			require: &Require{
				Command: []string{"bash", "update.sh"}},
			lines: 2,
		},
		"only docker": {
			require: &Require{
				Docker: &DockerCheck{Type: "ghcr"}},
			lines: 3,
		},
		"full require": {
			require: &Require{
				RegexContent: "content",
				RegexVersion: "version",
				Command:      []string{"bash", "update.sh"},
				Docker:       &DockerCheck{Type: "ghcr"}},
			lines: 6,
		},
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

func TestRequire_CheckValues(t *testing.T) {
	// GIVEN a Require
	tests := map[string]struct {
		require  *Require
		errRegex []string
	}{
		"nil": {
			require:  nil,
			errRegex: []string{"^$"},
		},
		"valid regex_content regex": {
			require: &Require{
				RegexContent: "[0-9]"},
			errRegex: []string{`^$`},
		},
		"invalid regex_content regex": {
			require: &Require{
				RegexContent: "[0-"},
			errRegex: []string{
				`^require:$`,
				`^  regex_content: .* <invalid>.*RegEx`},
		},
		"valid regex_content template": {
			require: &Require{
				RegexContent: `{% if  version $}.linux-amd64{% endif %}`},
			errRegex: []string{`^$`},
		},
		"invalid regex_content template": {
			require: &Require{
				RegexContent: "{% if  version }.linux-amd64"},
			errRegex: []string{
				`^require:$`,
				`^  regex_content: .* <invalid>.*templating`},
		},
		"valid regex_version": {
			require: &Require{
				RegexVersion: "[0-9]"},
			errRegex: []string{`^$`},
		},
		"invalid regex_version": {
			require: &Require{
				RegexVersion: "[0-"}, errRegex: []string{
				`^require:$`,
				`^  regex_version: .* <invalid>`},
		},
		"valid command": {
			require: &Require{
				Command: []string{
					"bash", "update.sh", "{{ version }}"}},
			errRegex: []string{`^$`},
		},
		"invalid command": {
			require: &Require{
				Command: []string{"{{ version }"}},
			errRegex: []string{
				`^require:$`,
				`^  command: .* <invalid>.*templating`},
		},
		"valid docker": {
			require: &Require{
				Docker: &DockerCheck{
					Type:  "ghcr",
					Image: "release-argus/argus",
					Tag:   "{{ version }}"}},
			errRegex: []string{`^$`},
		},
		"invalid docker": {
			require: &Require{
				Docker: &DockerCheck{
					Type: "foo"}},
			errRegex: []string{
				`^require:$`,
				`^  docker:$`,
				`^    type: .* <invalid>`},
		},
		"all possible errors": {
			require: &Require{
				RegexContent: "[0-",
				RegexVersion: "[0-",
				Docker: &DockerCheck{
					Type: "foo"}},
			errRegex: []string{
				`^require:$`,
				`^  regex_content: .* <invalid>`,
				`^  regex_version: .* <invalid>`,
				`^  docker:$`,
				`^    type: .* <invalid>`},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// WHEN CheckValues is called on it
			err := tc.require.CheckValues("")

			// THEN err is expected
			e := util.ErrorToString(err)
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

func TestRequire_FromStr(t *testing.T) {
	testLogging("WARN")
	// GIVEN a JSON string and a Require to use as defaults
	dflt := testRequire()
	tests := map[string]struct {
		jsonStr          *string
		dflt             *Require
		errRegex         string
		want             *Require
		pointerToDefault bool
	}{
		"nil": {
			jsonStr:  stringPtr(""),
			errRegex: "EOF$",
			want:     nil,
		},
		"nil with default": {
			jsonStr:          nil,
			dflt:             &dflt,
			pointerToDefault: true,
		},
		"empty": {
			jsonStr: stringPtr("{}"),
			want:    nil,
		},
		"empty with default": {
			jsonStr:          stringPtr("{}"),
			pointerToDefault: true,
		},
		"invalid JSON": {
			jsonStr:  stringPtr("{"),
			errRegex: `unexpected EOF$`,
		},
		"invalid JSON with default": {
			jsonStr:          stringPtr("{"),
			dflt:             &dflt,
			pointerToDefault: true,
			errRegex:         `unexpected EOF$`,
		},
		"invalid data type": {
			jsonStr: stringPtr(`{
"regex_content": 1}`),
			errRegex: `json: cannot unmarshal number into Go struct field Require.regex_content of type string$`,
		},
		"invalid data type with default": {
			jsonStr: stringPtr(`{
"regex_content": 1}`),
			dflt:             &dflt,
			pointerToDefault: true,
			errRegex:         `json: cannot unmarshal number into Go struct field Require.regex_content of type string$`,
		},
		"RegexContent defined": {
			jsonStr: stringPtr(`{
"regex_content": "foo"}`),
			want: &Require{
				RegexContent: "foo"},
		},
		"RegexVersion defined": {
			jsonStr: stringPtr(`{
"regex_version": "foo"}`),
			want: &Require{
				RegexVersion: "foo"},
		},
		"RegexContent from str, RegexVersion from default": {
			jsonStr: stringPtr(`{
"regex_content": "foo"}`),
			dflt: &Require{
				RegexVersion: "bar"},
			want: &Require{
				RegexContent: "foo",
				RegexVersion: "bar"},
		},
		"RegexContent from default, RegexVersion from str": {
			jsonStr: stringPtr(`{
"regex_version": "bar"}`),
			dflt: &Require{
				RegexContent: "foo"},
			want: &Require{
				RegexContent: "foo",
				RegexVersion: "bar"},
		},
		"Command defined": {
			jsonStr: stringPtr(`{
"command":[
	"foo",
	"bar"]}`),
			want: &Require{Command: []string{"foo", "bar"}},
		},
		"No Command JSON uses default": {
			jsonStr: stringPtr(`{
"regex_version": "foo"}`),
			dflt: &Require{
				RegexVersion: "bar",
				Command:      []string{"foo", "bar"}},
			want: &Require{
				RegexVersion: "foo",
				Command:      []string{"foo", "bar"}},
		},
		"Empty command overrides default": {
			jsonStr: stringPtr(`{
"regex_version": "foo",
"command": []}`),
			dflt: &Require{
				RegexVersion: "bar",
				Command:      []string{"foo", "bar"}},
			want: &Require{
				RegexVersion: "foo",
				Command:      []string{}},
		},
		"Only Docker.Type sent": {
			jsonStr: stringPtr(`{
"docker": {
	"type": "ghcr"}}`),
			want:     &Require{},
			errRegex: "^$",
		},
		"Docker defined": {
			jsonStr: stringPtr(`{
"docker": {
	"type": "ghcr",
	"image": "release-argus/argus",
	"tag": "latest",
	"username": "magic",
	"token": "admin"}}`),
			want: &Require{
				Docker: &DockerCheck{
					Type:     "ghcr",
					Image:    "release-argus/argus",
					Tag:      "latest",
					Username: "magic",
					Token:    "admin"}},
		},
		"Docker changing Type": {
			jsonStr: stringPtr(`{
"docker": {
	"type": "ghcr"}}`),
			dflt: &Require{
				Docker: &DockerCheck{
					Type:     "hub",
					Image:    "releaseargus/argus",
					Tag:      "latest",
					Username: "magic",
					Token:    "admin"}},
			want: &Require{
				Docker: &DockerCheck{
					Type:     "ghcr",
					Image:    "releaseargus/argus",
					Tag:      "latest",
					Username: "magic",
					Token:    "admin"}},
		},
		"Docker changing Image": {
			jsonStr: stringPtr(`{
"docker": {
	"image": "release-argus/argus"}}`),
			dflt: &Require{
				Docker: &DockerCheck{
					Type:     "hub",
					Image:    "releaseargus/argus",
					Tag:      "latest",
					Username: "magic",
					Token:    "admin"}},
			want: &Require{
				Docker: &DockerCheck{
					Type:     "hub",
					Image:    "release-argus/argus",
					Tag:      "latest",
					Username: "magic",
					Token:    "admin"}},
		},
		"Docker changing Tag": {
			jsonStr: stringPtr(`{
"docker": {
	"tag": "1.2.3"}}`),
			dflt: &Require{
				Docker: &DockerCheck{
					Type:     "hub",
					Image:    "releaseargus/argus",
					Tag:      "latest",
					Username: "magic",
					Token:    "admin"}},
			want: &Require{
				Docker: &DockerCheck{
					Type:     "hub",
					Image:    "releaseargus/argus",
					Tag:      "1.2.3",
					Username: "magic",
					Token:    "admin"}},
		},
		"Docker changing Username": {
			jsonStr: stringPtr(`{
"docker": {
	"username": "sir"}}`),
			dflt: &Require{
				Docker: &DockerCheck{
					Type:     "hub",
					Image:    "releaseargus/argus",
					Tag:      "latest",
					Username: "magic",
					Token:    "admin"}},
			want: &Require{
				Docker: &DockerCheck{
					Type:     "hub",
					Image:    "releaseargus/argus",
					Tag:      "latest",
					Username: "sir",
					Token:    "admin"}},
		},
		"Docker changing Token": {
			jsonStr: stringPtr(`{
"docker": {
	"token": "letmein"}}`),
			dflt: &Require{
				Docker: &DockerCheck{
					Type:     "hub",
					Image:    "releaseargus/argus",
					Tag:      "latest",
					Username: "magic",
					Token:    "admin"}},
			want: &Require{
				Docker: &DockerCheck{
					Type:     "hub",
					Image:    "releaseargus/argus",
					Tag:      "latest",
					Username: "magic",
					Token:    "letmein"}},
		},
		"Docker changing multiple (GHCR Argus)": {
			jsonStr: stringPtr(`{
"docker": {
	"type": "ghcr",
	"image": "release-argus/argus",
	"username": "",
	"token": "bar"}}`),
			dflt: &Require{
				Docker: &DockerCheck{
					Type:     "hub",
					Image:    "releaseargus/argus",
					Tag:      "latest",
					Username: "magic",
					Token:    "admin"}},
			want: &Require{
				Docker: &DockerCheck{
					Type:     "ghcr",
					Image:    "release-argus/argus",
					Tag:      "latest",
					Username: "",
					Token:    "bar"}},
		},
		"only RegexContent changed keeps default Docker": {
			jsonStr: stringPtr(`{
"regex_content": "foo"}`),
			dflt: &Require{Docker: &DockerCheck{
				Type:  "ghcr",
				Image: "release-argus/argus",
				Tag:   "latest"}},
			want: &Require{
				RegexContent: "foo",
				Docker: &DockerCheck{
					Type:  "ghcr",
					Image: "release-argus/argus",
					Tag:   "latest"}},
		},
		"no Docker type defaults to 'hub'": {
			jsonStr: stringPtr(`{
"docker": {
	"image": "releaseargus/argus",
	"tag": "latest"}}`),
			dflt: &Require{},
			want: &Require{
				Docker: &DockerCheck{
					Type:  "hub",
					Image: "releaseargus/argus",
					Tag:   "latest"}},
		},
		"Invalid Docker (no tag)": {
			jsonStr: stringPtr(`{
"docker": {
	"type": "ghcr",
	"image": "release-argus/argus"}}`),
			errRegex: `docker.*tag.*required`,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			// t.Parallel()
			if tc.errRegex == "" {
				tc.errRegex = "^$"
			}
			if tc.jsonStr != nil {
				*tc.jsonStr = strings.ReplaceAll(*tc.jsonStr, "\n", "")
				*tc.jsonStr = strings.ReplaceAll(*tc.jsonStr, "\t", "")
			}

			// WHEN RequireFromStr is called on it
			got, err := RequireFromStr(tc.jsonStr, tc.dflt, &util.LogFrom{Primary: name})

			// THEN err is expected
			re := regexp.MustCompile(tc.errRegex)
			error := util.ErrorToString(err)
			if !re.MatchString(error) {
				t.Errorf("want: %q\ngot:  %q",
					tc.errRegex, error)
			} else if err != nil {
				// no need to continue
				return
			}
			// AND got is expected
			if !tc.pointerToDefault {
				if !reflect.DeepEqual(got, tc.want) {
					if tc.want.Docker == nil || got.Docker == nil {
						t.Errorf("\nwant: %v\ngot:  %v",
							tc.want, got)
					} else {
						t.Errorf("\nwant: %v\ngot:  %v",
							*tc.want, *got)
					}
				}

				//(pointer to the default if jsonStr was invalid)
			} else if got != tc.dflt {
				t.Errorf("didn't get default\nwant: %v\ngot:  %v",
					tc.dflt, got)
			}
		})
	}
}

func TestRequire__String(t *testing.T) {
	tests := map[string]struct {
		require *Require
		want    string
	}{
		"nil": {
			require: nil,
			want:    "<nil>"},
		"empty": {
			require: &Require{},
			want:    "{}\n"},
		"all fields defined": {
			require: &Require{
				Status:       &svcstatus.Status{},
				RegexContent: "abc{{ version }}.tar.gz",
				RegexVersion: "v([0-9.]+)",
				Command:      command.Command{"ls", "-la"},
				Docker: &DockerCheck{
					Type: "hub"},
			},
			want: `
regex_content: abc{{ version }}.tar.gz
regex_version: v([0-9.]+)
command:
    - ls
    - -la
docker:
    type: hub
    image: ""
    tag: ""
`},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the Require is stringified with String
			got := tc.require.String()

			// THEN the result is as expected
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}
