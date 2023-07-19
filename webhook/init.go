// Copyright [2023] [Argus]
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

package webhook

import (
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	metric "github.com/release-argus/Argus/web/metrics"
)

// LogInit for this package.
func LogInit(log *util.JLog) {
	jLog = log
}

// Init the Slice metrics and hand out the defaults/notifiers.
func (w *Slice) Init(
	serviceStatus *svcstatus.Status,
	mains *SliceDefaults,
	defaults *WebHookDefaults,
	hardDefaults *WebHookDefaults,
	shoutrrrNotifiers *shoutrrr.Slice,
	parentInterval *string,
) {
	if w == nil || len(*w) == 0 {
		return
	}
	if mains == nil {
		mains = &SliceDefaults{}
	}

	for id := range *w {
		if (*w)[id] == nil {
			(*w)[id] = &WebHook{}
		}
		(*w)[id].ID = id
		(*w)[id].Init(
			serviceStatus,
			(*mains)[id], defaults, hardDefaults,
			shoutrrrNotifiers,
			parentInterval,
		)
	}
}

// Init the WebHook metrics and give the defaults/notifiers.
func (w *WebHook) Init(
	serviceStatus *svcstatus.Status,
	main *WebHookDefaults,
	defaults *WebHookDefaults,
	hardDefaults *WebHookDefaults,
	shoutrrrNotifiers *shoutrrr.Slice,
	parentInterval *string,
) {
	w.ParentInterval = parentInterval
	w.ServiceStatus = serviceStatus

	// Give the matching main
	w.Main = main
	// Create an empty Main if it's nil
	if w.Main == nil {
		w.Main = &WebHookDefaults{}
	}

	w.Failed = &w.ServiceStatus.Fails.WebHook
	w.Failed.Set(w.ID, nil)

	// Remove the type if it's the same as the main or the type is in the ID
	if w.Type == w.Main.Type || w.ID == w.Type {
		w.Type = ""
	}

	// Give the defaults
	w.Defaults = defaults
	w.HardDefaults = hardDefaults

	// WebHook fail notifiers
	w.Notifiers = &Notifiers{
		Shoutrrr: shoutrrrNotifiers,
	}
}

// InitMetrics of the Slice.
func (w *Slice) InitMetrics() {
	if w == nil {
		return
	}

	for id := range *w {
		(*w)[id].initMetrics()
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
	metric.InitPrometheusCounter(metric.WebHookMetric,
		w.ID,
		*w.ServiceStatus.ServiceID,
		"",
		"SUCCESS")
	metric.InitPrometheusCounter(metric.WebHookMetric,
		w.ID,
		*w.ServiceStatus.ServiceID,
		"",
		"FAIL")
}

// DeleteMetrics of the Slice.
func (w *Slice) DeleteMetrics() {
	if w == nil {
		return
	}

	for id := range *w {
		(*w)[id].deleteMetrics()
	}
}

// deleteMetrics of the WebHook.
func (w *WebHook) deleteMetrics() {
	if w == nil {
		return
	}

	metric.DeletePrometheusCounter(metric.WebHookMetric,
		w.ID,
		*w.ServiceStatus.ServiceID,
		"",
		"SUCCESS")
	metric.DeletePrometheusCounter(metric.WebHookMetric,
		w.ID,
		*w.ServiceStatus.ServiceID,
		"",
		"FAIL")
}
