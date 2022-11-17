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

	"github.com/release-argus/Argus/util"
	metric "github.com/release-argus/Argus/web/metrics"
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
			util.LogFrom{Primary: (*s)[key].ID},
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
			switch e := err.Error(); {
			case strings.HasPrefix(e, "regex "):
				metric.SetPrometheusGaugeWithID(metric.LatestVersionQueryLiveness, s.ID, 2)
			case strings.HasPrefix(e, "failed converting") && strings.Contains(e, " semantic version."):
				metric.SetPrometheusGaugeWithID(metric.LatestVersionQueryLiveness, s.ID, 3)
			case strings.HasPrefix(e, "queried version") && strings.Contains(e, " less than "):
				metric.SetPrometheusGaugeWithID(metric.LatestVersionQueryLiveness, s.ID, 4)
			default:
				metric.IncreasePrometheusCounterWithIDAndResult(metric.LatestVersionQueryMetric, s.ID, "FAIL")
				metric.SetPrometheusGaugeWithID(metric.LatestVersionQueryLiveness, s.ID, 0)
			}
		} else {
			metric.IncreasePrometheusCounterWithIDAndResult(metric.LatestVersionQueryMetric, s.ID, "SUCCESS")
			metric.SetPrometheusGaugeWithID(metric.LatestVersionQueryLiveness, s.ID, 1)
		}
		// Sleep interval between checks.
		time.Sleep(s.Options.GetIntervalDuration())
	}
}
