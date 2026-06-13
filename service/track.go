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

// Package service provides the service functionality for Argus.
package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/web/metric"
)

// Track will call Track on each Service, each in their own goroutine.
func (s *Services) Track(ordering *[]string, orderMu *sync.RWMutex) {
	metric.InitMetrics()

	orderMu.RLock()
	defer orderMu.RUnlock()
	for _, key := range *ordering {
		svc := (*s)[key]
		if svc.Options.GetActive() {
			svc.Options.Active = nil
		}

		// Track this Service in an infinite loop goroutine.
		go svc.Track()

		// Space out the tracking of each Service.
		time.Sleep(time.Second / 2)
	}
}

// Track the Service and send Notify messages and WebHooks when a new release is found.
// Pause for s.Interval between each check.
func (s *Service) Track() {
	s.initMetrics()
	// Skip inactive Services.
	if !s.Options.GetActive() {
		return
	}

	// Wait until the interval has elapsed.
	lastQueriedAt, _ := time.Parse(time.RFC3339, s.Status.LastQueried())
	if time.Since(lastQueriedAt) < s.Options.GetIntervalDuration() {
		time.Sleep(s.Options.GetIntervalDuration() - time.Since(lastQueriedAt))
	}

	// Track the deployed version in an infinite loop goroutine.
	if s.DeployedVersionLookup != nil {
		go func() {
			go s.DeployedVersionLookup.Track()
		}()
	}

	// If we have no LatestVersion, we can't track.
	if s.LatestVersion == nil {
		return
	}

	time.Sleep(2 * time.Second) // Give DeployedVersion some time to query first.
	// Track forever.
	logFrom := logx.LogFrom{Primary: s.ID}
	logx.Verbose(
		fmt.Sprintf(
			"Tracking %s at %s every %s",
			s.ID, s.LatestVersion.ServiceURL(), s.Options.GetInterval(),
		),
		logFrom,
		true,
	)
	for {
		// Stop tracking if deleting.
		if s.Status.Deleting() {
			return
		}

		// Query the Lookup.
		if newVersion, _ := s.LatestVersion.Query(true, logFrom); newVersion {
			go s.HandleUpdateActions(true)
		}

		// Sleep interval between checks.
		time.Sleep(s.Options.GetIntervalDuration())
	}
}
