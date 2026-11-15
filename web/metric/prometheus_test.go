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

//go:build unit

package metric

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/release-argus/Argus/internal/test"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
)

func TestPrometheusCounterVec(t *testing.T) {
	tests := []struct {
		name   string
		metric *prometheus.CounterVec
		// id, service_id, type, result.
		args     []string
		ordering []int
	}{
		{
			name:     "LatestVersionQueryResultTotal",
			metric:   LatestVersionQueryResultTotal,
			args:     []string{"ID", "", "TYPE", "RESULT"},
			ordering: []int{0, 1, 2},
		},
		{
			name:     "DeployedVersionQueryResultTotal",
			metric:   DeployedVersionQueryResultTotal,
			args:     []string{"ID", "", "TYPE", "RESULT"},
			ordering: []int{0, 1, 2},
		},
		{
			name:     "CommandResultTotal",
			metric:   CommandResultTotal,
			args:     []string{"COMMAND_ID", "SERVICE_ID", "", "RESULT"},
			ordering: []int{0, 2, 1},
		},
		{
			name:     "NotifyResultTotal",
			metric:   NotifyResultTotal,
			args:     []string{"NOTIFY_ID", "SERVICE_ID", "TYPE", "RESULT"},
			ordering: []int{0, 3, 1, 2},
		},
		{
			name:     "WebhookResultTotal",
			metric:   WebhookResultTotal,
			args:     []string{"WEBHOOK_ID", "SERVICE_ID", "", "RESULT"},
			ordering: []int{0, 2, 1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Make args unique and remove empty strings.
			labels := make([]string, len(tc.args))
			unorderedArgs := make([]string, 0, len(tc.args))
			for i, arg := range tc.args {
				if arg != "" {
					labels[i] = arg + "---" + tc.name
					unorderedArgs = append(unorderedArgs, labels[i])
				}
			}
			args := make([]string, len(unorderedArgs))
			for i, j := range tc.ordering {
				args[i] = unorderedArgs[j]
			}
			test.AssertMetricCount(t, "before InitPrometheusCounter: metric", tc.metric, 0)

			// id, service_id, type, result.
			InitPrometheusCounter(
				tc.metric,
				labels[0],
				labels[1],
				labels[2],
				labels[3],
			)
			test.AssertMetricCount(t, "after InitPrometheusCounter: metric", tc.metric, 1)
			var wantValue float64
			if gotValue := testutil.ToFloat64(tc.metric.WithLabelValues(args...)); gotValue != wantValue {
				t.Errorf(
					"after InitPrometheusCounter: metric mismatch\ngot:  %f\nwant: %f",
					gotValue, wantValue,
				)
			}

			wantValue++
			IncPrometheusCounter(
				tc.metric,
				labels[0],
				labels[1],
				labels[2],
				labels[3],
			)
			if gotValue := testutil.ToFloat64(tc.metric.WithLabelValues(args...)); gotValue != wantValue {
				t.Errorf(
					"after IncPrometheusCounter: metric mismatch\ngot:  %f\nwant: %f",
					gotValue, wantValue,
				)
			}

			wantValue = 0
			DeletePrometheusCounter(
				tc.metric,
				labels[0],
				labels[1],
				labels[2],
				labels[3],
			)
			if gotValue := testutil.ToFloat64(tc.metric.WithLabelValues(args...)); gotValue != wantValue {
				t.Errorf(
					"after DeletePrometheusCounter: metric mismatch\ngot:  %f\nwant: %f",
					gotValue, wantValue,
				)
			}
		})
	}
}

func TestPrometheusGaugeVec(t *testing.T) {
	tests := []struct {
		name   string
		metric *prometheus.GaugeVec
		args   []string
	}{
		{
			name:   "LatestVersionQueryResultLast",
			metric: LatestVersionQueryResultLast,
			args:   []string{"SERVICE_ID", "TYPE"},
		},
		{
			name:   "DeployedVersionQueryResultLast",
			metric: DeployedVersionQueryResultLast,
			args:   []string{"SERVICE_ID", "TYPE"},
		},
		{
			name:   "LatestVersionIsDeployed",
			metric: LatestVersionIsDeployed,
			args:   []string{"SERVICE_ID", ""},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Make args unique and remove empty strings.
			labels := make([]string, len(tc.args))
			args := make([]string, 0, len(tc.args))
			for i, arg := range tc.args {
				if arg != "" {
					labels[i] = arg + "---" + tc.name
					args = append(args, labels[i])
				}
			}
			test.AssertMetricCount(t, "before SetPrometheusGauge: metric", tc.metric, 0)

			wantValue := float64(3)
			SetPrometheusGauge(
				tc.metric,
				labels[0],
				labels[1],
				wantValue,
			)

			test.AssertMetricCount(t, "after first SetPrometheusGauge: metric", tc.metric, 1)
			gotValue := testutil.ToFloat64(tc.metric.WithLabelValues(args...))
			if gotValue != wantValue {
				t.Errorf(
					"after first SetPrometheusGauge: metric mismatch\ngot:  %f\nwant: %f",
					gotValue, wantValue,
				)
			}

			wantValue = float64(2)
			SetPrometheusGauge(
				tc.metric,
				labels[0],
				labels[1],
				wantValue,
			)

			gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(args...))
			if gotValue != wantValue {
				t.Errorf(
					"after second SetPrometheusGauge: metric mismatch\ngot:  %f\nwant: %f",
					gotValue, wantValue,
				)
			}

			wantValue = 0
			DeletePrometheusGauge(
				tc.metric,
				labels[0],
				labels[1],
			)

			gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(args...))
			if gotValue != wantValue {
				t.Errorf(
					"after DeletePrometheusGauge: metric mismatch\ngot:  %f\nwant: %f",
					gotValue, wantValue,
				)
			}
		})
	}
}

