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
		// Skip inactive Services (and services that were deleted on startup)
		if !(*s)[key].Options.GetActive() || (*s)[key] == nil {
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
	// Skip inactive Services
	if !s.Options.GetActive() {
		s.DeleteMetrics()
		return
	}
	s.ResetMetrics()

	// If this Service was last queried less than interval ago, wait until interval has elapsed.
	lastQueriedAt, _ := time.Parse(time.RFC3339, s.Status.GetLastQueried())
	if time.Since(lastQueriedAt) < s.Options.GetIntervalDuration() {
		time.Sleep(s.Options.GetIntervalDuration() - time.Since(lastQueriedAt))
	}

	// Track the deployed version in a infinite loop goroutine.
	go func() {
		time.Sleep(2 * time.Second) // Give LatestVersion some time to query first.

		go s.DeployedVersionLookup.Track()
	}()

	// Track forever.
	for {
		// If we're deleting this Service, stop tracking it.
		if s.Status.Deleting {
			return
		}

		// If new release found by this query.
		newVersion, err := s.LatestVersion.Query()

		// If a new version was found and we're not already on it
		if newVersion {
			// Send the Notify Message(s).
			//nolint:errcheck
			go s.Notify.Send("", "", s.GetServiceInfo(), true)

			// WebHook(s)/Command(s)
			go s.HandleUpdateActions()
		}

		// If it failed
		if err != nil {
			switch e := err.Error(); {
			case strings.HasPrefix(e, "regex "):
				metric.SetPrometheusGauge(metric.LatestVersionQueryLiveness,
					s.ID,
					2)
			case strings.HasPrefix(e, "failed converting") && strings.Contains(e, " semantic version."):
				metric.SetPrometheusGauge(metric.LatestVersionQueryLiveness,
					s.ID,
					3)
			case strings.HasPrefix(e, "queried version") && strings.Contains(e, " less than "):
				metric.SetPrometheusGauge(metric.LatestVersionQueryLiveness,
					s.ID,
					4)
			default:
				metric.IncreasePrometheusCounter(metric.LatestVersionQueryMetric,
					s.ID,
					"",
					"",
					"FAIL")
				metric.SetPrometheusGauge(metric.LatestVersionQueryLiveness,
					s.ID,
					0)
			}
		} else {
			metric.IncreasePrometheusCounter(metric.LatestVersionQueryMetric,
				s.ID,
				"",
				"",
				"SUCCESS")
			metric.SetPrometheusGauge(metric.LatestVersionQueryLiveness,
				s.ID,
				1)
		}
		// Sleep interval between checks.
		time.Sleep(s.Options.GetIntervalDuration())
	}
}
