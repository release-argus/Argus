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
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/release-argus/Argus/auth/rbac"
	"github.com/release-argus/Argus/auth/store"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// nonAdminGrants is a valid non-admin grant set.
func nonAdminGrants() []rbac.Grant {
	return []rbac.Grant{{
		Permission: rbac.Permission{Resource: rbac.ResourceService, Action: rbac.ActionRead},
		Scope:      rbac.Scope{Type: rbac.ScopeGlobal},
	}}
}

// seededGroupIDs maps the seeded groups' names to their IDs.
func seededGroupIDs(t *testing.T, deps *AuthDeps) map[string]string {
	t.Helper()

	groups, err := deps.Store.Groups(t.Context())
	if err != nil {
		t.Fatalf(
			"%s\nGroups() failed: %v",
			packageName, err,
		)
	}
	ids := make(map[string]string, len(groups))
	for _, group := range groups {
		ids[group.Name] = group.ID
	}
	return ids
}

func TestAPI_HTTPGroupList(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_HTTPGroupList.yml"
	api, deps, dbConn := testAuthServer(t, file)
	// AND: a logged-in admin.
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	authCtx := adminContext(t, api, deps)

	// WHEN: the admin lists groups.
	w := serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/groups", "", adminCookie))

	prefix := fmt.Sprintf("%s\nhttpGroupList()", packageName)

	// THEN: the seeded groups are returned.
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"%s\nstatus mismatch\ngot:  %d - %s\nwant: %d",
			prefix, got, w.Body.String(), want,
		)
	}
	var groups []store.Group
	if err := decode.Unmarshal("json", w.Body.Bytes(), &groups); err != nil {
		t.Fatalf(
			"%s\nparse list: %v",
			prefix, err,
		)
	}
	if len(groups) != 3 { // admin, operator, viewer.
		t.Errorf(
			"%s\ngroup count mismatch\ngot:  %d\nwant: 3",
			prefix, len(groups),
		)
	}

	// WHEN: the store breaks beneath a direct call.
	_ = dbConn.Close()
	recorder := httptest.NewRecorder()
	api.httpGroupList(recorder,
		withAuthCtx(httptest.NewRequest(http.MethodGet, "/api/v1/groups", nil), authCtx),
	)

	// THEN: the failure reads as 500.
	if recorder.Code != http.StatusInternalServerError {
		t.Errorf(
			"%s\nbroken store\ngot:  %d\nwant: 500",
			prefix, recorder.Code,
		)
	}
}

