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
	"strings"
	"testing"

	"github.com/release-argus/Argus/auth/store"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// preferencesPath is the route the signed-in user's preferences live on.
const preferencesPath = "/api/v1/auth/me/preferences"

// decodeAuthMe parses an '/auth/me'-shaped response body, failing on one without a user
// (such as an error response).
func decodeAuthMe(t *testing.T, body []byte) apitype.AuthMe {
	t.Helper()

	var me apitype.AuthMe
	if err := decode.Unmarshal("json", body, &me); err != nil {
		t.Fatalf("decode.Unmarshal() failed: %v", err)
	}
	if me.User.ID == "" {
		t.Fatalf("AuthMe.User.ID mismatch\ngot:  %q (body: %q)\nwant: non-empty", me.User.ID, body)
	}
	return me
}

func TestAPI_AuthMePreferences__roundTrip(t *testing.T) {
	api, _, _ := testAuthServer(t, t.Name()+".yml")
	cookie := loginCookie(t, api, "admin", "admin-password")

	req := authedRequest(http.MethodGet, "/api/v1/auth/me", "", cookie)
	w := serveAuth(api, req)
	me := decodeAuthMe(t, w.Body.Bytes())

	if me.Preferences != nil {
		t.Error("before save: AuthMe.Preferences mismatch\ngot:  present\nwant: absent")
	}

	req = authedRequest(
		http.MethodPut, preferencesPath,
		test.TrimJSON(`{
			"hide":       ["inactive", "up_to_date"],
			"view":       "table",
			"timestamps": ["found"]
		}`),
		cookie,
	)
	w = serveAuth(api, req)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"save user preferences: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	me = decodeAuthMe(t, w.Body.Bytes())
	if me.Preferences == nil {
		t.Fatal("save user preferences: AuthMe.Preferences mismatch\ngot:  absent\nwant: present")
	}
	// Returned in catalogue order, rather than as submitted.
	if testErr := test.AssertSlicesEqual(
		t,
		me.Preferences.Hide,
		[]string{store.HideUpToDate, store.HideInactive},
		"AuthMe.Preferences.Hide",
	); testErr != nil {
		t.Errorf("save user preferences: %v", testErr)
	}
	if testErr := test.AssertSlicesEqual(
		t, me.Preferences.Timestamps, []string{store.TimestampFound}, "AuthMe.Preferences.Timestamps",
	); testErr != nil {
		t.Errorf("save user preferences: %v", testErr)
	}
	if got, want := me.Preferences.View, store.ViewTable; got != want {
		t.Errorf(
			"save user preferences: AuthMe.Preferences.View mismatch\ngot:  %q\nwant: %q",
			got, want,
		)
	}

	req = authedRequest(http.MethodGet, "/api/v1/auth/me", "", cookie)
	w = serveAuth(api, req)
	me = decodeAuthMe(t, w.Body.Bytes())
	if me.Preferences == nil {
		t.Error("/auth/me: AuthMe.Preferences mismatch\ngot:  absent\nwant: present")
	} else if got, want := me.Preferences.View, store.ViewTable; got != want {
		t.Errorf("/auth/me: AuthMe.Preferences.View mismatch\ngot:  %q\nwant: %q", got, want)
	}

	req = authedRequest(http.MethodDelete, preferencesPath, "", cookie)
	w = serveAuth(api, req)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"delete user preferences: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	if deleted := decodeAuthMe(t, w.Body.Bytes()); deleted.Preferences != nil {
		t.Error("delete user preferences: AuthMe.Preferences mismatch\ngot:  present\nwant: absent")
	}

	req = authedRequest(http.MethodDelete, preferencesPath, "", cookie)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf(
			"delete user preferences (again): status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
}

