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
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

func TestSliceDefaults_Print(t *testing.T) {
	// GIVEN a SliceDefaults
	testValid := testShoutrrrDefaults(false, false)
	testInvalid := testShoutrrrDefaults(true, true)
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
			want:  "",
		},
		"single empty element slice": {
			slice: &SliceDefaults{
				"single": {}},
			want: `
notify:
  single: {}`,
		},
		"single element slice": {
			slice: &SliceDefaults{
				"single": testValid},
			want: `
notify:
  single:
    type: gotify
    options:
      max_tries: "` + testValid.GetSelfOption("max_tries") + `"
    url_fields:
      host: ` + testValid.GetSelfURLField("host") + `
      path: ` + testValid.GetSelfURLField("path") + `
      token: ` + testValid.GetSelfURLField("token"),
		},
		"multiple element slice": {
			slice: &SliceDefaults{
				"first":  testValid,
				"second": testInvalid},
			want: `
notify:
  first:
    type: gotify
    options:
      max_tries: "` + testValid.GetSelfOption("max_tries") + `"
    url_fields:
      host: ` + testValid.GetSelfURLField("host") + `
      path: ` + testValid.GetSelfURLField("path") + `
      token: ` + testValid.GetSelfURLField("token") + `
  second:
    type: gotify
    options:
      max_tries: "` + testInvalid.GetSelfOption("max_tries") + `"
    url_fields:
      host: ` + testInvalid.GetSelfURLField("host") + `
      path: ` + testInvalid.GetSelfURLField("path") + `
      token: ` + testInvalid.GetSelfURLField("token"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			if tc.want != "" {
				tc.want += "\n"
			}
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN Print is called
			tc.slice.Print("")

			// THEN it prints the expected output
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = stdout
			strOut := string(out)
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if strOut != tc.want {
				t.Errorf("Print should have given\n%q\nbut gave\n%q",
					tc.want, strOut)
			}
		})
	}
}

