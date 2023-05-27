// Copyright [2023] [Argus]
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
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/mux"
)

func TestHTTP_httpServiceOrder(t *testing.T) {
	// GIVEN an API and a request for the service order
	file := "TestHTTP_httpServiceOrder.yml"
	api := testAPI(file)
	apiMutex := sync.RWMutex{}
	defer func() {
		os.RemoveAll(file)
		if api.Config.Settings.Data.DatabaseFile != nil {
			os.RemoveAll(*api.Config.Settings.Data.DatabaseFile)
		}
	}()

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

			// WHEN that HTTP request is sent
			req := httptest.NewRequest(http.MethodGet, "/api/v1/service/order", nil)
			w := httptest.NewRecorder()
			apiMutex.RLock()
			api.httpServiceOrder(w, req)
			apiMutex.RUnlock()
			res := w.Result()
			defer res.Body.Close()

			// THEN the expected body is returned as expected
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("unexpected error - %v",
					err)
			}
			got := string(data)
			want := `{"order":[`
			if len(tc.order) > 0 {
				want += fmt.Sprintf(`"%s"`, strings.Join(tc.order, `","`))
			}
			want += "]}\n"
			if got != want {
				t.Errorf("want\n%q\nnot\n%q",
					want, got)
			}
		})
	}
}

func TestHTTP_httpServiceSummary(t *testing.T) {
	testSVC := testService("TestHTTP_httpServiceSummary")
	// GIVEN an API and a request for detail of a service
	file := "TestHTTP_httpServiceSummary.yml"
	api := testAPI(file)
	api.Config.Service[testSVC.ID] = testSVC
	api.Config.Order = append(api.Config.Order, testSVC.ID)
	defer func() {
		os.RemoveAll(file)
		if api.Config.Settings.Data.DatabaseFile != nil {
			os.RemoveAll(*api.Config.Settings.Data.DatabaseFile)
		}
	}()

	tests := map[string]struct {
		serviceName    string
		wantBody       string
		wantStatusCode int
	}{
		"known service": {
			serviceName:    (testSVC.ID),
			wantBody:       testSVC.Summary().String(),
			wantStatusCode: http.StatusOK,
		},
		"unknown service": {
			serviceName:    ("bish-bash-bosh"),
			wantBody:       `\{"message":"service .+ not found"`,
			wantStatusCode: http.StatusNotFound,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			target := "/api/v1/service/summary/"
			target += url.QueryEscape(tc.serviceName)

			// WHEN that HTTP request is sent
			req := httptest.NewRequest(http.MethodGet, target, nil)
			vars := map[string]string{
				"service_name": tc.serviceName}
			req = mux.SetURLVars(req, vars)
			w := httptest.NewRecorder()
			api.httpServiceSummary(w, req)
			res := w.Result()
			defer res.Body.Close()

			// THEN the expected status code is returned
			if res.StatusCode != tc.wantStatusCode {
				t.Errorf("Status code, expected a %d, not a %d",
					tc.wantStatusCode, res.StatusCode)
			}
			// AND the expected body is returned as expected
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("unexpected error - %v",
					err)
			}
			got := string(data)
			re := regexp.MustCompile(tc.wantBody)
			match := re.MatchString(got)
			if !match {
				t.Errorf("want match for %q\nnot: %q",
					tc.wantBody, got)
			}
		})
	}
}
