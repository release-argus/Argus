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
	"testing"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/internal/test"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	dvtest "github.com/release-argus/Argus/service/deployed_version/test"
	dvmanual "github.com/release-argus/Argus/service/deployed_version/types/manual"
	dvweb "github.com/release-argus/Argus/service/deployed_version/types/web"
	lvtest "github.com/release-argus/Argus/service/latest_version/test"
	lvgithub "github.com/release-argus/Argus/service/latest_version/types/github"
	lvweb "github.com/release-argus/Argus/service/latest_version/types/web"
	statustest "github.com/release-argus/Argus/service/status/test"
	"github.com/release-argus/Argus/util"
	whtest "github.com/release-argus/Argus/webhook/test"
)

func assertServiceCopyMutate(t *testing.T, svcI, svcJ *Service) {
	t.Helper()

	fieldTests := []test.FieldAssertion{
		{Name: "ID", Got: svcJ.ID, Want: svcI.ID, Mode: test.CompareEqual},
		{Name: "Name", Got: svcJ.Name, Want: svcI.Name, Mode: test.CompareEqual},
		{Name: "Comment", Got: svcJ.Comment, Want: svcI.Comment, Mode: test.CompareEqual},
	}
	if testErr := test.AssertFields(t, fieldTests, "Service"); testErr != nil {
		t.Fatal(testErr)
	}

	fieldTests = []test.FieldAssertion{
		{Name: "LatestVersion", Got: svcJ.LatestVersion, Want: svcI.LatestVersion, Mode: test.CopiedPointerMode(svcI.LatestVersion)},
		{Name: "DeployedVersionLookup", Got: svcJ.DeployedVersionLookup, Want: svcI.DeployedVersionLookup, Mode: test.CopiedPointerMode(svcI.DeployedVersionLookup)},
	}
	if testErr := test.AssertFields(t, fieldTests, "Service"); testErr != nil {
		t.Fatal(testErr)
	}

	if svcI.Dashboard.AutoApprove != nil {
		if svcJ.Dashboard.AutoApprove == nil {
			t.Fatal("Service.Dashboard.AutoApprove mismatch\ngot:  nil\nwant: non-nil")
		}
		hadDashboardAutoApproveJ := *svcJ.Dashboard.AutoApprove
		*svcI.Dashboard.AutoApprove = !*svcI.Dashboard.AutoApprove
		if gotJ := *svcJ.Dashboard.AutoApprove; gotJ != hadDashboardAutoApproveJ {
			t.Errorf(
				"Service.Dashboard.AutoApprove mutation mismatch\ngot:  %t\nwant: %t",
				gotJ, hadDashboardAutoApproveJ,
			)
		}
	}

	hadStatusLatestVersionJ := svcJ.Status.LatestVersion()
	svcI.Status.SetLatestVersion(hadStatusLatestVersionJ+"-mutated", "", false)
	if gotJ := svcJ.Status.LatestVersion(); gotJ != hadStatusLatestVersionJ {
		t.Errorf(
			"Service.Status.LatestVersion() mutation mismatch\ngot:  %q\nwant: %q",
			gotJ, hadStatusLatestVersionJ,
		)
	}

	if svcI.LatestVersion != nil {
		var hadLatestVersionValueJ string
		switch iLV := svcI.LatestVersion.(type) {
		case *lvgithub.Lookup:
			jLV, ok := svcJ.LatestVersion.(*lvgithub.Lookup)
			if ok {
				hadLatestVersionValueJ = jLV.URL
			} else {
				t.Fatalf(
					"Service.LatestVersion type mismatch\ngot:  %T\nwant: *lvgithub.Lookup",
					svcJ.LatestVersion,
				)
			}
			iLV.URL += "-mutated"
			if gotJ := jLV.URL; gotJ != hadLatestVersionValueJ {
				t.Errorf(
					"Service.LatestVersion.URL mutation mismatch\ngot:  %q\nwant: %q",
					gotJ, hadLatestVersionValueJ,
				)
			}
		case *lvweb.Lookup:
			jLV, ok := svcJ.LatestVersion.(*lvweb.Lookup)
			if ok {
				hadLatestVersionValueJ = jLV.URL
			} else {
				t.Fatalf(
					"Service.LatestVersion type mismatch\ngot:  %T\nwant: *lvweb.Lookup",
					svcJ.LatestVersion,
				)
			}
			iLV.URL += "-mutated"
			if gotJ := jLV.URL; gotJ != hadLatestVersionValueJ {
				t.Errorf(
					"Service.LatestVersion.URL mutation mismatch\ngot:  %q\nwant: %q",
					gotJ, hadLatestVersionValueJ,
				)
			}
		default:
			t.Fatalf("unhandled Service.LatestVersion type: %T", iLV)
		}
	}

	if svcI.DeployedVersionLookup != nil {
		var hadDeployedVersionValueJ string
		switch iDV := svcI.DeployedVersionLookup.(type) {
		case *dvmanual.Lookup:
			jDV, ok := svcJ.DeployedVersionLookup.(*dvmanual.Lookup)
			if ok {
				hadDeployedVersionValueJ = jDV.Version
			} else {
				t.Fatalf(
					"Service.DeployedVersionLookup type mismatch\ngot:  %T\nwant: *dvmanual.Lookup",
					svcJ.DeployedVersionLookup,
				)
			}
			iDV.Version += "-mutated"
			if gotJ := jDV.Version; gotJ != hadDeployedVersionValueJ {
				t.Errorf(
					"Service.DeployedVersionLookup.Version mutation mismatch\ngot:  %q\nwant: %q",
					gotJ, hadDeployedVersionValueJ,
				)
			}
		case *dvweb.Lookup:
			jDV, ok := svcJ.DeployedVersionLookup.(*dvweb.Lookup)
			if ok {
				hadDeployedVersionValueJ = jDV.URL
			} else {
				t.Fatalf(
					"Service.DeployedVersionLookup type mismatch\ngot:  %T\nwant: *dvweb.Lookup",
					svcJ.DeployedVersionLookup,
				)
			}
			iDV.URL += "-mutated"
			if gotJ := jDV.URL; gotJ != hadDeployedVersionValueJ {
				t.Errorf(
					"Service.DeployedVersionLookup.URL mutation mismatch\ngot:  %q\nwant: %q",
					gotJ, hadDeployedVersionValueJ,
				)
			}
		default:
			t.Fatalf("unhandled Service.DeployedVersionLookup type: %T", iDV)
		}
	}

	notifyKeysI := util.SortedKeys(svcI.Notify)
	notifyKeysJ := util.SortedKeys(svcJ.Notify)
	if !util.AreSlicesEqual(notifyKeysI, notifyKeysJ) {
		t.Fatalf("Service.Notify keys mismatch\ngot:  %q\nwant: %q", notifyKeysJ, notifyKeysI)
	}
	for _, key := range notifyKeysI {
		hadNotifyValueJ := svcJ.Notify[key].ID
		svcI.Notify[key].ID += "-mutated"
		if gotJ := svcJ.Notify[key].ID; gotJ != hadNotifyValueJ {
			t.Errorf(
				"Service.Notify[%q].ID mutation mismatch\ngot:  %q\nwant: %q",
				key, gotJ, hadNotifyValueJ,
			)
		}
	}

	if testErr := test.AssertSlicesEqualFunc(
		t,
		svcI.Command,
		svcJ.Command,
		func(i, j command.Command) bool {
			return i.JSON() == j.JSON()
		},
		"Service.Command",
	); testErr != nil {
		t.Fatal(testErr)
	}
	if len(svcI.Command) > 0 {
		hadCommandValueJ := svcJ.Command[0][0]
		svcI.Command[0][0] += "-mutated"
		if gotJ := svcJ.Command[0][0]; gotJ != hadCommandValueJ {
			t.Errorf(
				"Service.Command[0][0] mutation mismatch\ngot:  %q\nwant: %q",
				gotJ, hadCommandValueJ,
			)
		}
	}

	webHookKeysI := util.SortedKeys(svcI.Webhook)
	webHookKeysJ := util.SortedKeys(svcJ.Webhook)
	if !util.AreSlicesEqual(webHookKeysI, webHookKeysJ) {
		t.Fatalf("Service.Webhook keys mismatch\ngot:  %q\nwant: %q", webHookKeysJ, webHookKeysI)
	}
	for _, key := range webHookKeysI {
		hadWebhookValueJ := svcJ.Webhook[key].ID
		svcI.Webhook[key].ID += "-mutated"
		if gotJ := svcJ.Webhook[key].ID; gotJ != hadWebhookValueJ {
			t.Errorf(
				"Service.Webhook[%q].ID mutation mismatch\ngot:  %q\nwant: %q",
				key, gotJ, hadWebhookValueJ,
			)
		}
	}
}

