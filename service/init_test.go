// Copyright [2026] [Argus]
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
	"fmt"
	"strconv"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/release-argus/Argus/command"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	dvtest "github.com/release-argus/Argus/service/deployed_version/test"
	lvtest "github.com/release-argus/Argus/service/latest_version/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/web/metric"
	whtest "github.com/release-argus/Argus/webhook/test"
)

func TestService_IconURL(t *testing.T) {
	nilValue := "<nil>"
	// GIVEN: a Lookup.
	tests := []struct {
		name          string
		dashboardIcon string
		want          string
		notify        shoutrrr.Shoutrrrs
	}{
		{
			name:          "no dashboard.icon",
			want:          nilValue,
			dashboardIcon: "",
		},
		{
			name:          "no icon anywhere",
			want:          nilValue,
			dashboardIcon: "",
			notify: shoutrrr.Shoutrrrs{
				"test": {
					Main:         &shoutrrr.Defaults{},
					Defaults:     &shoutrrr.Defaults{},
					HardDefaults: &shoutrrr.Defaults{},
				},
			},
		},
		{
			name:          "emoji icon",
			want:          nilValue,
			dashboardIcon: ":smile:",
		},
		{
			name:          "web icon",
			want:          "https://example.com/icon.png",
			dashboardIcon: "https://example.com/icon.png",
		},
		{
			name: "notify icon only",
			want: "https://example.com/icon.png",
			notify: shoutrrr.Shoutrrrs{"test": shoutrrr.New(
				nil,
				"", "",
				nil, nil,
				map[string]string{
					"icon": "https://example.com/icon.png",
				},
				&shoutrrr.Defaults{}, &shoutrrr.Defaults{}, &shoutrrr.Defaults{},
			),
			},
		},
		{
			name:          "notify icon takes precedence over emoji",
			want:          "https://example.com/icon.png",
			dashboardIcon: ":smile:",
			notify: shoutrrr.Shoutrrrs{"test": shoutrrr.New(
				nil,
				"", "",
				nil, nil,
				map[string]string{
					"icon": "https://example.com/icon.png",
				},
				&shoutrrr.Defaults{}, &shoutrrr.Defaults{}, &shoutrrr.Defaults{},
			),
			},
		},
		{
			name:          "dashboard icon takes precedence over notify icon",
			want:          "https://root.com/icon.png",
			dashboardIcon: "https://root.com/icon.png",
			notify: shoutrrr.Shoutrrrs{"test": shoutrrr.New(
				nil,
				"", "",
				nil, nil,
				map[string]string{
					"icon": "https://example.com/icon.png",
				},
				&shoutrrr.Defaults{}, &shoutrrr.Defaults{}, &shoutrrr.Defaults{},
			),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			svc := testService(t, tc.name, "github", "url")

			svc.Dashboard.Icon = tc.dashboardIcon
			svc.Notify = tc.notify

			// WHEN: IconURL is called.
			got := svc.IconURL()

			// THEN: the function returns the correct result.
			if gotStr := util.DerefOr(got, nilValue); gotStr != tc.want {
				t.Errorf(
					"%s\nService.IconURL() value mismatch\ngot:  %q\nwant: %q",
					packageName, gotStr, tc.want,
				)
			}
		})
	}
}

