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

func TestWebHooksDefaults_Print(t *testing.T) {
	testValid := testDefaults(false, false)
	testInvalid := testDefaults(true, false)
	// GIVEN a WebHooksDefaults.
	tests := map[string]struct {
		webhooksDefaults *WebHooksDefaults
		want             string
	}{
		"nil map": {
			webhooksDefaults: nil,
			want:             "",
		},
		"single element map": {
			webhooksDefaults: &WebHooksDefaults{
				"single": testValid},
			want: test.TrimYAML(`
				webhook:
					single:
						type: ` + testValid.Type + `
						url: ` + testValid.URL + `
						allow_invalid_certs: ` + fmt.Sprint(*testValid.AllowInvalidCerts) + `
						secret: ` + testValid.Secret + `
						desired_status_code: ` + fmt.Sprint(*testValid.DesiredStatusCode) + `
						delay: ` + testValid.Delay + `
						max_tries: ` + fmt.Sprint(*testValid.MaxTries) + `
						silent_fails: ` + fmt.Sprint(*testValid.SilentFails)),
		},
		"multiple element map": {
			webhooksDefaults: &WebHooksDefaults{
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
						delay: ` + testValid.Delay + `
						max_tries: ` + fmt.Sprint(*testValid.MaxTries) + `
						silent_fails: ` + fmt.Sprint(*testValid.SilentFails) + `
					second:
						type: ` + testInvalid.Type + `
						url: ` + testInvalid.URL + `
						allow_invalid_certs: ` + fmt.Sprint(*testInvalid.AllowInvalidCerts) + `
						secret: ` + testInvalid.Secret + `
						desired_status_code: ` + fmt.Sprint(*testInvalid.DesiredStatusCode) + `
						delay: ` + testInvalid.Delay + `
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
			tc.webhooksDefaults.Print("")

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
	// GIVEN Defaults.
	tests := map[string]struct {
		webhook   *Defaults
		wantDelay string
		errRegex  string
		changed   bool
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
				headers:
					bar: "[^"]+" <invalid>`),
			webhook: NewDefaults(
				nil,
				&Headers{
					{Key: "foo", Value: "bar"},
					{Key: "bar", Value: "{{ version }"}},
				"", nil, nil, "", nil, "", ""),
		},
		"custom_headers -> headers": {
			webhook: &Defaults{
				Base: Base{
					CustomHeaders: &Headers{
						{Key: "foo", Value: "bar"}}}},
			changed: true,
		},
		"all errs": {
			errRegex: test.TrimYAML(`
				type: "[^"]+" <invalid>.*
				url: "[^"]+" <invalid>.*
				headers:
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
			err, changed := tc.webhook.CheckValues("")

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
			// AND the 'changed' flag matches expectation.
			if changed != tc.changed {
				t.Errorf("%s\nchanged flag mismatch\nwant: %t\ngot:  %t",
					packageName, tc.changed, changed)
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
		customHeaders    *Headers
		headers          *Headers
		errRegex         string
		changed          bool
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
		"valid headers": {
			headers: &Headers{
				{Key: "foo", Value: "bar"}},
			changed: false,
		},
		"invalid headers": {
			errRegex: test.TrimYAML(`
				^headers:
					bar: "[^"]+" <invalid>.*$`),
			headers: &Headers{
				{Key: "foo", Value: "bar"},
				{Key: "bar", Value: "{{ version }"}},
		},
		"custom_headers -> headers": {
			customHeaders: &Headers{
				{Key: "foo", Value: "bar"}},
			changed: true,
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
			webhook.CustomHeaders = tc.customHeaders
			webhook.Headers = tc.headers

			// WHEN CheckValues is called.
			err, changed := webhook.CheckValues("")

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
			// AND CustomHeaders are always moved to Headers.
			if webhook.CustomHeaders != nil && webhook.Headers == nil {
				t.Errorf("%s\nCustomHeaders not moved to Headers\nHeaders=%v\nCustomHeaders=%v",
					packageName, webhook.Headers, webhook.CustomHeaders)
			}
			// AND the 'changed' flag matches expectation.
			if changed != tc.changed {
				t.Errorf("%s\nchanged flag mismatch\nwant: %t\ngot:  %t",
					packageName, tc.changed, changed)
			}
		})
	}
}

