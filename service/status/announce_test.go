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

	"github.com/release-argus/Argus/config/decode"
	apitype "github.com/release-argus/Argus/web/api/types"
)

func TestStatus_AnnounceFirstVersion(t *testing.T) {
	// GIVEN: a Status and an AnnounceChannel that may be nil.
	tests := []struct {
		name       string
		nilChannel bool
	}{
		{name: "nil channel doesn't crash", nilChannel: true},
		{name: "non-nil sends correct data", nilChannel: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			status := testStatus()
			if tc.nilChannel {
				status.AnnounceChannel = nil
			}
			wantID := status.ServiceInfo.ID
			wantLatestVersion := status.LatestVersion()
			wantLatestVersionTimestamp := status.LatestVersionTimestamp()

			// WHEN: AnnounceFirstVersion is called on it.
			status.AnnounceFirstVersion()

			prefix := fmt.Sprintf("%s\nStatus.AnnounceFirstVersion()", packageName)

			// THEN: the message is received.
			if tc.nilChannel {
				return
			}
			gotData := <-status.AnnounceChannel
			var got apitype.WebSocketMessage
			_ = decode.Unmarshal("json", gotData, &got)
			if got.ServiceData.ID != wantID {
				t.Fatalf(
					"%s .ID mismatch\ngot:  %q\nwant: %q",
					prefix, got.ServiceData.ID, wantID,
				)
			}
			if gotLatestVersion := got.ServiceData.Status.LatestVersion; gotLatestVersion != wantLatestVersion {
				t.Errorf(
					"%s .LatestVersion mismatch\ngot:  %q\nwant: %q",
					prefix, gotLatestVersion, wantLatestVersion,
				)
			}
			if gotLatestVersionTimestamp := got.ServiceData.Status.LatestVersionTimestamp; gotLatestVersionTimestamp != wantLatestVersionTimestamp {
				t.Errorf(
					"%s .LatestVersionTimestamp mismatch\ngot:  %q\nwant: %q",
					prefix, gotLatestVersionTimestamp, wantLatestVersionTimestamp,
				)
			}
		})
	}
}

func TestStatus_AnnounceQuery(t *testing.T) {
	// GIVEN: an AnnounceChannel.
	tests := []struct {
		name       string
		nilChannel bool
	}{
		{name: "nil channel doesn't crash", nilChannel: true},
		{name: "non-nil sends correct data", nilChannel: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			status := testStatus()
			if tc.nilChannel {
				status.AnnounceChannel = nil
			}
			wantID := status.ServiceInfo.ID
			wantLastQueried := status.LastQueried()

			// WHEN: AnnounceQuery is called on it.
			status.AnnounceQuery()

			prefix := fmt.Sprintf("%s\nStatus.AnnounceQuery()", packageName)

			// THEN: the message is received.
			if tc.nilChannel {
				return
			}
			gotData := <-status.AnnounceChannel
			var got apitype.WebSocketMessage
			_ = decode.Unmarshal("json", gotData, &got)
			if got.ServiceData.ID != wantID {
				t.Fatalf(
					"%s .ID mismatch\ngot:  %q\nwant: %q",
					prefix, got.ServiceData.ID, wantID,
				)
			}
			if gotLastQueried := got.ServiceData.Status.LastQueried; gotLastQueried != wantLastQueried {
				t.Fatalf(
					"%s .LastQueried mismatch\ngot:  %q\nwant: %q",
					prefix, gotLastQueried, wantLastQueried,
				)
			}
		})
	}
}

func TestStatus_AnnounceQueryNewVersion(t *testing.T) {
	// GIVEN: a Status and an AnnounceChannel that may be nil.
	tests := []struct {
		name       string
		nilChannel bool
	}{
		{name: "nil channel doesn't crash", nilChannel: true},
		{name: "non-nil sends correct data", nilChannel: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			status := testStatus()
			if tc.nilChannel {
				status.AnnounceChannel = nil
			}
			wantID := status.ServiceInfo.ID
			wantLatestVersion := status.LatestVersion()
			wantLatestVersionTimestamp := status.LatestVersionTimestamp()

			// WHEN: AnnounceQueryNewVersion is called on it.
			status.AnnounceQueryNewVersion()

			prefix := fmt.Sprintf("%s\nStatus.AnnounceQueryNewVersion()", packageName)

			// THEN: the message is received.
			if tc.nilChannel {
				return
			}
			gotData := <-status.AnnounceChannel
			var got apitype.WebSocketMessage
			_ = decode.Unmarshal("json", gotData, &got)
			if got.ServiceData.ID != wantID {
				t.Fatalf(
					"%s .ID mismatch\ngot:  %q\nwant: %q",
					prefix, got.ServiceData.ID, wantID,
				)
			}
			if gotLatestVersion := got.ServiceData.Status.LatestVersion; gotLatestVersion != wantLatestVersion {
				t.Fatalf(
					"%s .LatestVersion mismatch\ngot:  %q\nwant: %q",
					prefix, gotLatestVersion, wantLatestVersion,
				)
			}
			if gotLatestVersionTimestamp := got.ServiceData.Status.LatestVersionTimestamp; gotLatestVersionTimestamp != wantLatestVersionTimestamp {
				t.Fatalf(
					"%s .LatestVersionTimestamp mismatch\ngot:  %q\nwant: %q",
					prefix, gotLatestVersionTimestamp, wantLatestVersionTimestamp,
				)
			}
		})
	}
}

