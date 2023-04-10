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

	"github.com/release-argus/Argus/util"
	metric "github.com/release-argus/Argus/web/metrics"
)

// UpdatedVersion will register the version change, setting `s.Status.DeployedVersion`
// to `s.Status.LatestVersion`
func (s *Service) UpdatedVersion(writeToDB bool) {
	if s.Status.GetDeployedVersion() == s.Status.GetLatestVersion() {
		return
	}

	// Check that no WebHook(s) failed
	if !s.Status.Fails.WebHook.AllPassed() {
		return
	}
	// Check that no Command(s) failed
	if !s.Status.Fails.Command.AllPassed() {
		return
	}
	// Don't update DeployedVersion to LatestVersion if we have a lookup check
	if s.DeployedVersionLookup != nil {
		//nolint:typecheck
		if (s.Command != nil && len(s.Command) != 0) ||
			s.WebHook != nil {
			// Update ApprovedVersion if there are Commands/WebHooks that should update DeployedVersion
			// (only having `deployed_version`,`command` or `webhook` would only use ApprovedVersion to track skips)
			// They should have all ran/sent successfully at this point
			s.UpdateLatestApproved()
		}
		return
	}
	s.Status.SetDeployedVersion(s.Status.GetLatestVersion(), writeToDB)

	// Announce version change to WebSocket clients
	s.Status.AnnounceUpdate()
}

// UpdateLatestApproved will check if all WebHook(s) have sent successfully for this Service,
// set the LatestVersion as approved in the Status, and announce the approval (if not previously).
func (s *Service) UpdateLatestApproved() {
	// Only announce once
	lv := s.Status.GetLatestVersion()
	if s.Status.GetApprovedVersion() != lv {
		s.Status.SetApprovedVersion(lv, true)
	}
}

// HandleUpdateActions will run all commands and send all WebHooks for this service if it has been called
// automatically and auto-approve is true. If new releases aren't auto-approved, then these will
// only be run/send if this is triggered fromUser (via the WebUI).
func (s *Service) HandleUpdateActions(writeToDB bool) {
	serviceInfo := s.GetServiceInfo()

	// Send the Notify Message(s).
	//nolint:errcheck
	go s.Notify.Send("", "", serviceInfo, true)

	//nolint:typecheck
	if s.WebHook != nil || s.Command != nil {
		if s.Dashboard.GetAutoApprove() {
			msg := fmt.Sprintf("Sending WebHooks/Running Commands for %q",
				s.Status.GetLatestVersion())
			jLog.Info(msg, util.LogFrom{Primary: s.ID}, true)

			// Run the Command(s)
			go func() {
				err := s.CommandController.Exec(&util.LogFrom{Primary: "Command", Secondary: s.ID})
				if err == nil && len(s.Command) != 0 {
					s.UpdatedVersion(writeToDB)
				}
			}()

			// Send the WebHook(s)
			go func() {
				err := s.WebHook.Send(serviceInfo, true)
				if err == nil && len(s.WebHook) != 0 {
					s.UpdatedVersion(writeToDB)
				}
			}()
		} else {
			jLog.Info("Waiting for approval on the Web UI", util.LogFrom{Primary: s.ID}, true)

			metric.SetPrometheusGauge(metric.AckWaiting,
				s.ID,
				1)
			s.Status.AnnounceQueryNewVersion()
		}
	} else {
		// Auto-update version for Service(s) without WebHook(s)
		s.UpdatedVersion(writeToDB)
	}
}

// HandleFailedActions will re-send all the WebHooks for this service
// that have either failed, or not been sent for this version. Otherwise,
// if all WebHooks have been sent successfully, then they'll all be resent.
func (s *Service) HandleFailedActions() {
	errChan := make(chan error)
	errored := false

	retryAll := s.shouldRetryAll()

	potentialErrors := 0
	// Send the WebHook(s).
	if len(s.WebHook) != 0 {
		potentialErrors += len(s.WebHook)
		for key := range s.WebHook {
			if retryAll || util.EvalNilPtr(s.Status.Fails.WebHook.Get(key), true) {
				// Skip if it's before NextRunnable
				if !s.WebHook[key].IsRunnable() {
					potentialErrors--
					continue
				}
				// Send
				go func(key string) {
					err := s.WebHook[key].Send(s.GetServiceInfo(), false)
					errChan <- err
				}(key)
				// Space out WebHooks.
				time.Sleep(250 * time.Millisecond)
			} else {
				potentialErrors--
			}
		}
	}
	// Run the Command(s)
	if len(s.Command) != 0 {
		potentialErrors += len(s.Command)
		logFrom := util.LogFrom{Primary: "Command", Secondary: s.ID}
		for key := range s.Command {
			if retryAll || util.EvalNilPtr(s.Status.Fails.Command.Get(key), true) {
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
				time.Sleep(250 * time.Millisecond)
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
				util.ErrorToString(errs), err)
		}
	}

	if !errored {
		s.UpdatedVersion(true)
	}
}

// HandleCommand will handle running the Command for this service
// to the matching Command.
func (s *Service) HandleCommand(command string) {
	// Find the command
	index := s.CommandController.Find(command)
	if index == nil {
		jLog.Warn(command+" not found", util.LogFrom{Primary: "Command", Secondary: s.ID}, true)
		return
	}

	// Skip if it ran less than 2*Interval ago
	if !(*s.CommandController).IsRunnable(*index) {
		return
	}

	// Send the Command.
	err := (*s.CommandController).ExecIndex(&util.LogFrom{Primary: "Command", Secondary: s.ID}, *index)
	if err == nil {
		s.UpdatedVersion(true)
	}
}

// HandleWebHook will handle sending the WebHook for this service
// to the WebHook with a matching ID.
func (s *Service) HandleWebHook(webhookID string) {
	//nolint:typecheck
	if s.WebHook == nil || s.WebHook[webhookID] == nil {
		return
	}

	// Skip if it's before NextRunnable.
	if !s.WebHook[webhookID].IsRunnable() {
		return
	}

	// Send the WebHook.
	err := s.WebHook[webhookID].Send(s.GetServiceInfo(), false)
	if err == nil {
		s.UpdatedVersion(true)
	}
}

// HandleSkip will set `version` to skipped and announce it to the websocket.
func (s *Service) HandleSkip(version string) {
	if version != s.Status.GetLatestVersion() {
		return
	}

	s.Status.SetApprovedVersion("SKIP_"+version, true)
}

func (s *Service) shouldRetryAll() (retry bool) {
	retry = true
	// retry all only if every WebHook has been sent successfully
	if len(s.WebHook) != 0 {
		for key := range s.WebHook {
			if util.EvalNilPtr(s.Status.Fails.WebHook.Get(key), true) {
				retry = false
				break
			}
		}
	}
	// AND every Command has been run successfully
	if retry && len(s.Command) != 0 {
		for key := range s.Command {
			if util.EvalNilPtr(s.Status.Fails.Command.Get(key), true) {
				retry = false
				break
			}
		}
	}
	return
}
