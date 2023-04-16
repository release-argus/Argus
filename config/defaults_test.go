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

package config

import (
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	opt "github.com/release-argus/Argus/service/options"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

func TestDefaults_String(t *testing.T) {
	// GIVEN a Defaults
	tests := map[string]struct {
		defaults *Defaults
		want     string
	}{
		"nil": {
			defaults: nil,
			want:     "<nil>",
		},
		"empty": {
			defaults: &Defaults{},
			want:     "{}\n",
		},
		"all fields": {
			defaults: &Defaults{
				Service: service.Service{
					Options: opt.Options{
						Interval: "1m",
					},
					DeployedVersionLookup: &deployedver.Lookup{
						URL:  "https://valid.release-argus.io/json",
						JSON: "foo.bar.version"},
					LatestVersion: latestver.Lookup{
						Type: "github",
						URL:  "release-argus/Argus"}},
				Notify: shoutrrr.Slice{
					"discord": {
						Params: map[string]string{
							"username": "Argus"}}},
				WebHook: webhook.WebHook{
					Delay: "0s"},
			},
			want: `
service:
    options:
        interval: 1m
    latest_version:
        type: github
        url: release-argus/Argus
    deployed_version:
        url: https://valid.release-argus.io/json
        json: foo.bar.version
notify:
    discord:
        params:
            username: Argus
webhook:
    delay: 0s
`,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the Defaults are stringified with String()
			got := tc.defaults.String()

			// THEN the result is as expected
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if got != tc.want {
				t.Errorf("want: %q\ngot:\n%q",
					tc.want, got)
			}
		})
	}
}

func TestDefaults_SetDefaults(t *testing.T) {
	// GIVEN nil defaults
	var defaults Defaults

	// WHEN SetDefaults is called on it
	defaults.SetDefaults()
	tests := map[string]struct {
		got  string
		want string
	}{
		"Service.Interval": {
			got:  defaults.Service.Options.Interval,
			want: "10m"},
		"Notify.discord.username": {
			got:  defaults.Notify["discord"].GetSelfParam("username"),
			want: "Argus"},
		"WebHook.Delay": {
			got:  defaults.WebHook.Delay,
			want: "0s"},
	}

	// THEN the defaults are set correctly
	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.got != tc.want {
				t.Log(name)
				t.Errorf("want: %s\ngot:  %s",
					tc.want, tc.got)
			}
		})
	}
}