func TestServiceCountCurrentAdd(t *testing.T) {
	tests := []struct {
		name   string
		delta  int
		active bool
	}{
		{name: "active=true, delta=3", delta: 3, active: true},
		{name: "active=false, delta=3", delta: 3, active: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var targetMetric prometheus.Gauge
			if tc.active {
				targetMetric = ServiceCountCurrent.WithLabelValues(ServiceStateActive)
			} else {
				targetMetric = ServiceCountCurrent.WithLabelValues(ServiceStateInactive)
			}
			hadValue := testutil.ToFloat64(targetMetric)

			ServiceCountCurrentAdd(new(tc.active), tc.delta)

			gotValue := testutil.ToFloat64(targetMetric)
			wantValue := hadValue + float64(tc.delta)
			if gotValue != wantValue {
				t.Errorf("+delta: metric mismatch\ngot:  %f\nwant: %f", gotValue, wantValue)
			}

			ServiceCountCurrentAdd(new(tc.active), -tc.delta)

			hadValue = gotValue
			gotValue = testutil.ToFloat64(targetMetric)
			wantValue = hadValue - float64(tc.delta)
			if gotValue != wantValue {
				t.Errorf("-delta: metric mismatch\ngot:  %f\nwant: %f", gotValue, wantValue)
			}
		})
	}
}

func TestMetricsAndVersionState(t *testing.T) {
	InitMetrics()

	tests := []struct {
		name            string
		approvedVersion string
		latestVersion   string
		deployedVersion string
		expectedState   LatestVersionDeployedState
		expectedMetrics map[string]float64
	}{
		{
			name:            "latest version deployed",
			approvedVersion: "1.0.0",
			latestVersion:   "1.2.0",
			deployedVersion: "1.2.0",
			expectedState:   LatestVersionDeployed,
			expectedMetrics: map[string]float64{
				"AVAILABLE": 0,
				"SKIPPED":   0,
			},
		},
		{
			name:            "latest version approved",
			approvedVersion: "1.2.0",
			latestVersion:   "1.2.0",
			deployedVersion: "1.0.0",
			expectedState:   LatestVersionApproved,
			expectedMetrics: map[string]float64{
				"AVAILABLE": 1,
				"SKIPPED":   0,
			},
		},
		{
			name:            "latest version skipped",
			approvedVersion: serviceinfo.SkippedVersion("1.2.0"),
			latestVersion:   "1.2.0",
			deployedVersion: "1.0.0",
			expectedState:   LatestVersionSkipped,
			expectedMetrics: map[string]float64{
				"AVAILABLE": 1,
				"SKIPPED":   1,
			},
		},
		{
			name:            "latest version neither deployed nor approved nor skipped",
			approvedVersion: "1.0.0",
			latestVersion:   "1.2.0",
			deployedVersion: "1.1.0",
			expectedState:   LatestVersionUnactioned,
			expectedMetrics: map[string]float64{
				"AVAILABLE": 1,
				"SKIPPED":   0,
			},
		},
		{
			name:            "latest version known, deployed version unknown",
			approvedVersion: "",
			latestVersion:   "1.2.3",
			deployedVersion: "",
			expectedState:   LatestVersionUnknown,
			expectedMetrics: map[string]float64{
				"AVAILABLE": 0,
				"SKIPPED":   0,
			},
		},
		{
			name:            "latest version unknown, deployed version known",
			approvedVersion: "",
			latestVersion:   "",
			deployedVersion: "1.2.3",
			expectedState:   LatestVersionUnknown,
			expectedMetrics: map[string]float64{
				"AVAILABLE": 0,
				"SKIPPED":   0,
			},
		},
		{
			name:            "latest+deployed version unknown",
			approvedVersion: "",
			latestVersion:   "",
			deployedVersion: "",
			expectedState:   LatestVersionUnknown,
			expectedMetrics: map[string]float64{
				"AVAILABLE": 0,
				"SKIPPED":   0,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: resets metric.UpdatesCurrent.

			InitMetrics()
			serviceInfo := serviceinfo.ServiceInfo{
				ApprovedVersion: tc.approvedVersion,
				LatestVersion:   tc.latestVersion,
				DeployedVersion: tc.deployedVersion,
			}

			state := GetVersionDeployedState(serviceInfo)

			if state != tc.expectedState {
				t.Errorf(
					"GetVersionDeployedState() mismatch\ngot:  %v\nwant: %v",
					state, tc.expectedState,
				)
			}

			var delta float64 = 1
			SetUpdatesCurrent(delta, state)

			for label, expected := range tc.expectedMetrics {
				metric := testutil.ToFloat64(UpdatesCurrent.WithLabelValues(label))
				if metric != expected {
					t.Errorf(
						"+1: UpdatesCurrent.WithLabelValues(%q) mismatch\ngot:  %v\nwant: %v",
						label, metric, expected,
					)
				}
			}

			delta = -1
			SetUpdatesCurrent(delta, state)

			for label := range tc.expectedMetrics {
				metric := testutil.ToFloat64(UpdatesCurrent.WithLabelValues(label))
				var want float64 = 0
				if metric != want {
					t.Errorf(
						"-1: UpdatesCurrent.WithLabelValues(%q) mismatch\ngot:  %v\nwant: %v",
						label, metric, want,
					)
				}
			}
		})
	}
}
