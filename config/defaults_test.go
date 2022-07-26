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

package config

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/deployed_version"
	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/webhook"
)

func TestDefaultsSetDefaults(t *testing.T) {
	// GIVEN nil defaults
	var defaults Defaults

	// WHEN SetDefaults is called on it
	defaults.SetDefaults()
	tests := map[string]struct {
		got  *string
		want string
	}{
		"Service.Interval": {
			got: defaults.Service.Interval, want: "10m"},
		"Notify.discord.username": {
			got: stringPtr(defaults.Notify["discord"].GetSelfParam("username")), want: "Argus"},
		"WebHook.Delay": {
			got: defaults.WebHook.Delay, want: "0s"},
	}

	// THEN the defaults are set correctly
	for name, tc := range tests {
		if utils.EvalNilPtr(tc.got, "") != tc.want {
			t.Errorf("%s:\nwant: %s\ngot:  %s",
				name, tc.want, utils.EvalNilPtr(tc.got, ""))
		}
	}
}

func TestDefaultsCheckValues(t *testing.T) {
	// GIVEN defaults with a test of invalid vars
	var defaults Defaults
	defaults.SetDefaults()
	tests := map[string]struct {
		input       Defaults
		errContains string
	}{
		"Service.Interval": {
			input: Defaults{Service: service.Service{
				Interval: stringPtr("10x")}},
			errContains: `interval: "10x" <invalid>`},
		"Service.DeployedVersionLookup.Regex": {
			input: Defaults{Service: service.Service{
				DeployedVersionLookup: &deployed_version.Lookup{
					Regex: "^something[0-"}}},
			errContains: `regex: "^something[0-" <invalid>`},
		"Notify.x.Delay": {
			input: Defaults{Notify: shoutrrr.Slice{
				"slack": &shoutrrr.Shoutrrr{
					Options: map[string]string{"delay": "10x"}}}},
			errContains: `delay: "10x" <invalid>`},
		"WebHook.x.Delay": {
			input: Defaults{WebHook: webhook.WebHook{
				Delay: stringPtr("10x")}},
			errContains: `delay: "10x" <invalid>`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// WHEN CheckValues is called on it
			err := utils.ErrorToString(tc.input.CheckValues())

			// THEN err matches expected
			if !strings.Contains(err, tc.errContains) {
				t.Errorf("invalid %s should have errored:\nwant: %s\ngot:  %s",
					name, tc.errContains, err)
			}
		})
	}
}

func TestDefaultsPrint(t *testing.T) {
	// GIVEN unmodified defaults from SetDefaults
	var defaults Defaults
	defaults.SetDefaults()
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Print is called
	defaults.Print()

	// THEN the expected number of lines are printed
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	want := 142
	got := strings.Count(string(out), "\n")
	if got != want {
		t.Errorf("Print should have given %d lines, but gave %d\n%s", want, got, out)
	}
}
