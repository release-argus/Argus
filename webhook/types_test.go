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
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestHeaders_UnmarshalYAML(t *testing.T) {
	// GIVEN a string to unmarshal as a Headers.
	tests := map[string]struct {
		input    string
		expected Headers
		errRegex string
	}{
		"empty": {
			input:    "",
			expected: Headers{},
			errRegex: `^$`,
		},
		"single map Header": {
			input: "foo: bar",
			expected: Headers{
				{Key: "foo", Value: "bar"}},
			errRegex: `^$`,
		},
		"multiple map Headers, sorted input": {
			input: test.TrimYAML(`
				bish: bash
				bosh: boom
				foo: bar`),
			expected: Headers{
				{Key: "bish", Value: "bash"},
				{Key: "bosh", Value: "boom"},
				{Key: "foo", Value: "bar"}},
			errRegex: `^$`,
		},
		"multiple map Headers, unsorted input - sorted output": {
			input: test.TrimYAML(`
				foo: bar
				bish: bash
				bosh: boom`),
			expected: Headers{
				{Key: "bish", Value: "bash"},
				{Key: "bosh", Value: "boom"},
				{Key: "foo", Value: "bar"}},
			errRegex: `^$`,
		},
		"expected []Headers format YAML": {
			input: test.TrimYAML(`
				- key: foo
					value: bar
				- key: bish
					value: bash
				- key: bosh
					value: boom`),
			expected: Headers{
				{Key: "foo", Value: "bar"},
				{Key: "bish", Value: "bash"},
				{Key: "bosh", Value: "boom"}},
			errRegex: `^$`,
		},
		"invalid YAML": {
			input:    "foo",
			errRegex: `cannot unmarshal .* into map\[string\]string`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// WHEN the string is unmarshalled.
			var headers Headers
			err := yaml.Unmarshal([]byte(tc.input), &headers)

			// THEN we get an error if expected.
			if tc.errRegex != "" || err != nil {
				e := util.ErrorToString(err)
				if !util.RegexCheck(tc.errRegex, e) {
					t.Fatalf("%s\nerror mismatch\nwant: %q\ngot:  %q",
						packageName, tc.errRegex, e)
				}
				return
			}
			// AND the Headers are as expected.
			if len(headers) != len(tc.expected) {
				t.Fatalf("%s\nheader length mismatch\nwant: %d (%+v)\ngot:  %d (%+v)",
					packageName,
					len(tc.expected), tc.expected,
					len(headers), headers)
			}
			for i, header := range headers {
				if header.Key != tc.expected[i].Key {
					t.Errorf("%s\nincorrect header key [%d]\nwant: %q\ngot:  %q",
						packageName, i,
						tc.expected[i].Key, header.Key)
				}
				if header.Value != tc.expected[i].Value {
					t.Errorf("%s\nincorrect header value value [%d]\nwant: %q\ngot:  %q",
						packageName, i,
						tc.expected[i].Value, header.Value)
				}
			}
		})
	}
}

func TestDefaults_String(t *testing.T) {
	// GIVEN Defaults.
	tests := map[string]struct {
		webhook *Defaults
		want    string
	}{
		"nil": {
			webhook: nil,
			want:    "",
		},
		"empty": {
			webhook: &Defaults{},
			want:    "{}",
		},
		"filled": {
			webhook: NewDefaults(
				test.BoolPtr(false),
				&Headers{
					{Key: "X-Header", Value: "val"},
					{Key: "X-Another", Value: "val2"}},
				"1h1m1s",
				test.UInt16Ptr(200),
				test.UInt8Ptr(4),
				"foobar",
				test.BoolPtr(true),
				"github",
				"https://example.com"),
			want: test.TrimYAML(`
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
				silent_fails: true`),
		},
		"quotes otherwise invalid YAML strings": {
			webhook: NewDefaults(
				nil,
				&Headers{
					{Key: ">123", Value: "{pass}"}},
				"", nil, nil, "", nil, "", ""),
			want: test.TrimYAML(`
				custom_headers:
					- key: '>123'
						value: '{pass}'`)},
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

				// WHEN the Defaults are stringified with String.
				got := tc.webhook.String(prefix)

				// THEN the result is as expected.
				want = strings.TrimPrefix(want, "\n")
				if got != want {
					t.Fatalf("%s\n(prefix=%q)\nwant: %q\ngot:  %q",
						packageName, prefix, want, got)
				}
			}
		})
	}
}

