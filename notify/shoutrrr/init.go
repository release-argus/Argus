// Copyright [2024] [Argus]
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

// Package shoutrrr provides the shoutrrr notification service to services.
package shoutrrr

import (
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/web/metric"
)

// LogInit for this package.
func LogInit(log *util.JLog) {
	jLog = log
}

// Init the Slice metrics amd hand out the defaults.
func (s *Slice) Init(
	serviceStatus *status.Status,
	mains *SliceDefaults,
	defaults, hardDefaults *SliceDefaults,
) {
	if s == nil {
		return
	}
	if mains == nil || len(*mains) == 0 {
		mains = &SliceDefaults{}
	}

	for key, shoutrrr := range *s {
		if shoutrrr == nil {
			shoutrrr = &Shoutrrr{}
			(*s)[key] = shoutrrr // Update the map.
		}
		shoutrrr.ID = key

		main := (*mains)[key]
		if main == nil {
			main = &Defaults{}
			main.InitMaps()
		}

		// Get Type from this, the associated Main, or the ID.
		notifyType := util.FirstNonDefault(
			shoutrrr.Type,
			main.Type,
			key)

		// Ensure defaults aren't nil.
		if len(*defaults) == 0 {
			defaults = &SliceDefaults{}
		}
		if (*defaults)[notifyType] == nil {
			(*defaults)[notifyType] = &Defaults{}
		}
		if (*hardDefaults)[notifyType] == nil {
			(*hardDefaults)[notifyType] = &Defaults{}
		}

		shoutrrr.Init(
			serviceStatus,
			main,
			(*defaults)[notifyType], (*hardDefaults)[notifyType])
	}
}

// Init the Shoutrrr.
func (b *Base) Init() {
	b.InitMaps()
}

// Init the Shoutrrr metrics and hand out the defaults.
func (s *Shoutrrr) Init(
	serviceStatus *status.Status,
	main *Defaults,
	defaults, hardDefaults *Defaults,
) {
	if s == nil {
		return
	}

	s.InitMaps()
	s.ServiceStatus = serviceStatus

	// Give the matching main.
	s.Main = main
	// Create a new main if nil.
	if main == nil {
		s.Main = &Defaults{}
	}

	s.Failed = &s.ServiceStatus.Fails.Shoutrrr
	s.Failed.Set(s.ID, nil)

	// Remove the Type if same as the main, or the Type is the ID.
	if s.Type == s.Main.Type || s.Type == s.ID {
		s.Type = ""
	}

	s.Main.Init()

	// Give Defaults.
	s.Defaults = defaults
	s.Defaults.Init()

	// Give Hard Defaults.
	s.HardDefaults = hardDefaults
	s.HardDefaults.Init()
}

// InitMaps will initialise all maps, converting all keys to lowercase.
func (s *Shoutrrr) InitMaps() {
	if s == nil {
		return
	}
	s.initOptions()
	s.initURLFields()
	s.initParams()
}

// initOptions mapping, converting all keys to lowercase.
func (b *Base) initOptions() {
	util.LowercaseStringStringMap(&b.Options)
}

// initURLFields mapping, converting all keys to lowercase.
func (b *Base) initURLFields() {
	util.LowercaseStringStringMap(&b.URLFields)
}

// initParams mapping, converting all keys to lowercase.
func (b *Base) initParams() {
	util.LowercaseStringStringMap(&b.Params)
}

// InitMaps will initialise all maps, converting all keys to lowercase.
func (b *Base) InitMaps() {
	b.initOptions()
	b.initURLFields()
	b.initParams()
}

// InitMetrics for this Slice.
func (s *Slice) InitMetrics() {
	if s == nil {
		return
	}

	for _, shoutrrr := range *s {
		shoutrrr.initMetrics()
	}
}

// initMetrics for this Shoutrrr.
func (s *Shoutrrr) initMetrics() {
	// ############
	// # Counters #
	// ############
	metric.InitPrometheusCounter(metric.NotifyMetric,
		s.ID,
		*s.ServiceStatus.ServiceID,
		s.GetType(),
		"SUCCESS")
	metric.InitPrometheusCounter(metric.NotifyMetric,
		s.ID,
		*s.ServiceStatus.ServiceID,
		s.GetType(),
		"FAIL")
}

// DeleteMetrics for this Slice.
func (s *Slice) DeleteMetrics() {
	if s == nil {
		return
	}

	for _, shoutrrr := range *s {
		shoutrrr.deleteMetrics()
	}
}

// deleteMetrics for this Shoutrrr.
func (s *Shoutrrr) deleteMetrics() {
	metric.DeletePrometheusCounter(metric.NotifyMetric,
		s.ID,
		*s.ServiceStatus.ServiceID,
		s.GetType(),
		"SUCCESS")
	metric.DeletePrometheusCounter(metric.NotifyMetric,
		s.ID,
		*s.ServiceStatus.ServiceID,
		s.GetType(),
		"FAIL")
}