func TestWebHooksDefaults_CheckValues(t *testing.T) {
	// GIVEN WebHooksDefaults.
	tests := map[string]struct {
		webhooksDefaults *WebHooksDefaults
		errRegex         string
		changed          bool
	}{
		"nil map": {},
		"valid single element map": {
			webhooksDefaults: &WebHooksDefaults{
				"a": testDefaults(true, false)},
		},
		"invalid single element map": {
			errRegex: test.TrimYAML(`
				^a:
					delay: .* <invalid>.*$`),
			webhooksDefaults: &WebHooksDefaults{
				"a": NewDefaults(
					nil, nil,
					"5x",
					nil, nil, "", nil, "", "")},
		},
		"valid multi element map": {
			webhooksDefaults: &WebHooksDefaults{
				"a": testDefaults(true, false),
				"b": testDefaults(false, false)},
		},
		"invalid multi element map": {
			errRegex: test.TrimYAML(`
				^a:
					delay: "[^"]+" <invalid>.*
				b:
					type: "[^"]+" <invalid>.*
					url: "[^"]+" <invalid>.*$`),
			webhooksDefaults: &WebHooksDefaults{
				"a": NewDefaults(
					nil, nil,
					"5x",
					nil, nil, "", nil, "", ""),
				"b": NewDefaults(
					nil, nil,
					"4",
					nil, nil, "", nil, "foo", "https://example.com/{{ version }")},
		},
		"custom_headers -> headers": {
			webhooksDefaults: &WebHooksDefaults{
				"a": &Defaults{
					Base: Base{
						CustomHeaders: &Headers{
							{Key: "foo", Value: "bar"}}}}},
			changed: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CheckValues is called.
			err, changed := tc.webhooksDefaults.CheckValues("")

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
			// AND the 'changed' flag matches expectation.
			if changed != tc.changed {
				t.Errorf("%s\nchanged flag mismatch\nwant: %t\ngot:  %t",
					packageName, tc.changed, changed)
			}
		})
	}
}

func TestWebHooks_CheckValues(t *testing.T) {
	// GIVEN WebHooks.
	tests := map[string]struct {
		webhooks *WebHooks
		errRegex string
		changed  bool
	}{
		"nil map": {},
		"valid single element map": {
			webhooks: &WebHooks{
				"a": testWebHook(true, false, false)},
		},
		"invalid single element map": {
			errRegex: test.TrimYAML(`
				^a:
					type: <required>.*
					delay: "5x" <invalid>.*
					url: <required>.*
					secret: <required>.*$`),
			webhooks: &WebHooks{
				"a": New(
					nil, nil,
					"5x",
					nil, nil,
					"a",
					nil, nil, nil,
					"",
					nil,
					"", "",
					nil, nil, nil)},
		},
		"valid multi element map": {
			webhooks: &WebHooks{
				"a": testWebHook(true, false, false),
				"b": testWebHook(false, false, false)},
		},
		"invalid multi element map": {
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
			webhooks: &WebHooks{
				"a": New(
					nil, nil,
					"5x",
					nil, nil,
					"a",
					nil, nil, nil,
					"",
					nil,
					"", "",
					nil, nil, nil),
				"b": New(
					nil, nil,
					"",
					nil, nil,
					"b",
					nil, nil, nil,
					"",
					nil,
					"foo", "",
					&Defaults{}, &Defaults{}, &Defaults{})},
		},
		"custom_headers -> headers": {
			webhooks: &WebHooks{
				"a": &WebHook{
					Base: Base{
						Type:   "github",
						URL:    "example.com",
						Secret: "Argus",
						CustomHeaders: &Headers{
							{Key: "foo", Value: "bar"}}}}},
			changed: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.webhooks != nil {
				svcStatus := status.Status{}
				svcStatus.Init(
					0, 0, len(*tc.webhooks),
					"", "", "",
					&dashboard.Options{})
				tc.webhooks.Init(
					&svcStatus,
					&WebHooksDefaults{},
					&Defaults{}, &Defaults{},
					nil, nil)
			}

			// WHEN CheckValues is called.
			err, changed := tc.webhooks.CheckValues("")

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
			// AND the 'changed' flag matches expectation.
			if changed != tc.changed {
				t.Errorf("%s\nchanged flag mismatch\nwant: %t\ngot:  %t",
					packageName, tc.changed, changed)
			}
		})
	}
}
