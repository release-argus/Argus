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

package v1

import (
	"testing"
	"time"

	command "github.com/release-argus/Argus/commands"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	api_type "github.com/release-argus/Argus/web/api/types"
)

func TestConvertAndCensorDeployedVersionLookup(t *testing.T) {
	// GIVEN a DeployedVersionLookup
	tests := map[string]struct {
		dvl                      *deployedver.Lookup
		dvlStatus                *svcstatus.Status
		approvedVersion          string
		deployedVersion          string
		deployedVersionTimestamp string
		latestVersion            string
		latestVersionTimestamp   string
		lastQueried              string
		regexMissesContent       int
		regexMissesVersion       int

		want *api_type.DeployedVersionLookup
	}{
		"nil": {
			dvl:  nil,
			want: nil,
		},
		"empty": {
			dvl:  &deployedver.Lookup{},
			want: &api_type.DeployedVersionLookup{},
		},
		"minimal": {
			dvl: &deployedver.Lookup{
				URL:  "https://example.com",
				JSON: "version"},
			want: &api_type.DeployedVersionLookup{
				URL:  "https://example.com",
				JSON: "version"},
		},
		"censor basic_auth.password": {
			dvl: &deployedver.Lookup{
				URL: "https://example.com",
				BasicAuth: &deployedver.BasicAuth{
					Username: "alan",
					Password: "pass123"}},
			want: &api_type.DeployedVersionLookup{
				URL: "https://example.com",
				BasicAuth: &api_type.BasicAuth{
					Username: "alan",
					Password: "<secret>"}},
		},
		"censor headers": {
			dvl: &deployedver.Lookup{
				URL: "https://example.com",
				Headers: []deployedver.Header{
					{Key: "X-Test-0", Value: "foo"},
					{Key: "X-Test-1", Value: "bar"}}},
			want: &api_type.DeployedVersionLookup{
				URL: "https://example.com",
				Headers: []api_type.Header{
					{Key: "X-Test-0", Value: "<secret>"},
					{Key: "X-Test-1", Value: "<secret>"},
				}},
		},
		"full": {
			regexMissesContent: 1,
			regexMissesVersion: 3,
			dvl: deployedver.New(
				boolPtr(true),
				&deployedver.BasicAuth{
					Username: "jim",
					Password: "whoops"},
				&[]deployedver.Header{
					{Key: "X-Test-0", Value: "foo"},
					{Key: "X-Test-1", Value: "bar"}},
				"version",
				opt.New(
					boolPtr(true), "10m", boolPtr(true),
					&opt.OptionsDefaults{},
					&opt.OptionsDefaults{}),
				`([0-9]+\.[0-9]+\.[0-9]+)`,
				&svcstatus.Status{},
				"https://release-argus.io",
				&deployedver.LookupDefaults{},
				&deployedver.LookupDefaults{}),
			dvlStatus: &svcstatus.Status{
				Fails:     svcstatus.Fails{},
				ServiceID: stringPtr("service-id"),
				WebURL:    stringPtr("https://release-argus.io")},
			want: &api_type.DeployedVersionLookup{
				URL:               "https://release-argus.io",
				AllowInvalidCerts: boolPtr(true),
				BasicAuth: &api_type.BasicAuth{
					Username: "jim",
					Password: "<secret>"},
				Headers: []api_type.Header{
					{Key: "X-Test-0", Value: "<secret>"},
					{Key: "X-Test-1", Value: "<secret>"}},
				JSON:  "version",
				Regex: `([0-9]+\.[0-9]+\.[0-9]+)`},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.dvl != nil {
				if tc.approvedVersion != "" {
					tc.dvl.Status.SetApprovedVersion("1.2.3", false)
					tc.dvl.Status.SetDeployedVersion("1.2.3", false)
					tc.dvl.Status.SetDeployedVersionTimestamp(time.Now().Format(time.RFC3339))
					tc.dvl.Status.SetLatestVersion("1.2.3", false)
					tc.dvl.Status.SetLatestVersionTimestamp(time.Now().Format(time.RFC3339))
					tc.dvl.Status.SetLastQueried(time.Now().Format(time.RFC3339))
				}
				tc.dvl.Status = tc.dvlStatus
				for i := 0; i < tc.regexMissesContent; i++ {
					tc.dvl.Status.RegexMissContent()
				}
				for i := 0; i < tc.regexMissesVersion; i++ {
					tc.dvl.Status.RegexMissVersion()
				}
			}

			// WHEN convertAndCensorDeployedVersionLookup is called on it
			got := convertAndCensorDeployedVersionLookup(tc.dvl)

			// THEN the WebHookSlice is converted correctly
			if got.String() != tc.want.String() {
				t.Errorf("want:\n%q\ngot:\n%q",
					tc.want.String(), got.String())
			}
		})
	}
}

