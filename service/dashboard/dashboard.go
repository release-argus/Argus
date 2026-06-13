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

// Package dashboard provides options for a Service.
package dashboard

import (
	"fmt"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/util"
)

// #########
// # TYPES #
// #########

// OptionsBase are the base options for the Dashboard.
type OptionsBase struct {
	AutoApprove *bool  `json:"auto_approve,omitempty" yaml:"auto_approve,omitempty"` // Default - true = Require approval before sending WebHooks for new releases.
	Icon        string `json:"icon,omitempty" yaml:"icon,omitempty"`                 // Icon URL to use for messages/Web UI.
	IconLinkTo  string `json:"icon_link_to,omitempty" yaml:"icon_link_to,omitempty"` // URL to redirect Icon clicks to.
	WebURL      string `json:"web_url,omitempty" yaml:"web_url,omitempty"`           // URL to provide on the Web UI.
}

// Options are options for the Dashboard.
type Options struct {
	OptionsBase `json:",inline" yaml:",inline"`

	iconExpanded       *string  // Icon URL after env var expansion. (nil if no env var expansion).
	iconNotify         *string  // Fallback icon URL from a Notify after env var expansion. (nil if we already have an Icon).
	iconLinkToExpanded *string  // URL after env var expansion. (nil if no env var expansion).
	webURLExpanded     *string  // URL after env var expansion. (nil if no env var expansion).
	Tags               []string `json:"tags,omitempty" yaml:"tags,omitempty"` // Tags for the Service.

	Defaults     *Defaults `json:"-" yaml:"-"` // Defaults.
	HardDefaults *Defaults `json:"-" yaml:"-"` // Hard defaults.
}

