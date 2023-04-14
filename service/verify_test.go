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
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	opt "github.com/release-argus/Argus/service/options"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

func TestService_Print(t *testing.T) {
	// GIVEN a Service
	tests := map[string]struct {
		svc              *Service
		options          opt.Options
		latestVersion    latestver.Lookup
		deployedVersion  *deployedver.Lookup
		commands         command.Slice
		webhooks         webhook.Slice
		notifies         shoutrrr.Slice
		dashboardOptions DashboardOptions
		lines            int
	}{
		"base fields only": {
			lines: 2,
			svc: &Service{ID: "test",
				Comment: "foo_comment"},
		},
		"base + latest_version": {
			lines: 4,
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			latestVersion: latestver.Lookup{
				Type: "github", URL: "release-argus/Argus"},
		},
		"base + latest_version + deployed_version": {
			lines: 6,
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			latestVersion: latestver.Lookup{
				Type: "github", URL: "release-argus/Argus"},
			deployedVersion: &deployedver.Lookup{
				URL: "https://release-argus.io/demo/api/v1/version"},
		},
		"base + latest_version + deployed_version + notifies": {
			lines: 9,
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			latestVersion: latestver.Lookup{
				Type: "github", URL: "release-argus/Argus"},
			deployedVersion: &deployedver.Lookup{
				URL: "https://release-argus.io/demo/api/v1/version"},
			notifies: shoutrrr.Slice{
				"foo": {Type: "discord"}},
		},
		"base + latest_version + deployed_version + notifies + commands": {
			lines: 11,
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			latestVersion: latestver.Lookup{
				Type: "github", URL: "release-argus/Argus"},
			deployedVersion: &deployedver.Lookup{
				URL: "https://release-argus.io/demo/api/v1/version"},
			notifies: shoutrrr.Slice{
				"foo": {Type: "discord"}},
			commands: command.Slice{
				{"ls", "-la"}},
		},
		"base + latest_version + deployed_version + notifies + commands + webhooks": {
			lines: 14,
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			latestVersion: latestver.Lookup{
				Type: "github", URL: "release-argus/Argus"},
			deployedVersion: &deployedver.Lookup{
				URL: "https://release-argus.io/demo/api/v1/version"},
			notifies: shoutrrr.Slice{
				"foo": &shoutrrr.Shoutrrr{Type: "discord"}},
			commands: command.Slice{
				{"ls", "-la"}},
			webhooks: webhook.Slice{
				"bar": &webhook.WebHook{URL: "https://example.com"}},
		},
		"base + latest_version + deployed_version + notifies + commands + webhooks + dashboard": {
			lines: 16,
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			latestVersion: latestver.Lookup{
				Type: "github", URL: "release-argus/Argus"},
			deployedVersion: &deployedver.Lookup{
				URL: "https://release-argus.io/demo/api/v1/version"},
			notifies: shoutrrr.Slice{
				"foo": &shoutrrr.Shoutrrr{Type: "discord"}},
			commands: command.Slice{
				{"ls", "-la"}},
			webhooks: webhook.Slice{
				"bar": &webhook.WebHook{URL: "https://example.com"}},
			dashboardOptions: DashboardOptions{
				Icon: "https://example.com/icon.png"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			tc.svc.LatestVersion = tc.latestVersion
			tc.svc.DeployedVersionLookup = tc.deployedVersion
			tc.svc.Command = tc.commands
			tc.svc.WebHook = tc.webhooks
			tc.svc.Notify = tc.notifies
			tc.svc.Dashboard = tc.dashboardOptions

			// WHEN Print is called
			tc.svc.Print("")

			// THEN it prints the expected number of lines
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = stdout
			got := strings.Count(string(out), "\n")
			if got != tc.lines {
				t.Errorf("Print should have given %d lines, but gave %d\n%s",
					tc.lines, got, out)
			}
		})
	}
}

