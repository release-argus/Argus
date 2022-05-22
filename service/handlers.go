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
	"time"

	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/web/metrics"
)

// UpdatedVersion will register the version change, setting `s.Status.DeployedVersion`
// to `s.Status.LatestVersion`
func (s *Service) UpdatedVersion() {
	// Only update if all webhooks have been sent
	// and none failed
	if s.WebHook != nil {
		for key := range *s.WebHook {
			// Default nil to true = failed
			if utils.EvalBoolPtr((*s.WebHook)[key].Failed, true) {
				return
			}
		}
	}
	// Don't update DeployedVersion to LatestVersion if we have a lookup check
	if s.DeployedVersionLookup != nil {
		s.UpdateLatestApproved()
		return
	}
	s.Status.SetDeployedVersion(s.Status.LatestVersion)

	// Announce version change to WebSocket clients
	s.AnnounceUpdate()
	if s.SaveChannel != nil {
		*s.SaveChannel <- true
	}
}

// UpdateLatestApproved will check if all WebHook(s) have sent successfully for this Service,
// set the LatestVersion as approved in the Status, and announce the approval (if not previously).
func (s *Service) UpdateLatestApproved() {
	// Only announce once
	if s.Status.ApprovedVersion != s.Status.LatestVersion {
		s.Status.ApprovedVersion = s.Status.LatestVersion
		s.AnnounceApproved()
	}
}

// HandleWebHooks will send all WebHooks for this service if it has been called automatically and
// auto-approve is true. If new releases aren't auto-approved, then the WebHooks will only be sent
// if this is triggered fromUser (via the WebUI).
func (s *Service) HandleWebHooks(fromUser bool) {
	if s.WebHook != nil {
		if s.GetAutoApprove() || fromUser {
			msg := fmt.Sprintf("Sending WebHooks for %q", s.Status.LatestVersion)
			jLog.Info(msg, utils.LogFrom{Primary: *s.ID}, true)

			// Send the WebHook(s).
			err := (*s.WebHook).Send(s.GetServiceInfo(), !fromUser)
			if err == nil {
				s.UpdatedVersion()
			}
		} else {
			jLog.Info("Waiting for approval on the Web UI", utils.LogFrom{Primary: *s.ID}, true)

			metrics.SetPrometheusGaugeWithID(metrics.AckWaiting, *s.ID, 1)
			s.AnnounceQueryNewVersion()
		}
	} else {
		// Auto-update version for Service(s) without WebHook(s)
		s.UpdatedVersion()
	}
}

// HandleFailedWebHooks will re-send all the WebHooks for this service
// that have either failed, or not been sent for this version.
func (s *Service) HandleFailedWebHooks() {
	if s.WebHook == nil {
		return
	}
	errs := make(chan error)
	errored := false

	// Send the WebHook(s).
	for key := range *s.WebHook {
		if utils.EvalBoolPtr((*s.WebHook)[key].Failed, true) {
			go func(key string) {
				err := (*s.WebHook)[key].Send(s.GetServiceInfo(), false)
				errs <- err
			}(key)
			// Don't send all WebHooks at the same time.
			time.Sleep(1 * time.Second)
		}
	}

	var err error
	for range *s.WebHook {
		errFound := <-errs
		if errFound != nil {
			errored = true
			if err == nil {
				err = errFound
			} else {
				err = fmt.Errorf("%s\n%s", err.Error(), errFound.Error())
			}
		}
	}
	if !errored {
		s.UpdatedVersion()
	}
}

// HandleWebHook will handle sending the WebHook for this service
// to the WebHook with a matching ID.
func (s *Service) HandleWebHook(webhookID string) {
	if s.WebHook == nil && (*s.WebHook)[webhookID] != nil {
		return
	}
	// Send the WebHook.
	err := (*s.WebHook)[webhookID].Send(s.GetServiceInfo(), false)
	if err == nil {
		s.UpdatedVersion()
	}
}

// HandleSkip will set `version` to skipped and announce it to the websocket.
func (s *Service) HandleSkip(version string) {
	if version == "" {
		return
	}

	s.Status.ApprovedVersion = "SKIP_" + version
	s.AnnounceApproved()

	if s.SaveChannel != nil {
		*s.SaveChannel <- true
	}
}