func TestAPI_HTTPGroupCreate(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_HTTPGroupCreate.yml"
	api, deps, _ := testAuthServer(t, file)
	// AND: a logged-in admin, and a logged-in non-admin.
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	if _, err := deps.Store.CreateGroup(t.Context(), "editors", "", nonAdminGrants()); err != nil {
		t.Fatalf(
			"%s\ncreate editors group: %v",
			packageName, err,
		)
	}
	createAuthUser(t, deps, "manager", "manager-password", "editors")
	managerCookie := loginCookie(t, api, "manager", "manager-password")

	// WHEN: the admin creates a group with scoped grants.
	w := serveAuth(api,
		authedRequest(http.MethodPost, "/api/v1/groups",
			test.TrimJSON(`{
				"name": "deployer",
				"description": "Deploy things.",
				"permissions": [
					{"resource": "service",        "action": "read",    "scope": {"type": "global"}},
					{"resource": "service_action", "action": "execute", "scope": {"type": "service", "ref": "argus"}}
				]
			}`),
			adminCookie,
		),
	)

	prefix := fmt.Sprintf("%s\nhttpGroupCreate()", packageName)

	// THEN: the group is created with its grants.
	if got, want := w.Code, http.StatusCreated; got != want {
		t.Fatalf(
			"%s\ncreate status mismatch\ngot:  %d - %s\nwant: %d",
			prefix, got, w.Body.String(), want,
		)
	}
	var created store.Group
	if err := decode.Unmarshal("json", w.Body.Bytes(), &created); err != nil {
		t.Fatalf(
			"%s\nparse create response: %v",
			prefix, err,
		)
	}
	if created.Name != "deployer" || len(created.Grants) != 2 {
		t.Errorf(
			"%s\ncreated group mismatch\ngot: %+v",
			prefix, created,
		)
	}

	// AND: a non-admin is forbidden from creating groups at all.
	if w := serveAuth(api,
		authedRequest(http.MethodPost, "/api/v1/groups",
			`{"name":"harmless"}`,
			managerCookie,
		),
	); w.Code != http.StatusForbidden {
		t.Errorf(
			"%s\nnon-admin group create\ngot:  %d - %s\nwant: 403",
			prefix, w.Code, w.Body.String(),
		)
	}

	// WHEN: an invalid or forbidden create arrives.
	tests := []struct {
		name       string
		body       string
		cookie     *http.Cookie
		wantStatus int
	}{
		{
			name:       "duplicate name",
			body:       `{"name":"deployer"}`,
			cookie:     adminCookie,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing name",
			body:       `{"description":"nameless"}`,
			cookie:     adminCookie,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unknown grant resource",
			body:       `{"name":"invalid","permissions":[{"resource":"nonsense","action":"read","scope":{"type":"global"}}]}`,
			cookie:     adminCookie,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "malformed body",
			body:       `{"name":`,
			cookie:     adminCookie,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "non-admin forbidden",
			body:       `{"name":"sneaky","permissions":[{"resource":"config","action":"read","scope":{"type":"global"}}]}`,
			cookie:     managerCookie,
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := serveAuth(api,
				authedRequest(http.MethodPost, "/api/v1/groups",
					tc.body,
					tc.cookie,
				),
			)

			// THEN: it is rejected with the expected status.
			if w.Code != tc.wantStatus {
				t.Errorf(
					"%s\n%s\ngot:  %d - %s\nwant: %d",
					prefix, tc.name, w.Code, w.Body.String(), tc.wantStatus,
				)
			}
		})
	}
}

func TestAPI_HTTPGroupGet(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_HTTPGroupGet.yml"
	api, deps, _ := testAuthServer(t, file)
	// AND: a logged-in admin and a target group.
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	target, err := deps.Store.CreateGroup(t.Context(), "target", "A target.", nil)
	if err != nil {
		t.Fatalf(
			"%s\ncreate target group: %v",
			packageName, err,
		)
	}

	// WHEN: the admin fetches the group by ID.
	w := serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/groups/"+target.ID, "", adminCookie),
	)

	prefix := fmt.Sprintf("%s\nhttpGroupGet()", packageName)

	// THEN: the group is returned.
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"%s\nstatus mismatch\ngot:  %d - %s\nwant: %d",
			prefix, got, w.Body.String(), want,
		)
	}
	var got store.Group
	if err := decode.Unmarshal("json", w.Body.Bytes(), &got); err != nil {
		t.Fatalf(
			"%s\nparse response: %v",
			prefix, err,
		)
	}
	if got.Name != "target" {
		t.Errorf(
			"%s\ngroup mismatch\ngot: %+v",
			prefix, got,
		)
	}

	// WHEN: an unknown ID is fetched.
	w = serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/groups/no-such-id", "", adminCookie))

	// THEN: the request 404s.
	if w.Code != http.StatusNotFound {
		t.Errorf(
			"%s\nget unknown\ngot:  %d\nwant: 404",
			prefix, w.Code,
		)
	}
}

