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
	"errors"
	"testing"
	"time"

	"github.com/release-argus/Argus/config/decode"
	apitype "github.com/release-argus/Argus/web/api/types"
)

func TestStatus_SendAnnouncePayload__marshalError(t *testing.T) {
	original := marshalAnnouncePayload
	customErr := errors.New("marshal failed")
	marshalAnnouncePayload = func(v any) ([]byte, error) {
		return nil, customErr
	}
	t.Cleanup(func() { marshalAnnouncePayload = original })

	status := testStatus()
	announceChannel := make(chan []byte, 1)
	status.AnnounceChannel = announceChannel

	status.sendAnnouncePayload(apitype.WebSocketMessage{})

	select {
	case msg := <-announceChannel:
		t.Fatalf("Status.AnnounceChannel message mismatch\ngot:  %q\nwant: absent", msg)
	default:
	}
}

func TestStatus_AnnounceFirstVersion(t *testing.T) {
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

			status.AnnounceFirstVersion()

			if tc.nilChannel {
				return
			}
			got := announcedStatus(t, status.AnnounceChannel, wantID)
			if got.LatestVersion != wantLatestVersion {
				t.Errorf(
					"WebSocketMessage.ServiceData.Status.LatestVersion mismatch\n"+
						"got:  %q\nwant: %q",
					got.LatestVersion, wantLatestVersion,
				)
			}
			if got.LatestVersionTimestamp != wantLatestVersionTimestamp {
				t.Errorf(
					"WebSocketMessage.ServiceData.Status.LatestVersionTimestamp mismatch\n"+
						"got:  %q\nwant: %q",
					got.LatestVersionTimestamp, wantLatestVersionTimestamp,
				)
			}
		})
	}
}

func TestStatus_AnnounceQuery(t *testing.T) {
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

			status.AnnounceQuery()

			if tc.nilChannel {
				return
			}
			got := announcedStatus(t, status.AnnounceChannel, wantID)
			if got.LastQueried != wantLastQueried {
				t.Fatalf(
					"WebSocketMessage.ServiceData.Status.LastQueried mismatch\ngot:  %q\nwant: %q",
					got.LastQueried, wantLastQueried,
				)
			}
		})
	}
}

func TestStatus_AnnounceQueryNewVersion(t *testing.T) {
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

			status.AnnounceQueryNewVersion()

			if tc.nilChannel {
				return
			}
			got := announcedStatus(t, status.AnnounceChannel, wantID)
			if got.LatestVersion != wantLatestVersion {
				t.Fatalf(
					"WebSocketMessage.ServiceData.Status.LatestVersion mismatch\n"+
						"got:  %q\nwant: %q",
					got.LatestVersion, wantLatestVersion,
				)
			}
			if got.LatestVersionTimestamp != wantLatestVersionTimestamp {
				t.Fatalf(
					"WebSocketMessage.ServiceData.Status.LatestVersionTimestamp mismatch\n"+
						"got:  %q\nwant: %q",
					got.LatestVersionTimestamp, wantLatestVersionTimestamp,
				)
			}
		})
	}
}

func TestStatus_AnnounceUpdate(t *testing.T) {
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

			status.AnnounceUpdate()

			if tc.nilChannel {
				return
			}
			got := announcedStatus(t, status.AnnounceChannel, wantID)
			if got.DeployedVersion != wantDeployedVersion {
				t.Fatalf(
					"WebSocketMessage.ServiceData.Status.DeployedVersion mismatch\n"+
						"got:  %q\nwant: %q",
					got.DeployedVersion, wantDeployedVersion,
				)
			}
			if got.DeployedVersionTimestamp != wantDeployedVersionTimestamp {
				t.Fatalf(
					"WebSocketMessage.ServiceData.Status.DeployedVersionTimestamp mismatch\n"+
						"got:  %q\nwant: %q",
					got.DeployedVersionTimestamp, wantDeployedVersionTimestamp,
				)
			}
		})
	}
}

func TestStatus_AnnounceApproved(t *testing.T) {
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
			wantApprovedVersion := "9.9.9"

			status.announceApproved(wantApprovedVersion)

			if tc.nilChannel {
				return
			}
			got := announcedStatus(t, status.AnnounceChannel, wantID)
			if got.ApprovedVersion != wantApprovedVersion {
				t.Fatalf(
					"WebSocketMessage.ServiceData.Status.ApprovedVersion mismatch\n"+
						"got:  %q\nwant: %q",
					got.ApprovedVersion, wantApprovedVersion,
				)
			}
		})
	}
}

// announcedStatus receives the next message on ch, for wantID, and returns its Status.
func announcedStatus(t *testing.T, ch chan []byte, wantID string) *apitype.Status {
	t.Helper()

	var data []byte
	select {
	case data = <-ch:
	case <-time.After(time.Second):
		t.Fatal("Status.AnnounceChannel message count mismatch\ngot:  0\nwant: 1")
	}

	var msg apitype.WebSocketMessage
	if err := decode.Unmarshal("json", data, &msg); err != nil {
		t.Fatalf("decode.Unmarshal() failed: %v", err)
	}

	if msg.ServiceData == nil {
		t.Fatalf(
			"WebSocketMessage.ServiceData mismatch\ngot:  nil\nwant: non-nil\nmessage: %s",
			data,
		)
	}
	if msg.ServiceData.ID != wantID {
		t.Fatalf(
			"WebSocketMessage.ServiceData.ID mismatch\ngot:  %q\nwant: %q",
			msg.ServiceData.ID, wantID,
		)
	}
	if msg.ServiceData.Status == nil {
		t.Fatalf(
			"WebSocketMessage.ServiceData.Status mismatch\ngot:  nil\nwant: non-nil\nmessage: %s",
			data,
		)
	}

	return msg.ServiceData.Status
}
