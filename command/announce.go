// Copyright [2026] [Argus]
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

// Package command provides CLI command execution for services.
package command

import (
	"errors"
	"fmt"

	"github.com/release-argus/Argus/config/decode"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// AnnounceCommand broadcasts the command run result and next-runnable time to WebSocket clients.
func (c *Controller) AnnounceCommand(index int, serviceInfo serviceinfo.ServiceInfo) {
	c.SetExecuting(index, false)
	commandSummary := make(map[string]*apitype.CommandSummary, 1)
	formatted := c.Command[index].ApplyTemplate(serviceInfo)
	commandSummary[formatted.String()] = &apitype.CommandSummary{
		Failed:       c.Failed.Get(index),
		NextRunnable: c.NextRunnable(index),
	}

	// Command success/fail.
	var payloadData []byte
	payloadData, _ = decode.Marshal(
		"json", apitype.WebSocketMessage{
			Page:    "APPROVALS",
			Type:    "COMMAND",
			SubType: "EVENT",
			ServiceData: &apitype.ServiceSummary{
				ID: serviceInfo.ID,
			},
			CommandData: commandSummary,
		},
	)

	c.ServiceStatus.SendAnnounce(payloadData)
}

// Find returns the index of the [Command] matching the given string.
func (c *Controller) Find(command string) (int, error) {
	if c == nil {
		return 0, errors.New("controller is nil")
	}
	svcInfo := c.ServiceStatus.GetServiceInfo()

	// Loop through all the Commands.
	for key, cmd := range c.Command {
		formatted := cmd.ApplyTemplate(svcInfo)
		if formatted.String() == command {
			return key, nil
		}
	}
	return 0, fmt.Errorf("command %q not found", command)
}
