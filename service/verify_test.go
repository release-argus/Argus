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

func TestSlice_Print(t *testing.T) {
	// GIVEN a Slice
	tests := map[string]struct {
		slice    *Slice
		ordering []string
		want     string
	}{
		"nil slice with no ordering": {
			slice: nil,
			want:  "",
		},
		"nil slice with ordering": {
			ordering: []string{"foo", "bar"},
			slice:    nil,
			want:     "",
		},
		"slice with nil Service and empty Service": {
			ordering: []string{"foo", "bar"},
			slice:    &Slice{"foo": nil, "bar": &Service{}},
			want: `
	service:
	  bar: {}`,
		},
		"respects ordering": {
			ordering: []string{"zulu", "alpha"},
			slice: &Slice{
				"zulu":  &Service{ID: "zulu", Comment: "a"},
				"alpha": &Service{ID: "alpha", Comment: "b"}},
			want: `
	service:
	  zulu:
	    comment: a
	  alpha:
	    comment: b`,
		},
		"respects reversedordering": {
			ordering: []string{"alpha", "zulu"},
			slice: &Slice{
				"zulu":  &Service{ID: "zulu", Comment: "a"},
				"alpha": &Service{ID: "alpha", Comment: "b"}},
			want: `
	service:
	  alpha:
	    comment: b
	  zulu:
	    comment: a`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			if tc.want != "" {
				tc.want += "\n"
			}
			tc.want = strings.ReplaceAll(tc.want, "\t", "")
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN Print is called
			tc.slice.Print("", tc.ordering)

			// THEN it prints the expected output
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = stdout
			strOut := string(out)
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if strOut != tc.want {
				t.Errorf("Print should have given\n%q\nbut gave\n%q",
					tc.want, strOut)
			}
		})
	}
}

