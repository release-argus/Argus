// Copyright [2022] [Argus]
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

	"github.com/release-argus/Argus/utils"
)

type DashboardOptions struct {
	Icon       string            `yaml:"icon,omitempty"`         // Icon URL to use for messages/Web UI
	IconLinkTo string            `yaml:"icon_link_to,omitempty"` // URL to redirect Icon clicks to
	WebURL     string            `yaml:"web_url,omitempty"`      // URL to provide on the Web UI
	Defaults   *DashboardOptions `yaml:"-"`                      // Defaults
}

// GetWebURL returns the Web URL.
func (d *DashboardOptions) GetWebURL(latestVersion string) string {
	if d.WebURL == "" {
		return ""
	}

	return utils.TemplateString(d.WebURL, utils.ServiceInfo{LatestVersion: latestVersion})
}

// Print the struct.
func (d *DashboardOptions) Print(prefix string) {
	fmt.Printf("%sdashboard:\n", prefix)
	utils.PrintlnIfNotDefault(d.WebURL, fmt.Sprintf("%sweb_url: %q", prefix, d.WebURL))
	utils.PrintlnIfNotDefault(d.Icon, fmt.Sprintf("%sicon: %q", prefix, d.Icon))
	utils.PrintlnIfNotDefault(d.IconLinkTo, fmt.Sprintf("%sicon_link_to: %q", prefix, d.IconLinkTo))
}
