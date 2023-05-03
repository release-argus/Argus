// Copyright [2023] [Argus]
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

package shoutrrr

import (
	"strconv"
	"time"

	"github.com/release-argus/Argus/util"
)

// GetOptions map from this Shoutrrr.
func (s *ShoutrrrBase) GetOptions() *map[string]string {
	return &s.Options
}

// GetOption from this/Main/Defaults/HardDefaults on FiFo.
func (s *Shoutrrr) GetOption(key string) string {
	return util.GetFirstNonDefault(
		s.Options[key],
		s.Main.Options[key],
		s.Defaults.Options[key],
		s.HardDefaults.Options[key])
}

// GetSelfOption gets Options[key] from this Shoutrrr
func (s *ShoutrrrBase) GetSelfOption(key string) string {
	return s.Options[key]
}

// SetOption[key] to value
func (s *ShoutrrrBase) SetOption(key string, value string) {
	s.Options[key] = value
}

// GetParams map from this Shoutrrr.
func (s *ShoutrrrBase) GetParams() *map[string]string {
	return &s.Params
}

// GetParam from this/Main/Defaults/HardDefaults on FiFo
func (s *Shoutrrr) GetParam(key string) string {
	return util.GetFirstNonDefault(
		s.Params[key],
		s.Main.Params[key],
		s.Defaults.Params[key],
		s.HardDefaults.Params[key])
}

// GetSelfParam gets Params[key] from this Shoutrrr
func (s *ShoutrrrBase) GetSelfParam(key string) string {
	return s.Params[key]
}

// SetParam[key] to value
func (s *ShoutrrrBase) SetParam(key string, value string) {
	s.Params[key] = value
}

// GetURLFields map from this Shoutrrr.
func (s *ShoutrrrBase) GetURLFields() *map[string]string {
	return &s.URLFields
}

// GetURLField from this/Main/Defaults/HardDefaults on FiFo
func (s *Shoutrrr) GetURLField(key string) string {
	return util.GetFirstNonDefault(
		s.URLFields[key],
		s.Main.URLFields[key],
		s.Defaults.URLFields[key],
		s.HardDefaults.URLFields[key])
}

// GetSelfURLField gets URLFields[key] from this Shoutrrr
func (s *ShoutrrrBase) GetSelfURLField(key string) string {
	return s.URLFields[key]
}

// SetURLField[key] to value
func (s *ShoutrrrBase) SetURLField(key string, value string) {
	s.URLFields[key] = value
}

// GetDelay before sending.
func (s *Shoutrrr) GetDelay() string {
	delay := s.GetOption("delay")
	if delay == "" {
		return "0s"
	}
	return delay
}

// GetDelayDuration before sending.
func (s *Shoutrrr) GetDelayDuration() (duration time.Duration) {
	duration, _ = time.ParseDuration(s.GetDelay())
	return
}

// GetMaxTries allowed for the Gotification.
func (s *Shoutrrr) GetMaxTries() uint {
	tries, _ := strconv.ParseUint(s.GetOption("max_tries"), 10, 32)
	return uint(tries)
}

// GetMessage of the Gotification after the context is applied and template evaluated.
func (s *Shoutrrr) GetMessage(context *util.ServiceInfo) string {
	return util.TemplateString(s.GetOption("message"), *context)
}

// GetTitle of the Shoutrrr after the context is applied and template evaluated.
func (s *Shoutrrr) GetTitle(serviceInfo *util.ServiceInfo) (title string) {
	title = util.GetFirstNonDefault(
		s.GetSelfParam("title"),
		s.Main.GetSelfParam("title"),
		s.Defaults.GetSelfParam("title"),
		s.HardDefaults.GetSelfParam("title"))
	title = util.TemplateString(title, *serviceInfo)
	return
}

// GetSelfType of this Shoutrrr.
func (s *ShoutrrrBase) GetSelfType() string {
	return s.Type
}

// GetType of this Shoutrrr.
func (s *Shoutrrr) GetType() string {
	// s.ID if the name is the same as the type
	return util.GetFirstNonDefault(s.Type, s.Main.Type, s.ID)
}