func TestService_CheckValues(t *testing.T) {
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
		errRegex         []string
	}{
		"options with errs": {
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			latestVersion: latestver.Lookup{
				Type: "github", URL: "release-argus/Argus"},
			options: *opt.New(
				nil, "10x", nil, nil, nil),
			errRegex: []string{
				`^test:$`,
				`^  options:$`,
				`^    interval: "[^"]+" <invalid>`},
		},
		"options,latest_version, with errs": {
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			options: *opt.New(
				nil, "10x", nil, nil, nil),
			latestVersion: latestver.Lookup{
				Type: "invalid", URL: "release-argus/Argus"},
			errRegex: []string{
				`^test:$`,
				`^  options:$`,
				`^    interval: "[^"]+" <invalid>`,
				`^  latest_version:$`,
				`^    type: "[^"]+" <invalid>`},
		},
		"latest_version, deployed_version with errs": {
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			options: *opt.New(
				nil, "10x", nil, nil, nil),
			latestVersion: latestver.Lookup{
				Type: "invalid", URL: "release-argus/Argus"},
			deployedVersion: &deployedver.Lookup{
				Regex: "[0-"},
			errRegex: []string{
				`^test:$`,
				`^  options:$`,
				`^    interval: "[^"]+" <invalid>`,
				`^  latest_version:$`,
				`^    type: "[^"]+" <invalid>`,
				`^  deployed_version:$`,
				`^    url: <required>`,
				`^    regex: "[^"]+" <invalid>`},
		},
		"latest_version, deployed_version, command with errs": {
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			options: *opt.New(
				nil, "10x", nil, nil, nil),
			latestVersion: latestver.Lookup{
				Type: "invalid", URL: "release-argus/Argus"},
			deployedVersion: &deployedver.Lookup{
				Regex: "[0-"},
			errRegex: []string{
				`^test:$`,
				`^  options:$`,
				`^    interval: "[^"]+" <invalid>`,
				`^  latest_version:$`,
				`^    type: "[^"]+" <invalid>`,
				`^  deployed_version:$`,
				`^    url: <required>`,
				`^    regex: "[^"]+" <invalid>`,
				`^  command:$`,
				`^    item_0: bash .* <invalid>.*templating`},
			commands: command.Slice{{"bash", "update.sh", "{{ version }"}},
		},
		"latest_version, deployed_version, notify with errs": {
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			options: *opt.New(
				nil, "10x", nil, nil, nil),
			latestVersion: latestver.Lookup{
				Type: "invalid", URL: "release-argus/Argus"},
			deployedVersion: &deployedver.Lookup{
				Regex: "[0-"},
			notifies: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", nil, nil,
					"discord",
					nil, nil, nil, nil)},
			commands: command.Slice{{
				"bash", "update.sh", "{{ version }"}},
			errRegex: []string{
				`^test:$`,
				`^  options:$`,
				`^    interval: "[^"]+" <invalid>`,
				`^  latest_version:$`,
				`^    type: "[^"]+" <invalid>`,
				`^  deployed_version:$`,
				`^    url: <required>`,
				`^    regex: "[^"]+" <invalid>`,
				`^  notify:`,
				`^    foo:$`,
				`^      url_fields:$`,
				`^        token: <required>`,
				`^        webhookid: <required>`,
				`^  command:$`,
				`^    item_0: bash .* <invalid>.*templating`},
		},
		"latest_version, deployed_version, webhook with errs": {
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			options: *opt.New(
				nil, "10x", nil, nil, nil),
			latestVersion: latestver.Lookup{
				Type: "invalid", URL: "release-argus/Argus"},
			deployedVersion: &deployedver.Lookup{
				Regex: "[0-"},
			commands: command.Slice{{
				"bash", "update.sh", "{{ version }"}},
			notifies: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", nil, nil,
					"discord",
					nil, nil, nil, nil)},
			webhooks: webhook.Slice{
				"wh": webhook.New(
					nil, nil,
					"0x",
					nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
			errRegex: []string{
				`^test:$`,
				`^  options:$`,
				`^    interval: "[^"]+" <invalid>`,
				`^  latest_version:$`,
				`^    type: "[^"]+" <invalid>`,
				`^  deployed_version:$`,
				`^    url: <required>`,
				`^    regex: "[^"]+" <invalid>`,
				`^  notify:`,
				`^    foo:$`,
				`^      url_fields:$`,
				`^        token: <required>`,
				`^        webhookid: <required>`,
				`^  command:$`,
				`^    item_0: bash .* <invalid>.*templating`,
				`^  webhook:$`,
				`^    wh:$`,
				`^      delay: "[^"]+" <invalid>`},
		},
		"has latest_version+deployed_version, webhook with errs": {
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			options: *opt.New(
				nil, "10x", nil, nil, nil),
			latestVersion: latestver.Lookup{
				Type: "invalid", URL: "release-argus/Argus"},
			deployedVersion: &deployedver.Lookup{
				Regex: "[0-"},
			commands: command.Slice{
				{"bash", "update.sh", "{{ version }"}},
			notifies: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", nil, nil,
					"discord",
					nil, nil, nil, nil)},
			webhooks: webhook.Slice{
				"wh": webhook.New(
					nil, nil,
					"0x",
					nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
			errRegex: []string{
				`^test:$`,
				`^  options:$`,
				`^    interval: "[^"]+" <invalid>`,
				`^  latest_version:$`,
				`^    type: "[^"]+" <invalid>`,
				`^  deployed_version:$`,
				`^    url: <required>`,
				`^    regex: "[^"]+" <invalid>`,
				`^  notify:`,
				`^    foo:$`,
				`^      url_fields:$`,
				`^        token: <required>`,
				`^        webhookid: <required>`,
				`^  command:$`,
				`^    item_0: bash .* <invalid>.*templating`,
				`^  webhook:$`,
				`^    wh:$`,
				`^      delay: "[^"]+" <invalid>`},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.svc.ID = "test"
			tc.svc.Options = tc.options
			tc.svc.LatestVersion = tc.latestVersion
			tc.svc.DeployedVersionLookup = tc.deployedVersion
			tc.svc.Command = tc.commands
			tc.svc.WebHook = tc.webhooks
			tc.svc.Notify = tc.notifies
			tc.svc.Dashboard = tc.dashboardOptions
			tc.svc.Init(
				&ServiceDefaults{}, &ServiceDefaults{},
				&shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{},
				&webhook.SliceDefaults{}, &webhook.WebHookDefaults{}, &webhook.WebHookDefaults{})

			// WHEN CheckValues is called
			err := tc.svc.CheckValues("")

			// THEN it err's when expected
			e := util.ErrorToString(err)
			lines := strings.Split(e, `\`)
			if len(tc.errRegex) > len(lines) {
				t.Fatalf("want %d errors:\n%v\ngot %d errors:\n%v",
					len(tc.errRegex), tc.errRegex, len(lines), lines)
			}
			for i := range tc.errRegex {
				re := regexp.MustCompile(tc.errRegex[i])
				match := re.MatchString(lines[i])
				if !match {
					t.Fatalf("%q didn't match %q\ngot:  %q",
						lines[i], tc.errRegex[i], e)
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
					Options: *opt.New(
						nil, "10s", nil, nil, nil),
					LatestVersion: latestver.Lookup{
						Type: "github", URL: "release-argus/Argus"}}},
			errRegex: `^$`,
		},
		"single invalid service": {
			slice: Slice{
				"first": {
					ID:      "test",
					Comment: "foo_comment",
					Options: *opt.New(
						nil, "10x", nil, nil, nil),
					LatestVersion: latestver.Lookup{
						Type: "github", URL: "release-argus/Argus"}}},
			errRegex: `interval: "[^"]+" <invalid>`,
		},
		"multiple invalid services": {
			slice: Slice{
				"foo": {
					ID:      "test",
					Comment: "foo_comment",
					Options: *opt.New(
						nil, "10x", nil, nil, nil),
					LatestVersion: latestver.Lookup{
						Type: "github", URL: "release-argus/Argus"}},
				"bar": {
					ID:      "test",
					Comment: "foo_comment",
					Options: *opt.New(
						nil, "10y", nil, nil, nil),
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
