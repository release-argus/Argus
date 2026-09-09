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

// Package webhook provides Webhook functionality to services.
package webhook

import (
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/web/metric"
)

// Init initialises each Webhook with the given status, config, notifiers, and parent interval.
func (w *Webhooks) Init(
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
			webhook = &Webhook{}
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

// init wires status, defaults, failure tracking, and notifiers into the Webhook.
func (w *Webhook) init(
	serviceStatus *status.Status,
	main *Defaults,
	cfg Config,
	shoutrrrNotifiers *shoutrrr.Shoutrrrs,
	parentInterval *string,
) {
	w.ParentInterval = parentInterval
	w.ServiceStatus = serviceStatus

	// Assign the matching main defaults.
	w.Main = main
	if w.Main == nil {
		w.Main = &Defaults{}
	}

	w.Failed = &w.ServiceStatus.Fails.Webhook
	w.SetFail(nil)

	// Remove the type if it matches the main type or matches the ID.
	if w.Type == w.Main.Type || w.ID == w.Type {
		w.Type = ""
	}

	w.Defaults = cfg.Defaults
	w.HardDefaults = cfg.HardDefaults

	// Webhook fail notifiers.
	w.Notifiers = Notifiers{
		Shoutrrr: shoutrrrNotifiers,
	}
}

// InitMetrics registers Prometheus counters for all Webhook elements.
func (w *Webhooks) InitMetrics() {
	if w == nil {
		return
	}

	for _, wh := range *w {
		wh.initMetrics()
	}
}

// DeleteMetrics removes Prometheus counters for all Webhook elements.
func (w *Webhooks) DeleteMetrics() {
	if w == nil {
		return
	}

	for _, wh := range *w {
		wh.deleteMetrics()
	}
}

// initMetrics registers Prometheus counters for Webhook success/failure results.
func (w *Webhook) initMetrics() {
	if w == nil {
		return
	}

	// ############
	// # Counters #
	// ############
	metric.InitPrometheusCounter(
		metric.WebhookResultTotal,
		w.ID,
		w.ServiceStatus.ServiceInfo.ID,
		"",
		metric.ActionResultSuccess,
	)
	metric.InitPrometheusCounter(
		metric.WebhookResultTotal,
		w.ID,
		w.ServiceStatus.ServiceInfo.ID,
		"",
		metric.ActionResultFail,
	)
}

// deleteMetrics removes Prometheus counters for Webhook success/failure results.
func (w *Webhook) deleteMetrics() {
	if w == nil {
		return
	}

	metric.DeletePrometheusCounter(
		metric.WebhookResultTotal,
		w.ID,
		w.ServiceStatus.ServiceInfo.ID,
		"",
		metric.ActionResultSuccess,
	)
	metric.DeletePrometheusCounter(
		metric.WebhookResultTotal,
		w.ID,
		w.ServiceStatus.ServiceInfo.ID,
		"",
		metric.ActionResultFail,
	)
}
