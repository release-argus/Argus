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

func assertServiceCopyMutate(t *testing.T, prefix string, svcI, svcJ *Service) {
	if svcI == nil {
		if svcJ != nil {
			t.Errorf("%s of nil\ngot:  non-nil\nwant: nil", prefix)
		}
		return
	}
	if svcJ == nil {
		t.Errorf("%s of non-nil\ngot:  nil\nwant: non-nil", prefix)
		return
	}

	// Basic fields.
	fieldTests := []test.FieldAssertion{
		{Name: "ID", Got: svcJ.ID, Want: svcI.ID, Mode: test.CompareEqual},
		{Name: "Name", Got: svcJ.Name, Want: svcI.Name, Mode: test.CompareEqual},
		{Name: "Comment", Got: svcJ.Comment, Want: svcI.Comment, Mode: test.CompareEqual},
	}
	if err := test.AssertFields(t, fieldTests, prefix, "basic field"); err != nil {
		t.Fatal(err)
	}

	// Pointers have been copied, not referenced.
	fieldTests = []test.FieldAssertion{
		{Name: "Options", Got: &svcJ.Options, Want: &svcI.Options, Mode: test.CompareDifferentPointer},
		{Name: "Dashboard", Got: &svcJ.Dashboard, Want: &svcI.Dashboard, Mode: test.CompareDifferentPointer},
		{Name: "Status", Got: &svcJ.Status, Want: &svcI.Status, Mode: test.CompareDifferentPointer},
		{Name: "LatestVersion", Got: svcJ.LatestVersion, Want: svcI.LatestVersion, Mode: test.CompareDifferentPointer},
		{Name: "DeployedVersionLookup", Got: svcJ.DeployedVersionLookup, Want: svcI.DeployedVersionLookup, Mode: test.CompareDifferentPointer},
		{Name: "Notify", Got: &svcJ.Notify, Want: &svcI.Notify, Mode: test.CompareDifferentPointer},
		{Name: "Command", Got: &svcJ.Command, Want: &svcI.Command, Mode: test.CompareDifferentPointer},
		{Name: "WebHook", Got: &svcJ.WebHook, Want: &svcI.WebHook, Mode: test.CompareDifferentPointer},
	}
	if err := test.AssertFields(t, fieldTests, prefix, "pointer"); err != nil {
		t.Fatal(err)
	}

	// Options.
	hadOptionsIntervalJ := svcJ.Options.Interval
	svcI.Options.Interval = "mutated"
	if gotJ := svcJ.Options.Interval; gotJ != hadOptionsIntervalJ {
		t.Errorf(
			"%s Options.Interval mutation mismatch\ngot:  %q\nwant: %q",
			prefix, gotJ, hadOptionsIntervalJ,
		)
	}

	// Dashboard.
	hadDashboardAutoApproveJ := svcJ.Dashboard.AutoApprove
	var newAutoApprove bool
	if hadDashboardAutoApproveJ != nil {
		newAutoApprove = !*hadDashboardAutoApproveJ
	}
	svcI.Dashboard.AutoApprove = &newAutoApprove
	if gotJ := svcJ.Dashboard.AutoApprove; gotJ != hadDashboardAutoApproveJ {
		t.Errorf(
			"%s Dashboard.AutoApprove mutation mismatch\ngot:  %v\nwant: %v",
			prefix, gotJ, hadDashboardAutoApproveJ,
		)
	}

	// Status.
	hadStatusLatestVersionJ := svcJ.Status.LatestVersion()
	svcI.Status.SetLatestVersion(hadStatusLatestVersionJ+"-mutated", "", false)
	if gotJ := svcJ.Status.LatestVersion(); gotJ != hadStatusLatestVersionJ {
		t.Errorf(
			"%s Status.LatestVersion mutation mismatch\ngot:  %q\nwant: %q",
			prefix, gotJ, hadStatusLatestVersionJ,
		)
	}

	// LatestVersion.
	if svcI.LatestVersion != nil {
		var hadLatestVersionValueJ string
		switch iLV := svcI.LatestVersion.(type) {
		case *lvgithub.Lookup:
			jLV, ok := svcJ.LatestVersion.(*lvgithub.Lookup)
			if ok {
				hadLatestVersionValueJ = jLV.URL
			} else {
				t.Fatalf(
					"%s LatestVersion type mismatch\ngot:  %T\nwant: *lvgithub.Lookup",
					prefix, svcJ.LatestVersion,
				)
			}
			iLV.URL += "-mutated"
			if gotJ := jLV.URL; gotJ != hadLatestVersionValueJ {
				t.Errorf(
					"%s LatestVersion.URL mutation mismatch\ngot:  %q\nwant: %q",
					prefix, gotJ, hadLatestVersionValueJ,
				)
			}
		case *lvweb.Lookup:
			jLV, ok := svcJ.LatestVersion.(*lvweb.Lookup)
			if ok {
				hadLatestVersionValueJ = jLV.URL
			} else {
				t.Fatalf(
					"%s LatestVersion type mismatch\ngot:  %T\nwant: *lvweb.Lookup",
					prefix, svcJ.LatestVersion,
				)
			}
			iLV.URL += "-mutated"
			if gotJ := jLV.URL; gotJ != hadLatestVersionValueJ {
				t.Errorf(
					"%s LatestVersion.URL mutation mismatch\ngot:  %q\nwant: %q",
					prefix, gotJ, hadLatestVersionValueJ,
				)
			}
		}
	}

	// DeployedVersion.
	if svcI.DeployedVersionLookup != nil {
		var hadDeployedVersionValueJ string
		switch iDV := svcI.DeployedVersionLookup.(type) {
		case *dvmanual.Lookup:
			jDV, ok := svcJ.DeployedVersionLookup.(*dvmanual.Lookup)
			if ok {
				hadDeployedVersionValueJ = jDV.Version
			} else {
				t.Fatalf(
					"%s DeployedVersion type mismatch\ngot:  %T\nwant: *lvgithub.Lookup",
					prefix, svcJ.DeployedVersionLookup,
				)
			}
			iDV.Version += "-mutated"
			if gotJ := jDV.Version; gotJ != hadDeployedVersionValueJ {
				t.Errorf(
					"%s DeployedVersion.URL mutation mismatch\ngot:  %q\nwant: %q",
					prefix, gotJ, hadDeployedVersionValueJ,
				)
			}
		case *dvweb.Lookup:
			jDV, ok := svcJ.DeployedVersionLookup.(*dvweb.Lookup)
			if ok {
				hadDeployedVersionValueJ = jDV.URL
			} else {
				t.Fatalf(
					"%s DeployedVersion type mismatch\ngot:  %T\nwant: *lvweb.Lookup",
					prefix, svcJ.DeployedVersionLookup,
				)
			}
			iDV.URL += "-mutated"
			if gotJ := jDV.URL; gotJ != hadDeployedVersionValueJ {
				t.Errorf(
					"%s DeployedVersion.URL mutation mismatch\ngot:  %q\nwant: %q",
					prefix, gotJ, hadDeployedVersionValueJ,
				)
			}
		}
	}

	// Notify.
	notifyKeysI := util.SortedKeys(svcI.Notify)
	notifyKeysJ := util.SortedKeys(svcJ.Notify)
	if !util.AreSlicesEqual(notifyKeysI, notifyKeysJ) {
		t.Fatalf(
			"%s Notify keys mismatch\ngot:  %v\nwant: %v",
			prefix, notifyKeysJ, notifyKeysI,
		)
	}
	for _, key := range notifyKeysI {
		hadNotifyValueJ := svcJ.Notify[key].ID
		svcI.Notify[key].ID += "-mutated"
		if gotJ := svcJ.Notify[key].ID; gotJ != hadNotifyValueJ {
			t.Errorf(
				"%s Notify[%q].ID mutation mismatch\ngot:  %q\nwant: %q",
				prefix, key, gotJ, hadNotifyValueJ,
			)
		}
	}

	// Command.
	if err := test.AssertSlicesEqualFunc(
		t,
		svcI.Command,
		svcJ.Command,
		func(i, j command.Command) bool {
			return i.JSON() == j.JSON()
		},
		prefix,
		"Command",
	); err != nil {
		t.Fatal(err)
	}
	if len(svcI.Command) > 0 {
		hadCommandValueJ := svcJ.Command[0][0]
		svcI.Command[0][0] += "-mutated"
		if gotJ := svcJ.Command[0][0]; gotJ != hadCommandValueJ {
			t.Errorf(
				"%s Command[0][0] mutation mismatch\ngot:  %q\nwant: %q",
				prefix, gotJ, hadCommandValueJ,
			)
		}
	}

	// WebHook.
	webHookKeysI := util.SortedKeys(svcI.WebHook)
	webHookKeysJ := util.SortedKeys(svcJ.WebHook)
	if !util.AreSlicesEqual(webHookKeysI, webHookKeysJ) {
		t.Fatalf(
			"%s WebHook keys mismatch\ngot:  %v\nwant: %v",
			prefix, webHookKeysJ, webHookKeysI,
		)
	}
	for _, key := range webHookKeysI {
		hadWebHookValueJ := svcJ.WebHook[key].ID
		svcI.WebHook[key].ID += "-mutated"
		if gotJ := svcJ.WebHook[key].ID; gotJ != hadWebHookValueJ {
			t.Errorf(
				"%s WebHook[%q].ID mutation mismatch\ngot:  %q\nwant: %q",
				prefix, key, gotJ, hadWebHookValueJ,
			)
		}
	}
}

