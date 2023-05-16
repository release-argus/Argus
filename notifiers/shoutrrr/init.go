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

package shoutrrr

import (
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	metric "github.com/release-argus/Argus/web/metrics"
)

// LogInit for this package.
func LogInit(log *util.JLog) {
	jLog = log
}

// Init the Slice metrics amd hand out the defaults.
func (s *Slice) Init(
	serviceStatus *svcstatus.Status,
	mains *SliceDefaults,
	defaults *SliceDefaults,
	hardDefaults *SliceDefaults,
) {
	if s == nil {
		return
	}
	if mains == nil || len(*mains) == 0 {
		mains = &SliceDefaults{}
	}

	for key := range *s {
		id := key
		if (*s)[key] == nil {
			(*s)[key] = &Shoutrrr{}
		}
		(*s)[key].ID = id

		if (*mains)[key] == nil {
			(*mains)[key] = &ShoutrrrDefaults{}
		}

		// Get Type from this, the associated Main, or the ID
		notifyType := util.FirstNonDefault(
			(*s)[key].Type,
			(*mains)[key].Type,
			id)

		// Ensure defaults aren't nil
		if len(*defaults) == 0 {
			defaults = &SliceDefaults{}
		}
		if (*defaults)[notifyType] == nil {
			(*defaults)[notifyType] = &ShoutrrrDefaults{}
		}
		if (*hardDefaults)[notifyType] == nil {
			(*hardDefaults)[notifyType] = &ShoutrrrDefaults{}
		}

		(*s)[key].Init(
			serviceStatus,
			(*mains)[key], (*defaults)[notifyType], (*hardDefaults)[notifyType])
	}
}

// Init the Shoutrrr.
func (s *ShoutrrrBase) Init() {
	s.InitMaps()
}

// Init the Shoutrrr metrics and hand out the defaults.
func (s *Shoutrrr) Init(
	serviceStatus *svcstatus.Status,
	main *ShoutrrrDefaults,
	defaults *ShoutrrrDefaults,
	hardDefaults *ShoutrrrDefaults,
) {
	if s == nil {
		return
	}

	s.InitMaps()
	s.ServiceStatus = serviceStatus

	// Give the matching main
	s.Main = main
	// Create a new main if it's nil and attached to a service
	if main == nil {
		s.Main = &ShoutrrrDefaults{}
	}

	s.Failed = &s.ServiceStatus.Fails.Shoutrrr
	s.Failed.Set(s.ID, nil)

	// Remove the type if it's the same as the main or the type is in the ID
	if s.Type == s.Main.Type || s.Type == s.ID {
		s.Type = ""
	}

	s.Main.Init()

	// Give Defaults
	s.Defaults = defaults
	s.Defaults.Init()

	// Give Hard Defaults
	s.HardDefaults = hardDefaults
	s.HardDefaults.Init()
}

// initOptions mapping, converting all keys to lowercase.
func (s *Shoutrrr) initOptions() {
	s.Options = util.LowercaseStringStringMap(&s.Options)
}

// initURLFields mapping, converting all keys to lowercase.
func (s *Shoutrrr) initURLFields() {
	s.URLFields = util.LowercaseStringStringMap(&s.URLFields)
}

// initParams mapping, converting all keys to lowercase.
func (s *Shoutrrr) initParams() {
	have := map[string]string(s.Params)
	s.Params = util.LowercaseStringStringMap(&have)
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
func (s *ShoutrrrBase) initOptions() {
	s.Options = util.LowercaseStringStringMap(&s.Options)
}

// initURLFields mapping, converting all keys to lowercase.
func (s *ShoutrrrBase) initURLFields() {
	s.URLFields = util.LowercaseStringStringMap(&s.URLFields)
}

// initParams mapping, converting all keys to lowercase.
func (s *ShoutrrrBase) initParams() {
	// have := map[string]string(s.Params)
	s.Params = util.LowercaseStringStringMap(&s.Params)
}

// InitMaps will initialise all maps, converting all keys to lowercase.
func (s *ShoutrrrBase) InitMaps() {
	s.initOptions()
	s.initURLFields()
	s.initParams()
}

// InitMetrics for this Slice.
func (s *Slice) InitMetrics() {
	if s == nil {
		return
	}

	for key := range *s {
		(*s)[key].initMetrics()
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

	for key := range *s {
		(*s)[key].deleteMetrics()
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
