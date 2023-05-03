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

package service

import (
	"fmt"

	"github.com/release-argus/Argus/util"
)

// DashboardOptionsBase are the base options for the Dashboard.
type DashboardOptionsBase struct {
	AutoApprove *bool `yaml:"auto_approve,omitempty" json:"auto_approve,omitempty"` // default - true = Requre approval before sending WebHook(s) for new releases
}

// DashboardOptionsDefaults are the default values for DashboardOptions.
type DashboardOptionsDefaults struct {
	DashboardOptionsBase `yaml:",inline" json:",inline"`
}

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

	Icon       string `yaml:"icon,omitempty" json:"icon,omitempty"`                 // Icon URL to use for messages/Web UI
	IconLinkTo string `yaml:"icon_link_to,omitempty" json:"icon_link_to,omitempty"` // URL to redirect Icon clicks to
	WebURL     string `yaml:"web_url,omitempty" json:"web_url,omitempty"`           // URL to provide on the Web UI

	Defaults     *DashboardOptionsDefaults `yaml:"-" json:"-"` // Defaults
	HardDefaults *DashboardOptionsDefaults `yaml:"-" json:"-"` // Hard defaults
}

// NewDashboardOptions creates a new DashboardOptions.
func NewDashboardOptions(
	autoApprove *bool,
	icon string,
	iconLinkTo string,
	webURL string,
	defaults *DashboardOptionsDefaults,
	hardDefaults *DashboardOptionsDefaults,
) *DashboardOptions {
	return &DashboardOptions{
		DashboardOptionsBase: DashboardOptionsBase{
			AutoApprove: autoApprove},
		Icon:         icon,
		IconLinkTo:   iconLinkTo,
		WebURL:       webURL,
		Defaults:     defaults,
		HardDefaults: hardDefaults}
}

// GetAutoApprove will return whether new releases should be auto-approved.
func (d *DashboardOptions) GetAutoApprove() bool {
	return *util.GetFirstNonDefault(
		d.AutoApprove,
		d.Defaults.AutoApprove,
		d.HardDefaults.AutoApprove)
}

// CheckValues of the Dashboardoption.
func (d *DashboardOptions) CheckValues(prefix string) (errs error) {
	if d == nil {
		return
	}

	if !util.CheckTemplate(d.WebURL) {
		errs = fmt.Errorf("%s  web_url: %q <invalid> (didn't pass templating)\\",
			prefix, d.WebURL)
	}

	if errs != nil {
		errs = fmt.Errorf("%sdashboard:\\%s",
			prefix, util.ErrorToString(errs))
	}
	return
}
