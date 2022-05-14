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

package slack

import "github.com/release-argus/Argus/utils"

var (
	jLog *utils.JLog
)

// Slice mapping of Slack.
type Slice map[string]*Slack

// Slack is a message w/ destination and from details.
type Slack struct {
	ID           *string `yaml:"-"`                    // Unique across the Slice
	URL          *string `yaml:"url,omitempty"`        // "https://example.com
	ServiceIcon  *string `yaml:"-"`                    // Service.Icon
	IconEmoji    *string `yaml:"icon_emoji,omitempty"` // ":github:"
	IconURL      *string `yaml:"icon_url,omitempty"`   // "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png"
	Username     *string `yaml:"username,omitempty"`   // "Argus"
	Message      *string `yaml:"message,omitempty"`    // "<{{ service_url }}|{{ service_id }}> - {{ version }} released"
	Delay        *string `yaml:"delay,omitempty"`      // The delay before sending the Slack message.
	MaxTries     *uint   `yaml:"max_tries,omitempty"`  // Number of times to attempt sending the Slack message if a 200 is not received.
	Failed       *bool   `yaml:"-"`                    // Whether the last attempt to send failed
	HardDefaults *Slack  `yaml:"-"`                    // Hardcoded default values
	Defaults     *Slack  `yaml:"-"`                    // Default values
	Main         *Slack  `yaml:"-"`                    // The Slack that this Slack is calling (and may override parts of)
}

// Payload to be to be sent as the Slack message.
type Payload struct {
	Username  string  `json:"username"`   // "Argus"
	IconEmoji *string `json:"icon_emoji"` // ":github:"
	IconURL   *string `json:"icon_url"`   // "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png"
	Text      string  `json:"text"`       // "${service} - {{ version }} released"
}
