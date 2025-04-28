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

package webhook

import (
	"fmt"
	"strings"
	"testing"

	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestSliceDefaults_Print(t *testing.T) {
	testValid := testDefaults(false, false)
	testInvalid := testDefaults(true, false)
	// GIVEN a SliceDefaults.
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
			want: test.TrimYAML(`
				webhook:
					single:
						type: ` + testValid.Type + `
						url: ` + testValid.URL + `
						allow_invalid_certs: ` + fmt.Sprint(*testValid.AllowInvalidCerts) + `
						secret: ` + testValid.Secret + `
						desired_status_code: ` + fmt.Sprint(*testValid.DesiredStatusCode) + `
						delay: ` + fmt.Sprint(testValid.Delay) + `
						max_tries: ` + fmt.Sprint(*testValid.MaxTries) + `
						silent_fails: ` + fmt.Sprint(*testValid.SilentFails)),
		},
		"multiple element slice": {
			slice: &SliceDefaults{
				"first":  testValid,
				"second": testInvalid},
			want: test.TrimYAML(`
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
						silent_fails: ` + fmt.Sprint(*testInvalid.SilentFails)),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout()

			if tc.want != "" {
				tc.want += "\n"
			}

			// WHEN Print is called.
			tc.slice.Print("")

			// THEN it prints the expected output.
			stdout := releaseStdout()
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if stdout != tc.want {
				t.Errorf("%s\nstdout mismatch\nwant: %q\ngot:  %q",
					packageName, tc.want, stdout)
			}
		})
	}
}

func TestDefaults_CheckValues(t *testing.T) {
	// GIVEN a Defaults.
	tests := map[string]struct {
		webhook   *Defaults
		wantDelay string
		errRegex  string
	}{
		"valid WebHook": {
			webhook: testDefaults(false, false),
		},
		"invalid delay": {
			errRegex: test.TrimYAML(`
				^delay: "4y" <invalid>`),
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
			errRegex: test.TrimYAML(`
				^type: .*foo.* <invalid>`),
			webhook: NewDefaults(
				nil, nil, "", nil, nil, "", nil,
				"foo",
				""),
		},
		"invalid url template": {
			errRegex: test.TrimYAML(`
				url: ".+" <invalid>`),
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
			errRegex: test.TrimYAML(`
				custom_headers:
					bar: "[^"]+" <invalid>`),
			webhook: NewDefaults(
				nil,
				&Headers{
					{Key: "foo", Value: "bar"},
					{Key: "bar", Value: "{{ version }"}},
				"", nil, nil, "", nil, "", ""),
		},
		"all errs": {
			errRegex: test.TrimYAML(`
				type: "[^"]+" <invalid>.*
				url: "[^"]+" <invalid>.*
				custom_headers:
					bar: "[^"]+" <invalid>.*
				delay: "[^"]+" <invalid>.*$`),
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

			// WHEN CheckValues is called.
			err := tc.webhook.CheckValues("")

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			lines := strings.Split(e, "\n")
			wantLines := strings.Count(tc.errRegex, "\n")
			if wantLines > len(lines) {
				t.Errorf("%s\nerror mismatch\nwant %d lines of error:\n%q\ngot %d lines:\n%q",
					packageName,
					wantLines, tc.errRegex,
					len(lines), lines)
				return
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
				return
			}
			// AND the delay is fixed when expected.
			if tc.wantDelay != "" && tc.webhook.Delay != tc.wantDelay {
				t.Errorf("%s\ndelay mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wantDelay, tc.webhook.Delay)
			}
		})
	}
}

