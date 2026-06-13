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
	// GIVEN: we have channels.
	announceChannel := make(chan []byte, 4)
	databaseChannel := make(chan dbtype.Message, 4)
	saveChannel := make(chan bool, 4)

	// WHEN: NewDefaults is called.
	statusDefaults := NewDefaults(announceChannel, databaseChannel, saveChannel)

	prefix := fmt.Sprintf("%s\nNewDefaults()", packageName)

	// THEN: the channels are handed out
	fieldTests := []test.FieldAssertion{
		{Name: "AnnounceChannel", Got: statusDefaults.AnnounceChannel, Want: announceChannel, Mode: test.CompareSamePointer},
		{Name: "DatabaseChannel", Got: statusDefaults.DatabaseChannel, Want: databaseChannel, Mode: test.CompareSamePointer},
		{Name: "SaveChannel", Got: statusDefaults.SaveChannel, Want: saveChannel, Mode: test.CompareSamePointer},
	}
	if err := test.AssertFields(t, fieldTests, prefix, "Status"); err != nil {
		t.Fatal(err)
	}
}

func TestStatus_Unmarshal(t *testing.T) {
	// GIVEN: a Status.
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

			// WHEN: UnmarshalX is called on the Status.
			data := []byte("")
			var status Status
			switch tc.format {
			case "YAML":
				_ = status.UnmarshalYAML(data)
			case "JSON":
				_ = status.UnmarshalJSON(data)
			default:
				t.Fatalf("unknown format %q", tc.format)
			}

			// THEN: the mutex is correctly handed to the ServiceInfo.
			_ = status.GetServiceInfo()

			// AND: when the mutex is locked, GetServiceInfo is held.
			unlockedAtChan := make(chan time.Time, 1)
			gotServiceInfoAtChan := make(chan time.Time, 1)
			status.mu.Lock()
			go func() {
				status.GetServiceInfo()
				gotServiceInfoAtChan <- time.Now()
			}()
			go func() {
				time.Sleep(100 * time.Millisecond)
				status.mu.Unlock()
				unlockedAtChan <- time.Now()
			}()
			unlockedAt := <-unlockedAtChan
			gotServiceInfoAt := <-gotServiceInfoAtChan
			if gotServiceInfoAt.Before(unlockedAt) {
				t.Errorf(
					"%s\nStatus.GetServiceInfo() was not held while mutex was locked!\ngot:  %v\nwant: %v",
					packageName, gotServiceInfoAt, unlockedAt,
				)
			}
		})
	}
}

func TestStatus_Copy(t *testing.T) {
	// GIVEN: a Status.
	announceChannel := make(chan []byte, 4)
	databaseChannel := make(chan dbtype.Message, 4)
	saveChannel := make(chan bool, 4)
	approvedVersion := "1.0.0"
	deployedVersion := "0.9.0"
	deployedVersionTimestamp := "2023-01-01T00:00:00Z"
	latestVersion := "1.0.0"
	latestVersionTimestamp := "2023-01-02T00:00:00Z"
	lastQueried := "2023-01-03T00:00:00Z"
	status := New(
		announceChannel, databaseChannel, saveChannel,
		approvedVersion,
		deployedVersion, deployedVersionTimestamp,
		latestVersion, latestVersionTimestamp,
		lastQueried,
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
			// WHEN: Copy is called on it.
			copiedStatus := status.Copy(tc.copyChannels)

			prefix := fmt.Sprintf(
				"%s\nStatus.Copy(withChannels=%t)",
				packageName, tc.copyChannels,
			)

			// THEN: the copied Status should have the same values as the original.
			fieldTests := []test.FieldAssertion{
				{Name: "ApprovedVersion", Got: copiedStatus.ServiceInfo.ApprovedVersion, Want: status.ServiceInfo.ApprovedVersion, Mode: test.CompareEqual},
				{Name: "DeployedVersion", Got: copiedStatus.ServiceInfo.DeployedVersion, Want: status.ServiceInfo.DeployedVersion, Mode: test.CompareEqual},
				{Name: "DeployedVersionTimestamp", Got: copiedStatus.deployedVersionTimestamp, Want: status.deployedVersionTimestamp, Mode: test.CompareEqual},
				{Name: "LatestVersion", Got: copiedStatus.ServiceInfo.LatestVersion, Want: status.ServiceInfo.LatestVersion, Mode: test.CompareEqual},
				{Name: "LatestVersionTimestamp", Got: copiedStatus.latestVersionTimestamp, Want: status.latestVersionTimestamp, Mode: test.CompareEqual},
				{Name: "LastQueried", Got: copiedStatus.lastQueried, Want: status.lastQueried, Mode: test.CompareEqual},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Status"); err != nil {
				t.Fatal(err)
			}

			// AND: the channels are only copied over if copyChannels is true.
			fieldTests = []test.FieldAssertion{
				{Name: "AnnounceChannel", Got: copiedStatus.AnnounceChannel, Want: status.AnnounceChannel, Mode: test.CompareSamePointer},
				{Name: "DatabaseChannel", Got: copiedStatus.DatabaseChannel, Want: status.DatabaseChannel, Mode: test.CompareSamePointer},
				{Name: "SaveChannel", Got: copiedStatus.SaveChannel, Want: status.SaveChannel, Mode: test.CompareSamePointer},
			}
			if !tc.copyChannels {
				for i := range fieldTests {
					fieldTests[i].Want = nil
					fieldTests[i].Check = func() (got any, want any, ok bool) {
						return got, want, got == nil
					}
				}
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Status"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestStatus_Copy__nil(t *testing.T) {
	// GIVEN: a nil Status.
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
			// WHEN: Copy is called on it.
			copiedStatus := status.Copy(tc.copyChannels)

			// THEN: nil is always returned.
			if copiedStatus != nil {
				t.Fatalf("%s\nStatus(nil).Copy(%t) mismatch\ngot:  %v\nwant: nil",
					packageName, tc.copyChannels, copiedStatus,
				)
			}
		})
	}
}

