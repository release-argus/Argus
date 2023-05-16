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

package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/release-argus/Argus/util"
)

// Track will call Track on all Services in this Slice.
func (s *Slice) Track(ordering *[]string, orderMutex *sync.RWMutex) {
	orderMutex.RLock()
	defer orderMutex.RUnlock()
	for _, key := range *ordering {
		// Skip inactive Services (and services that were deleted on startup)
		if !(*s)[key].Options.GetActive() || (*s)[key] == nil {
			continue
		}
		(*s)[key].Options.Active = nil

		jLog.Verbose(
			fmt.Sprintf("Tracking %s at %s every %s",
				(*s)[key].ID, (*s)[key].LatestVersion.ServiceURL(true), (*s)[key].Options.GetInterval()),
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
	lastQueriedAt, _ := time.Parse(time.RFC3339, s.Status.LastQueried())
	if time.Since(lastQueriedAt) < s.Options.GetIntervalDuration() {
		time.Sleep(s.Options.GetIntervalDuration() - time.Since(lastQueriedAt))
	}

	// Track the deployed version in an infinite loop goroutine.
	go func() {
		time.Sleep(2 * time.Second) // Give LatestVersion some time to query first.

		go s.DeployedVersionLookup.Track()
	}()

	// Track forever.
	logFrom := util.LogFrom{Primary: s.ID}
	for {
		// If we're deleting this Service, stop tracking it.
		if s.Status.Deleting() {
			return
		}

		// If new release found by this query.
		newVersion, _ := s.LatestVersion.Query(true, &logFrom)

		// If a new version was found
		if newVersion {
			go s.HandleUpdateActions(true)
		}

		// Sleep interval between checks.
		time.Sleep(s.Options.GetIntervalDuration())
	}
}
