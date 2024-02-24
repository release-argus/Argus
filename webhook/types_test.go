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

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/util"
	"gopkg.in/yaml.v3"
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

func TestWebHookDefaults_String(t *testing.T) {
	tests := map[string]struct {
		webhook *WebHookDefaults
		want    string
	}{
		"nil": {
			webhook: nil,
			want:    "",
		},
		"empty": {
			webhook: &WebHookDefaults{},
			want:    "{}",
		},
		"filled": {
			webhook: NewDefaults(
				boolPtr(false),
				&Headers{
					{Key: "X-Header", Value: "val"},
					{Key: "X-Another", Value: "val2"}},
				"1h1m1s",
				intPtr(200),
				uintPtr(4),
				"foobar",
				boolPtr(true),
				"github",
				"https://example.com"),
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
delay: 1h1m1s
max_tries: 4
silent_fails: true`,
		},
		"quotes otherwise invalid yaml strings": {
			webhook: NewDefaults(
				nil,
				&Headers{
					{Key: ">123", Value: "{pass}"}},
				"", nil, nil, "", nil, "", ""),
			want: `
custom_headers:
  - key: '>123'
    value: '{pass}'`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			prefixes := []string{"", " ", "  ", "    ", "- "}
			for _, prefix := range prefixes {
				want := strings.TrimPrefix(tc.want, "\n")
				if want != "" {
					if want != "{}" {
						want = prefix + strings.ReplaceAll(want, "\n", "\n"+prefix)
					}
					want += "\n"
				}

				// WHEN the WebHookDefaults are stringified with String
				got := tc.webhook.String(prefix)

				// THEN the result is as expected
				want = strings.TrimPrefix(want, "\n")
				if got != want {
					t.Fatalf("(prefix=%q) got:\n%q\nwant:\n%q",
						prefix, got, want)
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
			want:    "",
		},
		"empty": {
			webhook: &WebHook{},
			want:    "{}\n",
		},
		"filled": {
			webhook: New(
				boolPtr(false), // allow_invalid_certs
				&Headers{ // custom_headers
					{Key: "X-Header", Value: "val"},
					{Key: "X-Another", Value: "val2"}},
				"1h1m1s",    // delay
				intPtr(200), // desired_status_code
				nil,         // failed
				uintPtr(4),  // max_tries
				&Notifiers{ // notifiers
					Shoutrrr: &shoutrrr.Slice{
						"foo": shoutrrr.New(
							nil, "", nil, nil,
							"discord",
							nil, nil, nil, nil)}},
				stringPtr("3h2m1s"),   // parent_interval
				"foobar",              // secret
				boolPtr(true),         // silent_fails
				"github",              // type
				"https://example.com", // url
				NewDefaults( // main
					boolPtr(false),
					nil, "", nil, nil, "", nil, "", ""),
				NewDefaults( // defaults
					boolPtr(true),
					nil, "", nil, nil, "", nil, "", ""),
				NewDefaults( // hard_defaults
					boolPtr(true),
					nil, "", nil, nil, "", nil, "", "")),
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
delay: 1h1m1s
max_tries: 4
silent_fails: true
`,
		},
		"quotes otherwise invalid yaml strings": {
			webhook: New(
				nil,
				&Headers{
					{Key: ">123", Value: "{pass}"}},
				"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil),
			want: `
custom_headers:
  - key: '>123'
    value: '{pass}'
`},
	}

	for name, tc := range tests {
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

func TestSliceDefaults_String(t *testing.T) {
	tests := map[string]struct {
		slice *SliceDefaults
		want  string
	}{
		"nil": {
			slice: nil,
			want:  "",
		},
		"empty": {
			slice: &SliceDefaults{},
			want:  "{}",
		},
		"one empty and one nil": {
			slice: &SliceDefaults{
				"one": &WebHookDefaults{},
				"two": nil},
			want: `
one: {}`,
		},
		"one with data": {
			slice: &SliceDefaults{
				"one": NewDefaults(
					nil, nil, "", nil, nil, "", nil,
					"github",
					"https://example.com")},
			want: `
one:
  type: github
  url: https://example.com`,
		},
		"multiple": {
			slice: &SliceDefaults{
				"one": NewDefaults(
					nil, nil, "", nil, nil, "", nil,
					"github",
					"https://example.com"),
				"two": NewDefaults(
					nil, nil, "", nil, nil, "", nil,
					"gitlab",
					"https://other.com")},
			want: `
one:
  type: github
  url: https://example.com
two:
  type: gitlab
  url: https://other.com`,
		},
		"quotes otherwise invalid yaml strings": {
			slice: &SliceDefaults{
				"invalid": NewDefaults(
					nil,
					&Headers{
						{Key: ">123", Value: "{pass}"}},
					"", nil, nil, "", nil, "", "")},
			want: `
invalid:
  custom_headers:
    - key: '>123'
      value: '{pass}'`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel()

			prefixes := []string{"", " ", "  ", "    ", "- "}
			for _, prefix := range prefixes {
				want := strings.TrimPrefix(tc.want, "\n")
				if want != "" {
					if want != "{}" {
						want = prefix + strings.ReplaceAll(want, "\n", "\n"+prefix)
					}
					want += "\n"
				}

				// WHEN the Slice is stringified with String
				got := tc.slice.String(prefix)

				// THEN the result is as expected
				want = strings.TrimPrefix(want, "\n")
				if got != want {
					t.Fatalf("(prefix=%q) got:\n%q\nwant:\n%q",
						prefix, got, want)
				}
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
			want:  "",
		},
		"empty": {
			slice: &Slice{},
			want:  "{}\n",
		},
		"one": {
			slice: &Slice{
				"one": New(
					nil, nil, "", nil, nil, nil, nil, nil, "", nil,
					"github",
					"https://example.com",
					nil, nil, nil)},
			want: `
one:
  type: github
  url: https://example.com
`,
		},
		"multiple": {
			slice: &Slice{
				"one": New(
					nil, nil, "", nil, nil, nil, nil, nil, "", nil,
					"github",
					"https://example.com",
					nil, nil, nil),
				"two": New(
					nil, nil, "", nil, nil, nil, nil, nil, "", nil,
					"gitlab",
					"https://other.com",
					nil, nil, nil)},
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
				"invalid": New(
					nil,
					&Headers{
						{Key: ">123", Value: "{pass}"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
			want: `
invalid:
  custom_headers:
    - key: '>123'
      value: '{pass}'
`},
	}

	for name, tc := range tests {
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
