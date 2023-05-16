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
	"github.com/release-argus/Argus/service/latest_version/filter"
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
			want:     "",
		},
		"empty": {
			defaults: &Defaults{},
			want:     "{}",
		},
		"all fields": {
			defaults: &Defaults{
				Service: service.Defaults{
					Options: *opt.NewDefaults(
						"1m",            // interval
						boolPtr(false)), // semantic_versioning
					LatestVersion: *latestver.NewDefaults(
						stringPtr("foo"), // access_token
						boolPtr(true),    // allow_invalid_certs
						boolPtr(false),   // use_prerelease
						filter.NewRequireDefaults(
							filter.NewDockerCheckDefaults(
								"ghcr", // type
								"tokenGHCR",
								"tokenHub", "usernameHub",
								"tokenQuay",
								filter.NewDockerCheckDefaults(
									"quay", // type
									"otherTokenGHCR",
									"otherTokenHub", "otherUsernameHub",
									"otherTokenQuay",
									nil))),
					),
					DeployedVersionLookup: *deployedver.NewDefaults(
						boolPtr(false)), // allow_invalid_certs
					Dashboard: service.NewDashboardOptionsDefaults(
						boolPtr(true))}, // auto_approve
				Notify: shoutrrr.SliceDefaults{
					"discord": shoutrrr.NewDefaults(
						"",
						&map[string]string{ // options
							"message": "foo {{ version }}"},
						&map[string]string{ // params
							"username": "Argus"},
						&map[string]string{ // url_params
							"host": "example.com"})},
				WebHook: *webhook.NewDefaults(
					boolPtr(true), // allow_invalid_certs
					&webhook.Headers{ // custom_headers
						{Key: "X-Header", Value: "foo"}},
					"0s",                    // delay
					intPtr(203),             // desired_status_code
					uintPtr(2),              // max_tries
					"secret!!!",             // secret
					boolPtr(false),          // silent_fails
					"github",                // type
					"https://example.comm"), // url
			},
			want: `
service:
  options:
    interval: 1m
    semantic_versioning: false
  latest_version:
    access_token: foo
    allow_invalid_certs: true
    use_prerelease: false
    require:
      docker:
        type: ghcr
        ghcr:
          token: tokenGHCR
        hub:
          token: tokenHub
          username: usernameHub
        quay:
          token: tokenQuay
  deployed_version:
    allow_invalid_certs: false
  dashboard:
    auto_approve: true
notify:
  discord:
    options:
      message: foo {{ version }}
    url_fields:
      host: example.com
    params:
      username: Argus
webhook:
  type: github
  url: https://example.comm
  allow_invalid_certs: true
  custom_headers:
    - key: X-Header
      value: foo
  secret: secret!!!
  desired_status_code: 203
  delay: 0s
  max_tries: 2
  silent_fails: false`,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			prefixes := []string{"", " ", "  ", "    ", "- "}
			for _, prefix := range prefixes {
				want := strings.TrimPrefix(tc.want, "\n")
				if want != "" {
					if want != "{}" {
						want = prefix + strings.ReplaceAll(want, "\n", "\n"+prefix)
					}
					want += "\n"
				}

				// WHEN the Defaults are stringified with String()
				got := tc.defaults.String(prefix)

				// THEN the result is as expected
				if got != want {
					t.Errorf("(prefix=%q) want: %q\ngot:\n%q",
						prefix, want, got)
				}
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
			got:  defaults.Notify["discord"].GetParam("username"),
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
		"empty vars ignored": {
			env: map[string]string{
				"ARGUS_SERVICE_OPTIONS_INTERVAL":            "99m",
				"ARGUS_SERVICE_OPTIONS_SEMANTIC_VERSIONING": ""},
			want: &Defaults{
				Service: service.Defaults{
					Options: *opt.NewDefaults("99m", nil)}},
		},
		"service.options": {
			env: map[string]string{
				"ARGUS_SERVICE_OPTIONS_INTERVAL":            "99m",
				"ARGUS_SERVICE_OPTIONS_SEMANTIC_VERSIONING": "true"},
			want: &Defaults{
				Service: service.Defaults{
					Options: *opt.NewDefaults("99m", boolPtr(true))}},
		},
		"service.options - invalid time.duration - interval": {
			env: map[string]string{
				"ARGUS_SERVICE_OPTIONS_INTERVAL":            "99 something",
				"ARGUS_SERVICE_OPTIONS_SEMANTIC_VERSIONING": "true"},
			errRegex: `interval: "[^"]+" <invalid>`,
		},
		"service.options - invalid bool - semantic version": {
			env: map[string]string{
				"ARGUS_SERVICE_OPTIONS_INTERVAL":            "99",
				"ARGUS_SERVICE_OPTIONS_SEMANTIC_VERSIONING": "foo"},
			errRegex: `invalid bool for `,
		},
		"service.latest_version": {
			env: map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_ACCESS_TOKEN":        "ghp_something",
				"ARGUS_SERVICE_LATEST_VERSION_ALLOW_INVALID_CERTS": "true",
				"ARGUS_SERVICE_LATEST_VERSION_USE_PRERELEASE":      "true"},
			want: &Defaults{
				Service: service.Defaults{
					LatestVersion: *latestver.NewDefaults(
						stringPtr("ghp_something"),
						boolPtr(true),
						boolPtr(true),
						nil)}},
		},
		"service.latest_version.require": {
			env: map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_TYPE":         "ghcr",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_GHCR_TOKEN":   "tokenForGHCR",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_HUB_TOKEN":    "tokenForDockerHub",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_HUB_USERNAME": "usernameForDockerHub",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_QUAY_TOKEN":   "tokenForQuay"},
			want: &Defaults{
				Service: service.Defaults{
					LatestVersion: *latestver.NewDefaults(
						nil, nil, nil,
						filter.NewRequireDefaults(
							filter.NewDockerCheckDefaults(
								"ghcr",
								"tokenForGHCR",
								"tokenForDockerHub",
								"usernameForDockerHub",
								"tokenForQuay", nil)))}},
		},
		"service.latest_version.require - invalid type": {
			env: map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_TYPE":         "foo",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_GHCR_TOKEN":   "tokenForGHCR",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_HUB_TOKEN":    "tokenForDockerHub",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_HUB_USERNAME": "usernameForDockerHub",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_QUAY_TOKEN":   "tokenForQuay"},
			errRegex: `service:[^ ]+  latest_version:[^ ]+    require:[^ ]+      docker:[^ ]+        type: "foo" <invalid> `,
		},
		"service.latest_version - invalid bool - allow_invalid_certs": {
			env: map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_ACCESS_TOKEN":        "ghp_something",
				"ARGUS_SERVICE_LATEST_VERSION_ALLOW_INVALID_CERTS": "bar",
				"ARGUS_SERVICE_LATEST_VERSION_USE_PRERELEASE":      "true"},
			errRegex: `invalid bool for [^:]+`,
		},
		"service.latest_version - invalid bool - use_prerelease": {
			env: map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_ACCESS_TOKEN":        "ghp_something",
				"ARGUS_SERVICE_LATEST_VERSION_ALLOW_INVALID_CERTS": "true",
				"ARGUS_SERVICE_LATEST_VERSION_USE_PRERELEASE":      "bop"},
			errRegex: `invalid bool for [^:]+`,
		},
		"service.deployed_version": {
			env: map[string]string{
				"ARGUS_SERVICE_DEPLOYED_VERSION_ALLOW_INVALID_CERTS": "true"},
			want: &Defaults{
				Service: service.Defaults{
					DeployedVersionLookup: *deployedver.NewDefaults(
						boolPtr(true))}},
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
				Service: service.Defaults{
					Dashboard: service.NewDashboardOptionsDefaults(
						boolPtr(true))}},
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
				Notify: shoutrrr.SliceDefaults{
					"discord": shoutrrr.NewDefaults(
						"",
						&map[string]string{
							"message":   "bish",
							"max_tries": "1",
							"delay":     "1h"},
						&map[string]string{
							"avatar":     ":argus:",
							"color":      "0x50D9ff",
							"colordebug": "0x7b00ab",
							"colorerror": "0xd60510",
							"colorinfo":  "0x2488ff",
							"colorwarn":  "0xffc441",
							"json":       "no",
							"splitlines": "yes",
							"title":      "something",
							"username":   "test"},
						&map[string]string{
							"token":     "foo",
							"webhookid": "bar"})}},
		},
		"notify.discord - invalid options.delay": {
			env: map[string]string{
				"ARGUS_NOTIFY_DISCORD_OPTIONS_DELAY": "foo"},
			want: &Defaults{
				Notify: shoutrrr.SliceDefaults{
					"discord": shoutrrr.NewDefaults(
						"",
						&map[string]string{
							"delay": "foo"},
						nil, nil)}},
			errRegex: `notify:[^ ]+  discord:[^ ]+    options:[^ ]+      delay: "foo" <invalid>`,
		},
		"notify.smtp": {
			env: map[string]string{
				"ARGUS_NOTIFY_SMTP_OPTIONS_MESSAGE":     "bing",
				"ARGUS_NOTIFY_SMTP_OPTIONS_MAX_TRIES":   "2",
				"ARGUS_NOTIFY_SMTP_OPTIONS_DELAY":       "2m",
				"ARGUS_NOTIFY_SMTP_URL_FIELDS_USERNAME": "user",
				"ARGUS_NOTIFY_SMTP_URL_FIELDS_PASSWORD": "secret",
				"ARGUS_NOTIFY_SMTP_URL_FIELDS_HOST":     "smtp.example.com",
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
				Notify: shoutrrr.SliceDefaults{
					"smtp": shoutrrr.NewDefaults(
						"",
						&map[string]string{
							"message":   "bing",
							"max_tries": "2",
							"delay":     "2m"},
						&map[string]string{
							"fromaddress": "me@example.com",
							"toaddresses": "you@somewhere.com",
							"auth":        "Unknown",
							"clienthost":  "localhost",
							"encryption":  "auto",
							"fromname":    "someone",
							"subject":     "Argus SMTP Notification",
							"usehtml":     "no",
							"usestarttls": "yes"},
						&map[string]string{
							"username": "user",
							"password": "secret",
							"host":     "smtp.example.com",
							"port":     "25"})}},
		},
		"notify.gotify": {
			env: map[string]string{
				"ARGUS_NOTIFY_GOTIFY_OPTIONS_MESSAGE":   "shazam",
				"ARGUS_NOTIFY_GOTIFY_OPTIONS_MAX_TRIES": "3",
				"ARGUS_NOTIFY_GOTIFY_OPTIONS_DELAY":     "3s",
				"ARGUS_NOTIFY_GOTIFY_URL_FIELDS_HOST":   "gotify.example.com",
				"ARGUS_NOTIFY_GOTIFY_URL_FIELDS_PORT":   "443",
				"ARGUS_NOTIFY_GOTIFY_URL_FIELDS_PATH":   "gotify",
				"ARGUS_NOTIFY_GOTIFY_URL_FIELDS_TOKEN":  "SuperSecretToken",
				"ARGUS_NOTIFY_GOTIFY_PARAMS_DISABLETLS": "no",
				"ARGUS_NOTIFY_GOTIFY_PARAMS_PRIORITY":   "0",
				"ARGUS_NOTIFY_GOTIFY_PARAMS_TITLE":      "Argus Gotify Notification"},
			want: &Defaults{
				Notify: shoutrrr.SliceDefaults{
					"gotify": shoutrrr.NewDefaults(
						"",
						&map[string]string{
							"message":   "shazam",
							"max_tries": "3",
							"delay":     "3s"},
						&map[string]string{
							"disabletls": "no",
							"priority":   "0",
							"title":      "Argus Gotify Notification"},
						&map[string]string{
							"host":  "gotify.example.com",
							"port":  "443",
							"path":  "gotify",
							"token": "SuperSecretToken"})}},
		},
		"notify.googlechat": {
			env: map[string]string{
				"ARGUS_NOTIFY_GOOGLECHAT_OPTIONS_MESSAGE":   "whoosh",
				"ARGUS_NOTIFY_GOOGLECHAT_OPTIONS_MAX_TRIES": "4",
				"ARGUS_NOTIFY_GOOGLECHAT_OPTIONS_DELAY":     "4h",
				"ARGUS_NOTIFY_GOOGLECHAT_URL_FIELDS_RAW":    "chat.googleapis.com/v1/spaces/FOO/messages?key=bar&token=baz"},
			want: &Defaults{
				Notify: shoutrrr.SliceDefaults{
					"googlechat": shoutrrr.NewDefaults(
						"",
						&map[string]string{
							"message":   "whoosh",
							"max_tries": "4",
							"delay":     "4h"},
						nil,
						&map[string]string{
							"raw": "chat.googleapis.com/v1/spaces/FOO/messages?key=bar&token=baz"})}},
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
				Notify: shoutrrr.SliceDefaults{
					"ifttt": shoutrrr.NewDefaults(
						"",
						&map[string]string{
							"message":   "pow",
							"max_tries": "5",
							"delay":     "5m"},
						&map[string]string{
							"events":            "event1,event2",
							"title":             "Argus IFTTT Notification",
							"usemessageasvalue": "2",
							"usetitleasvalue":   "0",
							"value1":            "bish",
							"value2":            "bash",
							"value3":            "bosh"},
						&map[string]string{
							"webhookid": "secretWHID"})}},
		},
		"notify.join": {
			env: map[string]string{
				"ARGUS_NOTIFY_JOIN_OPTIONS_MESSAGE":   "pew",
				"ARGUS_NOTIFY_JOIN_OPTIONS_MAX_TRIES": "6",
				"ARGUS_NOTIFY_JOIN_OPTIONS_DELAY":     "6s",
				"ARGUS_NOTIFY_JOIN_URL_FIELDS_APIKEY": "apiKey",
				"ARGUS_NOTIFY_JOIN_PARAMS_DEVICES":    "device1,device2",
				"ARGUS_NOTIFY_JOIN_PARAMS_ICON":       "example.com/icon.png",
				"ARGUS_NOTIFY_JOIN_PARAMS_TITLE":      "Argus Join Notification"},
			want: &Defaults{
				Notify: shoutrrr.SliceDefaults{
					"join": shoutrrr.NewDefaults(
						"",
						&map[string]string{
							"message":   "pew",
							"max_tries": "6",
							"delay":     "6s"},
						&map[string]string{
							"devices": "device1,device2",
							"icon":    "example.com/icon.png",
							"title":   "Argus Join Notification"},
						&map[string]string{
							"apikey": "apiKey"})}},
		},
		"notify.mattermost": {
			env: map[string]string{
				"ARGUS_NOTIFY_MATTERMOST_OPTIONS_MESSAGE":     "ping",
				"ARGUS_NOTIFY_MATTERMOST_OPTIONS_MAX_TRIES":   "7",
				"ARGUS_NOTIFY_MATTERMOST_OPTIONS_DELAY":       "7h",
				"ARGUS_NOTIFY_MATTERMOST_URL_FIELDS_USERNAME": "Argus",
				"ARGUS_NOTIFY_MATTERMOST_URL_FIELDS_HOST":     "mattermost.example.com",
				"ARGUS_NOTIFY_MATTERMOST_URL_FIELDS_PORT":     "443",
				"ARGUS_NOTIFY_MATTERMOST_URL_FIELDS_PATH":     "mattermost",
				"ARGUS_NOTIFY_MATTERMOST_URL_FIELDS_TOKEN":    "mattermostToken",
				"ARGUS_NOTIFY_MATTERMOST_URL_FIELDS_CHANNEL":  "argus",
				"ARGUS_NOTIFY_MATTERMOST_PARAMS_ICON":         ":argus:",
				"ARGUS_NOTIFY_MATTERMOST_PARAMS_TITLE":        "Argus Mattermost Notification"},
			want: &Defaults{
				Notify: shoutrrr.SliceDefaults{
					"mattermost": shoutrrr.NewDefaults(
						"",
						&map[string]string{
							"message":   "ping",
							"max_tries": "7",
							"delay":     "7h"},
						&map[string]string{
							"icon":  ":argus:",
							"title": "Argus Mattermost Notification"},
						&map[string]string{
							"username": "Argus",
							"host":     "mattermost.example.com",
							"port":     "443",
							"path":     "mattermost",
							"token":    "mattermostToken",
							"channel":  "argus"})}},
		},
		"notify.matrix": {
			env: map[string]string{
				"ARGUS_NOTIFY_MATRIX_OPTIONS_MESSAGE":     "pong",
				"ARGUS_NOTIFY_MATRIX_OPTIONS_MAX_TRIES":   "8",
				"ARGUS_NOTIFY_MATRIX_OPTIONS_DELAY":       "8m",
				"ARGUS_NOTIFY_MATRIX_URL_FIELDS_USER":     "argus",
				"ARGUS_NOTIFY_MATRIX_URL_FIELDS_HOST":     "matrix.example.com",
				"ARGUS_NOTIFY_MATRIX_URL_FIELDS_PORT":     "443",
				"ARGUS_NOTIFY_MATRIX_URL_FIELDS_PATH":     "matrix",
				"ARGUS_NOTIFY_MATRIX_URL_FIELDS_PASSWORD": "matrixPassword",
				"ARGUS_NOTIFY_MATRIX_PARAMS_DISABLETLS":   "no",
				"ARGUS_NOTIFY_MATRIX_PARAMS_ROOMS":        "room1,room2",
				"ARGUS_NOTIFY_MATRIX_PARAMS_TITLE":        "Argus Matrix Notification"},
			want: &Defaults{
				Notify: shoutrrr.SliceDefaults{
					"matrix": shoutrrr.NewDefaults(
						"",
						&map[string]string{
							"message":   "pong",
							"max_tries": "8",
							"delay":     "8m"},
						&map[string]string{
							"disabletls": "no",
							"rooms":      "room1,room2",
							"title":      "Argus Matrix Notification"},
						&map[string]string{
							"user":     "argus",
							"host":     "matrix.example.com",
							"port":     "443",
							"path":     "matrix",
							"password": "matrixPassword"})}},
		},
		"notify.opsgenie": {
			env: map[string]string{
				"ARGUS_NOTIFY_OPSGENIE_OPTIONS_MESSAGE":    "pang",
				"ARGUS_NOTIFY_OPSGENIE_OPTIONS_MAX_TRIES":  "9",
				"ARGUS_NOTIFY_OPSGENIE_OPTIONS_DELAY":      "9s",
				"ARGUS_NOTIFY_OPSGENIE_URL_FIELDS_HOST":    "opsgenie.example.com",
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
				Notify: shoutrrr.SliceDefaults{
					"opsgenie": shoutrrr.NewDefaults(
						"",
						&map[string]string{
							"message":   "pang",
							"max_tries": "9",
							"delay":     "9s"},
						&map[string]string{
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
							"visibleto":   "visible1,visible2"},
						&map[string]string{
							"host":   "opsgenie.example.com",
							"port":   "443",
							"path":   "opsgenie",
							"apikey": "opsGenieApiKey"})}},
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
				Notify: shoutrrr.SliceDefaults{
					"pushbullet": shoutrrr.NewDefaults(
						"",
						&map[string]string{
							"message":   "pung",
							"max_tries": "10",
							"delay":     "10h"},
						&map[string]string{
							"title": "Argus Pushbullet Notification"},
						&map[string]string{
							"token":   "pushbulletToken",
							"targets": "target1,target2"})}},
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
				Notify: shoutrrr.SliceDefaults{
					"pushover": shoutrrr.NewDefaults(
						"",
						&map[string]string{
							"message":   "pung",
							"max_tries": "11",
							"delay":     "11m"},
						&map[string]string{
							"devices":  "device1,device2",
							"priority": "0",
							"title":    "Argus Pushbullet Notification"},
						&map[string]string{
							"token": "pushoverToken",
							"user":  "pushoverUser"})}},
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
				Notify: shoutrrr.SliceDefaults{
					"rocketchat": shoutrrr.NewDefaults(
						"",
						&map[string]string{
							"message":   "pung",
							"max_tries": "12",
							"delay":     "12s"},
						nil,
						&map[string]string{
							"username": "rocketchatUser",
							"host":     "rocketchat.example.com",
							"port":     "443",
							"path":     "rocketchat",
							"tokena":   "FIRST_token",
							"tokenb":   "SECOND_token",
							"channel":  "rocketchatChannel"})}},
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
				Notify: shoutrrr.SliceDefaults{
					"slack": shoutrrr.NewDefaults(
						"",
						&map[string]string{
							"message":   "slung",
							"max_tries": "13",
							"delay":     "13h"},
						&map[string]string{
							"botname":  "Argus",
							"color":    "#ff8000",
							"icon":     ":ghost:",
							"threadts": "1234567890.123456",
							"title":    "Argus Slack Notification"},
						&map[string]string{
							"token":   "slackToken",
							"channel": "somewhere"})}},
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
				Notify: shoutrrr.SliceDefaults{
					"teams": shoutrrr.NewDefaults(
						"",
						&map[string]string{
							"message":   "tung",
							"max_tries": "14",
							"delay":     "14m"},
						&map[string]string{
							"color": "#ff8000",
							"host":  "teams.example.com",
							"title": "Argus Teams Notification"},
						&map[string]string{
							"group":      "teamsGroup",
							"tenant":     "tenant",
							"altid":      "otherID?",
							"groupowner": "owner"})}},
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
				Notify: shoutrrr.SliceDefaults{
					"telegram": shoutrrr.NewDefaults(
						"",
						&map[string]string{
							"message":   "tong",
							"max_tries": "15",
							"delay":     "15s"},
						&map[string]string{
							"chats":        "chat1,chat2",
							"notification": "yes",
							"parsemode":    "None",
							"preview":      "yes",
							"title":        "Argus Telegram Notification"},
						&map[string]string{
							"token": "telegramToken"})}},
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
				Notify: shoutrrr.SliceDefaults{
					"zulip": shoutrrr.NewDefaults(
						"",
						&map[string]string{
							"message":   "zung",
							"max_tries": "16",
							"delay":     "16h"},
						&map[string]string{
							"stream": "stream",
							"topic":  "topic"},
						&map[string]string{
							"botmail": "botmail",
							"botkey":  "botkey",
							"host":    "zulip.example.com",
							"port":    "1234",
							"path":    "zulip"})}},
		},
		"webhook": {
			env: map[string]string{
				"ARGUS_WEBHOOK_ALLOW_INVALID_CERTS": "false",
				"ARGUS_WEBHOOK_DELAY":               "99s",
				"ARGUS_WEBHOOK_DESIRED_STATUS_CODE": "201",
				"ARGUS_WEBHOOK_MAX_TRIES":           "88",
				"ARGUS_WEBHOOK_SILENT_FAILS":        "true",
				"ARGUS_WEBHOOK_TYPE":                "github",
				"ARGUS_WEBHOOK_URL":                 "webhook.example.com"},
			want: &Defaults{
				WebHook: *webhook.NewDefaults(
					boolPtr(false),
					nil,
					"99s",
					intPtr(201),
					uintPtr(88),
					"",
					boolPtr(true),
					"github",
					"webhook.example.com")},
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
				Service: service.Defaults{
					DeployedVersionLookup: deployedver.LookupDefaults{}}}
			if tc.want == nil {
				tc.want = &Defaults{
					Notify: shoutrrr.SliceDefaults{}}
			}
			if tc.want.Notify != nil {
				dflts.Notify = shoutrrr.SliceDefaults{}
				for notifyType := range defaults.Notify {
					dflts.Notify[notifyType] = &shoutrrr.ShoutrrrDefaults{}
					dflts.Notify[notifyType].InitMaps()
					if tc.want.Notify[notifyType] == nil {
						tc.want.Notify[notifyType] = &shoutrrr.ShoutrrrDefaults{}
						tc.want.Notify[notifyType].InitMaps()
					}
				}
			}
			for k, v := range tc.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}
			// Catch fatal panics.
			defer func() {
				r := recover()
				if r != nil {
					if tc.errRegex == "" {
						t.Fatalf("unexpected panic: %v", r)
					}
					switch r.(type) {
					case string:
						if !regexp.MustCompile(tc.errRegex).MatchString(r.(string)) {
							t.Errorf("want error matching:\n%v\ngot:\n%v",
								tc.errRegex, t)
						}
					default:
						t.Fatalf("unexpected panic: %v", r)
					}
				}
			}()

			// WHEN MapEnvToStruct is called on it
			dflts.MapEnvToStruct()

			// THEN any error is as expected
			if tc.errRegex != "" { // Expected a FATAL panic to be caught above
				t.Fatalf("expected an error matching %q, but got none", tc.errRegex)
			}
			// AND the defaults are set to the appropriate env vars
			if dflts.String("") != tc.want.String("") {
				t.Errorf("want:\n%v\ngot:\n%v",
					tc.want.String(""), dflts.String(""))
			}
		})
	}
}

