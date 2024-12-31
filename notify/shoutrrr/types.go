// Copyright [2024] [Argus]
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
	"fmt"
	"strings"

	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

var (
	jLog           *util.JLog
	supportedTypes = []string{
		"bark", "discord", "smtp", "gotify", "googlechat", "ifttt", "join", "mattermost", "matrix", "ntfy",
		"opsgenie", "pushbullet", "pushover", "rocketchat", "slack", "teams", "telegram", "zulip", "generic", "shoutrrr"}
)

// Slice mapping of Shoutrrr.
type Slice map[string]*Shoutrrr

// String returns a string representation of the Slice.
func (s *Slice) String(prefix string) string {
	if s == nil {
		return ""
	}
	return util.ToYAMLString(s, prefix)
}

// Base is the base Shoutrrr.
type Base struct {
	Type      string            `yaml:"type,omitempty" json:"type,omitempty"`             // Notification type, e.g. slack.
	Options   map[string]string `yaml:"options,omitempty" json:"options,omitempty"`       // Options.
	URLFields map[string]string `yaml:"url_fields,omitempty" json:"url_fields,omitempty"` // URL Fields.
	Params    map[string]string `yaml:"params,omitempty" json:"params,omitempty"`         // Query/Param Props.
}

// SliceDefaults mapping of Defaults.
type SliceDefaults map[string]*Defaults

// String returns a string representation of the SliceDefaults.
func (s *SliceDefaults) String(prefix string) string {
	if s == nil {
		return ""
	}
	if len(*s) == 0 {
		return "{}\n"
	}

	var builder strings.Builder
	keys := util.SortedKeys(*s)

	for _, k := range keys {
		itemStr := (*s)[k].String(prefix + "  ")
		if itemStr != "" {
			delim := "\n"
			if itemStr == "{}\n" {
				delim = " "
			}
			builder.WriteString(fmt.Sprintf("%s%s:%s%s",
				prefix, k, delim, itemStr))
		}
	}

	return builder.String()
}

// Defaults are the default values for Shoutrrr.
type Defaults struct {
	Base `yaml:",inline" json:",inline"`
}

// NewDefaults returns a new Defaults.
func NewDefaults(
	sType string,
	options, urlFields, params map[string]string,
) (defaults *Defaults) {
	// Ensure they are non-nil for mapEnvToStruct.
	// (can edit the map after the struct has been created from the YAML).
	if options == nil {
		options = map[string]string{}
	}
	if params == nil {
		params = map[string]string{}
	}
	if urlFields == nil {
		urlFields = map[string]string{}
	}

	defaults = &Defaults{
		Base{
			Options:   options,
			URLFields: urlFields,
			Type:      sType,
			Params:    params}}
	defaults.InitMaps()
	return
}

// String returns a string representation of the Defaults.
func (d *Defaults) String(prefix string) string {
	if d == nil {
		return ""
	}
	return util.ToYAMLString(d, prefix)
}

// Shoutrrr contains the configuration for a Shoutrrr sender (e.g. Slack).
type Shoutrrr struct {
	Base `yaml:",inline" json:",inline"`

	ID string `yaml:"-" json:"-"` // ID for this Shoutrrr sender.

	Failed        *status.FailsShoutrrr `yaml:"-" json:"-"` // Whether the last send attempt failed.
	ServiceStatus *status.Status        `yaml:"-" json:"-"` // Status of the Service (used for templating commands).

	Main         *Defaults `yaml:"-" json:"-"` // The Shoutrrr that this Shoutrrr is calling (and may override parts of).
	Defaults     *Defaults `yaml:"-" json:"-"` // Default values.
	HardDefaults *Defaults `yaml:"-" json:"-"` // Hardcoded default values.
}

// New Shoutrrr.
func New(
	failed *status.FailsShoutrrr,
	id string,
	sType string,
	options, urlFields, params map[string]string,
	main *Defaults,
	defaults, hardDefaults *Defaults,
) (shoutrrr *Shoutrrr) {
	shoutrrr = &Shoutrrr{
		Base: Base{
			Type:      sType,
			Options:   options,
			URLFields: urlFields,
			Params:    params},
		Failed:        failed,
		ID:            id,
		ServiceStatus: nil,
		Main:          main,
		Defaults:      defaults,
		HardDefaults:  hardDefaults}
	shoutrrr.InitMaps()
	return
}

// String returns a string representation of the Shoutrrr.
func (s *Shoutrrr) String(prefix string) string {
	if s == nil {
		return ""
	}
	return util.ToYAMLString(s, prefix)
}
