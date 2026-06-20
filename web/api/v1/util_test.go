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

package v1

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"testing"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
)

func TestAPI_SendAnnouncePayload__marshalError(t *testing.T) {
	// GIVEN: a failing marshal function.
	original := marshalAnnouncePayload
	customErr := fmt.Errorf("marshal failed")
	marshalAnnouncePayload = func(v any) ([]byte, error) {
		return nil, customErr
	}
	t.Cleanup(func() { marshalAnnouncePayload = original })

	// AND: an API with an Announce Channel.
	announceChannel := make(chan []byte, 1)
	statusDefaults := status.NewDefaults(
		announceChannel,
		nil,
		nil,
	)
	api := &API{
		Config: &config.Config{
			HardDefaults: config.Defaults{
				Service: service.Defaults{
					Status: statusDefaults,
				},
			},
		},
	}

	// WHEN: sendAnnouncePayload is called.
	api.sendAnnouncePayload(apitype.WebSocketMessage{})

	prefix := fmt.Sprintf("%s\nAPI.sendAnnouncePayload(marshal error)", packageName)

	// THEN: no message is sent to the announce channel.
	select {
	case msg := <-announceChannel:
		t.Fatalf(
			"%s unexpected message on AnnounceChannel\ngot:  %q\nwant: none",
			prefix, msg,
		)
	default:
	}
}

func TestGetParam(t *testing.T) {
	// GIVEN: a map of query parameters and a parameter to retrieve.
	tests := []struct {
		name        string
		queryParams url.Values
		param       string
		want        *string
	}{
		{
			name:        "param exists",
			queryParams: url.Values{"key": {"value"}},
			param:       "key",
			want:        test.Ptr("value"),
		},
		{
			name:        "param does not exist",
			queryParams: url.Values{"key": {"value"}},
			param:       "nonexistent",
			want:        nil,
		},
		{
			name:        "empty param",
			queryParams: url.Values{"key": {""}},
			param:       "key",
			want:        test.Ptr(""),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: getParam is called.
			got := getParam(tc.queryParams, tc.param)

			prefix := fmt.Sprintf(
				"%s\ngetParam(params=%+v, key=%q)",
				packageName, tc.queryParams, tc.param,
			)

			// THEN: the result should be as expected.
			if (got == nil && tc.want != nil) ||
				(got != nil && tc.want == nil) ||
				(got != nil && *got != *tc.want) {
				t.Errorf(
					"%s value mismatch\ngot:  %v\nwant: %v",
					prefix, got, tc.want,
				)
			}
		})
	}
}

