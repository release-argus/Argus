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

//go:build unit

package service

import (
	"io"
	"os"
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
			dashboard.Defaults = &DashboardOptions{AutoApprove: tc.autoApproveDefault}
			dashboard.HardDefaults = &DashboardOptions{AutoApprove: tc.autoApproveHardDefault}

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

func TestDashboardOptions_Print(t *testing.T) {
	// GIVEN a Service
	tests := map[string]struct {
		dashboardOptions DashboardOptions
		want             string
	}{
		"default prints nothing": {
			dashboardOptions: DashboardOptions{},
			want:             "",
		},
		"print auto_approve": {
			dashboardOptions: DashboardOptions{
				AutoApprove: boolPtr(false)},
			want: `
dashboard:
  auto_approve: false
`,
		},
		"print icon": {
			dashboardOptions: DashboardOptions{
				Icon: "https://github.com/release-argus/Argus/raw/master/web/ui/static/favicon.svg"},
			want: `
dashboard:
  icon: "https://github.com/release-argus/Argus/raw/master/web/ui/static/favicon.svg"
`,
		},
		"print icon_link_to": {
			dashboardOptions: DashboardOptions{IconLinkTo: "https://release-argus.io/demo"},
			want: `
dashboard:
  icon_link_to: "https://release-argus.io/demo"
`,
		},
		"print web_url": {
			dashboardOptions: DashboardOptions{WebURL: "https://release-argus.io"},
			want: `
dashboard:
  web_url: "https://release-argus.io"
`,
		},
		"all options defined": {
			dashboardOptions: DashboardOptions{
				AutoApprove: boolPtr(false),
				WebURL:      "https://release-argus.io",
				Icon:        "https://github.com/release-argus/Argus/raw/master/web/ui/static/favicon.svg",
				IconLinkTo:  "https://release-argus.io/demo"},
			want: `
dashboard:
  auto_approve: false
  icon: "https://github.com/release-argus/Argus/raw/master/web/ui/static/favicon.svg"
  icon_link_to: "https://release-argus.io/demo"
  web_url: "https://release-argus.io"
`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN Print is called
			tc.dashboardOptions.Print("")

			// THEN it prints the expected number of lines
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = stdout
			got := string(out)
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
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
		"nil DashboardOptions": {
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
