// Copyright [2024]] [Argus]
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

// Package command provides the Command struct and methods for services.
package command

import (
	"encoding/json"
	"errors"
	"fmt"

	apitype "github.com/release-argus/Argus/web/api/types"
)

// AnnounceCommand will announce the Command fail status to `c.Announce` channel
// (Broadcast to all WebSocket clients).
func (c *Controller) AnnounceCommand(index int) {
	c.SetExecuting(index, false)
	commandSummary := make(map[string]*apitype.CommandSummary, 1)
	formatted := (*c.Command)[index].ApplyTemplate(c.ServiceStatus)
	commandSummary[formatted.String()] = &apitype.CommandSummary{
		Failed:       c.Failed.Get(index),
		NextRunnable: c.NextRunnable(index),
	}

	// Command success/fail.
	var payloadData []byte
	payloadData, _ = json.Marshal(apitype.WebSocketMessage{
		Page:    "APPROVALS",
		Type:    "COMMAND",
		SubType: "EVENT",
		ServiceData: &apitype.ServiceSummary{
			ID: *c.ServiceStatus.ServiceID},
		CommandData: commandSummary})

	c.ServiceStatus.SendAnnounce(&payloadData)
}

// Find `command` and return the index of it.
func (c *Controller) Find(command string) (int, error) {
	if c == nil {
		return 0, errors.New("controller is nil")
	}

	// Loop through all the Command(s).
	for key, cmd := range *c.Command {
		formatted := cmd.ApplyTemplate(c.ServiceStatus)
		// If this key is the command.
		if formatted.String() == command {
			return key, nil
		}
	}
	return 0, fmt.Errorf("command %q not found", command)
}
