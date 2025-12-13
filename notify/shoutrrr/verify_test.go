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

package shoutrrr

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestShoutrrr_checkValuesType(t *testing.T) {
	// GIVEN a Shoutrrr with a Type and possibly a Main.
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
				"",
				"", "",
				"", "",
				"",
				&dashboard.Options{})
			svcStatus.Init(
				1, 0, 0,
				name, "", "",
				&dashboard.Options{})
			shoutrrr.Init(
				svcStatus,
				tc.main,
				&Defaults{}, &Defaults{})

			// WHEN checkValuesType is called.
			errsType := shoutrrr.checkValuesType("")

			// THEN it errors when expected.
			if tc.errsRegex == "" {
				tc.errsRegex = "^$"
			}
			e := util.ErrorToString(errsType)
			if !util.RegexCheck(tc.errsRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant:  %q\ngot:  %q",
					packageName, tc.errsRegex, e)
			}
		})
	}
}

func TestBase_checkValuesOptions(t *testing.T) {
	// GIVEN a Base with Options.
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

			// WHEN checkValuesOptions is called.
			err := shoutrrr.checkValuesOptions("")

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q\n",
					packageName, tc.errRegex, e)
			}
			// AND the delay is set as expected if it didn't error on delay.
			if !util.RegexCheck("^delay:.*", e) &&
				shoutrrr.GetOption("delay") != tc.wantDelay {
				t.Errorf("%s\ndelay mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wantDelay, shoutrrr.GetOption("delay"))
			}
		})
	}
}

