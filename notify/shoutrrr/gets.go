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

// GetType resolves the notification type from instance, Main, or ID.
func (s *Shoutrrr) GetType() string {
	// s.ID if the name is the same as the type.
	return util.FirstNonDefault(
		s.Type,
		s.Main.Type,
		s.ID,
	)
}

// Title returns the notification title with info templates evaluated.
func (s *Shoutrrr) Title(info serviceinfo.ServiceInfo) string {
	return util.TemplateString(s.GetParam("title"), info)
}

// Message returns the notification message with info templates evaluated.
func (s *Shoutrrr) Message(info serviceinfo.ServiceInfo) string {
	return util.TemplateString(s.GetOption("message"), info)
}

// GetOption returns the value for key, resolved from instance, Main, Defaults, and HardDefaults in order.
func (s *Shoutrrr) GetOption(key string) string {
	return util.FirstNonDefaultWithEnv(
		s.Options[key],
		s.Main.Options[key],
		s.Defaults.Options[key],
		s.HardDefaults.Options[key],
	)
}

// GetURLField returns the value for key, resolved from instance, Main, Defaults, and HardDefaults in order.
func (s *Shoutrrr) GetURLField(key string) string {
	return util.FirstNonDefaultWithEnv(
		s.URLFields[key],
		s.Main.URLFields[key],
		s.Defaults.URLFields[key],
		s.HardDefaults.URLFields[key],
	)
}

// GetParam returns the value for key, resolved from instance, Main, Defaults, and HardDefaults in order.
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

// GetDelayDuration resolves the send delay as a time.Duration.
func (s *Shoutrrr) GetDelayDuration() (duration time.Duration) {
	duration, _ = time.ParseDuration(s.GetDelay())
	return
}

// GetMaxTries resolves the maximum number of send attempts.
func (s *Shoutrrr) GetMaxTries() uint8 {
	tries, _ := strconv.ParseUint(s.GetOption("max_tries"), 10, 8)
	return uint8(tries)
}

// getOption returns the value for key, or an empty string if not present.
func (b *Base) getOption(key string) string {
	return b.Options[key]
}

// setOption sets the value for key.
func (b *Base) setOption(key, value string) {
	b.Options[key] = value
}

// getURLField returns the value for key, or an empty string if not present.
func (b *Base) getURLField(key string) string {
	return b.URLFields[key]
}

// setURLField sets the value for key.
func (b *Base) setURLField(key, value string) {
	b.URLFields[key] = value
}

// GetParam returns the value for key, or an empty string if not present.
func (b *Base) GetParam(key string) string {
	return b.Params[key]
}

// setParam sets the value for key.
func (b *Base) setParam(key, value string) {
	b.Params[key] = value
}
