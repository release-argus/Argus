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

// Package webhook provides WebHook functionality to services.
package webhook

import (
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/web/metric"
)

// Init the WebHooks metrics and hand out the defaults/notifiers.
func (w *WebHooks) Init(
	serviceStatus *status.Status,
	cfg Config,
	shoutrrrNotifiers *shoutrrr.Shoutrrrs,
	parentInterval *string,
) {
	if w == nil || len(*w) == 0 {
		return
	}

	for id, webhook := range *w {
		if webhook == nil {
			webhook = &WebHook{}
			(*w)[id] = webhook // Update the map.
		}
		main := cfg.Root[id]

		webhook.ID = id
		webhook.init(
			serviceStatus,
			main,
			cfg,
			shoutrrrNotifiers,
			parentInterval,
		)
	}
}

// init wires dependencies and config into the WebHook and normalises its state.
// It assigns service status tracking, ensures a non-nil Main configuration,
// attaches failure tracking, applies defaults, and configures notifiers.
// It also clears Type when it matches inherited defaults or ID to avoid redundancy.
func (w *WebHook) init(
	serviceStatus *status.Status,
	main *Defaults,
	cfg Config,
	shoutrrrNotifiers *shoutrrr.Shoutrrrs,
	parentInterval *string,
) {
	w.ParentInterval = parentInterval
	w.ServiceStatus = serviceStatus

	// Give the matching main.
	w.Main = main
	// Create an empty Main if nil.
	if w.Main == nil {
		w.Main = &Defaults{}
	}

	w.Failed = &w.ServiceStatus.Fails.WebHook
	w.SetFail(nil)

	// Remove the type if it matches the main type or matches the ID.
	if w.Type == w.Main.Type || w.ID == w.Type {
		w.Type = ""
	}

	// Give the defaults.
	w.Defaults = cfg.Defaults
	w.HardDefaults = cfg.HardDefaults

	// WebHook fail notifiers.
	w.Notifiers = Notifiers{
		Shoutrrr: shoutrrrNotifiers,
	}
}

// InitMetrics registers Prometheus counters for all WebHook elements.
func (w *WebHooks) InitMetrics() {
	if w == nil {
		return
	}

	for _, wh := range *w {
		wh.initMetrics()
	}
}

// DeleteMetrics removes Prometheus counters for all WebHook elements.
func (w *WebHooks) DeleteMetrics() {
	if w == nil {
		return
	}

	for _, wh := range *w {
		wh.deleteMetrics()
	}
}

// initMetrics registers Prometheus counters for WebHook success/failure results.
func (w *WebHook) initMetrics() {
	if w == nil {
		return
	}

	// ############
	// # Counters #
	// ############
	metric.InitPrometheusCounter(
		metric.WebHookResultTotal,
		w.ID,
		w.ServiceStatus.ServiceInfo.ID,
		"",
		metric.ActionResultSuccess,
	)
	metric.InitPrometheusCounter(
		metric.WebHookResultTotal,
		w.ID,
		w.ServiceStatus.ServiceInfo.ID,
		"",
		metric.ActionResultFail,
	)
}

// deleteMetrics removes Prometheus counters for WebHook success/failure results.
func (w *WebHook) deleteMetrics() {
	if w == nil {
		return
	}

	metric.DeletePrometheusCounter(
		metric.WebHookResultTotal,
		w.ID,
		w.ServiceStatus.ServiceInfo.ID,
		"",
		metric.ActionResultSuccess,
	)
	metric.DeletePrometheusCounter(
		metric.WebHookResultTotal,
		w.ID,
		w.ServiceStatus.ServiceInfo.ID,
		"",
		metric.ActionResultFail,
	)
}