func TestService_Copy(t *testing.T) {
	tests := []struct {
		name string
		yaml string
	}{
		{
			name: "nil",
			yaml: "", // No YAML: a nil Service.
		},
		{
			name: "LatestVersion/github",
			yaml: `
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
			`,
		},
		{
			name: "LatestVersion/url",
			yaml: `
				latest_version:
				` + lvtest.Lookup(t, "url", false).String("    ") + `
			`,
		},
		{
			name: "DeployedVersionLookup",
			yaml: `
				deployed_version:
				` + dvtest.Lookup(t, "url", false, "").String("    ") + `
			`,
		},
		{
			name: "Dashboard",
			yaml: `
				dashboard:
					auto_approve: true
			`,
		},
		{
			name: "Notify",
			yaml: `
				notify:
					test:
				` + shoutrrrtest.Shoutrrr(false, false).String("    ") + `
			`,
		},
		{
			name: "Command",
			yaml: `
				command:
					- ['echo', 'Hello, World!']
			`,
		},
		{
			name: "Webhook",
			yaml: `
				webhook:
					test:
				` + whtest.Webhook(false, false, true).String("    ") + `
			`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var svc *Service
			if tc.yaml != "" {
				svc = decodeWithStatus(t, tc.name, tc.yaml)
			}

			got := svc.Copy(false)

			if svc == nil {
				if got != nil {
					t.Error("without channels: Service.Copy() mismatch\ngot:  non-nil\nwant: nil")
				}
				return
			}
			if got == nil {
				t.Fatal("without channels: Service.Copy() mismatch\ngot:  nil\nwant: non-nil")
			}

			gotStr, wantStr := got.String(""), svc.String("")
			if gotStr != wantStr {
				t.Errorf(
					"without channels: Service.String() mismatch\ngot:  %q\nwant: %q",
					gotStr, wantStr,
				)
			}

			gotStatus := &got.Status
			if gotStatus.AnnounceChannel != nil ||
				gotStatus.DatabaseChannel != nil ||
				gotStatus.SaveChannel != nil {
				t.Errorf(
					"without channels: Service.Status channels mismatch\n"+
						"got:  announce=%p, database=%p, save=%p\nwant: all nil",
					gotStatus.AnnounceChannel, gotStatus.DatabaseChannel, gotStatus.SaveChannel,
				)
			}

			assertServiceCopyMutate(t, svc, got)

			got = svc.Copy(true)
			if got == nil {
				t.Fatal("with channels: Service.Copy() mismatch\ngot:  nil\nwant: non-nil")
			}

			gotStr = got.String("")
			wantStr = svc.String("")
			if gotStr != wantStr {
				t.Errorf(
					"with channels: Service.String() mismatch\ngot:  %q\nwant: %q",
					gotStr, wantStr,
				)
			}

			hadStatus := &svc.Status
			gotStatus = &got.Status
			if hadStatus.AnnounceChannel != gotStatus.AnnounceChannel ||
				hadStatus.DatabaseChannel != gotStatus.DatabaseChannel ||
				hadStatus.SaveChannel != gotStatus.SaveChannel {
				t.Errorf(
					"with channels: Service.Status channels mismatch\n"+
						"got:  announce=%p, database=%p, save=%p\n"+
						"want: announce=%p, database=%p, save=%p",
					gotStatus.AnnounceChannel, gotStatus.DatabaseChannel, gotStatus.SaveChannel,
					hadStatus.AnnounceChannel, hadStatus.DatabaseChannel, hadStatus.SaveChannel,
				)
			}

			assertServiceCopyMutate(t, svc, got)
		})
	}
}

// decodeWithStatus decodes the YAML into a Service with ID id, with a Status that has channels.
func decodeWithStatus(t *testing.T, id, yaml string) *Service {
	t.Helper()

	svc, err := DecodeService(
		"yaml", []byte(test.TrimYAML(yaml)),
		id,
		plainDefaultsConfig(), shoutrrrtest.PlainConfig(), whtest.PlainConfig(),
	)
	if err != nil {
		t.Fatalf("DecodeService(id=%q) failed: %v", id, err)
	}

	svcStatus, _ := statustest.New("yaml", nil)
	svc.Status = *svcStatus.Copy(true)
	return svc
}