func TestWebHook_CheckValues(t *testing.T) {
	// GIVEN a WebHook.
	tests := map[string]struct {
		delay, wantDelay string
		whType           *string
		whMainType       string
		url, secret      *string
		customHeaders    Headers
		errRegex         string
	}{
		"valid WebHook": {},
		"invalid delay": {
			errRegex: `^delay: .* <invalid>`,
			delay:    "5x",
		},
		"fix int delay": {
			delay:     "5",
			wantDelay: "5s",
		},
		"invalid type": {
			errRegex: `^type: .*foo.* <invalid>`,
			whType:   test.StringPtr("foo"),
		},
		"invalid main type": {
			errRegex:   "", // Invalid, but caught in the Defaults CheckValues.
			whType:     test.StringPtr(""),
			whMainType: "bar",
		},
		"mismatching type and main type": {
			errRegex:   `^type: "github" != "gitlab" <invalid>.*$`,
			whType:     test.StringPtr("github"),
			whMainType: "gitlab",
		},
		"no type": {
			errRegex: `^type: <required>.*$`,
			whType:   test.StringPtr(""),
		},
		"invalid url template": {
			errRegex: `^url: .* <invalid>.*$`,
			url:      test.StringPtr("{{ version }"),
		},
		"no url": {
			errRegex: `^url: <required>.*$`,
			url:      test.StringPtr(""),
		},
		"no secret": {
			errRegex: `^secret: <required>.*$`,
			secret:   test.StringPtr(""),
		},
		"valid custom headers": {
			customHeaders: Headers{
				{Key: "foo", Value: "bar"}},
		},
		"invalid custom headers": {
			errRegex: test.TrimYAML(`
				^custom_headers:
					bar: "[^"]+" <invalid>.*$`),
			customHeaders: Headers{
				{Key: "foo", Value: "bar"},
				{Key: "bar", Value: "{{ version }"}},
		},
		"all errs": {
			errRegex: test.TrimYAML(`
				^type: "[^"]+" <invalid>.*
				delay: "[^"]+" <invalid>.*
				url: <required>.*
				secret: <required>.*$`),
			delay:  "5x",
			whType: test.StringPtr("foo"),
			url:    test.StringPtr(""),
			secret: test.StringPtr(""),
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

			// WHEN CheckValues is called.
			err := webhook.CheckValues("")

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			lines := strings.Split(e, "\n")
			wantLines := strings.Count(tc.errRegex, "\n")
			if wantLines > len(lines) {
				t.Fatalf("%s\nerror count mismatch\nwant: %d lines of error:\n%q\ngot %d:\n%v",
					packageName,
					wantLines, tc.errRegex,
					len(lines), lines)
				return
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
				return
			}
			// AND the delay is fixed when expected.
			if tc.wantDelay != "" && webhook.Delay != tc.wantDelay {
				t.Errorf("%s\ndelay mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wantDelay, webhook.Delay)
			}
		})
	}
}

func TestSliceDefaults_CheckValues(t *testing.T) {
	// GIVEN a SliceDefaults.
	tests := map[string]struct {
		slice    *SliceDefaults
		errRegex string
	}{
		"nil slice": {},
		"valid single element slice": {
			slice: &SliceDefaults{
				"a": testDefaults(true, false)},
		},
		"invalid single element slice": {
			errRegex: test.TrimYAML(`
				^a:
					delay: .* <invalid>.*$`),
			slice: &SliceDefaults{
				"a": NewDefaults(
					nil, nil,
					"5x",
					nil, nil, "", nil, "", "")},
		},
		"valid multi element slice": {
			slice: &SliceDefaults{
				"a": testDefaults(true, false),
				"b": testDefaults(false, false)},
		},
		"invalid multi element slice": {
			errRegex: test.TrimYAML(`
				^a:
					delay: "[^"]+" <invalid>.*
				b:
					type: "[^"]+" <invalid>.*
					url: "[^"]+" <invalid>.*$`),
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

			// WHEN CheckValues is called.
			err := tc.slice.CheckValues("")

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			lines := strings.Split(e, "\n")
			wantLines := strings.Count(tc.errRegex, "\n")
			if wantLines > len(lines) {
				t.Fatalf("%s\nerror count mismatch\nwant: %d lines of error:\n%q\ngot:  %d lines:\n%q",
					packageName,
					wantLines, tc.errRegex,
					len(lines), lines)
				return
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
				return
			}
		})
	}
}

func TestSlice_CheckValues(t *testing.T) {
	// GIVEN a Slice.
	tests := map[string]struct {
		slice    *Slice
		errRegex string
	}{
		"nil slice": {},
		"valid single element slice": {
			slice: &Slice{
				"a": testWebHook(true, false, false)},
		},
		"invalid single element slice": {
			errRegex: test.TrimYAML(`
				^a:
					type: <required>.*
					delay: "5x" <invalid>.*
					url: <required>.*
					secret: <required>.*$`),
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
			errRegex: test.TrimYAML(`
				^a:
					type: <required>.*
					delay: "5x" <invalid>.*
					url: <required>.*
					secret: <required>.*
				b:
					type: "foo" <invalid>.*
					url: <required>.*
					secret: <required>.*$`),
			slice: &Slice{
				"a": New(
					nil, nil,
					"5x",
					nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil),
				"b": New(
					nil, nil, "", nil, nil, nil, nil, nil, "", nil,
					"foo",
					"",
					&Defaults{},
					&Defaults{}, &Defaults{})},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.slice != nil {
				svcStatus := status.Status{}
				svcStatus.Init(
					0, 0, len(*tc.slice),
					"", "", "",
					&dashboard.Options{})
				tc.slice.Init(
					&svcStatus,
					&SliceDefaults{},
					&Defaults{}, &Defaults{},
					nil, nil)
			}

			// WHEN CheckValues is called.
			err := tc.slice.CheckValues("")

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			lines := strings.Split(e, "\n")
			wantLines := strings.Count(tc.errRegex, "\n")
			if wantLines > len(lines) {
				t.Fatalf("%s\nerror count mismatch\nwant: %d lines of error:\n%q\ngot:  %d lines:\n%q",
					packageName,
					wantLines, tc.errRegex,
					len(lines), lines)
				return
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
				return
			}
		})
	}
}
