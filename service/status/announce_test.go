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

func TestAnnounceFirstVersionWithNilAnnounce(t *testing.T) {
	// GIVEN a Service with a nil Announce
	service := testServiceGitHub()
	service.Announce = nil

	// WHEN AnnounceFirstVersion is called on it
	service.AnnounceFirstVersion()

	// THEN the function doesn't hang/err
}

func TestAnnounceFirstVersionWithAnnounce(t *testing.T) {
	// GIVEN a Service with an Announce channel
	service := testServiceGitHub()

	// WHEN AnnounceFirstVersion is called on it
	service.AnnounceFirstVersion()

	// THEN the function announces to the channel
	got := len(*service.Announce)
	want := 1
	if got != want {
		t.Errorf("%d messages in the channel from the announce. Should be %d",
			got, want)
	}
}

func TestAnnounceFirstVersion(t *testing.T) {
	// GIVEN a Service with an Announce channel
	service := testServiceGitHub()

	// WHEN AnnounceFirstVersion is called on it
	service.AnnounceFirstVersion()

	// THEN the message contains the correct data
	gotData := <-*service.Announce
	var got api_types.WebSocketMessage
	json.Unmarshal(gotData, &got)
	wantID := *service.ID
	wantLatestVersion := service.Status.LatestVersion
	wantLatestVersionTimestamp := service.Status.LatestVersionTimestamp
	if *got.ServiceData.ID != wantID {
		t.Errorf("ID - got %q, want %q",
			*got.ServiceData.ID, wantID)
	}
	if got.ServiceData.Status.LatestVersion != wantLatestVersion {
		t.Errorf("LatestVersion - got %q, want %q",
			got.ServiceData.Status.LatestVersion, wantLatestVersion)
	}
	if got.ServiceData.Status.LatestVersionTimestamp != wantLatestVersionTimestamp {
		t.Errorf("LatestVersionTimestamp - got %q, want %q",
			got.ServiceData.Status.LatestVersionTimestamp, wantLatestVersionTimestamp)
	}
}

func TestAnnounceQueryWithNilAnnounce(t *testing.T) {
	// GIVEN a Service with a nil Announce
	service := testServiceGitHub()
	service.Announce = nil

	// WHEN AnnounceQuery is called on it
	service.AnnounceQuery()

	// THEN the function doesn't hang/err
}

func TestAnnounceQueryWithAnnounce(t *testing.T) {
	// GIVEN a Service with an Announce channel
	service := testServiceGitHub()

	// WHEN AnnounceQuery is called on it
	service.AnnounceQuery()

	// THEN the function announces to the channel
	got := len(*service.Announce)
	want := 1
	if got != want {
		t.Errorf("%d messages in the channel from the announce. Should be %d",
			got, want)
	}
}

func TestAnnounceQuery(t *testing.T) {
	// GIVEN a Service with an Announce channel
	service := testServiceGitHub()

	// WHEN AnnounceQuery is called on it
	service.AnnounceQuery()

	// THEN the message contains the correct data
	gotData := <-*service.Announce
	var got api_types.WebSocketMessage
	json.Unmarshal(gotData, &got)
	wantID := *service.ID
	wantLastQueried := service.Status.LastQueried
	if *got.ServiceData.ID != wantID {
		t.Errorf("ID - got %q, want %q",
			*got.ServiceData.ID, wantID)
	}
	if got.ServiceData.Status.LastQueried != wantLastQueried {
		t.Errorf("LastQueried - got %q, want %q",
			got.ServiceData.Status.LastQueried, wantLastQueried)
	}
}

func TestAnnounceQueryNewVersionWithNilAnnounce(t *testing.T) {
	// GIVEN a Service with a nil Announce
	service := testServiceGitHub()
	service.Announce = nil

	// WHEN AnnounceQueryNewVersion is called on it
	service.AnnounceQueryNewVersion()

	// THEN the function doesn't hang/err
}

func TestAnnounceQueryNewVersionWithAnnounce(t *testing.T) {
	// GIVEN a Service with an Announce channel
	service := testServiceGitHub()

	// WHEN AnnounceQueryNewVersion is called on it
	service.AnnounceQueryNewVersion()

	// THEN the function announces to the channel
	got := len(*service.Announce)
	want := 1
	if got != want {
		t.Errorf("%d messages in the channel from the announce. Should be %d",
			got, want)
	}
}

