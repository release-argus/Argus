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
	"fmt"

	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

var (
	jLog           *util.JLog
	supportedTypes = []string{
		"bark", "discord", "smtp", "gotify", "googlechat", "ifttt", "join", "mattermost", "matrix", "ntfy",
		"opsgenie", "pushbullet", "pushover", "rocketchat", "slack", "teams", "telegram", "zulip", "shoutrrr"}
)

// Slice mapping of Shoutrrr.
type Slice map[string]*Shoutrrr

// String returns a string representation of the Slice.
func (s *Slice) String(prefix string) (str string) {
	if s != nil {
		str = util.ToYAMLString(s, prefix)
	}
	return
}

// ShoutrrrBase is the base Shoutrrr.
type ShoutrrrBase struct {
	Type      string            `yaml:"type,omitempty" json:"type,omitempty"`             // Notification type, e.g. slack
	Options   map[string]string `yaml:"options,omitempty" json:"options,omitempty"`       // Options
	URLFields map[string]string `yaml:"url_fields,omitempty" json:"url_fields,omitempty"` // URL Fields
	Params    map[string]string `yaml:"params,omitempty" json:"params,omitempty"`         // Query/Param Props
}

// SliceDefaults mapping of ShoutrrrDefaults.
type SliceDefaults map[string]*ShoutrrrDefaults

// String returns a string representation of the SliceDefaults.
func (s *SliceDefaults) String(prefix string) (str string) {
	if s == nil {
		return
	}

	keys := util.SortedKeys(*s)
	if len(keys) == 0 {
		str += "{}\n"
	}

	for _, k := range keys {
		itemStr := (*s)[k].String(prefix + "  ")
		if itemStr != "" {
			delim := "\n"
			if itemStr == "{}\n" {
				delim = " "
			}
			str += fmt.Sprintf("%s%s:%s%s",
				prefix, k, delim, itemStr)
		}
	}

	return
}

// ShoutrrrDefaults are the default values for Shoutrrr.
type ShoutrrrDefaults struct {
	ShoutrrrBase `yaml:",inline" json:",inline"`
}

// NewDefaults returns a new ShoutrrrDefaults.
func NewDefaults(
	sType string,
	options *map[string]string,
	params *map[string]string,
	urlFields *map[string]string,
) (defaults *ShoutrrrDefaults) {
	if options == nil {
		options = &map[string]string{}
	}
	if params == nil {
		params = &map[string]string{}
	}
	if urlFields == nil {
		urlFields = &map[string]string{}
	}
	defaults = &ShoutrrrDefaults{
		ShoutrrrBase{
			Options:   *options,
			URLFields: *urlFields,
			Type:      sType,
			Params:    *params}}
	defaults.InitMaps()
	return
}

// String returns a string representation of the ShoutrrrDefaults.
func (s *ShoutrrrDefaults) String(prefix string) (str string) {
	if s != nil {
		str = util.ToYAMLString(s, prefix)
	}
	return
}

type Shoutrrr struct {
	ShoutrrrBase `yaml:",inline" json:",inline"`

	ID string `yaml:"-" json:"-"` // ID for this Shoutrrr sender

	Failed        *svcstatus.FailsShoutrrr `yaml:"-" json:"-"` // Whether the last send attempt failed
	ServiceStatus *svcstatus.Status        `yaml:"-" json:"-"` // Status of the Service (used for templating commands)

	Main         *ShoutrrrDefaults `yaml:"-" json:"-"` // The Shoutrrr that this Shoutrrr is calling (and may override parts of)
	Defaults     *ShoutrrrDefaults `yaml:"-" json:"-"` // Default values
	HardDefaults *ShoutrrrDefaults `yaml:"-" json:"-"` // Harcoded default values
}

// New Shoutrrr.
func New(
	failed *svcstatus.FailsShoutrrr,
	id string,
	options *map[string]string,
	params *map[string]string,
	sType string,
	urlFields *map[string]string,
	main *ShoutrrrDefaults,
	defaults *ShoutrrrDefaults,
	hardDefaults *ShoutrrrDefaults,
) (shoutrrr *Shoutrrr) {
	shoutrrr = &Shoutrrr{
		ShoutrrrBase: ShoutrrrBase{
			Type: sType},
		Failed:        failed,
		ID:            id,
		ServiceStatus: nil,
		Main:          main,
		Defaults:      defaults,
		HardDefaults:  hardDefaults}
	if options != nil {
		shoutrrr.Options = *options
	}
	if params != nil {
		shoutrrr.Params = *params
	}
	if urlFields != nil {
		shoutrrr.URLFields = *urlFields
	}
	shoutrrr.InitMaps()
	return
}

// String returns a string representation of the Shoutrrr.
func (s *Shoutrrr) String(prefix string) (str string) {
	if s != nil {
		str = util.ToYAMLString(s, prefix)
	}
	return
}
