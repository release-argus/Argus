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

//go:build unit

package metric

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

var packageName = "metric"

func TestInitPrometheusCounterVec(t *testing.T) {
	// GIVEN a metric.
	tests := map[string]struct {
		metric *prometheus.CounterVec
		// id, service_id, type, result.
		args     []string
		ordering []int
	}{
		"LatestVersionQueryResultTotal": {
			metric:   LatestVersionQueryResultTotal,
			args:     []string{"ID", "", "TYPE", "RESULT"},
			ordering: []int{0, 1, 2}},
		"DeployedVersionQueryResultTotal": {
			metric:   DeployedVersionQueryResultTotal,
			args:     []string{"ID", "", "TYPE", "RESULT"},
			ordering: []int{0, 1, 2}},
		"CommandResultTotal": {
			metric:   CommandResultTotal,
			args:     []string{"COMMAND_ID", "SERVICE_ID", "", "RESULT"},
			ordering: []int{0, 2, 1}},
		"NotifyResultTotal": {
			metric:   NotifyResultTotal,
			args:     []string{"NOTIFY_ID", "SERVICE_ID", "TYPE", "RESULT"},
			ordering: []int{0, 3, 1, 2}},
		"WebHookResultTotal": {
			metric:   WebHookResultTotal,
			args:     []string{"WEBHOOK_ID", "SERVICE_ID", "", "RESULT"},
			ordering: []int{0, 2, 1}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Make args unique and remove empty strings.
			unorderedArgs := make([]string, 0, len(tc.args))
			for i := range tc.args {
				if tc.args[i] != "" {
					tc.args[i] += "---" + name
					unorderedArgs = append(unorderedArgs, tc.args[i])
				}
			}
			// Order the args.
			args := make([]string, len(unorderedArgs))
			for i, j := range tc.ordering {
				args[i] = unorderedArgs[j]
			}
			got := testutil.CollectAndCount(tc.metric)
			want := 0
			if got != want {
				t.Errorf("%s\nmetric count mismatch before InitPrometheusCounter()\nwant: %d metrics\ngot:  %d",
					packageName, want, got)
			}

			// WHEN it's initialised with InitPrometheusCounter.
			// id, service_id, type, result.
			InitPrometheusCounter(tc.metric,
				tc.args[0],
				tc.args[1],
				tc.args[2],
				tc.args[3])
			got = testutil.CollectAndCount(tc.metric)
			want = 1
			if got != want {
				t.Errorf("%s\nmetric count mismatch after InitPrometheusCounter()\nwant: %d metrics\ngot:  %d",
					packageName, want, got)
			}
			var wantValue float64
			var gotValue float64
			gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(args...))
			if gotValue != wantValue {
				t.Errorf("%s\nvalue mismatch after InitPrometheusCounter()\nwant: %f\ngot:  %f",
					packageName, wantValue, gotValue)
			}

			// THEN it can be increased.
			IncPrometheusCounter(tc.metric,
				tc.args[0],
				tc.args[1],
				tc.args[2],
				tc.args[3])
			gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(args...))
			wantValue++
			if gotValue != wantValue {
				t.Errorf("%s\nvalue mismatch after IncPrometheusCounter()\nwant: %f\ngot:  %f",
					packageName, wantValue, gotValue)
			}

			// AND it can be deleted.
			DeletePrometheusCounter(tc.metric,
				tc.args[0],
				tc.args[1],
				tc.args[2],
				tc.args[3])
			gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(args...))
			wantValue = 0
			if gotValue != wantValue {
				t.Errorf("%s\nvalue mismatch after DeletePrometheusCounter()\nwant: %f\ngot:  %f",
					packageName, wantValue, gotValue)
			}
		})
	}
}