func TestStatus_String(t *testing.T) {
	// GIVEN: a Status.
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
						failsBase: failsBase{
							fails: map[string]*bool{
								"bash": test.Ptr(false),
								"bish": nil,
								"bosh": test.Ptr(true),
							},
						},
					},
					Command: FailsCommand{
						fails: []*bool{
							nil,
							test.Ptr(false),
							test.Ptr(true),
						},
					},
					WebHook: FailsWebHook{
						failsBase: failsBase{
							fails: map[string]*bool{
								"bar": nil,
								"foo": test.Ptr(false),
							},
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
						failsBase: failsBase{
							fails: map[string]*bool{
								"bish": nil,
								"bash": test.Ptr(false),
								"bosh": test.Ptr(true),
							},
						},
					},
					Command: FailsCommand{
						fails: []*bool{
							nil,
							test.Ptr(false),
							test.Ptr(true),
						},
					},
					WebHook: FailsWebHook{
						failsBase: failsBase{
							fails: map[string]*bool{
								"foo": test.Ptr(false),
								"bar": nil,
							},
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

			{ // RegEz misses.
				for i := 0; i < tc.regexMissesContent; i++ {
					tc.status.RegexMissContent()
				}
				for i := 0; i < tc.regexMissesVersion; i++ {
					tc.status.RegexMissVersion()
				}
			}

			// WHEN: the Status is stringified with String.
			got := tc.status.String()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nStatus.String() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestStatus_Init(t *testing.T) {
	// GIVEN: we have a Status.
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

			// WHEN: Init is called.
			status.Init(
				tc.commands, tc.shoutrrrs, tc.webhooks,
				ServiceInfo{
					ID:   tc.serviceID,
					Name: tc.name,
				},
				&dashboard.Options{
					OptionsBase: dashboard.OptionsBase{
						WebURL: tc.webURL,
					},
				},
			)

			prefix := fmt.Sprintf("%s\nStatus.Init()", packageName)

			// THEN: the Status is initialised as expected:
			// 	ServiceID:
			if status.ServiceInfo.ID != tc.serviceID {
				t.Errorf(
					"%s .ServiceID mismatch\ngot:  %v\nwant: %v",
					packageName, status.ServiceInfo.ID, tc.serviceID,
				)
			}
			// 	Shoutrrr:
			want := 0
			if got := status.Fails.Shoutrrr.Length(); got != want {
				t.Errorf(
					"%s .Fails.Shoutrrr initial length mismatch\ngot:  %d\nwant: %d",
					prefix, got, want,
				)
			} else {
				for i := 0; i < tc.shoutrrrs; i++ {
					failed := false
					status.Fails.Shoutrrr.Set(fmt.Sprint(i), &failed)
				}
				if got := status.Fails.Shoutrrr.Length(); got != tc.shoutrrrs {
					t.Errorf(
						"%s .Fails.Shoutrrr capacity mismatch\ngot:  %d\nwant: %d",
						prefix, got, tc.shoutrrrs,
					)
				}
			}
			// 	Command:
			if got := status.Fails.Command.Length(); got != tc.commands {
				t.Errorf(
					"%s .Fails.Command length mismatch\ngot:  %d\nwant: %d",
					prefix, got, tc.commands,
				)
			}
			// 	WebHook:
			want = 0
			if got := status.Fails.WebHook.Length(); got != want {
				t.Errorf(
					"%s .Fails.WebHook initial length mismatch\ngot:  %d\nwant: %d",
					prefix, got, want,
				)
			} else {
				for i := 0; i < tc.webhooks; i++ {
					failed := false
					status.Fails.WebHook.Set(fmt.Sprint(i), &failed)
				}
				if got := status.Fails.WebHook.Length(); got != tc.webhooks {
					t.Errorf(
						"%s .Fails.WebHook capacity mismatch\ngot:  %d\nwant: %d",
						prefix, got, tc.webhooks,
					)
				}
			}
		})
	}
}

