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
	deployedver_base "github.com/release-argus/Argus/service/deployed_version/types/base"
	dv_web "github.com/release-argus/Argus/service/deployed_version/types/web"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	latestver_base "github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

func TestServices_Print(t *testing.T) {
	// GIVEN a Services.
	tests := map[string]struct {
		services *Services
		ordering []string
		want     string
	}{
		"nil map with no ordering": {
			services: nil,
			want:     "",
		},
		"nil map with ordering": {
			ordering: []string{"foo", "bar"},
			services: nil,
			want:     "",
		},
		"map with nil Service and empty Service": {
			ordering: []string{"foo", "bar"},
			services: &Services{"foo": nil, "bar": &Service{}},
			want: test.TrimYAML(`
				service:
					bar: {}`),
		},
		"respects ordering": {
			ordering: []string{"zulu", "alpha"},
			services: &Services{
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
			services: &Services{
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
			tc.services.Print("", tc.ordering)

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
	// GIVEN Defaults.
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

func TestServices_CheckValues(t *testing.T) {
	// GIVEN a Services.
	tests := map[string]struct {
		services *Services
		errRegex string
		changed  bool
	}{
		"nil map": {
			services: nil,
			errRegex: `^$`,
			changed:  false,
		},
		"single valid service": {
			services: &Services{
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
			changed:  false,
		},
		"multiple valid services": {
			services: &Services{
				"first": {
					ID:      "1",
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
					})},
				"second": {
					ID:      "2",
					Comment: "bar_comment",
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
			changed:  false,
		},
		"multiple valid services, 1+ changed": {
			services: &Services{
				"first": {
					ID:      "1",
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
					}),
					WebHook: map[string]*webhook.WebHook{
						"wh": {
							Base: webhook.Base{
								Type:   "github",
								URL:    "example.com",
								Secret: "Argus",
								CustomHeaders: &webhook.Headers{
									webhook.Header{
										Key: "foo", Value: "bar"}},
							}}}},
				"second": {
					ID:      "2",
					Comment: "bar_comment",
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
			changed:  true,
		},
		"multiple valid services, 1+ changed but some error, so changed false": {
			services: &Services{
				"first": {
					ID:      "1",
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
					}),
					WebHook: map[string]*webhook.WebHook{
						"wh": {
							Base: webhook.Base{
								Type:   "github",
								URL:    "example.com",
								Secret: "Argus",
								CustomHeaders: &webhook.Headers{
									webhook.Header{
										Key: "foo", Value: "bar"}},
							}}}},
				"second": {
					ID:      "2",
					Comment: "bar_comment",
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
			errRegex: test.TrimYAML(`
				^second:
					options:
						interval: "10x" <invalid>.*$`),
			changed: false,
		},
		"single invalid service": {
			services: &Services{
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
			changed:  false,
		},
		"multiple invalid services": {
			services: &Services{
				"foo": {
					ID:      "test",
					Comment: "foo_comment",
					Options: *opt.New(
						nil, "10x", nil, nil, nil),
					LatestVersion: nil},
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
					}),
					DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
						return deployedver.New(
							"url",
							"yaml", test.TrimYAML(`
								method: "SOMETHING"
							`),
							nil,
							nil,
							&deployedver_base.Defaults{}, &deployedver_base.Defaults{})
					}),
				}},
			errRegex: test.TrimYAML(`
				^bar:
					options:
						interval: "10y" <invalid>.*
					deployed_version:
						url: <required>.*
						method: "[^"]+" <invalid>.*
				foo:
					options:
						interval: "10x" <invalid>.*
					latest_version:
						.*nil.*$`),
			changed: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.services != nil {
				for _, svc := range *tc.services {
					svc.Init(
						&Defaults{}, &Defaults{},
						&shoutrrr.ShoutrrrsDefaults{}, &shoutrrr.ShoutrrrsDefaults{}, &shoutrrr.ShoutrrrsDefaults{},
						&webhook.WebHooksDefaults{}, &webhook.Defaults{}, &webhook.Defaults{})
				}
			}

			// WHEN CheckValues is called.
			err, changed := tc.services.CheckValues("")

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
			// AND the 'changed' flag matches expectation.
			if changed != tc.changed {
				t.Errorf("%s\nchanged flag mismatch\nwant: %t\ngot:  %t",
					packageName, tc.changed, changed)
			}
		})
	}
}

func TestService_CheckValues(t *testing.T) {
	// GIVEN a Service.
	tests := map[string]struct {
		svc      *Service
		errRegex string
		changed  bool
	}{
		"nil service": {
			svc:      nil,
			errRegex: `^$`,
			changed:  false,
		},
		"options with errs": {
			svc: &Service{
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
				}),
			},
			errRegex: test.TrimYAML(`
				^options:
					interval: "[^"]+" <invalid>.*$`),
			changed: false,
		},
		"options, latest_version is nil": {
			svc: &Service{
				ID:      "test",
				Comment: "foo_comment",
				Options: *opt.New(
					nil, "10x", nil, nil, nil),
				LatestVersion: nil,
			},
			errRegex: test.TrimYAML(`
				^options:
					interval: "[^"]+" <invalid>.*
				latest_version:
					.*nil.*$`),
			changed: false,
		},
		"options, latest_version with errs": {
			svc: &Service{
				ID:      "test",
				Comment: "foo_comment",
				Options: *opt.New(
					nil, "10x", nil, nil, nil),
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
			},
			errRegex: test.TrimYAML(`
				^options:
					interval: "[^"]+" <invalid>.*
				latest_version:
					url: <required>.*$`),
			changed: false,
		},
		"options, latest_version, deployed_version with errs": {
			svc: &Service{
				ID:      "test",
				Comment: "foo_comment",
				Options: *opt.New(
					nil, "10x", nil, nil, nil),
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
				DeployedVersionLookup: &dv_web.Lookup{
					Regex: `[0-`},
			},
			errRegex: test.TrimYAML(`
				^options:
					interval: "10x" <invalid>.*
				latest_version:
					url: <required>.*
				deployed_version:
					url: <required>.*
					method: <required>.*
					regex: "[^"]+" <invalid>.*$`),
			changed: false,
		},
		"options, latest_version, deployed_version, notify with errs": {
			svc: &Service{
				ID:      "test",
				Comment: "foo_comment",
				Options: *opt.New(
					nil, "10x", nil, nil, nil),
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
				DeployedVersionLookup: &dv_web.Lookup{
					Regex: `[0-`},
				Notify: shoutrrr.Shoutrrrs{
					"foo": shoutrrr.New(
						nil, "",
						"discord",
						nil, nil, nil,
						nil, nil, nil)},
			},
			errRegex: test.TrimYAML(`
				^options:
					interval: "[^"]+" <invalid>.*
				latest_version:
					url: <required>.*
				deployed_version:
					url: <required>.*
					method: <required>.*
					regex: "[^"]+" <invalid>.*
				notify:
					foo:
						url_fields:
							token: <required>.*
							webhookid: <required>.*$`),
			changed: false,
		},
		"options, latest_version, deployed_version, notify, command with errs": {
			svc: &Service{
				ID:      "test",
				Comment: "foo_comment",
				Options: *opt.New(
					nil, "10x", nil, nil, nil),
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
				DeployedVersionLookup: &dv_web.Lookup{
					Regex: `[0-`},
				Notify: shoutrrr.Shoutrrrs{
					"foo": shoutrrr.New(
						nil, "",
						"discord",
						nil, nil, nil,
						nil, nil, nil)},
				Command: command.Commands{{
					"bash", "update.sh", "{{ version }"}},
			},
			errRegex: test.TrimYAML(`
				^options:
					interval: "[^"]+" <invalid>.*
				latest_version:
					url: <required>.*
				deployed_version:
					url: <required>.*
					method: <required>.*
					regex: "[^"]+" <invalid>.*
				notify:
					foo:
						url_fields:
							token: <required>.*
							webhookid: <required>.*
				command:
					item_0: bash .* <invalid>.*templating.*$`),
			changed: false,
		},
		"options, latest_version, deployed_version, notify, command, webhook with errs": {
			svc: &Service{
				ID:      "test",
				Comment: "foo_comment",
				Options: *opt.New(
					nil, "10x", nil, nil, nil),
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
				DeployedVersionLookup: &dv_web.Lookup{
					Regex: `[0-`},
				Command: command.Commands{{
					"bash", "update.sh", "{{ version }"}},
				Notify: shoutrrr.Shoutrrrs{
					"foo": shoutrrr.New(
						nil,
						"", "discord",
						nil, nil, nil,
						nil, nil, nil)},
				WebHook: webhook.WebHooks{
					"wh": webhook.New(
						nil, nil,
						"0x",
						nil, nil,
						"wh",
						nil, nil, nil,
						"",
						nil,
						"", "",
						nil, nil, nil)},
			},
			errRegex: test.TrimYAML(`
				^options:
					interval: "10x" <invalid>.*
				latest_version:
					url: <required>.*
				deployed_version:
					url: <required>.*
					method: <required>.*
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
			changed: false,
		},
		"options, latest_version, deployed_version, notify, command, webhook with no errs": {
			svc: &Service{
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
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
						url: https://example.com
						method: GET
					`),
						nil,
						nil,
						nil, nil)
				}),
				Command: command.Commands{{
					"bash", "update.sh", "{{ version }}"}},
				Notify: shoutrrr.Shoutrrrs{
					"foo": shoutrrr.New(
						nil,
						"", "discord",
						nil,
						map[string]string{
							"token":     "x",
							"webhookid": "y"},
						nil,
						nil, nil, nil)},
				WebHook: webhook.WebHooks{
					"wh": webhook.New(
						nil, nil,
						"0s",
						nil, nil,
						"wh",
						nil,
						nil, nil,
						"secret",
						nil,
						"github", "https://example.com",
						nil, nil, nil)},
			},
			errRegex: "^$",
			changed:  false,
		},
		"notify changed": {
			svc: &Service{
				ID:      "test",
				Comment: "foo_comment",
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
						url: release-argus/Argus
					`),
						nil,
						nil,
						nil, nil)
				}),
				Notify: shoutrrr.Shoutrrrs{
					"foo": shoutrrr.New(
						nil,
						"", "generic",
						nil,
						map[string]string{
							"host":           "x",
							"secret":         "y",
							"custom_headers": `{"foo": "bar"}`},
						nil,
						nil, nil, nil)},
			},
			errRegex: `^$`,
			changed:  true,
		},
		"webhook changed": {
			svc: &Service{
				ID:      "test",
				Comment: "foo_comment",
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
						url: release-argus/Argus
					`),
						nil,
						nil,
						nil, nil)
				}),
				WebHook: map[string]*webhook.WebHook{
					"wh": {
						Base: webhook.Base{
							Type:   "github",
							URL:    "example.com",
							Secret: "Argus",
							CustomHeaders: &webhook.Headers{
								webhook.Header{
									Key: "foo", Value: "bar"}},
						}}},
			},
			errRegex: `^$`,
			changed:  true,
		},
		"not changed if we have errors": {
			svc: &Service{
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
				}),
				Notify: shoutrrr.Shoutrrrs{
					"foo": shoutrrr.New(
						nil,
						"", "generic",
						nil,
						map[string]string{
							"host":           "x",
							"secret":         "y",
							"custom_headers": `{"foo": "bar"}`},
						nil,
						nil, nil, nil)},
				WebHook: map[string]*webhook.WebHook{
					"wh": {
						Base: webhook.Base{
							Type:   "github",
							URL:    "example.com",
							Secret: "Argus",
							CustomHeaders: &webhook.Headers{
								webhook.Header{
									Key: "foo", Value: "bar"}},
						}}},
			},
			errRegex: test.TrimYAML(`
				^options:
					interval: "10x" <invalid>.*$`),
			changed: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.svc != nil {
				tc.svc.Init(
					&Defaults{}, &Defaults{},
					&shoutrrr.ShoutrrrsDefaults{}, &shoutrrr.ShoutrrrsDefaults{}, &shoutrrr.ShoutrrrsDefaults{},
					&webhook.WebHooksDefaults{}, &webhook.Defaults{}, &webhook.Defaults{})
			}

			// WHEN CheckValues is called.
			err, changed := tc.svc.CheckValues("")

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
			// AND the 'changed' flag matches expectation.
			if changed != tc.changed {
				t.Errorf("%s\nchanged flag mismatch\nwant: %t\ngot:  %t",
					packageName, tc.changed, changed)
			}
		})
	}
}
