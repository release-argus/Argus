// Copyright [2023] [Argus]
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

package svcstatus

import (
	"encoding/json"
	"testing"

	api_type "github.com/release-argus/Argus/web/api/types"
)

func TestStatus_AnnounceFirstVersion(t *testing.T) {
	// GIVEN a Status and an AnnounceChannel that may be nil
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
			wantID := *status.ServiceID
			wantLatestVersion := status.LatestVersion()
			wantLatestVersionTimestamp := status.LatestVersionTimestamp()

			// WHEN AnnounceFirstVersion is called on it
			status.AnnounceFirstVersion()

			// THEN the message is received
			if tc.nilChannel {
				return
			}
			gotData := <-*status.AnnounceChannel
			var got api_type.WebSocketMessage
			json.Unmarshal(gotData, &got)
			if got.ServiceData.ID != wantID {
				t.Fatalf("ID - got %q, want %q",
					got.ServiceData.ID, wantID)
			}
			if got.ServiceData.Status.LatestVersion != wantLatestVersion {
				t.Errorf("LatestVersion - got %q, want %q",
					got.ServiceData.Status.LatestVersion, wantLatestVersion)
			}
			if got.ServiceData.Status.LatestVersionTimestamp != wantLatestVersionTimestamp {
				t.Errorf("LatestVersionTimestamp - got %q, want %q",
					got.ServiceData.Status.LatestVersionTimestamp, wantLatestVersionTimestamp)
			}
		})
	}
}

func TestStatus_AnnounceQuery(t *testing.T) {
	// GIVEN an AnnounceChannel
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
			wantID := *status.ServiceID
			wantLastQueried := status.LastQueried()

			// WHEN AnnounceQuery is called on it
			status.AnnounceQuery()

			// THEN the message is received
			if tc.nilChannel {
				return
			}
			gotData := <-*status.AnnounceChannel
			var got api_type.WebSocketMessage
			json.Unmarshal(gotData, &got)
			if got.ServiceData.ID != wantID {
				t.Fatalf("ID - got %q, want %q\n%#v",
					got.ServiceData.ID, wantID, got.ServiceData.Status)
			}
			if got.ServiceData.Status.LastQueried != wantLastQueried {
				t.Errorf("LastQueried - got %q, want %q",
					got.ServiceData.Status.LatestVersion, wantLastQueried)
			}
		})
	}
}

func TestStatus_AnnounceQueryNewVersion(t *testing.T) {
	// GIVEN a Status and an AnnounceChannel that may be nil
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
			wantID := *status.ServiceID
			wantLatestVersion := status.LatestVersion()
			wantLatestVersionTimestamp := status.LatestVersionTimestamp()

			// WHEN AnnounceQueryNewVersion is called on it
			status.AnnounceQueryNewVersion()

			// THEN the message is received
			if tc.nilChannel {
				return
			}
			gotData := <-*status.AnnounceChannel
			var got api_type.WebSocketMessage
			json.Unmarshal(gotData, &got)
			if got.ServiceData.ID != wantID {
				t.Fatalf("ID - got %q, want %q\n%#v",
					got.ServiceData.ID, wantID, got.ServiceData.Status)
			}
			if got.ServiceData.Status.LatestVersion != wantLatestVersion {
				t.Errorf("LatestVersion - got %q, want %q",
					got.ServiceData.Status.LatestVersion, wantLatestVersion)
			}
			if got.ServiceData.Status.LatestVersionTimestamp != wantLatestVersionTimestamp {
				t.Errorf("LatestVersionTimestamp - got %q, want %q\n%#v",
					got.ServiceData.Status.LatestVersionTimestamp, wantLatestVersionTimestamp, got.ServiceData.Status)
			}
		})
	}
}

func TestStatus_AnnounceUpdate(t *testing.T) {
	// GIVEN a Status and an AnnounceChannel that may be nil
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
			wantID := *status.ServiceID
			wantDeployedVersion := status.DeployedVersion()
			wantDeployedVersionTimestamp := status.DeployedVersionTimestamp()

			// WHEN AnnounceUpdate is called on it
			status.AnnounceUpdate()

			// THEN the message is received
			if tc.nilChannel {
				return
			}
			gotData := <-*status.AnnounceChannel
			var got api_type.WebSocketMessage
			json.Unmarshal(gotData, &got)
			if got.ServiceData.ID != wantID {
				t.Fatalf("ID - got %q, want %q\n%#v",
					got.ServiceData.ID, wantID, got.ServiceData.Status)
			}
			if got.ServiceData.Status.DeployedVersion != wantDeployedVersion {
				t.Errorf("DeployedVersion - got %q, want %q",
					got.ServiceData.Status.DeployedVersion, wantDeployedVersion)
			}
			if got.ServiceData.Status.DeployedVersionTimestamp != wantDeployedVersionTimestamp {
				t.Errorf("DeployedVersionTimestamp - got %q, want %q\n%#v",
					got.ServiceData.Status.DeployedVersionTimestamp, wantDeployedVersionTimestamp, got.ServiceData.Status)
			}
		})
	}
}

func TestStatus_announceApproved(t *testing.T) {
	// GIVEN a Status and an AnnounceChannel
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
			wantID := *status.ServiceID
			wantApprovedVersion := status.ApprovedVersion()

			// WHEN announceApproved is called on it
			status.announceApproved()

			// THEN the message is received
			if tc.nilChannel {
				return
			}
			gotData := <-*status.AnnounceChannel
			var got api_type.WebSocketMessage
			json.Unmarshal(gotData, &got)
			if got.ServiceData.ID != wantID {
				t.Fatalf("ID - got %q, want %q\n%#v",
					got.ServiceData.ID, wantID, got.ServiceData.Status)
			}
			if got.ServiceData.Status.ApprovedVersion != wantApprovedVersion {
				t.Errorf("ApprovedVersion - got %q, want %q",
					got.ServiceData.Status.LatestVersion, wantApprovedVersion)
			}
		})
	}
}
