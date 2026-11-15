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
	nilValue := "nil"

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

			got := svc.IconURL()

			if gotStr := util.DerefOr(got, nilValue); gotStr != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", gotStr, tc.want)
			}
		})
	}
}

func TestService_Init(t *testing.T) {
	svcCfg := plainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig(t)

	tests := []struct {
		name                    string
		svc                     *Service
		wantIcon                string
		wantNotifyFromDefaults  bool
		wantCommandFromDefaults bool
		wantWebhookFromDefaults bool
	}{
		{
			name: "bare/Name defaulted to ID",
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
			name: "bare/Name",
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
			name: "notify/doesn't set fallback when Service has a Dashboard.Icon",
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
			name: "notify/doesn't set fallback from a non-URL icon",
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
					`)),
					"Init",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			wantIcon: "",
		},
		{
			name: "notify/doesn't set fallback from a URL icon when Service has a Dashboard.Icon",
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
									icon: https://example.com/notify-icon
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
			name: "notify/does set fallback when Service has no Dashboard.Icon",
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
			name: "notify + command + webhook",
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
						`+whtest.Webhook(t, false, false, false).String(".   ")+`
					`)),
					"Init",
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name: "notify/from defaults",
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
			wantNotifyFromDefaults: true,
		},
		{
			name: "notify/not from defaults",
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
			name: "command/from defaults",
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
			wantCommandFromDefaults: true,
		},
		{
			name: "command/not from defaults",
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
			name: "webhook/from defaults",
			svc: test.Must(t, func() (*Service, error) {
				svcCfg := plainDefaultsConfig()
				svcCfg.Soft.Webhook = map[string]struct{}{
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
			wantWebhookFromDefaults: true,
		},
		{
			name: "webhook/not from defaults",
			svc: test.Must(t, func() (*Service, error) {
				svcCfg := plainDefaultsConfig()
				svcCfg.Soft.Webhook = map[string]struct{}{
					"bar": {},
				}

				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
						webhook:
							test:
						`+whtest.Webhook(t, false, false, false).String(".   ")+`
					`)),
					"Init",
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name: "webhook + command - from defaults and notify overridden",
			svc: test.Must(t, func() (*Service, error) {
				svcCfg := plainDefaultsConfig()
				svcCfg.Soft.Notify = map[string]struct{}{
					"foo": {},
				}
				svcCfg.Soft.Command = command.Commands{
					{"ls"},
				}
				svcCfg.Soft.Webhook = map[string]struct{}{
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
			wantCommandFromDefaults: true,
			wantWebhookFromDefaults: true,
		},
		{
			name: "DeployedVersionLookup",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						deployed_version:
							type: url
							url: `+test.LookupPlain.URLValid+`
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
			hadWebhook := util.SortedKeys(s.Webhook)
			hadCommand := make(command.Commands, len(s.Command))
			copy(hadCommand, s.Command)

			announceChannel := make(chan []byte, 1)
			databaseChannel := make(chan dbtype.Message, 1)
			saveChannel := make(chan bool, 1)

			tc.svc.init(
				notifyCfg,
				whCfg,
				announceChannel,
				databaseChannel,
				saveChannel,
			)

			fieldTests := []test.FieldAssertion{
				{Name: "NotifyFromDefaults", Got: tc.svc.NotifyFromDefaults, Want: tc.wantNotifyFromDefaults, Mode: test.CompareEqual},
				{Name: "CommandFromDefaults", Got: tc.svc.CommandFromDefaults, Want: tc.wantCommandFromDefaults, Mode: test.CompareEqual},
				{Name: "WebhookFromDefaults", Got: tc.svc.WebhookFromDefaults, Want: tc.wantWebhookFromDefaults, Mode: test.CompareEqual},
			}
			if testErr := test.AssertFields(t, fieldTests, "Service"); testErr != nil {
				t.Error(testErr)
			}

			hadNotifyLength := len(hadNotify)
			gotNotifyLength := len(tc.svc.Notify)
			if gotNotifyLength != 0 {
				for i := range tc.svc.Notify {
					if tc.svc.Notify[i].Main == nil {
						t.Errorf("Service.Notify[%q].Main mismatch\ngot:  nil\nwant: non-nil", i)
					}
					if tc.svc.Notify[i].Defaults == nil {
						t.Errorf(
							"Service.Notify[%q].Defaults mismatch\ngot:  nil\nwant: non-nil",
							i,
						)
					}
					if tc.svc.Notify[i].HardDefaults == nil {
						t.Errorf(
							"Service.Notify[%q].HardDefaults mismatch\ngot:  nil\nwant: non-nil",
							i,
						)
					}
				}
			}

			if hadNotifyLength != 0 && gotNotifyLength != hadNotifyLength {
				t.Fatalf(
					"Service.Notify length mismatch\ngot:  %d (%q)\nwant: %d (%q)",
					gotNotifyLength, util.SortedKeys(tc.svc.Notify),
					hadNotifyLength, hadNotify,
				)
			}

			wantNotify := hadNotify
			if defaultNotifiers := tc.svc.Defaults.Notify; defaultNotifiers != nil && hadNotifyLength == 0 {
				wantNotify = util.SortedKeys(defaultNotifiers)
			}
			for _, i := range wantNotify {
				if tc.svc.Notify[i] == nil {
					t.Errorf("Service.Notify[%q] mismatch\ngot:  nil\nwant: non-nil", i)
				}
			}

			hadCommandLength := len(hadCommand)
			gotCommandLength := len(tc.svc.Command)
			if gotCommandLength != 0 {
				if tc.svc.CommandController == nil {
					t.Error("Service.CommandController mismatch\ngot:  nil\nwant: non-nil")
				}
			} else if tc.svc.CommandController != nil {
				t.Error("Service.CommandController mismatch\ngot:  non-nil\nwant: nil")
			}

			if hadCommandLength != 0 && gotCommandLength != hadCommandLength {
				t.Fatalf(
					"Service.Command length mismatch\ngot:  %d (%v)\nwant: %d (%v)",
					gotCommandLength, tc.svc.Command,
					hadCommandLength, hadCommand,
				)
			}

			wantCommand := hadCommand
			if defaultCommands := tc.svc.Defaults.Command; defaultCommands != nil && hadCommandLength == 0 {
				wantCommand = defaultCommands
			}
			for i := range wantCommand {
				got := tc.svc.Command[i].String()
				want := wantCommand[i].String()
				if got != want {
					t.Errorf(
						"Service.Command[%d].String() mismatch\ngot:  %q\nwant: %q",
						i, got, want,
					)
				}
			}

			hadWebhookLength := len(hadWebhook)
			gotWebhookLength := len(tc.svc.Webhook)
			if gotWebhookLength != 0 {
				for i := range tc.svc.Webhook {
					if tc.svc.Webhook[i].Main == nil {
						t.Errorf("Service.Webhook[%q].Main mismatch\ngot:  nil\nwant: non-nil", i)
					}
					if tc.svc.Webhook[i].Defaults == nil {
						t.Errorf(
							"Service.Webhook[%q].Defaults mismatch\ngot:  nil\nwant: non-nil",
							i,
						)
					}
					if tc.svc.Webhook[i].HardDefaults == nil {
						t.Errorf(
							"Service.Webhook[%q].HardDefaults mismatch\ngot:  nil\nwant: non-nil",
							i,
						)
					}
				}
			}

			if hadWebhookLength > 0 && gotWebhookLength != hadWebhookLength {
				t.Fatalf(
					"Service.Webhook length mismatch\ngot:  %d (%q)\nwant: %d (%q)",
					gotWebhookLength, util.SortedKeys(tc.svc.Webhook),
					hadWebhookLength, hadWebhook,
				)
			}

			wantWebhook := hadWebhook
			if defaultWebhooks := tc.svc.Defaults.Webhook; defaultWebhooks != nil && hadWebhookLength == 0 {
				wantWebhook = util.SortedKeys(defaultWebhooks)
			}
			for _, i := range wantWebhook {
				if tc.svc.Webhook[i] == nil {
					t.Errorf("Service.Webhook[%q] mismatch\ngot:  nil\nwant: non-nil", i)
				}
			}

			gotIcon := tc.svc.Dashboard.GetIcon()
			if tc.wantIcon != gotIcon {
				t.Errorf(
					"Service.Dashboard.GetIcon() mismatch\ngot:  %q\nwant: %q",
					gotIcon, tc.wantIcon,
				)
			}

			fieldTests = []test.FieldAssertion{
				{Name: "AnnounceChannel", Got: tc.svc.Status.AnnounceChannel, Want: announceChannel, Mode: test.CompareSamePointer},
				{Name: "DatabaseChannel", Got: tc.svc.Status.DatabaseChannel, Want: databaseChannel, Mode: test.CompareSamePointer},
				{Name: "SaveChannel", Got: tc.svc.Status.SaveChannel, Want: saveChannel, Mode: test.CompareSamePointer},
				{Name: "Dashboard", Got: tc.svc.Status.Dashboard, Want: &tc.svc.Dashboard, Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, fieldTests, "Status"); testErr != nil {
				t.Fatal(testErr)
			}

			serviceInfo := tc.svc.Status.GetServiceInfo()
			wantServiceInfoName := util.ValueOr(tc.svc.Name, tc.svc.ID)
			fieldTests = []test.FieldAssertion{
				{Name: "ID", Got: serviceInfo.ID, Want: tc.svc.ID, Mode: test.CompareEqual},
				{Name: "Name", Got: serviceInfo.Name, Want: wantServiceInfoName, Mode: test.CompareEqual},
				{Name: "Comment", Got: serviceInfo.Comment, Want: tc.svc.Comment, Mode: test.CompareEqual},
				{Name: "Icon", Got: serviceInfo.Icon, Want: tc.wantIcon, Mode: test.CompareEqual},
			}
			if testErr := test.AssertFields(t, fieldTests, "Status.ServiceInfo"); testErr != nil {
				t.Fatal(testErr)
			}
			if !util.AreSlicesEqual(serviceInfo.Tags, tc.svc.Dashboard.Tags) {
				t.Errorf(
					"Status.ServiceInfo.Tags mismatch\ngot:  %q\nwant: %q",
					serviceInfo.Tags, tc.svc.Dashboard.Tags,
				)
			}

			fieldTests = []test.FieldAssertion{
				{Name: "Command", Got: tc.svc.Status.Fails.Command.Length(), Want: len(tc.svc.Command), Mode: test.CompareEqual},
				{Name: "Shoutrrr", Got: tc.svc.Status.Fails.Shoutrrr.Length(), Want: len(tc.svc.Notify), Mode: test.CompareEqual},
				{Name: "Webhook", Got: tc.svc.Status.Fails.Webhook.Length(), Want: len(tc.svc.Webhook), Mode: test.CompareEqual},
			}
			if testErr := test.AssertFields(t, fieldTests, "Status.Fails"); testErr != nil {
				t.Fatal(testErr)
			}

			if dvl := tc.svc.DeployedVersionLookup; dvl != nil && dvl.GetStatus() != &tc.svc.Status {
				t.Errorf(
					"Service.DeployedVersionLookup.GetStatus() mismatch\ngot:  %p\nwant: %p",
					dvl.GetStatus(), &tc.svc.Status,
				)
			}

			// A Notify icon is kept as a fallback only when Dashboard.Icon is unset.
			wantFallbackIcon := tc.wantIcon
			if tc.svc.Dashboard.Icon != "" {
				wantFallbackIcon = ""
			}
			tc.svc.Dashboard.Icon = ""
			if got := tc.svc.Dashboard.GetIcon(); got != wantFallbackIcon {
				t.Errorf(
					"without Icon: Service.Dashboard.GetIcon() mismatch\ngot:  %q\nwant: %q",
					got, wantFallbackIcon,
				)
			}
		})
	}
}

func TestService_InitMetrics_then_DeleteMetrics(t *testing.T) {
	svcCfg := plainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig(t)

	tests := []struct {
		name               string
		nilDeployedVersion bool
		nilCommand         bool
		nilNotify          bool
		nilWebhook         bool
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
			name:       "nil Webhook",
			nilWebhook: true,
		},
		{
			name:               "nil all",
			nilDeployedVersion: true,
			nilCommand:         true,
			nilNotify:          true,
			nilWebhook:         true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: shares the LatestVersion and DeployedVersion metrics.

			activeStates := []bool{true, false}
			for _, active := range activeStates {
				t.Run("active="+strconv.FormatBool(active), func(t *testing.T) {
					// Serial: shares metric.ServiceCountCurrent and metric.LatestVersionQueryResultTotal.

					testCommand := command.Command{"ls"}
					testNotify := shoutrrrtest.Shoutrrr(t, false, false)
					testWebhook := whtest.Webhook(t, false, false, false)
					svc := test.Must(t, func() (s *Service, err error) {
						return DecodeService(
							"yaml", []byte(test.TrimYAML(`
								options:
									active: `+strconv.FormatBool(active)+`
								latest_version:
								`+lvtest.Lookup(t, "url", false).String("  ")+`
								deployed_version:
								`+dvtest.Lookup(t, "url", false, "").String("  ")+`
								command:
								- `+testCommand.JSON()+`
								notify:
									`+testNotify.ID+`:
								`+testNotify.String("    ")+`
								webhook:
									`+testWebhook.ID+`:
								`+testWebhook.String("    ")+`
							`)),
							"TestService_InitMetrics_then_DeleteMetrics--"+tc.name,
							svcCfg,
							notifyCfg,
							whCfg,
						)
					})

					version := "0.0.2"
					svc.Status.SetLatestVersion(version, "", false)
					svc.Status.SetDeployedVersion(version, "", false)

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
					if tc.nilWebhook {
						svc.Webhook = nil
					}

					latestVersionMetric := metric.LatestVersionIsDeployed.WithLabelValues(svc.ID)
					latestVersionType := svc.LatestVersion.GetType()
					latestVersionQueryMetric := metric.LatestVersionQueryResultLast.WithLabelValues(
						svc.ID, latestVersionType,
					)
					deployedVersionMetric := metric.DeployedVersionQueryResultLast.WithLabelValues(
						svc.ID, deployedVersionType,
					)
					commandMetric := metric.CommandResultTotal.WithLabelValues(
						testCommand.String(), metric.ActionResultSuccess, svc.ID,
					)
					notifyMetric := metric.NotifyResultTotal.WithLabelValues(
						testNotify.ID, metric.ActionResultSuccess, svc.ID, testNotify.GetType(),
					)
					webhookMetric := metric.WebhookResultTotal.WithLabelValues(
						testWebhook.ID, metric.ActionResultSuccess, svc.ID,
					)
					serviceCountCurrentActive := metric.ServiceCountCurrent.WithLabelValues(
						metric.ServiceStateActive,
					)
					initialServiceCountCurrentActive := testutil.ToFloat64(serviceCountCurrentActive)
					serviceCountCurrentInactive := metric.ServiceCountCurrent.WithLabelValues(
						metric.ServiceStateInactive,
					)
					initialServiceCountCurrentInactive := testutil.ToFloat64(serviceCountCurrentInactive)
					initialLatestVersionQueryTotals := testutil.CollectAndCount(
						metric.LatestVersionQueryResultTotal,
					)

					svc.initMetrics()

					wantIsDeployed := float64(0)
					if active {
						wantIsDeployed = float64(metric.LatestVersionDeployed)
					}
					if got := testutil.ToFloat64(latestVersionMetric); got != wantIsDeployed {
						t.Errorf(
							"after initMetrics: metric.LatestVersionIsDeployed mismatch\n"+
								"got:  %f\nwant: %f",
							got, wantIsDeployed,
						)
					}
					wantLatestVersionQueryTotals := initialLatestVersionQueryTotals
					if active {
						wantLatestVersionQueryTotals += 2
					}
					test.AssertMetricCount(
						t, "Service.initMetrics() LatestVersionQueryResultTotal",
						metric.LatestVersionQueryResultTotal, wantLatestVersionQueryTotals,
					)
					fieldTests := []test.FieldAssertion{
						{Name: "LatestVersionQueryResultLast", Got: testutil.ToFloat64(latestVersionQueryMetric), Want: float64(0), Mode: test.CompareEqual},
						{Name: "DeployedVersionQueryResultLast", Got: testutil.ToFloat64(deployedVersionMetric), Want: float64(0), Mode: test.CompareEqual},
						{Name: "CommandResultTotal", Got: testutil.ToFloat64(commandMetric), Want: float64(0), Mode: test.CompareEqual},
						{Name: "NotifyResultTotal", Got: testutil.ToFloat64(notifyMetric), Want: float64(0), Mode: test.CompareEqual},
						{Name: "WebhookResultTotal", Got: testutil.ToFloat64(webhookMetric), Want: float64(0), Mode: test.CompareEqual},
					}
					if testErr := test.AssertFields(t, fieldTests, "metric"); testErr != nil {
						t.Errorf("after initMetrics: %v", testErr)
					}
					wantActive, wantInactive := initialServiceCountCurrentActive, initialServiceCountCurrentInactive
					if active {
						wantActive++
					} else {
						wantInactive++
					}
					fieldTests = []test.FieldAssertion{
						{Name: "ServiceCountCurrent(active)", Got: testutil.ToFloat64(serviceCountCurrentActive), Want: wantActive, Mode: test.CompareEqual},
						{Name: "ServiceCountCurrent(inactive)", Got: testutil.ToFloat64(serviceCountCurrentInactive), Want: wantInactive, Mode: test.CompareEqual},
					}
					if testErr := test.AssertFields(t, fieldTests, "metric"); testErr != nil {
						t.Errorf("after initMetrics: %v", testErr)
					}

					// Give deleteMetrics values to clear.
					if active {
						latestVersionMetric.Set(3)
						latestVersionQueryMetric.Set(3)
						if !tc.nilDeployedVersion {
							deployedVersionMetric.Set(3)
						}
						if !tc.nilCommand {
							commandMetric.Add(3)
						}
						if !tc.nilNotify {
							notifyMetric.Add(3)
						}
						if !tc.nilWebhook {
							webhookMetric.Add(3)
						}
					}

					// ------------------------------------

					svc.deleteMetrics()

					latestVersionMetric = metric.LatestVersionIsDeployed.WithLabelValues(svc.ID)
					latestVersionQueryMetric = metric.LatestVersionQueryResultLast.WithLabelValues(
						svc.ID, latestVersionType,
					)
					deployedVersionMetric = metric.DeployedVersionQueryResultLast.WithLabelValues(
						svc.ID, deployedVersionType,
					)
					commandMetric = metric.CommandResultTotal.WithLabelValues(
						testCommand.String(), metric.ActionResultSuccess, svc.ID,
					)
					notifyMetric = metric.NotifyResultTotal.WithLabelValues(
						testNotify.ID, metric.ActionResultSuccess, svc.ID, testNotify.GetType(),
					)
					webhookMetric = metric.WebhookResultTotal.WithLabelValues(
						testWebhook.ID, metric.ActionResultSuccess, svc.ID,
					)

					fieldTests = []test.FieldAssertion{
						{Name: "ServiceCountCurrent(active)", Got: testutil.ToFloat64(serviceCountCurrentActive), Want: initialServiceCountCurrentActive, Mode: test.CompareEqual},
						{Name: "ServiceCountCurrent(inactive)", Got: testutil.ToFloat64(serviceCountCurrentInactive), Want: initialServiceCountCurrentInactive, Mode: test.CompareEqual},
					}
					if active {
						fieldTests = append(fieldTests, []test.FieldAssertion{
							{Name: "LatestVersionIsDeployed", Got: testutil.ToFloat64(latestVersionMetric), Want: float64(0), Mode: test.CompareEqual},
							{Name: "LatestVersionQueryResultLast", Got: testutil.ToFloat64(latestVersionQueryMetric), Want: float64(0), Mode: test.CompareEqual},
							{Name: "DeployedVersionQueryResultLast", Got: testutil.ToFloat64(deployedVersionMetric), Want: float64(0), Mode: test.CompareEqual},
							{Name: "CommandResultTotal", Got: testutil.ToFloat64(commandMetric), Want: float64(0), Mode: test.CompareEqual},
							{Name: "NotifyResultTotal", Got: testutil.ToFloat64(notifyMetric), Want: float64(0), Mode: test.CompareEqual},
							{Name: "WebhookResultTotal", Got: testutil.ToFloat64(webhookMetric), Want: float64(0), Mode: test.CompareEqual},
						}...)
					}
					if testErr := test.AssertFields(t, fieldTests, "metric"); testErr != nil {
						t.Fatalf("after deleteMetrics: %v", testErr)
					}
				})
			}
		})
	}
}
