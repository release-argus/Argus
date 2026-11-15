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
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/release-argus/Argus/auth/rbac"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service"
	svctest "github.com/release-argus/Argus/service/test"
	whtest "github.com/release-argus/Argus/webhook/test"
)

func TestAPI__auth__ServiceEdit__preventsSelfLockout(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	const editID = "retag-me"
	const outOfScopeID = "team-b-svc"
	const allTagsID = "all-tags-svc"
	const ownedID = "owned"

	svcA := testService(t, editID, "url", "url", true)
	svcA.Dashboard.Tags = []string{"team-a"}

	svcB := testService(t, outOfScopeID, "url", "url", true)
	svcB.Dashboard.Tags = []string{"team-b"}

	svcC := testService(t, allTagsID, "url", "url", true)
	svcC.Dashboard.Tags = []string{"team-a", "team-b"}

	svcOwned := testService(t, ownedID, "url", "url", true)

	api.Config.OrderMu.Lock()
	api.Config.Service[editID] = svcA
	api.Config.Service[outOfScopeID] = svcB
	api.Config.Service[allTagsID] = svcC
	api.Config.Service[ownedID] = svcOwned
	api.Config.Order = append(api.Config.Order, editID, outOfScopeID, allTagsID, ownedID)
	api.Config.OrderMu.Unlock()

	createAuthGroup(
		t,
		deps,
		"team-a-editors",
		"",
		[]rbac.Grant{
			{Resource: rbac.ResourceService, Action: rbac.ActionUpdate, Scope: rbac.Scope{Type: rbac.ScopeServiceTag, Ref: "team-a"}},
		},
	)
	createAuthUser(t, deps, "team-a-editor", "editor-password", "team-a-editors")

	createAuthGroup(
		t,
		deps,
		"owners",
		"",
		[]rbac.Grant{
			{Resource: rbac.ResourceService, Action: rbac.ActionUpdate, Scope: rbac.Scope{Type: rbac.ScopeService, Ref: ownedID}},
		},
	)
	createAuthUser(t, deps, "owner", "owner-password", "owners")

	tests := []struct {
		name       string
		user, pass string
		serviceID  string
		newID      string
		newTags    string
		wantStatus int
	}{
		{
			name: "valid/keep the authorising tag",
			user: "team-a-editor", pass: "editor-password",
			serviceID:  editID,
			newID:      editID,
			newTags:    `["team-a"]`,
			wantStatus: http.StatusOK,
		},
		{
			name: "valid/append a viewer tag",
			user: "team-a-editor", pass: "editor-password",
			serviceID:  editID,
			newID:      editID,
			newTags:    `["team-a", "team-b"]`,
			wantStatus: http.StatusOK,
		},
		{
			name: "valid/remove another tag",
			user: "team-a-editor", pass: "editor-password",
			serviceID:  allTagsID,
			newID:      allTagsID,
			newTags:    `["team-a"]`,
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid/tag editor replaces its authorising tag",
			user: "team-a-editor", pass: "editor-password",
			serviceID:  editID,
			newID:      editID,
			newTags:    `["team-b"]`,
			wantStatus: http.StatusForbidden, // The edit would lock the caller out.
		},
		{
			name: "invalid/tag editor strips all tags",
			user: "team-a-editor", pass: "editor-password",
			serviceID:  editID,
			newID:      editID,
			newTags:    `[]`,
			wantStatus: http.StatusForbidden, // The edit would lock the caller out.
		},
		{
			name: "invalid/tag editor edits a service outside its scope",
			user: "team-a-editor", pass: "editor-password",
			serviceID:  outOfScopeID,
			newID:      outOfScopeID,
			newTags:    `["team-a", "team-b"]`,
			wantStatus: http.StatusForbidden,
		},
		{
			name: "valid/service owner renames their own service",
			user: "owner", pass: "owner-password",
			serviceID:  ownedID,
			newID:      "owned-renamed", // The service-scoped grant is renamed too, so no lockout.
			newTags:    `[]`,
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: shares api.Config.Service.

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

			req := authedRequest(
				http.MethodPut, "/api/v1/service/config?service_id="+tc.serviceID, body, cookie,
			)
			w := serveAuth(api, req)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Fatalf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}

			if tc.wantStatus != http.StatusOK {
				return
			}
			var wantTags []string
			if err := decode.Unmarshal("json", []byte(tc.newTags), &wantTags); err != nil {
				t.Fatalf("decode.Unmarshal() failed: %v", err)
			}
			api.Config.OrderMu.RLock()
			svc := api.Config.Service[tc.newID]
			api.Config.OrderMu.RUnlock()
			if svc == nil {
				t.Fatalf("Config.Service[%q] mismatch\ngot:  absent\nwant: present", tc.newID)
			}
			if got := svc.Dashboard.Tags; !slices.Equal(got, wantTags) {
				t.Errorf("Service.Dashboard.Tags mismatch\ngot:  %q\nwant: %q", got, wantTags)
			}
		})
	}
}

