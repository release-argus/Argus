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

package config

import (
	"strings"

	"github.com/release-argus/Argus/conversions"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/utils"
)

// convertCurrentVersionToDeployedVersion handles the deprecation of `current_version*`,
// Renaming it to `deployed_version*`
func (c *Config) convertCurrentVersionToDeployedVersion() {
	for service_id := range c.Service {
		if c.Service[service_id].Status != nil {
			if c.Service[service_id].Status.CurrentVersion != "" {
				if c.Service[service_id].Status.DeployedVersion == "" {
					c.Service[service_id].Status.DeployedVersion = c.Service[service_id].Status.CurrentVersion
				}
				c.Service[service_id].Status.CurrentVersion = ""
			}
			if c.Service[service_id].Status.CurrentVersionTimestamp != "" {
				if c.Service[service_id].Status.DeployedVersionTimestamp == "" {
					c.Service[service_id].Status.DeployedVersionTimestamp = c.Service[service_id].Status.CurrentVersionTimestamp
				}
				c.Service[service_id].Status.CurrentVersionTimestamp = ""
			}
		}
	}
}

// convertDeprecatedSlackAndGotify will handle converting the old 'Gotify' and 'Slack' slices
// to the new 'Notify' format(*c.Service[serviceIndex]).Slack)
func (c *Config) convertDeprecatedSlackAndGotify() {
	// Check whether Defaults.Notify.(Slack|Mattermost) are wanted
	if c.Notify == nil {
		c.Notify = make(shoutrrr.Slice)
	}
	if c.Defaults.Notify == nil {
		c.Defaults.Notify = make(shoutrrr.Slice)
	}
	if c.Slack == nil {
		c.Slack = &conversions.SlackSlice{}
	}
	hadDefaultsSlack := true
	if c.Defaults.Slack == nil {
		c.Defaults.Slack = &conversions.Slack{}
		hadDefaultsSlack = false
	}
	hadDefaultsGotify := true
	if c.Gotify == nil {
		c.Gotify = &conversions.GotifySlice{}
		hadDefaultsGotify = false
	}
	if c.Defaults.Gotify == nil {
		c.Defaults.Gotify = &conversions.Gotify{}
	}

	// Gotify
	// Convert inside services
	for serviceIndex := range c.Service {
		if (*c.Service[serviceIndex]).Gotify == nil {
			continue
		}
		// Ensure the NotifySlice is not nil
		if (*c.Service[serviceIndex]).Notify == nil {
			notifySlice := make(shoutrrr.Slice)
			(*c.Service[serviceIndex]).Notify = &notifySlice
		}
		// Loop through the Gotifies
		for oldName := range *(*c.Service[serviceIndex]).Gotify {
			newName := oldName + "_gotify"
			// Convert this Gotify
			converted := (*(*c.Service[serviceIndex]).Gotify)[oldName].Convert("")
			// Give it to the Notify
			(*(*c.Service[serviceIndex]).Notify)[newName] = &converted
		}
	}
	// Convert mains
	if c.Gotify != nil {
		// Loop through the Gotifies
		for oldName := range *c.Gotify {
			newName := oldName + "_gotify"
			// Convert this Gotify
			converted := (*c.Gotify)[oldName].Convert("gotify")
			// Give it to the Notify
			c.Notify[newName] = &converted
		}
	}
	// Convert defaults
	if c.Defaults.Gotify != nil && hadDefaultsGotify {
		converted := (*c.Defaults.Gotify).Convert("gotify")
		c.Defaults.Notify["gotify"] = &converted
	}

	// Slack
	// Convert inside services
	for serviceIndex := range c.Service {
		if (*c.Service[serviceIndex]).Slack == nil {
			continue
		}
		// Ensure the NotifySlice is not nil
		if (*c.Service[serviceIndex]).Notify == nil {
			notifySlice := make(shoutrrr.Slice)
			(*c.Service[serviceIndex]).Notify = &notifySlice
		}
		for oldName := range *(*c.Service[serviceIndex]).Slack {
			newName := oldName
			// If oldName isn't unique, make it so
			// Keep looping just incase of XXX_slack_slack... names
			main := (*c.Slack)[oldName]
			mainURL := ""
			if main != nil {
				mainURL = utils.DefaultIfNil(main.URL)
			}
			dflt := c.Defaults.Slack
			dfltURL := ""
			if dflt != nil {
				dfltURL = utils.DefaultIfNil(dflt.URL)
			}
			url := utils.GetFirstNonDefault(
				utils.DefaultIfNil((*(*c.Service[serviceIndex]).Slack)[oldName].URL),
				mainURL,
				dfltURL)

			if isSlack := strings.Contains(url, "hooks.slack.com"); isSlack {
				newName = "_slack"
			} else {
				newName += "_mattermost"
			}

			// Convert this Slack/Mattermost
			converted := (*(*c.Service[serviceIndex]).Slack)[oldName].Convert(newName, url)
			// Give it to the notify
			(*(*c.Service[serviceIndex]).Notify)[newName] = &converted
		}
	}
	// Convert mains
	if c.Slack != nil {
		for oldName := range *c.Slack {
			newName := oldName
			url := utils.DefaultIfNil((*c.Slack)[oldName].URL)
			// Try and find a service with an oldName slack
			if url == "" {
				for serviceIndex := range c.Service {
					if (*c.Service[serviceIndex]).Slack != nil && (*(*c.Service[serviceIndex]).Slack)[oldName] != nil {
						url = utils.DefaultIfNil((*(*c.Service[serviceIndex]).Slack)[oldName].URL)
						if url != "" {
							break
						}
					}
				}
			}
			if isSlack := strings.Contains(url, "hooks.slack.com"); isSlack {
				newName += "_slack"
			} else {
				newName += "_mattermost"
			}
			converted := (*c.Slack)[oldName].Convert(newName, url)
			c.Notify[newName] = &converted
		}
	}
	// Convert defaults
	if c.Defaults.Slack != nil && hadDefaultsSlack {
		// Set the type to whatever is more common out of conversions, Slack or Mattermost
		converted := (*c.Defaults.Slack).Convert("", utils.DefaultIfNil((*c.Defaults.Slack).URL))
		c.Defaults.Notify["slack"] = &converted
		c.Defaults.Notify["mattermost"] = &converted
	}
	c.Defaults.Gotify = nil
	c.Gotify = nil
	c.Defaults.Slack = nil
	c.Slack = nil
	for i := range c.Service {
		c.Service[i].Gotify = nil
		c.Service[i].Slack = nil
	}
}

// convertDeprecatedURLCommands will convert the 'regex_submatch' URLCommand to 'regex'
func (c *Config) convertDeprecatedURLCommands() {
	for service_id := range c.Service {
		if c.Service[service_id].URLCommands != nil {
			for command_id := range *c.Service[service_id].URLCommands {
				if (*c.Service[service_id].URLCommands)[command_id].Type == "regex_submatch" {
					(*c.Service[service_id].URLCommands)[command_id].Type = "regex"
					// Use 0 now as that references the bracket group
					(*c.Service[service_id].URLCommands)[command_id].Index = 0
				}
			}
		}
	}
}
