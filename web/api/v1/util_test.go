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

package v1

import (
	"crypto/sha256"
	"encoding/json"
	"net/url"
	"testing"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
)

func TestGetParam(t *testing.T) {
	// GIVEN a map of query parameters and a parameter to retrieve
	tests := map[string]struct {
		queryParams url.Values
		param       string
		want        *string
	}{
		"param exists": {
			queryParams: url.Values{"key": {"value"}},
			param:       "key",
			want:        strPtr("value"),
		},
		"param does not exist": {
			queryParams: url.Values{"key": {"value"}},
			param:       "nonexistent",
			want:        nil,
		},
		"empty param": {
			queryParams: url.Values{"key": {""}},
			param:       "key",
			want:        strPtr(""),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN getParam is called
			got := getParam(&tc.queryParams, tc.param)

			// THEN the result should be as expected
			if (got == nil && tc.want != nil) ||
				(got != nil && tc.want == nil) ||
				(got != nil && *got != *tc.want) {
				t.Errorf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestAnnounceDelete(t *testing.T) {
	// GIVEN an API instance and a serviceID
	serviceID := "test-service"
	announceChannel := make(chan []byte, 2)
	statusDefaults := status.NewDefaults(
		&announceChannel,
		nil,
		nil)
	api := &API{
		Config: &config.Config{
			Order: []string{"some-order"},
			HardDefaults: config.Defaults{
				Service: service.Defaults{
					Status: statusDefaults}}}}

	// WHEN announceDelete is called
	api.announceDelete(serviceID)

	// THEN the message should be sent to the announce channel
	select {
	case msg := <-announceChannel:
		var wsMessage apitype.WebSocketMessage
		err := json.Unmarshal(msg, &wsMessage)
		if err != nil {
			t.Fatalf("failed to unmarshal message: %v", err)
		}

		if wsMessage.Page != "APPROVALS" ||
			wsMessage.Type != "DELETE" ||
			wsMessage.SubType != serviceID {
			t.Errorf("unexpected WebSocketMessage: %+v",
				wsMessage)
		}
	default:
		t.Fatal("expected message on announce channel, but got none")
	}
}

func strPtr(s string) *string {
	return &s
}

func TestConstantTimeCompare(t *testing.T) {
	// GIVEN two hashes
	tests := map[string]struct {
		hash1, hash2 string
	}{
		"equal - 1": {
			hash1: "a",
			hash2: "a",
		},
		"equal - 2": {
			hash1: "abc",
			hash2: "abc",
		},
		"not equal - 1": {
			hash1: "a",
			hash2: "b",
		},
		"not equal - 2": {
			hash1: "abc",
			hash2: "abb",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			hash1 := sha256.Sum256([]byte(tc.hash1))
			hash2 := sha256.Sum256([]byte(tc.hash2))

			// WHEN ConstantTimeCompare is called
			got := ConstantTimeCompare(hash1, hash2)

			// THEN the result should be as expected
			want := tc.hash1 == tc.hash2
			if got != (want) {
				t.Errorf("want %v, got %v",
					want, got)
			}
		})
	}
}

func TestAnnounceEdit(t *testing.T) {
	// GIVEN an API instance and old/new service data
	announceChannel := make(chan []byte, 2)
	statusDefaults := status.NewDefaults(
		&announceChannel,
		nil,
		nil)
	api := &API{
		Config: &config.Config{
			HardDefaults: config.Defaults{
				Service: service.Defaults{
					Status: statusDefaults}}}}

	tests := map[string]struct {
		oldData           *apitype.ServiceSummary
		newData           apitype.ServiceSummary
		wantedServiceData *apitype.ServiceSummary
	}{
		"edit with old data": {
			oldData:           &apitype.ServiceSummary{ID: "service-1", Icon: test.StringPtr("Service 1")},
			newData:           apitype.ServiceSummary{ID: "service-2", Icon: test.StringPtr("Service 1 Updated")},
			wantedServiceData: &apitype.ServiceSummary{ID: "service-2", Icon: test.StringPtr("Service 1 Updated")},
		},
		"edit with old data, no change": {
			oldData:           &apitype.ServiceSummary{ID: "service-1", Icon: test.StringPtr("Service 1")},
			newData:           apitype.ServiceSummary{ID: "service-1", Icon: test.StringPtr("Service 1")},
			wantedServiceData: &apitype.ServiceSummary{},
		},
		"edit with old data, only changes sent": {
			oldData:           &apitype.ServiceSummary{ID: "service-1", Icon: test.StringPtr("Service 1"), Type: "github"},
			newData:           apitype.ServiceSummary{ID: "service-1", Icon: test.StringPtr("Service 1"), Type: "url"},
			wantedServiceData: &apitype.ServiceSummary{Type: "url"},
		},
		"edit without old data": {
			oldData:           nil,
			newData:           apitype.ServiceSummary{ID: "service-2", Icon: test.StringPtr("Service 2")},
			wantedServiceData: &apitype.ServiceSummary{ID: "service-2", Icon: test.StringPtr("Service 2"), Status: &apitype.Status{}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.newData.Status = &apitype.Status{}
			if tc.oldData != nil {
				tc.oldData.Status = &apitype.Status{}
			}

			// WHEN announceEdit is called
			api.announceEdit(tc.oldData, &tc.newData)

			// THEN the message should be sent to the announce channel
			select {
			case msg := <-announceChannel:
				var wsMessage apitype.WebSocketMessage
				err := json.Unmarshal(msg, &wsMessage)
				if err != nil {
					t.Fatalf("failed to unmarshal message: %v",
						err)
				}

				// AND the ServiceData should be as expected
				wantedStr := util.ToYAMLString(tc.wantedServiceData, "")
				gotStr := util.ToYAMLString(wsMessage.ServiceData, "")
				if wsMessage.Page != "APPROVALS" ||
					wsMessage.Type != "EDIT" ||
					(tc.oldData != nil && wsMessage.SubType != tc.oldData.ID) ||
					(tc.oldData == nil && wsMessage.SubType != "") ||
					gotStr != wantedStr {
					t.Errorf("unexpected WebSocketMessage:\nwant:%q\ngot:  %q",
						wantedStr, gotStr)
				}
			default:
				t.Fatal("expected message on announce channel, but got none")
			}
		})
	}
}
