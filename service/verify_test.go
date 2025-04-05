// Copyright [2025] [Argus]
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

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/notify/shoutrrr"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	dv_web "github.com/release-argus/Argus/service/deployed_version/types/web"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	latestver_base "github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

func TestSlice_Print(t *testing.T) {
	// GIVEN a Slice.
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
			want: test.TrimYAML(`
				service:
					bar: {}`),
		},
		"respects ordering": {
			ordering: []string{"zulu", "alpha"},
			slice: &Slice{
				"zulu":  &Service{ID: "zulu", Comment: "a"},
				"alpha": &Service{ID: "alpha", Comment: "b"}},
			want: test.TrimYAML(`
				service:
					zulu:
						comment: a
					alpha:
						comment: b`),
		},
		"respects reversed ordering": {
			ordering: []string{"alpha", "zulu"},
			slice: &Slice{
				"zulu":  &Service{ID: "zulu", Comment: "a"},
				"alpha": &Service{ID: "alpha", Comment: "b"}},
			want: test.TrimYAML(`
				service:
					alpha:
						comment: b
					zulu:
						comment: a`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout()

			if tc.want != "" {
				tc.want += "\n"
			}
			tc.want = strings.ReplaceAll(tc.want, "\t", "")

			// WHEN Print is called.
			tc.slice.Print("", tc.ordering)

			// THEN it prints the expected stdout.
			stdout := releaseStdout()
			if stdout != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, stdout)
			}
		})
	}
}

func TestDefaults_CheckValues(t *testing.T) {
	// GIVEN a Defaults.
	tests := map[string]struct {
		options       opt.Defaults
		latestVersion latestver_base.Defaults
		errRegex      string
	}{
		"valid": {
			options: *opt.NewDefaults(
				"10s", nil),
			latestVersion: latestver_base.Defaults{
				Require: filter.RequireDefaults{
					Docker: *filter.NewDockerCheckDefaults(
						"ghcr", "", "", "", "", nil)}},
		},
		"options with errs": {
			options: *opt.NewDefaults(
				"10x", nil),
			errRegex: test.TrimYAML(`
				^options:
					interval: "[^"]+" <invalid>.*$`),
		},
		"latestVersion with errs": {
			options: *opt.NewDefaults(
				"10s", nil),
			latestVersion: latestver_base.Defaults{
				Require: filter.RequireDefaults{
					Docker: *filter.NewDockerCheckDefaults(
						"randomType", "", "", "", "", nil)}},
			errRegex: test.TrimYAML(`
				^latest_version:
					require:
						docker:
							type: "[^"]+" <invalid>`),
		},
		"all errs": {
			options: *opt.NewDefaults(
				"10x", nil),
			latestVersion: latestver_base.Defaults{
				Require: filter.RequireDefaults{
					Docker: *filter.NewDockerCheckDefaults(
						"randomType", "", "", "", "", nil)}},
			errRegex: test.TrimYAML(`
				^options:
					interval: "[^"]+" <invalid>.*
				latest_version:
					require:
						docker:
							type: "[^"]+" <invalid>.*$`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := &Defaults{
				Options:       tc.options,
				LatestVersion: tc.latestVersion}

			// WHEN CheckValues is called.
			err := svc.CheckValues("")

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			lines := strings.Split(e, "\n")
			wantLines := strings.Count(tc.errRegex, "\n")
			if wantLines > len(lines) {
				t.Fatalf("%s\nwant: %d lines of error:\n%q\ngot:  %d lines:\n%v\n\nstdout: %q",
					packageName,
					wantLines, tc.errRegex,
					len(lines), lines,
					e)
				return
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
				return
			}
		})
	}
}

