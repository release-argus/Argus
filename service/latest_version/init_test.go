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

package latestver

import (
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
)

func TestLookup_Init(t *testing.T) {
	lvCfg := plainDefaultsConfig()

	lookup := testLookup(t, "github", false)
	svcStatus := status.Status{}
	svcStatus.ServiceInfo.ID = "test"
	var options opt.Options

	lookup.Init(
		&options,
		&svcStatus,
		base.DefaultsConfig{
			Soft: &lvCfg.Soft.Common,
			Hard: &lvCfg.Hard.Common,
		},
	)

	fieldTests := []test.FieldAssertion{
		{Name: "GetOptions()", Got: lookup.GetOptions(), Want: &options, Mode: test.CompareSamePointer},
		{Name: "GetStatus()", Got: lookup.GetStatus(), Want: &svcStatus, Mode: test.CompareSamePointer},
		{Name: "GetDefaults()", Got: lookup.GetDefaults(), Want: &lvCfg.Soft.Common, Mode: test.CompareSamePointer},
		{Name: "GetHardDefaults()", Got: lookup.GetHardDefaults(), Want: &lvCfg.Hard.Common, Mode: test.CompareSamePointer},
	}
	if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
		t.Fatal(testErr)
	}
}
