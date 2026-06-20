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
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/release-argus/Argus/internal/test"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
)

var packageName = "metric"

func TestPrometheusCounterVec(t *testing.T) {
	// GIVEN: a metric.
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
			name:     "WebHookResultTotal",
			metric:   WebHookResultTotal,
			args:     []string{"WEBHOOK_ID", "SERVICE_ID", "", "RESULT"},
			ordering: []int{0, 2, 1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Make args unique and remove empty strings.
			unorderedArgs := make([]string, 0, len(tc.args))
			for i := range tc.args {
				if tc.args[i] != "" {
					tc.args[i] += "---" + tc.name
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
				t.Errorf(
					"%s\nmetric count mismatch before InitPrometheusCounter()\ngot:  %d\nwant: %d",
					packageName, got, want,
				)
			}

			prefix := fmt.Sprintf(
				"%s\nInitPrometheusCounter(metric=%q, id=%q, serviceID=%q, srcType=%q, result=%q)",
				packageName, tc.name, tc.args[0], tc.args[1], tc.args[2], tc.args[3],
			)

			// WHEN: it is initialised with InitPrometheusCounter.
			// id, service_id, type, result.
			InitPrometheusCounter(
				tc.metric,
				tc.args[0],
				tc.args[1],
				tc.args[2],
				tc.args[3],
			)
			got = testutil.CollectAndCount(tc.metric)
			want = 1
			if got != want {
				t.Errorf(
					"%s count mismatch\ngot:  %d\nwant: %d",
					prefix, got, want,
				)
			}
			var wantValue float64
			if gotValue := testutil.ToFloat64(tc.metric.WithLabelValues(args...)); gotValue != wantValue {
				t.Errorf(
					"%s value mismatch\ngot:  %f\nwant: %f",
					prefix, gotValue, wantValue,
				)
			}

			prefix = fmt.Sprintf(
				"%s\nIncPrometheusCounter(metric=%q, id=%q, serviceID=%q, srcType=%q, result=%q)",
				packageName, tc.name, tc.args[0], tc.args[1], tc.args[2], tc.args[3],
			)

			// THEN: it can be increased.
			wantValue++
			IncPrometheusCounter(
				tc.metric,
				tc.args[0],
				tc.args[1],
				tc.args[2],
				tc.args[3],
			)
			if gotValue := testutil.ToFloat64(tc.metric.WithLabelValues(args...)); gotValue != wantValue {
				t.Errorf(
					"%s value mismatch\ngot:  %f\nwant: %f",
					prefix, gotValue, wantValue,
				)
			}

			prefix = fmt.Sprintf(
				"%s\nDeletePrometheusCounter(metric=%q, id=%q, serviceID=%q, srcType=%q, result=%q)",
				packageName, tc.name, tc.args[0], tc.args[1], tc.args[2], tc.args[3],
			)

			// AND: it can be deleted.
			wantValue = 0
			DeletePrometheusCounter(
				tc.metric,
				tc.args[0],
				tc.args[1],
				tc.args[2],
				tc.args[3],
			)
			if gotValue := testutil.ToFloat64(tc.metric.WithLabelValues(args...)); gotValue != wantValue {
				t.Errorf(
					"%s value mismatch\ngot:  %f\nwant: %f",
					prefix, gotValue, wantValue,
				)
			}
		})
	}
}

func TestPrometheusGaugeVec(t *testing.T) {
	// GIVEN: a metric.
	tests := []struct {
		name       string
		metric     *prometheus.GaugeVec
		args       []string
		isGaugeVec bool
		value      float64
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
			args := make([]string, 0, len(tc.args))
			for i := range tc.args {
				if tc.args[i] != "" {
					tc.args[i] += "---" + tc.name
					args = append(args, tc.args[i])
				}
			}
			if got, want := testutil.CollectAndCount(tc.metric), 0; got != want {
				t.Errorf(
					"%s\nPrometheusGaugeVec() count mismatch on test start\ngot:  %d\nwant: %d",
					packageName, got, want,
				)
			}

			// WHEN: it is initialised with SetPrometheusGauge.
			wantValue := float64(3)
			SetPrometheusGauge(
				tc.metric,
				tc.args[0],
				tc.args[1],
				wantValue,
			)

			prefix := fmt.Sprintf(
				"%s\nSetPrometheusGauge(id=%q, srcType=%q, value=%d)",
				packageName, tc.name, tc.args[0], int(wantValue),
			)

			// THEN: changes can be noticed.
			if got, want := testutil.CollectAndCount(tc.metric), 1; got != want {
				t.Errorf(
					"%s count mismatch after Set\ngot:  %d\nwant: %d",
					prefix, got, want,
				)
			}
			gotValue := testutil.ToFloat64(tc.metric.WithLabelValues(args...))
			if gotValue != wantValue {
				t.Errorf(
					"%s value mismatch (first Set)\ngot:  %f\nwant: %f",
					prefix, gotValue, wantValue,
				)
			}

			// WHEN: changed.
			wantValue = float64(2)
			SetPrometheusGauge(
				tc.metric,
				tc.args[0],
				tc.args[1],
				wantValue,
			)

			prefix = fmt.Sprintf(
				"%s\nSetPrometheusGauge(id=%q, srcType=%q, value=%d)",
				packageName, tc.name, tc.args[0], int(wantValue),
			)

			// THEN: changes can be noticed.
			gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(args...))
			if gotValue != wantValue {
				t.Errorf(
					"%s value mismatch (second Set)\ngot:  %f\nwant: %f",
					prefix, gotValue, wantValue,
				)
			}

			// AND: it can be deleted.
			wantValue = 0
			DeletePrometheusGauge(
				tc.metric,
				tc.args[0],
				tc.args[1],
			)

			prefix = fmt.Sprintf(
				"%s\nDeletePrometheusGauge(id=%q, srcType=%q)",
				packageName, tc.args[0], tc.args[1],
			)

			gotValue = testutil.ToFloat64(tc.metric.WithLabelValues(args...))
			if gotValue != wantValue {
				t.Errorf(
					"%s value mismatch\ngot:  %f\nwant: %f",
					prefix, gotValue, wantValue,
				)
			}
		})
	}
}

