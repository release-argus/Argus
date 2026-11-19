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

// Package types provides Shoutrrr type-changes for the Web API.
package types

import (
	"slices"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/util"
)

// SafeFields are the keys of a notifier type whose values the API shows.
// It censors every other value.
type SafeFields struct {
	URLFields    []string // Shown as configured.
	URLFieldMaps []string // JSON objects, shown with each value censored.
	Params       []string // Shown as configured.
	ParamMaps    []string // JSON objects, shown with each value censored.
}

// SafeFieldsByType are the [SafeFields] of each notifier type.
var SafeFieldsByType = map[string]SafeFields{
	"bark": {
		URLFields: []string{"host", "path", "port"},
		Params: []string{
			"badge", "category", "copy", "disabletls", "group", "icon", "scheme", "skiptlsverify",
			"sound", "title",
		},
	},
	"discord": {
		Params: []string{
			"avatar", "avatarurl", "color", "colordebug", "colorerror", "colorinfo", "colorwarn",
			"json", "mentions", "splitlines", "thread_id", "thread_name", "title", "username",
		},
	},
	"generic": {
		URLFields:    []string{"host", "port"},
		URLFieldMaps: []string{"headers", "json_payload_vars", "query_vars"},
		Params: []string{
			"contenttype", "disabletls", "messagekey", "method", "skiptlsverify", "template",
			"title", "titlekey",
		},
	},
	"googlechat": {},
	"gotify": {
		URLFields: []string{"host", "path", "port"},
		Params: []string{
			"date", "disabletls", "extras", "insecureskipverify", "priority", "skiptlsverify",
			"title", "useheader",
		},
	},
	"homeassistant": {
		URLFields: []string{"host", "path", "port"},
		Params:    []string{"disabletls", "nid", "service", "skiptlsverify", "targets", "title"},
	},
	"ifttt": {
		Params: []string{
			"events", "messagevalue", "title", "titlevalue", "value1", "value2", "value3",
		},
	},
	"join": {
		Params: []string{"icon", "title"},
	},
	"lark": {
		URLFields: []string{"host"},
		Params:    []string{"link", "title"},
	},
	"matrix": {
		URLFields: []string{"host", "port", "user"},
		Params:    []string{"disabletls", "room", "rooms", "skiptlsverify", "title"},
	},
	"mattermost": {
		URLFields: []string{"channel", "host", "path", "port", "username"},
		Params: []string{
			"channel", "disabletls", "icon", "icon_emoji", "icon_url", "skiptlsverify", "title",
			"username",
		},
	},
	"mqtt": {
		URLFields: []string{"host", "port", "topic", "username"},
		Params: []string{
			"cleansession", "clientid", "disabletls", "disabletlsverification", "qos", "retained",
			"skiptlsverify",
		},
	},
	"notifiarr": {
		Params: []string{
			"channel", "color", "content", "fields", "footer", "icon", "id", "image", "name",
			"thumbnail", "update",
		},
	},
	"ntfy": {
		URLFields: []string{"host", "port", "topic", "username"},
		Params: []string{
			"actions", "at", "cache", "delay", "disabletls", "disabletls_",
			"disabletlsverification", "email", "filename", "firebase", "icon", "in", "markdown",
			"priority", "scheme", "skiptlsverify", "tags", "title",
		},
	},
	"opsgenie": {
		URLFields: []string{"host", "port"},
		Params: []string{
			"actions", "alias", "description", "entity", "note", "priority", "responders", "source",
			"tags", "title", "user", "visibleto",
		},
		ParamMaps: []string{"details"},
	},
	"pagerduty": {
		URLFields: []string{"host", "port"},
		Params: []string{
			"action", "client", "client_url", "contexts", "dedup_key", "severity", "source",
			"title",
		},
	},
	"pushbullet": {
		URLFields: []string{"targets"},
		Params:    []string{"title"},
	},
	"pushover": {
		Params: []string{"expire", "priority", "retry", "timestamp", "title"},
	},
	"rocketchat": {
		URLFields: []string{"channel", "host", "path", "port", "username"},
		Params:    []string{"channel", "disabletls", "skiptlsverify", "username"},
	},
	"shoutrrr": {},
	"signal": {
		URLFields: []string{"host", "port", "recipients", "source", "user"},
		Params: []string{
			"attachments", "disabletls", "host", "notifyself", "port", "recipients",
			"skiptlsverify", "source", "text_mode", "textmode", "title", "to", "user",
		},
	},
	"signalgrid": {
		Params: []string{"critical", "title", "type"},
	},
	"slack": {
		URLFields: []string{"channel"},
		Params: []string{
			"botname", "color", "icon", "icon_emoji", "icon_url", "thread_ts", "title", "username",
		},
	},
	"smtp": {
		URLFields: []string{"host", "port", "username"},
		Params: []string{
			"auth", "clienthost", "clientid", "encryption", "from", "fromaddress", "fromname",
			"requirestarttls", "scopes", "skiptlsverify", "starttls", "subject", "timeout", "title",
			"to", "toaddresses", "tokenurl", "usehtml", "usestarttls",
		},
	},
	"teams": {
		URLFields: []string{"group", "groupowner", "tenant"},
		Params:    []string{"color", "title"},
	},
	"telegram": {
		Params: []string{
			"apiurl", "channels", "chats", "notification", "parsemode", "preview", "skiptlsverify",
			"title",
		},
	},
	"twilio": {
		URLFields: []string{"accountsid", "fromnumber", "tonumbers"},
		Params:    []string{"title"},
	},
	"wecom": {
		Params: []string{"mentioned_list", "mentioned_mobile_list"},
	},
	"xmpp": {
		URLFields: []string{"host", "port", "user"},
		Params:    []string{"disabletls", "nick", "rooms", "skiptlsverify", "title", "to"},
	},
	"zulip": {
		URLFields: []string{"botmail", "host", "port"},
		Params: []string{
			"read_by_sender", "skiptlsverify", "stream", "title", "to", "topic", "type",
		},
	},
}

