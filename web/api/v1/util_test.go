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
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/release-argus/Argus/auth/store"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
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
			want:        new("value"),
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
			want:        new(""),
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
				Icon: new("Service 1"),
			},
			newData: apitype.ServiceSummary{
				ID:   "service-2",
				Icon: new("Service 1 Updated"),
			},
			wantedServiceData: &apitype.ServiceSummary{
				ID:   "service-2",
				Icon: new("Service 1 Updated"),
			},
		},
		{
			name: "edit with old data/no changes",
			oldData: &apitype.ServiceSummary{
				ID:   "service-1",
				Icon: new("Service 1"),
			},
			newData: apitype.ServiceSummary{
				ID:   "service-1",
				Icon: new("Service 1"),
			},
			wantedServiceData: nil,
		},
		{
			name: "edit with old data/only changes sent",
			oldData: &apitype.ServiceSummary{
				ID:   "service-1",
				Icon: new("Service 1"),
				Type: "github",
			},
			newData: apitype.ServiceSummary{
				ID:   "service-1",
				Icon: new("Service 1"),
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
				Icon: new("Service 2"),
			},
			wantedServiceData: &apitype.ServiceSummary{
				ID:   "service-2",
				Icon: new("Service 2"),
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
				// Every successful edit announces, even when nothing displayed changed.
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

func TestValidatePassword(t *testing.T) {
	// GIVEN: passwords of varying lengths.
	tests := []struct {
		name      string
		plaintext string
		errRegex  string
	}{
		{
			name:      "meets the minimum length",
			plaintext: "12345678",
			errRegex:  `^$`,
		},
		{
			name:      "longer than the minimum",
			plaintext: "a-much-longer-password",
			errRegex:  `^$`,
		},
		{
			name:      "below the minimum length",
			plaintext: "1234567",
			errRegex:  `^password: "\*" <invalid> \(must be at least 8 characters\)$`,
		},
		{
			name:      "empty",
			plaintext: "",
			errRegex:  `^password: <required> \(must be at least 8 characters\)$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: validatePassword is called.
			err := validatePassword(tc.plaintext)

			prefix := fmt.Sprintf(
				"%s\nvalidatePassword(%q)",
				packageName, tc.plaintext,
			)

			// THEN: the error matches expectation.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
		})
	}
}

func TestAPI_DecodeAuthBody(t *testing.T) {
	// GIVEN: request bodies in varying states.
	type target struct {
		Name string `json:"name"`
	}
	tests := []struct {
		name       string
		body       string
		wantOK     bool
		wantName   string
		wantStatus int
		bodyRegex  string
	}{
		{
			name:     "valid JSON",
			body:     `{"name":"argus"}`,
			wantOK:   true,
			wantName: "argus",
		},
		{
			name:       "malformed JSON",
			body:       `{"name":`,
			wantStatus: http.StatusBadRequest,
			bodyRegex:  `parse request`,
		},
		{
			name:       "body exceeding the size cap",
			body:       `{"name":"` + strings.Repeat("x", maxAuthBodySize+1) + `"}`,
			wantStatus: http.StatusBadRequest,
			bodyRegex:  `read request`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			api := &API{}
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
				strings.NewReader(tc.body))
			w := httptest.NewRecorder()

			// WHEN: the body is decoded.
			var got target
			ok := api.decodeAuthBody(w, req, &got)

			prefix := fmt.Sprintf("%s\ndecodeAuthBody", packageName)

			// THEN: the outcome matches expectation.
			if ok != tc.wantOK {
				t.Fatalf(
					"%s ok mismatch\ngot:  %t\nwant: %t",
					prefix, ok, tc.wantOK,
				)
			}
			if tc.wantOK {
				// AND: the target is filled.
				if got.Name != tc.wantName {
					t.Errorf(
						"%s decoded value mismatch\ngot:  %q\nwant: %q",
						prefix, got.Name, tc.wantName,
					)
				}
				return
			}

			// AND: the request is failed with the expected status and message.
			if w.Code != tc.wantStatus {
				t.Errorf(
					"%s status mismatch\ngot:  %d\nwant: %d",
					prefix, w.Code, tc.wantStatus,
				)
			}
			if !util.RegexCheck(tc.bodyRegex, w.Body.String()) {
				t.Errorf(
					"%s body mismatch\ngot:  %q\nwant: %q",
					prefix, w.Body.String(), tc.bodyRegex,
				)
			}
		})
	}
}

func TestAPI_FailAuthStoreRequest(t *testing.T) {
	// GIVEN: store errors of each classification.
	tests := []struct {
		name       string
		err        error
		wantStatus int
		bodyRegex  string
	}{
		{
			name:       "not found",
			err:        store.ErrNotFound,
			wantStatus: http.StatusNotFound,
			bodyRegex:  `"delete failed: not found"`,
		},
		{
			name:       "wrapped not found",
			err:        fmt.Errorf("load user: %w", store.ErrNotFound),
			wantStatus: http.StatusNotFound,
			bodyRegex:  `"delete failed: not found"`,
		},
		{
			name:       "username taken",
			err:        store.ErrUsernameTaken,
			wantStatus: http.StatusBadRequest,
			bodyRegex:  `delete failed:\\n  username already taken`,
		},
		{
			name:       "group name taken",
			err:        store.ErrGroupNameTaken,
			wantStatus: http.StatusBadRequest,
			bodyRegex:  `delete failed:\\n  group name already taken`,
		},
		{
			name:       "unknown group",
			err:        store.ErrUnknownGroup,
			wantStatus: http.StatusBadRequest,
			bodyRegex:  `delete failed:\\n  unknown group`,
		},
		{
			name:       "invalid grant",
			err:        store.ErrInvalidGrant,
			wantStatus: http.StatusBadRequest,
			bodyRegex:  `delete failed:\\n  invalid grant`,
		},
		{
			name:       "last admin",
			err:        store.ErrLastAdmin,
			wantStatus: http.StatusConflict,
			bodyRegex:  `delete failed:\\n  cannot delete, disable or demote the last enabled admin`,
		},
		{
			name:       "system group",
			err:        store.ErrSystemGroup,
			wantStatus: http.StatusConflict,
			bodyRegex:  `delete failed:\\n  system group cannot be modified this way`,
		},
		{
			name:       "infrastructure failure",
			err:        errors.New("db broke"),
			wantStatus: http.StatusInternalServerError,
			bodyRegex:  `"delete failed"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			api := &API{}
			w := httptest.NewRecorder()

			// WHEN: the error is mapped onto a response.
			api.failAuthStoreRequest(w, tc.err,
				logx.LogFrom{Primary: "TestAPI_FailAuthStoreRequest"}, "delete")

			prefix := fmt.Sprintf("%s\nfailAuthStoreRequest", packageName)

			// THEN: the status and message match expectations.
			if w.Code != tc.wantStatus {
				t.Errorf(
					"%s status mismatch\ngot:  %d\nwant: %d",
					prefix, w.Code, tc.wantStatus,
				)
			}
			if !util.RegexCheck(tc.bodyRegex, w.Body.String()) {
				t.Errorf(
					"%s body mismatch\ngot:  %q\nwant: %q",
					prefix, w.Body.String(), tc.bodyRegex,
				)
			}

			// AND: internal detail is never leaked on infrastructure failures.
			if tc.wantStatus == http.StatusInternalServerError &&
				strings.Contains(w.Body.String(), "db broke") {
				t.Errorf(
					"%s internal error detail leaked\ngot:  %q",
					prefix, w.Body.String(),
				)
			}
		})
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