func TestStatus_SetAnnounceChannel(t *testing.T) {
	// GIVEN: a Status with an initial AnnounceChannel.
	initialChannel := make(chan []byte, 4)
	status := New(
		initialChannel, nil, nil,
		"",
		"", "",
		"", "",
		"",
		&dashboard.Options{},
	)

	// WHEN: SetAnnounceChannel is called with a new channel.
	newChannel := make(chan []byte, 4)
	status.SetAnnounceChannel(newChannel)

	prefix := fmt.Sprintf(
		"%s\nStatus.SetAnnounceChannel(chan=%p)",
		packageName, newChannel,
	)

	// THEN: the AnnounceChannel should be updated to the new channel.
	if status.AnnounceChannel != newChannel {
		t.Errorf(
			"%s mismatch\ngot:  %p\nwant: %p",
			prefix, status.AnnounceChannel, newChannel,
		)
	}
}

func TestStatus_GetServiceInfo(t *testing.T) {
	// GIVEN: a Status.
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

	time.Sleep(10 * time.Millisecond)
	time.Sleep(time.Second)

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

	// When ServiceInfo is called on it.
	got := status.GetServiceInfo()

	// THEN: we get the correct ServiceInfo.
	gotStr := decode.ToJSONString(got)
	wantStr := decode.ToJSONString(want)
	if gotStr != wantStr {
		t.Errorf(
			"%s\nStatus.ServiceInfo() mismatch\ngot:  %#v\nwant: %#v",
			packageName, gotStr, wantStr,
		)
	}
}

func TestStatus_RefreshServiceInfo(t *testing.T) {
	// GIVEN: a Status with dashboard templates and a latest version.
	latestVersion := "1.2.3"
	status := Status{}
	status.Init(
		0, 0, 0,
		ServiceInfo{ID: "refresh-test"},
		&dashboard.Options{
			OptionsBase: dashboard.OptionsBase{
				Icon:       "https://example.com/{{ version }}.png",
				IconLinkTo: "https://example.com/link/{{ version }}",
				WebURL:     "https://example.com/{{ version }}",
			},
		},
	)
	status.SetLatestVersion(latestVersion, "", false)

	// AND: stale ServiceInfo fields.
	status.ServiceInfo.Icon = "stale"
	status.ServiceInfo.IconLinkTo = "stale"
	status.ServiceInfo.WebURL = "stale"

	// WHEN: RefreshServiceInfo is called.
	status.RefreshServiceInfo()

	prefix := fmt.Sprintf("%s\nStatus.RefreshServiceInfo()", packageName)

	// THEN: ServiceInfo fields are refreshed from the dashboard templates.
	fieldTests := []test.FieldAssertion{
		{Name: "Icon", Got: status.ServiceInfo.Icon, Want: "https://example.com/1.2.3.png"},
		{Name: "IconLinkTo", Got: status.ServiceInfo.IconLinkTo, Want: "https://example.com/link/1.2.3"},
		{Name: "WebURL", Got: status.ServiceInfo.WebURL, Want: "https://example.com/1.2.3"},
	}
	if err := test.AssertFields(t, fieldTests, prefix, "ServiceInfo"); err != nil {
		t.Fatal(err)
	}
}

