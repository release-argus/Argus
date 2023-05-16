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
	"bytes"
	"encoding/json"
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
	"time"

	"github.com/gorilla/mux"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/util"
	api_type "github.com/release-argus/Argus/web/api/types"
)

func TestHTTP_Version(t *testing.T) {
	// GIVEN an API and the Version,BuildDate and GoVersion vars defined
	api := API{}
	api.Log = util.NewJLog("WARN", false)
	util.Version = "1.2.3"
	util.BuildDate = "2022-01-01T01:01:01Z"

	// WHEN a HTTP request is made to the httpVersion handler
	req := httptest.NewRequest(http.MethodGet, "/api/v1/version", nil)
	w := httptest.NewRecorder()
	api.httpVersion(w, req)
	res := w.Result()
	defer res.Body.Close()

	// THEN the version is returned in JSON format
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected error to be nil got %v",
			err)
	}
	var got api_type.VersionAPI
	json.Unmarshal(data, &got)
	want := api_type.VersionAPI{
		Version:   util.Version,
		BuildDate: util.BuildDate,
		GoVersion: util.GoVersion,
	}
	if got != want {
		t.Errorf("Version HTTP should have returned %v, not %v",
			want, got)
	}
}

func TestHTTP_BasicAuth(t *testing.T) {
	// GIVEN an API with/without Basic Auth credentials
	tests := map[string]struct {
		basicAuth *config.WebSettingsBasicAuth
		fail      bool
		noHeader  bool
	}{
		"No basic auth": {
			basicAuth: nil,
			fail:      false},
		"basic auth fail invalid creds": {
			basicAuth: &config.WebSettingsBasicAuth{
				Username: "test", Password: "1234"},
			fail: true},
		"basic auth fail no Authorization header": {
			basicAuth: &config.WebSettingsBasicAuth{
				Username: "test", Password: "1234"},
			noHeader: true,
			fail:     true},
		"basic auth pass": {
			basicAuth: &config.WebSettingsBasicAuth{
				Username: "test", Password: "123"},
			fail: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			cfg := config.Config{}
			cfg.Settings.Web.BasicAuth = tc.basicAuth
			cfg.Settings.Web.RoutePrefix = stringPtr("")
			api := NewAPI(&cfg, util.NewJLog("WARN", false))
			api.Router.HandleFunc("/test", func(rw http.ResponseWriter, req *http.Request) {
				return
			})
			ts := httptest.NewServer(api.BaseRouter)
			defer ts.Close()

			// WHEN a HTTP request is made to this router
			client := http.Client{}
			req, err := http.NewRequest("GET", ts.URL+"/test", nil)
			if err != nil {
				t.Fatal(err)
			}
			if !tc.noHeader {
				req.Header = http.Header{
					// test:123
					"Authorization": {"Basic dGVzdDoxMjM="},
				}
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			// THEN the request passes only when expected
			got := resp.StatusCode
			want := 200
			if tc.fail {
				want = http.StatusUnauthorized
			}
			if got != want {
				t.Errorf("Expected a %d, not a %d",
					want, got)
			}
		})
	}
}