func TestWebHook_String(t *testing.T) {
	// GIVEN a WebHook.
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
				test.BoolPtr(false),
				&Headers{
					{Key: "X-Header", Value: "val"},
					{Key: "X-Another", Value: "val2"}},
				"1h1m1s",
				test.UInt16Ptr(200),
				nil,
				"filled",
				test.UInt8Ptr(4),
				&Notifiers{
					Shoutrrr: &shoutrrr.Shoutrrrs{
						"foo": shoutrrr.New(
							nil,
							"", "discord",
							nil, nil, nil,
							nil, nil, nil)}},
				test.StringPtr("3h2m1s"),
				"foobar",
				test.BoolPtr(true),
				"github", "https://example.com",
				NewDefaults(
					test.BoolPtr(false),
					nil, "", nil, nil, "", nil, "", ""),
				NewDefaults(
					test.BoolPtr(true),
					nil, "", nil, nil, "", nil, "", ""),
				NewDefaults(
					test.BoolPtr(true),
					nil, "", nil, nil, "", nil, "", "")),
			want: test.TrimYAML(`
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
			`),
		},
		"quotes otherwise invalid YAML strings": {
			webhook: New(
				nil,
				&Headers{
					{Key: ">123", Value: "{pass}"}},
				"",
				nil, nil,
				"wh",
				nil, nil, nil,
				"",
				nil,
				"", "",
				nil, nil, nil),
			want: test.TrimYAML(`
				custom_headers:
					- key: '>123'
						value: '{pass}'
			`)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the WebHook is stringified with String.
			got := tc.webhook.String()

			// THEN the result is as expected.
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestWebHooksDefaults_String(t *testing.T) {
	// GIVEN a WebHooksDefaults.
	tests := map[string]struct {
		webhooksDefaults *WebHooksDefaults
		want             string
	}{
		"nil": {
			webhooksDefaults: nil,
			want:             "",
		},
		"empty": {
			webhooksDefaults: &WebHooksDefaults{},
			want:             "{}",
		},
		"one empty and one nil": {
			webhooksDefaults: &WebHooksDefaults{
				"one": &Defaults{},
				"two": nil},
			want: test.TrimYAML(`
				one: {}`),
		},
		"one with data": {
			webhooksDefaults: &WebHooksDefaults{
				"one": NewDefaults(
					nil, nil, "", nil, nil, "", nil,
					"github",
					"https://example.com")},
			want: test.TrimYAML(`
				one:
					type: github
					url: https://example.com`),
		},
		"multiple": {
			webhooksDefaults: &WebHooksDefaults{
				"one": NewDefaults(
					nil, nil, "", nil, nil, "", nil,
					"github",
					"https://example.com"),
				"two": NewDefaults(
					nil, nil, "", nil, nil, "", nil,
					"gitlab",
					"https://example.com/other")},
			want: test.TrimYAML(`
				one:
					type: github
					url: https://example.com
				two:
					type: gitlab
					url: https://example.com/other`),
		},
		"quotes otherwise invalid YAML strings": {
			webhooksDefaults: &WebHooksDefaults{
				"invalid": NewDefaults(
					nil,
					&Headers{
						{Key: ">123", Value: "{pass}"}},
					"", nil, nil, "", nil, "", "")},
			want: test.TrimYAML(`
				invalid:
					custom_headers:
						- key: '>123'
							value: '{pass}'`),
		},
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

				// WHEN the WebHooks is stringified with String.
				got := tc.webhooksDefaults.String(prefix)

				// THEN the result is as expected.
				want = strings.TrimPrefix(want, "\n")
				if got != want {
					t.Fatalf("%s\n(prefix=%q)\nwant: %q\ngot:  %q",
						packageName, prefix, want, got)
				}
			}
		})
	}
}

