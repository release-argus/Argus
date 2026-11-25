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

package shoutrrr

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
	"github.com/release-argus/Argus/util/polymorphic"
)

func TestShoutrrr_TestSend(t *testing.T) {
	tests := []struct {
		name        string
		sType       *string
		nilShoutrrr bool
		errRegex    string
		wantSent    bool
	}{
		{
			name:        "nil shoutrrr",
			nilShoutrrr: true,
			errRegex:    `^shoutrrr is nil$`,
		},
		{
			name:  "invalid type",
			sType: new("somethingUnknown"),
			errRegex: test.TrimYAML(`
				^failed to create Shoutrrr sender:
					creating sender:
						error initializing router services: URL 0:
							unknown service: ""$`,
			),
		},
		{
			name:     "valid",
			errRegex: `^$`,
			wantSent: true,
		},
	}

	type gotifyRequest struct {
		path    string
		body    []byte
		readErr error
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			requests := make(chan gotifyRequest, 4)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				requests <- gotifyRequest{path: r.URL.Path, body: body, readErr: err}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":1}`))
			}))
			t.Cleanup(server.Close)
			serverURL, err := url.Parse(server.URL)
			if err != nil {
				t.Fatalf("url.Parse() failed: %v", err)
			}
			shoutrrr := testShoutrrr(false, false)
			shoutrrr.URLFields = map[string]string{
				"host":  serverURL.Hostname(),
				"port":  serverURL.Port(),
				"token": "AAAAAAAAAAAAAAA",
			}
			shoutrrr.Params["disabletls"] = "yes"
			shoutrrr.Params["title"] = "argus"
			shoutrrr.Options["message"] = "new release"
			if tc.sType != nil {
				shoutrrr.Type = *tc.sType
			}
			if tc.nilShoutrrr {
				shoutrrr = nil
			}

			err = shoutrrr.TestSend("https://example.com")
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			wantRequests := 0
			if tc.wantSent {
				wantRequests = 1
			}
			if got := len(requests); got != wantRequests {
				t.Fatalf("request count mismatch\ngot:  %d\nwant: %d", got, wantRequests)
			}
			if !tc.wantSent {
				return
			}
			req := <-requests
			if req.readErr != nil {
				t.Fatalf("io.ReadAll() failed: %v", req.readErr)
			}
			var body struct {
				Title   string `json:"title"`
				Message string `json:"message"`
			}
			if err = decode.Unmarshal("json", req.body, &body); err != nil {
				t.Fatalf("decode.Unmarshal() failed: %v", err)
			}
			if want := "/message"; req.path != want {
				t.Errorf("http.Request.URL.Path mismatch\ngot:  %q\nwant: %q", req.path, want)
			}
			if want := "TEST - argus"; body.Title != want {
				t.Errorf("request title mismatch\ngot:  %q\nwant: %q", body.Title, want)
			}
			if want := "TEST - new release"; body.Message != want {
				t.Errorf("request message mismatch\ngot:  %q\nwant: %q", body.Message, want)
			}
		})
	}
}

func TestShoutrrrs_CheckValues(t *testing.T) {
	tests := []struct {
		name     string
		input    *Shoutrrrs
		errRegex string
		changed  bool
	}{
		{
			name:     "nil map",
			input:    (*Shoutrrrs)(nil),
			errRegex: `^$`,
			changed:  false,
		},
		{
			name:     "valid map",
			errRegex: `^$`,
			input: &Shoutrrrs{
				"valid": testShoutrrr(false, false),
				"other": testShoutrrr(false, false),
			},
			changed: false,
		},
		{
			name:     "valid map, with changed",
			errRegex: `^$`,
			input: &Shoutrrrs{
				"valid": testShoutrrr(false, false),
				"other": testShoutrrr(false, false),
				"generic": &Shoutrrr{
					URLFields: map[string]string{
						"host":           "example.com",
						"custom_headers": `{"foo":"bar"}`},
				},
			},
			changed: true, // The "generic" ID types it, so custom_headers is renamed.
		},
		{
			name: "invalid map",
			errRegex: test.TrimYAML(`
				other:
					type: <required>`,
			),
			input: &Shoutrrrs{
				"valid": testShoutrrr(false, false),
				"other": New(
					nil, "", "",
					make(map[string]string),
					make(map[string]string),
					make(map[string]string),
					nil, nil, nil,
				),
			},
			changed: false,
		},
		{
			name: "ordered errors",
			errRegex: test.TrimYAML(`
				aNotify:
					type: <required>.*
				bNotify:
					type: <required>.*`,
			),
			input: &Shoutrrrs{
				"aNotify": New(
					nil, "", "",
					make(map[string]string),
					make(map[string]string),
					make(map[string]string),
					nil, nil, nil,
				),
				"bNotify": New(
					nil, "", "",
					make(map[string]string),
					make(map[string]string),
					make(map[string]string),
					nil, nil, nil,
				),
			},
			changed: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.input != nil {
				tc.input.Init(newTestStatus(tc.name, len(*tc.input)), plainConfig())
			}

			_, _ = test.AssertCheckValuesWithErrorAndChanged(
				t,
				tc.errRegex,
				tc.changed,
				tc.input.CheckValues,
			)
		})
	}
}

// testTwilioSID is split so secret scanners do not take it for a real account SID.
const testTwilioSID = "AC" + "0123456789abcdef" + "0123456789abcdef"

func TestShoutrrr_CheckValues(t *testing.T) {
	testS := testShoutrrr(false, false)
	tests := []struct {
		name                       string
		nilShoutrrr                bool
		sType                      string
		options, urlFields, params map[string]string
		wantURLFields, wantParams  map[string]string
		wantDelay                  string
		main                       *Defaults
		errRegex                   string
		changed                    bool
	}{
		{
			name:        "nil shoutrrr",
			nilShoutrrr: true,
			errRegex:    `^$`,
			changed:     false,
		},
		{
			name:      "empty",
			errRegex:  `^type: <required>[^:]+://[^:]+$`,
			urlFields: map[string]string{},
			params:    map[string]string{},
			changed:   false,
		},
		{
			name: "invalid delay",
			errRegex: test.TrimYAML(`
				^options:
					delay: .* <invalid>`,
			),
			sType:     testS.Type,
			urlFields: testS.URLFields,
			options: map[string]string{
				"delay": "5x",
			},
			changed: false,
		},
		{
			name:      "fixes delay",
			errRegex:  `^$`,
			sType:     testS.Type,
			urlFields: testS.URLFields,
			wantDelay: "5s",
			options: map[string]string{
				"delay": "5",
			},
			changed: false, // checkValuesOptions fixes the delay without reporting it.
		},
		{
			name: "invalid message template",
			errRegex: test.TrimYAML(`
				^options:
					message: ".+" <invalid>.*$`,
			),
			sType:     testS.Type,
			urlFields: testS.URLFields,
			options: map[string]string{
				"message": "{{ version }",
			},
			changed: false,
		},
		{
			name:      "valid message template",
			errRegex:  `^$`,
			sType:     testS.Type,
			urlFields: testS.URLFields,
			options: map[string]string{
				"message": "{{ version }}",
			},
			changed: false,
		},
		{
			name: "invalid title template",
			errRegex: test.TrimYAML(`
				^params:
					title: .* <invalid>.*$`,
			),
			sType:     testS.Type,
			urlFields: testS.URLFields,
			params: map[string]string{
				"title": "{{ version }",
			},
			changed: false,
		},
		{
			name:      "valid title template",
			errRegex:  `^$`,
			sType:     testS.Type,
			urlFields: testS.URLFields,
			params: map[string]string{
				"title": "{{ version }}",
			},
			changed: false,
		},
		{
			name:      "valid param template",
			errRegex:  `^$`,
			sType:     testS.Type,
			urlFields: testS.URLFields,
			params: map[string]string{
				"foo": "{{ version }}",
			},
			changed: false,
		},
		{
			name: "invalid param template",
			errRegex: test.TrimYAML(`
				^params:
					foo: .* <invalid>.*$`,
			),
			sType:     testS.Type,
			urlFields: testS.URLFields,
			params: map[string]string{
				"foo": "{{ version }",
			},
			changed: false,
		},
		{
			name: "invalid param and option",
			errRegex: test.TrimYAML(`
				^options:
					delay: [^<]+<invalid>.*
				params:
					title: [^<]+<invalid>.*$`,
			),
			sType:     testS.Type,
			urlFields: testS.URLFields,
			params: map[string]string{
				"title": "{{ version }",
			},
			options: map[string]string{
				"delay": "2x",
			},
			changed: false,
		},
		{
			name:     "does correctSelf",
			errRegex: `^$`,
			sType:    testS.Type,
			urlFields: map[string]string{
				"host":  "foo",
				"token": "AAAAAAAAAAAAAAA",
				"port":  ":8080",
				"path":  "/test",
			},
			wantURLFields: map[string]string{
				"host":  "foo",
				"token": "AAAAAAAAAAAAAAA",
				"port":  "8080",
				"path":  "test",
			},
			changed: true,
		},
		{
			name:     "generic.url_fields.custom_headers -> headers",
			errRegex: `^$`,
			sType:    "generic",
			urlFields: map[string]string{
				"host":           "example.com",
				"custom_headers": `{"foo":"bar"}`,
			},
			wantURLFields: map[string]string{
				"host":    "example.com",
				"headers": `{"foo":"bar"}`,
			},
			changed: true,
		},
		{
			name:     "generic.url_fields.custom_headers not pulled to headers if headers already defined",
			errRegex: `^$`,
			sType:    "generic",
			urlFields: map[string]string{
				"host":           "example.com",
				"custom_headers": `{"foo":"bar"}`,
				"headers":        `{"foo":"baz"}`,
			},
			wantURLFields: map[string]string{
				"host":    "example.com",
				"headers": `{"foo":"baz"}`,
			},
			changed: true,
		},
		{
			name:     "ntfy.params.disabletlsverification -> skiptlsverify",
			errRegex: `^$`,
			sType:    "ntfy",
			urlFields: map[string]string{
				"topic": "123",
			},
			wantURLFields: map[string]string{
				"topic": "123",
			},
			params: map[string]string{
				"disabletlsverification": "true",
			},
			wantParams: map[string]string{
				"skiptlsverify": "true",
			},
			changed: true,
		},
		{
			name:     "ntfy.params.disabletlsverification not pulled to skiptlsverify if skiptlsverify already defined",
			errRegex: `^$`,
			sType:    "ntfy",
			urlFields: map[string]string{
				"topic": "123",
			},
			wantURLFields: map[string]string{
				"topic": "123",
			},
			params: map[string]string{
				"disabletlsverification": "true",
				"skiptlsverify":          "false",
			},
			wantParams: map[string]string{
				"skiptlsverify": "false",
			},
			changed: true,
		},
		{
			name:      "valid",
			errRegex:  `^$`,
			urlFields: map[string]string{},
			main:      testDefaults(false, false),
			changed:   false,
		},
		{
			name:     "valid with self and main",
			errRegex: `^$`,
			urlFields: map[string]string{
				"host": "foo",
			},
			main: NewDefaults(
				testS.Type,
				nil,
				map[string]string{
					"token": "AAAAAAAAAAAAAAA",
				},
				nil,
			),
			changed: false,
		},
		{
			name: "invalid url_fields",
			errRegex: test.TrimYAML(`
				^url_fields:
					host: <required>.*
					token: <required>.*$`,
			),
			sType:   testS.Type,
			changed: false,
		},
		{
			name: "invalid params/locate skipped",
			errRegex: test.TrimYAML(`
				^params:
					fromaddress: <required>.*
					toaddresses: <required>.*$`,
			),
			urlFields: map[string]string{
				"host": "https://release-argus.io",
			},
			sType:   "smtp",
			changed: false,
		},
		{
			name:  "homeassistant/valid",
			sType: "homeassistant",
			urlFields: map[string]string{
				"host":  "homeassistant.example.com",
				"port":  "8123",
				"token": "bish",
			},
			params: map[string]string{
				"disabletls": "yes",
				"service":    "notify.mobile_app_phone",
			},
			errRegex: `^$`,
			changed:  false,
		},
		{
			name:  "homeassistant/host with scheme and port split out",
			sType: "homeassistant",
			urlFields: map[string]string{
				"host":  "http://homeassistant.example.com:8123",
				"token": "bish",
			},
			wantURLFields: map[string]string{
				"host":  "homeassistant.example.com",
				"port":  "8123",
				"token": "bish",
			},
			errRegex: `^$`,
			changed:  true,
		},
		{
			name:      "teams/new-style (Power Automate URL)",
			sType:     "teams",
			urlFields: map[string]string{},
			params: map[string]string{
				"host": "https://prod-00.westus.logic.azure.com:443/workflows/abc?sig=SIG",
			},
			errRegex: `^$`,
			changed:  false,
		},
		{
			name:  "teams/old-style (Office 365 Connectors) - validates without error",
			sType: "teams",
			urlFields: map[string]string{
				"group":      "GROUP",
				"tenant":     "TENANT",
				"altid":      "ALTID",
				"groupowner": "GROUPOWNER",
				"extraid":    "EXTRAID",
			},
			params: map[string]string{
				"host": "organization.webhook.office.com",
			},
			errRegex: `^$`,
			changed:  false,
		},
		{
			name:  "lark/valid",
			sType: "lark",
			urlFields: map[string]string{
				"host":  "open.larksuite.com",
				"token": "TOKEN",
			},
			errRegex: `^$`,
		},
		{
			name:  "mqtt/valid",
			sType: "mqtt",
			urlFields: map[string]string{
				"host":  "mqtt.example.com",
				"topic": "argus/releases",
			},
			params: map[string]string{
				"qos": "atleastonce",
			},
			wantParams: map[string]string{
				"qos": "AtLeastOnce",
			},
			errRegex: `^$`,
		},
		{
			name:  "pagerduty/valid",
			sType: "pagerduty",
			urlFields: map[string]string{
				"integrationkey": "0123456789abcdef0123456789abcdef",
			},
			errRegex: `^$`,
		},
		{
			name:  "rocketchat/valid without channel",
			sType: "rocketchat",
			urlFields: map[string]string{
				"host":   "rocketchat.example.com",
				"tokena": "TOKENA",
				"tokenb": "TOKENB",
			},
			errRegex: `^$`,
		},
		{
			name:  "signal/valid",
			sType: "signal",
			urlFields: map[string]string{
				"host":       "signal.example.com",
				"source":     "+15551234567",
				"recipients": "+15557654321,u:argus.01",
			},
			errRegex: `^$`,
		},
		{
			name:  "signalgrid/valid",
			sType: "signalgrid",
			urlFields: map[string]string{
				"clientkey": "CLIENTKEY",
				"channel":   "CHANNEL",
			},
			errRegex: `^$`,
		},
		{
			name:  "twilio/valid",
			sType: "twilio",
			urlFields: map[string]string{
				"accountsid": testTwilioSID,
				"authtoken":  "TOKEN",
				"fromnumber": "+15551234567",
				"tonumbers":  "+15557654321,+15550000000",
			},
			errRegex: `^$`,
		},
		{
			name:  "wecom/valid",
			sType: "wecom",
			urlFields: map[string]string{
				"key": "693a0006-7a0c-4bc4-97a0-0ec2a1fa5aaa",
			},
			errRegex: `^$`,
		},
		{
			name:  "xmpp/valid",
			sType: "xmpp",
			urlFields: map[string]string{
				"host":     "xmpp.example.com",
				"user":     "argus",
				"password": "pass",
			},
			params: map[string]string{
				"to": "user@example.com",
			},
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var input *Shoutrrr
			if !tc.nilShoutrrr {
				input = testShoutrrr(false, false)
				input.Type = tc.sType
				if tc.main == nil {
					tc.main = &Defaults{}
				}
				input.Main = tc.main
				input.Main.InitMaps()
				input.Options = util.CopyMap(tc.options)
				input.URLFields = util.CopyMap(tc.urlFields)
				input.Params = util.CopyMap(tc.params)
			}

			_, _ = test.AssertCheckValuesWithErrorAndChanged(
				t,
				tc.errRegex,
				tc.changed,
				input.CheckValues,
			)

			if tc.wantURLFields != nil {
				if testErr := test.AssertMapEqual(t, input.URLFields, tc.wantURLFields, "Shoutrrr.URLFields"); testErr != nil {
					t.Error(testErr)
				}
			}
			if tc.wantParams != nil {
				if testErr := test.AssertMapEqual(t, input.Params, tc.wantParams, "Shoutrrr.Params"); testErr != nil {
					t.Error(testErr)
				}
			}
			if tc.wantDelay != "" {
				if got := input.Options["delay"]; got != tc.wantDelay {
					t.Errorf(
						"Shoutrrr.Options[\"delay\"] mismatch\ngot:  %q\nwant: %q",
						got, tc.wantDelay,
					)
				}
			}
		})
	}
}