// OptionsDecode is an unmarshal-only helper for [Options].
type OptionsDecode struct {
	OptionsBase `json:",inline" yaml:",inline"`

	Tags any `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// ############
// # DECODING #
// ############

// UnmarshalJSON implements the json.Unmarshaler interface.
func (o *Options) UnmarshalJSON(data []byte) error {
	return o.unmarshal("json", data)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (o *Options) UnmarshalYAML(data []byte) error {
	return o.unmarshal("yaml", data)
}

// unmarshal implements the format.Unmarshaler interface.
func (o *Options) unmarshal(format string, data []byte) error {
	aux := OptionsDecode{
		OptionsBase: o.OptionsBase,
	}

	// Unmarshal into aux.
	if err := decode.Unmarshal(format, data, &aux); err != nil {
		return err //nolint:wrapcheck
	}
	o.OptionsBase = aux.OptionsBase
	o.unmarshalEnvVars()

	// Tags
	if aux.Tags != nil {
		tags, err := unmarshalStringOrStringList("tags", aux.Tags)
		if err != nil {
			return err
		}
		o.Tags = tags
	}

	return nil
}

// unmarshalEnvVars expands environment variables in icon and URL fields.
func (o *Options) unmarshalEnvVars() {
	// Unmarshal env vars for icon, iconLinkTo, and webURL.
	o.iconExpanded = util.TryExpandEnv(o.Icon)
	o.iconLinkToExpanded = util.TryExpandEnv(o.IconLinkTo)
	o.webURLExpanded = util.TryExpandEnv(o.WebURL)
}

// Decode creates and returns new [Options] from format-encoded data.
func Decode(
	format string,
	data []byte,
	cfg DefaultsConfig,
) (*Options, error) {
	field := Options{
		Defaults:     cfg.Soft,
		HardDefaults: cfg.Hard,
	}

	if err := decode.Unmarshal(format, data, &field); err != nil {
		return nil, &decode.KeyFieldError{
			Key: "dashboard",
			Err: err,
		}
	}

	return &field, nil
}

// unmarshalStringOrStringList decodes tags from either a string or string list.
func unmarshalStringOrStringList(field string, input any) ([]string, error) {
	switch tags := input.(type) {
	case string:
		return []string{tags}, nil
	case []any:
		res := make([]string, 0, len(tags))
		for _, tag := range tags {
			t, ok := tag.(string)
			if !ok {
				return nil, fmt.Errorf(
					"%s: expected a string inside the list, got %T",
					field, tag,
				)
			}
			res = append(res, t)
		}
		return res, nil
	default:
		return nil, fmt.Errorf(
			"%s: expected a string or a list of strings, got %T",
			field, tags,
		)
	}
}

// #########
// # STATE #
// #########

// IsZero implements the yaml.IsZeroer interface.
func (o *OptionsBase) IsZero() bool {
	return o.AutoApprove == nil &&
		o.Icon == "" &&
		o.IconLinkTo == "" &&
		o.WebURL == ""
}

// IsZero implements the yaml.IsZeroer interface.
func (o Options) IsZero() bool {
	return o.AutoApprove == nil &&
		o.Icon == "" &&
		o.IconLinkTo == "" &&
		o.WebURL == "" &&
		len(o.Tags) == 0 &&
		o.OptionsBase.IsZero()
}

// Copy returns a deep copy of the receiver.
func (o *Options) Copy() *Options {
	if o == nil {
		return nil
	}

	newOptions := &Options{
		OptionsBase: OptionsBase{
			Icon:       o.Icon,
			IconLinkTo: o.IconLinkTo,
			WebURL:     o.WebURL,
		},
		Tags:         o.Tags,
		Defaults:     o.Defaults,
		HardDefaults: o.HardDefaults,
	}
	newOptions.AutoApprove = copyPtr(o.AutoApprove)
	newOptions.iconExpanded = copyPtr(o.iconExpanded)
	newOptions.iconNotify = copyPtr(o.iconNotify)
	newOptions.iconLinkToExpanded = copyPtr(o.iconLinkToExpanded)
	newOptions.webURLExpanded = copyPtr(o.webURLExpanded)

	return newOptions
}

// copyPtr returns a copy of v when non-nil.
func copyPtr[T any](v *T) *T {
	if v == nil {
		return nil
	}
	cp := *v
	return &cp
}

// ##########
// # VALUES #
// ##########

// GetAutoApprove returns whether new releases are auto-approved.
func (o *Options) GetAutoApprove() bool {
	return *util.FirstNonDefault(
		o.AutoApprove,
		o.Defaults.AutoApprove,
		o.HardDefaults.AutoApprove,
	)
}

// SetFallbackIcon sets the icon URL from a Notify.
func (o *Options) SetFallbackIcon(iconNotify string) {
	o.iconNotify = &iconNotify
}

// GetIcon returns the icon URL.
func (o *Options) GetIcon() string {
	return util.ValueOr(
		util.DerefOr(o.iconExpanded, o.Icon),
		util.DerefOrZero(o.iconNotify),
	)
}

// GetIconLinkTo returns the icon link URL.
func (o *Options) GetIconLinkTo() string {
	return util.DerefOr(
		o.iconLinkToExpanded, o.IconLinkTo,
	)
}

// GetWebURL returns the web URL.
func (o *Options) GetWebURL() string {
	return util.DerefOr(
		o.webURLExpanded, o.WebURL,
	)
}

// ############
// # DEFAULTS #
// ############

// SetDefaults assigns defaults and hardDefaults to the receiver.
func (o *Options) SetDefaults(defaults, hardDefaults *Defaults) {
	o.Defaults = defaults
	o.HardDefaults = hardDefaults
}

// ##############
// # VALIDATION #
// ##############

// CheckValues validates the fields of the receiver.
func (o *Options) CheckValues() error {
	if o == nil {
		return nil
	}

	if !util.CheckTemplate(o.WebURL) {
		return &decode.FieldError{
			Key:         "web_url",
			Value:       o.WebURL,
			Description: "didn't pass templating",
		}
	}

	return nil
}
