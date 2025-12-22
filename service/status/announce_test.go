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

package status

import (
	"encoding/json"
	"testing"

	apitype "github.com/release-argus/Argus/web/api/types"
)

func TestStatus_AnnounceFirstVersion(t *testing.T) {
	// GIVEN a Status and an AnnounceChannel that may be nil.
	tests := map[string]struct {
		nilChannel bool
	}{
		"nil channel doesn't crash": {
			nilChannel: true},
		"non-nil sends correct data": {
			nilChannel: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			status := testStatus()
			if tc.nilChannel {
				status.AnnounceChannel = nil
			}
			wantID := status.ServiceInfo.ID
			wantLatestVersion := status.LatestVersion()
			wantLatestVersionTimestamp := status.LatestVersionTimestamp()

			// WHEN AnnounceFirstVersion is called on it.
			status.AnnounceFirstVersion()

			// THEN the message is received.
			if tc.nilChannel {
				return
			}
			gotData := <-status.AnnounceChannel
			var got apitype.WebSocketMessage
			_ = json.Unmarshal(gotData, &got)
			if got.ServiceData.ID != wantID {
				t.Fatalf("%s\nIDmismatch\nwant: %q\ngot:  %q",
					packageName, wantID, got.ServiceData.ID)
			}
			if got.ServiceData.Status.LatestVersion != wantLatestVersion {
				t.Errorf("%s\nLatestVersion mismatch\nwant: %q\ngot:  %q",
					packageName, wantLatestVersion, got.ServiceData.Status.LatestVersion)
			}
			if got.ServiceData.Status.LatestVersionTimestamp != wantLatestVersionTimestamp {
				t.Errorf("%s\nLatestVersionTimestamp mismatch\nwant: %q\ngot:  %q",
					packageName, wantLatestVersionTimestamp, got.ServiceData.Status.LatestVersionTimestamp)
			}
		})
	}
}

func TestStatus_AnnounceQuery(t *testing.T) {
	// GIVEN an AnnounceChannel.
	tests := map[string]struct {
		nilChannel bool
	}{
		"nil channel doesn't crash": {
			nilChannel: true},
		"non-nil sends correct data": {
			nilChannel: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			status := testStatus()
			if tc.nilChannel {
				status.AnnounceChannel = nil
			}
			wantID := status.ServiceInfo.ID
			wantLastQueried := status.LastQueried()

			// WHEN AnnounceQuery is called on it.
			status.AnnounceQuery()

			// THEN the message is received.
			if tc.nilChannel {
				return
			}
			gotData := <-status.AnnounceChannel
			var got apitype.WebSocketMessage
			_ = json.Unmarshal(gotData, &got)
			if got.ServiceData.ID != wantID {
				t.Fatalf("%s\nID mismatch\nwant: %q\ngot:  %q",
					packageName, wantID, got.ServiceData.ID)
			}
			if got.ServiceData.Status.LastQueried != wantLastQueried {
				t.Fatalf("%s\nLastQueried mismatch\nwant: %q\ngot:  %q",
					packageName, wantLastQueried, got.ServiceData.Status.LastQueried)
			}
		})
	}
}