func TestBase_CheckValues(t *testing.T) {
	tests := []struct {
		name       string
		input      *Base
		wantParams map[string]string
		errRegex   string
		changed    bool
	}{
		{
			name:     "nil Base",
			input:    (*Base)(nil),
			errRegex: `^$`,
			changed:  false,
		},
		{
			name: "valid Base",
			input: &Base{
				Type: "slack",
				Options: map[string]string{
					"delay": "10s",
				},
				Params: map[string]string{
					"color": "orange",
				},
			},
			errRegex: `^$`,
			changed:  false,
		},
		{
			name: "invalid delay option",
			input: &Base{
				Type: "slack",
				Options: map[string]string{
					"delay": "10x",
				},
			},
			errRegex: test.TrimYAML(`
				^options:
				  delay: "10x" <invalid>.*$`,
			),
			changed: false,
		},
		{
			name: "invalid param template",
			input: &Base{
				Type: "slack",
				Params: map[string]string{
					"color": "{{ invalid template }}",
				},
			},
			errRegex: test.TrimYAML(`
				^params:
					color: "{{ invalid template }}" <invalid>.*$`,
			),
			changed: false,
		},
		{
			name: "multiple errors",
			input: &Base{
				Type: "slack",
				Options: map[string]string{
					"delay": "10x",
				},
				Params: map[string]string{
					"color": "{{ invalid template }}",
				},
			},
			errRegex: test.TrimYAML(`
				^options:
					delay: "10x" <invalid>.*
				params:
					color: "{{ invalid template }}" <invalid>.*$`,
			),
			changed: false,
		},
		{
			name: "matrix/rooms/leading #",
			input: &Base{
				Type: "matrix",
				Params: map[string]string{
					"rooms": "#alias:server",
				},
			},
			errRegex: `^$`,
			wantParams: map[string]string{
				"rooms": "alias:server",
			},
			changed: true,
		},
		{
			name: "matrix/rooms/leading # already urlEncoded",
			input: &Base{
				Type: "matrix",
				Params: map[string]string{
					"rooms": "%23alias:server",
				},
			},
			errRegex: `^$`,
			wantParams: map[string]string{
				"rooms": "%23alias:server",
			},
			changed: false,
		},
		{
			name: "matrix/rooms/valid",
			input: &Base{
				Type: "matrix",
				Params: map[string]string{
					"rooms": "alias:server",
				},
			},
			errRegex: `^$`,
			wantParams: map[string]string{
				"rooms": "alias:server",
			},
			changed: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _ = test.AssertCheckValuesWithErrorAndChanged(
				t,
				tc.errRegex,
				tc.changed,
				func() (error, bool) { return tc.input.CheckValues("") },
			)

			if tc.wantParams != nil {
				if testErr := test.AssertMapEqual(t, tc.input.Params, tc.wantParams, "Base.Params"); testErr != nil {
					t.Error(testErr)
				}
			}
		})
	}
}

func TestShoutrrrsDefaults_CheckValues(t *testing.T) {
	tests := []struct {
		name     string
		input    *ShoutrrrsDefaults
		errRegex string
		changed  bool
	}{
		{
			name:     "nil map",
			input:    (*ShoutrrrsDefaults)(nil),
			errRegex: `^$`,
			changed:  false,
		},
		{
			name:     "valid map",
			errRegex: `^$`,
			input: &ShoutrrrsDefaults{
				"valid": testDefaults(false, false),
				"other": testDefaults(false, false),
			},
			changed: false,
		},
		{
			name:     "valid map, with changed",
			errRegex: `^$`,
			input: &ShoutrrrsDefaults{
				"valid": testDefaults(false, false),
				"other": testDefaults(false, false),
				"generic": &Defaults{
					URLFields: map[string]string{
						"host":           "example.com",
						"custom_headers": `{"foo":"bar"}`,
					},
				},
			},
			changed: true, // The "generic" ID types it, so custom_headers is renamed.
		},
		{
			name: "invalid type",
			errRegex: test.TrimYAML(`
				^other:
					type: "somethingUnknown" <invalid> \(supported values = .+\)$`,
			),
			input: &ShoutrrrsDefaults{
				"valid": testDefaults(false, false),
				"other": NewDefaults(
					"somethingUnknown",
					make(map[string]string),
					make(map[string]string),
					make(map[string]string),
				),
			},
			changed: false,
		},
		{
			name:     "delay without unit",
			errRegex: `^$`,
			input: &ShoutrrrsDefaults{
				"foo": NewDefaults(
					"gotify",
					map[string]string{
						"delay": "1",
					},
					nil,
					nil,
				),
			},
			changed: false, // checkValuesOptions fixes the delay without reporting it.
		},
		{
			name: "invalid delay",
			errRegex: test.TrimYAML(`
				^foo:
					options:
						delay: "1x" <invalid>`,
			),
			input: &ShoutrrrsDefaults{
				"foo": NewDefaults(
					"gotify",
					map[string]string{
						"delay": "1x",
					},
					nil,
					nil,
				),
			},
			changed: false,
		},
		{
			name: "invalid message template",
			errRegex: test.TrimYAML(`
				^bar:
					options:
						message: "[^"]+" <invalid>.*$`,
			),
			input: &ShoutrrrsDefaults{
				"bar": NewDefaults(
					"gotify",
					map[string]string{
						"message": "{{ .foo }",
					},
					nil,
					nil,
				),
			},
			changed: false,
		},
		{
			name: "invalid params template",
			errRegex: test.TrimYAML(`
				^bar:
					params:
						title: "[^"]+" <invalid>.*$`,
			),
			input: &ShoutrrrsDefaults{
				"bar": NewDefaults(
					"gotify",
					nil,
					nil,
					map[string]string{
						"title": "{{ .bar }",
					},
				),
			},
			changed: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _ = test.AssertCheckValuesWithErrorAndChanged(
				t,
				tc.errRegex,
				tc.changed,
				tc.input.CheckValues,
			)
		})
	}
}

func TestDefaults_CheckValues(t *testing.T) {
	tests := []struct {
		name     string
		input    *Defaults
		id       string
		errRegex string
		changed  bool
	}{
		{
			name:     "nil defaults/valid id",
			input:    (*Defaults)(nil),
			id:       "slack",
			errRegex: `^$`,
			changed:  false,
		},
		{
			name:     "nil defaults/invalid id",
			input:    (*Defaults)(nil),
			id:       "argus",
			errRegex: `^type: "argus" <invalid>.*gotify.*$`,
			changed:  false,
		},
		{
			name:     "empty Type uses id/valid",
			input:    &Defaults{},
			id:       "gotify",
			errRegex: `^$`,
			changed:  false,
		},
		{
			name:     "empty Type uses id/invalid",
			input:    &Defaults{},
			id:       "unknown",
			errRegex: `^type: "unknown" <invalid>.*$`,
			changed:  false,
		},
		{
			name: "Type set overrides id (both valid)",
			input: &Defaults{
				Type: "gotify",
			},
			id:       "slack",
			errRegex: `^$`,
			changed:  false,
		},
		{
			name: "Base error is propagated",
			input: &Defaults{
				Type: "slack",
				Options: map[string]string{
					"delay": "10x",
				},
			},
			id: "",
			errRegex: test.TrimYAML(`
				^options:
					delay: "10x" <invalid>.*$`,
			),
			changed: false,
		},
		{
			name: "Combines type invalid and base error",
			input: &Defaults{
				Type: "invalid",
				Params: map[string]string{
					"color": "{{ invalid template }}",
				},
			},
			id: "",
			errRegex: test.TrimYAML(`
				^type: "invalid" <invalid>.*
				params:
					color: "{{ invalid template }}" <invalid>.*$`,
			),
			changed: false,
		},
		{
			name: "both valid, no error",
			input: &Defaults{
				Type: "gotify",
				Options: map[string]string{
					"delay": "1s",
				},
				Params: map[string]string{
					"message": "release {{ version }}",
				},
			},
			id:       "",
			errRegex: `^$`,
			changed:  false,
		},
		{
			name:     "empty Type and id empty",
			input:    &Defaults{},
			id:       "",
			errRegex: `^type: <required>.*$`,
			changed:  false,
		},
		{
			name: "port with colon prefix is corrected",
			input: &Defaults{
				Type: "gotify",
				URLFields: map[string]string{
					"port": ":123",
				},
			},
			id:       "",
			errRegex: `^$`,
			changed:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _ = test.AssertCheckValuesWithErrorAndChanged(
				t,
				tc.errRegex,
				tc.changed,
				func() (error, bool) {
					return tc.input.CheckValues(tc.id)
				},
			)
		})
	}
}