func TestService_Init(t *testing.T) {
	// GIVEN: defaults.
	svcCfg := plainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()

	// AND: a Service.
	tests := []struct {
		name     string
		svc      *Service
		wantIcon string
	}{
		{
			name: "bare service - Name defaulted to ID",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
					`)),
					"Init",
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name: "service with Name",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						name: other-name
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
					`)),
					"Init",
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name: "service with notify - doesn't set fallback when Service has a Dashboard.Icon",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
						notify:
							test:
								type: discord
								params:
									icon: notify-icon
						dashboard:
							icon: dashboard-icon
					`)),
					"Init",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			wantIcon: "dashboard-icon",
		},
		{
			name: "service with notify - does set fallback when Service has no Dashboard.Icon",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
						notify:
							foo:
								type: discord
								params:
									icon: https://example.com/notify-icon-1
							baz:
							bar:
								type: discord
								params:
									icon: https://example.com/notify-icon-2
					`)),
					"Init",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			wantIcon: "https://example.com/notify-icon-2",
		},
		{
			name: "service with notify, command and webhook",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
						notify:
							test:
								type: discord
						command:
							- - ls
						webhook:
							test:
						`+whtest.WebHook(false, false, false).String(".   ")+`
					`)),
					"Init",
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name: "service with notifies from defaults",
			svc: test.Must(t, func() (*Service, error) {
				svcCfg := plainDefaultsConfig()
				svcCfg.Soft.Notify = map[string]struct{}{
					"foo": {},
				}

				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
					`)),
					"Init",
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name: "service with notifies not from defaults",
			svc: test.Must(t, func() (*Service, error) {
				svcCfg := plainDefaultsConfig()
				svcCfg.Soft.Notify = map[string]struct{}{
					"foo": {},
				}

				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
						notify:
							test: {}
					`)),
					"Init",
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name: "service with commands from defaults",
			svc: test.Must(t, func() (*Service, error) {
				svcCfg := plainDefaultsConfig()
				svcCfg.Soft.Command = command.Commands{
					{"ls"},
				}

				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
					`)),
					"Init",
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name: "service with commands not from defaults",
			svc: test.Must(t, func() (*Service, error) {
				svcCfg := plainDefaultsConfig()
				svcCfg.Soft.Command = command.Commands{
					{"ls"},
				}

				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
						command:
							- - test
					`)),
					"Init",
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name: "service with webhooks from defaults",
			svc: test.Must(t, func() (*Service, error) {
				svcCfg := plainDefaultsConfig()
				svcCfg.Soft.WebHook = map[string]struct{}{
					"bar": {},
				}

				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
					`)),
					"Init",
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name: "service with webhooks not from defaults",
			svc: test.Must(t, func() (*Service, error) {
				svcCfg := plainDefaultsConfig()
				svcCfg.Soft.WebHook = map[string]struct{}{
					"bar": {},
				}

				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
						webhook:
							test:
						`+whtest.WebHook(false, false, false).String(".   ")+`
					`)),
					"Init",
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name: "service with webhooks/commands from defaults and notify overridden",
			svc: test.Must(t, func() (*Service, error) {
				svcCfg := plainDefaultsConfig()
				svcCfg.Soft.Notify = map[string]struct{}{
					"foo": {},
				}
				svcCfg.Soft.Command = command.Commands{
					{"ls"},
				}
				svcCfg.Soft.WebHook = map[string]struct{}{
					"bar": {},
				}

				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
						notify:
							test: {}
					`)),
					"Init",
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name: "DeployedVersionLookup",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						deployed_version:
							type: url
							url: `+test.LookupPlain["url_valid"]+`
					`)),
					"Init",
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := tc.svc
			s.ID = tc.name
			hadNotify := util.SortedKeys(s.Notify)
			hadWebHook := util.SortedKeys(s.WebHook)
			hadCommand := make(command.Commands, len(s.Command))
			copy(hadCommand, s.Command)

			// AND: various channels.
			announceChannel := make(chan []byte, 1)
			databaseChannel := make(chan dbtype.Message, 1)
			saveChannel := make(chan bool, 1)

			// WHEN: init is called on it.
			tc.svc.init(
				notifyCfg,
				whCfg,
				announceChannel,
				databaseChannel,
				saveChannel,
			)

			prefix := fmt.Sprintf(
				"%s\nService.Init(notifyDefaults=%v, WebHookDefaults=%v)",
				packageName, notifyCfg, whCfg,
			)

			// THEN: Mains/Defaults/HardDefaults are handed to each WebHook.
			hadNotifyLength := len(hadNotify)
			gotNotifyLength := len(tc.svc.Notify)
			if gotNotifyLength != 0 {
				for i := range tc.svc.Notify {
					if tc.svc.Notify[i].Main == nil {
						t.Errorf("%s Notify[%q].Main is nil", prefix, i)
					}
					if tc.svc.Notify[i].Defaults == nil {
						t.Errorf("%s Notify[%q].Defaults is nil", prefix, i)
					}
					if tc.svc.Notify[i].HardDefaults == nil {
						t.Errorf("%s Notify[%q].HardDefaults is nil", prefix, i)
					}
				}
			}

			// AND: Notifiers are not overridden if non-empty originally.
			if hadNotifyLength != 0 && gotNotifyLength != hadNotifyLength {
				t.Fatalf(
					"%s Notify length mismatch\ngot:  %d (%v)\nwant: %d (%v)",
					prefix,
					gotNotifyLength, util.SortedKeys(tc.svc.Notify),
					hadNotifyLength, hadNotify,
				)
			}

			// AND: Notify is set to the default values when empty.
			wantNotify := hadNotify
			if defaultNotifiers := tc.svc.Defaults.Notify; defaultNotifiers != nil && hadNotifyLength == 0 {
				wantNotify = make([]string, len(defaultNotifiers))
				wantNotify = util.SortedKeys(defaultNotifiers)
			}
			for _, i := range wantNotify {
				if tc.svc.Notify[i] == nil {
					t.Errorf(
						"%s Notify[%q] is nil",
						prefix, i,
					)
				}
			}

			// THEN: CommandController is nil'd when we have no Command, non-nil otherwise.
			hadCommandLength := len(hadCommand)
			gotCommandLength := len(tc.svc.Command)
			if gotCommandLength != 0 {
				if tc.svc.CommandController == nil {
					t.Errorf(
						"%s CommandController is still nil with %v Commands present",
						prefix, tc.svc.Command,
					)
				}
			} else if tc.svc.CommandController != nil {
				t.Errorf(
					"%s CommandController should be nil with %v Commands present",
					prefix, tc.svc.Command,
				)
			}

			// AND: Commands are not overridden if non-empty originally.
			if hadCommandLength != 0 && gotCommandLength != hadCommandLength {
				t.Fatalf(
					"%s Command length changed unexpectedly\ngot:  %d (%v)\nwant: %d (%v)",
					prefix,
					gotCommandLength, tc.svc.Command,
					hadCommandLength, hadCommand,
				)
			}

			// AND: Command is set to the default values when empty.
			wantCommand := hadCommand
			if defaultCommands := tc.svc.Defaults.Command; defaultCommands != nil && hadCommandLength == 0 {
				wantCommand = make(command.Commands, len(defaultCommands))
				wantCommand = defaultCommands
			}
			for i := range wantCommand {
				got := tc.svc.Command[i].String()
				want := wantCommand[i].String()
				if got != want {
					t.Errorf(
						"%s Command[%d] changed unexpected\ngot:  %q\nwant: %q",
						prefix, i,
						got, want,
					)
				}
			}

			// THEN: Mains/Defaults/HardDefaults are handed to each WebHook.
			hadWebHookLength := len(hadWebHook)
			gotWebHookLength := len(tc.svc.WebHook)
			if gotWebHookLength != 0 {
				for i := range tc.svc.WebHook {
					if tc.svc.WebHook[i].Main == nil {
						t.Errorf("%s WebHook[%q].Main is nil", prefix, i)
					}
					if tc.svc.WebHook[i].Defaults == nil {
						t.Errorf("%s WebHook[%q].Defaults is nil", prefix, i)
					}
					if tc.svc.WebHook[i].HardDefaults == nil {
						t.Errorf("%s WebHook[%q].HardDefaults is nil", prefix, i)
					}
				}
			}

			// AND: WebHooks are not overridden if non-empty originally.
			if hadWebHookLength > 0 && gotWebHookLength != hadWebHookLength {
				t.Fatalf(
					"%s WebHook length changed\ngot:  %d (%v)\nwant: %d (%v)",
					prefix,
					gotWebHookLength, util.SortedKeys(tc.svc.WebHook),
					hadWebHookLength, hadWebHook,
				)
			}

			// AND: WebHook is set to the default values when empty.
			wantWebHook := hadWebHook
			if defaultWebHooks := tc.svc.Defaults.WebHook; defaultWebHooks != nil && hadWebHookLength == 0 {
				wantWebHook = make([]string, len(defaultWebHooks))
				wantWebHook = util.SortedKeys(defaultWebHooks)
			}
			for _, i := range wantWebHook {
				if tc.svc.WebHook[i] == nil {
					t.Errorf("%s WebHook[%q] is nil", prefix, i)
				}
			}
			// 	Dashboard
			gotIcon := tc.svc.Dashboard.GetIcon()
			if tc.wantIcon != gotIcon {
				t.Errorf(
					"%s Dashboard.GetIcon() value mismatch\ngot:  %q\nwant: %q",
					prefix, gotIcon, tc.wantIcon,
				)
			}
			// Status - Pointers.
			fieldTests := []test.FieldAssertion{
				{Name: "AnnounceChannel", Got: tc.svc.Status.AnnounceChannel, Want: announceChannel, Mode: test.CompareSamePointer},
				{Name: "DatabaseChannel", Got: tc.svc.Status.DatabaseChannel, Want: databaseChannel, Mode: test.CompareSamePointer},
				{Name: "SaveChannel", Got: tc.svc.Status.SaveChannel, Want: saveChannel, Mode: test.CompareSamePointer},
				{Name: "Dashboard", Got: tc.svc.Status.Dashboard, Want: &tc.svc.Dashboard, Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Status"); err != nil {
				t.Fatal(err)
			}
			// Status - ServiceInfo.
			serviceInfo := tc.svc.Status.GetServiceInfo()
			wantServiceInfoName := util.ValueOr(tc.svc.Name, tc.svc.ID)
			fieldTests = []test.FieldAssertion{
				{Name: "ID", Got: serviceInfo.ID, Want: tc.svc.ID, Mode: test.CompareEqual},
				{Name: "Name", Got: serviceInfo.Name, Want: wantServiceInfoName, Mode: test.CompareEqual},
				{Name: "Comment", Got: serviceInfo.Comment, Want: tc.svc.Comment, Mode: test.CompareEqual},
				{Name: "Icon", Got: serviceInfo.Icon, Want: tc.wantIcon, Mode: test.CompareEqual},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Status.ServiceInfo"); err != nil {
				t.Fatal(err)
			}
			if !util.AreSlicesEqual(serviceInfo.Tags, tc.svc.Dashboard.Tags) {
				t.Errorf(
					"%s Status.ServiceInfo.Tags mismatch\ngot:  %v\nwant: %v",
					prefix, serviceInfo.Tags, tc.svc.Dashboard.Tags,
				)
			}
			// Status - Fails.
			fieldTests = []test.FieldAssertion{
				{Name: "Command", Got: tc.svc.Status.Fails.Command.Length(), Want: len(tc.svc.Command), Mode: test.CompareEqual},
				{Name: "Shoutrrr", Got: tc.svc.Status.Fails.Shoutrrr.Length(), Want: len(tc.svc.Notify), Mode: test.CompareEqual},
				{Name: "WebHook", Got: tc.svc.Status.Fails.WebHook.Length(), Want: len(tc.svc.WebHook), Mode: test.CompareEqual},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Status.ServiceInfo"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestService_InitMetrics__ResetMetrics__DeleteMetrics(t *testing.T) {
	svcCfg := plainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()

	// GIVEN: a Service.
	tests := []struct {
		name               string
		nilDeployedVersion bool
		nilCommand         bool
		nilNotify          bool
		nilWebHook         bool
	}{
		{
			name: "all defined",
		},
		{
			name:               "nil DeployedVersionLookup",
			nilDeployedVersion: true,
		},
		{
			name:       "nil Command",
			nilCommand: true,
		},
		{
			name:      "nil Notify",
			nilNotify: true,
		},
		{
			name:       "nil WebHook",
			nilWebHook: true,
		},
		{
			name:               "nil all",
			nilDeployedVersion: true,
			nilCommand:         true,
			nilNotify:          true,
			nilWebHook:         true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're testing metrics.

			activeStates := []bool{true, false}
			for _, active := range activeStates {
				testCommand := command.Command{"ls"}
				testNotify := shoutrrrtest.Shoutrrr(false, false)
				testWebHook := whtest.WebHook(false, false, false)
				svc := test.Must(t, func() (s *Service, err error) {
					return DecodeService(
						"yaml", []byte(test.TrimYAML(`
							options:
								active: `+strconv.FormatBool(active)+`
							latest_version:
							`+lvtest.Lookup(t, "github", false).String("  ")+`
							deployed_version:
							`+dvtest.Lookup(t, "url", false, "").String("  ")+`
							command:
							- `+testCommand.JSON()+`
							notify:
								`+testNotify.ID+`:
							`+testNotify.String("    ")+`
							webhook:
								`+testWebHook.ID+`:
							`+testWebHook.String("    ")+`
						`)),
						"TestService_InitMetrics_ResetMetrics_DeleteMetrics--"+tc.name,
						svcCfg,
						notifyCfg,
						whCfg,
					)
				})

				// Set the versions.
				svc.Status.SetLatestVersion("0.0.2", "", false)
				svc.Status.SetDeployedVersion("0.0.2", "", false)

				// nil the vars.
				var deployedVersionType string
				if tc.nilDeployedVersion {
					svc.DeployedVersionLookup = nil
				} else {
					deployedVersionType = svc.DeployedVersionLookup.GetType()
				}
				if tc.nilCommand {
					svc.Command = nil
				}
				if tc.nilNotify {
					svc.Notify = nil
				}
				if tc.nilWebHook {
					svc.WebHook = nil
				}

				// metrics:
				// 	latest_version_query_result_total.
				latestVersionMetric := metric.LatestVersionIsDeployed.WithLabelValues(
					svc.ID,
				)
				// 	deployed_version_query_result_last.
				deployedVersionMetric := metric.DeployedVersionQueryResultLast.WithLabelValues(
					svc.ID, deployedVersionType,
				)
				// 	command_result_total.
				commandMetric := metric.CommandResultTotal.WithLabelValues(
					testCommand.String(), metric.ActionResultSuccess, svc.ID,
				)
				// 	notify_result_total.
				notifyMetric := metric.NotifyResultTotal.WithLabelValues(
					testNotify.ID, metric.ActionResultSuccess, svc.ID, testNotify.GetType(),
				)
				// 	webhook_result_total.
				webhookMetric := metric.WebHookResultTotal.WithLabelValues(
					testWebHook.ID, metric.ActionResultSuccess, svc.ID,
				)
				// 	service_count_current.
				serviceCountCurrentActive := metric.ServiceCountCurrent.WithLabelValues(
					metric.ServiceStateActive,
				)
				initialServiceCountCurrentActive := testutil.ToFloat64(serviceCountCurrentActive)
				serviceCountCurrentInactive := metric.ServiceCountCurrent.WithLabelValues(
					metric.ServiceStateInactive,
				)
				initialServiceCountCurrentInactive := testutil.ToFloat64(serviceCountCurrentInactive)

				// WHEN: initMetrics is called on it. #
				svc.initMetrics()

				prefix := fmt.Sprintf("%s\nService.initMetrics()", packageName)

				// THEN: the metrics are created:
				want := float64(3)
				oldWant := want
				// 	latest_version_is_deployed.
				if !active {
					want = 0
				} else {
					latestVersionMetric.Set(want)
				}
				if got := testutil.ToFloat64(latestVersionMetric); got != want {
					t.Errorf(
						"%s latestVersionMetric mismatch\ngot:  %f\nwant: %f",
						prefix, got, want,
					)
				}
				// 	deployed_version_query_result_last.
				if tc.nilDeployedVersion || !active {
					want = 0
				} else {
					deployedVersionMetric.Set(want)
				}
				if got := testutil.ToFloat64(deployedVersionMetric); got != want {
					t.Errorf(
						"%s deployedVersionMetric mismatch\ngot:  %f\nwant: %f",
						prefix, got, want,
					)
				}
				want = oldWant
				// 	command_result_total.
				if tc.nilCommand || !active {
					want = 0
				} else {
					commandMetric.Add(want)
				}
				if got := testutil.ToFloat64(commandMetric); got != want {
					t.Errorf(
						"%s commandMetric mismatch\ngot:  %f\nwant: %f",
						prefix, got, want,
					)
				}
				want = oldWant
				// 	notify_result_total.
				if tc.nilNotify || !active {
					want = 0
				} else {
					notifyMetric.Add(want)
				}
				if got := testutil.ToFloat64(notifyMetric); got != want {
					t.Errorf(
						"%s notifyMetric mismatch\ngot:  %f\nwant: %f",
						prefix, got, want,
					)
				}
				want = oldWant
				// 	webhook_result_total.
				if tc.nilWebHook || !active {
					want = 0
				} else {
					webhookMetric.Add(want)
				}
				if got := testutil.ToFloat64(webhookMetric); got != want {
					t.Errorf(
						"%s webhookMetric mismatch\ngot:  %f\nwant: %f",
						prefix, got, want,
					)
				}
				// 	service_count_current (active=true).
				want = initialServiceCountCurrentActive
				if active {
					want++
				}
				if got := testutil.ToFloat64(serviceCountCurrentActive); got != want {
					t.Errorf(
						"%s ServiceCountCurrent (active=true) mismatch\ngot:  %f\nwant: %f",
						prefix, got, want,
					)
				}
				// 	service_count_current (active=false).
				want = initialServiceCountCurrentInactive
				if !active {
					want++
				}
				if got := testutil.ToFloat64(serviceCountCurrentInactive); got != want {
					t.Errorf(
						"%s\nServiceCountCurrent (active=false) mismatch\ngot:  %f\nwant: %f",
						prefix, got, want,
					)
				}
				want = oldWant

				// ------------------------------------

				// WHEN: deleteMetrics is called on it.
				svc.deleteMetrics()

				// metrics:
				// 	latest_version_is_deployed.
				latestVersionMetric = metric.LatestVersionIsDeployed.WithLabelValues(
					svc.ID,
				)
				// 	deployed_version_query_result_last.
				deployedVersionMetric = metric.DeployedVersionQueryResultLast.WithLabelValues(
					svc.ID, deployedVersionType,
				)
				// 	command_result_total.
				commandMetric = metric.CommandResultTotal.WithLabelValues(
					testCommand.String(), metric.ActionResultSuccess, svc.ID,
				)
				// 	notify_result_total.
				notifyMetric = metric.NotifyResultTotal.WithLabelValues(
					testNotify.ID, metric.ActionResultSuccess, svc.ID, testNotify.GetType(),
				)
				// 	webhook_result_total.
				webhookMetric = metric.WebHookResultTotal.WithLabelValues(
					testWebHook.ID, metric.ActionResultSuccess, svc.ID,
				)

				prefix = fmt.Sprintf("%s\nService.deleteMetrics()", packageName)

				// THEN: the metrics are deleted:
				want = 0
				fieldTests := []test.FieldAssertion{
					// 	latest_version_is_deployed.
					{Name: "latestVersionMetric", Got: testutil.ToFloat64(latestVersionMetric), Want: want, Mode: test.CompareEqual},
					// 	deployed_version_query_result_last.
					{Name: "deployedVersionMetric", Got: testutil.ToFloat64(deployedVersionMetric), Want: want, Mode: test.CompareEqual},
					// 	command_result_total.
					{Name: "commandMetric", Got: testutil.ToFloat64(commandMetric), Want: want, Mode: test.CompareEqual},
					// 	notify_result_total.
					{Name: "notifyMetric", Got: testutil.ToFloat64(notifyMetric), Want: want, Mode: test.CompareEqual},
					// 	webhook_result_total.
					{Name: "webhookMetric", Got: testutil.ToFloat64(webhookMetric), Want: want, Mode: test.CompareEqual},
					// 	service_count_current.
					{Name: "ServiceCountCurrent (active=true)", Got: testutil.ToFloat64(serviceCountCurrentActive), Want: initialServiceCountCurrentActive, Mode: test.CompareEqual},
					{Name: "ServiceCountCurrent (active=false)", Got: testutil.ToFloat64(serviceCountCurrentInactive), Want: initialServiceCountCurrentInactive, Mode: test.CompareEqual},
				}
				if err := test.AssertFields(t, fieldTests, prefix, ""); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
