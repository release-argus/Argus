// Copyright [2024] [Argus]
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
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestShoutrrr_checkValuesType(t *testing.T) {
	// GIVEN a Shoutrrr with a Type and possibly a Main
	tests := map[string]struct {
		sType     string
		main      *Defaults
		errsRegex string
	}{
		"no type": {
			errsRegex: `^type: <required>.*$`,
			sType:     "",
		},
		"invalid type": {
			errsRegex: `^type: .* <invalid>.*$`,
			sType:     "argus",
		},
		"invalid type - type in main differs": {
			errsRegex: `^type: "gotify" != "discord" <invalid>.*$`,
			sType:     "gotify",
			main: NewDefaults("discord",
				make(map[string]string), make(map[string]string), make(map[string]string)),
		},
		"valid type": {
			errsRegex: `^$`,
			sType:     "gotify",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.Type = tc.sType
			svcStatus := status.New(
				nil, nil, nil,
				"", "", "", "", "", "")
			svcStatus.Init(
				1, 0, 0,
				test.StringPtr(name), nil)
			shoutrrr.Init(
				svcStatus,
				tc.main,
				&Defaults{}, &Defaults{})

			// WHEN checkValuesType is called
			errsType := shoutrrr.checkValuesType("")

			// THEN it errors when expected
			if tc.errsRegex == "" {
				tc.errsRegex = "^$"
			}
			e := util.ErrorToString(errsType)
			if !util.RegexCheck(tc.errsRegex, e) {
				t.Errorf("Shoutrrr.checkValuesType, want error on type to match:\n%q\ngot:\n%q\n",
					tc.errsRegex, e)
			}
		})
	}
}

func TestBase_checkValuesOptions(t *testing.T) {
	// GIVEN a Base with Options
	tests := map[string]struct {
		options   map[string]string
		wantDelay string
		errRegex  string
	}{
		"no options": {
			options:  map[string]string{},
			errRegex: `^$`,
		},
		"valid delay": {
			options: map[string]string{
				"delay": "5s"},
			wantDelay: "5s",
			errRegex:  `^$`,
		},
		"invalid delay": {
			options: map[string]string{
				"delay": "5x"},
			errRegex: test.TrimYAML(`
				^delay: "5x" <invalid>.*$`),
		},
		"fixes delay missing unit": {
			options: map[string]string{
				"delay": "5"},
			wantDelay: "5s",
			errRegex:  `^$`,
		},
		"valid message template": {
			options: map[string]string{
				"message": "release! {{ version }}"},
			errRegex: `^$`,
		},
		"invalid message template": {
			options: map[string]string{
				"message": "release! {{ version }"},
			errRegex: test.TrimYAML(`
					^message: "release! {{ version }" <invalid>.*$`),
		},
		"min max_tries": {
			options: map[string]string{
				"max_tries": "0"},
			errRegex: test.TrimYAML(`^$`),
		},
		"max max_tries": {
			options: map[string]string{
				"max_tries": fmt.Sprint(math.MaxUint8)},
			errRegex: test.TrimYAML(`^$`),
		},
		"invalid max_tries - too large": {
			options: map[string]string{
				"max_tries": fmt.Sprint(math.MaxUint16)},
			errRegex: test.TrimYAML(`
				^max_tries: "\d+" <invalid>.*$`),
		},
		"invalid max_tries - too large, >uint64": {
			options: map[string]string{
				"max_tries": fmt.Sprintf("1%d",
					uint(math.MaxUint64))},
			errRegex: test.TrimYAML(`
				^max_tries: "\d+" <invalid>.*$`),
		},
		"invalid max_tries - not a number": {
			options: map[string]string{
				"max_tries": "oneOrTwo"},
			errRegex: test.TrimYAML(`
				^max_tries: "oneOrTwo" <invalid>.*$`),
		},
		"invalid max_tries - negative": {
			options: map[string]string{
				"max_tries": "-1"},
			errRegex: test.TrimYAML(`
				^max_tries: "-1" <invalid>.*$`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := Base{}
			shoutrrr.Options = tc.options

			// WHEN checkValuesOptions is called
			err := shoutrrr.checkValuesOptions("")

			// THEN it errors when expected
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("Shoutrrr.checkValuesOptions(), want error on options to match:\n%q\ngot:\n%q\n",
					tc.errRegex, e)
			}
			// AND the delay is set as expected if it didn't error on delay
			if !util.RegexCheck("^delay:.*", e) &&
				shoutrrr.GetOption("delay") != tc.wantDelay {
				t.Errorf("Shoutrrr.checkValuesOptions(), want delay to be %q, got %q",
					tc.wantDelay, shoutrrr.GetOption("delay"))
			}
		})
	}
}