func TestShoutrrrsDefaults_Print(t *testing.T) {
	testValid := testDefaults(false, false)
	testInvalid := testDefaults(true, true)
	tests := []struct {
		name              string
		shoutrrrsDefaults *ShoutrrrsDefaults
		want              string
	}{
		{
			name:              "nil",
			shoutrrrsDefaults: nil,
			want:              "",
		},
		{
			name:              "empty",
			shoutrrrsDefaults: &ShoutrrrsDefaults{},
			want:              "",
		},
		{
			name: "single empty element map",
			shoutrrrsDefaults: &ShoutrrrsDefaults{
				"single": {},
			},
			want: test.TrimYAML(`
				notify:
					single: {}`,
			),
		},
		{
			name: "single element map",
			shoutrrrsDefaults: &ShoutrrrsDefaults{
				"single": testValid,
			},
			want: test.TrimYAML(`
				notify:
					single:
						type: gotify
						options:
							max_tries: '` + testValid.getOption("max_tries") + `'
						url_fields:
							host: ` + testValid.getURLField("host") + `
							path: ` + testValid.getURLField("path") + `
							token: ` + testValid.getURLField("token"),
			),
		},
		{
			name: "multiple element map",
			shoutrrrsDefaults: &ShoutrrrsDefaults{
				"first":  testValid,
				"second": testInvalid,
			},
			want: test.TrimYAML(`
				notify:
					first:
						type: gotify
						options:
							max_tries: '` + testValid.getOption("max_tries") + `'
						url_fields:
							host: ` + testValid.getURLField("host") + `
							path: ` + testValid.getURLField("path") + `
							token: ` + testValid.getURLField("token") + `
					second:
						type: gotify
						options:
							max_tries: '` + testInvalid.getOption("max_tries") + `'
						url_fields:
							host: ` + testInvalid.getURLField("host") + `
							path: ` + testInvalid.getURLField("path") + `
							token: ` + testInvalid.getURLField("token"),
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureStdout(t)

			if tc.want != "" {
				tc.want += "\n"
			}

			tc.shoutrrrsDefaults.Print("")

			stdout := releaseStdout()
			if stdout != tc.want {
				t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.want)
			}
		})
	}
}