func TestHTTP_VersionRefreshUncreated(t *testing.T) {
	// GIVEN an API and a request to refresh the x_version of a service
	file := "TestHTTP_VersionRefreshUncreated.yml"
	api := testAPI(file)
	defer func() {
		os.RemoveAll(file)
		if api.Config.Settings.Data.DatabaseFile != nil {
			os.RemoveAll(*api.Config.Settings.Data.DatabaseFile)
		}
	}()
	tests := map[string]struct {
		deployedVersion bool
		params          map[string]string
		wantBody        string
		wantStatusCode  int
	}{
		"latest version, no vars": {
			params:         map[string]string{},
			wantBody:       `"error":"values failed validity check:`,
			wantStatusCode: http.StatusBadRequest,
		},
		"latest version, valid vars": {
			params: map[string]string{
				"type":         "url",
				"url":          "https://valid.release-argus.io/plain",
				"url_commands": `[{"type":"regex","regex":"stable version: \"v?([0-9.]+)\""}]`},
			wantBody:       `^{"version":"[0-9.]+","timestamp":"[^"]+"}\s$`,
			wantStatusCode: http.StatusOK,
		},
		"latest version, invalid vars": {
			params: map[string]string{
				"type":         "url",
				"url":          "https://valid.release-argus.io/plain",
				"url_commands": `[{"type":"regex"}]`},
			wantBody:       `"error":"url_commands.*regex:.*required`,
			wantStatusCode: http.StatusBadRequest,
		},
		"deployed version, no vars": {
			deployedVersion: true,
			params:          map[string]string{},
			wantBody:        `"error":"values failed validity check:`,
			wantStatusCode:  http.StatusBadRequest,
		},
		"deployed version, valid vars": {
			deployedVersion: true,
			params: map[string]string{
				"url":   "https://valid.release-argus.io/plain",
				"regex": `stable version: "v?([0-9.]+)"`,
			},
			wantBody:       `^\{"version":"[0-9.]+","timestamp":"[^"]+"\}\s$`,
			wantStatusCode: http.StatusOK,
		},
		"deployed version, invalid vars": {
			deployedVersion: true,
			params: map[string]string{
				"url":   "https://valid.release-argus.io/plain",
				"regex": `stable version: "v?([0-9.+)"`,
			},
			wantBody:       `"error":"values failed validity check:.*Invalid RegEx`,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			target := "/api/v1/deployed_version/refresh"
			if !tc.deployedVersion {
				target = "/api/v1/latest_version/refresh"
			}
			// add the params to the URL
			params := url.Values{}
			for k, v := range tc.params {
				params.Set(k, v)
			}
			target += "?" + strings.Replace(params.Encode(), "+", "%20", -1)

			// WHEN that HTTP request is sent
			req := httptest.NewRequest(http.MethodGet, target, nil)
			w := httptest.NewRecorder()
			api.httpVersionRefreshUncreated(w, req)
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

func TestHTTP_VersionRefresh(t *testing.T) {
	testSVC := testService("TestHTTP_VersionRefresh")
	testSVC.LatestVersion.Status.SetLatestVersion("1.0.0", false)
	testSVC.LatestVersion.Query(true, &util.LogFrom{})
	v, _ := testSVC.DeployedVersionLookup.Query(true, &util.LogFrom{})
	testSVC.Status.SetDeployedVersion(v, false)
	// GIVEN an API and a request to refresh the x_version of a service
	file := "TestHTTP_VersionRefresh.yml"
	api := testAPI(file)
	apiMutex := sync.RWMutex{}
	defer func() {
		os.RemoveAll(file)
		if api.Config.Settings.Data.DatabaseFile != nil {
			os.RemoveAll(*api.Config.Settings.Data.DatabaseFile)
		}
	}()

	tests := map[string]struct {
		serviceName         *string
		svc                 *service.Service
		deployedVersion     bool
		nilDeployedVersion  bool
		params              map[string]string
		wantBody            string
		wantStatusCode      int
		wantLatestVersion   string
		wantDeployedVersion string
	}{
		"latest version, no changes": {
			params: map[string]string{},
			wantBody: fmt.Sprintf(`\{"version":%q,.*"\}`,
				testSVC.Status.LatestVersion()),
			wantStatusCode:    http.StatusOK,
			wantLatestVersion: testSVC.Status.LatestVersion(),
		},
		"latest version, different regex doesn't update service version": {
			params: map[string]string{
				"url_commands":        `[{"type":"regex","regex":"beta: \"v?([0-9.]+-beta)\""}]`,
				"semantic_versioning": "false"},
			wantBody:          `\{"version":"[0-9.]+-beta",.*"\}`,
			wantStatusCode:    http.StatusOK,
			wantLatestVersion: "",
		},
		"latest version, invalid vars": {
			params: map[string]string{
				"url_commands":        `[{"type":"regex","regex":"beta: \"v?([0-9.+-beta)\""}]`,
				"semantic_versioning": "false"},
			wantBody:          `{.*"error":".*Invalid RegEx`,
			wantStatusCode:    http.StatusOK,
			wantLatestVersion: "",
		},
		"latest version, unknown service": {
			serviceName: stringPtr("bish-bash-bosh"),
			params: map[string]string{
				"url_commands":        `\[\{"type":"regex","regex":"beta: \"v?([0-9.+-beta)\""\}\]`,
				"semantic_versioning": "false"},
			wantBody:          `\{"message":"service .+ not found"\}`,
			wantStatusCode:    http.StatusNotFound,
			wantLatestVersion: "",
		},
		"adding deployed version to service": {
			deployedVersion:    true,
			nilDeployedVersion: true,
			params: map[string]string{
				"url":                 "https://invalid.release-argus.io/json",
				"json":                "foo.bar.version",
				"allow_invalid_certs": "true"},
			wantBody: fmt.Sprintf(`\{"version":%q`,
				testSVC.Status.DeployedVersion()),
			wantStatusCode:      http.StatusOK,
			wantDeployedVersion: "",
		},
		"deployed version, no changes": {
			deployedVersion: true,
			params:          map[string]string{},
			wantBody: fmt.Sprintf(`\{"version":%q,.*"\}`,
				testSVC.Status.DeployedVersion()),
			wantStatusCode:      http.StatusOK,
			wantDeployedVersion: testSVC.Status.DeployedVersion(),
		},
		"deployed version, different json doesn't update service version": {
			deployedVersion: true,
			params: map[string]string{
				"json":                "version",
				"semantic_versioning": "false"},
			wantBody:          `\{"version":"[0-9.]+",.*"\}`,
			wantStatusCode:    http.StatusOK,
			wantLatestVersion: "",
		},
		"deployed version, invalid vars": {
			deployedVersion: true,
			params: map[string]string{
				"regex": "v?([0-9.+)"},
			wantBody:          `\{.*"error":".*Invalid RegEx`,
			wantStatusCode:    http.StatusOK,
			wantLatestVersion: "",
		},
		"deployed version, unknown service": {
			deployedVersion: true,
			serviceName:     stringPtr("bish-bash-bosh"),
			params: map[string]string{
				"semantic_versioning": "false"},
			wantBody:          `\{"message":"service .+ not found"`,
			wantStatusCode:    http.StatusNotFound,
			wantLatestVersion: "",
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := testService(name)
			apiMutex.Lock()
			api.Config.Service[svc.ID] = svc
			apiMutex.Unlock()
			if tc.nilDeployedVersion {
				svc.DeployedVersionLookup = nil
			}
			target := "/api/v1/deployed_version/refresh/"
			if !tc.deployedVersion {
				target = "/api/v1/latest_version/refresh/"
			}
			target += url.QueryEscape(svc.ID)
			// add the params to the URL
			params := url.Values{}
			for k, v := range tc.params {
				params.Set(k, v)
			}
			target += "?" + strings.Replace(params.Encode(), "+", "%20", -1)

			// WHEN that HTTP request is sent
			req := httptest.NewRequest(http.MethodGet, target, nil)
			// set service_name
			serviceName := svc.ID
			if tc.serviceName != nil {
				serviceName = *tc.serviceName
			}
			vars := map[string]string{
				"service_name": serviceName,
			}
			req = mux.SetURLVars(req, vars)
			w := httptest.NewRecorder()
			apiMutex.Lock()
			api.httpVersionRefresh(w, req)
			apiMutex.Unlock()
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
			// AND the LatestVersion is expected
			if svc.Status.LatestVersion() != tc.wantLatestVersion {
				t.Errorf("LatestVersion, expected %q, not %q",
					tc.wantLatestVersion, svc.Status.LatestVersion())
			}
			// AND the DeployedVersion is expected
			if svc.Status.DeployedVersion() != tc.wantDeployedVersion {
				t.Errorf("DeployedVersion, expected %q, not %q",
					tc.wantDeployedVersion, svc.Status.DeployedVersion())
			}
		})
	}
}

func TestHTTP_EditServiceGetDetail(t *testing.T) {
	testSVC := testService("TestHTTP_EditServiceGetDetail")
	// GIVEN an API and a request for detail of a service
	file := "TestHTTP_EditServiceGetDetail.yml"
	api := testAPI(file)
	apiMutex := sync.RWMutex{}
	defer func() {
		os.RemoveAll(file)
		if api.Config.Settings.Data.DatabaseFile != nil {
			os.RemoveAll(*api.Config.Settings.Data.DatabaseFile)
		}
	}()

	tests := map[string]struct {
		serviceName    *string
		wantBody       string
		wantStatusCode int
	}{
		"known service": {
			wantBody: fmt.Sprintf(
				`\{"comment":%q,.*"latest_version":{.*"url":%q.*,"deployed_version":{.*"url":%q,`,
				testSVC.Comment,
				testSVC.LatestVersion.URL,
				testSVC.DeployedVersionLookup.URL),
			wantStatusCode: http.StatusOK,
		},
		"unknown service": {
			serviceName:    stringPtr("bish-bash-bosh"),
			wantBody:       `\{"message":"service .+ not found"`,
			wantStatusCode: http.StatusNotFound,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := testService(name)
			apiMutex.Lock()
			api.Config.Service[svc.ID] = svc
			apiMutex.Unlock()
			// service_name
			serviceName := svc.ID
			if tc.serviceName != nil {
				serviceName = *tc.serviceName
			}
			target := "/api/v1/service/edit/"
			target += url.QueryEscape(serviceName)

			// WHEN that HTTP request is sent
			req := httptest.NewRequest(http.MethodGet, target, nil)
			vars := map[string]string{
				"service_name": serviceName,
			}
			req = mux.SetURLVars(req, vars)
			w := httptest.NewRecorder()
			apiMutex.RLock()
			api.httpEditServiceGetDetail(w, req)
			apiMutex.RUnlock()
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

func TestHTTP_EditServiceGetOtherDetails(t *testing.T) {
	// GIVEN an API and a request for detail of a service
	tests := map[string]struct {
		wantBody       string
		wantStatusCode int
	}{
		"get details": {
			wantBody: `
"hard_defaults":.*\{
"interval":"10m",
.*
"defaults":\{.*"notify":\{.*"webhook":\{`,
			wantStatusCode: http.StatusOK,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		file := name + ".test.yml"
		api := testAPI(file)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := testService(name)
			defer func() {
				os.RemoveAll(file)
				if api.Config.Settings.Data.DatabaseFile != nil {
					os.RemoveAll(*api.Config.Settings.Data.DatabaseFile)
				}
			}()
			api.Config.Service[svc.ID] = svc
			target := "/api/v1/service/edit/"
			target += url.QueryEscape(svc.ID)

			// WHEN that HTTP request is sent
			req := httptest.NewRequest(http.MethodGet, target, nil)
			w := httptest.NewRecorder()
			api.httpEditServiceGetOtherDetails(w, req)
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
			tc.wantBody = strings.ReplaceAll(tc.wantBody, "\n", "")
			re := regexp.MustCompile(tc.wantBody)
			match := re.MatchString(got)
			if !match {
				t.Errorf("want match for %q\nnot: %q",
					tc.wantBody, got)
			}
		})
	}
}

func TestHTTP_EditServiceEdit(t *testing.T) {
	testSVC := testService("TestHTTP_EditServiceEdit")
	testSVC.LatestVersion.Status.SetLatestVersion("1.0.0", false)
	testSVC.LatestVersion.Query(true, &util.LogFrom{})
	v, _ := testSVC.DeployedVersionLookup.Query(true, &util.LogFrom{})
	testSVC.Status.SetDeployedVersion(v, false)
	// GIVEN an API and a request to create/edit a service
	file := "TestHTTP_EditServiceEdit.yml"
	api := testAPI(file)
	apiMutex := sync.RWMutex{}
	defer func() {
		os.RemoveAll(file)
		if api.Config.Settings.Data.DatabaseFile != nil {
			os.RemoveAll(*api.Config.Settings.Data.DatabaseFile)
		}
	}()
	var svcName string
	for _, svc := range api.Config.Service {
		svcName = svc.ID
		break
	}

	tests := map[string]struct {
		serviceName         *string
		payload             string
		wantBody            string
		wantStatusCode      int
		wantLatestVersion   string
		wantDeployedVersion string
	}{
		"invalid json": {
			payload: `
		"name": "__name__-",
		"latest_version": {
		    "type":"github",
		    "url":"release-argus/Argus"
		`,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `\{"message":"create .* cannot unmarshal.*"\}`,
		},
		"create new service": {
			payload: `{
"name": "create new service-",
"latest_version": {
    "type":"github",
    "url":"release-argus/Argus"
}}`,
			wantStatusCode: http.StatusOK,
			wantBody:       "^$",
		},
		"create new service, but name already taken": {
			payload: `{
		"name": "` + svcName + `",
		"latest_version": {
		  "type":"github",
		  "url":"release-argus/Argus"
		}}`,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `\{"message":"create .* failed.*"\}`,
		},
		"create new service, but invalid interval": {
			payload: `{
		"name": "__name__-",
		"latest_version": {
		    "type":"github",
		    "url":"release-argus/Argus"
		},
		"options": {
			"interval": "foo"
		}}`,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `\{"message":"create .* failed.*options:.*interval:.*invalid.*"\}`,
		},
		"edit service": {
			serviceName: stringPtr("__name__"),
			payload: `{
		"name": "__name__",
		"latest_version": {
		    "type":"url",
		    "url":"https://valid.release-argus.io/plain",
		    "url_commands": [
		        {
		            "type":"regex",
		            "regex":"stable version: \"v?([0-9.]+)\""
		        }
		    ]
		},
		"options": {
		    "interval": "99m"
		}}`,
			wantStatusCode:      http.StatusOK,
			wantBody:            "^$",
			wantLatestVersion:   "[0-9.]+",
			wantDeployedVersion: "",
		},
		"edit service that doesn't exist": {
			serviceName: stringPtr("service that doesn't exist"),
			payload: `{
		"latest_version": {
		    "type":"url",
		    "url":"https://valid.release-argus.io/plain",
		    "url_commands": [
		        {
		            "type":"regex",
		            "regex":"stable version: \"v?([0-9.]+)\""
		        }
		    ]
		},
		"options": {
			"interval": "99m"
		}}`,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `^\{"message":"edit .* failed.*"\}`,
		},
		"edit service that doesn't query successfully": {
			serviceName: stringPtr("__name__"),
			payload: `{
		"name": "__name__",
		"latest_version": {
		    "type":"url",
		    "url":"https://valid.release-argus.io/plain",
		    "url_commands": [
		        {
		            "type":"regex",
		            "regex":"stable version: \"v-([0-9.]+)\""
		        }
		    ]
		},
		"options": {
			"interval": "99m"
		}}`,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `^\{"message":"edit .* failed.*didn't return any matches"\}`,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := testService(name)
			apiMutex.Lock()
			api.Config.Service[svc.ID] = svc
			api.Config.Order = append(api.Config.Order, svc.ID)
			apiMutex.Unlock()
			tc.payload = strings.ReplaceAll(tc.payload, "__name__", name)
			tc.payload = strings.ReplaceAll(tc.payload, "\n", "")
			payload := bytes.NewReader([]byte(tc.payload))
			var req *http.Request
			// CREATE
			target := "/api/v1/service/new"
			req = httptest.NewRequest(http.MethodPost, target, payload)
			// EDIT
			if tc.serviceName != nil {
				// set service_name
				*tc.serviceName = strings.ReplaceAll(
					*tc.serviceName, "__name__", name)
				vars := map[string]string{
					"service_name": url.PathEscape(*tc.serviceName),
				}
				target = "/api/v1/service/edit/" + url.PathEscape(*tc.serviceName)
				req = httptest.NewRequest(http.MethodPut, target, payload)
				req = mux.SetURLVars(req, vars)
			}

			// WHEN that HTTP request is sent
			w := httptest.NewRecorder()
			apiMutex.Lock()
			api.httpEditServiceEdit(w, req)
			apiMutex.Unlock()
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
			if tc.wantStatusCode != http.StatusOK {
				return
			}
			// AND the service was created
			serviceName := util.DefaultIfNil(tc.serviceName)
			// (CREATE)
			if serviceName == "" {
				var data map[string]interface{}
				json.Unmarshal([]byte(tc.payload), &data)
				serviceName = data["name"].(string)
			}
			apiMutex.RLock()
			if tc.serviceName != nil &&
				api.Config.Service[*tc.serviceName] == nil {
				t.Errorf("service %q not created",
					*tc.serviceName)
			}
			svc = api.Config.Service[serviceName]
			apiMutex.RUnlock()
			if svc == nil {
				if tc.wantLatestVersion != tc.wantDeployedVersion &&
					tc.wantLatestVersion != "" {
					t.Errorf("service %q not created",
						serviceName)
				}
				return
			}
			// AND the LatestVersion is expected
			re = regexp.MustCompile(tc.wantLatestVersion)
			match = re.MatchString(svc.Status.LatestVersion())
			if !match {
				t.Errorf("LatestVersion, expected %q, not %q",
					tc.wantLatestVersion, svc.Status.LatestVersion())
			}
			// AND the DeployedVersion is expected
			re = regexp.MustCompile(tc.wantDeployedVersion)
			match = re.MatchString(svc.Status.DeployedVersion())
			if !match {
				t.Errorf("DeployedVersion, expected %q, not %q",
					tc.wantDeployedVersion, svc.Status.DeployedVersion())
			}
		})
	}
}

func TestHTTP_EditServiceDelete(t *testing.T) {
	// GIVEN an API and a request to delete a service
	file := "TestHTTP_EditServiceDelete.yml"
	api := testAPI(file)
	defer func() {
		os.RemoveAll(file)
		if api.Config.Settings.Data.DatabaseFile != nil {
			os.RemoveAll(*api.Config.Settings.Data.DatabaseFile)
		}
	}()
	svc := testService("TestHTTP_EditServiceDelete")
	svc.Init(
		&api.Config.Defaults.Service, &api.Config.HardDefaults.Service,
		&api.Config.Notify, &api.Config.Defaults.Notify, &api.Config.HardDefaults.Notify,
		&api.Config.WebHook, &api.Config.Defaults.WebHook, &api.Config.HardDefaults.WebHook)
	api.Config.AddService("", svc)
	// drain db from the Service addition
	<-*api.Config.DatabaseChannel
	tests := []struct {
		name           string
		service        string
		wantBody       string
		wantStatusCode int
	}{
		{
			name:           "unknown service",
			service:        "foo",
			wantBody:       `{"message":"Delete .* failed, service not found"`,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "delete service",
			service:        svc.ID,
			wantBody:       `^$`,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "delete service again",
			service:        svc.ID,
			wantBody:       `{"message":"Delete .* failed, service not found"`,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		name, tc := tc.name, tc
		t.Run(name, func(t *testing.T) {

			target := "/api/v1/service/delete/"
			target += url.QueryEscape(tc.service)

			// WHEN that HTTP request is sent
			req := httptest.NewRequest(http.MethodGet, target, nil)
			vars := map[string]string{
				"service_name": tc.service,
			}
			req = mux.SetURLVars(req, vars)
			w := httptest.NewRecorder()
			api.httpEditServiceDelete(w, req)
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
			// AND the service is removed from the config
			if api.Config.Service[tc.service] != nil {
				t.Errorf("service %q was not removed from the config",
					tc.service)
			}
			if util.Contains(api.Config.Order, tc.service) {
				t.Errorf("service %q was not removed from the Order",
					tc.service)
			}
			// AND the service is removed from the database (if the req was OK)
			if tc.wantStatusCode == http.StatusOK {
				time.Sleep(time.Second)
				if len(*api.Config.DatabaseChannel) == 0 {
					t.Errorf("service %q was not removed from the database",
						tc.service)
				} else {
					msg := <-*api.Config.DatabaseChannel
					if msg.Delete != true {
						t.Errorf("message to the db should have been a deletion\n%+v",
							msg)
					}
				}
			}
		})
	}
}
