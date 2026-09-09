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
	"math/rand"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/release-argus/Argus/internal/logx"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

// HandleSkip marks the latest version as skipped.
func (s *Service) HandleSkip() {
	// Ignore skips if latest_version is deployed.
	if lv := s.Status.LatestVersion(); lv != s.Status.DeployedVersion() {
		s.Status.SetApprovedVersion(serviceinfo.SkippedVersion(lv), true)
	}
}

// HandleCommand finds and runs the named command on the Service if it is runnable.
func (s *Service) HandleCommand(command string) {
	// Find the command.
	index, err := s.CommandController.Find(command)
	if err != nil {
		logx.Warn(err, logx.LogFrom{Primary: "Command", Secondary: s.ID}, true)
		return
	}

	// Skip if it ran less than 2*Interval ago.
	if !(*s.CommandController).IsRunnable(index) {
		return
	}

	// Send the Command.
	err = (*s.CommandController).ExecIndex(
		logx.LogFrom{Primary: "Command", Secondary: s.ID},
		index,
		s.Status.GetServiceInfo(),
	)
	if err == nil {
		s.UpdatedVersion(true)
	}
}

// HandleWebhook finds the specified Webhook on this Service (and sends it if found).
func (s *Service) HandleWebhook(webhookID string) {
	if s.Webhook == nil || s.Webhook[webhookID] == nil {
		return
	}

	// Skip if before NextRunnable.
	if !s.Webhook[webhookID].IsRunnable() {
		return
	}

	// Send the Webhook.
	if err := s.Webhook[webhookID].Send(s.Status.GetServiceInfo(), false); err == nil {
		s.UpdatedVersion(true)
	}
}

// HandleUpdateActions runs all Commands and Webhooks for the service if auto_approve is enabled,
// otherwise waits for manual approval via the API.
func (s *Service) HandleUpdateActions(writeToDB bool) {
	svcInfo := s.Status.GetServiceInfo()

	// Send notify messages asynchronously.
	//nolint:errcheck
	go s.Notify.Send("", "", svcInfo, true)

	// Auto-update version for Services without Webhooks/Commands.
	if len(s.Webhook) == 0 && len(s.Command) == 0 {
		s.UpdatedVersion(writeToDB)
		return
	}

	// Approval required for Webhooks/Commands.
	if !s.Dashboard.GetAutoApprove() {
		logx.Info(
			"Waiting for approval on the Web UI",
			logx.LogFrom{Primary: s.ID},
			true,
		)

		s.Status.AnnounceQueryNewVersion()

		return
	}

	logx.Info(
		fmt.Sprintf(
			"Sending Webhooks/Running Commands for %q",
			s.Status.LatestVersion(),
		),
		logx.LogFrom{Primary: s.ID},
		true,
	)

	var g errgroup.Group

	// Run the Commands.
	if len(s.Command) != 0 {
		g.Go(func() error {
			return s.CommandController.Exec(
				logx.LogFrom{
					Primary:   "Command",
					Secondary: s.ID,
				},
			)
		})
	}

	// Send the Webhooks.
	if len(s.Webhook) != 0 {
		g.Go(func() error {
			return s.Webhook.Send(svcInfo, true)
		})
	}

	if err := g.Wait(); err == nil {
		s.UpdatedVersion(writeToDB)
	}
}

// HandleFailedActions re-sends all failed Webhooks and Commands, or re-sends all if all previously succeeded.
func (s *Service) HandleFailedActions() {
	svcInfo := s.Status.GetServiceInfo()
	errChan := make(chan error, len(s.Webhook)+len(s.Command))
	errored := false

	retryAll := s.shouldRetryAll()

	potentialErrors := 0
	// Send the Webhooks.
	if len(s.Webhook) != 0 {
		potentialErrors += len(s.Webhook)
		for key, wh := range s.Webhook {
			if retryAll || util.DerefOr(s.Status.Fails.Webhook.Get(key), true) {
				// Skip if before NextRunnable.
				if !wh.IsRunnable() {
					potentialErrors--
					continue
				}
				// Send.
				go func(w *webhook.Webhook) {
					err := w.Send(svcInfo, false)
					errChan <- err
				}(wh)
				// Space out Webhooks.
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
		logFrom := logx.LogFrom{Primary: "Command", Secondary: s.ID}
		for key := range s.Command {
			if retryAll || util.DerefOr(s.Status.Fails.Command.Get(key), true) {
				// Skip if before NextRunnable.
				if !s.CommandController.IsRunnable(key) {
					potentialErrors--
					continue
				}
				// Run.
				go func(key int) {
					err := s.CommandController.ExecIndex(logFrom, key, svcInfo)
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

// shouldRetryAll determines whether all Webhooks and Commands should be retried.
// It returns true if every Webhook has sent successfully and every Command
// has run successfully. If any Webhook or Command has failed, it returns false.
func (s *Service) shouldRetryAll() bool {
	// Retry all only if every Webhook has sent successfully.
	for key := range s.Webhook {
		if util.DerefOr(s.Status.Fails.Webhook.Get(key), true) {
			return false
		}
	}
	// And every Command has run successfully.
	for key := range s.Command {
		if util.DerefOr(s.Status.Fails.Command.Get(key), true) {
			return false
		}
	}

	return true
}

// UpdatedVersion registers the version change, setting DeployedVersion to LatestVersion
// if there is no DeployedVersionLookup, and announces the change.
func (s *Service) UpdatedVersion(writeToDB bool) {
	if s.Status.DeployedVersion() == s.Status.LatestVersion() {
		return
	}

	// Check that no Webhooks failed.
	if !s.Status.Fails.Webhook.AllPassed() {
		return
	}
	// Check that no Commands failed.
	if !s.Status.Fails.Command.AllPassed() {
		return
	}
	// Do not update DeployedVersion to LatestVersion if we have a deployed lookup check.
	if s.DeployedVersionLookup != nil {
		if len(s.Command) != 0 || len(s.Webhook) != 0 {
			// Update ApprovedVersion if Commands/Webhooks may update DeployedVersion.
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

// UpdateLatestApproved sets the latest version as approved if not already set.
func (s *Service) UpdateLatestApproved() {
	if lv := s.Status.LatestVersion(); lv != s.Status.ApprovedVersion() {
		s.Status.SetApprovedVersion(lv, true)
	}
}