func TestBase_CorrectSelf(t *testing.T) {
	tests := []struct {
		name          string
		sType         string
		mapTarget     string
		startAs, want map[string]string
		wantAbsent    []string
		wantChanged   bool
	}{
		{
			name:      "port/leading colon",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"port": ":8080",
			},
			want: map[string]string{
				"port": "8080",
			},
			wantChanged: true,
		},
		{
			name:      "port/valid",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"port": "8080",
			},
			want: map[string]string{
				"port": "8080",
			},
		},
		{
			name:      "path/leading slash",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"path": "/argus",
			},
			want: map[string]string{
				"path": "argus",
			},
			wantChanged: true,
		},
		{
			name:      "path/valid",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"path": "argus",
			},
			want: map[string]string{
				"path": "argus",
			},
		},
		{
			name:      "port/from url",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"host": "https://mattermost.example.com:8443",
				"port": "",
			},
			want: map[string]string{
				"host": "mattermost.example.com",
				"port": "8443",
			},
			wantChanged: true,
		},
		{
			name:      "host/valid, no change",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"host": "example.com",
			},
			want: map[string]string{
				"host": "example.com",
			},
		},
		{
			name:      "host/port without scheme",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"host": "example.com:9090",
				"port": "",
			},
			want: map[string]string{
				"host": "example.com",
				"port": "9090",
			},
			wantChanged: true,
		},
		{
			name:      "host/scheme without port",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"host": "https://example.com",
				"port": "",
			},
			want: map[string]string{
				"host": "example.com",
				"port": "",
			},
			wantChanged: true,
		},
		{
			name:      "host/encoded port overrides existing port",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"host": "https://example.com:8443",
				"port": "1234",
			},
			want: map[string]string{
				"host": "example.com",
				"port": "8443",
			},
			wantChanged: true,
		},
		{
			name:      "host/invalid host left unchanged",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"host": "https://\texample.com",
			},
			want: map[string]string{
				"host": "https://\texample.com",
			},
		},
		{
			name:      "bark/scheme http -> disabletls",
			sType:     "bark",
			mapTarget: "params",
			startAs: map[string]string{
				"scheme": "http",
			},
			want: map[string]string{
				"disabletls": "yes",
			},
			wantAbsent: []string{
				"scheme",
			},
			wantChanged: true,
		},
		{
			name:      "bark/scheme HTTP -> disabletls",
			sType:     "bark",
			mapTarget: "params",
			startAs: map[string]string{
				"scheme": "HTTP",
			},
			want: map[string]string{
				"disabletls": "yes",
			},
			wantAbsent: []string{
				"scheme",
			},
			wantChanged: true,
		},
		{
			name:      "bark/scheme https -> disabletls",
			sType:     "bark",
			mapTarget: "params",
			startAs: map[string]string{
				"scheme": "https",
			},
			want: map[string]string{
				"disabletls": "no",
			},
			wantAbsent: []string{
				"scheme",
			},
			wantChanged: true,
		},
		{
			name:      "bark/scheme HTTPS -> disabletls",
			sType:     "bark",
			mapTarget: "params",
			startAs: map[string]string{
				"scheme": "HTTPS",
			},
			want: map[string]string{
				"disabletls": "no",
			},
			wantAbsent: []string{
				"scheme",
			},
			wantChanged: true,
		},
		{
			name:      "bark/scheme -> disabletls (but disabletls already defined)",
			sType:     "bark",
			mapTarget: "params",
			startAs: map[string]string{
				"scheme":     "http",
				"disabletls": "no",
			},
			want: map[string]string{
				"disabletls": "no",
			},
			wantAbsent: []string{
				"scheme",
			},
			wantChanged: true,
		},
		{
			name:      "discord/threadid -> thread_id",
			sType:     "discord",
			mapTarget: "params",
			startAs: map[string]string{
				"threadid": "1234567890",
			},
			want: map[string]string{
				"thread_id": "1234567890",
			},
			wantAbsent: []string{
				"threadid",
			},
			wantChanged: true,
		},
		{
			name:      "discord/threadid -> thread_id (but thread_id already defined)",
			sType:     "discord",
			mapTarget: "params",
			startAs: map[string]string{
				"threadid":  "1234567890",
				"thread_id": "9876543210",
			},
			want: map[string]string{
				"thread_id": "9876543210",
			},
			wantAbsent: []string{
				"threadid",
			},
			wantChanged: true,
		},
		{
			name:      "generic/requestmethod -> method",
			sType:     "generic",
			mapTarget: "params",
			startAs: map[string]string{
				"requestmethod": "PUT",
			},
			want: map[string]string{
				"method": "PUT",
			},
			wantAbsent: []string{
				"requestmethod",
			},
			wantChanged: true,
		},
		{
			name:      "generic/requestmethod -> method (but method already defined)",
			sType:     "generic",
			mapTarget: "params",
			startAs: map[string]string{
				"requestmethod": "PUT",
				"method":        "GET",
			},
			want: map[string]string{
				"method": "GET",
			},
			wantAbsent: []string{
				"requestmethod",
			},
			wantChanged: true,
		},
		{
			name:      "generic/custom_headers -> headers",
			sType:     "generic",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"custom_headers": `{"foo":"bar"}`,
			},
			want: map[string]string{
				"headers": `{"foo":"bar"}`,
			},
			wantAbsent: []string{
				"custom_headers",
			},
			wantChanged: true,
		},
		{
			name:      "generic/custom_headers -> headers (but headers already defined)",
			sType:     "generic",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"custom_headers": `{"foo":"bar"}`,
				"headers":        `{"foo":"baz"}`,
			},
			want: map[string]string{
				"headers": `{"foo":"baz"}`,
			},
			wantAbsent: []string{
				"custom_headers",
			},
			wantChanged: true,
		},
		{
			name:      "gotify/insecureskipverify -> skiptlsverify",
			sType:     "gotify",
			mapTarget: "params",
			startAs: map[string]string{
				"insecureskipverify": "yes",
			},
			want: map[string]string{
				"skiptlsverify": "yes",
			},
			wantAbsent: []string{
				"insecureskipverify",
			},
			wantChanged: true,
		},
		{
			name:      "gotify/insecureskipverify -> skiptlsverify (but skiptlsverify already defined)",
			sType:     "gotify",
			mapTarget: "params",
			startAs: map[string]string{
				"insecureskipverify": "yes",
				"skiptlsverify":      "no",
			},
			want: map[string]string{
				"skiptlsverify": "no",
			},
			wantAbsent: []string{
				"insecureskipverify",
			},
			wantChanged: true,
		},
		{
			name:      "gotify/date removed",
			sType:     "gotify",
			mapTarget: "params",
			startAs: map[string]string{
				"date": "2006-01-02T15:04:05Z",
			},
			wantAbsent: []string{
				"date",
			},
			wantChanged: true,
		},
		{
			name:      "ifttt/usemessageasvalue -> messagevalue",
			sType:     "ifttt",
			mapTarget: "params",
			startAs: map[string]string{
				"usemessageasvalue": "2",
			},
			want: map[string]string{
				"messagevalue": "2",
			},
			wantAbsent: []string{
				"usemessageasvalue",
			},
			wantChanged: true,
		},
		{
			name:      "ifttt/usemessageasvalue -> messagevalue (but messagevalue already defined)",
			sType:     "ifttt",
			mapTarget: "params",
			startAs: map[string]string{
				"usemessageasvalue": "2",
				"messagevalue":      "3",
			},
			want: map[string]string{
				"messagevalue": "3",
			},
			wantAbsent: []string{
				"usemessageasvalue",
			},
			wantChanged: true,
		},
		{
			name:      "ifttt/usetitleasvalue -> titlevalue",
			sType:     "ifttt",
			mapTarget: "params",
			startAs: map[string]string{
				"usetitleasvalue": "2",
			},
			want: map[string]string{
				"titlevalue": "2",
			},
			wantAbsent: []string{
				"usetitleasvalue",
			},
			wantChanged: true,
		},
		{
			name:      "ifttt/usetitleasvalue -> titlevalue (but titlevalue already defined)",
			sType:     "ifttt",
			mapTarget: "params",
			startAs: map[string]string{
				"usetitleasvalue": "2",
				"titlevalue":      "3",
			},
			want: map[string]string{
				"titlevalue": "3",
			},
			wantAbsent: []string{
				"usetitleasvalue",
			},
			wantChanged: true,
		},
		{
			name:      "matrix/username -> user",
			sType:     "matrix",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"username": "argus",
			},
			want: map[string]string{
				"user": "argus",
			},
			wantAbsent: []string{
				"username",
			},
			wantChanged: true,
		},
		{
			name:      "matrix/username -> user (but user already defined)",
			sType:     "matrix",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"user":     "other",
				"username": "argus",
			},
			want: map[string]string{
				"user": "other",
			},
			wantAbsent: []string{
				"username",
			},
			wantChanged: true,
		},
		{
			name:      "matrix/rooms/leading #",
			sType:     "matrix",
			mapTarget: "params",
			startAs: map[string]string{
				"rooms": "#alias:server",
			},
			want: map[string]string{
				"rooms": "alias:server",
			},
			wantChanged: true,
		},
		{
			name:      "matrix/rooms/leading # already urlEncoded",
			sType:     "matrix",
			mapTarget: "params",
			startAs: map[string]string{
				"rooms": "%23alias:server",
			},
			want: map[string]string{
				"rooms": "%23alias:server",
			},
		},
		{
			name:      "matrix/rooms/valid",
			sType:     "matrix",
			mapTarget: "params",
			startAs: map[string]string{
				"rooms": "alias:server",
			},
			want: map[string]string{
				"rooms": "alias:server",
			},
		},
		{
			name:      "mqtt/qos, numeric alias",
			sType:     "mqtt",
			mapTarget: "params",
			startAs: map[string]string{
				"qos": "1",
			},
			want: map[string]string{
				"qos": "AtLeastOnce",
			},
			wantChanged: true,
		},
		{
			name:      "mqtt/qos, out of range left for CheckValues",
			sType:     "mqtt",
			mapTarget: "params",
			startAs: map[string]string{
				"qos": "3",
			},
			want: map[string]string{
				"qos": "3",
			},
		},
		{
			name:      "mattermost/channel, leading slash",
			sType:     "mattermost",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"channel": "/argus",
			},
			want: map[string]string{
				"channel": "argus",
			},
			wantChanged: true,
		},
		{
			name:      "mattermost/channel, valid",
			sType:     "mattermost",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"channel": "argus",
			},
			want: map[string]string{
				"channel": "argus",
			},
		},
		{
			name:      "opsgenie/priority 3 -> P3",
			sType:     "opsgenie",
			mapTarget: "params",
			startAs: map[string]string{
				"priority": "3",
			},
			want: map[string]string{
				"priority": "P3",
			},
			wantChanged: true,
		},
		{
			name:      "opsgenie/priority P2 kept",
			sType:     "opsgenie",
			mapTarget: "params",
			startAs: map[string]string{
				"priority": "P2",
			},
			want: map[string]string{
				"priority": "P2",
			},
		},
		{
			name:      "opsgenie/priority 6 kept",
			sType:     "opsgenie",
			mapTarget: "params",
			startAs: map[string]string{
				"priority": "6",
			},
			want: map[string]string{
				"priority": "6",
			},
		},
		{
			name:      "ntfy/disabletlsverification -> skiptlsverify",
			sType:     "ntfy",
			mapTarget: "params",
			startAs: map[string]string{
				"disabletlsverification": "true",
			},
			want: map[string]string{
				"skiptlsverify": "true",
			},
			wantAbsent: []string{
				"disabletlsverification",
			},
			wantChanged: true,
		},
		{
			name:      "ntfy/disabletlsverification -> skiptlsverify (but skiptlsverify already defined)",
			sType:     "ntfy",
			mapTarget: "params",
			startAs: map[string]string{
				"disabletlsverification": "true",
				"skiptlsverify":          "false",
			},
			want: map[string]string{
				"skiptlsverify": "false",
			},
			wantAbsent: []string{
				"disabletlsverification",
			},
			wantChanged: true,
		},
		{
			name:      "ntfy/disabletls -> skiptlsverify",
			sType:     "ntfy",
			mapTarget: "params",
			startAs: map[string]string{
				"disabletls": "yes",
			},
			want: map[string]string{
				"skiptlsverify": "yes",
			},
			wantAbsent: []string{
				"disabletls",
			},
			wantChanged: true,
		},
		{
			name:      "ntfy/disabletls -> skiptlsverify (but skiptlsverify already defined)",
			sType:     "ntfy",
			mapTarget: "params",
			startAs: map[string]string{
				"disabletls":    "yes",
				"skiptlsverify": "no",
			},
			want: map[string]string{
				"skiptlsverify": "no",
			},
			wantAbsent: []string{
				"disabletls",
			},
			wantChanged: true,
		},
		{
			name:      "ntfy/scheme http -> disabletls_",
			sType:     "ntfy",
			mapTarget: "params",
			startAs: map[string]string{
				"scheme": "http",
			},
			want: map[string]string{
				"disabletls_": "yes",
			},
			wantAbsent: []string{
				"scheme",
			},
			wantChanged: true,
		},
		{
			name:      "ntfy/scheme HTTP -> disabletls_",
			sType:     "ntfy",
			mapTarget: "params",
			startAs: map[string]string{
				"scheme": "HTTP",
			},
			want: map[string]string{
				"disabletls_": "yes",
			},
			wantAbsent: []string{
				"scheme",
			},
			wantChanged: true,
		},
		{
			name:      "ntfy/scheme https -> disabletls_",
			sType:     "ntfy",
			mapTarget: "params",
			startAs: map[string]string{
				"scheme": "https",
			},
			want: map[string]string{
				"disabletls_": "no",
			},
			wantAbsent: []string{
				"scheme",
			},
			wantChanged: true,
		},
		{
			name:      "ntfy/scheme HTTPS -> disabletls_",
			sType:     "ntfy",
			mapTarget: "params",
			startAs: map[string]string{
				"scheme": "HTTPS",
			},
			want: map[string]string{
				"disabletls_": "no",
			},
			wantAbsent: []string{
				"scheme",
			},
			wantChanged: true,
		},
		{
			name:      "ntfy/scheme -> disabletls_ (but disabletls_ already defined)",
			sType:     "ntfy",
			mapTarget: "params",
			startAs: map[string]string{
				"scheme":      "http",
				"disabletls_": "no",
			},
			want: map[string]string{
				"disabletls_": "no",
			},
			wantAbsent: []string{
				"scheme",
			},
			wantChanged: true,
		},
		{
			name:      "ntfy/disabletls and scheme -> skiptlsverify and disabletls_",
			sType:     "ntfy",
			mapTarget: "params",
			startAs: map[string]string{
				"disabletls": "yes",
				"scheme":     "http",
			},
			want: map[string]string{
				"disabletls_":   "yes",
				"skiptlsverify": "yes",
			},
			wantAbsent: []string{
				"disabletls",
				"scheme",
			},
			wantChanged: true,
		},
		{
			name:      "signal/apikey -> token",
			sType:     "signal",
			mapTarget: "params",
			startAs: map[string]string{
				"apikey": "secret",
			},
			want: map[string]string{
				"token": "secret",
			},
			wantAbsent: []string{
				"apikey",
			},
			wantChanged: true,
		},
		{
			name:      "signal/apikey -> token (but token already defined)",
			sType:     "signal",
			mapTarget: "params",
			startAs: map[string]string{
				"apikey": "secret",
				"token":  "other",
			},
			want: map[string]string{
				"token": "other",
			},
			wantAbsent: []string{
				"apikey",
			},
			wantChanged: true,
		},
		{
			name:      "slack/color, already #-prefixed",
			sType:     "slack",
			mapTarget: "params",
			startAs: map[string]string{
				"color": "#ffffff",
			},
			want: map[string]string{
				"color": "#ffffff",
			},
		},
		{
			name:      "slack/color, bare hex (no #)",
			sType:     "slack",
			mapTarget: "params",
			startAs: map[string]string{
				"color": "ffffff",
			},
			want: map[string]string{
				"color": "#ffffff",
			},
			wantChanged: true,
		},
		{
			name:      "slack/color, bare hex (no #), uppercase",
			sType:     "slack",
			mapTarget: "params",
			startAs: map[string]string{
				"color": "FFFFFF",
			},
			want: map[string]string{
				"color": "#FFFFFF",
			},
			wantChanged: true,
		},
		{
			name:      "slack/color, named color (non-hex, no # needed)",
			sType:     "slack",
			mapTarget: "params",
			startAs: map[string]string{
				"color": "good",
			},
			want: map[string]string{
				"color": "good",
			},
		},
		{
			name:      "slack/color, migrate old %23-encoded format",
			sType:     "slack",
			mapTarget: "params",
			startAs: map[string]string{
				"color": "%23ffffff",
			},
			want: map[string]string{
				"color": "#ffffff",
			},
			wantChanged: true,
		},
		{
			name:      "slack/threadts -> thread_ts",
			sType:     "slack",
			mapTarget: "params",
			startAs: map[string]string{
				"threadts": "1609459200.000001",
			},
			want: map[string]string{
				"thread_ts": "1609459200.000001",
			},
			wantAbsent: []string{
				"threadts",
			},
			wantChanged: true,
		},
		{
			name:      "slack/threadts -> thread_ts (but thread_ts already defined)",
			sType:     "slack",
			mapTarget: "params",
			startAs: map[string]string{
				"threadts":  "1609459200.000001",
				"thread_ts": "1700000000.000002",
			},
			want: map[string]string{
				"thread_ts": "1700000000.000002",
			},
			wantAbsent: []string{
				"threadts",
			},
			wantChanged: true,
		},
		{
			name:      "smtp/skiptlsverification -> skiptlsverify",
			sType:     "smtp",
			mapTarget: "params",
			startAs: map[string]string{
				"skiptlsverification": "true",
			},
			want: map[string]string{
				"skiptlsverify": "true",
			},
			wantAbsent: []string{
				"skiptlsverification",
			},
			wantChanged: true,
		},
		{
			name:      "smtp/skiptlsverification -> skiptlsverify (but skiptlsverify already defined)",
			sType:     "smtp",
			mapTarget: "params",
			startAs: map[string]string{
				"skiptlsverification": "true",
				"skiptlsverify":       "false",
			},
			want: map[string]string{
				"skiptlsverify": "false",
			},
			wantAbsent: []string{
				"skiptlsverification",
			},
			wantChanged: true,
		},
		{
			name:      "smtp/timeout, valid duration",
			sType:     "smtp",
			mapTarget: "params",
			startAs: map[string]string{
				"timeout": "0h0m10s",
			},
			want: map[string]string{
				"timeout": "0h0m10s",
			},
		},
		{
			name:      "smtp/timeout, integer treated as seconds",
			sType:     "smtp",
			mapTarget: "params",
			startAs: map[string]string{
				"timeout": "10",
			},
			want: map[string]string{
				"timeout": "10s",
			},
			wantChanged: true,
		},
		{
			name:      "smtp/timeout, invalid left for CheckValues",
			sType:     "smtp",
			mapTarget: "params",
			startAs: map[string]string{
				"timeout": "abc",
			},
			want: map[string]string{
				"timeout": "abc",
			},
		},
		{
			name:      "zulip/botmail, not urlEncoded",
			sType:     "zulip",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"botmail": "foo@bar.com",
			},
			want: map[string]string{
				"botmail": "foo%40bar.com",
			},
			wantChanged: true,
		},
		{
			name:      "zulip/botmail, valid",
			sType:     "zulip",
			mapTarget: "url_fields",
			startAs: map[string]string{
				"botmail": "foo%40bar.com",
			},
			want: map[string]string{
				"botmail": "foo%40bar.com",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := New(
				nil, "",
				tc.sType,
				make(map[string]string),
				make(map[string]string),
				make(map[string]string),
				NewDefaults(
					"",
					make(map[string]string),
					make(map[string]string),
					make(map[string]string),
				),
				NewDefaults(
					"",
					make(map[string]string),
					make(map[string]string),
					make(map[string]string),
				),
				NewDefaults(
					"",
					make(map[string]string),
					make(map[string]string),
					make(map[string]string),
				),
			)
			serviceStatus := newTestStatus(tc.name, 1)
			shoutrrr.Init(
				serviceStatus,
				shoutrrr.Main, shoutrrr.Defaults, shoutrrr.HardDefaults,
			)
			subTests := []struct {
				name      string
				layer     string
				URLFields map[string]string
				Params    map[string]string
			}{
				{
					name:      "root",
					layer:     "Shoutrrr",
					URLFields: shoutrrr.URLFields,
					Params:    shoutrrr.Params,
				},
				{
					name:      "main",
					layer:     "Shoutrrr.Main",
					URLFields: shoutrrr.Main.URLFields,
					Params:    shoutrrr.Main.Params,
				},
				{
					name:      "defaults",
					layer:     "Shoutrrr.Defaults",
					URLFields: shoutrrr.Defaults.URLFields,
					Params:    shoutrrr.Defaults.Params,
				},
				{
					name:      "hard_defaults",
					layer:     "Shoutrrr.HardDefaults",
					URLFields: shoutrrr.HardDefaults.URLFields,
					Params:    shoutrrr.HardDefaults.Params,
				},
			}
			// Set in each location in turn; only root is corrected.
			for _, subTest := range subTests {
				t.Run(subTest.name, func(t *testing.T) {
					// Serial: shares shoutrrr.

					targetMap, field := subTest.Params, subTest.layer+".Params"
					if tc.mapTarget == "url_fields" {
						targetMap, field = subTest.URLFields, subTest.layer+".URLFields"
					}
					maps.Copy(targetMap, tc.startAs)

					changed := shoutrrr.correctSelf(shoutrrr.GetType())

					// correctSelf edits only its receiver.
					wantChanged := tc.wantChanged && subTest.name == "root"
					if changed != wantChanged {
						t.Errorf("changed mismatch\ngot:  %t\nwant: %t", changed, wantChanged)
					}

					for k, v := range tc.want {
						want := v
						if subTest.name != "root" {
							want = tc.startAs[k]
						}
						if got := targetMap[k]; got != want {
							t.Errorf("%s[%q] mismatch\ngot:  %q\nwant: %q", field, k, got, want)
						}
						delete(targetMap, k)
					}

					if subTest.name == "root" {
						for _, k := range tc.wantAbsent {
							if _, found := targetMap[k]; found {
								t.Errorf("%s[%q] mismatch\ngot:  present\nwant: absent", field, k)
							}
						}
					}
					for _, k := range tc.wantAbsent {
						delete(subTest.URLFields, k)
						delete(subTest.Params, k)
					}
				})
			}
		})
	}
}

func TestBase_NormaliseParamSelect(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		allowed    []string
		startParam string
		ok         bool
		wantValue  string
	}{
		{
			name:      "exact match uses canonical case",
			value:     "Two",
			allowed:   []string{"One", "Two", "Three"},
			ok:        true,
			wantValue: "Two",
		},
		{
			name:      "case-insensitive match lower->upper",
			value:     "two",
			allowed:   []string{"One", "Two", "Three"},
			ok:        true,
			wantValue: "Two",
		},
		{
			name:      "case-insensitive match upper->proper",
			value:     "THREE",
			allowed:   []string{"One", "Two", "Three"},
			ok:        true,
			wantValue: "Three",
		},
		{
			name:       "non-match returns false and leaves value unchanged",
			value:      "four",
			allowed:    []string{"One", "Two", "Three"},
			startParam: "unchanged",
			ok:         false,
			wantValue:  "unchanged",
		},
		{
			name:       "empty value returns false and makes no change",
			value:      "",
			allowed:    []string{"One", "Two"},
			startParam: "keepme",
			ok:         false,
			wantValue:  "keepme",
		},
		{
			name:       "empty allowed set never matches",
			value:      "one",
			allowed:    []string{},
			startParam: "keepme",
			ok:         false,
			wantValue:  "keepme",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			key := "param"
			b := &Base{}
			b.InitMaps()
			if tc.startParam != "" {
				b.setParam(key, tc.startParam)
			}

			ok := b.normaliseParamSelect(key, tc.value, tc.allowed)

			if ok != tc.ok {
				t.Fatalf("ok mismatch\ngot:  %t\nwant: %t", ok, tc.ok)
			}

			if got := b.GetParam(key); got != tc.wantValue {
				t.Fatalf("Base.GetParam() mismatch\ngot:  %q\nwant: %q", got, tc.wantValue)
			}
		})
	}
}