func TestSlice_CheckValues(t *testing.T) {
	// GIVEN a Slice.
	tests := map[string]struct {
		slice    *Slice
		errRegex string
	}{
		"nil slice": {
			slice:    nil,
			errRegex: `^$`,
		},
		"single valid service": {
			slice: &Slice{
				"first": {
					ID:      "test",
					Comment: "foo_comment",
					Options: *opt.New(
						nil, "10s", nil, nil, nil),
					LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
						return latestver.New(
							"github",
							"yaml", test.TrimYAML(`
								url: release-argus/Argus
							`),
							nil,
							nil,
							nil, nil)
					})}},
			errRegex: `^$`,
		},
		"single invalid service": {
			slice: &Slice{
				"first": {
					ID:      "test",
					Comment: "foo_comment",
					Options: *opt.New(
						nil, "10x", nil, nil, nil),
					LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
						return latestver.New(
							"github",
							"yaml", test.TrimYAML(`
								url: release-argus/Argus
							`),
							nil,
							nil,
							nil, nil)
					})}},
			errRegex: `interval: "[^"]+" <invalid>.*$`,
		},
		"multiple invalid services": {
			slice: &Slice{
				"foo": {
					ID:      "test",
					Comment: "foo_comment",
					Options: *opt.New(
						nil, "10x", nil, nil, nil),
					LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
						return latestver.New(
							"github",
							"yaml", test.TrimYAML(`
								url: release-argus/Argus
							`),
							nil,
							nil,
							nil, nil)
					})},
				"bar": {
					ID:      "test",
					Comment: "foo_comment",
					Options: *opt.New(
						nil, "10y", nil, nil, nil),
					LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
						return latestver.New(
							"github",
							"yaml", test.TrimYAML(`
								url: release-argus/Argus
							`),
							nil,
							nil,
							nil, nil)
					})}},
			errRegex: test.TrimYAML(`
				^bar:
					options:
						interval: "10y" <invalid>.*
				foo:
					options:
						interval: "10x".*$`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CheckValues is called.
			err := tc.slice.CheckValues("")

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
		})
	}
}

