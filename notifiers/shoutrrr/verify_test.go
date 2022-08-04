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

package shoutrrr

import (
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/release-argus/Argus/utils"
)

func TestShoutrrrPrint(t *testing.T) {
	// GIVEN a Service
	tests := map[string]struct {
		sType     string
		options   map[string]string
		urlFields map[string]string
		params    map[string]string
		lines     int
	}{
		"all empty": {lines: 0},
		"only type": {lines: 1, sType: "gotify"},
		"type and options": {lines: 3, sType: "gotify",
			options: map[string]string{"foo": "bar"}},
		"type,options and urlFields": {lines: 5, sType: "gotify",
			options:   map[string]string{"foo": "bar"},
			urlFields: map[string]string{"foo": "bar"}},
		"type,options,urlFields and params": {lines: 7, sType: "gotify",
			options:   map[string]string{"foo": "bar"},
			urlFields: map[string]string{"foo": "bar"},
			params:    map[string]string{"foo": "bar"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			shoutrrr := Shoutrrr{
				Type:      tc.sType,
				Options:   tc.options,
				URLFields: tc.urlFields,
				Params:    tc.params,
			}

			// WHEN Print is called
			shoutrrr.Print("")

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
		"nil slice": {lines: 0, slice: nil},
		"single element slice": {lines: 9, slice: &Slice{"single": testShoutrrr(false, false, false)},
			regexMatches: []string{"^notify:$", "^  single:$", "^    type: ", "^    options:$", "^      max_tries: ", "^    url_fields:$", "^      token: "}},
		"multiple element slice": {lines: 17, slice: &Slice{"first": testShoutrrr(false, false, false), "second": testShoutrrr(true, true, true)},
			regexMatches: []string{"^notify:$", "^  first:$", "^    type: ", "^    options:$", "^      max_tries: ", "^    url_fields:$", "^      token: ", "^  second:$"}},
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
			strOut := string(out)
			got := strings.Count(strOut, "\n")
			if got != tc.lines {
				t.Errorf("Print should have given %d lines, but gave %d\n%s",
					tc.lines, got, out)
			}
			lines := strings.Split(strOut, "\n")
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
						regex, strOut)
				}
			}
		})
	}
}

