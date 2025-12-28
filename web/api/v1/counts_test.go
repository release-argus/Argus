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
	"testing"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/web/metric"
	"github.com/release-argus/Argus/webhook"
)

var router *mux.Router

func TestHTTP_Counts(t *testing.T) {
	// GIVEN values of metrics.
	tests := map[string]struct {
		serviceCountCurrentActive   int
		serviceCountCurrentInactive int
		updatesCurrentAvailable     int
		updatesCurrentSkipped       int
		updateDetails               *[]UpdateDetails
		updateDetailsOther          *[]UpdateDetails
	}{
		"empty": {},
		"ServiceCount": {
			serviceCountCurrentActive:   2,
			serviceCountCurrentInactive: 1,
		},
		"UpdatesCurrent('AVAILABLE')": {
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
		"UpdatesCurrent('SKIPPED')": {
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
				}},
		},
		"all": {
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
		"up-to-date services ignored": {
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
				}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
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
				var newService service.Service
				if err := yaml.Unmarshal(tmpl, &newService); err != nil {
					t.Fatalf("%s\nfailed to unmarshal service template: %v",
						packageName, err)
				}
				newService.Dashboard = *dashboard.NewOptions(
					nil, "", "", "", nil,
					&dashboard.OptionsDefaults{}, &dashboard.OptionsDefaults{})
				newService.Init(
					&service.Defaults{}, &service.Defaults{},
					&shoutrrr.ShoutrrrsDefaults{}, &shoutrrr.ShoutrrrsDefaults{}, &shoutrrr.ShoutrrrsDefaults{},
					&webhook.WebHooksDefaults{}, &webhook.Defaults{}, &webhook.Defaults{})

				api.Config.Service[id] = &newService
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
					approvedVersion = "SKIP_" + detail.LatestVersion
				}
				newService.Status.SetDeployedVersion(detail.DeployedVersion, "", false)
				newService.Status.SetLatestVersion(detail.LatestVersion, "", false)
				newService.Status.SetApprovedVersion(approvedVersion, false)
				newService.Status.SetLastQueried(detail.LastChecked)
				newService.Dashboard.AutoApprove = &detail.AutoApprove
				updateDetailIndex++
			}

			metric.ServiceCountCurrent.Reset()
			metric.ServiceCountCurrentAdd(test.BoolPtr(true), tc.serviceCountCurrentActive)
			metric.ServiceCountCurrentAdd(test.BoolPtr(false), tc.serviceCountCurrentInactive)
			t.Cleanup(func() {
				metric.ServiceCountCurrent.Reset()
			})
			metric.UpdatesCurrent.WithLabelValues("AVAILABLE").Set(float64(tc.updatesCurrentAvailable))
			metric.UpdatesCurrent.WithLabelValues("SKIPPED").Set(float64(tc.updatesCurrentSkipped))
			var updateDetailsStr string
			if tc.updateDetails != nil {
				updateDetailsStr = fmt.Sprintf(`,"update_details": %v`,
					util.ToJSONString(*tc.updateDetails))
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

			// WHEN a HTTP request is sent to the /counts endpoint.
			req := httptest.NewRequest(http.MethodGet, "/api/v1/counts", nil)
			w := httptest.NewRecorder()
			api.httpCounts(w, req)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			// THEN the set values are returned in the JSON response.
			data, _ := io.ReadAll(res.Body)
			if dataStr := string(data); dataStr != wantJSON {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, wantJSON, dataStr)
			}
		})
	}
}
