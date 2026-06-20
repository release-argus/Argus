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

package webhook

import (
	"fmt"
	"testing"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

// ############
// # DECODING #
// ############

func TestDecodeDefaults(t *testing.T) {
	// GIVEN: data in a given format to Decode into a WebHook.
	tests := []struct {
		name         string
		format, data string
		want         string
		errRegex     string
	}{
		{
			name:   "JSON/empty",
			format: "json",
			data:   "",
			want:   "",
			errRegex: test.TrimYAML(`
				^jsontext:
					unexpected EOF$`,
			),
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:   "JSON/filled",
			format: "json",
			data: test.TrimJSON(`{
				"type": "github",
				"url": "https://example.com",
				"allow_invalid_certs": false,
				"headers": [
					{"key": "X-Header",  "value": "val" },
					{"key": "X-Another", "value": "val2" }
				],
				"secret": "foobar",
				"desired_status_code": 200,
				"delay": "1h2m3s",
				"max_tries": 4,
				"silent_fails": true,
				"other": "foo"
			}`),
			want: test.TrimYAML(`
				type: github
				url: https://example.com
				allow_invalid_certs: false
				headers:
					- key: X-Header
					  value: val
					- key: X-Another
					  value: val2
				secret: foobar
				desired_status_code: 200
				delay: 1h2m3s
				max_tries: 4
				silent_fails: true
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/filled",
			format: "yaml",
			data: test.TrimYAML(`
				type: github
				url: https://example.com
				allow_invalid_certs: false
				headers:
					- key: X-Header
					  value: val
					- key: X-Another
					  value: val2
				secret: foobar
				desired_status_code: 200
				delay: 1h2m3s
				max_tries: 4
				silent_fails: true
				other: foo
			`),
			want: test.TrimYAML(`
				type: github
				url: https://example.com
				allow_invalid_certs: false
				headers:
					- key: X-Header
					  value: val
					- key: X-Another
					  value: val2
				secret: foobar
				desired_status_code: 200
				delay: 1h2m3s
				max_tries: 4
				silent_fails: true
			`),
			errRegex: `^$`,
		},
		{
			name:     "JSON/invalid data type",
			format:   "json",
			data:     `"url: [https://example.com]"`,
			errRegex: `^json: .*unmarshal .*`,
		},
		{
			name:     "YAML/invalid data type",
			format:   "yaml",
			data:     `url: [https://example.com]`,
			errRegex: `^[^\s]+ .*unmarshal .*`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				DecodeDefaults,
				tc.format, tc.data,
				func(v *Defaults) string { return v.String("") },
				tc.want,
				tc.errRegex,
				packageName,
				"DecodeDefaults",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestWebHooks_MarshalJSON(t *testing.T) {
	// GIVEN: various WebHooks states to marshal.
	tests := []struct {
		name     string
		webhooks *WebHooks
		wantStr  string
	}{
		{
			name:     "nil map -> null",
			webhooks: nil,
			wantStr:  "null",
		},
		{
			name:     "empty map -> empty array",
			webhooks: &WebHooks{},
			wantStr:  "[]",
		},
		{
			name: "two items",
			webhooks: &WebHooks{
				"a": &WebHook{
					Base: Base{
						Type: "github",
					},
					ID: "a",
				},
				"b": &WebHook{
					Base: Base{
						Type: "gitlab",
					},
					ID: "b",
				},
			},
			wantStr: test.TrimJSON(`[
				{"type": "github", "name": "a"},
				{"type": "gitlab", "name": "b"}
			]`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prefix := fmt.Sprintf("%s\nWebHooks.MarshalJSON()", packageName)

			// WHEN: marshaling the WebHooks.
			data, err := tc.webhooks.MarshalJSON()
			if err != nil {
				t.Fatalf("%s returned error: %v", prefix, err)
			}

			if got := string(data); got != tc.wantStr {
				t.Errorf(
					"%s value mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wantStr,
				)
			}
		})
	}
}

func TestWebHooks_UnmarshalJSON(t *testing.T) {
	// GIVEN: a string in a given format to unmarshal into Defaults.
	tests := []struct {
		name     string
		data     string
		errRegex string
		wantKeys map[string]string
	}{
		{
			name: "valid array with two items",
			data: test.TrimJSON(`[
				{"name": "a", "type": "github"},
				{"name": "b", "type": "gitlab"}
			]`),
			errRegex: `^$`,
			wantKeys: map[string]string{
				"a": "github",
				"b": "gitlab",
			},
		},
		{
			name:     "empty array becomes empty map",
			data:     `[]`,
			errRegex: `^$`,
			wantKeys: map[string]string{},
		},
		{
			name:     "null becomes empty map",
			data:     `null`,
			errRegex: `^$`,
			wantKeys: map[string]string{},
		},
		{
			name: "duplicate ids - last wins",
			data: test.TrimJSON(`[
				{"name": "dupe", "type": "github"},
				{"name": "dupe", "type": "gitlab"}
			]`),
			errRegex: `^$`,
			wantKeys: map[string]string{
				"dupe": "gitlab",
			},
		},
		{
			name:     "invalid JSON",
			data:     `{`,
			errRegex: `.+`,
		},
		{
			name: "wrong shape (object instead of array)",
			data: test.TrimJSON(`{
				"name": "a",
				"type": "github"
			}`),
			errRegex: `json: .*unmarshal.* object.+$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: unmarshaling JSON into a WebHooks.
			var s WebHooks
			err := s.UnmarshalJSON([]byte(tc.data))

			prefix := fmt.Sprintf(
				"%s\nWebHooks.UnmarshalJSON(%q)",
				packageName, tc.data,
			)

			// THEN: errors produced match the regex.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			if e != "" {
				return
			}

			// AND: map keys and types are as expected.
			if got, want := len(s), len(tc.wantKeys); got != want {
				t.Fatalf(
					"%s length mismatch\ngot:  %d (%+v)\nwant: %d (%+v)",
					prefix,
					len(s), s,
					len(tc.wantKeys), tc.wantKeys,
				)
			}
			for id, wantType := range tc.wantKeys {
				got, ok := s[id]
				if !ok {
					t.Errorf("%s missing key %q", prefix, id)
					continue
				}
				if got == nil {
					t.Errorf("%s value for key %q is nil", prefix, id)
				}
				if got.Type != wantType {
					t.Errorf(
						"%s .Type mismatch for %q\ngot:  %q\nwant: %q",
						prefix, id,
						got.Type, wantType,
					)
				}
				if got.ID != id {
					t.Errorf(
						"%s .ID mismatch for key %q\ngot:  %q\nwant: %q",
						prefix, id,
						got.ID, id,
					)
				}
			}
		})
	}
}

func assertFailsWebHookState(
	t *testing.T,
	got *status.FailsWebHook,
	wantFails map[string]*bool,
	prefix, target string,
) {
	t.Helper()

	for key, wantFail := range wantFails {
		// Each Fail should be value-equal.
		gotFail := got.Get(key)
		gotStr := test.StringifyPtr(gotFail)
		wantStr := test.StringifyPtr(wantFail)
		if gotStr != wantStr {
			t.Errorf(
				"%s %s[%q] mismatch\ngot:  %s\nwant: %s",
				prefix, target, key,
				gotStr, wantStr,
			)
		}
		// And mutating the fail should not affect the original map.
		wantFails[key] = test.Ptr(!*wantFail)
		gotFail = got.Get(key)
		gotStr = test.StringifyPtr(gotFail)
		if gotStr != wantStr {
			t.Errorf(
				"%s %s[%q] changed from mutating original map\ngot:  %s\nwant: %s",
				prefix, target, key,
				gotStr, wantStr,
			)
		}
	}
}

func testWebHookForCopy(t *testing.T, svcStatus *status.Status, id string) *WebHook {
	t.Helper()

	main := &Defaults{
		Base: Base{
			Type: "github",
			URL:  "https://main.example.com",
		},
	}
	defaults := &Defaults{
		Base: Base{
			Type: "github",
			URL:  "https://defaults.example.com",
		},
	}
	hardDefaults := &Defaults{
		Base: Base{
			Type: "github",
			URL:  "https://hard-defaults.example.com",
		},
	}

	notifiers := Notifiers{
		Shoutrrr: &shoutrrr.Shoutrrrs{},
	}

	wh := New(
		test.Ptr(true),
		Headers{
			{Key: "X-Header", Value: "A"},
			{Key: "X-Other", Value: "B"},
		},
		"1s",
		test.Ptr(uint16(202)),
		&svcStatus.Fails.WebHook,
		id,
		test.Ptr(uint8(3)),
		notifiers,
		test.Ptr("5m"),
		"secret",
		test.Ptr(false),
		"github",
		"https://example.com",
		main,
		defaults,
		hardDefaults,
	)
	wh.ServiceStatus = svcStatus
	wh.Failed.Set(id, test.Ptr(false))
	wh.Failed.SetNextRunnable(id, time.Unix(123, 0))

	return wh
}

func TestWebHook_Copy(t *testing.T) {
	tests := []struct {
		name     string
		wantNil  bool
		wantFail map[string]*bool
		// Mutations to verify that Copy doesn't alias the underlying map.
		mutateOriginalHeaderValue string
		mutateCopyHeaderValue     string
	}{
		{
			name:    "nil receiver",
			wantNil: true,
		},
		{
			name:                      "copies fields and deep-copies pointers, slices, and fails",
			wantFail:                  map[string]*bool{"notify": test.Ptr(false)},
			mutateOriginalHeaderValue: "mutated",
			mutateCopyHeaderValue:     "copy-mutated",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// GIVEN: a Status.
			origStatus, _ := statustest.New("yaml", nil)
			origStatus.Fails.WebHook.Init(2)
			copyStatus, _ := statustest.New("yaml", nil)
			copyStatus.Fails.WebHook.Init(2)

			// AND: Notifiers.
			copyNotifiers := Notifiers{Shoutrrr: &shoutrrr.Shoutrrrs{}}

			// AND: a WebHook.
			var orig *WebHook
			if !tc.wantNil {
				orig = testWebHookForCopy(t, origStatus, "notify")
			}

			wantStr := decode.ToYAMLString(orig, "")

			// WHEN: Copy is called on it.
			got := orig.Copy(copyStatus, copyNotifiers)

			prefix := fmt.Sprintf(
				"%s\nWebHook.Copy(status=%p, notifiers=%v)",
				packageName, copyNotifiers, copyNotifiers,
			)

			// THEN: nil handling.
			if tc.wantNil {
				if got != nil {
					t.Fatalf("%s got %v want nil", prefix, got)
				}
				return
			}

			// AND: the copy is non-nil.
			if got == nil {
				t.Fatalf("%s got nil want non-nil", prefix)
			}

			// AND: the copy is distinct.
			if got == orig {
				t.Fatalf(
					"%s should return a distinct copy\ngot:  %p\nwant: %p",
					prefix, got, orig,
				)
			}

			// AND: the copy unmarshals the same.
			if gotStr := decode.ToYAMLString(got, ""); gotStr != wantStr {
				t.Fatalf(
					"%s stringified mismatch:\ngot;  %q\nwant: %q",
					prefix, gotStr, wantStr,
				)
			}

			// AND: the fields are copied as expected.
			fieldTests := []test.FieldAssertion{
				{Name: "Type", Got: got.Type, Want: orig.Type, Mode: test.CompareEqual},
				{Name: "URL", Got: got.URL, Want: orig.URL, Mode: test.CompareEqual},
				{Name: "Secret", Got: got.Secret, Want: orig.Secret, Mode: test.CompareEqual},
				{Name: "Delay", Got: got.Delay, Want: orig.Delay, Mode: test.CompareEqual},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "WebHook"); err != nil {
				t.Fatal(err)
			}

			// AND: copied pointers should be value-equal and non-aliased.
			fieldTests = []test.FieldAssertion{
				{Name: "AllowInvalidCerts", Got: got.AllowInvalidCerts, Want: orig.AllowInvalidCerts, Mode: test.CompareDifferentPointer},
				{Name: "DesiredStatusCode", Got: got.DesiredStatusCode, Want: orig.DesiredStatusCode, Mode: test.CompareDifferentPointer},
				{Name: "MaxTries", Got: got.MaxTries, Want: orig.MaxTries, Mode: test.CompareDifferentPointer},
				{Name: "SilentFails", Got: got.SilentFails, Want: orig.SilentFails, Mode: test.CompareDifferentPointer},
				{Name: "ParentInterval", Got: got.ParentInterval, Want: orig.ParentInterval, Mode: test.CompareDifferentPointer},
				{Name: "Failed", Got: got.Failed, Want: orig.Failed, Mode: test.CompareDifferentPointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "WebHook"); err != nil {
				t.Fatal(err)
			}

			// AND: non-base fields are copied/shared.
			fieldTests = []test.FieldAssertion{
				{Name: "ID", Got: got.ID, Want: orig.ID, Mode: test.CompareEqual},
				{Name: "ServiceStatus", Got: got.ServiceStatus, Want: copyStatus, Mode: test.CompareSamePointer},
				{Name: "Notifiers", Got: got.Notifiers.Shoutrrr, Want: copyNotifiers.Shoutrrr, Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "WebHook"); err != nil {
				t.Fatal(err)
			}

			// AND: defaults pointers are shared.
			fieldTests = []test.FieldAssertion{
				{Name: "Main", Got: got.Main, Want: orig.Main, Mode: test.CompareSamePointer},
				{Name: "Defaults", Got: got.Defaults, Want: orig.Defaults, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: got.HardDefaults, Want: orig.HardDefaults, Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "WebHook"); err != nil {
				t.Fatal(err)
			}

			// AND: Failed should be deep-copied.
			if got.Failed == nil || orig.Failed == nil {
				t.Fatalf("%s Failed unexpectedly nil", prefix)
			}
			if got.Failed == orig.Failed {
				t.Errorf("%s Failed should not alias", prefix)
			}
			assertFailsWebHookState(
				t,
				got.Failed,
				map[string]*bool{"notify": test.Ptr(false)},
				prefix,
				"Failed",
			)
			if got.Failed.NextRunnable("notify") != orig.Failed.NextRunnable("notify") {
				t.Errorf(
					"%s Failed.NextRunnable(%q) mismatch\ngot:  %v\nwant: %v",
					prefix, "notify",
					got.Failed.NextRunnable("notify"),
					orig.Failed.NextRunnable("notify"),
				)
			}
		})
	}
}

func TestWebHooks_Copy(t *testing.T) {
	tests := []struct {
		name     string
		webhooks *WebHooks
		wantNil  bool
		wantLen  int
		// Mutations to verify that Copy doesn't alias the underlying map.
		reassignOriginalKey string
	}{
		{
			name:     "nil receiver",
			webhooks: nil,
			wantNil:  true,
		},
		{
			name:     "pointer to nil map becomes empty map",
			webhooks: func() *WebHooks { var wh WebHooks; return &wh }(),
			wantLen:  0,
		},
		{
			name: "copies each entry",
			webhooks: func() *WebHooks {
				origStatus, _ := statustest.New("yaml", nil)
				origStatus.Fails.WebHook.Init(2)
				wh := WebHooks{
					"foo": testWebHookForCopy(t, origStatus, "foo"),
					"bar": testWebHookForCopy(t, origStatus, "bar"),
				}
				return &wh
			}(),
			wantLen: 2,
		},
		{
			name: "reassigning original map entry does not affect copy",
			webhooks: func() *WebHooks {
				origStatus, _ := statustest.New("yaml", nil)
				origStatus.Fails.WebHook.Init(2)
				wh := WebHooks{
					"foo": testWebHookForCopy(t, origStatus, "foo"),
				}
				return &wh
			}(),
			wantLen:             1,
			reassignOriginalKey: "foo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// GIVEN: a Status.
			copyStatus, _ := statustest.New("yaml", nil)
			copyStatus.Fails.WebHook.Init(2)

			// AND: Notifiers.
			copyNotifiers := Notifiers{Shoutrrr: &shoutrrr.Shoutrrrs{}}

			// WHEN: Copy is called on it.
			got := tc.webhooks.Copy(copyStatus, copyNotifiers)

			prefix := fmt.Sprintf("%s\nWebHooks.Copy()", packageName)

			// THEN: nil handling.
			if tc.wantNil {
				if got != nil {
					t.Fatalf("%s got %v want nil", prefix, got)
				}
				return
			}

			// AND: the lengths match.
			if gotLen := len(got); gotLen != tc.wantLen {
				t.Fatalf("%s length mismatch\ngot:  %d\nwant: %d", prefix, gotLen, tc.wantLen)
			}

			// AND: each entry is a distinct copy.
			orig := *tc.webhooks
			for key, want := range orig {
				gotEntry := got[key]
				if gotEntry == nil {
					t.Fatalf("%s missing copied entry %q", prefix, key)
				}
				fieldTests := []test.FieldAssertion{
					{Name: "ID", Got: gotEntry.ID, Want: want.ID, Mode: test.CompareEqual},
					{Name: "Notifiers.Shoutrrr", Got: gotEntry.Notifiers.Shoutrrr, Want: copyNotifiers.Shoutrrr, Mode: test.CompareSamePointer},
					{Name: "ServiceStatus", Got: gotEntry.ServiceStatus, Want: copyStatus, Mode: test.CompareSamePointer},
				}
				if err := test.AssertFields(t, fieldTests, prefix, "WebHook"); err != nil {
					t.Fatal(err)
				}
			}

			// AND: reassigning original map entry does not affect copy.
			if tc.reassignOriginalKey != "" {
				newStatus, _ := statustest.New("yaml", nil)
				newStatus.Fails.WebHook.Init(2)
				replacement := testWebHookForCopy(t, newStatus, "replacement")
				orig[tc.reassignOriginalKey] = replacement
				if got[tc.reassignOriginalKey] == replacement {
					t.Fatalf(
						"%s entry %q should not point at replacement value",
						prefix, tc.reassignOriginalKey,
					)
				}
			}
		})
	}
}

// #########
// # STATE #
// #########

func TestWebHooksDefaults_IsZero(t *testing.T) {
	// GIVEN: a WebHooksDefaults.
	tests := []struct {
		name     string
		defaults *WebHooksDefaults
		want     bool
	}{
		{
			name:     "empty/0 items",
			defaults: &WebHooksDefaults{},
			want:     true,
		},
		{
			name: "empty/1 item",
			defaults: &WebHooksDefaults{
				"a": &Defaults{},
			},
			want: true,
		},
		{
			name: "empty/2 items",
			defaults: &WebHooksDefaults{
				"a": &Defaults{},
				"b": &Defaults{},
			},
			want: true,
		},
		{
			name: "non-empty/1 item",
			defaults: &WebHooksDefaults{
				"a": &Defaults{
					Base: Base{
						Type: "github",
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty/1 item",
			defaults: &WebHooksDefaults{
				"a": &Defaults{
					Base: Base{
						Type: "github",
					},
				},
				"b": &Defaults{
					Base: Base{
						Type: "gitkav",
					},
				},
			},
			want: false,
		},
		{
			name: "mixed",
			defaults: &WebHooksDefaults{
				"a": &Defaults{},
				"b": &Defaults{
					Base: Base{
						Type: "github",
					},
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero is called on it.
			got := tc.defaults.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nWebHooksDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestDefaults_IsZero(t *testing.T) {
	// GIVEN: a Defaults.
	tests := []struct {
		name     string
		defaults *Defaults
		want     bool
	}{
		{
			name:     "nil",
			defaults: nil,
			want:     true,
		},
		{
			name:     "empty",
			defaults: &Defaults{},
			want:     true,
		},
		{
			name: "non-empty/Type",
			defaults: &Defaults{
				Base: Base{
					Type: "github",
				},
			},
			want: false,
		},
		{
			name: "non-empty/URL",
			defaults: &Defaults{
				Base: Base{
					URL: "https://example.com",
				},
			},
			want: false,
		},
		{
			name: "non-empty/AllowInvalidCerts",
			defaults: &Defaults{
				Base: Base{
					AllowInvalidCerts: test.Ptr(false),
				},
			},
			want: false,
		},
		{
			name: "non-empty/Headers",
			defaults: &Defaults{
				Base: Base{
					Headers: Headers{
						{Key: "X-Header", Value: "val"},
						{Key: "X-Another", Value: "val2"},
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty/Secret",
			defaults: &Defaults{
				Base: Base{
					Secret: "foobar",
				},
			},
			want: false,
		},
		{
			name: "non-empty/DesiredStatusCode",
			defaults: &Defaults{
				Base: Base{
					DesiredStatusCode: test.Ptr[uint16](200),
				},
			},
			want: false,
		},
		{
			name: "non-empty/Delay",
			defaults: &Defaults{
				Base: Base{
					Delay: "1h2m3s",
				},
			},
			want: false,
		},
		{
			name: "non-empty/MaxTries",
			defaults: &Defaults{
				Base: Base{
					MaxTries: test.Ptr[uint8](4),
				},
			},
			want: false,
		},
		{
			name: "non-empty/SilentFails",
			defaults: &Defaults{
				Base: Base{
					SilentFails: test.Ptr(true),
				},
			},
			want: false,
		},
		{
			name: "non-empty/all",
			defaults: &Defaults{
				Base: Base{
					Type:              "github",
					URL:               "https://example.com",
					AllowInvalidCerts: test.Ptr(false),
					Headers: Headers{
						{Key: "X-Header", Value: "val"},
						{Key: "X-Another", Value: "val2"},
					},
					Secret:            "foobar",
					DesiredStatusCode: test.Ptr[uint16](200),
					Delay:             "1h2m3s",
					MaxTries:          test.Ptr[uint8](4),
					SilentFails:       test.Ptr(true),
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.defaults.IsZero()
			if got != tc.want {
				t.Errorf(
					"%s\nDefaults.IsZero() mismatch\ngot:  %t\nwant: %v",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestWebHooks_IsZero(t *testing.T) {
	// GIVEN: a WebHooks.
	tests := []struct {
		name     string
		webhooks *WebHooks
		want     bool
	}{
		{
			name:     "nil",
			webhooks: nil,
			want:     true,
		},
		{
			name:     "empty",
			webhooks: &WebHooks{},
			want:     true,
		},
		{
			name: "non-empty/1 item",
			webhooks: &WebHooks{
				"a": &WebHook{},
			},
			want: false,
		},
		{
			name: "non-empty/2 items",
			webhooks: &WebHooks{
				"a": &WebHook{},
				"b": &WebHook{},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero is called on it.
			got := tc.webhooks.IsZero()

			if got != tc.want {
				t.Errorf(
					"%s\nWebHooks.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestWebHook_IsDefault(t *testing.T) {
	// GIVEN: a WebHook.
	tests := []struct {
		name    string
		webhook *WebHook
		want    bool
	}{
		{
			name:    "empty",
			webhook: &WebHook{},
			want:    true,
		},
		{
			name: "non-empty/Type",
			webhook: &WebHook{
				Base: Base{
					Type: "github",
				},
			},
			want: false,
		},
		{
			name: "non-empty/URL",
			webhook: &WebHook{
				Base: Base{
					URL: "https://example.com",
				},
			},
			want: false,
		},
		{
			name: "non-empty/AllowInvalidCerts",
			webhook: &WebHook{
				Base: Base{
					AllowInvalidCerts: test.Ptr(false),
				},
			},
			want: false,
		},
		{
			name: "non-empty/Headers",
			webhook: &WebHook{
				Base: Base{
					Headers: Headers{
						{Key: "X-Header", Value: "val"},
						{Key: "X-Another", Value: "val2"},
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty/Secret",
			webhook: &WebHook{
				Base: Base{
					Secret: "foobar",
				},
			},
			want: false,
		},
		{
			name: "non-empty/DesiredStatusCode",
			webhook: &WebHook{
				Base: Base{
					DesiredStatusCode: test.Ptr[uint16](200),
				},
			},
			want: false,
		},
		{
			name: "non-empty/Delay",
			webhook: &WebHook{
				Base: Base{
					Delay: "1h2m3s",
				},
			},
			want: false,
		},
		{
			name: "non-empty/MaxTries",
			webhook: &WebHook{
				Base: Base{
					MaxTries: test.Ptr[uint8](4),
				},
			},
			want: false,
		},
		{
			name: "non-empty/SilentFails",
			webhook: &WebHook{
				Base: Base{
					SilentFails: test.Ptr(true),
				},
			},
			want: false,
		},
		{
			name: "non-empty/all",
			webhook: &WebHook{
				Base: Base{
					Type:              "github",
					URL:               "https://example.com",
					AllowInvalidCerts: test.Ptr(false),
					Headers: Headers{
						{Key: "X-Header", Value: "val"},
						{Key: "X-Another", Value: "val2"},
					},
					Secret:            "foobar",
					DesiredStatusCode: test.Ptr[uint16](200),
					Delay:             "1h2m3s",
					MaxTries:          test.Ptr[uint8](4),
					SilentFails:       test.Ptr(true),
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsDefault is called on it.
			got := tc.webhook.IsDefault()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nWebHook.IsDefault() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

// #############
// # STRINGIFY #
// #############

func TestDefaults_String(t *testing.T) {
	// GIVEN: Defaults.
	tests := []struct {
		name     string
		defaults *Defaults
		want     string
	}{
		{
			name:     "empty",
			defaults: &Defaults{},
			want:     "{}\n",
		},
		{
			name: "filled",
			defaults: test.Must(t, func() (*Defaults, error) {
				return DecodeDefaults(
					"yaml", []byte(test.TrimYAML(`
						allow_invalid_certs: false
						headers:
							- key: X-Header
							  value: val
							- key: X-Another
							  value: val2
						delay: 1h1m1s
						desired_status_code: 200
						max_tries: 4
						secret: foobar
						silent_fails: true
						type: github
						url: https://example.com
					`)),
				)
			}),
			want: test.TrimYAML(`
				type: github
				url: https://example.com
				allow_invalid_certs: false
				headers:
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
		{
			name: "quotes otherwise invalid YAML strings",
			defaults: test.Must(t, func() (*Defaults, error) {
				return DecodeDefaults(
					"yaml", []byte(test.TrimYAML(`
						headers:
							- key: '>123'
							  value: '{pass}'
					`)),
				)
			}),
			want: test.TrimYAML(`
				headers:
					- key: '>123'
						value: '{pass}'
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				packageName,
				tc.defaults.String,
				tc.want,
			)
		})
	}
}

func TestWebHooksDefaults_String(t *testing.T) {
	// GIVEN: a WebHooksDefaults.
	tests := []struct {
		name             string
		webhooksDefaults *WebHooksDefaults
		want             string
	}{
		{
			name:             "nil",
			webhooksDefaults: nil,
			want:             "",
		},
		{
			name:             "empty",
			webhooksDefaults: &WebHooksDefaults{},
			want:             "{}\n",
		},
		{
			name: "two empty",
			webhooksDefaults: &WebHooksDefaults{
				"one": &Defaults{},
				"two": &Defaults{},
				// "two": nil,
			},
			want: test.TrimYAML(`
				one: {}
				two: {}
			`),
		},
		{
			name: "one with data",
			webhooksDefaults: &WebHooksDefaults{
				"one": test.Must(t, func() (*Defaults, error) {
					return DecodeDefaults(
						"yaml", []byte(test.TrimYAML(`
							type: github
							url: https://example.com
						`)),
					)
				}),
			},
			want: test.TrimYAML(`
				one:
					type: github
					url: https://example.com
			`),
		},
		{
			name: "multiple",
			webhooksDefaults: &WebHooksDefaults{
				"one": test.Must(t, func() (*Defaults, error) {
					return DecodeDefaults(
						"yaml", []byte(test.TrimYAML(`
							type: github
							url: https://example.com
						`)),
					)
				}),
				"two": test.Must(t, func() (*Defaults, error) {
					return DecodeDefaults(
						"yaml", []byte(test.TrimYAML(`
							type: gitlab
							url: https://example.com/other
						`)),
					)
				}),
			},
			want: test.TrimYAML(`
				one:
					type: github
					url: https://example.com
				two:
					type: gitlab
					url: https://example.com/other
			`),
		},
		{
			name: "quotes otherwise invalid YAML strings",
			webhooksDefaults: &WebHooksDefaults{
				"invalid": test.Must(t, func() (*Defaults, error) {
					return DecodeDefaults(
						"yaml", []byte(test.TrimYAML(`
							headers:
								- key: '>123'
								  value: '{pass}'
						`)),
					)
				}),
			},
			want: test.TrimYAML(`
				invalid:
					headers:
						- key: '>123'
							value: '{pass}'
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				packageName,
				tc.webhooksDefaults.String,
				tc.want,
			)
		})
	}
}

func TestWebHooks_String(t *testing.T) {
	// GIVEN: webHooks.
	tests := []struct {
		name     string
		webhooks *WebHooks
		want     string
	}{
		{
			name:     "nil",
			webhooks: nil,
			want:     "",
		},
		{
			name:     "empty",
			webhooks: &WebHooks{},
			want:     "{}\n",
		},
		{
			name: "one",
			webhooks: &WebHooks{
				"one": New(
					nil, nil,
					"",
					nil, nil,
					"one",
					nil, Notifiers{}, nil,
					"", nil,
					"github", "https://example.com",
					nil, nil, nil,
				),
			},
			want: test.TrimYAML(`
				one:
					type: github
					url: https://example.com
			`),
		},
		{
			name: "multiple",
			webhooks: &WebHooks{
				"one": New(
					nil, nil,
					"",
					nil, nil,
					"one",
					nil, Notifiers{}, nil,
					"",
					nil,
					"github", "https://example.com",
					nil, nil, nil,
				),
				"two": New(
					nil, nil,
					"",
					nil, nil,
					"two",
					nil, Notifiers{}, nil,
					"",
					nil,
					"gitlab", "https://example.com/other",
					nil, nil, nil,
				),
			},
			want: test.TrimYAML(`
				one:
					type: github
					url: https://example.com
				two:
					type: gitlab
					url: https://example.com/other
			`),
		},
		{
			name: "quotes otherwise invalid YAML strings",
			webhooks: &WebHooks{
				"invalid": New(
					nil,
					Headers{
						{Key: ">123", Value: "{pass}"},
					},
					"",
					nil, nil,
					"invalid",
					nil, Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			want: test.TrimYAML(`
				invalid:
					headers:
						- key: '>123'
							value: '{pass}'
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: the WebHooks is stringified with String.
			got := tc.webhooks.String()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nWebHooks.String() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestWebHook_String(t *testing.T) {
	// GIVEN: a WebHook.
	tests := []struct {
		name    string
		webhook *WebHook
		want    string
	}{
		{
			name:    "nil",
			webhook: nil,
			want:    "",
		},
		{
			name:    "empty",
			webhook: &WebHook{},
			want:    "{}\n",
		},
		{
			name: "filled",
			webhook: New(
				test.Ptr(false),
				Headers{
					{Key: "X-Header", Value: "val"},
					{Key: "X-Another", Value: "val2"},
				},
				"1h1m1s",
				test.Ptr[uint16](200),
				nil,
				"filled",
				test.Ptr[uint8](4),
				Notifiers{
					Shoutrrr: &shoutrrr.Shoutrrrs{
						"foo": shoutrrr.New(
							nil,
							"", "discord",
							nil, nil, nil,
							nil, nil, nil,
						),
					},
				},
				test.Ptr("3h2m1s"),
				"foobar",
				test.Ptr(true),
				"github", "https://example.com",
				test.Must(t, func() (*Defaults, error) {
					return DecodeDefaults("yaml", []byte("allow_invalid_certs: false"))
				}),
				test.Must(t, func() (*Defaults, error) {
					return DecodeDefaults("yaml", []byte("allow_invalid_certs: true"))
				}),
				test.Must(t, func() (*Defaults, error) {
					return DecodeDefaults("yaml", []byte("allow_invalid_certs: true"))
				}),
			),
			want: test.TrimYAML(`
				type: github
				url: https://example.com
				allow_invalid_certs: false
				headers:
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
		{
			name: "quotes otherwise invalid YAML strings",
			webhook: New(
				nil,
				Headers{
					{Key: ">123", Value: "{pass}"},
				},
				"",
				nil, nil,
				"wh",
				nil, Notifiers{}, nil,
				"",
				nil,
				"", "",
				nil, nil, nil,
			),
			want: test.TrimYAML(`
				headers:
					- key: '>123'
						value: '{pass}'
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				packageName,
				tc.webhook.String,
				tc.want,
			)
		})
	}
}
