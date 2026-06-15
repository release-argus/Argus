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

package shoutrrr

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
	"github.com/release-argus/Argus/util/polymorphic"
)

func TestShoutrrr_TestSend(t *testing.T) {
	// GIVEN: a Shoutrrr.
	tests := []struct {
		name        string
		sType       *string
		nilShoutrrr bool
		errRegex    string
	}{
		{
			name:        "nil shoutrrr",
			nilShoutrrr: true,
			errRegex:    `^shoutrrr is nil$`,
		},
		{
			name:  "invalid type",
			sType: test.Ptr("somethingUnknown"),
			errRegex: test.TrimYAML(`
				^failed to create Shoutrrr sender:
					creating sender for URLs \[\]:
						error initializing router services:
							unknown service: ""
								unknown service$`,
			),
		},
		{
			name:     "valid",
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			if tc.sType != nil {
				shoutrrr.Type = *tc.sType
			}
			if tc.nilShoutrrr {
				shoutrrr = nil
			}
			testURL := "https://example.com"

			// WHEN: TestSend is called.
			err := shoutrrr.TestSend(testURL)

			prefix := fmt.Sprintf(
				"%s\nShoutrrr.TestSend(%q)",
				packageName, testURL,
			)

			// THEN: it errors when expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
		})
	}
}

func TestShoutrrrs_CheckValues(t *testing.T) {
	notifyCfg := plainConfig(t)

	// GIVEN: Shoutrrrs.
	tests := []struct {
		name     string
		input    *Shoutrrrs
		errRegex string
		changed  bool
	}{
		{
			name:     "nil map",
			input:    (*Shoutrrrs)(nil),
			errRegex: `^$`,
			changed:  false,
		},
		{
			name:     "valid map",
			errRegex: `^$`,
			input: &Shoutrrrs{
				"valid": testShoutrrr(false, false),
				"other": testShoutrrr(false, false),
			},
			changed: false,
		},
		{
			name:     "valid map, with changed",
			errRegex: `^$`,
			input: &Shoutrrrs{
				"valid": testShoutrrr(false, false),
				"other": testShoutrrr(false, false),
				"generic": &Shoutrrr{
					Base: Base{
						URLFields: map[string]string{
							"host":           "example.com",
							"custom_headers": `{"foo":"bar"}`},
					},
				},
			},
			changed: true,
		},
		{
			name: "invalid map",
			errRegex: test.TrimYAML(`
				other:
					type: <required>`,
			),
			input: &Shoutrrrs{
				"valid": testShoutrrr(false, false),
				"other": New(
					nil, "", "",
					make(map[string]string),
					make(map[string]string),
					make(map[string]string),
					nil, nil, nil,
				),
			},
			changed: false,
		},
		{
			name: "ordered errors",
			errRegex: test.TrimYAML(`
				aNotify:
					type: <required>.*
				bNotify:
					type: <required>.*`,
			),
			input: &Shoutrrrs{
				"aNotify": New(
					nil, "", "",
					make(map[string]string),
					make(map[string]string),
					make(map[string]string),
					nil, nil, nil,
				),
				"bNotify": New(
					nil, "", "",
					make(map[string]string),
					make(map[string]string),
					make(map[string]string),
					nil, nil, nil,
				),
			},
			changed: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.input != nil {
				svcStatus, _ := statustest.New("yaml", nil)
				svcStatus.Init(
					0, len(*tc.input), 0,
					status.ServiceInfo{
						ID: tc.name,
					},
					&dashboard.Options{},
				)
				tc.input.Init(svcStatus, notifyCfg)
			}

			_, _ = test.AssertCheckValuesWithErrorAndChanged(
				t,
				packageName,
				tc.errRegex,
				tc.changed,
				tc.input.CheckValues,
			)
		})
	}
}
func TestShoutrrr_CheckValues(t *testing.T) {
	// GIVEN: a Shoutrrr.
	testS := testShoutrrr(false, false)
	tests := []struct {
		name                       string
		nilShoutrrr                bool
		sType                      string
		options, urlFields, params map[string]string
		wantURLFields, wantParams  map[string]string
		wantDelay                  string
		main                       *Defaults
		errRegex                   string
		changed                    bool
	}{
		{
			name:        "nil shoutrrr",
			nilShoutrrr: true,
			errRegex:    `^$`,
			changed:     false,
		},
		{
			name:      "empty",
			errRegex:  `^type: <required>[^:]+://[^:]+$`,
			urlFields: map[string]string{},
			params:    map[string]string{},
			changed:   false,
		},
		{
			name: "invalid delay",
			errRegex: test.TrimYAML(`
				^options:
					delay: .* <invalid>`,
			),
			sType:     testS.Type,
			urlFields: testS.URLFields,
			options: map[string]string{
				"delay": "5x",
			},
			changed: false,
		},
		{
			name:      "fixes delay",
			errRegex:  `^$`,
			sType:     testS.Type,
			urlFields: testS.URLFields,
			wantDelay: "5s",
			options: map[string]string{
				"delay": "5",
			},
			changed: false,
		},
		{
			name: "invalid message template",
			errRegex: test.TrimYAML(`
				^options:
					message: ".+" <invalid>.*$`,
			),
			sType:     testS.Type,
			urlFields: testS.URLFields,
			options: map[string]string{
				"message": "{{ version }",
			},
			changed: false,
		},
		{
			name:      "valid message template",
			errRegex:  `^$`,
			sType:     testS.Type,
			urlFields: testS.URLFields,
			options: map[string]string{
				"message": "{{ version }}",
			},
			changed: false,
		},
		{
			name: "invalid title template",
			errRegex: test.TrimYAML(`
				^params:
					title: .* <invalid>.*$`,
			),
			sType:     testS.Type,
			urlFields: testS.URLFields,
			params: map[string]string{
				"title": "{{ version }",
			},
			changed: false,
		},
		{
			name:      "valid title template",
			errRegex:  `^$`,
			sType:     testS.Type,
			urlFields: testS.URLFields,
			params: map[string]string{
				"title": "{{ version }}",
			},
			changed: false,
		},
		{
			name:      "valid param template",
			errRegex:  `^$`,
			sType:     testS.Type,
			urlFields: testS.URLFields,
			params: map[string]string{
				"foo": "{{ version }}",
			},
			changed: false,
		},
		{
			name: "invalid param template",
			errRegex: test.TrimYAML(`
				^params:
					foo: .* <invalid>.*$`,
			),
			sType:     testS.Type,
			urlFields: testS.URLFields,
			params: map[string]string{
				"foo": "{{ version }",
			},
			changed: false,
		},
		{
			name: "invalid param and option",
			errRegex: test.TrimYAML(`
				^options:
					delay: [^<]+<invalid>.*
				params:
					title: [^<]+<invalid>.*$`,
			),
			sType:     testS.Type,
			urlFields: testS.URLFields,
			params: map[string]string{
				"title": "{{ version }",
			},
			options: map[string]string{
				"delay": "2x",
			},
			changed: false,
		},
		{
			name:     "does correctSelf",
			errRegex: `^$`,
			sType:    testS.Type,
			urlFields: map[string]string{
				"host":  "foo",
				"token": "bar",
				"port":  ":8080",
				"path":  "/test",
			},
			wantURLFields: map[string]string{
				"port": "8080",
				"path": "test",
			},
			changed: true,
		},
		{
			name:  "generic.url_fields.custom_headers -> headers",
			sType: "generic",
			urlFields: map[string]string{
				"host":           "example.com",
				"custom_headers": `{"foo":"bar"}`,
			},
			wantURLFields: map[string]string{
				"custom_headers": `{"foo":"bar"}`,
			},
			changed: true,
		},
		{
			name:  "generic.url_fields.custom_headers not pulled to headers if headers already defined",
			sType: "generic",
			urlFields: map[string]string{
				"host":           "example.com",
				"custom_headers": `{"foo":"bar"}`,
				"headers":        `{"foo":"baz"}`,
			},
			wantURLFields: map[string]string{
				"headers": `{"foo":"baz"}`,
			},
			changed: true,
		},
		{
			name:  "ntfy.params.disabletls -> disabletlsverification",
			sType: "ntfy",
			urlFields: map[string]string{
				"topic": "123",
			},
			wantURLFields: map[string]string{
				"topic": "123",
			},
			params: map[string]string{
				"disabletls": "true",
			},
			wantParams: map[string]string{
				"disabletlsverification": "true",
			},
			changed: true,
		},
		{
			name:  "ntfy.params.disabletls not pulled to disabletlsverification if disabletlsverification already defined",
			sType: "ntfy",
			urlFields: map[string]string{
				"topic": "123",
			},
			wantURLFields: map[string]string{
				"topic": "123",
			},
			params: map[string]string{
				"disabletls":             "true",
				"disabletlsverification": "false",
			},
			wantParams: map[string]string{
				"disabletlsverification": "false",
			},
			changed: true,
		},
		{
			name:      "valid",
			errRegex:  `^$`,
			urlFields: map[string]string{},
			main:      testDefaults(false, false),
			changed:   false,
		},
		{
			name:     "valid with self and main",
			errRegex: `^$`,
			urlFields: map[string]string{
				"host": "foo",
			},
			main: NewDefaults(
				testS.Type,
				nil,
				map[string]string{
					"token": "bar",
				},
				nil,
			),
			changed: false,
		},
		{
			name: "invalid url_fields",
			errRegex: test.TrimYAML(`
				^url_fields:
					host: <required>.*
					token: <required>.*$`,
			),
			sType:   testS.Type,
			changed: false,
		},
		{
			name: "invalid params + locate fail",
			errRegex: test.TrimYAML(`
				^params:
					fromaddress: <required>.*
					toaddresses: <required>.*$`,
			),
			urlFields: map[string]string{
				"host": "https://release-argus.io",
			},
			sType:   "smtp",
			changed: false,
		},
		{
			name:  "gotify - fail CreateSender",
			sType: "gotify",
			urlFields: map[string]string{
				"host":  "https://	example.com",
				"token": "bish",
			},
			errRegex: `failed to parse URL`,
			changed:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var input *Shoutrrr
			if !tc.nilShoutrrr {
				input = testShoutrrr(false, false)
				input.Type = tc.sType
				if tc.main == nil {
					tc.main = &Defaults{}
				}
				input.Main = tc.main
				input.Main.InitMaps()
				input.Options = util.CopyMap(tc.options)
				input.URLFields = util.CopyMap(tc.urlFields)
				input.Params = util.CopyMap(tc.params)
			}

			_, _ = test.AssertCheckValuesWithErrorAndChanged(
				t,
				packageName,
				tc.errRegex,
				tc.changed,
				input.CheckValues,
			)
		})
	}
}
func TestBase_CheckValues(t *testing.T) {
	tests := []struct {
		name     string
		input    *Base
		want     *Base
		id       string
		errRegex string
		changed  bool
	}{
		{
			name:     "nil Base",
			input:    (*Base)(nil),
			errRegex: `^$`,
			changed:  false,
		},
		{
			name: "valid Base",
			input: &Base{
				Type: "slack",
				Options: map[string]string{
					"delay": "10s",
				},
				Params: map[string]string{
					"color": "orange",
				},
			},
			errRegex: `^$`,
			changed:  false,
		},
		{
			name: "invalid delay option",
			input: &Base{
				Type: "slack",
				Options: map[string]string{
					"delay": "10x",
				},
			},
			errRegex: test.TrimYAML(`
				^options:
				  delay: "10x" <invalid>.*$`,
			),
			changed: false,
		},
		{
			name: "invalid param template",
			input: &Base{
				Type: "slack",
				Params: map[string]string{
					"color": "{{ invalid template }}",
				},
			},
			errRegex: test.TrimYAML(`
				^params:
					color: "{{ invalid template }}" <invalid>.*$`,
			),
			changed: false,
		},
		{
			name: "multiple errors",
			input: &Base{
				Type: "slack",
				Options: map[string]string{
					"delay": "10x",
				},
				Params: map[string]string{
					"color": "{{ invalid template }}",
				},
			},
			errRegex: test.TrimYAML(`
				^options:
					delay: "10x" <invalid>.*
				params:
					color: "{{ invalid template }}" <invalid>.*$`,
			),
			changed: false,
		},
		{
			name: "matrix - rooms, leading #",
			input: &Base{
				Type: "matrix",
				Params: map[string]string{
					"rooms": "#alias:server",
				},
			},
			errRegex: `^$`,
			want: &Base{
				Type: "matrix",
				Params: map[string]string{
					"rooms": "alias:server",
				},
			},
			changed: true,
		},
		{
			name: "matrix - rooms, leading # already urlEncoded",
			input: &Base{
				Type: "matrix",
				Params: map[string]string{
					"rooms": "%23alias:server",
				},
			},
			errRegex: `^$`,
			want: &Base{
				Type: "matrix",
				Params: map[string]string{
					"rooms": "%23alias:server",
				},
			},
			changed: false,
		},
		{
			name: "matrix - rooms, valid",
			input: &Base{
				Type: "matrix",
				Params: map[string]string{
					"rooms": "alias:server",
				},
			},
			errRegex: `^$`,
			want: &Base{
				Type: "matrix",
				Params: map[string]string{
					"rooms": "alias:server",
				},
			},
			changed: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _ = test.AssertCheckValuesWithErrorAndChanged(
				t,
				packageName,
				tc.errRegex,
				tc.changed,
				func() (error, bool) { return tc.input.CheckValues(tc.id) },
			)
		})
	}
}

