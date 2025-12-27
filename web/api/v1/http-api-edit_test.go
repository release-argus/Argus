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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrr_test "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service"
	dv_web "github.com/release-argus/Argus/service/deployed_version/types/web"
	lv_web "github.com/release-argus/Argus/service/latest_version/types/web"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

func TestHTTP_LatestVersionRefreshUncreated(t *testing.T) {
	type wants struct {
		bodyRegex  string
		statusCode int
	}

	// GIVEN an API and a request to refresh the latest_version of a service.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	t.Cleanup(func() {
		if api.Config.Settings.Data.DatabaseFile != "" {
			_ = os.RemoveAll(api.Config.Settings.Data.DatabaseFile)
		}
	})
	tests := map[string]struct {
		params map[string]string
		wants  wants
	}{
		"no overrides": {
			params: map[string]string{},
			wants: wants{
				bodyRegex:  `"message":"overrides: .*required`,
				statusCode: http.StatusBadRequest},
		},
		"invalid JSON": {
			params: map[string]string{
				"overrides": `"type": "url", "url": "` + test.LookupPlain["url_valid"] + `"}`,
			},
			wants: wants{
				bodyRegex:  `^\{"message":"invalid JSON[^"]+"\}$`,
				statusCode: http.StatusBadRequest},
		},
		"invalid vars - New fail": {
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"type": "something"
				}`),
			},
			wants: wants{
				bodyRegex:  `"message":".*type: .*invalid`,
				statusCode: http.StatusBadRequest},
		},
		"invalid vars - CheckValues fail": {
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"type":         "url",
					"url":          "` + test.LookupPlain["url_valid"] + `",
					"url_commands": "[{\"type\": \"regex\"}]"
				}`),
			},
			wants: wants{
				bodyRegex:  `"message":"url_commands.*regex:.*required`,
				statusCode: http.StatusBadRequest},
		},
		"valid vars": {
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"type":         "url",
					"url":          "` + test.LookupPlain["url_valid"] + `",
					"url_commands": "[{\"type\": \"regex\", \"regex\": \"stable version: \\\"v?([0-9.]+)\\\"\"}]"
				}`),
			},
			wants: wants{
				bodyRegex:  `^{"version":"[0-9.]+","timestamp":"[^"]+"}\s$`,
				statusCode: http.StatusOK},
		},
		"query fail": {
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"type":         "url",
					"url":          "` + test.LookupPlain["url_invalid"] + `",
					"url_commands": "[{\"type\": \"regex\", \"regex\": \"stable version: \\\"v?([0-9.]+)\\\"\"}]"
				}`),
			},
			wants: wants{
				bodyRegex:  `"message":"x509 `,
				statusCode: http.StatusBadRequest},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			target := "/api/v1/latest_version/refresh_uncreated"
			params := url.Values{}
			for k, v := range tc.params {
				params.Set(k, v)
			}

			// WHEN that HTTP request is sent.
			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			api.httpLatestVersionRefreshUncreated(w, req)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			// THEN the expected status code is returned.
			if res.StatusCode != tc.wants.statusCode {
				t.Errorf("%s\nstatus code mismatch\nwant: %d\ngot:  %d",
					packageName, tc.wants.statusCode, res.StatusCode)
			}
			// AND the expected body is returned.
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("%s\nunexpected error - %v",
					packageName, err)
			}
			got := string(data)
			if !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf("%s\nbody mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wants.bodyRegex, got)
			}
		})
	}
}

func TestHTTP_DeployedVersionRefreshUncreated(t *testing.T) {
	type wants struct {
		bodyRegex  string
		statusCode int
	}

	// GIVEN an API and a request to refresh the deployed_version of a service.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	t.Cleanup(func() {
		if api.Config.Settings.Data.DatabaseFile != "" {
			_ = os.RemoveAll(api.Config.Settings.Data.DatabaseFile)
		}
	})
	tests := map[string]struct {
		params map[string]string
		wants  wants
	}{
		"no vars": {
			params: map[string]string{},
			wants: wants{
				bodyRegex:  `"message":"overrides: .*required`,
				statusCode: http.StatusBadRequest},
		},
		"invalid JSON": {
			params: map[string]string{
				"overrides": `"type": "url", "url": "` + test.LookupPlain["url_valid"] + `"}`,
			},
			wants: wants{
				bodyRegex:  `^\{"message":"invalid JSON[^"]+"\}$`,
				statusCode: http.StatusBadRequest},
		},
		"invalid vars - New fail": {
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"type": "something"
				}`),
			},
			wants: wants{
				bodyRegex:  `"message":".*type: .*invalid`,
				statusCode: http.StatusBadRequest},
		},
		"invalid vars - CheckValues fail": {
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"type":  "url",
					"url":   "` + test.LookupPlain["url_valid"] + `",
					"regex": "stable version: \"v?([0-9.+)\""
				}`),
			},
			wants: wants{
				bodyRegex:  `\{"message":"regex: .*invalid`,
				statusCode: http.StatusBadRequest},
		},
		"valid vars": {
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"type":  "url",
					"url":   "` + test.LookupPlain["url_valid"] + `",
					"regex": "stable version: \"v?([0-9.]+)\""
				}`)},
			wants: wants{
				bodyRegex:  `^\{"version":"[0-9.]+","timestamp":"[^"]+"\}\s$`,
				statusCode: http.StatusOK},
		},
		"query fail": {
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"type":  "url",
					"url":   "` + test.LookupPlain["url_invalid"] + `",
					"regex": "stable version: \"v?([0-9.]+)\""
				}`),
			},
			wants: wants{
				bodyRegex:  `"message":"x509 `,
				statusCode: http.StatusBadRequest},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			target := "/api/v1/deployed_version/refresh_uncreated"
			params := url.Values{}
			for k, v := range tc.params {
				params.Set(k, v)
			}

			// WHEN that HTTP request is sent.
			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			api.httpDeployedVersionRefreshUncreated(w, req)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			// THEN the expected status code is returned.
			if res.StatusCode != tc.wants.statusCode {
				t.Errorf("%s\nstatus code mismatch\nwant: %d\ngot:  %d",
					packageName, tc.wants.statusCode, res.StatusCode)
			}
			// AND the expected body is returned.
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("%s\nunexpected error - %v",
					packageName, err)
			}
			got := string(data)
			if !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf("%s\nbody mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wants.bodyRegex, got)
			}
		})
	}
}

