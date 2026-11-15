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
				"DecodeDefaults",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestWebhooks_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		webhooks *Webhooks
		wantStr  string
	}{
		{
			name:     "nil map -> null",
			webhooks: nil,
			wantStr:  "null",
		},
		{
			name:     "empty map -> empty array",
			webhooks: &Webhooks{},
			wantStr:  "[]",
		},
		{
			name: "two items",
			webhooks: &Webhooks{
				"a": &Webhook{
					Type: "github",
					ID:   "a",
				},
				"b": &Webhook{
					Type: "gitlab",
					ID:   "b",
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

			data, err := tc.webhooks.MarshalJSON()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got := string(data); got != tc.wantStr {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.wantStr)
			}
		})
	}
}

func TestWebhooks_UnmarshalJSON(t *testing.T) {
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
			name: "duplicate names rejected",
			data: test.TrimJSON(`[
				{"name": "dupe", "type": "github"},
				{"name": "dupe", "type": "gitlab"}
			]`),
			errRegex: test.TrimYAML(`
				^webhook\[1\]:
					name: "dupe" <invalid> \(must be unique\)$`,
			),
		},
		{
			name: "missing name rejected",
			data: test.TrimJSON(`[
				{"name": "a", "type": "github"},
				{"type": "gitlab"}
			]`),
			errRegex: test.TrimYAML(`
				^webhook\[1\]:
					name: <required>$`,
			),
		},
		{
			name: "empty name rejected",
			data: `[{"name": "", "type": "github"}]`,
			errRegex: test.TrimYAML(`
				^webhook\[0\]:
					name: <required>$`,
			),
		},
		{
			name:     "invalid JSON",
			data:     `[`,
			errRegex: `^jsontext: unexpected EOF after offset`,
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

			var s Webhooks
			err := s.UnmarshalJSON([]byte(tc.data))
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			if got, want := len(s), len(tc.wantKeys); got != want {
				t.Fatalf(
					"length mismatch\ngot:  %d %q\nwant: %d %q",
					got, util.SortedKeys(s),
					want, util.SortedKeys(tc.wantKeys),
				)
			}
			for id, wantType := range tc.wantKeys {
				got, ok := s[id]
				if !ok {
					t.Errorf("Webhooks[%q] mismatch\ngot:  absent\nwant: present", id)
					continue
				}
				if got == nil {
					t.Errorf("Webhooks[%q] mismatch\ngot:  nil\nwant: non-nil", id)
					continue
				}
				if got.Type != wantType {
					t.Errorf(
						"Webhooks[%q].Type mismatch\ngot:  %q\nwant: %q",
						id, got.Type, wantType,
					)
				}
				if got.ID != id {
					t.Errorf("Webhooks[%q].ID mismatch\ngot:  %q\nwant: %q", id, got.ID, id)
				}
			}
		})
	}
}

// testCopyStatus returns a Status with Webhook fails initialised.
func testCopyStatus(t *testing.T) *status.Status {
	t.Helper()

	svcStatus, _ := statustest.New("yaml", nil)
	svcStatus.Fails.Webhook.Init(2)

	return svcStatus
}

func testWebhookForCopy(svcStatus *status.Status, id string) *Webhook {
	main := &Defaults{
		Type: "github",
		URL:  "https://main.example.com",
	}
	defaults := &Defaults{
		Type: "github",
		URL:  "https://defaults.example.com",
	}
	hardDefaults := &Defaults{
		Type: "github",
		URL:  "https://hard-defaults.example.com",
	}

	notifiers := Notifiers{
		Shoutrrr: &shoutrrr.Shoutrrrs{},
	}

	wh := New(
		new(true),
		Headers{
			{Key: "X-Header", Value: "A"},
			{Key: "X-Other", Value: "B"},
		},
		"1s",
		new(uint16(202)),
		&svcStatus.Fails.Webhook,
		id,
		new(uint8(3)),
		notifiers,
		new("5m"),
		"secret",
		new(false),
		"github",
		"https://example.com",
		main,
		defaults,
		hardDefaults,
	)
	wh.ServiceStatus = svcStatus
	wh.Failed.Set(id, new(false))
	wh.Failed.SetNextRunnable(id, time.Unix(123, 0))

	return wh
}

