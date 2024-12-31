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
		// id, service_id, type, result
		args     []string
		ordering []int
	}{
		"LatestVersionQueryMetric": {
			metric:   LatestVersionQueryMetric,
			args:     []string{"ID", "", "TYPE", "RESULT"},
			ordering: []int{0, 1, 2}},
		"DeployedVersionQueryMetric": {
			metric:   DeployedVersionQueryMetric,
			args:     []string{"ID", "", "", "RESULT"},
			ordering: []int{0, 1}},
		"CommandMetric": {
			metric:   CommandMetric,
			args:     []string{"COMMAND_ID", "SERVICE_ID", "", "RESULT"},
			ordering: []int{0, 2, 1}},
		"NotifyMetric": {
			metric:   NotifyMetric,
			args:     []string{"NOTIFY_ID", "SERVICE_ID", "TYPE", "RESULT"},
			ordering: []int{0, 3, 1, 2}},
		"WebHookMetric": {
			metric:   WebHookMetric,
			args:     []string{"WEBHOOK_ID", "SERVICE_ID", "", "RESULT"},
			ordering: []int{0, 2, 1}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Make args unique and remove empty strings
			unorderedArgs := make([]string, 0, len(tc.args))
			for i := range tc.args {
				if tc.args[i] != "" {
					tc.args[i] += "---" + name
					unorderedArgs = append(unorderedArgs, tc.args[i])
				}
			}
			// Order the args
			args := make([]string, len(unorderedArgs))
			for i, j := range tc.ordering {
				args[i] = unorderedArgs[j]
			}
			got := testutil.CollectAndCount(tc.metric)
			want := 0
			if got != want {
				t.Errorf("haven't initialised yet but got %d metrics, expecting %d",
					got, want)
			}

			// WHEN it's initialised with InitPrometheusCounter
			// id, service_id, type, result
			InitPrometheusCounter(tc.metric,
				tc.args[0],
				tc.args[1],
				tc.args[2],
				tc.args[3])
			got = testutil.CollectAndCount(tc.metric)
			want = 1
			if got != want {
				t.Errorf("has been initialised but got %d metrics, expecting %d",
					got, want)
			}
			var wantValue float64
			var gotValue float64
			gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(args...))
			// }
			if gotValue != wantValue {
				t.Errorf("has been initialised but got %f, expecting %f",
					gotValue, wantValue)
			}

			// THEN it can be increased
			IncreasePrometheusCounter(tc.metric,
				tc.args[0],
				tc.args[1],
				tc.args[2],
				tc.args[3])
			gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(args...))
			wantValue++
			if gotValue != wantValue {
				t.Errorf("has been changed but got %f, expecting %f",
					gotValue, wantValue)
			}

			// AND it can be deleted
			DeletePrometheusCounter(tc.metric,
				tc.args[0],
				tc.args[1],
				tc.args[2],
				tc.args[3])
			gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(args...))
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
			args:   []string{"SERVICE_ID", "TYPE"}},
		"DeployedVersionQueryLiveness": {
			metric: DeployedVersionQueryLiveness,
			args:   []string{"SERVICE_ID", ""}},
		"LatestVersionIsDeployed": {
			metric: LatestVersionIsDeployed,
			args:   []string{"SERVICE_ID", ""}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Make args unique and remove empty strings
			args := make([]string, 0, len(tc.args))
			for i := range tc.args {
				if tc.args[i] != "" {
					tc.args[i] += "---" + name
					args = append(args, tc.args[i])
				}
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
				tc.args[1],
				wantValue)
			got = testutil.CollectAndCount(tc.metric)
			want = 1
			if got != want {
				t.Errorf("has been initialised but got %d metrics, expecting %d",
					got, want)
			}
			gotValue := testutil.ToFloat64(tc.metric.WithLabelValues(args...))
			if gotValue != wantValue {
				t.Errorf("has been initialised but got %f, expecting %f",
					gotValue, wantValue)
			}

			// THEN changes can be noticed
			wantValue = float64(2)
			SetPrometheusGauge(tc.metric,
				tc.args[0],
				tc.args[1],
				wantValue)
			gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(args...))
			if gotValue != wantValue {
				t.Errorf("has been changed but got %f, expecting %f",
					gotValue, wantValue)
			}

			// AND it can be deleted
			DeletePrometheusGauge(tc.metric,
				tc.args[0],
				tc.args[1])
			wantValue = float64(0)
			gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(args...))
			if gotValue != wantValue {
				t.Errorf("has been deleted but got %f, expecting %f",
					gotValue, wantValue)
			}
		})
	}
}
