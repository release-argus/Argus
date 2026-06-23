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
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
)

func TestShoutrrr_BuildURL(t *testing.T) {
	// GIVEN: a Shoutrrr.
	tests := []struct {
		name                       string
		sType                      string
		options, urlFields, params map[string]string
		want                       string
	}{
		{
			name:  "bark/base",
			sType: "bark",
			want:  "bark://:KEY@HOST:8080",
			urlFields: map[string]string{
				"devicekey": "KEY",
				"host":      "HOST",
				"port":      "8080",
			},
		},
		{
			name:  "bark/base + path",
			sType: "bark",
			want:  "bark://:KEY@HOST:8080/shazam",
			urlFields: map[string]string{
				"devicekey": "KEY",
				"host":      "HOST",
				"port":      "8080",
				"path":      "shazam",
			},
		},
		{
			name:  "discord/base",
			sType: "discord",
			want:  "discord://TOKEN@WEBHOOKID",
			urlFields: map[string]string{
				"token":     "TOKEN",
				"webhookid": "WEBHOOKID",
			},
		},
		{
			name:  "discord/base + thread_id",
			sType: "discord",
			want:  "discord://TOKEN@WEBHOOKID?thread_id=1234567890",
			urlFields: map[string]string{
				"token":     "TOKEN",
				"webhookid": "WEBHOOKID",
			},
			params: map[string]string{
				"thread_id": "1234567890",
			},
		},
		{
			name:  "generic/base",
			sType: "generic",
			want:  "generic://HOST",
			urlFields: map[string]string{
				"host": "HOST",
			},
		},
		{
			name:  "generic/base + headers",
			sType: "generic",
			want:  "generic://HOST?@contentType=val2&@fooBar=val1",
			urlFields: map[string]string{
				"host":    "HOST",
				"headers": `{"fooBar":"val1","contentType":"val2"}`,
			},
		},
		{
			name:  "generic/base + json_payload_vars",
			sType: "generic",
			want:  "generic://HOST?$key1=val1",
			urlFields: map[string]string{
				"host":              "HOST",
				"json_payload_vars": `{"key1":"val1"}`,
			},
		},
		{
			name:  "generic/base + query_vars",
			sType: "generic",
			want:  "generic://HOST?foo=bar",
			urlFields: map[string]string{
				"host":       "HOST",
				"query_vars": `{"foo":"bar"}`,
			},
		},
		{
			name:  "generic/base + headers + json_payload_vars + query_vars",
			sType: "generic",
			want:  "generic://HOST?@contentType=val2&@fooBar=val1&$key1=val1&foo=bar",
			urlFields: map[string]string{
				"host":              "HOST",
				"headers":           `{"fooBar":"val1","contentType":"val2"}`,
				"json_payload_vars": `{"key1":"val1"}`,
				"query_vars":        `{"foo":"bar"}`,
			},
		},
		{
			name:  "gotify/base",
			sType: "gotify",
			want:  "gotify://HOST/TOKEN",
			urlFields: map[string]string{
				"host": "HOST", "token": "TOKEN",
			},
		},
		{
			name:  "gotify/base + port",
			sType: "gotify",
			want:  "gotify://HOST:8443/TOKEN",
			urlFields: map[string]string{
				"host":  "HOST",
				"token": "TOKEN",
				"port":  "8443",
			},
		},
		{
			name:  "gotify/base + port + path",
			sType: "gotify",
			want:  "gotify://HOST:8443/PATH/TOKEN",
			urlFields: map[string]string{
				"host":  "HOST",
				"token": "TOKEN",
				"path":  "PATH",
				"port":  "8443",
			},
		},
		{
			name:  "googlechat",
			sType: "googlechat",
			want:  "googlechat://RAW",
			urlFields: map[string]string{
				"raw": "RAW",
			},
		},
		{
			name:  "ifttt",
			sType: "ifttt",
			want:  "ifttt://WEBHOOKID/?events=EVENT1,EVENT2",
			urlFields: map[string]string{
				"webhookid": "WEBHOOKID",
			},
			params: map[string]string{
				"events": "EVENT1,EVENT2",
			},
		},
		{
			name:  "join",
			sType: "join",
			want:  "join://shoutrrr:APIKEY@join/?devices=DEVICE1,DEVICE2",
			urlFields: map[string]string{
				"apikey": "APIKEY",
			},
			params: map[string]string{
				"devices": "DEVICE1,DEVICE2",
			},
		},
		{
			name:  "mattermost/base",
			sType: "mattermost",
			want:  "mattermost://HOST/TOKEN",
			urlFields: map[string]string{
				"host":  "HOST",
				"token": "TOKEN",
			},
		},
		{
			name:  "mattermost/base + username",
			sType: "mattermost",
			want:  "mattermost://USERNAME@HOST/TOKEN",
			urlFields: map[string]string{
				"host":     "HOST",
				"token":    "TOKEN",
				"username": "USERNAME",
			},
		},
		{
			name:  "mattermost/base + username + port",
			sType: "mattermost",
			want:  "mattermost://USERNAME@HOST:8443/TOKEN",
			urlFields: map[string]string{
				"host":     "HOST",
				"token":    "TOKEN",
				"username": "USERNAME",
				"port":     "8443",
			},
		},
		{
			name:  "matrix/base",
			sType: "matrix",
			want:  "matrix://:PASSWORD@HOST/",
			urlFields: map[string]string{
				"host":     "HOST",
				"password": "PASSWORD",
			},
		},
		{
			name:  "matrix/base + user",
			sType: "matrix",
			want:  "matrix://USER:PASSWORD@HOST/",
			urlFields: map[string]string{
				"host":     "HOST",
				"password": "PASSWORD",
				"user":     "USER",
			},
		},
		{
			name:  "matrix/base + user + port",
			sType: "matrix",
			want:  "matrix://USER:PASSWORD@HOST:8443/",
			urlFields: map[string]string{
				"host":     "HOST",
				"password": "PASSWORD",
				"user":     "USER",
				"port":     "8443",
			},
		},
		{
			name:  "matrix/base + user + port + rooms",
			sType: "matrix",
			want:  "matrix://USER:PASSWORD@HOST:8443/?rooms=ROOMS",
			urlFields: map[string]string{
				"host":     "HOST",
				"password": "PASSWORD",
				"user":     "USER",
				"port":     "8443",
			},
			params: map[string]string{
				"rooms": "ROOMS",
			},
		},
		{
			name:  "matrix/base + user + port + disabletls",
			sType: "matrix",
			want:  "matrix://USER:PASSWORD@HOST:8443/?disableTLS=yes",
			urlFields: map[string]string{
				"host":     "HOST",
				"password": "PASSWORD",
				"user":     "USER",
				"port":     "8443",
			},
			params: map[string]string{
				"disabletls": "yes",
			},
		},
		{
			name:  "matrix/base + user + port + rooms + disabletls",
			sType: "matrix",
			want:  "matrix://USER:PASSWORD@HOST:8443/?disableTLS=yes&rooms=ROOMS",
			urlFields: map[string]string{
				"host":     "HOST",
				"password": "PASSWORD",
				"user":     "USER",
				"port":     "8443",
			},
			params: map[string]string{
				"rooms":      "ROOMS",
				"disabletls": "yes",
			},
		},
		{
			name:  "ntfy/base",
			sType: "ntfy",
			want:  "ntfy://:@/TOPIC",
			urlFields: map[string]string{
				"topic": "TOPIC",
			},
		},
		{
			name:  "ntfy/base + username",
			sType: "ntfy",
			want:  "ntfy://USER:@/TOPIC",
			urlFields: map[string]string{
				"topic":    "TOPIC",
				"username": "USER",
			},
		},
		{
			name:  "ntfy/base + username + password",
			sType: "ntfy",
			want:  "ntfy://USER:PASS@/TOPIC",
			urlFields: map[string]string{
				"topic":    "TOPIC",
				"username": "USER",
				"password": "PASS",
			},
		},
		{
			name:  "ntfy/base + username + password + host",
			sType: "ntfy",
			want:  "ntfy://USER:PASS@HOST/TOPIC",
			urlFields: map[string]string{
				"topic":    "TOPIC",
				"username": "USER",
				"password": "PASS",
				"host":     "HOST",
			},
		},
		{
			name:  "ntfy/base + username + password + host + port",
			sType: "ntfy",
			want:  "ntfy://USER:PASS@HOST:8443/TOPIC",
			urlFields: map[string]string{
				"topic":    "TOPIC",
				"username": "USER",
				"password": "PASS",
				"host":     "HOST",
				"port":     "8443",
			},
		},
		{
			name:  "opsgenie/base",
			sType: "opsgenie",
			want:  "opsgenie://DEFAULT_HOST/APIKEY",
			urlFields: map[string]string{
				"host":   "DEFAULT_HOST",
				"apikey": "APIKEY",
			},
		},
		{
			name:  "opsgenie/base + port",
			sType: "opsgenie",
			want:  "opsgenie://DEFAULT_HOST:8443/APIKEY",
			urlFields: map[string]string{
				"host":   "DEFAULT_HOST",
				"apikey": "APIKEY",
				"port":   "8443",
			},
		},
		{
			name:  "notifiarr/base",
			sType: "notifiarr",
			want:  "notifiarr://APIKEY",
			urlFields: map[string]string{
				"apikey": "APIKEY",
			},
		},
		{
			name:  "notifiarr/base + channel",
			sType: "notifiarr",
			want:  "notifiarr://APIKEY?channel=CHANNEL",
			urlFields: map[string]string{
				"apikey": "APIKEY",
			},
			params: map[string]string{
				"channel": "CHANNEL",
			},
		},
		{
			name:  "pushbullet/base",
			sType: "pushbullet",
			want:  "pushbullet://TOKEN/TARGETS",
			urlFields: map[string]string{
				"token":   "TOKEN",
				"targets": "TARGETS",
			},
		},
		{
			name:  "pushover/base",
			sType: "pushover",
			want:  "pushover://shoutrrr:TOKEN@USER/",
			urlFields: map[string]string{
				"token": "TOKEN",
				"user":  "USER",
			},
		},
		{
			name:  "pushover/base + devices",
			sType: "pushover",
			want:  "pushover://shoutrrr:TOKEN@USER/?devices=DEVICES",
			urlFields: map[string]string{
				"token": "TOKEN",
				"user":  "USER",
			},
			params: map[string]string{
				"devices": "DEVICES",
			},
		},
		{
			name:  "rocketchat/base",
			sType: "rocketchat",
			want:  "rocketchat://HOST/TOKENA/TOKENB/CHANNEL",
			urlFields: map[string]string{
				"host":    "HOST",
				"tokena":  "TOKENA",
				"tokenb":  "TOKENB",
				"channel": "CHANNEL",
			},
		},
		{
			name:  "rocketchat/base + port",
			sType: "rocketchat",
			want:  "rocketchat://HOST:8443/TOKENA/TOKENB/CHANNEL",
			urlFields: map[string]string{
				"host":    "HOST",
				"tokena":  "TOKENA",
				"tokenb":  "TOKENB",
				"channel": "CHANNEL",
				"port":    "8443",
			},
		},
		{
			name:  "slack/base",
			sType: "slack",
			want:  "slack://TOKEN@CHANNEL",
			urlFields: map[string]string{
				"token":   "TOKEN",
				"channel": "CHANNEL",
			},
		},
		{
			name:  "slack/base + thread_ts",
			sType: "slack",
			want:  "slack://TOKEN@CHANNEL?thread_ts=1609459200.000001",
			urlFields: map[string]string{
				"token":   "TOKEN",
				"channel": "CHANNEL",
			},
			params: map[string]string{
				"thread_ts": "1609459200.000001",
			},
		},
		{
			name:  "smtp/base",
			sType: "smtp",
			want:  "smtp://HOST/?fromaddress=FROMADDRESS&toaddresses=TO_ADDRESS1%2CTO_ADDRESS2",
			urlFields: map[string]string{
				"host": "HOST",
			},
			params: map[string]string{
				"fromaddress": "FROMADDRESS",
				"toaddresses": "TO_ADDRESS1,TO_ADDRESS2",
			},
		},
		{
			name:  "smtp/base + login",
			sType: "smtp",
			want:  "smtp://USERNAME:PASSWORD@HOST/?fromaddress=FROMADDRESS&toaddresses=TO_ADDRESS1%2CTO_ADDRESS2",
			urlFields: map[string]string{
				"host":     "HOST",
				"username": "USERNAME",
				"password": "PASSWORD",
			},
			params: map[string]string{
				"fromaddress": "FROMADDRESS",
				"toaddresses": "TO_ADDRESS1,TO_ADDRESS2",
			},
		},
		{
			name:  "smtp/base + login + port",
			sType: "smtp",
			want:  "smtp://USERNAME:PASSWORD@HOST:587/?fromaddress=FROMADDRESS&toaddresses=TO_ADDRESS1%2CTO_ADDRESS2",
			urlFields: map[string]string{
				"host":     "HOST",
				"username": "USERNAME",
				"password": "PASSWORD",
				"port":     "587",
			},
			params: map[string]string{
				"fromaddress": "FROMADDRESS",
				"toaddresses": "TO_ADDRESS1,TO_ADDRESS2",
			},
		},
		{
			name:  "teams/base",
			sType: "teams",
			want:  "teams://?host=https%3A%2F%2Fprod-00.westus.logic.azure.com%3A443%2Fworkflows%2Fabc",
			params: map[string]string{
				"host": "https://prod-00.westus.logic.azure.com:443/workflows/abc",
			},
		},
		{
			name:  "teams/base + color",
			sType: "teams",
			want:  "teams://?host=https%3A%2F%2Fprod-00.westus.logic.azure.com%3A443%2Fworkflows%2Fabc&color=blue",
			params: map[string]string{
				"host":  "https://prod-00.westus.logic.azure.com:443/workflows/abc",
				"color": "blue",
			},
		},
		{
			name:  "teams/base + color + title",
			sType: "teams",
			want:  "teams://?host=https%3A%2F%2Fprod-00.westus.logic.azure.com%3A443%2Fworkflows%2Fabc&color=blue&title=My+Title",
			params: map[string]string{
				"host":  "https://prod-00.westus.logic.azure.com:443/workflows/abc",
				"color": "blue",
				"title": "My Title",
			},
		},
		{
			name:  "telegram",
			sType: "telegram",
			want:  "telegram://TOKEN@telegram?chats=CHATS",
			urlFields: map[string]string{
				"token": "TOKEN",
			},
			params: map[string]string{
				"chats": "CHATS",
			},
		},
		{
			name:  "zulip/base",
			sType: "zulip",
			want:  "zulip://BOTMAIL:BOTKEY@HOST",
			urlFields: map[string]string{
				"host":    "HOST",
				"botmail": "BOTMAIL",
				"botkey":  "BOTKEY",
			},
		},
		{
			name:  "zulip/base + token",
			sType: "zulip",
			want:  "zulip://BOTMAIL:BOTKEY@HOST?topic=TOPIC",
			urlFields: map[string]string{
				"host":    "HOST",
				"botmail": "BOTMAIL",
				"botkey":  "BOTKEY",
			},
			params: map[string]string{
				"topic": "TOPIC",
			},
		},
		{
			name:  "zulip/base + stream",
			sType: "zulip",
			want:  "zulip://BOTMAIL:BOTKEY@HOST?stream=STREAM",
			urlFields: map[string]string{
				"host":    "HOST",
				"botmail": "BOTMAIL",
				"botkey":  "BOTKEY",
			},
			params: map[string]string{
				"stream": "STREAM",
			},
		},
		{
			name:  "zulip/base + stream + token ",
			sType: "zulip",
			want:  "zulip://BOTMAIL:BOTKEY@HOST?stream=STREAM&topic=TOPIC",
			urlFields: map[string]string{
				"host":    "HOST",
				"botmail": "BOTMAIL",
				"botkey":  "BOTKEY",
			},
			params: map[string]string{
				"stream": "STREAM",
				"topic":  "TOPIC",
			},
		},
		{
			name:  "zulip//base + direct message",
			sType: "zulip",
			want:  "zulip://BOTMAIL:BOTKEY@HOST?to=user%40example.com&type=direct",
			urlFields: map[string]string{
				"host":    "HOST",
				"botmail": "BOTMAIL",
				"botkey":  "BOTKEY",
			},
			params: map[string]string{
				"type": "direct",
				"to":   "user@example.com",
			},
		},
		{
			name:  "zulip/base + title",
			sType: "zulip",
			want:  "zulip://BOTMAIL:BOTKEY@HOST?title=My+Title",
			urlFields: map[string]string{
				"host":    "HOST",
				"botmail": "BOTMAIL",
				"botkey":  "BOTKEY",
			},
			params: map[string]string{
				"title": "My Title",
			},
		},
		{
			name:  "zulip/base + stream + topic + title + read_by_sender",
			sType: "zulip",
			want:  "zulip://BOTMAIL:BOTKEY@HOST?read_by_sender=true&stream=STREAM&title=Alert&topic=TOPIC",
			urlFields: map[string]string{
				"host":    "HOST",
				"botmail": "BOTMAIL",
				"botkey":  "BOTKEY",
			},
			params: map[string]string{
				"stream":         "STREAM",
				"topic":          "TOPIC",
				"title":          "Alert",
				"read_by_sender": "true",
			},
		},
		{
			name:  "shoutrrr",
			sType: "shoutrrr",
			want:  "RAW",
			urlFields: map[string]string{
				"raw": "RAW",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.Type = tc.sType
			shoutrrr.URLFields = tc.urlFields
			shoutrrr.Params = tc.params

			// WHEN: BuildURL is called.
			got := shoutrrr.BuildURL()

			// THEN: the expected URL is returned.
			if got != tc.want {
				t.Errorf(
					"%s\nShoutrrr.BuildURL() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestShoutrrr_BuildParams(t *testing.T) {
	// GIVEN: a Shoutrrr and ServiceInfo.
	svcInfo := serviceinfo.ServiceInfo{
		ID:              "service_id",
		Name:            "service_name",
		URL:             "service_url",
		Icon:            "https://example.com/icon.png",
		IconLinkTo:      "https://example.com/icon_link_to",
		WebURL:          "service_web_url",
		ApprovedVersion: "1.2.3a",
		DeployedVersion: "1.2.3b",
		LatestVersion:   "1.2.3c",
	}
	tests := []struct {
		name                                                 string
		rootValue, mainValue, defaultValue, hardDefaultValue *string
		env                                                  map[string]string
		want                                                 string
	}{
		{
			name:             "root overrides all",
			want:             "this",
			rootValue:        test.Ptr("this"),
			mainValue:        test.Ptr("not_this"),
			defaultValue:     test.Ptr("not_this"),
			hardDefaultValue: test.Ptr("not_this"),
		},
		{
			name:             "main overrides default and hardDefault",
			want:             "this",
			rootValue:        nil,
			mainValue:        test.Ptr("this"),
			defaultValue:     test.Ptr("not_this"),
			hardDefaultValue: test.Ptr("not_this"),
		},
		{
			name:             "default overrides hardDefault",
			want:             "this",
			rootValue:        nil,
			mainValue:        nil,
			defaultValue:     test.Ptr("this"),
			hardDefaultValue: test.Ptr("not_this"),
		},
		{
			name:             "hardDefaultValue is last resort",
			want:             "this",
			rootValue:        nil,
			mainValue:        nil,
			defaultValue:     nil,
			hardDefaultValue: test.Ptr("this"),
		},
		{
			name:             "django templating",
			want:             "this",
			rootValue:        test.Ptr("{% if 'a' == 'a' %}this{% endif %}"),
			mainValue:        nil,
			defaultValue:     test.Ptr("not_this"),
			hardDefaultValue: test.Ptr("not_this"),
		},
		{
			name: "django vars",
			want: fmt.Sprintf(
				"foo%s-%s",
				svcInfo.ID, svcInfo.LatestVersion,
			),
			rootValue:        test.Ptr("foo{{ service_id }}-{{ version }}"),
			mainValue:        nil,
			defaultValue:     test.Ptr("not_this"),
			hardDefaultValue: test.Ptr("not_this"),
		},
		{
			name: "all django vars",
			want: fmt.Sprintf(
				"foo-%s-%s-%s--%s-%s-%s--%s-%s-%s-%s",
				svcInfo.ID, svcInfo.Name, svcInfo.URL,
				svcInfo.Icon, svcInfo.IconLinkTo, svcInfo.WebURL,
				svcInfo.LatestVersion, svcInfo.ApprovedVersion, svcInfo.DeployedVersion, svcInfo.LatestVersion,
			),
			rootValue:        test.Ptr("foo-{{ service_id }}-{{ service_name }}-{{ service_url }}--{{ icon }}-{{ icon_link_to }}-{{ web_url }}--{{ version }}-{{ approved_version }}-{{ deployed_version }}-{{ latest_version }}"),
			mainValue:        nil,
			defaultValue:     test.Ptr("not_this"),
			hardDefaultValue: test.Ptr("not_this"),
		},
		{
			name:             "env vars",
			want:             "this",
			rootValue:        test.Ptr("${ARGUS_TEST_SHOUTRRR_BUILD_PARAMS}"),
			mainValue:        test.Ptr("not_this"),
			defaultValue:     test.Ptr("not_this"),
			hardDefaultValue: test.Ptr("not_this"),
			env: map[string]string{
				"ARGUS_TEST_SHOUTRRR_BUILD_PARAMS": "this",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.SetEnv(t, tc.env)

			key := "test"
			shoutrrr := testShoutrrr(false, false)
			if tc.rootValue != nil {
				shoutrrr.Params[key] = *tc.rootValue
			}
			if tc.mainValue != nil {
				shoutrrr.Main.Params[key] = *tc.mainValue
			}
			if tc.defaultValue != nil {
				shoutrrr.Defaults.Params[key] = *tc.defaultValue
			}
			if tc.hardDefaultValue != nil {
				shoutrrr.HardDefaults.Params[key] = *tc.hardDefaultValue
			}

			// WHEN: BuildParams is called.
			result := shoutrrr.BuildParams(svcInfo)

			// THEN: the function returns the params to use.
			if got := (*result)[key]; got != tc.want {
				t.Fatalf(
					"%s\nShoutrrr.BuildParams(%+v) mismatch\ngot:  %q\nwant: %q",
					packageName, svcInfo,
					got, tc.want,
				)
			}
		})
	}
}

func TestShoutrrr_GetSender(t *testing.T) {
	type wants struct {
		err        bool
		title, msg string
		params     string
	}
	// GIVEN: a Shoutrrr, ServiceInfo and the hard defaults.
	svcInfo := serviceinfo.ServiceInfo{
		ID:            "service_id",
		LatestVersion: "1.2.3",
	}
	hardDefaults := ShoutrrrsDefaults{}
	hardDefaults.Default()
	tests := []struct {
		name         string
		title        string
		msg          string
		url          string
		shoutrrrYAML string
		mainYAML     string
		wants        wants
	}{
		{
			name:  "valid sender with title and message",
			title: "Test Title",
			msg:   "Test Message {{ version }}",
			shoutrrrYAML: test.TrimYAML(`
				type: gotify
				url_fields:
					host: HOST
					token: TOKEN
				params:
					root: "root"
			`),
			mainYAML: test.TrimYAML(`
				params:
					root: "main"
					main: "main"
			`),
			url: "gotify://HOST:443/TOKEN",
			wants: wants{
				title: "Test Title",
				msg:   "Test Message {{ version }}",
			},
		},
		{
			name: "valid sender without title or message",
			url:  "gotify://HOST:443/TOKEN",
			shoutrrrYAML: test.TrimYAML(`
				type: gotify
				url_fields:
					host: HOST
					token: TOKEN
			`),
			wants: wants{
				title: "",
				msg: fmt.Sprintf(
					`%s - %s released`,
					svcInfo.ID, svcInfo.LatestVersion,
				),
			},
		},
		{
			name: "invalid sender URL",
			shoutrrrYAML: test.TrimYAML(`
				type: gotify
				url_fields:
					host: "invalid:	//example.com"
			`),
			wants: wants{
				err: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := &Shoutrrr{}
			if err := decode.Unmarshal("yaml", []byte(tc.shoutrrrYAML), shoutrrr); err != nil {
				t.Fatalf(
					"%s\nfailed to unmarshal Shoutrrr: %v",
					packageName, err,
				)
			}
			main := &Defaults{}
			if err := decode.Unmarshal("yaml", []byte(tc.mainYAML), main); err != nil {
				t.Fatalf(
					"%s\nfailed to unmarshal Shoutrrr.Main: %v",
					packageName, err,
				)
			}
			svcStatus := status.Status{}
			svcStatus.Init(
				0, 1, 0,
				status.ServiceInfo{
					ID: tc.name,
				},
				&dashboard.Options{},
			)
			shoutrrr.Init(
				&svcStatus,
				main,
				&Defaults{}, hardDefaults[shoutrrr.Type],
			)

			// WHEN: getSender is called.
			_, message, params, url, err := shoutrrr.getSender(tc.title, tc.msg, svcInfo)

			prefix := fmt.Sprintf("%s\nShoutrrr.getSender()", packageName)

			// THEN: the expected results are returned.
			if gotErr := (err != nil); gotErr != tc.wants.err {
				t.Fatalf(
					"%s error mismatch\ngot:  decode=%t\nwant: decode=%t",
					prefix, gotErr, tc.wants.err,
				)
			}
			if err == nil {
				if url != tc.url {
					t.Errorf(
						"%s url mismatch\ngot:  %q\n\nwant: %q",
						prefix, url, tc.url,
					)
				}
				if message != tc.wants.msg {
					t.Errorf(
						"%s message mismatch\ngot:  %q\n\nwant: %q",
						prefix, message, tc.wants.msg,
					)
				}
				if got := (*params)["title"]; tc.wants.title != "" && got != tc.wants.title {
					t.Errorf(
						"%s params.title mismatch\ngot:  %q\n\nwant: %q",
						prefix, got, tc.wants.title,
					)
				}
			}
		})
	}
}

func TestShoutrrr_ParseSend(t *testing.T) {
	// GIVEN: a possible list of errors from a send operation.
	tests := []struct {
		name        string
		errs        []error
		serviceName string
		deleting    bool
		wantFailed  bool
		errCounts   map[string]int
	}{
		{
			name:        "no errors, service not deleting",
			errs:        []error{nil, nil},
			serviceName: "service1",
			wantFailed:  false,
			errCounts:   map[string]int{},
		},
		{
			name: "errors/service not deleting",
			errs: []error{
				errors.New("error1"),
				errors.New("error2"),
				errors.New("error1"),
			},
			serviceName: "service1",
			wantFailed:  true,
			errCounts: map[string]int{
				"error1": 2,
				"error2": 1,
			},
		},
		{
			name: "errors/service deleting",
			errs: []error{
				errors.New("error1"),
				errors.New("error2"),
			},
			serviceName: "service1",
			deleting:    true,
			wantFailed:  false,
			errCounts:   map[string]int{},
		},
		{
			name: "no service name",
			errs: []error{
				errors.New("error1"),
				errors.New("error1"),
				errors.New("error1"),
			},
			serviceName: "",
			wantFailed:  true,
			errCounts: map[string]int{
				"error1": 3,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're testing metrics.

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.ServiceStatus = &status.Status{}
			if tc.deleting {
				shoutrrr.ServiceStatus.SetDeleting()
			}
			logFrom := logx.LogFrom{Primary: "test", Secondary: "test"}
			combinedErrs := map[string]int{}

			// WHEN: parseSend is called on them.
			failed := shoutrrr.parseSend(tc.errs, combinedErrs, tc.serviceName, logFrom)

			prefix := fmt.Sprintf("%s\nShoutrrr.parseSend()", packageName)

			// THEN: the boolean result is as expected.
			if failed != tc.wantFailed {
				t.Errorf(
					"%s resi;t mismatch\ngot:  %t\nwant: %t",
					prefix, tc.wantFailed, failed,
				)
			}

			// AND: the errors are combined as expected.
			if testErr := test.AssertMapEqual(
				t,
				combinedErrs,
				tc.errCounts,
				prefix,
				"combinedErrs",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestQueryParam(t *testing.T) {
	// GIVEN: a key and value.
	tests := []struct {
		name string
		key  string
		val  string
		want string
	}{
		{
			name: "empty value returns empty string",
			key:  "foo",
			val:  "",
			want: "",
		},
		{
			name: "plain ASCII value",
			key:  "foo",
			val:  "bar",
			want: "foo=bar",
		},
		{
			name: "space is encoded as +",
			key:  "title",
			val:  "My Title",
			want: "title=My+Title",
		},
		{
			name: "@ is percent-encoded",
			key:  "to",
			val:  "user@example.com",
			want: "to=user%40example.com",
		},
		{
			name: "comma is percent-encoded",
			key:  "toaddresses",
			val:  "a@b.com,c@d.com",
			want: "toaddresses=a%40b.com%2Cc%40d.com",
		},
		{
			name: "& is percent-encoded",
			key:  "q",
			val:  "foo&bar",
			want: "q=foo%26bar",
		},
		{
			name: "= is percent-encoded",
			key:  "q",
			val:  "a=b",
			want: "q=a%3Db",
		},
		{
			name: "URL value is percent-encoded",
			key:  "host",
			val:  "https://example.com:443/path",
			want: "host=https%3A%2F%2Fexample.com%3A443%2Fpath",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: queryParam is called.
			got := queryParam(tc.key, tc.val)

			// THEN: the expected string is returned.
			if got != tc.want {
				t.Errorf(
					"%s\nqueryParam(%q, %q) mismatch\ngot:  %q\nwant: %q",
					packageName, tc.key, tc.val,
					got, tc.want,
				)
			}
		})
	}
}

func TestJSONMapToString(t *testing.T) {
	// GIVEN: a JSON string.
	tests := []struct {
		name    string
		jsonStr string
		want    string
	}{
		{
			name:    "empty",
			jsonStr: ``,
			want:    "",
		},
		{
			name:    "empty map",
			jsonStr: `{}`,
			want:    "",
		},
		{
			name:    "single",
			jsonStr: `{"foo":"bar"}`,
			want:    "@foo=bar",
		},
		{
			name:    "multiple",
			jsonStr: `{"foo":"bar","bar":"foo","hi":"there"}`,
			want:    "@bar=foo&@foo=bar&@hi=there",
		},
		{
			name:    "invalid (lists are not supported)",
			jsonStr: `{"foo":["alpha","bravo","charlie"]}`,
			want:    "",
		},
		{
			name:    "invalid JSON",
			jsonStr: `{"foo":"bar","bar":"foo`,
			want:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: jsonMapToString is called.
			got := jsonMapToString(tc.jsonStr, "@")

			// THEN: the expected URL is returned.
			if got != tc.want {
				t.Errorf(
					"%s\njsonMapToString(%q, @) mismatch\ngot:  %q\nwant: %q",
					packageName, tc.jsonStr,
					got, tc.want,
				)
			}
		})
	}
}
