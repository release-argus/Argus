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
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrr_test "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service/dashboard"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	metric "github.com/release-argus/Argus/web/metric"
	"github.com/release-argus/Argus/webhook"
	webhook_test "github.com/release-argus/Argus/webhook/test"
)

func TestService_IconURL(t *testing.T) {
	nilValue := "<nil>"
	// GIVEN a Lookup.
	tests := map[string]struct {
		dashboardIcon string
		want          string
		notify        shoutrrr.Slice
	}{
		"no dashboard.icon": {
			want:          nilValue,
			dashboardIcon: "",
		},
		"no icon anywhere": {
			want:          nilValue,
			dashboardIcon: "",
			notify: shoutrrr.Slice{"test": {
				Main:         &shoutrrr.Defaults{},
				Defaults:     &shoutrrr.Defaults{},
				HardDefaults: &shoutrrr.Defaults{},
			}},
		},
		"emoji icon": {
			want:          nilValue,
			dashboardIcon: ":smile:",
		},
		"web icon": {
			want:          "https://example.com/icon.png",
			dashboardIcon: "https://example.com/icon.png",
		},
		"notify icon only": {
			want: "https://example.com/icon.png",
			notify: shoutrrr.Slice{"test": shoutrrr.New(
				nil, "", "",
				nil, nil,
				map[string]string{
					"icon": "https://example.com/icon.png"},
				&shoutrrr.Defaults{},
				&shoutrrr.Defaults{},
				&shoutrrr.Defaults{})},
		},
		"notify icon takes precedence over emoji": {
			want:          "https://example.com/icon.png",
			dashboardIcon: ":smile:",
			notify: shoutrrr.Slice{"test": shoutrrr.New(
				nil, "", "",
				nil, nil,
				map[string]string{
					"icon": "https://example.com/icon.png"},
				&shoutrrr.Defaults{},
				&shoutrrr.Defaults{},
				&shoutrrr.Defaults{})},
		},
		"dashboard icon takes precedence over notify icon": {
			want:          "https://root.com/icon.png",
			dashboardIcon: "https://root.com/icon.png",
			notify: shoutrrr.Slice{"test": shoutrrr.New(
				nil, "", "",
				nil, nil,
				map[string]string{
					"icon": "https://example.com/icon.png"},
				&shoutrrr.Defaults{},
				&shoutrrr.Defaults{},
				&shoutrrr.Defaults{})},
		},
	}

	for name, tc := range tests {
		svc := testService(t, name, "github")

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc.Dashboard.Icon = tc.dashboardIcon
			svc.Notify = tc.notify

			// WHEN IconURL is called.
			got := svc.IconURL()

			// THEN the function returns the correct result.
			gotStr := util.DereferenceOrValue(got, nilValue)
			if gotStr != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, gotStr)
			}
		})
	}
}

