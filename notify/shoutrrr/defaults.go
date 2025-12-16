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

// Package shoutrrr provides the shoutrrr notification service to services.
package shoutrrr

import "net/http"

// notifyDefaultOptions are the default options for all notifiers.
func notifyDefaultOptions() map[string]string {
	return map[string]string{
		"message":   "{{ service_name | default:service_id }} - {{ version }} released",
		"max_tries": "3",
		"delay":     "0s"}
}

// Default sets these ShoutrrrsDefaults to the default values.
func (s *ShoutrrrsDefaults) Default() {
	defaults := make(ShoutrrrsDefaults, len(supportedTypes))
	defaults["bark"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		map[string]string{
			"port": "443"},
		map[string]string{
			"title": "Argus"})
	defaults["discord"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil,
		map[string]string{
			"splitlines": "yes",
			"username":   "Argus"})
	defaults["smtp"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil,
		map[string]string{
			"requirestarttls": "no",
			"skiptlsverify":   "no",
			"usehtml":         "no",
			"usestarttls":     "yes"})
	defaults["googlechat"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil, nil)
	defaults["gotify"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		map[string]string{
			"port": "443"},
		map[string]string{
			"disabletls":         "no",
			"insecureskipverify": "no",
			"priority":           "0",
			"title":              "Argus",
			"useheader":          "no"})
	defaults["ifttt"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil,
		map[string]string{
			"usemessageasvalue": "2",
			"usetitleasvalue":   "0"})
	defaults["join"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil, nil)
	defaults["mattermost"] = NewDefaults(
		"",
		map[string]string{
			"message": "<{{ service_url }}|{{ service_name | default:service_id }}>" +
				" - {{ version }} released" +
				"{% if web_url %} (<{{ web_url }}|changelog>){% endif %}",
			"max_tries": "3",
			"delay":     "0s"},
		map[string]string{
			"username": "Argus",
			"port":     "443"},
		map[string]string{
			"disabletls": "no",
		})
	defaults["matrix"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		map[string]string{
			"port": "443"},
		map[string]string{
			"disabletls": "no",
		})
	defaults["ntfy"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		map[string]string{
			"host": "ntfy.sh"},
		map[string]string{
			"disabletls": "no",
			"title":      "Argus"})
	defaults["opsgenie"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil, nil)
	defaults["pushbullet"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil,
		map[string]string{
			"title": "Argus"})
	defaults["pushover"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil, nil)
	defaults["rocketchat"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		map[string]string{
			"port": "443"},
		nil)
	defaults["slack"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil,
		map[string]string{
			"botname": "Argus"})
	defaults["teams"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil, nil)
	defaults["telegram"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil,
		map[string]string{
			"notification": "yes",
			"preview":      "yes"})
	defaults["zulip"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil, nil)
	defaults["generic"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil,
		map[string]string{
			"contenttype":   "application/json",
			"disabletls":    "no",
			"messagekey":    "message",
			"requestmethod": http.MethodPost,
			"titlekey":      "title"})

	// Initialise maps.
	for _, notify := range defaults {
		notify.InitMaps()
	}

	// Overwrite the receiver.
	*s = defaults
}