func TestShoutrrrsDefaults_CheckValues(t *testing.T) {
	// GIVEN: ShoutrrrsDefaults.
	tests := []struct {
		name     string
		input    *ShoutrrrsDefaults
		errRegex string
		changed  bool
	}{
		{
			name:     "nil map",
			input:    (*ShoutrrrsDefaults)(nil),
			errRegex: `^$`,
			changed:  false,
		},
		{
			name:     "valid map",
			errRegex: `^$`,
			input: &ShoutrrrsDefaults{
				"valid": testDefaults(false, false),
				"other": testDefaults(false, false),
			},
			changed: false,
		},
		{
			name:     "valid map, with changed",
			errRegex: `^$`,
			input: &ShoutrrrsDefaults{
				"valid": testDefaults(false, false),
				"other": testDefaults(false, false),
				"generic": &Defaults{
					Base: Base{
						URLFields: map[string]string{
							"host":           "example.com",
							"custom_headers": `{"foo":"bar"}`,
						},
					},
				},
			},
			changed: true,
		},
		{
			name:     "invalid type",
			errRegex: "", // Caught by Shoutrrr.CheckValues.
			input: &ShoutrrrsDefaults{
				"valid": testDefaults(false, false),
				"other": NewDefaults(
					"somethingUnknown",
					make(map[string]string),
					make(map[string]string),
					make(map[string]string),
				),
			},
			changed: false,
		},
		{
			name:     "delay without unit",
			errRegex: `^$`,
			input: &ShoutrrrsDefaults{
				"foo": NewDefaults(
					"gotify",
					map[string]string{
						"delay": "1",
					},
					nil,
					nil,
				),
			},
			changed: false,
		},
		{
			name: "invalid delay",
			errRegex: test.TrimYAML(`
				^foo:
					options:
						delay: "1x" <invalid>`,
			),
			input: &ShoutrrrsDefaults{
				"foo": NewDefaults(
					"gotify",
					map[string]string{
						"delay": "1x",
					},
					nil,
					nil,
				),
			},
			changed: false,
		},
		{
			name: "invalid message template",
			errRegex: test.TrimYAML(`
				^bar:
					options:
						message: "[^"]+" <invalid>.*$`,
			),
			input: &ShoutrrrsDefaults{
				"bar": NewDefaults(
					"gotify",
					map[string]string{
						"message": "{{ .foo }",
					},
					nil,
					nil,
				),
			},
			changed: false,
		},
		{
			name: "invalid params template",
			errRegex: test.TrimYAML(`
				^bar:
					params:
						title: "[^"]+" <invalid>.*$`,
			),
			input: &ShoutrrrsDefaults{
				"bar": NewDefaults(
					"gotify",
					nil,
					nil,
					map[string]string{
						"title": "{{ .bar }",
					},
				),
			},
			changed: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _ = test.AssertCheckValuesWithErrorAndChanged(
				t,
				packageName,
				tc.errRegex,
				tc.changed,
				tc.input.CheckValues,
			)
		})
	}
}
func TestDefaults_CheckValues(t *testing.T) {
	// GIVEN: Defaults and various ids.
	tests := []struct {
		name     string
		input    *Defaults
		id       string
		errRegex string
		changed  bool
	}{
		{
			name:     "nil defaults - valid id",
			input:    (*Defaults)(nil),
			id:       "slack",
			errRegex: `^$`,
			changed:  false,
		},
		{
			name:     "nil defaults - invalid id",
			input:    (*Defaults)(nil),
			id:       "argus",
			errRegex: `^type: "argus" <invalid>.*gotify.*$`,
			changed:  false,
		},
		{
			name: "empty Type uses id - valid",
			input: &Defaults{
				Base: Base{},
			},
			id:       "gotify",
			errRegex: `^$`,
			changed:  false,
		},
		{
			name: "empty Type uses id - invalid",
			input: &Defaults{
				Base: Base{},
			},
			id:       "unknown",
			errRegex: `^type: "unknown" <invalid>.*$`,
			changed:  false,
		},
		{
			name: "Type set overrides id (both valid)",
			input: &Defaults{
				Base: Base{
					Type: "gotify",
				},
			},
			id:       "slack",
			errRegex: `^$`,
			changed:  false,
		},
		{
			name: "Base error is propagated",
			input: &Defaults{
				Base: Base{
					Type: "slack",
					Options: map[string]string{
						"delay": "10x",
					},
				},
			},
			id: "",
			errRegex: test.TrimYAML(`
				^options:
					delay: "10x" <invalid>.*$`,
			),
			changed: false,
		},
		{
			name: "Combines type invalid and base error",
			input: &Defaults{
				Base: Base{
					Type: "invalid",
					Params: map[string]string{
						"color": "{{ invalid template }}",
					},
				},
			},
			id: "",
			errRegex: test.TrimYAML(`
				^type: "invalid" <invalid>.*
				params:
					color: "{{ invalid template }}" <invalid>.*$`,
			),
			changed: false,
		},
		{
			name: "both valid - no error",
			input: &Defaults{
				Base: Base{
					Type: "gotify",
					Options: map[string]string{
						"delay": "1s",
					},
					Params: map[string]string{
						"message": "release {{ version }}",
					},
				},
			},
			id:       "",
			errRegex: `^$`,
			changed:  false,
		},
		{
			name: "empty Type and id empty",
			input: &Defaults{
				Base: Base{},
			},
			id:       "",
			errRegex: `^type: <required>.*$`,
			changed:  false,
		},
		{
			name: "port with colon prefix is corrected",
			input: &Defaults{
				Base: Base{
					Type: "gotify",
					URLFields: map[string]string{
						"port": ":123",
					},
				},
			},
			id:       "",
			errRegex: `^$`,
			changed:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _ = test.AssertCheckValuesWithErrorAndChanged(
				t,
				packageName,
				tc.errRegex,
				tc.changed,
				func() (error, bool) {
					return tc.input.CheckValues(tc.id)
				},
			)
		})
	}
}