func TestStatus_GetWebURL(t *testing.T) {
	// GIVEN: we have a Status.
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
					OptionsBase: dashboard.OptionsBase{
						WebURL: tc.webURL,
					},
				},
			)
			status.SetLatestVersion(latestVersion, "", false)

			// WHEN: GetWebURL is called.
			got := status.GetWebURL()

			// THEN: the returned WebURL is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nStatus.WebURL() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestStatus_SetAndGetLastQueried(t *testing.T) {
	// GIVEN: a lastQueried string.
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

			// AND: a Status.
			status := testStatus()

			// WHEN: SetLastQueried is called.
			status.SetLastQueried(tc.lastQueried)

			prefix := fmt.Sprintf(
				"%s\nStatus.SetLastQueried(timestamp=%q)",
				packageName, tc.lastQueried,
			)

			// THEN: LastQueried will have been set to the given timestamp (if provided).
			lastQueried := status.LastQueried()
			if tc.lastQueried != "" {
				if lastQueried != tc.lastQueried {
					t.Errorf(
						"%s mismatch on LastQueried()\ngot:  %s\nwant: %s",
						prefix, lastQueried, tc.lastQueried,
					)
				}
			} else {
				// AND: if no timestamp was provided, then LastQueried() should be <1s ago.
				lastQueriedTime, _ := time.Parse(time.RFC3339, lastQueried)
				if since := time.Since(lastQueriedTime); since > time.Second {
					t.Errorf(
						"%s mismatch on LastQueried() (too far in the past)\ngot:  %s ago\nwant: <1s",
						prefix, since,
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
	// GIVEN: different Status version combinations.
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

			// WHEN: comparing versions.
			result := status1.SameVersions(status2)

			// THEN: the result matches expected.
			if result != tc.expected {
				t.Errorf(
					"%s\nStatus.SameVersions(%+v) mismatch (on %+v)\ngot:  %t\nwant: %t",
					packageName, status1, status2,
					result, tc.expected,
				)
			}
		})
	}
}

