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
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/util"
)

func TestHTTP_HTTPServiceOrderGet(t *testing.T) {
	// GIVEN: an API and a request for the service order.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	apiMu := sync.RWMutex{}

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

			api.Config.Order = tc.order

			// WHEN: that HTTP request is sent.
			req := httptest.NewRequest(http.MethodGet, "/api/v1/service/order", nil)
			w := httptest.NewRecorder()
			apiMu.RLock()
			api.httpServiceOrderGet(w, req)
			apiMu.RUnlock()
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			prefix := fmt.Sprintf("%s\nAPI.httpServiceOrderGet()", packageName)

			// THEN: the expected body is returned.
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf(
					"%s unexpected error:\n%v",
					prefix, err,
				)
			}
			var wantBuilder strings.Builder
			wantBuilder.WriteString(`{"order":[`)
			if len(tc.order) > 0 {
				wantBuilder.WriteString(`"`)
				wantBuilder.WriteString(strings.Join(tc.order, `","`))
				wantBuilder.WriteString(`"`)
			}
			wantBuilder.WriteString("]}\n")
			if gotBody, wantBody := string(data), wantBuilder.String(); gotBody != wantBody {
				t.Errorf(
					"%s body mismatch\ngot:  %q\nwant: %q",
					packageName, gotBody, wantBody,
				)
			}
		})
	}
}

func TestHTTP_HTTPServiceOrderSet(t *testing.T) {
	// GIVEN: an API and a request to set the service order.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	apiMu := sync.RWMutex{}
	t.Cleanup(func() {
		// Give time for save before TempDir clean-up.
		time.Sleep(2 * config.DebounceDuration)
	})

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
			name:           "valid order",
			hadOrder:       testOrder,
			body:           `{"order":["service1"]}`,
			wantStatusCode: http.StatusOK,
			wantOrder:      []string{"service1"},
			bodyRegex:      successMessage,
		},
		{
			name:           "empty order",
			hadOrder:       testOrder,
			body:           `{"order":[]}`,
			wantStatusCode: http.StatusOK,
			wantOrder:      []string{},
			bodyRegex:      successMessage,
		},
		{
			name:           "body with no order",
			hadOrder:       testOrder,
			body:           `{"invalid":"data"}`,
			wantStatusCode: http.StatusOK,
			wantOrder:      []string{},
			bodyRegex:      successMessage,
		},
		{
			name:           "payload too large",
			hadOrder:       testOrder,
			body:           strings.Repeat("a", 1024),
			wantStatusCode: http.StatusBadRequest,
			wantOrder:      nil,
			bodyRegex:      `{"message":"http: request body too large"}`,
		},
		{
			name:           "invalid JSON",
			hadOrder:       testOrder,
			body:           `{"order":["service1","service2","service3"}`,
			wantStatusCode: http.StatusBadRequest,
			wantOrder:      nil,
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
			// t.Parallel() - Cannot run in parallel since we're sharing the API.

			api.Config.Order = tc.hadOrder
			services := make(map[string]*service.Service, len(tc.hadOrder))
			for _, id := range tc.hadOrder {
				services[id] = testService(t, id, "url", "url", true)
			}
			api.Config.Service = services

			// WHEN: that HTTP request is sent.
			req := httptest.NewRequest(http.MethodPost, "/api/v1/service/order", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			apiMu.Lock()
			api.httpServiceOrderSet(w, req)
			apiMu.Unlock()
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			prefix := fmt.Sprintf("%s\nAPI.httpServiceOrderSet()", packageName)

			// THEN: the expected status code is returned.
			if got, want := res.StatusCode, tc.wantStatusCode; got != want {
				t.Errorf(
					"%s status code mismatch\ngot:  %d\nwant: %d",
					prefix, got, want,
				)
			}

			// AND: the expected body is returned.
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf(
					"%s unexpected error:\n%v",
					prefix, err,
				)
			}
			if got := string(data); !util.RegexCheck(tc.bodyRegex, got) {
				t.Errorf(
					"%s body mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.bodyRegex,
				)
			}

			// AND: the service order is updated as expected.
			if tc.wantOrder != nil && !util.AreSlicesEqual(api.Config.Order, tc.wantOrder) {
				t.Errorf(
					"%s Order mismatch\ngot:  %q\nwant: %q",
					prefix, api.Config.Order, tc.wantOrder,
				)
			}
		})
	}
}

func TestHTTP_HTTPServiceSummary(t *testing.T) {
	testSVC := testService(t, "TestHTTP_HTTPServiceSummary", "url", "url", true)
	// GIVEN: an API and a request for detail of a service.
	file := filepath.Join(t.TempDir(), "config.yml")
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
			wantBody:       testSVC.Summary().String(),
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

			// WHEN: that HTTP request is sent.
			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			api.httpServiceSummary(w, req)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			prefix := fmt.Sprintf("%s\nAPI.httpServiceSummary()", packageName)

			// THEN: the expected status code is returned.
			if got, want := res.StatusCode, tc.wantStatusCode; got != want {
				t.Errorf(
					"%s status code mismatch\ngot:  %d\nwant: %d",
					prefix, got, want,
				)
			}

			// AND: the expected body is returned.
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf(
					"%s unexpected error:\n%v",
					prefix, err,
				)
			}
			if got := string(data); !util.RegexCheck(tc.wantBody, got) {
				t.Errorf(
					"%s body mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wantBody,
				)
			}
		})
	}
}