func TestShoutrrr_CheckURL(t *testing.T) {
	tests := []struct {
		name              string
		sType             string
		urlFields, params map[string]string
		errRegex          string
	}{
		{
			name:  "homeassistant/fail Locate on an out-of-range port",
			sType: "homeassistant",
			urlFields: map[string]string{
				"host":  "homeassistant.example.com",
				"port":  "99999",
				"token": "bish",
			},
			errRegex: `invalid URL port`,
		},
		{
			name:  "gotify/fail CreateSender",
			sType: "gotify",
			urlFields: map[string]string{
				"host":  "https://	example.com",
				"token": "bish",
			},
			errRegex: `failed to parse URL`,
		},
		{
			name:  "pagerduty/invalid integrationkey",
			sType: "pagerduty",
			urlFields: map[string]string{
				"integrationkey": "KEY",
			},
			errRegex: `^pagerduty:\n.*invalid integration key format`,
		},
		{
			name:  "twilio/invalid accountsid",
			sType: "twilio",
			urlFields: map[string]string{
				"accountsid": "SID",
				"authtoken":  "TOKEN",
				"fromnumber": "+15551234567",
				"tonumbers":  "+15557654321",
			},
			errRegex: `^twilio:\n.*account SID must be AC followed by 32 hexadecimal characters`,
		},
		{
			name:  "teams/legacy skips Locate",
			sType: "teams",
			urlFields: map[string]string{
				"altid":      "ALTID",
				"group":      "GROUP",
				"groupowner": "GROUPOWNER",
				"tenant":     "TENANT",
			},
			errRegex: `^$`,
		},
		{
			name:  "twilio/valid",
			sType: "twilio",
			urlFields: map[string]string{
				"accountsid": testTwilioSID,
				"authtoken":  "TOKEN",
				"fromnumber": "+15551234567",
				"tonumbers":  "+15557654321",
			},
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := testShoutrrr(false, false)
			s.Type = tc.sType
			s.Main = &Defaults{}
			s.Main.InitMaps()
			s.URLFields = util.CopyMap(tc.urlFields)
			s.Params = util.CopyMap(tc.params)

			err := s.CheckURL()

			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

func TestShoutrrrs_CheckURLs(t *testing.T) {
	pagerDuty := func(integrationKey string) *Shoutrrr {
		return New(
			nil, "", "pagerduty",
			nil, map[string]string{"integrationkey": integrationKey}, nil,
			&Defaults{}, &Defaults{}, &Defaults{},
		)
	}
	validKey := "0123456789abcdef0123456789abcdef"

	tests := []struct {
		name      string
		shoutrrrs *Shoutrrrs
		errRegex  string
	}{
		{
			name:      "nil",
			shoutrrrs: nil,
			errRegex:  `^$`,
		},
		{
			name:      "valid",
			shoutrrrs: &Shoutrrrs{"bar": pagerDuty(validKey)},
			errRegex:  `^$`,
		},
		{
			name:      "invalid/keyed by ID",
			shoutrrrs: &Shoutrrrs{"bar": pagerDuty(validKey), "foo": pagerDuty("KEY")},
			errRegex:  `^foo:\n  pagerduty:\n    invalid integration key format`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.shoutrrrs.CheckURLs()

			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

func TestShoutrrrs_LogWarnings(t *testing.T) {
	notify := func(sType string, urlFields, params map[string]string) *Shoutrrr {
		return New(
			nil, "", sType,
			nil, urlFields, params,
			&Defaults{}, &Defaults{}, &Defaults{},
		)
	}
	gotifyURLFields := map[string]string{"host": "example.com", "token": "AAAAAAAAAAAAAAA"}

	tests := []struct {
		name      string
		shoutrrrs *Shoutrrrs
		want      string
	}{
		{
			name:      "nil",
			shoutrrrs: nil,
			want:      `^$`,
		},
		{
			name:      "valid",
			shoutrrrs: &Shoutrrrs{"foo": notify("gotify", gotifyURLFields, nil)},
			want:      `^$`,
		},
		{
			name: "unlocatable",
			shoutrrrs: &Shoutrrrs{
				"foo": notify("pagerduty", map[string]string{"integrationkey": "KEY"}, nil),
			},
			want: `^WARNING: service_id \(foo\), failed to locate notify, so it will fail to send: pagerduty: invalid integration key`,
		},
		{
			name: "unknown params",
			shoutrrrs: &Shoutrrrs{
				"foo": notify("gotify", gotifyURLFields, map[string]string{"bogus": "x", "priority": "1"}),
			},
			want: `^WARNING: service_id \(foo\), dropping unknown params \["bogus"\]\n$`,
		},
		{
			name: "ntfy disabletls_ known",
			shoutrrrs: &Shoutrrrs{
				"foo": notify("ntfy", map[string]string{"topic": "TOPIC"}, map[string]string{"disabletls_": "yes"}),
			},
			want: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			tc.shoutrrrs.LogWarnings("service_id")

			stdout := releaseStdout()
			if !util.RegexCheck(tc.want, stdout) {
				t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.want)
			}
		})
	}
}

func TestShoutrrr_CheckValuesType(t *testing.T) {
	tests := []struct {
		name     string
		sType    string
		main     *Defaults
		errRegex string
	}{
		{
			name:     "no type",
			errRegex: `^type: <required>.*$`,
			sType:    "",
		},
		{
			name:     "invalid type/unknown",
			errRegex: `^type: .* <invalid>.*$`,
			sType:    "argus",
		},
		{
			name:     "invalid type/type in main differs",
			errRegex: `type: "gotify" <invalid> .*\(discord\)\)`,
			sType:    "gotify",
			main: NewDefaults(
				"discord",
				make(map[string]string),
				make(map[string]string),
				make(map[string]string),
			),
		},
		{
			name:     "valid type",
			errRegex: `^$`,
			sType:    "gotify",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := testShoutrrr(false, false)
			input.Type = tc.sType
			input.Init(
				newTestStatus(tc.name, 1),
				tc.main,
				&Defaults{}, &Defaults{},
			)

			_, _ = test.AssertCheckValuesWithErrorAndChanged(
				t,
				tc.errRegex,
				false,
				input.CheckValues,
			)
		})
	}
}

func TestBase_CheckValuesOptions(t *testing.T) {
	tests := []struct {
		name      string
		options   map[string]string
		wantDelay string
		errRegex  string
	}{
		{
			name:     "no options",
			options:  map[string]string{},
			errRegex: `^$`,
		},
		{
			name: "valid delay",
			options: map[string]string{
				"delay": "5s",
			},
			wantDelay: "5s",
			errRegex:  `^$`,
		},
		{
			name: "invalid delay",
			options: map[string]string{
				"delay": "5x",
			},
			errRegex: `^delay: "5x" <invalid>.*$`,
		},
		{
			name: "fixes delay missing unit",
			options: map[string]string{
				"delay": "5",
			},
			wantDelay: "5s",
			errRegex:  `^$`,
		},
		{
			name: "valid message template",
			options: map[string]string{
				"message": "release! {{ version }}",
			},
			errRegex: `^$`,
		},
		{
			name: "invalid message template",
			options: map[string]string{
				"message": "release! {{ version }",
			},
			errRegex: `^message: "release! {{ version }" <invalid>.*$`,
		},
		{
			name: "min max_tries",
			options: map[string]string{
				"max_tries": "0",
			},
			errRegex: `^$`,
		},
		{
			name: "max max_tries",
			options: map[string]string{
				"max_tries": strconv.Itoa(math.MaxUint8),
			},
			errRegex: `^$`,
		},
		{
			name: "invalid max_tries/too large, MaxUint16",
			options: map[string]string{
				"max_tries": strconv.Itoa(math.MaxUint16),
			},
			errRegex: `^max_tries: "\d+" <invalid>.*$`,
		},
		{
			name: "invalid max_tries/too large, >MaxUint64",
			options: map[string]string{
				"max_tries": fmt.Sprintf("1%d", uint(math.MaxUint64)),
			},
			errRegex: `^max_tries: "\d+" <invalid>.*$`,
		},
		{
			name: "invalid max_tries/not a number",
			options: map[string]string{
				"max_tries": "oneOrTwo",
			},
			errRegex: `^max_tries: "oneOrTwo" <invalid>.*$`,
		},
		{
			name: "invalid max_tries/negative",
			options: map[string]string{
				"max_tries": "-1",
			},
			errRegex: `^max_tries: "-1" <invalid>.*$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := Base{
				Options: tc.options,
			}

			err := test.AssertCheckValuesWithError(
				t,
				tc.errRegex,
				input.checkValuesOptions,
			)
			e := errfmt.FormatError(err)
			if !util.RegexCheck("^delay: .*", e) {
				if got := input.getOption("delay"); got != tc.wantDelay {
					t.Errorf("Base.getOption() mismatch\ngot:  %q\nwant: %q", got, tc.wantDelay)
				}
			}
		})
	}
}

func TestShoutrrr_CheckValuesURLFields(t *testing.T) {
	tests := []struct {
		name      string
		sType     string
		urlFields map[string]string
		main      *Defaults
		errRegex  string
	}{
		{
			name:  "bark/invalid",
			sType: "bark",
			errRegex: test.TrimYAML(`
				^devicekey: <required>.*
				host: <required>.*$`,
			),
		},
		{
			name:  "bark/no devicekey",
			sType: "bark",
			urlFields: map[string]string{
				"host": "https://example.com",
			},
			errRegex: `^devicekey: <required>.*$`,
		},
		{
			name:  "bark/no host",
			sType: "bark",
			urlFields: map[string]string{
				"devicekey": "foo",
			},
			errRegex: `^host: <required>.*$`,
		},
		{
			name:  "bark/valid",
			sType: "bark",
			urlFields: map[string]string{
				"devicekey": "foo",
				"host":      "https://example.com",
			},
			errRegex: `^$`,
		},
		{
			name:  "discord/invalid",
			sType: "discord",
			errRegex: test.TrimYAML(`
				^token: <required>.*
				webhookid: <required>.*$`,
			),
		},
		{
			name:  "discord/no token",
			sType: "discord",
			urlFields: map[string]string{
				"webhookid": "bash",
			},
			errRegex: `^token: <required>.*$`,
		},
		{
			name:  "discord/no webhookid",
			sType: "discord",
			urlFields: map[string]string{
				"token": "bish",
			},
			errRegex: `^webhookid: <required>.*$`,
		},
		{
			name:  "discord/valid",
			sType: "discord",
			urlFields: map[string]string{
				"token":     "bish",
				"webhookid": "webhookid",
			},
			errRegex: `^$`,
		},
		{
			name:  "discord/valid with main",
			sType: "discord",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"token":     "bish",
					"webhookid": "bash",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:      "smtp/no host",
			sType:     "smtp",
			urlFields: map[string]string{},
			errRegex:  `^host: <required>.*$`,
		},
		{
			name:  "smtp/valid",
			sType: "smtp",
			urlFields: map[string]string{
				"host": "bish",
			},
			errRegex: `^$`,
		},
		{
			name:  "smtp/valid with main",
			sType: "smtp",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"host": "bish",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:  "gotify/invalid",
			sType: "gotify",
			errRegex: test.TrimYAML(`
				^host: <required>.*
				token: <required>.*$`,
			),
		},
		{
			name:  "gotify/no host",
			sType: "gotify",
			urlFields: map[string]string{
				"token": "bash",
			},
			errRegex: `^host: <required>.*$`,
		},
		{
			name:  "gotify/no token",
			sType: "gotify",
			urlFields: map[string]string{
				"host": "bish",
			},
			errRegex: `^token: <required>.*$`,
		},
		{
			name:  "gotify/valid",
			sType: "gotify",
			urlFields: map[string]string{
				"host":  "bish",
				"token": "bash",
			},
			errRegex: `^$`,
		},
		{
			name:  "gotify/valid with main",
			sType: "gotify",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"host":  "bish",
					"token": "bash",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:     "googlechat/invalid",
			sType:    "googlechat",
			errRegex: `^raw: <required>.*$`,
		},
		{
			name:  "googlechat/valid",
			sType: "googlechat",
			urlFields: map[string]string{
				"raw": "bish",
			},
			errRegex: `^$`,
		},
		{
			name:  "googlechat/valid with main",
			sType: "googlechat",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"raw": "bish",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:  "homeassistant/invalid",
			sType: "homeassistant",
			errRegex: test.TrimYAML(`
				^token: <required>.*
				host: <required>.*$`,
			),
		},
		{
			name:  "homeassistant/no token",
			sType: "homeassistant",
			urlFields: map[string]string{
				"host": "homeassistant.example.com",
			},
			errRegex: `^token: <required>.*$`,
		},
		{
			name:  "homeassistant/no host",
			sType: "homeassistant",
			urlFields: map[string]string{
				"token": "bash",
			},
			errRegex: `^host: <required>.*$`,
		},
		{
			name:  "homeassistant/valid",
			sType: "homeassistant",
			urlFields: map[string]string{
				"host":  "homeassistant.example.com",
				"token": "bash",
			},
			errRegex: `^$`,
		},
		{
			name:  "homeassistant/valid with main",
			sType: "homeassistant",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"host":  "homeassistant.example.com",
					"token": "bash",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:      "ifttt/no webhookid",
			sType:     "ifttt",
			urlFields: map[string]string{},
			errRegex:  `^webhookid: <required>.*$`,
		},
		{
			name:  "ifttt/valid",
			sType: "ifttt",
			urlFields: map[string]string{
				"webhookid": "bish",
			},
			errRegex: `^$`,
		},
		{
			name:  "ifttt/valid with main",
			sType: "ifttt",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"webhookid": "webhookid",
				},
				map[string]string{
					"events": "events",
				},
			),
			errRegex: `^$`,
		},
		{
			name:      "join/no apikey",
			sType:     "join",
			urlFields: map[string]string{},
			errRegex:  `^apikey: <required>.*$`,
		},
		{
			name:  "join/valid",
			sType: "join",
			urlFields: map[string]string{
				"apikey": "bish",
			},
			errRegex: `^$`,
		},
		{
			name:  "join/valid with main",
			sType: "join",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"apikey": "apikey",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:  "mattermost/invalid",
			sType: "mattermost",
			errRegex: test.TrimYAML(`
				^host: <required>.*
				token: <required>.*$`,
			),
		},
		{
			name:  "mattermost/no host",
			sType: "mattermost",
			urlFields: map[string]string{
				"token": "bash",
			},
			errRegex: `^host: <required>.*$`,
		},
		{
			name:  "mattermost/no token",
			sType: "mattermost",
			urlFields: map[string]string{
				"host": "bish",
			},
			errRegex: `^token: <required>.*$`,
		},
		{
			name:  "mattermost/valid",
			sType: "mattermost",
			urlFields: map[string]string{
				"host":  "bish",
				"token": "bash",
			},
			errRegex: `^$`,
		},
		{
			name:  "mattermost/valid with main",
			sType: "mattermost",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"host":  "bish",
					"token": "bash",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:  "matrix/invalid",
			sType: "matrix",
			errRegex: test.TrimYAML(`
				^host: <required>.*
				password: <required>.*$`,
			),
		},
		{
			name:  "matrix/no host",
			sType: "matrix",
			urlFields: map[string]string{
				"password": "bash",
			},
			errRegex: `^host: <required>.*$`,
		},
		{
			name:  "matrix/no password",
			sType: "matrix",
			urlFields: map[string]string{
				"host": "bish",
			},
			errRegex: `^password: <required>.*$`,
		},
		{
			name:  "matrix/valid",
			sType: "matrix",
			urlFields: map[string]string{
				"host":     "bish",
				"password": "password",
			},
			errRegex: `^$`,
		},
		{
			name:  "matrix/valid with main",
			sType: "matrix",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"host":     "bish",
					"password": "bash",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:     "notifiarr/invalid",
			sType:    "notifiarr",
			errRegex: `^apikey: <required>.*$`,
		},
		{
			name:  "notifiarr/valid",
			sType: "notifiarr",
			urlFields: map[string]string{
				"apikey": "abc123",
			},
			errRegex: `^$`,
		},
		{
			name:  "notifiarr/valid with main",
			sType: "notifiarr",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"apikey": "abc123",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:     "ntfy/invalid",
			sType:    "ntfy",
			errRegex: `^topic: <required>.*$`,
		},
		{
			name:  "ntfy/valid",
			sType: "ntfy",
			urlFields: map[string]string{
				"topic": "foo",
			},
			errRegex: `^$`,
		},
		{
			name:     "opsgenie/invalid",
			sType:    "opsgenie",
			errRegex: `^apikey: <required>.*$`,
		},
		{
			name:  "opsgenie/valid",
			sType: "opsgenie",
			urlFields: map[string]string{
				"apikey": "apikey",
			},
			errRegex: `^$`,
		},
		{
			name:  "opsgenie/valid with main",
			sType: "opsgenie",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"apikey": "apikey",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:  "pushbullet/invalid",
			sType: "pushbullet",
			errRegex: test.TrimYAML(`
				^token: <required>.*
				targets: <required>.*$`,
			),
		},
		{
			name:  "pushbullet/no token",
			sType: "pushbullet",
			urlFields: map[string]string{
				"targets": "bash",
			},
			errRegex: `^token: <required>.*$`,
		},
		{
			name:  "pushbullet/no targets",
			sType: "pushbullet",
			urlFields: map[string]string{
				"token": "bish",
			},
			errRegex: `^targets: <required>.*$`,
		},
		{
			name:  "pushbullet/valid",
			sType: "pushbullet",
			urlFields: map[string]string{
				"token":   "bish",
				"targets": "targets",
			},
			errRegex: `^$`,
		},
		{
			name:  "pushbullet/valid with main",
			sType: "pushbullet",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"token":   "bish",
					"targets": "bash",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:  "pushover/invalid",
			sType: "pushover",
			errRegex: test.TrimYAML(`
				^token: <required>.*
				user: <required>.*$`,
			),
		},
		{
			name:  "pushover/no token",
			sType: "pushover",
			urlFields: map[string]string{
				"user": "bash",
			},
			errRegex: `^token: <required>.*$`,
		},
		{
			name:  "pushover/no user",
			sType: "pushover",
			urlFields: map[string]string{
				"token": "bish",
			},
			errRegex: `^user: <required>.*$`,
		},
		{
			name:  "pushover/valid",
			sType: "pushover",
			urlFields: map[string]string{
				"token": "bish",
				"user":  "user",
			},
			errRegex: `^$`,
		},
		{
			name:  "pushover/valid with main",
			sType: "pushover",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"token": "bish",
					"user":  "bash",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:  "rocketchat/invalid",
			sType: "rocketchat",
			errRegex: test.TrimYAML(`
				^host: <required>.*
				tokena: <required>.*
				tokenb: <required>.*$`,
			),
		},
		{
			name:  "rocketchat/no host",
			sType: "rocketchat",
			urlFields: map[string]string{
				"tokena":  "bash",
				"tokenb":  "bosh",
				"channel": "bing",
			},
			errRegex: `^host: <required>.*$`,
		},
		{
			name:  "rocketchat/no tokena",
			sType: "rocketchat",
			urlFields: map[string]string{
				"host":    "bish",
				"tokenb":  "bash",
				"channel": "bing",
			},
			errRegex: `^tokena: <required>.*$`,
		},
		{
			name:  "rocketchat/no tokenb",
			sType: "rocketchat",
			urlFields: map[string]string{
				"host":    "bish",
				"tokena":  "bash",
				"channel": "bing",
			},
			errRegex: `^tokenb: <required>.*$`,
		},
		{
			name:  "rocketchat/no channel",
			sType: "rocketchat",
			urlFields: map[string]string{
				"host":   "bish",
				"tokena": "bash",
				"tokenb": "bosh",
			},
			errRegex: `^$`,
		},
		{
			name:  "rocketchat/valid",
			sType: "rocketchat",
			urlFields: map[string]string{
				"host":    "bish",
				"tokena":  "bash",
				"tokenb":  "bosh",
				"channel": "bing",
			},
			errRegex: `^$`,
		},
		{
			name:  "rocketchat/valid with main",
			sType: "rocketchat",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"host":    "bish",
					"tokena":  "bash",
					"tokenb":  "bosh",
					"channel": "bing",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:  "slack/invalid",
			sType: "slack",
			errRegex: test.TrimYAML(`
				^token: <required>.*
				channel: <required>.*$`,
			),
		},
		{
			name:  "slack/no token",
			sType: "slack",
			urlFields: map[string]string{
				"channel": "bash",
			},
			errRegex: `^token: <required>.*$`,
		},
		{
			name:  "slack/no channel",
			sType: "slack",
			urlFields: map[string]string{
				"token": "bish",
			},
			errRegex: `^channel: <required>.*$`,
		},
		{
			name:  "slack/valid",
			sType: "slack",
			urlFields: map[string]string{
				"token":   "bish",
				"channel": "channel",
			},
			errRegex: `^$`,
		},
		{
			name:  "slack/valid with main",
			sType: "slack",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"token":   "bish",
					"channel": "bash",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:      "teams/no url_fields (Power Automate format - no url fields needed)",
			sType:     "teams",
			urlFields: map[string]string{},
			errRegex:  `^$`,
		},
		{
			name:      "shoutrrr/no raw",
			sType:     "shoutrrr",
			urlFields: map[string]string{},
			errRegex:  `^raw: <required>.*$`,
		},
		{
			name:  "shoutrrr/valid",
			sType: "shoutrrr",
			urlFields: map[string]string{
				"raw": "slack://TOKEN@CHANNEL",
			},
			errRegex: `^$`,
		},
		{
			name:  "shoutrrr/valid with main",
			sType: "shoutrrr",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"raw": "slack://TOKEN@CHANNEL",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:      "telegram/no token",
			sType:     "telegram",
			urlFields: map[string]string{},
			errRegex:  `^token: <required>.*$`,
		},
		{
			name:  "telegram/valid",
			sType: "telegram",
			urlFields: map[string]string{
				"token": "bish",
			},
			errRegex: `^$`,
		},
		{
			name:  "telegram/valid with main",
			sType: "telegram",
			main: NewDefaults(
				"",
				nil,
				map[string]string{
					"token": "bish",
				},
				map[string]string{
					"chats": "chats",
				},
			),
			errRegex: `^$`,
		},
		{
			name:  "zulip/invalid",
			sType: "zulip",
			errRegex: test.TrimYAML(`
				^host: <required>.*
				botmail: <required>.*
				botkey: <required>.*$`,
			),
		},
		{
			name:  "zulip/no host",
			sType: "zulip",
			urlFields: map[string]string{
				"botmail": "bash",
				"botkey":  "bosh",
			},
			errRegex: `^host: <required>.*$`,
		},
		{
			name:  "zulip/no botmail",
			sType: "zulip",
			urlFields: map[string]string{
				"host":   "bish",
				"botkey": "bash",
			},
			errRegex: `^botmail: <required>.*$`,
		},
		{
			name:  "zulip/no botkey",
			sType: "zulip",
			urlFields: map[string]string{
				"host":    "bish",
				"botmail": "bash",
			},
			errRegex: `^botkey: <required>.*$`,
		},
		{
			name:  "zulip/valid",
			sType: "zulip",
			urlFields: map[string]string{
				"host":    "bish",
				"botmail": "bash",
				"botkey":  "bosh",
			},
			errRegex: `^$`,
		},
		{
			name:  "zulip/valid with main",
			sType: "zulip",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"host":    "bish",
					"botmail": "bash",
					"botkey":  "bosh",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:     "generic/invalid",
			sType:    "generic",
			errRegex: `^host: <required>.*$`,
		},
		{
			name:  "generic/valid",
			sType: "generic",
			urlFields: map[string]string{
				"host": "example.com",
			},
			errRegex: `^$`,
		},
		{
			name:  "generic/valid with main",
			sType: "generic",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"host": "example.com",
				},
				nil,
			),
			errRegex: `^$`,
		},
		{
			name:  "generic/valid with headers, json_payload_vars, query_vars",
			sType: "generic",
			urlFields: map[string]string{
				"host":              "example.com",
				"headers":           `{"foo":"bar"}`,
				"json_payload_vars": `{"bish":"bash","bosh":"bing"}`,
				"query_vars":        `{"me":"work"}`,
			},
			errRegex: `^$`,
		},
		{
			name:  "generic/invalid headers",
			sType: "generic",
			urlFields: map[string]string{
				"host":    "example.com",
				"headers": `"foo":"bar"}`,
			},
			errRegex: `^headers: "\\\"foo\\\":\\\"bar\\\"}" <invalid>.*$`,
		},
		{
			name:  "generic/invalid json_payload_vars",
			sType: "generic",
			urlFields: map[string]string{
				"host":              "example.com",
				"json_payload_vars": `{foo":"bar`,
			},
			errRegex: `^json_payload_vars: "{foo\\\":\\\"bar" <invalid>.*$`,
		},
		{
			name:  "generic/invalid query_vars",
			sType: "generic",
			urlFields: map[string]string{
				"host":       "example.com",
				"query_vars": `{foo:bar}`,
			},
			errRegex: `^query_vars: "{foo:bar}" <invalid>.*$`,
		},
		{
			name:     "lark/invalid",
			sType:    "lark",
			errRegex: `^token: <required>.*$`,
		},
		{
			name:  "mqtt/invalid",
			sType: "mqtt",
			errRegex: test.TrimYAML(`
				^host: <required>.*
				topic: <required>.*$`,
			),
		},
		{
			name:     "pagerduty/invalid",
			sType:    "pagerduty",
			errRegex: `^integrationkey: <required>.*$`,
		},
		{
			name:  "signal/invalid",
			sType: "signal",
			errRegex: test.TrimYAML(`
				^host: <required>.*
				source: <required>.*
				recipients: <required>.*$`,
			),
		},
		{
			name:  "signalgrid/invalid",
			sType: "signalgrid",
			errRegex: test.TrimYAML(`
				^clientkey: <required>
				channel: <required>$`,
			),
		},
		{
			name:  "twilio/invalid",
			sType: "twilio",
			errRegex: test.TrimYAML(`
				^accountsid: <required>.*
				authtoken: <required>
				fromnumber: <required>.*
				tonumbers: <required>.*$`,
			),
		},
		{
			name:     "wecom/invalid",
			sType:    "wecom",
			errRegex: `^key: <required>.*$`,
		},
		{
			name:  "xmpp/invalid",
			sType: "xmpp",
			errRegex: test.TrimYAML(`
				^host: <required>.*
				user: <required>.*
				password: <required>$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := testShoutrrr(false, false)
			input.Type = tc.sType
			input.URLFields = tc.urlFields
			input.Init(
				newTestStatus(tc.name, 1),
				tc.main,
				&Defaults{}, &Defaults{},
			)

			_ = test.AssertCheckValuesWithError(
				t,
				tc.errRegex,
				input.checkValuesURLFields,
			)
		})
	}
}

func TestShoutrrr_CheckValuesParams(t *testing.T) {
	tests := []struct {
		name      string
		sType     string
		params    map[string]string
		urlFields map[string]string
		main      *Defaults
		errRegex  string
	}{
		{
			name:     "no params",
			params:   map[string]string{},
			errRegex: `^$`,
		},
		{
			name: "valid message template",
			params: map[string]string{
				"message": "release! {{ version }}",
			},
			errRegex: `^$`,
		},
		{
			name: "invalid message template",
			params: map[string]string{
				"a": "release! {{ version }",
			},
			errRegex: `^a: "release! {{ version }" <invalid>.*$`,
		},
		{
			name:     "homeassistant/no service",
			sType:    "homeassistant",
			errRegex: `^$`,
		},
		{
			name:  "homeassistant/valid service, bare action",
			sType: "homeassistant",
			params: map[string]string{
				"service": "mobile_app_phone",
			},
			errRegex: `^$`,
		},
		{
			name:  "homeassistant/valid service, domain.action",
			sType: "homeassistant",
			params: map[string]string{
				"service": "notify.mobile_app_phone",
			},
			errRegex: `^$`,
		},
		{
			name:  "homeassistant/invalid service, trailing dot",
			sType: "homeassistant",
			params: map[string]string{
				"service": "notify.",
			},
			errRegex: `^service: "notify." <invalid>.*$`,
		},
		{
			name:  "homeassistant/invalid service, path traversal",
			sType: "homeassistant",
			params: map[string]string{
				"service": "../config",
			},
			errRegex: `^service: "../config" <invalid>.*$`,
		},
		{
			name:  "homeassistant/invalid service, API path",
			sType: "homeassistant",
			params: map[string]string{
				"service": "/api/events/foo",
			},
			errRegex: `^service: "/api/events/foo" <invalid>.*$`,
		},
		{
			name:     "ifttt/no events",
			sType:    "ifttt",
			errRegex: `^events: <required>.*$`,
		},
		{
			name:  "ifttt/valid",
			sType: "ifttt",
			params: map[string]string{
				"events": "events",
			},
			errRegex: `^$`,
		},
		{
			name:  "ifttt/valid with main",
			sType: "ifttt",
			main: NewDefaults(
				"",
				nil,
				nil,
				map[string]string{
					"events": "events",
				},
			),
			errRegex: `^$`,
		},
		{
			name:     "join/no devices",
			sType:    "join",
			params:   map[string]string{},
			errRegex: `^devices: <required>.*$`,
		},
		{
			name:  "join/valid",
			sType: "join",
			params: map[string]string{
				"devices": "foo,bar",
			},
			errRegex: `^$`,
		},
		{
			name:  "join/valid with main",
			sType: "join",
			main: NewDefaults(
				"",
				nil,
				nil,
				map[string]string{
					"devices": "devices",
				},
			),
			errRegex: `^$`,
		},
		{
			name:  "smtp/invalid",
			sType: "smtp",
			errRegex: test.TrimYAML(`
				^fromaddress: <required>.*
				toaddresses: <required>.*$`,
			),
		},
		{
			name:  "smtp/no fromaddress",
			sType: "smtp",
			params: map[string]string{
				"toaddresses": "bosh",
			},
			errRegex: `^fromaddress: <required>.*$`,
		},
		{
			name:  "smtp/no toaddresses",
			sType: "smtp",
			params: map[string]string{
				"fromaddress": "bash",
			},
			errRegex: `^toaddresses: <required>.*$`,
		},
		{
			name:  "smtp/valid timeout",
			sType: "smtp",
			params: map[string]string{
				"fromaddress": "bash",
				"toaddresses": "bosh",
				"timeout":     "0h0m10s",
			},
			errRegex: `^$`,
		},
		{
			name:  "smtp/invalid timeout",
			sType: "smtp",
			params: map[string]string{
				"fromaddress": "bash",
				"toaddresses": "bosh",
				"timeout":     "abc",
			},
			errRegex: `^timeout: "abc" <invalid>.*positive.*duration.*$`,
		},
		{
			name:  "smtp/zero timeout",
			sType: "smtp",
			params: map[string]string{
				"fromaddress": "bash",
				"toaddresses": "bosh",
				"timeout":     "0s",
			},
			errRegex: `^timeout: "0s" <invalid>.*positive.*duration.*$`,
		},
		{
			name:  "smtp/negative timeout",
			sType: "smtp",
			params: map[string]string{
				"fromaddress": "bash",
				"toaddresses": "bosh",
				"timeout":     "-5s",
			},
			errRegex: `^timeout: "-5s" <invalid>.*positive.*duration.*$`,
		},
		{
			name:  "smtp/valid",
			sType: "smtp",
			params: map[string]string{
				"fromaddress": "bash",
				"toaddresses": "bosh",
			},
			errRegex: `^$`,
		},
		{
			name:  "smtp/valid with main",
			sType: "smtp",
			params: map[string]string{
				"fromaddress": "bash",
			},
			main: NewDefaults(
				"",
				nil,
				nil,
				map[string]string{
					"toaddresses": "bosh",
				},
			),
			errRegex: `^$`,
		},
		{
			name:     "teams/no host",
			sType:    "teams",
			params:   map[string]string{},
			errRegex: `^host: <required>.*$`,
		},
		{
			name:  "teams/old-style (no host) - no error",
			sType: "teams",
			urlFields: map[string]string{
				"group":      "GROUP",
				"tenant":     "TENANT",
				"altid":      "ALTID",
				"groupowner": "GROUPOWNER",
				"extraid":    "EXTRAID",
			},
			params:   map[string]string{},
			errRegex: `^$`,
		},
		{
			name:  "teams/valid (Power Automate URL)",
			sType: "teams",
			params: map[string]string{
				"host": "https://prod-00.westus.logic.azure.com:443/workflows/abc",
			},
			errRegex: `^$`,
		},
		{
			name:  "teams/valid with main",
			sType: "teams",
			main: NewDefaults(
				"",
				nil,
				nil,
				map[string]string{
					"host": "https://prod-00.westus.logic.azure.com:443/workflows/abc",
				},
			),
			errRegex: `^$`,
		},
		{
			name:     "telegram/no chats",
			sType:    "telegram",
			errRegex: `^chats: <required>.*$`,
		},
		{
			name:  "telegram/valid",
			sType: "telegram",
			params: map[string]string{
				"chats": "chats",
			},
			errRegex: `^$`,
		},
		{
			name:  "telegram/valid with main",
			sType: "telegram",
			main: NewDefaults(
				"",
				nil,
				nil,
				map[string]string{
					"chats": "chats",
				},
			),
			errRegex: `^$`,
		},
		{
			name:     "xmpp/no to or rooms",
			sType:    "xmpp",
			errRegex: `^to: <required>.*rooms.*$`,
		},
		{
			name:  "xmpp/rooms only",
			sType: "xmpp",
			params: map[string]string{
				"rooms": "room@conf.example.com",
			},
			errRegex: `^$`,
		},
		{
			name:  "zulip/type=invalid",
			sType: "zulip",
			params: map[string]string{
				"type": "bad",
			},
			errRegex: `^type: "bad" <invalid>.*$`,
		},
		{
			name:  "zulip/type=channel",
			sType: "zulip",
			params: map[string]string{
				"type": "channel",
			},
			errRegex: `^$`,
		},
		{
			name:  "zulip/type=direct",
			sType: "zulip",
			params: map[string]string{
				"type": "direct",
			},
			errRegex: `^$`,
		},
		{
			name:  "zulip/type=empty",
			sType: "zulip",
			params: map[string]string{
				"type": "",
			},
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := testShoutrrr(false, false)
			input.Type = tc.sType
			input.Params = tc.params
			input.URLFields = tc.urlFields
			input.Init(
				newTestStatus(tc.name, 1),
				tc.main,
				&Defaults{}, &Defaults{},
			)

			_ = test.AssertCheckValuesWithError(
				t,
				tc.errRegex,
				input.checkValuesParams,
			)
		})
	}
}

func TestBase_CheckValuesParams(t *testing.T) {
	t.Setenv("ARGUS_TEST_NTFY_DISABLETLS", "yes")
	tests := []struct {
		name     string
		itemType string
		params   map[string]string
		errRegex string
	}{
		{
			name:     "no params",
			params:   map[string]string{},
			errRegex: `^$`,
		},
		{
			name: "valid message template",
			params: map[string]string{
				"message": "release! {{ version }}",
			},
			errRegex: `^$`,
		},
		{
			name: "invalid message template",
			params: map[string]string{
				"a": "release! {{ version }",
			},
			errRegex: `^a: "release! {{ version }" <invalid>.*$`,
		},
		{
			name:     "valid select param",
			itemType: "smtp",
			params: map[string]string{
				"auth": "OAuth2",
			},
			errRegex: `^$`,
		},
		{
			name:     "invalid select param",
			itemType: "smtp",
			params: map[string]string{
				"auth": "-",
			},
			errRegex: `^auth: "-" <invalid>.*OAuth2.*$`,
		},
		{
			name:     "smtp/valid timeout",
			itemType: "smtp",
			params: map[string]string{
				"timeout": "0h0m10s",
			},
			errRegex: `^$`,
		},
		{
			name:     "pushover/valid retry and expire",
			itemType: "pushover",
			params: map[string]string{
				"retry":  "30s",
				"expire": "1h",
			},
			errRegex: `^$`,
		},
		{
			name:     "pushover/invalid retry and expire",
			itemType: "pushover",
			params: map[string]string{
				"retry":  "30",
				"expire": "0s",
			},
			errRegex: test.TrimYAML(`
				^retry: "30" <invalid>.*
				expire: "0s" <invalid>.*$`),
		},
		{
			name:     "smtp/invalid timeout",
			itemType: "smtp",
			params: map[string]string{
				"timeout": "abc",
			},
			errRegex: `^timeout: "abc" <invalid>.*positive.*duration.*$`,
		},
		{
			name:     "smtp/zero timeout",
			itemType: "smtp",
			params: map[string]string{
				"timeout": "0s",
			},
			errRegex: `^timeout: "0s" <invalid>.*positive.*duration.*$`,
		},
		{
			name:     "smtp/negative timeout",
			itemType: "smtp",
			params: map[string]string{
				"timeout": "-5s",
			},
			errRegex: `^timeout: "-5s" <invalid>.*positive.*duration.*$`,
		},
		{
			name:     "ntfy/valid disabletls_",
			itemType: "ntfy",
			params: map[string]string{
				"disabletls_": "Yes",
			},
			errRegex: `^$`,
		},
		{
			name:     "ntfy/valid disabletls_ from env var",
			itemType: "ntfy",
			params: map[string]string{
				"disabletls_": "${ARGUS_TEST_NTFY_DISABLETLS}",
			},
			errRegex: `^$`,
		},
		{
			name:     "ntfy/invalid disabletls_",
			itemType: "ntfy",
			params: map[string]string{
				"disabletls_": "maybe",
			},
			errRegex: `^disabletls_: "maybe" <invalid>.*'yes' or 'no'.*$`,
		},
		{
			name:     "ntfy/invalid disabletls_ from unset env var",
			itemType: "ntfy",
			params: map[string]string{
				"disabletls_": "${ARGUS_TEST_NTFY_DISABLETLS_UNSET}",
			},
			errRegex: `^disabletls_: "\$\{ARGUS_TEST_NTFY_DISABLETLS_UNSET\}" <invalid>.*$`,
		},
		{
			name:     "telegram/valid preview",
			itemType: "telegram",
			params: map[string]string{
				"preview": "false",
			},
			errRegex: `^$`,
		},
		{
			name:     "smtp/invalid usestarttls",
			itemType: "smtp",
			params: map[string]string{
				"usestarttls": "on",
			},
			errRegex: `^usestarttls: "on" <invalid>.*'yes' or 'no'.*$`,
		},
		{
			name:     "mqtt/invalid deprecated disabletlsverification",
			itemType: "mqtt",
			params: map[string]string{
				"disabletlsverification": "maybe",
			},
			errRegex: `^disabletlsverification: "maybe" <invalid>.*$`,
		},
		{
			name:     "gotify/invalid skiptlsverify and useheader each reported",
			itemType: "gotify",
			params: map[string]string{
				"skiptlsverify": "maybe",
				"useheader":     "sometimes",
			},
			errRegex: `^skiptlsverify: "maybe" <invalid>.*\nuseheader: "sometimes" <invalid>.*$`,
		},
		{
			name: "non-ntfy/disabletls_ not checked",
			params: map[string]string{
				"disabletls_": "maybe",
			},
			errRegex: `^$`,
		},
		{
			name: "non-smtp/timeout not checked",
			params: map[string]string{
				"timeout": "abc",
			},
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := Base{
				Params: tc.params,
			}

			_ = test.AssertCheckValuesWithError(
				t,
				tc.errRegex,
				func() error {
					return input.checkValuesParams(tc.itemType)
				},
			)
		})
	}
}