func TestAPI_HTTPGroupUpdate(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_HTTPGroupUpdate.yml"
	api, deps, _ := testAuthServer(t, file)
	// AND: a logged-in admin.
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	if _, err := deps.Store.CreateGroup(t.Context(), "editors", "", nonAdminGrants()); err != nil {
		t.Fatalf(
			"%s\ncreate editors group: %v",
			packageName, err,
		)
	}
	// AND: a logged-in non-admin.
	createAuthUser(t, deps, "manager", "manager-password", "editors")
	managerCookie := loginCookie(t, api, "manager", "manager-password")
	// AND: a target group.
	target, err := deps.Store.CreateGroup(t.Context(), "target", "", nil)
	if err != nil {
		t.Fatalf(
			"%s\ncreate target group: %v",
			packageName, err,
		)
	}
	ids := seededGroupIDs(t, deps)

	// WHEN: the admin patches the group.
	w := serveAuth(api,
		authedRequest(http.MethodPatch, "/api/v1/groups/"+target.ID,
			test.TrimJSON(`{
				"name": "target-2",
				"permissions": [
					{"resource": "version_refresh", "action": "execute", "scope": {"type": "service_tag", "ref": "prod"}}
				]
			}`),
			adminCookie,
		),
	)

	prefix := fmt.Sprintf("%s\nhttpGroupUpdate()", packageName)

	// THEN: the patch is applied.
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"%s\npatch status mismatch\ngot:  %d - %s\nwant: %d",
			prefix, want, w.Body.String(), got,
		)
	}
	var patched store.Group
	if err := decode.Unmarshal("json", w.Body.Bytes(), &patched); err != nil {
		t.Fatalf(
			"%s\nparse patch response: %v",
			prefix, err,
		)
	}
	if patched.Name != "target-2" || len(patched.Grants) != 1 {
		t.Errorf(
			"%s\npatched group mismatch\ngot: %+v",
			prefix, patched,
		)
	}

	// AND: a non-admin is forbidden from patching groups.
	if w := serveAuth(api,
		authedRequest(http.MethodPatch, "/api/v1/groups/"+target.ID,
			`{"description":"renamed"}`,
			managerCookie,
		),
	); w.Code != http.StatusForbidden {
		t.Errorf(
			"%s\nnon-admin group update\ngot:  %d - %s\nwant: 403",
			prefix, w.Code, w.Body.String(),
		)
	}

	// WHEN: an invalid or forbidden patch arrives.
	tests := []struct {
		name       string
		target     string
		body       string
		cookie     *http.Cookie
		wantStatus int
	}{
		{
			name:       "non-admin forbidden",
			target:     target.ID,
			body:       `{"permissions":[{"resource":"config","action":"read","scope":{"type":"global"}}]}`,
			cookie:     managerCookie,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "rename the admin system group",
			target:     ids[store.GroupAdmin],
			body:       `{"name":"superusers"}`,
			cookie:     adminCookie,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "edit admin system group grants",
			target:     ids[store.GroupAdmin],
			body:       `{"permissions":[]}`,
			cookie:     adminCookie,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "duplicate name",
			target:     target.ID,
			body:       `{"name":"` + store.GroupOperator + `"}`,
			cookie:     adminCookie,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unknown group",
			target:     "no-such-id",
			body:       `{"description":"x"}`,
			cookie:     adminCookie,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "malformed body",
			target:     target.ID,
			body:       `{"name":`,
			cookie:     adminCookie,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := serveAuth(api,
				authedRequest(http.MethodPatch, "/api/v1/groups/"+tc.target,
					tc.body,
					tc.cookie,
				),
			)

			// THEN: it is rejected with the expected status.
			if w.Code != tc.wantStatus {
				t.Errorf(
					"%s\n%s\ngot:  %d - %s\nwant: %d",
					prefix, tc.name, w.Code, w.Body.String(), tc.wantStatus,
				)
			}
		})
	}

	// AND: a starter group can still be renamed (only admin is railed).
	if w := serveAuth(api,
		authedRequest(http.MethodPatch, "/api/v1/groups/"+ids[store.GroupViewer],
			`{"name":"watchers"}`,
			adminCookie,
		),
	); w.Code != http.StatusOK {
		t.Errorf(
			"%s\nrename starter group\ngot:  %d - %s\nwant: 200",
			prefix, w.Code, w.Body.String(),
		)
	}
}