func TestDefaults_MapEnvToStruct(t *testing.T) {
	var defaults Defaults
	defaults.SetDefaults()
	// GIVEN a defaults and a bunch of env vars
	test := map[string]struct {
		env      map[string]string
		want     *Defaults
		errRegex string
	}{
		"Service.options": {
			env: map[string]string{
				"ARGUS_SERVICE_OPTIONS_INTERVAL":            "99m",
				"ARGUS_SERVICE_OPTIONS_SEMANTIC_VERSIONING": "true"},
			want: &Defaults{
				Service: service.Service{
					Options: opt.Options{
						Interval:           "99m",
						SemanticVersioning: boolPtr(true)}}},
		},
		"Service.options - invalid time.duration - interval": {
			env: map[string]string{
				"ARGUS_SERVICE_OPTIONS_INTERVAL":            "99 something",
				"ARGUS_SERVICE_OPTIONS_SEMANTIC_VERSIONING": "true"},
			errRegex: `interval: "[^"]+" <invalid>`,
		},
		"Service.options - invalid bool - semantic version": {
			env: map[string]string{
				"ARGUS_SERVICE_OPTIONS_INTERVAL":            "99",
				"ARGUS_SERVICE_OPTIONS_SEMANTIC_VERSIONING": "foo"},
			errRegex: `invalid bool for `,
		},
		"service.latest_version": {
			env: map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_ALLOW_INVALID_CERTS": "true",
				"ARGUS_SERVICE_LATEST_VERSION_USE_PRERELEASE":      "true"},
			want: &Defaults{
				Service: service.Service{
					LatestVersion: latestver.Lookup{
						AllowInvalidCerts: boolPtr(true),
						UsePreRelease:     boolPtr(true)}}},
		},
		"service.latest_version - invalid bool - allow_invalid_certs": {
			env: map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_ALLOW_INVALID_CERTS": "bar",
				"ARGUS_SERVICE_LATEST_VERSION_USE_PRERELEASE":      "true"},
			errRegex: `invalid bool for [^:]+`,
		},
		"service.latest_version - invalid bool - use_prerelease": {
			env: map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_ALLOW_INVALID_CERTS": "true",
				"ARGUS_SERVICE_LATEST_VERSION_USE_PRERELEASE":      "bop"},
			errRegex: `invalid bool for [^:]+`,
		},
		"service.deployed_version": {
			env: map[string]string{
				"ARGUS_SERVICE_DEPLOYED_VERSION_ALLOW_INVALID_CERTS": "true"},
			want: &Defaults{
				Service: service.Service{
					DeployedVersionLookup: &deployedver.Lookup{
						AllowInvalidCerts: boolPtr(true)}}},
		},
		"service.deployed_version - invalid bool - allow_invalid_certs": {
			env: map[string]string{
				"ARGUS_SERVICE_DEPLOYED_VERSION_ALLOW_INVALID_CERTS": "bang"},
			errRegex: `invalid bool for [^:]+`,
		},
		"service.dashboard": {
			env: map[string]string{
				"ARGUS_SERVICE_DASHBOARD_AUTO_APPROVE": "true"},
			want: &Defaults{
				Service: service.Service{
					Dashboard: service.DashboardOptions{
						AutoApprove: boolPtr(true)}}},
		},
		"service.dashboard - invalid bool - auto_approve": {
			env: map[string]string{
				"ARGUS_SERVICE_DASHBOARD_AUTO_APPROVE": "zap"},
			errRegex: `invalid bool for [^:]+`,
		},
		"notify.discord": {
			env: map[string]string{
				"ARGUS_NOTIFY_DISCORD_OPTIONS_MESSAGE":      "bish",
				"ARGUS_NOTIFY_DISCORD_OPTIONS_MAX_TRIES":    "1",
				"ARGUS_NOTIFY_DISCORD_OPTIONS_DELAY":        "1h",
				"ARGUS_NOTIFY_DISCORD_URL_FIELDS_TOKEN":     "foo",
				"ARGUS_NOTIFY_DISCORD_URL_FIELDS_WEBHOOKID": "bar",
				"ARGUS_NOTIFY_DISCORD_PARAMS_AVATAR":        ":argus:",
				"ARGUS_NOTIFY_DISCORD_PARAMS_COLOR":         "0x50D9ff",
				"ARGUS_NOTIFY_DISCORD_PARAMS_COLORDEBUG":    "0x7b00ab",
				"ARGUS_NOTIFY_DISCORD_PARAMS_COLORERROR":    "0xd60510",
				"ARGUS_NOTIFY_DISCORD_PARAMS_COLORINFO":     "0x2488ff",
				"ARGUS_NOTIFY_DISCORD_PARAMS_COLORWARN":     "0xffc441",
				"ARGUS_NOTIFY_DISCORD_PARAMS_JSON":          "no",
				"ARGUS_NOTIFY_DISCORD_PARAMS_SPLITLINES":    "yes",
				"ARGUS_NOTIFY_DISCORD_PARAMS_TITLE":         "something",
				"ARGUS_NOTIFY_DISCORD_PARAMS_USERNAME":      "test",
			},
			want: &Defaults{
				Notify: shoutrrr.Slice{
					"discord": {
						Options: map[string]string{
							"message":   "bish",
							"max_tries": "1",
							"delay":     "1h"},
						URLFields: map[string]string{
							"token":     "foo",
							"webhookid": "bar"},
						Params: map[string]string{
							"avatar":     ":argus:",
							"color":      "0x50D9ff",
							"colordebug": "0x7b00ab",
							"colorerror": "0xd60510",
							"colorinfo":  "0x2488ff",
							"colorwarn":  "0xffc441",
							"json":       "no",
							"splitlines": "yes",
							"title":      "something",
							"username":   "test"}}}},
		},
		"notify.discord - invalid options.delay": {
			env: map[string]string{
				"ARGUS_NOTIFY_DISCORD_OPTIONS_DELAY": "foo"},
			errRegex: `discord.*delay: "[^"]+" <invalid>`,
		},
		"notify.smtp": {
			env: map[string]string{
				"ARGUS_NOTIFY_SMTP_OPTIONS_MESSAGE":     "bing",
				"ARGUS_NOTIFY_SMTP_OPTIONS_MAX_TRIES":   "2",
				"ARGUS_NOTIFY_SMTP_OPTIONS_DELAY":       "2m",
				"ARGUS_NOTIFY_SMTP_URL_FIELDS_USERNAME": "user",
				"ARGUS_NOTIFY_SMTP_URL_FIELDS_PASSWORD": "secret",
				"ARGUS_NOTIFY_SMTP_URL_FIELDS_HOST":     "https://smtp.example.com",
				"ARGUS_NOTIFY_SMTP_URL_FIELDS_PORT":     "25",
				"ARGUS_NOTIFY_SMTP_PARAMS_FROMADDRESS":  "me@example.com",
				"ARGUS_NOTIFY_SMTP_PARAMS_TOADDRESSES":  "you@somewhere.com",
				"ARGUS_NOTIFY_SMTP_PARAMS_AUTH":         "Unknown",
				"ARGUS_NOTIFY_SMTP_PARAMS_CLIENTHOST":   "localhost",
				"ARGUS_NOTIFY_SMTP_PARAMS_ENCRYPTION":   "auto",
				"ARGUS_NOTIFY_SMTP_PARAMS_FROMNAME":     "someone",
				"ARGUS_NOTIFY_SMTP_PARAMS_SUBJECT":      "Argus SMTP Notification",
				"ARGUS_NOTIFY_SMTP_PARAMS_USEHTML":      "no",
				"ARGUS_NOTIFY_SMTP_PARAMS_USESTARTTLS":  "yes"},
			want: &Defaults{
				Notify: shoutrrr.Slice{
					"smtp": {
						Options: map[string]string{
							"message":   "bing",
							"max_tries": "2",
							"delay":     "2m"},
						URLFields: map[string]string{
							"username": "user",
							"password": "secret",
							"host":     "https://smtp.example.com",
							"port":     "25"},
						Params: map[string]string{
							"fromaddress": "me@example.com",
							"toaddresses": "you@somewhere.com",
							"auth":        "Unknown",
							"clienthost":  "localhost",
							"encryption":  "auto",
							"fromname":    "someone",
							"subject":     "Argus SMTP Notification",
							"usehtml":     "no",
							"usestarttls": "yes"}}}},
		},
		"notify.gotify": {
			env: map[string]string{
				"ARGUS_NOTIFY_GOTIFY_OPTIONS_MESSAGE":   "shazam",
				"ARGUS_NOTIFY_GOTIFY_OPTIONS_MAX_TRIES": "3",
				"ARGUS_NOTIFY_GOTIFY_OPTIONS_DELAY":     "3s",
				"ARGUS_NOTIFY_GOTIFY_URL_FIELDS_HOST":   "https://gotify.example.com",
				"ARGUS_NOTIFY_GOTIFY_URL_FIELDS_PORT":   "443",
				"ARGUS_NOTIFY_GOTIFY_URL_FIELDS_PATH":   "gotify",
				"ARGUS_NOTIFY_GOTIFY_URL_FIELDS_TOKEN":  "SuperSecretToken",
				"ARGUS_NOTIFY_GOTIFY_PARAMS_DISABLETLS": "no",
				"ARGUS_NOTIFY_GOTIFY_PARAMS_PRIORITY":   "0",
				"ARGUS_NOTIFY_GOTIFY_PARAMS_TITLE":      "Argus Gotify Notification"},
			want: &Defaults{
				Notify: shoutrrr.Slice{
					"gotify": {
						Options: map[string]string{
							"message":   "shazam",
							"max_tries": "3",
							"delay":     "3s"},
						URLFields: map[string]string{
							"host":  "https://gotify.example.com",
							"port":  "443",
							"path":  "gotify",
							"token": "SuperSecretToken"},
						Params: map[string]string{
							"disabletls": "no",
							"priority":   "0",
							"title":      "Argus Gotify Notification"}}}},
		},
		"notify.googlechat": {
			env: map[string]string{
				"ARGUS_NOTIFY_GOOGLECHAT_OPTIONS_MESSAGE":   "whoosh",
				"ARGUS_NOTIFY_GOOGLECHAT_OPTIONS_MAX_TRIES": "4",
				"ARGUS_NOTIFY_GOOGLECHAT_OPTIONS_DELAY":     "4h",
				"ARGUS_NOTIFY_GOOGLECHAT_URL_FIELDS_RAW":    "chat.googleapis.com/v1/spaces/FOO/messages?key=bar&token=baz"},
			want: &Defaults{
				Notify: shoutrrr.Slice{
					"googlechat": {
						Options: map[string]string{
							"message":   "whoosh",
							"max_tries": "4",
							"delay":     "4h"},
						URLFields: map[string]string{
							"raw": "chat.googleapis.com/v1/spaces/FOO/messages?key=bar&token=baz"}}}},
		},
		"notify.ifttt": {
			env: map[string]string{
				"ARGUS_NOTIFY_IFTTT_OPTIONS_MESSAGE":          "pow",
				"ARGUS_NOTIFY_IFTTT_OPTIONS_MAX_TRIES":        "5",
				"ARGUS_NOTIFY_IFTTT_OPTIONS_DELAY":            "5m",
				"ARGUS_NOTIFY_IFTTT_URL_FIELDS_WEBHOOKID":     "secretWHID",
				"ARGUS_NOTIFY_IFTTT_PARAMS_EVENTS":            "event1,event2",
				"ARGUS_NOTIFY_IFTTT_PARAMS_TITLE":             "Argus IFTTT Notification",
				"ARGUS_NOTIFY_IFTTT_PARAMS_USEMESSAGEASVALUE": "2",
				"ARGUS_NOTIFY_IFTTT_PARAMS_USETITLEASVALUE":   "0",
				"ARGUS_NOTIFY_IFTTT_PARAMS_VALUE1":            "bish",
				"ARGUS_NOTIFY_IFTTT_PARAMS_VALUE2":            "bash",
				"ARGUS_NOTIFY_IFTTT_PARAMS_VALUE3":            "bosh"},
			want: &Defaults{
				Notify: shoutrrr.Slice{
					"ifttt": {
						Options: map[string]string{
							"message":   "pow",
							"max_tries": "5",
							"delay":     "5m"},
						URLFields: map[string]string{
							"webhookid": "secretWHID"},
						Params: map[string]string{
							"events":            "event1,event2",
							"title":             "Argus IFTTT Notification",
							"usemessageasvalue": "2",
							"usetitleasvalue":   "0",
							"value1":            "bish",
							"value2":            "bash",
							"value3":            "bosh"}}}},
		},
		"notify.join": {
			env: map[string]string{
				"ARGUS_NOTIFY_JOIN_OPTIONS_MESSAGE":   "pew",
				"ARGUS_NOTIFY_JOIN_OPTIONS_MAX_TRIES": "6",
				"ARGUS_NOTIFY_JOIN_OPTIONS_DELAY":     "6s",
				"ARGUS_NOTIFY_JOIN_URL_FIELDS_APIKEY": "apiKey",
				"ARGUS_NOTIFY_JOIN_PARAMS_DEVICES":    "device1,device2",
				"ARGUS_NOTIFY_JOIN_PARAMS_ICON":       "https://example.com/icon.png",
				"ARGUS_NOTIFY_JOIN_PARAMS_TITLE":      "Argus Join Notification"},
			want: &Defaults{
				Notify: shoutrrr.Slice{
					"join": {
						Options: map[string]string{
							"message":   "pew",
							"max_tries": "6",
							"delay":     "6s"},
						URLFields: map[string]string{
							"apikey": "apiKey"},
						Params: map[string]string{
							"devices": "device1,device2",
							"icon":    "https://example.com/icon.png",
							"title":   "Argus Join Notification"}}}},
		},
		"notify.mattermost": {
			env: map[string]string{
				"ARGUS_NOTIFY_MATTERMOST_OPTIONS_MESSAGE":     "ping",
				"ARGUS_NOTIFY_MATTERMOST_OPTIONS_MAX_TRIES":   "7",
				"ARGUS_NOTIFY_MATTERMOST_OPTIONS_DELAY":       "7h",
				"ARGUS_NOTIFY_MATTERMOST_URL_FIELDS_USERNAME": "Argus",
				"ARGUS_NOTIFY_MATTERMOST_URL_FIELDS_HOST":     "https://mattermost.example.com",
				"ARGUS_NOTIFY_MATTERMOST_URL_FIELDS_PORT":     "443",
				"ARGUS_NOTIFY_MATTERMOST_URL_FIELDS_PATH":     "mattermost",
				"ARGUS_NOTIFY_MATTERMOST_URL_FIELDS_TOKEN":    "mattermostToken",
				"ARGUS_NOTIFY_MATTERMOST_URL_FIELDS_CHANNEL":  "argus",
				"ARGUS_NOTIFY_MATTERMOST_PARAMS_ICON":         ":argus:",
				"ARGUS_NOTIFY_MATTERMOST_PARAMS_TITLE":        "Argus Mattermost Notification"},
			want: &Defaults{
				Notify: shoutrrr.Slice{
					"mattermost": {
						Options: map[string]string{
							"message":   "ping",
							"max_tries": "7",
							"delay":     "7h"},
						URLFields: map[string]string{
							"username": "Argus",
							"host":     "https://mattermost.example.com",
							"port":     "443",
							"path":     "mattermost",
							"token":    "mattermostToken",
							"channel":  "argus"},
						Params: map[string]string{
							"icon":  ":argus:",
							"title": "Argus Mattermost Notification"}}}},
		},
		"notify.matrix": {
			env: map[string]string{
				"ARGUS_NOTIFY_MATRIX_OPTIONS_MESSAGE":     "pong",
				"ARGUS_NOTIFY_MATRIX_OPTIONS_MAX_TRIES":   "8",
				"ARGUS_NOTIFY_MATRIX_OPTIONS_DELAY":       "8m",
				"ARGUS_NOTIFY_MATRIX_URL_FIELDS_USER":     "argus",
				"ARGUS_NOTIFY_MATRIX_URL_FIELDS_HOST":     "https://matrix.example.com",
				"ARGUS_NOTIFY_MATRIX_URL_FIELDS_PORT":     "443",
				"ARGUS_NOTIFY_MATRIX_URL_FIELDS_PATH":     "matrix",
				"ARGUS_NOTIFY_MATRIX_URL_FIELDS_PASSWORD": "matrixPassword",
				"ARGUS_NOTIFY_MATRIX_PARAMS_DISABLETLS":   "no",
				"ARGUS_NOTIFY_MATRIX_PARAMS_ROOMS":        "room1,room2",
				"ARGUS_NOTIFY_MATRIX_PARAMS_TITLE":        "Argus Matrix Notification"},
			want: &Defaults{
				Notify: shoutrrr.Slice{
					"matrix": {
						Options: map[string]string{
							"message":   "pong",
							"max_tries": "8",
							"delay":     "8m"},
						URLFields: map[string]string{
							"user":     "argus",
							"host":     "https://matrix.example.com",
							"port":     "443",
							"path":     "matrix",
							"password": "matrixPassword"},
						Params: map[string]string{
							"disabletls": "no",
							"rooms":      "room1,room2",
							"title":      "Argus Matrix Notification"}}}},
		},
		"notify.opsgenie": {
			env: map[string]string{
				"ARGUS_NOTIFY_OPSGENIE_OPTIONS_MESSAGE":    "pang",
				"ARGUS_NOTIFY_OPSGENIE_OPTIONS_MAX_TRIES":  "9",
				"ARGUS_NOTIFY_OPSGENIE_OPTIONS_DELAY":      "9s",
				"ARGUS_NOTIFY_OPSGENIE_URL_FIELDS_HOST":    "https://opsgenie.example.com",
				"ARGUS_NOTIFY_OPSGENIE_URL_FIELDS_PORT":    "443",
				"ARGUS_NOTIFY_OPSGENIE_URL_FIELDS_PATH":    "opsgenie",
				"ARGUS_NOTIFY_OPSGENIE_URL_FIELDS_APIKEY":  "opsGenieApiKey",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_ACTIONS":     "action1,action2",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_ALIAS":       "argus",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_DESCRIPTION": "Argus OpsGenie DESC",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_DETAILS":     "foo=bar",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_ENTITY":      "argus",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_NOTE":        "testing OpsGenie",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_PRIORITY":    "P1",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_RESPONDERS":  "responder1,responder2",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_SOURCE":      "argus",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_TAGS":        "tag1,tag2",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_TITLE":       "Argus OpsGenie Notification",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_USER":        "argus",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_VISIBLETO":   "visible1,visible2"},
			want: &Defaults{
				Notify: shoutrrr.Slice{
					"opsgenie": {
						Options: map[string]string{
							"message":   "pang",
							"max_tries": "9",
							"delay":     "9s"},
						URLFields: map[string]string{
							"host":   "https://opsgenie.example.com",
							"port":   "443",
							"path":   "opsgenie",
							"apikey": "opsGenieApiKey"},
						Params: map[string]string{
							"actions":     "action1,action2",
							"alias":       "argus",
							"description": "Argus OpsGenie DESC",
							"details":     "foo=bar",
							"entity":      "argus",
							"note":        "testing OpsGenie",
							"priority":    "P1",
							"responders":  "responder1,responder2",
							"source":      "argus",
							"tags":        "tag1,tag2",
							"title":       "Argus OpsGenie Notification",
							"user":        "argus",
							"visibleto":   "visible1,visible2"}}}},
		},
		"notify.pushbullet": {
			env: map[string]string{
				"ARGUS_NOTIFY_PUSHBULLET_OPTIONS_MESSAGE":    "pung",
				"ARGUS_NOTIFY_PUSHBULLET_OPTIONS_MAX_TRIES":  "10",
				"ARGUS_NOTIFY_PUSHBULLET_OPTIONS_DELAY":      "10h",
				"ARGUS_NOTIFY_PUSHBULLET_URL_FIELDS_TOKEN":   "pushbulletToken",
				"ARGUS_NOTIFY_PUSHBULLET_URL_FIELDS_TARGETS": "target1,target2",
				"ARGUS_NOTIFY_PUSHBULLET_PARAMS_TITLE":       "Argus Pushbullet Notification"},
			want: &Defaults{
				Notify: shoutrrr.Slice{
					"pushbullet": {
						Options: map[string]string{
							"message":   "pung",
							"max_tries": "10",
							"delay":     "10h"},
						URLFields: map[string]string{
							"token":   "pushbulletToken",
							"targets": "target1,target2"},
						Params: map[string]string{
							"title": "Argus Pushbullet Notification"}}}},
		},
		"notify.pushover": {
			env: map[string]string{
				"ARGUS_NOTIFY_PUSHOVER_OPTIONS_MESSAGE":   "pung",
				"ARGUS_NOTIFY_PUSHOVER_OPTIONS_MAX_TRIES": "11",
				"ARGUS_NOTIFY_PUSHOVER_OPTIONS_DELAY":     "11m",
				"ARGUS_NOTIFY_PUSHOVER_URL_FIELDS_TOKEN":  "pushoverToken",
				"ARGUS_NOTIFY_PUSHOVER_URL_FIELDS_USER":   "pushoverUser",
				"ARGUS_NOTIFY_PUSHOVER_PARAMS_DEVICES":    "device1,device2",
				"ARGUS_NOTIFY_PUSHOVER_PARAMS_PRIORITY":   "0",
				"ARGUS_NOTIFY_PUSHOVER_PARAMS_TITLE":      "Argus Pushbullet Notification"},
			want: &Defaults{
				Notify: shoutrrr.Slice{
					"pushover": {
						Options: map[string]string{
							"message":   "pung",
							"max_tries": "11",
							"delay":     "11m"},
						URLFields: map[string]string{
							"token": "pushoverToken",
							"user":  "pushoverUser"},
						Params: map[string]string{
							"devices":  "device1,device2",
							"priority": "0",
							"title":    "Argus Pushbullet Notification"}}}},
		},
		"notify.rocketchat": {
			env: map[string]string{
				"ARGUS_NOTIFY_ROCKETCHAT_OPTIONS_MESSAGE":     "pung",
				"ARGUS_NOTIFY_ROCKETCHAT_OPTIONS_MAX_TRIES":   "12",
				"ARGUS_NOTIFY_ROCKETCHAT_OPTIONS_DELAY":       "12s",
				"ARGUS_NOTIFY_ROCKETCHAT_URL_FIELDS_USERNAME": "rocketchatUser",
				"ARGUS_NOTIFY_ROCKETCHAT_URL_FIELDS_HOST":     "rocketchat.example.com",
				"ARGUS_NOTIFY_ROCKETCHAT_URL_FIELDS_PORT":     "443",
				"ARGUS_NOTIFY_ROCKETCHAT_URL_FIELDS_PATH":     "rocketchat",
				"ARGUS_NOTIFY_ROCKETCHAT_URL_FIELDS_TOKENA":   "FIRST_token",
				"ARGUS_NOTIFY_ROCKETCHAT_URL_FIELDS_TOKENB":   "SECOND_token",
				"ARGUS_NOTIFY_ROCKETCHAT_URL_FIELDS_CHANNEL":  "rocketchatChannel"},
			want: &Defaults{
				Notify: shoutrrr.Slice{
					"rocketchat": {
						Options: map[string]string{
							"message":   "pung",
							"max_tries": "12",
							"delay":     "12s"},
						URLFields: map[string]string{
							"username": "rocketchatUser",
							"host":     "rocketchat.example.com",
							"port":     "443",
							"path":     "rocketchat",
							"tokena":   "FIRST_token",
							"tokenb":   "SECOND_token",
							"channel":  "rocketchatChannel"}}}},
		},
		"notify.slack": {
			env: map[string]string{
				"ARGUS_NOTIFY_SLACK_OPTIONS_MESSAGE":    "slung",
				"ARGUS_NOTIFY_SLACK_OPTIONS_MAX_TRIES":  "13",
				"ARGUS_NOTIFY_SLACK_OPTIONS_DELAY":      "13h",
				"ARGUS_NOTIFY_SLACK_URL_FIELDS_TOKEN":   "slackToken",
				"ARGUS_NOTIFY_SLACK_URL_FIELDS_CHANNEL": "somewhere",
				"ARGUS_NOTIFY_SLACK_PARAMS_BOTNAME":     "Argus",
				"ARGUS_NOTIFY_SLACK_PARAMS_COLOR":       "#ff8000",
				"ARGUS_NOTIFY_SLACK_PARAMS_ICON":        ":ghost:",
				"ARGUS_NOTIFY_SLACK_PARAMS_THREADTS":    "1234567890.123456",
				"ARGUS_NOTIFY_SLACK_PARAMS_TITLE":       "Argus Slack Notification"},
			want: &Defaults{
				Notify: shoutrrr.Slice{
					"slack": {
						Options: map[string]string{
							"message":   "slung",
							"max_tries": "13",
							"delay":     "13h"},
						URLFields: map[string]string{
							"token":   "slackToken",
							"channel": "somewhere"},
						Params: map[string]string{
							"botname":  "Argus",
							"color":    "#ff8000",
							"icon":     ":ghost:",
							"threadts": "1234567890.123456",
							"title":    "Argus Slack Notification"}}}},
		},
		"notify.teams": {
			env: map[string]string{
				"ARGUS_NOTIFY_TEAMS_OPTIONS_MESSAGE":       "tung",
				"ARGUS_NOTIFY_TEAMS_OPTIONS_MAX_TRIES":     "14",
				"ARGUS_NOTIFY_TEAMS_OPTIONS_DELAY":         "14m",
				"ARGUS_NOTIFY_TEAMS_URL_FIELDS_GROUP":      "teamsGroup",
				"ARGUS_NOTIFY_TEAMS_URL_FIELDS_TENANT":     "tenant",
				"ARGUS_NOTIFY_TEAMS_URL_FIELDS_ALTID":      "otherID?",
				"ARGUS_NOTIFY_TEAMS_URL_FIELDS_GROUPOWNER": "owner",
				"ARGUS_NOTIFY_TEAMS_PARAMS_COLOR":          "#ff8000",
				"ARGUS_NOTIFY_TEAMS_PARAMS_HOST":           "teams.example.com",
				"ARGUS_NOTIFY_TEAMS_PARAMS_TITLE":          "Argus Teams Notification"},
			want: &Defaults{
				Notify: shoutrrr.Slice{
					"teams": {
						Options: map[string]string{
							"message":   "tung",
							"max_tries": "14",
							"delay":     "14m"},
						URLFields: map[string]string{
							"group":      "teamsGroup",
							"tenant":     "tenant",
							"altid":      "otherID?",
							"groupowner": "owner"},
						Params: map[string]string{
							"color": "#ff8000",
							"host":  "teams.example.com",
							"title": "Argus Teams Notification"}}}},
		},
		"notify.telegram": {
			env: map[string]string{
				"ARGUS_NOTIFY_TELEGRAM_OPTIONS_MESSAGE":     "tong",
				"ARGUS_NOTIFY_TELEGRAM_OPTIONS_MAX_TRIES":   "15",
				"ARGUS_NOTIFY_TELEGRAM_OPTIONS_DELAY":       "15s",
				"ARGUS_NOTIFY_TELEGRAM_URL_FIELDS_TOKEN":    "telegramToken",
				"ARGUS_NOTIFY_TELEGRAM_PARAMS_CHATS":        "chat1,chat2",
				"ARGUS_NOTIFY_TELEGRAM_PARAMS_NOTIFICATION": "yes",
				"ARGUS_NOTIFY_TELEGRAM_PARAMS_PARSEMODE":    "None",
				"ARGUS_NOTIFY_TELEGRAM_PARAMS_PREVIEW":      "yes",
				"ARGUS_NOTIFY_TELEGRAM_PARAMS_TITLE":        "Argus Telegram Notification"},
			want: &Defaults{
				Notify: shoutrrr.Slice{
					"telegram": {
						Options: map[string]string{
							"message":   "tong",
							"max_tries": "15",
							"delay":     "15s"},
						URLFields: map[string]string{
							"token": "telegramToken"},
						Params: map[string]string{
							"chats":        "chat1,chat2",
							"notification": "yes",
							"parsemode":    "None",
							"preview":      "yes",
							"title":        "Argus Telegram Notification"}}}},
		},
		"notify.zulip": {
			env: map[string]string{
				"ARGUS_NOTIFY_ZULIP_OPTIONS_MESSAGE":    "zung",
				"ARGUS_NOTIFY_ZULIP_OPTIONS_MAX_TRIES":  "16",
				"ARGUS_NOTIFY_ZULIP_OPTIONS_DELAY":      "16h",
				"ARGUS_NOTIFY_ZULIP_URL_FIELDS_BOTMAIL": "botmail",
				"ARGUS_NOTIFY_ZULIP_URL_FIELDS_BOTKEY":  "botkey",
				"ARGUS_NOTIFY_ZULIP_URL_FIELDS_HOST":    "zulip.example.com",
				"ARGUS_NOTIFY_ZULIP_URL_FIELDS_PORT":    "1234",
				"ARGUS_NOTIFY_ZULIP_URL_FIELDS_PATH":    "zulip",
				"ARGUS_NOTIFY_ZULIP_PARAMS_STREAM":      "stream",
				"ARGUS_NOTIFY_ZULIP_PARAMS_TOPIC":       "topic"},
			want: &Defaults{
				Notify: shoutrrr.Slice{
					"zulip": {
						Options: map[string]string{
							"message":   "zung",
							"max_tries": "16",
							"delay":     "16h"},
						URLFields: map[string]string{
							"botmail": "botmail",
							"botkey":  "botkey",
							"host":    "zulip.example.com",
							"port":    "1234",
							"path":    "zulip"},
						Params: map[string]string{
							"stream": "stream",
							"topic":  "topic"}}}},
		},
		"webhook": {
			env: map[string]string{
				"ARGUS_WEBHOOK_TYPE":                "gitlab",
				"ARGUS_WEBHOOK_DELAY":               "99s",
				"ARGUS_WEBHOOK_MAX_TRIES":           "88",
				"ARGUS_WEBHOOK_ALLOW_INVALID_CERTS": "true",
				"ARGUS_WEBHOOK_DESIRED_STATUS_CODE": "201",
				"ARGUS_WEBHOOK_SILENT_FAILS":        "true"},
			want: &Defaults{
				WebHook: webhook.WebHook{
					Type:              "gitlab",
					Delay:             "99s",
					MaxTries:          uintPtr(88),
					AllowInvalidCerts: boolPtr(true),
					DesiredStatusCode: intPtr(201),
					SilentFails:       boolPtr(true)}},
		},
		"webhook - invalid str - type": {
			env: map[string]string{
				"ARGUS_WEBHOOK_TYPE": "pizza"},
			errRegex: "a",
		},
		"webhook - invalid time.duration - delay": {
			env: map[string]string{
				"ARGUS_WEBHOOK_DELAY": "pasta"},
			errRegex: `delay: "[^"]+" <invalid>`,
		},
		"webhook - invalid uint - max_tries": {
			env: map[string]string{
				"ARGUS_WEBHOOK_MAX_TRIES": "-1"},
			errRegex: "invalid uint for ",
		},
		"webhook - invalid bool - allow_invalid_certs": {
			env: map[string]string{
				"ARGUS_WEBHOOK_ALLOW_INVALID_CERTS": "foo"},
			errRegex: "invalid bool for ",
		},
		"webhook - invalid int - desired_status_code": {
			env: map[string]string{
				"ARGUS_WEBHOOK_DESIRED_STATUS_CODE": "okay"},
			errRegex: "invalid integer for ",
		},
		"webhook - invalid bool - silent_fails": {
			env: map[string]string{
				"ARGUS_WEBHOOK_SILENT_FAILS": "bar"},
			errRegex: "invalid bool for ",
		},
		"no env vars": {
			want: &Defaults{},
		},
		"no 'ARGUS_' env vars": {
			env: map[string]string{
				"NOT_ARGUS": "foo"},
			want: &Defaults{},
		},
	}

	for name, tc := range test {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {

			dflts := Defaults{
				Service: service.Service{
					DeployedVersionLookup: &deployedver.Lookup{}}}
			if tc.want == nil {
				tc.want = &Defaults{
					Notify: shoutrrr.Slice{}}
			}
			if tc.want.Service.DeployedVersionLookup == nil {
				tc.want.Service.DeployedVersionLookup = &deployedver.Lookup{}
			}
			if tc.want.Notify != nil {
				dflts.Notify = shoutrrr.Slice{}
				for notifyType := range defaults.Notify {
					dflts.Notify[notifyType] = &shoutrrr.Shoutrrr{}
					dflts.Notify[notifyType].InitMaps()
					if tc.want.Notify[notifyType] == nil {
						tc.want.Notify[notifyType] = &shoutrrr.Shoutrrr{}
						tc.want.Notify[notifyType].InitMaps()
					}
				}
			}
			for k, v := range tc.env {
				os.Setenv(k, v)
			}
			// Cleanup
			defer func() {
				for k := range tc.env {
					os.Unsetenv(k)
				}
			}()

			// WHEN MapEnvToStruct is called on it
			err := dflts.MapEnvToStruct()

			// THEN any error is as expected
			if tc.errRegex == "" {
				tc.errRegex = `^$`
			}
			errStr := util.ErrorToString(err)
			if !regexp.MustCompile(tc.errRegex).MatchString(errStr) {
				t.Errorf("want error matching:\n%v\ngot:\n%v",
					tc.errRegex, errStr)
			}
			if tc.errRegex != `^$` {
				return // no further checks if there was an error
			}
			// AND the defaults are set to the appropriate env vars
			if dflts.String() != tc.want.String() {
				t.Errorf("want:\n%v\ngot:\n%v",
					tc.want.String(), dflts.String())
			}
		})
	}
}