func TestAPI__auth__notifyTestInheritRequiresServiceUpdate(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)

	createAuthGroup(
		t,
		deps,
		"notifier",
		"",
		[]rbac.Grant{
			{Resource: rbac.ResourceNotify, Action: rbac.ActionExecute, Scope: rbac.Scope{Type: rbac.ScopeGlobal}},
		},
	)
	createAuthUser(t, deps, "notifier-user", "notifier-password", "notifier")

	createAuthGroup(
		t,
		deps,
		"notifier-updater",
		"",
		[]rbac.Grant{
			{Resource: rbac.ResourceNotify, Action: rbac.ActionExecute, Scope: rbac.Scope{Type: rbac.ScopeGlobal}},
			{Resource: rbac.ResourceService, Action: rbac.ActionUpdate, Scope: rbac.Scope{Type: rbac.ScopeService, Ref: "test"}},
		},
	)
	createAuthUser(t, deps, "updater-user", "updater-password", "notifier-updater")

	tests := []struct {
		name       string
		username   string
		password   string
		body       string
		wantStatus int
	}{
		{
			name:       "invalid/inherit without service:update on the source",
			username:   "notifier-user",
			password:   "notifier-password",
			body:       `{"service_id_previous":"test","name_previous":"test"}`,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "valid/inherit with service:update on the source",
			username:   "updater-user",
			password:   "updater-password",
			body:       `{"service_id_previous":"test","name_previous":"test"}`,
			wantStatus: http.StatusBadRequest, // Passes the auth check, then fails on its payload.
		},
		{
			name:       "valid/no inheritance needs no service:update",
			username:   "notifier-user",
			password:   "notifier-password",
			body:       `{"service_id":"standalone"}`,
			wantStatus: http.StatusBadRequest, // Passes the auth check, then fails on its payload.
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cookie := loginCookie(t, api, tc.username, tc.password)

			req := authedRequest(http.MethodPost, "/api/v1/notify/test", tc.body, cookie)
			w := serveAuth(api, req)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
		})
	}
}

func TestAPI__auth__notifyTestRootInheritRequiresGlobalServiceUpdate(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	api.Config.Notify = shoutrrr.ShoutrrrsDefaults{
		"root-gotify": shoutrrr.NewDefaults(
			"gotify",
			nil,
			shoutrrr.MapStringStringOmitNull{"host": "gotify.example.com", "token": "root-secret"},
			nil,
		),
	}

	createAuthGroup(
		t,
		deps,
		"root-notifier",
		"",
		[]rbac.Grant{
			{Resource: rbac.ResourceNotify, Action: rbac.ActionExecute, Scope: rbac.Scope{Type: rbac.ScopeGlobal}},
		},
	)
	createAuthUser(t, deps, "root-notifier-user", "root-notifier-password", "root-notifier")

	createAuthGroup(
		t,
		deps,
		"root-notifier-scoped",
		"",
		[]rbac.Grant{
			{Resource: rbac.ResourceNotify, Action: rbac.ActionExecute, Scope: rbac.Scope{Type: rbac.ScopeGlobal}},
			{Resource: rbac.ResourceService, Action: rbac.ActionUpdate, Scope: rbac.Scope{Type: rbac.ScopeService, Ref: "test"}},
		},
	)
	createAuthUser(t, deps, "scoped-user", "scoped-password", "root-notifier-scoped")

	tests := []struct {
		name       string
		username   string
		password   string
		body       string
		wantStatus int
	}{
		{
			name:       "invalid/root inherit with notify:execute alone",
			username:   "root-notifier-user",
			password:   "root-notifier-password",
			body:       `{"service_id":"x","service_name":"x","name":"root-gotify"}`,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "invalid/root inherit with only service-scoped service:update",
			username:   "scoped-user",
			password:   "scoped-password",
			body:       `{"service_id":"x","service_name":"x","name":"root-gotify"}`,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "valid/root inherit with global service:update",
			username:   "admin",
			password:   "admin-password",
			body:       `{"service_id":"x","service_name":"x","name":"root-gotify"}`,
			wantStatus: http.StatusBadRequest, // Passes the auth check, then fails on its payload.
		},
		{
			name:       "valid/naming no root notifier needs no service:update",
			username:   "root-notifier-user",
			password:   "root-notifier-password",
			body:       `{"service_id":"x","service_name":"x","name":"not-a-root-notifier"}`,
			wantStatus: http.StatusBadRequest, // Passes the auth check, then fails on its payload.
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cookie := loginCookie(t, api, tc.username, tc.password)

			req := authedRequest(http.MethodPost, "/api/v1/notify/test", tc.body, cookie)
			w := serveAuth(api, req)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
		})
	}
}

