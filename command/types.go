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
	"strings"
	"sync"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service/status"
)

// #########
// # TYPES #
// #########

// Commands is a slice of Command.
type Commands []Command

// Command is a slice of strings representing a cli command.
type Command []string

// Controller holds the commands to run and the status of the last run.
type Controller struct {
	mu           sync.RWMutex         // Mutex for concurrent access.
	Command      Commands             `json:"-" yaml:"-"` // Commands to run (with args).
	nextRunnable []time.Time          // Time the Commands can next be run (for staggering).
	Failed       *status.FailsCommand `json:"-" yaml:"-"` // Whether the last execution attempt failed.

	Notifiers      Notifiers      `json:"-" yaml:"-"` // The Notifiers to notify on failures.
	ServiceStatus  *status.Status `json:"-" yaml:"-"` // Status of the Service (used for templating commands).
	ParentInterval *string        `json:"-" yaml:"-"` // Interval between the parent Service's queries.
}

// Notifiers holds the notifiers used when a command fails.
type Notifiers struct {
	Shoutrrr shoutrrr.Shoutrrrs
}

// #########
// # STATE #
// #########

// Copy returns a deep copy of the receiver.
func (c *Commands) Copy() Commands {
	if c == nil {
		return nil
	}

	newCommands := make(Commands, len(*c))
	for i, cmd := range *c {
		newCommands[i] = cmd.Copy()
	}

	return newCommands
}

// Copy returns a deep copy of the receiver.
func (c *Command) Copy() Command {
	if c == nil {
		return nil
	}

	oldCommand := *c
	newCommand := make(Command, len(oldCommand))
	_ = copy(newCommand, oldCommand)
	return newCommand
}

// CopyFailsFrom copies failure and cooldown state from target into the receiver.
func (c *Controller) CopyFailsFrom(target *Controller) {
	if c == nil || c.Command == nil ||
		target == nil {
		return
	}

	target.mu.RLock()
	defer target.mu.RUnlock()
	c.mu.Lock()
	defer c.mu.Unlock()

	// Loop through old fails.
	for i := 0; i < target.Failed.Length(); i++ {
		// Loop through new commands and look for a match.
		for j := 0; j < c.Failed.Length(); j++ {
			// If the command has been run (has a failed state),
			// and the commands match, copy the failed status.
			if target.Failed.Get(i) != nil &&
				target.Command[i].JSON() == c.Command[j].JSON() {
				failed := target.Failed.Get(i)
				c.Failed.Set(j, *failed)
				c.nextRunnable[j] = target.nextRunnable[i]
				break
			}
		}
	}
}

// #############
// # STRINGIFY #
// #############

// String implements fmt.Stringer.
func (c *Command) String() string {
	if c == nil {
		return ""
	}
	return strings.Join(*c, " ")
}

// JSON returns the receiver encoded as a JSON array of strings.
func (c *Command) JSON() string {
	b, err := decode.Marshal("json", []string(*c))
	if err != nil {
		return "[]"
	}
	return string(b)
}
