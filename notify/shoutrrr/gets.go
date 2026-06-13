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

// Package shoutrrr provides the shoutrrr notification service to services.
package shoutrrr

import (
	"strconv"
	"time"

	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/util"
)

// GetType of this Shoutrrr.
func (s *Shoutrrr) GetType() string {
	// s.ID if the name is the same as the type.
	return util.FirstNonDefault(
		s.Type,
		s.Main.Type,
		s.ID,
	)
}

// Title of the Shoutrrr after the service info is applied and template evaluated.
func (s *Shoutrrr) Title(info serviceinfo.ServiceInfo) string {
	return util.TemplateString(s.GetParam("title"), info)
}

// Message of the Shoutrrr after the service info is applied and template evaluated.
func (s *Shoutrrr) Message(info serviceinfo.ServiceInfo) string {
	return util.TemplateString(s.GetOption("message"), info)
}

// GetOption from this/Main/Defaults/HardDefaults on FiFo.
func (s *Shoutrrr) GetOption(key string) string {
	return util.FirstNonDefaultWithEnv(
		s.Options[key],
		s.Main.Options[key],
		s.Defaults.Options[key],
		s.HardDefaults.Options[key],
	)
}

// GetURLField from this/Main/Defaults/HardDefaults on FiFo.
func (s *Shoutrrr) GetURLField(key string) string {
	return util.FirstNonDefaultWithEnv(
		s.URLFields[key],
		s.Main.URLFields[key],
		s.Defaults.URLFields[key],
		s.HardDefaults.URLFields[key],
	)
}

// GetParam from this/Main/Defaults/HardDefaults on FiFo.
func (s *Shoutrrr) GetParam(key string) string {
	return util.FirstNonDefaultWithEnv(
		s.Params[key],
		s.Main.Params[key],
		s.Defaults.Params[key],
		s.HardDefaults.Params[key],
	)
}

// GetDelay returns the delay to wait before sending this notification.
func (s *Shoutrrr) GetDelay() string {
	delay := s.GetOption("delay")
	if delay == "" {
		return "0s"
	}
	return delay
}

// GetDelayDuration returns the time.Duration to wait before sending this notification.
func (s *Shoutrrr) GetDelayDuration() (duration time.Duration) {
	duration, _ = time.ParseDuration(s.GetDelay())
	return
}

// GetMaxTries returns the max number of tries allowed for this notification.
func (s *Shoutrrr) GetMaxTries() uint8 {
	tries, _ := strconv.ParseUint(s.GetOption("max_tries"), 10, 8)
	return uint8(tries)
}

// GetOption returns the value for key, or an empty string if it is not present.
func (b *Base) getOption(key string) string {
	return b.Options[key]
}

// setOption sets the value for key.
func (b *Base) setOption(key, value string) {
	b.Options[key] = value
}

// getURLField returns the value for key, or an empty string if it is not present.
func (b *Base) getURLField(key string) string {
	return b.URLFields[key]
}

// setURLField sets the value for key.
func (b *Base) setURLField(key, value string) {
	b.URLFields[key] = value
}

// GetParam returns the value for key, or an empty string if it is not present.
func (b *Base) GetParam(key string) string {
	return b.Params[key]
}

// setParam sets the value for key.
func (b *Base) setParam(key, value string) {
	b.Params[key] = value
}