func TestAPI_AuthMePreferencesUpdate(t *testing.T) {
	api, _, _ := testAuthServer(t, t.Name()+".yml")
	cookie := loginCookie(t, api, "admin", "admin-password")

	tests := []struct {
		name        string
		body        string
		noCookie    bool
		wantStatus  int
		wantMessage string
	}{
		{
			name: "valid/every field",
			body: test.TrimJSON(`{
				"hide":["skipped"],
				"view":"grid",
				"timestamps":["deployed","queried"]
			}`),
			wantStatus: http.StatusOK,
		},
		{
			name: "valid/explicitly hide nothing",
			body: test.TrimJSON(`{
				"hide":[],
				"view":"grid",
				"timestamps":[]
			}`),
			wantStatus: http.StatusOK,
		},
		{
			name:       "valid/only one field, the rest inherit",
			body:       `{"view":"table"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:        "invalid/unknown hide name",
			body:        `{"hide":["upToDate"]}`,
			wantStatus:  http.StatusBadRequest,
			wantMessage: `hide: "upToDate" <invalid> (`,
		},
		{
			name:        "invalid/unknown view",
			body:        `{"view":"list"}`,
			wantStatus:  http.StatusBadRequest,
			wantMessage: `view: "list" <invalid> (`,
		},
		{
			name:        "invalid/unknown timestamp name",
			body:        `{"timestamps":["last_queried"]}`,
			wantStatus:  http.StatusBadRequest,
			wantMessage: `timestamps: "last_queried" <invalid> (`,
		},
		{
			name:        "invalid/malformed body",
			body:        `{"view":`,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "parse request",
		},
		{
			name:       "invalid/no session",
			body:       `{"view":"grid"}`,
			noCookie:   true,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sessionCookie := cookie
			if tc.noCookie {
				sessionCookie = nil
			}
			req := authedRequest(http.MethodPut, preferencesPath, tc.body, sessionCookie)
			w := serveAuth(api, req)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Fatalf(
					"save user preferences: status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}

			if tc.wantMessage == "" {
				return
			}
			var body map[string]string
			if err := decode.Unmarshal("json", w.Body.Bytes(), &body); err != nil {
				t.Fatalf("save user preferences: decode.Unmarshal() failed: %v", err)
			}
			if msg := body["message"]; !strings.Contains(msg, tc.wantMessage) {
				t.Errorf(
					"save user preferences: body[\"message\"] mismatch\ngot:  %q\nwant: %q",
					msg, tc.wantMessage,
				)
			}
		})
	}
}

func TestAPI_AuthMePreferences__perUser(t *testing.T) {
	api, deps, _ := testAuthServer(t, t.Name()+".yml")
	createAuthUser(t, deps, "other", "other-password")
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	otherCookie := loginCookie(t, api, "other", "other-password")

	req := authedRequest(http.MethodPut, preferencesPath, `{"view":"table"}`, adminCookie)
	w := serveAuth(api, req)
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"save user preferences: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	req = authedRequest(http.MethodGet, "/api/v1/auth/me", "", otherCookie)
	w = serveAuth(api, req)
	me := decodeAuthMe(t, w.Body.Bytes())
	if me.Preferences != nil {
		t.Error("other user: AuthMe.Preferences mismatch\ngot:  present\nwant: absent")
	}
}

func TestAPI_AuthMePreferencesUpdate__bearerRefused(t *testing.T) {
	api, deps, _ := testAuthServer(t, t.Name()+".yml")
	authCtx := adminContext(t, api, deps)
	plaintext, _ := createAPIToken(t, deps, authCtx.User.ID, "ci", nil)

	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			req := authedRequest(method, preferencesPath, `{"view":"table"}`, nil)
			req.Header.Set("Authorization", "Bearer "+plaintext)
			w := serveAuth(api, req)

			// Only a session may change preferences.
			if got, want := w.Code, http.StatusForbidden; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
		})
	}
}

func TestAPI_AuthMePreferences__explicitEmpty(t *testing.T) {
	api, _, _ := testAuthServer(t, t.Name()+".yml")
	cookie := loginCookie(t, api, "admin", "admin-password")

	req := authedRequest(
		http.MethodPut, preferencesPath, `{"hide":[],"view":"grid","timestamps":[]}`, cookie,
	)
	w := serveAuth(api, req)
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"save user preferences: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	for _, tc := range []struct {
		name string
		read func() []byte
	}{
		{
			name: "PUT response",
			read: func() []byte {
				return w.Body.Bytes()
			},
		},
		{
			name: "later /auth/me",
			read: func() []byte {
				req := authedRequest(http.MethodGet, "/api/v1/auth/me", "", cookie)
				return serveAuth(api, req).Body.Bytes()
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			body := tc.read()
			me := decodeAuthMe(t, body)
			if me.Preferences == nil {
				t.Fatal("AuthMe.Preferences mismatch\ngot:  absent\nwant: present")
			}
			// Present and empty, rather than absent (inheriting).
			if testErr := test.AssertSlicesEqual(
				t, me.Preferences.Hide, []string{}, "AuthMe.Preferences.Hide",
			); testErr != nil {
				t.Error(testErr)
			}
			if testErr := test.AssertSlicesEqual(
				t, me.Preferences.Timestamps, []string{}, "AuthMe.Preferences.Timestamps",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestAPI_AuthMePreferences__unauthenticatedAndBrokenStore(t *testing.T) {
	api, deps, dbConn := testAuthServer(t, t.Name()+".yml")
	authCtx := adminContext(t, api, deps)

	tests := []struct {
		name    string
		method  string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{name: "update", method: http.MethodPut, handler: api.httpAuthMePreferencesUpdate},
		{name: "delete", method: http.MethodDelete, handler: api.httpAuthMePreferencesDelete},
	}

	for _, tc := range tests {
		t.Run("no auth context/"+tc.name, func(t *testing.T) {
			// Serial: shares dbConn, which the parent closes after the loop.

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, preferencesPath, strings.NewReader(`{"view":"grid"}`))
			tc.handler(w, req)

			if got, want := w.Code, http.StatusUnauthorized; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
		})
	}

	_ = dbConn.Close()

	for _, tc := range tests {
		t.Run("broken store/"+tc.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, preferencesPath, strings.NewReader(`{"view":"grid"}`))
			req = withAuthCtx(req, authCtx)
			tc.handler(w, req)

			if got, want := w.Code, http.StatusInternalServerError; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
		})
	}
}

func TestAPI_AuthMePreferences__partialSaveInherits(t *testing.T) {
	api, _, _ := testAuthServer(t, t.Name()+".yml")
	cookie := loginCookie(t, api, "admin", "admin-password")

	req := authedRequest(http.MethodPut, preferencesPath, `{"hide":["skipped"]}`, cookie)
	w := serveAuth(api, req)
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"save user preferences: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	me := decodeAuthMe(t, w.Body.Bytes())
	if me.Preferences == nil {
		t.Fatal("save user preferences: AuthMe.Preferences mismatch\ngot:  absent\nwant: present")
	}
	if testErr := test.AssertSlicesEqual(
		t, me.Preferences.Hide, []string{store.HideSkipped}, "AuthMe.Preferences.Hide",
	); testErr != nil {
		t.Errorf("save user preferences: %v", testErr)
	}
	// Layout and timestamps stay absent, so they keep inheriting.
	if testErr := test.AssertSlicesEqual(
		t, me.Preferences.Timestamps, nil, "AuthMe.Preferences.Timestamps",
	); testErr != nil {
		t.Errorf("save user preferences: %v", testErr)
	}
	if got, want := me.Preferences.View, ""; got != want {
		t.Errorf(
			"save user preferences: AuthMe.Preferences.View mismatch\ngot:  %q\nwant: %q",
			got, want,
		)
	}
}

func TestAPI_AuthMe__preferencesReadFails(t *testing.T) {
	api, deps, dbConn := testAuthServer(t, t.Name()+".yml")
	authCtx := adminContext(t, api, deps)
	prefs := store.DashboardPreferences{View: store.ViewTable}
	if _, err := deps.Store.SetPreferences(t.Context(), authCtx.User.ID, prefs); err != nil {
		t.Fatalf("Store.SetPreferences() failed: %v", err)
	}

	test.DropTable(t, dbConn, "user_preferences")

	req := httptest.NewRequest(
		http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"username":"admin","password":"admin-password"}`),
	)
	w := serveAuth(api, req)

	// Login degrades, rather than failing, without the preferences it could not read.
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("login: status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
	}
	if me := decodeAuthMe(t, w.Body.Bytes()); me.Preferences != nil {
		t.Error("login: AuthMe.Preferences mismatch\ngot:  present\nwant: absent")
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req = withAuthCtx(req, authCtx)
	api.httpAuthMe(w, req)

	// 500, rather than reporting preferences the user has not lost.
	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf(
			"read user preferences: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
}