func TestPrometheusGaugeVec(t *testing.T) {
	// GIVEN a metric.
	tests := map[string]struct {
		metric     *prometheus.GaugeVec
		args       []string
		isGaugeVec bool
		value      float64
	}{
		"LatestVersionQueryResultLast": {
			metric: LatestVersionQueryResultLast,
			args:   []string{"SERVICE_ID", "TYPE"}},
		"DeployedVersionQueryResultLast": {
			metric: DeployedVersionQueryResultLast,
			args:   []string{"SERVICE_ID", "TYPE"}},
		"LatestVersionIsDeployed": {
			metric: LatestVersionIsDeployed,
			args:   []string{"SERVICE_ID", ""}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Make args unique and remove empty strings.
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
				t.Errorf("%s\nmetric count mismatch before InitPrometheusGauge()\nwant: %d metrics\ngot:  %d",
					packageName, want, got)
			}

			// WHEN it's initialised with SetPrometheusGauge.
			wantValue := float64(3)
			SetPrometheusGauge(tc.metric,
				tc.args[0],
				tc.args[1],
				wantValue)
			got = testutil.CollectAndCount(tc.metric)
			want = 1
			if got != want {
				t.Errorf("%s\nmetric count mismatch after SetPrometheusGauge()\nwant: %d metrics\ngot:  %d",
					packageName, want, got)
			}
			gotValue := testutil.ToFloat64(tc.metric.WithLabelValues(args...))
			if gotValue != wantValue {
				t.Errorf("%s\nvalue mismatch after SetPrometheusGauge()\nwant: %f\ngot:  %f",
					packageName, wantValue, gotValue)
			}

			// THEN changes can be noticed.
			wantValue = float64(2)
			SetPrometheusGauge(tc.metric,
				tc.args[0],
				tc.args[1],
				wantValue)
			gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(args...))
			if gotValue != wantValue {
				t.Errorf("%s\nvalue mismatch after SetPrometheusGauge()\nwant: %f\ngot:  %f",
					packageName, wantValue, gotValue)
			}

			// AND it can be deleted.
			DeletePrometheusGauge(tc.metric,
				tc.args[0],
				tc.args[1])
			wantValue = float64(0)
			gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(args...))
			if gotValue != wantValue {
				t.Errorf("%s\nvalue mismatch after DeletePrometheusGauge()\nwant: %f\ngot:  %f",
					packageName, wantValue, gotValue)
			}
		})
	}
}

func TestMetricsAndVersionState(t *testing.T) {
	// GIVEN the Prometheus metrics are initialized.
	InitMetrics()

	tests := map[string]struct {
		approvedVersion string
		latestVersion   string
		deployedVersion string
		expectedState   float64
		expectedMetrics map[string]float64
	}{
		"latest version deployed": {
			approvedVersion: "1.0.0",
			latestVersion:   "1.2.0",
			deployedVersion: "1.2.0",
			expectedState:   1, // Latest version deployed.
			expectedMetrics: map[string]float64{
				"AVAILABLE": 0,
				"SKIPPED":   0,
			},
		},
		"latest version approved": {
			approvedVersion: "1.2.0",
			latestVersion:   "1.2.0",
			deployedVersion: "1.0.0",
			expectedState:   2, // Latest version approved.
			expectedMetrics: map[string]float64{
				"AVAILABLE": 1,
				"SKIPPED":   0,
			},
		},
		"latest version skipped": {
			approvedVersion: "SKIP_1.2.0",
			latestVersion:   "1.2.0",
			deployedVersion: "1.0.0",
			expectedState:   3, // Latest version skipped.
			expectedMetrics: map[string]float64{
				"AVAILABLE": 1,
				"SKIPPED":   1,
			},
		},
		"latest version neither deployed nor approved nor skipped": {
			approvedVersion: "1.0.0",
			latestVersion:   "1.2.0",
			deployedVersion: "1.1.0",
			expectedState:   0, // Latest version not deployed/approved/skipped.
			expectedMetrics: map[string]float64{
				"AVAILABLE": 1,
				"SKIPPED":   0,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// Reset metrics for each test.
			InitMetrics()

			// WHEN GetVersionDeployedState is called.
			state := GetVersionDeployedState(
				tc.approvedVersion,
				tc.latestVersion,
				tc.deployedVersion)

			// THEN the returned state should match the expected state.
			if state != tc.expectedState {
				t.Errorf("%s\nGetVersionDeployedState(%q, %q, %q)\nwant: %v\ngot:  %v",
					packageName,
					tc.approvedVersion, tc.latestVersion, tc.deployedVersion,
					tc.expectedState, state)
			}

			// WHEN SetUpdatesCurrent is called.
			SetUpdatesCurrent(1, state)

			// THEN metrics should match the expected values.
			for label, expected := range tc.expectedMetrics {
				metric := testutil.ToFloat64(UpdatesCurrent.WithLabelValues(label))
				if metric != expected {
					t.Errorf("%s\nUpdatesCurrent[%q]\nwant: %v\ngot:  %v",
						packageName, label,
						expected, metric)
				}
			}

			// WHEN SetUpdatesCurrent is called again with the inverse delta.
			SetUpdatesCurrent(-1, state)

			// THEN metrics should reset back to 0.
			for label := range tc.expectedMetrics {
				metric := testutil.ToFloat64(UpdatesCurrent.WithLabelValues(label))
				var want float64 = 0
				if metric != want {
					t.Errorf("%s\nUpdatesCurrent[%q]\nwant: %f\ngot:  %f",
						packageName, label,
						want, metric)
				}
			}
		})
	}
}