func TestStatus_ApprovedVersion(t *testing.T) {
	versions := serviceinfo.ServiceInfo{
		DeployedVersion: "0.0.1",
		LatestVersion:   "0.0.3",
	}
	// GIVEN: a Status.
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

	// Changing UpdatesCurrent.
	metricsMu.RLock()
	t.Cleanup(metricsMu.RUnlock)

	for _, tc := range tests {

		subTests := []bool{true, false}
		for _, writeDB := range subTests {
			name := fmt.Sprintf("%s, writeToDB=%t", tc.name, writeDB)
			t.Run(name, func(t *testing.T) {
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
						OptionsBase: dashboard.OptionsBase{
							WebURL: "https://example.com",
						},
					},
				)
				status.Init(
					0, 0, 0,
					ServiceInfo{
						ID: name,
					},
					status.Dashboard,
				)
				status.SetLatestVersion(versions.LatestVersion, "", false)
				status.SetDeployedVersion(versions.DeployedVersion, "", false)

				// WHEN: SetApprovedVersion is called.
				status.SetApprovedVersion(tc.approving, writeDB)

				prefix := fmt.Sprintf(
					"%s\nStatus(%+v).SetApprovedVersion(approving=%q, writeDB=%t)",
					packageName, versions, tc.approving, writeDB,
				)

				// THEN: the Status is as expected:
				var wantMessages int
				if writeDB {
					wantMessages = tc.wantMessages
				}
				fieldTests := []test.FieldAssertion{
					{Name: "ApprovedVersion", Got: status.ApprovedVersion(), Want: tc.approving, Mode: test.CompareEqual},
					{Name: "DeployedVersion", Got: status.DeployedVersion(), Want: versions.DeployedVersion, Mode: test.CompareEqual},
					{Name: "LatestVersion", Got: status.LatestVersion(), Want: versions.LatestVersion, Mode: test.CompareEqual},
					{Name: "AnnounceChannel", Got: len(status.AnnounceChannel), Want: wantMessages, Mode: test.CompareEqual},
					{Name: "DatabaseChannel", Got: len(status.DatabaseChannel), Want: wantMessages, Mode: test.CompareEqual},
				}
				if err := test.AssertFields(t, fieldTests, prefix, "Status"); err != nil {
					t.Fatal(err)
				}

				// AND: LatestVersionIsDeployedVersion metric is updated.
				got := int(testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(status.ServiceInfo.ID)))
				var want int
				if writeDB {
					want = int(tc.latestVersionIsDeployedMetric)
				}
				if got != want {
					t.Errorf(
						"%s LatestVersionIsDeployedVersion metric mismatch\ngot:  %d\nwant: %d",
						prefix, got, want,
					)
				}
			})
		}
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
	// GIVEN: a Status.
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

	// Changing UpdatesCurrent.
	metricsMu.RLock()
	t.Cleanup(metricsMu.RUnlock)

	for _, tc := range tests {

		subTests := []bool{true, false}
		for _, writeDB := range subTests {
			name := fmt.Sprintf("%s, writeToDB=%t", tc.name, writeDB)
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				dbChannel := make(chan dbtype.Message, 4)
				status := New(
					nil, dbChannel, nil,
					tc.hadVersions.ApprovedVersion,
					tc.hadVersions.DeployedVersion, "",
					tc.hadVersions.LatestVersion, "",
					"",
					&dashboard.Options{
						OptionsBase: dashboard.OptionsBase{
							WebURL: "https://example.com",
						},
					},
				)
				status.Init(
					0, 0, 0,
					ServiceInfo{
						ID: name,
					},
					status.Dashboard,
				)
				hadDeployedVersionTimestamp := status.deployedVersionTimestamp

				// WHEN: SetDeployedVersion is called on it.
				status.SetDeployedVersion(tc.args.version, tc.args.timestamp, writeDB)

				prefix := fmt.Sprintf(
					"%s\nStatus(%+v).SetDeployedVersion(version=%q, timestamp=%q, writeDB=%t)",
					packageName, tc.hadVersions, tc.args.version, tc.args.timestamp, writeDB,
				)

				// THEN: DeployedVersion is set to this version.
				fieldTests := []test.FieldAssertion{
					{Name: "ApprovedVersion", Got: status.ApprovedVersion(), Want: tc.want.versions.ApprovedVersion, Mode: test.CompareEqual},
					{Name: "DeployedVersion", Got: status.DeployedVersion(), Want: tc.want.versions.DeployedVersion, Mode: test.CompareEqual},
					{Name: "LatestVersion", Got: status.LatestVersion(), Want: tc.want.versions.LatestVersion, Mode: test.CompareEqual},
				}
				if err := test.AssertFields(t, fieldTests, prefix, "Status"); err != nil {
					t.Fatal(err)
				}

				// AND: the DeployedVersionTimestamp is unchanged when DeployedVersion is unchanged.
				if tc.hadVersions.DeployedVersion == tc.args.version {
					if timestamp := status.DeployedVersionTimestamp(); timestamp != hadDeployedVersionTimestamp {
						t.Errorf(
							"%s Status.DeployedVersionTimestamp() changed when DeployedVersion did not\ngot:  %s\nwant: %s",
							prefix, timestamp, hadDeployedVersionTimestamp,
						)
					}
				} else {
					// AND: the DeployedVersionTimestamp is set to the provided date (when provided).
					if tc.want.deployedVersionTimestamp != "" {
						if timestamp := status.DeployedVersionTimestamp(); timestamp != tc.want.deployedVersionTimestamp {
							t.Errorf(
								"%s Status.DeployedVersionTimestamp() did not change when DeployedVersion did\ngot:  %s\nwant: %s",
								prefix, timestamp, tc.want.deployedVersionTimestamp,
							)
						}
					} else {
						// AND: the DeployedVersionTimestamp is set to current time (when no releaseDate was given).
						d, _ := time.Parse(time.RFC3339, status.DeployedVersionTimestamp())
						since := time.Since(d)
						if since > time.Second {
							t.Errorf(
								"%s DeployedVersionTimestamp() was %v ago, not recent enough!",
								prefix, since,
							)
						}
					}
				}

				// AND: the LatestVersionIsDeployedVersion metric is updated as expected (only when we're writing to DB).
				got := int(testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(status.ServiceInfo.ID)))
				var want int
				if writeDB && status.LatestVersion() == status.DeployedVersion() {
					want = int(metric.LatestVersionDeployed)
				}
				if got != want {
					t.Errorf(
						"%s LatestVersionIsDeployedVersion metric mismatch\ngot:  %d\nwant: %d",
						prefix, got, want,
					)
				}
			})
		}
	}
}

