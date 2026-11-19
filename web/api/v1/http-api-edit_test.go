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
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/release-argus/Argus/auth/rbac"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/dashboard"
	dvweb "github.com/release-argus/Argus/service/deployed_version/types/web"
	"github.com/release-argus/Argus/service/shared"
	svctest "github.com/release-argus/Argus/service/test"
	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
	whtest "github.com/release-argus/Argus/webhook/test"
)

func TestHTTP_LatestVersionRefreshUncreated(t *testing.T) {
	type wants struct {
		bodyRegex  string
		statusCode int
	}

	file := t.Name() + ".yml"
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
				"overrides": `"type": "url", "url": "` + test.LookupPlain.URLValid + `"}`,
			},
			wants: wants{
				bodyRegex: `` +
					`^{"message":"latest_version:\\n` +
					`  .* unmarshal[^"]+"}$`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid vars/Decode fail",
			params: map[string]string{
				"overrides": `{"type": "something"}`,
			},
			wants: wants{
				bodyRegex:  `"message":".*type: .*invalid`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid vars/CheckValues fail",
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"type":         "url",
					"url":          "` + test.LookupPlain.URLValid + `",
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
					"url":          "` + test.LookupPlain.URLValid + `",
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
					"url":          "` + test.LookupPlain.URLInvalid + `",
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

			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			api.httpLatestVersionRefreshUncreated(w, req)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			if got, want := res.StatusCode, tc.wants.statusCode; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}

			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}
			if got := string(data); !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, tc.wants.bodyRegex)
			}
		})
	}
}

