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

package web

import (
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestLookup_GetVersion(t *testing.T) {
	const (
		jsonBody  = `{"bar":"1.2.2","foo":{"bar":{"version":"3.2.1"}}}`
		plainBody = `version: "1.2.1"` + "\n" +
			`non-semantic: "ver1.2.2"`
	)

	tests := []struct {
		name                 string
		body                 []byte
		json                 string
		regex, regexTemplate string
		semVer               bool
		wantVersion          string
		errRegex             string
	}{
		{
			name:     "empty body",
			body:     nil,
			semVer:   true,
			errRegex: `^no version found in`,
		},
		{
			name:        "plain body without regex or JSON",
			body:        []byte("1.2.3"),
			semVer:      true,
			wantVersion: "1.2.3",
			errRegex:    `^$`,
		},
		{
			name:     "JSON/empty string value",
			body:     []byte(`{"bar":""}`),
			json:     "bar",
			semVer:   true,
			errRegex: `^no version found in.*$`,
		},
		{
			name:        "JSON/top-level key",
			body:        []byte(jsonBody),
			json:        "bar",
			semVer:      true,
			wantVersion: "1.2.2",
			errRegex:    `^$`,
		},
		{
			name:        "JSON/nested key",
			body:        []byte(jsonBody),
			json:        "foo.bar.version",
			semVer:      true,
			wantVersion: "3.2.1",
			errRegex:    `^$`,
		},
		{
			name:     "JSON/invalid body",
			body:     []byte(plainBody),
			json:     "bar",
			semVer:   true,
			errRegex: `^failed to unmarshal response from.*`,
		},
		{
			name:   "JSON/missing key",
			body:   []byte(`{}`),
			json:   "missing",
			semVer: true,
			errRegex: test.TrimYAML(`
				^failed to navigate JSON:
					failed to find value for .*$`,
			),
		},
		{
			name:        "regex with capture group",
			body:        []byte(plainBody),
			regex:       `version: "([^"]+)"`,
			semVer:      true,
			wantVersion: "1.2.1",
			errRegex:    `^$`,
		},
		{
			name:        "regex without capture group",
			body:        []byte(plainBody),
			regex:       `[0-9]+\.[0-9]+\.[0-9]+`,
			semVer:      true,
			wantVersion: "1.2.1",
			errRegex:    `^$`,
		},
		{
			name:     "regex no match",
			body:     []byte(plainBody),
			regex:    `^bishBashBosh$`,
			semVer:   true,
			errRegex: `^regex .* didn't return any matches on.*$`,
		},
		{
			name: "regex with template",
			body: []byte(
				`stable release version info "1.2.1" and more`,
			),
			regex:         `(stable).*(version).*"([\d.]+).*(and)`,
			regexTemplate: `$2 $1 $4, $3`,
			semVer:        false,
			wantVersion:   "version stable and, 1.2.1",
			errRegex:      `^$`,
		},
		{
			name:        "JSON/then regex",
			body:        []byte(jsonBody),
			json:        "bar",
			regex:       `^([0-9.]+)$`,
			semVer:      true,
			wantVersion: "1.2.2",
			errRegex:    `^$`,
		},
		{
			name:     "semantic versioning rejects non-semantic version",
			body:     []byte(plainBody),
			regex:    `non-semantic: "([^"]+)"`,
			semVer:   true,
			errRegex: `^failed to convert .* to a semantic version.*$`,
		},
		{
			name:        "semantic versioning disabled allows non-semantic version",
			body:        []byte(plainBody),
			regex:       `non-semantic: "([^"]+)"`,
			semVer:      false,
			wantVersion: "ver1.2.2",
			errRegex:    `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t)
			lookup.JSON = tc.json
			lookup.Regex = tc.regex
			lookup.RegexTemplate = tc.regexTemplate
			lookup.Options.SemanticVersioning = &tc.semVer

			version, err := lookup.getVersion(tc.body, logx.LogFrom{})
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if version != tc.wantVersion {
				t.Errorf("version mismatch\ngot:  %q\nwant: %q", version, tc.wantVersion)
			}
		})
	}
}
