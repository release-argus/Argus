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

package status

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/release-argus/Argus/config/decode"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/web/metric"
)

func TestNewDefaults(t *testing.T) {
	announceChannel := make(chan []byte, 4)
	databaseChannel := make(chan dbtype.Message, 4)
	saveChannel := make(chan bool, 4)

	statusDefaults := NewDefaults(announceChannel, databaseChannel, saveChannel)

	fieldTests := []test.FieldAssertion{
		{Name: "AnnounceChannel", Got: statusDefaults.AnnounceChannel, Want: announceChannel, Mode: test.CompareSamePointer},
		{Name: "DatabaseChannel", Got: statusDefaults.DatabaseChannel, Want: databaseChannel, Mode: test.CompareSamePointer},
		{Name: "SaveChannel", Got: statusDefaults.SaveChannel, Want: saveChannel, Mode: test.CompareSamePointer},
	}
	if testErr := test.AssertFields(t, fieldTests, "Defaults"); testErr != nil {
		t.Fatal(testErr)
	}
}

func TestStatus_Unmarshal(t *testing.T) {
	tests := []struct {
		format string
	}{
		{
			format: "YAML",
		},
		{
			format: "JSON",
		},
	}

	for _, tc := range tests {
		t.Run(tc.format, func(t *testing.T) {
			t.Parallel()

			data := []byte("")
			var status Status
			var err error
			switch tc.format {
			case "YAML":
				err = status.UnmarshalYAML(data)
			case "JSON":
				err = status.UnmarshalJSON(data)
			default:
				t.Fatal("unknown format")
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			unlockedAtChan := make(chan time.Time, 1)
			gotServiceInfoAtChan := make(chan time.Time, 1)
			status.mu.Lock()
			go func() {
				status.GetServiceInfo()
				gotServiceInfoAtChan <- time.Now()
			}()
			go func() {
				time.Sleep(100 * time.Millisecond)
				unlockedAtChan <- time.Now()
				status.mu.Unlock()
			}()
			var unlockedAt time.Time
			select {
			case unlockedAt = <-unlockedAtChan:
			case <-time.After(time.Second):
				t.Fatal("Status.mu unlock mismatch\ngot:  still locked\nwant: unlocked")
			}
			var gotServiceInfoAt time.Time
			select {
			case gotServiceInfoAt = <-gotServiceInfoAtChan:
			case <-time.After(time.Second):
				t.Fatal("Status.GetServiceInfo() mismatch\ngot:  still blocked\nwant: returned")
			}
			if gotServiceInfoAt.Before(unlockedAt) {
				t.Errorf(
					"Status.GetServiceInfo() return time mismatch\ngot:  %v\nwant: NOT before %v",
					gotServiceInfoAt, unlockedAt,
				)
			}
		})
	}
}

func TestStatus_Copy(t *testing.T) {
	announceChannel := make(chan []byte, 4)
	databaseChannel := make(chan dbtype.Message, 4)
	saveChannel := make(chan bool, 4)
	status := New(
		announceChannel, databaseChannel, saveChannel,
		"1.0.0",
		"0.9.0", "2023-01-01T00:00:00Z",
		"1.0.0", "2023-01-02T00:00:00Z",
		"2023-01-03T00:00:00Z",
		&dashboard.Options{},
	)

	tests := []struct {
		name         string
		copyChannels bool
	}{
		{name: "copy channels", copyChannels: true},
		{name: "don't copy channels", copyChannels: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			copiedStatus := status.Copy(tc.copyChannels)

			fieldTests := []test.FieldAssertion{
				{Name: "ServiceInfo.ApprovedVersion", Got: copiedStatus.ServiceInfo.ApprovedVersion, Want: status.ServiceInfo.ApprovedVersion, Mode: test.CompareEqual},
				{Name: "ServiceInfo.DeployedVersion", Got: copiedStatus.ServiceInfo.DeployedVersion, Want: status.ServiceInfo.DeployedVersion, Mode: test.CompareEqual},
				{Name: "deployedVersionTimestamp", Got: copiedStatus.deployedVersionTimestamp, Want: status.deployedVersionTimestamp, Mode: test.CompareEqual},
				{Name: "ServiceInfo.LatestVersion", Got: copiedStatus.ServiceInfo.LatestVersion, Want: status.ServiceInfo.LatestVersion, Mode: test.CompareEqual},
				{Name: "latestVersionTimestamp", Got: copiedStatus.latestVersionTimestamp, Want: status.latestVersionTimestamp, Mode: test.CompareEqual},
				{Name: "lastQueried", Got: copiedStatus.lastQueried, Want: status.lastQueried, Mode: test.CompareEqual},
			}
			if testErr := test.AssertFields(t, fieldTests, "Status"); testErr != nil {
				t.Fatal(testErr)
			}

			fieldTests = []test.FieldAssertion{
				{Name: "AnnounceChannel", Got: copiedStatus.AnnounceChannel, Want: status.AnnounceChannel, Mode: test.CompareSamePointer},
				{Name: "DatabaseChannel", Got: copiedStatus.DatabaseChannel, Want: status.DatabaseChannel, Mode: test.CompareSamePointer},
				{Name: "SaveChannel", Got: copiedStatus.SaveChannel, Want: status.SaveChannel, Mode: test.CompareSamePointer},
			}
			if !tc.copyChannels {
				for i := range fieldTests {
					fieldTests[i].Want = nil
				}
			}
			if testErr := test.AssertFields(t, fieldTests, "Status"); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestStatus_Copy__nil(t *testing.T) {
	var status *Status

	tests := []struct {
		name         string
		copyChannels bool
	}{
		{name: "copy channels", copyChannels: true},
		{name: "don't copy channels", copyChannels: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			copiedStatus := status.Copy(tc.copyChannels)

			if copiedStatus != nil {
				t.Fatal("mismatch\ngot:  non-nil\nwant: nil")
			}
		})
	}
}

func TestStatus_String(t *testing.T) {
	tests := []struct {
		name                                   string
		status                                 *Status
		regexMissesContent, regexMissesVersion int
		want                                   string
	}{
		{
			name:   "empty status",
			status: &Status{},
			want:   "",
		},
		{
			name: "only fails",
			status: &Status{
				Fails: Fails{
					Shoutrrr: FailsShoutrrr{
						fails: map[string]*bool{
							"bash": new(false),
							"bish": nil,
							"bosh": new(true),
						},
					},
					Command: FailsCommand{
						fails: []*bool{
							nil,
							new(false),
							new(true),
						},
					},
					Webhook: FailsWebhook{
						fails: map[string]*bool{
							"bar": nil,
							"foo": new(false),
						},
					},
				},
			},
			want: test.TrimYAML(`
				fails:
					shoutrrr:
						bash: false
						bish: nil
						bosh: true
					command:
						- 0: nil
						- 1: false
						- 2: true
					webhook:
						bar: nil
						foo: false
			`),
		},
		{
			name:               "filled",
			regexMissesContent: 1,
			regexMissesVersion: 2,
			status: &Status{
				ServiceInfo: serviceinfo.ServiceInfo{
					ApprovedVersion: "1.2.4",
					DeployedVersion: "1.2.3",
					LatestVersion:   "1.2.4",
				},
				deployedVersionTimestamp: "2022-01-01T01:01:02Z",
				latestVersionTimestamp:   "2022-01-01T01:01:01Z",
				lastQueried:              "2022-01-01T01:01:01Z",
				Fails: Fails{
					Shoutrrr: FailsShoutrrr{
						fails: map[string]*bool{
							"bish": nil,
							"bash": new(false),
							"bosh": new(true),
						},
					},
					Command: FailsCommand{
						fails: []*bool{
							nil,
							new(false),
							new(true),
						},
					},
					Webhook: FailsWebhook{
						fails: map[string]*bool{
							"foo": new(false),
							"bar": nil,
						},
					},
				},
			},
			want: test.TrimYAML(`
				approved_version: 1.2.4
				deployed_version: 1.2.3
				deployed_version_timestamp: 2022-01-01T01:01:02Z
				latest_version: 1.2.4
				latest_version_timestamp: 2022-01-01T01:01:01Z
				last_queried: 2022-01-01T01:01:01Z
				regex_misses_content: 1
				regex_misses_version: 2
				fails:
					shoutrrr:
						bash: false
						bish: nil
						bosh: true
					command:
						- 0: nil
						- 1: false
						- 2: true
					webhook:
						bar: nil
						foo: false
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for i := 0; i < tc.regexMissesContent; i++ {
				tc.status.RegexMissContent()
			}
			for i := 0; i < tc.regexMissesVersion; i++ {
				tc.status.RegexMissVersion()
			}

			got := tc.status.String()

			if got != tc.want {
				t.Errorf("Status.String() mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestStatus_Init(t *testing.T) {
	tests := []struct {
		name                          string
		shoutrrrs, commands, webhooks int
		serviceID                     string
		webURL                        string
	}{
		{
			name:      "ServiceID",
			serviceID: "test",
		},
		{
			name:   "WebURL",
			webURL: "https://example.com",
		},
		{
			name:      "shoutrrrs",
			shoutrrrs: 2,
		},
		{
			name:     "commands",
			commands: 3,
		},
		{
			name:     "webhooks",
			webhooks: 4,
		},
		{
			name:      "all",
			serviceID: "argus",
			webURL:    "https://release-argus.io",
			shoutrrrs: 5,
			commands:  5,
			webhooks:  5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var status Status

			status.Init(
				tc.commands, tc.shoutrrrs, tc.webhooks,
				ServiceInfo{
					ID:   tc.serviceID,
					Name: tc.name,
				},
				&dashboard.Options{
					WebURL: tc.webURL,
				},
			)

			if status.ServiceInfo.ID != tc.serviceID {
				t.Errorf(
					"after Init: Status.ServiceInfo.ID mismatch\ngot:  %q\nwant: %q",
					status.ServiceInfo.ID, tc.serviceID,
				)
			}
			if got := status.GetWebURL(); got != tc.webURL {
				t.Errorf(
					"after Init: Status.GetWebURL() mismatch\ngot:  %q\nwant: %q",
					got, tc.webURL,
				)
			}
			// Init sets only the map's capacity.
			want := 0
			if got := status.Fails.Shoutrrr.Length(); got != want {
				t.Errorf(
					"after Init: Status.Fails.Shoutrrr.Length() mismatch\ngot:  %d\nwant: %d",
					got, want,
				)
			} else {
				for i := 0; i < tc.shoutrrrs; i++ {
					failed := false
					status.Fails.Shoutrrr.Set(strconv.Itoa(i), &failed)
				}
				if got := status.Fails.Shoutrrr.Length(); got != tc.shoutrrrs {
					t.Errorf(
						"after Set: Status.Fails.Shoutrrr.Length() mismatch\ngot:  %d\nwant: %d",
						got, tc.shoutrrrs,
					)
				}
			}
			if got := status.Fails.Command.Length(); got != tc.commands {
				t.Errorf(
					"after Init: Status.Fails.Command.Length() mismatch\ngot:  %d\nwant: %d",
					got, tc.commands,
				)
			}
			want = 0
			if got := status.Fails.Webhook.Length(); got != want {
				t.Errorf(
					"after Init: Status.Fails.Webhook.Length() mismatch\ngot:  %d\nwant: %d",
					got, want,
				)
			} else {
				for i := 0; i < tc.webhooks; i++ {
					failed := false
					status.Fails.Webhook.Set(strconv.Itoa(i), &failed)
				}
				if got := status.Fails.Webhook.Length(); got != tc.webhooks {
					t.Errorf(
						"after Set: Status.Fails.Webhook.Length() mismatch\ngot:  %d\nwant: %d",
						got, tc.webhooks,
					)
				}
			}
		})
	}
}

func TestStatus_SetAnnounceChannel(t *testing.T) {
	initialChannel := make(chan []byte, 4)
	status := New(
		initialChannel, nil, nil,
		"",
		"", "",
		"", "",
		"",
		&dashboard.Options{},
	)

	newChannel := make(chan []byte, 4)
	status.SetAnnounceChannel(newChannel)

	if status.AnnounceChannel != newChannel {
		t.Errorf(
			"Status.AnnounceChannel mismatch\ngot:  %p\nwant: %p",
			status.AnnounceChannel, newChannel,
		)
	}
}

func TestStatus_RefreshServiceInfo(t *testing.T) {
	status := Status{}
	status.Init(
		0, 0, 0,
		ServiceInfo{ID: "refresh-test"},
		&dashboard.Options{
			Icon:       "https://example.com/{{ version }}.png",
			IconLinkTo: "https://example.com/link/{{ version }}",
			WebURL:     "https://example.com/{{ version }}",
		},
	)
	status.SetLatestVersion("1.2.3", "", false)

	status.ServiceInfo.Icon = "stale"
	status.ServiceInfo.IconLinkTo = "stale"
	status.ServiceInfo.WebURL = "stale"

	status.RefreshServiceInfo()

	fieldTests := []test.FieldAssertion{
		{Name: "ServiceInfo.Icon", Got: status.ServiceInfo.Icon, Want: "https://example.com/1.2.3.png", Mode: test.CompareEqual},
		{Name: "ServiceInfo.IconLinkTo", Got: status.ServiceInfo.IconLinkTo, Want: "https://example.com/link/1.2.3", Mode: test.CompareEqual},
		{Name: "ServiceInfo.WebURL", Got: status.ServiceInfo.WebURL, Want: "https://example.com/1.2.3", Mode: test.CompareEqual},
	}
	if testErr := test.AssertFields(t, fieldTests, "Status"); testErr != nil {
		t.Fatal(testErr)
	}
}

func TestStatus_GetServiceInfo(t *testing.T) {
	status := testStatus()

	id := "test_id"
	status.ServiceInfo.ID = id
	name := "test_name"
	status.ServiceInfo.Name = name
	url := "https://example.com"
	status.ServiceInfo.URL = url

	icon := "https://example.com/icon"
	status.Dashboard.Icon = icon
	iconLinkTo := "https://example.com/icon_link"
	status.Dashboard.IconLinkTo = iconLinkTo
	webURL := "https://example.com/web"
	status.Dashboard.WebURL = webURL
	tags := []string{"tag1", "tag2"}
	status.ServiceInfo.Tags = tags

	deployedVersion := "deployed.version"
	status.SetDeployedVersion(deployedVersion, "", false)
	latestVersion := "latest.version"
	status.SetLatestVersion(latestVersion, "", false)
	approvedVersion := "approved.version"
	status.SetApprovedVersion(approvedVersion, false)

	want := serviceinfo.ServiceInfo{
		ID:   id,
		Name: name,
		URL:  url,

		Icon:       icon,
		IconLinkTo: iconLinkTo,
		WebURL:     webURL,
		Tags:       tags,

		DeployedVersion: deployedVersion,
		ApprovedVersion: approvedVersion,
		LatestVersion:   latestVersion,
	}

	got := status.GetServiceInfo()

	// The Set* calls refreshed the icon and URLs from the dashboard.
	gotStr := decode.ToJSONString(got)
	wantStr := decode.ToJSONString(want)
	if gotStr != wantStr {
		t.Errorf("mismatch\ngot:  %q\nwant: %q", gotStr, wantStr)
	}
}

func TestStatus_GetWebURL(t *testing.T) {
	latestVersion := "1.2.3"
	tests := []struct {
		name   string
		webURL string
		want   string
	}{
		{
			name:   "empty string",
			webURL: "",
			want:   "",
		},
		{
			name:   "string without templating",
			webURL: "https://example.com/somewhere",
			want:   "https://example.com/somewhere",
		},
		{
			name:   "string with templating",
			webURL: "https://example.com/somewhere/{{ version }}",
			want:   "https://example.com/somewhere/" + latestVersion,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			status := Status{}
			status.Init(
				0, 0, 0,
				ServiceInfo{
					ID: tc.name,
				},
				&dashboard.Options{
					WebURL: tc.webURL,
				},
			)
			status.SetLatestVersion(latestVersion, "", false)

			got := status.GetWebURL()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestStatus_SetAndGetLastQueried(t *testing.T) {
	tests := []struct {
		name        string
		lastQueried string
	}{
		{name: "empty string", lastQueried: ""},
		{name: "non-empty string", lastQueried: "2020-01-01T00:00:00Z"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			status := testStatus()

			status.SetLastQueried(tc.lastQueried)

			lastQueried := status.LastQueried()
			if tc.lastQueried != "" {
				if lastQueried != tc.lastQueried {
					t.Errorf(
						"Status.LastQueried() mismatch\ngot:  %q\nwant: %q",
						lastQueried, tc.lastQueried,
					)
				}
			} else {
				lastQueriedTime, err := time.Parse(time.RFC3339, lastQueried)
				if since := time.Since(lastQueriedTime); err != nil || since > time.Second {
					t.Errorf(
						"Status.LastQueried() mismatch\ngot:  %q (%s ago, err=%v)\nwant: <1s ago",
						lastQueried, since, err,
					)
				}
			}
		})
	}
}

func TestStatus_SameVersions(t *testing.T) {
	type versions struct {
		approvedVersion, deployedVersion, latestVersion string
	}

	tests := []struct {
		name             string
		status1, status2 versions
		expected         bool
	}{
		{
			name: "identical versions",
			status1: versions{
				approvedVersion: "1.0.0",
				deployedVersion: "0.9.0",
				latestVersion:   "1.0.0",
			},
			status2: versions{
				approvedVersion: "1.0.0",
				deployedVersion: "0.9.0",
				latestVersion:   "1.0.0",
			},
			expected: true,
		},
		{
			name: "different approved version",
			status1: versions{
				approvedVersion: "1.0.0",
				deployedVersion: "0.9.0",
				latestVersion:   "1.0.0",
			},
			status2: versions{
				approvedVersion: "1.1.0",
				deployedVersion: "0.9.0",
				latestVersion:   "1.0.0",
			},
			expected: false,
		},
		{
			name: "different deployed version",
			status1: versions{
				approvedVersion: "1.0.0",
				deployedVersion: "0.9.0",
				latestVersion:   "1.0.0",
			},
			status2: versions{
				approvedVersion: "1.0.0",
				deployedVersion: "1.0.0",
				latestVersion:   "1.0.0",
			},
			expected: false,
		},
		{
			name: "different latest version",
			status1: versions{
				approvedVersion: "1.0.0",
				deployedVersion: "0.9.0",
				latestVersion:   "1.0.0",
			},
			status2: versions{
				approvedVersion: "1.0.0",
				deployedVersion: "0.9.0",
				latestVersion:   "1.1.0",
			},
			expected: false,
		},
		{
			name: "all empty versions match",
			status1: versions{
				approvedVersion: "",
				deployedVersion: "",
				latestVersion:   "",
			},
			status2: versions{
				approvedVersion: "",
				deployedVersion: "",
				latestVersion:   "",
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			status1 := New(
				nil, nil, nil,
				tc.status1.approvedVersion,
				tc.status1.deployedVersion, "",
				tc.status1.latestVersion, "",
				"",
				&dashboard.Options{},
			)

			status2 := New(
				nil, nil, nil,
				tc.status2.approvedVersion,
				tc.status2.deployedVersion, "",
				tc.status2.latestVersion, "",
				"",
				&dashboard.Options{},
			)

			result := status1.SameVersions(status2)

			if result != tc.expected {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", result, tc.expected)
			}
		})
	}
}

func TestStatus_ApprovedVersion(t *testing.T) {
	versions := serviceinfo.ServiceInfo{
		DeployedVersion: "0.0.1",
		LatestVersion:   "0.0.3",
	}

	tests := []struct {
		name                          string
		hadApprovedVersion            string
		approving                     string
		latestVersionIsDeployedMetric metric.LatestVersionDeployedState
		wantMessages                  int
	}{
		{
			name:                          "Inherit version",
			hadApprovedVersion:            "0.0.0",
			approving:                     "0.0.0",
			latestVersionIsDeployedMetric: metric.LatestVersionUnactioned,
			wantMessages:                  0,
		},
		{
			name:                          "Approving LatestVersion",
			approving:                     versions.LatestVersion,
			latestVersionIsDeployedMetric: metric.LatestVersionApproved,
			wantMessages:                  1,
		},
		{
			name:                          "Skipping LatestVersion",
			approving:                     serviceinfo.SkippedVersion(versions.LatestVersion),
			latestVersionIsDeployedMetric: metric.LatestVersionSkipped,
			wantMessages:                  1,
		},
		{
			name:                          "Approving non-LatestVersion",
			approving:                     "0.0.2a",
			latestVersionIsDeployedMetric: metric.LatestVersionUnactioned,
			wantMessages:                  1,
		},
	}

	// The subtests change metric.UpdatesCurrent.
	metricsMu.RLock()
	t.Cleanup(metricsMu.RUnlock)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for _, writeDB := range []bool{true, false} {
				t.Run(fmt.Sprintf("writeToDB=%t", writeDB), func(t *testing.T) {
					t.Parallel()

					announceChannel := make(chan []byte, 4)
					databaseChannel := make(chan dbtype.Message, 4)
					status := New(
						announceChannel, databaseChannel, nil,
						tc.hadApprovedVersion,
						"", "",
						"", "",
						"",
						&dashboard.Options{
							WebURL: "https://example.com",
						},
					)
					status.Init(
						0, 0, 0,
						ServiceInfo{
							ID: t.Name(),
						},
						status.Dashboard,
					)
					status.SetLatestVersion(versions.LatestVersion, "", false)
					status.SetDeployedVersion(versions.DeployedVersion, "", false)

					status.SetApprovedVersion(tc.approving, writeDB)

					var wantMessages int
					if writeDB {
						wantMessages = tc.wantMessages
					}
					fieldTests := []test.FieldAssertion{
						{Name: "ApprovedVersion()", Got: status.ApprovedVersion(), Want: tc.approving, Mode: test.CompareEqual},
						{Name: "DeployedVersion()", Got: status.DeployedVersion(), Want: versions.DeployedVersion, Mode: test.CompareEqual},
						{Name: "LatestVersion()", Got: status.LatestVersion(), Want: versions.LatestVersion, Mode: test.CompareEqual},
					}
					if testErr := test.AssertFields(t, fieldTests, "Status"); testErr != nil {
						t.Fatal(testErr)
					}
					channelTests := []test.FieldAssertion{
						{Name: "len(Status.AnnounceChannel)", Got: len(status.AnnounceChannel), Want: wantMessages, Mode: test.CompareEqual},
						{Name: "len(Status.DatabaseChannel)", Got: len(status.DatabaseChannel), Want: wantMessages, Mode: test.CompareEqual},
					}
					if testErr := test.AssertFields(t, channelTests, ""); testErr != nil {
						t.Fatal(testErr)
					}

					got := int(testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(status.ServiceInfo.ID)))
					var want int
					if writeDB {
						want = int(tc.latestVersionIsDeployedMetric)
					}
					if got != want {
						t.Errorf(
							"metric.LatestVersionIsDeployed mismatch\ngot:  %d\nwant: %d",
							got, want,
						)
					}
				})
			}
		})
	}
}

func TestStatus_DeployedVersion(t *testing.T) {
	type values struct {
		versions                 serviceinfo.ServiceInfo
		deployedVersionTimestamp string
	}
	type args struct {
		version, timestamp string
	}

	tests := []struct {
		name        string
		hadVersions serviceinfo.ServiceInfo
		args        args
		want        values
	}{
		{
			name: "Inherit version",
			hadVersions: serviceinfo.ServiceInfo{
				ApprovedVersion: "0.0.2",
				DeployedVersion: "0.0.1",
				LatestVersion:   "0.0.3",
			},
			args: args{
				version:   "0.0.1",
				timestamp: "2020-01-01T00:00:00Z",
			},
			want: values{
				versions: serviceinfo.ServiceInfo{
					ApprovedVersion: "0.0.2",
					DeployedVersion: "0.0.1",
					LatestVersion:   "0.0.3",
				},
			},
		},
		{
			name: "Deploying ApprovedVersion - DeployedVersion becomes ApprovedVersion and resets ApprovedVersion",
			hadVersions: serviceinfo.ServiceInfo{
				ApprovedVersion: "0.0.2",
				DeployedVersion: "0.0.1",
				LatestVersion:   "0.0.3",
			},
			args: args{
				version: "0.0.2",
			},
			want: values{
				versions: serviceinfo.ServiceInfo{
					ApprovedVersion: "",
					DeployedVersion: "0.0.2",
					LatestVersion:   "0.0.3",
				},
			},
		},
		{
			name: "Deploying unknown Version - DeployedVersion becomes this version",
			hadVersions: serviceinfo.ServiceInfo{
				ApprovedVersion: "0.0.2",
				DeployedVersion: "0.0.1",
				LatestVersion:   "0.0.3",
			},
			args: args{
				version: "0.0.4-dev",
			},
			want: values{
				versions: serviceinfo.ServiceInfo{
					ApprovedVersion: "0.0.2",
					DeployedVersion: "0.0.4-dev",
					LatestVersion:   "0.0.3",
				},
			},
		},
		{
			name: "Deploying LatestVersion - DeployedVersion becomes this version",
			hadVersions: serviceinfo.ServiceInfo{
				ApprovedVersion: "0.0.2",
				DeployedVersion: "0.0.1",
				LatestVersion:   "0.0.3",
			},
			args: args{
				version: "0.0.3",
			},
			want: values{
				versions: serviceinfo.ServiceInfo{
					ApprovedVersion: "0.0.2",
					DeployedVersion: "0.0.3",
					LatestVersion:   "0.0.3",
				},
			},
		},
		{
			name: "Deploying X with timestamp - DeployedVersion and DeployedVersionTimestamp are set",
			hadVersions: serviceinfo.ServiceInfo{
				ApprovedVersion: "0.0.2",
				DeployedVersion: "0.0.1",
				LatestVersion:   "0.0.3",
			},
			args: args{
				version:   "0.0.4",
				timestamp: "2020-01-01T00:00:00Z",
			},
			want: values{
				versions: serviceinfo.ServiceInfo{
					ApprovedVersion: "0.0.2",
					DeployedVersion: "0.0.4",
					LatestVersion:   "0.0.3",
				},
				deployedVersionTimestamp: "2020-01-01T00:00:00Z",
			},
		},
	}

	// The subtests change metric.UpdatesCurrent.
	metricsMu.RLock()
	t.Cleanup(metricsMu.RUnlock)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for _, writeDB := range []bool{true, false} {
				t.Run(fmt.Sprintf("writeToDB=%t", writeDB), func(t *testing.T) {
					t.Parallel()

					dbChannel := make(chan dbtype.Message, 4)
					status := New(
						nil, dbChannel, nil,
						tc.hadVersions.ApprovedVersion,
						tc.hadVersions.DeployedVersion, "",
						tc.hadVersions.LatestVersion, "",
						"",
						&dashboard.Options{
							WebURL: "https://example.com",
						},
					)
					status.Init(
						0, 0, 0,
						ServiceInfo{
							ID: t.Name(),
						},
						status.Dashboard,
					)
					hadDeployedVersionTimestamp := status.deployedVersionTimestamp

					status.SetDeployedVersion(tc.args.version, tc.args.timestamp, writeDB)

					fieldTests := []test.FieldAssertion{
						{Name: "ApprovedVersion()", Got: status.ApprovedVersion(), Want: tc.want.versions.ApprovedVersion, Mode: test.CompareEqual},
						{Name: "DeployedVersion()", Got: status.DeployedVersion(), Want: tc.want.versions.DeployedVersion, Mode: test.CompareEqual},
						{Name: "LatestVersion()", Got: status.LatestVersion(), Want: tc.want.versions.LatestVersion, Mode: test.CompareEqual},
					}
					if testErr := test.AssertFields(t, fieldTests, "Status"); testErr != nil {
						t.Fatal(testErr)
					}

					if tc.hadVersions.DeployedVersion == tc.args.version {
						if timestamp := status.DeployedVersionTimestamp(); timestamp != hadDeployedVersionTimestamp {
							t.Errorf(
								"Status.DeployedVersionTimestamp() mismatch\ngot:  %q\nwant: %q",
								timestamp, hadDeployedVersionTimestamp,
							)
						}
					} else {
						if tc.want.deployedVersionTimestamp != "" {
							if timestamp := status.DeployedVersionTimestamp(); timestamp != tc.want.deployedVersionTimestamp {
								t.Errorf(
									"Status.DeployedVersionTimestamp() mismatch\n"+
										"got:  %q\nwant: %q",
									timestamp, tc.want.deployedVersionTimestamp,
								)
							}
						} else {
							timestamp := status.DeployedVersionTimestamp()
							d, err := time.Parse(time.RFC3339, timestamp)
							since := time.Since(d)
							if err != nil || since > time.Second {
								t.Errorf(
									"Status.DeployedVersionTimestamp() mismatch\n"+
										"got:  %q (%s ago, err=%v)\nwant: <1s ago",
									timestamp, since, err,
								)
							}
						}
					}

					got := int(testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(status.ServiceInfo.ID)))
					var want int
					if writeDB && status.LatestVersion() == status.DeployedVersion() {
						want = int(metric.LatestVersionDeployed)
					}
					if got != want {
						t.Errorf(
							"metric.LatestVersionIsDeployed mismatch\ngot:  %d\nwant: %d",
							got, want,
						)
					}
				})
			}
		})
	}
}

func TestStatus_LatestVersion(t *testing.T) {
	type values struct {
		version, timestamp string
	}

	lastQueried := "2021-01-01T00:00:00Z"
	tests := []struct {
		name      string
		had, args values
		want      *values // Default to args.
	}{
		{
			name: "same version",
			had: values{
				version: "1.2.3", timestamp: "2020-01-01T00:00:00Z",
			},
			args: values{
				version: "1.2.3", timestamp: "2020-01-01T00:00:00Z",
			},
		},
		{
			name: "timestamp/Empty == Set to lastQueried",
			had: values{
				version: "0.0.0", timestamp: "2021-01-01T00:00:00Z",
			},
			args: values{
				version: "0.0.1", timestamp: "",
			},
			want: &values{
				version: "0.0.1", timestamp: lastQueried,
			},
		},
		{
			name: "timestamp/Given == Set to value given",
			had: values{
				version: "0.0.0", timestamp: "2022-01-01T00:00:00Z",
			},
			args: values{
				version: "0.0.1", timestamp: "2022-01-01T00:00:00Z",
			},
		},
	}

	// The subtests change metric.UpdatesCurrent.
	metricsMu.RLock()
	t.Cleanup(metricsMu.RUnlock)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for _, writeDB := range []bool{true, false} {
				t.Run(fmt.Sprintf("writeToDB=%t", writeDB), func(t *testing.T) {
					t.Parallel()

					dbChannel := make(chan dbtype.Message, 8)
					status := New(
						nil, dbChannel, nil,
						"",
						"0.0.0", "",
						tc.had.version, tc.had.timestamp,
						lastQueried,
						&dashboard.Options{
							WebURL: "https://example.com",
						},
					)
					status.Init(
						0, 0, 0,
						ServiceInfo{
							ID: t.Name(),
						},
						status.Dashboard,
					)
					wantValues := tc.want
					if wantValues == nil {
						wantValues = &tc.args
					}

					status.SetLatestVersion(tc.args.version, tc.args.timestamp, writeDB)

					fieldTests := []test.FieldAssertion{
						{Name: "LatestVersion()", Got: status.LatestVersion(), Want: wantValues.version, Mode: test.CompareEqual},
						{Name: "LatestVersionTimestamp()", Got: status.LatestVersionTimestamp(), Want: wantValues.timestamp, Mode: test.CompareEqual},
					}
					if testErr := test.AssertFields(t, fieldTests, "Status"); testErr != nil {
						t.Fatal(testErr)
					}

					got := int(testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(status.ServiceInfo.ID)))
					var want int
					if writeDB && status.LatestVersion() == status.DeployedVersion() {
						want = int(metric.LatestVersionDeployed)
					}
					if got != want {
						t.Errorf(
							"metric.LatestVersionIsDeployed mismatch\ngot:  %d\nwant: %d",
							got, want,
						)
					}
				})
			}
		})
	}
}

func TestStatus_RegexMissesContent(t *testing.T) {
	status := Status{}
	var want uint = 0
	for i := range 3 {
		status.RegexMissContent()
		want++

		if got := status.RegexMissesContent(); got != want {
			t.Errorf(
				"miss %d: Status.RegexMissesContent() mismatch\ngot:  %d\nwant: %d",
				i, got, want,
			)
		}
	}

	status.ResetRegexMisses()

	want = 0
	if got := status.RegexMissesContent(); got != want {
		t.Errorf(
			"after ResetRegexMisses: Status.RegexMissesContent() mismatch\ngot:  %d\nwant: %d",
			got, want,
		)
	}
}

func TestStatus_RegexMissesVersion(t *testing.T) {
	status := Status{}
	var want uint = 0
	for i := range 3 {
		status.RegexMissVersion()
		want++

		if got := status.RegexMissesVersion(); got != want {
			t.Errorf(
				"miss %d: Status.RegexMissesVersion() mismatch\ngot:  %d\nwant: %d",
				i, got, want,
			)
		}
	}

	status.ResetRegexMisses()

	want = 0
	if got := status.RegexMissesVersion(); got != want {
		t.Errorf(
			"after ResetRegexMisses: Status.RegexMissesVersion() mismatch\ngot:  %d\nwant: %d",
			got, want,
		)
	}
}

func TestStatus_SetDeleting(t *testing.T) {
	status := Status{}

	status.SetDeleting()

	if !status.Deleting() {
		t.Error("first SetDeleting: Status.Deleting() mismatch\ngot:  false\nwant: true")
	}

	status.SetDeleting()

	if !status.Deleting() {
		t.Error("second SetDeleting: Status.Deleting() mismatch\ngot:  false\nwant: true")
	}
}

func TestStatus_SendAnnounce(t *testing.T) {
	tests := []struct {
		name       string
		deleting   bool
		nilChannel bool
	}{
		{name: "active"},
		{name: "deleting", deleting: true},
		{name: "nil channel", nilChannel: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			announceChannel := make(chan []byte, 4)
			status := New(
				announceChannel, nil, nil,
				"",
				"", "",
				"", "",
				"",
				&dashboard.Options{},
			)
			if tc.nilChannel {
				status.AnnounceChannel = nil
			}
			if tc.deleting {
				status.SetDeleting()
			}

			status.SendAnnounce([]byte{})

			got := 0
			if status.AnnounceChannel != nil {
				got = len(status.AnnounceChannel)
			}
			want := 1
			if tc.deleting || tc.nilChannel {
				want = 0
			}
			if got != want {
				t.Errorf(
					"Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					got, want,
				)
			}
		})
	}
}

func TestStatus_SendDatabase(t *testing.T) {
	tests := []struct {
		name       string
		deleting   bool
		nilChannel bool
	}{
		{name: "active"},
		{name: "deleting", deleting: true},
		{name: "nil channel", nilChannel: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			databaseChannel := make(chan dbtype.Message, 4)
			status := New(
				nil, databaseChannel, nil,
				"",
				"", "",
				"", "",
				"",
				&dashboard.Options{},
			)
			if tc.nilChannel {
				status.DatabaseChannel = nil
			}
			if tc.deleting {
				status.SetDeleting()
			}

			status.sendDatabase(&dbtype.Message{})

			want := 1
			if tc.deleting || tc.nilChannel {
				want = 0
			}
			got := 0
			if status.DatabaseChannel != nil {
				got = len(status.DatabaseChannel)
			}
			if got != want {
				t.Errorf(
					"Status.DatabaseChannel message count mismatch\ngot:  %d\nwant: %d",
					got, want,
				)
			}
		})
	}
}

func TestStatus_SendSave(t *testing.T) {
	tests := []struct {
		name       string
		deleting   bool
		nilChannel bool
	}{
		{name: "active"},
		{name: "deleting", deleting: true},
		{name: "nil channel", nilChannel: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			saveChannel := make(chan bool, 4)
			status := New(
				nil, nil, saveChannel,
				"",
				"", "",
				"", "",
				"",
				&dashboard.Options{},
			)
			if tc.nilChannel {
				status.SaveChannel = nil
			}
			if tc.deleting {
				status.SetDeleting()
			}

			status.SendSave()

			want := 1
			if tc.deleting || tc.nilChannel {
				want = 0
			}
			got := 0
			if status.SaveChannel != nil {
				got = len(status.SaveChannel)
			}
			if got != want {
				t.Errorf("Status.SaveChannel message count mismatch\ngot:  %d\nwant: %d", got, want)
			}
		})
	}
}

func TestSetLatestVersionIsDeployedMetric(t *testing.T) {
	tests := []struct {
		name     string
		versions serviceinfo.ServiceInfo
		want     metric.LatestVersionDeployedState
	}{
		{
			name: "latest version is deployed",
			versions: serviceinfo.ServiceInfo{
				LatestVersion:   "1.2.3",
				DeployedVersion: "1.2.3",
			},
			want: metric.LatestVersionDeployed,
		},
		{
			name: "latest version is not deployed/only",
			versions: serviceinfo.ServiceInfo{
				LatestVersion:   "1.2.3",
				DeployedVersion: "1.2.4",
			},
			want: metric.LatestVersionUnactioned,
		},
		{
			name: "latest version is not deployed/is approved",
			versions: serviceinfo.ServiceInfo{
				ApprovedVersion: "1.2.3",
				LatestVersion:   "1.2.3",
				DeployedVersion: "1.2.4",
			},
			want: metric.LatestVersionApproved,
		},
		{
			name: "latest version is not deployed/is skipped",
			versions: serviceinfo.ServiceInfo{
				ApprovedVersion: serviceinfo.SkippedVersion("1.2.3"),
				LatestVersion:   "1.2.3",
				DeployedVersion: "1.2.4",
			},
			want: metric.LatestVersionSkipped,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.versions.ID = tc.name

			setLatestVersionIsDeployedMetric(tc.versions)
			got := metric.LatestVersionDeployedState(
				testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(tc.versions.ID)),
			)

			if got != tc.want {
				t.Errorf(
					"metric.LatestVersionIsDeployed mismatch\ngot:  %v\nwant: %v",
					got, tc.want,
				)
			}
		})
	}
}

type fuzzSetVersionsOp uint8

const (
	opSetApprovedLatest fuzzSetVersionsOp = iota
	opSetApprovedSkip
	opSetLatest
	opSetDeployed
)

func (o fuzzSetVersionsOp) String() string {
	return [...]string{"setApprovedLatest", "setApprovedSkip", "setLatest", "setDeployed"}[o]
}

func FuzzSetVersions(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3}, []byte{0, 1, 2, 3})

	svcStatus := testStatus()
	svcStatus.ServiceInfo.LatestVersion = "0.0.0"
	svcStatus.ServiceInfo.DeployedVersion = "0.0.0"
	svcStatus.ServiceInfo.ApprovedVersion = ""

	announceChan := make(chan []byte, 4)
	svcStatus.AnnounceChannel = announceChan
	databaseChan := make(chan dbtype.Message, 4)
	svcStatus.DatabaseChannel = databaseChan
	saveChan := make(chan bool, 4)
	svcStatus.SaveChannel = saveChan
	stopDrain := make(chan struct{})
	f.Cleanup(func() {
		close(stopDrain)
	})
	// Drain to prevent blocking.
	go func() {
		for {
			select {
			case <-announceChan:
			case <-databaseChan:
			case <-saveChan:
			case <-stopDrain:
				return
			}
		}
	}()

	hadUpdatesCurrentAvailable := testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("AVAILABLE"))
	hadUpdatesCurrentSkipped := testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("SKIPPED"))

	f.Fuzz(func(t *testing.T, opsRaw []byte, versionsRaw []byte) {
		if len(opsRaw) == 0 || len(versionsRaw) == 0 {
			return
		}

		versions := []string{
			"1.0.0",
			"1.1.0",
			"2.0.0",
		}

		for i := range len(opsRaw) {
			op := fuzzSetVersionsOp(opsRaw[i] % 4)
			version := versions[int(versionsRaw[i%len(versionsRaw)])%len(versions)]
			t.Logf("op %d: kind=%s, version=%q", i, op, version)

			switch op {
			case opSetApprovedLatest:
				svcStatus.SetApprovedVersion(version, true)

			case opSetApprovedSkip:
				svcStatus.SetApprovedVersion(
					serviceinfo.SkippedVersion(svcStatus.ServiceInfo.LatestVersion),
					true,
				)

			case opSetLatest:
				svcStatus.SetLatestVersion(version, "", true)

			case opSetDeployed:
				svcStatus.SetDeployedVersion(version, "", true)
			}

			assertSetVersionsResult(t, svcStatus, hadUpdatesCurrentAvailable, hadUpdatesCurrentSkipped)
		}
	})
}

func assertSetVersionsResult(t *testing.T, s *Status, hadAvailable, hadSkipped float64) {
	t.Helper()

	info := s.GetServiceInfo()
	state := fmt.Sprintf(
		"approved=%q, deployed=%q, latest=%q",
		info.ApprovedVersion, info.DeployedVersion, info.LatestVersion,
	)

	// Approved must not equal Latest if Latest == Deployed.
	if info.LatestVersion == info.DeployedVersion &&
		info.ApprovedVersion == info.LatestVersion {
		t.Fatalf(
			"ServiceInfo.ApprovedVersion mismatch\ngot:  %q\nwant: NOT %q\nstate: %s",
			info.ApprovedVersion, info.LatestVersion, state,
		)
	}

	if serviceinfo.SkippedVersion(info.LatestVersion) == info.DeployedVersion &&
		info.ApprovedVersion != "" {
		t.Fatalf(
			"ServiceInfo.ApprovedVersion mismatch\ngot:  %q\nwant: \"\" (reset)\nstate: %s",
			info.ApprovedVersion, state,
		)
	}

	// Timestamps must exist if version exists.
	if info.DeployedVersion != "" && s.DeployedVersionTimestamp() == "" {
		t.Fatalf(
			"Status.DeployedVersionTimestamp() mismatch\ngot:  \"\"\nwant: non-empty\nstate: %s",
			state,
		)
	}

	want := hadAvailable
	if info.LatestVersion != info.DeployedVersion {
		want += 1
	}
	got := testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("AVAILABLE"))
	if got != want {
		t.Fatalf(
			"metric.UpdatesCurrent.WithLabelValues(\"AVAILABLE\") mismatch\n"+
				"got:  %f\nwant: %f\nstate: %s",
			got, want, state,
		)
	}

	want = hadSkipped
	if info.ApprovedVersion == serviceinfo.SkippedVersion(info.LatestVersion) {
		want += 1
	}
	got = testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("SKIPPED"))
	if got != want {
		t.Fatalf(
			"metric.UpdatesCurrent.WithLabelValues(\"SKIPPED\") mismatch\n"+
				"got:  %f\nwant: %f\nstate: %s",
			got, want, state,
		)
	}
}

func TestStatus_InitMetrics_then_DeleteMetrics(t *testing.T) {
	// The test counts the gauge's series.
	metricsMu.Lock()
	t.Cleanup(metricsMu.Unlock)

	status := testStatus()
	status.ServiceInfo.ID = "TestStatus_InitMetrics_DeleteMetrics"
	status.ServiceInfo.LatestVersion = "1.1.0"
	status.ServiceInfo.DeployedVersion = "1.0.0"
	status.ServiceInfo.ApprovedVersion = serviceinfo.SkippedVersion("1.0.0")

	had := testutil.CollectAndCount(metric.LatestVersionIsDeployed)

	status.InitMetrics()

	test.AssertMetricCount(
		t, "Status.InitMetrics() LatestVersionIsDeployed",
		metric.LatestVersionIsDeployed, had+1,
	)

	// 0 (unactioned): the skip is of 1.0.0, not latest 1.1.0.
	if got, want := testutil.ToFloat64(
		metric.LatestVersionIsDeployed.WithLabelValues(status.ServiceInfo.ID),
	), 0.0; got != want {
		t.Fatalf(
			"after InitMetrics: metric.LatestVersionIsDeployed mismatch\ngot:  %f\nwant: %f",
			got, want,
		)
	}

	status.DeleteMetrics()

	test.AssertMetricCount(
		t, "Status.DeleteMetrics() LatestVersionIsDeployed",
		metric.LatestVersionIsDeployed, had,
	)
}
