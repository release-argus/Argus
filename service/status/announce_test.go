// Copyright [2022] [Argus]
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

package service_status

import (
	"encoding/json"
	"testing"

	api_types "github.com/release-argus/Argus/web/api/types"
)

func TestAnnounceFirstVersion(t *testing.T) {
	// GIVEN a Status
	tests := map[string]struct {
		nilChannel bool
	}{
		"nil channel doesn't crash":  {nilChannel: true},
		"non-nil sends correct data": {nilChannel: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			status := testStatus()
			if tc.nilChannel {
				status.AnnounceChannel = nil
			}
			wantID := *status.ServiceID
			wantLatestVersion := status.LatestVersion
			wantLatestVersionTimestamp := status.LatestVersionTimestamp

			// WHEN AnnounceFirstVersion is called on it
			status.AnnounceFirstVersion()

			// THEN the message is received
			if tc.nilChannel {
				return
			}
			gotData := <-*status.AnnounceChannel
			var got api_types.WebSocketMessage
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

func TestAnnounceQuery(t *testing.T) {
	// GIVEN a Status
	tests := map[string]struct {
		nilChannel bool
	}{
		"nil channel doesn't crash":  {nilChannel: true},
		"non-nil sends correct data": {nilChannel: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			status := testStatus()
			if tc.nilChannel {
				status.AnnounceChannel = nil
			}
			wantID := *status.ServiceID
			wantLastQueried := status.LastQueried

			// WHEN AnnounceQuery is called on it
			status.AnnounceQuery()

			// THEN the message is received
			if tc.nilChannel {
				return
			}
			gotData := <-*status.AnnounceChannel
			var got api_types.WebSocketMessage
			json.Unmarshal(gotData, &got)
			if got.ServiceData.ID != wantID {
				t.Fatalf("ID - got %q, want %q\n%#v",
					got.ServiceData.ID, wantID, got.ServiceData.Status)
			}
			if got.ServiceData.Status.LastQueried != wantLastQueried {
				t.Errorf("%s:\nLastQueried - got %q, want %q",
					name, got.ServiceData.Status.LatestVersion, wantLastQueried)
			}
		})
	}
}

func TestAnnounceQueryNewVersion(t *testing.T) {
	// GIVEN a Status
	tests := map[string]struct {
		nilChannel bool
	}{
		"nil channel doesn't crash":  {nilChannel: true},
		"non-nil sends correct data": {nilChannel: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			status := testStatus()
			if tc.nilChannel {
				status.AnnounceChannel = nil
			}
			wantID := *status.ServiceID
			wantLatestVersion := status.LatestVersion
			wantLatestVersionTimestamp := status.LatestVersionTimestamp

			// WHEN AnnounceQueryNewVersion is called on it
			status.AnnounceQueryNewVersion()

			// THEN the message is received
			if tc.nilChannel {
				return
			}
			gotData := <-*status.AnnounceChannel
			var got api_types.WebSocketMessage
			json.Unmarshal(gotData, &got)
			if got.ServiceData.ID != wantID {
				t.Fatalf("ID - got %q, want %q\n%#v",
					got.ServiceData.ID, wantID, got.ServiceData.Status)
			}
			if got.ServiceData.Status.LatestVersion != wantLatestVersion {
				t.Errorf("%s:\nLatestVersion - got %q, want %q",
					name, got.ServiceData.Status.LatestVersion, wantLatestVersion)
			}
			if got.ServiceData.Status.LatestVersionTimestamp != wantLatestVersionTimestamp {
				t.Errorf("%s:\nLatestVersionTimestamp - got %q, want %q\n%#v",
					name, got.ServiceData.Status.LatestVersionTimestamp, wantLatestVersionTimestamp, got.ServiceData.Status)
			}
		})
	}
}

func TestAnnounceUpdate(t *testing.T) {
	// GIVEN a Status
	tests := map[string]struct {
		nilChannel bool
	}{
		"nil channel doesn't crash":  {nilChannel: true},
		"non-nil sends correct data": {nilChannel: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			status := testStatus()
			if tc.nilChannel {
				status.AnnounceChannel = nil
			}
			wantID := *status.ServiceID
			wantDeployedVersion := status.DeployedVersion
			wantDeployedVersionTimestamp := status.DeployedVersionTimestamp

			// WHEN AnnounceUpdate is called on it
			status.AnnounceUpdate()

			// THEN the message is received
			if tc.nilChannel {
				return
			}
			gotData := <-*status.AnnounceChannel
			var got api_types.WebSocketMessage
			json.Unmarshal(gotData, &got)
			if got.ServiceData.ID != wantID {
				t.Fatalf("ID - got %q, want %q\n%#v",
					got.ServiceData.ID, wantID, got.ServiceData.Status)
			}
			if got.ServiceData.Status.DeployedVersion != wantDeployedVersion {
				t.Errorf("%s:\nDeployedVersion - got %q, want %q",
					name, got.ServiceData.Status.DeployedVersion, wantDeployedVersion)
			}
			if got.ServiceData.Status.DeployedVersionTimestamp != wantDeployedVersionTimestamp {
				t.Errorf("%s:\nDeployedVersionTimestamp - got %q, want %q\n%#v",
					name, got.ServiceData.Status.DeployedVersionTimestamp, wantDeployedVersionTimestamp, got.ServiceData.Status)
			}
		})
	}
}

func TestAnnounceApprovedWithNilAnnounce(t *testing.T) {
	// GIVEN a Status with a nil Announce
	status := testStatus()
	status.AnnounceChannel = nil

	// WHEN AnnounceApproved is called on it
	status.AnnounceApproved()

	// THEN the function doesn't hang/err
}

func TestAnnounceApprovedWithAnnounce(t *testing.T) {
	// GIVEN a Status with an Announce channel
	status := testStatus()

	// WHEN AnnounceApproved is called on it
	status.AnnounceApproved()

	// THEN the function announces to the channel
	got := len(*status.AnnounceChannel)
	want := 1
	if got != want {
		t.Errorf("%d messages in the channel from the announce. Should be %d",
			got, want)
	}
}

func TestAnnounceApproved(t *testing.T) {
	// GIVEN a Status
	tests := map[string]struct {
		nilChannel bool
	}{
		"nil channel doesn't crash":  {nilChannel: true},
		"non-nil sends correct data": {nilChannel: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			status := testStatus()
			if tc.nilChannel {
				status.AnnounceChannel = nil
			}
			wantID := *status.ServiceID
			wantApprovedVersion := status.ApprovedVersion

			// WHEN AnnounceApproved is called on it
			status.AnnounceApproved()

			// THEN the message is received
			if tc.nilChannel {
				return
			}
			gotData := <-*status.AnnounceChannel
			var got api_types.WebSocketMessage
			json.Unmarshal(gotData, &got)
			if got.ServiceData.ID != wantID {
				t.Fatalf("ID - got %q, want %q\n%#v",
					got.ServiceData.ID, wantID, got.ServiceData.Status)
			}
			if got.ServiceData.Status.ApprovedVersion != wantApprovedVersion {
				t.Errorf("%s:\nApprovedVersion - got %q, want %q",
					name, got.ServiceData.Status.LatestVersion, wantApprovedVersion)
			}
		})
	}
}