// SafeFieldsOf returns the [SafeFields] of sType, which are empty when sType is unknown.
func SafeFieldsOf(sType string) SafeFields {
	return SafeFieldsByType[sType]
}

// CensorURLFields returns a copy of urlFields with each value the receiver does not show censored.
func (f SafeFields) CensorURLFields(urlFields map[string]string) map[string]string {
	return censor(urlFields, f.URLFields, f.URLFieldMaps)
}

// CensorParams returns a copy of params with each value the receiver does not show censored.
func (f SafeFields) CensorParams(params map[string]string) map[string]string {
	return censor(params, f.Params, f.ParamMaps)
}

// censor keeps empty and safe values, censors JSON objects under maps by value and the rest whole.
func censor(values map[string]string, safe, maps []string) map[string]string {
	out := make(map[string]string, len(values))
	for key, value := range values {
		switch {
		case value == "" || slices.Contains(safe, key):
			out[key] = value
		case slices.Contains(maps, key):
			out[key] = CensorMapValues(value)
		default:
			out[key] = util.SecretValue
		}
	}
	return out
}

// CensorMapValues replaces each value in the JSON object with [util.SecretValue], keeping the keys.
// A value that is not a JSON object string is replaced whole.
func CensorMapValues(value string) string {
	if value == "" {
		return ""
	}

	var m map[string]string
	if err := decode.Unmarshal("json", []byte(value), &m); err != nil {
		return util.SecretValue
	}

	for k := range m {
		m[k] = util.SecretValue
	}
	// Re-encoding decoded JSON cannot fail.
	out, _ := decode.Marshal("json", m)

	return string(out)
}

// RestoreMapValues restores the [util.SecretValue] values in the JSON object from original.
// A whole-value [util.SecretValue] is replaced with original.
func RestoreMapValues(original, value string) string {
	if original == "" {
		return value
	}
	if value == util.SecretValue {
		return original
	}

	var target, orig map[string]string
	if err := decode.Unmarshal("json", []byte(value), &target); err != nil {
		return value
	}
	if err := decode.Unmarshal("json", []byte(original), &orig); err != nil {
		return value
	}

	changed := false
	for k, v := range target {
		if o := orig[k]; v == util.SecretValue && o != "" {
			target[k] = o
			changed = true
		}
	}
	if !changed {
		return value
	}
	out, err := decode.Marshal("json", target)
	if err != nil {
		return value
	}

	return string(out)
}

// RestoreSecrets restores target's censored values from original, including inside JSON objects.
func RestoreSecrets(original, target map[string]string) {
	for key, value := range target {
		target[key] = RestoreMapValues(original[key], value)
	}
}
