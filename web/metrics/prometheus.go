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

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus metrics.
var (
	QueryMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "query_result_total",
		Help: "Number of times the version check has passed/failed.",
	},
		[]string{
			"id",
			"result",
		})
	CommandMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "command_result_total",
		Help: "Number of times a Command has passed/failed.",
	},
		[]string{
			"id",
			"result",
			"service_id",
		})
	NotifyMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notify_result_total",
		Help: "Number of times a Notify message has passed/failed.",
	},
		[]string{
			"id",
			"result",
			"service_id",
			"type",
		})
	WebHookMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "webhook_result_total",
		Help: "Number of times a WebHook has passed/failed.",
	},
		[]string{
			"id",
			"service_id",
			"result",
		})
	QueryLiveness = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "query_result_last",
		Help: "Whether this service's last query was successful (0=no, 1=yes, 2=no_regex_match, 3=semantic_version_fail, 4=progressive_version_fail).",
	},
		[]string{
			"id",
		})
	AckWaiting = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ack_waiting",
		Help: "Whether a new release is waiting to be acknowledged (skipped/approved; 0=no, 1=yes).",
	},
		[]string{
			"id",
		})
)

// InitPrometheusCounterWithIDAndResult will set the `metric` counter for this service to 0.
func InitPrometheusCounterWithIDAndResult(metric *prometheus.CounterVec, id string, result string) {
	metric.With(prometheus.Labels{"id": id, "result": result}).Add(float64(0))
}

// InitPrometheusCounterActions will set the `metric` counter for this service to 0.
func InitPrometheusCounterActions(metric *prometheus.CounterVec, id string, serviceID string, src_type string, result string) {
	if src_type == "" {
		metric.With(prometheus.Labels{"id": id, "service_id": serviceID, "result": result}).Add(float64(0))
	} else {
		metric.With(prometheus.Labels{"id": id, "service_id": serviceID, "type": src_type, "result": result}).Add(float64(0))
	}
}

// IncreasePrometheusCounterWithIDAndResult will increase the `metric` counter for this id.
func IncreasePrometheusCounterWithIDAndResult(metric *prometheus.CounterVec, id string, result string) {
	metric.With(prometheus.Labels{"id": id, "result": result}).Inc()
}

// IncreasePrometheusCounterActions will increase the `metric` counter for this id and serviceID.
func IncreasePrometheusCounterActions(metric *prometheus.CounterVec, id string, serviceID string, src_type string, result string) {
	if src_type == "" {
		metric.With(prometheus.Labels{"id": id, "service_id": serviceID, "result": result}).Inc()
	} else {
		metric.With(prometheus.Labels{"id": id, "service_id": serviceID, "type": src_type, "result": result}).Inc()
	}
}

// SetPrometheusGaugeWithID will set the `metric` gauge for this service to `value`.
func SetPrometheusGaugeWithID(metric *prometheus.GaugeVec, id string, value float64) {
	metric.With(prometheus.Labels{"id": id}).Set(value)
}