func TestStatus_LatestVersion(t *testing.T) {
	type values struct {
		version, timestamp string
	}
	// GIVEN: a Status.
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
			name: "timestamp - Empty == Set to lastQueried",
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
			name: "Timestamp - Given == Set to value given",
			had: values{
				version: "0.0.0", timestamp: "2022-01-01T00:00:00Z",
			},
			args: values{
				version: "0.0.1", timestamp: "2022-01-01T00:00:00Z",
			},
		},
	}

	// Changing UpdatesCurrent.
	metricsMu.RLock()
	t.Cleanup(metricsMu.RUnlock)

	for _, tc := range tests {

		subTests := []bool{true, false}
		for _, writeDB := range subTests {
			name := fmt.Sprintf(
				"%s, writeToDB=%t",
				tc.name, writeDB,
			)
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				dbChannel := make(chan dbtype.Message, 8)
				status := New(
					nil, dbChannel, nil,
					"",
					"0.0.0", "",
					tc.had.version, tc.had.timestamp,
					lastQueried,
					&dashboard.Options{
						OptionsBase: dashboard.OptionsBase{
							WebURL: "https://example.com",
						},
					},
				)
				status.Init(
					0, 0, 0,
					ServiceInfo{
						ID: name,
					},
					status.Dashboard,
				)
				if tc.want == nil {
					tc.want = &tc.args
				}

				// WHEN: SetLatestVersion is called on it.
				status.SetLatestVersion(tc.args.version, tc.args.timestamp, writeDB)

				prefix := fmt.Sprintf(
					"%s\nStatus(latest_version=%q).SetDeployedVersion(version=%q, timestamp=%q, writeDB=%t)",
					packageName, tc.had.version, tc.args.version, tc.args.timestamp, writeDB,
				)

				// THEN: LatestVersion is set to this version.
				fieldTests := []test.FieldAssertion{
					{Name: "LatestVersion", Got: status.LatestVersion(), Want: tc.want.version, Mode: test.CompareEqual},
					{Name: "LatestVersionTimestamp", Got: status.LatestVersionTimestamp(), Want: tc.want.timestamp, Mode: test.CompareEqual},
				}
				if err := test.AssertFields(t, fieldTests, prefix, "Status"); err != nil {
					t.Fatal(err)
				}

				// AND: the LatestVersionIsDeployedVersion metric is updated as expected.
				got := int(testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(status.ServiceInfo.ID)))
				var want int
				if writeDB && status.LatestVersion() == status.DeployedVersion() {
					want = int(metric.LatestVersionDeployed)
				}
				if got != want {
					t.Errorf(
						"%s LatestVersionIsDeployedVersion metric mismatch\ngot:  %d\nwant: %d",
						prefix, got, want,
					)
				}
			})
		}
	}
}

func TestStatus_RegexMissesContent(t *testing.T) {
	// GIVEN: a Status.
	status := Status{}

	var want uint = 0
	for i := 1; i <= 3; i++ {
		// WHEN: RegexMissContent is called on it.
		status.RegexMissContent()
		want++

		// THEN: RegexMissesVersion is incremented.
		if got := status.RegexMissesContent(); got != want {
			t.Errorf(
				"%s\nStatus.RegexMissesContent() mismatch (increment %d)\ngot:  %d\nwant: %d",
				packageName, i,
				got, want,
			)
		}
	}

	// WHEN: ResetRegexMisses is called on it.
	status.ResetRegexMisses()

	// THEN: RegexMisses is reset.
	want = 0
	if got := status.RegexMissesContent(); got != want {
		t.Errorf(
			"%s\nStatus.RegexMissesContent() mismatch (after ResetRegexMisses())\ngot:  %d\nwant: %d",
			packageName, got, want,
		)
	}
}

func TestStatus_RegexMissesVersion(t *testing.T) {
	// GIVEN: a Status.
	status := Status{}

	var want uint = 0
	for i := 1; i <= 3; i++ {
		// WHEN: RegexMissVersion is called on it.
		status.RegexMissVersion()
		want++

		// THEN: RegexMissesVersion is incremented.
		if got := status.RegexMissesVersion(); got != want {
			t.Errorf(
				"%s\nStatus.RegexMissesVersion() mismatch (increment %d)\ngot:  %d\nwant: %d",
				packageName, i,
				got, want,
			)
		}
	}

	// WHEN: ResetRegexMisses is called on it.
	status.ResetRegexMisses()

	// THEN: RegexMisses is reset.
	want = 0
	if got := status.RegexMissesVersion(); got != want {
		t.Errorf(
			"%s\nStatus.RegexMissesVersion() mismatch (after ResetRegexMisses())\ngot:  %d\nwant: %d",
			packageName, got, want,
		)
	}
}

