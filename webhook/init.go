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
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/web/metric"
)

// LogInit for this package.
func LogInit(log *util.JLog) {
	jLog = log
}

// Init the Slice metrics and hand out the defaults/notifiers.
func (s *Slice) Init(
	serviceStatus *status.Status,
	mains *SliceDefaults,
	defaults, hardDefaults *Defaults,
	shoutrrrNotifiers *shoutrrr.Slice,
	parentInterval *string,
) {
	if s == nil || len(*s) == 0 {
		return
	}
	if mains == nil {
		mains = &SliceDefaults{}
	}

	for id, wh := range *s {
		if wh == nil {
			wh = &WebHook{}
			(*s)[id] = wh // Update the map.
		}
		wh.ID = id
		wh.Init(
			serviceStatus,
			(*mains)[id], defaults, hardDefaults,
			shoutrrrNotifiers,
			parentInterval,
		)
	}
}

// Init the WebHook metrics and give the defaults/notifiers.
func (w *WebHook) Init(
	serviceStatus *status.Status,
	main *Defaults,
	defaults, hardDefaults *Defaults,
	shoutrrrNotifiers *shoutrrr.Slice,
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
	w.Failed.Set(w.ID, nil)

	// Remove the type if it matches the main type or matches the ID.
	if w.Type == w.Main.Type || w.ID == w.Type {
		w.Type = ""
	}

	// Give the defaults.
	w.Defaults = defaults
	w.HardDefaults = hardDefaults

	// WebHook fail notifiers.
	w.Notifiers = &Notifiers{
		Shoutrrr: shoutrrrNotifiers,
	}
}

// InitMetrics of the Slice.
func (s *Slice) InitMetrics() {
	if s == nil {
		return
	}

	for _, wh := range *s {
		wh.initMetrics()
	}
}

// initMetrics, giving them all a starting value.
func (w *WebHook) initMetrics() {
	if w == nil {
		return
	}

	// ############
	// # Counters #
	// ############
	metric.InitPrometheusCounter(metric.WebHookResultTotal,
		w.ID,
		*w.ServiceStatus.ServiceID,
		"",
		"SUCCESS")
	metric.InitPrometheusCounter(metric.WebHookResultTotal,
		w.ID,
		*w.ServiceStatus.ServiceID,
		"",
		"FAIL")
}

// DeleteMetrics of the Slice.
func (s *Slice) DeleteMetrics() {
	if s == nil {
		return
	}

	for _, wh := range *s {
		wh.deleteMetrics()
	}
}

// deleteMetrics of the WebHook.
func (w *WebHook) deleteMetrics() {
	if w == nil {
		return
	}

	metric.DeletePrometheusCounter(metric.WebHookResultTotal,
		w.ID,
		*w.ServiceStatus.ServiceID,
		"",
		"SUCCESS")
	metric.DeletePrometheusCounter(metric.WebHookResultTotal,
		w.ID,
		*w.ServiceStatus.ServiceID,
		"",
		"FAIL")
}
