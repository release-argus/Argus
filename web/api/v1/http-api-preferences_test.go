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

// decodeAuthMe parses an '/auth/me'-shaped response body.
func decodeAuthMe(t *testing.T, prefix string, body []byte) apitype.AuthMe {
	t.Helper()

	var me apitype.AuthMe
	if err := decode.Unmarshal("json", body, &me); err != nil {
		t.Fatalf(
			"%s\nparse response: %v",
			prefix, err,
		)
	}
	return me
}

func TestAPI_AuthMePreferences__roundTrip(t *testing.T) {
	// GIVEN: an auth-enabled API and a signed-in admin.
	file := "TestAPI_AuthMePreferences__roundTrip.yml"
	api, _, _ := testAuthServer(t, file)
	cookie := loginCookie(t, api, "admin", "admin-password")

	prefix := fmt.Sprintf("%s\nhttpAuthMePreferences()", packageName)

	// WHEN: /auth/me is read before anything is saved.
	me := decodeAuthMe(t, prefix,
		serveAuth(api, authedRequest(
			http.MethodGet, "/api/v1/auth/me",
			"",
			cookie,
		)).Body.Bytes(),
	)

	// THEN: it carries no preferences.
	if me.Preferences != nil {
		t.Errorf(
			"%s\nunsaved preferences present\ngot: %+v",
			prefix, *me.Preferences,
		)
	}

	// WHEN: preferences are saved.
	w := serveAuth(api, authedRequest(
		http.MethodPut, preferencesPath,
		test.TrimJSON(`{
			"hide":       ["inactive", "up_to_date"],
			"view":       "table",
			"timestamps": ["found"]
		}`),
		cookie,
	))

	// THEN: the response carries them back, in catalogue order.
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"%s\nstatus mismatch\ngot:  %d - %s\nwant: %d",
			prefix,
			got, w.Body.String(),
			want,
		)
	}
	me = decodeAuthMe(t, prefix, w.Body.Bytes())
	if me.Preferences == nil {
		t.Fatalf(
			"%s\nsaved preferences missing from the response\ngot: %s",
			prefix, w.Body.String(),
		)
	}
	if testErr := test.AssertSlicesEqual(t,
		me.Preferences.Hide, []string{store.HideUpToDate, store.HideInactive},
		prefix, "hide",
	); testErr != nil {
		t.Error(testErr)
	}
	if testErr := test.AssertSlicesEqual(t,
		me.Preferences.Timestamps, []string{store.TimestampFound},
		prefix, "timestamps",
	); testErr != nil {
		t.Error(testErr)
	}
	if got, want := me.Preferences.View, store.ViewTable; got != want {
		t.Errorf(
			"%s\nview mismatch\ngot:  %q\nwant: %q",
			prefix, got, want,
		)
	}

	// AND: /auth/me carries the same preferences.
	me = decodeAuthMe(t, prefix,
		serveAuth(api, authedRequest(
			http.MethodGet, "/api/v1/auth/me",
			"",
			cookie,
		)).Body.Bytes(),
	)
	if me.Preferences == nil || me.Preferences.View != store.ViewTable {
		t.Errorf(
			"%s\n/auth/me lost the saved preferences\ngot: %+v",
			prefix, me.Preferences,
		)
	}

	// WHEN: they are discarded.
	w = serveAuth(api, authedRequest(http.MethodDelete, preferencesPath, "", cookie))

	// THEN: the response, and /auth/me, carry none.
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"%s\ndelete status mismatch\ngot:  %d - %s\nwant: %d",
			prefix,
			got, w.Body.String(),
			want,
		)
	}
	if me := decodeAuthMe(t, prefix, w.Body.Bytes()); me.Preferences != nil {
		t.Errorf(
			"%s\npreferences survived the delete\ngot: %+v",
			prefix, *me.Preferences,
		)
	}

	// AND: discarding again is not an error.
	if w := serveAuth(api, authedRequest(
		http.MethodDelete, preferencesPath,
		"",
		cookie,
	)); w.Code != http.StatusOK {
		t.Errorf(
			"%s\nsecond delete status mismatch\ngot:  %d - %s\nwant: %d",
			prefix,
			w.Code, w.Body.String(),
			http.StatusOK,
		)
	}
}

