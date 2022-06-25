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

// Slice mapping of Gotify.
type GotifySlice map[string]*Gotify

// Gotify is a Gotification w/ destination and sender details.
type Gotify struct {
	URL      *string `yaml:"url,omitempty"`       // "https://example.com
	Token    *string `yaml:"token,omitempty"`     // apptoken
	Title    *string `yaml:"title,omitempty"`     // "{{ service_id }} - {{ version }} released"
	Message  *string `yaml:"message,omitempty"`   // "Argus"
	Priority *int    `yaml:"priority,omitempty"`  // <1 = Min, 1-3 = Low, 4-7 = Med, >7 = High
	Delay    *string `yaml:"delay,omitempty"`     // The delay before sending the Gotify message.
	MaxTries *uint   `yaml:"max_tries,omitempty"` // Number of times to attempt sending the Gotify message if a 200 is not received.
}

// TODO: Remove V
func (g *Gotify) Convert(id string) (converted shoutrrr.Shoutrrr) {
	if g == nil {
		return
	}
	converted.InitMaps()
	convertedID := id
	converted.ID = &convertedID
	converted.Type = "gotify"

	if g.URL != nil {
		url := *g.URL
		url = strings.TrimPrefix(url, "https://")
		if strings.HasPrefix(url, "http://") {
			converted.SetParam("disabletls", "yes")
		}
		url = strings.TrimPrefix(url, "http://")

		parts := strings.Split(url, "/")
		convertedHost := parts[0]
		converted.SetURLField("host", convertedHost)
		if strings.Contains(parts[0], ":") {
			hostSplit := strings.Split(parts[0], ":")
			converted.SetURLField("host", hostSplit[0])
		}
		converted.SetURLField("port", utils.GetPortFromURL(*g.URL, "80"))

		// gotify.example.io -> [ "gotify.example.io" ]
		// gotify.example.io/test/123 -> [ "gotify.example.io", "test", "123" ]
		convertedPath := ""
		if len(parts) > 1 {
			convertedPath = strings.Join(parts[1:], "/")
			converted.SetURLField("path", convertedPath)
		}

	}

	if g.Token != nil {
		converted.SetURLField("token", *g.Token)
	}

	if g.Title != nil {
		converted.SetParam("title", *g.Title)
	}

	if g.Message != nil {
		converted.SetOption("message", *g.Message)
	}

	if g.Priority != nil {
		converted.SetParam("priority", fmt.Sprintf("%d", *g.Priority))
	}

	if g.Delay != nil {
		converted.SetOption("delay", *g.Delay)
	}

	if g.MaxTries != nil {
		converted.SetOption("max_tries", fmt.Sprintf("%d", *g.MaxTries))
	}
	return
}