func TestBase_checkValuesParams(t *testing.T) {
	// GIVEN a Base with Params
	tests := map[string]struct {
		params   map[string]string
		errRegex string
	}{
		"no params": {
			params:   map[string]string{},
			errRegex: `^$`,
		},
		"valid message template": {
			params: map[string]string{
				"message": "release! {{ version }}"},
			errRegex: `^$`,
		},
		"invalid message template": {
			params: map[string]string{
				"a": "release! {{ version }"},
			errRegex: `^a: "release! {{ version }" <invalid>.*$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := Base{}
			shoutrrr.Params = tc.params

			// WHEN checkValuesParams is called
			err := shoutrrr.checkValuesParams("")

			// THEN it errors when expected
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("Shoutrrr.checkValuesParams(), want error on params to match:\n%q\ngot:\n%q\n",
					tc.errRegex, e)
			}
		})
	}
}

func TestBase_CheckValues(t *testing.T) {
	tests := map[string]struct {
		base     *Base
		want     *Base
		id       string
		errRegex string
	}{
		"nil Base": {
			base:     nil,
			errRegex: `^$`,
		},
		"valid Base": {
			base: &Base{
				Type: "slack",
				Options: map[string]string{
					"delay": "10s"},
				Params: map[string]string{
					"color": "orange"},
			},
			errRegex: `^$`,
		},
		"invalid delay option": {
			base: &Base{
				Type: "slack",
				Options: map[string]string{
					"delay": "10x"},
			},
			errRegex: test.TrimYAML(`
				^options:
					delay: "10x" <invalid>.*$`),
		},
		"invalid param template": {
			base: &Base{
				Type: "slack",
				Params: map[string]string{
					"color": "{{ invalid template }}"},
			},
			errRegex: test.TrimYAML(`
				^params:
					color: "{{ invalid template }}" <invalid>.*$`),
		},
		"multiple errors": {
			base: &Base{
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
					color: "{{ invalid template }}" <invalid>.*$`),
		},
		"matrix - rooms, leading #": {
			base: &Base{
				Type: "matrix",
				Params: map[string]string{
					"rooms": "#alias:server"}},
			errRegex: `^$`,
			want: &Base{
				Type: "matrix",
				Params: map[string]string{
					"rooms": "alias:server"}},
		},
		"matrix - rooms, leading # already urlEncoded": {
			base: &Base{
				Type: "matrix",
				Params: map[string]string{
					"rooms": "%23alias:server"}},
			errRegex: `^$`,
			want: &Base{
				Type: "matrix",
				Params: map[string]string{
					"rooms": "%23alias:server"}},
		},
		"matrix - rooms, valid": {
			base: &Base{
				Type: "matrix",
				Params: map[string]string{
					"rooms": "alias:server"}},
			errRegex: `^$`,
			want: &Base{
				Type: "matrix",
				Params: map[string]string{
					"rooms": "alias:server"}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.want == nil && tc.base != nil {
				tc.want = &Base{
					Type:      tc.base.Type,
					Options:   util.CopyMap(tc.base.Options),
					Params:    util.CopyMap(tc.base.Params),
					URLFields: util.CopyMap(tc.base.URLFields),
				}
			}

			// WHEN CheckValues is called
			err := tc.base.CheckValues("", tc.id)

			// THEN it errors when expected
			e := util.ErrorToString(err)
			lines := strings.Split(e, "\n")
			wantLines := strings.Count(tc.errRegex, "\n")
			if wantLines > len(lines) {
				t.Fatalf("Base.CheckValues() want %d lines of error:\n%q\ngot %d lines:\n%v\nstdout: %q",
					wantLines, tc.errRegex, len(lines), lines, e)
				return
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("Base.CheckValues() error mismatch\nwant match for:\n%q\ngot:\n%q",
					tc.errRegex, e)
				return
			}
			// AND the Base is as expected
			wantStr := util.ToYAMLString(tc.want, "")
			gotStr := util.ToYAMLString(tc.base, "")
			if gotStr != wantStr {
				t.Errorf("Base.CheckValues() mismatch\nwant:\n%q\ngot:\n%q",
					wantStr, gotStr)
			}
		})
	}
}

