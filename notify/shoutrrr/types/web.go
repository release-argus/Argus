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
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"slices"
	"strconv"
	"strings"

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
	// JSON values, shown with each string and number censored bar those at the listed paths.
	// A path joins object keys with "." and marks an array element with "[]", e.g. "[].label".
	ParamJSON map[string][]string
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
		URLFields:    []string{"host", "port", "username"},
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
			"date", "disabletls", "insecureskipverify", "priority", "skiptlsverify", "title",
			"useheader",
		},
		ParamJSON: map[string][]string{"extras": {"client::display.contentType"}},
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
			"at", "cache", "delay", "disabletls", "disabletls_", "disabletlsverification", "email",
			"filename", "firebase", "icon", "in", "markdown", "priority", "scheme", "skiptlsverify",
			"tags", "title",
		},
		ParamJSON: map[string][]string{
			"actions": {"[].action", "[].intent", "[].label", "[].method"},
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
	return censor(urlFields, f.URLFields, f.URLFieldMaps, nil)
}

// CensorParams returns a copy of params with each value the receiver does not show censored.
func (f SafeFields) CensorParams(params map[string]string) map[string]string {
	return censor(params, f.Params, f.ParamMaps, f.ParamJSON)
}

// censor keeps empty and safe values, censors JSON objects under maps by value,
// JSON values under jsonPaths by string and number, and the rest whole.
func censor(
	values map[string]string,
	safe, maps []string,
	jsonPaths map[string][]string,
) map[string]string {
	out := make(map[string]string, len(values))
	for key, value := range values {
		safePaths, isJSON := jsonPaths[key]
		switch {
		case value == "" || slices.Contains(safe, key):
			out[key] = value
		case slices.Contains(maps, key):
			out[key] = CensorMapValues(value)
		case isJSON:
			out[key] = CensorJSON(value, safePaths)
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

// CensorJSON replaces each string and number in the JSON value with [util.SecretValue],
// bar those at safePaths. A value that is not JSON is replaced whole.
func CensorJSON(value string, safePaths []string) string {
	if value == "" {
		return ""
	}

	node, err := unmarshalJSON(value)
	if err != nil {
		return util.SecretValue
	}
	// Re-encoding decoded JSON cannot fail.
	out, _ := decode.Marshal("json", censorNode(node, "", safePaths))

	return string(out)
}

// censorNode censors the strings and numbers under node, which sits at path, bar those at safePaths.
func censorNode(node any, path string, safePaths []string) any {
	switch n := node.(type) {
	case map[string]any:
		for key, value := range n {
			keyPath := key
			if path != "" {
				keyPath = path + "." + key
			}
			if !slices.Contains(safePaths, keyPath) {
				n[key] = censorNode(value, keyPath, safePaths)
			}
		}
	case []any:
		for i, value := range n {
			n[i] = censorNode(value, path+"[]", safePaths)
		}
	case string, jsontext.Value:
		return util.SecretValue
	}
	return node
}

// oldIndexKey is the key of a JSON array element that holds its index in the original array.
const oldIndexKey = "old_index"

// RestoreJSONValues restores the [util.SecretValue] values in value from original.
// Object values are matched by key, and array elements by their oldIndexKey, which is dropped.
// A whole-value [util.SecretValue] is replaced with original.
func RestoreJSONValues(original, value string) string {
	if value == util.SecretValue && original != "" {
		return original
	}
	nothingToRestore := original == "" || !strings.Contains(value, util.SecretValue)
	if nothingToRestore && !strings.Contains(value, oldIndexKey) {
		return value
	}

	node, err := unmarshalJSON(value)
	if err != nil {
		return value
	}
	orig, err := unmarshalJSON(original)
	if err != nil {
		orig = nil
	}
	// Re-encoding decoded JSON cannot fail.
	out, _ := decode.Marshal("json", restoreNode(orig, node))

	return string(out)
}

// restoreNode restores the [util.SecretValue] values under node from original.
func restoreNode(original, node any) any {
	switch n := node.(type) {
	case map[string]any:
		orig, _ := original.(map[string]any)
		for key, value := range n {
			n[key] = restoreNode(orig[key], value)
		}
	case []any:
		orig, _ := original.([]any)
		for i, value := range n {
			// Restore only from the element's own old index, so no other element's secret reaches it.
			var origElem any
			if elem, ok := value.(map[string]any); ok {
				raw, _ := elem[oldIndexKey].(jsontext.Value)
				if index, err := strconv.Atoi(string(raw)); err == nil && index >= 0 && index < len(orig) {
					origElem = orig[index]
				}
				delete(elem, oldIndexKey)
			}
			n[i] = restoreNode(origElem, value)
		}
	case string:
		if n == util.SecretValue && original != nil {
			return original
		}
	}
	return node
}

// unmarshalJSON decodes value, keeping each number as written rather than rounding it to a float64.
func unmarshalJSON(value string) (any, error) {
	var node any
	err := json.Unmarshal([]byte(value), &node, json.WithUnmarshalers(json.UnmarshalFromFunc(
		func(dec *jsontext.Decoder, v *any) error {
			if dec.PeekKind() != '0' {
				return errors.ErrUnsupported
			}
			raw, err := dec.ReadValue()
			if err != nil {
				return err //nolint:wrapcheck
			}
			*v = raw.Clone()
			return nil
		},
	)))
	return node, err //nolint:wrapcheck
}

// RestoreSecrets restores target's censored values from original, including inside JSON values.
func RestoreSecrets(original, target map[string]string) {
	for key, value := range target {
		target[key] = RestoreJSONValues(original[key], value)
	}
}
