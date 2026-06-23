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

//go:build integration

package shoutrrr

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/release-argus/Argus/service/status"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/web/metric"
)

func TestShoutrrrs_Metrics(t *testing.T) {
	// GIVEN: Shoutrrrs.
	tests := []struct {
		name      string
		shoutrrrs *Shoutrrrs
	}{
		{
			name:      "nil",
			shoutrrrs: nil,
		},
		{
			name:      "empty",
			shoutrrrs: &Shoutrrrs{},
		},
		{
			name: "with one",
			shoutrrrs: &Shoutrrrs{
				"foo": &Shoutrrr{},
			},
		},
		{
			name: "multiple",
			shoutrrrs: &Shoutrrrs{
				"bish": &Shoutrrr{},
				"bash": &Shoutrrr{},
				"bosh": &Shoutrrr{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're testing metrics.

			if tc.shoutrrrs != nil {
				for name, s := range *tc.shoutrrrs {
					s.ID = name
					s.ServiceStatus = &status.Status{
						ServiceInfo: serviceinfo.ServiceInfo{
							ID: name + "-service",
						},
					}
					s.Main = &Defaults{}
					s.Type = "gotify"
				}
			}

			// WHEN: the Prometheus metrics are initialised with initMetrics.
			had := testutil.CollectAndCount(metric.NotifyResultTotal)
			tc.shoutrrrs.InitMetrics()

			// THEN: it can be counted.
			got := testutil.CollectAndCount(metric.NotifyResultTotal)
			want := had
			if tc.shoutrrrs != nil {
				want += 2 * len(*tc.shoutrrrs)
			}
			if got != want {
				t.Errorf(
					"%s\nShoutrrrs.InitMetrics() metrics count mismatch\ngot:  %d\nwant: %d",
					packageName, got, want,
				)
			}

			// AND: the metrics can be deleted.
			tc.shoutrrrs.DeleteMetrics()
			got = testutil.CollectAndCount(metric.NotifyResultTotal)
			if got != had {
				t.Errorf(
					"%s\nShoutrrrs.DeleteMetrics() metrics count mismatch\ngot:  %d\nwant: %d",
					packageName, got, had,
				)
			}
		})
	}
}

func TestShoutrrr_Metrics(t *testing.T) {
	// GIVEN: a Shoutrrr.
	tests := []string{
		"a service",
		"another service",
	}

	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're testing metrics.

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.ServiceStatus.ServiceInfo.ID = name

			// WHEN: the Prometheus metrics are initialised with initMetrics.
			had := testutil.CollectAndCount(metric.NotifyResultTotal)
			shoutrrr.initMetrics()

			// THEN: it can be collected.
			// counters:
			got := testutil.CollectAndCount(metric.NotifyResultTotal)
			want := 2
			if (got - had) != want {
				t.Errorf(
					"%s\nShoutrrr.initMetrics() metrics count mismatch\ngot:  %d\nwant: %d",
					packageName, got-had, want,
				)
			}

			// AND: it can be deleted.
			shoutrrr.deleteMetrics()
			got = testutil.CollectAndCount(metric.NotifyResultTotal)
			if got != had {
				t.Errorf(
					"%s\nShoutrrr.deleteMetrics() metrics count mismatch\ngot:  %d\nwant: %d",
					packageName, got, had,
				)
			}
		})
	}
}
