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
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
	apitype "github.com/release-argus/Argus/web/api/types"
)

func TestAPI_SendAnnouncePayload(t *testing.T) {
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

	api.sendAnnouncePayload(apitype.WebSocketMessage{})
	select {
	case <-announceChannel:
	default:
		t.Fatal("working marshal: Status.AnnounceChannel message count mismatch\ngot:  0\nwant: 1")
	}

	original := marshalAnnouncePayload
	customErr := errors.New("marshal failed")
	marshalAnnouncePayload = func(v any) ([]byte, error) {
		return nil, customErr
	}
	t.Cleanup(func() { marshalAnnouncePayload = original })

	releaseStdout := test.CaptureLog(t, logx.Default())
	api.sendAnnouncePayload(apitype.WebSocketMessage{})
	stdout := releaseStdout()

	select {
	case msg := <-announceChannel:
		t.Fatalf(
			"failing marshal: Status.AnnounceChannel message mismatch\ngot:  %q\nwant: absent",
			msg,
		)
	default:
	}

	stdoutRegex := `^ERROR: API sendAnnouncePayload, marshal failed\s$`
	if !util.RegexCheck(stdoutRegex, stdout) {
		t.Errorf("failing marshal: stdout mismatch\ngot:  %q\nwant: %q", stdout, stdoutRegex)
	}
}

func TestGetParam(t *testing.T) {
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

			got := getParam(tc.queryParams, tc.param)

			if (got == nil && tc.want != nil) ||
				(got != nil && tc.want == nil) ||
				(got != nil && *got != *tc.want) {
				t.Errorf(
					"mismatch\ngot:  %s\nwant: %s",
					test.StringifyPtr(got), test.StringifyPtr(tc.want),
				)
			}
		})
	}
}

func TestAPI_AnnounceEdit(t *testing.T) {
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
			// Serial: shares announceChannel.

			tc.newData.Status = &apitype.Status{}
			if tc.oldData != nil {
				tc.oldData.Status = &apitype.Status{}
			}

			api.announceEdit(tc.oldData, &tc.newData)

			select {
			case msg := <-announceChannel:
				var wsMessage apitype.WebSocketMessage
				err := decode.Unmarshal("json", msg, &wsMessage)
				if err != nil {
					t.Fatalf("decode.Unmarshal() failed: %v", err)
				}

				wantedStr := decode.ToYAMLString(tc.wantedServiceData, "")
				gotStr := decode.ToYAMLString(wsMessage.ServiceData, "")
				wantSubType := ""
				if tc.oldData != nil {
					wantSubType = tc.oldData.ID
				}
				if wsMessage.Page != "APPROVALS" ||
					wsMessage.Type != "EDIT" ||
					wsMessage.SubType != wantSubType ||
					gotStr != wantedStr {
					t.Errorf(
						"WebSocketMessage mismatch\n"+
							"got:  page=%q type=%q sub_type=%q service_data=%q\n"+
							"want: page=%q type=%q sub_type=%q service_data=%q",
						wsMessage.Page, wsMessage.Type, wsMessage.SubType, gotStr,
						"APPROVALS", "EDIT", wantSubType, wantedStr,
					)
				}
			default:
				// Every successful edit announces, even when nothing displayed changed.
				t.Fatal("Status.AnnounceChannel message count mismatch\ngot:  0\nwant: 1")
			}
		})
	}
}

func TestAPI_AnnounceDelete(t *testing.T) {
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

	api.announceDelete(serviceID)

	select {
	case msg := <-announceChannel:
		var wsMessage apitype.WebSocketMessage
		err := decode.Unmarshal("json", msg, &wsMessage)
		if err != nil {
			t.Fatalf("decode.Unmarshal() failed: %v", err)
		}

		want := apitype.WebSocketMessage{
			Page:    "APPROVALS",
			Type:    "DELETE",
			SubType: serviceID,
		}
		if wsMessage.Page != want.Page ||
			wsMessage.Type != want.Type ||
			wsMessage.SubType != want.SubType {
			t.Errorf(
				"WebSocketMessage mismatch\n"+
					"got:  page=%q type=%q sub_type=%q\n"+
					"want: page=%q type=%q sub_type=%q",
				wsMessage.Page, wsMessage.Type, wsMessage.SubType,
				want.Page, want.Type, want.SubType,
			)
		}
	default:
		t.Fatal("Status.AnnounceChannel message count mismatch\ngot:  0\nwant: 1")
	}
}

