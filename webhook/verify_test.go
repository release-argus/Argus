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
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/release-argus/Argus/util"
)

func TestWebHookPrint(t *testing.T) {
	// GIVEN a Service
	tests := map[string]struct {
		webhook WebHook
		lines   int
	}{
		"all fields": {
			lines:   10,
			webhook: *testWebHook(true, true, false, true)},
		"partial fields": {
			lines: 2,
			webhook: WebHook{
				Type: "github",
				URL:  "https://release-argus.io"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN Print is called
			tc.webhook.Print("")

			// THEN it prints the expected number of lines
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

func TestSlicePrint(t *testing.T) {
	// GIVEN a Service
	tests := map[string]struct {
		slice        *Slice
		lines        int
		regexMatches []string
	}{
		"nil slice": {
			lines: 0, slice: nil,
		},
		"single element slice": {
			lines: 4,
			slice: &Slice{
				"single": &WebHook{
					Type: "github",
					URL:  "https://release-argus.io"}},
			regexMatches: []string{
				"^webhook:$",
				"^  single:$",
				"^    type: "},
		},
		"multiple element slice": {
			lines: 13,
			slice: &Slice{
				"first": &WebHook{
					Type: "github",
					URL:  "https://release-argus.io"},
				"second": testWebHook(true, true, false, false)},
			regexMatches: []string{
				"^webhook:$",
				"^  first:$",
				"^    type: ",
				"^  second:$",
				"^    delay"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN Print is called
			tc.slice.Print("")

			// THEN it prints the expected number of lines
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = stdout
			output := string(out)
			got := strings.Count(output, "\n")
			if got != tc.lines {
				t.Errorf("Print should have given %d lines, but gave %d\n%s",
					tc.lines, got, out)
			}
			lines := strings.Split(output, "\n")
			for _, regex := range tc.regexMatches {
				foundMatch := false
				re := regexp.MustCompile(regex)
				for _, line := range lines {
					match := re.MatchString(line)
					if match {
						foundMatch = true
						break
					}
				}
				if !foundMatch {
					t.Errorf("match on %q not found in\n%q",
						regex, output)
				}
			}
		})
	}
}

func TestWebHookCheckValues(t *testing.T) {
	// GIVEN a WebHook
	tests := map[string]struct {
		delay         string
		wantDelay     string
		noMain        bool
		whType        string
		url           *string
		secret        *string
		customHeaders Headers
		errRegex      string
	}{
		"valid WebHook": {
			errRegex: "^$",
		},
		"invalid delay": {
			errRegex: "delay: .* <invalid>",
			delay:    "5x",
		},
		"fix int delay": {
			errRegex:  "^$",
			delay:     "5",
			wantDelay: "5s",
		},
		"invalid type": {
			errRegex: "type: .* <invalid>",
			whType:   "foo",
		},
		"invalid url template": {
			errRegex: "url: .* <invalid>",
			url:      stringPtr("{{ version }"),
		},
		"no url": {
			errRegex: "url: <required>",
			url:      stringPtr(""),
		},
		"no secret": {
			errRegex: "secret: <required>",
			secret:   stringPtr(""),
		},
		"valid custom headers": {
			errRegex: "^$",
			customHeaders: Headers{
				{Key: "foo", Value: "bar"}},
		},
		"invalid custom headers": {
			errRegex: `\  bar: .* <invalid>`,
			customHeaders: Headers{
				{Key: "foo", Value: "bar"},
				{Key: "bar", Value: "{{ version }"}},
		},
		"all errs": {
			errRegex: "delay: .* <invalid>.*type: .* <invalid>.*url: <required>.*secret: <required>",
			delay:    "5x",
			whType:   "foo",
			url:      stringPtr(""),
			secret:   stringPtr(""),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, !tc.noMain, false, false)
			if tc.delay != "" {
				webhook.Delay = tc.delay
			}
			if tc.whType != "" {
				webhook.Type = tc.whType
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
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Errorf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
			if tc.wantDelay != "" && webhook.Delay != tc.wantDelay {
				t.Errorf("want delay=%q\ngot  delay=%q",
					tc.wantDelay, webhook.Delay)
			}
		})
	}
}

func TestSliceCheckValues(t *testing.T) {
	// GIVEN a Slice
	tests := map[string]struct {
		slice    *Slice
		errRegex string
	}{
		"nil slice": {
			errRegex: "^$",
		},
		"valid single element slice": {
			errRegex: "^$",
			slice: &Slice{
				"a": testWebHook(true, true, false, false)},
		},
		"invalid single element slice": {
			errRegex: "delay: .* <invalid>",
			slice: &Slice{
				"a": &WebHook{Delay: "5x"}},
		},
		"valid multi element slice": {
			errRegex: "^$",
			slice: &Slice{
				"a": testWebHook(true, true, false, false),
				"b": testWebHook(false, true, false, false)},
		},
		"invalid multi element slice": {
			errRegex: "delay: .* <invalid>.*type: .* <invalid>",
			slice: &Slice{
				"a": &WebHook{
					Delay: "5x"},
				"b": &WebHook{
					Type:         "foo",
					Main:         &WebHook{},
					Defaults:     &WebHook{},
					HardDefaults: &WebHook{}}},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CheckValues is called
			err := tc.slice.CheckValues("")

			// THEN it err's when expected
			e := util.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Errorf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
		})
	}
}
