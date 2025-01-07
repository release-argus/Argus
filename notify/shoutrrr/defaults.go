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

// notifyDefaultOptions are the default options for all notifiers.
func notifyDefaultOptions() map[string]string {
	return map[string]string{
		"message":   "{{ service_name | default:service_id }} - {{ version }} released",
		"max_tries": "3",
		"delay":     "0s"}
}

// Default sets this SliceDefaults to the default values.
func (s *SliceDefaults) Default() {
	newSlice := make(SliceDefaults, len(supportedTypes))
	newSlice["bark"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		map[string]string{
			"port": "443"},
		map[string]string{
			"title": "Argus"})
	newSlice["discord"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil,
		map[string]string{
			"splitlines": "yes",
			"username":   "Argus"})
	newSlice["smtp"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil, nil)
	newSlice["googlechat"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil, nil)
	newSlice["gotify"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		map[string]string{
			"port": "443"},
		map[string]string{
			"disabletls": "no",
			"priority":   "0",
			"title":      "Argus"})
	newSlice["ifttt"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil,
		map[string]string{
			"usemessageasvalue": "2",
			"usetitleasvalue":   "0"})
	newSlice["join"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil, nil)
	newSlice["mattermost"] = NewDefaults(
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
		nil)
	newSlice["matrix"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		map[string]string{
			"disabletls": "no",
			"port":       "443"},
		nil)
	newSlice["ntfy"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		map[string]string{
			"host": "ntfy.sh"},
		map[string]string{
			"title": "Argus"})
	newSlice["opsgenie"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil, nil)
	newSlice["pushbullet"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil,
		map[string]string{
			"title": "Argus"})
	newSlice["pushover"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil, nil)
	newSlice["rocketchat"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		map[string]string{
			"port": "443"},
		nil)
	newSlice["slack"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil,
		map[string]string{
			"botname": "Argus"})
	newSlice["teams"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil, nil)
	newSlice["telegram"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil,
		map[string]string{
			"notification": "yes",
			"preview":      "yes"})
	newSlice["zulip"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil, nil)
	newSlice["generic"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil,
		map[string]string{
			"contenttype":   "application/json",
			"disabletls":    "no",
			"messagekey":    "message",
			"requestmethod": "POST",
			"titlekey":      "title"})
	newSlice["shoutrrr"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil, nil)

	// Initialise maps.
	for _, notify := range newSlice {
		notify.InitMaps()
	}

	// Set the new slice as the default slice.
	*s = newSlice
}