func TestBase_CheckValuesParamsSelects(t *testing.T) {
	tests := []struct {
		name      string
		itemType  string
		params    map[string]string
		errRegex  string
		wantParam map[string]string
	}{
		{
			name:     "bark/valid sound normalised",
			itemType: "bark",
			params: map[string]string{
				"sound": "BELL",
			},
			errRegex: `^$`,
			wantParam: map[string]string{
				"sound": "bell",
			},
		},
		{
			name:     "bark/invalid sound",
			itemType: "bark",
			params: map[string]string{
				"sound": "nope",
			},
			errRegex: "^" + regexp.QuoteMeta(
				polymorphic.ErrInvalidValue{
					Key:     "sound",
					Value:   "nope",
					Allowed: barkParamSound,
				}.Error(),
			) + "$",
		},
		{
			name:     "generic/valid method normalised",
			itemType: "generic",
			params: map[string]string{
				"method": "post",
			},
			errRegex: `^$`,
			wantParam: map[string]string{
				"method": "POST",
			},
		},
		{
			name:     "generic/invalid method",
			itemType: "generic",
			params: map[string]string{
				"method": "FETCH",
			},
			errRegex: "^" + regexp.QuoteMeta(
				polymorphic.ErrInvalidValue{
					Key:     "method",
					Value:   "FETCH",
					Allowed: genericParamMethod,
				}.Error(),
			) + "$",
		},
		{
			name:     "ntfy/valid priority",
			itemType: "ntfy",
			params: map[string]string{
				"priority": "DEFAULT",
			},
			errRegex: `^$`,
			wantParam: map[string]string{
				"priority": "default",
			},
		},
		{
			name:     "ntfy/invalid priority",
			itemType: "ntfy",
			params: map[string]string{
				"priority": "urgENT",
			},
			errRegex: "^" + regexp.QuoteMeta(
				polymorphic.ErrInvalidValue{
					Key:     "priority",
					Value:   "urgENT",
					Allowed: ntfyParamPriority,
				}.Error(),
			) + "$",
		},
		{
			name:     "smtp/valid auth and encryption",
			itemType: "smtp",
			params: map[string]string{
				"auth":       "oauth2",
				"encryption": "explicittls",
			},
			errRegex: `^$`,
			wantParam: map[string]string{
				"auth":       "OAuth2",
				"encryption": "ExplicitTLS",
			},
		},
		{
			name:     "smtp/valid auth login",
			itemType: "smtp",
			params: map[string]string{
				"auth": "login",
			},
			errRegex: `^$`,
			wantParam: map[string]string{
				"auth": "Login",
			},
		},
		{
			name:     "smtp/invalid auth and encryption aggregated",
			itemType: "smtp",
			params: map[string]string{
				"auth":       "basic",
				"encryption": "tls1.3",
			},
			errRegex: "^" + regexp.QuoteMeta(
				errfmt.FormatError(
					errors.Join(
						&polymorphic.ErrInvalidValue{
							Key:     "auth",
							Value:   "basic",
							Allowed: smtpParamAuth,
						},
						&polymorphic.ErrInvalidValue{
							Key:     "encryption",
							Value:   "tls1.3",
							Allowed: smtpParamEncryption,
						},
					),
				),
			) + "$",
		},
		{
			name:     "telegram/valid parsemode",
			itemType: "telegram",
			params: map[string]string{
				"parsemode": "markdownv2",
			},
			errRegex: `^$`,
			wantParam: map[string]string{
				"parsemode": "MarkdownV2",
			},
		},
		{
			name:     "telegram/invalid parsemode",
			itemType: "telegram",
			params: map[string]string{
				"parsemode": "mdx",
			},
			errRegex: "^" + regexp.QuoteMeta(
				polymorphic.ErrInvalidValue{
					Key:     "parsemode",
					Value:   "mdx",
					Allowed: telegramParamParsemode,
				}.Error(),
			) + "$",
		},
		{
			name:     "unknown/type is no-op even with values",
			itemType: "unknown",
			params: map[string]string{
				"scheme": "ftp",
			},
			errRegex: `^$`,
			wantParam: map[string]string{
				"scheme": "ftp",
			},
		},
		{
			name:     "nil/empty params produces no error",
			itemType: "smtp",
			params:   map[string]string{},
			errRegex: `^$`,
		},
		{
			name:     "mqtt/valid qos normalised",
			itemType: "mqtt",
			params: map[string]string{
				"qos": "exactlyonce",
			},
			errRegex: `^$`,
			wantParam: map[string]string{
				"qos": "ExactlyOnce",
			},
		},
		{
			name:     "mqtt/invalid qos",
			itemType: "mqtt",
			params: map[string]string{
				"qos": "Sometimes",
			},
			errRegex: "^" + regexp.QuoteMeta(
				polymorphic.ErrInvalidValue{
					Key:     "qos",
					Value:   "Sometimes",
					Allowed: mqttParamQoS,
				}.Error(),
			) + "$",
		},
		{
			name:     "pagerduty/valid action and severity normalised",
			itemType: "pagerduty",
			params: map[string]string{
				"action":   "RESOLVE",
				"severity": "Warning",
			},
			errRegex: `^$`,
			wantParam: map[string]string{
				"action":   "resolve",
				"severity": "warning",
			},
		},
		{
			name:     "pagerduty/invalid action",
			itemType: "pagerduty",
			params: map[string]string{
				"action": "snooze",
			},
			errRegex: "^" + regexp.QuoteMeta(
				polymorphic.ErrInvalidValue{
					Key:     "action",
					Value:   "snooze",
					Allowed: pagerDutyParamAction,
				}.Error(),
			) + "$",
		},
		{
			name:     "pagerduty/invalid severity",
			itemType: "pagerduty",
			params: map[string]string{
				"severity": "fatal",
			},
			errRegex: "^" + regexp.QuoteMeta(
				polymorphic.ErrInvalidValue{
					Key:     "severity",
					Value:   "fatal",
					Allowed: pagerDutyParamSeverity,
				}.Error(),
			) + "$",
		},
		{
			name:     "signal/valid textmode normalised",
			itemType: "signal",
			params: map[string]string{
				"textmode": "styled",
			},
			errRegex: `^$`,
			wantParam: map[string]string{
				"textmode": "Styled",
			},
		},
		{
			name:     "signal/invalid textmode",
			itemType: "signal",
			params: map[string]string{
				"textmode": "Bold",
			},
			errRegex: "^" + regexp.QuoteMeta(
				polymorphic.ErrInvalidValue{
					Key:     "textmode",
					Value:   "Bold",
					Allowed: signalParamTextMode,
				}.Error(),
			) + "$",
		},
		{
			name:     "signalgrid/invalid type",
			itemType: "signalgrid",
			params: map[string]string{
				"type": "PANIC",
			},
			errRegex: "^" + regexp.QuoteMeta(
				polymorphic.ErrInvalidValue{
					Key:     "type",
					Value:   "PANIC",
					Allowed: signalgridParamType,
				}.Error(),
			) + "$",
		},
		{
			name:     "signalgrid/valid type normalised",
			itemType: "signalgrid",
			params: map[string]string{
				"type": "crit",
			},
			errRegex: `^$`,
			wantParam: map[string]string{
				"type": "CRIT",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := Base{
				Params: tc.params,
			}

			_ = test.AssertCheckValuesWithError(
				t,
				tc.errRegex,
				func() error {
					return input.checkValuesParams(tc.itemType)
				},
			)

			for k, v := range tc.wantParam {
				if got := input.GetParam(k); got != v {
					t.Fatalf("Base.GetParam(%q) mismatch\ngot:  %q\nwant: %q", k, got, v)
				}
			}
		})
	}
}