func TestShoutrrrsDefaults_Print(t *testing.T) {
	// GIVEN: a ShoutrrrsDefaults.
	testValid := testDefaults(false, false)
	testInvalid := testDefaults(true, true)
	tests := []struct {
		name              string
		shoutrrrsDefaults *ShoutrrrsDefaults
		want              string
	}{
		{
			name:              "nil",
			shoutrrrsDefaults: nil,
			want:              "",
		},
		{
			name:              "empty",
			shoutrrrsDefaults: &ShoutrrrsDefaults{},
			want:              "",
		},
		{
			name: "single empty element map",
			shoutrrrsDefaults: &ShoutrrrsDefaults{
				"single": {},
			},
			want: test.TrimYAML(`
				notify:
					single: {}`,
			),
		},
		{
			name: "single element map",
			shoutrrrsDefaults: &ShoutrrrsDefaults{
				"single": testValid,
			},
			want: test.TrimYAML(`
				notify:
					single:
						type: gotify
						options:
							max_tries: '` + testValid.getOption("max_tries") + `'
						url_fields:
							host: ` + testValid.getURLField("host") + `
							path: ` + testValid.getURLField("path") + `
							token: ` + testValid.getURLField("token"),
			),
		},
		{
			name: "multiple element map",
			shoutrrrsDefaults: &ShoutrrrsDefaults{
				"first":  testValid,
				"second": testInvalid,
			},
			want: test.TrimYAML(`
				notify:
					first:
						type: gotify
						options:
							max_tries: '` + testValid.getOption("max_tries") + `'
						url_fields:
							host: ` + testValid.getURLField("host") + `
							path: ` + testValid.getURLField("path") + `
							token: ` + testValid.getURLField("token") + `
					second:
						type: gotify
						options:
							max_tries: '` + testInvalid.getOption("max_tries") + `'
						url_fields:
							host: ` + testInvalid.getURLField("host") + `
							path: ` + testInvalid.getURLField("path") + `
							token: ` + testInvalid.getURLField("token"),
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout(t)

			if tc.want != "" {
				tc.want += "\n"
			}

			// WHEN: Print is called.
			tc.shoutrrrsDefaults.Print("")

			// THEN: it prints the expected stdout.
			stdout := releaseStdout()
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if stdout != tc.want {
				t.Errorf(
					"%s\nShoutrrrsDefaults.Print() stdout mismatch\ngot:  %q\nwant: %q",
					packageName, stdout, tc.want,
				)
			}
		})
	}
}

