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

package service

import (
	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/webhook"
)

// Copy returns a deep copy of the receiver, optionally including status channels.
func (s *Service) Copy(withChannels bool) *Service {
	if s == nil {
		return nil
	}

	svc := &Service{
		ID:           s.ID,
		Name:         s.Name,
		Comment:      s.Comment,
		Options:      *s.Options.Copy(),
		Dashboard:    *s.Dashboard.Copy(),
		Status:       *s.Status.Copy(withChannels),
		Defaults:     s.Defaults,
		HardDefaults: s.HardDefaults,
	}

	// LatestVersion.
	if s.LatestVersion != nil {
		svc.LatestVersion = s.LatestVersion.Copy(&svc.Status)
		svc.LatestVersion.SetStatus(&svc.Status)
	}
	// DeployedVersionLookup.
	if s.DeployedVersionLookup != nil {
		svc.DeployedVersionLookup = s.DeployedVersionLookup.Copy(&svc.Status)
		svc.DeployedVersionLookup.SetStatus(&svc.Status)
	}
	// Notify.
	svc.Notify = s.Notify.Copy(&svc.Status)
	// Command.
	if len(s.Command) != 0 {
		svc.CommandController = command.NewController(
			&svc.Status,
			s.Command,
			svc.Notify,
			svc.Options.GetIntervalPointer(),
		)
		svc.Command = s.Command.Copy()
	}
	// WebHook.
	svc.WebHook = s.WebHook.Copy(&svc.Status, webhook.Notifiers{Shoutrrr: &svc.Notify})

	return svc
}