func TestAPI_AuthMePreferencesUpdate(t *testing.T) {
	// GIVEN: an auth-enabled API and a signed-in admin.
	file := "TestAPI_AuthMePreferencesUpdate.yml"
	api, _, _ := testAuthServer(t, file)
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
			prefix := fmt.Sprintf("%s\nhttpAuthMePreferencesUpdate(%s)", packageName, tc.name)

			// WHEN: the preferences are PUT.
			sessionCookie := cookie
			if tc.noCookie {
				sessionCookie = nil
			}
			w := serveAuth(api, authedRequest(
				http.MethodPut, preferencesPath,
				tc.body,
				sessionCookie,
			))

			// THEN: the status is as expected.
			if got := w.Code; got != tc.wantStatus {
				t.Fatalf(
					"%s\nstatus mismatch\ngot:  %d - %s\nwant: %d",
					prefix, got, w.Body.String(), tc.wantStatus,
				)
			}

			// AND: a rejection says which value it objected to.
			if tc.wantMessage == "" {
				return
			}
			var body map[string]string
			if err := decode.Unmarshal("json", w.Body.Bytes(), &body); err != nil {
				t.Fatalf(
					"%s\nparse error response: %v",
					prefix, err,
				)
			}
			if !strings.Contains(body["message"], tc.wantMessage) {
				t.Errorf(
					"%s\nmessage mismatch\ngot:  %q\nwant it to contain: %q",
					prefix, body["message"], tc.wantMessage,
				)
			}
		})
	}
}

func TestAPI_AuthMePreferences__perUser(t *testing.T) {
	// GIVEN: an auth-enabled API with two signed-in users.
	file := "TestAPI_AuthMePreferences__perUser.yml"
	api, deps, _ := testAuthServer(t, file)
	createAuthUser(t, deps, "other", "other-password")
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	otherCookie := loginCookie(t, api, "other", "other-password")

	prefix := fmt.Sprintf("%s\nhttpAuthMePreferences() per user", packageName)

	// WHEN: one of them saves preferences.
	if w := serveAuth(api, authedRequest(
		http.MethodPut, preferencesPath,
		`{"view":"table"}`,
		adminCookie,
	)); w.Code != http.StatusOK {
		t.Fatalf(
			"%s\nsave failed\ngot: %d - %s",
			prefix, w.Code, w.Body.String(),
		)
	}

	// THEN: the other user still has none.
	me := decodeAuthMe(t, prefix,
		serveAuth(api, authedRequest(
			http.MethodGet, "/api/v1/auth/me",
			"",
			otherCookie,
		)).Body.Bytes(),
	)
	if me.Preferences != nil {
		t.Errorf(
			"%s\nanother user's preferences leaked\ngot: %+v",
			prefix, *me.Preferences,
		)
	}
}

func TestAPI_AuthMePreferencesUpdate__bearerRefused(t *testing.T) {
	// GIVEN: an auth-enabled API and an admin owning an API token.
	file := "TestAPI_AuthMePreferencesUpdate__bearerRefused.yml"
	api, deps, _ := testAuthServer(t, file)
	authCtx := adminContext(t, api, deps)
	plaintext, _, err := deps.Store.CreateAPIToken(
		t.Context(), authCtx.User.ID, "ci", nil,
	)
	if err != nil {
		t.Fatalf(
			"%s\nsetup CreateAPIToken failed: %v",
			packageName, err,
		)
	}

	prefix := fmt.Sprintf("%s\nhttpAuthMePreferences() from a token", packageName)

	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		t.Run("method="+method, func(t *testing.T) {
			// WHEN: the token tries to change the owner's preferences.
			req := authedRequest(method, preferencesPath, `{"view":"table"}`, nil)
			req.Header.Set("Authorization", "Bearer "+plaintext)
			w := serveAuth(api, req)

			// THEN: the request is refused.
			if got, want := w.Code, http.StatusForbidden; got != want {
				t.Errorf(
					"%s\nstatus mismatch\ngot:  %d - %s\nwant: %d",
					prefix, got, w.Body.String(), want,
				)
			}
		})
	}
}

