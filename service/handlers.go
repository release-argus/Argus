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

// Package service provides the service functionality for Argus.
package service

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
	"github.com/release-argus/Argus/webhook"
)

// HandleSkip will set `version` to skipped and announce it to the websocket.
func (s *Service) HandleSkip() {
	// Ignore skips if latest_version is deployed.
	if s.Status.DeployedVersion() != s.Status.LatestVersion() {
		s.Status.SetApprovedVersion("SKIP_"+s.Status.LatestVersion(), true)
	}
}

// HandleCommand will find the specified command on this Service (and run it if found).
func (s *Service) HandleCommand(command string) {
	// Find the command.
	index, err := s.CommandController.Find(command)
	if err != nil {
		logutil.Log.Warn(err, logutil.LogFrom{Primary: "Command", Secondary: s.ID}, true)
		return
	}

	// Skip if it ran less than 2*Interval ago.
	if !(*s.CommandController).IsRunnable(index) {
		return
	}

	// Send the Command.
	err = (*s.CommandController).ExecIndex(
		logutil.LogFrom{Primary: "Command", Secondary: s.ID},
		index,
		s.Status.GetServiceInfo())
	if err == nil {
		s.UpdatedVersion(true)
	}
}

// HandleWebHook finds the specified WebHook on this Service (and sends it if found).
func (s *Service) HandleWebHook(webhookID string) {
	//nolint:typecheck
	if s.WebHook == nil || s.WebHook[webhookID] == nil {
		return
	}

	// Skip if before NextRunnable.
	if !s.WebHook[webhookID].IsRunnable() {
		return
	}

	// Send the WebHook.
	err := s.WebHook[webhookID].Send(s.Status.GetServiceInfo(), false)
	if err == nil {
		s.UpdatedVersion(true)
	}
}

// HandleUpdateActions runs all commands and send all WebHooks for this service if auto-approve true.
// If new releases are not auto-approved, then these will
// only run/send if manually triggered fromUser (via the WebUI).
func (s *Service) HandleUpdateActions(writeToDB bool) {
	serviceInfo := s.Status.GetServiceInfo()

	// Send the Notify Messages.
	//nolint:errcheck
	go s.Notify.Send("", "", serviceInfo, true)

	//nolint:typecheck
	if s.WebHook != nil || s.Command != nil {
		if s.Dashboard.GetAutoApprove() {
			msg := fmt.Sprintf("Sending WebHooks/Running Commands for %q",
				s.Status.LatestVersion())
			logutil.Log.Info(msg, logutil.LogFrom{Primary: s.ID}, true)

			// Run the Commands.
			go func() {
				err := s.CommandController.Exec(logutil.LogFrom{Primary: "Command", Secondary: s.ID})
				if err == nil && len(s.Command) != 0 {
					s.UpdatedVersion(writeToDB)
				}
			}()

			// Send the WebHooks.
			go func() {
				err := s.WebHook.Send(serviceInfo, true)
				if err == nil && len(s.WebHook) != 0 {
					s.UpdatedVersion(writeToDB)
				}
			}()
		} else {
			logutil.Log.Info("Waiting for approval on the Web UI", logutil.LogFrom{Primary: s.ID}, true)

			s.Status.AnnounceQueryNewVersion()
		}
	} else {
		// Auto-update version for Services without WebHooks.
		s.UpdatedVersion(writeToDB)
	}
}

