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

//go:build unit

package service

import (
	"strings"
	"testing"

	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestDashboardOptions_GetAutoApprove(t *testing.T) {
	// GIVEN a DashboardOptions
	tests := map[string]struct {
		rootValue, defaultValue, hardDefaultValue *bool
		want                                      bool
	}{
		"root overrides all": {
			want:             true,
			rootValue:        test.BoolPtr(true),
			defaultValue:     test.BoolPtr(false),
			hardDefaultValue: test.BoolPtr(false)},
		"default overrides hardDefault": {
			want:             true,
			defaultValue:     test.BoolPtr(true),
			hardDefaultValue: test.BoolPtr(false)},
		"hardDefault is last resort": {
			want:             true,
			hardDefaultValue: test.BoolPtr(true)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			dashboard := DashboardOptions{}
			dashboard.AutoApprove = tc.rootValue
			defaults := NewDashboardOptionsDefaults(tc.defaultValue)
			dashboard.Defaults = &defaults
			hardDefaultValues := NewDashboardOptionsDefaults(tc.hardDefaultValue)
			dashboard.HardDefaults = &hardDefaultValues

			// WHEN GetAutoApprove is called
			got := dashboard.GetAutoApprove()

			// THEN the function returns the correct result
			if got != tc.want {
				t.Errorf("want: %t\ngot:  %t",
					tc.want, got)
			}
		})
	}
}

func TestDashboardOptions_CheckValues(t *testing.T) {
	// GIVEN DashboardOptions
	tests := map[string]struct {
		dashboardOptions *DashboardOptions
		errRegex         string
	}{
		"nil": {
			errRegex:         `^$`,
			dashboardOptions: nil},
		"invalid web_url template": {
			errRegex:         `^web_url: ".*" <invalid>.*$`,
			dashboardOptions: &DashboardOptions{WebURL: "https://release-argus.io/{{ version }"}},
		"valid web_url template": {
			errRegex:         `^$`,
			dashboardOptions: &DashboardOptions{WebURL: "https://release-argus.io"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// WHEN CheckValues is called on it
			err := tc.dashboardOptions.CheckValues("")

			// THEN the err is what we expect
			e := util.ErrorToString(err)
			lines := strings.Split(e, "\n")
			wantLines := strings.Count(tc.errRegex, "\n")
			if wantLines > len(lines) {
				t.Fatalf("DashboardOptions.CheckValues() want %d lines of error:\n%q\ngot %d lines:\n%v\nstdout: %q",
					wantLines, tc.errRegex, len(lines), lines, e)
				return
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("DashboardOptions.CheckValues() error mismatch\nwant match for:\n%q\ngot:\n%q",
					tc.errRegex, e)
				return
			}
		})
	}
}