func TestAPI_AuthMePreferences__explicitEmpty(t *testing.T) {
	// GIVEN: an auth-enabled API and a signed-in admin.
	file := "TestAPI_AuthMePreferences__explicitEmpty.yml"
	api, _, _ := testAuthServer(t, file)
	cookie := loginCookie(t, api, "admin", "admin-password")

	prefix := fmt.Sprintf("%s\nhttpAuthMePreferences() explicit none", packageName)

	// WHEN: empty lists are saved.
	w := serveAuth(api, authedRequest(http.MethodPut, preferencesPath,
		`{"hide":[],"view":"grid","timestamps":[]}`, cookie))
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"%s\nstatus mismatch\ngot:  %d - %s\nwant: %d",
			prefix, got, w.Body.String(), want,
		)
	}

	// THEN: they come back present and empty.
	for _, body := range []struct {
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
				return serveAuth(api,
					authedRequest(http.MethodGet, "/api/v1/auth/me", "", cookie)).Body.Bytes()
			},
		},
	} {
		me := decodeAuthMe(t, prefix, body.read())
		if me.Preferences == nil {
			t.Fatalf(
				"%s\n%s dropped the saved preferences",
				prefix, body.name,
			)
		}
		if testErr := test.AssertSlicesEqual(
			t, me.Preferences.Hide, []string{}, prefix, body.name+" hide",
		); testErr != nil {
			t.Error(testErr)
		}
		if testErr := test.AssertSlicesEqual(
			t, me.Preferences.Timestamps, []string{}, prefix, body.name+" timestamps",
		); testErr != nil {
			t.Error(testErr)
		}
	}
}

func TestAPI_AuthMePreferences__unauthenticatedAndBrokenStore(t *testing.T) {
	// GIVEN: an auth-enabled API, an admin, and a handle on its database.
	file := "TestAPI_AuthMePreferences__unauthenticatedAndBrokenStore.yml"
	api, deps, dbConn := testAuthServer(t, file)
	authCtx := adminContext(t, api, deps)

	prefix := fmt.Sprintf("%s\nhttpAuthMePreferences() failure paths", packageName)

	tests := []struct {
		name    string
		method  string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{name: "update", method: http.MethodPut, handler: api.httpAuthMePreferencesUpdate},
		{name: "delete", method: http.MethodDelete, handler: api.httpAuthMePreferencesDelete},
	}

	// WHEN: each handler is called with no authenticated user.
	for _, tc := range tests {
		t.Run("no auth context/"+tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			tc.handler(recorder,
				httptest.NewRequest(tc.method, preferencesPath, strings.NewReader(`{"view":"grid"}`)),
			)

			// THEN: it answers 401.
			if got, want := recorder.Code, http.StatusUnauthorized; got != want {
				t.Errorf(
					"%s\nstatus mismatch\ngot:  %d - %s\nwant: %d",
					prefix, got, recorder.Body.String(), want,
				)
			}
		})
	}

	// AND: the store breaks beneath them.
	_ = dbConn.Close()

	for _, tc := range tests {
		t.Run("broken store/"+tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			tc.handler(recorder, withAuthCtx(
				httptest.NewRequest(tc.method, preferencesPath, strings.NewReader(`{"view":"grid"}`)),
				authCtx,
			))

			// THEN: the failure reads as 500.
			if got, want := recorder.Code, http.StatusInternalServerError; got != want {
				t.Errorf(
					"%s\nstatus mismatch\ngot:  %d - %s\nwant: %d",
					prefix, got, recorder.Body.String(), want,
				)
			}
		})
	}
}