func TestAPI_AnnounceEdit(t *testing.T) {
	// GIVEN: an API instance and old/new service data.
	announceChannel := make(chan []byte, 2)
	statusDefaults := status.NewDefaults(
		announceChannel,
		nil,
		nil,
	)
	api := &API{
		Config: &config.Config{
			HardDefaults: config.Defaults{
				Service: service.Defaults{
					Status: statusDefaults,
				},
			},
		},
	}

	tests := []struct {
		name              string
		oldData           *apitype.ServiceSummary
		newData           apitype.ServiceSummary
		wantedServiceData *apitype.ServiceSummary
	}{
		{
			name: "edit with old data/all change",
			oldData: &apitype.ServiceSummary{
				ID:   "service-1",
				Icon: test.Ptr("Service 1"),
			},
			newData: apitype.ServiceSummary{
				ID:   "service-2",
				Icon: test.Ptr("Service 1 Updated"),
			},
			wantedServiceData: &apitype.ServiceSummary{
				ID:   "service-2",
				Icon: test.Ptr("Service 1 Updated"),
			},
		},
		{
			name: "edit with old data/no changes",
			oldData: &apitype.ServiceSummary{
				ID:   "service-1",
				Icon: test.Ptr("Service 1"),
			},
			newData: apitype.ServiceSummary{
				ID:   "service-1",
				Icon: test.Ptr("Service 1"),
			},
			wantedServiceData: nil,
		},
		{
			name: "edit with old data/only changes sent",
			oldData: &apitype.ServiceSummary{
				ID:   "service-1",
				Icon: test.Ptr("Service 1"),
				Type: "github",
			},
			newData: apitype.ServiceSummary{
				ID:   "service-1",
				Icon: test.Ptr("Service 1"),
				Type: "url",
			},
			wantedServiceData: &apitype.ServiceSummary{
				Type: "url",
			},
		},
		{
			name:    "edit without old data",
			oldData: nil,
			newData: apitype.ServiceSummary{
				ID:   "service-2",
				Icon: test.Ptr("Service 2"),
			},
			wantedServiceData: &apitype.ServiceSummary{
				ID:   "service-2",
				Icon: test.Ptr("Service 2"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using a shared channel.

			tc.newData.Status = &apitype.Status{}
			if tc.oldData != nil {
				tc.oldData.Status = &apitype.Status{}
			}

			// WHEN: announceEdit is called.
			api.announceEdit(tc.oldData, &tc.newData)

			prefix := fmt.Sprintf("%s\nAPI.announceEdit()", packageName)

			// THEN: the message should be sent to the announce channel.
			select {
			case msg := <-announceChannel:
				var wsMessage apitype.WebSocketMessage
				err := decode.Unmarshal("json", msg, &wsMessage)
				if err != nil {
					t.Fatalf(
						"%s failed to unmarshal message from Announce channel: %v",
						prefix, err,
					)
				}

				// AND: the ServiceData should be as expected.
				wantedStr := decode.ToYAMLString(tc.wantedServiceData, "")
				gotStr := decode.ToYAMLString(wsMessage.ServiceData, "")
				if wsMessage.Page != "APPROVALS" ||
					wsMessage.Type != "EDIT" ||
					(tc.oldData != nil && wsMessage.SubType != tc.oldData.ID) ||
					(tc.oldData == nil && wsMessage.SubType != "") ||
					gotStr != wantedStr {
					t.Errorf(
						"%s unexpected WebSocketMessage in AnnounceChannel\ngot:  %q:\nwant: %q",
						prefix, gotStr, wantedStr,
					)
				}
			default:
				// Message not wanted.
				if tc.wantedServiceData == nil {
					return
				}
				t.Fatalf("%s Announce channel mismatch\ngot:  none\nwant: message", prefix)
			}
		})
	}
}

func TestAPI_AnnounceDelete(t *testing.T) {
	// GIVEN: an API instance and a serviceID.
	serviceID := "test-service"
	announceChannel := make(chan []byte, 2)
	statusDefaults := status.NewDefaults(
		announceChannel,
		nil,
		nil,
	)
	api := &API{
		Config: &config.Config{
			Order: []string{"some-order"},
			HardDefaults: config.Defaults{
				Service: service.Defaults{
					Status: statusDefaults,
				},
			},
		},
	}

	// WHEN: announceDelete is called.
	api.announceDelete(serviceID)

	prefix := fmt.Sprintf("%s\nAPI.announceDelete()", packageName)

	// THEN: the message should be sent to the announce channel.
	select {
	case msg := <-announceChannel:
		var wsMessage apitype.WebSocketMessage
		err := decode.Unmarshal("json", msg, &wsMessage)
		if err != nil {
			t.Fatalf(
				"%s failed to unmarshal message from Announce channel: %v",
				prefix, err,
			)
		}

		want := apitype.WebSocketMessage{
			Page:    "APPROVALS",
			Type:    "DELETE",
			SubType: serviceID,
		}
		if wsMessage.Page != "APPROVALS" ||
			wsMessage.Type != "DELETE" ||
			wsMessage.SubType != serviceID {
			t.Errorf(
				"%s unexpected WebSocketMessage in AnnounceChannel\ngot:  %+v:\nwant: %+v",
				prefix, wsMessage, want,
			)
		}
	default:
		t.Fatalf(
			"%s Announce channel mismatch\ngot:  none\nwant: message",
			prefix,
		)
	}
}

func TestAPI_AnnounceOrder(t *testing.T) {
	// GIVEN: an API instance with a service order.
	announceChannel := make(chan []byte, 2)
	statusDefaults := status.NewDefaults(
		announceChannel,
		nil,
		nil,
	)
	order := []string{"some-order"}
	api := &API{
		Config: &config.Config{
			Order: order,
			HardDefaults: config.Defaults{
				Service: service.Defaults{
					Status: statusDefaults,
				},
			},
		},
	}

	// WHEN: announceOrder is called.
	api.announceOrder()

	prefix := fmt.Sprintf("%s\nAPI.announceOrder()", packageName)

	// THEN: the message should be sent to the announce channel.
	select {
	case msg := <-announceChannel:
		var wsMessage apitype.WebSocketMessage
		err := decode.Unmarshal("json", msg, &wsMessage)
		if err != nil {
			t.Fatalf(
				"%s failed to unmarshal message from Announce channel: %v",
				prefix, err,
			)
		}

		want := apitype.WebSocketMessage{
			Page:    "APPROVALS",
			Type:    "SERVICE",
			SubType: "ORDER",
		}
		if wsMessage.Page != "APPROVALS" ||
			wsMessage.Type != "SERVICE" ||
			wsMessage.SubType != "ORDER" {
			t.Errorf(
				"%s unexpected WebSocketMessage in AnnounceChannel\ngot:  %+v:\nwant: %+v",
				prefix, wsMessage, want,
			)
		}

		if wsMessage.Order == nil {
			t.Fatalf(
				"%s Order missing from WebSocketMessage in AnnounceChannel\ngot:  none\nwant: order",
				prefix,
			)
		} else {
			if match := util.AreSlicesEqual(*wsMessage.Order, order); !match {
				t.Errorf(
					"%s Order mismatch in WebSocketMessage in AnnounceChannel\ngot:  %+v\nwant: %+v",
					prefix, *wsMessage.Order, order,
				)
			}
		}
	default:
		t.Fatalf(
			"%s Announce channel mismatch\ngot:  none\nwant: message",
			prefix,
		)
	}
}

func TestConstantTimeCompare(t *testing.T) {
	// GIVEN: two hashes.
	tests := []struct {
		name         string
		hash1, hash2 string
	}{
		{
			name:  "equal/1",
			hash1: "a",
			hash2: "a",
		},
		{
			name:  "equal/2",
			hash1: "abc",
			hash2: "abc",
		},
		{
			name:  "not equal/1",
			hash1: "a",
			hash2: "b",
		},
		{
			name:  "not equal/2",
			hash1: "abc",
			hash2: "abb",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			hash1 := sha256.Sum256([]byte(tc.hash1))
			hash2 := sha256.Sum256([]byte(tc.hash2))

			// WHEN: ConstantTimeCompare is called.
			got := ConstantTimeCompare(hash1, hash2)

			prefix := fmt.Sprintf(
				"%s\nConstantTimeCompare(a=%q, b=%q)",
				packageName, tc.hash1, tc.hash2,
			)

			// THEN: the result should be as expected.
			want := tc.hash1 == tc.hash2
			if got != want {
				t.Errorf(
					"%s value mismatch\ngot:  %t\nwant: %t",
					prefix, got, want,
				)
			}
		})
	}
}
