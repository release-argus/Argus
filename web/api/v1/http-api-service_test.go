// Copyright [2025] [Argus]
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
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/mux"

	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestHTTP_httpServiceOrderGet(t *testing.T) {
	// GIVEN an API and a request for the service order.
	file := "TestHTTP_httpServiceOrderGet.yml"
	api := testAPI(file)
	apiMutex := sync.RWMutex{}
	t.Cleanup(func() {
		os.RemoveAll(file)
		if api.Config.Settings.Data.DatabaseFile != "" {
			os.RemoveAll(api.Config.Settings.Data.DatabaseFile)
		}
	})

	tests := map[string]struct {
		order []string
	}{
		"empty": {
			order: []string{},
		},
		"one": {
			order: []string{"service1"},
		},
		"multiple": {
			order: []string{"service1", "service2", "service3"},
		},
		"multiple - other order": {
			order: []string{"service2", "service3", "service1"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			api.Config.Order = tc.order

			// WHEN that HTTP request is sent.
			req := httptest.NewRequest(http.MethodGet, "/api/v1/service/order", nil)
			w := httptest.NewRecorder()
			apiMutex.RLock()
			api.httpServiceOrderGet(w, req)
			apiMutex.RUnlock()
			res := w.Result()
			t.Cleanup(func() { res.Body.Close() })

			// THEN the expected body is returned as expected.
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("%s\nunexpected error - %v",
					packageName, err)
			}
			got := string(data)
			want := `{"order":[`
			if len(tc.order) > 0 {
				want += fmt.Sprintf(`"%s"`, strings.Join(tc.order, `","`))
			}
			want += "]}\n"
			if got != want {
				t.Errorf("%s\nwant %q\ngot:  %q",
					packageName, want, got)
			}
		})
	}
}

func TestHTTP_httpServiceSummary(t *testing.T) {
	testSVC := testService("TestHTTP_httpServiceSummary", true)
	// GIVEN an API and a request for detail of a service.
	file := "TestHTTP_httpServiceSummary.yml"
	api := testAPI(file)
	api.Config.Service[testSVC.ID] = testSVC
	api.Config.Order = append(api.Config.Order, testSVC.ID)
	t.Cleanup(func() {
		os.RemoveAll(file)
		if api.Config.Settings.Data.DatabaseFile != "" {
			os.RemoveAll(api.Config.Settings.Data.DatabaseFile)
		}
	})

	tests := map[string]struct {
		serviceID      string
		wantBody       string
		wantStatusCode int
	}{
		"known service": {
			serviceID:      testSVC.ID,
			wantBody:       testSVC.Summary().String(),
			wantStatusCode: http.StatusOK,
		},
		"unknown service": {
			serviceID:      "bish-bash-bosh",
			wantBody:       `\{"message":"service .+ not found"`,
			wantStatusCode: http.StatusNotFound,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			target := "/api/v1/service/summary/"
			target += url.QueryEscape(tc.serviceID)

			// WHEN that HTTP request is sent.
			req := httptest.NewRequest(http.MethodGet, target, nil)
			vars := map[string]string{
				"service_id": tc.serviceID}
			req = mux.SetURLVars(req, vars)
			w := httptest.NewRecorder()
			api.httpServiceSummary(w, req)
			res := w.Result()
			t.Cleanup(func() { res.Body.Close() })

			// THEN the expected status code is returned.
			if res.StatusCode != tc.wantStatusCode {
				t.Errorf("%s\nstatus code mismatch\nwant: %d\ngot:  %d",
					packageName, tc.wantStatusCode, res.StatusCode)
			}
			// AND the expected body is returned as expected.
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("%s\nunexpected error - %v",
					packageName, err)
			}
			got := string(data)
			if !util.RegexCheck(tc.wantBody, got) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wantBody, got)
			}
		})
	}
}

func TestHTTP_httpServiceOrderSet(t *testing.T) {
	// GIVEN an API and a request to set the service order.
	file := "TestHTTP_httpServiceOrderSet.yml"
	api := testAPI(file)
	apiMutex := sync.RWMutex{}
	t.Cleanup(func() {
		os.RemoveAll(file)
		if api.Config.Settings.Data.DatabaseFile != "" {
			os.RemoveAll(api.Config.Settings.Data.DatabaseFile)
		}
	})

	testOrder := []string{"service1", "service2", "service3"}
	successMessage := `{"message":"order updated"}` + "\n"
	tests := map[string]struct {
		body           string
		wantStatusCode int
		hadOrder       []string
		wantOrder      []string
		wantBody       string
	}{
		"valid order": {
			hadOrder: testOrder,
			body: test.TrimJSON(`{
				"order":["service1"]
			}`),
			wantStatusCode: http.StatusOK,
			wantOrder:      []string{"service1"},
			wantBody:       successMessage,
		},
		"empty order": {
			hadOrder:       testOrder,
			body:           `{"order":[]}`,
			wantStatusCode: http.StatusOK,
			wantOrder:      []string{},
			wantBody:       successMessage,
		},
		"body with no order": {
			hadOrder:       testOrder,
			body:           `{"invalid":"data"}`,
			wantStatusCode: http.StatusOK,
			wantOrder:      []string{},
			wantBody:       successMessage,
		},
		"payload too large": {
			hadOrder:       testOrder,
			body:           strings.Repeat("a", 1024),
			wantStatusCode: http.StatusBadRequest,
			wantOrder:      nil,
			wantBody:       `{"message":"http: request body too large"}`,
		},
		"invalid JSON": {
			hadOrder:       testOrder,
			body:           `{"order":["service1","service2","service3"}`,
			wantStatusCode: http.StatusBadRequest,
			wantOrder:      nil,
			wantBody:       `{"message":"Invalid JSON - invalid character '}' after array element"}`,
		},
		"trim unknown services": {
			hadOrder:       testOrder,
			body:           `{"order":["service1","service2","service3","service4"]}`,
			wantStatusCode: http.StatusOK,
			wantOrder:      testOrder,
			wantBody:       successMessage,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() -- Cannot run in parallel since we're sharing the API.

			api.Config.Order = tc.hadOrder
			service := make(map[string]*service.Service, len(tc.hadOrder))
			for _, svc := range tc.hadOrder {
				service[svc] = testService(svc, true)
			}
			api.Config.Service = service

			// WHEN that HTTP request is sent.
			req := httptest.NewRequest(http.MethodPost, "/api/v1/service/order", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			apiMutex.Lock()
			api.httpServiceOrderSet(w, req)
			apiMutex.Unlock()
			res := w.Result()
			t.Cleanup(func() { res.Body.Close() })

			// THEN the expected status code is returned.
			if res.StatusCode != tc.wantStatusCode {
				t.Errorf("%s\nstatus code mismatch\nwant: %d\ngot:  %d",
					packageName, tc.wantStatusCode, res.StatusCode)
			}
			// AND the expected body is returned as expected.
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("%s\nunexpected error - %v",
					packageName, err)
			}
			got := string(data)
			if got != tc.wantBody {
				t.Errorf("%s\nbody mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wantBody, got)
			}
			// AND the service order is updated as expected.
			if tc.wantOrder != nil && !test.EqualSlices(api.Config.Order, tc.wantOrder) {
				t.Errorf("%s\nordering mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wantOrder, api.Config.Order)
			}
		})
	}
}