func TestBase_CorrectSelf(t *testing.T) {
	// GIVEN: a Service.
	tests := []struct {
		name          string
		sType         string
		mapTarget     string
		startAs, want map[string]string
		renamedVar    bool
	}{
		{
			name:      "port - leading colon",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"port": ":8080",
			},
			want: map[string]string{
				"port": "8080",
			},
		},
		{
			name:      "port - valid",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"port": "8080",
			},
			want: map[string]string{
				"port": "8080",
			},
		},
		{
			name:      "path - leading slash",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"path": "/argus",
			},
			want: map[string]string{
				"path": "argus",
			},
		},
		{
			name:      "path - valid",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"path": "argus",
			},
			want: map[string]string{
				"path": "argus",
			},
		},
		{
			name:      "port - from url",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"host": "https://mattermost.example.com:8443", "port": "",
			},
			want: map[string]string{
				"host": "mattermost.example.com", "port": "8443",
			},
		},
		{
			name:      "generic - custom_headers -> headers",
			sType:     "generic",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"custom_headers": `{"foo":"bar"}`,
			},
			want: map[string]string{
				"headers": `{"foo":"bar"}`,
			},
			renamedVar: true,
		},
		{
			name:      "generic - custom_headers -> headers (but headers already defined)",
			sType:     "generic",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"custom_headers": `{"foo":"bar"}`,
				"headers":        `{"foo":"baz"}`,
			},
			want: map[string]string{
				"headers": `{"foo":"baz"}`,
			},
			renamedVar: true,
		},
		{
			name:      "matrix - rooms, leading #",
			sType:     "matrix",
			mapTarget: "params",
			startAs: map[string]string{
				"rooms": "#alias:server",
			},
			want: map[string]string{
				"rooms": "alias:server",
			},
		},
		{
			name:      "matrix - rooms, leading # already urlEncoded",
			sType:     "matrix",
			mapTarget: "params",
			startAs: map[string]string{
				"rooms": "%23alias:server",
			},
			want: map[string]string{
				"rooms": "%23alias:server",
			},
		},
		{
			name:      "matrix - rooms, valid",
			sType:     "matrix",
			mapTarget: "params",
			startAs: map[string]string{
				"rooms": "alias:server",
			},
			want: map[string]string{
				"rooms": "alias:server",
			},
		},
		{
			name:      "mattermost - channel, leading slash",
			sType:     "mattermost",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"channel": "/argus",
			},
			want: map[string]string{
				"channel": "argus",
			},
		},
		{
			name:      "mattermost - channel, valid",
			sType:     "mattermost",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"channel": "argus",
			},
			want: map[string]string{
				"channel": "argus",
			},
		},
		{
			name:      "ntfy - disabletls -> disabletlsverification",
			sType:     "ntfy",
			mapTarget: "params",
			startAs: map[string]string{
				"disabletls": "true",
			},
			want: map[string]string{
				"disabletlsverification": "true",
			},
			renamedVar: true,
		},
		{
			name:      "ntfy - disabletls -> disabletlsverification (but disabletlsverification already defined)",
			sType:     "ntfy",
			mapTarget: "params",
			startAs: map[string]string{
				"disabletls": "true", "disabletlsverification": "false",
			},
			want: map[string]string{
				"disabletlsverification": "false",
			},
			renamedVar: true,
		},
		{
			name:      "slack - color, not urlEncoded",
			sType:     "slack",
			mapTarget: "params",
			startAs: map[string]string{
				"color": "#ffffff",
			},
			want: map[string]string{
				"color": "%23ffffff",
			},
		},
		{
			name:      "slack - color, valid",
			sType:     "slack",
			mapTarget: "params",
			startAs: map[string]string{
				"color": "%23ffffff",
			},
			want: map[string]string{
				"color": "%23ffffff",
			},
		},
		{
			name:      "teams - altid, leading slash",
			sType:     "teams",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"altid": "/argus",
			},
			want: map[string]string{
				"altid": "argus",
			},
		},
		{
			name:      "teams - altid, valid",
			sType:     "teams",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"altid": "argus",
			},
			want: map[string]string{
				"altid": "argus",
			},
		},
		{
			name:      "teams - groupowner, leading slash",
			sType:     "teams",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"groupowner": "/argus",
			},
			want: map[string]string{
				"groupowner": "argus",
			},
		},
		{
			name:      "teams - groupowner, valid",
			sType:     "teams",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"groupowner": "argus",
			},
			want: map[string]string{
				"groupowner": "argus",
			},
		},
		{
			name:      "zulip - botmail, not urlEncoded",
			sType:     "zulip",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"botmail": "foo@bar.com",
			},
			want: map[string]string{
				"botmail": "foo%40bar.com",
			},
		},
		{
			name:      "zulip - botmail, valid",
			sType:     "zulip",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"botmail": "foo%40bar.com",
			},
			want: map[string]string{
				"botmail": "foo%40bar.com",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := New(
				nil, "",
				tc.sType,
				make(map[string]string),
				make(map[string]string),
				make(map[string]string),
				NewDefaults(
					"",
					make(map[string]string),
					make(map[string]string),
					make(map[string]string),
				),
				NewDefaults(
					"",
					make(map[string]string),
					make(map[string]string),
					make(map[string]string),
				),
				NewDefaults(
					"",
					make(map[string]string),
					make(map[string]string),
					make(map[string]string),
				),
			)
			serviceStatus := status.Status{}
			serviceStatus.Init(
				0, 1, 0,
				status.ServiceInfo{
					ID: tc.name,
				},
				&dashboard.Options{},
			)
			shoutrrr.Init(
				&serviceStatus,
				shoutrrr.Main, shoutrrr.Defaults, shoutrrr.HardDefaults,
			)
			var subTestMap = map[string]struct {
				defaults  *Defaults
				URLFields map[string]string
				Params    map[string]string
			}{
				"root": {
					URLFields: shoutrrr.URLFields,
					Params:    shoutrrr.Params,
				},
				"main": {
					URLFields: shoutrrr.Main.URLFields,
					Params:    shoutrrr.Main.Params,
				},
				"defaults": {
					URLFields: shoutrrr.Defaults.URLFields,
					Params:    shoutrrr.Defaults.Params,
				},
				"hard_defaults": {
					URLFields: shoutrrr.HardDefaults.URLFields,
					Params:    shoutrrr.HardDefaults.Params,
				},
			}
			// Sub tests - set in different locations and check it's corrected there.
			for subTest := range subTestMap {
				t.Logf(
					"%s - sub_test: %s",
					packageName, subTest,
				)
				if tc.mapTarget == "url_fields" {
					for k, v := range tc.startAs {
						subTestMap[subTest].URLFields[k] = v
					}
				} else {
					for k, v := range tc.startAs {
						subTestMap[subTest].Params[k] = v
					}
				}

				// WHEN: correctSelf is called.
				typ := shoutrrr.GetType()
				shoutrrr.correctSelf(typ)

				for k, v := range tc.want {
					want := v
					// root is the only one that gets corrected.
					if subTest != "root" {
						want = tc.startAs[k]
					}
					got := shoutrrr.GetURLField(k)
					if tc.mapTarget != "url_fields" {
						got = shoutrrr.GetParam(k)
					}
					if got != want {
						t.Errorf(
							"%s\nBase.correctSelf(%q) %s[%q] mismatch\ngot:  %q\nwant: %q",
							packageName, typ, tc.mapTarget, k,
							got, want,
						)
					}
					// Reset.
					if tc.mapTarget == "url_fields" {
						delete(subTestMap[subTest].URLFields, k)
					} else {
						delete(subTestMap[subTest].Params, k)
					}
				}
			}
		})
	}
}

func TestBase_NormaliseParamSelect(t *testing.T) {
	// GIVEN: a Base and various inputs to normaliseParamSelect.
	tests := []struct {
		name       string
		value      string
		allowed    []string
		startParam string
		ok         bool
		wantValue  string
	}{
		{
			name:      "exact match uses canonical case",
			value:     "Two",
			allowed:   []string{"One", "Two", "Three"},
			ok:        true,
			wantValue: "Two",
		},
		{
			name:      "case-insensitive match lower->upper",
			value:     "two",
			allowed:   []string{"One", "Two", "Three"},
			ok:        true,
			wantValue: "Two",
		},
		{
			name:      "case-insensitive match upper->proper",
			value:     "THREE",
			allowed:   []string{"One", "Two", "Three"},
			ok:        true,
			wantValue: "Three",
		},
		{
			name:       "non-match returns false and leaves value unchanged",
			value:      "four",
			allowed:    []string{"One", "Two", "Three"},
			startParam: "unchanged",
			ok:         false,
			wantValue:  "unchanged",
		},
		{
			name:      "empty value returns false and makes no change",
			value:     "",
			allowed:   []string{"One", "Two"},
			ok:        false,
			wantValue: "",
		},
		{
			name:       "empty allowed set never matches",
			value:      "one",
			allowed:    []string{},
			startParam: "keepme",
			ok:         false,
			wantValue:  "keepme",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b := &Base{}
			b.InitMaps()
			if tc.startParam != "" {
				b.setParam(tc.name, tc.startParam)
			}

			resultChannel := make(chan bool, 1)
			// WHEN: normaliseParamSelect is called.
			resultChannel <- b.normaliseParamSelect(tc.name, tc.value, tc.allowed)

			prefix := fmt.Sprintf("%s\nnormaliseParamSelect()", packageName)

			// THEN: it returns the expected boolean.
			if err := test.AssertChannelBool(
				t,
				tc.ok,
				resultChannel,
				nil,
				nil,
			); err != nil {
				t.Fatal(prefix + err.Error())
			}

			// AND: Params[key] is set/unchanged as expected.
			if got := b.GetParam(tc.name); got != tc.wantValue {
				t.Fatalf(
					"%s Params[x] mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wantValue,
				)
			}
		})
	}
}

func TestShoutrrr_CheckValuesType(t *testing.T) {
	// GIVEN: a Shoutrrr with a Type and possibly a Main.
	tests := []struct {
		name     string
		sType    string
		main     *Defaults
		errRegex string
		changed  bool
	}{
		{
			name:     "no type",
			errRegex: `^type: <required>.*$`,
			sType:    "",
		},
		{
			name:     "invalid type",
			errRegex: `^type: .* <invalid>.*$`,
			sType:    "argus",
		},
		{
			name:     "invalid type - type in main differs",
			errRegex: `type: "gotify" <invalid> .*\(discord\)\)`,
			sType:    "gotify",
			main: NewDefaults(
				"discord",
				make(map[string]string),
				make(map[string]string),
				make(map[string]string),
			),
		},
		{
			name:     "valid type",
			errRegex: `^$`,
			sType:    "gotify",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := testShoutrrr(false, false)
			input.Type = tc.sType
			svcStatus, _ := statustest.New("yaml", nil)
			svcStatus.Init(
				0, 1, 0,
				status.ServiceInfo{
					ID: tc.name,
				},
				&dashboard.Options{},
			)
			input.Init(
				svcStatus,
				tc.main,
				&Defaults{}, &Defaults{},
			)

			_, _ = test.AssertCheckValuesWithErrorAndChanged(
				t,
				packageName,
				tc.errRegex,
				tc.changed,
				input.CheckValues,
			)
		})
	}
}

