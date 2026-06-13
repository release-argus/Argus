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

//go:build integration

package v1

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service"
	dvweb "github.com/release-argus/Argus/service/deployed_version/types/web"
	"github.com/release-argus/Argus/service/shared"
	svctest "github.com/release-argus/Argus/service/test"
	"github.com/release-argus/Argus/util"
	whtest "github.com/release-argus/Argus/webhook/test"
)

func TestHTTP_LatestVersionRefreshUncreated(t *testing.T) {
	type wants struct {
		bodyRegex  string
		statusCode int
	}

	// GIVEN: an API and a request to refresh the latest_version of a service.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	tests := []struct {
		name   string
		params map[string]string
		wants  wants
	}{
		{
			name:   "no overrides",
			params: map[string]string{},
			wants: wants{
				bodyRegex:  `"message":"overrides: .*required`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid JSON",
			params: map[string]string{
				"overrides": `"type": "url", "url": "` + test.LookupPlain["url_valid"] + `"}`,
			},
			wants: wants{
				bodyRegex: `` +
					`^{"message":"latest_version:\\n` +
					`  .*cannot unmarshal[^"]+"}$`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid vars - Decode fail",
			params: map[string]string{
				"overrides": `{"type": "something"}`,
			},
			wants: wants{
				bodyRegex:  `"message":".*type: .*invalid`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid vars - CheckValues fail",
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"type":         "url",
					"url":          "` + test.LookupPlain["url_valid"] + `",
					"url_commands": "[{\"type\": \"regex\"}]"
				}`),
			},
			wants: wants{
				bodyRegex:  `"message":"url_commands.*regex:.*required`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "valid vars",
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"type":         "url",
					"url":          "` + test.LookupPlain["url_valid"] + `",
					"url_commands": "[{\"type\": \"regex\", \"regex\": \"stable version: \\\"v?([0-9.]+)\\\"\"}]"
				}`),
			},
			wants: wants{
				bodyRegex:  `^{"version":"[0-9.]+","timestamp":"[^"]+"}\s$`,
				statusCode: http.StatusOK,
			},
		},
		{
			name: "query fail",
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"type":         "url",
					"url":          "` + test.LookupPlain["url_invalid"] + `",
					"url_commands": "[{\"type\": \"regex\", \"regex\": \"stable version: \\\"v?([0-9.]+)\\\"\"}]"
				}`),
			},
			wants: wants{
				bodyRegex:  `"message":"x509 `,
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			target := "/api/v1/latest_version/refresh_uncreated"
			params := url.Values{}
			for k, v := range tc.params {
				params.Set(k, v)
			}

			// WHEN: that HTTP request is sent.
			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			api.httpLatestVersionRefreshUncreated(w, req)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			prefix := fmt.Sprintf("%s\nAPI.httpLatestVersionRefreshUncreated()", packageName)

			// THEN: the expected status code is returned.
			if got, want := res.StatusCode, tc.wants.statusCode; got != want {
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
			if got := string(data); !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf(
					"%s body mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wants.bodyRegex,
				)
			}
		})
	}
}

func TestHTTP_DeployedVersionRefreshUncreated(t *testing.T) {
	type wants struct {
		bodyRegex  string
		statusCode int
	}

	// GIVEN: an API and a request to refresh the deployed_version of a service.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	tests := []struct {
		name   string
		params map[string]string
		wants  wants
	}{
		{
			name:   "no vars",
			params: map[string]string{},
			wants: wants{
				bodyRegex:  `"message":"overrides: .*required`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid JSON",
			params: map[string]string{
				"overrides": `"type": "url", "url": "` + test.LookupPlain["url_valid"] + `"}`,
			},
			wants: wants{
				bodyRegex: `` +
					`^{"message":"deployed_version:\\n` +
					`  .*cannot unmarshal[^"]+"}$`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid vars - Decode fail",
			params: map[string]string{
				"overrides": `{"type": "something"}`,
			},
			wants: wants{
				bodyRegex:  `"message":".*type: .*invalid`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid vars - CheckValues fail",
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"type":  "url",
					"url":   "` + test.LookupPlain["url_valid"] + `",
					"regex": "stable version: \"v?([0-9.+)\""
				}`),
			},
			wants: wants{
				bodyRegex:  `{"message":"regex: .*invalid`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "valid vars",
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"type":  "url",
					"url":   "` + test.LookupPlain["url_valid"] + `",
					"regex": "stable version: \"v?([0-9.]+)\""
				}`),
			},
			wants: wants{
				bodyRegex:  `^{"version":"[0-9.]+","timestamp":"[^"]+"}\s$`,
				statusCode: http.StatusOK,
			},
		},
		{
			name: "query fail",
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"type":  "url",
					"url":   "` + test.LookupPlain["url_invalid"] + `",
					"regex": "stable version: \"v?([0-9.]+)\""
				}`),
			},
			wants: wants{
				bodyRegex:  `"message":"x509 `,
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			target := "/api/v1/deployed_version/refresh_uncreated"
			params := url.Values{}
			for k, v := range tc.params {
				params.Set(k, v)
			}

			// WHEN: that HTTP request is sent.
			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			api.httpDeployedVersionRefreshUncreated(w, req)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			prefix := fmt.Sprintf("%s\nAPI.httpDeployedVersionRefreshUncreated()", packageName)

			// THEN: the expected status code is returned.
			if got, want := res.StatusCode, tc.wants.statusCode; got != want {
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
			if got := string(data); !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf(
					"%s body mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wants.bodyRegex,
				)
			}
		})
	}
}