func TestService_CheckValues(t *testing.T) {
	// GIVEN a Service.
	tests := map[string]struct {
		svc              *Service
		options          opt.Options
		latestVersion    latestver.Lookup
		deployedVersion  deployedver.Lookup
		commands         command.Slice
		webhooks         webhook.Slice
		notifies         shoutrrr.Slice
		dashboardOptions DashboardOptions
		errRegex         string
	}{
		"nil service": {
			svc:      nil,
			errRegex: `^$`,
		},
		"options with errs": {
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			latestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New(
					"github",
					"yaml", test.TrimYAML(`
						url: release-argus/Argus
					`),
					nil,
					nil,
					nil, nil)
			}),
			options: *opt.New(
				nil, "10x", nil, nil, nil),
			errRegex: test.TrimYAML(`
				^options:
					interval: "[^"]+" <invalid>.*$`),
		},
		"options, latest_version with errs": {
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			options: *opt.New(
				nil, "10x", nil, nil, nil),
			latestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New(
					"github",
					"yaml", "",
					nil,
					nil,
					nil, nil)
			}),
			errRegex: test.TrimYAML(`
				^options:
					interval: "[^"]+" <invalid>.*
				latest_version:
					url: <required>.*$`),
		},
		"options, latest_version, deployed_version with errs": {
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			options: *opt.New(
				nil, "10x", nil, nil, nil),
			latestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New(
					"github",
					"yaml", "",
					nil,
					nil,
					nil, nil)
			}),
			deployedVersion: &dv_web.Lookup{
				Regex: `[0-`},
			errRegex: test.TrimYAML(`
				^options:
					interval: "10x" <invalid>.*
				latest_version:
					url: <required>.*
				deployed_version:
					url: <required>.*
					regex: "[^"]+" <invalid>.*$`),
		},
		"options, latest_version, deployed_version, notify with errs": {
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			options: *opt.New(
				nil, "10x", nil, nil, nil),
			latestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New(
					"github",
					"yaml", "",
					nil,
					nil,
					nil, nil)
			}),
			deployedVersion: &dv_web.Lookup{
				Regex: `[0-`},
			notifies: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "",
					"discord",
					nil, nil, nil,
					nil, nil, nil)},
			errRegex: test.TrimYAML(`
				^options:
					interval: "[^"]+" <invalid>.*
				latest_version:
					url: <required>.*
				deployed_version:
					url: <required>.*
					regex: "[^"]+" <invalid>.*
				notify:
					foo:
						url_fields:
							token: <required>.*
							webhookid: <required>.*$`),
		},
		"options, latest_version, deployed_version, notify, command with errs": {
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			options: *opt.New(
				nil, "10x", nil, nil, nil),
			latestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New(
					"github",
					"yaml", "",
					nil,
					nil,
					nil, nil)
			}),
			deployedVersion: &dv_web.Lookup{
				Regex: `[0-`},
			notifies: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "",
					"discord",
					nil, nil, nil,
					nil, nil, nil)},
			commands: command.Slice{{
				"bash", "update.sh", "{{ version }"}},
			errRegex: test.TrimYAML(`
				^options:
					interval: "[^"]+" <invalid>.*
				latest_version:
					url: <required>.*
				deployed_version:
					url: <required>.*
					regex: "[^"]+" <invalid>.*
				notify:
					foo:
						url_fields:
							token: <required>.*
							webhookid: <required>.*
				command:
					item_0: bash .* <invalid>.*templating.*$`),
		},
		"options, latest_version, deployed_version, notify, command, webhook with errs": {
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			options: *opt.New(
				nil, "10x", nil, nil, nil),
			latestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New(
					"github",
					"yaml", "",
					nil,
					nil,
					nil, nil)
			}),
			deployedVersion: &dv_web.Lookup{
				Regex: `[0-`},
			commands: command.Slice{{
				"bash", "update.sh", "{{ version }"}},
			notifies: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "",
					"discord",
					nil, nil, nil,
					nil, nil, nil)},
			webhooks: webhook.Slice{
				"wh": webhook.New(
					nil, nil,
					"0x",
					nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
			errRegex: test.TrimYAML(`
				^options:
					interval: "10x" <invalid>.*
				latest_version:
					url: <required>.*
				deployed_version:
					url: <required>.*
					regex: "[^"]+" <invalid>.*
				notify:
					foo:
						url_fields:
							token: <required>.*
							webhookid: <required>.*
				command:
					item_0: bash .* <invalid>.*templating.*
				webhook:
					wh:
						type: <required>.*
						delay: "[^"]+" <invalid>.*
						url: <required>.*
						secret: <required>.*$`),
		},
		"options, latest_version, deployed_version, notify, command, webhook with no errs": {
			svc: &Service{
				ID: "test", Comment: "foo_comment"},
			options: *opt.New(
				nil, "10s", nil, nil, nil),
			latestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New(
					"github",
					"yaml", test.TrimYAML(`
						url: release-argus/Argus
					`),
					nil,
					nil,
					nil, nil)
			}),
			deployedVersion: test.IgnoreError(t, func() (deployedver.Lookup, error) {
				return deployedver.New(
					"url",
					"yaml", test.TrimYAML(`
						url: https://example.com
					`),
					nil,
					nil,
					nil, nil)
			}),
			commands: command.Slice{{
				"bash", "update.sh", "{{ version }}"}},
			notifies: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "",
					"discord",
					nil,
					map[string]string{
						"token":     "x",
						"webhookid": "y"},
					nil,
					nil, nil, nil)},
			webhooks: webhook.Slice{
				"wh": webhook.New(
					nil, nil,
					"0s",
					nil, nil, nil, nil, nil, "", nil,
					"", "",
					nil, nil, nil)},
			errRegex: "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.svc != nil {
				tc.svc.ID = "test"
				tc.svc.Options = tc.options
				tc.svc.LatestVersion = tc.latestVersion
				tc.svc.DeployedVersionLookup = tc.deployedVersion
				tc.svc.Command = tc.commands
				tc.svc.WebHook = tc.webhooks
				tc.svc.Notify = tc.notifies
				tc.svc.Dashboard = tc.dashboardOptions
				tc.svc.Init(
					&Defaults{}, &Defaults{},
					&shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{},
					&webhook.SliceDefaults{}, &webhook.Defaults{}, &webhook.Defaults{})
			}

			// WHEN CheckValues is called.
			err := tc.svc.CheckValues("")

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			lines := strings.Split(e, "\n")
			wantLines := strings.Count(tc.errRegex, "\n")
			if wantLines > len(lines) {
				t.Fatalf("%s\nwant: %d lines of error:\n%q\ngot:  %d lines:\n%v\n\nstdout: %q",
					packageName,
					wantLines, tc.errRegex,
					len(lines), lines,
					e)
				return
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
				return
			}
		})
	}
}