func TestBase_CheckValuesOptions(t *testing.T) {
	// GIVEN: a Base with Options.
	tests := []struct {
		name      string
		options   map[string]string
		wantDelay string
		errRegex  string
	}{
		{
			name:     "no options",
			options:  map[string]string{},
			errRegex: `^$`,
		},
		{
			name: "valid delay",
			options: map[string]string{
				"delay": "5s",
			},
			wantDelay: "5s",
			errRegex:  `^$`,
		},
		{
			name: "invalid delay",
			options: map[string]string{
				"delay": "5x",
			},
			errRegex: `^delay: "5x" <invalid>.*$`,
		},
		{
			name: "fixes delay missing unit",
			options: map[string]string{
				"delay": "5",
			},
			wantDelay: "5s",
			errRegex:  `^$`,
		},
		{
			name: "valid message template",
			options: map[string]string{
				"message": "release! {{ version }}",
			},
			errRegex: `^$`,
		},
		{
			name: "invalid message template",
			options: map[string]string{
				"message": "release! {{ version }",
			},
			errRegex: `^message: "release! {{ version }" <invalid>.*$`,
		},
		{
			name: "min max_tries",
			options: map[string]string{
				"max_tries": "0",
			},
			errRegex: `^$`,
		},
		{
			name: "max max_tries",
			options: map[string]string{
				"max_tries": strconv.Itoa(math.MaxUint8),
			},
			errRegex: `^$`,
		},
		{
			name: "invalid max_tries - too large",
			options: map[string]string{
				"max_tries": strconv.Itoa(math.MaxUint16),
			},
			errRegex: `^max_tries: "\d+" <invalid>.*$`,
		},
		{
			name: "invalid max_tries - too large, >uint64",
			options: map[string]string{
				"max_tries": fmt.Sprintf("1%d", uint(math.MaxUint64)),
			},
			errRegex: `^max_tries: "\d+" <invalid>.*$`,
		},
		{
			name: "invalid max_tries - not a number",
			options: map[string]string{
				"max_tries": "oneOrTwo",
			},
			errRegex: `^max_tries: "oneOrTwo" <invalid>.*$`,
		},
		{
			name: "invalid max_tries - negative",
			options: map[string]string{
				"max_tries": "-1",
			},
			errRegex: `^max_tries: "-1" <invalid>.*$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := Base{
				Options: tc.options,
			}

			err := test.AssertCheckValuesWithError(
				t,
				packageName,
				tc.errRegex,
				input.checkValuesOptions,
			)

			// AND: the delay is set as expected if it didn't error on delay.
			e := errfmt.FormatError(err)
			if !util.RegexCheck("^delay: .*", e) {
				if got := input.getOption("delay"); got != tc.wantDelay {
					t.Errorf(
						"%s\nBase.Option.delay mismatch after checkValuesOptions()\ngot:  %q\nwant: %q",
						packageName, got, tc.wantDelay,
					)
				}
			}
		})
	}
}

func TestShoutrrr_CheckValuesURLFields(t *testing.T) {
	// GIVEN: a Shoutrrr with Params.
	tests := []struct {
		name      string
		sType     string
		urlFields map[string]string
		main      *Defaults
		errRegex  string
	}{
		{
			name:  "bark - invalid",
			sType: "bark",
			errRegex: test.TrimYAML(`
				^devicekey: <required>.*
				host: <required>.*$`,
			),
		},
		{
			name:  "bark - no devicekey",
			sType: "bark",
			urlFields: map[string]string{
				"host": "https://example.com",
			},
			errRegex: `^devicekey: <required>.*$`,
		},
		{
			name:  "bark - no host",
			sType: "bark",
			urlFields: map[string]string{
				"devicekey": "foo",
			},
			errRegex: `^host: <required>.*$`,
		},
		{
			name:  "bark - valid",
			sType: "bark",
			urlFields: map[string]string{
				"devicekey": "foo",
				"host":      "https://example.com",
			},
			errRegex: `^$`,
		},
		{
			name:  "discord - invalid",
			sType: "discord",
			errRegex: test.TrimYAML(`
				^token: <required>.*
				webhookid: <required>.*$`,
			),
		},
		{
			name:  "discord - no token",
			sType: "discord",
			urlFields: map[string]string{
				"webhookid": "bash",
			},
			errRegex: `^token: <required>.*$`,
		},
		{
			name:  "discord - no webhookid",
			sType: "discord",
			urlFields: map[string]string{
				"token": "bish",
			},
			errRegex: `^webhookid: <required>.*$`,
		},
		{
			name:  "discord - valid",
			sType: "discord",
			urlFields: map[string]string{
				"token":     "bish",
				"webhookid": "webhookid",
			},
			errRegex: `^$`,
		},
		{
			name:  "discord - valid with main",
			sType: "discord",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"token":     "bish",
					"webhookid": "bash",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:      "smtp - no host",
			sType:     "smtp",
			urlFields: map[string]string{},
			errRegex:  `^host: <required>.*$`,
		},
		{
			name:  "smtp - valid",
			sType: "smtp",
			urlFields: map[string]string{
				"host": "bish",
			},
			errRegex: `^$`,
		},
		{
			name:  "smtp - valid with main",
			sType: "smtp",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"host": "bish",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:  "gotify - invalid",
			sType: "gotify",
			errRegex: test.TrimYAML(`
				^host: <required>.*
				token: <required>.*$`,
			),
		},
		{
			name:  "gotify - no host",
			sType: "gotify",
			urlFields: map[string]string{
				"token": "bash",
			},
			errRegex: `^host: <required>.*$`,
		},
		{
			name:  "gotify - no token",
			sType: "gotify",
			urlFields: map[string]string{
				"host": "bish",
			},
			errRegex: `^token: <required>.*$`,
		},
		{
			name:  "gotify - valid",
			sType: "gotify",
			urlFields: map[string]string{
				"host":  "bish",
				"token": "bash",
			},
			errRegex: `^$`,
		},
		{
			name:  "gotify - valid with main",
			sType: "gotify",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"host":  "bish",
					"token": "bash",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:     "googlechat - invalid",
			sType:    "googlechat",
			errRegex: `^raw: <required>.*$`,
		},
		{
			name:  "googlechat - valid",
			sType: "googlechat",
			urlFields: map[string]string{
				"raw": "bish",
			},
			errRegex: `^$`,
		},
		{
			name:  "googlechat - valid with main",
			sType: "googlechat",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"raw": "bish",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:      "ifttt - no webhookid",
			sType:     "ifttt",
			urlFields: map[string]string{},
			errRegex:  `^webhookid: <required>.*$`,
		},
		{
			name:  "ifttt - valid",
			sType: "ifttt",
			urlFields: map[string]string{
				"webhookid": "bish",
			},
			errRegex: `^$`,
		},
		{
			name:  "ifttt - valid with main",
			sType: "ifttt",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"webhookid": "webhookid",
				},
				map[string]string{
					"events": "events",
				},
			),
			errRegex: `^$`,
		},
		{
			name:      "join - no apikey",
			sType:     "join",
			urlFields: map[string]string{},
			errRegex:  `^apikey: <required>.*$`,
		},
		{
			name:  "join - valid",
			sType: "join",
			urlFields: map[string]string{
				"apikey": "bish",
			},
			errRegex: `^$`,
		},
		{
			name:  "join - valid with main",
			sType: "join",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"apikey": "apikey",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:  "mattermost - invalid",
			sType: "mattermost",
			errRegex: test.TrimYAML(`
				^host: <required>.*
				token: <required>.*$`,
			),
		},
		{
			name:  "mattermost - no host",
			sType: "mattermost",
			urlFields: map[string]string{
				"token": "bash",
			},
			errRegex: `^host: <required>.*$`,
		},
		{
			name:  "mattermost - no token",
			sType: "mattermost",
			urlFields: map[string]string{
				"host": "bish",
			},
			errRegex: `^token: <required>.*$`,
		},
		{
			name:  "mattermost - valid",
			sType: "mattermost",
			urlFields: map[string]string{
				"host":  "bish",
				"token": "bash",
			},
			errRegex: `^$`,
		},
		{
			name:  "mattermost - valid with main",
			sType: "mattermost",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"host":  "bish",
					"token": "bash",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:  "matrix - invalid",
			sType: "matrix",
			errRegex: test.TrimYAML(`
				^host: <required>.*
				password: <required>.*$`,
			),
		},
		{
			name:  "matrix - no host",
			sType: "matrix",
			urlFields: map[string]string{
				"password": "bash",
			},
			errRegex: `^host: <required>.*$`,
		},
		{
			name:  "matrix - no password",
			sType: "matrix",
			urlFields: map[string]string{
				"host": "bish",
			},
			errRegex: `^password: <required>.*$`,
		},
		{
			name:  "matrix - valid",
			sType: "matrix",
			urlFields: map[string]string{
				"host":     "bish",
				"password": "password",
			},
			errRegex: `^$`,
		},
		{
			name:  "matrix - valid with main",
			sType: "matrix",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"host":     "bish",
					"password": "bash",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:     "ntfy - invalid",
			sType:    "ntfy",
			errRegex: `^topic: <required>.*$`,
		},
		{
			name:  "ntfy - valid",
			sType: "ntfy",
			urlFields: map[string]string{
				"topic": "foo",
			},
			errRegex: `^$`,
		},
		{
			name:     "opsgenie - invalid",
			sType:    "opsgenie",
			errRegex: `^apikey: <required>.*$`,
		},
		{
			name:  "opsgenie - valid",
			sType: "opsgenie",
			urlFields: map[string]string{
				"apikey": "apikey",
			},
			errRegex: `^$`,
		},
		{
			name:  "opsgenie - valid with main",
			sType: "opsgenie",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"apikey": "apikey",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:  "pushbullet - invalid",
			sType: "pushbullet",
			errRegex: test.TrimYAML(`
				^token: <required>.*
				targets: <required>.*$`,
			),
		},
		{
			name:  "pushbullet - no token",
			sType: "pushbullet",
			urlFields: map[string]string{
				"targets": "bash",
			},
			errRegex: `^token: <required>.*$`,
		},
		{
			name:  "pushbullet - no targets",
			sType: "pushbullet",
			urlFields: map[string]string{
				"token": "bish",
			},
			errRegex: `^targets: <required>.*$`,
		},
		{
			name:  "pushbullet - valid",
			sType: "pushbullet",
			urlFields: map[string]string{
				"token":   "bish",
				"targets": "targets",
			},
			errRegex: `^$`,
		},
		{
			name:  "pushbullet - valid with main",
			sType: "pushbullet",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"token":   "bish",
					"targets": "bash",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:  "pushover - invalid",
			sType: "pushover",
			errRegex: test.TrimYAML(`
				^token: <required>.*
				user: <required>.*$`,
			),
		},
		{
			name:  "pushover - no token",
			sType: "pushover",
			urlFields: map[string]string{
				"user": "bash",
			},
			errRegex: `^token: <required>.*$`,
		},
		{
			name:  "pushover - no user",
			sType: "pushover",
			urlFields: map[string]string{
				"token": "bish",
			},
			errRegex: `^user: <required>.*$`,
		},
		{
			name:  "pushover - valid",
			sType: "pushover",
			urlFields: map[string]string{
				"token": "bish",
				"user":  "user",
			},
			errRegex: `^$`,
		},
		{
			name:  "pushover - valid with main",
			sType: "pushover",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"token": "bish",
					"user":  "bash",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:  "rocketchat - invalid",
			sType: "rocketchat",
			errRegex: test.TrimYAML(`
				^host: <required>.*
				tokena: <required>.*
				tokenb: <required>.*
				channel: <required>.*$`,
			),
		},
		{
			name:  "rocketchat - no host",
			sType: "rocketchat",
			urlFields: map[string]string{
				"tokena":  "bash",
				"tokenb":  "bosh",
				"channel": "bing",
			},
			errRegex: `^host: <required>.*$`,
		},
		{
			name:  "rocketchat - no tokena",
			sType: "rocketchat",
			urlFields: map[string]string{
				"host":    "bish",
				"tokenb":  "bash",
				"channel": "bing",
			},
			errRegex: `^tokena: <required>.*$`,
		},
		{
			name:  "rocketchat - no tokenb",
			sType: "rocketchat",
			urlFields: map[string]string{
				"host":    "bish",
				"tokena":  "bash",
				"channel": "bing",
			},
			errRegex: `^tokenb: <required>.*$`,
		},
		{
			name:  "rocketchat - no channel",
			sType: "rocketchat",
			urlFields: map[string]string{
				"host":   "bish",
				"tokena": "bash",
				"tokenb": "bosh",
			},
			errRegex: `^channel: <required>.*$`,
		},
		{
			name:  "rocketchat - valid",
			sType: "rocketchat",
			urlFields: map[string]string{
				"host":    "bish",
				"tokena":  "bash",
				"tokenb":  "bosh",
				"channel": "bing",
			},
			errRegex: `^$`,
		},
		{
			name:  "rocketchat - valid with main",
			sType: "rocketchat",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"host":    "bish",
					"tokena":  "bash",
					"tokenb":  "bosh",
					"channel": "bing",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:  "slack - invalid",
			sType: "slack",
			errRegex: test.TrimYAML(`
				^token: <required>.*
				channel: <required>.*$`,
			),
		},
		{
			name:  "slack - no token",
			sType: "slack",
			urlFields: map[string]string{
				"channel": "bash",
			},
			errRegex: `^token: <required>.*$`,
		},
		{
			name:  "slack - no channel",
			sType: "slack",
			urlFields: map[string]string{
				"token": "bish",
			},
			errRegex: `^channel: <required>.*$`,
		},
		{
			name:  "slack - valid",
			sType: "slack",
			urlFields: map[string]string{
				"token":   "bish",
				"channel": "channel",
			},
			errRegex: `^$`,
		},
		{
			name:  "slack - valid with main",
			sType: "slack",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"token":   "bish",
					"channel": "bash",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:  "teams - invalid",
			sType: "teams",
			errRegex: test.TrimYAML(`
				^group: <required>.*
				tenant: <required>.*
				altid: <required>.*
				groupowner: <required>.*
				extraid: <required>.*$`,
			),
		},
		{
			name:  "teams - no group",
			sType: "teams",
			urlFields: map[string]string{
				"tenant":     "bash",
				"altid":      "bosh",
				"groupowner": "bing",
				"extraid":    "boop",
			},
			errRegex: `^group: <required>.*$`,
		},
		{
			name:  "teams - no tenant",
			sType: "teams",
			urlFields: map[string]string{
				"group":      "bish",
				"altid":      "bash",
				"groupowner": "bing",
				"extraid":    "boop",
			},
			errRegex: `^tenant: <required>.*$`,
		},
		{
			name:  "teams - no altid",
			sType: "teams",
			urlFields: map[string]string{
				"group":      "bish",
				"tenant":     "bash",
				"groupowner": "bing",
				"extraid":    "boop",
			},
			errRegex: `^altid: <required>.*$`,
		},
		{
			name:  "teams - no groupowner",
			sType: "teams",
			urlFields: map[string]string{
				"group":   "bish",
				"tenant":  "bash",
				"altid":   "bosh",
				"extraid": "boop",
			},
			errRegex: `^groupowner: <required>.*$`,
		},
		{
			name:  "teams - no extraid",
			sType: "teams",
			urlFields: map[string]string{
				"group":      "bish",
				"tenant":     "bash",
				"altid":      "bosh",
				"groupowner": "bing",
			},
			errRegex: `^extraid: <required>.*$`,
		},
		{
			name:  "teams - valid",
			sType: "teams",
			urlFields: map[string]string{
				"group":      "bish",
				"tenant":     "bash",
				"altid":      "bosh",
				"groupowner": "bing",
				"extraid":    "boop",
			},
			errRegex: `^$`,
		},
		{
			name:  "teams - valid with main",
			sType: "teams",
			main: NewDefaults(
				"",
				map[string]string{
					"host": "https://release-argus.io",
				},
				map[string]string{
					"group":      "bish",
					"tenant":     "bash",
					"altid":      "bosh",
					"groupowner": "bing",
					"extraid":    "boop",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:      "telegram - no token",
			sType:     "telegram",
			urlFields: map[string]string{},
			errRegex:  `^token: <required>.*$`,
		},
		{
			name:  "telegram - valid",
			sType: "telegram",
			urlFields: map[string]string{
				"token": "bish",
			},
			errRegex: `^$`,
		},
		{
			name:  "telegram - valid with main",
			sType: "telegram",
			main: NewDefaults(
				"",
				nil,
				map[string]string{
					"token": "bish",
				},
				map[string]string{
					"chats": "chats",
				},
			),
			errRegex: `^$`,
		},
		{
			name:  "zulip - invalid",
			sType: "zulip",
			errRegex: test.TrimYAML(`
				^host: <required>.*
				botmail: <required>.*
				botkey: <required>.*$`,
			),
		},
		{
			name:  "zulip - no host",
			sType: "zulip",
			urlFields: map[string]string{
				"botmail": "bash",
				"botkey":  "bosh",
			},
			errRegex: `^host: <required>.*$`,
		},
		{
			name:  "zulip - no botmail",
			sType: "zulip",
			urlFields: map[string]string{
				"host":   "bish",
				"botkey": "bash",
			},
			errRegex: `^botmail: <required>.*$`,
		},
		{
			name:  "zulip - no botkey",
			sType: "zulip",
			urlFields: map[string]string{
				"host":    "bish",
				"botmail": "bash",
			},
			errRegex: `^botkey: <required>.*$`,
		},
		{
			name:  "zulip - valid",
			sType: "zulip",
			urlFields: map[string]string{
				"host":    "bish",
				"botmail": "bash",
				"botkey":  "bosh",
			},
			errRegex: `^$`,
		},
		{
			name:  "zulip - valid with main",
			sType: "zulip",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"host":    "bish",
					"botmail": "bash",
					"botkey":  "bosh",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:     "generic - invalid",
			sType:    "generic",
			errRegex: `^host: <required>.*$`,
		},
		{
			name:  "generic - valid",
			sType: "generic",
			urlFields: map[string]string{
				"host": "example.com",
			},
			errRegex: `^$`,
		},
		{
			name:  "generic - valid with main",
			sType: "generic",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"host": "example.com",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:  "generic - valid with headers/json_payload_vars/query_vars",
			sType: "generic",
			urlFields: map[string]string{
				"host":              "example.com",
				"headers":           `{"foo":"bar"}`,
				"json_payload_vars": `{"bish":"bash","bosh":"bing"}`,
				"query_vars":        `{"me":"work"}`,
			},
			errRegex: `^$`,
		},
		{
			name:  "generic - invalid headers",
			sType: "generic",
			urlFields: map[string]string{
				"host":    "example.com",
				"headers": `"foo":"bar"}`,
			},
			errRegex: `^headers: "\\\"foo\\\":\\\"bar\\\"}" <invalid>.*$`,
		},
		{
			name:  "generic - invalid json_payload_vars",
			sType: "generic",
			urlFields: map[string]string{
				"host":              "example.com",
				"json_payload_vars": `{foo":"bar`,
			},
			errRegex: `^json_payload_vars: "{foo\\\":\\\"bar" <invalid>.*$`,
		},
		{
			name:  "generic - invalid query_vars",
			sType: "generic",
			urlFields: map[string]string{
				"host":       "example.com",
				"query_vars": `{foo:bar}`,
			},
			errRegex: `^query_vars: "{foo:bar}" <invalid>.*$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := testShoutrrr(false, false)
			input.Type = tc.sType
			input.URLFields = tc.urlFields
			svcStatus, _ := statustest.New("yaml", nil)
			svcStatus.Init(
				0, 1, 0,
				status.ServiceInfo{
					ID: tc.name,
				},
				svcStatus.Dashboard,
			)
			input.Init(
				svcStatus,
				tc.main,
				&Defaults{}, &Defaults{},
			)

			_ = test.AssertCheckValuesWithError(
				t,
				packageName,
				tc.errRegex,
				input.checkValuesURLFields,
			)
		})
	}
}