func TestDefaults_CheckValues(t *testing.T) {
	// GIVEN defaults with a test of invalid vars
	var defaults Defaults
	defaults.SetDefaults()
	tests := map[string]struct {
		input       *Defaults
		errContains []string
	}{
		"Service.Interval": {
			input: &Defaults{Service: service.Service{
				Options: opt.Options{
					Interval: "10x"}}},
			errContains: []string{
				`^  service:$`,
				`^      interval: "10x" <invalid>`},
		},
		"Service.DeployedVersionLookup.Regex": {
			input: &Defaults{Service: service.Service{
				DeployedVersionLookup: &deployedver.Lookup{
					Regex: `^something[0-`}}},
			errContains: []string{
				`^  service:$`,
				`^    deployed_version:$`,
				`^      regex: "\^something\[0\-" <invalid>`},
		},
		"Service.Interval + Service.DeployedVersionLookup.Regex": {
			input: &Defaults{Service: service.Service{
				Options: opt.Options{
					Interval: "10x"},
				DeployedVersionLookup: &deployedver.Lookup{
					Regex: `^something[0-`}}},
			errContains: []string{
				`^  service:$`,
				`^    deployed_version:$`,
				`^      regex: "\^something\[0\-" <invalid>`},
		},
		"Notify.x.Delay": {
			input: &Defaults{Notify: shoutrrr.Slice{
				"slack": &shoutrrr.Shoutrrr{
					Options: map[string]string{"delay": "10x"}}}},
			errContains: []string{
				`^  notify:$`,
				`^    slack:$`,
				`^      options:`,
				`^        delay: "10x" <invalid>`},
		},
		"WebHook.Delay": {
			input: &Defaults{WebHook: webhook.WebHook{
				Delay: "10x"}},
			errContains: []string{
				`^  webhook:$`,
				`^  delay: "10x" <invalid>`},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CheckValues is called on it
			err := util.ErrorToString(tc.input.CheckValues())

			// THEN err matches expected
			lines := strings.Split(err, "\\")
			for i := range tc.errContains {
				re := regexp.MustCompile(tc.errContains[i])
				found := false
				for j := range lines {
					match := re.MatchString(lines[j])
					if match {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("invalid %s should have errored:\nwant: %q\ngot:  %q",
						name, tc.errContains[i], strings.ReplaceAll(err, `\`, "\n"))
				}
			}
		})
	}
}

func TestDefaults_Print(t *testing.T) {
	// GIVEN unmodified defaults from SetDefaults
	var defaults Defaults
	defaults.SetDefaults()
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Print is called
	defaults.Print()

	// THEN the expected number of lines are printed
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = stdout
	want := 129
	got := strings.Count(string(out), "\n")
	if got != want {
		t.Errorf("Print should have given %d lines, but gave %d\n%s", want, got, out)
	}
}
