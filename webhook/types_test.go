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
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"gopkg.in/yaml.v2"
)

func TestHeaders_UnmarshalYAML(t *testing.T) {
	// GIVEN a string to unmarshal as a Headers
	tests := map[string]struct {
		input    string
		expected Headers
		errRegex string
	}{
		"empty": {
			input:    "",
			expected: Headers{},
		},
		"single map Header": {
			input: "foo: bar",
			expected: Headers{
				{Key: "foo", Value: "bar"}},
		},
		"multiple map Headers, sorted input": {
			input: `
bish: bash
bosh: boom
foo: bar`,
			expected: Headers{
				{Key: "bish", Value: "bash"},
				{Key: "bosh", Value: "boom"},
				{Key: "foo", Value: "bar"}},
		},
		"multiple map Headers, unsorted input - sorted output": {
			input: `
foo: bar
bish: bash
bosh: boom`,
			expected: Headers{
				{Key: "bish", Value: "bash"},
				{Key: "bosh", Value: "boom"},
				{Key: "foo", Value: "bar"}},
		},
		"expecteder []Headers format yaml": {
			input: `
- key: foo
  value: bar
- key: bish
  value: bash
- key: bosh
  value: boom`,
			expected: Headers{
				{Key: "foo", Value: "bar"},
				{Key: "bish", Value: "bash"},
				{Key: "bosh", Value: "boom"}},
		},
		"invalid YAML": {
			input:    "foo",
			errRegex: `cannot unmarshal .* into map\[string\]string`,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {

			// WHEN the string is unmarshaled
			var headers Headers
			err := yaml.Unmarshal([]byte(tc.input), &headers)

			// THEN we get an error if expected
			if tc.errRegex != "" || err != nil {
				// No error expected
				if tc.errRegex == "" {
					tc.errRegex = "^$"
				}
				e := util.ErrorToString(err)
				re := regexp.MustCompile(tc.errRegex)
				match := re.MatchString(e)
				if !match {
					t.Fatalf("want match for %q\nnot: %q",
						tc.errRegex, e)
				}
				return
			}
			// AND the Headers are as expected
			if len(headers) != len(tc.expected) {
				t.Fatalf("Got differing amounts of headers\ngot: %v\nwant: %v", headers, tc.expected)
			}
			for i, header := range headers {
				if header.Key != tc.expected[i].Key {
					t.Errorf("Incorrect header key: %v, want %v", header.Key, tc.expected[i].Key)
				}
				if header.Value != tc.expected[i].Value {
					t.Errorf("Incorrect header value: %v, want %v", header.Value, tc.expected[i].Value)
				}
			}
		})
	}
}

func TestWebHook_String(t *testing.T) {
	tests := map[string]struct {
		webhook *WebHook
		want    string
	}{
		"nil": {
			webhook: nil,
			want:    "<nil>",
		},
		"empty": {
			webhook: &WebHook{},
			want:    "{}\n",
		},
		"filled": {
			webhook: &WebHook{
				ID:                "something",
				Type:              "github",
				URL:               "https://example.com",
				AllowInvalidCerts: boolPtr(false),
				CustomHeaders: &Headers{
					{Key: "X-Header", Value: "val"},
					{Key: "X-Another", Value: "val2"}},
				Secret:            "foobar",
				DesiredStatusCode: intPtr(200),
				Delay:             "1h1mm1s",
				MaxTries:          uintPtr(4),
				SilentFails:       boolPtr(true),
				NextRunnable:      time.Time{},
				Main:              &WebHook{},
				Defaults:          &WebHook{AllowInvalidCerts: boolPtr(false)},
				HardDefaults:      &WebHook{AllowInvalidCerts: boolPtr(false)},
				Notifiers: &Notifiers{
					Shoutrrr: &shoutrrr.Slice{
						"foo": &shoutrrr.Shoutrrr{Type: "discord"}}},
				ServiceStatus:  &svcstatus.Status{},
				ParentInterval: stringPtr("3h2m1s"),
			},
			want: `
type: github
url: https://example.com
allow_invalid_certs: false
custom_headers:
- key: X-Header
  value: val
- key: X-Another
  value: val2
secret: foobar
desired_status_code: 200
delay: 1h1mm1s
max_tries: 4
silent_fails: true
`,
		},
		"quotes otherwise invalid yaml strings": {
			webhook: &WebHook{
				CustomHeaders: &Headers{
					{Key: ">123", Value: "{pass}"}}},
			want: `
custom_headers:
- key: '>123'
  value: '{pass}'
`},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the WebHook is stringified with String
			got := tc.webhook.String()

			// THEN the result is as expected
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}

func TestSlice_String(t *testing.T) {
	tests := map[string]struct {
		slice *Slice
		want  string
	}{
		"nil": {
			slice: nil,
			want:  "<nil>",
		},
		"empty": {
			slice: &Slice{},
			want:  "{}\n",
		},
		"one": {
			slice: &Slice{
				"one": {
					Type: "github",
					URL:  "https://example.com"}},
			want: `
one:
  type: github
  url: https://example.com
`,
		},
		"multiple": {
			slice: &Slice{
				"one": {
					Type: "github",
					URL:  "https://example.com"},
				"two": {
					Type: "gitlab",
					URL:  "https://other.com"}},
			want: `
one:
  type: github
  url: https://example.com
two:
  type: gitlab
  url: https://other.com
`,
		},
		"quotes otherwise invalid yaml strings": {
			slice: &Slice{
				"invalid": {CustomHeaders: &Headers{
					{Key: ">123", Value: "{pass}"}}}},
			want: `
invalid:
  custom_headers:
  - key: '>123'
    value: '{pass}'
`},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the Slice is stringified with String
			got := tc.slice.String()

			// THEN the result is as expected
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}
