// Copyright [2022] [Argus]
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
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
	metrics "github.com/release-argus/Argus/web/metrics"
)

// Init the Slice metrics.
func (s *Slice) Init(
	log *utils.JLog,
	serviceID *string,
	serviceStatus *service_status.Status,
	mains *Slice,
	defaults *Slice,
	hardDefaults *Slice,
) {
	jLog = log
	if s == nil {
		return
	}
	if mains == nil {
		mains = &Slice{}
	}

	for key := range *s {
		id := key
		if (*s)[key] == nil {
			(*s)[key] = &Shoutrrr{}
		}
		(*s)[key].ID = &id
		(*s)[key].Failed = &serviceStatus.Fails.Shoutrrr

		if len(*mains) == 0 {
			mains = &Slice{}
		}
		if (*mains)[key] == nil {
			(*mains)[key] = &Shoutrrr{}
		}

		// Get Type from this or the associated Main
		notifyType := utils.GetFirstNonDefault(
			(*s)[key].Type,
			(*mains)[key].Type,
		)

		// Ensure defaults aren't nil
		if len(*defaults) == 0 {
			defaults = &Slice{}
		}
		if (*defaults)[notifyType] == nil {
			(*defaults)[notifyType] = &Shoutrrr{}
		}
		if (*hardDefaults)[notifyType] == nil {
			(*hardDefaults)[notifyType] = &Shoutrrr{}
		}

		(*s)[key].Init(serviceID, (*mains)[key], (*defaults)[notifyType], (*hardDefaults)[notifyType])
	}
}

// Init the Shoutrrr metrics and hand out the defaults.
func (s *Shoutrrr) Init(
	serviceID *string,
	main *Shoutrrr,
	defaults *Shoutrrr,
	hardDefaults *Shoutrrr,
) {
	if s == nil {
		return
	}

	s.InitMaps()

	// Give the matching main
	(*s).Main = main
	if main == nil && utils.DefaultIfNil(serviceID) != "" {
		s.Main = &Shoutrrr{}
	}
	s.Main.InitMaps()

	// Give Defaults
	(*s).Defaults = defaults
	s.Defaults.InitMaps()

	// Give Hard Defaults
	(*s).HardDefaults = hardDefaults
	s.HardDefaults.InitMaps()

	s.initMetrics(serviceID)
}

// initOptions mapping, converting all keys to lowercase.
func (s *Shoutrrr) initOptions() {
	s.Options = utils.LowercaseStringStringMap(&s.Options)
}

// initURLFields mapping, converting all keys to lowercase.
func (s *Shoutrrr) initURLFields() {
	s.URLFields = utils.LowercaseStringStringMap(&s.URLFields)
}

// initParams mapping, converting all keys to lowercase.
func (s *Shoutrrr) initParams() {
	have := map[string]string(s.Params)
	s.Params = utils.LowercaseStringStringMap(&have)
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

// initMetrics, giving them all a starting value.
func (s *Shoutrrr) initMetrics(serviceID *string) {
	// Only record metrics for Shoutrrrs attached to a Service
	if s.Main == nil || s.GetType() == "" {
		return
	}

	// ############
	// # Counters #
	// ############
	metrics.InitPrometheusCounterActions(metrics.NotifyMetric, *(*s).ID, *serviceID, s.GetType(), "SUCCESS")
	metrics.InitPrometheusCounterActions(metrics.NotifyMetric, *(*s).ID, *serviceID, s.GetType(), "FAIL")
}

// SetLog will set the logger for the package
func SetLog(log *utils.JLog) {
	jLog = log
}
