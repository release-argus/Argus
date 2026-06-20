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

// Package shoutrrr provides the shoutrrr notification service to services.
package shoutrrr

import (
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/web/metric"
)

// Init initialises the receiver, ensuring maps are non-nil with lowercase keys.
func (b *Base) Init() {
	b.InitMaps()
}

// Init wires defaults, status tracking, and failure state into the Shoutrrr.
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

	// Assign the matching main.
	s.Main = main
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

	s.Defaults = defaults
	s.HardDefaults = hardDefaults
}

// Init assigns defaults and failure tracking to each Shoutrrr in the map.
func (s *Shoutrrrs) Init(serviceStatus *status.Status, cfg Config) {
	if s == nil {
		return
	}

	for key, shoutrrr := range *s {
		if shoutrrr == nil {
			shoutrrr = &Shoutrrr{}
			(*s)[key] = shoutrrr // Update the map.
		}
		shoutrrr.ID = key

		main := cfg.Root[key]
		if main == nil {
			main = &Defaults{}
			main.InitMaps()
		}

		// Get Type from this, the associated Main, or the ID.
		notifyType := util.FirstNonDefault(
			shoutrrr.Type,
			main.Type,
			key,
		)

		// Ensure defaults aren't nil.
		if len(cfg.Defaults) == 0 {
			cfg.Defaults = ShoutrrrsDefaults{}
		}
		if cfg.Defaults[notifyType] == nil {
			cfg.Defaults[notifyType] = &Defaults{}
		}
		if cfg.HardDefaults[notifyType] == nil {
			cfg.HardDefaults[notifyType] = &Defaults{}
		}

		shoutrrr.Init(
			serviceStatus,
			main,
			(cfg.Defaults)[notifyType], (cfg.HardDefaults)[notifyType],
		)
	}
}

// InitMaps initialises all maps and lowercases all keys.
func (b *Base) InitMaps() {
	b.Options = util.EnsureMap(b.Options)
	b.Options = util.LowercaseKeys(b.Options)

	b.URLFields = util.EnsureMap(b.URLFields)
	b.URLFields = util.LowercaseKeys(b.URLFields)

	b.Params = util.EnsureMap(b.Params)
	b.Params = util.LowercaseKeys(b.Params)
}

// InitMaps initialises all maps and lowercases all keys.
func (s *Shoutrrr) InitMaps() {
	if s == nil {
		return
	}
	s.Base.InitMaps()
}

// InitMetrics registers Prometheus counters for all Shoutrrr elements.
func (s *Shoutrrrs) InitMetrics() {
	if s == nil {
		return
	}

	for _, shoutrrr := range *s {
		shoutrrr.initMetrics()
	}
}

// DeleteMetrics removes Prometheus counters for all Shoutrrr elements.
func (s *Shoutrrrs) DeleteMetrics() {
	if s == nil {
		return
	}

	for _, shoutrrr := range *s {
		shoutrrr.deleteMetrics()
	}
}

// initMetrics registers Prometheus counters for Shoutrrr success/failure results.
func (s *Shoutrrr) initMetrics() {
	serviceID := s.ServiceStatus.ServiceInfo.ID

	// ############
	// # Counters #
	// ############
	metric.InitPrometheusCounter(
		metric.NotifyResultTotal,
		s.ID,
		serviceID,
		s.GetType(),
		metric.ActionResultSuccess,
	)
	metric.InitPrometheusCounter(
		metric.NotifyResultTotal,
		s.ID,
		serviceID,
		s.GetType(),
		metric.ActionResultFail,
	)
}

// deleteMetrics removes Prometheus counters for Notify success/failure results.
func (s *Shoutrrr) deleteMetrics() {
	serviceID := s.ServiceStatus.ServiceInfo.ID

	// ############
	// # Counters #
	// ############
	metric.DeletePrometheusCounter(
		metric.NotifyResultTotal,
		s.ID,
		serviceID,
		s.GetType(),
		metric.ActionResultSuccess,
	)
	metric.DeletePrometheusCounter(
		metric.NotifyResultTotal,
		s.ID,
		serviceID,
		s.GetType(),
		metric.ActionResultFail,
	)
}
