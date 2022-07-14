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
	// Check that no WebHook(s) failed
	if s.WebHook != nil {
		for key := range *s.WebHook {
			// Default nil to true = failed
			if utils.EvalNilPtr((*s.WebHook)[key].Failed, true) {
				return
			}
		}
	}
	// Check that no Command(s) failed
	if s.Command != nil {
		for key := range *s.Command {
			// Default nil to true = failed
			if utils.EvalNilPtr(s.CommandController.Failed[key], true) {
				return
			}
		}
	}
	// Don't update DeployedVersion to LatestVersion if we have a lookup check
	if s.DeployedVersionLookup != nil {
		if s.Command != nil || s.WebHook != nil {
			// Update ApprovedVersion if there are Commands/WebHooks that should update DeployedVersion
			// (only having `deployed_version`,`command` or `webhook` would only use ApprovedVersion to track skips)
			// They should have all ran/sent successfully at this point
			s.UpdateLatestApproved()
		}
		return
	}
	s.SetDeployedVersion(s.Status.LatestVersion)

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

// HandleUpdateActions will run all commands and send all WebHooks for this service if it has been called
// automatically and auto-approve is true. If new releases aren't auto-approved, then these will
// only be run/send if this is triggered fromUser (via the WebUI).
func (s *Service) HandleUpdateActions() {
	if s.WebHook != nil || s.Command != nil {
		if s.GetAutoApprove() {
			msg := fmt.Sprintf("Sending WebHooks/Running Commands for %q", s.Status.LatestVersion)
			jLog.Info(msg, utils.LogFrom{Primary: *s.ID}, true)

			// Run the Command(s)
			go func() {
				err := s.CommandController.Exec(&utils.LogFrom{Primary: "Command", Secondary: *s.ID})
				if err == nil {
					s.UpdatedVersion()
				}
			}()

			// Send the WebHook(s)
			go func() {
				err := s.WebHook.Send(s.GetServiceInfo(), true)
				if err == nil {
					s.UpdatedVersion()
				}
			}()
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

// HandleFailedActions will re-send all the WebHooks for this service
// that have either failed, or not been sent for this version. Otherwise,
// if all WebHooks have been sent successfully, then they'll all be resent.
func (s *Service) HandleFailedActions() {
	errChan := make(chan error)
	errored := false

	retryAll := true
	// retryAll only if every WebHook has been sent successfully
	if s.WebHook != nil {
		for key := range *s.WebHook {
			if !utils.EvalNilPtr((*s.WebHook)[key].Failed, true) {
				retryAll = false
				break
			}
		}
	}
	// AND every Command has been run successfully
	if retryAll && s.Command != nil {
		for key := range *s.Command {
			if utils.EvalNilPtr(s.CommandController.Failed[key], true) {
				retryAll = false
				break
			}
		}
	}

	potentialErrors := 0
	// Send the WebHook(s).
	if s.WebHook != nil {
		potentialErrors += len(*s.WebHook)
		for key := range *s.WebHook {
			if retryAll || utils.EvalNilPtr((*s.WebHook)[key].Failed, true) {
				// Skip if it's before NextRunnable
				if !(*s.WebHook)[key].IsRunnable() {
					potentialErrors--
					continue
				}
				// Send
				go func(key string) {
					err := (*s.WebHook)[key].Send(s.GetServiceInfo(), false)
					errChan <- err
				}(key)
				// Space out WebHooks.
				time.Sleep(1 * time.Second)
			} else {
				potentialErrors--
			}
		}
	}
	// Run the Command(s)
	if s.Command != nil {
		potentialErrors += len(*s.Command)
		logFrom := utils.LogFrom{Primary: "Command", Secondary: *s.ID}
		for key := range *s.Command {
			if retryAll || utils.EvalNilPtr(s.CommandController.Failed[key], true) {
				// Skip if it's before NextRunnable
				if !s.CommandController.IsRunnable(key) {
					potentialErrors--
					continue
				}
				// Run
				go func(key int) {
					err := s.CommandController.ExecIndex(&logFrom, key)
					errChan <- err
				}(key)
				// Space out Commands.
				time.Sleep(1 * time.Second)
			} else {
				potentialErrors--
			}
		}
	}

	var errs error
	for potentialErrors != 0 {
		err := <-errChan
		potentialErrors--
		if err != nil {
			errored = true
			errs = fmt.Errorf("%s\n%w",
				utils.ErrorToString(errs), err)
		}
	}

	if !errored {
		s.UpdatedVersion()
	}
}

// HandleCommand will handle running the Command for this service
// to the matching Command.
func (s *Service) HandleCommand(command string) {
	if s.Command == nil {
		return
	}

	// Find the command
	index := s.CommandController.Find(command)
	if index == nil {
		jLog.Warn(command+" not found", utils.LogFrom{Primary: "Command", Secondary: *s.ID}, true)
		return
	}

	// Skip if it ran less than 2*Interval ago
	if !(*s.CommandController).IsRunnable(*index) {
		return
	}

	// Send the Command.
	err := (*s.CommandController).ExecIndex(&utils.LogFrom{Primary: "Command", Secondary: *s.ID}, *index)
	if err == nil {
		s.UpdatedVersion()
	}
}

// HandleWebHook will handle sending the WebHook for this service
// to the WebHook with a matching ID.
func (s *Service) HandleWebHook(webhookID string) {
	if s.WebHook == nil || (*s.WebHook)[webhookID] == nil {
		return
	}

	// Skip if it's before NextRunnable
	if !(*s.WebHook)[webhookID].IsRunnable() {
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
	if version != s.Status.LatestVersion {
		return
	}

	s.Status.ApprovedVersion = "SKIP_" + version
	s.AnnounceApproved()

	if s.SaveChannel != nil {
		*s.SaveChannel <- true
	}
}
