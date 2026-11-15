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

package manual

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/web/metric"
)

func TestLookup_InitMetrics(t *testing.T) {
	lookup := Lookup{}
	hadC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)

	lookup.InitMetrics(&lookup)

	test.AssertMetricCount(
		t, "Lookup.InitMetrics() DeployedVersionQueryResultTotal",
		metric.DeployedVersionQueryResultTotal, hadC,
	)
}

func TestLookup_DeleteMetrics(t *testing.T) {
	originalC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)
	lookup := testLookup(t, "")
	lookup.Lookup.InitMetrics(lookup)
	hadC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)
	if hadC == originalC {
		t.Fatalf(
			"base.Lookup.InitMetrics() DeployedVersionQueryResultTotal count mismatch\n"+
				"got:  %d\nwant: NOT %d",
			hadC, originalC,
		)
	}

	lookup.DeleteMetrics(lookup)

	test.AssertMetricCount(
		t, "Lookup.DeleteMetrics() DeployedVersionQueryResultTotal",
		metric.DeployedVersionQueryResultTotal, hadC,
	)

	lookup.Lookup.DeleteMetrics(lookup)
	test.AssertMetricCount(
		t, "base.Lookup.DeleteMetrics() DeployedVersionQueryResultTotal",
		metric.DeployedVersionQueryResultTotal, originalC,
	)
}

func TestLookup_QueryMetrics(t *testing.T) {
	lookup := Lookup{}
	hadC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)

	lookup.QueryMetrics(&lookup, nil)

	test.AssertMetricCount(
		t, "Lookup.QueryMetrics() DeployedVersionQueryResultTotal",
		metric.DeployedVersionQueryResultTotal, hadC,
	)
}
