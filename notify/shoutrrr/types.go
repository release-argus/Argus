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
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

// #############
// # CONSTANTS #
// #############

var (
	supportedTypes = []string{
		"bark", "discord", "smtp", "gotify", "googlechat", "ifttt", "join", "mattermost", "matrix", "ntfy",
		"opsgenie", "pushbullet", "pushover", "rocketchat", "slack", "teams", "telegram", "zulip", "generic",
		"shoutrrr",
	}
)

// #########
// # TYPES #
// #########

// Config holds root, default, and hard-default Shoutrrr configuration.
type Config struct {
	Root         ShoutrrrsDefaults
	Defaults     ShoutrrrsDefaults
	HardDefaults ShoutrrrsDefaults
}

// Shoutrrrs is a string map of Shoutrrr.
type Shoutrrrs map[string]*Shoutrrr

// Base is the base Shoutrrr.
type Base struct {
	Type      string                  `json:"type,omitempty" yaml:"type,omitempty"`             // Notification type, e.g. slack.
	Options   MapStringStringOmitNull `json:"options,omitempty" yaml:"options,omitempty"`       // Options.
	URLFields MapStringStringOmitNull `json:"url_fields,omitempty" yaml:"url_fields,omitempty"` // URL Fields.
	Params    MapStringStringOmitNull `json:"params,omitempty" yaml:"params,omitempty"`         // Query/Param Props.
}

// Shoutrrr contains the configuration for a Shoutrrr sender (e.g. Slack).
type Shoutrrr struct {
	Base `json:",inline" yaml:",inline"`

	ID string `json:"name,omitempty" yaml:"-"` // ID for this Shoutrrr sender.

	Failed        *status.FailsShoutrrr `json:"-" yaml:"-"` // Whether the last send attempt failed.
	ServiceStatus *status.Status        `json:"-" yaml:"-"` // Status of the Service (used for templating commands).

	Main         *Defaults `json:"-" yaml:"-"` // The root Shoutrrr configuration that this instance may partially override.
	Defaults     *Defaults `json:"-" yaml:"-"` // Default values.
	HardDefaults *Defaults `json:"-" yaml:"-"` // Hardcoded default values.
}

// ShoutrrrsDefaults is a string map of Defaults.
type ShoutrrrsDefaults map[string]*Defaults

// Defaults are the default values for Shoutrrr.
type Defaults struct {
	Base `json:",inline" yaml:",inline"`
}

// ############
// # DECODING #
// ############

// MarshalJSON implements json.Marshaler for Shoutrrrs.
//
// Each map entry is encoded as a Shoutrrr object in a JSON array.
// The map key is not included in the output; it is already stored in the Shoutrrr.ID field.
// The output array is sorted by key to ensure deterministic ordering.
//
// For example (input order is not preserved):
//
//	{
//		"b": { id: "b", ... },
//		"a": { id: "a", ... }
//	}
//
// is marshaled as:
//
//	[
//		{ id: "a", ... },
//		{ id: "b", ... }
//	]
func (s *Shoutrrrs) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}

	keys := util.SortedKeys(*s)
	arr := make([]*Shoutrrr, 0, len(*s))
	for _, key := range keys {
		arr = append(arr, (*s)[key])
	}
	return decode.Marshal("json", arr) //nolint:wrapcheck
}

// UnmarshalJSON implements json.Unmarshaler for Shoutrrrs.
//
// It supports a JSON array of Shoutrrr entries:
//
//	[
//		{ id: "a", ... },
//		{ id: "b", ... }
//	]
//
// which is converted into a map keyed by each entry's ID field.
func (s *Shoutrrrs) UnmarshalJSON(data []byte) error {
	var arr []Shoutrrr
	if err := decode.Unmarshal("json", data, &arr); err != nil {
		return err //nolint:wrapcheck
	}
	*s = make(Shoutrrrs, len(arr))

	for i := range arr {
		(*s)[arr[i].ID] = &arr[i]
	}
	return nil
}