func TestStatus_SetDeleting(t *testing.T) {
	// GIVEN: a Status.
	status := Status{}

	// WHEN: SetDeleting is called on it.
	status.SetDeleting()

	// THEN: the deleting flag should be set to true.
	if !status.Deleting() {
		t.Errorf(
			"%s\nStatus.SetDeleting() mismatch\ngot:  false\nwant: true",
			packageName,
		)
	}
	// WHEN: SetDeleting is called on it again.
	status.SetDeleting()

	// THEN: the deleting flag should still be true.
	if !status.Deleting() {
		t.Errorf(
			"%s\nStatus.SetDeleting() mismatch after second call\ngot:  false\nwant: true",
			packageName,
		)
	}
}

func TestStatus_SendAnnounce(t *testing.T) {
	// GIVEN: a Status with channels.
	tests := []struct {
		name       string
		deleting   bool
		nilChannel bool
	}{
		{name: "not deleting, or nil channel"},
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

			// WHEN: SendAnnounce is called on it.
			status.SendAnnounce(&[]byte{})

			// THEN: the AnnounceChannel is sent a message if not deleting or nil.
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
					"%s\nStatus.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					packageName, got, want,
				)
			}
		})
	}
}

func TestStatus_SendDatabase(t *testing.T) {
	// GIVEN: a Status with channels.
	tests := []struct {
		name       string
		deleting   bool
		nilChannel bool
	}{
		{name: "not deleting, or nil channel"},
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

			// AND: an empty message.
			msg := &dbtype.Message{}

			// WHEN: sendDatabase is called on it.
			status.sendDatabase(msg)

			prefix := fmt.Sprintf(
				"%s\nStatus.SendDatabase(msg=%+v)",
				packageName, msg,
			)

			// THEN: the DatabaseChannel is sent a message if not deleting or nil.
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
					"%s .DatabaseChannel message count mismatch\ngot:  %d\nwant: %d",
					prefix, got, want,
				)
			}
		})
	}
}

func TestStatus_SendSave(t *testing.T) {
	// GIVEN: a Status with channels.
	tests := []struct {
		name       string
		deleting   bool
		nilChannel bool
	}{
		{name: "not deleting, or nil channel"},
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

			// WHEN: SendSave is called on it.
			status.SendSave()

			// THEN: the SaveChannel is sent a message if not deleting or nil.
			want := 1
			if tc.deleting || tc.nilChannel {
				want = 0
			}
			got := 0
			if status.SaveChannel != nil {
				got = len(status.SaveChannel)
			}
			if got != want {
				t.Errorf(
					"%s\nStatus.SendSave() .SaveChannel message count mismatch\ngot:  %d\nwant: %d",
					packageName, got, want,
				)
			}
		})
	}
}