func TestShoutrrr_checkValuesURLFields(t *testing.T) {
	// GIVEN a Shoutrrr with Params
	tests := map[string]struct {
		sType     string
		urlFields map[string]string
		main      *Defaults
		errRegex  string
	}{
		"bark - invalid": {
			sType: "bark",
			errRegex: test.TrimYAML(`
				^devicekey: <required>.*
				host: <required>.*$`),
		},
		"bark - no devicekey": {
			sType: "bark",
			urlFields: map[string]string{
				"host": "https://example.com"},
			errRegex: test.TrimYAML(`
				^devicekey: <required>.*$`),
		},
		"bark - no host": {
			sType: "bark",
			urlFields: map[string]string{
				"devicekey": "foo"},
			errRegex: test.TrimYAML(`
				^host: <required>.*$`),
		},
		"bark - valid": {
			sType: "bark",
			urlFields: map[string]string{
				"devicekey": "foo",
				"host":      "https://example.com"},
			errRegex: `^$`,
		},
		"discord - invalid": {
			sType: "discord",
			errRegex: test.TrimYAML(`
				^token: <required>.*
				webhookid: <required>.*$`),
		},
		"discord - no token": {
			sType: "discord",
			urlFields: map[string]string{
				"webhookid": "bash"},
			errRegex: test.TrimYAML(`
				^token: <required>.*$`),
		},
		"discord - no webhookid": {
			sType: "discord",
			urlFields: map[string]string{
				"token": "bish"},
			errRegex: test.TrimYAML(`
				^webhookid: <required>.*$`),
		},
		"discord - valid": {
			sType: "discord",
			urlFields: map[string]string{
				"token":     "bish",
				"webhookid": "webhookid"},
			errRegex: `^$`,
		},
		"discord - valid with main": {
			sType: "discord",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"token":     "bish",
					"webhookid": "bash"},
				nil),
			errRegex: `^$`,
		},
		"smtp - no host": {
			sType:     "smtp",
			urlFields: map[string]string{},
			errRegex: test.TrimYAML(`
				^host: <required>.*$`),
		},
		"smtp - valid": {
			sType: "smtp",
			urlFields: map[string]string{
				"host": "bish"},
			errRegex: `^$`,
		},
		"smtp - valid with main": {
			sType: "smtp",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"host": "bish"},
				nil),
			errRegex: `^$`,
		},
		"gotify - invalid": {
			sType: "gotify",
			errRegex: test.TrimYAML(`
				^host: <required>.*
				token: <required>.*$`),
		},
		"gotify - no host": {
			sType: "gotify",
			urlFields: map[string]string{
				"token": "bash"},
			errRegex: test.TrimYAML(`
				^host: <required>.*$`),
		},
		"gotify - no token": {
			sType: "gotify",
			urlFields: map[string]string{
				"host": "bish"},
			errRegex: test.TrimYAML(`
				^token: <required>.*$`),
		},
		"gotify - valid": {
			sType: "gotify",
			urlFields: map[string]string{
				"host":  "bish",
				"token": "bash"},
			errRegex: `^$`,
		},
		"gotify - valid with main": {
			sType: "gotify",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"host":  "bish",
					"token": "bash"},
				nil),
			errRegex: `^$`,
		},
		"googlechat - invalid": {
			sType: "googlechat",
			errRegex: test.TrimYAML(`
				^raw: <required>.*$`),
		},
		"googlechat - valid": {
			sType: "googlechat",
			urlFields: map[string]string{
				"raw": "bish"},
			errRegex: `^$`,
		},
		"googlechat - valid with main": {
			sType: "googlechat",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"raw": "bish"},
				nil),
			errRegex: `^$`,
		},
		"ifttt - no webhookid": {
			sType:     "ifttt",
			urlFields: map[string]string{},
			errRegex: test.TrimYAML(`
				^webhookid: <required>.*$`),
		},
		"ifttt - valid": {
			sType: "ifttt",
			urlFields: map[string]string{
				"webhookid": "bish"},
			errRegex: `^$`,
		},
		"ifttt - valid with main": {
			sType: "ifttt",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"webhookid": "webhookid"},
				map[string]string{
					"events": "events"}),
			errRegex: `^$`,
		},
		"join - no apikey": {
			sType:     "join",
			urlFields: map[string]string{},
			errRegex: test.TrimYAML(`
				^apikey: <required>.*$`),
		},
		"join - valid": {
			sType: "join",
			urlFields: map[string]string{
				"apikey": "bish"},
			errRegex: `^$`,
		},
		"join - valid with main": {
			sType: "join",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"apikey": "apikey"},
				nil),
			errRegex: `^$`,
		},
		"mattermost - invalid": {
			sType: "mattermost",
			errRegex: test.TrimYAML(`
				^host: <required>.*
				token: <required>.*$`),
		},
		"mattermost - no host": {
			sType: "mattermost",
			urlFields: map[string]string{
				"token": "bash"},
			errRegex: test.TrimYAML(`
				^host: <required>.*$`),
		},
		"mattermost - no token": {
			sType: "mattermost",
			urlFields: map[string]string{
				"host": "bish"},
			errRegex: test.TrimYAML(`
				^token: <required>.*$`),
		},
		"mattermost - valid": {
			sType: "mattermost",
			urlFields: map[string]string{
				"host":  "bish",
				"token": "bash"},
			errRegex: `^$`,
		},
		"mattermost - valid with main": {
			sType: "mattermost",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"host":  "bish",
					"token": "bash"},
				nil),
			errRegex: `^$`,
		},
		"matrix - invalid": {
			sType: "matrix",
			errRegex: test.TrimYAML(`
				^host: <required>.*
				password: <required>.*$`),
		},
		"matrix - no host": {
			sType: "matrix",
			urlFields: map[string]string{
				"password": "bash"},
			errRegex: test.TrimYAML(`
					^host: <required>.*$`),
		},
		"matrix - no password": {
			sType: "matrix",
			urlFields: map[string]string{
				"host": "bish"},
			errRegex: test.TrimYAML(`
					^password: <required>.*$`),
		},
		"matrix - valid": {
			sType: "matrix",
			urlFields: map[string]string{
				"host":     "bish",
				"password": "password"},
			errRegex: `^$`,
		},
		"matrix - valid with main": {
			sType: "matrix",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"host":     "bish",
					"password": "bash"},
				nil),
			errRegex: `^$`,
		},
		"ntfy - invalid": {
			sType: "ntfy",
			errRegex: test.TrimYAML(`
				^topic: <required>.*$`),
		},
		"ntfy - valid": {
			sType: "ntfy",
			urlFields: map[string]string{
				"topic": "foo"},
			errRegex: `^$`,
		},
		"opsgenie - invalid": {
			sType: "opsgenie",
			errRegex: test.TrimYAML(`
				^apikey: <required>.*$`),
		},
		"opsgenie - valid": {
			sType: "opsgenie",
			urlFields: map[string]string{
				"apikey": "apikey"},
			errRegex: `^$`,
		},
		"opsgenie - valid with main": {
			sType: "opsgenie",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"apikey": "apikey"},
				nil),
			errRegex: `^$`,
		},
		"pushbullet - invalid": {
			sType: "pushbullet",
			errRegex: test.TrimYAML(`
				^token: <required>.*
				targets: <required>.*$`),
		},
		"pushbullet - no token": {
			sType: "pushbullet",
			urlFields: map[string]string{
				"targets": "bash"},
			errRegex: test.TrimYAML(`
					^token: <required>.*$`),
		},
		"pushbullet - no targets": {
			sType: "pushbullet",
			urlFields: map[string]string{
				"token": "bish"},
			errRegex: test.TrimYAML(`
					^targets: <required>.*$`),
		},
		"pushbullet - valid": {
			sType: "pushbullet",
			urlFields: map[string]string{
				"token":   "bish",
				"targets": "targets"},
			errRegex: `^$`,
		},
		"pushbullet - valid with main": {
			sType: "pushbullet",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"token":   "bish",
					"targets": "bash"},
				nil),
			errRegex: `^$`,
		},
		"pushover - invalid": {
			sType: "pushover",
			errRegex: test.TrimYAML(`
				^token: <required>.*
				user: <required>.*$`),
		},
		"pushover - no token": {
			sType: "pushover",
			urlFields: map[string]string{
				"user": "bash"},
			errRegex: test.TrimYAML(`
					^token: <required>.*$`),
		},
		"pushover - no user": {
			sType: "pushover",
			urlFields: map[string]string{
				"token": "bish"},
			errRegex: test.TrimYAML(`
					^user: <required>.*$`),
		},
		"pushover - valid": {
			sType: "pushover",
			urlFields: map[string]string{
				"token": "bish",
				"user":  "user"},
			errRegex: `^$`,
		},
		"pushover - valid with main": {
			sType: "pushover",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"token": "bish",
					"user":  "bash"},
				nil),
			errRegex: `^$`,
		},
		"rocketchat - invalid": {
			sType: "rocketchat",
			errRegex: test.TrimYAML(`
				^host: <required>.*
				tokena: <required>.*
				tokenb: <required>.*
				channel: <required>.*$`),
		},
		"rocketchat - no host": {
			sType: "rocketchat",
			urlFields: map[string]string{
				"tokena":  "bash",
				"tokenb":  "bosh",
				"channel": "bing"},
			errRegex: test.TrimYAML(`
					^host: <required>.*$`),
		},
		"rocketchat - no tokena": {
			sType: "rocketchat",
			urlFields: map[string]string{
				"host":    "bish",
				"tokenb":  "bash",
				"channel": "bing"},
			errRegex: test.TrimYAML(`
					^tokena: <required>.*$`),
		},
		"rocketchat - no tokenb": {
			sType: "rocketchat",
			urlFields: map[string]string{
				"host":    "bish",
				"tokena":  "bash",
				"channel": "bing"},
			errRegex: test.TrimYAML(`
					^tokenb: <required>.*$`),
		},
		"rocketchat - no channel": {
			sType: "rocketchat",
			urlFields: map[string]string{
				"host":   "bish",
				"tokena": "bash",
				"tokenb": "bosh"},
			errRegex: test.TrimYAML(`
					^channel: <required>.*$`),
		},
		"rocketchat - valid": {
			sType: "rocketchat",
			urlFields: map[string]string{
				"host":    "bish",
				"tokena":  "bash",
				"tokenb":  "bosh",
				"channel": "bing"},
			errRegex: `^$`,
		},
		"rocketchat - valid with main": {
			sType: "rocketchat",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"host":    "bish",
					"tokena":  "bash",
					"tokenb":  "bosh",
					"channel": "bing"},
				nil),
			errRegex: `^$`,
		},
		"slack - invalid": {
			sType: "slack",
			errRegex: test.TrimYAML(`
				^token: <required>.*
				channel: <required>.*$`),
		},
		"slack - no token": {
			sType: "slack",
			urlFields: map[string]string{
				"channel": "bash"},
			errRegex: test.TrimYAML(`
					^token: <required>.*$`),
		},
		"slack - no channel": {
			sType: "slack",
			urlFields: map[string]string{
				"token": "bish"},
			errRegex: test.TrimYAML(`
					^channel: <required>.*$`),
		},
		"slack - valid": {
			sType: "slack",
			urlFields: map[string]string{
				"token":   "bish",
				"channel": "channel"},
			errRegex: `^$`,
		},
		"slack - valid with main": {
			sType: "slack",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"token":   "bish",
					"channel": "bash"},
				nil),
			errRegex: `^$`,
		},
		"teams - invalid": {
			sType: "teams",
			errRegex: test.TrimYAML(`
				^group: <required>.*
				tenant: <required>.*
				altid: <required>.*
				groupowner: <required>.*$`),
		},
		"teams - no group": {
			sType: "teams",
			urlFields: map[string]string{
				"tenant":     "bash",
				"altid":      "bosh",
				"groupowner": "bing"},
			errRegex: test.TrimYAML(`
					^group: <required>.*$`),
		},
		"teams - no tenant": {
			sType: "teams",
			urlFields: map[string]string{
				"group":      "bish",
				"altid":      "bash",
				"groupowner": "bing"},
			errRegex: test.TrimYAML(`
					^tenant: <required>.*$`),
		},
		"teams - no altid": {
			sType: "teams",
			urlFields: map[string]string{
				"group":      "bish",
				"tenant":     "bash",
				"groupowner": "bing"},
			errRegex: test.TrimYAML(`
					^altid: <required>.*$`),
		},
		"teams - no groupowner": {
			sType: "teams",
			urlFields: map[string]string{
				"group":  "bish",
				"tenant": "bash",
				"altid":  "bosh"},
			errRegex: test.TrimYAML(`
					^groupowner: <required>.*$`),
		},
		"teams - valid": {
			sType: "teams",
			urlFields: map[string]string{
				"group":      "bish",
				"tenant":     "bash",
				"altid":      "bosh",
				"groupowner": "bing"},
			errRegex: `^$`,
		},
		"teams - valid with main": {
			sType: "teams",
			main: NewDefaults(
				"",
				map[string]string{
					"host": "https://release-argus.io"},
				map[string]string{
					"group":      "bish",
					"tenant":     "bash",
					"altid":      "bosh",
					"groupowner": "bing"},
				nil),
			errRegex: `^$`,
		},
		"telegram - no token": {
			sType:     "telegram",
			urlFields: map[string]string{},
			errRegex: test.TrimYAML(`
				^token: <required>.*$`),
		},
		"telegram - valid": {
			sType: "telegram",
			urlFields: map[string]string{
				"token": "bish"},
			errRegex: `^$`,
		},
		"telegram - valid with main": {
			sType: "telegram",
			main: NewDefaults(
				"",
				nil,
				map[string]string{
					"token": "bish"},
				map[string]string{
					"chats": "chats"}),
			errRegex: `^$`,
		},
		"zulip - invalid": {
			sType: "zulip",
			errRegex: test.TrimYAML(`
				^host: <required>.*
				botmail: <required>.*
				botkey: <required>.*$`),
		},
		"zulip - no host": {
			sType: "zulip",
			urlFields: map[string]string{
				"botmail": "bash",
				"botkey":  "bosh"},
			errRegex: test.TrimYAML(`
					^host: <required>.*$`),
		},
		"zulip - no botmail": {
			sType: "zulip",
			urlFields: map[string]string{
				"host":   "bish",
				"botkey": "bash"},
			errRegex: test.TrimYAML(`
					^botmail: <required>.*$`),
		},
		"zulip - no botkey": {
			sType: "zulip",
			urlFields: map[string]string{
				"host":    "bish",
				"botmail": "bash"},
			errRegex: test.TrimYAML(`
					^botkey: <required>.*$`),
		},
		"zulip - valid": {
			sType: "zulip",
			urlFields: map[string]string{
				"host":    "bish",
				"botmail": "bash",
				"botkey":  "bosh"},
			errRegex: `^$`,
		},
		"zulip - valid with main": {
			sType: "zulip",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"host":    "bish",
					"botmail": "bash",
					"botkey":  "bosh"},
				nil),
			errRegex: `^$`,
		},
		"generic - invalid": {
			sType: "generic",
			errRegex: test.TrimYAML(`
				^host: <required>.*$`),
		},
		"generic - valid": {
			sType: "generic",
			urlFields: map[string]string{
				"host": "example.com"},
			errRegex: `^$`,
		},
		"generic - valid with main": {
			sType: "generic",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"host": "example.com"},
				nil),
			errRegex: `^$`,
		},
		"generic - valid with custom_headers/json_payload_vars/query_vars": {
			sType: "generic",
			urlFields: map[string]string{
				"host":              "example.com",
				"custom_headers":    `{"foo":"bar"}`,
				"json_payload_vars": `{"bish":"bash","bosh":"bing"}`,
				"query_vars":        `{"me":"work"}`},
			errRegex: `^$`,
		},
		"generic - invalid custom_headers": {
			sType: "generic",
			urlFields: map[string]string{
				"host":           "example.com",
				"custom_headers": `"foo":"bar"}`},
			errRegex: test.TrimYAML(`
				^custom_headers: "\\\"foo\\\":\\\"bar\\\"\}" <invalid>.*$`),
		},
		"generic - invalid json_payload_vars": {
			sType: "generic",
			urlFields: map[string]string{
				"host":              "example.com",
				"json_payload_vars": `{foo":"bar`},
			errRegex: test.TrimYAML(`
				^json_payload_vars: "\{foo\\\":\\\"bar" <invalid>.*$`),
		},
		"generic - invalid query_vars": {
			sType: "generic",
			urlFields: map[string]string{
				"host":       "example.com",
				"query_vars": `{foo:bar}`},
			errRegex: test.TrimYAML(`
				^query_vars: "\{foo:bar}" <invalid>.*$`),
		},
		"shoutrrr - invalid": {
			sType: "shoutrrr",
			errRegex: test.TrimYAML(`
				^raw: <required>.*$`),
		},
		"shoutrrr - valid": {
			sType: "shoutrrr",
			urlFields: map[string]string{
				"raw": "bish"},
			errRegex: `^$`,
		},
		"shoutrrr - valid with main": {
			sType: "shoutrrr",
			main: NewDefaults(
				"", nil,
				map[string]string{
					"raw": "bish"},
				nil),
			errRegex: `^$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.Type = tc.sType
			shoutrrr.URLFields = tc.urlFields
			svcStatus := status.New(
				nil, nil, nil,
				"", "", "", "", "", "")
			svcStatus.Init(
				1, 0, 0,
				test.StringPtr("serviceID"), nil)
			shoutrrr.Init(
				svcStatus,
				tc.main,
				&Defaults{}, &Defaults{})

			// WHEN checkValuesURLFields is called
			err := shoutrrr.checkValuesURLFields("")

			// THEN it errors when expected
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("Shoutrrr.checkValuesURLFields(), want error on url_fields to match:\n%q\ngot:\n%q\n",
					tc.errRegex, e)
			}
		})
	}
}

func TestShoutrrr_checkValuesParams(t *testing.T) {
	// GIVEN a Shoutrrr with Params
	tests := map[string]struct {
		sType    string
		params   map[string]string
		main     *Defaults
		errRegex string
	}{
		"no params": {
			params:   map[string]string{},
			errRegex: `^$`,
		},
		"valid message template": {
			params: map[string]string{
				"message": "release! {{ version }}"},
			errRegex: `^$`,
		},
		"invalid message template": {
			params: map[string]string{
				"a": "release! {{ version }"},
			errRegex: test.TrimYAML(`
				^a: "release! {{ version }" <invalid>.*$`),
		},
		"smtp - invalid": {
			sType: "smtp",
			errRegex: test.TrimYAML(`
				^fromaddress: <required>.*
				toaddresses: <required>.*$`),
		},
		"smtp - no fromaddress": {
			sType: "smtp",
			params: map[string]string{
				"toaddresses": "bosh"},
			errRegex: test.TrimYAML(`
				^fromaddress: <required>.*$`),
		},
		"smtp - no toaddresses": {
			sType: "smtp",
			params: map[string]string{
				"fromaddress": "bash"},
			errRegex: test.TrimYAML(`
				^toaddresses: <required>.*$`),
		},
		"smtp - valid": {
			sType: "smtp",
			params: map[string]string{
				"fromaddress": "bash",
				"toaddresses": "bosh"},
			errRegex: `^$`,
		},
		"smtp - valid with main": {
			sType: "smtp",
			params: map[string]string{
				"fromaddress": "bash"},
			main: NewDefaults(
				"",
				nil,
				nil,
				map[string]string{
					"toaddresses": "bosh"}),
			errRegex: `^$`,
		},
		"ifttt - no events": {
			sType: "ifttt",
			errRegex: test.TrimYAML(`
				^events: <required>.*$`),
		},
		"ifttt - valid": {
			sType: "ifttt",
			params: map[string]string{
				"events": "events"},
			errRegex: `^$`,
		},
		"ifttt - valid with main": {
			sType: "ifttt",
			main: NewDefaults(
				"",
				nil,
				nil,
				map[string]string{
					"events": "events"}),
			errRegex: `^$`,
		},
		"join - no devices": {
			sType:  "join",
			params: map[string]string{},
			errRegex: test.TrimYAML(`
				^devices: <required>.*$`),
		},
		"join - valid": {
			sType: "join",
			params: map[string]string{
				"devices": "foo,bar"},
			errRegex: `^$`,
		},
		"join - valid with main": {
			sType: "join",
			main: NewDefaults(
				"",
				nil,
				nil,
				map[string]string{
					"devices": "devices"}),
			errRegex: `^$`,
		},
		"teams - no host": {
			sType:  "teams",
			params: map[string]string{},
			errRegex: test.TrimYAML(`
				^host: <required>.*$`),
		},
		"teams - valid": {
			sType: "teams",
			params: map[string]string{
				"host": "https://release-argus.io"},
			errRegex: `^$`,
		},
		"teams - valid with main": {
			sType: "teams",
			main: NewDefaults(
				"",
				nil,
				nil,
				map[string]string{
					"host": "https://release-argus.io"}),
			errRegex: `^$`,
		},
		"telegram - no chats": {
			sType: "telegram",
			errRegex: test.TrimYAML(`
				^chats: <required>.*$`),
		},
		"telegram - valid": {
			sType: "telegram",
			params: map[string]string{
				"chats": "chats"},
			errRegex: `^$`,
		},
		"telegram - valid with main": {
			sType: "telegram",
			main: NewDefaults(
				"",
				nil,
				nil,
				map[string]string{
					"chats": "chats"}),
			errRegex: `^$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.Type = tc.sType
			shoutrrr.Params = tc.params
			svcStatus := status.New(
				nil, nil, nil,
				"", "", "", "", "", "")
			svcStatus.Init(
				1, 0, 0,
				test.StringPtr("serviceID"), nil)
			shoutrrr.Init(
				svcStatus,
				tc.main,
				&Defaults{}, &Defaults{})

			// WHEN checkValuesParams is called
			err := shoutrrr.checkValuesParams("")

			// THEN it errors when expected
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("Shoutrrr.checkValuesParams(), want error on params to match:\n%q\ngot:\n%q\n",
					tc.errRegex, e)
			}
		})
	}
}

