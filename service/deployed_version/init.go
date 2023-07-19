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

package deployedver

import (
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	metric "github.com/release-argus/Argus/web/metrics"
)

// LogInit for this package.
func LogInit(log *util.JLog) {
	jLog = log
}

// Init will initialise the Service metric.
func (l *Lookup) Init(
	defaults *LookupDefaults,
	hardDefaults *LookupDefaults,
	status *svcstatus.Status,
	options *opt.Options,
) {
	if l == nil {
		return
	}

	l.Defaults = defaults
	l.HardDefaults = hardDefaults
	l.Status = status
	l.Options = options
}

// InitMetrics for this Lookup.
func (l *Lookup) InitMetrics() {
	if l == nil {
		return
	}
	// ############
	// # Counters #
	// ############
	metric.InitPrometheusCounter(metric.DeployedVersionQueryMetric,
		*l.Status.ServiceID,
		"",
		"",
		"SUCCESS")
	metric.InitPrometheusCounter(metric.DeployedVersionQueryMetric,
		*l.Status.ServiceID,
		"",
		"",
		"FAIL")
}

// DeleteMetrics for this Lookup.
func (l *Lookup) DeleteMetrics() {
	if l == nil {
		return
	}

	// Liveness
	metric.DeletePrometheusGauge(metric.DeployedVersionQueryLiveness,
		*l.Status.ServiceID)
	// Counters
	metric.DeletePrometheusCounter(metric.DeployedVersionQueryMetric,
		*l.Status.ServiceID,
		"",
		"",
		"SUCCESS")
	metric.DeletePrometheusCounter(metric.DeployedVersionQueryMetric,
		*l.Status.ServiceID,
		"",
		"",
		"FAIL")
}
