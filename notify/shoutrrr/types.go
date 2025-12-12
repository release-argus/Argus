// Copyright [2025] [Argus]
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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

var (
	supportedTypes = []string{
		"bark", "discord", "smtp", "gotify", "googlechat", "ifttt", "join", "mattermost", "matrix", "ntfy",
		"opsgenie", "pushbullet", "pushover", "rocketchat", "slack", "teams", "telegram", "zulip", "generic",
		"shoutrrr"}
)

// Shoutrrrs is a string map of Shoutrrr.
type Shoutrrrs map[string]*Shoutrrr

// UnmarshalJSON converts a JSON array to a Shoutrrrs map.
func (s *Shoutrrrs) UnmarshalJSON(data []byte) error {
	var arr []Shoutrrr
	if err := json.Unmarshal(data, &arr); err != nil {
		return err //nolint:wrapcheck
	}
	*s = make(Shoutrrrs, len(arr))

	for i := range arr {
		(*s)[arr[i].ID] = &arr[i]
	}
	return nil
}

// MarshalJSON marshals into a JSON array of Shoutrrr values.
func (s *Shoutrrrs) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}

	keys := util.SortedKeys(*s)
	arr := make([]*Shoutrrr, 0, len(*s))
	for _, key := range keys {
		arr = append(arr, (*s)[key])
	}
	return json.Marshal(arr) //nolint:wrapcheck
}

// String returns a string representation of the Shoutrrrs.
func (s *Shoutrrrs) String(prefix string) string {
	if s == nil {
		return ""
	}
	return util.ToYAMLString(s, prefix)
}

// Base is the base Shoutrrr.
type Base struct {
	Type      string                       `json:"type,omitempty" yaml:"type,omitempty"`             // Notification type, e.g. slack.
	Options   util.MapStringStringOmitNull `json:"options,omitempty" yaml:"options,omitempty"`       // Options.
	URLFields util.MapStringStringOmitNull `json:"url_fields,omitempty" yaml:"url_fields,omitempty"` // URL Fields.
	Params    util.MapStringStringOmitNull `json:"params,omitempty" yaml:"params,omitempty"`         // Query/Param Props.
}

// ShoutrrrsDefaults is a string map of Defaults.
type ShoutrrrsDefaults map[string]*Defaults

// String returns a string representation of the ShoutrrrsDefaults.
func (s *ShoutrrrsDefaults) String(prefix string) string {
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
	Base `json:",inline" yaml:",inline"`
}

// NewDefaults returns a new Defaults.
func NewDefaults(
	sType string,
	options, urlFields, params util.MapStringStringOmitNull,
) (defaults *Defaults) {
	// Ensure they are non-nil for mapEnvToStruct.
	// (can edit the map after the struct has been created from the YAML).
	if options == nil {
		options = util.MapStringStringOmitNull{}
	}
	if params == nil {
		params = util.MapStringStringOmitNull{}
	}
	if urlFields == nil {
		urlFields = util.MapStringStringOmitNull{}
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
	Base `json:",inline" yaml:",inline"`

	ID string `json:"name,omitempty" yaml:"-"` // ID for this Shoutrrr sender.

	Failed        *status.FailsShoutrrr `json:"-" yaml:"-"` // Whether the last send attempt failed.
	ServiceStatus *status.Status        `json:"-" yaml:"-"` // Status of the Service (used for templating commands).

	Main         *Defaults `json:"-" yaml:"-"` // The Shoutrrr that this Shoutrrr is calling (and may override parts of).
	Defaults     *Defaults `json:"-" yaml:"-"` // Default values.
	HardDefaults *Defaults `json:"-" yaml:"-"` // Hardcoded default values.
}

// New Shoutrrr.
func New(
	failed *status.FailsShoutrrr,
	id string,
	sType string,
	options, urlFields, params util.MapStringStringOmitNull,
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
