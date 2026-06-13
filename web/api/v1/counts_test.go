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
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/release-argus/Argus/config/decode"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	svctest "github.com/release-argus/Argus/service/test"
	whtest "github.com/release-argus/Argus/webhook/test"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/web/metric"
)

var router *mux.Router

func TestHTTP_Counts(t *testing.T) {
	svcCfg := svctest.PlainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()

	// GIVEN: values of metrics.
	tests := []struct {
		name                                                   string
		serviceCountCurrentActive, serviceCountCurrentInactive int
		updatesCurrentAvailable, updatesCurrentSkipped         int
		updateDetails, updateDetailsOther                      *[]UpdateDetails
	}{
		{
			name: "empty",
		},
		{
			name:                        "ServiceCount",
			serviceCountCurrentActive:   2,
			serviceCountCurrentInactive: 1,
		},
		{
			name:                      "UpdatesCurrent('AVAILABLE')",
			serviceCountCurrentActive: 5,
			updatesCurrentAvailable:   3,
			updateDetails: &[]UpdateDetails{
				{
					ServiceName:     "test-0",
					DeployedVersion: "0.0.0",
					LatestVersion:   "1.0.0",
					LastChecked:     time.Now().UTC().Format(time.RFC3339),
					AutoApprove:     false,
					Approved:        false,
					Skipped:         false,
				},
				{
					ServiceName:     "test-1",
					DeployedVersion: "0.0.0",
					LatestVersion:   "1.0.0",
					LastChecked:     time.Now().UTC().Format(time.RFC3339),
					AutoApprove:     true,
					Approved:        true,
					Skipped:         false,
				},
				{
					ServiceName:     "test-2",
					DeployedVersion: "0.0.0",
					LatestVersion:   "1.0.0",
					LastChecked:     time.Now().UTC().Format(time.RFC3339),
					AutoApprove:     true,
					Approved:        false,
					Skipped:         false,
				},
			},
		},
		{
			name:                        "UpdatesCurrent('SKIPPED')",
			serviceCountCurrentActive:   2,
			serviceCountCurrentInactive: 4,
			updatesCurrentSkipped:       1,
			updateDetails: &[]UpdateDetails{
				{
					ServiceName:     "test-0",
					DeployedVersion: "0.0.0",
					LatestVersion:   "1.0.0",
					LastChecked:     time.Now().UTC().Format(time.RFC3339),
					AutoApprove:     false,
					Approved:        false,
					Skipped:         true,
				},
			},
		},
		{
			name:                        "all",
			serviceCountCurrentActive:   3,
			serviceCountCurrentInactive: 2,
			updatesCurrentAvailable:     2,
			updatesCurrentSkipped:       2,
			updateDetails: &[]UpdateDetails{
				{
					ServiceName:     "test-0",
					DeployedVersion: "0.0.0",
					LatestVersion:   "1.0.0",
					LastChecked:     time.Now().UTC().Format(time.RFC3339),
					AutoApprove:     true,
					Approved:        false,
					Skipped:         true,
				},
				{
					ServiceName:     "test-1",
					DeployedVersion: "0.0.0",
					LatestVersion:   "1.0.0",
					LastChecked:     time.Now().UTC().Format(time.RFC3339),
					AutoApprove:     true,
					Approved:        true,
					Skipped:         false,
				},
				{
					ServiceName:     "test-2",
					DeployedVersion: "0.0.0",
					LatestVersion:   "1.0.0",
					LastChecked:     time.Now().UTC().Format(time.RFC3339),
					AutoApprove:     true,
					Approved:        false,
					Skipped:         true,
				},
				{
					ServiceName:     "test-3",
					DeployedVersion: "0.0.0",
					LatestVersion:   "1.0.0",
					LastChecked:     time.Now().UTC().Format(time.RFC3339),
					AutoApprove:     true,
					Approved:        false,
					Skipped:         false,
				},
			},
		},
		{
			name:                        "up-to-date services ignored",
			serviceCountCurrentActive:   1,
			serviceCountCurrentInactive: 1,
			updatesCurrentAvailable:     1,
			updatesCurrentSkipped:       1,
			updateDetails:               nil,
			updateDetailsOther: &[]UpdateDetails{
				{
					ServiceName:     "test-0",
					DeployedVersion: "1.0.0",
					LatestVersion:   "1.0.0",
					LastChecked:     time.Now().UTC().Format(time.RFC3339),
					AutoApprove:     false,
					Approved:        false,
					Skipped:         false,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're testing metrics.

			api := API{}

			tmpl := []byte(test.TrimYAML(`
				id: test
				latest_version:
					type: url
					url: ` + test.LookupPlain["url_valid"] + `
			`))
			serviceTotal := tc.serviceCountCurrentActive + tc.serviceCountCurrentInactive
			api.Config = &config.Config{}
			api.Config.Service = make(service.Services, serviceTotal)
			updateDetailIndex := 0
			for i := 0; i < serviceTotal; i++ {
				id := fmt.Sprintf("test-%d", i)
				newService, err := service.DecodeService(
					"yaml", tmpl,
					id,
					svcCfg, notifyCfg, whCfg,
				)
				if err != nil {
					t.Fatalf(
						"%s\nfailed to unmarshal service template: %v",
						packageName, err,
					)
				}

				api.Config.Service[id] = newService
				api.Config.Order = append(api.Config.Order, id)

				if tc.updateDetails == nil || updateDetailIndex >= len(*tc.updateDetails) {
					updateDetailsLength := 0
					if tc.updateDetails != nil {
						updateDetailsLength = len(*tc.updateDetails)
					}
					if tc.updateDetailsOther != nil && updateDetailIndex < (updateDetailsLength+len(*tc.updateDetailsOther)) {
						otherIndex := updateDetailIndex - updateDetailsLength
						detail := (*tc.updateDetailsOther)[otherIndex]

						newService.Status.SetDeployedVersion(detail.DeployedVersion, "", false)
						newService.Status.SetLatestVersion(detail.LatestVersion, "", false)
						updateDetailIndex++
					}
					continue
				}
				detail := (*tc.updateDetails)[updateDetailIndex]

				newService.ID = detail.ServiceName
				approvedVersion := "something"
				if detail.Approved {
					approvedVersion = detail.LatestVersion
				} else if detail.Skipped {
					approvedVersion = serviceinfo.SkippedVersion(detail.LatestVersion)
				}
				newService.Status.SetDeployedVersion(detail.DeployedVersion, "", false)
				newService.Status.SetLatestVersion(detail.LatestVersion, "", false)
				newService.Status.SetApprovedVersion(approvedVersion, false)
				newService.Status.SetLastQueried(detail.LastChecked)
				newService.Dashboard.AutoApprove = &detail.AutoApprove
				updateDetailIndex++
			}

			metric.ServiceCountCurrent.Reset()
			metric.ServiceCountCurrentAdd(test.Ptr(true), tc.serviceCountCurrentActive)
			metric.ServiceCountCurrentAdd(test.Ptr(false), tc.serviceCountCurrentInactive)
			t.Cleanup(func() {
				metric.ServiceCountCurrent.Reset()
			})
			metric.UpdatesCurrent.WithLabelValues("AVAILABLE").Set(float64(tc.updatesCurrentAvailable))
			metric.UpdatesCurrent.WithLabelValues("SKIPPED").Set(float64(tc.updatesCurrentSkipped))
			var updateDetailsStr string
			if tc.updateDetails != nil {
				updateDetailsStr = fmt.Sprintf(
					`,"update_details": %v`,
					decode.ToJSONString(*tc.updateDetails),
				)
			}
			wantJSON := test.TrimJSON(fmt.Sprintf(`{
				"service_count": %d,
				"service_count_active": %d,
				"service_count_inactive": %d,
				"updates_available": %d,
				"updates_skipped": %d%s
			}`,
				tc.serviceCountCurrentActive+tc.serviceCountCurrentInactive,
				tc.serviceCountCurrentActive,
				tc.serviceCountCurrentInactive,
				tc.updatesCurrentAvailable,
				tc.updatesCurrentSkipped,
				updateDetailsStr,
			)) + "\n"

			// WHEN: a HTTP request is sent to the /counts endpoint.
			req := httptest.NewRequest(http.MethodGet, "/api/v1/counts", nil)
			w := httptest.NewRecorder()
			api.httpCounts(w, req)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			// THEN: the set values are returned in the JSON response.
			data, _ := io.ReadAll(res.Body)
			if got := string(data); got != wantJSON {
				t.Errorf(
					"%s\nAPI.httpCounts() response mismatch\ngot:  %q\nwant: %q",
					packageName, got, wantJSON,
				)
			}
		})
	}
}
