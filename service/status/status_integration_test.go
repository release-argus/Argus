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

package status

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/web/metric"
)

func TestUpdateUpdatesCurrentMetric(t *testing.T) {
	tests := []struct {
		name                                     string
		previousVersions, newVersions            serviceinfo.ServiceInfo
		updateCountAvailable, updateCountSkipped float64
	}{
		{
			name: "0 to 0 - versions unchanged",
			previousVersions: serviceinfo.ServiceInfo{
				ApprovedVersion: "",
				DeployedVersion: "1.1.0",
				LatestVersion:   "1.2.0",
			},
			newVersions: serviceinfo.ServiceInfo{
				ApprovedVersion: "",
				DeployedVersion: "1.1.0",
				LatestVersion:   "1.2.0",
			},
			updateCountAvailable: 0,
			updateCountSkipped:   0,
		},
		{
			name: "0 to 1 - Latest version not deployed, approved, or skipped -> Latest version deployed",
			previousVersions: serviceinfo.ServiceInfo{
				ApprovedVersion: "",
				DeployedVersion: "1.1.0",
				LatestVersion:   "1.2.0",
			},
			newVersions: serviceinfo.ServiceInfo{
				DeployedVersion: "1.2.0",
			},
			updateCountAvailable: -1,
			updateCountSkipped:   0,
		},
		{
			name: "0 to 2 - Latest version not deployed, approved, or skipped -> Latest version approved",
			previousVersions: serviceinfo.ServiceInfo{
				ApprovedVersion: "",
				DeployedVersion: "1.1.0",
				LatestVersion:   "1.2.0",
			},
			newVersions: serviceinfo.ServiceInfo{
				ApprovedVersion: "1.2.0",
			},
			updateCountAvailable: 0,
			updateCountSkipped:   0,
		},
		{
			name: "0 to 3 - Latest version not deployed, approved, or skipped -> Latest version skipped",
			previousVersions: serviceinfo.ServiceInfo{
				ApprovedVersion: "",
				DeployedVersion: "1.1.0",
				LatestVersion:   "1.2.0",
			},
			newVersions: serviceinfo.ServiceInfo{
				ApprovedVersion: serviceinfo.SkippedVersion("1.2.0"),
			},
			updateCountAvailable: 0,
			updateCountSkipped:   1,
		},
		{
			name: "1 to 0 - Latest version deployed -> Latest version not deployed, approved, or skipped",
			previousVersions: serviceinfo.ServiceInfo{
				ApprovedVersion: "",
				DeployedVersion: "1.2.0",
				LatestVersion:   "1.2.0",
			},
			newVersions: serviceinfo.ServiceInfo{
				LatestVersion: "1.3.0",
			},
			updateCountAvailable: 1,
			updateCountSkipped:   0,
		},
		// Deployed cannot go to approved/skipped (1 to 2, 1 to 3) without first being available.
		{
			name: "2 to 0 - Latest version approved -> Latest version not deployed, approved, or skipped",
			previousVersions: serviceinfo.ServiceInfo{
				ApprovedVersion: "1.2.0",
				DeployedVersion: "1.1.0",
				LatestVersion:   "1.2.0",
			},
			newVersions: serviceinfo.ServiceInfo{
				ApprovedVersion: "",
			},
			updateCountAvailable: 0,
			updateCountSkipped:   0,
		},
		{
			name: "2 to 1 - Latest version approved -> Latest version deployed",
			previousVersions: serviceinfo.ServiceInfo{
				ApprovedVersion: "1.2.0",
				DeployedVersion: "1.1.0",
				LatestVersion:   "1.2.0",
			},
			newVersions: serviceinfo.ServiceInfo{
				DeployedVersion: "1.2.0",
			},
			updateCountAvailable: -1,
			updateCountSkipped:   0,
		},
		{
			name: "2 to 3 - Latest version approved -> Latest version skipped",
			previousVersions: serviceinfo.ServiceInfo{
				ApprovedVersion: "1.2.0",
				DeployedVersion: "1.1.0",
				LatestVersion:   "1.2.0",
			},
			newVersions: serviceinfo.ServiceInfo{
				ApprovedVersion: serviceinfo.SkippedVersion("1.2.0"),
			},
			updateCountAvailable: 0,
			updateCountSkipped:   1,
		},
		{
			name: "3 to 0 - Latest version skipped -> Latest version not deployed, approved, or skipped",
			previousVersions: serviceinfo.ServiceInfo{
				ApprovedVersion: serviceinfo.SkippedVersion("1.2.0"),
				DeployedVersion: "1.1.0",
				LatestVersion:   "1.2.0",
			},
			newVersions: serviceinfo.ServiceInfo{
				ApprovedVersion: "",
			},
			updateCountAvailable: 0,
			updateCountSkipped:   -1,
		},
		{
			name: "3 to 1 - Latest version skipped -> Latest version deployed",
			previousVersions: serviceinfo.ServiceInfo{
				ApprovedVersion: serviceinfo.SkippedVersion("1.2.0"),
				DeployedVersion: "1.1.0",
				LatestVersion:   "1.2.0",
			},
			newVersions: serviceinfo.ServiceInfo{
				DeployedVersion: "1.2.0",
			},
			updateCountAvailable: -1,
			updateCountSkipped:   -1,
		},
		{
			name: "3 to 2 - Latest version skipped -> Latest version approved",
			previousVersions: serviceinfo.ServiceInfo{
				ApprovedVersion: serviceinfo.SkippedVersion("1.2.0"),
				DeployedVersion: "1.1.0",
				LatestVersion:   "1.2.0",
			},
			newVersions: serviceinfo.ServiceInfo{
				ApprovedVersion: "1.2.0",
			},
			updateCountAvailable: 0,
			updateCountSkipped:   -1,
		},
	}

	// The subtests change and read metric.UpdatesCurrent.
	metricsMu.Lock()
	t.Cleanup(metricsMu.Unlock)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: shares metric.UpdatesCurrent.

			tc.newVersions.DeployedVersion = util.ValueOr(tc.newVersions.DeployedVersion, tc.previousVersions.DeployedVersion)
			tc.newVersions.LatestVersion = util.ValueOr(tc.newVersions.LatestVersion, tc.previousVersions.LatestVersion)
			hadUpdatesCurrentAvailable := testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("AVAILABLE"))
			hadUpdatesCurrentSkipped := testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("SKIPPED"))

			updateUpdatesCurrentMetric(tc.previousVersions, tc.newVersions)

			want := hadUpdatesCurrentAvailable + tc.updateCountAvailable
			got := testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("AVAILABLE"))
			if got != want {
				t.Errorf(
					"metric.UpdatesCurrent.WithLabelValues(\"AVAILABLE\") mismatch\n"+
						"got:  %f\nwant: %f",
					got, want,
				)
			}
			want = hadUpdatesCurrentSkipped + tc.updateCountSkipped
			got = testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("SKIPPED"))
			if got != want {
				t.Errorf(
					"metric.UpdatesCurrent.WithLabelValues(\"SKIPPED\") mismatch\n"+
						"got:  %f\nwant: %f",
					got, want,
				)
			}
		})
	}
}