func TestShoutrrr_checkValuesMaster(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		sType              *string
		options            map[string]string
		urlFields          map[string]string
		params             map[string]string
		main               *ShoutrrrDefaults
		errsRegex          string
		errsOptionsRegex   string
		errsURLFieldsRegex string
		errsParamsRegex    string
	}{
		"no type": {
			errsRegex: "type: <required>",
			sType:     stringPtr(""),
		},
		"invalid type": {
			errsRegex: "type: .* <invalid>",
			sType:     stringPtr("argus"),
		},
		"invalid type - type in main differs": {
			errsRegex:          `type: "gotify" != "discord" <invalid>`,
			errsURLFieldsRegex: `host: <required>.*token: <required>`,
			sType:              stringPtr("gotify"),
			main:               NewDefaults("discord", nil, nil, nil),
		},
		"bark - invalid": {
			sType:              stringPtr("bark"),
			errsURLFieldsRegex: `^  devicekey: <required>[^:]+host: <required>[^:]+$`,
		},
		"bark - no devicekey": {
			sType:              stringPtr("bark"),
			errsURLFieldsRegex: `^  devicekey: <required>[^:]+$`,
			urlFields: map[string]string{
				"host": "https://example.com"},
		},
		"bark - no host": {
			sType:              stringPtr("bark"),
			errsURLFieldsRegex: `^  host: <required>[^:]+$`,
			urlFields: map[string]string{
				"devicekey": "foo"},
		},
		"bark - valid": {
			sType: stringPtr("bark"),
			urlFields: map[string]string{
				"devicekey": "foo",
				"host":      "https://example.com"},
		},
		"discord - invalid": {
			sType:              stringPtr("discord"),
			errsURLFieldsRegex: "token: <required>.*webhookid: <required>",
		},
		"discord - no token": {
			sType:              stringPtr("discord"),
			errsURLFieldsRegex: "token: <required>",
			urlFields: map[string]string{
				"webhookid": "bash"},
		},
		"discord - no webhookid": {
			sType:              stringPtr("discord"),
			errsURLFieldsRegex: "webhookid: <required>",
			urlFields: map[string]string{
				"token": "bish"},
		},
		"discord - valid": {
			sType: stringPtr("discord"),
			urlFields: map[string]string{
				"token":     "bish",
				"webhookid": "webhookid"},
		},
		"discord - valid with main": {
			sType: stringPtr("discord"),
			main: NewDefaults(
				"", nil, nil,
				&map[string]string{
					"token":     "bish",
					"webhookid": "bash"}),
		},
		"smtp - invalid": {
			sType:              stringPtr("smtp"),
			errsURLFieldsRegex: "host: <required>.*",
			errsParamsRegex:    "fromaddress: <required>.*toaddresses: <required>",
		},
		"smtp - no host": {
			sType:              stringPtr("smtp"),
			errsURLFieldsRegex: "host: <required>",
			params: map[string]string{
				"fromaddress": "bash",
				"toaddresses": "bosh"},
		},
		"smtp - no fromaddress": {
			sType:           stringPtr("smtp"),
			errsParamsRegex: "fromaddress: <required>",
			urlFields: map[string]string{
				"host": "bish"},
			params: map[string]string{
				"toaddresses": "bosh"},
		},
		"smtp - no toaddresses": {
			sType:           stringPtr("smtp"),
			errsParamsRegex: "toaddresses: <required>",
			urlFields: map[string]string{
				"host": "bish"},
			params: map[string]string{
				"fromaddress": "bash"},
		},
		"smtp - valid": {
			sType: stringPtr("smtp"),
			urlFields: map[string]string{
				"host": "bish"},
			params: map[string]string{
				"fromaddress": "bash",
				"toaddresses": "bosh"},
		},
		"smtp - valid with main": {
			sType: stringPtr("smtp"),
			main: NewDefaults(
				"", nil,
				&map[string]string{
					"fromaddress": "bash",
					"toaddresses": "bosh"},
				&map[string]string{
					"host": "bish"}),
		},
		"gotify - invalid": {
			sType:              stringPtr("gotify"),
			errsURLFieldsRegex: "host: <required>.*token: <required>",
		},
		"gotify - no host": {
			sType:              stringPtr("gotify"),
			errsURLFieldsRegex: "host: <required>",
			urlFields: map[string]string{
				"token": "bash"},
		},
		"gotify - no token": {
			sType:              stringPtr("gotify"),
			errsURLFieldsRegex: "token: <required>",
			urlFields: map[string]string{
				"host": "bish"},
		},
		"gotify - valid": {
			sType: stringPtr("gotify"),
			urlFields: map[string]string{
				"host":  "bish",
				"token": "bash"},
		},
		"gotify - valid with main": {
			sType: stringPtr("gotify"),
			main: NewDefaults(
				"", nil, nil,
				&map[string]string{
					"host":  "bish",
					"token": "bash"}),
		},
		"googlechat - invalid": {
			sType:              stringPtr("googlechat"),
			errsURLFieldsRegex: "raw: <required>",
		},
		"googlechat - valid": {
			sType: stringPtr("googlechat"),
			urlFields: map[string]string{
				"raw": "bish"},
		},
		"googlechat - valid with main": {
			sType: stringPtr("googlechat"),
			main: NewDefaults(
				"", nil, nil,
				&map[string]string{
					"raw": "bish"}),
		},
		"ifttt - invalid": {
			sType:              stringPtr("ifttt"),
			errsURLFieldsRegex: "webhookid: <required>", errsParamsRegex: "events: <required>",
		},
		"ifttt - no webhookid": {
			sType:              stringPtr("ifttt"),
			errsURLFieldsRegex: "webhookid: <required>",
			urlFields:          map[string]string{},
			params: map[string]string{
				"events": "bash"},
		},
		"ifttt - no events": {
			sType:           stringPtr("ifttt"),
			errsParamsRegex: "events: <required>",
			urlFields: map[string]string{
				"webhookid": "bish"},
		},
		"ifttt - valid": {
			sType: stringPtr("ifttt"),
			urlFields: map[string]string{
				"webhookid": "bish"},
			params: map[string]string{
				"events": "events"},
		},
		"ifttt - valid with main": {
			sType: stringPtr("ifttt"),
			main: NewDefaults(
				"", nil,
				&map[string]string{
					"events": "events"},
				&map[string]string{
					"webhookid": "webhookid"}),
		},
		"join - invalid": {
			sType:              stringPtr("join"),
			errsURLFieldsRegex: "apikey: <required>", errsParamsRegex: "devices: <required>",
		},
		"join - no apikey": {
			sType:              stringPtr("join"),
			errsURLFieldsRegex: "apikey: <required>",
			urlFields:          map[string]string{},
			params: map[string]string{
				"devices": "bash"},
		},
		"join - no devices": {
			sType:           stringPtr("join"),
			errsParamsRegex: "devices: <required>",
			urlFields: map[string]string{
				"apikey": "bish"},
		},
		"join - valid": {
			sType: stringPtr("join"),
			urlFields: map[string]string{
				"apikey": "bish"},
			params: map[string]string{
				"devices": "devices"},
		},
		"join - valid with main": {
			sType: stringPtr("join"),
			main: NewDefaults(
				"", nil,
				&map[string]string{
					"devices": "devices"},
				&map[string]string{
					"apikey": "apikey"}),
		},
		"mattermost - invalid": {
			sType:              stringPtr("mattermost"),
			errsURLFieldsRegex: "host: <required>.*token: <required>",
		},
		"mattermost - no host": {
			sType:              stringPtr("mattermost"),
			errsURLFieldsRegex: "host: <required>",
			urlFields: map[string]string{
				"token": "bash"},
		},
		"mattermost - no token": {
			sType:              stringPtr("mattermost"),
			errsURLFieldsRegex: "token: <required>",
			urlFields: map[string]string{
				"host": "bish"},
		},
		"mattermost - valid": {
			sType: stringPtr("mattermost"),
			urlFields: map[string]string{
				"host":  "bish",
				"token": "bash"},
		},
		"mattermost - valid with main": {
			sType: stringPtr("mattermost"),
			main: NewDefaults(
				"", nil, nil,
				&map[string]string{
					"host":  "bish",
					"token": "bash"}),
		},
		"matrix - invalid": {
			sType:              stringPtr("matrix"),
			errsURLFieldsRegex: "host: <required>.*password: <required>",
		},
		"matrix - no host": {
			sType:              stringPtr("matrix"),
			errsURLFieldsRegex: "host: <required>",
			urlFields: map[string]string{
				"password": "bash"},
		},
		"matrix - no password": {
			sType:              stringPtr("matrix"),
			errsURLFieldsRegex: "password: <required>",
			urlFields: map[string]string{
				"host": "bish"},
		},
		"matrix - valid": {
			sType: stringPtr("matrix"),
			urlFields: map[string]string{
				"host":     "bish",
				"password": "password"},
		},
		"matrix - valid with main": {
			sType: stringPtr("matrix"),
			main: NewDefaults(
				"", nil, nil,
				&map[string]string{
					"host":     "bish",
					"password": "bash"}),
		},
		"ntfy - invalid": {
			sType:              stringPtr("ntfy"),
			errsURLFieldsRegex: "topic: <required>",
		},
		"ntfy - valid": {
			sType: stringPtr("ntfy"),
			urlFields: map[string]string{
				"topic": "foo"},
		},
		"opsgenie - invalid": {
			sType:              stringPtr("opsgenie"),
			errsURLFieldsRegex: "apikey: <required>",
		},
		"opsgenie - valid": {
			sType: stringPtr("opsgenie"),
			urlFields: map[string]string{
				"apikey": "apikey"},
		},
		"opsgenie - valid with main": {
			sType: stringPtr("opsgenie"),
			main: NewDefaults(
				"", nil, nil,
				&map[string]string{
					"apikey": "apikey"}),
		},
		"pushbullet - invalid": {
			sType:              stringPtr("pushbullet"),
			errsURLFieldsRegex: "token: <required>.*targets: <required>",
		},
		"pushbullet - no token": {
			sType:              stringPtr("pushbullet"),
			errsURLFieldsRegex: "token: <required>",
			urlFields: map[string]string{
				"targets": "bash"},
		},
		"pushbullet - no targets": {
			sType:              stringPtr("pushbullet"),
			errsURLFieldsRegex: "targets: <required>",
			urlFields: map[string]string{
				"token": "bish"},
		},
		"pushbullet - valid": {
			sType: stringPtr("pushbullet"),
			urlFields: map[string]string{
				"token":   "bish",
				"targets": "targets"},
		},
		"pushbullet - valid with main": {
			sType: stringPtr("pushbullet"),
			main: NewDefaults(
				"", nil, nil,
				&map[string]string{
					"token":   "bish",
					"targets": "bash"}),
		},
		"pushover - invalid": {
			sType:              stringPtr("pushover"),
			errsURLFieldsRegex: "token: <required>.*user: <required>",
		},
		"pushover - no token": {
			sType:              stringPtr("pushover"),
			errsURLFieldsRegex: "token: <required>",
			urlFields: map[string]string{
				"user": "bash"},
		},
		"pushover - no user": {
			sType:              stringPtr("pushover"),
			errsURLFieldsRegex: "user: <required>",
			urlFields: map[string]string{
				"token": "bish"},
		},
		"pushover - valid": {
			sType: stringPtr("pushover"),
			urlFields: map[string]string{
				"token": "bish",
				"user":  "user"},
		},
		"pushover - valid with main": {
			sType: stringPtr("pushover"),
			main: NewDefaults(
				"", nil, nil,
				&map[string]string{
					"token": "bish",
					"user":  "bash"}),
		},
		"rocketchat - invalid": {
			sType:              stringPtr("rocketchat"),
			errsURLFieldsRegex: "host: <required>.*tokena: <required>.*tokenb: <required>.*channel: <required>",
		},
		"rocketchat - no host": {
			sType:              stringPtr("rocketchat"),
			errsURLFieldsRegex: "host: <required>",
			urlFields: map[string]string{
				"tokena":  "bash",
				"tokenb":  "bosh",
				"channel": "bing"},
		},
		"rocketchat - no tokena": {
			sType:              stringPtr("rocketchat"),
			errsURLFieldsRegex: "tokena: <required>",
			urlFields: map[string]string{
				"host":    "bish",
				"tokenb":  "bash",
				"channel": "bing"},
		},
		"rocketchat - no tokenb": {
			sType:              stringPtr("rocketchat"),
			errsURLFieldsRegex: "tokenb: <required>",
			urlFields: map[string]string{
				"host":    "bish",
				"tokena":  "bash",
				"channel": "bing"},
		},
		"rocketchat - no channel": {
			sType:              stringPtr("rocketchat"),
			errsURLFieldsRegex: "channel: <required>",
			urlFields: map[string]string{
				"host":   "bish",
				"tokena": "bash",
				"tokenb": "bosh"},
		},
		"rocketchat - valid": {
			sType: stringPtr("rocketchat"),
			urlFields: map[string]string{
				"host":    "bish",
				"tokena":  "bash",
				"tokenb":  "bosh",
				"channel": "bing"},
		},
		"rocketchat - valid with main": {
			sType: stringPtr("rocketchat"),
			main: NewDefaults(
				"", nil, nil,
				&map[string]string{
					"host":    "bish",
					"tokena":  "bash",
					"tokenb":  "bosh",
					"channel": "bing"}),
		},
		"slack - invalid": {
			sType:              stringPtr("slack"),
			errsURLFieldsRegex: "token: <required>.*channel: <required>",
		},
		"slack - no token": {
			sType:              stringPtr("slack"),
			errsURLFieldsRegex: "token: <required>",
			urlFields: map[string]string{
				"channel": "bash"},
		},
		"slack - no channel": {
			sType:              stringPtr("slack"),
			errsURLFieldsRegex: "channel: <required>",
			urlFields: map[string]string{
				"token": "bish"},
		},
		"slack - valid": {
			sType: stringPtr("slack"),
			urlFields: map[string]string{
				"token":   "bish",
				"channel": "channel"},
		},
		"slack - valid with main": {
			sType: stringPtr("slack"),
			main: NewDefaults(
				"", nil, nil,
				&map[string]string{
					"token":   "bish",
					"channel": "bash"}),
		},
		"teams - invalid": {
			sType:              stringPtr("teams"),
			errsURLFieldsRegex: "group: <required>.*tenant: <required>.*altid: <required>.*groupowner: <required>", errsParamsRegex: "host: <required>",
		},
		"teams - no group": {
			sType:              stringPtr("teams"),
			errsURLFieldsRegex: "group: <required>",
			urlFields: map[string]string{
				"tenant":     "bash",
				"altid":      "bosh",
				"groupowner": "bing"},
			params: map[string]string{
				"host": "https://release-argus.io"},
		},
		"teams - no tenant": {
			sType:              stringPtr("teams"),
			errsURLFieldsRegex: "tenant: <required>",
			urlFields: map[string]string{
				"group":      "bish",
				"altid":      "bash",
				"groupowner": "bing"},
			params: map[string]string{
				"host": "https://release-argus.io"},
		},
		"teams - no altid": {
			sType:              stringPtr("teams"),
			errsURLFieldsRegex: "altid: <required>",
			urlFields: map[string]string{
				"group":      "bish",
				"tenant":     "bash",
				"groupowner": "bing"},
			params: map[string]string{
				"host": "https://release-argus.io"},
		},
		"teams - no groupowner": {
			sType:              stringPtr("teams"),
			errsURLFieldsRegex: "groupowner: <required>",
			urlFields: map[string]string{
				"group":  "bish",
				"tenant": "bash",
				"altid":  "bosh"},
			params: map[string]string{
				"host": "https://release-argus.io"},
		},
		"teams - no host": {
			sType:           stringPtr("teams"),
			errsParamsRegex: "host: <required>",
			urlFields: map[string]string{
				"group":      "bish",
				"tenant":     "bash",
				"altid":      "bosh",
				"groupowner": "bing"},
		},
		"teams - valid": {
			sType: stringPtr("teams"),
			urlFields: map[string]string{
				"group":      "bish",
				"tenant":     "bash",
				"altid":      "bosh",
				"groupowner": "bing"},
			params: map[string]string{
				"host": "https://release-argus.io"},
		},
		"teams - valid with main": {
			sType: stringPtr("teams"),
			main: NewDefaults(
				"", nil,
				&map[string]string{
					"host": "https://release-argus.io"},
				&map[string]string{
					"group":      "bish",
					"tenant":     "bash",
					"altid":      "bosh",
					"groupowner": "bing"}),
		},
		"telegram - invalid": {
			sType:              stringPtr("telegram"),
			errsURLFieldsRegex: "token: <required>",
			errsParamsRegex:    "chats: <required>",
		},
		"telegram - no token": {
			sType:              stringPtr("telegram"),
			errsURLFieldsRegex: "token: <required>",
			urlFields:          map[string]string{},
			params: map[string]string{
				"chats": "bash"},
		},
		"telegram - no chats": {
			sType:           stringPtr("telegram"),
			errsParamsRegex: "chats: <required>",
			urlFields: map[string]string{
				"token": "bish"},
		},
		"telegram - valid": {
			sType: stringPtr("telegram"),
			urlFields: map[string]string{
				"token": "bish"},
			params: map[string]string{
				"chats": "chats"},
		},
		"telegram - valid with main": {
			sType: stringPtr("telegram"),
			main: NewDefaults(
				"", nil,
				&map[string]string{
					"chats": "chats"},
				&map[string]string{
					"token": "bish"}),
		},
		"zulip - invalid": {
			sType:              stringPtr("zulip"),
			errsURLFieldsRegex: "host: <required>.*botmail: <required>.*botkey: <required>",
		},
		"zulip - no host": {
			sType:              stringPtr("zulip"),
			errsURLFieldsRegex: "host: <required>",
			urlFields: map[string]string{
				"botmail": "bash",
				"botkey":  "bosh"},
		},
		"zulip - no botmail": {
			sType:              stringPtr("zulip"),
			errsURLFieldsRegex: "botmail: <required>",
			urlFields: map[string]string{
				"host":   "bish",
				"botkey": "bash"},
		},
		"zulip - no botkey": {
			sType:              stringPtr("zulip"),
			errsURLFieldsRegex: "botkey: <required>",
			urlFields: map[string]string{
				"host":    "bish",
				"botmail": "bash"},
		},
		"zulip - valid": {
			sType: stringPtr("zulip"),
			urlFields: map[string]string{
				"host":    "bish",
				"botmail": "bash",
				"botkey":  "bosh"},
		},
		"zulip - valid with main": {
			sType: stringPtr("zulip"),
			main: NewDefaults(
				"", nil, nil,
				&map[string]string{
					"host":    "bish",
					"botmail": "bash",
					"botkey":  "bosh"}),
		},
		"shoutrrr - invalid": {
			sType:              stringPtr("shoutrrr"),
			errsURLFieldsRegex: "raw: <required>",
		},
		"shoutrrr - valid": {
			sType: stringPtr("shoutrrr"),
			urlFields: map[string]string{
				"raw": "bish"},
		},
		"shoutrrr - valid with main": {
			sType: stringPtr("shoutrrr"),
			main: NewDefaults(
				"", nil, nil,
				&map[string]string{
					"raw": "bish"}),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			if tc.sType != nil {
				shoutrrr.Type = *tc.sType
			}
			shoutrrr.Options = tc.options
			shoutrrr.URLFields = tc.urlFields
			shoutrrr.Params = tc.params
			svcStatus := svcstatus.New(
				nil, nil, nil,
				"", "", "", "", "", "")
			svcStatus.Init(
				1, 0, 0,
				stringPtr("serviceID"), nil)
			shoutrrr.Init(
				svcStatus,
				tc.main,
				&ShoutrrrDefaults{}, &ShoutrrrDefaults{})

			// WHEN checkValuesMaster is called
			var (
				errs          error
				errsOptions   error
				errsURLFields error
				errsParams    error
			)
			shoutrrr.checkValuesMaster("", &errs, &errsOptions, &errsURLFields, &errsParams)

			// THEN it err's when expected
			// errs
			if tc.errsRegex == "" {
				tc.errsRegex = "^$"
			}
			e := util.ErrorToString(errs)
			re := regexp.MustCompile(tc.errsRegex)
			match := re.MatchString(e)
			if !match {
				t.Errorf("want match for %q\nnot: %q",
					tc.errsRegex, e)
			}
			// errsOptions
			if tc.errsOptionsRegex == "" {
				tc.errsOptionsRegex = "^$"
			}
			e = util.ErrorToString(errsOptions)
			re = regexp.MustCompile(tc.errsOptionsRegex)
			match = re.MatchString(e)
			if !match {
				t.Errorf("want match for %q\nnot: %q",
					tc.errsOptionsRegex, e)
			}
			// errsURLFields
			if tc.errsURLFieldsRegex == "" {
				tc.errsURLFieldsRegex = "^$"
			}
			e = util.ErrorToString(errsURLFields)
			re = regexp.MustCompile(tc.errsURLFieldsRegex)
			match = re.MatchString(e)
			if !match {
				t.Errorf("want match for %q\nnot: %q",
					tc.errsURLFieldsRegex, e)
			}
			// errsParams
			if tc.errsParamsRegex == "" {
				tc.errsParamsRegex = "^$"
			}
			e = util.ErrorToString(errsParams)
			re = regexp.MustCompile(tc.errsParamsRegex)
			match = re.MatchString(e)
			if !match {
				t.Errorf("want match for %q\nnot: %q",
					tc.errsParamsRegex, e)
			}
		})
	}
}