func TestHTTP_LatestVersionRefresh(t *testing.T) {
	testSVC := testService("TestHTTP_LatestVersionRefresh", false)
	testSVC.LatestVersion.GetStatus().SetLatestVersion("1.0.0", "", false)
	_, _ = testSVC.LatestVersion.Query(true, logutil.LogFrom{})
	_ = testSVC.DeployedVersionLookup.Query(true, logutil.LogFrom{})
	type wants struct {
		bodyRegex                      string
		statusCode                     int
		latestVersion, deployedVersion string
		announce                       bool
	}

	// GIVEN an API and a request to refresh the latest_version of a service.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	apiMutex := sync.RWMutex{}
	t.Cleanup(func() {
		if api.Config.Settings.Data.DatabaseFile != "" {
			_ = os.RemoveAll(api.Config.Settings.Data.DatabaseFile)
		}
	})

	tests := map[string]struct {
		serviceID *string
		svc       *service.Service
		params    map[string]string
		wants     wants
	}{
		"no changes": {
			params: map[string]string{},
			wants: wants{
				bodyRegex: fmt.Sprintf(`\{"version":%q,.*"\}`,
					testSVC.Status.LatestVersion()),
				statusCode:      http.StatusOK,
				latestVersion:   testSVC.Status.LatestVersion(),
				deployedVersion: testSVC.Status.LatestVersion(),
				announce:        true},
		},
		"semantic_versioning not sent - refreshes service": {
			wants: wants{
				bodyRegex:       `\{"version":"ver[0-9.]+",.*"\}`,
				statusCode:      http.StatusOK,
				latestVersion:   testSVC.Status.LatestVersion(),
				deployedVersion: testSVC.Status.LatestVersion(),
				announce:        true},
		},
		"semantic_versioning=null - fail as default=true": {
			params: map[string]string{
				"semantic_versioning": "null"},
			wants: wants{
				bodyRegex:     `failed to convert \\"[^"]+\\" to a semantic version`,
				statusCode:    http.StatusBadRequest,
				latestVersion: ""},
		},
		"semantic_versioning=same - refreshes service": {
			params: map[string]string{
				"semantic_versioning": "false"},
			wants: wants{
				bodyRegex:       `\{"version":"ver[0-9.]+",.*"\}`,
				statusCode:      http.StatusOK,
				latestVersion:   testSVC.Status.LatestVersion(),
				deployedVersion: testSVC.Status.LatestVersion(),
				announce:        true},
		},
		"semantic_versioning=diff - not applied to service": {
			params: map[string]string{
				"semantic_versioning": "true"},
			wants: wants{
				bodyRegex:     `failed to convert \\"[^"]+\\" to a semantic version`,
				statusCode:    http.StatusBadRequest,
				latestVersion: ""},
		},
		"different regex doesn't update service version": {
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"url_commands": [{"type":"regex","regex":"beta: \"v?([0-9.]+-beta)\""}]
				}`)},
			wants: wants{
				bodyRegex:     `\{"version":"[0-9.]+-beta",.*"\}`,
				statusCode:    http.StatusOK,
				latestVersion: ""},
		},
		"invalid vars": {
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"url_commands": [{"type":"regex","regex":"beta: \\\"v?([0-9.+-beta)\\\""}]
				}`),
				"semantic_versioning": "false"},
			wants: wants{
				bodyRegex:     `{"message":".*regex: .*invalid`,
				statusCode:    http.StatusBadRequest,
				latestVersion: ""},
		},
		"unknown service": {
			serviceID: test.StringPtr("bish-bash-bosh"),
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"url_commands": [{"type":"regex","regex":"beta: \\\"v?([0-9.+-beta)\\\""}]
				}`),
				"semantic_versioning": "false"},
			wants: wants{
				bodyRegex:     `\{"message":"service .+ not found"\}`,
				statusCode:    http.StatusNotFound,
				latestVersion: ""},
		},
		"no service_id provided": {
			serviceID: test.StringPtr(""),
			wants: wants{
				bodyRegex:     `\{"message":"missing required query parameter: service_id"\}`,
				statusCode:    http.StatusBadRequest,
				latestVersion: ""},
		},
		"use secretRefs": {
			svc: test.IgnoreError(t, func() (*service.Service, error) {
				base := testService("TestHTTP_LatestVersionRefresh", false)
				if lv, ok := base.LatestVersion.(*lv_web.Lookup); ok {
					lv.URL = test.LookupWithHeaderAuth["url_valid"]
					lv.Headers = shared.Headers{
						{Key: test.LookupWithHeaderAuth["header_key"], Value: test.LookupWithHeaderAuth["header_value_pass"]}}
				}
				return base, nil
			}),
			params: map[string]string{
				"overrides": test.TrimJSON(`{
						"headers": [
							{
								"key": "` + test.LookupWithHeaderAuth["header_key"] + `",
								"value": "` + util.SecretValue + `",
								"old_index": 0
							}
						]
					}`,
				),
			},
			wants: wants{
				bodyRegex:  `\{"version":"ver[0-9.]+",.*"\}`,
				statusCode: http.StatusOK,
				announce:   false},
		},
		"invalid secretRefs": {
			svc: test.IgnoreError(t, func() (*service.Service, error) {
				base := testService("TestHTTP_LatestVersionRefresh", false)
				if lv, ok := base.LatestVersion.(*lv_web.Lookup); ok {
					lv.URL = test.LookupWithHeaderAuth["url_valid"]
					lv.Headers = shared.Headers{
						{Key: test.LookupWithHeaderAuth["header_key"], Value: test.LookupWithHeaderAuth["header_value_pass"]}}
				}
				return base, nil
			}),
			params: map[string]string{
				"overrides": test.TrimJSON(`{
						"headers": [
							{
								"key": "` + test.LookupWithHeaderAuth["header_key"] + `",
								"value": "` + util.SecretValue + `",
								"old_index": [0]
							}
						]
					}`,
				),
			},
			wants: wants{
				bodyRegex:  `cannot unmarshal array into Go struct field`,
				statusCode: http.StatusBadRequest,
				announce:   false},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := testService(name, false)
			apiMutex.Lock()
			api.Config.Service[svc.ID] = svc
			apiMutex.Unlock()
			target := "/api/v1/latest_version/refresh"
			// Add the params to the URL.
			params := url.Values{}
			for k, v := range tc.params {
				params.Set(k, v)
			}
			// Set service_id.
			serviceID := util.DereferenceOrValue(tc.serviceID, svc.ID)
			params.Set("service_id", serviceID)

			// WHEN that HTTP request is sent.
			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			apiMutex.Lock()
			api.httpLatestVersionRefresh(w, req)
			apiMutex.Unlock()
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			// THEN the expected status code is returned.
			if res.StatusCode != tc.wants.statusCode {
				t.Errorf("%s\nstatus code mismatch\nwant: %d\ngot:  %d",
					packageName, tc.wants.statusCode, res.StatusCode)
			}
			// AND the expected body is returned.
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("%s\nunexpected error - %v",
					packageName, err)
			}
			got := string(data)
			if !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf("%s\nbody mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wants.bodyRegex, got)
			}
			// AND the LatestVersion is expected.
			if svc.Status.LatestVersion() != tc.wants.latestVersion {
				t.Errorf("%s\nVersionRefresh - LatestVersion\nwant: %q\ngot:  %q",
					packageName, tc.wants.latestVersion, svc.Status.LatestVersion())
			}
			// AND the DeployedVersion is expected.
			if svc.Status.DeployedVersion() != tc.wants.deployedVersion {
				t.Errorf("%s\nVersionRefresh - DeployedVersion\nwant: %q\ngot:  %q",
					packageName, tc.wants.deployedVersion, svc.Status.DeployedVersion())
			}
			// AND the expected announcement is made.
			wantAnnounces := 0
			if tc.wants.announce {
				wantAnnounces = 1
			}
			if got := len(svc.Status.AnnounceChannel); got != wantAnnounces {
				t.Errorf("%s\nDeployedVersionRefresh - Announcements length mismatch\nwant: %d\ngot:  %d",
					packageName, wantAnnounces, got)
			}
		})
	}
}

