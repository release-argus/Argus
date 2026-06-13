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
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/web/metric"
)

func TestUpdateUpdatesCurrentMetric(t *testing.T) {
	// GIVEN: a Status that has just had a version change.
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
			name: "0 to 1 - Latest version not deployed/approved/skipped -> Latest version deployed",
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
			name: "0 to 2 - Latest version not deployed/approved/skipped -> Latest version approved",
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
			name: "0 to 3 - Latest version not deployed/approved/skipped -> Latest version skipped",
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
			name: "1 to 0 - Latest version deployed -> Latest version not deployed/approved/skipped",
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
		// Cannot go from deployed to approved/skipped without first being available.
		// "1 to 2 - Latest version deployed -> Latest version approved": {}.
		// "1 to 3 - Latest version deployed -> Latest version skipped": {}.
		{
			name: "2 to 0 - Latest version approved -> Latest version not deployed/approved/skipped",
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
			name: "3 to 0 - Latest version skipped -> Latest version not deployed/approved/skipped",
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

	// Changing and reading UpdatesCurrent.
	metricsMu.Lock()
	t.Cleanup(metricsMu.Unlock)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're testing metrics.

			tc.newVersions.DeployedVersion = util.ValueOr(tc.newVersions.DeployedVersion, tc.previousVersions.DeployedVersion)
			tc.newVersions.LatestVersion = util.ValueOr(tc.newVersions.LatestVersion, tc.previousVersions.LatestVersion)
			hadUpdatesCurrentAvailable := testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("AVAILABLE"))
			hadUpdatesCurrentSkipped := testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("SKIPPED"))

			// WHEN: updateUpdatesCurrentMetric is called with the previous and new versions.
			updateUpdatesCurrentMetric(tc.previousVersions, tc.newVersions)

			prefix := fmt.Sprintf(
				"%s\nUpdateUpdatesCurrentMetric(previous=%+v, new=%+v)",
				packageName, tc.previousVersions, tc.newVersions,
			)

			// Validate the update counts for both the approved and skipped metrics.
			// For this, we assume that `SetUpdatesCurrent` has been correctly implemented,
			// and the metrics have been updated accordingly.
			want := hadUpdatesCurrentAvailable + tc.updateCountAvailable
			got := testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("AVAILABLE"))
			if got != want {
				t.Errorf(
					"%s 'updates available' metric mismatch\ngot:  %f\nwant: %f",
					prefix, got, want,
				)
			}
			want = hadUpdatesCurrentSkipped + tc.updateCountSkipped
			got = testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("SKIPPED"))
			if got != want {
				t.Errorf(
					"%s 'updates skipped' metric mismatch\ngot:  %f\nwant: %f",
					prefix, got, want,
				)
			}
		})
	}
}