func TestShoutrrr_CheckValuesParams(t *testing.T) {
	// GIVEN: a Shoutrrr with Params.
	tests := []struct {
		name     string
		sType    string
		params   map[string]string
		main     *Defaults
		errRegex string
	}{
		{
			name:     "no params",
			params:   map[string]string{},
			errRegex: `^$`,
		},
		{
			name: "valid message template",
			params: map[string]string{
				"message": "release! {{ version }}",
			},
			errRegex: `^$`,
		},
		{
			name: "invalid message template",
			params: map[string]string{
				"a": "release! {{ version }",
			},
			errRegex: `^a: "release! {{ version }" <invalid>.*$`,
		},
		{
			name:  "smtp - invalid",
			sType: "smtp",
			errRegex: test.TrimYAML(`
				^fromaddress: <required>.*
				toaddresses: <required>.*$`,
			),
		},
		{
			name:  "smtp - no fromaddress",
			sType: "smtp",
			params: map[string]string{
				"toaddresses": "bosh",
			},
			errRegex: `^fromaddress: <required>.*$`,
		},
		{
			name:  "smtp - no toaddresses",
			sType: "smtp",
			params: map[string]string{
				"fromaddress": "bash",
			},
			errRegex: `^toaddresses: <required>.*$`,
		},
		{
			name:  "smtp - valid",
			sType: "smtp",
			params: map[string]string{
				"fromaddress": "bash",
				"toaddresses": "bosh",
			},
			errRegex: `^$`,
		},
		{
			name:  "smtp - valid with main",
			sType: "smtp",
			params: map[string]string{
				"fromaddress": "bash",
			},
			main: NewDefaults(
				"",
				nil,
				nil,
				map[string]string{
					"toaddresses": "bosh",
				},
			),
			errRegex: `^$`,
		},
		{
			name:     "ifttt - no events",
			sType:    "ifttt",
			errRegex: `^events: <required>.*$`,
		},
		{
			name:  "ifttt - valid",
			sType: "ifttt",
			params: map[string]string{
				"events": "events",
			},
			errRegex: `^$`,
		},
		{
			name:  "ifttt - valid with main",
			sType: "ifttt",
			main: NewDefaults(
				"",
				nil,
				nil,
				map[string]string{
					"events": "events",
				},
			),
			errRegex: `^$`,
		},
		{
			name:     "join - no devices",
			sType:    "join",
			params:   map[string]string{},
			errRegex: `^devices: <required>.*$`,
		},
		{
			name:  "join - valid",
			sType: "join",
			params: map[string]string{
				"devices": "foo,bar",
			},
			errRegex: `^$`,
		},
		{
			name:  "join - valid with main",
			sType: "join",
			main: NewDefaults(
				"",
				nil,
				nil,
				map[string]string{
					"devices": "devices",
				},
			),
			errRegex: `^$`,
		},
		{
			name:     "teams - no host",
			sType:    "teams",
			params:   map[string]string{},
			errRegex: `^host: <required>.*$`,
		},
		{
			name:  "teams - valid",
			sType: "teams",
			params: map[string]string{
				"host": "https://release-argus.io",
			},
			errRegex: `^$`,
		},
		{
			name:  "teams - valid with main",
			sType: "teams",
			main: NewDefaults(
				"",
				nil,
				nil,
				map[string]string{
					"host": "https://release-argus.io",
				},
			),
			errRegex: `^$`,
		},
		{
			name:     "telegram - no chats",
			sType:    "telegram",
			errRegex: `^chats: <required>.*$`,
		},
		{
			name:  "telegram - valid",
			sType: "telegram",
			params: map[string]string{
				"chats": "chats",
			},
			errRegex: `^$`,
		},
		{
			name:  "telegram - valid with main",
			sType: "telegram",
			main: NewDefaults(
				"",
				nil,
				nil,
				map[string]string{
					"chats": "chats",
				},
			),
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := testShoutrrr(false, false)
			input.Type = tc.sType
			input.Params = tc.params
			svcStatus, _ := statustest.New("yaml", nil)
			svcStatus.Init(
				0, 1, 0,
				status.ServiceInfo{
					ID: tc.name,
				},
				svcStatus.Dashboard,
			)
			input.Init(
				svcStatus,
				tc.main,
				&Defaults{}, &Defaults{},
			)

			_ = test.AssertCheckValuesWithError(
				t,
				packageName,
				tc.errRegex,
				input.checkValuesParams,
			)
		})
	}
}