func TestAPI__auth__versionRefreshOverridesRequireServiceUpdate(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)

	createAuthGroup(
		t,
		deps,
		"refresher",
		"",
		[]rbac.Grant{
			{Resource: rbac.ResourceVersionRefresh, Action: rbac.ActionExecute, Scope: rbac.Scope{Type: rbac.ScopeService, Ref: "test"}},
			{Resource: rbac.ResourceService, Action: rbac.ActionRead, Scope: rbac.Scope{Type: rbac.ScopeService, Ref: "test"}},
		},
	)
	createAuthUser(t, deps, "refresher-user", "refresher-password", "refresher")

	// "test" and any override query a local version server.
	serveVersion := func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("1.2.3")) }
	versionServer := httptest.NewServer(http.HandlerFunc(serveVersion))
	t.Cleanup(versionServer.Close)
	svc := test.Must(t, func() (*service.Service, error) {
		return service.DecodeService(
			"yaml", []byte(test.TrimYAML(`
				latest_version:
					type: url
					url: `+versionServer.URL+`
				deployed_version:
					type: url
					url: `+versionServer.URL+`
			`)),
			"test",
			svctest.PlainDefaultsConfig(t), shoutrrrtest.PlainConfig(), whtest.PlainConfig(t),
		)
	})
	api.Config.OrderMu.Lock()
	api.Config.Service["test"] = svc
	api.Config.OrderMu.Unlock()
	localOverrides := `{"url":"` + versionServer.URL + `"}`

	tests := []struct {
		name       string
		username   string
		password   string
		endpoint   string
		wantStatus int
	}{
		{
			name:       "valid/latest_version refresh without overrides",
			username:   "refresher-user",
			password:   "refresher-password",
			endpoint:   "/api/v1/latest_version/refresh?service_id=test",
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid/latest_version refresh with overrides",
			username:   "refresher-user",
			password:   "refresher-password",
			endpoint:   `/api/v1/latest_version/refresh?service_id=test&overrides={"url":"https://attacker.example.com"}`,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "valid/deployed_version refresh without overrides",
			username:   "refresher-user",
			password:   "refresher-password",
			endpoint:   "/api/v1/deployed_version/refresh?service_id=test",
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid/deployed_version refresh with overrides",
			username:   "refresher-user",
			password:   "refresher-password",
			endpoint:   `/api/v1/deployed_version/refresh?service_id=test&overrides={"url":"https://attacker.example.com"}`,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "valid/latest_version overrides with service:update",
			username:   "admin",
			password:   "admin-password",
			endpoint:   "/api/v1/latest_version/refresh?service_id=test&overrides=" + localOverrides,
			wantStatus: http.StatusOK,
		},
		{
			name:       "valid/deployed_version overrides with service:update",
			username:   "admin",
			password:   "admin-password",
			endpoint:   "/api/v1/deployed_version/refresh?service_id=test&overrides=" + localOverrides,
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cookie := loginCookie(t, api, tc.username, tc.password)

			req := authedRequest(http.MethodGet, tc.endpoint, "", cookie)
			w := serveAuth(api, req)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
		})
	}
}
