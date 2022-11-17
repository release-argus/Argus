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

package latestver

import (
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	metric "github.com/release-argus/Argus/web/metrics"
)

// Init will initialise the Service metric.
func (l *Lookup) Init(
	log *util.JLog,
	defaults *Lookup,
	hardDefaults *Lookup,
	status *svcstatus.Status,
	options *opt.Options,
) {
	jLog = log

	if l.Type == "github" {
		l.GitHubData = &GitHubData{}
	}

	l.Defaults = defaults
	l.HardDefaults = hardDefaults
	l.Status = status
	l.Options = options
	l.initMetrics()
	l.URLCommands.Init(jLog)
	l.Require.Init(log, status)
}

// initMetrics will initialise the Prometheus metric.
func (l *Lookup) initMetrics() {
	// ############
	// # Counters #
	// ############
	metric.InitPrometheusCounterWithIDAndResult(metric.LatestVersionQueryMetric, *l.Status.ServiceID, "SUCCESS")
	metric.InitPrometheusCounterWithIDAndResult(metric.LatestVersionQueryMetric, *l.Status.ServiceID, "FAIL")
}