func TestWebhooks_Copy(t *testing.T) {
	tests := []struct {
		name     string
		webhooks *Webhooks
		wantLen  int
		// Mutations to verify that Copy doesn't alias the underlying map.
		reassignOriginalKey string
	}{
		{
			name:     "nil receiver",
			webhooks: nil,
		},
		{
			name:     "pointer to nil map becomes empty map",
			webhooks: func() *Webhooks { var wh Webhooks; return &wh }(),
			wantLen:  0,
		},
		{
			name: "copies each entry",
			webhooks: func() *Webhooks {
				origStatus := testCopyStatus(t)
				wh := Webhooks{
					"foo": testWebhookForCopy(origStatus, "foo"),
					"bar": testWebhookForCopy(origStatus, "bar"),
				}
				return &wh
			}(),
			wantLen: 2,
		},
		{
			name: "reassigning original map entry does not affect copy",
			webhooks: func() *Webhooks {
				origStatus := testCopyStatus(t)
				wh := Webhooks{
					"foo": testWebhookForCopy(origStatus, "foo"),
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

			copyStatus := testCopyStatus(t)
			copyNotifiers := Notifiers{Shoutrrr: &shoutrrr.Shoutrrrs{}}

			got := tc.webhooks.Copy(copyStatus, copyNotifiers)

			if tc.webhooks == nil {
				if got != nil {
					t.Fatal("copy mismatch\ngot:  non-nil\nwant: nil")
				}
				return
			}

			if gotLen := len(got); gotLen != tc.wantLen {
				t.Fatalf(
					"length mismatch\ngot:  %d %q\nwant: %d %q",
					gotLen, util.SortedKeys(got),
					tc.wantLen, util.SortedKeys(*tc.webhooks),
				)
			}

			orig := *tc.webhooks
			for key, want := range orig {
				t.Run(key, func(t *testing.T) {
					// Serial: reads the original's entries, which the parent reassigns after the loop.

					gotEntry := got[key]
					if gotEntry == nil {
						t.Fatal("Webhook mismatch\ngot:  nil\nwant: non-nil")
					}
					fieldTests := []test.FieldAssertion{
						{Name: "ID", Got: gotEntry.ID, Want: want.ID, Mode: test.CompareEqual},
						{Name: "Notifiers.Shoutrrr", Got: gotEntry.Notifiers.Shoutrrr, Want: copyNotifiers.Shoutrrr, Mode: test.CompareSamePointer},
						{Name: "ServiceStatus", Got: gotEntry.ServiceStatus, Want: copyStatus, Mode: test.CompareSamePointer},
					}
					if testErr := test.AssertFields(t, fieldTests, "Webhook"); testErr != nil {
						t.Fatal(testErr)
					}
				})
			}

			if tc.reassignOriginalKey != "" {
				newStatus := testCopyStatus(t)
				replacement := testWebhookForCopy(newStatus, "replacement")
				orig[tc.reassignOriginalKey] = replacement
				if got[tc.reassignOriginalKey] == replacement {
					t.Fatalf(
						"Webhooks: copy aliases the original\ngot:  %p\nwant: NOT %p",
						got[tc.reassignOriginalKey], replacement,
					)
				}
			}
		})
	}
}

func TestWebhook_Copy(t *testing.T) {
	tests := []struct {
		name     string
		wantNil  bool
		wantFail map[string]*bool
	}{
		{
			name:    "nil receiver",
			wantNil: true,
		},
		{
			name:     "copies fields and deep-copies pointers, slices, and fails",
			wantFail: map[string]*bool{"notify": new(false)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			origStatus := testCopyStatus(t)
			copyStatus := testCopyStatus(t)

			copyNotifiers := Notifiers{Shoutrrr: &shoutrrr.Shoutrrrs{}}

			var orig *Webhook
			if !tc.wantNil {
				orig = testWebhookForCopy(origStatus, "notify")
			}

			wantStr := decode.ToYAMLString(orig, "")

			got := orig.Copy(copyStatus, copyNotifiers)

			if tc.wantNil {
				if got != nil {
					t.Fatal("after Copy: copy mismatch\ngot:  non-nil\nwant: nil")
				}
				return
			}

			if got == nil {
				t.Fatal("after Copy: copy mismatch\ngot:  nil\nwant: non-nil")
			}

			if got == orig {
				t.Fatalf("after Copy: copy aliases the original\ngot:  %p\nwant: NOT %p", got, orig)
			}

			if gotStr := decode.ToYAMLString(got, ""); gotStr != wantStr {
				t.Fatalf("after Copy: Webhook mismatch\ngot:  %q\nwant: %q", gotStr, wantStr)
			}

			fieldTests := []test.FieldAssertion{
				{Name: "Type", Got: got.Type, Want: orig.Type, Mode: test.CompareEqual},
				{Name: "URL", Got: got.URL, Want: orig.URL, Mode: test.CompareEqual},
				{Name: "Secret", Got: got.Secret, Want: orig.Secret, Mode: test.CompareEqual},
				{Name: "Delay", Got: got.Delay, Want: orig.Delay, Mode: test.CompareEqual},
			}
			if testErr := test.AssertFields(t, fieldTests, "Webhook"); testErr != nil {
				t.Fatalf("after Copy: %v", testErr)
			}

			fieldTests = []test.FieldAssertion{
				{Name: "AllowInvalidCerts", Got: got.AllowInvalidCerts, Want: orig.AllowInvalidCerts, Mode: test.CopiedPointerMode(orig.AllowInvalidCerts)},
				{Name: "DesiredStatusCode", Got: got.DesiredStatusCode, Want: orig.DesiredStatusCode, Mode: test.CopiedPointerMode(orig.DesiredStatusCode)},
				{Name: "MaxTries", Got: got.MaxTries, Want: orig.MaxTries, Mode: test.CopiedPointerMode(orig.MaxTries)},
				{Name: "SilentFails", Got: got.SilentFails, Want: orig.SilentFails, Mode: test.CopiedPointerMode(orig.SilentFails)},
				{Name: "ParentInterval", Got: got.ParentInterval, Want: orig.ParentInterval, Mode: test.CopiedPointerMode(orig.ParentInterval)},
				{Name: "Failed", Got: got.Failed, Want: orig.Failed, Mode: test.CopiedPointerMode(orig.Failed)},
			}
			if testErr := test.AssertFields(t, fieldTests, "Webhook"); testErr != nil {
				t.Fatalf("after Copy: %v", testErr)
			}

			fieldTests = []test.FieldAssertion{
				{Name: "ID", Got: got.ID, Want: orig.ID, Mode: test.CompareEqual},
				{Name: "ServiceStatus", Got: got.ServiceStatus, Want: copyStatus, Mode: test.CompareSamePointer},
				{Name: "Notifiers.Shoutrrr", Got: got.Notifiers.Shoutrrr, Want: copyNotifiers.Shoutrrr, Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, fieldTests, "Webhook"); testErr != nil {
				t.Fatalf("after Copy: %v", testErr)
			}

			fieldTests = []test.FieldAssertion{
				{Name: "Main", Got: got.Main, Want: orig.Main, Mode: test.CompareSamePointer},
				{Name: "Defaults", Got: got.Defaults, Want: orig.Defaults, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: got.HardDefaults, Want: orig.HardDefaults, Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, fieldTests, "Webhook"); testErr != nil {
				t.Fatalf("after Copy: %v", testErr)
			}

			if got.Failed == nil {
				t.Fatal("after Copy: Webhook.Failed mismatch\ngot:  nil\nwant: non-nil")
			}
			for key, wantFail := range tc.wantFail {
				gotFailStr := test.StringifyPtr(got.Failed.Get(key))
				wantFailStr := test.StringifyPtr(wantFail)
				if gotFailStr != wantFailStr {
					t.Errorf(
						"after Copy: Webhook.Failed.Get(%q) mismatch\ngot:  %s\nwant: %s",
						key, gotFailStr, wantFailStr,
					)
				}
				gotNR, wantNR := got.Failed.NextRunnable(key), orig.Failed.NextRunnable(key)
				if !gotNR.Equal(wantNR) {
					t.Errorf(
						"after Copy: Webhook.Failed.NextRunnable(%q) mismatch\ngot:  %s\nwant: %s",
						key, gotNR, wantNR,
					)
				}

				got.Failed.Set(key, new(!*wantFail))
				origFailStr := test.StringifyPtr(orig.Failed.Get(key))
				if origFailStr != wantFailStr {
					t.Errorf(
						"after Set: original Webhook.Failed.Get(%q) mismatch\ngot:  %s\nwant: %s",
						key, origFailStr, wantFailStr,
					)
				}
			}

			if len(got.Headers) == 0 || len(orig.Headers) == 0 {
				t.Fatalf(
					"after Copy: len(Webhook.Headers) mismatch\n"+
						"got:  copy=%d, original=%d\nwant: non-zero",
					len(got.Headers), len(orig.Headers),
				)
			}
			if &got.Headers[0] == &orig.Headers[0] {
				t.Errorf(
					"after Copy: Webhook.Headers: copy aliases the original\n"+
						"got:  %p\nwant: NOT %p",
					&got.Headers[0], &orig.Headers[0],
				)
			}
		})
	}
}

// #########
// # STATE #
// #########

func TestWebhooksDefaults_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		defaults *WebhooksDefaults
		want     bool
	}{
		{
			name:     "empty/0 items",
			defaults: &WebhooksDefaults{},
			want:     true,
		},
		{
			name: "empty/1 item",
			defaults: &WebhooksDefaults{
				"a": &Defaults{},
			},
			want: true,
		},
		{
			name: "empty/2 items",
			defaults: &WebhooksDefaults{
				"a": &Defaults{},
				"b": &Defaults{},
			},
			want: true,
		},
		{
			name: "non-empty/1 item",
			defaults: &WebhooksDefaults{
				"a": &Defaults{
					Type: "github",
				},
			},
			want: false,
		},
		{
			name: "non-empty/2 items",
			defaults: &WebhooksDefaults{
				"a": &Defaults{
					Type: "github",
				},
				"b": &Defaults{
					Type: "gitkav",
				},
			},
			want: false,
		},
		{
			name: "mixed",
			defaults: &WebhooksDefaults{
				"a": &Defaults{},
				"b": &Defaults{
					Type: "github",
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
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestDefaults_IsZero(t *testing.T) {
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
				Type: "github",
			},
			want: false,
		},
		{
			name: "non-empty/URL",
			defaults: &Defaults{
				URL: "https://example.com",
			},
			want: false,
		},
		{
			name: "non-empty/AllowInvalidCerts",
			defaults: &Defaults{
				AllowInvalidCerts: new(false),
			},
			want: false,
		},
		{
			name: "non-empty/Headers",
			defaults: &Defaults{
				Headers: Headers{
					{Key: "X-Header", Value: "val"},
					{Key: "X-Another", Value: "val2"},
				},
			},
			want: false,
		},
		{
			name: "non-empty/Secret",
			defaults: &Defaults{
				Secret: "foobar",
			},
			want: false,
		},
		{
			name: "non-empty/DesiredStatusCode",
			defaults: &Defaults{
				DesiredStatusCode: new(uint16(200)),
			},
			want: false,
		},
		{
			name: "non-empty/Delay",
			defaults: &Defaults{
				Delay: "1h2m3s",
			},
			want: false,
		},
		{
			name: "non-empty/MaxTries",
			defaults: &Defaults{
				MaxTries: new(uint8(4)),
			},
			want: false,
		},
		{
			name: "non-empty/SilentFails",
			defaults: &Defaults{
				SilentFails: new(true),
			},
			want: false,
		},
		{
			name: "non-empty/all",
			defaults: &Defaults{
				Type:              "github",
				URL:               "https://example.com",
				AllowInvalidCerts: new(false),
				Headers: Headers{
					{Key: "X-Header", Value: "val"},
					{Key: "X-Another", Value: "val2"},
				},
				Secret:            "foobar",
				DesiredStatusCode: new(uint16(200)),
				Delay:             "1h2m3s",
				MaxTries:          new(uint8(4)),
				SilentFails:       new(true),
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.defaults.IsZero()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestWebhooks_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		webhooks *Webhooks
		want     bool
	}{
		{
			name:     "nil",
			webhooks: nil,
			want:     true,
		},
		{
			name:     "empty",
			webhooks: &Webhooks{},
			want:     true,
		},
		{
			name: "non-empty/1 item",
			webhooks: &Webhooks{
				"a": &Webhook{},
			},
			want: false, // An empty entry still counts.
		},
		{
			name: "non-empty/2 items",
			webhooks: &Webhooks{
				"a": &Webhook{},
				"b": &Webhook{},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.webhooks.IsZero()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestWebhook_IsDefault(t *testing.T) {
	tests := []struct {
		name    string
		webhook *Webhook
		want    bool
	}{
		{
			name:    "empty",
			webhook: &Webhook{},
			want:    true,
		},
		{
			name: "non-empty/Type",
			webhook: &Webhook{
				Type: "github",
			},
			want: false,
		},
		{
			name: "non-empty/URL",
			webhook: &Webhook{
				URL: "https://example.com",
			},
			want: false,
		},
		{
			name: "non-empty/AllowInvalidCerts",
			webhook: &Webhook{
				AllowInvalidCerts: new(false),
			},
			want: false,
		},
		{
			name: "non-empty/Headers",
			webhook: &Webhook{
				Headers: Headers{
					{Key: "X-Header", Value: "val"},
					{Key: "X-Another", Value: "val2"},
				},
			},
			want: false,
		},
		{
			name: "non-empty/Secret",
			webhook: &Webhook{
				Secret: "foobar",
			},
			want: false,
		},
		{
			name: "non-empty/DesiredStatusCode",
			webhook: &Webhook{
				DesiredStatusCode: new(uint16(200)),
			},
			want: false,
		},
		{
			name: "non-empty/Delay",
			webhook: &Webhook{
				Delay: "1h2m3s",
			},
			want: false,
		},
		{
			name: "non-empty/MaxTries",
			webhook: &Webhook{
				MaxTries: new(uint8(4)),
			},
			want: false,
		},
		{
			name: "non-empty/SilentFails",
			webhook: &Webhook{
				SilentFails: new(true),
			},
			want: false,
		},
		{
			name: "non-empty/all",
			webhook: &Webhook{
				Type:              "github",
				URL:               "https://example.com",
				AllowInvalidCerts: new(false),
				Headers: Headers{
					{Key: "X-Header", Value: "val"},
					{Key: "X-Another", Value: "val2"},
				},
				Secret:            "foobar",
				DesiredStatusCode: new(uint16(200)),
				Delay:             "1h2m3s",
				MaxTries:          new(uint8(4)),
				SilentFails:       new(true),
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.webhook.IsDefault()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

// #############
// # STRINGIFY #
// #############

func TestDefaults_String(t *testing.T) {
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
				"Defaults.String",
				tc.defaults.String,
				tc.want,
			)
		})
	}
}

func TestWebhooksDefaults_String(t *testing.T) {
	tests := []struct {
		name             string
		webhooksDefaults *WebhooksDefaults
		want             string
	}{
		{
			name:             "nil",
			webhooksDefaults: nil,
			want:             "",
		},
		{
			name:             "empty",
			webhooksDefaults: &WebhooksDefaults{},
			want:             "{}\n",
		},
		{
			name: "two empty",
			webhooksDefaults: &WebhooksDefaults{
				"one": &Defaults{},
				"two": &Defaults{},
			},
			want: test.TrimYAML(`
				one: {}
				two: {}
			`),
		},
		{
			name: "one with data",
			webhooksDefaults: &WebhooksDefaults{
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
			webhooksDefaults: &WebhooksDefaults{
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
			webhooksDefaults: &WebhooksDefaults{
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
				"WebhooksDefaults.String",
				tc.webhooksDefaults.String,
				tc.want,
			)
		})
	}
}

func TestWebhooks_String(t *testing.T) {
	tests := []struct {
		name     string
		webhooks *Webhooks
		want     string
	}{
		{
			name:     "nil",
			webhooks: nil,
			want:     "",
		},
		{
			name:     "empty",
			webhooks: &Webhooks{},
			want:     "{}\n",
		},
		{
			name: "one",
			webhooks: &Webhooks{
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
			webhooks: &Webhooks{
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
			webhooks: &Webhooks{
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

			got := tc.webhooks.String()

			if got != tc.want {
				t.Errorf("Webhooks.String() mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestWebhook_String(t *testing.T) {
	tests := []struct {
		name    string
		webhook *Webhook
		want    string
	}{
		{
			name:    "nil",
			webhook: nil,
			want:    "",
		},
		{
			name:    "empty",
			webhook: &Webhook{},
			want:    "{}\n",
		},
		{
			name: "filled",
			webhook: New(
				new(false),
				Headers{
					{Key: "X-Header", Value: "val"},
					{Key: "X-Another", Value: "val2"},
				},
				"1h1m1s",
				new(uint16(200)),
				nil,
				"filled",
				new(uint8(4)),
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
				new("3h2m1s"),
				"foobar",
				new(true),
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
				"Webhook.String",
				tc.webhook.String,
				tc.want,
			)
		})
	}
}
