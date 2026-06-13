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
	"github.com/release-argus/Argus/web/metric"
)

func TestLookup_InitMetrics(t *testing.T) {
	// GIVEN: a Manual Lookup.
	lookup := Lookup{}
	hadC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)

	// WHEN: InitMetrics is called.
	lookup.InitMetrics(&lookup)

	// THEN: no metrics should be initialised as the function does nothing.
	gotC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)
	if gotC != hadC {
		t.Errorf(
			"%s\nLookup.InitMetrics() Counter metrics initialised unexpectedly\ngot:  %d\nwant: %d",
			packageName, gotC, hadC,
		)
	}
}

func TestLookup_DeleteMetrics(t *testing.T) {
	// GIVEN: a Manual Lookup with metrics.
	originalC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)
	lookup := testLookup(t, "")
	lookup.Lookup.InitMetrics(lookup)
	hadC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)
	if hadC == 0 {
		t.Fatalf(
			"%s\nLookup.InitMetrics() Counter metrics were not initialised (before DeleteMetrics())",
			packageName,
		)
	}

	// WHEN: DeleteMetrics is called.
	lookup.DeleteMetrics(lookup)

	// THEN: no metrics should be deleted as the function does nothing.
	if gotC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal); gotC != hadC {
		t.Errorf(
			"%s\nLookup.DeleteMetrics() Lookup Counter metrics deleted unexpectedly\ngot:  %d\nwant: %d",
			packageName, hadC, gotC,
		)
	}

	// AND: they can be deleted with DeleteMetrics on the base Lookup.
	lookup.Lookup.DeleteMetrics(lookup)
	if gotC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal); gotC != originalC {
		t.Errorf(
			"%s\nBase.QueryMetrics() Lookup Counter metrics not deleted\ngot:  %d\nwant: %d",
			packageName, originalC, gotC,
		)
	}
}

func TestLookup_QueryMetrics(t *testing.T) {
	// GIVEN: a Manual Lookup.
	lookup := Lookup{}
	hadC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)

	// WHEN: QueryMetrics is called.
	lookup.QueryMetrics(&lookup, nil)

	// THEN: no metrics should be updated as the function does nothing.
	if gotC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal); gotC != hadC {
		t.Errorf(
			"%s\nLookup.QueryMetrics() unexpected Lookup Counter metric updates\ngot:  %d\nwant: %d",
			packageName, hadC, gotC,
		)
	}
}