func TestShoutrrr_CorrectSelf(t *testing.T) {
	// GIVEN a Service
	tests := map[string]struct {
		sType         string
		mapTarget     string
		startAs, want map[string]string
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
	subTests := []string{
		"root", "main", "defaults", "hard_defaults"}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := New(
				nil, "",
				tc.sType,
				make(map[string]string), make(map[string]string), make(map[string]string),
				NewDefaults("",
					make(map[string]string), make(map[string]string), make(map[string]string)),
				NewDefaults("",
					make(map[string]string), make(map[string]string), make(map[string]string)),
				NewDefaults("",
					make(map[string]string), make(map[string]string), make(map[string]string)))
			serviceStatus := status.Status{}
			serviceStatus.Init(1, 0, 0, &name, nil)
			shoutrrr.Init(
				&serviceStatus,
				shoutrrr.Main, shoutrrr.Defaults, shoutrrr.HardDefaults)
			var subTestMap = map[string]struct {
				URLFields map[string]string
				Params    map[string]string
			}{
				"root": {
					URLFields: shoutrrr.URLFields,
					Params:    shoutrrr.Params},
				"main": {
					URLFields: shoutrrr.Main.URLFields,
					Params:    shoutrrr.Main.Params},
				"defaults": {
					URLFields: shoutrrr.Defaults.URLFields,
					Params:    shoutrrr.Defaults.Params},
				"hard_defaults": {
					URLFields: shoutrrr.HardDefaults.URLFields,
					Params:    shoutrrr.HardDefaults.Params},
			}
			// sub tests - set in different locations and check its corrected there
			for sub_test := range subTestMap {
				t.Logf("sub_test: %s", sub_test)
				if tc.mapTarget == "url_fields" {
					for k, v := range tc.startAs {
						subTestMap[sub_test].URLFields[k] = v
					}
				} else {
					for k, v := range tc.startAs {
						subTestMap[sub_test].Params[k] = v
					}
				}

				// WHEN correctSelf is called
				shoutrrr.correctSelf(shoutrrr.GetType())

				// THEN the fields are corrected as necessary
				for k, v := range tc.want {
					want := v
					// root is the only one that gets corrected
					if sub_test != "root" {
						want = tc.startAs[k]
					}
					got := shoutrrr.GetURLField(k)
					if tc.mapTarget != "url_fields" {
						got = shoutrrr.GetParam(k)
					}
					if got != want {
						t.Errorf("want %s:%q, not %q",
							k, want, got)
					} else {
						for _, sub_test_check := range subTests {
							if sub_test_check != sub_test {
								testData := subTestMap[sub_test_check]
								if len(testData.URLFields) > 0 || len(testData.Params) > 0 {
									t.Errorf("want empty %s, not %v/%v",
										sub_test_check, testData.URLFields, testData.Params)
								}
							}
						}
					}
					// reset
					if tc.mapTarget == "url_fields" {
						delete(subTestMap[sub_test].URLFields, k)
					} else {
						delete(subTestMap[sub_test].Params, k)
					}
				}
			}
		})
	}
}