func TestAPI_HTTPGroupUpdate__kicks(t *testing.T) {
	// GIVEN: an auth-enabled API with WebSocket clients for the admin,
	// a group member, and an outsider.
	file := "TestAPI_HTTPGroupUpdate__kicks.yml"
	api, deps, _ := testAuthServer(t, file)
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	group, err := deps.Store.CreateGroup(t.Context(), "crew", "", nil)
	if err != nil {
		t.Fatalf(
			"%s\ncreate crew group: %v",
			packageName, err,
		)
	}
	member := createAuthUser(t, deps, "member", "member-password", "crew")
	outsider := createAuthUser(t, deps, "outsider", "outsider-password")
	adminID := adminContext(t, api, deps).User.ID
	connect := wireHub(t, api)
	adminClient := connect(adminID, "admin-session")
	memberClient := connect(member.ID, "member-session")
	outsiderClient := connect(outsider.ID, "outsider-session")

	prefix := fmt.Sprintf("%s\nhttpGroupUpdate() kicks", packageName)

	// WHEN: only the group's name/description change.
	w := serveAuth(api,
		authedRequest(http.MethodPatch, "/api/v1/groups/"+group.ID,
			`{"name":"crew-renamed","description":"renamed"}`,
			adminCookie,
		),
	)
	if w.Code != http.StatusOK {
		t.Fatalf(
			"%s\nrename patch\ngot:  %d - %s\nwant: 200",
			prefix, w.Code, w.Body.String(),
		)
	}

	// THEN: nobody's grants changed - nobody is kicked.
	if !api.hub.hasClient(memberClient) ||
		!api.hub.hasClient(outsiderClient) ||
		!api.hub.hasClient(adminClient) {
		t.Errorf(
			"%s\nrename/description-only patch should kick nobody",
			prefix,
		)
	}

	// WHEN: the group's grants change.
	w = serveAuth(api,
		authedRequest(http.MethodPatch, "/api/v1/groups/"+group.ID,
			test.TrimJSON(`{
				"permissions": [
					{"resource": "service", "action": "read", "scope": {"type": "service_tag", "ref": "prod"}}
				]
			}`),
			adminCookie,
		),
	)
	if w.Code != http.StatusOK {
		t.Fatalf(
			"%s\ngrants patch\ngot:  %d - %s\nwant: 200",
			prefix, w.Code, w.Body.String(),
		)
	}

	// THEN: only the member's client is kicked - not the outsider's,
	// nor the editing admin's.
	if api.hub.hasClient(memberClient) ||
		!api.hub.hasClient(outsiderClient) ||
		!api.hub.hasClient(adminClient) {
		t.Errorf(
			"%s\ngrants patch should kick exactly the members' clients",
			prefix,
		)
	}
}

func TestAPI_HTTPGroupDelete(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_HTTPGroupDelete.yml"
	api, deps, _ := testAuthServer(t, file)
	// AND: a logged-in admin and a target group.
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	target, err := deps.Store.CreateGroup(t.Context(), "target", "", nil)
	if err != nil {
		t.Fatalf(
			"%s\ncreate target group: %v",
			packageName, err,
		)
	}
	ids := seededGroupIDs(t, deps)

	prefix := fmt.Sprintf("%s\nhttpGroupDelete()", packageName)

	// WHEN/THEN: deleting the admin system group is 409.
	if w := serveAuth(api,
		authedRequest(http.MethodDelete, "/api/v1/groups/"+ids[store.GroupAdmin], "", adminCookie),
	); w.Code != http.StatusConflict {
		t.Errorf(
			"%s\ndelete admin\ngot:  %d\nwant: 409",
			prefix, w.Code,
		)
	}

	// AND: unknown IDs 404.
	if w := serveAuth(api,
		authedRequest(http.MethodDelete, "/api/v1/groups/no-such-id", "", adminCookie),
	); w.Code != http.StatusNotFound {
		t.Errorf(
			"%s\ndelete unknown\ngot:  %d\nwant: 404",
			prefix, w.Code,
		)
	}

	// AND: a starter group (operator) can be deleted (only admin is railed).
	if w := serveAuth(api,
		authedRequest(http.MethodDelete, "/api/v1/groups/"+ids[store.GroupOperator], "", adminCookie),
	); w.Code != http.StatusNoContent {
		t.Errorf(
			"%s\ndelete starter group\ngot:  %d\nwant: 204",
			prefix, w.Code,
		)
	}

	// WHEN: the admin deletes the target group.
	w := serveAuth(api,
		authedRequest(http.MethodDelete, "/api/v1/groups/"+target.ID, "", adminCookie))

	// THEN: the group is gone.
	if got, want := w.Code, http.StatusNoContent; got != want {
		t.Fatalf(
			"%s\ndelete status mismatch\ngot:  %d - %s\nwant: %d",
			prefix, got, w.Body.String(), want,
		)
	}
	if _, err := deps.Store.GroupByID(t.Context(), target.ID); !errors.Is(err, store.ErrNotFound) {
		t.Errorf(
			"%s\ngroup should be deleted\ngot: %v",
			prefix, err,
		)
	}
}

