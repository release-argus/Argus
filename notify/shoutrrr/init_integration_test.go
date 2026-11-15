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

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/status"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/web/metric"
)

func TestShoutrrrs_Metrics(t *testing.T) {
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
			// Serial: shares metric.NotifyResultTotal.

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

			had := testutil.CollectAndCount(metric.NotifyResultTotal)
			tc.shoutrrrs.InitMetrics()

			want := had
			if tc.shoutrrrs != nil {
				want += 2 * len(*tc.shoutrrrs)
			}
			test.AssertMetricCount(
				t, "Shoutrrrs.InitMetrics() NotifyResultTotal",
				metric.NotifyResultTotal, want,
			)

			tc.shoutrrrs.DeleteMetrics()
			test.AssertMetricCount(
				t, "Shoutrrrs.DeleteMetrics() NotifyResultTotal",
				metric.NotifyResultTotal, had,
			)
		})
	}
}

func TestShoutrrr_Metrics(t *testing.T) {
	tests := []string{
		"a service",
		"another service",
	}

	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			// Serial: shares metric.NotifyResultTotal.

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.ServiceStatus.ServiceInfo.ID = name

			had := testutil.CollectAndCount(metric.NotifyResultTotal)
			shoutrrr.initMetrics()

			test.AssertMetricCount(
				t, "Shoutrrr.initMetrics() NotifyResultTotal",
				metric.NotifyResultTotal, had+2,
			)

			shoutrrr.deleteMetrics()
			test.AssertMetricCount(
				t, "Shoutrrr.deleteMetrics() NotifyResultTotal",
				metric.NotifyResultTotal, had,
			)
		})
	}
}