func TestServiceCountCurrentAdd(t *testing.T) {
	// GIVEN: a delta amount and an active state.
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

			// WHEN: ServiceCountCurrentAdd is called with the active state and the delta amount.
			ServiceCountCurrentAdd(test.Ptr(tc.active), tc.delta)

			// THEN: the metric should be incremented by this delta.
			gotValue := testutil.ToFloat64(targetMetric)
			wantValue := hadValue + float64(tc.delta)
			if gotValue != wantValue {
				t.Errorf(
					"%s\nServiceCountCurrentAdd(active=%t, delta=%d) metric mismatch\ngot:  %f\nwant: %f",
					packageName, tc.active, tc.delta,
					gotValue, wantValue,
				)
			}

			// WHEN: ServiceCountCurrentAdd is called again with the inverse delta.
			ServiceCountCurrentAdd(test.Ptr(tc.active), -tc.delta)

			// THEN: the metric should be decremented by this delta.
			hadValue = gotValue
			gotValue = testutil.ToFloat64(targetMetric)
			wantValue = hadValue - float64(tc.delta)
			if gotValue != wantValue {
				t.Errorf(
					"%s\nServiceCountCurrentAdd(active=%t, delta=%d) metric mismatch\ngot:  %f\nwant: %f",
					packageName, tc.active, tc.delta,
					gotValue, wantValue,
				)
			}
		})
	}
}

func TestMetricsAndVersionState(t *testing.T) {
	// GIVEN: the Prometheus metrics are initialised.
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
			expectedState:   LatestVersionDeployed, // Latest version deployed.
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
			expectedState:   LatestVersionApproved, // Latest version approved.
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
			expectedState:   LatestVersionSkipped, // Latest version skipped.
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
			expectedState:   LatestVersionUnactioned, // Latest version not deployed/approved/skipped.
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
		},
		{
			name:            "latest version unknown, deployed version known",
			approvedVersion: "",
			latestVersion:   "",
			deployedVersion: "1.2.3",
			expectedState:   LatestVersionUnknown,
		},
		{
			name:            "latest+deployed version unknown",
			approvedVersion: "",
			latestVersion:   "",
			deployedVersion: "",
			expectedState:   LatestVersionUnknown,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			// Reset metrics for each test.
			InitMetrics()
			serviceInfo := serviceinfo.ServiceInfo{
				ApprovedVersion: tc.approvedVersion,
				LatestVersion:   tc.latestVersion,
				DeployedVersion: tc.deployedVersion,
			}

			// WHEN: GetVersionDeployedState is called.
			state := GetVersionDeployedState(serviceInfo)

			// THEN: the returned state should match the expected state.
			if state != tc.expectedState {
				t.Errorf(
					"%s\nGetVersionDeployedState(%+v) mismatch\ngot:  %v\nwant: %v",
					packageName, serviceInfo,
					state, tc.expectedState,
				)
			}

			// WHEN: SetUpdatesCurrent is called.
			var delta float64 = 1
			SetUpdatesCurrent(delta, state)

			// THEN: metrics should match the expected values.
			for label, expected := range tc.expectedMetrics {
				metric := testutil.ToFloat64(UpdatesCurrent.WithLabelValues(label))
				if metric != expected {
					t.Errorf(
						"%s\nSetUpdatesCurrent(delta=%f, result=%d) metric mismatch for %q\ngot:  %v\nwant: %v",
						packageName, delta, state, label,
						metric, expected,
					)
				}
			}

			// WHEN: SetUpdatesCurrent is called again with the inverse delta.
			delta = -1
			SetUpdatesCurrent(delta, state)

			// THEN: metrics should reset back to 0.
			for label := range tc.expectedMetrics {
				metric := testutil.ToFloat64(UpdatesCurrent.WithLabelValues(label))
				var want float64 = 0
				if metric != want {
					t.Errorf(
						"%s\nSetUpdatesCurrent(delta=%f, result=%d) metric mismatch for %q\ngot:  %f\nwant: %f",
						packageName, delta, state, label,
						metric, want,
					)
				}
			}
		})
	}
}
