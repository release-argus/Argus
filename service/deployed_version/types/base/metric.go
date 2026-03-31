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
		metric.ActionResultSuccess)
	metric.InitPrometheusCounter(metric.DeployedVersionQueryResultTotal,
		l.GetServiceID(),
		"",
		lookupType,
		metric.ActionResultFail)
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
		metric.ActionResultSuccess)
	metric.DeletePrometheusCounter(metric.DeployedVersionQueryResultTotal,
		l.GetServiceID(),
		"",
		lookupType,
		metric.ActionResultFail)
}

// QueryMetrics sets the Prometheus metrics for the DeployedVersion query.
func (l *Lookup) QueryMetrics(parentLookup Interface, err error) {
	serviceID := l.GetServiceID()
	serviceType := parentLookup.GetType()
	// Default to success.
	liveness := metric.DeployedVersionQueryResultSuccess
	result := metric.ActionResultSuccess

	// If it failed.
	if err != nil {
		// Increase failure count.
		result = metric.ActionResultFail
		// Liveness.
		liveness = metric.DeployedVersionQueryResultFailed
	}

	// Set liveness.
	metric.SetPrometheusGauge(metric.DeployedVersionQueryResultLast,
		serviceID, serviceType,
		float64(liveness))
	// Increase query result count.
	metric.IncPrometheusCounter(metric.DeployedVersionQueryResultTotal,
		serviceID,
		"",
		serviceType,
		result)
}
