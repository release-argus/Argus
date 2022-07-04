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
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestInitPrometheusCounterWithIDAndResult(t *testing.T) {
	// GIVEN a QueryMetric
	metric := QueryMetric

	// WHEN it's initialised with InitPrometheusCounterWithIDAndResult
	InitPrometheusCounterWithIDAndResult(metric, "SERVICE_ID", "RESULT")

	// THEN it can be collected
	got := testutil.CollectAndCount(metric)
	want := 1
	if got != want {
		t.Errorf("%d QueryMetric's were initialised, expecting %d",
			got, want)
	}
}

func TestIncreasePrometheusCounterWithIDAndResult(t *testing.T) {
	// GIVEN a QueryMetric that's been initialised
	metric := QueryMetric
	InitPrometheusCounterWithIDAndResult(metric, "SERVICE_ID", "RESULT")

	// WHEN it's incremented with IncreasePrometheusCounterWithIDAndResult
	IncreasePrometheusCounterWithIDAndResult(metric, "SERVICE_ID", "RESULT")

	// THEN this increase can be collected
	got := testutil.ToFloat64(metric.WithLabelValues("SERVICE_ID", "RESULT"))
	want := float64(1)
	if got != want {
		t.Errorf("QueryMetric was incremented. Got %g, want %g",
			got, want)
	}
}

func TestInitPrometheusCounterActionsWithWebHook(t *testing.T) {
	// GIVEN a WebHookMetric
	metric := WebHookMetric

	// WHEN it's initialised with InitPrometheusCounterActions
	InitPrometheusCounterActions(metric, "WEBHOOK_ID", "SERVICE_ID", "", "RESULT")

	// THEN it can be collected
	got := testutil.CollectAndCount(metric)
	want := 1
	if got != want {
		t.Errorf("%d Metrics's were initialised, expecting %d",
			got, want)
	}
}

func TestIncreasePrometheusCounterActionsWithWebHook(t *testing.T) {
	// GIVEN a WebHookMetric that's been initialised
	metric := WebHookMetric
	InitPrometheusCounterActions(metric, "WEBHOOK_ID", "SERVICE_ID", "", "RESULT")

	// WHEN it's incremented with IncreasePrometheusCounterActions
	IncreasePrometheusCounterActions(metric, "WEBHOOK_ID", "SERVICE_ID", "", "RESULT")

	// THEN this increase can be collected
	got := testutil.ToFloat64(metric.WithLabelValues("WEBHOOK_ID", "SERVICE_ID", "RESULT"))
	want := float64(1)
	if got != want {
		t.Errorf("WebHookMetric was incremented. Got %g, want %g",
			got, want)
	}
}

func TestInitPrometheusCounterActionsWithNotify(t *testing.T) {
	// GIVEN a NotifyMetric
	metric := NotifyMetric

	// WHEN it's initialised with InitPrometheusCounterActions
	InitPrometheusCounterActions(metric, "NOTIFY_ID", "SERVICE_ID", "NOTIFY_TYPE", "RESULT")

	// THEN it can be collected
	got := testutil.CollectAndCount(metric)
	want := 1
	if got != want {
		t.Errorf("%d NotifyMetric's were initialised, expecting %d",
			got, want)
	}
}

func TestIncreasePrometheusCounterActionsWithNotify(t *testing.T) {
	// GIVEN a NotifyMetric that's been initialised
	metric := NotifyMetric
	InitPrometheusCounterActions(metric, "NOTIFY_ID", "SERVICE_ID", "NOTIFY_TYPE", "RESULT")

	// WHEN it's incremented with IncreasePrometheusCounterActions
	IncreasePrometheusCounterActions(metric, "NOTIFY_ID", "SERVICE_ID", "NOTIFY_TYPE", "RESULT")

	// THEN this increase can be collected
	got := testutil.ToFloat64(metric.WithLabelValues("NOTIFY_ID", "RESULT", "SERVICE_ID", "NOTIFY_TYPE"))
	want := float64(1)
	if got != want {
		t.Errorf("NotifyMetric was incremented. Got %g, want %g",
			got, want)
	}
}

func TestSetPrometheusGaugeWithIDDidInitialise(t *testing.T) {
	// GIVEN a QueryLiveness
	metric := QueryLiveness

	// WHEN it's initialised with SetPrometheusGaugeWithID
	SetPrometheusGaugeWithID(metric, "SERVICE_ID", 5)

	// THEN it can be collected
	got := testutil.CollectAndCount(metric)
	want := 1
	if got != want {
		t.Errorf("%d QueryLiveness's were initialised, expecting %d",
			got, want)
	}
}

func TestSetPrometheusGaugeWithIDDidChange(t *testing.T) {
	// GIVEN a QueryLiveness that's been initialised
	metric := QueryLiveness
	was := float64(0)
	SetPrometheusGaugeWithID(metric, "SERVICE_ID", was)

	// WHEN it's changed with SetPrometheusGaugeWithID
	now := float64(1)
	SetPrometheusGaugeWithID(metric, "SERVICE_ID", now)

	// THEN this change can be collected
	got := testutil.ToFloat64(metric.WithLabelValues("SERVICE_ID"))
	want := float64(1)
	if got != want {
		t.Errorf("QueryLiveness should've been changed to %g but got %g",
			want, got)
	}
}