func TestService_Init(t *testing.T) {
	// GIVEN a Service.
	tests := map[string]struct {
		svc      *Service
		defaults *Defaults
		wantIcon string
	}{
		"bare service - Name defaulted to ID": {
			svc: &Service{
				ID: "Init",
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New("github",
						"yaml", test.TrimYAML(`
							url: release-argus/Argus
						`),
						nil,
						nil,
						nil, nil)
				}),
			},
		},
		"service with Name": {
			svc: &Service{
				ID:   "Init",
				Name: "other-name",
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New("github",
						"yaml", test.TrimYAML(`
							url: release-argus/Argus
						`),
						nil,
						nil,
						nil, nil)
				}),
			},
		},
		"service with notify - doesn't set fallback when Service has a Dashboard.Icon": {
			svc: &Service{
				ID: "Init",
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New("github",
						"yaml", test.TrimYAML(`
								url: release-argus/Argus
							`),
						nil,
						nil,
						nil, nil)
				}),
				Notify: shoutrrr.Slice{
					"test": shoutrrr.New(
						nil, "", "discord",
						nil, nil,
						map[string]string{
							"icon": "notify-icon"},
						nil, nil, nil)},
				Dashboard: *dashboard.NewOptions(
					nil,
					"dashboard-icon", "",
					"", nil,
					nil, nil)},
			wantIcon: "dashboard-icon",
		},
		"service with notify - does set fallback when Service has no Dashboard.Icon": {
			svc: &Service{
				ID: "Init",
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New("github",
						"yaml", test.TrimYAML(`
								url: release-argus/Argus
							`),
						nil,
						nil,
						nil, nil)
				}),
				Notify: shoutrrr.Slice{
					"baz": nil,
					"foo": shoutrrr.New(
						nil, "", "discord",
						nil, nil,
						map[string]string{
							"icon": "example.com/notify-icon-1"},
						nil, nil, nil),
					"bar": shoutrrr.New(
						nil, "", "discord",
						nil, nil,
						map[string]string{
							"icon": "https://example.com/notify-icon-2"},
						nil, nil, nil)},
				Dashboard: *dashboard.NewOptions(
					nil,
					"", "",
					"", nil,
					nil, nil)},
			wantIcon: "https://example.com/notify-icon-2",
		},
		"service with notify, command and webhook": {
			svc: &Service{
				ID: "Init",
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New("github",
						"yaml", test.TrimYAML(`
							url: release-argus/Argus
						`),
						nil,
						nil,
						nil, nil)
				}),
				Notify: shoutrrr.Slice{
					"test": shoutrrr.New(
						nil, "", "discord",
						nil, nil, nil,
						nil, nil, nil)},
				Command: command.Slice{
					{"ls"}},
				WebHook: webhook.Slice{
					"test": webhook_test.WebHook(false, false, false)}},
		},
		"service with notifies from defaults": {
			svc: &Service{
				ID: "Init",
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New("github",
						"yaml", test.TrimYAML(`
							url: release-argus/Argus
						`),
						nil,
						nil,
						nil, nil)
				})},
			defaults: &Defaults{
				Notify: map[string]struct{}{
					"foo": {}}},
		},
		"service with notifies not from defaults": {
			svc: &Service{
				ID: "Init",
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New("github",
						"yaml", test.TrimYAML(`
							url: release-argus/Argus
						`),
						nil,
						nil,
						nil, nil)
				}),
				Notify: shoutrrr.Slice{
					"test": &shoutrrr.Shoutrrr{}}},
			defaults: &Defaults{
				Notify: map[string]struct{}{
					"foo": {}}},
		},
		"service with commands from defaults": {
			svc: &Service{
				ID: "Init",
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New("github",
						"yaml", test.TrimYAML(`
							url: release-argus/Argus
						`),
						nil,
						nil,
						nil, nil)
				})},
			defaults: &Defaults{
				Command: command.Slice{
					{"ls"}}},
		},
		"service with commands not from defaults": {
			svc: &Service{
				ID: "Init",
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New("github",
						"yaml", test.TrimYAML(`
							url: release-argus/Argus
						`),
						nil,
						nil,
						nil, nil)
				}),
				Command: command.Slice{
					{"test"}}},
			defaults: &Defaults{
				Command: command.Slice{
					{"ls"}}},
		},
		"service with webhooks from defaults": {
			svc: &Service{
				ID: "Init",
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New("github",
						"yaml", test.TrimYAML(`
							url: release-argus/Argus
						`),
						nil,
						nil,
						nil, nil)
				})},
			defaults: &Defaults{
				WebHook: map[string]struct{}{
					"bar": {}}},
		},
		"service with webhooks not from defaults": {
			svc: &Service{
				ID: "Init",
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New("github",
						"yaml", test.TrimYAML(`
							url: release-argus/Argus
						`),
						nil,
						nil,
						nil, nil)
				}),
				WebHook: webhook.Slice{
					"test": webhook_test.WebHook(false, false, false)}},
			defaults: &Defaults{
				WebHook: map[string]struct{}{
					"bar": {}}},
		},
		"service with webhooks/commands from defaults and notify overridden": {
			svc: &Service{
				ID: "Init",
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New("github",
						"yaml", test.TrimYAML(`
							url: release-argus/Argus
						`),
						nil,
						nil,
						nil, nil)
				}),
				Notify: shoutrrr.Slice{
					"test": &shoutrrr.Shoutrrr{}}},
			defaults: &Defaults{
				Notify: map[string]struct{}{
					"foo": {}},
				Command: command.Slice{
					{"ls"}},
				WebHook: map[string]struct{}{
					"bar": {}}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			if tc.defaults == nil {
				tc.defaults = &Defaults{}
			}
			var hardDefaults Defaults
			tc.svc.ID = name
			hadName := tc.svc.Name
			hadNotify := util.SortedKeys(tc.svc.Notify)
			hadWebHook := util.SortedKeys(tc.svc.WebHook)
			hadCommand := make(command.Slice, len(tc.svc.Command))
			copy(hadCommand, tc.svc.Command)

			// WHEN Init is called on it.
			tc.svc.Init(
				tc.defaults, &hardDefaults,
				&shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{},
				&webhook.SliceDefaults{}, &webhook.Defaults{}, &webhook.Defaults{})

			// THEN the Name is set to the ID if not set.
			if (hadName != "" && tc.svc.Name != hadName) || (hadName == "" && tc.svc.Name != tc.svc.ID) {
				t.Errorf("%s\nName mismatch\nwant: %q\ngot:  %q",
					packageName, tc.svc.ID, tc.svc.Name)
			}
			// THEN pointers to those vars are handed out to the Lookup::
			// 	Defaults.
			if tc.svc.Defaults != tc.defaults {
				t.Errorf("%s\nDefaults mismatch\nwant: %v\ngot:  %v",
					packageName, tc.defaults, tc.svc.Defaults)
			}
			// 	Dashboard.Defaults.
			if tc.svc.Dashboard.Defaults != &tc.defaults.Dashboard {
				t.Errorf("%s\nDashboard.Defaults mismatch\nwant: %v\ngot:  %v",
					packageName, &tc.defaults.Dashboard, tc.svc.Dashboard.Defaults)
			}
			// 	Options.defaults.
			if tc.svc.Options.Defaults != &tc.defaults.Options {
				t.Errorf("%s\nOptions.Defaults mismatch\nwant: %v\ngot:  %v",
					packageName, &tc.defaults.Options, tc.svc.Options.Defaults)
			}
			// 	HardDefaults.
			if tc.svc.HardDefaults != &hardDefaults {
				t.Errorf("%s\nHardDefaults mismatch\nwant: %v\ngot:  %v",
					packageName, &hardDefaults, tc.svc.HardDefaults)
			}
			// 	Dashboard.HardDefaults.
			if tc.svc.Dashboard.HardDefaults != &hardDefaults.Dashboard {
				t.Errorf("%s\nDashboard.HardDefaults mismatch\nwant: %v\ngot:  %v",
					packageName, &hardDefaults.Dashboard, tc.svc.Dashboard.HardDefaults)
			}
			// 	Options.HardDefaults.
			if tc.svc.Options.HardDefaults != &hardDefaults.Options {
				t.Errorf("%s\nOptions.HardDefaults mismatch\nwant: %v\ngot:  %v",
					packageName, &hardDefaults.Options, tc.svc.Options.HardDefaults)
			}
			// 	Notify.
			if len(tc.svc.Notify) != 0 {
				for i := range tc.svc.Notify {
					if tc.svc.Notify[i].Main == nil {
						t.Errorf("%s\nNotify init didn't initialise the Main",
							packageName)
					}
				}
			}
			// 		Notifiers are not overridden if non-empty.
			if len(hadNotify) != 0 && len(tc.svc.Notify) != len(hadNotify) {
				t.Fatalf("%s\nNotify length mismatch\nwant: %d (%v)\ngot:  %d (%v)",
					packageName,
					len(hadNotify), hadNotify,
					len(tc.svc.Notify), util.SortedKeys(tc.svc.Notify))
			}
			wantNotify := hadNotify
			if len(hadNotify) == 0 && tc.defaults != nil {
				wantNotify = make([]string, len(tc.defaults.Notify))
				wantNotify = util.SortedKeys(tc.defaults.Notify)
			}
			for _, i := range wantNotify {
				if tc.svc.Notify[i] == nil {
					t.Errorf("%s - Notify[%s] was nil",
						packageName, i)
				}
			}
			// 	Command.
			if len(tc.svc.Command) != 0 {
				if tc.svc.CommandController == nil {
					t.Errorf("%s\nCommandController is still nil with %v Commands present",
						packageName, tc.svc.Command)
				}
			} else if tc.svc.CommandController != nil {
				t.Errorf("%s\nCommandController should be nil with %v Commands present",
					packageName, tc.svc.Command)
			}
			// 		Command is not overridden if non-empty.
			if len(hadCommand) != 0 && len(tc.svc.Command) != len(hadCommand) {
				t.Fatalf("%s\nCommand length changed\nwant: %d (%v)\ngot:  %d (%v)",
					packageName,
					len(hadCommand), hadCommand,
					len(tc.svc.Command), tc.svc.Command)
			}
			wantCommand := hadCommand
			if len(hadCommand) == 0 && tc.defaults != nil {
				wantCommand = make(command.Slice, len(tc.defaults.Command))
				wantCommand = tc.defaults.Command
			}
			for i := range wantCommand {
				if tc.svc.Command[i].String() != wantCommand[i].String() {
					t.Errorf("%s - Command[%d] changed\nwant: %q\ngot:  %q",
						packageName, i,
						wantCommand[i].String(), tc.svc.Command[i].String())
				}
			}
			// 	WebHook.
			if len(tc.svc.WebHook) != 0 {
				for i := range tc.svc.WebHook {
					if tc.svc.WebHook[i].Main == nil {
						t.Errorf("%s\nWebHook init didn't initialise the Main",
							packageName)
					}
				}
			}
			// 		WebHooks are not overridden if non-empty.
			if len(hadWebHook) != 0 && len(tc.svc.WebHook) != len(hadWebHook) {
				t.Fatalf("%s\nWebHook length changed\nwant: %d (%v)\ngot:  %d (%v)",
					packageName,
					len(hadWebHook), hadWebHook,
					len(tc.svc.WebHook), util.SortedKeys(tc.svc.WebHook))
			}
			wantWebHook := hadWebHook
			if len(hadWebHook) == 0 && tc.defaults != nil {
				wantWebHook = make([]string, len(tc.defaults.WebHook))
				wantWebHook = util.SortedKeys(tc.defaults.WebHook)
			}
			for _, i := range wantWebHook {
				if tc.svc.WebHook[i] == nil {
					t.Errorf("%s - hadWebHook[%s] was nil",
						packageName, i)
				}
			}
			// 	Dashboard
			gotIcon := tc.svc.Dashboard.GetIcon()
			if tc.wantIcon != gotIcon {
				t.Errorf("%s\nDashboard icon fallback mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wantIcon, gotIcon)
			}
		})
	}
}

