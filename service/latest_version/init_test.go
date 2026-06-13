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
	"fmt"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
)

func TestLookup_Init(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)

	// GIVEN: a Lookup and vars for the Init.
	lookup := testLookup(t, "github", false)
	lookup.GetStatus().ServiceInfo.ID += "TestInit"
	svcStatus := status.Status{}
	svcStatus.ServiceInfo.ID = "test"
	var options opt.Options

	// WHEN: Init is called on it.
	lookup.Init(
		&options,
		&svcStatus,
		lvCfg,
	)

	prefix := fmt.Sprintf("%s\nLookup.Init()", packageName)

	// THEN: pointers to those vars are handed out to the Lookup.
	fieldTests := []test.FieldAssertion{
		{Name: "Options", Got: lookup.GetOptions(), Want: &options, Mode: test.CompareSamePointer},
		{Name: "Status", Got: lookup.GetStatus(), Want: &svcStatus, Mode: test.CompareSamePointer},
		{Name: "Defaults", Got: lookup.GetDefaults(), Want: lvCfg.Soft, Mode: test.CompareSamePointer},
		{Name: "HardDefaults", Got: lookup.GetHardDefaults(), Want: lvCfg.Hard, Mode: test.CompareSamePointer},
	}
	if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
		t.Fatal(err)
	}
}
