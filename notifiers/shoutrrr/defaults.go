// Copyright [2023] [Argus]
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

package shoutrrr

// notifyDefaultOptions are the default options for all notifiers.
func notifyDefaultOptions() *map[string]string {
	return &map[string]string{
		"message":   "{{ service_id }} - {{ version }} released",
		"max_tries": "3",
		"delay":     "0s"}
}

// SetDefaults for Shoutrrr.
func (s *SliceDefaults) SetDefaults() {
	newSlice := make(SliceDefaults, len(supportedTypes))
	newSlice["bark"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		&map[string]string{
			"title": "Argus"},
		&map[string]string{
			"port": "443"})
	newSlice["discord"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		&map[string]string{
			"splitlines": "yes",
			"username":   "Argus"},
		nil)
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
		&map[string]string{
			"disabletls": "no",
			"priority":   "0",
			"title":      "Argus"},
		&map[string]string{
			"port": "443"})
	newSlice["ifttt"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		&map[string]string{
			"usemessageasvalue": "2",
			"usetitleasvalue":   "0"},
		nil)
	newSlice["join"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil, nil)
	newSlice["mattermost"] = NewDefaults(
		"",
		&map[string]string{
			"message":   "<{{ service_url }}|{{ service_id }}> - {{ version }} released{% if web_url %} (<{{ web_url }}|changelog>){% endif %}",
			"max_tries": "3",
			"delay":     "0s"},
		nil,
		&map[string]string{
			"username": "Argus",
			"port":     "443"})
	newSlice["matrix"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil,
		&map[string]string{
			"disabletls": "no",
			"port":       "443"})
	newSlice["ntfy"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		&map[string]string{
			"title": "Argus"},
		&map[string]string{
			"host": "ntfy.sh"})
	newSlice["opsgenie"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil, nil)
	newSlice["pushbullet"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		&map[string]string{
			"title": "Argus"},
		nil)
	newSlice["pushover"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil, nil)
	newSlice["rocketchat"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil,
		&map[string]string{
			"port": "443"})
	newSlice["slack"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		&map[string]string{
			"botname": "Argus"},
		nil)
	newSlice["teams"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil, nil)
	newSlice["telegram"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		&map[string]string{
			"notification": "yes",
			"preview":      "yes"},
		nil)
	newSlice["zulip"] = NewDefaults(
		"",
		notifyDefaultOptions(),
		nil, nil)
	// InitMaps
	for _, notify := range newSlice {
		notify.InitMaps()
	}

	// Set the new slice as the default slice
	*s = newSlice
}