func TestService_InitMetrics_ResetMetrics_DeleteMetrics(t *testing.T) {
	// GIVEN a Service.
	tests := map[string]struct {
		nilDeployedVersion bool
		nilCommand         bool
		nilNotify          bool
		nilWebHook         bool
	}{
		"all defined": {},
		"nil DeployedVersionLookup": {
			nilDeployedVersion: true},
		"nil Command": {
			nilCommand: true},
		"nil Notify": {
			nilNotify: true},
		"nil WebHook": {
			nilWebHook: true},
		"nil all": {
			nilDeployedVersion: true,
			nilCommand:         true,
			nilNotify:          true,
			nilWebHook:         true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're testing metrics.

			testCommand := command.Command{"ls"}
			testNotify := shoutrrr_test.Shoutrrr(false, false)
			testWebHook := webhook_test.WebHook(false, false, false)
			service := &Service{
				ID:                    "TestService_InitMetrics_ResetMetrics_DeleteMetrics--" + name,
				LatestVersion:         testLatestVersion(t, "github", false),
				DeployedVersionLookup: testDeployedVersionLookup(t, false),
				Command: command.Slice{
					testCommand},
				Notify: shoutrrr.Slice{
					testNotify.ID: testNotify},
				WebHook: webhook.Slice{
					testWebHook.ID: testWebHook},
			}

			// Init the service.
			service.Init(
				&Defaults{}, &Defaults{},
				&shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{},
				&webhook.SliceDefaults{}, &webhook.Defaults{}, &webhook.Defaults{},
			)
			service.Status.SetLatestVersion("0.0.2", "", false)
			service.Status.SetDeployedVersion("0.0.2", "", false)

			// nil the vars.
			var deployedVersionType string
			if tc.nilDeployedVersion {
				service.DeployedVersionLookup = nil
			} else {
				deployedVersionType = service.DeployedVersionLookup.GetType()
			}
			if tc.nilCommand {
				service.Command = nil
			}
			if tc.nilNotify {
				service.Notify = nil
			}
			if tc.nilWebHook {
				service.WebHook = nil
			}

			// metrics:
			// 	latest_version_query_result_total.
			latestVersionMetric := metric.LatestVersionIsDeployed.WithLabelValues(
				service.ID)
			// 	deployed_version_query_result_last.
			deployedVersionMetric := metric.DeployedVersionQueryResultLast.WithLabelValues(
				service.ID, deployedVersionType)
			// 	command_result_total.
			commandMetric := metric.CommandResultTotal.WithLabelValues(
				testCommand.String(), "SUCCESS", service.ID)
			// 	notify_result_total.
			notifyMetric := metric.NotifyResultTotal.WithLabelValues(
				testNotify.ID, "SUCCESS", service.ID, testNotify.GetType())
			// 	webhook_result_total.
			webhookMetric := metric.WebHookResultTotal.WithLabelValues(
				testWebHook.ID, "SUCCESS", service.ID)
			// 	service_count_total.
			serviceCountTotal := metric.ServiceCountCurrent
			initialServiceCountCurrent := testutil.ToFloat64(serviceCountTotal)

			// #################################
			// WHEN initMetrics is called on it.
			// #################################
			service.initMetrics()

			// THEN the metrics are created:
			want := float64(3)
			oldWant := want
			// 	latest_version_is_deployed.
			latestVersionMetric.Set(want)
			if got := testutil.ToFloat64(latestVersionMetric); got != want {
				t.Errorf("%s\nlatestVersionMetric mismatch after initMetrics()\nwant: %f\ngot:  %f",
					packageName, want, got)
			}
			// 	deployed_version_query_result_last.
			if tc.nilDeployedVersion {
				want = 0
			} else {
				deployedVersionMetric.Set(want)
			}
			if got := testutil.ToFloat64(deployedVersionMetric); got != want {
				t.Errorf("%s\ndeployedVersionMetric mismatch after initMetrics()\nwant: %f\ngot:  %f",
					packageName, want, got)
			}
			want = oldWant
			// 	command_result_total.
			if tc.nilCommand {
				want = 0
			} else {
				commandMetric.Add(want)
			}
			if got := testutil.ToFloat64(commandMetric); got != want {
				t.Errorf("%s\ncommandMetric mismatch after initMetrics()\nwant: %f\ngot:  %f",
					packageName, want, got)
			}
			want = oldWant
			// 	notify_result_total.
			if tc.nilNotify {
				want = 0
			} else {
				notifyMetric.Add(want)
			}
			if got := testutil.ToFloat64(notifyMetric); got != want {
				t.Errorf("%s\nnotifyMetric mismatch after initMetrics()\nwant: %f\ngot:  %f",
					packageName, want, got)
			}
			// 	webhook_result_total.
			want = oldWant
			if tc.nilWebHook {
				want = 0
			} else {
				webhookMetric.Add(want)
			}
			if got := testutil.ToFloat64(webhookMetric); got != want {
				t.Errorf("%s\nwebhookMetric mismatch after initMetrics()\nwant: %f\ngot:  %f",
					packageName, want, got)
			}
			want = oldWant
			// 	service_count_total.
			wantServiceCountCurrent := initialServiceCountCurrent + 1
			if got := testutil.ToFloat64(serviceCountTotal); got != wantServiceCountCurrent {
				t.Errorf("%s\nServiceCountCurrent mismatch after initMetrics()\nwant: %f\ngot:  %f",
					packageName, wantServiceCountCurrent, got)
			}

			// ###################################
			// WHEN deleteMetrics is called on it.
			// ###################################
			service.deleteMetrics()

			// metrics:
			// 	latest_version_is_deployed.
			latestVersionMetric = metric.LatestVersionIsDeployed.WithLabelValues(
				service.ID)
			// 	deployed_version_query_result_last.
			deployedVersionMetric = metric.DeployedVersionQueryResultLast.WithLabelValues(
				service.ID, deployedVersionType)
			// 	command_result_total.
			commandMetric = metric.CommandResultTotal.WithLabelValues(
				testCommand.String(), "SUCCESS", service.ID)
			// 	notify_result_total.
			notifyMetric = metric.NotifyResultTotal.WithLabelValues(
				testNotify.ID, "SUCCESS", service.ID, testNotify.GetType())
			// 	webhook_result_total.
			webhookMetric = metric.WebHookResultTotal.WithLabelValues(
				testWebHook.ID, "SUCCESS", service.ID)

			// THEN the metrics are deleted:
			want = 0
			// 	latest_version_is_deployed.
			if got := testutil.ToFloat64(latestVersionMetric); got != want {
				t.Errorf("%s\nlatestVersionMetric mismatch after deleteMetrics()\nwant: %f\ngot:  %f",
					packageName, want, got)
			}
			// 	deployed_version_query_result_last.
			if got := testutil.ToFloat64(deployedVersionMetric); got != want {
				t.Errorf("%s\ndeployedVersionMetric mismatch after deleteMetrics()\nwant: %f\ngot:  %f",
					packageName, want, got)
			}
			// 	command_result_total.
			if got := testutil.ToFloat64(commandMetric); got != want {
				t.Errorf("%s\ncommandMetric mismatch after deleteMetrics()\nwant: %f\ngot:  %f",
					packageName, want, got)
			}
			// 	notify_result_total.
			if got := testutil.ToFloat64(notifyMetric); got != want {
				t.Errorf("%s\nnotifyMetric mismatch after deleteMetrics()\nwant: %f\ngot:  %f",
					packageName, want, got)
			}
			// 	webhook_result_total.
			if got := testutil.ToFloat64(webhookMetric); got != want {
				t.Errorf("%s\nwebhookMetric mismatch after deleteMetrics()\nwant: %f\ngot:  %f",
					packageName, want, got)
			}
			// 	service_count_total.
			if got := testutil.ToFloat64(serviceCountTotal); got != initialServiceCountCurrent {
				t.Errorf("%s\nServiceCountCurrent mismatch after deleteMetrics()\nwant: %f\ngot:  %f",
					packageName, wantServiceCountCurrent, got)
			}
		})
	}
}