func TestShoutrrr_CheckValues(t *testing.T) {
	// GIVEN a Shoutrrr
	testS := testShoutrrr(false, false)
	tests := map[string]struct {
		nilShoutrrr                bool
		sType                      string
		options, urlFields, params map[string]string
		wantURLFields              map[string]string
		wantDelay                  string
		main                       *Defaults
		errRegex                   string
	}{
		"nil shoutrrr": {
			nilShoutrrr: true,
			errRegex:    `^$`,
		},
		"empty": {
			errRegex:  `^type: <required>[^:]+://[^:]+$`,
			urlFields: map[string]string{},
			params:    map[string]string{},
		},
		"invalid delay": {
			errRegex: test.TrimYAML(`
				^options:
					delay: .* <invalid>`),
			sType:     testS.Type,
			urlFields: testS.URLFields,
			options: map[string]string{
				"delay": "5x"},
		},
		"fixes delay": {
			errRegex:  `^$`,
			sType:     testS.Type,
			urlFields: testS.URLFields,
			wantDelay: "5s",
			options: map[string]string{
				"delay": "5"},
		},
		"invalid message template": {
			errRegex: test.TrimYAML(`
				^options:
					message: ".+" <invalid>.*$`),
			sType:     testS.Type,
			urlFields: testS.URLFields,
			options: map[string]string{
				"message": "{{ version }"},
		},
		"valid message template": {
			errRegex:  `^$`,
			sType:     testS.Type,
			urlFields: testS.URLFields,
			options: map[string]string{
				"message": "{{ version }}"},
		},
		"invalid title template": {
			errRegex: test.TrimYAML(`
				^params:
					title: .* <invalid>.*$`),
			sType:     testS.Type,
			urlFields: testS.URLFields,
			params: map[string]string{
				"title": "{{ version }"},
		},
		"valid title template": {
			errRegex:  `^$`,
			sType:     testS.Type,
			urlFields: testS.URLFields,
			params: map[string]string{
				"title": "{{ version }}"},
		},
		"valid param template": {
			errRegex:  `^$`,
			sType:     testS.Type,
			urlFields: testS.URLFields,
			params: map[string]string{
				"foo": "{{ version }}"},
		},
		"invalid param template": {
			errRegex: test.TrimYAML(`
				^params:
					foo: .* <invalid>.*$`),
			sType:     testS.Type,
			urlFields: testS.URLFields,
			params: map[string]string{
				"foo": "{{ version }"},
		},
		"invalid param and option": {
			errRegex: test.TrimYAML(`
				^options:
					delay: [^<]+<invalid>.*
				params:
					title: [^<]+<invalid>.*$`),
			sType:     testS.Type,
			urlFields: testS.URLFields,
			params: map[string]string{
				"title": "{{ version }"},
			options: map[string]string{
				"delay": "2x"},
		},
		"does correctSelf": {
			errRegex: `^$`,
			sType:    testS.Type,
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
			errRegex:  `^$`,
			urlFields: map[string]string{},
			main:      testDefaults(false, false),
		},
		"valid with self and main": {
			errRegex: `^$`,
			urlFields: map[string]string{
				"host": "foo"},
			main: NewDefaults(
				testS.Type,
				nil,
				map[string]string{
					"token": "bar"},
				nil),
		},
		"invalid url_fields": {
			errRegex: test.TrimYAML(`
				^url_fields:
					host: <required>.*
					token: <required>.*$`),
			sType: testS.Type,
		},
		"invalid params + locate fail": {
			errRegex: test.TrimYAML(`
				^params:
					fromaddress: <required>.*
					toaddresses: <required>.*$`),
			urlFields: map[string]string{
				"host": "https://release-argus.io"},
			sType: "smtp",
		},
		"gotify - fail CreateSender": {
			sType: "gotify",
			urlFields: map[string]string{
				"host":  "https://	example.com",
				"token": "bish"},
			errRegex: `invalid control character in URL`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.Type = tc.sType
			if tc.main == nil {
				tc.main = &Defaults{}
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

			// THEN it errors when expected
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
			if tc.nilShoutrrr {
				return
			}
			if tc.wantDelay != "" && shoutrrr.GetOption("delay") != tc.wantDelay {
				t.Errorf("delay not set/corrected. want match for %q\nnot: %q",
					tc.wantDelay, shoutrrr.GetOption("delay"))
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
			slice: nil, errRegex: `^$`},
		"valid slice": {
			errRegex: `^$`,
			slice: &Slice{
				"valid": testShoutrrr(false, false),
				"other": testShoutrrr(false, false)}},
		"invalid slice": {
			errRegex: test.TrimYAML(`
				other:
					type: <required>`),
			slice: &Slice{
				"valid": testShoutrrr(false, false),
				"other": New(
					nil, "", "",
					make(map[string]string), make(map[string]string), make(map[string]string),
					nil, nil, nil)}},
		"ordered errors": {
			errRegex: test.TrimYAML(`
				aNotify:
					type: <required>.*
				bNotify:
					type: <required>.*`),
			slice: &Slice{
				"aNotify": New(
					nil, "", "",
					make(map[string]string), make(map[string]string), make(map[string]string),
					nil, nil, nil),
				"bNotify": New(
					nil, "", "",
					make(map[string]string), make(map[string]string), make(map[string]string),
					nil, nil, nil)}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.slice != nil {
				svcStatus := &status.Status{}
				svcStatus.Init(
					len(*tc.slice), 0, 0, nil, nil)
				tc.slice.Init(
					svcStatus,
					&SliceDefaults{}, &SliceDefaults{}, &SliceDefaults{})
			}

			// WHEN CheckValues is called
			err := tc.slice.CheckValues("")

			// THEN it errors when expected
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
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
			slice:    nil,
			errRegex: `^$`,
		},
		"valid slice": {
			errRegex: `^$`,
			slice: &SliceDefaults{
				"valid": testDefaults(false, false),
				"other": testDefaults(false, false)},
		},
		"invalid type": {
			errRegex: "", // Caught by Shoutrrr.CheckValues
			slice: &SliceDefaults{
				"valid": testDefaults(false, false),
				"other": NewDefaults(
					"somethingUnknown",
					make(map[string]string), make(map[string]string), make(map[string]string))},
		},
		"delay without unit": {
			errRegex: `^$`,
			slice: &SliceDefaults{
				"foo": NewDefaults(
					"gotify",
					map[string]string{
						"delay": "1"},
					nil, nil)},
		},
		"invalid delay": {
			errRegex: test.TrimYAML(`
				^foo:
					options:
						delay: "1x" <invalid>`),
			slice: &SliceDefaults{
				"foo": NewDefaults(
					"gotify",
					map[string]string{
						"delay": "1x"},
					nil, nil)},
		},
		"invalid message template": {
			errRegex: test.TrimYAML(`
				^bar:
					options:
						message: "[^"]+" <invalid>.*$`),
			slice: &SliceDefaults{
				"bar": NewDefaults(
					"gotify",
					map[string]string{
						"message": "{{ .foo }"},
					nil, nil)},
		},
		"invalid params template": {
			errRegex: test.TrimYAML(`
				^bar:
					params:
						title: "[^"]+" <invalid>.*$`),
			slice: &SliceDefaults{
				"bar": NewDefaults(
					"gotify",
					nil,
					nil,
					map[string]string{
						"title": "{{ .bar }"})},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CheckValues is called
			err := tc.slice.CheckValues("")

			// THEN it errors when expected
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("ShoutrrrSlice.CheckValues() mismatch\n%q\ngot:\n%q",
					tc.errRegex, e)
			}
		})
	}
}