func TestCheckValuesMaster(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		sType              *string
		options            map[string]string
		urlFields          map[string]string
		params             map[string]string
		main               Shoutrrr
		errsRegex          string
		errsOptionsRegex   string
		errsURLFieldsRegex string
		errsParamsRegex    string
	}{
		"no type":           {errsRegex: "type: <required>", sType: stringPtr("")},
		"invalid type":      {errsRegex: "type: .* <invalid>", sType: stringPtr("argus")},
		"discord - invalid": {sType: stringPtr("discord"), errsURLFieldsRegex: "token: <required>.*webhookid: <required>"},
		"discord - no token": {sType: stringPtr("discord"), errsURLFieldsRegex: "token: <required>",
			urlFields: map[string]string{"webhookid": "bash"}},
		"discord - no webhookid": {sType: stringPtr("discord"), errsURLFieldsRegex: "webhookid: <required>",
			urlFields: map[string]string{"token": "bish"}},
		"discord - valid": {sType: stringPtr("discord"),
			urlFields: map[string]string{"token": "bish", "webhookid": "webhookid"}},
		"discord - valid with main": {sType: stringPtr("discord"),
			main: Shoutrrr{URLFields: map[string]string{"token": "bish", "webhookid": "bash"}}},
		"smtp - invalid": {sType: stringPtr("smtp"), errsURLFieldsRegex: "host: <required>.*", errsParamsRegex: "fromaddress: <required>.*toaddresses: <required>"},
		"smtp - no host": {sType: stringPtr("smtp"), errsURLFieldsRegex: "host: <required>",
			params: map[string]string{"fromaddress": "bash", "toaddresses": "bosh"}},
		"smtp - no fromaddress": {sType: stringPtr("smtp"), errsParamsRegex: "fromaddress: <required>",
			urlFields: map[string]string{"host": "bish"},
			params:    map[string]string{"toaddresses": "bosh"}},
		"smtp - no toaddresses": {sType: stringPtr("smtp"), errsParamsRegex: "toaddresses: <required>",
			urlFields: map[string]string{"host": "bish"},
			params:    map[string]string{"fromaddress": "bash"}},
		"smtp - valid": {sType: stringPtr("smtp"),
			urlFields: map[string]string{"host": "bish"},
			params:    map[string]string{"fromaddress": "bash", "toaddresses": "bosh"}},
		"smtp - valid with main": {sType: stringPtr("smtp"),
			main: Shoutrrr{URLFields: map[string]string{"host": "bish"},
				Params: map[string]string{"fromaddress": "bash", "toaddresses": "bosh"}}},
		"gotify - invalid": {sType: stringPtr("gotify"), errsURLFieldsRegex: "host: <required>.*token: <required>"},
		"gotify - no host": {sType: stringPtr("gotify"), errsURLFieldsRegex: "host: <required>",
			urlFields: map[string]string{"token": "bash"}},
		"gotify - no token": {sType: stringPtr("gotify"), errsURLFieldsRegex: "token: <required>",
			urlFields: map[string]string{"host": "bish"}},
		"gotify - valid": {sType: stringPtr("gotify"),
			urlFields: map[string]string{"host": "bish", "token": "bash"}},
		"gotify - valid with main": {sType: stringPtr("gotify"),
			main: Shoutrrr{URLFields: map[string]string{"host": "bish", "token": "bash"}}},
		"googlechat - invalid": {sType: stringPtr("googlechat"), errsURLFieldsRegex: "raw: <required>"},
		"googlechat - valid": {sType: stringPtr("googlechat"),
			urlFields: map[string]string{"raw": "bish"}},
		"googlechat - valid with main": {sType: stringPtr("googlechat"),
			main: Shoutrrr{URLFields: map[string]string{"raw": "bish"}}},
		"ifttt - invalid": {sType: stringPtr("ifttt"), errsURLFieldsRegex: "webhookid: <required>", errsParamsRegex: "events: <required>"},
		"ifttt - no webhookid": {sType: stringPtr("ifttt"), errsURLFieldsRegex: "webhookid: <required>",
			urlFields: map[string]string{},
			params:    map[string]string{"events": "bash"}},
		"ifttt - no events": {sType: stringPtr("ifttt"), errsParamsRegex: "events: <required>",
			urlFields: map[string]string{"webhookid": "bish"}},
		"ifttt - valid": {sType: stringPtr("ifttt"),
			urlFields: map[string]string{"webhookid": "bish"},
			params:    map[string]string{"events": "events"}},
		"ifttt - valid with main": {sType: stringPtr("ifttt"),
			main: Shoutrrr{URLFields: map[string]string{"webhookid": "webhookid"}, Params: map[string]string{"events": "events"}}},
		"join - invalid": {sType: stringPtr("join"), errsURLFieldsRegex: "apikey: <required>", errsParamsRegex: "devices: <required>"},
		"join - no apikey": {sType: stringPtr("join"), errsURLFieldsRegex: "apikey: <required>",
			urlFields: map[string]string{},
			params:    map[string]string{"devices": "bash"}},
		"join - no devices": {sType: stringPtr("join"), errsParamsRegex: "devices: <required>",
			urlFields: map[string]string{"apikey": "bish"}},
		"join - valid": {sType: stringPtr("join"),
			urlFields: map[string]string{"apikey": "bish"},
			params:    map[string]string{"devices": "devices"}},
		"join - valid with main": {sType: stringPtr("join"),
			main: Shoutrrr{URLFields: map[string]string{"apikey": "apikey"}, Params: map[string]string{"devices": "devices"}}},
		"mattermost - invalid": {sType: stringPtr("mattermost"), errsURLFieldsRegex: "host: <required>.*token: <required>"},
		"mattermost - no host": {sType: stringPtr("mattermost"), errsURLFieldsRegex: "host: <required>",
			urlFields: map[string]string{"token": "bash"}},
		"mattermost - no token": {sType: stringPtr("mattermost"), errsURLFieldsRegex: "token: <required>",
			urlFields: map[string]string{"host": "bish"}},
		"mattermost - valid": {sType: stringPtr("mattermost"),
			urlFields: map[string]string{"host": "bish", "token": "bash"}},
		"mattermost - valid with main": {sType: stringPtr("mattermost"),
			main: Shoutrrr{URLFields: map[string]string{"host": "bish", "token": "bash"}}},
		"matrix - invalid": {sType: stringPtr("matrix"), errsURLFieldsRegex: "host: <required>.*password: <required>"},
		"matrix - no host": {sType: stringPtr("matrix"), errsURLFieldsRegex: "host: <required>",
			urlFields: map[string]string{"password": "bash"}},
		"matrix - no password": {sType: stringPtr("matrix"), errsURLFieldsRegex: "password: <required>",
			urlFields: map[string]string{"host": "bish"}},
		"matrix - valid": {sType: stringPtr("matrix"),
			urlFields: map[string]string{"host": "bish", "password": "password"}},
		"matrix - valid with main": {sType: stringPtr("matrix"),
			main: Shoutrrr{URLFields: map[string]string{"host": "bish", "password": "bash"}}},
		"opsgenie - invalid": {sType: stringPtr("opsgenie"), errsURLFieldsRegex: "apikey: <required>"},
		"opsgenie - valid": {sType: stringPtr("opsgenie"),
			urlFields: map[string]string{"apikey": "apikey"}},
		"opsgenie - valid with main": {sType: stringPtr("opsgenie"),
			main: Shoutrrr{URLFields: map[string]string{"apikey": "apikey"}}},
		"pushbullet - invalid": {sType: stringPtr("pushbullet"), errsURLFieldsRegex: "token: <required>.*targets: <required>"},
		"pushbullet - no token": {sType: stringPtr("pushbullet"), errsURLFieldsRegex: "token: <required>",
			urlFields: map[string]string{"targets": "bash"}},
		"pushbullet - no targets": {sType: stringPtr("pushbullet"), errsURLFieldsRegex: "targets: <required>",
			urlFields: map[string]string{"token": "bish"}},
		"pushbullet - valid": {sType: stringPtr("pushbullet"),
			urlFields: map[string]string{"token": "bish", "targets": "targets"}},
		"pushbullet - valid with main": {sType: stringPtr("pushbullet"),
			main: Shoutrrr{URLFields: map[string]string{"token": "bish", "targets": "bash"}}},
		"pushover - invalid": {sType: stringPtr("pushover"), errsURLFieldsRegex: "token: <required>.*user: <required>"},
		"pushover - no token": {sType: stringPtr("pushover"), errsURLFieldsRegex: "token: <required>",
			urlFields: map[string]string{"user": "bash"}},
		"pushover - no user": {sType: stringPtr("pushover"), errsURLFieldsRegex: "user: <required>",
			urlFields: map[string]string{"token": "bish"}},
		"pushover - valid": {sType: stringPtr("pushover"),
			urlFields: map[string]string{"token": "bish", "user": "user"}},
		"pushover - valid with main": {sType: stringPtr("pushover"),
			main: Shoutrrr{URLFields: map[string]string{"token": "bish", "user": "bash"}}},
		"rocketchat - invalid": {sType: stringPtr("rocketchat"), errsURLFieldsRegex: "host: <required>.*tokena: <required>.*tokenb: <required>.*channel: <required>"},
		"rocketchat - no host": {sType: stringPtr("rocketchat"), errsURLFieldsRegex: "host: <required>",
			urlFields: map[string]string{"tokena": "bash", "tokenb": "bosh", "channel": "bing"}},
		"rocketchat - no tokena": {sType: stringPtr("rocketchat"), errsURLFieldsRegex: "tokena: <required>",
			urlFields: map[string]string{"host": "bish", "tokenb": "bash", "channel": "bing"}},
		"rocketchat - no tokenb": {sType: stringPtr("rocketchat"), errsURLFieldsRegex: "tokenb: <required>",
			urlFields: map[string]string{"host": "bish", "tokena": "bash", "channel": "bing"}},
		"rocketchat - no channel": {sType: stringPtr("rocketchat"), errsURLFieldsRegex: "channel: <required>",
			urlFields: map[string]string{"host": "bish", "tokena": "bash", "tokenb": "bosh"}},
		"rocketchat - valid": {sType: stringPtr("rocketchat"),
			urlFields: map[string]string{"host": "bish", "tokena": "bash", "tokenb": "bosh", "channel": "bing"}},
		"rocketchat - valid with main": {sType: stringPtr("rocketchat"),
			main: Shoutrrr{URLFields: map[string]string{"host": "bish", "tokena": "bash", "tokenb": "bosh", "channel": "bing"}}},
		"slack - invalid": {sType: stringPtr("slack"), errsURLFieldsRegex: "token: <required>.*channel: <required>"},
		"slack - no token": {sType: stringPtr("slack"), errsURLFieldsRegex: "token: <required>",
			urlFields: map[string]string{"channel": "bash"}},
		"slack - no channel": {sType: stringPtr("slack"), errsURLFieldsRegex: "channel: <required>",
			urlFields: map[string]string{"token": "bish"}},
		"slack - valid": {sType: stringPtr("slack"),
			urlFields: map[string]string{"token": "bish", "channel": "channel"}},
		"slack - valid with main": {sType: stringPtr("slack"),
			main: Shoutrrr{URLFields: map[string]string{"token": "bish", "channel": "bash"}}},
		"teams - invalid": {sType: stringPtr("teams"), errsURLFieldsRegex: "group: <required>.*tenant: <required>.*altid: <required>.*groupowner: <required>", errsParamsRegex: "host: <required>"},
		"teams - no group": {sType: stringPtr("teams"), errsURLFieldsRegex: "group: <required>",
			urlFields: map[string]string{"tenant": "bash", "altid": "bosh", "groupowner": "bing"},
			params:    map[string]string{"host": "https://release-argus.io"}},
		"teams - no tenant": {sType: stringPtr("teams"), errsURLFieldsRegex: "tenant: <required>",
			urlFields: map[string]string{"group": "bish", "altid": "bash", "groupowner": "bing"},
			params:    map[string]string{"host": "https://release-argus.io"}},
		"teams - no altid": {sType: stringPtr("teams"), errsURLFieldsRegex: "altid: <required>",
			urlFields: map[string]string{"group": "bish", "tenant": "bash", "groupowner": "bing"},
			params:    map[string]string{"host": "https://release-argus.io"}},
		"teams - no groupowner": {sType: stringPtr("teams"), errsURLFieldsRegex: "groupowner: <required>",
			urlFields: map[string]string{"group": "bish", "tenant": "bash", "altid": "bosh"},
			params:    map[string]string{"host": "https://release-argus.io"}},
		"teams - no host": {sType: stringPtr("teams"), errsParamsRegex: "host: <required>",
			urlFields: map[string]string{"group": "bish", "tenant": "bash", "altid": "bosh", "groupowner": "bing"}},
		"teams - valid": {sType: stringPtr("teams"),
			urlFields: map[string]string{"group": "bish", "tenant": "bash", "altid": "bosh", "groupowner": "bing"},
			params:    map[string]string{"host": "https://release-argus.io"}},
		"teams - valid with main": {sType: stringPtr("teams"),
			main: Shoutrrr{URLFields: map[string]string{"group": "bish", "tenant": "bash", "altid": "bosh", "groupowner": "bing"},
				Params: map[string]string{"host": "https://release-argus.io"}}},
		"telegram - invalid": {sType: stringPtr("telegram"), errsURLFieldsRegex: "token: <required>", errsParamsRegex: "chats: <required>"},
		"telegram - no token": {sType: stringPtr("telegram"), errsURLFieldsRegex: "token: <required>",
			urlFields: map[string]string{},
			params:    map[string]string{"chats": "bash"}},
		"telegram - no chats": {sType: stringPtr("telegram"), errsParamsRegex: "chats: <required>",
			urlFields: map[string]string{"token": "bish"}},
		"telegram - valid": {sType: stringPtr("telegram"),
			urlFields: map[string]string{"token": "bish"},
			params:    map[string]string{"chats": "chats"}},
		"telegram - valid with main": {sType: stringPtr("telegram"),
			main: Shoutrrr{URLFields: map[string]string{"token": "bish"}, Params: map[string]string{"chats": "chats"}}},
		"zulip_chat - invalid": {sType: stringPtr("zulip_chat"), errsURLFieldsRegex: "host: <required>.*botmail: <required>.*botkey: <required>"},
		"zulip_chat - no host": {sType: stringPtr("zulip_chat"), errsURLFieldsRegex: "host: <required>",
			urlFields: map[string]string{"botmail": "bash", "botkey": "bosh"}},
		"zulip_chat - no botmail": {sType: stringPtr("zulip_chat"), errsURLFieldsRegex: "botmail: <required>",
			urlFields: map[string]string{"host": "bish", "botkey": "bash"}},
		"zulip_chat - no botkey": {sType: stringPtr("zulip_chat"), errsURLFieldsRegex: "botkey: <required>",
			urlFields: map[string]string{"host": "bish", "botmail": "bash"}},
		"zulip_chat - valid": {sType: stringPtr("zulip_chat"),
			urlFields: map[string]string{"host": "bish", "botmail": "bash", "botkey": "bosh"}},
		"zulip_chat - valid with main": {sType: stringPtr("zulip_chat"),
			main: Shoutrrr{URLFields: map[string]string{"host": "bish", "botmail": "bash", "botkey": "bosh"}}},
		"shoutrrr - invalid": {sType: stringPtr("shoutrrr"), errsURLFieldsRegex: "raw: <required>"},
		"shoutrrr - valid": {sType: stringPtr("shoutrrr"),
			urlFields: map[string]string{"raw": "bish"}},
		"shoutrrr - valid with main": {sType: stringPtr("shoutrrr"),
			main: Shoutrrr{URLFields: map[string]string{"raw": "bish"}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			shoutrrr := testShoutrrr(false, true, false)
			if tc.sType != nil {
				shoutrrr.Type = *tc.sType
			}
			shoutrrr.Main = &tc.main
			shoutrrr.Main.InitMaps()
			shoutrrr.Options = tc.options
			shoutrrr.URLFields = tc.urlFields
			shoutrrr.Params = tc.params

			// WHEN CheckValues is called
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
			e := utils.ErrorToString(errs)
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
			e = utils.ErrorToString(errsOptions)
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
			e = utils.ErrorToString(errsURLFields)
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
			e = utils.ErrorToString(errsParams)
			re = regexp.MustCompile(tc.errsParamsRegex)
			match = re.MatchString(e)
			if !match {
				t.Errorf("want match for %q\nnot: %q",
					tc.errsParamsRegex, e)
			}
		})
	}
}