func TestHTTP_DeployedVersionRefresh(t *testing.T) {
	testSVC := testService("TestHTTP_DeployedVersionRefresh", false)
	testSVC.LatestVersion.GetStatus().SetLatestVersion("1.0.0", "", false)
	_, _ = testSVC.LatestVersion.Query(true, logutil.LogFrom{})
	_ = testSVC.DeployedVersionLookup.Query(true, logutil.LogFrom{})
	type wants struct {
		bodyRegex       string
		statusCode      int
		deployedVersion string
	}

	// GIVEN an API and a request to refresh the deployed_version of a service.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	apiMutex := sync.RWMutex{}
	t.Cleanup(func() {
		if api.Config.Settings.Data.DatabaseFile != "" {
			_ = os.RemoveAll(api.Config.Settings.Data.DatabaseFile)
		}
	})

	tests := map[string]struct {
		serviceID          *string
		svc                *service.Service
		nilDeployedVersion bool
		params             map[string]string
		wants              wants
	}{
		"adding deployed version to service - success": {
			nilDeployedVersion: true,
			params: map[string]string{
				"overrides": test.TrimJSON(`{
				"type":                "url",
				"url":                 "` + test.LookupJSON["url_invalid"] + `",
				"json":                "nonSemVer",
				"allow_invalid_certs": true
			}`)},
			wants: wants{
				bodyRegex: fmt.Sprintf(`\{"version":%q`,
					testSVC.Status.DeployedVersion()),
				statusCode:      http.StatusOK,
				deployedVersion: ""},
		},
		"adding deployed version to service - 'type' missing": {
			nilDeployedVersion: true,
			params: map[string]string{
				"overrides": test.TrimJSON(`{
			"url":                 "` + test.LookupJSON["url_invalid"] + `",
			"json":                "nonSemVer",
			"allow_invalid_certs": true
			}`)},
			wants: wants{
				bodyRegex:  `\{"message":"missing required parameter: overrides.type"`,
				statusCode: http.StatusBadRequest},
		},
		"adding deployed version to service - no overrides": {
			nilDeployedVersion: true,
			wants: wants{
				bodyRegex:  `\{"message":"missing required parameter: overrides.type"`,
				statusCode: http.StatusBadRequest},
		},
		"no changes": {
			params: map[string]string{},
			wants: wants{
				bodyRegex: fmt.Sprintf(`\{"version":%q,.*"\}`,
					testSVC.Status.DeployedVersion()),
				statusCode:      http.StatusOK,
				deployedVersion: testSVC.Status.DeployedVersion()},
		},
		"semantic_versioning not sent - refreshes service": {
			wants: wants{
				bodyRegex:       `\{"version":"ver[\d.]+",.*"\}`,
				statusCode:      http.StatusOK,
				deployedVersion: testSVC.Status.DeployedVersion()},
		},
		"semantic_versioning=null - fail as default=true": {
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"url_commands": [{"type":"regex","regex":"beta: \"v?([0-9.]+-beta\")"}]
				}`),
				"semantic_versioning": "null"},
			wants: wants{
				bodyRegex:       `failed to convert \\"[^"]+\\" to a semantic version`,
				statusCode:      http.StatusBadRequest,
				deployedVersion: ""},
		},
		"semantic_versioning=same - refreshes service": {
			params: map[string]string{
				"semantic_versioning": "false",
			},
			wants: wants{
				bodyRegex:       `\{"version":"ver[0-9.]+",.*"\}`,
				statusCode:      http.StatusOK,
				deployedVersion: testSVC.Status.DeployedVersion()},
		},
		"semantic_versioning=diff - not applied to service": {
			params: map[string]string{
				"semantic_versioning": "true",
			},
			wants: wants{
				bodyRegex:       `failed to convert \\"[^"]+\\" to a semantic version`,
				statusCode:      http.StatusBadRequest,
				deployedVersion: ""},
		},
		"different JSON doesn't update service version": {
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"json": "version"
				}`),
				"semantic_versioning": "false"},
			wants: wants{
				bodyRegex:  `\{"version":"[0-9.]+",.*"\}`,
				statusCode: http.StatusOK},
		},
		"invalid JSON - existing DVL": {
			params: map[string]string{
				"overrides": `{"json": "x.y"}`,
			},
			wants: wants{
				bodyRegex:  `^\{"message":"failed to find value for \\"x\.y\\" in .*"\}$`,
				statusCode: http.StatusBadRequest},
		},
		"invalid JSON - new DVL": {
			nilDeployedVersion: true,
			params: map[string]string{
				"overrides": `{
					"type": "url",
					"method": "GET",
					"url": "` + test.LookupJSON["url_valid"] + `",
					"json": "x.y"
				}`,
			},
			wants: wants{
				bodyRegex:  `^\{"message":"failed to find value for \\"x\.y\\" in .*"\}$`,
				statusCode: http.StatusBadRequest},
		},
		"invalid vars - CheckValues fail": {
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"regex": "v?([0-9.+)"
				}`)},
			wants: wants{
				bodyRegex:  `\{"message":".*regex: .*invalid`,
				statusCode: http.StatusBadRequest},
		},
		"unknown service": {
			serviceID: test.StringPtr("bish-bash-bosh"),
			params: map[string]string{
				"semantic_versioning": "false"},
			wants: wants{
				bodyRegex:  `\{"message":"service .+ not found"`,
				statusCode: http.StatusNotFound},
		},
		"no service_id provided": {
			serviceID: test.StringPtr(""),
			wants: wants{
				bodyRegex:       `\{"message":"missing required query parameter: service_id"\}`,
				statusCode:      http.StatusBadRequest,
				deployedVersion: ""},
		},
		"use secretRefs": {
			svc: test.IgnoreError(t, func() (*service.Service, error) {
				base := testService("TestHTTP_LatestVersionRefresh", false)
				if dv, ok := base.DeployedVersionLookup.(*dv_web.Lookup); ok {
					dv.URL = test.LookupWithHeaderAuth["url_valid"]
					dv.Headers = shared.Headers{
						{Key: test.LookupWithHeaderAuth["header_key"], Value: test.LookupWithHeaderAuth["header_value_pass"]}}
				}
				return base, nil
			}),
			params: map[string]string{
				"overrides": test.TrimJSON(`{
						"headers": [
							{
								"key": "` + test.LookupWithHeaderAuth["header_key"] + `",
								"value": "` + util.SecretValue + `",
								"old_index": 0
							}
						]
					}`,
				),
			},
			wants: wants{
				bodyRegex:  `\{"version":"ver[0-9.]+",.*"\}`,
				statusCode: http.StatusOK},
		},
		"invalid secretRefs": {
			svc: test.IgnoreError(t, func() (*service.Service, error) {
				base := testService("TestHTTP_LatestVersionRefresh", false)
				if dv, ok := base.DeployedVersionLookup.(*dv_web.Lookup); ok {
					dv.URL = test.LookupWithHeaderAuth["url_valid"]
					dv.Headers = shared.Headers{
						{Key: test.LookupWithHeaderAuth["header_key"], Value: test.LookupWithHeaderAuth["header_value_pass"]}}
				}
				return base, nil
			}),
			params: map[string]string{
				"overrides": test.TrimJSON(`{
						"headers": [
							{
								"key": "` + test.LookupWithHeaderAuth["header_key"] + `",
								"value": "` + util.SecretValue + `",
								"old_index": [0]
							}
						]
					}`,
				),
			},
			wants: wants{
				bodyRegex:  `cannot unmarshal array into Go struct field`,
				statusCode: http.StatusBadRequest},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := testService(name, false)
			apiMutex.Lock()
			api.Config.Service[svc.ID] = svc
			apiMutex.Unlock()
			if tc.nilDeployedVersion {
				svc.DeployedVersionLookup = nil
			}
			initialLatestVersion := svc.Status.LatestVersion()
			target := "/api/v1/deployed_version/refresh"
			// add the params to the URL.
			params := url.Values{}
			for k, v := range tc.params {
				params.Set(k, v)
			}
			// Set service_id.
			serviceID := util.DereferenceOrValue(tc.serviceID, svc.ID)
			params.Set("service_id", serviceID)

			// WHEN that HTTP request is sent.
			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			apiMutex.Lock()
			api.httpDeployedVersionRefresh(w, req)
			apiMutex.Unlock()
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			// THEN the expected status code is returned.
			if res.StatusCode != tc.wants.statusCode {
				t.Errorf("%s\nstatus code mismatch\nwant: %d\ngot:  %d",
					packageName, tc.wants.statusCode, res.StatusCode)
			}
			// AND the expected body is returned.
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("%s\nunexpected error - %v",
					packageName, err)
			}
			got := string(data)
			if !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf("%s\nbody mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wants.bodyRegex, got)
			}
			// AND the LatestVersion is unchanged.
			if gotLatestVersion := svc.Status.LatestVersion(); gotLatestVersion != initialLatestVersion {
				t.Errorf("%s\nVersionRefresh - LatestVersion\nwant: %q\ngot:  %q",
					packageName, initialLatestVersion, svc.Status.LatestVersion())
			}
			// AND the DeployedVersion is expected.
			if svc.Status.DeployedVersion() != tc.wants.deployedVersion {
				t.Errorf("%s\nVersionRefresh - DeployedVersion\nwant: %q\ngot:  %q",
					packageName, tc.wants.deployedVersion, svc.Status.DeployedVersion())
			}
		})
	}
}

