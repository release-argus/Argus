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
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/test"
	logutil "github.com/release-argus/Argus/util/log"
)

func TestShoutrrr_getSender(t *testing.T) {
	type wants struct {
		err        bool
		title, msg string
		params     string
	}
	// GIVEN a Shoutrrr, ServiceInfo and the hard defaults.
	svcInfo := serviceinfo.ServiceInfo{
		ID:            "service_id",
		LatestVersion: "1.2.3",
	}
	hardDefaults := ShoutrrrsDefaults{}
	hardDefaults.Default()
	tests := map[string]struct {
		title        string
		msg          string
		url          string
		shoutrrrYAML string
		mainYAML     string
		wants        wants
	}{
		"valid sender with title and message": {
			title: "Test Title",
			msg:   "Test Message {{ version }}",
			shoutrrrYAML: test.TrimYAML(`
				type: gotify
				url_fields:
					host: HOST
					token: TOKEN
				params:
					root: "root"`),
			mainYAML: test.TrimYAML(`
				params:
					root: "main"
					main: "main"`),
			url: "gotify://HOST:443/TOKEN",
			wants: wants{
				title: "Test Title",
				msg:   "Test Message {{ version }}"},
		},
		"valid sender without title or message": {
			url: "gotify://HOST:443/TOKEN",
			shoutrrrYAML: test.TrimYAML(`
				type: gotify
				url_fields:
					host: HOST
					token: TOKEN`),
			wants: wants{
				title: "",
				msg: fmt.Sprintf(`%s - %s released`,
					svcInfo.ID, svcInfo.LatestVersion)},
		},
		"invalid sender URL": {
			shoutrrrYAML: test.TrimYAML(`
				type: gotify
				url_fields:
					host: "invalid:	//example.com"`),
			wants: wants{err: true},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := &Shoutrrr{}
			if err := yaml.Unmarshal([]byte(tc.shoutrrrYAML), shoutrrr); err != nil {
				t.Fatalf("%s\nfailed to unmarshal shoutrrr YAML: %v",
					packageName, err)
			}
			main := &Defaults{}
			if err := yaml.Unmarshal([]byte(tc.mainYAML), main); err != nil {
				t.Fatalf("%s\nfailed to unmarshal main YAML: %v",
					packageName, err)
			}
			svcStatus := status.Status{}
			svcStatus.Init(1, 0, 0,
				name, "", "",
				&dashboard.Options{})
			shoutrrr.Init(
				&svcStatus,
				main,
				&Defaults{}, hardDefaults[shoutrrr.Type])

			// WHEN getSender is called.
			_, message, params, url, err := shoutrrr.getSender(tc.title, tc.msg, svcInfo)

			// THEN the expected results are returned.
			if (err != nil) != tc.wants.err {
				t.Fatalf("%s\nerror mismatch\nwant: err=%t\ngot:  err=%t",
					packageName, tc.wants.err, err != nil)
			}
			if err == nil {
				if url != tc.url {
					t.Errorf("%s\nurl mismatch\nwant: %q\ngot:  %q\n",
						packageName, tc.url, url)
				}
				if message != tc.wants.msg {
					t.Errorf("%s\nmessage mismatch\nwant: %q\ngot:  %q\n",
						packageName, tc.wants.msg, message)
				}
				if tc.wants.title != "" && (*params)["title"] != tc.wants.title {
					t.Errorf("%s\ntitle mismatch\nwant: %q\ngot:  %q\n",
						packageName, tc.wants.title, (*params)["title"])
				}
			}
		})
	}
}