func TestShoutrrr_CorrectSelf(t *testing.T) {
	// GIVEN a Service
	tests := map[string]struct {
		sType     string
		mapTarget string
		startAs   map[string]string
		want      map[string]string
	}{
		"port - leading colon": {
			mapTarget: "url_fields",
			startAs:   map[string]string{"port": ":8080"},
			want:      map[string]string{"port": "8080"},
		},
		"port - valid": {
			mapTarget: "url_fields",
			startAs:   map[string]string{"port": "8080"},
			want:      map[string]string{"port": "8080"},
		},
		"path - leading slash": {
			mapTarget: "url_fields",
			startAs:   map[string]string{"path": "/argus"},
			want:      map[string]string{"path": "argus"},
		},
		"path - valid": {
			mapTarget: "url_fields",
			startAs:   map[string]string{"path": "argus"},
			want:      map[string]string{"path": "argus"},
		},
		"port - from url": {
			mapTarget: "url_fields",
			startAs:   map[string]string{"host": "https://mattermost.example.com:8443", "port": ""},
			want:      map[string]string{"host": "mattermost.example.com", "port": "8443"},
		},
		"matrix - rooms, leading #": {
			sType:     "matrix",
			mapTarget: "params",
			startAs:   map[string]string{"rooms": "#alias:server"},
			want:      map[string]string{"rooms": "alias:server"},
		},
		"matrix - rooms, leading # already urlEncoded": {
			sType:     "matrix",
			mapTarget: "params",
			startAs:   map[string]string{"rooms": "%23alias:server"},
			want:      map[string]string{"rooms": "%23alias:server"},
		},
		"matrix - rooms, valid": {
			sType:     "matrix",
			mapTarget: "params",
			startAs:   map[string]string{"rooms": "alias:server"},
			want:      map[string]string{"rooms": "alias:server"},
		},
		"mattermost - channel, leading slash": {
			sType:     "mattermost",
			mapTarget: "url_fields",
			startAs:   map[string]string{"channel": "/argus"},
			want:      map[string]string{"channel": "argus"},
		},
		"mattermost - channel, valid": {
			sType:     "mattermost",
			mapTarget: "url_fields",
			startAs:   map[string]string{"channel": "argus"},
			want:      map[string]string{"channel": "argus"},
		},
		"slack - color, not urlEncoded": {
			sType:     "slack",
			mapTarget: "params",
			startAs:   map[string]string{"color": "#ffffff"},
			want:      map[string]string{"color": "%23ffffff"},
		},
		"slack - color, valid": {
			sType:     "slack",
			mapTarget: "params",
			startAs:   map[string]string{"color": "%23ffffff"},
			want:      map[string]string{"color": "%23ffffff"},
		},
		"teams - altid, leading slash": {
			sType:     "teams",
			mapTarget: "url_fields",
			startAs:   map[string]string{"altid": "/argus"},
			want:      map[string]string{"altid": "argus"},
		},
		"teams - altid, valid": {
			sType:     "teams",
			mapTarget: "url_fields",
			startAs:   map[string]string{"altid": "argus"},
			want:      map[string]string{"altid": "argus"},
		},
		"teams - groupowner, leading slash": {
			sType:     "teams",
			mapTarget: "url_fields",
			startAs:   map[string]string{"groupowner": "/argus"},
			want:      map[string]string{"groupowner": "argus"},
		},
		"teams - groupowner, valid": {
			sType:     "teams",
			mapTarget: "url_fields",
			startAs:   map[string]string{"groupowner": "argus"},
			want:      map[string]string{"groupowner": "argus"},
		},
		"zulip - botmail, not urlEncoded": {
			sType:     "zulip",
			mapTarget: "url_fields",
			startAs:   map[string]string{"botmail": "foo@bar.com"},
			want:      map[string]string{"botmail": "foo%40bar.com"},
		},
		"zulip - botmail, valid": {
			sType:     "zulip",
			mapTarget: "url_fields",
			startAs:   map[string]string{"botmail": "foo%40bar.com"},
			want:      map[string]string{"botmail": "foo%40bar.com"},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := New(
				nil, "", nil, nil,
				tc.sType,
				nil, nil, nil, nil)
			shoutrrr.InitMaps()
			if tc.mapTarget == "url_fields" {
				for k, v := range tc.startAs {
					shoutrrr.SetURLField(k, v)
				}
			} else {
				for k, v := range tc.startAs {
					shoutrrr.SetParam(k, v)
				}
			}

			// WHEN correctSelf is called
			shoutrrr.correctSelf()

			// THEN the fields are corrected as necessary
			for k, v := range tc.want {
				got := shoutrrr.GetSelfURLField(k)
				if tc.mapTarget != "url_fields" {
					got = shoutrrr.GetSelfParam(k)
				}
				if got != v {
					t.Errorf("want %s:%q, not %q",
						k, v, got)
				}
			}
		})
	}
}

