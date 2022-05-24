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
	hasSlack := false
	hasMattermost := false
	if c.Defaults.Notify["slack"] != nil {
		hasSlack = true
	}
	if c.Defaults.Notify["mattermost"] != nil {
		hasMattermost = true
	}
	if c.Notify == nil {
		c.Notify = make(shoutrrr.Slice)
	}
	if c.Defaults.Notify == nil {
		c.Defaults.Notify = make(shoutrrr.Slice)
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
			newName := oldName
			// If oldName isn't unique, make it so
			// Keep looping just incase of XXX_gotify_gotify... names
			for {
				// add suffix until newName is unique
				if (*(*c.Service[serviceIndex]).Gotify)[newName] != nil || (*(*c.Service[serviceIndex]).Slack)[newName] != nil || (*(*c.Service[serviceIndex]).Notify)[newName] != nil {
					newName += "_gotify"
				} else {
					break
				}
			}

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
			newName := oldName
			// If oldName isn't unique, make it so
			// Keep looping just incase of XXX_gotify_gotify... names
			for {
				// add suffix until newName is unique
				if (c.Gotify != nil && (*c.Gotify)[newName] != nil) || (c.Slack != nil && (*c.Slack)[newName] != nil) || c.Notify[newName] != nil {
					newName += "_gotify"
				} else {
					break
				}
			}
			converted := (*c.Gotify)[oldName].Convert("gotify")
			c.Notify[newName] = &converted
		}
	}
	// Convert defaults
	if c.Defaults.Gotify != nil {
		converted := (*c.Defaults.Gotify).Convert("gotify")
		c.Defaults.Notify["gotify"] = &converted
	}

	// Slack
	if c.Slack != nil {
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
				isSlack := strings.Contains(url, "hooks.slack.com")
				// add suffix to names that aren't unique between 'gotify' and 'slack'
				suffix := "_mattermost"
				if isSlack {
					suffix = "_slack"
				}
				for {
					// add suffix until newName is unique
					if (*(*c.Service[serviceIndex]).Slack)[newName] != nil || (*(*c.Service[serviceIndex]).Notify)[newName] != nil {
						newName += suffix
					} else {
						break
					}
				}

				// Convert this Slack/Mattermost
				converted := (*(*c.Service[serviceIndex]).Slack)[oldName].Convert(newName, url)
				// Give it to the notify
				(*(*c.Service[serviceIndex]).Notify)[newName] = &converted
			}
		}
		// Convert mains
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
			isSlack := strings.Contains(url, "hooks.slack.com")
			suffix := "_mattermost"
			if isSlack {
				suffix = "_slack"
				hasSlack = true
			} else {
				hasMattermost = true
			}
			// If oldName isn't unique, make it so
			// Keep looping just incase of XXX_slack_slack... names
			for {
				// add suffix until newName is unique
				if (c.Slack != nil && (*c.Slack)[newName] != nil) || c.Notify[newName] != nil {
					newName += suffix
				} else {
					break
				}
			}
			converted := (*c.Slack)[oldName].Convert(newName, url)
			c.Notify[newName] = &converted
		}
		// Convert defaults
		if c.Defaults.Slack != nil {
			// Set the type to whatever is more common out of conversions, Slack or Mattermost
			converted := (*c.Defaults.Slack).Convert("", utils.DefaultIfNil((*c.Defaults.Slack).URL))
			c.Defaults.Notify["slack"] = &converted
			c.Defaults.Notify["mattermost"] = &converted
		}
	}
	c.Defaults.Gotify = nil
	c.Gotify = nil
	c.Defaults.Slack = nil
	c.Slack = nil
	for i := range c.Service {
		c.Service[i].Gotify = nil
		c.Service[i].Slack = nil
	}

	// Delete defaults if they're most likely not wanted
	if !hasSlack && hasMattermost {
		delete(c.Defaults.Notify, "slack")
	}
	if !hasMattermost && hasSlack {
		delete(c.Defaults.Notify, "mattermost")
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