func TestWebHooks_String(t *testing.T) {
	// GIVEN a WebHooks.
	tests := map[string]struct {
		webhooks *WebHooks
		want     string
	}{
		"nil": {
			webhooks: nil,
			want:     "",
		},
		"empty": {
			webhooks: &WebHooks{},
			want:     "{}\n",
		},
		"one": {
			webhooks: &WebHooks{
				"one": New(
					nil, nil,
					"",
					nil, nil,
					"one",
					nil, nil, nil,
					"", nil,
					"github", "https://example.com",
					nil, nil, nil)},
			want: test.TrimYAML(`
				one:
					type: github
					url: https://example.com
			`),
		},
		"multiple": {
			webhooks: &WebHooks{
				"one": New(
					nil, nil,
					"",
					nil, nil,
					"one",
					nil, nil, nil,
					"",
					nil,
					"github", "https://example.com",
					nil, nil, nil),
				"two": New(
					nil, nil,
					"",
					nil, nil,
					"two",
					nil, nil, nil,
					"",
					nil,
					"gitlab", "https://example.com/other",
					nil, nil, nil)},
			want: test.TrimYAML(`
				one:
					type: github
					url: https://example.com
				two:
					type: gitlab
					url: https://example.com/other
			`),
		},
		"quotes otherwise invalid YAML strings": {
			webhooks: &WebHooks{
				"invalid": New(
					nil,
					&Headers{
						{Key: ">123", Value: "{pass}"}},
					"",
					nil, nil,
					"invalid",
					nil, nil, nil,
					"",
					nil,
					"", "",
					nil, nil, nil)},
			want: test.TrimYAML(`
				invalid:
					custom_headers:
						- key: '>123'
							value: '{pass}'
			`)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the WebHooks is stringified with String.
			got := tc.webhooks.String()

			// THEN the result is as expected.
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestWebHooks_UnmarshalJSON(t *testing.T) {
	// GIVEN various JSON inputs to unmarshal into WebHooks.
	tests := map[string]struct {
		json     string
		wantErr  string
		wantKeys map[string]string
	}{
		"valid array with two items": {
			json: test.TrimJSON(`[
                {"name": "a", "type": "github"},
                {"name": "b", "type": "gitlab"}
            ]`),
			wantErr: `^$`,
			wantKeys: map[string]string{
				"a": "github",
				"b": "gitlab",
			},
		},
		"empty array becomes empty map": {
			json:     `[]`,
			wantErr:  `^$`,
			wantKeys: map[string]string{},
		},
		"null becomes empty map": {
			json:     `null`,
			wantErr:  `^$`,
			wantKeys: map[string]string{},
		},
		"duplicate ids - last wins": {
			json: test.TrimJSON(`[
                {"name": "dupe", "type": "github"},
                {"name": "dupe", "type": "gitlab"}
            ]`),
			wantErr: `^$`,
			wantKeys: map[string]string{
				"dupe": "gitlab",
			},
		},
		"invalid JSON": {
			json:    `{`,
			wantErr: `.+`,
		},
		"wrong shape (object instead of array)": {
			json: test.TrimJSON(`{
                "name": "a", "type": "github"
            }`),
			wantErr: `json: cannot unmarshal object.+$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN unmarshaling JSON into a WebHooks.
			var s WebHooks
			err := s.UnmarshalJSON([]byte(tc.json))

			// THEN errors produced match the regex.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.wantErr, e) {
				t.Fatalf("%s\nerror mismatch\nwant: %q\ngot: %q", packageName, tc.wantErr, e)
			}
			if e != "" {
				return
			}

			// AND map keys and types are as expected.
			if len(s) != len(tc.wantKeys) {
				t.Fatalf("%s\nlength mismatch\nwant: %d\ngot: %d", packageName, len(tc.wantKeys), len(s))
			}
			for id, wantType := range tc.wantKeys {
				got, ok := s[id]
				if !ok {
					t.Errorf("%s\nmissing key %q", packageName, id)
				}
				if got == nil {
					t.Errorf("%s\nvalue for key %q is nil", packageName, id)
				}
				if got.Type != wantType {
					t.Errorf("%s\nType mismatch for %q\nwant: %q\n got: %q", packageName, id, wantType, got.Type)
				}
				if got.ID != id {
					t.Errorf("%s\nID mismatch for key %q\nwant: %q\n got: %q", packageName, id, id, got.ID)
				}
			}
		})
	}
}

func TestWebHooks_MarshalJSON(t *testing.T) {
	// GIVEN various WebHooks states to marshal.
	tests := map[string]struct {
		webhooks *WebHooks
		wantStr  string
	}{
		"nil map -> null": {
			webhooks: nil,
			wantStr:  "null",
		},
		"empty map -> empty array": {
			webhooks: &WebHooks{},
			wantStr:  "[]",
		},
		"two items": {
			webhooks: &WebHooks{
				"a": &WebHook{Base: Base{Type: "github"}, ID: "a"},
				"b": &WebHook{Base: Base{Type: "gitlab"}, ID: "b"},
			},
			wantStr: test.TrimJSON(`[
                {"type": "github", "name": "a"},
                {"type": "gitlab", "name": "b"}
            ]`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN marshaling the WebHooks.
			data, err := tc.webhooks.MarshalJSON()
			if err != nil {
				t.Fatalf("%s\nMarshalJSON returned error: %v", packageName, err)
			}

			// THEN the result matches the expected JSON.
			dataStr := string(data)
			if dataStr != tc.wantStr {
				t.Errorf("%s\nJSON mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wantStr, dataStr)
			}
		})
	}
}