func TestShoutrrr_CheckValues(t *testing.T) {
	// GIVEN a Shoutrrr
	test := testShoutrrr(false, false)
	tests := map[string]struct {
		nilShoutrrr   bool
		sType         string
		options       map[string]string
		wantDelay     string
		urlFields     map[string]string
		wantURLFields map[string]string
		params        map[string]string
		main          *ShoutrrrDefaults
		errRegex      string
	}{
		"nil shoutrrr": {
			nilShoutrrr: true,
			errRegex:    "^$",
		},
		"empty": {
			errRegex:  "^type: <required>[^:]+://[^:]+$",
			urlFields: map[string]string{},
			params:    map[string]string{},
		},
		"invalid delay": {
			errRegex:  "delay: .* <invalid>",
			sType:     test.Type,
			urlFields: test.URLFields,
			options: map[string]string{
				"delay": "5x"},
		},
		"fixes delay": {
			errRegex:  "^$",
			sType:     test.Type,
			urlFields: test.URLFields,
			wantDelay: "5s",
			options: map[string]string{
				"delay": "5"},
		},
		"invalid message template": {
			errRegex:  "message: .* <invalid>",
			sType:     test.Type,
			urlFields: test.URLFields,
			options: map[string]string{
				"message": "{{ vesrion }"},
		},
		"valid message template": {
			errRegex:  "^$",
			sType:     test.Type,
			urlFields: test.URLFields,
			options: map[string]string{
				"message": "{{ vesrion }}"},
		},
		"invalid title template": {
			errRegex:  "title: .* <invalid>",
			sType:     test.Type,
			urlFields: test.URLFields,
			params: map[string]string{
				"title": "{{ version }"},
		},
		"valid title template": {
			errRegex:  "^$",
			sType:     test.Type,
			urlFields: test.URLFields,
			params: map[string]string{
				"title": "{{ version }}"},
		},
		"valid param template": {
			errRegex:  "^$",
			sType:     test.Type,
			urlFields: test.URLFields,
			params: map[string]string{
				"foo": "{{ vesrion }}"},
		},
		"invalid param template": {
			errRegex:  "foo: .* <invalid>",
			sType:     test.Type,
			urlFields: test.URLFields,
			params: map[string]string{
				"foo": "{{ version }"},
		},
		"invalid param and option": {
			errRegex:  `options:[^ ]+  delay: [^<]+<invalid>.*params:[^ ]+  title: [^<]+<invalid>`,
			sType:     test.Type,
			urlFields: test.URLFields,
			params: map[string]string{
				"title": "{{ version }"},
			options: map[string]string{
				"delay": "2x"},
		},
		"does correctSelf": {
			errRegex: "^$",
			sType:    test.Type,
			urlFields: map[string]string{
				"host":  "foo",
				"token": "bar",
				"port":  ":8080",
				"path":  "/test"},
			wantURLFields: map[string]string{
				"port": "8080",
				"path": "test"},
		},
		"valid": {
			errRegex:  "^$",
			urlFields: map[string]string{},
			main:      testShoutrrrDefaults(false, false),
		},
		"valid with self and main": {
			errRegex: "^$",
			urlFields: map[string]string{
				"host": "foo"},
			main: NewDefaults(
				test.Type,
				nil, nil,
				&map[string]string{
					"token": "bar"}),
		},
		"invalid url_fields": {
			errRegex: "^url_fields:.*host: <required>.*token: <required>[^:]+$",
			sType:    test.Type,
		},
		"invalid params + locate fail": {
			errRegex: "fromaddress: <required>.*toaddresses: <required>",
			urlFields: map[string]string{
				"host": "https://release-argus.io"},
			sType: "smtp",
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			// t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.Type = tc.sType
			if tc.main == nil {
				tc.main = &ShoutrrrDefaults{}
			}
			shoutrrr.Main = tc.main
			shoutrrr.Main.InitMaps()
			shoutrrr.Options = tc.options
			shoutrrr.URLFields = tc.urlFields
			shoutrrr.Params = tc.params
			if tc.nilShoutrrr {
				shoutrrr = nil
			}

			// WHEN CheckValues is called
			err := shoutrrr.CheckValues("")

			// THEN it err's when expected
			e := util.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Errorf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
			if tc.nilShoutrrr {
				return
			}
			if tc.wantDelay != "" && shoutrrr.GetSelfOption("delay") != tc.wantDelay {
				t.Errorf("delay not set/corrected. want match for %q\nnot: %q",
					tc.wantDelay, shoutrrr.GetSelfOption("delay"))
			}
			for key := range tc.wantURLFields {
				if shoutrrr.URLFields[key] != tc.wantURLFields[key] {
					t.Errorf("want: %q:%q\ngot:  %q:%q\n%v\n%v",
						key, tc.wantURLFields[key], key, shoutrrr.URLFields[key], tc.wantURLFields, shoutrrr.URLFields)
				}
			}
		})
	}
}