func TestAPI_AnnounceOrder(t *testing.T) {
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

	api.announceOrder()

	select {
	case msg := <-announceChannel:
		var wsMessage apitype.WebSocketMessage
		err := decode.Unmarshal("json", msg, &wsMessage)
		if err != nil {
			t.Fatalf("decode.Unmarshal() failed: %v", err)
		}

		want := apitype.WebSocketMessage{
			Page:    "APPROVALS",
			Type:    "SERVICE",
			SubType: "ORDER",
		}
		if wsMessage.Page != want.Page ||
			wsMessage.Type != want.Type ||
			wsMessage.SubType != want.SubType {
			t.Errorf(
				"WebSocketMessage mismatch\n"+
					"got:  page=%q type=%q sub_type=%q\n"+
					"want: page=%q type=%q sub_type=%q",
				wsMessage.Page, wsMessage.Type, wsMessage.SubType,
				want.Page, want.Type, want.SubType,
			)
		}

		if wsMessage.Order == nil {
			t.Fatal("WebSocketMessage.Order mismatch\ngot:  nil\nwant: non-nil")
		} else {
			if match := util.AreSlicesEqual(*wsMessage.Order, order); !match {
				t.Errorf(
					"WebSocketMessage.Order mismatch\ngot:  %q\nwant: %q",
					*wsMessage.Order, order,
				)
			}
		}
	default:
		t.Fatal("Status.AnnounceChannel message count mismatch\ngot:  0\nwant: 1")
	}
}

func TestValidatePassword(t *testing.T) {
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
		{
			name:      "above the maximum length",
			plaintext: strings.Repeat("a", maxPasswordLength+1),
			errRegex:  `^password: "\*" <invalid> \(must be at most \d+ characters\)$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validatePassword(tc.plaintext)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

func TestAPI_DecodeAuthBody(t *testing.T) {
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
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(tc.body))
			w := httptest.NewRecorder()

			var got target
			ok := api.decodeAuthBody(w, req, &got)

			if ok != tc.wantOK {
				t.Fatalf("ok mismatch\ngot:  %t\nwant: %t", ok, tc.wantOK)
			}
			if tc.wantOK {
				if got.Name != tc.wantName {
					t.Errorf("target.Name mismatch\ngot:  %q\nwant: %q", got.Name, tc.wantName)
				}
				return
			}

			if gotCode, wantCode := w.Code, tc.wantStatus; gotCode != wantCode {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					gotCode, w.Body.String(), wantCode,
				)
			}
			if !util.RegexCheck(tc.bodyRegex, w.Body.String()) {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", w.Body.String(), tc.bodyRegex)
			}
		})
	}
}

func TestAPI_FailAuthStoreRequest(t *testing.T) {
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
			wantStatus: http.StatusConflict,
			bodyRegex:  `delete failed:\\n  username already taken`,
		},
		{
			name:       "group name taken",
			err:        store.ErrGroupNameTaken,
			wantStatus: http.StatusConflict,
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
			bodyRegex:  `delete failed:\\n  cannot delete, disable, or demote the last enabled admin`,
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

			logFrom := logx.LogFrom{Primary: "TestAPI_FailAuthStoreRequest"}
			api.failAuthStoreRequest(w, tc.err, logFrom, "delete")

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
			gotBody := w.Body.String()
			if !util.RegexCheck(tc.bodyRegex, gotBody) {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", gotBody, tc.bodyRegex)
			}

			if tc.wantStatus == http.StatusInternalServerError &&
				strings.Contains(gotBody, "db broke") {
				t.Errorf("body mismatch\ngot:  %q\nwant: NOT %q", gotBody, "db broke")
			}
		})
	}
}

func TestConstantTimeCompare(t *testing.T) {
	tests := []struct {
		name           string
		input1, input2 string
	}{
		{
			name:   "equal/1",
			input1: "a",
			input2: "a",
		},
		{
			name:   "equal/2",
			input1: "abc",
			input2: "abc",
		},
		{
			name:   "not equal/1",
			input1: "a",
			input2: "b",
		},
		{
			name:   "not equal/2",
			input1: "abc",
			input2: "abb",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			hash1 := sha256.Sum256([]byte(tc.input1))
			hash2 := sha256.Sum256([]byte(tc.input2))

			got := ConstantTimeCompare(hash1, hash2)

			want := tc.input1 == tc.input2
			if got != want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, want)
			}
		})
	}
}
