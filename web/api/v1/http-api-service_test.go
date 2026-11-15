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
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/release-argus/Argus/auth/rbac"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/util"
)

func TestAPI_HTTPServiceOrderGet(t *testing.T) {
	file := t.Name() + ".yml"
	api := testAPI(t, file)

	tests := []struct {
		name  string
		order []string
	}{
		{
			name:  "empty",
			order: []string{},
		},
		{
			name:  "one",
			order: []string{"service1"},
		},
		{
			name:  "multiple/alphabetical ordering",
			order: []string{"service1", "service2", "service3"},
		},
		{
			name:  "multiple/custom ordering",
			order: []string{"service2", "service3", "service1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: shares api.Config.

			api.Config.OrderMu.Lock()
			api.Config.Order = tc.order
			api.Config.OrderMu.Unlock()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/service/order", nil)
			w := httptest.NewRecorder()
			api.httpServiceOrderGet(w, req)
			resp := w.Result()
			t.Cleanup(func() { _ = resp.Body.Close() })

			data, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}
			var wantBuilder strings.Builder
			wantBuilder.WriteString(`{"order":[`)
			if len(tc.order) > 0 {
				wantBuilder.WriteString(`"`)
				wantBuilder.WriteString(strings.Join(tc.order, `","`))
				wantBuilder.WriteString(`"`)
			}
			wantBuilder.WriteString("]}\n")
			// Unfiltered, as auth is disabled.
			if gotBody, wantBody := string(data), wantBuilder.String(); gotBody != wantBody {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", gotBody, wantBody)
			}
		})
	}
}