func TestShoutrrr_TestSend(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		sType       *string
		nilShoutrrr bool
		wantErr     bool
	}{
		"nil shoutrrr": {
			nilShoutrrr: true, wantErr: true},
		"invalid type": {
			sType: test.StringPtr("somethingUnknown"), wantErr: true},
		"valid": {
			wantErr: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			if tc.sType != nil {
				shoutrrr.Type = *tc.sType
			}
			if tc.nilShoutrrr {
				shoutrrr = nil
			}

			// WHEN TestSend is called
			err := shoutrrr.TestSend("https://example.com")

			// THEN it errors when expected
			if tc.wantErr && err == nil {
				t.Errorf("want err, not nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("want nil, not err: %v", err)
			}
		})
	}

}

func TestSliceDefaults_Print(t *testing.T) {
	// GIVEN a SliceDefaults
	testValid := testDefaults(false, false)
	testInvalid := testDefaults(true, true)
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
			want: test.TrimYAML(`
				notify:
					single: {}`),
		},
		"single element slice": {
			slice: &SliceDefaults{
				"single": testValid},
			want: test.TrimYAML(`
				notify:
					single:
						type: gotify
						options:
							max_tries: "` + testValid.GetOption("max_tries") + `"
						url_fields:
							host: ` + testValid.GetURLField("host") + `
							path: ` + testValid.GetURLField("path") + `
							token: ` + testValid.GetURLField("token")),
		},
		"multiple element slice": {
			slice: &SliceDefaults{
				"first":  testValid,
				"second": testInvalid},
			want: test.TrimYAML(`
				notify:
					first:
						type: gotify
						options:
							max_tries: "` + testValid.GetOption("max_tries") + `"
						url_fields:
							host: ` + testValid.GetURLField("host") + `
							path: ` + testValid.GetURLField("path") + `
							token: ` + testValid.GetURLField("token") + `
					second:
						type: gotify
						options:
							max_tries: "` + testInvalid.GetOption("max_tries") + `"
						url_fields:
							host: ` + testInvalid.GetURLField("host") + `
							path: ` + testInvalid.GetURLField("path") + `
							token: ` + testInvalid.GetURLField("token")),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout
			releaseStdout := test.CaptureStdout()

			if tc.want != "" {
				tc.want += "\n"
			}

			// WHEN Print is called
			tc.slice.Print("")

			// THEN it prints the expected stdout
			stdout := releaseStdout()
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if stdout != tc.want {
				t.Errorf("Print should have given\n%q\nbut gave\n%q",
					tc.want, stdout)
			}
		})
	}
}