func TestSlice_Print(t *testing.T) {
	// GIVEN a Slice
	tests := map[string]struct {
		slice      *Slice
		ordering   []string
		lines      int
		regexMatch string
	}{
		"nil slice with no ordering": {
			lines: 0,
			slice: nil,
		},
		"nil slice with ordering": {
			lines:    0,
			ordering: []string{"foo", "bar"},
			slice:    nil,
		},
		"respects ordering": {
			lines:    7,
			ordering: []string{"zulu", "alpha"},
			slice: &Slice{
				"zulu":  &Service{ID: "zulu", Comment: "a"},
				"alpha": &Service{ID: "alpha", Comment: "b"}},
			regexMatch: `zulu(.|\s)+alpha`,
		},
		"respects reversedordering": {
			lines:    7,
			ordering: []string{"alpha", "zulu"},
			slice: &Slice{
				"zulu":  &Service{ID: "zulu", Comment: "a"},
				"alpha": &Service{ID: "alpha", Comment: "b"}},
			regexMatch: `alpha(.|\s)+zulu`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN Print is called
			tc.slice.Print("", tc.ordering)

			// THEN it prints the expected number of lines
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = stdout
			got := strings.Count(string(out), "\n")
			if got != tc.lines {
				t.Errorf("Print should have given %d lines, but gave %d\n%s",
					tc.lines, got, out)
			}
			// in the right order
			re := regexp.MustCompile(tc.regexMatch)
			match := re.MatchString(string(out))
			if !match {
				t.Fatalf("want match for %q\nnot: %q",
					tc.regexMatch, string(out))
			}
		})
	}
}