func TestBase_checkValuesParams(t *testing.T) {
	// GIVEN a Base with Params.
	tests := map[string]struct {
		itemType string
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
		"valid select param": {
			itemType: "smtp",
			params: map[string]string{
				"auth": "OAuth2"},
			errRegex: `^$`,
		},
		"invalid select param": {
			itemType: "smtp",
			params: map[string]string{
				"auth": "-"},
			errRegex: `^auth: "-" <invalid>.*OAuth2.*$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := Base{}
			shoutrrr.Params = tc.params

			// WHEN checkValuesParams is called.
			err := shoutrrr.checkValuesParams("", tc.itemType)

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q\n",
					packageName, tc.errRegex, e)
			}
		})
	}
}

func TestBase_normaliseParamSelect(t *testing.T) {
	// GIVEN a Base and various inputs to normaliseParamSelect.
	tests := map[string]struct {
		value      string
		allowed    []string
		startParam string
		wantOK     bool
		wantValue  string
	}{
		"exact match uses canonical case": {
			value:     "Two",
			allowed:   []string{"One", "Two", "Three"},
			wantOK:    true,
			wantValue: "Two",
		},
		"case-insensitive match lower->upper": {
			value:     "two",
			allowed:   []string{"One", "Two", "Three"},
			wantOK:    true,
			wantValue: "Two",
		},
		"case-insensitive match upper->proper": {
			value:     "THREE",
			allowed:   []string{"One", "Two", "Three"},
			wantOK:    true,
			wantValue: "Three",
		},
		"non-match returns false and leaves value unchanged": {
			value:      "four",
			allowed:    []string{"One", "Two", "Three"},
			startParam: "unchanged",
			wantOK:     false,
			wantValue:  "unchanged",
		},
		"empty value returns false and makes no change": {
			value:     "",
			allowed:   []string{"One", "Two"},
			wantOK:    false,
			wantValue: "",
		},
		"empty allowed set never matches": {
			value:      "one",
			allowed:    []string{},
			startParam: "keepme",
			wantOK:     false,
			wantValue:  "keepme",
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			b := &Base{}
			b.InitMaps()
			if tc.startParam != "" {
				b.SetParam(name, tc.startParam)
			}

			// WHEN normaliseParamSelect is called.
			ok := b.normaliseParamSelect(name, tc.value, tc.allowed)

			// THEN it returns the expected boolean.
			if ok != tc.wantOK {
				t.Fatalf("%s\nnormaliseParamSelect() ok mismatch\nwant: %t\ngot:  %t",
					packageName, tc.wantOK, ok)
			}

			// AND Params[key] is set/unchanged as expected.
			got := b.GetParam(name)
			if got != tc.wantValue {
				t.Fatalf("%s\nnormaliseParamSelect() Params[x] mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wantValue, got)
			}
		})
	}
}

func TestBase_validateParamSelect(t *testing.T) {
	key := "test"
	// GIVEN a Base and various inputs to validateParamSelect.
	tests := map[string]struct {
		value     string
		allowed   []string
		wantErr   string
		wantValue string
	}{
		"empty value returns nil": {
			value:     "",
			allowed:   []string{"low", "default", "high"},
			wantErr:   `^$`,
			wantValue: "",
		},
		"valid value normalises and returns nil": {
			value:     "HiGh",
			allowed:   []string{"min", "low", "default", "high", "max"},
			wantErr:   `^$`,
			wantValue: "high",
		},
		"invalid value returns error and leaves unchanged": {
			value:     "nope",
			allowed:   []string{"None", "Unknown", "Plain", "CramMD5", "OAuth2"},
			wantErr:   "^" + key + `: "nope" <invalid> .*'CramMD5'.*$`,
			wantValue: "nope",
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			b := &Base{}
			b.InitMaps()
			b.SetParam(key, tc.value)

			// WHEN validateParamSelect is called.
			err := b.validateParamSelect("", key, tc.allowed)

			// THEN error matches expectation.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.wantErr, e) {
				t.Fatalf("%s\nvalidateParamSelect() error mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wantErr, e)
			}

			// AND Params[key] is set/unchanged as expected.
			got := b.GetParam(key)
			if got != tc.wantValue {
				t.Fatalf("%s\nvalidateParamSelect() Params[x] mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wantValue, got)
			}
		})
	}
}

func TestBase_checkValuesParamsSelects(t *testing.T) {
	// GIVEN a Base with Params and different item types.
	tests := map[string]struct {
		itemType  string
		params    map[string]string
		wantErr   string
		wantParam map[string]string
	}{
		// bark
		"bark - valid scheme+sound normalised": {
			itemType: "bark",
			params: map[string]string{
				"scheme": "HTTP",
				"sound":  "BELL",
			},
			wantErr: `^$`,
			wantParam: map[string]string{
				"scheme": "http",
				"sound":  "bell",
			},
		},
		"bark - invalid scheme and sound aggregated": {
			itemType: "bark",
			params: map[string]string{
				"scheme": "-",
				"sound":  "nope",
			},
			wantErr: test.TrimYAML(`
                ^scheme: "-" <invalid> \(supported = \['http', 'https'\]\)
                sound: "nope" <invalid> \(supported = \['alarm', 'anticipate', 'bell'.*\)$`),
		},
		// generic
		"generic - valid requestmethod normalised": {
			itemType: "generic",
			params: map[string]string{
				"requestmethod": "post",
			},
			wantErr: `^$`,
			wantParam: map[string]string{
				"requestmethod": "POST",
			},
		},
		"generic - invalid requestmethod": {
			itemType: "generic",
			params: map[string]string{
				"requestmethod": "FETCH",
			},
			wantErr: `^requestmethod: "FETCH" <invalid> \(supported = \['CONNECT', 'DELETE', 'GET', 'HEAD', 'OPTIONS', 'POST', 'PUT', 'TRACE'\]\)$`,
		},
		// ntfy
		"ntfy - valid priority and scheme": {
			itemType: "ntfy",
			params: map[string]string{
				"priority": "DEFAULT",
				"scheme":   "HTTPS",
			},
			wantErr: `^$`,
			wantParam: map[string]string{
				"priority": "default",
				"scheme":   "https",
			},
		},
		"ntfy - invalid priority and scheme aggregated": {
			itemType: "ntfy",
			params: map[string]string{
				"priority": "urgENT",
				"scheme":   "ftp",
			},
			wantErr: test.TrimYAML(`
                ^priority: "urgENT" <invalid> \(supported = \['min', 'low', 'default', 'high', 'max'\]\)
                scheme: "ftp" <invalid> \(supported = \['http', 'https'\]\)$`),
		},
		// smtp
		"smtp - valid auth and encryption": {
			itemType: "smtp",
			params: map[string]string{
				"auth":       "oauth2",
				"encryption": "explicittls",
			},
			wantErr: `^$`,
			wantParam: map[string]string{
				"auth":       "OAuth2",
				"encryption": "ExplicitTLS",
			},
		},
		"smtp - invalid auth and encryption aggregated": {
			itemType: "smtp",
			params: map[string]string{
				"auth":       "basic",
				"encryption": "tls1.3",
			},
			wantErr: test.TrimYAML(`
                ^auth: "basic" <invalid> \(supported = \['None', 'Unknown', 'Plain', 'CramMD5', 'OAuth2'\]\)
                encryption: "tls1.3" <invalid> \(supported = \['Auto', 'ExplicitTLS', 'ImplicitTLS', 'None'\]\)$`),
		},
		// telegram
		"telegram - valid parsemode": {
			itemType: "telegram",
			params: map[string]string{
				"parsemode": "markdown",
			},
			wantErr: `^$`,
			wantParam: map[string]string{
				"parsemode": "Markdown",
			},
		},
		"telegram - invalid parsemode": {
			itemType: "telegram",
			params: map[string]string{
				"parsemode": "mdx",
			},
			wantErr: `^parsemode: "mdx" <invalid> \(supported = \['None', 'HTML', 'Markdown'\]\)$`,
		},
		// unknown type and empty params
		"unknown - type is no-op even with values": {
			itemType: "unknown",
			params: map[string]string{
				"scheme": "ftp",
			},
			wantErr:   `^$`,
			wantParam: map[string]string{},
		},
		"nil/empty params produces no error": {
			itemType: "smtp",
			params:   map[string]string{},
			wantErr:  `^$`,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			b := &Base{}
			b.InitMaps()
			for k, v := range tc.params {
				b.SetParam(k, v)
			}

			// WHEN checkValuesParamsSelects is called.
			err := b.checkValuesParamsSelects("", tc.itemType)

			// THEN error matches expectation.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.wantErr, e) {
				t.Fatalf("%s\ncheckValuesParamsSelects() error mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wantErr, e)
			}

			// AND any expected normalisations took place.
			for k, v := range tc.wantParam {
				got := b.GetParam(k)
				if got != v {
					t.Fatalf("%s\ncheckValuesParamsSelects() normalisation mismatch for %q\nwant: %q\ngot:  %q",
						packageName, k, v, got)
				}
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

			// WHEN CheckValues is called.
			err := tc.base.CheckValues("", tc.id)

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			lines := strings.Split(e, "\n")
			wantLines := strings.Count(tc.errRegex, "\n")
			if wantLines > len(lines) {
				t.Fatalf("%s\nwant: %d lines of error:\n%q\ngot:  %d lines:\n%v\n\nstdout: %q",
					packageName, wantLines, tc.errRegex, len(lines), lines, e)
				return
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
				return
			}
			// AND the Base is as expected.
			wantStr := util.ToYAMLString(tc.want, "")
			gotStr := util.ToYAMLString(tc.base, "")
			if gotStr != wantStr {
				t.Errorf("%s\nstringified mismatch\nwant:\n%q\ngot:\n%q",
					packageName, wantStr, gotStr)
			}
		})
	}
}

func TestShoutrrr_checkValuesURLFields(t *testing.T) {
	// GIVEN a Shoutrrr with Params.
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
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.Type = tc.sType
			shoutrrr.URLFields = tc.urlFields
			svcStatus := status.New(
				nil, nil, nil,
				"",
				"", "",
				"", "",
				"",
				&dashboard.Options{})
			svcStatus.Init(
				1, 0, 0,
				"serviceID", "", "",
				svcStatus.Dashboard)
			shoutrrr.Init(
				svcStatus,
				tc.main,
				&Defaults{}, &Defaults{})

			// WHEN checkValuesURLFields is called.
			err := shoutrrr.checkValuesURLFields("")

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
		})
	}
}

func TestShoutrrr_checkValuesParams(t *testing.T) {
	// GIVEN a Shoutrrr with Params.
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
				"",
				"", "",
				"", "",
				"",
				&dashboard.Options{})
			svcStatus.Init(
				1, 0, 0,
				"serviceID", "", "",
				svcStatus.Dashboard)
			shoutrrr.Init(
				svcStatus,
				tc.main,
				&Defaults{}, &Defaults{})

			// WHEN checkValuesParams is called.
			err := shoutrrr.checkValuesParams("")

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q\n",
					packageName, tc.errRegex, e)
			}
		})
	}
}

func TestShoutrrr_CorrectSelf(t *testing.T) {
	// GIVEN a Service.
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
			serviceStatus.Init(
				1, 0, 0,
				name, "", "",
				&dashboard.Options{})
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
			// Sub tests - set in different locations and check it's corrected there.
			for subTest := range subTestMap {
				t.Logf("%s - sub_test: %s",
					packageName, subTest)
				if tc.mapTarget == "url_fields" {
					for k, v := range tc.startAs {
						subTestMap[subTest].URLFields[k] = v
					}
				} else {
					for k, v := range tc.startAs {
						subTestMap[subTest].Params[k] = v
					}
				}

				// WHEN correctSelf is called.
				shoutrrr.correctSelf(shoutrrr.GetType())

				// THEN the fields are corrected as necessary.
				for k, v := range tc.want {
					want := v
					// root is the only one that gets corrected.
					if subTest != "root" {
						want = tc.startAs[k]
					}
					got := shoutrrr.GetURLField(k)
					if tc.mapTarget != "url_fields" {
						got = shoutrrr.GetParam(k)
					}
					if got != want {
						t.Errorf("%s\nwant %s:%q, not %q",
							packageName, k,
							want, got)
					} else {
						for _, subTestCheck := range subTests {
							if subTestCheck != subTest {
								testData := subTestMap[subTestCheck]
								if len(testData.URLFields) > 0 || len(testData.Params) > 0 {
									t.Errorf("%s\nmismatch\nwant: empty %s\ngot:  url_fields=%+v / params=%+v",
										packageName, subTestCheck,
										testData.URLFields, testData.Params)
								}
							}
						}
					}
					// Reset.
					if tc.mapTarget == "url_fields" {
						delete(subTestMap[subTest].URLFields, k)
					} else {
						delete(subTestMap[subTest].Params, k)
					}
				}
			}
		})
	}
}

func TestShoutrrr_CheckValues(t *testing.T) {
	// GIVEN a Shoutrrr.
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
			errRegex: `failed to parse URL`,
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

			// WHEN CheckValues is called.
			err := shoutrrr.CheckValues("")

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
			if tc.nilShoutrrr {
				return
			}
			if tc.wantDelay != "" && shoutrrr.GetOption("delay") != tc.wantDelay {
				t.Errorf("%s\ndelay not set/corrected\nwant: %q\ngot:  %q",
					packageName, tc.wantDelay, shoutrrr.GetOption("delay"))
			}
			for key := range tc.wantURLFields {
				if shoutrrr.URLFields[key] != tc.wantURLFields[key] {
					t.Errorf("%s\nmismatch on %q\nwant: %q (%v)\ngot:  %q (%v)",
						packageName, key,
						tc.wantURLFields[key], tc.wantURLFields,
						shoutrrr.URLFields[key], shoutrrr.URLFields)
				}
			}
		})
	}
}

func TestShoutrrrs_CheckValues(t *testing.T) {
	// GIVEN Shoutrrrs.
	tests := map[string]struct {
		shoutrrrs *Shoutrrrs
		errRegex  string
	}{
		"nil map": {
			shoutrrrs: nil, errRegex: `^$`},
		"valid map": {
			errRegex: `^$`,
			shoutrrrs: &Shoutrrrs{
				"valid": testShoutrrr(false, false),
				"other": testShoutrrr(false, false)}},
		"invalid map": {
			errRegex: test.TrimYAML(`
				other:
					type: <required>`),
			shoutrrrs: &Shoutrrrs{
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
			shoutrrrs: &Shoutrrrs{
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

			if tc.shoutrrrs != nil {
				svcStatus := &status.Status{}
				svcStatus.Init(
					len(*tc.shoutrrrs), 0, 0,
					"", "", "",
					&dashboard.Options{})
				tc.shoutrrrs.Init(
					svcStatus,
					&ShoutrrrsDefaults{}, &ShoutrrrsDefaults{}, &ShoutrrrsDefaults{})
			}

			// WHEN CheckValues is called.
			err := tc.shoutrrrs.CheckValues("")

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
		})
	}
}

func TestShoutrrrsDefaults_CheckValues(t *testing.T) {
	// GIVEN ShoutrrrsDefaults.
	tests := map[string]struct {
		shoutrrrsDefaults *ShoutrrrsDefaults
		errRegex          string
	}{
		"nil map": {
			shoutrrrsDefaults: nil,
			errRegex:          `^$`,
		},
		"valid map": {
			errRegex: `^$`,
			shoutrrrsDefaults: &ShoutrrrsDefaults{
				"valid": testDefaults(false, false),
				"other": testDefaults(false, false)},
		},
		"invalid type": {
			errRegex: "", // Caught by Shoutrrr.CheckValues.
			shoutrrrsDefaults: &ShoutrrrsDefaults{
				"valid": testDefaults(false, false),
				"other": NewDefaults(
					"somethingUnknown",
					make(map[string]string), make(map[string]string), make(map[string]string))},
		},
		"delay without unit": {
			errRegex: `^$`,
			shoutrrrsDefaults: &ShoutrrrsDefaults{
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
			shoutrrrsDefaults: &ShoutrrrsDefaults{
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
			shoutrrrsDefaults: &ShoutrrrsDefaults{
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
			shoutrrrsDefaults: &ShoutrrrsDefaults{
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

			// WHEN CheckValues is called.
			err := tc.shoutrrrsDefaults.CheckValues("")

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
		})
	}
}

func TestDefaults_CheckValues(t *testing.T) {
	// GIVEN Defaults and various ids.
	tests := map[string]struct {
		d        *Defaults
		id       string
		errRegex string
	}{
		"nil defaults - valid id": {
			d:        nil,
			id:       "slack",
			errRegex: `^$`,
		},
		"nil defaults - invalid id": {
			d:        nil,
			id:       "argus",
			errRegex: `^type: "argus" <invalid>.*gotify.*$`,
		},
		"empty Type uses id - valid": {
			d:        &Defaults{Base: Base{}},
			id:       "gotify",
			errRegex: `^$`,
		},
		"empty Type uses id - invalid": {
			d:        &Defaults{Base: Base{}},
			id:       "unknown",
			errRegex: `^type: "unknown" <invalid>.*$`,
		},
		"Type set overrides id (both valid)": {
			d:        &Defaults{Base: Base{Type: "gotify"}},
			id:       "slack",
			errRegex: `^$`,
		},
		"Base error is propagated": {
			d: &Defaults{Base: Base{
				Type:    "slack",
				Options: map[string]string{"delay": "10x"},
			}},
			id: "",
			errRegex: test.TrimYAML(`
                ^options:
                  delay: "10x" <invalid>.*$`),
		},
		"Combines type invalid and base error": {
			d: &Defaults{Base: Base{
				Type:   "invalid",
				Params: map[string]string{"color": "{{ invalid template }}"},
			}},
			id: "",
			errRegex: test.TrimYAML(`
                ^type: "invalid" <invalid>.*
                params:
                  color: "{{ invalid template }}" <invalid>.*$`),
		},
		"both valid - no error": {
			d: &Defaults{Base: Base{
				Type: "gotify",
				Options: map[string]string{
					"delay": "1s"},
				Params: map[string]string{
					"message": "release {{ version }}"},
			}},
			id:       "",
			errRegex: `^$`,
		},
		"empty Type and id empty": {
			d:        &Defaults{Base: Base{}},
			id:       "",
			errRegex: `^type: "" <invalid>.*$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CheckValues is called.
			err := tc.d.CheckValues("", tc.id)

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
		})
	}
}