func TestHTTP_DeployedVersionRefreshUncreated(t *testing.T) {
	type wants struct {
		bodyRegex  string
		statusCode int
	}

	file := t.Name() + ".yml"
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
				"overrides": `"type": "url", "url": "` + test.LookupPlain.URLValid + `"}`,
			},
			wants: wants{
				bodyRegex: `` +
					`^{"message":"deployed_version:\\n` +
					`  .* unmarshal[^"]+"}$`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid vars/Decode fail",
			params: map[string]string{
				"overrides": `{"type": "something"}`,
			},
			wants: wants{
				bodyRegex:  `"message":".*type: .*invalid`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid vars/CheckValues fail",
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"type":  "url",
					"url":   "` + test.LookupPlain.URLValid + `",
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
					"url":   "` + test.LookupPlain.URLValid + `",
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
					"url":   "` + test.LookupPlain.URLInvalid + `",
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

			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			api.httpDeployedVersionRefreshUncreated(w, req)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			if got, want := res.StatusCode, tc.wants.statusCode; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}

			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}
			if got := string(data); !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, tc.wants.bodyRegex)
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

	file := t.Name() + ".yml"
	api := testAPI(t, file)
	apiMu := sync.RWMutex{}
	headerAuthURL := test.HeaderAuthServer(t, "1.2.3")

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
							url: `+test.LookupBare.URLValid+`/ver1.2.3
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
			name: "semantic_versioning/null, fail as default=true",
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
							url: `+test.LookupBare.URLValid+`/ver1.2.3
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
			name: "semantic_versioning/same refreshes service",
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
							url: `+test.LookupBare.URLValid+`/ver1.2.3
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
			name: "semantic_versioning/diff, not applied to service",
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
							url: `+test.LookupBare.URLValid+`/ver1.2.3
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
							url: `+test.LookupBare.URLValid+`/v1.2.3-beta
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
			serviceID: new("bash-bosh"),
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
			serviceID: new(""),
			wants: wants{
				bodyRegex:     `{"message":"missing required query parameter: service_id"}`,
				statusCode:    http.StatusBadRequest,
				latestVersion: "",
			},
		},
		{
			name: "secretRefs/used",
			params: map[string]string{
				"overrides": test.TrimJSON(`{
						"headers": [
							{
								"key": "` + test.LookupWithHeaderAuth.HeaderKey + `",
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
							url: `+headerAuthURL+`
							headers:
								- key: `+test.LookupWithHeaderAuth.HeaderKey+`
									value: `+test.LookupWithHeaderAuth.HeaderValuePass+`
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
			name: "secretRefs/invalid",
			params: map[string]string{
				"overrides": test.TrimJSON(`{
						"headers": [
							{
								"key": "` + test.LookupWithHeaderAuth.HeaderKey + `",
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
							url: `+headerAuthURL+`
							headers:
								- key: `+test.LookupWithHeaderAuth.HeaderKey+`
									value: `+test.LookupWithHeaderAuth.HeaderValuePass+`
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
			params := url.Values{}
			for k, v := range tc.params {
				params.Set(k, v)
			}
			serviceID := util.DerefOr(tc.serviceID, svc.ID)
			params.Set("service_id", serviceID)

			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			apiMu.Lock()
			api.httpLatestVersionRefresh(w, req)
			apiMu.Unlock()
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			if got, want := res.StatusCode, tc.wants.statusCode; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}

			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}
			if got := string(data); !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, tc.wants.bodyRegex)
			}

			// Moves only on an override-free refresh matching the service's semantic_versioning.
			gotLV := svc.Status.LatestVersion()
			wantLV := tc.wants.latestVersion
			if gotLV != wantLV {
				t.Errorf("Status.LatestVersion() mismatch\ngot:  %q\nwant: %q", gotLV, wantLV)
			}

			// With no deployed_version, webhook or command, a changed version also sets this.
			gotDV := svc.Status.DeployedVersion()
			wantDV := tc.wants.deployedVersion
			if gotDV != wantDV {
				t.Errorf("Status.DeployedVersion() mismatch\ngot:  %q\nwant: %q", gotDV, wantDV)
			}

			wantAnnounces := 0
			if tc.wants.announce {
				wantAnnounces = 1
			}
			if gotAnnounces := len(svc.Status.AnnounceChannel); gotAnnounces != wantAnnounces {
				t.Errorf(
					"Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					gotAnnounces, wantAnnounces,
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

	file := t.Name() + ".yml"
	api := testAPI(t, file)
	apiMu := sync.RWMutex{}
	headerAuthURL := test.HeaderAuthServer(t, "1.2.3")

	tests := []struct {
		name               string
		serviceID          *string
		svc                *service.Service
		nilDeployedVersion bool
		params             map[string]string
		wants              wants
	}{
		{
			name:               "uncreated/success",
			nilDeployedVersion: true,
			params: map[string]string{
				"overrides": test.TrimJSON(`{
					"type":                "url",
					"url":                 "` + test.LookupBare.URLInvalid + "/" + url.QueryEscape(`{"foo":"ver1.2.3-beta"}`) + `",
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
			name:               "uncreated/no overrides",
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
							url: `+test.LookupBare.URLValid+`/ver1.2.3
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
			name: "semantic_versioning/null, fail as default=true",
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
							url: `+test.LookupBare.URLValid+`/ver1.2.3
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
			name: "semantic_versioning/diff, not applied to service",
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
							url: `+test.LookupBare.URLValid+`/ver1.2.3
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
							url: `+test.LookupBare.URLValid+"/"+url.QueryEscape(`{"foo":"ver1.2.3-beta","bar":"ver1.2.3-beta"}`)+`
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
			name: "existing DVL, invalid JSON",
			params: map[string]string{
				"overrides": `{"json": "x.y"}`,
			},
			wants: wants{
				bodyRegex:  `^{"message":"failed to unmarshal response from .*"}$`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:               "new DVL, JSON key not found",
			nilDeployedVersion: true,
			params: map[string]string{
				"overrides": `{
					"type": "url",
					"method": "GET",
					"url": "` + test.LookupJSON.URLValid + `",
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
			name: "invalid vars, CheckValues fail",
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
			serviceID: new("bish-bash-bosh"),
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
			serviceID: new(""),
			wants: wants{
				bodyRegex:       `{"message":"missing required query parameter: service_id"}`,
				statusCode:      http.StatusBadRequest,
				deployedVersion: "",
			},
		},
		{
			name: "secretRefs/used",
			svc: test.Must(t, func() (*service.Service, error) {
				base := testService(t, "__name__", "url", "url", false)
				dv, ok := base.DeployedVersionLookup.(*dvweb.Lookup)
				if !ok {
					return nil, fmt.Errorf(
						"Service.DeployedVersionLookup type mismatch\ngot:  %T\nwant: *web.Lookup",
						base.DeployedVersionLookup,
					)
				}
				dv.URL = headerAuthURL
				dv.Headers = shared.Headers{
					{Key: test.LookupWithHeaderAuth.HeaderKey, Value: test.LookupWithHeaderAuth.HeaderValuePass},
				}
				return base, nil
			}),
			params: map[string]string{
				"overrides": test.TrimJSON(`{
						"headers": [
							{
								"key": "` + test.LookupWithHeaderAuth.HeaderKey + `",
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
			name: "secretRefs/invalid",
			svc: test.Must(t, func() (*service.Service, error) {
				base := testService(t, "__name__", "url", "url", false)
				dv, ok := base.DeployedVersionLookup.(*dvweb.Lookup)
				if !ok {
					return nil, fmt.Errorf(
						"Service.DeployedVersionLookup type mismatch\ngot:  %T\nwant: *web.Lookup",
						base.DeployedVersionLookup,
					)
				}
				dv.URL = headerAuthURL
				dv.Headers = shared.Headers{
					{Key: test.LookupWithHeaderAuth.HeaderKey, Value: test.LookupWithHeaderAuth.HeaderValuePass},
				}
				return base, nil
			}),
			params: map[string]string{
				"overrides": test.TrimJSON(`{
						"headers": [
							{
								"key": "` + test.LookupWithHeaderAuth.HeaderKey + `",
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
			params := url.Values{}
			for k, v := range tc.params {
				params.Set(k, v)
			}
			serviceID := util.DerefOr(tc.serviceID, svc.ID)
			params.Set("service_id", serviceID)

			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			apiMu.Lock()
			api.httpDeployedVersionRefresh(w, req)
			apiMu.Unlock()
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			if got, want := res.StatusCode, tc.wants.statusCode; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}

			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}
			if got := string(data); !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, tc.wants.bodyRegex)
			}

			if got, want := svc.Status.LatestVersion(), initialLatestVersion; got != want {
				t.Errorf("Status.LatestVersion() mismatch\ngot:  %q\nwant: %q", got, want)
			}

			// Only a refresh without overrides moves it.
			if got, want := svc.Status.DeployedVersion(), tc.wants.deployedVersion; got != want {
				t.Errorf("Status.DeployedVersion() mismatch\ngot:  %q\nwant: %q", got, want)
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

	file := t.Name() + ".yml"
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
							url: `+test.LookupBare.URLValid+`/ver1.2.3
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
						"url": "` + test.LookupBare.URLValid + `/ver1.2.3"
					}
				}`),
				statusCode: http.StatusOK,
			},
		},
		{
			name:      "unknown service",
			serviceID: new("bish-bash-bosh"),
			wants: wants{
				bodyRegex:  `{"message":"service .+ not found"`,
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:      "no service_id provided",
			serviceID: new(""),
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
			serviceID := util.DerefOr(tc.serviceID, svc.ID)
			params.Set("service_id", serviceID)

			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			apiMu.RLock()
			api.httpServiceDetail(w, req)
			apiMu.RUnlock()
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			if got, want := res.StatusCode, tc.wants.statusCode; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}

			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}
			if got := string(data); !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, tc.wants.bodyRegex)
			}
		})
	}
}

func TestHTTP_OtherServiceDetails(t *testing.T) {
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
			target := "/api/v1/service/defaults"

			req := httptest.NewRequest(http.MethodGet, target, nil)
			w := httptest.NewRecorder()
			api.httpOtherServiceDetails(w, req)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			if got, want := res.StatusCode, tc.wantStatusCode; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}

			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}
			got := string(data)
			if !util.RegexCheck(tc.wantBody, got) {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, tc.wantBody)
			}
		})
	}
}

func TestHTTP_TemplateParse(t *testing.T) {
	type wants struct {
		bodyRegex  string
		statusCode int
	}

	file := t.Name() + ".yml"
	api := testAPI(t, file)
	apiMu := sync.RWMutex{}

	testSVC := testService(t, t.Name(), "url", "url", true)
	apiMu.Lock()
	api.Config.Service[testSVC.ID] = testSVC
	apiMu.Unlock()

	tests := []struct {
		name        string
		queryParams map[string]string
		wants       wants
	}{
		{
			name:        "missing required parameters/both",
			queryParams: map[string]string{},
			wants: wants{
				bodyRegex:  `{"message":"missing required query parameter: service_id"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "missing required parameters/service_id",
			queryParams: map[string]string{
				"template": "{{ service_name }}",
			},
			wants: wants{
				bodyRegex:  `{"message":"missing required query parameter: service_id"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "missing required parameters/template",
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
					`  jsontext: invalid character '}'(?:[^\\]|\\")*"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "valid template/default parameters",
			queryParams: map[string]string{
				"service_id": testSVC.ID,
				"template":   `{{service_name }} - {{ version }}`,
			},
			wants: wants{
				bodyRegex: fmt.Sprintf(
					`{"parsed":"%s - %s"}`,
					regexp.QuoteMeta(testSVC.GetName()), regexp.QuoteMeta(testSVC.Status.LatestVersion()),
				),
				statusCode: http.StatusOK,
			},
		},
		{
			name: "valid template/overridden parameters",
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
				// An unknown service parses with empty defaults, not a 404.
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

			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			api.httpTemplateParse(w, req)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			if got, want := res.StatusCode, tc.wants.statusCode; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}

			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}
			if got := string(data); !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, tc.wants.bodyRegex)
			}
		})
	}
}

func TestHTTP_ServiceEdit__routeValidation(t *testing.T) {
	file := t.Name() + ".yml"
	api := testAPI(t, file)

	tests := []struct {
		name          string
		handler       func(http.ResponseWriter, *http.Request)
		serviceID     string
		wantBodyRegex string
	}{
		{
			name:          "create/rejects a service_id (would be an edit)",
			handler:       api.httpServiceCreate,
			serviceID:     "existing",
			wantBodyRegex: "service_id must be empty",
		},
		{
			name:          "update/rejects a missing service_id (would be a create)",
			handler:       api.httpServiceUpdate,
			serviceID:     "",
			wantBodyRegex: "service_id is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			target := "/api/v1/service/config"
			if tc.serviceID != "" {
				target += "?service_id=" + url.QueryEscape(tc.serviceID)
			}
			req := httptest.NewRequest(http.MethodPut, target, strings.NewReader(`{}`))
			w := httptest.NewRecorder()

			tc.handler(w, req)

			if got, want := w.Code, http.StatusBadRequest; got != want {
				t.Fatalf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
			if got := w.Body.String(); !util.RegexCheck(tc.wantBodyRegex, got) {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, tc.wantBodyRegex)
			}
		})
	}
}

func TestHTTP_ServiceEdit__create(t *testing.T) {
	type wants struct {
		bodyRegex                                string
		statusCode                               int
		latestVersionRegex, deployedVersionRegex string
		serviceYAML                              string
	}

	file := t.Name() + ".yml"
	api := testAPI(t, file)
	apiMu := sync.RWMutex{}

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
			name: "no id",
			payload: test.TrimJSON(`{
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `"
				}
			}`),
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex: `` +
					`^{"message":"create .* failed:\\n` +
					`  id: <required>"}$`,
			},
		},
		{
			name: "empty id",
			payload: test.TrimJSON(`{
				"id": "",
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `"
				}
			}`),
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex: `` +
					`^{"message":"create .* failed:\\n` +
					`  id: <required>"}$`,
			},
		},
		{
			name: "lv-github",
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
				statusCode:           http.StatusOK,
				bodyRegex:            `^{"message":"created service[^}]+"}`,
				latestVersionRegex:   `^[0-9.]+$`,
				deployedVersionRegex: `^[0-9.]+$`,
				// GitHub has no allow_invalid_certs, so it is dropped.
				serviceYAML: test.TrimYAML(`
					name: __name__-foo
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
			name: "lv-url",
			payload: test.TrimJSON(`{
				"id": "__name__-new",
				"comment": "goodbye",
				"options": {
					"semantic_versioning": false
				},
				"latest_version": {
					"type": "url",
					"url": "` + test.LookupBare.URLInvalid + "/" + url.QueryEscape(`versions here: "ver1.2.3", release=1.2.3.exe`) + `",
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
				statusCode:           http.StatusOK,
				bodyRegex:            `{"message":"created service[^}]+"}`,
				latestVersionRegex:   `^1\.2\.3$`,
				deployedVersionRegex: `^1\.2\.3$`,
				serviceYAML: test.TrimYAML(`
					comment: goodbye
					options:
						semantic_versioning: false
					latest_version:
						type: url
						url: ` + test.LookupBare.URLInvalid + "/" + url.QueryEscape(`versions here: "ver1.2.3", release=1.2.3.exe`) + `
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
			name: "dv-manual",
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
				statusCode:           http.StatusOK,
				bodyRegex:            `{"message":"created service[^}]+"}`,
				latestVersionRegex:   `^$`,
				deployedVersionRegex: `^1\.2\.3$`,
				serviceYAML: test.TrimYAML(`
					name: __name__-foo
					comment: hi
					options:
						semantic_versioning: true
					deployed_version:
						type: manual
				`),
			},
		},
		{
			name: "dv-url",
			payload: test.TrimJSON(`{
					"id": "__name__-new",
					"comment": "bye",
					"options": {
						"semantic_versioning": true
					},
					"deployed_version": {
						"type": "url",
						"method": "GET",
						"url": "` + test.LookupBare.URLInvalid + "/" + url.QueryEscape(`{"foo":"1.2.3-beta"}`) + `",
						"allow_invalid_certs": true,
						"json": "foo",
						"regex": "v?(\\d+)\\.(\\d+)\\.(\\d+)",
						"regex_template": "$3.$2.$1"
					}
				}`),
			serviceID: "__name__-new",
			wants: wants{
				statusCode:           http.StatusOK,
				bodyRegex:            `{"message":"created service[^}]+"}`,
				latestVersionRegex:   `^$`,
				deployedVersionRegex: `^3\.2\.1$`,
				serviceYAML: test.TrimYAML(`
					comment: bye
					options:
						semantic_versioning: true
					deployed_version:
						type: url
						method: GET
						url: ` + test.LookupBare.URLInvalid + "/" + url.QueryEscape(`{"foo":"1.2.3-beta"}`) + `
						allow_invalid_certs: true
						json: foo
						regex: 'v?(\d+)\.(\d+)\.(\d+)'
						regex_template: $3.$2.$1
				`),
			},
		},
		{
			name: "ID already taken",
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
			name: "name already taken",
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
			name: "invalid interval",
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

			svc := testService(t, tc.name, "url", "url", true)
			svc.Name = svc.ID
			tc.wants.serviceYAML = strings.ReplaceAll(tc.wants.serviceYAML, "__name__", tc.name)
			api.Config.OrderMu.Lock()
			api.Config.Service[svc.ID] = svc
			api.Config.Order = append(api.Config.Order, svc.ID)
			api.Config.OrderMu.Unlock()

			tc.payload = strings.ReplaceAll(tc.payload, "__name__", tc.name)
			payload := bytes.NewReader([]byte(tc.payload))
			req := httptest.NewRequest(http.MethodPost, "/api/v1/service/new", payload)

			w := httptest.NewRecorder()
			apiMu.Lock()
			api.httpServiceEdit(w, req, actionCreate)
			apiMu.Unlock()
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			if got, want := res.StatusCode, tc.wants.statusCode; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}

			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}
			if got := string(data); !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, tc.wants.bodyRegex)
			}
			if tc.wants.statusCode != http.StatusOK {
				return
			}

			serviceID := tc.name
			if tc.serviceID != "" {
				serviceID = strings.ReplaceAll(tc.serviceID, "__name__", tc.name)
			}
			api.Config.OrderMu.RLock()
			svc = api.Config.Service[serviceID]
			api.Config.OrderMu.RUnlock()
			if svc == nil {
				t.Fatalf("Config.Service[%q] mismatch\ngot:  nil\nwant: non-nil", serviceID)
			}

			if got, want := svc.String(""), tc.wants.serviceYAML; got != want {
				t.Errorf("Service.String() mismatch\ngot:  %q\nwant: %q", got, want)
			}

			gotLV := svc.Status.LatestVersion()
			wantLV := tc.wants.latestVersionRegex
			if !util.RegexCheck(wantLV, gotLV) {
				t.Errorf("Status.LatestVersion() mismatch\ngot:  %q\nwant: %q", gotLV, wantLV)
			}

			// Without a deployed_version, it copies LatestVersion().
			gotDV := svc.Status.DeployedVersion()
			wantDV := tc.wants.deployedVersionRegex
			if !util.RegexCheck(wantDV, gotDV) {
				t.Errorf("Status.DeployedVersion() mismatch\ngot:  %q\nwant: %q", gotDV, wantDV)
			}
		})
	}
}

func TestHTTP_ServiceEdit__create__concurrentConflict(t *testing.T) {
	file := t.Name() + ".yml"
	api := testAPI(t, file)
	const serviceID = "TestHTTP_ServiceEdit_createConflict"

	held := api.acquireServiceOp(serviceID)
	held.mu.Lock()
	t.Cleanup(func() {
		held.mu.Unlock()
		api.releaseServiceOp(serviceID, held)
	})

	payload := bytes.NewReader([]byte(test.TrimJSON(`{
		"id": "` + serviceID + `",
		"options": {
			"active": false
		},
		"latest_version": {
			"type": "github",
			"url": "` + test.ArgusGitHubRepo + `"
		}
	}`)))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/service/new", payload)
	w := httptest.NewRecorder()
	api.httpServiceEdit(w, req, actionCreate)
	res := w.Result()
	t.Cleanup(func() { _ = res.Body.Close() })

	// 409, rather than racing the in-flight op.
	if got, want := res.StatusCode, http.StatusConflict; got != want {
		t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
	}

	if got, want := w.Body.String(), `another operation is in progress`; !util.RegexCheck(want, got) {
		t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestHTTP_ServiceEdit__edit(t *testing.T) {
	svcCfg := svctest.PlainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()

	type wants struct {
		bodyRegex                                string
		statusCode                               int
		latestVersionRegex, deployedVersionRegex string
		serviceYAML                              string
	}

	file := t.Name() + ".yml"
	api := testAPI(t, file)
	apiMu := sync.RWMutex{}

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
					`    json: .* unmarshal[^"]+"}`,
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
							url: `+test.LookupBare.URLValid+`/1.2.3
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
						"url":  "` + test.LookupBare.URLValid + `/version is v1.2.3",
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
				statusCode:           http.StatusOK,
				bodyRegex:            `{"message":"edited service[^}]+"}`,
				latestVersionRegex:   `^1\.2\.3$`,
				deployedVersionRegex: `^1\.2\.3$`,
				serviceYAML: test.TrimYAML(`
					options:
						interval: 99m
					latest_version:
						type: url
						url: ` + test.LookupBare.URLValid + `/version is v1.2.3
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
							url: `+test.LookupBare.URLInvalid+`/v1.2.3
							allow_invalid_certs: true
							regex: v?([0-9.]+)
						notify:
							discord: {}
						webhook:
							wh: {}
						dashboard:
							icon: https://example.com/icon.png
					`)),
					"successful dv edit",
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
						"url":  "` + test.LookupBare.URLInvalid + `/v1.2.3",
						"allow_invalid_certs": true,
						"regex": "v?([0-9.]+)"
					}
				}`,
			wants: wants{
				statusCode:           http.StatusOK,
				bodyRegex:            `{"message":"edited service[^}]+"}`,
				latestVersionRegex:   `^$`,
				deployedVersionRegex: `^1\.2\.3$`,
				serviceYAML: test.TrimYAML(`
					options:
						interval: 99m
					deployed_version:
						type: url
						url: ` + test.LookupBare.URLInvalid + `/v1.2.3
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
						"url":  "` + test.LookupPlain.URLValid + `",
						"url_commands": [
							{
								"type": "regex",
								"regex": "stable version: \"v?([0-9.]+)\""
							}
						]
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
			svc:  testService(t, "edit that doesn't query (lv) successfully", "url", "url", false),
			payload: `
				{
					"id": "__name__",
					"latest_version": {
						"type": "url",
						"url":  "` + test.LookupPlain.URLValid + `",
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
			svc:  testService(t, "edit that doesn't query (dv) successfully", "url", "url", false),
			payload: `
				{
					"id": "__name__",
					"deployed_version": {
						"type": "url",
						"url":  "` + test.LookupPlain.URLValid + `",
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
				api.Config.OrderMu.Lock()
				tc.svc.ID = serviceID
				api.Config.Service[tc.svc.ID] = tc.svc
				api.Config.Order = append(api.Config.Order, tc.svc.ID)
				api.Config.OrderMu.Unlock()
			}
			tc.payload = strings.ReplaceAll(tc.payload, "__name__", serviceID)
			tc.payload = test.TrimJSON(tc.payload)
			payload := bytes.NewReader([]byte(tc.payload))
			target := "/api/v1/service/config"
			params := url.Values{}
			params.Set("service_id", serviceID)
			req := httptest.NewRequest(http.MethodPut, target, payload)
			req.URL.RawQuery = params.Encode()

			w := httptest.NewRecorder()
			apiMu.Lock()
			api.httpServiceEdit(w, req, actionEdit)
			apiMu.Unlock()
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			if got, want := res.StatusCode, tc.wants.statusCode; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}

			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}
			if got := string(data); !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, tc.wants.bodyRegex)
			}
			if tc.wants.statusCode != http.StatusOK {
				return
			}

			api.Config.OrderMu.RLock()
			svc := api.Config.Service[serviceID]
			api.Config.OrderMu.RUnlock()
			if svc == nil {
				t.Fatalf("Config.Service[%q] mismatch\ngot:  nil\nwant: non-nil", serviceID)
			}

			// An edit replaces rather than merges.
			if got := svc.String(""); got != tc.wants.serviceYAML {
				t.Errorf("Service.String() mismatch\ngot:  %q\nwant: %q", got, tc.wants.serviceYAML)
			}

			gotLV := svc.Status.LatestVersion()
			wantLV := tc.wants.latestVersionRegex
			if !util.RegexCheck(wantLV, gotLV) {
				t.Errorf("Status.LatestVersion() mismatch\ngot:  %q\nwant: %q", gotLV, wantLV)
			}

			// Without a deployed_version, it copies LatestVersion().
			gotDV := svc.Status.DeployedVersion()
			wantDV := tc.wants.deployedVersionRegex
			if !util.RegexCheck(wantDV, gotDV) {
				t.Errorf("Status.DeployedVersion() mismatch\ngot:  %q\nwant: %q", gotDV, wantDV)
			}
		})
	}
}

func TestHTTP_ServiceEdit__edit__missingID(t *testing.T) {
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)

	const editID = "TestHTTP_ServiceEdit__edit__missingID"
	svc := testService(t, editID, "url", "url", true)
	api.Config.Service[svc.ID] = svc
	wantOrder := []string{"before", editID, "after"}
	api.Config.Order = slices.Clone(wantOrder)

	payload := bytes.NewReader([]byte(test.TrimJSON(`{
		"options": {
			"active": false
		},
		"latest_version": {
			"type": "github",
			"url": "` + test.ArgusGitHubRepo + `"
		}
	}`)))
	params := url.Values{}
	params.Set("service_id", editID)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/service/config", payload)
	req.URL.RawQuery = params.Encode()
	w := httptest.NewRecorder()
	api.httpServiceEdit(w, req, actionEdit)
	res := w.Result()
	t.Cleanup(func() { _ = res.Body.Close() })

	if got, want := res.StatusCode, http.StatusBadRequest; got != want {
		t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
	}

	if got, want := w.Body.String(), `failed:\\n  id: <required>"}$`; !util.RegexCheck(want, got) {
		t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, want)
	}

	if got := api.Config.Service[editID]; got != svc {
		t.Errorf("Config.Service[%q] mismatch\ngot:  %p\nwant: %p", editID, got, svc)
	}

	if got := api.Config.Service[""]; got != nil {
		t.Errorf("Config.Service[\"\"] mismatch\ngot:  %q\nwant: nil", got.String(""))
	}

	if got := api.Config.Order; !slices.Equal(got, wantOrder) {
		t.Errorf("Config.Order mismatch\ngot:  %q\nwant: %q", got, wantOrder)
	}
}

func TestHTTP_ServiceEdit__edit__renameToExisting(t *testing.T) {
	const (
		editID  = "rename-source"
		otherID = "rename-other"
	)
	tests := []struct {
		name      string
		otherName string
		newID     string
	}{
		{
			name:      "ID",
			otherName: "rename-other-name",
			newID:     otherID,
		},
		{
			name:      "Name",
			otherName: "rename-taken-name",
			newID:     "rename-taken-name",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: writes the package-level flag vars via testAPI's Config.Load.

			api := testAPI(t, filepath.Join(t.TempDir(), "argus.yml"))
			source := testService(t, editID, "url", "url", true)
			source.Name = "rename-source-name"
			other := testService(t, otherID, "url", "url", true)
			other.Name = tc.otherName
			api.Config.Service[source.ID] = source
			api.Config.Service[other.ID] = other
			api.Config.Order = append(api.Config.Order, source.ID, other.ID)

			// A unique name of its own, so only the ID collides.
			payload := bytes.NewReader([]byte(test.TrimJSON(`{
				"id": "` + tc.newID + `",
				"name": "rename-source-renamed",
				"options": {
					"active": false
				},
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `"
				}
			}`)))
			params := url.Values{}
			params.Set("service_id", editID)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/service/update", payload)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			api.httpServiceEdit(w, req, actionEdit)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			if got, want := res.StatusCode, http.StatusBadRequest; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}

			if got, want := w.Body.String(), `already exists`; !util.RegexCheck(want, got) {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, want)
			}

			if _, ok := api.Config.Service[editID]; !ok {
				t.Errorf("Config.Service[%q] mismatch\ngot:  absent\nwant: present", editID)
			}
		})
	}
}

func TestHTTP_ServiceEdit__edit__rejectsAnIDHeldAsAnotherServiceName(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)

	const (
		editID    = "restore-grants-source"
		holderID  = "restore-grants-holder"
		takenName = "restore-grants-taken-name"
	)
	source := testService(t, editID, "url", "url", true)
	source.Name = "restore-grants-source-name"
	holder := testService(t, holderID, "url", "url", true)
	holder.Name = takenName
	api.Config.Service[source.ID] = source
	api.Config.Service[holder.ID] = holder
	api.Config.Order = append(api.Config.Order, source.ID, holder.ID)

	group := createAuthGroup(t, deps, "scoped", "", []rbac.Grant{
		{Resource: rbac.ResourceService, Action: rbac.ActionRead, Scope: rbac.Scope{Type: rbac.ScopeService, Ref: editID}},
	})

	// An ID matching another service's name trips the guard AddService applies.
	payload := bytes.NewReader([]byte(test.TrimJSON(`{
		"id": "` + takenName + `",
		"name": "restore-grants-renamed",
		"options": {
			"active": false
		},
		"latest_version": {
			"type": "github",
			"url": "` + test.ArgusGitHubRepo + `"
		}
	}`)))
	params := url.Values{}
	params.Set("service_id", editID)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/service/update", payload)
	req.URL.RawQuery = params.Encode()
	req = withAuthCtx(req, adminContext(t, api, deps))
	w := httptest.NewRecorder()
	api.httpServiceEdit(w, req, actionEdit)

	if got, want := w.Code, http.StatusBadRequest; got != want {
		t.Fatalf("status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
	}

	// The handler's pre-check refused it, so no grant was moved.
	if got, want := w.Body.String(), "a service with this id or name already exists"; !strings.Contains(got, want) {
		t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, want)
	}

	if _, ok := api.Config.Service[editID]; !ok {
		t.Fatalf("Config.Service[%q] mismatch\ngot:  absent\nwant: present", editID)
	}
	got, err := deps.Store.GroupByID(t.Context(), group.ID)
	if err != nil {
		t.Fatalf("Store.GroupByID() failed: %v", err)
	}
	if gotLen, want := len(got.Grants), 1; gotLen != want {
		t.Fatalf("Group.Grants length mismatch\ngot:  %d\nwant: %d", gotLen, want)
	}
	for i, grant := range got.Grants {
		if grant.Scope.Ref != editID {
			t.Errorf(
				"Group.Grants[%d].Scope.Ref mismatch\ngot:  %q\nwant: %q",
				i, grant.Scope.Ref, editID,
			)
		}
	}
}

func TestHTTP_ServiceEdit__edit__secrets(t *testing.T) {
	svcCfg := svctest.PlainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()
	type wants struct {
		statusCode  int
		serviceYAML string
	}

	file := t.Name() + ".yml"
	api := testAPI(t, file)
	apiMu := sync.RWMutex{}

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
					"lv-url",
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
					"url": "` + test.LookupPlain.URLValid + `",
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
					name: lv-url
					comment: foo
					latest_version:
						type: url
						url: ` + test.LookupPlain.URLValid + `
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
							url: `+test.LookupPlainPOST.URLValid+`
							body: '`+test.LookupPlainPOST.DataPass+`'
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
					"url": "` + test.LookupPlainPOST.URLValid + `",
					"body": "` + strings.ReplaceAll(test.LookupPlainPOST.DataPass, `"`, `\"`) + `",
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
						url: ` + test.LookupPlainPOST.URLValid + `
						body: '` + test.LookupPlainPOST.DataPass + `'
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
									token: AAAAAAAAAAAAAAA
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
								token: AAAAAAAAAAAAAAA
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
								url: `+test.WebhookGitHub.URLValid+`
								secret: `+test.WebhookGitHub.SecretPass+`
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
						"url": "` + test.WebhookGitHub.URLValid + `",
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
							url: ` + test.WebhookGitHub.URLValid + `
							secret: ` + test.WebhookGitHub.SecretPass + `
				`),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			serviceID := tc.svc.ID
			tc.wants.serviceYAML = strings.ReplaceAll(tc.wants.serviceYAML, "__name__", tc.name)
			api.Config.OrderMu.Lock()
			api.Config.Service[serviceID] = tc.svc
			api.Config.Order = append(api.Config.Order, serviceID)
			api.Config.OrderMu.Unlock()
			tc.payload = strings.ReplaceAll(tc.payload, "__name__", serviceID)
			tc.payload = test.TrimJSON(tc.payload)
			payload := bytes.NewReader([]byte(tc.payload))
			target := "/api/v1/service/config"
			params := url.Values{}
			params.Set("service_id", serviceID)
			req := httptest.NewRequest(http.MethodPut, target, payload)
			req.URL.RawQuery = params.Encode()

			w := httptest.NewRecorder()
			apiMu.Lock()
			api.httpServiceEdit(w, req, actionEdit)
			apiMu.Unlock()
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			if got, want := res.StatusCode, tc.wants.statusCode; got != want {
				t.Fatalf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}

			api.Config.OrderMu.RLock()
			svc := api.Config.Service[serviceID]
			api.Config.OrderMu.RUnlock()
			if svc == nil {
				t.Fatalf("Config.Service[%q] mismatch\ngot:  nil\nwant: non-nil", serviceID)
			}

			// Each <secret> takes the old value, matched by old_index where given.
			got := svc.String("")
			if got != tc.wants.serviceYAML {
				tokenReplacer := regexp.MustCompile(`(access_token|secret|token): [^\s<]+`)
				tokenReplacement := "$1: SECRET"
				got = tokenReplacer.ReplaceAllString(got, tokenReplacement)
				want := tokenReplacer.ReplaceAllString(tc.wants.serviceYAML, tokenReplacement)
				if got == want {
					t.Errorf(
						"Service.String() mismatch: only the redacted secret values differ\n"+
							"got:  %q\nwant: %q",
						got, want,
					)
				} else {
					t.Errorf("Service.String() mismatch\ngot:  %q\nwant: %q", got, want)
				}
			}
		})
	}
}

func TestHTTP_ServiceEdit__edit__waitsForInFlightOp(t *testing.T) {
	file := t.Name() + ".yml"
	api := testAPI(t, file)
	const serviceID = "TestHTTP_ServiceEdit_waits-absent"

	params := url.Values{}
	params.Set("service_id", serviceID)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/service/update", nil)
	req.URL.RawQuery = params.Encode()

	// The service is absent, so it 404s once the lock is released.
	assertWaitsForOpLock(
		t, &api, serviceID,
		func(w http.ResponseWriter) { api.httpServiceEdit(w, req, actionEdit) },
		http.StatusNotFound,
	)
}

// assertWaitsForOpLock asserts that serve blocks while serviceID's op lock is held,
// rather than failing fast with 409 as a create does, and responds wantStatus once
// it is released.
func assertWaitsForOpLock(
	t *testing.T,
	api *API,
	serviceID string,
	serve func(http.ResponseWriter),
	wantStatus int,
) {
	t.Helper()

	held := api.acquireServiceOp(serviceID)
	held.mu.Lock() // In-flight operation holds the lock.

	done := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		w := httptest.NewRecorder()
		serve(w)
		done <- w
	}()

	select {
	case w := <-done:
		t.Fatalf(
			"held lock: request mismatch\ngot:  completed with %d - %q\nwant: blocked",
			w.Code, w.Body.String(),
		)
	case <-time.After(1 * time.Second):
	}

	// The in-flight operation releases the lock.
	held.mu.Unlock()
	api.releaseServiceOp(serviceID, held)

	select {
	case w := <-done:
		if w.Code != wantStatus {
			t.Errorf(
				"released lock: status code mismatch\ngot:  %d - %q\nwant: %d",
				w.Code, w.Body.String(), wantStatus,
			)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf(
			"released lock: status code mismatch\ngot:  timeout after 5s\nwant: %d",
			wantStatus,
		)
	}
}

func TestServiceEditAffectsVisibility(t *testing.T) {
	summaryFor := func(id string, tags []string) *apitype.ServiceSummary {
		summary := &apitype.ServiceSummary{ID: id}
		if len(tags) != 0 {
			summary.Tags = &tags
		}
		return summary
	}
	serviceFor := func(id string, tags []string) *service.Service {
		return &service.Service{
			ID:        id,
			Dashboard: dashboard.Options{Tags: tags},
		}
	}

	tests := []struct {
		name       string
		oldService *apitype.ServiceSummary
		newService *service.Service
		want       bool
	}{
		{
			name:       "create",
			oldService: nil,
			newService: serviceFor("new", nil),
			want:       true,
		},
		{
			name:       "edit/rename",
			oldService: summaryFor("old", nil),
			newService: serviceFor("new", nil),
			want:       true,
		},
		{
			name:       "edit/tag added",
			oldService: summaryFor("svc", nil),
			newService: serviceFor("svc", []string{"prod"}),
			want:       true,
		},
		{
			name:       "edit/tag removed",
			oldService: summaryFor("svc", []string{"prod"}),
			newService: serviceFor("svc", nil),
			want:       true,
		},
		{
			name:       "edit/tag swapped",
			oldService: summaryFor("svc", []string{"prod"}),
			newService: serviceFor("svc", []string{"staging"}),
			want:       true,
		},
		{
			name:       "edit/plain/same ID, no tags",
			oldService: summaryFor("svc", nil),
			newService: serviceFor("svc", nil),
			want:       false,
		},
		{
			name:       "edit/plain/same ID, same tags",
			oldService: summaryFor("svc", []string{"prod", "web"}),
			newService: serviceFor("svc", []string{"prod", "web"}),
			want:       false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := serviceEditAffectsVisibility(tc.oldService, tc.newService)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestHTTP_ServiceDelete(t *testing.T) {
	type wants struct {
		bodyRegex  string
		statusCode int
	}

	file := t.Name() + ".yml"
	api := testAPI(t, file)
	svc := testService(t, t.Name(), "url", "url", true)
	svc.HardDefaults.Status.DatabaseChannel = api.Config.DatabaseChannel
	if err := api.Config.AddService("", svc); err != nil {
		t.Fatalf("Config.AddService() failed: %v", err)
	}
	// Drain the addition's database message.
	select {
	case <-api.Config.DatabaseChannel:
	case <-time.After(time.Second):
		t.Fatal("Config.DatabaseChannel message count mismatch\ngot:  0\nwant: 1")
	}
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
			// Serial: shares api.Config.

			target := "/api/v1/service/delete"
			params := url.Values{}
			params.Set("service_id", tc.serviceID)

			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			api.httpServiceDelete(w, req)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			if got, want := res.StatusCode, tc.wants.statusCode; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}

			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}
			if got := string(data); !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, tc.wants.bodyRegex)
			}

			if api.Config.Service[tc.serviceID] != nil {
				t.Errorf(
					"Config.Service[%q] mismatch\ngot:  present\nwant: absent (deleted)",
					tc.serviceID,
				)
			}
			if slices.Contains(api.Config.Order, tc.serviceID) {
				t.Errorf(
					"Config.Order mismatch\ngot:  %q\nwant: NOT %q",
					api.Config.Order, tc.serviceID,
				)
			}

			if tc.wants.statusCode == http.StatusOK {
				select {
				case msg := <-api.Config.DatabaseChannel:
					if !msg.Delete {
						t.Errorf(
							"Message.Delete mismatch\n"+
								"got:  false (ServiceID=%q, %d cells)\nwant: true",
							msg.ServiceID, len(msg.Cells),
						)
					}
				case <-time.After(time.Second):
					t.Error("Config.DatabaseChannel message count mismatch\ngot:  0\nwant: 1")
				}
			}
		})
	}
}

func TestHTTP_ServiceDelete__waitsForInFlightOp(t *testing.T) {
	file := t.Name() + ".yml"
	api := testAPI(t, file)
	svc := testService(t, "TestHTTP_ServiceDelete_waits", "url", "url", true)
	svc.HardDefaults.Status.DatabaseChannel = api.Config.DatabaseChannel
	if err := api.Config.AddService("", svc); err != nil {
		t.Fatalf("Config.AddService() failed: %v", err)
	}
	// Drain the addition's database message.
	select {
	case <-api.Config.DatabaseChannel:
	case <-time.After(time.Second):
		t.Fatal("Config.DatabaseChannel message count mismatch\ngot:  0\nwant: 1")
	}

	params := url.Values{}
	params.Set("service_id", svc.ID)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/service/delete", nil)
	req.URL.RawQuery = params.Encode()

	assertWaitsForOpLock(
		t, &api, svc.ID,
		func(w http.ResponseWriter) { api.httpServiceDelete(w, req) },
		http.StatusOK,
	)

	api.serviceOpMu.Lock()
	_, leaked := api.serviceOps[svc.ID]
	api.serviceOpMu.Unlock()
	if leaked {
		t.Errorf("API.serviceOps[%q] mismatch\ngot:  present\nwant: absent (released)", svc.ID)
	}
}

func TestHTTP_NotifyTest(t *testing.T) {
	type wants struct {
		bodyRegex  string
		statusCode int
	}

	file := t.Name() + ".yml"
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
		name    string
		payload string
		wants   wants
	}{
		{
			name: "no body",
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  "body required",
			},
		},
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
			name: "unknown notify ref",
			payload: `{
				"name": "new_notify"
			}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `invalid type "[^"]+"`,
			},
		},
		{
			name: "no name or name_previous",
			payload: `{
				"service_id": "new_service"
			}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `name and/or name_previous are required`,
			},
		},
		{
			name: "new service/no main/topic required",
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
			name: "new service/no main/no service_id",
			payload: `{
				"name": "test_notify",
				"type": "ntfy"}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  "url_fields:[^ ]+ +topic: .*required",
			},
		},
		{
			name: "new service/no main/invalid JSON/options",
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
						jsontext: invalid character.*$`,
				),
			},
		},
		{
			name: "new service/no main/options, invalid",
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
			name: "new service/have main/options applied, delay ignored",
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
			name: "new service/no main/invalid JSON/url_fields",
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
						jsontext: invalid character.*$`,
				),
			},
		},
		{
			name: "new service/have main/url_fields, invalid",
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
			name: "new service/no main/no type",
			payload: `{
				"service_id": "also_unknown",
				"name": "test_notify"}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `invalid type "test_notify"`,
			},
		},
		{
			name: "new service/no main/unknown type",
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
			name: "new service/no main/type from ID",
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
			name: "new service/have main/type from Main",
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
			name: "same service/have main/type from original",
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
			name: "same service/no main/can remove vars",
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
			name: "same service/no main/unsent vars inherited",
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
			name: "same service/have main/fail send",
			payload: `{
				"service_id_previous": "test",
				"service_id": "test",
				"name": "test",
				"name_previous": "test",
				"type": "` + validNotify.Type + `",
				"url_fields": {
					"host": "` + validNotify.URLFields["host"] + `",
					"path": "` + validNotify.URLFields["path"] + `",
					"token": "AAAAAAAAAAAAAAA"
				}
			}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `failed to send .*Unauthorized \(401\)`,
			},
		},
		{
			name: "same service/have main/new name, also fail send",
			payload: `{
				"service_id_previous": "test",
				"service_id": "new_name",
				"name": "test",
				"name_previous": "test",
				"type": "` + validNotify.Type + `",
				"url_fields": {
					"host": "` + validNotify.URLFields["host"] + `",
					"path": "` + validNotify.URLFields["path"] + `",
					"token": "AAAAAAAAAAAAAAA"
				}
			}`,
			wants: wants{
				statusCode: http.StatusBadRequest,
				bodyRegex:  `failed to send .*Unauthorized \(401\)`,
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
				// With nothing to inherit, the name is taken as the type.
				bodyRegex: `invalid type "new_notify"`,
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
				// With nothing to inherit, the name is taken as the type.
				bodyRegex: `invalid type "new_notify"`,
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
				// With nothing to inherit, the name is taken as the type.
				bodyRegex: `invalid type "new_notify"`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.wants.bodyRegex = util.ValueOr(tc.wants.bodyRegex, `^$`)
			tc.payload = test.TrimJSON(tc.payload)
			payload := bytes.NewReader([]byte(tc.payload))

			req := httptest.NewRequest(http.MethodGet, "/api/v1/notify/test", payload)
			w := httptest.NewRecorder()
			api.httpNotifyTest(w, req)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			if got, want := res.StatusCode, tc.wants.statusCode; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}

			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}
			var body map[string]string
			if err := decode.Unmarshal("json", data, &body); err != nil {
				t.Fatalf("decode.Unmarshal() failed: %v", err)
			}
			if got := body["message"]; !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, tc.wants.bodyRegex)
			}
		})
	}
}