func TestAPI_HTTPServiceOrderGet__filteredByGrants(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file) // Config holds service "test".

	api.Config.OrderMu.Lock()
	api.Config.Service["middle"] = &service.Service{}
	api.Config.Service["tagged"] = &service.Service{
		Dashboard: dashboard.Options{Tags: []string{"prod"}},
	}
	api.Config.Order = []string{"test", "middle", "tagged"}
	api.Config.OrderMu.Unlock()

	for groupName, grants := range map[string][]rbac.Grant{
		"scoped": {
			{Resource: rbac.ResourceService, Action: rbac.ActionRead, Scope: rbac.Scope{Type: rbac.ScopeServiceTag, Ref: "prod"}},
		},
		"pair": {
			{Resource: rbac.ResourceService, Action: rbac.ActionRead, Scope: rbac.Scope{Type: rbac.ScopeService, Ref: "test"}},
			{Resource: rbac.ResourceService, Action: rbac.ActionRead, Scope: rbac.Scope{Type: rbac.ScopeServiceTag, Ref: "prod"}},
		},
	} {
		createAuthGroup(t, deps, groupName, "", grants)
	}
	createAuthUser(t, deps, "scoped-user", "scoped-password", "scoped")
	createAuthUser(t, deps, "pair-user", "pair-password", "pair")
	createAuthUser(t, deps, "loner", "loner-password")

	tests := []struct {
		name     string
		username string
		password string
		want     []string
	}{
		{
			name:     "global read sees the full order",
			username: "admin", password: "admin-password",
			want: []string{"test", "middle", "tagged"},
		},
		{
			name:     "scoped read sees only permitted services",
			username: "scoped-user", password: "scoped-password",
			want: []string{"tagged"},
		},
		{
			name:     "filtering preserves the relative order",
			username: "pair-user", password: "pair-password",
			want: []string{"test", "tagged"},
		},
		{
			name:     "no grants sees an empty order",
			username: "loner", password: "loner-password",
			want: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cookie := loginCookie(t, api, tc.username, tc.password)

			req := authedRequest(http.MethodGet, "/api/v1/service/order", "", cookie)
			w := serveAuth(api, req)

			if got, want := w.Code, http.StatusOK; got != want {
				t.Fatalf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
			var response ServiceOrderAPI
			if err := decode.Unmarshal("json", w.Body.Bytes(), &response); err != nil {
				t.Fatalf("decode.Unmarshal() failed: %v", err)
			}
			if got := response.Order; !slices.Equal(got, tc.want) {
				t.Errorf("ServiceOrderAPI.Order mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestAPI_HTTPServiceOrderSet(t *testing.T) {
	file := t.Name() + ".yml"
	api := testAPI(t, file)

	testOrder := []string{"service1", "service2", "service3"}
	successMessage := `{"message":"order updated"}` + "\n"
	tests := []struct {
		name                string
		body                string
		wantStatusCode      int
		hadOrder, wantOrder []string
		bodyRegex           string
	}{
		{
			name:           "valid reorder",
			hadOrder:       testOrder,
			body:           `{"order":["service3","service2","service1"]}`,
			wantStatusCode: http.StatusOK,
			wantOrder:      []string{"service3", "service2", "service1"},
			bodyRegex:      successMessage,
		},
		{
			name:           "partial submission reorders only the services submitted",
			hadOrder:       testOrder,
			body:           `{"order":["service3","service1"]}`,
			wantStatusCode: http.StatusOK,
			wantOrder:      []string{"service3", "service2", "service1"},
			bodyRegex:      successMessage,
		},
		{
			name:           "a lone submitted service keeps its place",
			hadOrder:       testOrder,
			body:           `{"order":["service2"]}`,
			wantStatusCode: http.StatusOK,
			wantOrder:      testOrder,
			bodyRegex:      successMessage,
		},
		{
			name:           "duplicates pruned",
			hadOrder:       testOrder,
			body:           `{"order":["service2", "service2", "service1", "service3"]}`,
			wantStatusCode: http.StatusOK,
			wantOrder:      []string{"service2", "service1", "service3"},
			bodyRegex:      successMessage,
		},
		{
			name:           "empty submission keeps all services",
			hadOrder:       testOrder,
			body:           `{"order":[]}`,
			wantStatusCode: http.StatusOK,
			wantOrder:      testOrder,
			bodyRegex:      successMessage,
		},
		{
			name:           "body with no order keeps all services",
			hadOrder:       testOrder,
			body:           `{"invalid":"data"}`,
			wantStatusCode: http.StatusOK,
			wantOrder:      testOrder,
			bodyRegex:      successMessage,
		},
		{
			name:           "payload too large",
			hadOrder:       testOrder,
			body:           strings.Repeat("a", 1024),
			wantStatusCode: http.StatusBadRequest,
			wantOrder:      testOrder,
			bodyRegex:      `{"message":"http: request body too large"}`,
		},
		{
			name:           "invalid JSON",
			hadOrder:       testOrder,
			body:           `{"order":["service1","service2","service3"}`,
			wantStatusCode: http.StatusBadRequest,
			wantOrder:      testOrder,
			bodyRegex:      `{"message":"invalid JSON:\\n  jsontext: invalid character '}' after array element`,
		},
		{
			name:           "trim unknown services",
			hadOrder:       testOrder,
			body:           `{"order":["service1","service2","service3","service4"]}`,
			wantStatusCode: http.StatusOK,
			wantOrder:      testOrder,
			bodyRegex:      successMessage,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: shares api.Config.

			services := make(map[string]*service.Service, len(tc.hadOrder))
			for _, id := range tc.hadOrder {
				services[id] = testService(t, id, "url", "url", true)
			}
			api.Config.OrderMu.Lock()
			api.Config.Order = tc.hadOrder
			api.Config.Service = services
			api.Config.OrderMu.Unlock()

			req := httptest.NewRequest(http.MethodPost, "/api/v1/service/order", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			api.httpServiceOrderSet(w, req)
			resp := w.Result()
			t.Cleanup(func() { _ = resp.Body.Close() })

			if got, want := resp.StatusCode, tc.wantStatusCode; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}

			data, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}
			if got := string(data); !util.RegexCheck(tc.bodyRegex, got) {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, tc.bodyRegex)
			}

			// Submitted services are re-seated into the positions they already held.
			if !util.AreSlicesEqual(api.Config.Order, tc.wantOrder) {
				t.Errorf(
					"Config.Order mismatch\ngot:  %q\nwant: %q",
					api.Config.Order, tc.wantOrder,
				)
			}
		})
	}
}

func TestAPI_HTTPServiceSummary(t *testing.T) {
	testSVC := testService(t, t.Name(), "url", "url", true)

	file := t.Name() + ".yml"
	api := testAPI(t, file)
	api.Config.Service[testSVC.ID] = testSVC
	api.Config.Order = append(api.Config.Order, testSVC.ID)

	tests := []struct {
		name           string
		serviceID      string
		wantBody       string
		wantStatusCode int
	}{
		{
			name:           "known service",
			serviceID:      testSVC.ID,
			wantBody:       "^" + regexp.QuoteMeta(testSVC.Summary().String()) + "\n$",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "unknown service",
			serviceID:      "bish-bash-bosh",
			wantBody:       `{"message":"service .+ not found"`,
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "no service_id provided",
			serviceID:      "",
			wantBody:       `{"message":"missing required query parameter: service_id"}`,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			target := "/api/v1/service/summary"
			params := url.Values{}
			params.Set("service_id", tc.serviceID)

			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			api.httpServiceSummary(w, req)
			resp := w.Result()
			t.Cleanup(func() { _ = resp.Body.Close() })

			if got, want := resp.StatusCode, tc.wantStatusCode; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}

			data, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}
			if got := string(data); !util.RegexCheck(tc.wantBody, got) {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, tc.wantBody)
			}
		})
	}
}
