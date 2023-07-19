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

//go:build unit

package metric

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestInitPrometheusCounterVec(t *testing.T) {
	// GIVEN a metric
	tests := map[string]struct {
		metric *prometheus.CounterVec
		args   []string
	}{
		"LatestVersionQueryMetric": {
			metric: LatestVersionQueryMetric,
			args:   []string{"ID", "RESULT"}},
		"DeployedVersionQueryMetric": {
			metric: DeployedVersionQueryMetric,
			args:   []string{"ID", "RESULT"}},
		"CommandMetric": {
			metric: CommandMetric,
			args:   []string{"COMMAND_ID", "RESULT", "SERVICE_ID"}},
		"NotifyMetric": {
			metric: NotifyMetric,
			args:   []string{"NOTIFY_ID", "RESULT", "SERVICE_ID", "TYPE"}},
		"WebHookMetric": {
			metric: WebHookMetric,
			args:   []string{"WEBHOOK_ID", "RESULT", "SERVICE_ID"}},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for i := range tc.args {
				tc.args[i] += "---" + name
			}
			got := testutil.CollectAndCount(tc.metric)
			want := 0
			if got != want {
				t.Errorf("haven't initialised yet but got %d metrics, expecting %d",
					got, want)
			}

			// WHEN it's initialised with InitPrometheusCounter
			switch args := len(tc.args); {
			case args == 2:
				// id, result
				InitPrometheusCounter(tc.metric,
					tc.args[0],
					"",
					"",
					tc.args[1])
			case args == 3:
				// id, service_id, result
				InitPrometheusCounter(tc.metric,
					tc.args[0],
					tc.args[2],
					"",
					tc.args[1])
			default:
				// id, service_id, type, result
				InitPrometheusCounter(tc.metric,
					tc.args[0],
					tc.args[2],
					tc.args[3],
					tc.args[1])
			}
			got = testutil.CollectAndCount(tc.metric)
			want = 1
			if got != want {
				t.Errorf("has been initialised but got %d metrics, expecting %d",
					got, want)
			}
			var wantValue float64
			var gotValue float64
			switch args := len(tc.args); {
			case args == 2:
				gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(tc.args[0], tc.args[1]))
			case args == 3:
				gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(tc.args[0], tc.args[1], tc.args[2]))
			default:
				gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(tc.args[0], tc.args[1], tc.args[2], tc.args[3]))
			}
			if gotValue != wantValue {
				t.Errorf("has been initialised but got %f, expecting %f",
					gotValue, wantValue)
			}

			// THEN it can be increased
			switch args := len(tc.args); {
			case args == 2:
				IncreasePrometheusCounter(tc.metric,
					tc.args[0],
					"",
					"",
					tc.args[1])
				gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(tc.args[0], tc.args[1]))
			case args == 3:
				IncreasePrometheusCounter(tc.metric,
					tc.args[0],
					tc.args[2],
					"",
					tc.args[1])
				gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(tc.args[0], tc.args[1], tc.args[2]))
			default:
				IncreasePrometheusCounter(tc.metric,
					tc.args[0],
					tc.args[2],
					tc.args[3],
					tc.args[1])
				gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(tc.args[0], tc.args[1], tc.args[2], tc.args[3]))
			}
			wantValue++
			if gotValue != wantValue {
				t.Errorf("has been changed but got %f, expecting %f",
					gotValue, wantValue)
			}

			// AND it can be deleted
			switch args := len(tc.args); {
			case args == 2:
				DeletePrometheusCounter(tc.metric,
					tc.args[0],
					"",
					"",
					tc.args[1])
				gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(tc.args[0], tc.args[1]))
			case args == 3:
				DeletePrometheusCounter(tc.metric,
					tc.args[0],
					tc.args[2],
					"",
					tc.args[1])
				gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(tc.args[0], tc.args[1], tc.args[2]))
			default:
				DeletePrometheusCounter(tc.metric,
					tc.args[0],
					tc.args[2],
					tc.args[3],
					tc.args[1])
				gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(tc.args[0], tc.args[1], tc.args[2], tc.args[3]))
			}
			wantValue = 0
			if gotValue != wantValue {
				t.Errorf("has been deleted but got %f, expecting %f",
					gotValue, wantValue)
			}
		})
	}
}

func TestPrometheusGaugeVec(t *testing.T) {
	// GIVEN a metric
	tests := map[string]struct {
		metric     *prometheus.GaugeVec
		args       []string
		isGaugeVec bool
		value      float64
	}{
		"LatestVersionQueryLiveness": {
			metric: LatestVersionQueryLiveness,
			args:   []string{"SERVICE_ID"}},
		"DeployedVersionQueryLiveness": {
			metric: DeployedVersionQueryLiveness,
			args:   []string{"SERVICE_ID"}},
		"LatestVersionIsDeployed": {
			metric: LatestVersionIsDeployed,
			args:   []string{"SERVICE_ID"}},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for i := range tc.args {
				tc.args[i] += name
			}
			got := testutil.CollectAndCount(tc.metric)
			want := 0
			if got != want {
				t.Errorf("haven't initialised yet but got %d metrics, expecting %d",
					got, want)
			}

			// WHEN it's initialised with SetPrometheusGauge
			wantValue := float64(3)
			SetPrometheusGauge(tc.metric,
				tc.args[0],
				wantValue)
			got = testutil.CollectAndCount(tc.metric)
			want = 1
			if got != want {
				t.Errorf("has been initialised but got %d metrics, expecting %d",
					got, want)
			}
			gotValue := testutil.ToFloat64(tc.metric.WithLabelValues(tc.args[0]))
			if gotValue != wantValue {
				t.Errorf("has been initialised but got %f, expecting %f",
					gotValue, wantValue)
			}

			// THEN changes can be noticed
			wantValue = float64(2)
			SetPrometheusGauge(tc.metric,
				tc.args[0],
				wantValue)
			gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(tc.args[0]))
			if gotValue != wantValue {
				t.Errorf("has been changed but got %f, expecting %f",
					gotValue, wantValue)
			}

			// AND it can be deleted
			DeletePrometheusGauge(tc.metric,
				tc.args[0])
			wantValue = float64(0)
			gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(tc.args[0]))
			if gotValue != wantValue {
				t.Errorf("has been deleted but got %f, expecting %f",
					gotValue, wantValue)
			}
		})
	}
}
