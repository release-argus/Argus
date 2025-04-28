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

// Package base provides the base struct for deployed_version lookups.
package base

import (
	"github.com/release-argus/Argus/web/metric"
)

// InitMetrics for this Lookup.
func (l *Lookup) InitMetrics(parentLookup Interface) {
	lookupType := parentLookup.GetType()

	// ############
	// # Counters #
	// ############
	metric.InitPrometheusCounter(metric.DeployedVersionQueryResultTotal,
		l.GetServiceID(),
		"",
		lookupType,
		"SUCCESS")
	metric.InitPrometheusCounter(metric.DeployedVersionQueryResultTotal,
		l.GetServiceID(),
		"",
		lookupType,
		"FAIL")
}

// DeleteMetrics for this Lookup.
func (l *Lookup) DeleteMetrics(parentLookup Interface) {
	lookupType := parentLookup.GetType()

	// Liveness.
	metric.DeletePrometheusGauge(metric.DeployedVersionQueryResultLast,
		l.GetServiceID(),
		lookupType,
	)
	// Counters.
	metric.DeletePrometheusCounter(metric.DeployedVersionQueryResultTotal,
		l.GetServiceID(),
		"",
		lookupType,
		"SUCCESS")
	metric.DeletePrometheusCounter(metric.DeployedVersionQueryResultTotal,
		l.GetServiceID(),
		"",
		lookupType,
		"FAIL")
}

// QueryMetrics sets the Prometheus metrics for the DeployedVersion query.
func (l *Lookup) QueryMetrics(parentLookup Interface, err error) {
	// If it failed.
	if err != nil {
		// Increase failure count.
		metric.IncPrometheusCounter(metric.DeployedVersionQueryResultTotal,
			l.GetServiceID(),
			"",
			parentLookup.GetType(),
			"FAIL")
		// Set liveness.
		metric.SetPrometheusGauge(metric.DeployedVersionQueryResultLast,
			l.GetServiceID(), parentLookup.GetType(),
			0)
		// If it succeeded.
	} else {
		metric.IncPrometheusCounter(metric.DeployedVersionQueryResultTotal,
			l.GetServiceID(),
			"",
			parentLookup.GetType(),
			"SUCCESS")
		metric.SetPrometheusGauge(metric.DeployedVersionQueryResultLast,
			l.GetServiceID(), parentLookup.GetType(),
			1)
	}
}