func TestAnnounceQueryNewVersion(t *testing.T) {
	// GIVEN a Service with an Announce channel
	service := testServiceGitHub()

	// WHEN AnnounceQueryNewVersion is called on it
	service.AnnounceQueryNewVersion()

	// THEN the message contains the correct data
	gotData := <-*service.Announce
	var got api_types.WebSocketMessage
	json.Unmarshal(gotData, &got)
	wantID := *service.ID
	wantLatestVersion := service.Status.LatestVersion
	wantLatestVersionTimestamp := service.Status.LatestVersionTimestamp
	if *got.ServiceData.ID != wantID {
		t.Errorf("ID - got %q, want %q",
			*got.ServiceData.ID, wantID)
	}
	if got.ServiceData.Status.LatestVersion != wantLatestVersion {
		t.Errorf("LatestVersion - got %q, want %q",
			got.ServiceData.Status.LatestVersion, wantLatestVersion)
	}
	if got.ServiceData.Status.LatestVersionTimestamp != wantLatestVersionTimestamp {
		t.Errorf("LatestVersionTimestamp - got %q, want %q",
			got.ServiceData.Status.LatestVersionTimestamp, wantLatestVersionTimestamp)
	}
}

func TestAnnounceUpdateWithNilAnnounce(t *testing.T) {
	// GIVEN a Service with a nil Announce
	service := testServiceGitHub()
	service.Announce = nil

	// WHEN AnnounceUpdate is called on it
	service.AnnounceUpdate()

	// THEN the function doesn't hang/err
}

func TestAnnounceUpdateWithAnnounce(t *testing.T) {
	// GIVEN a Service with an Announce channel
	service := testServiceGitHub()

	// WHEN AnnounceUpdate is called on it
	service.AnnounceUpdate()

	// THEN the function announces to the channel
	got := len(*service.Announce)
	want := 1
	if got != want {
		t.Errorf("%d messages in the channel from the announce. Should be %d",
			got, want)
	}
}

func TestAnnounceUpdate(t *testing.T) {
	// GIVEN a Service with an Announce channel
	service := testServiceGitHub()

	// WHEN AnnounceUpdate is called on it
	service.AnnounceUpdate()

	// THEN the message contains the correct data
	gotData := <-*service.Announce
	var got api_types.WebSocketMessage
	json.Unmarshal(gotData, &got)
	wantID := *service.ID
	wantDeployedVersion := service.Status.DeployedVersion
	wantDeployedVersionTimestamp := service.Status.DeployedVersionTimestamp
	if *got.ServiceData.ID != wantID {
		t.Errorf("ID - got %q, want %q",
			*got.ServiceData.ID, wantID)
	}
	if got.ServiceData.Status.DeployedVersion != wantDeployedVersion {
		t.Errorf("DeployedVersion - got %q, want %q",
			got.ServiceData.Status.DeployedVersion, wantDeployedVersion)
	}
	if got.ServiceData.Status.DeployedVersionTimestamp != wantDeployedVersionTimestamp {
		t.Errorf("DeployedVersionTimestamp - got %q, want %q",
			got.ServiceData.Status.DeployedVersionTimestamp, wantDeployedVersionTimestamp)
	}
}

func TestAnnounceApprovedWithNilAnnounce(t *testing.T) {
	// GIVEN a Service with a nil Announce
	service := testServiceGitHub()
	service.Announce = nil

	// WHEN AnnounceApproved is called on it
	service.AnnounceApproved()

	// THEN the function doesn't hang/err
}

func TestAnnounceApprovedWithAnnounce(t *testing.T) {
	// GIVEN a Service with an Announce channel
	service := testServiceGitHub()

	// WHEN AnnounceApproved is called on it
	service.AnnounceApproved()

	// THEN the function announces to the channel
	got := len(*service.Announce)
	want := 1
	if got != want {
		t.Errorf("%d messages in the channel from the announce. Should be %d",
			got, want)
	}
}

func TestAnnounceApproved(t *testing.T) {
	// GIVEN a Service with an Announce channel
	service := testServiceGitHub()

	// WHEN AnnounceApproved is called on it
	service.AnnounceApproved()

	// THEN the message contains the correct data
	gotData := <-*service.Announce
	var got api_types.WebSocketMessage
	json.Unmarshal(gotData, &got)
	wantID := *service.ID
	wantApprovedVersion := service.Status.ApprovedVersion
	if *got.ServiceData.ID != wantID {
		t.Errorf("ID - got %q, want %q",
			*got.ServiceData.ID, wantID)
	}
	if got.ServiceData.Status.ApprovedVersion != wantApprovedVersion {
		t.Errorf("ApprovedVersion - got %q, want %q",
			got.ServiceData.Status.ApprovedVersion, wantApprovedVersion)
	}
}
