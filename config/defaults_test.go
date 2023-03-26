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
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	opt "github.com/release-argus/Argus/service/options"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

func TestDefaults_SetDefaults(t *testing.T) {
	// GIVEN nil defaults
	var defaults Defaults

	// WHEN SetDefaults is called on it
	defaults.SetDefaults()
	tests := map[string]struct {
		got  string
		want string
	}{
		"Service.Interval": {
			got:  defaults.Service.Options.Interval,
			want: "10m"},
		"Notify.discord.username": {
			got:  defaults.Notify["discord"].GetSelfParam("username"),
			want: "Argus"},
		"WebHook.Delay": {
			got:  defaults.WebHook.Delay,
			want: "0s"},
	}

	// THEN the defaults are set correctly
	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.got != tc.want {
				t.Log(name)
				t.Errorf("want: %s\ngot:  %s",
					tc.want, tc.got)
			}
		})
	}
}

func TestDefaults_CheckValues(t *testing.T) {
	// GIVEN defaults with a test of invalid vars
	var defaults Defaults
	defaults.SetDefaults()
	tests := map[string]struct {
		input       Defaults
		errContains []string
	}{
		"Service.Interval": {
			input: Defaults{Service: service.Service{
				Options: opt.Options{
					Interval: "10x"}}},
			errContains: []string{
				`^  service:$`,
				`^      interval: "10x" <invalid>`},
		},
		"Service.DeployedVersionLookup.Regex": {
			input: Defaults{Service: service.Service{
				DeployedVersionLookup: &deployedver.Lookup{
					Regex: `^something[0-`}}},
			errContains: []string{
				`^  service:$`,
				`^    deployed_version:$`,
				`^      regex: "\^something\[0\-" <invalid>`},
		},
		"Service.Interval + Service.DeployedVersionLookup.Regex": {
			input: Defaults{Service: service.Service{
				Options: opt.Options{
					Interval: "10x"},
				DeployedVersionLookup: &deployedver.Lookup{
					Regex: `^something[0-`}}},
			errContains: []string{
				`^  service:$`,
				`^    deployed_version:$`,
				`^      regex: "\^something\[0\-" <invalid>`},
		},
		"Notify.x.Delay": {
			input: Defaults{Notify: shoutrrr.Slice{
				"slack": &shoutrrr.Shoutrrr{
					Options: map[string]string{"delay": "10x"}}}},
			errContains: []string{
				`^  notify:$`,
				`^    slack:$`,
				`^      options:`,
				`^        delay: "10x" <invalid>`},
		},
		"WebHook.Delay": {
			input: Defaults{WebHook: webhook.WebHook{
				Delay: "10x"}},
			errContains: []string{
				`^  webhook:$`,
				`^  delay: "10x" <invalid>`}},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// WHEN CheckValues is called on it
			err := util.ErrorToString(tc.input.CheckValues())

			// THEN err matches expected
			lines := strings.Split(err, "\\")
			for i := range tc.errContains {
				re := regexp.MustCompile(tc.errContains[i])
				found := false
				for j := range lines {
					match := re.MatchString(lines[j])
					if match {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("invalid %s should have errored:\nwant: %q\ngot:  %q",
						name, tc.errContains[i], strings.ReplaceAll(err, `\`, "\n"))
				}
			}
		})
	}
}

func TestDefaults_Print(t *testing.T) {
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
	out, _ := io.ReadAll(r)
	os.Stdout = stdout
	want := 129
	got := strings.Count(string(out), "\n")
	if got != want {
		t.Errorf("Print should have given %d lines, but gave %d\n%s", want, got, out)
	}
}
