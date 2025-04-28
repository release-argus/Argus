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

// Package dashboard provides options for a Service in the
package dashboard

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/util"
)

// OptionsBase are the base options for the Dashboard.
type OptionsBase struct {
	AutoApprove *bool `json:"auto_approve,omitempty" yaml:"auto_approve,omitempty"` // Default - true = Require approval before sending WebHook(s) for new releases.
}

// Options are options for the Dashboard.
type Options struct {
	OptionsBase `json:",inline" yaml:",inline"`

	Icon               string   `json:"icon,omitempty" yaml:"icon,omitempty"` // Icon URL to use for messages/Web UI.
	iconExpanded       *string  // Icon URL after env var expansion. (nil if no env var expansion).
	iconNotify         *string  // Faillback icon URL from a Notify after env var expansion. (nil if we already have an Icon).
	IconLinkTo         string   `json:"icon_link_to,omitempty" yaml:"icon_link_to,omitempty"` // URL to redirect Icon clicks to.
	iconLinkToExpanded *string  // URL after env var expansion. (nil if no env var expansion).
	WebURL             string   `json:"web_url,omitempty" yaml:"web_url,omitempty"` // URL to provide on the Web UI.
	webURLExpanded     *string  // URL after env var expansion. (nil if no env var expansion).
	Tags               []string `json:"tags,omitempty" yaml:"tags,omitempty"` // Tags for the Service.

	Defaults     *OptionsDefaults `json:"-" yaml:"-"` // Defaults.
	HardDefaults *OptionsDefaults `json:"-" yaml:"-"` // Hard defaults.
}

// NewOptions creates a new Options.
func NewOptions(
	autoApprove *bool,
	icon string,
	iconLinkTo string,
	webURL string,
	tags []string,
	defaults, hardDefaults *OptionsDefaults,
) *Options {
	return &Options{
		OptionsBase: OptionsBase{
			AutoApprove: autoApprove},
		Icon:         icon,
		IconLinkTo:   iconLinkTo,
		WebURL:       webURL,
		Tags:         tags,
		Defaults:     defaults,
		HardDefaults: hardDefaults}
}

func (o *Options) unmarshalEnvVars() {
	// Unmarshal env vars for icon, iconLinkTo, and webURL.
	o.iconExpanded = util.TryExpandEnv(o.Icon)
	o.iconLinkToExpanded = util.TryExpandEnv(o.IconLinkTo)
	o.webURLExpanded = util.TryExpandEnv(o.WebURL)
}

// UnmarshalJSON handles the unmarshalling of Options.
func (o *Options) UnmarshalJSON(data []byte) error {
	baseErr := "failed to unmarshal service.Dashboard:"

	aux := &struct {
		*OptionsBase `json:",inline"`

		Icon       *string         `json:"icon,omitempty"`         // Icon URL to use for messages/Web UI.
		IconLinkTo *string         `json:"icon_link_to,omitempty"` // URL to redirect Icon clicks to.
		WebURL     *string         `json:"web_url,omitempty"`      // URL to provide on the Web UI.
		Tags       json.RawMessage `json:"tags,omitempty"`         // Tags for the Service.
	}{
		OptionsBase: &o.OptionsBase,
		Icon:        &o.Icon,
		IconLinkTo:  &o.IconLinkTo,
		WebURL:      &o.WebURL,
	}

	// Unmarshal into aux.
	if err := json.Unmarshal(data, &aux); err != nil {
		errStr := util.FormatUnmarshalError("json", err)
		errStr = strings.ReplaceAll(errStr, "\n", "\n  ")
		return fmt.Errorf("%s\n  %s",
			baseErr, errStr)
	}
	o.unmarshalEnvVars()

	// Tags
	if len(aux.Tags) > 0 {
		tags, err := unmarshalStringOrStringList(
			func(out *[]string) error { return json.Unmarshal(aux.Tags, out) },
			func(out *string) error { return json.Unmarshal(aux.Tags, out) },
		)

		if err != nil {
			return errors.New(baseErr + "\n  tags: <invalid> (expected a string or a list of strings)")
		}
		o.Tags = tags
	}

	return nil
}

