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
	"strings"
	"time"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

var (
	jLog *util.JLog
)

// Slice mapping of WebHook.
type Slice []Command

type Command []string

// String returns a string representation of the Command in the format of 'arg0 arg1'.
func (c *Command) String() (str string) {
	if c == nil {
		return
	}
	str = strings.Join(*c, " ")
	return
}

type Controller struct {
	Command        *Slice            `yaml:"-" json:"-"` // command(s) to run (with args)
	NextRunnable   []time.Time       `yaml:"-" json:"-"` // Time the Commands can next be run (for staggering)
	Failed         *[]*bool          `yaml:"-" json:"-"` // Whether the last execution attempt failed
	Notifiers      Notifiers         `yaml:"-" json:"-"` // The Notify's to notify on failures
	ServiceStatus  *svcstatus.Status `yaml:"-" json:"-"` // Status of the Service (used for templating commands)
	ParentInterval *string           `yaml:"-" json:"-"` // Interval between the parent Service's queries
}

// Notifiers to use when their WebHook fails.
type Notifiers struct {
	Shoutrrr *shoutrrr.Slice // Shoutrrr
}

// CopyFailsFrom target
func (c *Controller) CopyFailsFrom(target *Controller) {
	if c == nil || c.Command == nil ||
		target == nil || target.Failed == nil {
		return
	}

	if c.Failed == nil {
		fails := make([]*bool, len(*c.Command))
		c.Failed = &fails
		c.NextRunnable = make([]time.Time, len(*c.Command))
	}

	// Loop through old fails
	for i := range *target.Failed {
		// Loop through new fails to find try and find this command
		for j := range *c.Failed {
			// If the command has been run (has a failed state)
			// and the commands match, copy the failed status
			if (*target.Failed)[i] != nil &&
				(*target.Command)[i].FormattedString() == (*c.Command)[j].FormattedString() {
				failed := *(*target.Failed)[i]
				(*c.Failed)[j] = &failed
				c.NextRunnable[j] = target.NextRunnable[i]
				break
			}
		}
	}

}
