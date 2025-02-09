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

// Package metric provides Prometheus metrics for the Argus service.
package metric

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus metric.
var (
	// ServiceCountCurrent holds the amount of services in the configuration.
	ServiceCountCurrent = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "service_count_current",
		Help: "Number of services in the configuration."})
	// LatestVersionQueryResultLast holds the last latest version query result.
	LatestVersionQueryResultLast = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "latest_version_query_result_last",
		Help: "Whether this service's last latest version query was successful (0=no, 1=yes, 2=no_match__url_command_or_require, 3=semantic_version_fail, 4=progressive_version_fail)."},
		[]string{
			"id",
			"type",
		})
	// LatestVersionQueryResultTotal counts the amount of times the latest version query has passed or failed.
	LatestVersionQueryResultTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "latest_version_query_result_total",
		Help: "Number of times the latest version check has passed/failed."},
		[]string{
			"id",
			"type",
			"result",
		})
	// DeployedVersionQueryResultLast holds the state of the latest deployed version query.
	DeployedVersionQueryResultLast = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "deployed_version_query_result_last",
		Help: "Whether this service's last deployed version query was successful (0=no, 1=yes)."},
		[]string{
			"id",
			"type",
		})
	// DeployedVersionQueryResultTotal counts the amount of times the deployed version query has passed or failed.
	DeployedVersionQueryResultTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "deployed_version_query_result_total",
		Help: "Number of times the deployed version check has passed/failed."},
		[]string{
			"id",
			"type",
			"result",
		})
	// LatestVersionIsDeployed tracks the deployment state of the latest version.
	LatestVersionIsDeployed = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "latest_version_is_deployed",
		Help: "Whether this service's latest version is the same as its deployed version (0=no, 1=yes, 2=approved, 3=skipped)."},
		[]string{
			"id",
		})
	// UpdatesCurrent tracks the count of updates available/skipped.
	UpdatesCurrent = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "updates_current",
		Help: "Total number of updates available/skipped."},
		[]string{
			"type"})
	// CommandResultTotal counts the amount of times a Command has passed or failed.
	CommandResultTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "command_result_total",
		Help: "Number of times a Command has passed/failed."},
		[]string{
			"id",
			"result",
			"service_id",
		})
	// NotifyResultTotal counts the amount of times a Notify has passed or failed.
	NotifyResultTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notify_result_total",
		Help: "Number of times a Notify message has passed/failed."},
		[]string{
			"id",
			"result",
			"service_id",
			"type",
		})
	// WebHookResultTotal counts the amount of times a WebHook has passed or failed.
	WebHookResultTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "webhook_result_total",
		Help: "Number of times a WebHook has passed/failed."},
		[]string{
			"id",
			"result",
			"service_id",
		})
)

// InitPrometheusCounter will set the `metric` counter for the given label(s) to 0.
//
// Required labels:
//
//	id: id
//	result: result
//
// Optional labels:
//
//	serviceID: service_id
//	srcType: type
func InitPrometheusCounter(
	metric *prometheus.CounterVec,
	id, serviceID, srcType, result string,
) {
	metric.With(mergeCounterLabels(id, serviceID, srcType, result)).Add(0)
}

// DeletePrometheusCounter will delete the `metric` counter with the given label(s).
//
// Required labels:
//
//	id: id
//	result: result
//
// Optional labels:
//
//	serviceID: service_id
//	srcType: type
func DeletePrometheusCounter(
	metric *prometheus.CounterVec,
	id, serviceID, srcType, result string,
) {
	metric.Delete(mergeCounterLabels(id, serviceID, srcType, result))
}

// IncPrometheusCounter will increment the `metric` counter with the given label(s).
//
// Required labels:
//
//	id: id
//	result: result
//
// Optional labels:
//
//	serviceID: service_id
//	srcType: type
func IncPrometheusCounter(
	metric *prometheus.CounterVec,
	id, serviceID, srcType, result string,
) {
	metric.With(mergeCounterLabels(id, serviceID, srcType, result)).Inc()
}

// mergeCounterLabels creates a prometheus.Labels map with common labels for counters.
func mergeCounterLabels(
	id, serviceID, srcType, result string,
) prometheus.Labels {
	labels := prometheus.Labels{
		"id":     id,
		"result": result,
	}
	if serviceID != "" {
		labels["service_id"] = serviceID
	}
	if srcType != "" {
		labels["type"] = srcType
	}
	return labels
}

// SetPrometheusGauge will set the `metric` gauge for the given label(s) to `value`.
//
// Required labels:
//
//	id: id
//
// Optional labels:
//
//	srcType: type
func SetPrometheusGauge(
	metric *prometheus.GaugeVec,
	id, srcType string,
	value float64,
) {
	metric.With(mergeGaugeLabels(id, srcType)).Set(value)
}

// DeletePrometheusGauge will delete the `metric` gauge with the given label(s).
//
// Required labels:
//
//	id: id
//
// Optional labels:
//
//	srcType: type
func DeletePrometheusGauge(
	metric *prometheus.GaugeVec,
	id, srcType string,
) {
	metric.Delete(mergeGaugeLabels(id, srcType))
}

// mergeGaugeLabels creates a prometheus.Labels map with common labels for gauges.
func mergeGaugeLabels(id, srcType string) prometheus.Labels {
	labels := prometheus.Labels{
		"id": id,
	}
	if srcType != "" {
		labels["type"] = srcType
	}
	return labels
}

// getLatestVersionIsDeployedState determines the deployment state of the latest version.
//
// Returns:
// - 1: The latest version is deployed (latestVersion matches deployedVersion).
// - 2: The latest version is approved (approvedVersion matches latestVersion).
// - 3: The latest version is skipped (approvedVersion is SKIP_latestVersion).
// - 0: The latest version is neither deployed, approved, nor skipped.
func GetVersionDeployedState(approvedVersion, latestVersion, deployedVersion string) float64 {
	switch {
	case latestVersion == deployedVersion:
		return 1 // Latest version is deployed.
	case approvedVersion == latestVersion:
		return 2 // Latest version is approved.
	case strings.HasSuffix(approvedVersion, latestVersion):
		return 3 // Latest version is skipped.
	default:
		return 0 // Latest version is not deployed/approved/skipped.
	}
}

// SetUpdatesCurrent updates the UpdatesCurrent Prometheus metric with the given delta.
// The metric is updated based on the given result value, which indicates the status:
//   - 0: Latest version not deployed/approved/skipped.
//   - 1: Latest version deployed (does not modify metric).
//   - 2: Latest version approved.
//   - 3: Latest version skipped.
func SetUpdatesCurrent(delta, result float64) {
	switch result {
	case 0, 2:
		UpdatesCurrent.WithLabelValues("AVAILABLE").Add(delta)
	case 3:
		UpdatesCurrent.WithLabelValues("AVAILABLE").Add(delta)
		UpdatesCurrent.WithLabelValues("SKIPPED").Add(delta)
	}
}

// InitMetrics will initialise all global metrics.
func InitMetrics() {
	// service_count.
	ServiceCountCurrent.Set(0)
	// updates_current.
	UpdatesCurrent.With(prometheus.Labels{
		"type": "AVAILABLE"}).Set(0)
	UpdatesCurrent.With(prometheus.Labels{
		"type": "SKIPPED"}).Set(0)
}
