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
	"testing"
)

func TestGetAutoApprove(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		autoApproveRoot        *bool
		autoApproveDefault     *bool
		autoApproveHardDefault *bool
		wantBool               bool
	}{
		"root overrides all": {wantBool: true, autoApproveRoot: boolPtr(true),
			autoApproveDefault: boolPtr(false), autoApproveHardDefault: boolPtr(false)},
		"default overrides hardDefault": {wantBool: true, autoApproveRoot: nil,
			autoApproveDefault: boolPtr(false), autoApproveHardDefault: boolPtr(false)},
		"hardDefault is last resort": {wantBool: true, autoApproveRoot: nil, autoApproveDefault: nil,
			autoApproveHardDefault: boolPtr(true)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dashboard := DashboardOptions{}
			dashboard.AutoApprove = tc.autoApproveRoot
			dashboard.Defaults.AutoApprove = tc.autoApproveDefault
			dashboard.HardDefaults.AutoApprove = tc.autoApproveHardDefault

			// WHEN GetAutoApprove is called
			got := dashboard.GetAutoApprove()

			// THEN the function returns the correct result
			if got != tc.wantBool {
				t.Errorf("%s:\nwant: %t\ngot:  %t",
					name, tc.wantBool, got)
			}
		})
	}
}