func TestStatus_AnnounceUpdate(t *testing.T) {
	// GIVEN: a Status and an AnnounceChannel that may be nil.
	tests := []struct {
		name       string
		nilChannel bool
	}{
		{name: "nil channel doesn't crash", nilChannel: true},
		{name: "non-nil sends correct data", nilChannel: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			status := testStatus()
			if tc.nilChannel {
				status.AnnounceChannel = nil
			}
			wantID := status.ServiceInfo.ID
			wantDeployedVersion := status.DeployedVersion()
			wantDeployedVersionTimestamp := status.DeployedVersionTimestamp()

			// WHEN: AnnounceUpdate is called on it.
			status.AnnounceUpdate()

			prefix := fmt.Sprintf("%s\nStatus.AnnounceUpdate()", packageName)

			// THEN: the message is received.
			if tc.nilChannel {
				return
			}
			gotData := <-status.AnnounceChannel
			var got apitype.WebSocketMessage
			_ = decode.Unmarshal("json", gotData, &got)
			if got.ServiceData.ID != wantID {
				t.Fatalf(
					"%s .ID mismatch\ngot:  %q\nwant: %q",
					prefix, wantID, got.ServiceData.ID,
				)
			}
			if gotDeployedVersion := got.ServiceData.Status.DeployedVersion; gotDeployedVersion != wantDeployedVersion {
				t.Fatalf(
					"%s .DeployedVersion mismatch\ngot:  %q\nwant: %q",
					prefix, gotDeployedVersion, wantDeployedVersion,
				)
			}
			if gotDeployedVersionTimestamp := got.ServiceData.Status.DeployedVersionTimestamp; gotDeployedVersionTimestamp != wantDeployedVersionTimestamp {
				t.Fatalf(
					"%s .DeployedVersionTimestamp mismatch\ngot:  %q\nwant: %q",
					prefix, gotDeployedVersionTimestamp, wantDeployedVersionTimestamp,
				)
			}
		})
	}
}

func TestStatus_AnnounceApproved(t *testing.T) {
	// GIVEN: a Status and an AnnounceChannel.
	tests := []struct {
		name       string
		nilChannel bool
	}{
		{name: "nil channel", nilChannel: true},
		{name: "non-nil", nilChannel: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			status := testStatus()
			if tc.nilChannel {
				status.AnnounceChannel = nil
			}
			wantID := status.ServiceInfo.ID
			wantApprovedVersion := status.ApprovedVersion()

			// WHEN: announceApproved is called on it.
			status.announceApproved()

			prefix := fmt.Sprintf("%s\nStatus.announceApproved()", packageName)

			// THEN: the message is received.
			if tc.nilChannel {
				return
			}
			gotData := <-status.AnnounceChannel
			var got apitype.WebSocketMessage
			_ = decode.Unmarshal("json", gotData, &got)
			if got.ServiceData.ID != wantID {
				t.Fatalf(
					"%s .ID mismatch\ngot:  %q\nwant: %q",
					prefix, got.ServiceData.ID, wantID,
				)
			}
			if got.ServiceData.Status == nil {
				t.Fatalf(
					"%s ServiceData.Status is nil in message (%+v)",
					prefix, got,
				)
			}
			gotApprovedVersion := got.ServiceData.Status.ApprovedVersion
			if gotApprovedVersion != wantApprovedVersion {
				t.Fatalf(
					"%s .ApprovedVersion mismatch\ngot:  %q\nwant: %q",
					prefix, gotApprovedVersion, wantApprovedVersion,
				)
			}
		})
	}
}
