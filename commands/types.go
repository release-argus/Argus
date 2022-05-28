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
	"github.com/release-argus/Argus/utils"
)

var (
	jLog *utils.JLog
)

// Slice mapping of WebHook.
type Slice []Command

type Command []string

type Fails []*bool

type Controller struct {
	ServiceID *string      `yaml:"-"` // ID of the service this Controller is attached to
	Command   *Slice       `yaml:"-"` // command to run (with args)
	Failed    Fails        `yaml:"-"` // Whether the last execution attempt failed
	Notifiers Notifiers    `yaml:"-"` // The Notify's to notify on failures
	Announce  *chan []byte `yaml:"-"` // Announce to the WebSocket
}

// Notifiers to use when their WebHook fails.
type Notifiers struct {
	Shoutrrr *shoutrrr.Slice // Shoutrrr
}
