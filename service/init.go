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
	"strings"

	"github.com/release-argus/Argus/command"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/web/metric"
	"github.com/release-argus/Argus/webhook"
)

// IconURL returns the URL Icon for the Service.
func (s *Service) IconURL() *string {
	// Service.Icon
	if strings.HasPrefix(s.Dashboard.Icon, "http") {
		return &s.Dashboard.Icon
	}

	if s.Notify != nil {
		// Search for a web icon.
		for _, notify := range s.Notify {
			// `Params.Icon`
			if icon := notify.GetParam("icon"); icon != "" &&
				strings.HasPrefix(icon, "http") {
				return &icon
			}
		}
	}

	return nil
}

// init initialises the Service, wiring channels and assigning command/notify/webhook defaults where unset.
func (s *Service) init(
	notifyCfg shoutrrr.Config,
	whCfg webhook.Config,
	announceChannel chan []byte,
	databaseChannel chan dbtype.Message,
	saveChannel chan bool,
) {
	// Status.
	s.Status.AnnounceChannel = announceChannel
	s.Status.DatabaseChannel = databaseChannel
	s.Status.SaveChannel = saveChannel
	var serviceURL string
	if s.LatestVersion != nil {
		serviceURL = s.LatestVersion.ServiceURL()
	}
	s.Status.Init(
		len(s.Command), len(s.Notify), len(s.WebHook),
		status.ServiceInfo{
			ID:         s.ID,
			Name:       s.Name,
			Comment:    s.Comment,
			ServiceURL: serviceURL,
		},
		&s.Dashboard,
	)

	// Command.
	commandDefaults := util.FirstNonEmptySlice(s.Defaults.Command, s.HardDefaults.Command)
	if len(s.Command) == 0 && len(commandDefaults) != 0 {
		s.Command = make(command.Commands, len(commandDefaults))
		copy(s.Command, commandDefaults)
		s.CommandFromDefaults = true
	}
	if len(s.Command) != 0 {
		s.CommandController = command.NewController(
			&s.Status,
			s.Command,
			s.Notify,
			s.Options.GetIntervalPointer(),
		)
	}

	// Notify.
	notifyDefaults := util.FirstNonEmptyMap(s.Defaults.Notify, s.HardDefaults.Notify)
	if len(s.Notify) == 0 && len(notifyDefaults) != 0 {
		s.Notify = make(shoutrrr.Shoutrrrs, len(notifyDefaults))
		for key := range notifyDefaults {
			s.Notify[key] = &shoutrrr.Shoutrrr{}
		}
		s.NotifyFromDefaults = true
	}
	s.Notify.Init(
		&s.Status,
		notifyCfg,
	)

	// 	If the dashboard icon is not set, use the first icon from a Notify.
	if s.Dashboard.GetIcon() == "" && s.Notify != nil {
		orderedNotifyKeys := util.SortedKeys(s.Notify)
		for _, key := range orderedNotifyKeys {
			// `Params.Icon`
			if icon := util.EvalEnvVars(s.Notify[key].GetParam("icon")); icon != "" &&
				(strings.HasPrefix(icon, "https://") || strings.HasPrefix(icon, "http://")) {
				s.Dashboard.SetFallbackIcon(icon)
				s.Status.RefreshServiceInfo()
				break
			}
		}
	}

	// WebHook.
	webhookDefaults := util.FirstNonEmptyMap(s.Defaults.WebHook, s.HardDefaults.WebHook)
	if s.WebHook == nil && len(webhookDefaults) != 0 {
		s.WebHook = make(webhook.WebHooks, len(webhookDefaults))
		for key := range webhookDefaults {
			s.WebHook[key] = &webhook.WebHook{}
		}
		s.WebHookFromDefaults = true
	}
	s.WebHook.Init(
		&s.Status,
		whCfg,
		&s.Notify,
		s.Options.GetIntervalPointer(),
	)
}

// initMetrics registers all Prometheus metrics for the Service.
func (s *Service) initMetrics() {
	metric.ServiceCountCurrentAdd(s.Options.Active, 1)
	if !s.Options.GetActive() {
		return
	}

	if s.LatestVersion != nil {
		s.LatestVersion.InitMetrics(s.LatestVersion)
	}
	if s.DeployedVersionLookup != nil {
		s.DeployedVersionLookup.InitMetrics(s.DeployedVersionLookup)
	}
	s.Notify.InitMetrics()
	s.CommandController.InitMetrics()
	s.WebHook.InitMetrics()
	s.Status.InitMetrics()
}

// DeleteMetrics removes all Prometheus metrics for the Service.
func (s *Service) deleteMetrics() {
	metric.ServiceCountCurrentAdd(s.Options.Active, -1)
	if !s.Options.GetActive() {
		return
	}

	if s.LatestVersion != nil {
		s.LatestVersion.DeleteMetrics(s.LatestVersion)
	}
	if s.DeployedVersionLookup != nil {
		s.DeployedVersionLookup.DeleteMetrics(s.DeployedVersionLookup)
	}
	s.Notify.DeleteMetrics()
	s.CommandController.DeleteMetrics()
	s.WebHook.DeleteMetrics()
	s.Status.DeleteMetrics()
}
