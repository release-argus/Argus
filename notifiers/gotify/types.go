// Copyright [2022] [Hymenaios]
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

package gotify

import (
	"github.com/hymenaios-io/Hymenaios/utils"
)

var jLog *utils.JLog

// Slice mapping of Gotify.
type Slice map[string]*Gotify

// Gotify is a Gotification w/ destination and sender details.
type Gotify struct {
	ID           *string `yaml:"-"`                   // ID for this Gotify sender
	URL          *string `yaml:"url,omitempty"`       // "https://example.com
	Token        *string `yaml:"token,omitempty"`     // apptoken
	Title        *string `yaml:"title,omitempty"`     // "{{ service_id }} - {{ version }} released"
	Message      *string `yaml:"message,omitempty"`   // "Hymenaios"
	Extras       *Extras `yaml:"extras,omitempty"`    // Message extras
	Priority     *int    `yaml:"priority,omitempty"`  // <1 = Min, 1-3 = Low, 4-7 = Med, >7 = High
	Delay        *string `yaml:"delay,omitempty"`     // The delay before sending the Gotify message.
	MaxTries     *uint   `yaml:"max_tries,omitempty"` // Number of times to attempt sending the Gotify message if a 200 is not received.
	Failed       *bool   `yaml:"-"`                   // Whether the last send attempt failed
	HardDefaults *Gotify `yaml:"-"`                   // Harcoded default values
	Defaults     *Gotify `yaml:"-"`                   // Default values
	Main         *Gotify `yaml:"-"`                   // The Gotify that this Gotify is calling (and may override parts of)
}

// Payload to be to be sent as the Gotification.
type Payload struct {
	Extras   map[string]interface{} `json:"extras,omitempty"`
	Message  string                 `form:"message" query:"message" json:"message" binding:"required,omitempty"`
	Priority int                    `json:"priority,omitempty"`
	Title    string                 `form:"title" query:"title" json:"title,omitempty"`
}

// Extras (https://gotify.net/docs/msgextras) for the Gotifications.
type Extras struct {
	AndroidAction      *string `yaml:"android_action,omitempty"`      // URL to open on notification delivery
	ClientDisplay      *string `yaml:"client_display,omitempty"`      // Render message in 'text/plain' or 'text/markdown'
	ClientNotification *string `yaml:"client_notification,omitempty"` // URL to open on notification click
}