func TestAPI_HTTPGroupDelete__kicks(t *testing.T) {
	// GIVEN: an auth-enabled API with WebSocket clients for the admin,
	// a group member, and an outsider.
	file := "TestAPI_HTTPGroupDelete__kicks.yml"
	api, deps, _ := testAuthServer(t, file)
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	group, err := deps.Store.CreateGroup(t.Context(), "crew", "", nil)
	if err != nil {
		t.Fatalf(
			"%s\ncreate crew group: %v",
			packageName, err,
		)
	}
	member := createAuthUser(t, deps, "member", "member-password", "crew")
	outsider := createAuthUser(t, deps, "outsider", "outsider-password")
	adminID := adminContext(t, api, deps).User.ID
	connect := wireHub(t, api)
	adminClient := connect(adminID, "admin-session")
	memberClient := connect(member.ID, "member-session")
	outsiderClient := connect(outsider.ID, "outsider-session")

	prefix := fmt.Sprintf("%s\nhttpGroupDelete() kicks", packageName)

	// WHEN: the admin deletes the group.
	w := serveAuth(api,
		authedRequest(http.MethodDelete, "/api/v1/groups/"+group.ID, "", adminCookie))
	if w.Code != http.StatusNoContent {
		t.Fatalf(
			"%s\ndelete\ngot:  %d - %s\nwant: 204",
			prefix, w.Code, w.Body.String(),
		)
	}

	// THEN: only the member's client is kicked (membership captured before
	// the delete's cascade) - not the outsider's, nor the editing admin's.
	if api.hub.hasClient(memberClient) ||
		!api.hub.hasClient(outsiderClient) ||
		!api.hub.hasClient(adminClient) {
		t.Errorf(
			"%s\ndelete should kick exactly the members' clients",
			prefix,
		)
	}
}

func TestAPI_HTTPPermissionCatalogue(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_HTTPPermissionCatalogue.yml"
	api, deps, _ := testAuthServer(t, file)
	// AND: a logged-in admin and viewer.
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	createAuthUser(t, deps, "viewer", "viewer-password", store.GroupViewer)
	viewerCookie := loginCookie(t, api, "viewer", "viewer-password")

	prefix := fmt.Sprintf("%s\nhttpPermissionCatalogue()", packageName)

	// WHEN: the admin fetches the permission catalogue.
	w := serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/permissions", "", adminCookie))

	// THEN: the valid matrix is returned.
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"%s\nstatus mismatch\ngot:  %d - %s\nwant: %d",
			prefix, got, w.Body.String(), want,
		)
	}
	var catalogue apitype.PermissionCatalogue
	if err := decode.Unmarshal("json", w.Body.Bytes(), &catalogue); err != nil {
		t.Fatalf(
			"%s\nparse response: %v",
			prefix, err,
		)
	}
	if len(catalogue.Resources) == 0 {
		t.Errorf("%s\ncatalogue should not be empty", prefix)
	}

	// AND: a viewer (no group:read) gets 403.
	if w := serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/permissions", "", viewerCookie),
	); w.Code != http.StatusForbidden {
		t.Errorf(
			"%s\nviewer access\ngot:  %d\nwant: 403",
			prefix, w.Code,
		)
	}
}