func TestShoutrrr_TestSend(t *testing.T) {
	// GIVEN a Shoutrrr.
	tests := map[string]struct {
		sType       *string
		nilShoutrrr bool
		errRegex    string
	}{
		"nil shoutrrr": {
			nilShoutrrr: true, errRegex: `^shoutrrr is nil$`},
		"invalid type": {
			sType:    test.StringPtr("somethingUnknown"),
			errRegex: `^failed to create Shoutrrr sender.*unknown service: ""$`},
		"valid": {
			errRegex: `^$`},
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

			// WHEN TestSend is called.
			err := shoutrrr.TestSend("https://example.com")

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
		})
	}
}

func TestShoutrrrsDefaults_Print(t *testing.T) {
	// GIVEN a ShoutrrrsDefaults.
	testValid := testDefaults(false, false)
	testInvalid := testDefaults(true, true)
	tests := map[string]struct {
		shoutrrrsDefaults *ShoutrrrsDefaults
		want              string
	}{
		"nil": {
			shoutrrrsDefaults: nil,
			want:              "",
		},
		"empty": {
			shoutrrrsDefaults: &ShoutrrrsDefaults{},
			want:              "",
		},
		"single empty element map": {
			shoutrrrsDefaults: &ShoutrrrsDefaults{
				"single": {}},
			want: test.TrimYAML(`
				notify:
					single: {}`),
		},
		"single element map": {
			shoutrrrsDefaults: &ShoutrrrsDefaults{
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
		"multiple element map": {
			shoutrrrsDefaults: &ShoutrrrsDefaults{
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
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout()

			if tc.want != "" {
				tc.want += "\n"
			}

			// WHEN Print is called.
			tc.shoutrrrsDefaults.Print("")

			// THEN it prints the expected stdout.
			stdout := releaseStdout()
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if stdout != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, stdout)
			}
		})
	}
}