func TestService_CheckValues(t *testing.T) {
	// GIVEN a Service
	tests := map[string]struct {
		svc              *Service
		defaults         *Service
		options          opt.Options
		latestVersion    latestver.Lookup
		deployedVersion  *deployedver.Lookup
		commands         command.Slice
		webhooks         webhook.Slice
		notifies         shoutrrr.Slice
		dashboardOptions DashboardOptions
		errRegex         []string
	}{
		"options with errs": {
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			latestVersion: latestver.Lookup{
				Type: "github", URL: "release-argus/Argus"},
			options: opt.Options{
				Interval: "10x"},
			errRegex: []string{
				`^  options:$`,
				`^    interval: .* <invalid>`},
		},
		"options,latest_version, with errs": {
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			options: opt.Options{
				Interval: "10x"},
			latestVersion: latestver.Lookup{
				Type: "invalid", URL: "release-argus/Argus"},
			errRegex: []string{
				`^  options:$`, `^    interval: .* <invalid>`,
				`^  latest_version:$`,
				`^    type: .* <invalid>`},
		},
		"latest_version, deployed_version with errs": {
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			options: opt.Options{
				Interval: "10x"},
			latestVersion: latestver.Lookup{
				Type: "invalid", URL: "release-argus/Argus"},
			deployedVersion: &deployedver.Lookup{
				Regex: "[0-"},
			errRegex: []string{
				`^  options:$`,
				`^    interval: .* <invalid>`,
				`^  latest_version:$`,
				`^    type: .* <invalid>`,
				`^  deployed_version:$`,
				`^    regex: .* <invalid>`},
		},
		"latest_version, deployed_version, command with errs": {
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			options: opt.Options{
				Interval: "10x"},
			latestVersion: latestver.Lookup{
				Type: "invalid", URL: "release-argus/Argus"},
			deployedVersion: &deployedver.Lookup{
				Regex: "[0-"},
			errRegex: []string{
				`^  options:$`,
				`^    interval: .* <invalid>`,
				`^  latest_version:$`,
				`^    type: .* <invalid>`,
				`^  deployed_version:$`,
				`^    regex: .* <invalid>`,
				`^  command:$`,
				`^    item_0: bash .* <invalid>.*templating`},
			commands: command.Slice{{"bash", "update.sh", "{{ version }"}},
		},
		"latest_version, deployed_version, notify with errs": {
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			options: opt.Options{
				Interval: "10x"},
			latestVersion: latestver.Lookup{
				Type: "invalid", URL: "release-argus/Argus"},
			deployedVersion: &deployedver.Lookup{
				Regex: "[0-"},
			commands: command.Slice{{
				"bash", "update.sh", "{{ version }"}},
			notifies: shoutrrr.Slice{
				"foo": &shoutrrr.Shoutrrr{Type: "discord"}}, errRegex: []string{
				`^  options:$`, `^    interval: .* <invalid>`,
				`^  latest_version:$`,
				`^    type: .* <invalid>`,
				`^  deployed_version:$`,
				`^    regex: .* <invalid>`,
				`^  command:$`,
				`^    item_0: bash .* <invalid>.*templating`,
				`^  notify:`,
				`^    foo:$`,
				`^      url_fields:$`,
				`^        token: <required>`,
				`^        webhookid: <required>`},
		},
		"latest_version, deployed_version, webhook with errs": {
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			options: opt.Options{
				Interval: "10x"},
			latestVersion: latestver.Lookup{
				Type: "invalid", URL: "release-argus/Argus"},
			deployedVersion: &deployedver.Lookup{
				Regex: "[0-"},
			commands: command.Slice{{
				"bash", "update.sh", "{{ version }"}},
			notifies: shoutrrr.Slice{
				"foo": &shoutrrr.Shoutrrr{Type: "discord"}},
			webhooks: webhook.Slice{
				"wh": &webhook.WebHook{Delay: "0x"}}, errRegex: []string{
				`^  options:$`,
				`^    interval: .* <invalid>`,
				`^  latest_version:$`,
				`^    type: .* <invalid>`,
				`^  deployed_version:$`,
				`^    regex: .* <invalid>`,
				`^  command:$`,
				`^    item_0: bash .* <invalid>.*templating`,
				`^  notify:`,
				`^    foo:$`,
				`^      url_fields:$`,
				`^        token: <required>`,
				`^        webhookid: <required>`,
				`^  webhook:$`,
				`^    wh:$`,
				`^      delay: .* <invalid>`},
		},
		"has defaults. latest_version, deployed_version, webhook with errs": {
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			defaults: &Service{},
			options: opt.Options{
				Interval: "10x"},
			latestVersion: latestver.Lookup{
				Type: "invalid", URL: "release-argus/Argus"},
			deployedVersion: &deployedver.Lookup{
				Regex: "[0-"},
			commands: command.Slice{
				{"bash", "update.sh", "{{ version }"}},
			notifies: shoutrrr.Slice{
				"foo": &shoutrrr.Shoutrrr{Type: "discord"}},
			webhooks: webhook.Slice{
				"wh": &webhook.WebHook{Delay: "0x"}},
			errRegex: []string{
				`^  test:$`,
				`^  options:$`,
				`^    interval: .* <invalid>`,
				`^  latest_version:$`,
				`^    type: .* <invalid>`,
				`^  deployed_version:$`,
				`^    regex: .* <invalid>`,
				`^  command:$`,
				`^    item_0: bash .* <invalid>.*templating`,
				`^  notify:`,
				`^    foo:$`,
				`^      url_fields:$`, `^        token: <required>`,
				`^        webhookid: <required>`,
				`^  webhook:$`,
				`^    wh:$`, `^      delay: .* <invalid>`},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			tc.svc.ID = "test"
			tc.svc.Defaults = tc.defaults
			tc.svc.Options = tc.options
			tc.svc.LatestVersion = tc.latestVersion
			tc.svc.DeployedVersionLookup = tc.deployedVersion
			tc.svc.Command = tc.commands
			tc.svc.WebHook = tc.webhooks
			tc.svc.Notify = tc.notifies
			for i := range tc.notifies {
				tc.notifies[i].Main = &shoutrrr.Shoutrrr{}
				tc.notifies[i].Defaults = &shoutrrr.Shoutrrr{}
				tc.notifies[i].HardDefaults = &shoutrrr.Shoutrrr{}
			}
			tc.svc.Dashboard = tc.dashboardOptions

			// WHEN CheckValues is called
			err := tc.svc.CheckValues("")

			// THEN it err's when expected
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

func TestSlice_CheckValues(t *testing.T) {
	// GIVEN a Slice
	tests := map[string]struct {
		slice         Slice
		errRegex      string
		errRegexOther string
	}{
		"single valid service": {
			slice: Slice{
				"first": {
					ID:      "test",
					Comment: "foo_comment",
					Options: opt.Options{Interval: "10s"},
					LatestVersion: latestver.Lookup{
						Type: "github", URL: "release-argus/Argus"}}},
			errRegex: `^$`,
		},
		"single invalid service": {
			slice: Slice{
				"first": {
					ID:      "test",
					Comment: "foo_comment",
					Options: opt.Options{Interval: "10x"},
					LatestVersion: latestver.Lookup{
						Type: "github", URL: "release-argus/Argus"}}},
			errRegex: `interval: .* <invalid>`,
		},
		"multiple invalid services": {
			slice: Slice{
				"foo": {
					ID:      "test",
					Comment: "foo_comment",
					Options: opt.Options{Interval: "10x"},
					LatestVersion: latestver.Lookup{
						Type: "github", URL: "release-argus/Argus"}},
				"bar": {
					ID:      "test",
					Comment: "foo_comment",
					Options: opt.Options{Interval: "10y"},
					LatestVersion: latestver.Lookup{
						Type: "github", URL: "release-argus/Argus"}}},
			errRegex:      `interval: .*10x.* <invalid>.*interval: .*10y.* <invalid>`,
			errRegexOther: "interval: .*10y.* <invalid>.*interval: .*10x.* <invalid>",
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CheckValues is called
			err := tc.slice.CheckValues("")

			// THEN it err's when expected
			e := util.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				if tc.errRegexOther != "" {
					re = regexp.MustCompile(tc.errRegexOther)
					match = re.MatchString(e)
				}
				if !match {
					t.Fatalf("want match for %q\nnot: %q",
						tc.errRegex, e)
				}
			}
		})
	}
}
