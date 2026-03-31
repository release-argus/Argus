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

// Package base provides the base struct for latest_version lookups.
package base

import (
	"strings"

	"github.com/release-argus/Argus/web/metric"
)

// InitMetrics for this Lookup.
func (l *Lookup) InitMetrics(parentLookup Interface) {
	lookupType := parentLookup.GetType()
	serviceID := l.GetServiceID()

	// ############
	// # Counters #
	// ############
	metric.InitPrometheusCounter(metric.LatestVersionQueryResultTotal,
		serviceID,
		"",
		lookupType,
		metric.ActionResultSuccess)
	metric.InitPrometheusCounter(metric.LatestVersionQueryResultTotal,
		serviceID,
		"",
		lookupType,
		metric.ActionResultFail)
}

// DeleteMetrics for this Lookup.
func (l *Lookup) DeleteMetrics(parentLookup Interface) {
	lookupType := parentLookup.GetType()
	serviceID := l.GetServiceID()

	// ############
	// #  Gauges  #
	// ############
	// Liveness.
	metric.DeletePrometheusGauge(metric.LatestVersionQueryResultLast,
		serviceID,
		lookupType,
	)

	// ############
	// # Counters #
	// ############
	metric.DeletePrometheusCounter(metric.LatestVersionQueryResultTotal,
		serviceID,
		"",
		lookupType,
		metric.ActionResultSuccess)
	metric.DeletePrometheusCounter(metric.LatestVersionQueryResultTotal,
		serviceID,
		"",
		lookupType,
		metric.ActionResultFail)
}

// QueryMetrics sets the Prometheus metrics for the LatestVersion query.
func (l *Lookup) QueryMetrics(parentLookup Interface, err error) {
	serviceID := l.GetServiceID()
	serviceType := parentLookup.GetType()
	// Default to success.
	liveness := metric.LatestVersionQueryResultSuccess
	result := metric.ActionResultSuccess

	// If it failed.
	if err != nil {
		// Increase failure count.
		result = metric.ActionResultFail
		// Liveness.
		switch e := err.Error(); {
		case strings.HasPrefix(e, "no releases were found matching"):
			liveness = metric.LatestVersionQueryResultNoMatch
		case strings.HasPrefix(e, "failed to convert") && strings.Contains(e, " semantic version."):
			liveness = metric.LatestVersionQueryResultSemanticVersionFail
		case strings.HasPrefix(e, "queried version") && strings.Contains(e, " less than "):
			liveness = metric.LatestVersionQueryResultProgressiveVersionFail
		default:
			liveness = metric.LatestVersionQueryResultFailed
		}
	}

	// Set liveness.
	metric.SetPrometheusGauge(metric.LatestVersionQueryResultLast,
		serviceID, serviceType,
		float64(liveness))
	// Increase query result count.
	metric.IncPrometheusCounter(metric.LatestVersionQueryResultTotal,
		serviceID,
		"",
		serviceType,
		result)
}