func TestHTTP_LatestVersionRefresh(t *testing.T) {
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()

	type wants struct {
		bodyRegex                      string
		statusCode                     int
		latestVersion, deployedVersion string
		announce                       bool
	}

	// GIVEN: an API and a request to refresh the latest_version of a service.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	apiMu := sync.RWMutex{}

	tests := []struct {
		name      string
		serviceID *string
		svc       *service.Service
		params    map[string]string
		wants     wants
	}{
		{
			name:   "no changes",
			params: map[string]string{},
			svc: test.Must(t, func() (*service.Service, error) {
				svcCfg := svctest.PlainDefaultsConfig()
				return service.DecodeService(
					"yaml", []byte(test.TrimYAML(`
						options:
							semantic_versioning: false
						latest_version:
							type: url
							url: `+test.LookupBare["url_valid"]+`/ver1.2.3
					`)),
					"__name__",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			wants: wants{
				bodyRegex:       `{"version":"ver[\d.]+",.*"}`,
				statusCode:      http.StatusOK,
				latestVersion:   "ver1.2.3",
				deployedVersion: "ver1.2.3",
				announce:        true,
			},
		},
		{
			name: "semantic_versioning=null - fail as default=true",
			params: map[string]string{
				"semantic_versioning": "null",
			},
			svc: test.Must(t, func() (*service.Service, error) {
				svcCfg := svctest.PlainDefaultsConfig()
				return service.DecodeService(
					"yaml", []byte(test.TrimYAML(`
						options:
							semantic_versioning: false
						latest_version:
							type: url
							url: `+test.LookupBare["url_valid"]+`/ver1.2.3
					`)),
					"__name__",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			wants: wants{
				bodyRegex:     `failed to convert \\"[^"]+\\" to a semantic version`,
				statusCode:    http.StatusBadRequest,
				latestVersion: "",
			},
		},
		{
			name: "semantic_versioning=same - refreshes service",
			params: map[string]string{
				"semantic_versioning": "false",
			},
			svc: test.Must(t, func() (*service.Service, error) {
				svcCfg := svctest.PlainDefaultsConfig()
				return service.DecodeService(
					"yaml", []byte(test.TrimYAML(`
						options:
							semantic_versioning: false
						latest_version:
							type: url
							url: `+test.LookupBare["url_valid"]+`/ver1.2.3
					`)),
					"__name__",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			wants: wants{
				bodyRegex:       `{"version":"ver[\d.]+",.*"}`,
				statusCode:      http.StatusOK,
				latestVersion:   "ver1.2.3",
				deployedVersion: "ver1.2.3",
				announce:        true,
			},
		},
		{
			name: "semantic_versioning=diff - not applied to service",
			params: map[string]string{
				"semantic_versioning": "true",
			},
			svc: test.Must(t, func() (*service.Service, error) {
				svcCfg := svctest.PlainDefaultsConfig()
				return service.DecodeService(
					"yaml", []byte(test.TrimYAML(`
						options:
							semantic_versioning: false
						latest_version:
							type: url
							url: `+test.LookupBare["url_valid"]+`/ver1.2.3
					`)),
					"__name__",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			wants: wants{
				bodyRegex:     `failed to convert \\"[^"]+\\" to a semantic version`,
				statusCode:    http.StatusBadRequest,
				latestVersion: "",
			},
		},
		{
			name: "different regex doesn't update service version",
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"url_commands": [
						{"type":"regex", "regex":"v?([0-9.]+-beta)"}
					]
				}`),
			},
			svc: test.Must(t, func() (*service.Service, error) {
				svcCfg := svctest.PlainDefaultsConfig()
				return service.DecodeService(
					"yaml", []byte(test.TrimYAML(`
						options:
							semantic_versioning: false
						latest_version:
							type: url
							url: `+test.LookupBare["url_valid"]+`/v1.2.3-beta
							url_commands:
								- type: regex
								  regex: "v([0-9.]+-beta)"
					`)),
					"__name__",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			wants: wants{
				bodyRegex:     `{"version":"[0-9.]+-beta",.*"}`,
				statusCode:    http.StatusOK,
				latestVersion: "",
				announce:      false,
			},
		},
		{
			name: "invalid vars",
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"url_commands": [
						{
							"type":"regex",
							"regex":"beta: \\\"v?([0-9.+-beta)\\\""
						}
					]
				}`),
				"semantic_versioning": "false",
			},
			wants: wants{
				bodyRegex:     `{"message":".*regex: .*invalid`,
				statusCode:    http.StatusBadRequest,
				latestVersion: "",
			},
		},
		{
			name:      "unknown service",
			serviceID: test.Ptr("bash-bosh"),
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"url_commands": [
						{"type":"regex","regex":"beta: \\\"v?([0-9.+-beta)\\\""}
					]
				}`),
				"semantic_versioning": "false",
			},
			wants: wants{
				bodyRegex:     `{"message":"service .+ not found"}`,
				statusCode:    http.StatusNotFound,
				latestVersion: "",
			},
		},
		{
			name:      "no service_id provided",
			serviceID: test.Ptr(""),
			wants: wants{
				bodyRegex:     `{"message":"missing required query parameter: service_id"}`,
				statusCode:    http.StatusBadRequest,
				latestVersion: "",
			},
		},
		{
			name: "use secretRefs",
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
			svc: test.Must(t, func() (*service.Service, error) {
				svcCfg := svctest.PlainDefaultsConfig()
				return service.DecodeService(
					"yaml", []byte(test.TrimYAML(`
						options:
							semantic_versioning: false
						latest_version:
							type: url
							url: `+test.LookupWithHeaderAuth["url_valid"]+`
							headers:
								- key: `+test.LookupWithHeaderAuth["header_key"]+`
									value: `+test.LookupWithHeaderAuth["header_value_pass"]+`
					`)),
					"__name__",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			wants: wants{
				bodyRegex:  `{"version":"[0-9.]+",.*"}`,
				statusCode: http.StatusOK,
				announce:   false,
			},
		},
		{
			name: "invalid secretRefs",
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
			svc: test.Must(t, func() (*service.Service, error) {
				svcCfg := svctest.PlainDefaultsConfig()
				return service.DecodeService(
					"yaml", []byte(test.TrimYAML(`
						options:
							semantic_versioning: false
						latest_version:
							type: url
							url: `+test.LookupWithHeaderAuth["url_valid"]+`
							headers:
								- key: `+test.LookupWithHeaderAuth["header_key"]+`
									value: `+test.LookupWithHeaderAuth["header_value_pass"]+`
					`)),
					"__name__",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			wants: wants{
				bodyRegex:  `.*unmarshal.* array`,
				statusCode: http.StatusBadRequest,
				announce:   false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var svc *service.Service
			if tc.svc != nil {
				svc = tc.svc
				if svc.ID == "__name__" {
					svc.ID = tc.name
				}
			} else {
				svc = testService(t, tc.name, "url", "url", false)
			}
			apiMu.Lock()
			api.Config.Service[svc.ID] = svc
			apiMu.Unlock()

			target := "/api/v1/latest_version/refresh"
			// Add the params to the URL.
			params := url.Values{}
			for k, v := range tc.params {
				params.Set(k, v)
			}
			// Set service_id.
			serviceID := util.DerefOr(tc.serviceID, svc.ID)
			params.Set("service_id", serviceID)

			// WHEN: that HTTP request is sent.
			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			apiMu.Lock()
			api.httpLatestVersionRefresh(w, req)
			apiMu.Unlock()
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			prefix := fmt.Sprintf("%s\nAPI.httpLatestVersionRefresh()", packageName)

			// THEN: the expected status code is returned.
			if got, want := res.StatusCode, tc.wants.statusCode; got != want {
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
			if got := string(data); !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf(
					"%s body mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wants.bodyRegex,
				)
			}

			// AND: the LatestVersion is expected.
			gotLV := svc.Status.LatestVersion()
			wantLV := tc.wants.latestVersion
			if gotLV != wantLV {
				t.Errorf(
					"%s LatestVersion mismatch\ngot:  %q\nwant: %q",
					prefix, gotLV, wantLV,
				)
			}

			// AND: the DeployedVersion is expected.
			gotDV := svc.Status.DeployedVersion()
			wantDV := tc.wants.deployedVersion
			if gotDV != wantDV {
				t.Errorf(
					"%s DeployedVersion mismatch\ngot:  %q\nwant: %q",
					prefix, gotDV, wantDV,
				)
			}

			// AND: the expected announcement is made.
			wantAnnounces := 0
			if tc.wants.announce {
				wantAnnounces = 1
			}
			if gotAnnounces := len(svc.Status.AnnounceChannel); gotAnnounces != wantAnnounces {
				t.Errorf(
					"%s AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					prefix, gotAnnounces, wantAnnounces,
				)
			}
		})
	}
}

func TestHTTP_DeployedVersionRefresh(t *testing.T) {
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()

	type wants struct {
		bodyRegex       string
		statusCode      int
		deployedVersion string
	}

	// GIVEN: an API and a request to refresh the deployed_version of a service.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	apiMu := sync.RWMutex{}

	tests := []struct {
		name               string
		serviceID          *string
		svc                *service.Service
		nilDeployedVersion bool
		params             map[string]string
		wants              wants
	}{
		{
			name:               "adding deployed version to service - success",
			nilDeployedVersion: true,
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"type":                "url",
					"url":                 "` + test.LookupBare["url_invalid"] + "/" + url.QueryEscape(`{"foo":"ver1.2.3-beta"}`) + `",
					"json":                "foo",
					"allow_invalid_certs": true
				}`),
			},
			wants: wants{
				bodyRegex:       `{"version":"ver1.2.3-beta",.*}`,
				statusCode:      http.StatusOK,
				deployedVersion: "",
			},
		},
		{
			name:               "adding deployed version to service - no overrides",
			nilDeployedVersion: true,
			wants: wants{
				bodyRegex:  `{"message":"missing required parameter: overrides"`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:   "no changes",
			params: map[string]string{},
			svc: test.Must(t, func() (*service.Service, error) {
				svcCfg := svctest.PlainDefaultsConfig()
				return service.DecodeService(
					"yaml", []byte(test.TrimYAML(`
						options:
							semantic_versioning: false
						deployed_version:
							type: url
							url: `+test.LookupBare["url_valid"]+`/ver1.2.3
					`)),
					"__name__",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			wants: wants{
				bodyRegex:       `{"version":"ver[\d.]+",.*"}`,
				statusCode:      http.StatusOK,
				deployedVersion: "ver1.2.3",
			},
		},
		{
			name: "semantic_versioning=null - fail as default=true",
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"url_commands": [
						{"type":"regex","regex":"beta: \"v?([0-9.]+-beta\")"}
					]
				}`),
				"semantic_versioning": "null",
			},
			svc: test.Must(t, func() (*service.Service, error) {
				svcCfg := svctest.PlainDefaultsConfig()
				return service.DecodeService(
					"yaml", []byte(test.TrimYAML(`
						options:
							semantic_versioning: false
						deployed_version:
							type: url
							url: `+test.LookupBare["url_valid"]+`/ver1.2.3
					`)),
					"__name__",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			wants: wants{
				bodyRegex:       `failed to convert \\"[^"]+\\" to a semantic version`,
				statusCode:      http.StatusBadRequest,
				deployedVersion: "",
			},
		},
		{
			name: "semantic_versioning=diff - not applied to service",
			params: map[string]string{
				"semantic_versioning": "true",
			},
			svc: test.Must(t, func() (*service.Service, error) {
				svcCfg := svctest.PlainDefaultsConfig()
				return service.DecodeService(
					"yaml", []byte(test.TrimYAML(`
						options:
							semantic_versioning: false
						deployed_version:
							type: url
							url: `+test.LookupBare["url_valid"]+`/ver1.2.3
					`)),
					"__name__",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			wants: wants{
				bodyRegex:       `failed to convert \\"[^"]+\\" to a semantic version`,
				statusCode:      http.StatusBadRequest,
				deployedVersion: "",
			},
		},
		{
			name: "different JSON doesn't update service version",
			params: map[string]string{
				"overrides":           `{"json": "bar"}`,
				"semantic_versioning": "false",
			},
			svc: test.Must(t, func() (*service.Service, error) {
				svcCfg := svctest.PlainDefaultsConfig()
				return service.DecodeService(
					"yaml", []byte(test.TrimYAML(`
						options:
							semantic_versioning: false
						deployed_version:
							type: url
							url: `+test.LookupBare["url_valid"]+"/"+url.QueryEscape(`{"foo":"ver1.2.3-beta","bar":"ver1.2.3-beta"}`)+`
							json: foo
					`)),
					"__name__",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			wants: wants{
				bodyRegex:       `{"version":"ver[\d.]+-beta",.*"}`,
				statusCode:      http.StatusOK,
				deployedVersion: "",
			},
		},
		{
			name: "invalid JSON - existing DVL",
			params: map[string]string{
				"overrides": `{"json": "x.y"}`,
			},
			wants: wants{
				bodyRegex:  `^{"message":"failed to unmarshal response from .*"}$`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:               "JSON key not found - new DVL",
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
				bodyRegex: `` +
					`^{"message":"failed to navigate JSON:\\n` +
					`  failed to find value for \\"x\.y\\" in .*"}$`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid vars - CheckValues fail",
			params: map[string]string{
				"overrides": `{"regex": "v?([0-9.+)"}`,
			},
			wants: wants{
				bodyRegex:  `{"message":".*regex: .*invalid`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:      "unknown service",
			serviceID: test.Ptr("bish-bash-bosh"),
			params: map[string]string{
				"semantic_versioning": "false",
			},
			wants: wants{
				bodyRegex:  `{"message":"service .+ not found"`,
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:      "no service_id provided",
			serviceID: test.Ptr(""),
			wants: wants{
				bodyRegex:       `{"message":"missing required query parameter: service_id"}`,
				statusCode:      http.StatusBadRequest,
				deployedVersion: "",
			},
		},
		{
			name: "use secretRefs",
			svc: test.Must(t, func() (*service.Service, error) {
				base := testService(t, "TestHTTP_LatestVersionRefresh", "url", "url", false)
				if dv, ok := base.DeployedVersionLookup.(*dvweb.Lookup); ok {
					dv.URL = test.LookupWithHeaderAuth["url_valid"]
					dv.Headers = shared.Headers{
						{Key: test.LookupWithHeaderAuth["header_key"], Value: test.LookupWithHeaderAuth["header_value_pass"]},
					}
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
				bodyRegex:  `{"version":"[0-9.]+",.*"}`,
				statusCode: http.StatusOK,
			},
		},
		{
			name: "invalid secretRefs",
			svc: test.Must(t, func() (*service.Service, error) {
				base := testService(t, "TestHTTP_LatestVersionRefresh", "url", "url", false)
				if dv, ok := base.DeployedVersionLookup.(*dvweb.Lookup); ok {
					dv.URL = test.LookupWithHeaderAuth["url_valid"]
					dv.Headers = shared.Headers{
						{Key: test.LookupWithHeaderAuth["header_key"], Value: test.LookupWithHeaderAuth["header_value_pass"]},
					}
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
				bodyRegex:  `.*unmarshal.* array`,
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var svc *service.Service
			if tc.svc != nil {
				svc = tc.svc
				if svc.ID == "__name__" {
					svc.ID = tc.name
				}
			} else {
				svc = testService(t, tc.name, "url", "url", false)
			}
			apiMu.Lock()
			api.Config.Service[svc.ID] = svc
			apiMu.Unlock()
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
			serviceID := util.DerefOr(tc.serviceID, svc.ID)
			params.Set("service_id", serviceID)

			// WHEN: that HTTP request is sent.
			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			apiMu.Lock()
			api.httpDeployedVersionRefresh(w, req)
			apiMu.Unlock()
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			prefix := fmt.Sprintf("%s\nAPI.httpDeployedVersionRefresh()", packageName)

			// THEN: the expected status code is returned.
			if got, want := res.StatusCode, tc.wants.statusCode; got != want {
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
			if got := string(data); !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf(
					"%s body mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wants.bodyRegex,
				)
			}

			// AND: the LatestVersion is unchanged.
			if gotLatestVersion := svc.Status.LatestVersion(); gotLatestVersion != initialLatestVersion {
				t.Errorf(
					"%s LatestVersion mismatch\ngot:  %q\nwant: %q",
					prefix, gotLatestVersion, initialLatestVersion,
				)
			}

			// AND: the DeployedVersion is expected.
			if gotDeployedVersion := svc.Status.DeployedVersion(); gotDeployedVersion != tc.wants.deployedVersion {
				t.Errorf(
					"%s DeployedVersion mismatch\ngot:  %q\nwant: %q",
					prefix, gotDeployedVersion, tc.wants.deployedVersion,
				)
			}
		})
	}
}

func TestHTTP_ServiceDetail(t *testing.T) {
	svcCfg := svctest.PlainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()

	type wants struct {
		bodyRegex  string
		statusCode int
	}

	// GIVEN: an API and a request for detail of a service.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	apiMu := sync.RWMutex{}

	tests := []struct {
		name      string
		svc       *service.Service
		serviceID *string
		wants     wants
	}{
		{
			name: "known service",
			svc: test.Must(t, func() (*service.Service, error) {
				return service.DecodeService(
					"yaml", []byte(test.TrimYAML(`
						name: foo
						comment: hello
						options:
							semantic_versioning: false
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
						deployed_version:
							type: url
							url: `+test.LookupBare["url_valid"]+`/ver1.2.3
					`)),
					"__name__",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			wants: wants{
				bodyRegex: test.TrimJSON(`{
					"name": "foo",
					"comment": "hello",
					"options": {
						"semantic_versioning": false
					},
					"latest_version": {
						"type": "github",
						"url": "` + test.ArgusGitHubRepo + `"
					},
					"deployed_version": {
						"type": "url",
						"url": "` + test.LookupBare["url_valid"] + `/ver1.2.3"
					}
				}`),
				statusCode: http.StatusOK,
			},
		},
		{
			name:      "unknown service",
			serviceID: test.Ptr("bish-bash-bosh"),
			wants: wants{
				bodyRegex:  `{"message":"service .+ not found"`,
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:      "no service_id provided",
			serviceID: test.Ptr(""),
			wants: wants{
				bodyRegex:  `{"message":"missing required query parameter: service_id"}`,
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var svc *service.Service
			if tc.svc != nil {
				svc = tc.svc
				if svc.ID == "__name__" {
					svc.ID = tc.name
				}
			} else {
				svc = testService(t, tc.name, "url", "url", true)
			}
			apiMu.Lock()
			api.Config.Service[svc.ID] = svc
			apiMu.Unlock()
			target := "/api/v1/service/config"
			params := url.Values{}
			// Set service_id.
			serviceID := util.DerefOr(tc.serviceID, svc.ID)
			params.Set("service_id", serviceID)

			// WHEN: that HTTP request is sent.
			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			apiMu.RLock()
			api.httpServiceDetail(w, req)
			apiMu.RUnlock()
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			prefix := fmt.Sprintf("%s\nAPI.httpServiceDetail()", packageName)

			// THEN: the expected status code is returned.
			if got, want := res.StatusCode, tc.wants.statusCode; got != want {
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
			if got := string(data); !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf(
					"%s body mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wants.bodyRegex,
				)
			}
		})
	}
}

func TestHTTP_OtherServiceDetails(t *testing.T) {
	// GIVEN: an API and a request for detail of a service.
	tests := []struct {
		name           string
		wantBody       string
		wantStatusCode int
	}{
		{
			name: "get details",
			wantBody: test.TrimJSON(`
				"hard_defaults": {.*
					"service": {.*
						"options": {.*
							"interval": "10m",.*
					"notify": {.*
					"webhook": {.*
				}
				.*
				"defaults": {.*
					"notify": {.*
					"webhook": {.*
				}`,
			),
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), tc.name+".yaml")
			api := testAPI(t, path)
			svc := testService(t, tc.name, "url", "url", true)
			api.Config.Service[svc.ID] = svc
			target := "/api/v1/service/defaults"

			// WHEN: that HTTP request is sent.
			req := httptest.NewRequest(http.MethodGet, target, nil)
			w := httptest.NewRecorder()
			api.httpOtherServiceDetails(w, req)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			prefix := fmt.Sprintf("%s\nAPI.httpOtherServiceDetails()", packageName)

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
			got := string(data)
			if !util.RegexCheck(tc.wantBody, got) {
				t.Errorf(
					"%s body mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wantBody,
				)
			}
		})
	}
}

func TestHTTP_TemplateParse(t *testing.T) {
	type wants struct {
		bodyRegex  string
		statusCode int
	}

	// GIVEN: an API and a request to parse a template.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	apiMu := sync.RWMutex{}

	testSVC := testService(t, "TestHTTP_TemplateParse", "url", "url", true)
	apiMu.Lock()
	api.Config.Service[testSVC.ID] = testSVC
	apiMu.Unlock()

	tests := []struct {
		name        string
		queryParams map[string]string
		wants       wants
	}{
		{
			name:        "missing required parameters",
			queryParams: map[string]string{},
			wants: wants{
				bodyRegex:  `{"message":"missing required query parameter: service_id"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "missing required parameters - service_id",
			queryParams: map[string]string{
				"template": "{{ service_name }}",
			},
			wants: wants{
				bodyRegex:  `{"message":"missing required query parameter: service_id"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "missing required parameters - template",
			queryParams: map[string]string{
				"service_id": test.ArgusGitHubRepo,
			},
			wants: wants{
				bodyRegex:  `{"message":"missing required query parameter: template"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid template",
			queryParams: map[string]string{
				"service_id": testSVC.ID,
				"template":   "{{.InvalidField}",
			},
			wants: wants{
				bodyRegex:  `{"message":"failed to parse template"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid params JSON",
			queryParams: map[string]string{
				"service_id": testSVC.ID,
				"template":   "{{ service_name }}",
				"params":     `{"invalid":}`,
			},
			wants: wants{
				bodyRegex: `` +
					`{"message":"invalid 'params' query parameter format -\\n` +
					`  jsontext: invalid character '}'.*\\n` +
					`    invalid character.*"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "valid template with default parameters",
			queryParams: map[string]string{
				"service_id": testSVC.ID,
				"template":   `{{service_name }} - {{ version }}`,
			},
			wants: wants{
				bodyRegex: fmt.Sprintf(
					`{"parsed":"%s - %s"}`,
					testSVC.GetName(), testSVC.Status.LatestVersion(),
				),
				statusCode: http.StatusOK,
			},
		},
		{
			name: "valid template with overridden parameters",
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
		{
			name: "unknown service",
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

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			target := "/api/v1/template/parse"
			params := url.Values{}
			for k, v := range tc.queryParams {
				params.Set(k, v)
			}

			// WHEN: that HTTP request is sent.
			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			api.httpTemplateParse(w, req)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			prefix := fmt.Sprintf("%s\nAPI.httpTemplateParse()", packageName)

			// THEN: the expected status code is returned.
			if got, want := res.StatusCode, tc.wants.statusCode; got != want {
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
			if got := string(data); !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf(
					"%s body mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wants.bodyRegex,
				)
			}
		})
	}
}

func TestHTTP_ServiceEdit__Create(t *testing.T) {
	testSVC := testService(t, "TestHTTP_ServiceEdit_Create", "url", "url", true)
	testSVC.LatestVersion.GetStatus().SetLatestVersion("1.0.0", "", false)
	_, _ = testSVC.LatestVersion.Query(true, logx.LogFrom{})
	_ = testSVC.DeployedVersionLookup.Query(true, logx.LogFrom{})
	type wants struct {
		bodyRegex                      string
		statusCode                     int
		latestVersion, deployedVersion string
		serviceYAML                    string
	}

	// GIVEN: an API and a request to create a service.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	apiMu := sync.RWMutex{}

	// Give time for save before TempDir clean-up.
	t.Cleanup(func() { time.Sleep(2 * config.DebounceDuration) })

	tests := []struct {
		name      string
		payload   string
		serviceID string
		wants     wants
	}{
		{
			name: "invalid JSON",
			payload: test.TrimJSON(`
				"id": "__name__-",
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `"
			`),
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex: `` +
					`^{"message":"create .* failed:\\n` +
					`  unmarshal service payload:\\n` +
					`    json: .*unmarshal[^"]+"}`,
			},
		},
		{
			name: "new service - lv-github",
			payload: test.TrimJSON(`{
				"id": "__name__-new",
				"name": "__name__-foo",
				"comment": "hello",
				"options": {
					"semantic_versioning": true
				},
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `",
					"url_commands": [
						{
							"type": "regex",
							"regex": "v?([\\d.]+)"
						}
					],
					"require": {
						"regex_content": "(ver)?{{ version }}",
						"regex_version": "^[\\d.]+$"
					},
					"allow_invalid_certs": true
				}
			}`),
			serviceID: "__name__-new",
			wants: wants{
				statusCode: http.StatusOK,
				bodyRegex:  `^{"message":"created service[^}]+"}`,
				serviceYAML: test.TrimYAML(`
					name: '__name__-foo'
					comment: hello
					options:
						semantic_versioning: true
					latest_version:
						type: github
						url: ` + test.ArgusGitHubRepo + `
						url_commands:
							- type: regex
								regex: 'v?([\d.]+)'
						require:
							regex_content: (ver)?{{ version }}
							regex_version: '^[\d.]+$'
				`),
			},
		},
		{
			name: "new service - lv-url",
			payload: test.TrimJSON(`{
				"id": "__name__-new",
				"comment": "goodbye",
				"options": {
					"semantic_versioning": false
				},
				"latest_version": {
					"type": "url",
					"url": "` + test.LookupBare["url_invalid"] + "/" + url.QueryEscape(`versions here: "ver1.2.3", release=1.2.3.exe`) + `",
					"url_commands": [
						{
							"type": "regex",
							"regex": "v?([0-9.]+)"
						}
					],
					"require": {
						"regex_content": "{{ version }}.exe",
						"regex_version": "^[\\d.]+$"
					},
					"allow_invalid_certs": true
				}
			}`),
			serviceID: "__name__-new",
			wants: wants{
				statusCode:    http.StatusOK,
				bodyRegex:     `{"message":"created service[^}]+"}`,
				latestVersion: "1.2.3",
				serviceYAML: test.TrimYAML(`
					comment: goodbye
					options:
						semantic_versioning: false
					latest_version:
						type: url
						url: ` + test.LookupBare["url_invalid"] + "/" + url.QueryEscape(`versions here: "ver1.2.3", release=1.2.3.exe`) + `
						url_commands:
							- type: regex
								regex: v?([0-9.]+)
						require:
							regex_content: '{{ version }}.exe'
							regex_version: '^[\d.]+$'
						allow_invalid_certs: true
				`),
			},
		},
		{
			name: "new service - dv-manual",
			payload: test.TrimJSON(`{
				"id": "__name__-new",
				"name": "__name__-foo",
				"comment": "hi",
				"options": {
					"semantic_versioning": true
				},
				"deployed_version": {
					"type": "manual",
					"version": "1.2.3"
				}
			}`),
			serviceID: "__name__-new",
			wants: wants{
				statusCode:      http.StatusOK,
				bodyRegex:       `{"message":"created service[^}]+"}`,
				deployedVersion: "1.2.3",
				serviceYAML: test.TrimYAML(`
					name: '__name__-foo'
					comment: hi
					options:
						semantic_versioning: true
					deployed_version:
						type: manual
				`),
			},
		},
		{
			name: "new service - dv-url",
			payload: test.TrimJSON(`{
					"id": "__name__-new",
					"comment": "bye",
					"options": {
						"semantic_versioning": true
					},
					"deployed_version": {
						"type": "url",
						"method": "GET",
						"url": "` + test.LookupBare["url_invalid"] + "/" + url.QueryEscape(`{"foo":"1.2.3-beta"}`) + `",
						"allow_invalid_certs": true,
						"json": "foo",
						"regex": "v?(\\d+)\\.(\\d+)\\.(\\d+)",
						"regex_template": "$3.$2.$1"
					}
				}`),
			serviceID: "__name__-new",
			wants: wants{
				statusCode:      http.StatusOK,
				bodyRegex:       `{"message":"created service[^}]+"}`,
				deployedVersion: "3.2.1",
				serviceYAML: test.TrimYAML(`
					comment: bye
					options:
						semantic_versioning: true
					deployed_version:
						type: url
						method: GET
						url: ` + test.LookupBare["url_invalid"] + "/" + url.QueryEscape(`{"foo":"1.2.3-beta"}`) + `
						allow_invalid_certs: true
						json: foo
						regex: 'v?(\d+)\.(\d+)\.(\d+)'
						regex_template: $3.$2.$1
				`),
			},
		},
		{
			name: "new service, ID already taken",
			payload: test.TrimJSON(`{
					"id": "__name__",
					"latest_version": {
						"type": "github",
						"url": "` + test.ArgusGitHubRepo + `"
					}
				}`),
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `{"message":"create .* failed.*"}`,
			},
		},
		{
			name: "new service, name already taken",
			payload: test.TrimJSON(`{
					"id": "__name__-new",
					"name": "__name__",
					"latest_version": {
						"type": "github",
						"url": "` + test.ArgusGitHubRepo + `"
					}
				}`),
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `{"message":"create .* failed.*"}`,
			},
		},
		{
			name: "new service, invalid interval",
			payload: test.TrimJSON(`{
				"id": "__name__-new",
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `"
				},
				"options": {
					"interval": "foo"
				}
			}`),
			serviceID: "__name__-new",
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex: `` +
					`^{"message":"create .* failed:\\n` +
					`  options:\\n` +
					`    interval: \\"foo\\"[^"]+"}`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: a service with ID: 'test name' already exists.
			svc := testService(t, tc.name, "url", "url", true)
			svc.Name = svc.ID
			tc.wants.serviceYAML = strings.ReplaceAll(tc.wants.serviceYAML, "__name__", tc.name)
			apiMu.Lock()
			api.Config.Service[svc.ID] = svc
			api.Config.Order = append(api.Config.Order, svc.ID)
			apiMu.Unlock()

			tc.payload = strings.ReplaceAll(tc.payload, "__name__", tc.name)
			payload := bytes.NewReader([]byte(tc.payload))
			var req *http.Request
			// CREATE.
			target := "/api/v1/service/new"
			req = httptest.NewRequest(http.MethodPost, target, payload)

			// WHEN: that HTTP request is sent.
			w := httptest.NewRecorder()
			apiMu.Lock()
			api.httpServiceEdit(w, req)
			apiMu.Unlock()
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			prefix := fmt.Sprintf("%s\nAPI.httpServiceEdit() (create)", packageName)

			// THEN: the expected status code is returned.
			if got, want := res.StatusCode, tc.wants.statusCode; got != want {
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
			if got := string(data); !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf(
					"%s body mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wants.bodyRegex,
				)
			}
			if tc.wants.statusCode != http.StatusOK {
				return
			}

			// AND: the service was created.
			serviceID := tc.name
			if tc.serviceID != "" {
				serviceID = strings.ReplaceAll(tc.serviceID, "__name__", tc.name)
			}
			apiMu.RLock()
			svc = api.Config.Service[serviceID]
			apiMu.RUnlock()
			if svc == nil {
				if tc.wants.latestVersion != tc.wants.deployedVersion &&
					tc.wants.latestVersion != "" {
					t.Fatalf(
						"%s service %q not created",
						prefix, serviceID,
					)
				}
				return
			}

			// AND: the service stringifies as expected.
			if got, want := svc.String(""), tc.wants.serviceYAML; got != want {
				t.Errorf(
					"%s stringified service mismatch\ngot:  %q\nwant: %q",
					prefix, got, want,
				)
			}

			// AND: the LatestVersion is expected.
			gotLV := svc.Status.LatestVersion()
			wantLV := tc.wants.latestVersion
			if !util.RegexCheck(wantLV, gotLV) {
				t.Errorf(
					"%s LatestVersion mismatch\ngot:  %q\nwant: %q",
					prefix, gotLV, tc.wants.latestVersion,
				)
			}

			// AND: the DeployedVersion is expected.
			gotDV := svc.Status.DeployedVersion()
			wantDV := tc.wants.deployedVersion
			if !util.RegexCheck(wantDV, gotDV) {
				t.Errorf(
					"%s DeployedVersion mismatch\ngot:  %q\nwant: %q",
					prefix, gotDV, tc.wants.deployedVersion,
				)
			}
		})
	}
}

func TestHTTP_ServiceEdit__Edit(t *testing.T) {
	svcCfg := svctest.PlainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()

	type wants struct {
		bodyRegex                      string
		statusCode                     int
		latestVersion, deployedVersion string
		serviceYAML                    string
	}

	// GIVEN: an API and a request to edit a service.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	apiMu := sync.RWMutex{}

	// Give time for save before TempDir clean-up.
	t.Cleanup(func() { time.Sleep(2 * config.DebounceDuration) })

	tests := []struct {
		name    string
		svc     *service.Service
		payload string
		wants   wants
	}{
		{
			name: "invalid JSON",
			svc:  testService(t, "invalid JSON", "url", "url", false),
			payload: `
				"id": "__name__-",
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `"
				`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex: `` +
					`^{"message":"edit .* failed:\\n` +
					`  unmarshal service payload:\\n` +
					`    json: cannot unmarshal[^"]+"}`,
			},
		},
		{
			name: "successful lv edit",
			svc: test.Must(t, func() (*service.Service, error) {
				return service.DecodeService(
					"yaml", []byte(test.TrimYAML(`
						name: successful lv edit
						comment: hi
						options:
							active: false
						latest_version:
							type: github
							url: "`+test.ArgusGitHubRepo+`"
							access_token: token
							require:
								regex_content: abc
								docker:
									image: foo/bar
									tag: '{{ version }}'
									auth:
										token: def
						deployed_version:
							type: url
							url: `+test.LookupBare["url_valid"]+`/1.2.3
						notify:
							discord: {}
						webhook:
							wh: {}
						dashboard:
							icon: https://example.com/icon.png
					`)),
					"successful lv edit",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			payload: `
				{
					"id": "__name__",
					"options": {
						"interval": "99m"
					},
					"latest_version": {
						"type": "url",
						"url":  "` + test.LookupBare["url_valid"] + `/version is v1.2.3",
						"allow_invalid_certs": true,
						"url_commands": [
							{
								"type": "regex",
								"regex": "v?([0-9.]+)"
							}
						]
					}
				}`,
			wants: wants{
				statusCode:      http.StatusOK,
				bodyRegex:       `{"message":"edited service[^}]+"}`,
				latestVersion:   `1\.2\.3`,
				deployedVersion: "",
				serviceYAML: test.TrimYAML(`
					options:
						interval: 99m
					latest_version:
						type: url
						url: ` + test.LookupBare["url_valid"] + `/version is v1.2.3
						url_commands:
							- type: regex
								regex: v?([0-9.]+)
						allow_invalid_certs: true
				`),
			},
		},
		{
			name: "successful dv edit",
			svc: test.Must(t, func() (*service.Service, error) {
				return service.DecodeService(
					"yaml", []byte(test.TrimYAML(`
						name: successful dv edit
						comment: hello
						options:
							active: false
						latest_version:
							type: github
							url: "`+test.ArgusGitHubRepo+`"
							access_token: token
							require:
								regex_content: abc
								docker:
									image: foo/bar
									tag: '{{ version }}'
									auth:
										token: def
						deployed_version:
							type: url
							url: `+test.LookupBare["url_invalid"]+`/v1.2.3
							allow_invalid_certs: true
							regex: v?([0-9.]+)
						notify:
							discord: {}
						webhook:
							wh: {}
						dashboard:
							icon: https://example.com/icon.png
					`)),
					"successful lv edit",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			payload: `
				{
					"id": "__name__",
					"options": {
						"interval": "99m"
					},
					"deployed_version": {
						"type": "url",
						"url":  "` + test.LookupBare["url_invalid"] + `/v1.2.3",
						"allow_invalid_certs": true,
						"regex": "v?([0-9.]+)"
					}
				}`,
			wants: wants{
				statusCode:      http.StatusOK,
				bodyRegex:       `{"message":"edited service[^}]+"}`,
				latestVersion:   "",
				deployedVersion: `1\.2\.3`,
				serviceYAML: test.TrimYAML(`
					options:
						interval: 99m
					deployed_version:
						type: url
						url: ` + test.LookupBare["url_invalid"] + `/v1.2.3
						allow_invalid_certs: true
						regex: v?([0-9.]+)
				`),
			},
		},
		{
			name: "edit service that doesn't exist",
			svc:  nil,
			payload: `
				{
					"latest_version": {
						"type": "url",
						"url":  "` + test.LookupPlain["url_valid"] + `",
						"url_commands": [
							{
								"type": "regex",
								"regex": "stable version: \"v?([0-9.]+)\""
							}
						],
					},
					"options": {
						"interval": "99m"
					}
				}`,
			wants: wants{
				statusCode: http.StatusNotFound,
				bodyRegex:  `^{"message":"edit .* failed.*"}`,
			},
		},
		{
			name: "edit that doesn't query (lv) successfully",
			svc:  testService(t, "invalid JSON", "url", "url", false),
			payload: `
				{
					"id": "__name__",
					"latest_version": {
						"type": "url",
						"url":  "` + test.LookupPlain["url_valid"] + `",
						"url_commands": [
							{
								"type": "regex",
								"regex": "stable version: \"v-([0-9.]+)\""
							}
						]
					},
					"options": {
						"interval": "99m"
					}
				}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex: `` +
					`^{"message":"edit \\"[^"]+\\" failed.*:\\n` +
					`  latest_version fetches failed:\\n` +
					`    no releases were found.*\\n` +
					`      regex \\".+\\" didn't return any matches on \\".+\\""}`,
			},
		},
		{
			name: "edit that doesn't query (dv) successfully",
			svc:  testService(t, "invalid JSON", "url", "url", false),
			payload: `
				{
					"id": "__name__",
					"deployed_version": {
						"type": "url",
						"url":  "` + test.LookupPlain["url_valid"] + `",
						"regex": "stable version: \"v-([0-9.]+)\""
					},
					"options": {
						"interval": "99m"
					}
				}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex: `` +
					`^{"message":"edit \\"[^"]+\\" failed.*:\\n` +
					`  deployed_version fetches failed:\\n` +
					`    regex .* didn't return any matches.*"}$`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			serviceID := tc.name
			tc.wants.serviceYAML = strings.ReplaceAll(tc.wants.serviceYAML, "__name__", serviceID)
			if tc.svc != nil {
				apiMu.Lock()
				tc.svc.ID = serviceID
				api.Config.Service[tc.svc.ID] = tc.svc
				api.Config.Order = append(api.Config.Order, tc.svc.ID)
				apiMu.Unlock()
			}
			tc.payload = strings.ReplaceAll(tc.payload, "__name__", serviceID)
			tc.payload = test.TrimJSON(tc.payload)
			payload := bytes.NewReader([]byte(tc.payload))
			// EDIT.
			target := "/api/v1/service/config"
			params := url.Values{}
			// Set service_id.
			params.Set("service_id", serviceID)
			req := httptest.NewRequest(http.MethodPut, target, payload)
			req.URL.RawQuery = params.Encode()

			// WHEN: that HTTP request is sent.
			w := httptest.NewRecorder()
			apiMu.Lock()
			api.httpServiceEdit(w, req)
			apiMu.Unlock()
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			prefix := fmt.Sprintf("%s\nAPI.httpServiceEdit() (edit)", packageName)

			// THEN: the expected status code is returned.
			if got, want := res.StatusCode, tc.wants.statusCode; got != want {
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
			if got := string(data); !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf(
					"%s body mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wants.bodyRegex,
				)
			}
			if tc.wants.statusCode != http.StatusOK {
				return
			}

			// AND: the service was created.
			apiMu.RLock()
			svc := api.Config.Service[serviceID]
			apiMu.RUnlock()
			if svc == nil {
				if tc.wants.latestVersion != tc.wants.deployedVersion &&
					tc.wants.latestVersion != "" {
					t.Errorf(
						"%s service %q not created",
						prefix, serviceID,
					)
				}
				return
			}

			// AND: the stringified service is expected.
			if got := svc.String(""); got != tc.wants.serviceYAML {
				t.Errorf(
					"%s stringified Service mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wants.serviceYAML,
				)
			}

			// AND: the LatestVersion is expected.
			gotLV := svc.Status.LatestVersion()
			wantLV := tc.wants.latestVersion
			if !util.RegexCheck(wantLV, gotLV) {
				t.Errorf(
					"%s LatestVersion mismatch\ngot:  %q\nwant: %q",
					prefix, gotLV, tc.wants.latestVersion,
				)
			}

			// AND: the DeployedVersion is expected.
			gotDV := svc.Status.DeployedVersion()
			wantDV := tc.wants.deployedVersion
			if !util.RegexCheck(wantDV, gotDV) {
				t.Errorf(
					"%s DeployedVersion mismatch\ngot:  %q\nwant: %q",
					prefix, gotDV, tc.wants.deployedVersion,
				)
			}
		})
	}
}

func TestHTTP_ServiceEdit__Edit__Secrets(t *testing.T) {
	svcCfg := svctest.PlainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()
	type wants struct {
		statusCode  int
		serviceYAML string
	}

	// GIVEN: an API and a request to edit a service.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	apiMu := sync.RWMutex{}

	// Give time for save before TempDir clean-up.
	t.Cleanup(func() { time.Sleep(2 * config.DebounceDuration) })

	tests := []struct {
		name    string
		svc     *service.Service
		payload string
		wants   wants
	}{
		{
			name: "lv-github",
			svc: test.Must(t, func() (*service.Service, error) {
				return service.DecodeService(
					"yaml", []byte(test.TrimYAML(`
						name: lv-github
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
							access_token: `+test.GitHubToken(t)+`
					`)),
					"lv-github",
					svcCfg,
					notifyCfg,
					whCfg,
				)
			}),
			payload: test.TrimJSON(`{
				"id": "__name__",
				"name": "__name__",
				"comment": "foo",
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `",
					"access_token": "<secret>"
				}
			}`),
			wants: wants{
				statusCode: http.StatusOK,
				serviceYAML: test.TrimYAML(`
					name: lv-github
					comment: foo
					latest_version:
						type: github
						url: ` + test.ArgusGitHubRepo + `
						access_token: ` + test.GitHubToken(t) + `
				`),
			},
		},
		{
			name: "lv-github-require",
			svc: test.Must(t, func() (*service.Service, error) {
				return service.DecodeService(
					"yaml", []byte(test.TrimYAML(`
						name: lv-github-require
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
							access_token: `+test.GitHubToken(t)+`
							require:
								docker:
									type: ghcr
									image: `+strings.ToLower(test.ArgusGitHubRepo)+`
									tag: '{{ version }}'
									auth:
										token: `+test.GitHubToken(t)+`
					`)),
					"lv-github-require",
					svcCfg,
					notifyCfg,
					whCfg,
				)
			}),
			payload: test.TrimJSON(`{
				"id": "__name__",
				"name": "__name__",
				"comment": "foo",
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `",
					"access_token": "<secret>",
					"require": {
						"docker": {
							"type": "ghcr",
							"image": "` + strings.ToLower(test.ArgusGitHubRepo) + `",
							"tag": "{{ version }}",
							"auth": {
								"token": "<secret>"
							}
						}
					}
				}
			}`),
			wants: wants{
				statusCode: http.StatusOK,
				serviceYAML: test.TrimYAML(`
					name: lv-github-require
					comment: foo
					latest_version:
						type: github
						url: ` + test.ArgusGitHubRepo + `
						require:
							docker:
								type: ghcr
								image: ` + strings.ToLower(test.ArgusGitHubRepo) + `
								tag: '{{ version }}'
								auth:
									token: ` + test.GitHubToken(t) + `
						access_token: ` + test.GitHubToken(t) + `
				`),
			},
		},
		{
			name: "lv-url",
			svc: test.Must(t, func() (*service.Service, error) {
				return service.DecodeService(
					"yaml", []byte(test.TrimYAML(`
						name: lv-url
						latest_version:
							type: url
							url: `+test.ArgusGitHubRepo+`
							headers:
								- key: X-A
								  value: a
								- key: X-B
								  value: b
								- key: X-C
								  value: c
					`)),
					"lv-github",
					svcCfg,
					notifyCfg,
					whCfg,
				)
			}),
			payload: test.TrimJSON(`{
				"id": "__name__",
				"name": "__name__",
				"comment": "foo",
				"latest_version": {
					"type": "url",
					"url": "` + test.LookupPlain["url_valid"] + `",
					"url_commands": [
						{"type": "regex", "regex": "\"(\\d+\\.\\d+\\.\\d+)\""}
					],
					"headers": [
						{"key": "X-Alpha",   "value": "<secret>", "old_index": 0},
						{"key": "X-Charlie", "value": "<secret>", "old_index": 2},
						{"key": "X-Bravo",   "value": "<secret>", "old_index": 1}
					]
				}
			}`),
			wants: wants{
				statusCode: http.StatusOK,
				serviceYAML: test.TrimYAML(`
					name: lv-github
					comment: foo
					latest_version:
						type: url
						url: ` + test.LookupPlain["url_valid"] + `
						url_commands:
							- type: regex
								regex: '"(\d+\.\d+\.\d+)"'
						headers:
							- key: X-Alpha
							  value: a
							- key: X-Charlie
							  value: c
							- key: X-Bravo
							  value: b
				`),
			},
		},
		{
			name: "dv-url",
			svc: test.Must(t, func() (*service.Service, error) {
				return service.DecodeService(
					"yaml", []byte(test.TrimYAML(`
						name: dv-url
						deployed_version:
							type: url
							method: POST
							url: `+test.LookupPlainPOST["url_valid"]+`
							body: '`+test.LookupPlainPOST["data_pass"]+`'
							regex: ver([0-9.]+)
					`)),
					"dv-url",
					svcCfg,
					notifyCfg,
					whCfg,
				)
			}),
			payload: test.TrimJSON(`{
				"id": "__name__",
				"name": "__name__",
				"comment": "foo",
				"deployed_version": {
					"type": "url",
					"method": "POST",
					"url": "` + test.LookupPlainPOST["url_valid"] + `",
					"body": "` + strings.ReplaceAll(test.LookupPlainPOST["data_pass"], `"`, `\"`) + `",
					"regex": "ver([0-9.]+)"
				}
			}`),
			wants: wants{
				statusCode: http.StatusOK,
				serviceYAML: test.TrimYAML(`
					name: dv-url
					comment: foo
					deployed_version:
						type: url
						method: POST
						url: ` + test.LookupPlainPOST["url_valid"] + `
						body: '` + test.LookupPlainPOST["data_pass"] + `'
						regex: ver([0-9.]+)
				`),
			},
		},
		{
			name: "notify",
			svc: test.Must(t, func() (*service.Service, error) {
				return service.DecodeService(
					"yaml", []byte(test.TrimYAML(`
						name: notify
						deployed_version:
							type: manual
						notify:
							test:
								type: gotify
								url_fields:
									host: https://example.com
									token: abc
					`)),
					"notify",
					svcCfg,
					notifyCfg,
					whCfg,
				)
			}),
			payload: test.TrimJSON(`{
				"id": "__name__",
				"name": "__name__",
				"comment": "foo",
				"deployed_version": {
					"type": "manual"
				},
				"notify": [
					{
						"name": "test",
						"old_index": "test",
						"type": "gotify",
						"url_fields": {
							"host": "https://example.com",
							"token": "<secret>"
						}
					}
				]
			}`),
			wants: wants{
				statusCode: http.StatusOK,
				serviceYAML: test.TrimYAML(`
					name: notify
					comment: foo
					deployed_version:
						type: manual
					notify:
						test:
							type: gotify
							url_fields:
								host: example.com
								token: abc
				`),
			},
		},
		{
			name: "webhook",
			svc: test.Must(t, func() (*service.Service, error) {
				return service.DecodeService(
					"yaml", []byte(test.TrimYAML(`
						name: webhook
						deployed_version:
							type: manual
						webhook:
							test:
								type: github
								url: `+test.WebHookGitHub["url_valid"]+`
								secret: `+test.WebHookGitHub["secret_pass"]+`
					`)),
					"webhook",
					svcCfg,
					notifyCfg,
					whCfg,
				)
			}),
			payload: test.TrimJSON(`{
				"id": "__name__",
				"name": "__name__",
				"comment": "foo",
				"deployed_version": {
					"type": "manual"
				},
				"webhook": [
					{
						"name": "test",
						"old_index": "test",
						"type": "github",
						"url": "` + test.WebHookGitHub["url_valid"] + `",
						"secret": "<secret>"
					}
				]
			}`),
			wants: wants{
				statusCode: http.StatusOK,
				serviceYAML: test.TrimYAML(`
					name: webhook
					comment: foo
					deployed_version:
						type: manual
					webhook:
						test:
							type: github
							url: ` + test.WebHookGitHub["url_valid"] + `
							secret: ` + test.WebHookGitHub["secret_pass"] + `
				`),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			serviceID := tc.svc.ID
			tc.wants.serviceYAML = strings.ReplaceAll(tc.wants.serviceYAML, "__name__", tc.name)
			apiMu.Lock()
			api.Config.Service[serviceID] = tc.svc
			api.Config.Order = append(api.Config.Order, serviceID)
			apiMu.Unlock()
			tc.payload = strings.ReplaceAll(tc.payload, "__name__", serviceID)
			tc.payload = test.TrimJSON(tc.payload)
			payload := bytes.NewReader([]byte(tc.payload))
			// EDIT.
			target := "/api/v1/service/config"
			params := url.Values{}
			// Set service_id.
			params.Set("service_id", serviceID)
			req := httptest.NewRequest(http.MethodPut, target, payload)
			req.URL.RawQuery = params.Encode()

			// WHEN: that HTTP request is sent.
			w := httptest.NewRecorder()
			apiMu.Lock()
			api.httpServiceEdit(w, req)
			apiMu.Unlock()
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			prefix := fmt.Sprintf("%s\nAPI.httpServiceEdit() (edit)", packageName)

			// THEN: the expected status code is returned.
			if got, want := res.StatusCode, tc.wants.statusCode; got != want {
				data, _ := io.ReadAll(res.Body)
				t.Fatalf(
					"%s status code mismatch\ngot:  %d\nwant: %d\n%q",
					prefix, got, want, string(data),
				)
			}

			// AND: the service was created.
			apiMu.RLock()
			svc := api.Config.Service[serviceID]
			apiMu.RUnlock()
			if svc == nil {
				t.Fatalf(
					"%s service %q not created",
					prefix, serviceID,
				)
			}

			// AND: the stringified service is expected.
			got := svc.String("")
			if got != tc.wants.serviceYAML {
				tokenReplacer := regexp.MustCompile(`(access_token|secret|token): [^\s<]+`)
				tokenReplacement := "$1: SECRET"
				got = tokenReplacer.ReplaceAllString(got, tokenReplacement)
				tc.wants.serviceYAML = tokenReplacer.ReplaceAllString(tc.wants.serviceYAML, tokenReplacement)

				t.Errorf(
					"%s stringified Service mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wants.serviceYAML,
				)
			}
		})
	}
}

func TestHTTP_ServiceDelete(t *testing.T) {
	type wants struct {
		bodyRegex  string
		statusCode int
	}

	// GIVEN: an API and a request to delete a service.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	t.Cleanup(func() {
		// Give time for save before TempDir clean-up.
		time.Sleep(2 * config.DebounceDuration)
	})
	svc := testService(t, "TestHTTP_ServiceDelete", "url", "url", true)
	svc.HardDefaults.Status.DatabaseChannel = api.Config.DatabaseChannel
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
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:      "delete service",
			serviceID: svc.ID,
			wants: wants{
				bodyRegex:  `{"message":"deleted service[^}]+"}`,
				statusCode: http.StatusOK,
			},
		},
		{
			name:      "delete service again",
			serviceID: svc.ID,
			wants: wants{
				bodyRegex:  `{"message":"delete .* failed, service not found"`,
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:      "no service_id provided",
			serviceID: "",
			wants: wants{
				bodyRegex:  `{"message":"missing required query parameter: service_id"}`,
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() -- Cannot run in parallel since we're sharing the API.

			target := "/api/v1/service/delete"
			params := url.Values{}
			// Set service_id.
			params.Set("service_id", tc.serviceID)

			// WHEN: that HTTP request is sent.
			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			api.httpServiceDelete(w, req)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			prefix := fmt.Sprintf("%s\nAPI.httpServiceDelete()", packageName)

			// THEN: the expected status code is returned.
			if got, want := res.StatusCode, tc.wants.statusCode; got != want {
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
			if got := string(data); !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf(
					"%s body mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wants.bodyRegex,
				)
			}

			// AND: the service is removed from the config.
			if api.Config.Service[tc.serviceID] != nil {
				t.Errorf(
					"%s\nservice %q not removed from Config.Service[]",
					packageName, tc.serviceID,
				)
			}
			if util.Contains(api.Config.Order, tc.serviceID) {
				t.Errorf(
					"%s\nservice %q not removed from Order",
					packageName, tc.serviceID,
				)
			}

			// AND: the service is removed from the database (if the req was OK).
			if tc.wants.statusCode == http.StatusOK {
				time.Sleep(time.Second)
				if len(api.Config.DatabaseChannel) == 0 {
					t.Errorf(
						"%s service %q not removed from database",
						prefix, tc.serviceID,
					)
				} else {
					msg := <-api.Config.DatabaseChannel
					if msg.Delete != true {
						t.Errorf(
							"%s should have sent a deletion to the db, not\n%+v",
							prefix, msg,
						)
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

	// GIVEN: an API and a request to test a notify.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)

	validNotify := shoutrrrtest.Shoutrrr(false, false)
	api.Config.Notify = shoutrrr.ShoutrrrsDefaults{}
	options := util.CopyMap(validNotify.Options)
	params := util.CopyMap(validNotify.Params)
	urlFields := util.CopyMap(validNotify.URLFields)
	api.Config.Notify["test"] = shoutrrr.NewDefaults(
		"gotify",
		options, urlFields, params,
	)
	api.Config.Service["test"].Notify = map[string]*shoutrrr.Shoutrrr{
		"test":    shoutrrrtest.Shoutrrr(false, false),
		"no_main": shoutrrrtest.Shoutrrr(false, false),
	}
	tests := []struct {
		name        string
		queryParams map[string]string
		payload     string
		wants       wants
	}{
		{
			name: "body too large",
			payload: `{
				"test": "` + strings.Repeat(strings.Repeat("abcdefghijklmnopqrstuvwxyz", 100), 100) + `"}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  "request body too large",
			},
		},
		{
			name: "no bodyRegex",
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  "name and/or name_previous are required",
			},
		},
		{
			name: "no service, new notify",
			payload: `{
				"name": "new_notify"}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `invalid type "[^"]+"`,
			},
		},
		{
			name: "new service, no new/old notify",
			payload: `{
				"service_id": "new_service"}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `name and/or name_previous are required`,
			},
		},
		{
			name: "new service, no main",
			payload: `{
				"service_id": "also_unknown",
				"name": "test_notify",
				"type": "ntfy"}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  "url_fields:[^ ]+ +topic: .*required",
			},
		},
		{
			name: "new service, no main - no service_id",
			payload: `{
				"name": "test_notify",
				"type": "ntfy"}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  "url_fields:[^ ]+ +topic: .*required",
			},
		},
		{
			name: "new service, no main - invalid JSON, options",
			payload: `{
				"service_id": "also_unknown",
				"name": "test_notify",
				"type": "ntfy",
				"options": {
					"delay": "1s",
					"something" "else"
				}
			}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex: test.TrimYAML(`
					failed to unmarshal payload:
						jsontext: invalid character.*
				`),
			},
		},
		{
			name: "new service, no main - options, invalid",
			payload: `{
				"service_id": "also_unknown",
				"name": "test_notify",
				"type": "ntfy",
				"options": {
					"delay": "time"
				}
			}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `options:[^ ]+  delay: "[^"]+" <invalid>`,
			},
		},
		{
			name: "new service, have main - options, applied, delay ignored",
			payload: `{
				"service_id": "also_unknown",
				"name": "test",
				"options": {
					"delay": "24h"
				}
			}`,
			wants: wants{
				statusCode: http.StatusOK,
				bodyRegex:  "message sent",
			},
		},
		{
			name: "new service, no main - invalid JSON, url_fields",
			payload: `{
				"service_id": "also_unknown",
				"name": "test_notify",
				"type": "ntfy",
				"url_fields": {
					"host" "example.com"
				}
			}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex: test.TrimYAML(`
					failed to unmarshal payload:
						jsontext: invalid character.*
				`),
			},
		},
		{
			name: "new service, have main - url_fields, invalid",
			payload: `{
				"service_id": "also_unknown",
				"name": "test",
				"url_fields": {
					"port": "number"
				}
			}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `failed to parse URL`,
			},
		},
		{
			name: "new service, no main - no type",
			payload: `{
				"service_id": "also_unknown",
				"name": "test_notify"}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `invalid type "test_notify"`,
			},
		},
		{
			name: "new service, no main - unknown type",
			payload: `{
				"service_id": "unknown",
				"name": "test_notify",
				"type": "something"}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `invalid type "something"`,
			},
		},
		{
			name: "new service, no main - type from ID",
			payload: `{
				"service_id": "unknown",
				"name": "` + validNotify.Type + `",
				"url_fields": {
					"host": "` + validNotify.URLFields["host"] + `",
					"path": "` + validNotify.URLFields["path"] + `",
					"token": "` + validNotify.URLFields["token"] + `"
				}
			}`,
			wants: wants{
				statusCode: http.StatusOK,
				bodyRegex:  "message sent",
			},
		},
		{
			name: "new service, have main - type from Main",
			payload: `{
				"service_id": "unknown",
				"name": "test",
				"url_fields": {
					"host": "` + validNotify.URLFields["host"] + `",
					"path": "` + validNotify.URLFields["path"] + `",
					"token": "` + validNotify.URLFields["token"] + `"
				}
			}`,
			wants: wants{
				statusCode: http.StatusOK,
				bodyRegex:  "message sent",
			},
		},
		{
			name: "same service, have main - type from original",
			payload: `{
				"service_id_previous": "test",
				"service_id": "test",
				"name": "new_notify",
				"name_previous": "test",
				"url_fields": {
					"host": "` + validNotify.URLFields["host"] + `",
					"path": "` + validNotify.URLFields["path"] + `",
					"token": "` + util.SecretValue + `"
				}
			}`,
			wants: wants{
				statusCode: http.StatusOK,
				bodyRegex:  "message sent",
			},
		},
		{
			name: "same service, no main - can remove vars",
			payload: `{
				"service_id_previous": "test",
				"service_id": "test",
				"name": "new_notify",
				"name_previous": "no_main",
				"url_fields": {
					"host": "` + validNotify.URLFields["host"] + `",
					"path": "` + validNotify.URLFields["path"] + `",
					"token": ""
				}
			}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex: test.TrimYAML(`
					^url_fields:
						token: <required>.*$`,
				),
			},
		},
		{
			name: "same service, no main - unsent vars inherited",
			payload: `{
				"service_id_previous": "test",
				"service_id": "test",
				"name": "new_notify",
				"name_previous": "no_main",
				"type": "` + validNotify.Type + `",
				"url_fields": {
					"host": "` + validNotify.URLFields["host"] + `",
					"path": "` + validNotify.URLFields["path"] + `"
				}
			}`,
			wants: wants{
				statusCode: http.StatusOK,
				bodyRegex:  "message sent",
			},
		},
		{
			name: "same service, have main - fail send",
			payload: `{
				"service_id_previous": "test",
				"service_id": "test",
				"name": "test",
				"name_previous": "test",
				"type": "` + validNotify.Type + `",
				"url_fields": {
					"host": "` + validNotify.URLFields["host"] + `",
					"path": "` + validNotify.URLFields["path"] + `",
					"token": "invalid"
				}
			}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  "invalid .* token",
			},
		},
		{
			name: "same service, have main - new name, also fail send",
			payload: `{
				"service_id_previous": "test",
				"service_id": "new_name",
				"name": "test",
				"name_previous": "test",
				"type": "` + validNotify.Type + `",
				"url_fields": {
					"host": "` + validNotify.URLFields["host"] + `",
					"path": "` + validNotify.URLFields["path"] + `",
					"token": "invalid"
				}
			}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  "invalid .* token",
			},
		},
		{
			name: "service_id_previous that doesn't exist",
			payload: `{
				"service_id_previous": "does_not_exist",
				"service_id": "test",
				"name": "new_notify",
				"name_previous": "test",
				"url_fields": {
					"host": "` + validNotify.URLFields["host"] + `",
					"path": "` + validNotify.URLFields["path"] + `",
					"token": "` + util.SecretValue + `"
				}
			}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `invalid type "new_notify"`,
			},
		},
		{
			name: "name_previous that doesn't exist",
			payload: `{
				"service_id_previous": "test",
				"service_id": "test",
				"name": "new_notify",
				"name_previous": "does_not_exist",
				"url_fields": {
					"host": "` + validNotify.URLFields["host"] + `",
					"path": "` + validNotify.URLFields["path"] + `",
					"token": "` + util.SecretValue + `"
				}
			}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `invalid type "new_notify"`,
			},
		},
		{
			name: "service_id_previous and name_previous that doesn't exist",
			payload: `{
				"service_id_previous": "does_not_exist",
				"service_id": "test",
				"name": "new_notify",
				"name_previous": "also_does_not_exist",
				"url_fields": {
					"host": "` + validNotify.URLFields["host"] + `",
					"path": "` + validNotify.URLFields["path"] + `",
					"token": "` + util.SecretValue + `"
				}
			}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `invalid type "new_notify"`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.wants.bodyRegex = util.ValueOr(tc.wants.bodyRegex, `^$`)
			tc.payload = test.TrimJSON(tc.payload)
			payload := bytes.NewReader([]byte(tc.payload))

			// WHEN: that request is sent.
			req := httptest.NewRequest(http.MethodGet, "/api/v1/notify/test", payload)
			w := httptest.NewRecorder()
			api.httpNotifyTest(w, req)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			prefix := fmt.Sprintf("%s\nAPI.httpNotifyTest()", packageName)

			// THEN: the expected status code is returned.
			if got, want := res.StatusCode, tc.wants.statusCode; got != want {
				t.Errorf(
					"%s status code mismatch\ngot:  %d\nwant: %d",
					prefix, got, want,
				)
			}

			// AND: the expected message is contained in the bodyRegex.
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("%s unexpected error:\n%v", packageName, err)
			}
			// Marshal message out of JSON data {"message": text}.
			var body map[string]string
			_ = decode.Unmarshal("json", data, &body)
			if got := body["message"]; !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf(
					"%s body mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wants.bodyRegex,
				)
			}
		})
	}
}
