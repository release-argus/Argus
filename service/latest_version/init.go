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

package latestver

import (
	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	metric "github.com/release-argus/Argus/web/metrics"
	"github.com/release-argus/Argus/webhook"
)

// LogInit for this package.
func LogInit(log *util.JLog) {
	jLog = log

	filter.LogInit(log)
	command.LogInit(log)
	shoutrrr.LogInit(log)
	webhook.LogInit(log)
}

// Init the Lookup, assigning Defaults and initialising child structs.
func (l *Lookup) Init(
	defaults *LookupDefaults,
	hardDefaults *LookupDefaults,
	status *svcstatus.Status,
	options *opt.Options,
) {
	if l.Type == "github" {
		l.GitHubData = NewGitHubData("", nil)
	}

	l.Defaults = defaults
	l.HardDefaults = hardDefaults
	l.Status = status
	l.Options = options

	l.Require.Init(status, &defaults.Require)
}

// initMetrics for this Lookup.
func (l *Lookup) InitMetrics() {
	// ############
	// # Counters #
	// ############
	metric.InitPrometheusCounter(metric.LatestVersionQueryMetric,
		*l.Status.ServiceID,
		"",
		"",
		"SUCCESS")
	metric.InitPrometheusCounter(metric.LatestVersionQueryMetric,
		*l.Status.ServiceID,
		"",
		"",
		"FAIL")
}

// DeleteMetrics for this Lookup.
func (l *Lookup) DeleteMetrics() {
	metric.DeletePrometheusCounter(metric.LatestVersionQueryMetric,
		*l.Status.ServiceID,
		"",
		"",
		"SUCCESS")
	metric.DeletePrometheusCounter(metric.LatestVersionQueryMetric,
		*l.Status.ServiceID,
		"",
		"",
		"FAIL")
}
