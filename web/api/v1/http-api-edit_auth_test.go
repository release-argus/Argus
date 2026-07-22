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
	"github.com/release-argus/Argus/notify/shoutrrr"
)

func TestAPI__auth__ServiceEdit__preventsSelfLockout(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI__auth__ServiceEdit__preventsSelfLockout.yml"
	api, deps, _ := testAuthServer(t, file)
	const editID = "retag-me"
	const outOfScopeID = "team-b-svc"
	const allTagsID = "all-tags-svc"
	const ownedID = "owned"

	// AND: a service tagged "team-a"
	svcA := testService(t, editID, "url", "url", true)
	svcA.Dashboard.Tags = []string{"team-a"}
	// AND: a service tagged "team-b"
	svcB := testService(t, outOfScopeID, "url", "url", true)
	svcB.Dashboard.Tags = []string{"team-b"}
	// AND: a service tagged with both "team-a" and "team-b"
	svcC := testService(t, outOfScopeID, "url", "url", true)
	svcC.Dashboard.Tags = []string{"team-a", "team-b"}
	// AND: a service owned by a service-scoped grant.
	svcOwned := testService(t, ownedID, "url", "url", true)

	api.Config.OrderMu.Lock()
	api.Config.Service[editID] = svcA
	api.Config.Service[outOfScopeID] = svcB
	api.Config.Service[allTagsID] = svcC
	api.Config.Service[ownedID] = svcOwned
	api.Config.Order = append(api.Config.Order, editID, outOfScopeID, ownedID)
	api.Config.OrderMu.Unlock()

	// AND: a user whose only grant is service:update scoped to tag "team-a".
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
			"%s\ncreate 'team-a-editors' group: %v",
			packageName, err,
		)
	}
	createAuthUser(t, deps, "team-a-editor", "editor-password", "team-a-editors")

	// AND: a user with service:update scoped to the "owned" service.
	if _, err := deps.Store.CreateGroup(
		t.Context(),
		"owners",
		"",
		[]rbac.Grant{
			{
				Permission: rbac.Permission{Resource: rbac.ResourceService, Action: rbac.ActionUpdate},
				Scope:      rbac.Scope{Type: rbac.ScopeService, Ref: ownedID},
			},
		},
	); err != nil {
		t.Fatalf(
			"%s\ncreate 'owners' group: %v",
			packageName, err,
		)
	}
	createAuthUser(t, deps, "owner", "owner-password", "owners")

	// Adding tags is fine.
	// Removing other tags is fine.
	// Stripping the last tag/scope authorising the caller locks them out and is refused.
	// A rename also renames a service-scoped grant, so it is never a lockout.
	tests := []struct {
		name          string
		user, pass    string
		serviceID     string
		newID         string
		newTags       string
		wantForbidden bool
	}{
		{
			name: "valid/keep the authorising tag",
			user: "team-a-editor", pass: "editor-password",
			serviceID:     editID,
			newID:         editID,
			newTags:       `["team-a"]`,
			wantForbidden: false,
		},
		{
			name: "valid/append a viewer tag",
			user: "team-a-editor", pass: "editor-password",
			serviceID:     editID,
			newID:         editID,
			newTags:       `["team-a", "team-b"]`,
			wantForbidden: false,
		},
		{
			name: "valid/remove another tag",
			user: "team-a-editor", pass: "editor-password",
			serviceID:     allTagsID,
			newID:         allTagsID,
			newTags:       `["team-a"]`,
			wantForbidden: false,
		},
		{
			name: "invalid/tag editor replaces its authorising tag",
			user: "team-a-editor", pass: "editor-password",
			serviceID:     editID,
			newID:         editID,
			newTags:       `["team-b"]`,
			wantForbidden: true,
		},
		{
			name: "invalid/tag editor strips all tags",
			user: "team-a-editor", pass: "editor-password",
			serviceID:     editID,
			newID:         editID,
			newTags:       `[]`,
			wantForbidden: true,
		},
		{
			name: "invalid/tag editor edits a service outside its scope",
			user: "team-a-editor", pass: "editor-password",
			serviceID:     outOfScopeID,
			newID:         outOfScopeID,
			newTags:       `["team-a". "team-b"]`,
			wantForbidden: true,
		},
		{
			name: "valid/service owner renames their own service",
			user: "owner", pass: "owner-password",
			serviceID:     ownedID,
			newID:         "owned-renamed",
			newTags:       `[]`,
			wantForbidden: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Restore the starting tags.
			hadTags := api.Config.Service[editID].Dashboard.Tags
			t.Cleanup(func() {
				api.Config.OrderMu.Lock()
				api.Config.Service[editID].Dashboard.Tags = hadTags
				api.Config.OrderMu.Unlock()
			})

			cookie := loginCookie(t, api, tc.user, tc.pass)
			body := test.TrimJSON(`{
				"id": "` + tc.newID + `",
				"options": {"active": false},
				"latest_version": {"type": "github", "url": "` + test.ArgusGitHubRepo + `"},
				"dashboard": {"tags": ` + tc.newTags + `}
			}`)

			// WHEN: the service is edited.
			w := serveAuth(api, authedRequest(
				http.MethodPut,
				"/api/v1/service/config?service_id="+tc.serviceID,
				body, cookie,
			))

			prefix := fmt.Sprintf("%s\nservice edit self-lockout", packageName)

			// THEN: only an edit that removes the caller's own access is refused.
			gotForbidden := w.Code == http.StatusForbidden
			if gotForbidden != tc.wantForbidden {
				t.Errorf(
					"%s forbidden mismatch\ngot:  %t (status %d - %s)\nwant: %t",
					prefix,
					gotForbidden, w.Code, w.Body.String(),
					tc.wantForbidden,
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
			cookie := loginCookie(t, api, tc.username, tc.password)

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

func TestAPI__auth__notifyTestRootInheritRequiresGlobalServiceUpdate(t *testing.T) {
	// GIVEN: an auth-enabled API with a root notifier holding a secret.
	file := "TestAPI__Auth__NotifyTestRootInherit.yml"
	api, deps, _ := testAuthServer(t, file)
	api.Config.Notify = shoutrrr.ShoutrrrsDefaults{
		"root-gotify": shoutrrr.NewDefaults(
			"gotify",
			nil,
			shoutrrr.MapStringStringOmitNull{"host": "gotify.example.com", "token": "root-secret"},
			nil,
		),
	}

	// AND: a user holding notify:execute, but no service:update.
	if _, err := deps.Store.CreateGroup(
		t.Context(),
		"root-notifier",
		"",
		[]rbac.Grant{
			{
				Permission: rbac.Permission{Resource: rbac.ResourceNotify, Action: rbac.ActionExecute},
				Scope:      rbac.Scope{Type: rbac.ScopeGlobal},
			},
		},
	); err != nil {
		t.Fatalf(
			"%s\ncreate root-notifier group: %v",
			packageName, err,
		)
	}
	createAuthUser(t, deps, "root-notifier-user", "root-notifier-password", "root-notifier")

	// AND: a user holding notify:execute, and service:update on "test" only.
	if _, err := deps.Store.CreateGroup(
		t.Context(),
		"root-notifier-scoped",
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
			"%s\ncreate root-notifier-scoped group: %v",
			packageName, err,
		)
	}
	createAuthUser(t, deps, "scoped-user", "scoped-password", "root-notifier-scoped")

	tests := []struct {
		name          string
		username      string
		password      string
		body          string
		wantForbidden bool
	}{
		{
			name:          "invalid/root inherit with notify:execute alone",
			username:      "root-notifier-user",
			password:      "root-notifier-password",
			body:          `{"service_id":"x","service_name":"x","name":"root-gotify"}`,
			wantForbidden: true,
		},
		{
			name:          "invalid/root inherit with only service-scoped service:update",
			username:      "scoped-user",
			password:      "scoped-password",
			body:          `{"service_id":"x","service_name":"x","name":"root-gotify"}`,
			wantForbidden: true,
		},
		{
			name:          "valid/root inherit with global service:update",
			username:      "admin",
			password:      "admin-password",
			body:          `{"service_id":"x","service_name":"x","name":"root-gotify"}`,
			wantForbidden: false,
		},
		{
			name:          "valid/naming no root notifier needs no service:update",
			username:      "root-notifier-user",
			password:      "root-notifier-password",
			body:          `{"service_id":"x","service_name":"x","name":"not-a-root-notifier"}`,
			wantForbidden: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cookie := loginCookie(t, api, tc.username, tc.password)

			// WHEN: a notify test naming that notifier is requested.
			w := serveAuth(
				api,
				authedRequest(http.MethodPost, "/api/v1/notify/test", tc.body, cookie),
			)

			prefix := fmt.Sprintf("%s\nroot notify test inheritance", packageName)

			// THEN: only a caller lacking global service:update is refused.
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

func TestAPI__auth__versionRefreshOverridesRequireServiceUpdate(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI__auth__versionRefreshOverridesRequireServiceUpdate.yml"
	api, deps, _ := testAuthServer(t, file)

	// AND: a user able to refresh and read "test", but not update it.
	if _, err := deps.Store.CreateGroup(
		t.Context(),
		"refresher",
		"",
		[]rbac.Grant{
			{
				Permission: rbac.Permission{Resource: rbac.ResourceVersionRefresh, Action: rbac.ActionExecute},
				Scope:      rbac.Scope{Type: rbac.ScopeService, Ref: "test"},
			},
			{
				Permission: rbac.Permission{Resource: rbac.ResourceService, Action: rbac.ActionRead},
				Scope:      rbac.Scope{Type: rbac.ScopeService, Ref: "test"},
			},
		},
	); err != nil {
		t.Fatalf(
			"%s\ncreate refresher group: %v",
			packageName, err,
		)
	}
	createAuthUser(t, deps, "refresher-user", "refresher-password", "refresher")

	tests := []struct {
		name          string
		username      string
		password      string
		endpoint      string
		wantForbidden bool
	}{
		{
			name:          "valid/latest_version refresh without overrides",
			username:      "refresher-user",
			password:      "refresher-password",
			endpoint:      "/api/v1/latest_version/refresh?service_id=test",
			wantForbidden: false,
		},
		{
			name:          "invalid/latest_version refresh with overrides",
			username:      "refresher-user",
			password:      "refresher-password",
			endpoint:      `/api/v1/latest_version/refresh?service_id=test&overrides={"url":"https://attacker.example.com"}`,
			wantForbidden: true,
		},
		{
			name:          "valid/deployed_version refresh without overrides",
			username:      "refresher-user",
			password:      "refresher-password",
			endpoint:      "/api/v1/deployed_version/refresh?service_id=test",
			wantForbidden: false,
		},
		{
			name:          "invalid/deployed_version refresh with overrides",
			username:      "refresher-user",
			password:      "refresher-password",
			endpoint:      `/api/v1/deployed_version/refresh?service_id=test&overrides={"url":"https://attacker.example.com"}`,
			wantForbidden: true,
		},
		{
			name:          "valid/latest_version overrides with service:update",
			username:      "admin",
			password:      "admin-password",
			endpoint:      `/api/v1/latest_version/refresh?service_id=test&overrides={"url":"https://attacker.example.com"}`,
			wantForbidden: false,
		},
		{
			name:          "valid/deployed_version overrides with service:update",
			username:      "admin",
			password:      "admin-password",
			endpoint:      `/api/v1/deployed_version/refresh?service_id=test&overrides={"url":"https://attacker.example.com"}`,
			wantForbidden: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cookie := loginCookie(t, api, tc.username, tc.password)

			// WHEN: the refresh is requested.
			w := serveAuth(api, authedRequest(http.MethodGet, tc.endpoint, "", cookie))

			prefix := fmt.Sprintf("%s\nversion refresh overrides", packageName)

			// THEN: only overrides from a caller lacking service:update are refused.
			gotForbidden := w.Code == http.StatusForbidden
			if gotForbidden != tc.wantForbidden {
				t.Errorf(
					"%s\nforbidden mismatch\ngot:  %t (status %d - %s)\nwant: %t",
					prefix,
					gotForbidden, w.Code, w.Body.String(),
					tc.wantForbidden,
				)
			}
		})
	}
}