func TestCorrectSelf(t *testing.T) {
	// GIVEN a Service
	tests := map[string]struct {
		sType     string
		mapTarget string
		key       string
		startAs   string
		want      string
	}{
		"port - leading colon":                   {mapTarget: "url_fields", key: "port", startAs: ":8080", want: "8080"},
		"port - valid":                           {mapTarget: "url_fields", key: "port", startAs: "8080", want: "8080"},
		"path - leading slash":                   {mapTarget: "url_fields", key: "path", startAs: "/argus", want: "argus"},
		"path - valid":                           {mapTarget: "url_fields", key: "path", startAs: "argus", want: "argus"},
		"mattermost - channel, leading slash":    {sType: "mattermost", mapTarget: "url_fields", key: "channel", startAs: "/argus", want: "argus"},
		"mattermost - channel, valid":            {sType: "mattermost", mapTarget: "url_fields", key: "channel", startAs: "argus", want: "argus"},
		"slack - color, # instead of %23":        {sType: "slack", mapTarget: "params", key: "color", startAs: "#ffffff", want: "%23ffffff"},
		"slack - color, valid":                   {sType: "slack", mapTarget: "params", key: "color", startAs: "%23ffffff", want: "%23ffffff"},
		"teams - altid, leading slash":           {sType: "teams", mapTarget: "url_fields", key: "altid", startAs: "/argus", want: "argus"},
		"teams - altid, valid":                   {sType: "teams", mapTarget: "url_fields", key: "altid", startAs: "argus", want: "argus"},
		"teams - groupowner, leading slash":      {sType: "teams", mapTarget: "url_fields", key: "groupowner", startAs: "/argus", want: "argus"},
		"teams - groupowner, valid":              {sType: "teams", mapTarget: "url_fields", key: "groupowner", startAs: "argus", want: "argus"},
		"zulip_chat - botmail, @ instead of %40": {sType: "zulip_chat", mapTarget: "url_fields", key: "botmail", startAs: "foo@bar.com", want: "foo%40bar.com"},
		"zulip_chat - botmail, valid":            {sType: "zulip_chat", mapTarget: "url_fields", key: "botmail", startAs: "foo%40bar.com", want: "foo%40bar.com"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			shoutrrr := Shoutrrr{Type: tc.sType}
			shoutrrr.InitMaps()
			if tc.mapTarget == "url_fields" {
				shoutrrr.SetURLField(tc.key, tc.startAs)
			} else {
				shoutrrr.SetParam(tc.key, tc.startAs)
			}

			// WHEN correctSelf is called
			shoutrrr.correctSelf()

			// THEN the field is corrected when necessary
			got := shoutrrr.GetSelfURLField(tc.key)
			if tc.mapTarget != "url_fields" {
				got = shoutrrr.GetSelfParam(tc.key)
			}
			if got != tc.want {
				t.Errorf("want: %s:%q\ngot:  %s:%q",
					tc.key, tc.want, tc.key, got)
			}
		})
	}
}

