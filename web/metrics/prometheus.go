// Copyright [2022] [Hymenaios]
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
	GotifyMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gotify_result_total",
		Help: "Number of times a Gotify message has passed/failed.",
	},
		[]string{
			"id",
			"service_id",
			"result",
		})
	SlackMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "slack_result_total",
		Help: "Number of times a Slack message has passed/failed.",
	},
		[]string{
			"id",
			"service_id",
			"result",
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

// InitPrometheusCounterWithID will set the `metric` counter for this service to 0.
func InitPrometheusCounterWithID(metric *prometheus.CounterVec, id string) {
	metric.With(prometheus.Labels{"id": id}).Add(float64(0))
}

// InitPrometheusCounterWithIDAndResult will set the `metric` counter for this service to 0.
func InitPrometheusCounterWithIDAndResult(metric *prometheus.CounterVec, id string, result string) {
	metric.With(prometheus.Labels{"id": id, "result": result}).Add(float64(0))
}

// InitPrometheusCounterWithIDAndServiceIDAndResult will set the `metric` counter for this service to 0.
func InitPrometheusCounterWithIDAndServiceIDAndResult(metric *prometheus.CounterVec, id string, serviceID string, result string) {
	metric.With(prometheus.Labels{"id": id, "service_id": serviceID, "result": result}).Add(float64(0))
}

// IncreasePrometheusCounterWithIDAndResult will increase the `metric` counter for this id.
func IncreasePrometheusCounterWithIDAndResult(metric *prometheus.CounterVec, id string, result string) {
	metric.With(prometheus.Labels{"id": id, "result": result}).Inc()
}

// IncreasePrometheusCounterWithIDAndServiceIDAndResult will increase the `metric` counter for this id and serviceID.
func IncreasePrometheusCounterWithIDAndServiceIDAndResult(metric *prometheus.CounterVec, id string, serviceID string, result string) {
	metric.With(prometheus.Labels{"id": id, "service_id": serviceID, "result": result}).Inc()
}

// SetPrometheusGaugeWithID will set the `metric` gauge for this service to `value`.
func SetPrometheusGaugeWithID(metric *prometheus.GaugeVec, id string, value float64) {
	metric.With(prometheus.Labels{"id": id}).Set(value)
}