func TestShoutrrr_BuildURL(t *testing.T) {
	// GIVEN a Shoutrrr.
	tests := map[string]struct {
		sType                      string
		options, urlFields, params map[string]string
		want                       string
	}{
		"bark - base": {
			sType: "bark",
			want:  "bark://:KEY@HOST:8080",
			urlFields: map[string]string{
				"devicekey": "KEY",
				"host":      "HOST",
				"port":      "8080"},
		},
		"bark - base + path": {
			sType: "bark",
			want:  "bark://:KEY@HOST:8080/shazam",
			urlFields: map[string]string{
				"devicekey": "KEY",
				"host":      "HOST",
				"port":      "8080",
				"path":      "shazam"},
		},
		"discord - base": {
			sType: "discord",
			want:  "discord://TOKEN@WEBHOOKID",
			urlFields: map[string]string{
				"token":     "TOKEN",
				"webhookid": "WEBHOOKID"},
		},
		"smtp - base": {
			sType: "smtp",
			want:  "smtp://HOST/?fromaddress=FROMADDRESS&toaddresses=TO_ADDRESS1,TO_ADDRESS2",
			urlFields: map[string]string{
				"host": "HOST"},
			params: map[string]string{
				"fromaddress": "FROMADDRESS",
				"toaddresses": "TO_ADDRESS1,TO_ADDRESS2"},
		},
		"smtp - base + login": {
			sType: "smtp",
			want:  "smtp://USERNAME:PASSWORD@HOST/?fromaddress=FROMADDRESS&toaddresses=TO_ADDRESS1,TO_ADDRESS2",
			urlFields: map[string]string{
				"host":     "HOST",
				"username": "USERNAME",
				"password": "PASSWORD"},
			params: map[string]string{
				"fromaddress": "FROMADDRESS",
				"toaddresses": "TO_ADDRESS1,TO_ADDRESS2"},
		},
		"smtp - base + login + port": {
			sType: "smtp",
			want:  "smtp://USERNAME:PASSWORD@HOST:587/?fromaddress=FROMADDRESS&toaddresses=TO_ADDRESS1,TO_ADDRESS2",
			urlFields: map[string]string{
				"host":     "HOST",
				"username": "USERNAME",
				"password": "PASSWORD",
				"port":     "587"},
			params: map[string]string{
				"fromaddress": "FROMADDRESS",
				"toaddresses": "TO_ADDRESS1,TO_ADDRESS2"},
		},
		"gotify - base": {
			sType: "gotify",
			want:  "gotify://HOST/TOKEN",
			urlFields: map[string]string{
				"host": "HOST", "token": "TOKEN"},
		},
		"gotify - base + port": {
			sType: "gotify",
			want:  "gotify://HOST:8443/TOKEN",
			urlFields: map[string]string{
				"host":  "HOST",
				"token": "TOKEN",
				"port":  "8443"},
		},
		"gotify - base + port + path": {
			sType: "gotify",
			want:  "gotify://HOST:8443/PATH/TOKEN",
			urlFields: map[string]string{
				"host":  "HOST",
				"token": "TOKEN",
				"path":  "PATH",
				"port":  "8443"},
		},
		"googlechat - base": {
			sType: "googlechat",
			want:  "googlechat://RAW",
			urlFields: map[string]string{
				"raw": "RAW"},
		},
		"ifttt - base": {
			sType: "ifttt",
			want:  "ifttt://WEBHOOKID/?events=EVENT1,EVENT2",
			urlFields: map[string]string{
				"webhookid": "WEBHOOKID"},
			params: map[string]string{
				"events": "EVENT1,EVENT2"},
		},
		"join - base": {
			sType: "join",
			want:  "join://shoutrrr:APIKEY@join/?devices=DEVICE1,DEVICE2",
			urlFields: map[string]string{
				"apikey": "APIKEY"},
			params: map[string]string{
				"devices": "DEVICE1,DEVICE2"},
		},
		"mattermost - base": {
			sType: "mattermost",
			want:  "mattermost://HOST/TOKEN",
			urlFields: map[string]string{
				"host":  "HOST",
				"token": "TOKEN"},
		},
		"mattermost - base + username": {
			sType: "mattermost",
			want:  "mattermost://USERNAME@HOST/TOKEN",
			urlFields: map[string]string{
				"host":     "HOST",
				"token":    "TOKEN",
				"username": "USERNAME"},
		},
		"mattermost - base + port": {
			sType: "mattermost",
			want:  "mattermost://USERNAME@HOST:8443/TOKEN",
			urlFields: map[string]string{
				"host":     "HOST",
				"token":    "TOKEN",
				"username": "USERNAME",
				"port":     "8443"},
		},
		"mattermost - base + port + path": {
			sType: "mattermost",
			want:  "mattermost://USERNAME@HOST:8443/PATH/TOKEN",
			urlFields: map[string]string{
				"host":     "HOST",
				"token":    "TOKEN",
				"username": "USERNAME",
				"path":     "PATH",
				"port":     "8443"},
		},
		"matrix - base": {
			sType: "matrix",
			want:  "matrix://:PASSWORD@HOST/",
			urlFields: map[string]string{
				"host":     "HOST",
				"password": "PASSWORD"},
		},
		"matrix - base + user": {
			sType: "matrix",
			want:  "matrix://USER:PASSWORD@HOST/",
			urlFields: map[string]string{
				"host":     "HOST",
				"password": "PASSWORD",
				"user":     "USER"},
		},
		"matrix - base + user + port": {
			sType: "matrix",
			want:  "matrix://USER:PASSWORD@HOST:8443/",
			urlFields: map[string]string{
				"host":     "HOST",
				"password": "PASSWORD",
				"user":     "USER",
				"port":     "8443"},
		},
		"matrix - base + user + port + path": {
			sType: "matrix",
			want:  "matrix://USER:PASSWORD@HOST:8443/PATH/",
			urlFields: map[string]string{
				"host":     "HOST",
				"password": "PASSWORD",
				"user":     "USER",
				"port":     "8443",
				"path":     "PATH"},
		},
		"matrix - base + user + port + path + rooms": {
			sType: "matrix",
			want:  "matrix://USER:PASSWORD@HOST:8443/PATH/?rooms=ROOMS",
			urlFields: map[string]string{
				"host":     "HOST",
				"password": "PASSWORD",
				"user":     "USER",
				"port":     "8443",
				"path":     "PATH"},
			params: map[string]string{
				"rooms": "ROOMS"},
		},
		"matrix - base + user + port + path + disabletls": {
			sType: "matrix",
			want:  "matrix://USER:PASSWORD@HOST:8443/PATH/?disableTLS=yes",
			urlFields: map[string]string{
				"host":     "HOST",
				"password": "PASSWORD",
				"user":     "USER",
				"port":     "8443",
				"path":     "PATH"},
			params: map[string]string{
				"disabletls": "yes"},
		},
		"matrix - base + user + port + path + rooms + disabletls": {
			sType: "matrix",
			want:  "matrix://USER:PASSWORD@HOST:8443/PATH/?rooms=ROOMS&disableTLS=yes",
			urlFields: map[string]string{
				"host":     "HOST",
				"password": "PASSWORD",
				"user":     "USER",
				"port":     "8443",
				"path":     "PATH"},
			params: map[string]string{
				"rooms":      "ROOMS",
				"disabletls": "yes"},
		},
		"ntfy - base": {
			sType: "ntfy",
			want:  "ntfy://:@/TOPIC",
			urlFields: map[string]string{
				"topic": "TOPIC"},
		},
		"ntfy - base + username": {
			sType: "ntfy",
			want:  "ntfy://USER:@/TOPIC",
			urlFields: map[string]string{
				"topic":    "TOPIC",
				"username": "USER"},
		},
		"ntfy - base + username + password": {
			sType: "ntfy",
			want:  "ntfy://USER:PASS@/TOPIC",
			urlFields: map[string]string{
				"topic":    "TOPIC",
				"username": "USER",
				"password": "PASS"},
		},
		"ntfy - base + username + password + host": {
			sType: "ntfy",
			want:  "ntfy://USER:PASS@HOST/TOPIC",
			urlFields: map[string]string{
				"topic":    "TOPIC",
				"username": "USER",
				"password": "PASS",
				"host":     "HOST"},
		},
		"ntfy - base + username + password + host + port": {
			sType: "ntfy",
			want:  "ntfy://USER:PASS@HOST:8443/TOPIC",
			urlFields: map[string]string{
				"topic":    "TOPIC",
				"username": "USER",
				"password": "PASS",
				"host":     "HOST",
				"port":     "8443"},
		},
		"opsgenie - base": {
			sType: "opsgenie",
			want:  "opsgenie://DEFAULT_HOST/APIKEY",
			urlFields: map[string]string{
				"host":   "DEFAULT_HOST",
				"apikey": "APIKEY"},
		},
		"opsgenie - base + port": {
			sType: "opsgenie",
			want:  "opsgenie://DEFAULT_HOST:8443/APIKEY",
			urlFields: map[string]string{
				"host":   "DEFAULT_HOST",
				"apikey": "APIKEY",
				"port":   "8443"},
		},
		"opsgenie - base + port + path": {
			sType: "opsgenie",
			want:  "opsgenie://DEFAULT_HOST:8443/PATH/APIKEY",
			urlFields: map[string]string{
				"host":   "DEFAULT_HOST",
				"apikey": "APIKEY",
				"port":   "8443",
				"path":   "PATH"},
		},
		"pushbullet - base": {
			sType: "pushbullet",
			want:  "pushbullet://TOKEN/TARGETS",
			urlFields: map[string]string{
				"token":   "TOKEN",
				"targets": "TARGETS"},
		},
		"pushover - base": {
			sType: "pushover",
			want:  "pushover://shoutrrr:TOKEN@USER/",
			urlFields: map[string]string{
				"token": "TOKEN",
				"user":  "USER"},
		},
		"pushover - base + devices": {
			sType: "pushover",
			want:  "pushover://shoutrrr:TOKEN@USER/?devices=DEVICES",
			urlFields: map[string]string{
				"token": "TOKEN",
				"user":  "USER"},
			params: map[string]string{
				"devices": "DEVICES"},
		},
		"rocketchat - base": {
			sType: "rocketchat",
			want:  "rocketchat://HOST/TOKENA/TOKENB/CHANNEL",
			urlFields: map[string]string{
				"host":    "HOST",
				"tokena":  "TOKENA",
				"tokenb":  "TOKENB",
				"channel": "CHANNEL"},
		},
		"rocketchat - base + port": {
			sType: "rocketchat",
			want:  "rocketchat://HOST:8443/TOKENA/TOKENB/CHANNEL",
			urlFields: map[string]string{
				"host":    "HOST",
				"tokena":  "TOKENA",
				"tokenb":  "TOKENB",
				"channel": "CHANNEL",
				"port":    "8443"},
		},
		"rocketchat - base + port + path": {
			sType: "rocketchat",
			want:  "rocketchat://HOST:8443/PATH/TOKENA/TOKENB/CHANNEL",
			urlFields: map[string]string{
				"host":    "HOST",
				"tokena":  "TOKENA",
				"tokenb":  "TOKENB",
				"channel": "CHANNEL",
				"port":    "8443",
				"path":    "PATH"},
		},
		"slack - base": {
			sType: "slack",
			want:  "slack://TOKEN@CHANNEL",
			urlFields: map[string]string{
				"token":   "TOKEN",
				"channel": "CHANNEL"},
		},
		"teams - base": {
			sType: "teams",
			want:  "teams://GROUP@TENANT/ALTID/GROUPOWNER?host=HOST",
			urlFields: map[string]string{
				"group":      "GROUP",
				"tenant":     "TENANT",
				"altid":      "ALTID",
				"groupowner": "GROUPOWNER"},
			params: map[string]string{
				"host": "HOST"},
		},
		"telegram - base": {
			sType: "telegram",
			want:  "telegram://TOKEN@telegram?chats=CHATS",
			urlFields: map[string]string{
				"token": "TOKEN"},
			params: map[string]string{
				"chats": "CHATS"},
		},
		"zulip - base": {
			sType: "zulip",
			want:  "zulip://BOTMAIL:BOTKEY@HOST",
			urlFields: map[string]string{
				"host":    "HOST",
				"botmail": "BOTMAIL",
				"botkey":  "BOTKEY"},
		},
		"zulip - base + token": {
			sType: "zulip",
			want:  "zulip://BOTMAIL:BOTKEY@HOST?topic=TOPIC",
			urlFields: map[string]string{
				"host":    "HOST",
				"botmail": "BOTMAIL",
				"botkey":  "BOTKEY"},
			params: map[string]string{
				"topic": "TOPIC"},
		},
		"zulip - base + stream": {
			sType: "zulip",
			want:  "zulip://BOTMAIL:BOTKEY@HOST?stream=STREAM",
			urlFields: map[string]string{
				"host":    "HOST",
				"botmail": "BOTMAIL",
				"botkey":  "BOTKEY"},
			params: map[string]string{
				"stream": "STREAM"},
		},
		"zulip - base + token + stream": {
			sType: "zulip",
			want:  "zulip://BOTMAIL:BOTKEY@HOST?stream=STREAM&topic=TOPIC",
			urlFields: map[string]string{
				"host": "HOST", "botmail": "BOTMAIL", "botkey": "BOTKEY"},
			params: map[string]string{
				"topic":  "TOPIC",
				"stream": "STREAM"},
		},
		"generic - base": {
			sType: "generic",
			want:  "generic://HOST",
			urlFields: map[string]string{
				"host": "HOST"},
		},
		"generic - base + custom_headers": {
			sType: "generic",
			want:  "generic://HOST?@contentType=val2&@fooBar=val1",
			urlFields: map[string]string{
				"host":           "HOST",
				"custom_headers": `{"fooBar":"val1","contentType":"val2"}`},
		},
		"generic - base + json_payload_vars": {
			sType: "generic",
			want:  "generic://HOST?$key1=val1",
			urlFields: map[string]string{
				"host":              "HOST",
				"json_payload_vars": `{"key1":"val1"}`},
		},
		"generic - base + query_vars": {
			sType: "generic",
			want:  "generic://HOST?foo=bar",
			urlFields: map[string]string{
				"host":       "HOST",
				"query_vars": `{"foo":"bar"}`},
		},
		"generic - base + custom_headers + json_payload_vars + query_vars": {
			sType: "generic",
			want:  "generic://HOST?@contentType=val2&@fooBar=val1&$key1=val1&foo=bar",
			urlFields: map[string]string{
				"host":              "HOST",
				"custom_headers":    `{"fooBar":"val1","contentType":"val2"}`,
				"json_payload_vars": `{"key1":"val1"}`,
				"query_vars":        `{"foo":"bar"}`},
		},
		"shoutrrr - base": {
			sType:     "shoutrrr",
			want:      "RAW",
			urlFields: map[string]string{"raw": "RAW"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.Type = tc.sType
			shoutrrr.URLFields = tc.urlFields
			shoutrrr.Params = tc.params

			// WHEN BuildURL is called.
			got := shoutrrr.BuildURL()

			// THEN the expected URL is returned.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func Test_jsonMapToString(t *testing.T) {
	// GIVEN a JSON string.
	tests := map[string]struct {
		jsonStr string
		want    string
	}{
		"empty": {
			jsonStr: ``,
			want:    "",
		},
		"empty map": {
			jsonStr: `{}`,
			want:    "",
		},
		"single": {
			jsonStr: `{"foo":"bar"}`,
			want:    "@foo=bar",
		},
		"multiple": {
			jsonStr: `{"foo":"bar","bar":"foo","hi":"there"}`,
			want:    "@bar=foo&@foo=bar&@hi=there",
		},
		"invalid (lists are not supported)": {
			jsonStr: `{"foo":["alpha","bravo","charlie"]}`,
			want:    "",
		},
		"invalid JSON": {
			jsonStr: `{"foo":"bar","bar":"foo`,
			want:    "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN jsonMapToString is called.
			got := jsonMapToString(tc.jsonStr, "@")

			// THEN the expected URL is returned.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestShoutrrr_BuildParams(t *testing.T) {
	// GIVEN a Shoutrrr and ServiceInfo.
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
	tests := map[string]struct {
		rootValue, mainValue, defaultValue, hardDefaultValue *string
		envVars                                              map[string]string
		want                                                 string
	}{
		"root overrides all": {
			want:             "this",
			rootValue:        test.StringPtr("this"),
			mainValue:        test.StringPtr("not_this"),
			defaultValue:     test.StringPtr("not_this"),
			hardDefaultValue: test.StringPtr("not_this"),
		},
		"main overrides default and hardDefault": {
			want:             "this",
			mainValue:        test.StringPtr("this"),
			defaultValue:     test.StringPtr("not_this"),
			hardDefaultValue: test.StringPtr("not_this"),
		},
		"default overrides hardDefault": {
			want:             "this",
			defaultValue:     test.StringPtr("this"),
			hardDefaultValue: test.StringPtr("not_this"),
		},
		"hardDefaultValue is last resort": {
			want:             "this",
			hardDefaultValue: test.StringPtr("this"),
		},
		"django templating": {
			want:             "this",
			rootValue:        test.StringPtr("{% if 'a' == 'a' %}this{% endif %}"),
			defaultValue:     test.StringPtr("not_this"),
			hardDefaultValue: test.StringPtr("not_this"),
		},
		"django vars": {
			want: fmt.Sprintf("foo%s-%s",
				svcInfo.ID, svcInfo.LatestVersion),
			rootValue:        test.StringPtr("foo{{ service_id }}-{{ version }}"),
			defaultValue:     test.StringPtr("not_this"),
			hardDefaultValue: test.StringPtr("not_this"),
		},
		"all django vars": {
			want: fmt.Sprintf("foo-%s-%s-%s--%s-%s-%s--%s-%s-%s-%s",
				svcInfo.ID, svcInfo.Name, svcInfo.URL,
				svcInfo.Icon, svcInfo.IconLinkTo, svcInfo.WebURL,
				svcInfo.LatestVersion, svcInfo.ApprovedVersion, svcInfo.DeployedVersion, svcInfo.LatestVersion),
			rootValue:        test.StringPtr("foo-{{ service_id }}-{{ service_name }}-{{ service_url }}--{{ icon }}-{{ icon_link_to }}-{{ web_url }}--{{ version }}-{{ approved_version }}-{{ deployed_version }}-{{ latest_version }}"),
			defaultValue:     test.StringPtr("not_this"),
			hardDefaultValue: test.StringPtr("not_this"),
		},
		"env vars": {
			want:             "this",
			rootValue:        test.StringPtr("${ARGUS_TEST_SHOUTRRR_BUILD_PARAMS}"),
			mainValue:        test.StringPtr("not_this"),
			defaultValue:     test.StringPtr("not_this"),
			hardDefaultValue: test.StringPtr("not_this"),
			envVars: map[string]string{
				"ARGUS_TEST_SHOUTRRR_BUILD_PARAMS": "this",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.envVars {
				os.Setenv(k, v)
				t.Cleanup(func() { os.Unsetenv(k) })
			}

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

			// WHEN BuildParams is called.
			got := shoutrrr.BuildParams(svcInfo)

			// THEN the function returns the params to use.
			if (*got)[key] != tc.want {
				t.Fatalf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestShoutrrr_parseSend(t *testing.T) {
	// GIVEN a possible list of errors from a send operation.
	tests := map[string]struct {
		errs        []error
		serviceName string
		deleting    bool
		wantFailed  bool
		wantErrs    map[string]int
	}{
		"no errors, service not deleting": {
			errs:        []error{nil, nil},
			serviceName: "service1",
			wantFailed:  false,
			wantErrs:    map[string]int{},
		},
		"errors, service not deleting": {
			errs: []error{
				errors.New("error1"),
				errors.New("error2"),
				errors.New("error1")},
			serviceName: "service1",
			wantFailed:  true,
			wantErrs:    map[string]int{"error1": 2, "error2": 1},
		},
		"errors, service deleting": {
			errs: []error{
				errors.New("error1"),
				errors.New("error2")},
			serviceName: "service1",
			deleting:    true,
			wantFailed:  false,
			wantErrs:    map[string]int{},
		},
		"no service name": {
			errs: []error{
				errors.New("error1"),
				errors.New("error1"),
				errors.New("error1")},
			serviceName: "",
			wantFailed:  true,
			wantErrs:    map[string]int{"error1": 3},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're testing metrics.

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.ServiceStatus = &status.Status{}
			if tc.deleting {
				shoutrrr.ServiceStatus.SetDeleting()
			}
			logFrom := logutil.LogFrom{Primary: "test", Secondary: "test"}
			combinedErrs := map[string]int{}

			failed := shoutrrr.parseSend(tc.errs, combinedErrs, tc.serviceName, logFrom)

			if failed != tc.wantFailed {
				t.Errorf("%s\nfailed mismatch\nwant: %t\ngot:  %t",
					packageName, failed, tc.wantFailed)
			}
			if !reflect.DeepEqual(combinedErrs, tc.wantErrs) {
				t.Errorf("%s\ncombinedErrs mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wantErrs, combinedErrs)
			}
		})
	}
}