func TestHTTP_ServiceDetail(t *testing.T) {
	type wants struct {
		bodyRegex  string
		statusCode int
	}

	testSVC := testService("TestHTTP_ServiceDetail", true)
	// GIVEN an API and a request for detail of a service.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	apiMutex := sync.RWMutex{}
	t.Cleanup(func() {
		if api.Config.Settings.Data.DatabaseFile != "" {
			_ = os.RemoveAll(api.Config.Settings.Data.DatabaseFile)
		}
	})

	tests := map[string]struct {
		serviceID *string
		wants     wants
	}{
		"known service": {
			wants: wants{
				bodyRegex: fmt.Sprintf(`\{"comment":%q,.*"latest_version":{.*"url":%q.*,"deployed_version":{.*"url":%q,`,
					testSVC.Comment,
					testSVC.LatestVersion.(*lv_web.Lookup).URL,
					testSVC.DeployedVersionLookup.(*dv_web.Lookup).URL),
				statusCode: http.StatusOK},
		},
		"unknown service": {
			serviceID: test.StringPtr("bish-bash-bosh"),
			wants: wants{
				bodyRegex:  `\{"message":"service .+ not found"`,
				statusCode: http.StatusNotFound},
		},
		"no service_id provided": {
			serviceID: test.StringPtr(""),
			wants: wants{
				bodyRegex:  `\{"message":"missing required query parameter: service_id"\}`,
				statusCode: http.StatusBadRequest},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := testService(name, true)
			apiMutex.Lock()
			api.Config.Service[svc.ID] = svc
			apiMutex.Unlock()
			target := "/api/v1/service/config"
			params := url.Values{}
			// Set service_id.
			serviceID := util.DereferenceOrValue(tc.serviceID, svc.ID)
			params.Set("service_id", serviceID)

			// WHEN that HTTP request is sent.
			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			apiMutex.RLock()
			api.httpServiceDetail(w, req)
			apiMutex.RUnlock()
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			// THEN the expected status code is returned.
			if res.StatusCode != tc.wants.statusCode {
				t.Errorf("%s\nstatus code mismatch\nwant: %d\ngot:  %d",
					packageName, tc.wants.statusCode, res.StatusCode)
			}
			// AND the expected body is returned.
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("%s\nunexpected error - %v",
					packageName, err)
			}
			got := string(data)
			if !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf("%s\nbody mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wants.bodyRegex, got)
			}
		})
	}
}

