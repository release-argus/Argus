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

// Package webhook provides WebHook functionality to services.
package webhook

import (
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/web/metric"
)

// Init the WebHooks metrics and hand out the defaults/notifiers.
func (wh *WebHooks) Init(
	serviceStatus *status.Status,
	mains *WebHooksDefaults,
	defaults, hardDefaults *Defaults,
	shoutrrrNotifiers *shoutrrr.Shoutrrrs,
	parentInterval *string,
) {
	if wh == nil || len(*wh) == 0 {
		return
	}
	if mains == nil {
		mains = &WebHooksDefaults{}
	}

	for id, webhook := range *wh {
		if webhook == nil {
			webhook = &WebHook{}
			(*wh)[id] = webhook // Update the map.
		}
		webhook.ID = id
		webhook.Init(
			serviceStatus,
			(*mains)[id], defaults, hardDefaults,
			shoutrrrNotifiers,
			parentInterval,
		)
	}
}

// Init the WebHook metrics and give the defaults/notifiers.
func (wh *WebHook) Init(
	serviceStatus *status.Status,
	main *Defaults,
	defaults, hardDefaults *Defaults,
	shoutrrrNotifiers *shoutrrr.Shoutrrrs,
	parentInterval *string,
) {
	wh.ParentInterval = parentInterval
	wh.ServiceStatus = serviceStatus

	// Give the matching main.
	wh.Main = main
	// Create an empty Main if nil.
	if wh.Main == nil {
		wh.Main = &Defaults{}
	}

	wh.Failed = &wh.ServiceStatus.Fails.WebHook
	wh.Failed.Set(wh.ID, nil)

	// Remove the type if it matches the main type or matches the ID.
	if wh.Type == wh.Main.Type || wh.ID == wh.Type {
		wh.Type = ""
	}

	// Give the defaults.
	wh.Defaults = defaults
	wh.HardDefaults = hardDefaults

	// WebHook fail notifiers.
	wh.Notifiers = &Notifiers{
		Shoutrrr: shoutrrrNotifiers,
	}
}

// InitMetrics of the WebHooks.
func (wh *WebHooks) InitMetrics() {
	if wh == nil {
		return
	}

	for _, wh := range *wh {
		wh.initMetrics()
	}
}

// initMetrics, giving them all a starting value.
func (wh *WebHook) initMetrics() {
	if wh == nil {
		return
	}

	// ############
	// # Counters #
	// ############
	metric.InitPrometheusCounter(metric.WebHookResultTotal,
		wh.ID,
		wh.ServiceStatus.ServiceInfo.ID,
		"",
		"SUCCESS")
	metric.InitPrometheusCounter(metric.WebHookResultTotal,
		wh.ID,
		wh.ServiceStatus.ServiceInfo.ID,
		"",
		"FAIL")
}

// DeleteMetrics of the WebHooks.
func (wh *WebHooks) DeleteMetrics() {
	if wh == nil {
		return
	}

	for _, wh := range *wh {
		wh.deleteMetrics()
	}
}

// deleteMetrics of the WebHook.
func (wh *WebHook) deleteMetrics() {
	if wh == nil {
		return
	}

	metric.DeletePrometheusCounter(metric.WebHookResultTotal,
		wh.ID,
		wh.ServiceStatus.ServiceInfo.ID,
		"",
		"SUCCESS")
	metric.DeletePrometheusCounter(metric.WebHookResultTotal,
		wh.ID,
		wh.ServiceStatus.ServiceInfo.ID,
		"",
		"FAIL")
}