// New returns a new Shoutrrr.
func New(
	failed *status.FailsShoutrrr,
	id string,
	sType string,
	options, urlFields, params MapStringStringOmitNull,
	main *Defaults,
	defaults, hardDefaults *Defaults,
) (shoutrrr *Shoutrrr) {
	shoutrrr = &Shoutrrr{
		Base: Base{
			Type:      sType,
			Options:   options,
			URLFields: urlFields,
			Params:    params,
		},
		Failed:        failed,
		ID:            id,
		ServiceStatus: nil,
		Main:          main,
		Defaults:      defaults,
		HardDefaults:  hardDefaults,
	}
	shoutrrr.InitMaps()
	return
}

// NewDefaults returns a new Defaults.
func NewDefaults(
	sType string,
	options, urlFields, params MapStringStringOmitNull,
) (defaults *Defaults) {
	// Ensure they are non-nil for mapEnvToStruct.
	// (can edit the map after the struct has been created from the YAML).
	if options == nil {
		options = MapStringStringOmitNull{}
	}
	if params == nil {
		params = MapStringStringOmitNull{}
	}
	if urlFields == nil {
		urlFields = MapStringStringOmitNull{}
	}

	defaults = &Defaults{
		Base{
			Options:   options,
			URLFields: urlFields,
			Type:      sType,
			Params:    params,
		},
	}
	defaults.InitMaps()
	return
}

// Copy returns a deep copy of the receiver.
func (s Shoutrrrs) Copy(serviceStatus *status.Status) Shoutrrrs {
	field := make(Shoutrrrs, len(s))
	for k, v := range s {
		field[k] = v.Copy(serviceStatus)
	}
	return field
}

// Copy returns a deep copy of the receiver.
func (s *Shoutrrr) Copy(serviceStatus *status.Status) *Shoutrrr {
	if s == nil {
		return nil
	}

	field := Shoutrrr{
		Base: Base{
			Type:      s.Type,
			Options:   *s.Options.Copy(),
			URLFields: *s.URLFields.Copy(),
			Params:    *s.Params.Copy(),
		},
		ID:            s.ID,
		Failed:        s.Failed.Copy(),
		ServiceStatus: serviceStatus,
		Main:          s.Main,
		Defaults:      s.Defaults,
		HardDefaults:  s.HardDefaults,
	}

	return &field
}

// #########
// # STATE #
// #########

// IsZero implements the yaml.IsZeroer interface.
func (s ShoutrrrsDefaults) IsZero() bool {
	for _, v := range s {
		if !v.IsZero() {
			return false
		}
	}
	return true
}

// IsZero implements the yaml.IsZeroer interface.
func (d *Defaults) IsZero() bool {
	if d == nil {
		return true
	}

	return d.Type == "" &&
		len(d.Options) == 0 &&
		len(d.URLFields) == 0 &&
		len(d.Params) == 0
}

// IsZero implements the yaml.IsZeroer interface.
func (s *Shoutrrrs) IsZero() bool {
	if s == nil {
		return true
	}

	return len(*s) == 0
}

// IsDefault reports whether all Shoutrrr fields are at their default (zero) values.
func (s *Shoutrrr) IsDefault() bool {
	return len(s.Options) == 0 &&
		len(s.URLFields) == 0 &&
		len(s.Params) == 0
}

// #############
// # STRINGIFY #
// #############

// String returns a string representation of the receiver.
func (s *Shoutrrrs) String(prefix string) string {
	if s == nil {
		return ""
	}
	return decode.ToYAMLString(s, prefix)
}

// String returns a string representation of the receiver.
func (s *Shoutrrr) String(prefix string) string {
	if s == nil {
		return ""
	}
	return decode.ToYAMLString(s, prefix)
}

// String returns a string representation of the receiver.
func (s *ShoutrrrsDefaults) String(prefix string) string {
	if s == nil {
		return ""
	}

	return decode.ToYAMLString(s, prefix)
}

// String returns a string representation of the receiver.
func (d *Defaults) String(prefix string) string {
	if d == nil {
		return ""
	}
	return decode.ToYAMLString(d, prefix)
}
