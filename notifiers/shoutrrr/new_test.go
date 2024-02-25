// Copyright [2023] [Argus]
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
	"bytes"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/release-argus/Argus/util"
)

func TestFromPayload_ReadFromFail(t *testing.T) {
	// GIVEN an invalid payload
	payloadStr := "this is a long payload"
	payload := io.NopCloser(bytes.NewReader([]byte(payloadStr)))
	payload = http.MaxBytesReader(nil, payload, 5)

	// WHEN we call New
	_, err := FromPayload(
		&payload,
		&Shoutrrr{},
		&util.LogFrom{Primary: "TestFromPayload_ReadFromFail"},
	)

	// THEN we should get an error
	if err == nil {
		t.Errorf("Want error, got nil")
	}
}

func TestShoutrrr_FromPayload(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		payload  string
		original *Shoutrrr
		want     *Shoutrrr
		err      string
	}{
		"empty": {
			payload:  "",
			original: &Shoutrrr{},
			want:     &Shoutrrr{},
			err:      "EOF",
		},
		"empty JSON": {
			payload:  "{}",
			original: &Shoutrrr{},
			want:     &Shoutrrr{},
			err:      "type is required",
		},
		"invalid type": {
			payload:  `{"type": "foo"}`,
			original: &Shoutrrr{},
			want: &Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					Type: "foo"}},
			err: "type.*foo.*invalid",
		},
		"valid": {
			payload: `{
"type": "gotify",
"url_fields": {
	"host": "example.com",
	"token": "foo"}
}`,
			original: &Shoutrrr{},
			want: &Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					Type: "gotify",
					URLFields: map[string]string{
						"host":  "example.com",
						"token": "foo"},
				}},
		},
		"inherit secrets": {
			payload: `{
"type": "gotify",
"url_fields": {
	"host": "example.com",
	"token": "<secret>"}
}`,
			original: &Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					Type: "gotify",
					URLFields: map[string]string{
						"host":  "release-argus.com",
						"token": "bar",
					}}},
			want: &Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					Type: "gotify",
					URLFields: map[string]string{
						"host":  "example.com",
						"token": "bar",
					}}},
		},
		"invalid CheckValues": {
			payload: `{
"type": "gotify",
"url_fields": {
	"host": "release-argus.com"}
}`,
			original: &Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					Type: "gotify",
					URLFields: map[string]string{
						"host":  "example.com",
						"token": "bar",
					}}},
			want: &Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					Type: "gotify",
					URLFields: map[string]string{
						"host": "release-argus.com",
					}}},
			err: "token: <required",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.original.Main = &ShoutrrrDefaults{}
			tc.original.Defaults = &ShoutrrrDefaults{}
			tc.original.HardDefaults = &ShoutrrrDefaults{}
			tc.payload = strings.TrimSpace(tc.payload)
			tc.payload = strings.ReplaceAll(tc.payload, "\n", "")
			tc.payload = strings.ReplaceAll(tc.payload, "\t", "")
			tc.payload = strings.ReplaceAll(tc.payload, ": ", ":")
			reader := strings.NewReader(tc.payload)
			payload := io.NopCloser(reader)

			// WHEN creating a Shoutrrr from a payload
			got, err := FromPayload(&payload, tc.original, &util.LogFrom{
				Primary: "TestShoutrrr_FromPayload", Secondary: name,
			})

			// THEN the Shoutrrr is created as expected
			if tc.err == "" {
				tc.err = "^$"
			}
			gotErr := util.ErrorToString(err)
			if !regexp.MustCompile(tc.err).MatchString(gotErr) {
				t.Errorf("Expecting error to match %q, got %q",
					tc.err, gotErr)
			}
			if got.String("") != tc.want.String("") {
				t.Errorf("FromPayload() mismatch:\ngot:  %q\nwant: %q",
					got.String(""), tc.want.String(""))
			}
		})
	}

}

func TestShoutrrr_InheritSecrets(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		from *Shoutrrr
		to   *Shoutrrr
		want *Shoutrrr
	}{
		"empty": {
			from: &Shoutrrr{},
			to:   &Shoutrrr{},
			want: &Shoutrrr{},
		},
		"params": {
			from: &Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					Params: map[string]string{
						"devices":      "foo",
						"not_a_secret": "bar",
					}}},
			to: &Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					Params: map[string]string{
						"devices":      "<secret>",
						"not_a_secret": "<secret>",
					}}},
			want: &Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					Params: map[string]string{
						"devices":      "foo",
						"not_a_secret": "<secret>",
					}}},
		},
		"url_fields": {
			from: &Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					URLFields: map[string]string{
						"altid":        "bish",
						"not_a_secret": "what",
					}}},
			to: &Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					URLFields: map[string]string{
						"altid":        "<secret>",
						"apikey":       "<secret>",
						"botKey":       "bosh",
						"not_a_secret": "<secret>",
					}}},
			want: &Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					URLFields: map[string]string{
						"altid":        "bish",
						"apikey":       "<secret>",
						"botKey":       "bosh",
						"not_a_secret": "<secret>",
					}}},
		},
		"both": {
			from: &Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					Params: map[string]string{
						"devices":        "foo",
						"something_else": "bar",
					},
					URLFields: map[string]string{
						"altid":        "bish",
						"not_a_secret": "what",
					}}},
			to: &Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					Params: map[string]string{
						"devices":        "foo",
						"something_else": "<secret>",
					},
					URLFields: map[string]string{
						"altid":        "<secret>",
						"apikey":       "<secret>",
						"botKey":       "bosh",
						"not_a_secret": "<secret>",
					}}},
			want: &Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					Params: map[string]string{
						"devices":        "foo",
						"something_else": "<secret>",
					},
					URLFields: map[string]string{
						"altid":        "bish",
						"apikey":       "<secret>",
						"botKey":       "bosh",
						"not_a_secret": "<secret>",
					}}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN inheriting secrets
			tc.to.inheritSecrets(tc.from)

			// THEN the secrets are inherited
			got := tc.to.String("")
			want := tc.want.String("")
			if got != want {
				t.Errorf("InheritSecrets() mismatch: got %q, want %q",
					got, want)
			}
		})
	}
}