func TestBase_ValidateParamSelect(t *testing.T) {
	key := "test"

	tests := []struct {
		name     string
		value    string
		allowed  []string
		errRegex string
		want     string
	}{
		{
			name:     "empty value returns nil",
			value:    "",
			allowed:  []string{"low", "default", "high"},
			errRegex: `^$`,
			want:     "",
		},
		{
			name:     "valid value normalises and returns nil",
			value:    "HiGh",
			allowed:  []string{"min", "low", "default", "high", "max"},
			errRegex: `^$`,
			want:     "high",
		},
		{
			name:     "invalid value returns error and leaves unchanged",
			value:    "nope",
			allowed:  []string{"None", "Unknown", "Plain", "CRAMMD5", "OAuth2"},
			errRegex: "^" + key + `: "nope" <invalid> .*'CRAMMD5'.*$`,
			want:     "nope",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := Base{}
			input.InitMaps()
			input.setParam(key, tc.value)

			err := input.validateParamSelect(key, tc.allowed)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if got := input.GetParam(key); got != tc.want {
				t.Fatalf("Base.GetParam() mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestParseHostPort(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantHost string
		wantPort string
	}{
		{
			name:     "empty",
			input:    "",
			wantHost: "",
			wantPort: "",
		},
		{
			name:     "bare hostname",
			input:    "example.com",
			wantHost: "example.com",
			wantPort: "",
		},
		{
			name:     "hostname with port, no scheme",
			input:    "example.com:8080",
			wantHost: "example.com",
			wantPort: "8080",
		},
		{
			name:     "scheme and hostname",
			input:    "https://example.com",
			wantHost: "example.com",
			wantPort: "",
		},
		{
			name:     "scheme, hostname, and port",
			input:    "https://example.com:8443",
			wantHost: "example.com",
			wantPort: "8443",
		},
		{
			name:     "scheme, hostname, port, and path",
			input:    "https://example.com:8443/path",
			wantHost: "example.com",
			wantPort: "8443",
		},
		{
			name:     "userinfo is stripped",
			input:    "user:pass@example.com",
			wantHost: "example.com",
			wantPort: "",
		},
		{
			name:     "IPv4 address/bare",
			input:    "127.0.0.1",
			wantHost: "127.0.0.1",
			wantPort: "",
		},
		{
			name:     "IPv4 address/scheme, hostname, and port",
			input:    "https://127.0.0.1:9000",
			wantHost: "127.0.0.1",
			wantPort: "9000",
		},
		{
			name:     "IPv6 address/bare",
			input:    "https://[::1]",
			wantHost: "::1",
			wantPort: "",
		},
		{
			name:     "IPv6 address/scheme, hostname, and port",
			input:    "https://[::1]:8080",
			wantHost: "::1",
			wantPort: "8080",
		},
		{
			name:     "port only",
			input:    ":8080",
			wantHost: "",
			wantPort: "8080",
		},
		{
			name:     "trailing colon, no port",
			input:    "example.com:",
			wantHost: "example.com",
			wantPort: "",
		},
		{
			name:     "scheme only",
			input:    "https://",
			wantHost: "",
			wantPort: "",
		},
		{
			name:     "case is preserved",
			input:    "EXAMPLE.COM",
			wantHost: "EXAMPLE.COM",
			wantPort: "",
		},
		{
			name:     "invalid/control character returns input unchanged",
			input:    "https://\texample.com",
			wantHost: "https://\texample.com",
			wantPort: "",
		},
		{
			name:     "invalid/non-numeric port returns input unchanged",
			input:    "example.com:notaport",
			wantHost: "example.com:notaport",
			wantPort: "",
		},
		{
			name:     "leading trailing spaces trimmed/hostname",
			input:    "  example.com  ",
			wantHost: "example.com",
			wantPort: "",
		},
		{
			name:     "leading trailing spaces trimmed/scheme, hostname, and port",
			input:    "  https://example.com:123  ",
			wantHost: "example.com",
			wantPort: "123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host, port := parseHostPort(tc.input)

			if host != tc.wantHost || port != tc.wantPort {
				t.Errorf(
					"mismatch\ngot:  host=%q, port=%q\nwant: host=%q, port=%q",
					host, port,
					tc.wantHost, tc.wantPort,
				)
			}
		})
	}
}