func TestSetLatestVersionIsDeployedMetric(t *testing.T) {
	// GIVEN: a ServiceInfo.
	tests := []struct {
		name      string
		serviceID string
		versions  serviceinfo.ServiceInfo
		isSkipped bool
		had, want float64
	}{
		{
			name: "latest version is deployed",
			versions: serviceinfo.ServiceInfo{
				LatestVersion:   "1.2.3",
				DeployedVersion: "1.2.3",
			},
			want: 1,
		},
		{
			name: "latest version is not deployed",
			versions: serviceinfo.ServiceInfo{
				LatestVersion:   "1.2.3",
				DeployedVersion: "1.2.4",
			},
			want: 0,
		},
		{
			name: "latest version is not deployed, but is approved",
			versions: serviceinfo.ServiceInfo{
				ApprovedVersion: "1.2.3",
				LatestVersion:   "1.2.3",
				DeployedVersion: "1.2.4",
			},
			want: 2,
		},
		{
			name: "latest version is not deployed, but is skipped",
			versions: serviceinfo.ServiceInfo{
				ApprovedVersion: serviceinfo.SkippedVersion("1.2.3"),
				LatestVersion:   "1.2.3",
				DeployedVersion: "1.2.4",
			},
			want: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.versions.ID = tc.name

			// WHEN: setLatestVersion is called on it.
			setLatestVersionIsDeployedMetric(tc.versions)
			got := testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(tc.versions.ID))

			// THEN: the metric is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nStatus.SetLatestVersionIsDeployedMetric() mismatch\ngot:  %f\nwant: %f",
					packageName, got, tc.want,
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

func FuzzSetVersions(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3}, []byte{0, 1, 2, 3})

	testStatus := testStatus()
	testStatus.ServiceInfo.LatestVersion = "0.0.0"
	testStatus.ServiceInfo.DeployedVersion = "0.0.0"
	testStatus.ServiceInfo.ApprovedVersion = ""

	announceChan := make(chan []byte, 4)
	testStatus.AnnounceChannel = announceChan
	databaseChan := make(chan dbtype.Message, 4)
	testStatus.DatabaseChannel = databaseChan
	saveChan := make(chan bool, 4)
	testStatus.SaveChannel = saveChan
	running := true
	f.Cleanup(func() {
		running = false
	})
	// Drain to prevent blocking.
	go func() {
		for running {
			select {
			case <-announceChan:
				continue
			case <-databaseChan:
				continue
			case <-saveChan:
				continue
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

			switch op {
			case opSetApprovedLatest:
				testStatus.SetApprovedVersion(version, true)

			case opSetApprovedSkip:
				testStatus.SetApprovedVersion(
					serviceinfo.SkippedVersion(testStatus.ServiceInfo.LatestVersion),
					true,
				)

			case opSetLatest:
				testStatus.SetLatestVersion(version, "", true)

			case opSetDeployed:
				testStatus.SetDeployedVersion(version, "", true)
			}

			assertSetVersionsResult(t, testStatus, hadUpdatesCurrentAvailable, hadUpdatesCurrentSkipped)
		}
	})
}

func assertSetVersionsResult(t *testing.T, s *Status, hadAvailable, hadSkipped float64) {
	info := s.GetServiceInfo()

	// Approved must not equal Latest if Latest == Deployed.
	if info.LatestVersion == info.DeployedVersion &&
		info.ApprovedVersion == info.LatestVersion {
		t.Fatalf("invalid state: approved == latest == deployed")
	}

	// SkippedVersion logic correctness.
	if serviceinfo.SkippedVersion(info.LatestVersion) == info.DeployedVersion &&
		info.ApprovedVersion != "" {
		t.Fatalf("approved should reset after skipped deploy")
	}

	// Timestamps must exist if version exists.
	if info.DeployedVersion != "" && s.DeployedVersionTimestamp() == "" {
		t.Fatalf("deployed version without timestamp")
	}

	// Updates Available.
	want := hadAvailable
	//   Increase when Latest is not Deployed.
	if info.LatestVersion != info.DeployedVersion {
		want += 1
	}
	got := testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("AVAILABLE"))
	if got != want {
		t.Fatalf(
			"%s\navailable updates metric count mismatch\nwant: %f\ngot:  %f\nserviceInfo: %+v",
			packageName, want, got, info,
		)
	}
	// Updates Skipped.
	want = hadSkipped
	//   Increase when Approved is Skip of Latest.
	if info.ApprovedVersion == serviceinfo.SkippedVersion(info.LatestVersion) {
		want += 1
	}
	got = testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("SKIPPED"))
	if got != want {
		t.Fatalf(
			"%s\nskipped updates metric count mismatch\nwant: %f\ngot:  %f\nserviceInfo: %+v",
			packageName, want, got, info,
		)
	}
}

func TestStatus_InitMetrics__DeleteMetrics(t *testing.T) {
	// GIVEN: a Status.
	status := testStatus()
	status.ServiceInfo.ID = "TestStatus_InitMetrics_DeleteMetrics"
	status.ServiceInfo.LatestVersion = "1.1.0"
	status.ServiceInfo.DeployedVersion = "1.0.0"
	status.ServiceInfo.ApprovedVersion = serviceinfo.SkippedVersion("1.0.0")

	// WHEN: InitMetrics is called on it.
	status.InitMetrics()

	// THEN: the metrics are created.
	got := testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(status.ServiceInfo.ID))
	if got != 0 {
		t.Fatalf("latest version is deployed metric not created")
	}

	// WHEN: DeleteMetrics is called on it.
	status.DeleteMetrics()

	// THEN: the metrics are deleted.
	got = testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(status.ServiceInfo.ID))
	if got != 0 {
		t.Fatalf("latest version is deployed metric not deleted")
	}
}