func TestConvertURLCommandSlice(t *testing.T) {
	// GIVEN a URL Command slice
	tests := map[string]struct {
		slice *filter.URLCommandSlice
		want  *api_type.URLCommandSlice
	}{
		"nil": {
			slice: nil,
			want:  nil,
		},
		"empty": {
			slice: &filter.URLCommandSlice{},
			want:  &api_type.URLCommandSlice{},
		},
		"regex": {
			slice: &filter.URLCommandSlice{
				{Type: "regex", Regex: stringPtr("[0-9.]+")}},
			want: &api_type.URLCommandSlice{
				{Type: "regex", Regex: stringPtr("[0-9.]+")}},
		},
		"replace": {
			slice: &filter.URLCommandSlice{
				{Type: "replace", Old: stringPtr("foo"), New: stringPtr("bar")}},
			want: &api_type.URLCommandSlice{
				{Type: "replace", Old: stringPtr("foo"), New: stringPtr("bar")}},
		},
		"split": {
			slice: &filter.URLCommandSlice{
				{Type: "split", Index: 7}},
			want: &api_type.URLCommandSlice{
				{Type: "split", Index: 7}},
		},
		"one of each": {
			slice: &filter.URLCommandSlice{
				{Type: "regex", Regex: stringPtr("[0-9.]+")},
				{Type: "replace", Old: stringPtr("foo"), New: stringPtr("bar")},
				{Type: "split", Index: 7}},
			want: &api_type.URLCommandSlice{
				{Type: "regex", Regex: stringPtr("[0-9.]+")},
				{Type: "replace", Old: stringPtr("foo"), New: stringPtr("bar")},
				{Type: "split", Index: 7}},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertURLCommandSlice is called on it
			got := convertURLCommandSlice(tc.slice)

			// THEN the WebHookSlice is converted correctly
			if got.String() != tc.want.String() {
				t.Errorf("want: %q, got: %q",
					tc.want.String(), got.String())
			}
		})
	}
}

func TestConvertCommandSlice(t *testing.T) {
	// GIVEN a Command slice
	tests := map[string]struct {
		slice *command.Slice
		want  *api_type.CommandSlice
	}{
		"nil": {
			slice: nil,
			want:  nil,
		},
		"empty": {
			slice: &command.Slice{},
			want:  &api_type.CommandSlice{},
		},
		"one": {
			slice: &command.Slice{
				{"ls", "-lah"}},
			want: &api_type.CommandSlice{
				{"ls", "-lah"}},
		},
		"two": {
			slice: &command.Slice{
				{"ls", "-lah"},
				{"/bin/bash", "something.sh"}},
			want: &api_type.CommandSlice{
				{"ls", "-lah"},
				{"/bin/bash", "something.sh"}},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertCommandSlice is called on it
			got := convertCommandSlice(tc.slice)

			// THEN the CommandSlice is converted correctly
			if got == tc.want { // both nil
				return
			}
			// check number of commands
			if len(*got) != len(*tc.want) {
				t.Errorf("want:\n%v\ngot:\n%v",
					tc.want, got)
				return
			}
			for cI := range *got {
				// check number of args
				if len((*got)[cI]) != len((*tc.want)[cI]) {
					t.Errorf("want:\n%v\ngot:\n%v",
						tc.want, got)
				}
				// check args
				for aI := range (*got)[cI] {
					if (*got)[cI][aI] != (*tc.want)[cI][aI] {
						t.Errorf("want:\n%v\ngot:\n%v",
							tc.want, got)
					}
				}
			}
		})
	}
}
