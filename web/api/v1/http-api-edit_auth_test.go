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
	"fmt"
	"net/http"
	"testing"

	"github.com/release-argus/Argus/auth/rbac"
	"github.com/release-argus/Argus/internal/test"
)

func TestAPI__auth__ServiceEdit__reauthorisesResultingTarget(t *testing.T) {
	// GIVEN: an auth-enabled API with a service tagged "team-a".
	file := "TestAPI__auth__ServiceEdit__reauthorisesResultingTarget.yml"
	api, deps, _ := testAuthServer(t, file)
	const editID = "retag-me"
	svc := testService(t, editID, "url", "url", true)
	svc.Dashboard.Tags = []string{"team-a"}
	api.Config.OrderMu.Lock()
	api.Config.Service[editID] = svc
	api.Config.Order = append(api.Config.Order, editID)
	api.Config.OrderMu.Unlock()

	// AND: a user whose only grant is service:update scoped to that tag.
	if _, err := deps.Store.CreateGroup(
		t.Context(),
		"team-a-editors",
		"",
		[]rbac.Grant{
			{
				Permission: rbac.Permission{Resource: rbac.ResourceService, Action: rbac.ActionUpdate},
				Scope:      rbac.Scope{Type: rbac.ScopeServiceTag, Ref: "team-a"},
			},
		},
	); err != nil {
		t.Fatalf(
			"%s\ncreate team-a-editors group: %v",
			packageName, err,
		)
	}
	createAuthUser(t, deps, "team-a-editor", "editor-password", "team-a-editors")

	tests := []struct {
		name          string
		tags          string
		wantForbidden bool
	}{
		{
			name:          "valid/edit keeping the service in its own scope",
			tags:          `["team-a"]`,
			wantForbidden: false,
		},
		{
			name:          "invalid/retag into a scope the editor does not hold",
			tags:          `["team-b"]`,
			wantForbidden: true,
		},
		{
			name:          "invalid/strip the tag that authorised the edit",
			tags:          `[]`,
			wantForbidden: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Restore the starting tag.
			api.Config.OrderMu.Lock()
			api.Config.Service[editID].Dashboard.Tags = []string{"team-a"}
			api.Config.OrderMu.Unlock()

			cookie := loginCookie(t, api, "team-a-editor", "editor-password")
			body := test.TrimJSON(`{
				"id": "` + editID + `",
				"options": {"active": false},
				"latest_version": {"type": "github", "url": "` + test.ArgusGitHubRepo + `"},
				"dashboard": {"tags": ` + tc.tags + `}
			}`)

			// WHEN: the service is edited with a given tag set.
			w := serveAuth(api, authedRequest(
				http.MethodPut,
				"/api/v1/service/config?service_id="+editID,
				body, cookie,
			))

			prefix := fmt.Sprintf("%s\nservice edit re-authorisation", packageName)

			// THEN: an edit whose result leaves the editor's scope is refused.
			gotForbidden := w.Code == http.StatusForbidden
			if gotForbidden != tc.wantForbidden {
				t.Errorf(
					"%s forbidden mismatch\ngot:  %t (status %d - %s)\nwant: %t",
					prefix, gotForbidden, w.Code, w.Body.String(), tc.wantForbidden,
				)
			}
		})
	}
}

func TestAPI__auth__notifyTestInheritRequiresServiceUpdate(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI__Auth__NotifyTestInherit.yml"
	api, deps, _ := testAuthServer(t, file)

	// AND: a user holding notify:execute but no service:update.
	if _, err := deps.Store.CreateGroup(
		t.Context(),
		"notifier",
		"",
		[]rbac.Grant{
			{
				Permission: rbac.Permission{Resource: rbac.ResourceNotify, Action: rbac.ActionExecute},
				Scope:      rbac.Scope{Type: rbac.ScopeGlobal},
			},
		},
	); err != nil {
		t.Fatalf(
			"%s\ncreate notifier group: %v",
			packageName, err,
		)
	}
	createAuthUser(t, deps, "notifier-user", "notifier-password", "notifier")

	// AND: a user holding notify:execute, and service:update on "test".
	if _, err := deps.Store.CreateGroup(
		t.Context(),
		"notifier-updater",
		"",
		[]rbac.Grant{
			{
				Permission: rbac.Permission{Resource: rbac.ResourceNotify, Action: rbac.ActionExecute},
				Scope:      rbac.Scope{Type: rbac.ScopeGlobal},
			},
			{
				Permission: rbac.Permission{Resource: rbac.ResourceService, Action: rbac.ActionUpdate},
				Scope:      rbac.Scope{Type: rbac.ScopeService, Ref: "test"},
			},
		},
	); err != nil {
		t.Fatalf(
			"%s\ncreate notifier-updater group: %v",
			packageName, err,
		)
	}
	createAuthUser(t, deps, "updater-user", "updater-password", "notifier-updater")

	tests := []struct {
		name          string
		username      string
		password      string
		body          string
		wantForbidden bool
	}{
		{
			name:          "invalid/inherit without auth",
			username:      "notifier-user",
			password:      "notifier-password",
			body:          `{"service_id_previous":"test","name_previous":"test"}`,
			wantForbidden: true,
		},
		{
			name:          "invalid/inherit without service:update on the source",
			username:      "notifier-user",
			password:      "notifier-password",
			body:          `{"service_id_previous":"test","name_previous":"test"}`,
			wantForbidden: true,
		},
		{
			name:          "valid/inherit with service:update on the source",
			username:      "updater-user",
			password:      "updater-password",
			body:          `{"service_id_previous":"test","name_previous":"test"}`,
			wantForbidden: false,
		},
		{
			name:          "valid/no inheritance needs no service:update",
			username:      "notifier-user",
			password:      "notifier-password",
			body:          `{"service_id":"standalone"}`,
			wantForbidden: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var cookie *http.Cookie
			if tc.username != "" && tc.password != "" {
				cookie = loginCookie(t, api, tc.username, tc.password)
			}

			// WHEN: a notify test is requested.
			w := serveAuth(
				api,
				authedRequest(http.MethodPost, "/api/v1/notify/test", tc.body, cookie),
			)

			prefix := fmt.Sprintf("%s\nnotify test inheritance", packageName)

			// THEN: only a caller lacking update on the source service is refused.
			gotForbidden := w.Code == http.StatusForbidden
			if gotForbidden != tc.wantForbidden {
				t.Errorf(
					"%s\nforbidden mismatch\ngot:  %t (status %d - %s)\nwant: %t",
					prefix, gotForbidden, w.Code, w.Body.String(), tc.wantForbidden,
				)
			}
		})
	}
}