func TestStatus_AnnounceQueryNewVersion(t *testing.T) {
	// GIVEN a Status and an AnnounceChannel that may be nil.
	tests := map[string]struct {
		nilChannel bool
	}{
		"nil channel doesn't crash": {
			nilChannel: true},
		"non-nil sends correct data": {
			nilChannel: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			status := testStatus()
			if tc.nilChannel {
				status.AnnounceChannel = nil
			}
			wantID := status.ServiceInfo.ID
			wantLatestVersion := status.LatestVersion()
			wantLatestVersionTimestamp := status.LatestVersionTimestamp()

			// WHEN AnnounceQueryNewVersion is called on it.
			status.AnnounceQueryNewVersion()

			// THEN the message is received.
			if tc.nilChannel {
				return
			}
			gotData := <-status.AnnounceChannel
			var got apitype.WebSocketMessage
			_ = json.Unmarshal(gotData, &got)
			if got.ServiceData.ID != wantID {
				t.Fatalf("%s\nID mismatch\nwant: %q\ngot:  %q",
					packageName, wantID, got.ServiceData.ID)
			}
			if got.ServiceData.Status.LatestVersion != wantLatestVersion {
				t.Fatalf("%s\nLatestVersion mismatch\nwant: %q\ngot:  %q",
					packageName, wantLatestVersion, got.ServiceData.Status.LatestVersion)
			}
			if got.ServiceData.Status.LatestVersionTimestamp != wantLatestVersionTimestamp {
				t.Fatalf("%s\nLatestVersionTimestamp mismatch\nwant: %q\ngot:  %q",
					packageName, wantLatestVersionTimestamp, got.ServiceData.Status.LatestVersionTimestamp)
			}
		})
	}
}

func TestStatus_AnnounceUpdate(t *testing.T) {
	// GIVEN a Status and an AnnounceChannel that may be nil.
	tests := map[string]struct {
		nilChannel bool
	}{
		"nil channel doesn't crash": {
			nilChannel: true},
		"non-nil sends correct data": {
			nilChannel: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			status := testStatus()
			if tc.nilChannel {
				status.AnnounceChannel = nil
			}
			wantID := status.ServiceInfo.ID
			wantDeployedVersion := status.DeployedVersion()
			wantDeployedVersionTimestamp := status.DeployedVersionTimestamp()

			// WHEN AnnounceUpdate is called on it.
			status.AnnounceUpdate()

			// THEN the message is received.
			if tc.nilChannel {
				return
			}
			gotData := <-status.AnnounceChannel
			var got apitype.WebSocketMessage
			_ = json.Unmarshal(gotData, &got)
			if got.ServiceData.ID != wantID {
				t.Fatalf("%s\nID mismatch\nwant: %q\ngot:  %q",
					packageName, wantID, got.ServiceData.ID)
			}
			if got.ServiceData.Status.DeployedVersion != wantDeployedVersion {
				t.Fatalf("%s\nDeployedVersion mismatch\nwant: %q\ngot:  %q",
					packageName, wantDeployedVersion, got.ServiceData.Status.DeployedVersion)
			}
			if got.ServiceData.Status.DeployedVersionTimestamp != wantDeployedVersionTimestamp {
				t.Fatalf("%s\nDeployedVersionTimestamp mismatch\nwant: %q\ngot:  %q",
					packageName, wantDeployedVersionTimestamp, got.ServiceData.Status.DeployedVersionTimestamp)
			}
		})
	}
}

func TestStatus_announceApproved(t *testing.T) {
	// GIVEN a Status and an AnnounceChannel.
	tests := map[string]struct {
		nilChannel bool
	}{
		"nil channel": {
			nilChannel: true},
		"non-nil": {
			nilChannel: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			status := testStatus()
			if tc.nilChannel {
				status.AnnounceChannel = nil
			}
			wantID := status.ServiceInfo.ID
			wantApprovedVersion := status.ApprovedVersion()

			// WHEN announceApproved is called on it.
			status.announceApproved()

			// THEN the message is received.
			if tc.nilChannel {
				return
			}
			gotData := <-status.AnnounceChannel
			var got apitype.WebSocketMessage
			_ =json.Unmarshal(gotData, &got)
			if got.ServiceData.ID != wantID {
				t.Fatalf("%s\nID mismatch\nwant: %q\ngot:  %q",
					packageName, wantID, got.ServiceData.ID)
			}
			if got.ServiceData.Status.ApprovedVersion != wantApprovedVersion {
				t.Fatalf("%s\nApprovedVersion mismatch\nwant: %q\ngot:  %q",
					packageName, wantApprovedVersion, got.ServiceData.Status.ApprovedVersion)
			}
		})
	}
}