func TestService_Copy(t *testing.T) {
	svcCfg := plainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig(t)
	whCfg := whtest.PlainConfig(t)

	// GIVEN: a Service.
	tests := []struct {
		name string
		svc  *Service
	}{
		{
			name: "nil",
			svc:  nil,
		},
		{
			name: "LatestVersion",
			svc: test.Must(t, func() (*Service, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				svc, err := DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
						`+lvtest.Lookup(t, "url", false).String("    ")+`
					`)),
					"LatestVersion",
					svcCfg, notifyCfg, whCfg,
				)
				if err != nil {
					return nil, err
				}
				svc.Status = *svcStatus.Copy(true)
				return svc, nil
			}),
		},
		{
			name: "DeployedVersionLookup",
			svc: test.Must(t, func() (*Service, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				svc, err := DecodeService(
					"yaml", []byte(test.TrimYAML(`
						deployed_version:
						`+dvtest.Lookup(t, "url", false, "").String("    ")+`
					`)),
					"DeployedVersionLookup",
					svcCfg, notifyCfg, whCfg,
				)
				if err != nil {
					return nil, err
				}
				svc.Status = *svcStatus.Copy(true)
				return svc, nil
			}),
		},
		{
			name: "Notify",
			svc: test.Must(t, func() (*Service, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				svc, err := DecodeService(
					"yaml", []byte(test.TrimYAML(`
						notify:
							test:
						`+shoutrrrtest.Shoutrrr(t, false, false).String("    ")+`
					`)),
					"Notify",
					svcCfg, notifyCfg, whCfg,
				)
				if err != nil {
					return nil, err
				}
				svc.Status = *svcStatus.Copy(true)
				return svc, nil
			}),
		},
		{
			name: "Command",
			svc: test.Must(t, func() (*Service, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				svc, err := DecodeService(
					"yaml", []byte(test.TrimYAML(`
						command:
							- ['echo', 'Hello, World!']
					`)),
					"Command",
					svcCfg, notifyCfg, whCfg,
				)
				if err != nil {
					return nil, err
				}
				svc.Status = *svcStatus.Copy(true)
				return svc, nil
			}),
		},
		{
			name: "WebHook",
			svc: test.Must(t, func() (*Service, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				svc, err := DecodeService(
					"yaml", []byte(test.TrimYAML(`
						webhook:
							test:
						`+whtest.WebHook(t, false, false, true).String("    ")+`
					`)),
					"WebHook",
					svcCfg, notifyCfg, whCfg,
				)
				if err != nil {
					return nil, err
				}
				svc.Status = *svcStatus.Copy(true)
				return svc, nil
			}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Copy is called on it.
			withChannels := false
			got := tc.svc.Copy(withChannels)

			prefix := fmt.Sprintf(
				"%s\nService.Copy(withChannels=%t)",
				packageName, withChannels,
			)

			// THEN: nil is received only when nil was copied.
			if got == nil {
				if tc.svc != nil {
					t.Errorf("%s of nil\ngot:  non-nil\nwant: nil", prefix)
				}
				return
			}

			// AND: the struct stringifies exactly the same.
			gotStr, wantStr := got.String(""), tc.svc.String("")
			if gotStr != wantStr {
				t.Errorf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, gotStr, wantStr,
				)
			}

			// AND: the Status channels are not copied over.
			gotStatus := &got.Status
			if gotStatus.AnnounceChannel != nil ||
				gotStatus.DatabaseChannel != nil ||
				gotStatus.SaveChannel != nil {
				t.Errorf(
					"%s .Status channel(s) mismatch\ngot:  %+v\nwant: all nil",
					prefix, gotStatus,
				)
			}

			// AND: mutating the copy does not mutate the original.
			assertServiceCopyMutate(t, prefix, tc.svc, got)

			// GIVEN: withChannels is true.
			withChannels = true

			// WHEN: Copy() is called again.
			got = tc.svc.Copy(withChannels)

			prefix = fmt.Sprintf(
				"%s\nLookup.Copy(withChannels=%t)",
				packageName, withChannels,
			)

			// THEN: the same data is returned.
			gotStr = tc.svc.String("")
			wantStr = got.String("")
			if gotStr != wantStr {
				t.Errorf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, gotStr, wantStr,
				)
			}

			// AND: the Status channels are copied over.
			hadStatus := &tc.svc.Status
			gotStatus = &got.Status
			if hadStatus.AnnounceChannel != gotStatus.AnnounceChannel ||
				hadStatus.DatabaseChannel != gotStatus.DatabaseChannel ||
				hadStatus.SaveChannel != gotStatus.SaveChannel {
				t.Errorf(
					"%s .Status channel(s) mismatch\ngot:  %+v\nwant: %+v",
					prefix, gotStatus, hadStatus,
				)
			}

			// AND: mutating the copy does not mutate the original.
			assertServiceCopyMutate(t, prefix, tc.svc, got)
		})
	}
}
