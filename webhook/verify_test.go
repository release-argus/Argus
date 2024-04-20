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

package webhook

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestSliceDefaults_Print(t *testing.T) {
	testValid := testWebHookDefaults(false, false, false)
	testInvalid := testWebHookDefaults(true, true, false)
	// GIVEN a SliceDefaults
	tests := map[string]struct {
		slice *SliceDefaults
		want  string
	}{
		"nil slice": {
			slice: nil,
			want:  "",
		},
		"single element slice": {
			slice: &SliceDefaults{
				"single": testValid},
			want: `
webhook:
  single:
    type: ` + testValid.Type + `
    url: ` + testValid.URL + `
    allow_invalid_certs: ` + fmt.Sprint(*testValid.AllowInvalidCerts) + `
    secret: ` + testValid.Secret + `
    desired_status_code: ` + fmt.Sprint(*testValid.DesiredStatusCode) + `
    delay: ` + fmt.Sprint(testValid.Delay) + `
    max_tries: ` + fmt.Sprint(*testValid.MaxTries) + `
    silent_fails: ` + fmt.Sprint(*testValid.SilentFails),
		},
		"multiple element slice": {
			slice: &SliceDefaults{
				"first":  testValid,
				"second": testInvalid},
			want: `
webhook:
  first:
    type: ` + testValid.Type + `
    url: ` + testValid.URL + `
    allow_invalid_certs: ` + fmt.Sprint(*testValid.AllowInvalidCerts) + `
    secret: ` + testValid.Secret + `
    desired_status_code: ` + fmt.Sprint(*testValid.DesiredStatusCode) + `
    delay: ` + fmt.Sprint(testValid.Delay) + `
    max_tries: ` + fmt.Sprint(*testValid.MaxTries) + `
    silent_fails: ` + fmt.Sprint(*testValid.SilentFails) + `
  second:
    type: ` + testInvalid.Type + `
    url: ` + testInvalid.URL + `
    allow_invalid_certs: ` + fmt.Sprint(*testInvalid.AllowInvalidCerts) + `
    secret: ` + testInvalid.Secret + `
    desired_status_code: ` + fmt.Sprint(*testInvalid.DesiredStatusCode) + `
    delay: ` + fmt.Sprint(testInvalid.Delay) + `
    max_tries: ` + fmt.Sprint(*testInvalid.MaxTries) + `
    silent_fails: ` + fmt.Sprint(*testInvalid.SilentFails),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout
			test.StdoutMutex.Lock()
			defer test.StdoutMutex.Unlock()

			if tc.want != "" {
				tc.want += "\n"
			}
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN Print is called
			tc.slice.Print("")

			// THEN it prints the expected output
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = stdout
			strOut := string(out)
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if strOut != tc.want {
				t.Errorf("Print should have given\n%q\nbut gave\n%q",
					tc.want, strOut)
			}
		})
	}
}

