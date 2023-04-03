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

package command

import (
	"encoding/json"

	api_type "github.com/release-argus/Argus/web/api/types"
)

// AnnounceCommand will announce the Command fail status to `c.Announce` channel
// (Broadcast to all WebSocket clients).
func (c *Controller) AnnounceCommand(index int) {
	c.SetNextRunnable(index, false)
	commandSummary := make(map[string]*api_type.CommandSummary, 1)
	formatted := (*c.Command)[index].ApplyTemplate(c.ServiceStatus)
	commandSummary[formatted.String()] = &api_type.CommandSummary{
		Failed:       c.Failed.Get(index),
		NextRunnable: c.NextRunnable[index],
	}

	// Command success/fail
	var payloadData []byte
	payloadData, _ = json.Marshal(api_type.WebSocketMessage{
		Page:    "APPROVALS",
		Type:    "COMMAND",
		SubType: "EVENT",
		ServiceData: &api_type.ServiceSummary{
			ID: *c.ServiceStatus.ServiceID,
		},
		CommandData: commandSummary,
	})

	if c.ServiceStatus.AnnounceChannel != nil {
		*c.ServiceStatus.AnnounceChannel <- payloadData
	}
}

// Find `command`.
func (c *Controller) Find(command string) (index *int) {
	if c == nil {
		return
	}

	// Loop through all the Command(s)
	for key := range *c.Command {
		formatted := (*c.Command)[key].ApplyTemplate(c.ServiceStatus)
		// If this key is the command
		if formatted.String() == command {
			index = &key
			return
		}
	}
	return
}

// ResetFails of this Controller's Commands
func (c *Controller) ResetFails() {
	if c == nil {
		return
	}
	c.Failed.Reset()
}
