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
	"fmt"
	"testing"

	"github.com/release-argus/Argus/util"
)

func TestShoutrrr_GetURL(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		sType     string
		options   map[string]string
		urlFields map[string]string
		params    map[string]string
		want      string
	}{
		"discord - base": {
			sType: "discord",
			want:  "discord://TOKEN@WEBHOOKID",
			urlFields: map[string]string{
				"token":     "TOKEN",
				"webhookid": "WEBHOOKID"},
		},
		"smtp - base": {
			sType: "smtp",
			want:  "smtp://HOST/?fromaddress=FROMADDRESS&toaddresses=TOADDRESS1,TOADDRESS2",
			urlFields: map[string]string{
				"host": "HOST"},
			params: map[string]string{
				"fromaddress": "FROMADDRESS",
				"toaddresses": "TOADDRESS1,TOADDRESS2"},
		},
		"smtp - base + login": {
			sType: "smtp",
			want:  "smtp://USERNAME:PASSWORD@HOST/?fromaddress=FROMADDRESS&toaddresses=TOADDRESS1,TOADDRESS2",
			urlFields: map[string]string{
				"host":     "HOST",
				"username": "USERNAME",
				"password": "PASSWORD"},
			params: map[string]string{
				"fromaddress": "FROMADDRESS",
				"toaddresses": "TOADDRESS1,TOADDRESS2"},
		},
		"smtp - base + login + port": {
			sType: "smtp",
			want:  "smtp://USERNAME:PASSWORD@HOST:587/?fromaddress=FROMADDRESS&toaddresses=TOADDRESS1,TOADDRESS2",
			urlFields: map[string]string{
				"host":     "HOST",
				"username": "USERNAME",
				"password": "PASSWORD",
				"port":     "587"},
			params: map[string]string{
				"fromaddress": "FROMADDRESS",
				"toaddresses": "TOADDRESS1,TOADDRESS2"},
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
		"shoutrrr - base": {
			sType:     "shoutrrr",
			want:      "RAW",
			urlFields: map[string]string{"raw": "RAW"}},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			shoutrrr := testShoutrrr(false, true, false)
			shoutrrr.Type = tc.sType
			shoutrrr.URLFields = tc.urlFields
			shoutrrr.Params = tc.params

			// WHEN GetURL is called
			got := shoutrrr.GetURL()

			// THEN the expected URL is returned
			if got != tc.want {
				t.Errorf("\nwant: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestShoutrrr_GetParams(t *testing.T) {
	// GIVEN a Shoutrrr and ServiceInfo
	serviceInfo := util.ServiceInfo{
		ID:            "service_id",
		LatestVersion: "1.2.3",
	}
	tests := map[string]struct {
		paramsRoot        *string
		paramsMain        *string
		paramsDefault     *string
		paramsHardDefault *string
		wantString        string
	}{
		"root overrides all": {
			wantString:        "this",
			paramsRoot:        stringPtr("this"),
			paramsDefault:     stringPtr("not_this"),
			paramsHardDefault: stringPtr("not_this"),
		},
		"main overrides default and hardDefault": {
			wantString:        "this",
			paramsRoot:        nil,
			paramsMain:        stringPtr("this"),
			paramsDefault:     stringPtr("not_this"),
			paramsHardDefault: stringPtr("not_this"),
		},
		"default overrides hardDefault": {
			wantString:        "this",
			paramsRoot:        nil,
			paramsDefault:     stringPtr("this"),
			paramsHardDefault: stringPtr("not_this"),
		},
		"hardDefault is last resort": {
			wantString:        "this",
			paramsRoot:        nil,
			paramsDefault:     nil,
			paramsHardDefault: stringPtr("this"),
		},
		"jinja templating": {
			wantString:        "this",
			paramsRoot:        stringPtr("{% if 'a' == 'a' %}this{% endif %}"),
			paramsDefault:     stringPtr("not_this"),
			paramsHardDefault: stringPtr("not_this"),
		},
		"jinja vars": {
			wantString:        fmt.Sprintf("foo%s-%s", serviceInfo.ID, serviceInfo.LatestVersion),
			paramsRoot:        stringPtr("foo{{ service_id }}-{{ version }}"),
			paramsDefault:     stringPtr("not_this"),
			paramsHardDefault: stringPtr("not_this"),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			key := "test"
			shoutrrr := testShoutrrr(false, true, false)
			if tc.paramsRoot != nil {
				shoutrrr.Params[key] = *tc.paramsRoot
			}
			if tc.paramsMain != nil {
				shoutrrr.Main.Params[key] = *tc.paramsMain
			}
			if tc.paramsDefault != nil {
				shoutrrr.Defaults.Params[key] = *tc.paramsDefault
			}
			if tc.paramsHardDefault != nil {
				shoutrrr.HardDefaults.Params[key] = *tc.paramsHardDefault
			}

			// WHEN GetParams is called
			got := shoutrrr.GetParams(&serviceInfo)

			// THEN the function returns the params to use
			if (*got)[key] != tc.wantString {
				t.Fatalf("want: %q\ngot:  %q",
					tc.wantString, got)
			}
		})
	}
}