// UnmarshalYAML handles the unmarshalling of Options.
func (o *Options) UnmarshalYAML(value *yaml.Node) error {
	baseErr := "failed to unmarshal service.Dashboard:"

	aux := &struct {
		*OptionsBase `yaml:",inline"`

		Icon       *string      `yaml:"icon,omitempty"`         // Icon URL to use for messages/Web UI.
		IconLinkTo *string      `yaml:"icon_link_to,omitempty"` // URL to redirect Icon clicks to.
		WebURL     *string      `yaml:"web_url,omitempty"`      // URL to provide on the Web UI.
		Tags       util.RawNode `yaml:"tags,omitempty"`         // Tags for the Service.
	}{
		OptionsBase: &o.OptionsBase,
		Icon:        &o.Icon,
		IconLinkTo:  &o.IconLinkTo,
		WebURL:      &o.WebURL,
	}

	// Unmarshal into aux.
	if err := value.Decode(&aux); err != nil {
		errStr := util.FormatUnmarshalError("yaml", err)
		errStr = strings.ReplaceAll(errStr, "\n", "\n  ")
		return fmt.Errorf("%s\n  %s",
			baseErr, errStr)
	}
	o.unmarshalEnvVars()

	// Tags
	if aux.Tags.Node != nil {
		tags, err := unmarshalStringOrStringList(
			func(out *[]string) error { return aux.Tags.Decode(out) },
			func(out *string) error { return aux.Tags.Decode(out) },
		)

		if err != nil {
			return errors.New(baseErr + "\n  tags: <invalid> (expected a string or a list of strings)")
		}
		o.Tags = tags
	}

	return nil
}

func unmarshalStringOrStringList(
	unmarshalList func(out *[]string) error,
	unmarshalString func(out *string) error,
) ([]string, error) {
	var (
		list   []string
		single string
	)
	if err := unmarshalList(&list); err == nil {
		return list, nil
	}
	if err := unmarshalString(&single); err == nil {
		return []string{single}, nil
	}
	return nil, errors.New("expected a string or a list of strings")
}

// Copy creates a copy of the Options struct.
func (o *Options) Copy() *Options {
	if o == nil {
		return nil
	}

	newOptions := &Options{
		Icon:         o.Icon,
		IconLinkTo:   o.IconLinkTo,
		WebURL:       o.WebURL,
		Tags:         o.Tags,
		Defaults:     o.Defaults,
		HardDefaults: o.HardDefaults}

	if o.AutoApprove != nil {
		autoApprove := *o.AutoApprove
		newOptions.AutoApprove = &autoApprove
	}
	if o.iconExpanded != nil {
		iconExpanded := *o.iconExpanded
		newOptions.iconExpanded = &iconExpanded
	}
	if o.iconNotify != nil {
		iconNotify := *o.iconNotify
		newOptions.iconNotify = &iconNotify
	}
	if o.iconLinkToExpanded != nil {
		iconLinkToExpanded := *o.iconLinkToExpanded
		newOptions.iconLinkToExpanded = &iconLinkToExpanded
	}
	if o.webURLExpanded != nil {
		webURLExpanded := *o.webURLExpanded
		newOptions.webURLExpanded = &webURLExpanded
	}

	return newOptions
}

// GetAutoApprove returns whether new releases are auto-approved.
func (o *Options) GetAutoApprove() bool {
	return *util.FirstNonDefault(
		o.AutoApprove,
		o.Defaults.AutoApprove,
		o.HardDefaults.AutoApprove)
}

// SetFallbackIcon sets the icon URL from a Notify.
func (o *Options) SetFallbackIcon(iconNotify string) {
	o.iconNotify = &iconNotify
}

// GetIcon returns the icon URL.
func (o *Options) GetIcon() string {
	return util.ValueOrValue(
		util.DereferenceOrValue(
			o.iconExpanded, o.Icon),
		util.DereferenceOrDefault(
			o.iconNotify))
}

// GetIconLinkTo returns the icon link URL.
func (o *Options) GetIconLinkTo() string {
	return util.DereferenceOrValue(
		o.iconLinkToExpanded, o.IconLinkTo)
}

// GetWebURL returns the web URL.
func (o *Options) GetWebURL() string {
	return util.DereferenceOrValue(
		o.webURLExpanded, o.WebURL)
}

// CheckValues validates the fields of the Options struct.
func (o *Options) CheckValues(prefix string) error {
	if o == nil {
		return nil
	}

	if !util.CheckTemplate(o.WebURL) {
		return fmt.Errorf("%sweb_url: %q <invalid> (didn't pass templating)",
			prefix, o.WebURL)
	}

	return nil
}
