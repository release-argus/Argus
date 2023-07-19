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

package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus metric.
var (
	// Latest version query successful - 0=no, 1=yes, 2=no_regex_match, 3=semantic_version_fail, 4=progressive_version_fail
	LatestVersionQueryLiveness = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "latest_version_query_result_last",
		Help: "Whether this service's last latest version query was successful (0=no, 1=yes, 2=no_regex_match, 3=semantic_version_fail, 4=progressive_version_fail)."},
		[]string{
			"id",
		})
	// Count of the number of times each latest version query has passed/failed
	LatestVersionQueryMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "latest_version_query_result_total",
		Help: "Number of times the latest version check has passed/failed."},
		[]string{
			"id",
			"result",
		})
	// Lateest deployed version query successful - 0=no, 1=yes
	DeployedVersionQueryLiveness = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "deployed_version_query_result_last",
		Help: "Whether this service's last deployed version query was successful (0=no, 1=yes)."},
		[]string{
			"id",
		})
	// Count of the number of times each deployed version query has passed/failed
	DeployedVersionQueryMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "deployed_version_query_result_total",
		Help: "Number of times the deployed version check has passed/failed."},
		[]string{
			"id",
			"result",
		})
	// Latest version is deployed - 0=no, 1=yes, 2=approved, 3=skipped
	LatestVersionIsDeployed = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "latest_version_is_deployed",
		Help: "Whether this service's latest version is the same as its deployed version (0=no, 1=yes, 2=approved, 3=skipped)."},
		[]string{
			"id",
		})
	// Count of the number of times each Command has passed/failed
	CommandMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "command_result_total",
		Help: "Number of times a Command has passed/failed."},
		[]string{
			"id",
			"result",
			"service_id",
		})
	// Count of the number of times each Notify has passed/failed
	NotifyMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notify_result_total",
		Help: "Number of times a Notify message has passed/failed."},
		[]string{
			"id",
			"result",
			"service_id",
			"type",
		})
	// Count of the number of times each WebHook has passed/failed
	WebHookMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "webhook_result_total",
		Help: "Number of times a WebHook has passed/failed."},
		[]string{
			"id",
			"result",
			"service_id",
		})
)

// InitPrometheusCounter will set the `metric` counter for the given labels to 0.
func InitPrometheusCounter(
	metric *prometheus.CounterVec,
	id string,
	serviceID string,
	srcType string,
	result string,
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

// DeletePrometheusCounter will delete the `metric` counter for the given labels.
func DeletePrometheusCounter(
	metric *prometheus.CounterVec,
	id string,
	serviceID string,
	srcType string,
	result string,
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

// IncreasePrometheusCounter will increement the `metric` counter for the given labels.
func IncreasePrometheusCounter(
	metric *prometheus.CounterVec,
	id string,
	serviceID string,
	srcType string,
	result string,
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

// SetPrometheusGauge will set the `metric` gauge for the given label to `value`.
func SetPrometheusGauge(
	metric *prometheus.GaugeVec,
	id string,
	value float64,
) {
	metric.With(prometheus.Labels{"id": id}).Set(value)
}

// DeletePrometheusGaug will delete the `metric` gauge for the given label.
func DeletePrometheusGauge(
	metric *prometheus.GaugeVec,
	id string,
) {
	metric.Delete(prometheus.Labels{"id": id})
}
