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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	"github.com/release-argus/Argus/auth/rbac"
	"github.com/release-argus/Argus/auth/store"
	storetest "github.com/release-argus/Argus/auth/store/test"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// nonAdminGrants is a valid non-admin grant set.
func nonAdminGrants() []rbac.Grant {
	return []rbac.Grant{
		{Resource: rbac.ResourceService, Action: rbac.ActionRead, Scope: rbac.Scope{Type: rbac.ScopeGlobal}},
	}
}

// seededGroupIDs maps the seeded groups' names to their IDs.
func seededGroupIDs(t *testing.T, deps *AuthDeps) map[string]string {
	t.Helper()

	groups, err := deps.Store.Groups(t.Context())
	if err != nil {
		t.Fatalf("Store.Groups() failed: %v", err)
	}
	ids := make(map[string]string, len(groups))
	for _, group := range groups {
		ids[group.Name] = group.ID
	}
	return ids
}

func TestAPI_HTTPGroupList(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, dbConn := testAuthServer(t, file)

	adminCookie := loginCookie(t, api, "admin", "admin-password")
	authCtx := adminContext(t, api, deps)

	req := authedRequest(http.MethodGet, "/api/v1/groups", "", adminCookie)
	w := serveAuth(api, req)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("list: status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
	}
	var groups []store.Group
	if err := decode.Unmarshal("json", w.Body.Bytes(), &groups); err != nil {
		t.Fatalf("list: decode.Unmarshal() failed: %v", err)
	}
	groupNames := test.AttributesSorted(groups, func(group store.Group) string { return group.Name })
	if testErr := test.AssertSlicesEqual(
		t, groupNames, []string{store.GroupAdmin, store.GroupOperator, store.GroupViewer}, "Group.Name",
	); testErr != nil {
		t.Errorf("list: %v", testErr)
	}

	_ = dbConn.Close()
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/groups", nil)
	req = withAuthCtx(req, authCtx)
	api.httpGroupList(w, req)

	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf(
			"broken store: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
}

func TestAPI_HTTPGroupCreate(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)

	adminCookie := loginCookie(t, api, "admin", "admin-password")
	createAuthGroup(t, deps, "editors", "", nonAdminGrants())
	createAuthUser(t, deps, "manager", "manager-password", "editors")
	managerCookie := loginCookie(t, api, "manager", "manager-password")

	req := authedRequest(
		http.MethodPost, "/api/v1/groups",
		test.TrimJSON(`{
			"name": "deployer",
			"description": "Deploy things.",
			"permissions": [
				{"resource": "service",        "action": "read",    "scope": {"type": "global"}},
				{"resource": "service_action", "action": "execute", "scope": {"type": "service", "ref": "argus"}}
			]
		}`),
		adminCookie,
	)
	w := serveAuth(api, req)

	if got, want := w.Code, http.StatusCreated; got != want {
		t.Fatalf(
			"admin creating group: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	var created store.Group
	if err := decode.Unmarshal("json", w.Body.Bytes(), &created); err != nil {
		t.Fatalf("admin creating group: decode.Unmarshal() failed: %v", err)
	}
	wantName, wantGrants := "deployer", 2
	if created.Name != wantName || len(created.Grants) != wantGrants {
		t.Errorf(
			"admin creating group: Group mismatch\n"+
				"got:  Name=%q, len(Grants)=%d\nwant: Name=%q, len(Grants)=%d",
			created.Name, len(created.Grants),
			wantName, wantGrants,
		)
	}

	req = authedRequest(http.MethodPost, "/api/v1/groups", `{"name":"harmless"}`, managerCookie)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusForbidden; got != want {
		t.Errorf(
			"non-admin creating group: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

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
			wantStatus: http.StatusConflict,
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
			req := authedRequest(http.MethodPost, "/api/v1/groups", tc.body, tc.cookie)
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

func TestAPI_HTTPGroupCreate__nameBounds(t *testing.T) {
	file := t.Name() + ".yml"
	api, _, _ := testAuthServer(t, file)
	adminCookie := loginCookie(t, api, "admin", "admin-password")

	tooLong := strings.Repeat("x", maxFieldLength+1)
	atLimit := strings.Repeat("y", maxFieldLength)
	atLimit2 := strings.Repeat("z", maxFieldLength)

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantName   string // The name the group is stored under ("" to skip).
	}{
		{
			name:       "valid/surrounding whitespace is trimmed",
			body:       `{"name":"  spaced  "}`,
			wantStatus: http.StatusCreated,
			wantName:   "spaced",
		},
		{
			name:       "valid/name at the limit",
			body:       `{"name":"` + atLimit + `"}`,
			wantStatus: http.StatusCreated,
			wantName:   atLimit,
		},
		{
			name:       "valid/name at the limit after whitespace trimmed",
			body:       `{"name":"` + atLimit2 + ` "}`,
			wantStatus: http.StatusCreated,
			wantName:   atLimit2,
		},
		{
			name:       "invalid/name over the limit",
			body:       `{"name":"` + tooLong + `"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid/whitespace-only name",
			body:       `{"name":"   "}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid/description over the limit",
			body:       `{"name":"described","description":"` + tooLong + `"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid/padded system group name collides",
			body:       `{"name":"  admin  "}`,
			wantStatus: http.StatusConflict,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := authedRequest(http.MethodPost, "/api/v1/groups", tc.body, adminCookie)
			w := serveAuth(api, req)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Fatalf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
			if tc.wantName == "" {
				return
			}

			var created store.Group
			if err := decode.Unmarshal("json", w.Body.Bytes(), &created); err != nil {
				t.Fatalf("decode.Unmarshal() failed: %v", err)
			}
			if got, want := created.Name, tc.wantName; got != want {
				t.Errorf("Group.Name mismatch\ngot:  %q\nwant: %q", got, want)
			}
		})
	}
}

func TestAPI_HTTPGroupGet(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)

	adminCookie := loginCookie(t, api, "admin", "admin-password")
	target := createAuthGroup(t, deps, "target", "A target.", nil)

	req := authedRequest(http.MethodGet, "/api/v1/groups/"+target.ID, "", adminCookie)
	w := serveAuth(api, req)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"get group by ID: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	var got store.Group
	if err := decode.Unmarshal("json", w.Body.Bytes(), &got); err != nil {
		t.Fatalf("get group by ID: decode.Unmarshal() failed: %v", err)
	}
	if want := "target"; got.Name != want {
		t.Errorf("get group by ID: Group.Name mismatch\ngot:  %q\nwant: %q", got.Name, want)
	}

	req = authedRequest(http.MethodGet, "/api/v1/groups/no-such-id", "", adminCookie)
	w = serveAuth(api, req)

	if got, want := w.Code, http.StatusNotFound; got != want {
		t.Errorf(
			"get unknown group: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
}

func TestAPI_HTTPGroupUpdate(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)

	adminCookie := loginCookie(t, api, "admin", "admin-password")
	createAuthGroup(t, deps, "editors", "", nonAdminGrants())

	createAuthUser(t, deps, "manager", "manager-password", "editors")
	managerCookie := loginCookie(t, api, "manager", "manager-password")

	target := createAuthGroup(t, deps, "target", "", nil)
	ids := seededGroupIDs(t, deps)

	req := authedRequest(
		http.MethodPatch, "/api/v1/groups/"+target.ID,
		test.TrimJSON(`{
			"name": "target-2",
			"permissions": [
				{"resource": "version_refresh", "action": "execute", "scope": {"type": "service_tag", "ref": "prod"}}
			]
		}`),
		adminCookie,
	)
	w := serveAuth(api, req)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"admin patching group: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	var patched store.Group
	if err := decode.Unmarshal("json", w.Body.Bytes(), &patched); err != nil {
		t.Fatalf("admin patching group: decode.Unmarshal() failed: %v", err)
	}
	wantName, wantGrants := "target-2", 1
	if patched.Name != wantName || len(patched.Grants) != wantGrants {
		t.Errorf(
			"admin patching group: Group mismatch\n"+
				"got:  Name=%q, len(Grants)=%d\nwant: Name=%q, len(Grants)=%d",
			patched.Name, len(patched.Grants),
			wantName, wantGrants,
		)
	}

	req = authedRequest(
		http.MethodPatch, "/api/v1/groups/"+target.ID, `{"description":"renamed"}`, managerCookie,
	)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusForbidden; got != want {
		t.Errorf(
			"non-admin patching group: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

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
			wantStatus: http.StatusConflict,
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
		{
			name:       "blank name",
			target:     target.ID,
			body:       `{"name":"   "}`,
			cookie:     adminCookie,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "over-long description",
			target:     target.ID,
			body:       `{"description":"` + strings.Repeat("x", maxFieldLength+1) + `"}`,
			cookie:     adminCookie,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := authedRequest(http.MethodPatch, "/api/v1/groups/"+tc.target, tc.body, tc.cookie)
			w := serveAuth(api, req)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
		})
	}

	// Only the admin system group is protected, so a starter group can be renamed.
	req = authedRequest(
		http.MethodPatch, "/api/v1/groups/"+ids[store.GroupViewer],
		`{"name":"watchers"}`, adminCookie,
	)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf(
			"rename starter group: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
}

func TestAPI_HTTPGroupUpdate__kicks(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	group := createAuthGroup(t, deps, "crew", "", nil)
	member := createAuthUser(t, deps, "member", "member-password", "crew")
	outsider := createAuthUser(t, deps, "outsider", "outsider-password")
	adminID := adminContext(t, api, deps).User.ID
	connect := wireHub(t, api)
	adminClient := connect(adminID, "admin-session")
	memberClient := connect(member.ID, "member-session")
	outsiderClient := connect(outsider.ID, "outsider-session")

	req := authedRequest(
		http.MethodPatch, "/api/v1/groups/"+group.ID,
		`{"name":"crew-renamed","description":"renamed"}`, adminCookie,
	)
	w := serveAuth(api, req)
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"rename/description patch: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	// No grants changed, so nobody is kicked.
	memberKept := api.hub.hasClient(memberClient)
	outsiderKept := api.hub.hasClient(outsiderClient)
	adminKept := api.hub.hasClient(adminClient)
	if !memberKept || !outsiderKept || !adminKept {
		t.Errorf(
			"rename/description patch: Hub.hasClient() mismatch\n"+
				"got:  member=%t, outsider=%t, admin=%t\n"+
				"want: member=true, outsider=true, admin=true",
			memberKept, outsiderKept, adminKept,
		)
	}

	req = authedRequest(
		http.MethodPatch, "/api/v1/groups/"+group.ID,
		test.TrimJSON(`{
			"permissions": [
				{"resource": "service", "action": "read", "scope": {"type": "service_tag", "ref": "prod"}}
			]
		}`),
		adminCookie,
	)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"grants patch: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	memberKept = api.hub.hasClient(memberClient)
	outsiderKept = api.hub.hasClient(outsiderClient)
	adminKept = api.hub.hasClient(adminClient)
	if memberKept || !outsiderKept || !adminKept {
		t.Errorf(
			"grants patch: Hub.hasClient() mismatch\n"+
				"got:  member=%t, outsider=%t, admin=%t\n"+
				"want: member=false, outsider=true, admin=true",
			memberKept, outsiderKept, adminKept,
		)
	}
}

func TestAPI_HTTPGroupDelete(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)

	adminCookie := loginCookie(t, api, "admin", "admin-password")
	target := createAuthGroup(t, deps, "target", "", nil)
	ids := seededGroupIDs(t, deps)

	req := authedRequest(http.MethodDelete, "/api/v1/groups/"+ids[store.GroupAdmin], "", adminCookie)
	w := serveAuth(api, req)
	if got, want := w.Code, http.StatusConflict; got != want {
		t.Errorf(
			"delete admin group: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	req = authedRequest(http.MethodDelete, "/api/v1/groups/no-such-id", "", adminCookie)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusNotFound; got != want {
		t.Errorf(
			"delete unknown group: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	// Only the admin system group is protected, so a starter group can be deleted.
	req = authedRequest(http.MethodDelete, "/api/v1/groups/"+ids[store.GroupOperator], "", adminCookie)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusNoContent; got != want {
		t.Errorf(
			"delete starter group: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	req = authedRequest(http.MethodDelete, "/api/v1/groups/"+target.ID, "", adminCookie)
	w = serveAuth(api, req)

	if got, want := w.Code, http.StatusNoContent; got != want {
		t.Fatalf(
			"delete known group: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	if _, err := deps.Store.GroupByID(t.Context(), target.ID); !errors.Is(err, store.ErrNotFound) {
		t.Errorf(
			"delete known group: Store.GroupByID() error type mismatch\n"+
				"got:  %v\nwant: %v (deleted)",
			err, store.ErrNotFound,
		)
	}
}

func TestAPI_HTTPGroupDelete__kicks(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	group := createAuthGroup(t, deps, "crew", "", nil)
	member := createAuthUser(t, deps, "member", "member-password", "crew")
	outsider := createAuthUser(t, deps, "outsider", "outsider-password")
	adminID := adminContext(t, api, deps).User.ID
	connect := wireHub(t, api)
	adminClient := connect(adminID, "admin-session")
	memberClient := connect(member.ID, "member-session")
	outsiderClient := connect(outsider.ID, "outsider-session")

	req := authedRequest(http.MethodDelete, "/api/v1/groups/"+group.ID, "", adminCookie)
	w := serveAuth(api, req)
	if got, want := w.Code, http.StatusNoContent; got != want {
		t.Fatalf("status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
	}

	// Membership is captured before the delete's cascade.
	memberKept := api.hub.hasClient(memberClient)
	outsiderKept := api.hub.hasClient(outsiderClient)
	adminKept := api.hub.hasClient(adminClient)
	if memberKept || !outsiderKept || !adminKept {
		t.Errorf(
			"Hub.hasClient() mismatch\n"+
				"got:  member=%t, outsider=%t, admin=%t\n"+
				"want: member=false, outsider=true, admin=true",
			memberKept, outsiderKept, adminKept,
		)
	}
}

func TestAPI_HTTPGroupDelete__membershipLookupFailure(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, dbConn := testAuthServer(t, file)
	authCtx := adminContext(t, api, deps)
	target := createAuthGroup(t, deps, "doomed", "", nil)
	test.DropTable(t, dbConn, "user_groups")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/groups/"+target.ID, nil)
	req = mux.SetURLVars(withAuthCtx(req, authCtx), map[string]string{"id": target.ID})
	w := httptest.NewRecorder()
	api.httpGroupDelete(w, req)

	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
	}
}

func TestAPI_HTTPGroupUpdate__kickLookupFailure(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	authCtx := adminContext(t, api, deps)

	faultDB, state := storetest.FaultDB(t)
	faultStore, err := store.New(t.Context(), faultDB)
	if err != nil {
		t.Fatalf("store.New() failed: %v", err)
	}
	target, err := faultStore.CreateGroup(t.Context(), "kicked", "", nonAdminGrants())
	if err != nil {
		t.Fatalf("Store.CreateGroup() failed: %v", err)
	}
	api.auth.Store = faultStore

	// Fail only the membership listing; the update itself succeeds.
	state.Set(`ORDER BY user_id`)
	t.Cleanup(func() { state.Set("") })
	req := httptest.NewRequest(
		http.MethodPatch, "/api/v1/groups/"+target.ID,
		strings.NewReader(`{"permissions":[]}`),
	)
	req = mux.SetURLVars(withAuthCtx(req, authCtx), map[string]string{"id": target.ID})
	w := httptest.NewRecorder()
	releaseStdout := test.CaptureLog(t, logx.Default())
	api.httpGroupUpdate(w, req)
	stdout := releaseStdout()

	// Failing to kick members is only logged.
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
	}

	wantStdout := `^ERROR: httpGroupUpdate \(.*\), list group's users:\s+injected fault`
	if !util.RegexCheck(wantStdout, stdout) {
		t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, wantStdout)
	}

	group, err := faultStore.GroupByID(t.Context(), target.ID)
	if err != nil {
		t.Fatalf("Store.GroupByID() failed: %v", err)
	}
	if got, want := len(group.Grants), 0; got != want {
		t.Errorf("len(Group.Grants) mismatch\ngot:  %d\nwant: %d", got, want)
	}
}

func TestAPI_HTTPPermissionCatalogue(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)

	adminCookie := loginCookie(t, api, "admin", "admin-password")
	createAuthUser(t, deps, "viewer", "viewer-password", store.GroupViewer)
	viewerCookie := loginCookie(t, api, "viewer", "viewer-password")

	req := authedRequest(http.MethodGet, "/api/v1/permissions", "", adminCookie)
	w := serveAuth(api, req)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"admin access to permissions: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	var catalogue apitype.PermissionCatalogue
	if err := decode.Unmarshal("json", w.Body.Bytes(), &catalogue); err != nil {
		t.Fatalf("admin access to permissions: decode.Unmarshal() failed: %v", err)
	}
	if len(catalogue.Resources) == 0 {
		t.Error(
			"admin access to permissions: len(PermissionCatalogue.Resources) mismatch\n" +
				"got:  0\nwant: > 0",
		)
	}

	req = authedRequest(http.MethodGet, "/api/v1/permissions", "", viewerCookie)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusForbidden; got != want {
		t.Errorf(
			"viewer access to permissions: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
}