func TestSlice_CheckValues(t *testing.T) {
	// GIVEN a Slice
	tests := map[string]struct {
		slice    *Slice
		errRegex string
	}{
		"nil slice": {
			slice: nil, errRegex: "^$"},
		"valid slice": {
			errRegex: "^$",
			slice: &Slice{
				"valid": testShoutrrr(false, false),
				"other": testShoutrrr(false, false)}},
		"invalid slice": {
			errRegex: "type: <required>",
			slice: &Slice{
				"valid": testShoutrrr(false, false),
				"other": New(
					nil, "", nil, nil, "", nil, nil, nil, nil)}},
		"ordered errors": {
			errRegex: "aNotify.*type: <required>.*bNotify.*type: <required>",
			slice: &Slice{
				"aNotify": New(
					nil, "", nil, nil, "", nil, nil, nil, nil),
				"bNotify": New(
					nil, "", nil, nil, "", nil, nil, nil, nil)}},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.slice != nil {
				svcStatus := &svcstatus.Status{}
				svcStatus.Init(
					len(*tc.slice), 0, 0, nil, nil)
				tc.slice.Init(
					svcStatus,
					&SliceDefaults{}, &SliceDefaults{}, &SliceDefaults{})
			}

			// WHEN CheckValues is called
			eer := tc.slice.CheckValues("")

			// THEN it err's when expected
			e := util.ErrorToString(eer)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Errorf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
		})
	}
}

