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

package deployed_version

import (
	"github.com/release-argus/Argus/service/options"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/web/metrics"
)

// Init will initialise the Service metrics.
func (l *Lookup) Init(
	log *utils.JLog,
	defaults **Lookup,
	hardDefaults **Lookup,
	status *service_status.Status,
	options *options.Options,
) {
	if l == nil {
		return
	}
	jLog = log

	l.Defaults = defaults
	l.HardDefaults = hardDefaults
	l.Status = status
	l.options = options
	l.initMetrics()
}

// initMetrics will initialise the Prometheus metrics.
func (l *Lookup) initMetrics() {
	// ############
	// # Counters #
	// ############
	metrics.InitPrometheusCounterWithIDAndResult(metrics.DeployedVersionQueryMetric, *(*l.Status).ServiceID, "SUCCESS")
	metrics.InitPrometheusCounterWithIDAndResult(metrics.DeployedVersionQueryMetric, *(*l.Status).ServiceID, "FAIL")
}
