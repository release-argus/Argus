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

package conversions

import (
	"fmt"
	"strings"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/utils"
)

// Slice mapping of Slack.
type SlackSlice map[string]*Slack

// Slack is a message w/ destination and from details.
type Slack struct {
	ID        *string `yaml:"-"`                    // Unique across the Slice
	URL       *string `yaml:"url,omitempty"`        // "https://example.com
	IconEmoji *string `yaml:"icon_emoji,omitempty"` // ":github:"
	IconURL   *string `yaml:"icon_url,omitempty"`   // "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png"
	Username  *string `yaml:"username,omitempty"`   // "Argus"
	Message   *string `yaml:"message,omitempty"`    // "<{{ service_url }}|{{ service_id }}> - {{ version }} released"
	Delay     *string `yaml:"delay,omitempty"`      // The delay before sending the Slack message.
	MaxTries  *uint   `yaml:"max_tries,omitempty"`  // Number of times to attempt sending the Slack message if a 200 is not received.
	Failed    *bool   `yaml:"-"`                    // Whether the last attempt to send failed
	Type      *string `yaml:"-"`                    // slack/mattermost
}

// TODO: Remove V
func (s *Slack) Convert(
	id string,
	url string,
) (converted shoutrrr.Shoutrrr) {
	if s == nil {
		return
	}
	converted.InitMaps()
	convertedID := id
	converted.ID = &convertedID

	isSlack := strings.Contains(url, "hooks.slack.com")
	var convertedType string
	if isSlack {
		convertedType = "slack"
		url = strings.TrimPrefix(url, "https://")
		url = strings.TrimPrefix(url, "http://")

		token := "hook:" + strings.Join(strings.Split(url, "/")[1:], "-")
		converted.SetURLField("token", token)

		converted.SetURLField("channel", "webhook")
		// mattermost
	} else {
		if s.URL != nil {
			url := *s.URL

			// Host
			url = strings.TrimPrefix(url, "https://")
			url = strings.TrimPrefix(url, "http://")
			parts := strings.Split(url, "/")
			converted.SetURLField("host", strings.Split(parts[0], ":")[0])
			converted.SetURLField("port", utils.GetPortFromURL(*s.URL, "443"))

			convertedPath := ""
			// mattermost.example.io/hooks/TOKEN -> [ "hooks", "TOKEN"]
			// mattermost.example.io/sub/path/hooks/TOKEN -> [ "sub", "path", "hooks", "TOKEN" ]
			if len(parts) > 2 {
				convertedPath = strings.Join(parts[1:len(parts)-2], "/")
			}
			if convertedPath != "" {
				converted.SetURLField("path", convertedPath)
			}

			convertedToken := parts[len(parts)-1]
			converted.SetURLField("token", convertedToken)
		}
		convertedType = "mattermost"
	}
	converted.Type = convertedType

	if s.Message != nil {
		converted.SetOption("message", *s.Message)
	}

	if s.Delay != nil {
		converted.SetOption("delay", *s.Delay)
	}

	if s.MaxTries != nil {
		converted.SetOption("max_tries", fmt.Sprint(*s.MaxTries))
	}

	if s.Username != nil {
		converted.SetParam("botname", *s.Username)
	}

	if s.IconURL != nil {
		converted.SetParam("icon", *s.IconURL)
	} else if s.IconEmoji != nil {
		converted.SetParam("icon", *s.IconEmoji)
	}
	return
}
