// Copyright [2024] [Argus]
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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus metric.
var (
	// LatestVersionQueryLiveness holds the last latest version query result.
	LatestVersionQueryLiveness = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "latest_version_query_result_last",
		Help: "Whether this service's last latest version query was successful (0=no, 1=yes, 2=no_match__url_command_or_require, 3=semantic_version_fail, 4=progressive_version_fail)."},
		[]string{
			"id",
			"type",
		})
	// LatestVersionQueryMetric counts the amount of times the latest version query has passed or failed.
	LatestVersionQueryMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "latest_version_query_result_total",
		Help: "Number of times the latest version check has passed/failed."},
		[]string{
			"id",
			"type",
			"result",
		})
	// DeployedVersionQueryLiveness holds the state of the latest deployed version query.
	DeployedVersionQueryLiveness = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "deployed_version_query_result_last",
		Help: "Whether this service's last deployed version query was successful (0=no, 1=yes)."},
		[]string{
			"id",
		})
	// DeployedVersionQueryMetric counts the amount of times the deployed version query has passed or failed.
	DeployedVersionQueryMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "deployed_version_query_result_total",
		Help: "Number of times the deployed version check has passed/failed."},
		[]string{
			"id",
			"result",
		})
	// LatestVersionIsDeployed tracks the deployment state of the latest version.
	LatestVersionIsDeployed = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "latest_version_is_deployed",
		Help: "Whether this service's latest version is the same as its deployed version (0=no, 1=yes, 2=approved, 3=skipped)."},
		[]string{
			"id",
		})
	// CommandMetric counts the amount of times a Command has passed or failed.
	CommandMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "command_result_total",
		Help: "Number of times a Command has passed/failed."},
		[]string{
			"id",
			"result",
			"service_id",
		})
	// NotifyMetric counts the amount of times a Notify has passed or failed.
	NotifyMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notify_result_total",
		Help: "Number of times a Notify message has passed/failed."},
		[]string{
			"id",
			"result",
			"service_id",
			"type",
		})
	// WebHookMetric counts the amount of times a WebHook has passed or failed.
	WebHookMetric = promauto.NewCounterVec(prometheus.CounterOpts{
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
//	id
//	result
//
// Optional labels:
//
//	serviceID
//	srcType
func InitPrometheusCounter(
	metric *prometheus.CounterVec,
	id, serviceID, srcType, result string,
) {
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
	metric.With(labels).Add(float64(0))
}

// DeletePrometheusCounter will delete the `metric` counter with the given label(s).
//
// Required labels:
//
//	id
//	result
//
// Optional labels:
//
//	serviceID
//	srcType
func DeletePrometheusCounter(
	metric *prometheus.CounterVec,
	id, serviceID, srcType, result string,
) {
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
	metric.Delete(labels)
}

// IncreasePrometheusCounter will increment the `metric` counter with the given label(s).
//
// Required labels:
//
//	id
//	result
//
// Optional labels:
//
//	serviceID
//	srcType
func IncreasePrometheusCounter(
	metric *prometheus.CounterVec,
	id, serviceID, srcType, result string,
) {
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
	metric.With(labels).Inc()
}

// SetPrometheusGauge will set the `metric` gauge for the given label(s) to `value`.
//
// Required labels:
//
//	id
//
// Optional labels:
//
//	srcType
func SetPrometheusGauge(
	metric *prometheus.GaugeVec,
	id, srcType string,
	value float64,
) {
	labels := prometheus.Labels{
		"id": id,
	}
	if srcType != "" {
		labels["type"] = srcType
	}
	metric.With(labels).Set(value)
}

// DeletePrometheusGauge will delete the `metric` gauge with the given label(s).
//
// Required labels:
//
//	id
//	result
//
// Optional labels:
//
//	serviceID
//	srcType
func DeletePrometheusGauge(
	metric *prometheus.GaugeVec,
	id, srcType string,
) {
	labels := prometheus.Labels{
		"id": id,
	}
	if srcType != "" {
		labels["type"] = srcType
	}
	metric.Delete(labels)
}