func TestWebHookDefaults_CheckValues(t *testing.T) {
	// GIVEN a WebHookDefaults
	tests := map[string]struct {
		webhook   *WebHookDefaults
		wantDelay string
		errRegex  []string
	}{
		"valid WebHook": {
			webhook: testWebHookDefaults(false, false, false),
		},
		"invalid delay": {
			errRegex: []string{
				"^delay: .* <invalid>"},
			webhook: NewDefaults(
				nil, nil,
				"4y",
				nil, nil, "", nil, "", ""),
		},
		"fix int delay": {
			webhook: NewDefaults(
				nil, nil,
				"3",
				nil, nil, "", nil, "", ""),
			wantDelay: "3s",
		},
		"invalid type": {
			errRegex: []string{
				"^type: .*foo.* <invalid>"},
			webhook: NewDefaults(
				nil, nil, "", nil, nil, "", nil,
				"foo",
				""),
		},
		"invalid url template": {
			errRegex: []string{
				"url: .* <invalid>"},
			webhook: NewDefaults(
				nil, nil, "", nil, nil, "", nil, "",
				"https://example.com/{{ version }"),
		},
		"valid custom headers": {
			webhook: NewDefaults(
				nil,
				&Headers{
					{Key: "foo", Value: "bar"}},
				"", nil, nil, "", nil, "",
				"https://example.com/{{ version }"),
		},
		"invalid custom headers": {
			errRegex: []string{
				`^custom_headers:$`,
				`^  bar: "[^"]+" <invalid>`},
			webhook: NewDefaults(
				nil,
				&Headers{
					{Key: "foo", Value: "bar"},
					{Key: "bar", Value: "{{ version }"}},
				"", nil, nil, "", nil, "", ""),
		},
		"all errs": {
			errRegex: []string{
				`^type: "[^"]+" <invalid>`,
				`^delay: "[^"]+" <invalid>`,
				`^url: "[^"]+" <invalid>`},
			webhook: NewDefaults(
				nil,
				&Headers{
					{Key: "foo", Value: "bar"},
					{Key: "bar", Value: "{{ version }"}},
				"5x",
				nil, nil, "", nil,
				"shazam",
				"https://example.com/{{ version }"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CheckValues is called
			err := tc.webhook.CheckValues("")

			// THEN it err's when expected
			e := util.ErrorToString(err)
			lines := strings.Split(e, `\`)
			if len(tc.errRegex) > len(lines) {
				t.Errorf("want %d errors:\n%v\ngot %d errors:\n%v",
					len(tc.errRegex), tc.errRegex, len(lines), lines)
				return
			}
			for i := range tc.errRegex {
				re := regexp.MustCompile(tc.errRegex[i])
				match := re.MatchString(lines[i])
				if !match {
					t.Errorf("want match for: %q\ngot:  %q",
						tc.errRegex[i], e)
					return
				}
			}
			if tc.wantDelay != "" && tc.webhook.Delay != tc.wantDelay {
				t.Errorf("want delay=%q\ngot  delay=%q",
					tc.wantDelay, tc.webhook.Delay)
			}
		})
	}
}

func TestWebHook_CheckValues(t *testing.T) {
	// GIVEN a WebHook
	tests := map[string]struct {
		delay         string
		wantDelay     string
		whType        *string
		whMainType    string
		url           *string
		secret        *string
		customHeaders Headers
		errRegex      []string
	}{
		"valid WebHook": {},
		"invalid delay": {
			errRegex: []string{
				"^delay: .* <invalid>"},
			delay: "5x",
		},
		"fix int delay": {
			delay:     "5",
			wantDelay: "5s",
		},
		"invalid type": {
			errRegex: []string{
				"^type: .*foo.* <invalid>"},
			whType: stringPtr("foo"),
		},
		"invalid main type": {
			errRegex:   []string{}, // Invalid, but caught in the Defaults CheckValues
			whType:     stringPtr(""),
			whMainType: "bar",
		},
		"mismatching type and main type": {
			errRegex: []string{
				`^type: "github" != "gitlab" <invalid>`},
			whType:     stringPtr("github"),
			whMainType: "gitlab",
		},
		"no type": {
			errRegex: []string{
				"^type: <required>"},
			whType: stringPtr(""),
		},
		"invalid url template": {
			errRegex: []string{
				"url: .* <invalid>"},
			url: stringPtr("{{ version }"),
		},
		"no url": {
			errRegex: []string{
				"^url: <required>"},
			url: stringPtr(""),
		},
		"no secret": {
			errRegex: []string{
				"^secret: <required>"},
			secret: stringPtr(""),
		},
		"valid custom headers": {
			customHeaders: Headers{
				{Key: "foo", Value: "bar"}},
		},
		"invalid custom headers": {
			errRegex: []string{
				`^custom_headers:$`,
				`^  bar: "[^"]+" <invalid>`},
			customHeaders: Headers{
				{Key: "foo", Value: "bar"},
				{Key: "bar", Value: "{{ version }"}},
		},
		"all errs": {
			errRegex: []string{
				`^type: "[^"]+" <invalid>`,
				`^delay: "[^"]+" <invalid>`,
				`^url: <required>`,
				`^secret: <required>`},
			delay:  "5x",
			whType: stringPtr("foo"),
			url:    stringPtr(""),
			secret: stringPtr(""),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			if tc.whMainType != "" {
				webhook.Main.Type = tc.whMainType
			}
			if tc.delay != "" {
				webhook.Delay = tc.delay
			}
			if tc.whType != nil {
				webhook.Type = *tc.whType
			}
			if tc.url != nil {
				webhook.URL = *tc.url
			}
			if tc.secret != nil {
				webhook.Secret = *tc.secret
			}
			webhook.CustomHeaders = &tc.customHeaders

			// WHEN CheckValues is called
			err := webhook.CheckValues("")

			// THEN it err's when expected
			e := util.ErrorToString(err)
			lines := strings.Split(e, `\`)
			if len(tc.errRegex) > len(lines) {
				t.Errorf("want %d errors:\n%v\ngot %d errors:\n%v",
					len(tc.errRegex), tc.errRegex, len(lines), lines)
				return
			}
			for i := range tc.errRegex {
				re := regexp.MustCompile(tc.errRegex[i])
				match := re.MatchString(lines[i])
				if !match {
					t.Errorf("want match for: %q\ngot:  %q",
						tc.errRegex[i], e)
					return
				}
			}
			if tc.wantDelay != "" && webhook.Delay != tc.wantDelay {
				t.Errorf("want delay=%q\ngot  delay=%q",
					tc.wantDelay, webhook.Delay)
			}
		})
	}
}

func TestSliceDefaults_CheckValues(t *testing.T) {
	// GIVEN a SliceDefaults
	tests := map[string]struct {
		slice    *SliceDefaults
		errRegex []string
	}{
		"nil slice": {},
		"valid single element slice": {
			slice: &SliceDefaults{
				"a": testWebHookDefaults(true, false, false)},
		},
		"invalid single element slice": {
			errRegex: []string{
				`^webhook:$`,
				`^  a:$`,
				`^    delay: .* <invalid>`},
			slice: &SliceDefaults{
				"a": NewDefaults(
					nil, nil,
					"5x",
					nil, nil, "", nil, "", "")},
		},
		"valid multi element slice": {
			slice: &SliceDefaults{
				"a": testWebHookDefaults(true, false, false),
				"b": testWebHookDefaults(false, false, false)},
		},
		"invalid multi element slice": {
			errRegex: []string{
				`^webhook:$`,
				`^  a:$`,
				`^    delay: "[^"]+" <invalid>`,
				`^  b:$`,
				`^    type: "[^"]+" <invalid>`,
				`^    url: "[^"]+" <invalid>`},
			slice: &SliceDefaults{
				"a": NewDefaults(
					nil, nil,
					"5x",
					nil, nil, "", nil, "", ""),
				"b": NewDefaults(
					nil, nil,
					"4",
					nil, nil, "", nil, "foo", "https://example.com/{{ version }")},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CheckValues is called
			err := tc.slice.CheckValues("")

			// THEN it err's when expected
			e := util.ErrorToString(err)
			lines := strings.Split(e, `\`)
			if len(tc.errRegex) > len(lines) {
				t.Errorf("want %d errors:\n%v\ngot %d errors:\n%v",
					len(tc.errRegex), tc.errRegex, len(lines), lines)
				return
			}
			for i := range tc.errRegex {
				re := regexp.MustCompile(tc.errRegex[i])
				match := re.MatchString(lines[i])
				if !match {
					t.Errorf("want match for: %q\ngot:  %q",
						tc.errRegex[i], e)
					return
				}
			}
		})
	}
}

func TestSlice_CheckValues(t *testing.T) {
	// GIVEN a Slice
	tests := map[string]struct {
		slice    *Slice
		errRegex []string
	}{
		"nil slice": {},
		"valid single element slice": {
			slice: &Slice{
				"a": testWebHook(true, false, false)},
		},
		"invalid single element slice": {
			errRegex: []string{
				`^webhook:$`,
				`^  a:$`,
				`^    delay: .* <invalid>`},
			slice: &Slice{
				"a": New(
					nil, nil,
					"5x",
					nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
		},
		"valid multi element slice": {
			slice: &Slice{
				"a": testWebHook(true, false, false),
				"b": testWebHook(false, false, false)},
		},
		"invalid multi element slice": {
			errRegex: []string{
				`^webhook:$`,
				`^  a:$`,
				`^    delay: "[^"]+" <invalid>`,
				`^    type: <required>`,
				`^    url: <required>`,
				`^    secret: <required>`,
				`^  b:$`,
				`^    type: "[^"]+" <invalid>`,
				`^    url: <required>`,
				`^    secret: <required>`},
			slice: &Slice{
				"a": New(
					nil, nil,
					"5x",
					nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil),
				"b": New(
					nil, nil, "", nil, nil, nil, nil, nil, "", nil,
					"foo",
					"",
					&WebHookDefaults{},
					&WebHookDefaults{},
					&WebHookDefaults{})},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.slice != nil {
				svcStatus := svcstatus.Status{}
				svcStatus.Init(
					0, 0, len(*tc.slice),
					nil, nil)
				tc.slice.Init(
					&svcStatus,
					&SliceDefaults{}, &WebHookDefaults{}, &WebHookDefaults{},
					nil, nil)
			}

			// WHEN CheckValues is called
			err := tc.slice.CheckValues("")

			// THEN it err's when expected
			e := util.ErrorToString(err)
			lines := strings.Split(e, `\`)
			if len(tc.errRegex) > len(lines) {
				t.Errorf("want %d errors:\n%v\ngot %d errors:\n%v",
					len(tc.errRegex), tc.errRegex, len(lines), lines)
				return
			}
			for i := range tc.errRegex {
				re := regexp.MustCompile(tc.errRegex[i])
				match := re.MatchString(lines[i])
				if !match {
					t.Errorf("want match for: %q\ngot:  %q",
						tc.errRegex[i], e)
					return
				}
			}
		})
	}
}
