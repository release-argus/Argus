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

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service/deployed_version"
)

func TestServiceGetIconURL(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		icon   *string
		want   string
		notify shoutrrr.Slice
	}{
		"nil icon":   {want: "", icon: nil},
		"emoji icon": {want: ":smile:", icon: stringPtr(":smile:")},
		"web icon":   {want: "https://example.com/icon.png", icon: stringPtr("https://example.com/icon.png")},
		"notify icon only": {want: "https://example.com/icon.png", notify: shoutrrr.Slice{"test": &shoutrrr.Shoutrrr{
			Params: map[string]string{
				"icon": "https://example.com/icon.png",
			},
			Main:         &shoutrrr.Shoutrrr{},
			Defaults:     &shoutrrr.Shoutrrr{},
			HardDefaults: &shoutrrr.Shoutrrr{},
		}}},
		"no icon anywhere": {want: "https://example.com/icon.png", notify: shoutrrr.Slice{"test": &shoutrrr.Shoutrrr{
			Main:         &shoutrrr.Shoutrrr{},
			Defaults:     &shoutrrr.Shoutrrr{},
			HardDefaults: &shoutrrr.Shoutrrr{},
		}}},
		"notify icon overriden by icon": {want: ":smile:", icon: stringPtr(":smile:"),
			notify: shoutrrr.Slice{"test": &shoutrrr.Shoutrrr{
				Params: map[string]string{
					"icon": "https://example.com/icon.png",
				},
				Main:         &shoutrrr.Shoutrrr{},
				Defaults:     &shoutrrr.Shoutrrr{},
				HardDefaults: &shoutrrr.Shoutrrr{},
			}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			service := testServiceGitHub()
			service.Icon = tc.icon
			service.Notify = tc.notify

			// WHEN GetIconURL is called
			got := service.GetIconURL()

			// THEN the function returns the correct result
			if got != tc.want {
				t.Errorf("%s:\nwant: %q\ngot:  %q",
					name, tc.want, got)
			}
		})
	}
}

func TestServiceInitHandsOutDefaults(t *testing.T) {
	// GIVEN a Service with nil Defaults
	service := testServiceGitHub()
	service.Defaults = nil
	defaults := Service{
		ID: "test",
	}

	// WHEN Init is called on it
	service.Init(nil, &defaults, &Service{})

	// THEN ApprovedVersion is reset
	got := service.Defaults
	if got != &defaults {
		t.Errorf("Service should've been given %v Defaults, not %v",
			defaults, got)
	}
}

func TestServiceInitHandsOutHardDefaults(t *testing.T) {
	// GIVEN a Service with nil HardDefaults
	service := testServiceGitHub()
	service.HardDefaults = nil
	defaults := Service{ID: "test"}
	// WHEN Init is called on it
	service.Init(nil, &Service{}, &defaults)

	// THEN ApprovedVersion is reset
	got := service.HardDefaults
	if got != &defaults {
		t.Errorf("Service should've been given %v HardDefaults, not %v",
			defaults, got)
	}
}

func TestServiceInitWithDeployedVersionLookupHandsOutDefaults(t *testing.T) {
	// GIVEN a Service with DeployedVersionLookup
	service := testServiceGitHub()
	service.DeployedVersionLookup = &deployed_version.Lookup{}
	defaults := deployed_version.Lookup{Regex: "test"}

	// WHEN Init is called on it
	service.Init(nil, &Service{DeployedVersionLookup: &defaults}, &Service{DeployedVersionLookup: &deployed_version.Lookup{}})

	// THEN ApprovedVersion is reset
	got := service.DeployedVersionLookup.Defaults
	if *got != &defaults {
		t.Errorf("DeployedVersionLookup should've been given %v Defaults, not %v",
			defaults, got)
	}
}

func TestServiceInitWithDeployedVersionLookupHandsOutHardDefaults(t *testing.T) {
	// GIVEN a Service with DeployedVersionLookup
	service := testServiceGitHub()
	service.DeployedVersionLookup = &deployed_version.Lookup{}
	defaults := deployed_version.Lookup{
		Regex: "test",
	}

	// WHEN Init is called on it
	service.Init(nil, &Service{DeployedVersionLookup: &deployed_version.Lookup{}}, &Service{DeployedVersionLookup: &defaults})

	// THEN ApprovedVersion is reset
	got := service.DeployedVersionLookup.HardDefaults
	if *got != &defaults {
		t.Errorf("DeployedVersionLookup should've been given %v Defaults, not %v",
			defaults, got)
	}
}