func TestBase_CheckValuesParams(t *testing.T) {
	// GIVEN: a Base with Params.
	tests := []struct {
		name     string
		itemType string
		params   map[string]string
		errRegex string
	}{
		{
			name:     "no params",
			params:   map[string]string{},
			errRegex: `^$`,
		},
		{
			name: "valid message template",
			params: map[string]string{
				"message": "release! {{ version }}",
			},
			errRegex: `^$`,
		},
		{
			name: "invalid message template",
			params: map[string]string{
				"a": "release! {{ version }",
			},
			errRegex: `^a: "release! {{ version }" <invalid>.*$`,
		},
		{
			name:     "valid select param",
			itemType: "smtp",
			params: map[string]string{
				"auth": "OAuth2",
			},
			errRegex: `^$`,
		},
		{
			name:     "invalid select param",
			itemType: "smtp",
			params: map[string]string{
				"auth": "-",
			},
			errRegex: `^auth: "-" <invalid>.*OAuth2.*$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := Base{
				Params: tc.params,
			}

			_ = test.AssertCheckValuesWithError(
				t,
				packageName,
				tc.errRegex,
				func() error {
					return input.checkValuesParams(tc.itemType)
				},
			)
		})
	}
}

func TestBase_CheckValuesParamsSelects(t *testing.T) {
	// GIVEN: a Base with Params and different item types.
	tests := []struct {
		name      string
		itemType  string
		params    map[string]string
		errRegex  string
		wantParam map[string]string
	}{
		// bark
		{
			name:     "bark - valid scheme+sound normalised",
			itemType: "bark",
			params: map[string]string{
				"scheme": "HTTP",
				"sound":  "BELL",
			},
			errRegex: `^$`,
			wantParam: map[string]string{
				"scheme": "http",
				"sound":  "bell",
			},
		},
		{
			name:     "bark - invalid scheme and sound aggregated",
			itemType: "bark",
			params: map[string]string{
				"scheme": "-",
				"sound":  "nope",
			},
			errRegex: "^" + test.RegexBracketEscaper.Replace(
				errfmt.FormatError(
					errors.Join(
						polymorphic.InvalidTypeError{
							Key:     "scheme",
							Value:   "-",
							Allowed: barkNtfyParamScheme,
						},
						polymorphic.InvalidTypeError{
							Key:     "sound",
							Value:   "nope",
							Allowed: barkParamSound,
						},
					),
				),
			) + "$",
		},
		// generic
		{
			name:     "generic - valid requestmethod normalised",
			itemType: "generic",
			params: map[string]string{
				"requestmethod": "post",
			},
			errRegex: `^$`,
			wantParam: map[string]string{
				"requestmethod": "POST",
			},
		},
		{
			name:     "generic - invalid requestmethod",
			itemType: "generic",
			params: map[string]string{
				"requestmethod": "FETCH",
			},
			errRegex: "^" + test.RegexBracketEscaper.Replace(
				polymorphic.InvalidTypeError{
					Key:     "requestmethod",
					Value:   "FETCH",
					Allowed: genericParamRequestmethod,
				}.Error(),
			) + "$",
		},
		// ntfy
		{
			name:     "ntfy - valid priority and scheme",
			itemType: "ntfy",
			params: map[string]string{
				"priority": "DEFAULT",
				"scheme":   "HTTPS",
			},
			errRegex: `^$`,
			wantParam: map[string]string{
				"priority": "default",
				"scheme":   "https",
			},
		},
		{
			name:     "ntfy - invalid priority and scheme aggregated",
			itemType: "ntfy",
			params: map[string]string{
				"priority": "urgENT",
				"scheme":   "ftp",
			},
			errRegex: "^" + test.RegexBracketEscaper.Replace(
				errfmt.FormatError(
					errors.Join(
						polymorphic.InvalidTypeError{
							Key:     "priority",
							Value:   "urgENT",
							Allowed: ntfyParamPriority,
						},
						polymorphic.InvalidTypeError{
							Key:     "scheme",
							Value:   "ftp",
							Allowed: barkNtfyParamScheme,
						},
					),
				),
			) + "$",
		},
		// smtp
		{
			name:     "smtp - valid auth and encryption",
			itemType: "smtp",
			params: map[string]string{
				"auth":       "oauth2",
				"encryption": "explicittls",
			},
			errRegex: `^$`,
			wantParam: map[string]string{
				"auth":       "OAuth2",
				"encryption": "ExplicitTLS",
			},
		},
		{
			name:     "smtp - invalid auth and encryption aggregated",
			itemType: "smtp",
			params: map[string]string{
				"auth":       "basic",
				"encryption": "tls1.3",
			},
			errRegex: "^" + test.RegexBracketEscaper.Replace(
				errfmt.FormatError(
					errors.Join(
						polymorphic.InvalidTypeError{
							Key:     "auth",
							Value:   "basic",
							Allowed: smtpParamAuth,
						},
						polymorphic.InvalidTypeError{
							Key:     "encryption",
							Value:   "tls1.3",
							Allowed: smtpParamEncryption,
						},
					),
				),
			) + "$",
		},
		// telegram
		{
			name:     "telegram - valid parsemode",
			itemType: "telegram",
			params: map[string]string{
				"parsemode": "markdown",
			},
			errRegex: `^$`,
			wantParam: map[string]string{
				"parsemode": "Markdown",
			},
		},
		{
			name:     "telegram - invalid parsemode",
			itemType: "telegram",
			params: map[string]string{
				"parsemode": "mdx",
			},
			errRegex: "^" + test.RegexBracketEscaper.Replace(
				polymorphic.InvalidTypeError{
					Key:     "parsemode",
					Value:   "mdx",
					Allowed: telegramParamParsemode,
				}.Error(),
			) + "$",
		},
		// unknown type and empty params
		{
			name:     "unknown - type is no-op even with values",
			itemType: "unknown",
			params: map[string]string{
				"scheme": "ftp",
			},
			errRegex:  `^$`,
			wantParam: map[string]string{},
		},
		{
			name:     "nil/empty params produces no error",
			itemType: "smtp",
			params:   map[string]string{},
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := Base{
				Params: tc.params,
			}

			_ = test.AssertCheckValuesWithError(
				t,
				packageName,
				tc.errRegex,
				func() error {
					return input.checkValuesParams(tc.itemType)
				},
			)

			// THEN: any expected normalisations took place.
			for k, v := range tc.wantParam {
				if got := input.GetParam(k); got != v {
					t.Fatalf(
						"%s\nBase checkValuesParams() normalisation mismatch for %q\ngot:  %q\nwant: %q",
						packageName, k, got, v,
					)
				}
			}
		})
	}
}

func TestBase_ValidateParamSelect(t *testing.T) {
	key := "test"
	// GIVEN: a Base and various inputs to validateParamSelect.
	tests := []struct {
		name     string
		value    string
		allowed  []string
		errRegex string
		want     string
	}{
		{
			name:     "empty value returns nil",
			value:    "",
			allowed:  []string{"low", "default", "high"},
			errRegex: `^$`,
			want:     "",
		},
		{
			name:     "valid value normalises and returns nil",
			value:    "HiGh",
			allowed:  []string{"min", "low", "default", "high", "max"},
			errRegex: `^$`,
			want:     "high",
		},
		{
			name:     "invalid value returns error and leaves unchanged",
			value:    "nope",
			allowed:  []string{"None", "Unknown", "Plain", "CramMD5", "OAuth2"},
			errRegex: "^" + key + `: "nope" <invalid> .*'CramMD5'.*$`,
			want:     "nope",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := Base{}
			input.InitMaps()
			input.setParam(key, tc.value)

			// WHEN: validateParamSelect is called.
			err := input.validateParamSelect(key, tc.allowed)

			prefix := fmt.Sprintf("%s\nBase.validateParamSelect()", packageName)

			// THEN: error matches expectation.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}

			// AND: Params[key] is set/unchanged as expected.
			if got := input.GetParam(key); got != tc.want {
				t.Fatalf(
					"%s Params[%q] mismatch\ngot:  %q\nwant: %q",
					prefix, key,
					got, tc.want,
				)
			}
		})
	}
}