func TestSliceDefaults_CheckValues(t *testing.T) {
	// GIVEN a SliceDefaults
	tests := map[string]struct {
		slice    *SliceDefaults
		errRegex string
	}{
		"nil slice": {
			slice: nil, errRegex: "^$"},
		"valid slice": {
			errRegex: "^$",
			slice: &SliceDefaults{
				"valid": testShoutrrrDefaults(false, false),
				"other": testShoutrrrDefaults(false, false)}},
		"invalid type": {
			errRegex: `type: "[^"]+" <invalid>`,
			slice: &SliceDefaults{
				"valid": testShoutrrrDefaults(false, false),
				"other": NewDefaults(
					"sommethingUnknown",
					nil, nil, nil)}},
		"delay without unit": {
			errRegex: `^$`,
			slice: &SliceDefaults{
				"foo": NewDefaults(
					"gotify",
					&map[string]string{
						"delay": "1"},
					nil, nil)}},
		"invalid delay": {
			errRegex: `delay: "[^"]+" <invalid>`,
			slice: &SliceDefaults{
				"foo": NewDefaults(
					"gotify",
					&map[string]string{
						"delay": "1x"},
					nil, nil)}},
		"invalid message template": {
			errRegex: `message: "[^"]+" <invalid>`,
			slice: &SliceDefaults{
				"bar": NewDefaults(
					"gotify",
					&map[string]string{
						"message": "{{ .foo }"},
					nil, nil)}},
		"invalid params template": {
			errRegex: `title: "[^"]+" <invalid>`,
			slice: &SliceDefaults{
				"bar": NewDefaults(
					"gotify",
					nil,
					&map[string]string{
						"title": "{{ .bar }"},
					nil)}},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CheckValues is called
			eer := tc.slice.CheckValues("")

			// THEN it err's when expected
			e := util.ErrorToString(eer)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Errorf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
		})
	}
}