func TestHTTP_OtherServiceDetails(t *testing.T) {
	// GIVEN an API and a request for detail of a service.
	tests := map[string]struct {
		wantBody       string
		wantStatusCode int
	}{
		"get details": {
			wantBody: `
				"hard_defaults": .*\{
				"interval": "10m",
				.*
				"defaults": \{.*"notify": \{.*"webhook": \{`,
			wantStatusCode: http.StatusOK,
		},
	}

	for name, tc := range tests {
		path := filepath.Join(t.TempDir(), name+".yaml")
		api := testAPI(t, path)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.wantBody = test.TrimJSON(tc.wantBody)
			svc := testService(name, true)
			t.Cleanup(func() {
				if api.Config.Settings.Data.DatabaseFile != "" {
					_ = os.RemoveAll(api.Config.Settings.Data.DatabaseFile)
				}
			})
			api.Config.Service[svc.ID] = svc
			target := "/api/v1/service/defaults"

			// WHEN that HTTP request is sent.
			req := httptest.NewRequest(http.MethodGet, target, nil)
			w := httptest.NewRecorder()
			api.httpOtherServiceDetails(w, req)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			// THEN the expected status code is returned.
			if res.StatusCode != tc.wantStatusCode {
				t.Errorf("%s\nstatus code mismatch\nwant: %d\ngot:  %d",
					packageName, tc.wantStatusCode, res.StatusCode)
			}
			// AND the expected body is returned.
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("%s\nunexpected error - %v",
					packageName, err)
			}
			got := string(data)
			tc.wantBody = strings.ReplaceAll(tc.wantBody, "\n", "")
			if !util.RegexCheck(tc.wantBody, got) {
				t.Errorf("%s\nbody mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wantBody, got)
			}
		})
	}
}