func TestHTTP_ServiceOpLock__conflict(t *testing.T) {
	// hold describes how the test pre-holds a service's op lock before a request.
	type hold int
	const (
		holdNone  hold = iota
		holdWrite      // Simulate an in-flight edit/delete (exclusive).
		holdRead       // Simulate an in-flight refresh (shared).
	)

	file := t.Name() + ".yml"
	api := testAPI(t, file)

	tests := []struct {
		name       string
		handler    func(http.ResponseWriter, *http.Request)
		hold       hold
		statusCode int
		bodyRegex  string
	}{
		{
			name:       "latest-version refresh allowed with no op in flight",
			handler:    api.httpLatestVersionRefresh,
			hold:       holdNone,
			statusCode: http.StatusNotFound, // Past the lock, then the absent service.
			bodyRegex:  `not found`,
		},
		{
			name:       "latest-version refresh rejected while an exclusive op is in flight",
			handler:    api.httpLatestVersionRefresh,
			hold:       holdWrite,
			statusCode: http.StatusConflict,
			bodyRegex:  `another operation is in progress`,
		},
		{
			name:       "latest-version refresh allowed alongside another refresh (shared)",
			handler:    api.httpLatestVersionRefresh,
			hold:       holdRead,
			statusCode: http.StatusNotFound, // Past the lock, then the absent service.
			bodyRegex:  `not found`,
		},
		{
			name:       "deployed-version refresh rejected while an exclusive op is in flight",
			handler:    api.httpDeployedVersionRefresh,
			hold:       holdWrite,
			statusCode: http.StatusConflict,
			bodyRegex:  `another operation is in progress`,
		},
		{
			name:       "deployed-version refresh allowed alongside another refresh (shared)",
			handler:    api.httpDeployedVersionRefresh,
			hold:       holdRead,
			statusCode: http.StatusNotFound, // Past the lock, then the absent service.
			bodyRegex:  `not found`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			serviceID := tc.name

			held := api.acquireServiceOp(serviceID)
			switch tc.hold {
			case holdWrite:
				held.mu.Lock()
			case holdRead:
				held.mu.RLock()
			}
			t.Cleanup(func() {
				switch tc.hold {
				case holdWrite:
					held.mu.Unlock()
				case holdRead:
					held.mu.RUnlock()
				}
				api.releaseServiceOp(serviceID, held)
			})

			params := url.Values{}
			params.Set("service_id", serviceID)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/service", nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			tc.handler(w, req)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			if got, want := res.StatusCode, tc.statusCode; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}

			if got := w.Body.String(); !util.RegexCheck(tc.bodyRegex, got) {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, tc.bodyRegex)
			}
		})
	}
}