func TestDefaults_CheckValues(t *testing.T) {
	// GIVEN defaults with a test of invalid vars
	var defaults Defaults
	defaults.SetDefaults()
	tests := map[string]struct {
		input    *Defaults
		errRegex []string
	}{
		"Service.Interval": {
			input: &Defaults{Service: service.Defaults{
				Options: *opt.NewDefaults("10x", nil)}},
			errRegex: []string{
				`^service:$`,
				`^  options:$`,
				`^    interval: "10x" <invalid>`},
		},
		"Service.LatestVersion.Require.Docker.Type": {
			input: &Defaults{Service: service.Defaults{
				LatestVersion: latestver.LookupDefaults{
					Require: filter.RequireDefaults{
						Docker: *filter.NewDockerCheckDefaults(
							"pizza",
							"", "", "", "", nil)}}}},
			errRegex: []string{
				`^service:$`,
				`^  latest_version:$`,
				`^    require:$`,
				`^      docker:$`,
				`^        type: "pizza" <invalid>`},
		},
		"Service.Interval + Service.DeployedVersionLookup.Regex": {
			input: &Defaults{Service: service.Defaults{
				Options: *opt.NewDefaults("10x", nil),
				LatestVersion: latestver.LookupDefaults{
					Require: filter.RequireDefaults{
						Docker: *filter.NewDockerCheckDefaults(
							"pizza",
							"", "", "", "", nil)}}}},
			errRegex: []string{
				`^service:$`,
				`^  options:$`,
				`^    interval: "10x" <invalid>`,
				`^  latest_version:$`,
				`^    require:$`,
				`^      docker:$`,
				`^        type: "pizza" <invalid>`},
		},
		"Notify.x.Delay": {
			input: &Defaults{Notify: shoutrrr.SliceDefaults{
				"slack": shoutrrr.NewDefaults(
					"",
					&map[string]string{"delay": "10x"},
					nil, nil)}},
			errRegex: []string{
				`^notify:$`,
				`^  slack:$`,
				`^    options:`,
				`^      delay: "10x" <invalid>`},
		},
		"WebHook.Delay": {
			input: &Defaults{WebHook: *webhook.NewDefaults(
				nil, nil, "10x", nil, nil, "", nil, "", "")},
			errRegex: []string{
				`^webhook:$`,
				`^  delay: "10x" <invalid>`},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			prefixes := []string{"", " ", "  ", "    ", "- "}
			for _, prefix := range prefixes {
				errRegex := make([]string, len(tc.errRegex))
				for i := range errRegex {
					errRegex[i] = strings.ReplaceAll(tc.errRegex[i], `^`, `^`+prefix)
				}

				// WHEN CheckValues is called on it
				err := util.ErrorToString(tc.input.CheckValues(prefix))

				// THEN err matches expected
				lines := strings.Split(err, "\\")
				for i := range errRegex {
					re := regexp.MustCompile(errRegex[i])
					match := re.MatchString(lines[i])
					if !match {
						t.Errorf("(prefix=%q) want match for: %q on %q\n%q",
							prefix, errRegex[i], lines[i], err)
						return
					}
				}
			}
		})
	}
}

func TestDefaults_Print(t *testing.T) {
	// GIVEN a set of Defaults
	var defaults Defaults
	defaults.SetDefaults()
	tests := map[string]struct {
		input *Defaults
		lines int
	}{
		"unmodified hard defaults": {
			input: &defaults,
			lines: 148 + len(defaults.Notify)},
		"empty defaults": {
			input: &Defaults{},
			lines: 1},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN Print is called
			tc.input.Print("")

			// THEN the expected number of lines are printed
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