// HandleFailedActions re-sends all the WebHooks for this service
// that have either failed or not sent for this version. Otherwise,
// if all WebHooks have sent successfully, then they all resend.
func (s *Service) HandleFailedActions() {
	serviceInfo := s.Status.GetServiceInfo()
	errChan := make(chan error, len(s.WebHook)+len(s.Command))
	errored := false

	retryAll := s.shouldRetryAll()

	potentialErrors := 0
	// Send the WebHooks.
	if len(s.WebHook) != 0 {
		potentialErrors += len(s.WebHook)
		for key, wh := range s.WebHook {
			if retryAll || util.DereferenceOrValue(s.Status.Fails.WebHook.Get(key), true) {
				// Skip if before NextRunnable.
				if !wh.IsRunnable() {
					potentialErrors--
					continue
				}
				// Send.
				go func(wh *webhook.WebHook) {
					err := wh.Send(serviceInfo, false)
					errChan <- err
				}(wh)
				// Space out WebHooks.
				//#nosec G404 -- sleep does not need cryptographic security.
				time.Sleep(time.Duration(100+rand.Intn(150)) * time.Millisecond)
			} else {
				potentialErrors--
			}
		}
	}
	// Run the Commands.
	if len(s.Command) != 0 {
		potentialErrors += len(s.Command)
		logFrom := logutil.LogFrom{Primary: "Command", Secondary: s.ID}
		for key := range s.Command {
			if retryAll || util.DereferenceOrValue(s.Status.Fails.Command.Get(key), true) {
				// Skip if before NextRunnable.
				if !s.CommandController.IsRunnable(key) {
					potentialErrors--
					continue
				}
				// Run.
				go func(key int) {
					err := s.CommandController.ExecIndex(logFrom, key, serviceInfo)
					errChan <- err
				}(key)
				// Space out Commands.
				//#nosec G404 -- sleep does not need cryptographic security.
				time.Sleep(time.Duration(100+rand.Intn(150)) * time.Millisecond)
			} else {
				potentialErrors--
			}
		}
	}

	for potentialErrors != 0 {
		if err := <-errChan; err != nil {
			errored = true
		}
		potentialErrors--
	}

	if !errored {
		s.UpdatedVersion(true)
	}
}

// shouldRetryAll determines whether all WebHooks and Commands should be retried.
// It returns true if every WebHook has sent successfully and every Command
// has run successfully. If any WebHook or Command has failed, it returns false.
func (s *Service) shouldRetryAll() bool {
	// Retry all only if every WebHook has sent successfully.
	for key := range s.WebHook {
		if util.DereferenceOrValue(s.Status.Fails.WebHook.Get(key), true) {
			return false
		}
	}
	// AND every Command has run successfully.
	for key := range s.Command {
		if util.DereferenceOrValue(s.Status.Fails.Command.Get(key), true) {
			return false
		}
	}

	return true
}

// UpdatedVersion will register the version change, setting `s.Status.DeployedVersion`
// to `s.Status.LatestVersion` if there's no DeployedVersionLookup and announce the change.
func (s *Service) UpdatedVersion(writeToDB bool) {
	if s.Status.DeployedVersion() == s.Status.LatestVersion() {
		return
	}

	// Check that no WebHook(s) failed.
	if !s.Status.Fails.WebHook.AllPassed() {
		return
	}
	// Check that no Command(s) failed.
	if !s.Status.Fails.Command.AllPassed() {
		return
	}
	// Do not update DeployedVersion to LatestVersion if we have a deployed lookup check.
	if s.DeployedVersionLookup != nil {
		if len(s.Command) != 0 || len(s.WebHook) != 0 {
			// Update ApprovedVersion if Commands/WebHooks may update DeployedVersion.
			// (only having `deployed_version`, `command` or `webhook` would only use ApprovedVersion to track skips)
			// They should have all ran/sent successfully at this point.
			s.UpdateLatestApproved()
		}
		return
	}
	s.Status.SetDeployedVersion(s.Status.LatestVersion(), "", writeToDB)

	// Announce version change to WebSocket clients.
	s.Status.AnnounceUpdate()
}

// UpdateLatestApproved will check if all WebHook(s) have sent successfully for this Service,
// set the LatestVersion as approved in the Status, and announce the approval (if not previously).
func (s *Service) UpdateLatestApproved() {
	// Only announce once.
	lv := s.Status.LatestVersion()
	if s.Status.ApprovedVersion() != lv {
		s.Status.SetApprovedVersion(lv, true)
	}
}
