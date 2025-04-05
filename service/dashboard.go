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

// Package service provides the service functionality for Argus.
package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/util"
)

// DashboardOptionsBase are the base options for the Dashboard.
type DashboardOptionsBase struct {
	AutoApprove *bool `yaml:"auto_approve,omitempty" json:"auto_approve,omitempty"` // Default - true = Require approval before sending WebHook(s) for new releases.
}

// DashboardOptionsDefaults are the default values for DashboardOptions.
type DashboardOptionsDefaults struct {
	DashboardOptionsBase `yaml:",inline" json:",inline"`
}

// NewDashboardOptionsDefaults creates a new DashboardOptionsDefaults.
func NewDashboardOptionsDefaults(
	autoApprove *bool,
) DashboardOptionsDefaults {
	return DashboardOptionsDefaults{
		DashboardOptionsBase: DashboardOptionsBase{
			AutoApprove: autoApprove}}
}

// DashboardOptions are options for the Dashboard.
type DashboardOptions struct {
	DashboardOptionsBase `yaml:",inline" json:",inline"`

	Icon       string   `yaml:"icon,omitempty" json:"icon,omitempty"`                 // Icon URL to use for messages/Web UI.
	IconLinkTo string   `yaml:"icon_link_to,omitempty" json:"icon_link_to,omitempty"` // URL to redirect Icon clicks to.
	WebURL     string   `yaml:"web_url,omitempty" json:"web_url,omitempty"`           // URL to provide on the Web UI.
	Tags       []string `yaml:"tags,omitempty" json:"tags,omitempty"`                 // Tags for the Service.

	Defaults     *DashboardOptionsDefaults `yaml:"-" json:"-"` // Defaults.
	HardDefaults *DashboardOptionsDefaults `yaml:"-" json:"-"` // Hard defaults.
}

// NewDashboardOptions creates a new DashboardOptions.
func NewDashboardOptions(
	autoApprove *bool,
	icon string,
	iconLinkTo string,
	webURL string,
	tags []string,
	defaults, hardDefaults *DashboardOptionsDefaults,
) *DashboardOptions {
	return &DashboardOptions{
		DashboardOptionsBase: DashboardOptionsBase{
			AutoApprove: autoApprove},
		Icon:         icon,
		IconLinkTo:   iconLinkTo,
		WebURL:       webURL,
		Tags:         tags,
		Defaults:     defaults,
		HardDefaults: hardDefaults}
}

// UnmarshalJSON handles the unmarshalling of a DashboardOptions.
func (d *DashboardOptions) UnmarshalJSON(data []byte) error {
	baseErr := "failed to unmarshal service.DashboardOptions:"

	aux := &struct {
		*DashboardOptionsBase `json:",inline"`

		Icon       *string         `json:"icon,omitempty"`         // Icon URL to use for messages/Web UI.
		IconLinkTo *string         `json:"icon_link_to,omitempty"` // URL to redirect Icon clicks to.
		WebURL     *string         `json:"web_url,omitempty"`      // URL to provide on the Web UI.
		Tags       json.RawMessage `json:"tags,omitempty"`         // Tags for the Service.
	}{
		DashboardOptionsBase: &d.DashboardOptionsBase,
		Icon:                 &d.Icon,
		IconLinkTo:           &d.IconLinkTo,
		WebURL:               &d.WebURL,
	}

	// Unmarshal into aux.
	if err := json.Unmarshal(data, &aux); err != nil {
		errStr := util.FormatUnmarshalError("json", err)
		errStr = strings.ReplaceAll(errStr, "\n", "\n  ")
		return fmt.Errorf("%s\n  %s",
			baseErr, errStr)
	}

	// Tags
	if len(aux.Tags) > 0 {
		var tagsAsString string
		var tagsAsArray []string

		// Try to unmarshal as a list of strings
		if err := json.Unmarshal(aux.Tags, &tagsAsArray); err == nil {
			d.Tags = tagsAsArray
			// Try to unmarshal as a single string
		} else if err := json.Unmarshal(aux.Tags, &tagsAsString); err == nil {
			d.Tags = []string{tagsAsString}
		} else {
			return errors.New(baseErr + "\n  tags: <invalid> (expected a string or a list of strings)")
		}
	}

	return nil
}

// UnmarshalYAML handles the unmarshalling of a DashboardOptions.
func (d *DashboardOptions) UnmarshalYAML(value *yaml.Node) error {
	baseErr := "failed to unmarshal service.DashboardOptions:"

	aux := &struct {
		*DashboardOptionsBase `yaml:",inline"`

		Icon       *string      `yaml:"icon,omitempty"`         // Icon URL to use for messages/Web UI.
		IconLinkTo *string      `yaml:"icon_link_to,omitempty"` // URL to redirect Icon clicks to.
		WebURL     *string      `yaml:"web_url,omitempty"`      // URL to provide on the Web UI.
		Tags       util.RawNode `yaml:"tags,omitempty"`         // Tags for the Service.
	}{
		DashboardOptionsBase: &d.DashboardOptionsBase,
		Icon:                 &d.Icon,
		IconLinkTo:           &d.IconLinkTo,
		WebURL:               &d.WebURL,
	}

	// Unmarshal into aux.
	if err := value.Decode(&aux); err != nil {
		errStr := util.FormatUnmarshalError("yaml", err)
		errStr = strings.ReplaceAll(errStr, "\n", "\n  ")
		return fmt.Errorf("%s\n  %s",
			baseErr, errStr)
	}

	// Tags
	if aux.Tags.Node != nil {
		var tagsAsString string
		var tagsAsArray []string

		// Try to unmarshal as a list of strings
		if err := aux.Tags.Decode(&tagsAsArray); err == nil {
			d.Tags = tagsAsArray
			// Try to unmarshal as a single string
		} else if err := aux.Tags.Decode(&tagsAsString); err == nil {
			d.Tags = []string{tagsAsString}
		} else {
			return errors.New(baseErr + "\n  tags: <invalid> (expected a string or a list of strings)")
		}
	}

	return nil
}

// GetAutoApprove returns whether new releases are auto-approved.
func (d *DashboardOptions) GetAutoApprove() bool {
	return *util.FirstNonDefault(
		d.AutoApprove,
		d.Defaults.AutoApprove,
		d.HardDefaults.AutoApprove)
}

// CheckValues validates the fields of the DashboardOptions struct.
func (d *DashboardOptions) CheckValues(prefix string) error {
	if d == nil {
		return nil
	}

	if !util.CheckTemplate(d.WebURL) {
		return fmt.Errorf("%sweb_url: %q <invalid> (didn't pass templating)",
			prefix, d.WebURL)
	}

	return nil
}