func TestAPI_AuthMePreferences__partialSaveInherits(t *testing.T) {
	// GIVEN: an auth-enabled API and a signed-in admin.
	file := "TestAPI_AuthMePreferences__partialSaveInherits.yml"
	api, _, _ := testAuthServer(t, file)
	cookie := loginCookie(t, api, "admin", "admin-password")

	// WHEN: only the filters are saved.
	body := `{"hide":["skipped"]}`
	w := serveAuth(api, authedRequest(
		http.MethodPut, preferencesPath,
		body,
		cookie,
	))
	prefix := fmt.Sprintf(
		"%s\nhttpAuthMePreferences(%s) partial save",
		packageName, body,
	)
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"%s\nstatus mismatch\ngot:  %d - %s\nwant: %d",
			prefix, got, w.Body.String(), want,
		)
	}

	// THEN: the layout and timestamps stay absent, so they keep inheriting.
	me := decodeAuthMe(t, prefix, w.Body.Bytes())
	if me.Preferences == nil {
		t.Fatalf(
			"%s\nsaved preferences missing\ngot: %s",
			prefix, w.Body.String(),
		)
	}
	if testErr := test.AssertSlicesEqual(
		t, me.Preferences.Hide, []string{store.HideSkipped}, prefix, "hide",
	); testErr != nil {
		t.Error(testErr)
	}
	if testErr := test.AssertSlicesEqual(
		t, me.Preferences.Timestamps, nil, prefix, "timestamps",
	); testErr != nil {
		t.Error(testErr)
	}
	if me.Preferences.View != "" {
		t.Errorf(
			"%s\nview should be absent\ngot: %q",
			prefix, me.Preferences.View,
		)
	}
}

func TestAPI_AuthMe__preferencesReadFails(t *testing.T) {
	// GIVEN: an auth-enabled API and an admin with saved preferences.
	file := "TestAPI_AuthMe__preferencesReadFails.yml"
	api, deps, dbConn := testAuthServer(t, file)
	authCtx := adminContext(t, api, deps)
	prefs := store.DashboardPreferences{View: store.ViewTable}
	if _, err := deps.Store.SetPreferences(t.Context(), authCtx.User.ID, prefs); err != nil {
		t.Fatalf(
			"%s\nsetup SetPreferences(%s) failed: %v",
			packageName, prefs, err,
		)
	}

	prefix := fmt.Sprintf("%s\npreferences read failure", packageName)

	// AND: reading them now fails.
	if _, err := dbConn.ExecContext(t.Context(), `DROP TABLE user_preferences;`); err != nil {
		t.Fatalf(
			"%s drop table failed: %v",
			prefix, err,
		)
	}

	// WHEN: the caller signs in.
	body := `{"username":"admin","password":"admin-password"}`
	w := serveAuth(api, httptest.NewRequest(
		http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(body),
	))

	// THEN: sign-in still succeeds, without the preferences it could not read.
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"%s\nlogin should degrade, not fail\ngot:  %d - %s\nwant: %d",
			prefix, got, w.Body.String(), want,
		)
	}
	if me := decodeAuthMe(t, prefix, w.Body.Bytes()); me.Preferences != nil {
		t.Errorf(
			"%s\nlogin carried preferences it could not read\ngot: %+v",
			prefix, *me.Preferences,
		)
	}

	// WHEN: /auth/me is read instead.
	recorder := httptest.NewRecorder()
	api.httpAuthMe(recorder, withAuthCtx(
		httptest.NewRequest(
			http.MethodGet, "/api/v1/auth/me",
			nil,
		),
		authCtx,
	))

	// THEN: it fails rather than reporting preferences the user has not lost.
	if got, want := recorder.Code, http.StatusInternalServerError; got != want {
		t.Errorf(
			"%s\nstatus mismatch\ngot:  %d - %s\nwant: %d",
			prefix, got, recorder.Body.String(), want,
		)
	}
}
