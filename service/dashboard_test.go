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

//go:build unit

package service

import (
	"regexp"
	"strings"
	"testing"

	"github.com/release-argus/Argus/util"
)

func TestDashboardOptions_GetAutoApprove(t *testing.T) {
	// GIVEN a DashboardOptions
	tests := map[string]struct {
		autoApproveRoot        *bool
		autoApproveDefault     *bool
		autoApproveHardDefault *bool
		wantBool               bool
	}{
		"root overrides all": {
			wantBool:               true,
			autoApproveRoot:        boolPtr(true),
			autoApproveDefault:     boolPtr(false),
			autoApproveHardDefault: boolPtr(false)},
		"default overrides hardDefault": {
			wantBool:               true,
			autoApproveRoot:        nil,
			autoApproveDefault:     boolPtr(true),
			autoApproveHardDefault: boolPtr(false)},
		"hardDefault is last resort": {
			wantBool:               true,
			autoApproveRoot:        nil,
			autoApproveDefault:     nil,
			autoApproveHardDefault: boolPtr(true)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			dashboard := DashboardOptions{}
			dashboard.AutoApprove = tc.autoApproveRoot
			defaults := NewDashboardOptionsDefaults(tc.autoApproveDefault)
			dashboard.Defaults = &defaults
			hardDefaults := NewDashboardOptionsDefaults(tc.autoApproveHardDefault)
			dashboard.HardDefaults = &hardDefaults

			// WHEN GetAutoApprove is called
			got := dashboard.GetAutoApprove()

			// THEN the function returns the correct result
			if got != tc.wantBool {
				t.Errorf("want: %t\ngot:  %t",
					tc.wantBool, got)
			}
		})
	}
}

func TestDashboardOptions_CheckValues(t *testing.T) {
	// GIVEN DashboardOptions
	jLog = util.NewJLog("WARN", false)
	tests := map[string]struct {
		dashboardOptions *DashboardOptions
		errRegex         []string
	}{
		"nil": {
			errRegex:         []string{"^$"},
			dashboardOptions: nil},
		"invalid web_url template": {
			errRegex:         []string{"^-dashboard:$", "^-  web_url: .* <invalid>"},
			dashboardOptions: &DashboardOptions{WebURL: "https://release-argus.io/{{ version }"}},
		"valid web_url template": {
			errRegex:         []string{"^$"},
			dashboardOptions: &DashboardOptions{WebURL: "https://release-argus.io"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// WHEN CheckValues is called on it
			err := tc.dashboardOptions.CheckValues("-")

			// THEN the err is what we expect
			e := util.ErrorToString(err)
			lines := strings.Split(e, `\`)
			for i := range tc.errRegex {
				re := regexp.MustCompile(tc.errRegex[i])
				found := false
				for j := range lines {
					match := re.MatchString(lines[j])
					if match {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("want match for: %q\ngot:  %q",
						tc.errRegex[i], strings.ReplaceAll(e, `\`, "\n"))
				}
			}
		})
	}
}