func TestShoutrrrCheckValues(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		nilShoutrrr     bool
		serviceShoutrrr bool
		sType           string
		options         map[string]string
		wantDelay       string
		urlFields       map[string]string
		wantURLFields   map[string]string
		params          map[string]string
		main            *Shoutrrr
		errRegex        string
	}{
		"nil shoutrrr": {nilShoutrrr: true, errRegex: "^$"},
		"invalid delay": {errRegex: "delay: .* <invalid>",
			options: map[string]string{"delay": "5x"}},
		"fixes delay": {errRegex: "^$", wantDelay: "5s",
			options: map[string]string{"delay": "5"}},
		"runs correctSelf": {errRegex: "^$", urlFields: map[string]string{"port": ":8080", "path": "/test"},
			wantURLFields: map[string]string{"port": "8080", "path": "test"}},
		"doesn't check non-service url_fields/params": {errRegex: "^$", urlFields: map[string]string{}, params: map[string]string{}},
		"does field check service shoutrrrs - valid": {errRegex: "^$", serviceShoutrrr: true, urlFields: map[string]string{},
			main: testShoutrrr(false, false, false)},
		"does field check service shoutrrrs - valid with self and main": {errRegex: "^$", serviceShoutrrr: true, urlFields: map[string]string{"host": "foo"},
			main: &Shoutrrr{URLFields: map[string]string{"token": "bar"}}},
		"does field check service shoutrrrs - invalid urlfields": {errRegex: "host: <required>.*token: <required>", serviceShoutrrr: true,
			main: &Shoutrrr{}, sType: "gotify"},
		"does field check service shoutrrrs - invalid params + locate fail": {errRegex: "fromaddress: <required>.*toaddresses: <required>", serviceShoutrrr: true,
			urlFields: map[string]string{"host": "https://release-argus.io"}, main: &Shoutrrr{}, sType: "smtp"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			shoutrrr := testShoutrrr(false, tc.serviceShoutrrr, false)
			if tc.sType != "" {
				shoutrrr.Type = tc.sType
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
			eer := shoutrrr.CheckValues("")

			// THEN it err's when expected
			e := utils.ErrorToString(eer)
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

func TestSliceCheckValues(t *testing.T) {
	// GIVEN a Slice
	tests := map[string]struct {
		slice    *Slice
		errRegex string
	}{
		"nil shoutrrr": {slice: nil, errRegex: "^$"},
		"valid slice": {errRegex: "^$",
			slice: &Slice{"valid": testShoutrrr(false, true, false), "other": testShoutrrr(false, true, false)}},
		"invalid slice": {errRegex: "type: <required>",
			slice: &Slice{"valid": testShoutrrr(false, true, false), "other": &Shoutrrr{Main: &Shoutrrr{}}}},
		"ordered errors": {errRegex: "aNotify.*type: <required>.*bNotify.*type: <required>",
			slice: &Slice{"aNotify": &Shoutrrr{Main: &Shoutrrr{}}, "bNotify": &Shoutrrr{Main: &Shoutrrr{}}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// WHEN CheckValues is called
			eer := tc.slice.CheckValues("")

			// THEN it err's when expected
			e := utils.ErrorToString(eer)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Errorf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
		})
	}
}