func TestHTTP_TemplateParse(t *testing.T) {
	type wants struct {
		bodyRegex  string
		statusCode int
	}

	// GIVEN an API and a request to parse a template.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	apiMutex := sync.RWMutex{}
	t.Cleanup(func() {
		if api.Config.Settings.Data.DatabaseFile != "" {
			_ = os.RemoveAll(api.Config.Settings.Data.DatabaseFile)
		}
	})

	testSVC := testService("TestHTTP_TemplateParse", true)
	apiMutex.Lock()
	api.Config.Service[testSVC.ID] = testSVC
	apiMutex.Unlock()

	tests := map[string]struct {
		queryParams map[string]string
		wants       wants
	}{
		"missing required parameters": {
			queryParams: map[string]string{},
			wants: wants{
				bodyRegex:  `{"message":"missing required query parameter: service_id"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		"missing required parameters - service_id": {
			queryParams: map[string]string{
				"template": "{{ service_name }}"},
			wants: wants{
				bodyRegex:  `{"message":"missing required query parameter: service_id"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		"missing required parameters - template": {
			queryParams: map[string]string{
				"service_id": "release-argus/Argus"},
			wants: wants{
				bodyRegex:  `{"message":"missing required query parameter: template"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		"invalid template": {
			queryParams: map[string]string{
				"service_id": testSVC.ID,
				"template":   "{{.InvalidField}",
			},
			wants: wants{
				bodyRegex:  `{"message":"failed to parse template"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		"invalid params JSON": {
			queryParams: map[string]string{
				"service_id": testSVC.ID,
				"template":   "{{ service_name }}",
				"params":     `{"invalid":}`,
			},
			wants: wants{
				bodyRegex:  `{"message":"Invalid 'params' query parameter format - invalid character '}' looking for beginning of value"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		"valid template with default parameters": {
			queryParams: map[string]string{
				"service_id": testSVC.ID,
				"template":   `{{service_name }} - {{ version }}`,
			},
			wants: wants{
				bodyRegex: fmt.Sprintf(`{"parsed":"%s - %s"}`,
					testSVC.Name, testSVC.Status.LatestVersion()),
				statusCode: http.StatusOK,
			},
		},
		"valid template with overridden parameters": {
			queryParams: map[string]string{
				"service_id": testSVC.ID,
				"template":   `{{ service_id}} - {{ version }}`,
				"params": test.TrimJSON(`{
					"id": "OverriddenName",
					"latest_version": "2.0.0"
				}`),
			},
			wants: wants{
				bodyRegex:  `{"parsed":"OverriddenName - 2.0.0"}`,
				statusCode: http.StatusOK,
			},
		},
		"unknown service": {
			queryParams: map[string]string{
				"service_id": "unknown_service",
				"template":   "{{ service_name }}",
			},
			wants: wants{
				bodyRegex:  `{"parsed":""}`,
				statusCode: http.StatusOK,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			target := "/api/v1/template/parse"
			params := url.Values{}
			for k, v := range tc.queryParams {
				params.Set(k, v)
			}

			// WHEN that HTTP request is sent.
			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			api.httpTemplateParse(w, req)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			// THEN the expected status code is returned.
			if res.StatusCode != tc.wants.statusCode {
				t.Errorf("%s\nstatus code mismatch\nwant: %d\ngot:  %d",
					packageName, tc.wants.statusCode, res.StatusCode)
			}
			// AND the expected body is returned.
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("%s\nunexpected error - %v",
					packageName, err)
			}
			got := string(data)
			if !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf("%s\nbody mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wants.bodyRegex, got)
			}
		})
	}
}

func TestHTTP_ServiceEdit(t *testing.T) {
	testSVC := testService("TestHTTP_ServiceEdit", true)
	testSVC.LatestVersion.GetStatus().SetLatestVersion("1.0.0", "", false)
	_, _ = testSVC.LatestVersion.Query(true, logutil.LogFrom{})
	_ = testSVC.DeployedVersionLookup.Query(true, logutil.LogFrom{})
	type wants struct {
		bodyRegex                      string
		statusCode                     int
		latestVersion, deployedVersion string
	}

	// GIVEN an API and a request to create/edit a service.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	apiMutex := sync.RWMutex{}
	t.Cleanup(func() {
		if api.Config.Settings.Data.DatabaseFile != "" {
			_ = os.RemoveAll(api.Config.Settings.Data.DatabaseFile)
		}
	})
	var svcName string
	for _, svc := range api.Config.Service {
		svcName = svc.ID
		break
	}

	// Give time for save before TempDir clean-up.
	t.Cleanup(func() { time.Sleep(2 * config.DebounceDuration) })

	tests := map[string]struct {
		serviceID *string
		payload   string
		wants     wants
	}{
		"invalid JSON": {
			payload: `
				"id": "__name__-",
				"latest_version": {
					"type": "github",
					"url": "release-argus/Argus"
				`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `\{"message":"create .* cannot unmarshal.*"\}`},
		},
		"create new service": {
			payload: `
				{
					"id": "create new service-",
					"latest_version": {
						"type": "github",
						"url": "release-argus/Argus"}
				}`,
			wants: wants{
				statusCode: http.StatusOK,
				bodyRegex:  `\{"message":"created service[^}]+"\}`},
		},
		"create new service, but ID already taken": {
			payload: `
				{
					"id": "` + svcName + `",
					"latest_version": {
						"type": "github",
						"url": "release-argus/Argus"}
				}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `\{"message":"create .* failed.*"\}`},
		},
		"create new service, but name already taken": {
			payload: `
				{
					"id": "__name__",
					"name": "` + svcName + `",
					"latest_version": {
						"type": "github",
						"url": "release-argus/Argus"}
				}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `\{"message":"create .* failed.*"\}`},
		},
		"create new service, but invalid interval": {
			payload: `
				{
					"id": "__name__-",
					"latest_version": {
						"type": "github",
						"url": "release-argus/Argus"},
					"options": {
						"interval": "foo"}
				}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `\{"message":"create .* failed.*options:.*interval:.*invalid.*"\}`},
		},
		"edit service": {
			serviceID: test.StringPtr("__name__"),
			payload: `
				{
					"id": "__name__",
					"latest_version": {
						"type": "url",
						"url":  "` + test.LookupPlain["url_valid"] + `",
						"url_commands": [
							{
								"type": "regex",
								"regex": "stable version: \"v?([0-9.]+)\""}]},
					"options": {
						"interval": "99m"}
				}`,
			wants: wants{
				statusCode:      http.StatusOK,
				bodyRegex:       `\{"message":"edited service[^}]+"\}`,
				latestVersion:   "[0-9.]+",
				deployedVersion: ""},
		},
		"edit service that doesn't exist": {
			serviceID: test.StringPtr("service that doesn't exist"),
			payload: `
				{
					"latest_version": {
						"type": "url",
						"url":  "` + test.LookupPlain["url_valid"] + `",
						"url_commands": [
							{
								"type": "regex",
								"regex": "stable version: \"v?([0-9.]+)\""}]},
					"options": {
						"interval": "99m"}
				}`,
			wants: wants{
				statusCode: http.StatusNotFound,
				bodyRegex:  `^\{"message":"edit .* failed.*"\}`},
		},
		"edit service that doesn't query successfully": {
			serviceID: test.StringPtr("__name__"),
			payload: `
				{
					"id": "__name__",
					"latest_version": {
						"type": "url",
						"url":  "` + test.LookupPlain["url_valid"] + `",
						"url_commands": [
							{
								"type": "regex",
								"regex": "stable version: \"v-([0-9.]+)\""}]},
					"options": {
						"interval": "99m"}
				}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `^\{"message":"edit .* failed.*\\nlatest_version - no releases were found.*\\nregex \\".+\\" didn't return any matches on \\".+\\""\}`},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := testService(name, true)
			apiMutex.Lock()
			api.Config.Service[svc.ID] = svc
			api.Config.Order = append(api.Config.Order, svc.ID)
			apiMutex.Unlock()
			tc.payload = strings.ReplaceAll(tc.payload, "__name__", name)
			tc.payload = test.TrimJSON(tc.payload)
			payload := bytes.NewReader([]byte(tc.payload))
			var req *http.Request
			// CREATE.
			target := "/api/v1/service/new"
			req = httptest.NewRequest(http.MethodPost, target, payload)
			// EDIT.
			if tc.serviceID != nil {
				// set service_id.
				*tc.serviceID = strings.ReplaceAll(
					*tc.serviceID, "__name__", name)
				target = "/api/v1/service/config"
				params := url.Values{}
				// Set service_id.
				serviceID := util.DereferenceOrValue(tc.serviceID, svc.ID)
				params.Set("service_id", serviceID)
				req = httptest.NewRequest(http.MethodPut, target, payload)
				req.URL.RawQuery = params.Encode()
			}

			// WHEN that HTTP request is sent.
			w := httptest.NewRecorder()
			apiMutex.Lock()
			api.httpServiceEdit(w, req)
			apiMutex.Unlock()
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			// THEN the expected status code is returned.
			if res.StatusCode != tc.wants.statusCode {
				t.Errorf("%s\nstatus code mismatch\nwant: %d\ngot:  %d",
					packageName, tc.wants.statusCode, res.StatusCode)
			}
			// AND the expected body is returned.
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("%s\nunexpected error - %v",
					packageName, err)
			}
			got := string(data)
			if !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf("%s\nbody mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wants.bodyRegex, got)
			}
			if tc.wants.statusCode != http.StatusOK {
				return
			}
			// AND the service was created.
			serviceID := util.DereferenceOrDefault(tc.serviceID)
			// CREATE.
			if serviceID == "" {
				var data map[string]any
				_ = json.Unmarshal([]byte(tc.payload), &data)
				serviceID = data["id"].(string)
			}
			apiMutex.RLock()
			if tc.serviceID != nil &&
				api.Config.Service[*tc.serviceID] == nil {
				t.Errorf("%s\nservice %q not created",
					packageName, *tc.serviceID)
			}
			svc = api.Config.Service[serviceID]
			apiMutex.RUnlock()
			if svc == nil {
				if tc.wants.latestVersion != tc.wants.deployedVersion &&
					tc.wants.latestVersion != "" {
					t.Errorf("%s\nservice %q not created",
						packageName, serviceID)
				}
				return
			}
			// AND the LatestVersion is expected.
			if !util.RegexCheck(tc.wants.latestVersion, svc.Status.LatestVersion()) {
				t.Errorf("%s\nServiceEdit - LatestVersion\nwant: %q\ngot:  %q",
					packageName, tc.wants.latestVersion, svc.Status.LatestVersion())
			}
			// AND the DeployedVersion is expected.
			if !util.RegexCheck(tc.wants.deployedVersion, svc.Status.DeployedVersion()) {
				t.Errorf("%s\nServiceEdit - DeployedVersion\nwant: %q\ngot:  %q",
					packageName, tc.wants.deployedVersion, svc.Status.DeployedVersion())
			}
		})
	}
}

func TestHTTP_ServiceDelete(t *testing.T) {
	type wants struct {
		bodyRegex  string
		statusCode int
	}

	// GIVEN an API and a request to delete a service.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	t.Cleanup(func() {
		if api.Config.Settings.Data.DatabaseFile != "" {
			_ = os.RemoveAll(api.Config.Settings.Data.DatabaseFile)
		}
		// Give time for save before TempDir clean-up.
		time.Sleep(2 * config.DebounceDuration)
	})
	svc := testService("TestHTTP_ServiceDelete", true)
	svc.Init(
		&api.Config.Defaults.Service, &api.Config.HardDefaults.Service,
		&api.Config.Notify, &api.Config.Defaults.Notify, &api.Config.HardDefaults.Notify,
		&api.Config.WebHook, &api.Config.Defaults.WebHook, &api.Config.HardDefaults.WebHook)
	_ = api.Config.AddService("", svc)
	// Drain db from the Service addition.
	<-api.Config.DatabaseChannel
	tests := []struct {
		name      string
		serviceID string
		wants     wants
	}{
		{
			name:      "unknown service",
			serviceID: "foo",
			wants: wants{
				bodyRegex:  `{"message":"delete .* failed, service not found"`,
				statusCode: http.StatusNotFound},
		}, {
			name:      "delete service",
			serviceID: svc.ID,
			wants: wants{
				bodyRegex:  `\{"message":"deleted service[^}]+"\}`,
				statusCode: http.StatusOK},
		},
		{
			name:      "delete service again",
			serviceID: svc.ID,
			wants: wants{
				bodyRegex:  `{"message":"delete .* failed, service not found"`,
				statusCode: http.StatusNotFound},
		},
		{
			name:      "no service_id provided",
			serviceID: "",
			wants: wants{
				bodyRegex:  `{"message":"missing required query parameter: service_id"}`,
				statusCode: http.StatusBadRequest},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() -- Cannot run in parallel since we're sharing the API.

			target := "/api/v1/service/delete"
			params := url.Values{}
			// Set service_id.
			params.Set("service_id", tc.serviceID)

			// WHEN that HTTP request is sent.
			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			api.httpServiceDelete(w, req)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			// THEN the expected status code is returned.
			if res.StatusCode != tc.wants.statusCode {
				t.Errorf("%s\nstatus code mismatch\nwant: %d\ngot:  %d",
					packageName, tc.wants.statusCode, res.StatusCode)
			}
			// AND the expected body is returned.
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("%s\nunexpected error - %v",
					packageName, err)
			}
			got := string(data)
			if !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf("%s\nbody mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wants.bodyRegex, got)
			}
			// AND the service is removed from the config.
			if api.Config.Service[tc.serviceID] != nil {
				t.Errorf("%s\nservice %q not removed from Config.Service[]",
					packageName, tc.serviceID)
			}
			if util.Contains(api.Config.Order, tc.serviceID) {
				t.Errorf("%s\nservice %q not removed from Order",
					packageName, tc.serviceID)
			}
			// AND the service is removed from the database (if the req was OK).
			if tc.wants.statusCode == http.StatusOK {
				time.Sleep(time.Second)
				if len(api.Config.DatabaseChannel) == 0 {
					t.Errorf("%s\nservice %q not removed from database",
						packageName, tc.serviceID)
				} else {
					msg := <-api.Config.DatabaseChannel
					if msg.Delete != true {
						t.Errorf("%s\nServiceDelete should have sent a deletion to the db, not\n%+v",
							packageName, msg)
					}
				}
			}
		})
	}
}

func TestHTTP_NotifyTest(t *testing.T) {
	type wants struct {
		bodyRegex  string
		statusCode int
	}

	// GIVEN an API and a request to test a notify.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	t.Cleanup(func() {
		if api.Config.Settings.Data.DatabaseFile != "" {
			_ = os.RemoveAll(api.Config.Settings.Data.DatabaseFile)
		}
	})
	validNotify := shoutrrr_test.Shoutrrr(false, false)
	api.Config.Notify = shoutrrr.ShoutrrrsDefaults{}
	options := util.CopyMap(validNotify.Options)
	params := util.CopyMap(validNotify.Params)
	urlFields := util.CopyMap(validNotify.URLFields)
	api.Config.Notify["test"] = shoutrrr.NewDefaults(
		"gotify",
		options, urlFields, params)
	api.Config.Service["test"].Notify = map[string]*shoutrrr.Shoutrrr{
		"test":    shoutrrr_test.Shoutrrr(false, false),
		"no_main": shoutrrr_test.Shoutrrr(false, false)}
	tests := map[string]struct {
		queryParams map[string]string
		payload     string
		wants       wants
	}{
		"body too large": {
			payload: `{
				"test": "` + strings.Repeat(strings.Repeat("abcdefghijklmnopqrstuvwxyz", 100), 100) + `"}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  "request body too large"},
		},
		"no bodyRegex": {
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  "unexpected end of JSON input"},
		},
		"no service, new notify": {
			payload: `{
				"name": "new_notify"}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `invalid type "[^"]+"`},
		},
		"new service, no new/old notify": {
			payload: `{
				"service_id": "new_service"}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `name and/or name_previous are required`},
		},
		"new service, no main": {
			payload: `{
				"service_id": "also_unknown",
				"name": "test_notify",
				"type": "ntfy"}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  "url_fields:[^ ]+ +topic: .*required"},
		},
		"new service, no main - no service_id": {
			payload: `{
				"name": "test_notify",
				"type": "ntfy"}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  "url_fields:[^ ]+ +topic: .*required"},
		},
		"new service, no main - invalid JSON, options": {
			payload: `{
				"service_id": "also_unknown",
				"name": "test_notify",
				"type": "ntfy",
				"options": {
					"delay": "1s",
					"something" "else"}}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  "invalid character .* after object key"},
		},
		"new service, no main - options, invalid": {
			payload: `{
				"service_id": "also_unknown",
				"name": "test_notify",
				"type": "ntfy",
				"options": {
					"delay": "time"}}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `options:[^ ]+  delay: "[^"]+" <invalid>`},
		},
		"new service, have main - options, applied, delay ignored": {
			payload: `{
				"service_id": "also_unknown",
				"name": "test",
				"options": {
					"delay": "24h"}}`,
			wants: wants{
				statusCode: http.StatusOK,
				bodyRegex:  "message sent"},
		},
		"new service, no main - invalid JSON, url_fields": {
			payload: `{
				"service_id": "also_unknown",
				"name": "test_notify",
				"type": "ntfy",
				"url_fields": {
					"host" "example.com"}}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  "invalid character .* after object key"},
		},
		"new service, have main - url_fields, invalid": {
			payload: `{
				"service_id": "also_unknown",
				"name": "test",
				"url_fields": {
					"port": "number"}}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `failed to parse URL`},
		},
		"new service, no main - no type": {
			payload: `{
				"service_id": "also_unknown",
				"name": "test_notify"}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `invalid type "test_notify"`},
		},
		"new service, no main - unknown type": {
			payload: `{
				"service_id": "unknown",
				"name": "test_notify",
				"type": "something"}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `invalid type "something"`},
		},
		"new service, no main - type from ID": {
			payload: `{
				"service_id": "unknown",
				"name": "` + validNotify.Type + `",
				"url_fields": {
					"host": "` + validNotify.URLFields["host"] + `",
					"path": "` + validNotify.URLFields["path"] + `",
					"token": "` + validNotify.URLFields["token"] + `"}}`,
			wants: wants{
				statusCode: http.StatusOK,
				bodyRegex:  "message sent"},
		},
		"new service, have main - type from Main": {
			payload: `{
				"service_id": "unknown",
				"name": "test",
				"url_fields": {
					"host": "` + validNotify.URLFields["host"] + `",
					"path": "` + validNotify.URLFields["path"] + `",
					"token": "` + validNotify.URLFields["token"] + `"}}`,
			wants: wants{
				statusCode: http.StatusOK,
				bodyRegex:  "message sent"},
		},
		"same service, have main - type from original": {
			payload: `{
				"service_id_previous": "test",
				"service_id": "test",
				"name": "new_notify",
				"name_previous": "test",
				"url_fields": {
					"host": "` + validNotify.URLFields["host"] + `",
					"path": "` + validNotify.URLFields["path"] + `",
					"token": "` + util.SecretValue + `"}}`,
			wants: wants{
				statusCode: http.StatusOK,
				bodyRegex:  "message sent"},
		},
		"same service, no main - can remove vars": {
			payload: `{
				"service_id_previous": "test",
				"service_id": "test",
				"name": "new_notify",
				"name_previous": "no_main",
				"url_fields": {
					"host": "` + validNotify.URLFields["host"] + `",
					"path": "` + validNotify.URLFields["path"] + `",
					"token": ""}}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex: test.TrimYAML(`
				^url_fields:
					token: <required>.*$`)},
		},
		"same service, no main - unsent vars inherited": {
			payload: `{
				"service_id_previous": "test",
				"service_id": "test",
				"name": "new_notify",
				"name_previous": "no_main",
				"type": "` + validNotify.Type + `",
				"url_fields": {
					"host": "` + validNotify.URLFields["host"] + `",
					"path": "` + validNotify.URLFields["path"] + `"}}`,
			wants: wants{
				statusCode: http.StatusOK,
				bodyRegex:  "message sent"},
		},
		"same service, have main - fail send": {
			payload: `{
				"service_id_previous": "test",
				"service_id": "test",
				"name": "test",
				"name_previous": "test",
				"type": "` + validNotify.Type + `",
				"url_fields": {
					"host": "` + validNotify.URLFields["host"] + `",
					"path": "` + validNotify.URLFields["path"] + `",
					"token": "invalid"}}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  "invalid .* token"},
		},
		"same service, have main - new name, also fail send": {
			payload: `{
				"service_id_previous": "test",
				"service_id": "new_name",
				"name": "test",
				"name_previous": "test",
				"type": "` + validNotify.Type + `",
				"url_fields": {
					"host": "` + validNotify.URLFields["host"] + `",
					"path": "` + validNotify.URLFields["path"] + `",
					"token": "invalid"}}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  "invalid .* token"},
		},
		"service_id_previous that doesn't exist": {
			payload: `{
				"service_id_previous": "does_not_exist",
				"service_id": "test",
				"name": "new_notify",
				"name_previous": "test",
				"url_fields": {
					"host": "` + validNotify.URLFields["host"] + `",
					"path": "` + validNotify.URLFields["path"] + `",
					"token": "` + util.SecretValue + `"}}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `invalid type "new_notify"`},
		},
		"name_previous that doesn't exist": {
			payload: `{
				"service_id_previous": "test",
				"service_id": "test",
				"name": "new_notify",
				"name_previous": "does_not_exist",
				"url_fields": {
					"host": "` + validNotify.URLFields["host"] + `",
					"path": "` + validNotify.URLFields["path"] + `",
					"token": "` + util.SecretValue + `"}}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `invalid type "new_notify"`},
		},
		"service_id_previous and name_previous that doesn't exist": {
			payload: `{
				"service_id_previous": "does_not_exist",
				"service_id": "test",
				"name": "new_notify",
				"name_previous": "also_does_not_exist",
				"url_fields": {
					"host": "` + validNotify.URLFields["host"] + `",
					"path": "` + validNotify.URLFields["path"] + `",
					"token": "` + util.SecretValue + `"}}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `invalid type "new_notify"`},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.wants.bodyRegex = util.ValueOrValue(tc.wants.bodyRegex, `^$`)
			tc.payload = test.TrimJSON(tc.payload)
			payload := bytes.NewReader([]byte(tc.payload))

			// WHEN that request is sent.
			req := httptest.NewRequest(http.MethodGet,
				"/api/v1/notify/test",
				payload)
			w := httptest.NewRecorder()
			api.httpNotifyTest(w, req)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			// THEN the expected status code is returned.
			if res.StatusCode != tc.wants.statusCode {
				t.Errorf("%s\nstatus code mismatch\nwant: %d\ngot:  %d",
					packageName, tc.wants.statusCode, res.StatusCode)
			}
			// AND the expected message is contained in the bodyRegex.
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("%s\nunexpected error - %v",
					packageName, err)
			}
			// Marshal message out of JSON data {"message": text}.
			var body map[string]string
			_ = json.Unmarshal(data, &body)
			if !util.RegexCheck(tc.wants.bodyRegex, body["message"]) {
				t.Errorf("%s\nbody mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wants.bodyRegex, body["message"])
			}
		})
	}
}
