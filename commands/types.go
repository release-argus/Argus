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
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
	"time"
)

var (
	jLog *utils.JLog
)

// Slice mapping of WebHook.
type Slice []Command

type Command []string

type Fails []*bool

type Controller struct {
	ServiceID      *string                `yaml:"-"` // ID of the service this Controller is attached to
	Command        *Slice                 `yaml:"-"` // command to run (with args)
	NextRunnable   []time.Time            `yaml:"-"` // Time the Commands can next be run (for staggering)
	Failed         Fails                  `yaml:"-"` // Whether the last execution attempt failed
	Notifiers      Notifiers              `yaml:"-"` // The Notify's to notify on failures
	Announce       *chan []byte           `yaml:"-"` // Announce to the WebSocket
	ServiceStatus  *service_status.Status `yaml:"-"` // Status of the Service (used for templating commands)
	ParentInterval *string                `yaml:"-"` // Interval between the parent Service's queries
}

// Notifiers to use when their WebHook fails.
type Notifiers struct {
	Shoutrrr *shoutrrr.Slice // Shoutrrr
}
