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
	"strings"

	api_types "github.com/release-argus/Argus/web/api/types"
)

// AnnounceCommand will announce the failed status of `c.Announce` channel
// (Broadcast to all WebSocket clients).
func (c *Controller) AnnounceCommand(index int) {
	var payloadData []byte

	commandSummary := make(map[string]*api_types.CommandSummary)
	commandSummary[strings.Join((*c.Command)[index], " ")] = &api_types.CommandSummary{Failed: c.Failed[index]}

	// Command success/fail
	wsPage := "APPROVALS"
	wsType := "COMMAND"
	wsSubType := "EVENT"
	payloadData, _ = json.Marshal(api_types.WebSocketMessage{
		Page:    &wsPage,
		Type:    &wsType,
		SubType: &wsSubType,
		ServiceData: &api_types.ServiceSummary{
			ID: c.ServiceID,
		},
		CommandData: commandSummary,
	})

	if c.Announce != nil {
		*c.Announce <- payloadData
	}
}

// Find `command`.
func (c *Controller) Find(command string) *int {
	// Loop through all the Command(s)
	for i := range *c.Command {
		// If this command starts with the same text
		if (*c.Command)[i].String() == command {
			return &i
		}
	}
	return nil
}

// ResetFails of this Controller's Commands
func (c *Controller) ResetFails() {
	if c == nil {
		return
	}
	for i := range c.Failed {
		c.Failed[i] = nil
	}
}
