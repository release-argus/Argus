// Copyright [2022] [Argus]
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

package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/web/metrics"
)

// Track will call Track on all Services in this Slice.
func (s *Slice) Track(ordering *[]string) {
	for _, key := range *ordering {
		// Skip disabled Services
		if !(*s)[key].Options.GetActive() {
			continue
		}
		(*s)[key].Options.Active = nil

		jLog.Verbose(
			fmt.Sprintf("Tracking %s at %s every %s", (*s)[key].ID, (*s)[key].LatestVersion.GetServiceURL(true), (*s)[key].Options.GetInterval()),
			utils.LogFrom{Primary: (*s)[key].ID},
			true)

		// Track this Service in a infinite loop goroutine.
		go (*s)[key].Track()

		// Space out the tracking of each Service.
		time.Sleep(time.Second / 2)
	}
}

// Track the Service and send Notify messages (Service.Notify) as
// well as WebHooks (Service.WebHook) when a new release is spotted.
// It sleeps for Service.Interval between each check.
func (s *Service) Track() {
	serviceInfo := s.GetServiceInfo()

	// Track the deployed version in a infinite loop goroutine.
	go s.DeployedVersionLookup.Track()

	// Track forever.
	time.Sleep(2 * time.Second) // Give DeployedVersion some time to query first
	for {
		// If new release found by this query.
		newVersion, err := s.LatestVersion.Query()

		// If a new version was found and we're not already on it
		if newVersion {
			// Get updated serviceInfo
			serviceInfo = s.GetServiceInfo()

			// Send the Notify Message(s).
			//nolint:errcheck
			go s.Notify.Send("", "", &serviceInfo, true)

			// WebHook(s)/Command(s)
			go s.HandleUpdateActions()
		}

		// If it failed
		if err != nil {
			if strings.HasPrefix(err.Error(), "regex ") {
				metrics.SetPrometheusGaugeWithID(metrics.LatestVersionQueryLiveness, s.ID, 2)
			} else if strings.HasPrefix(err.Error(), "failed converting") && strings.Contains(err.Error(), " semantic version.") {
				metrics.SetPrometheusGaugeWithID(metrics.LatestVersionQueryLiveness, s.ID, 3)
			} else if strings.HasPrefix(err.Error(), "queried version") && strings.Contains(err.Error(), " less than ") {
				metrics.SetPrometheusGaugeWithID(metrics.LatestVersionQueryLiveness, s.ID, 4)
			} else {
				metrics.IncreasePrometheusCounterWithIDAndResult(metrics.LatestVersionQueryMetric, s.ID, "FAIL")
				metrics.SetPrometheusGaugeWithID(metrics.LatestVersionQueryLiveness, s.ID, 0)
			}
		} else {
			metrics.IncreasePrometheusCounterWithIDAndResult(metrics.LatestVersionQueryMetric, s.ID, "SUCCESS")
			metrics.SetPrometheusGaugeWithID(metrics.LatestVersionQueryLiveness, s.ID, 1)
		}
		// Sleep interval between checks.
		time.Sleep(s.Options.GetIntervalDuration())
	}
}
