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

package webhook

import (
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/utils"
)

var (
	jLog *utils.JLog
)

// Slice mapping of WebHook.
type Slice map[string]*WebHook

// WebHook to send.
type WebHook struct {
	ID                *string      `yaml:"-"`                             // Unique across the Slice
	ServiceID         *string      `yaml:"-"`                             // ID of the service this WebHook is attached to
	Type              *string      `yaml:"type,omitempty"`                // "github"/"url"
	URL               *string      `yaml:"url,omitempty"`                 // "https://example.com"
	Secret            *string      `yaml:"secret,omitempty"`              // "SECRET"
	DesiredStatusCode *int         `yaml:"desired_status_code,omitempty"` // e.g. 202
	Delay             *string      `yaml:"delay,omitempty"`               // The delay before sending the WebHook.
	MaxTries          *uint        `yaml:"max_tries,omitempty"`           // Number of times to attempt sending the WebHook if the desired status code is not received.
	SilentFails       *bool        `yaml:"silent_fails,omitempty"`        // Whether to notify if this WebHook fails MaxTries times.
	Failed            *bool        `yaml:"-"`                             // Whether the last send attempt failed
	HardDefaults      *WebHook     `yaml:"-"`                             // Hardcoded default values
	Defaults          *WebHook     `yaml:"-"`                             // Default values
	Main              *WebHook     `yaml:"-"`                             // The Webhook that this Webhook is calling (and may override parts of)
	Notifiers         *Notifiers   `yaml:"-"`                             // The Notify's to notify on failures
	Announce          *chan []byte `yaml:"-"`                             // Announce to the WebSocket
}

// Notifiers to use when their WebHook fails.
type Notifiers struct {
	Shoutrrr *shoutrrr.Slice // Shoutrrr
}
