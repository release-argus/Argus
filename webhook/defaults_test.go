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

package webhook

import (
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func TestDefaults_Default(t *testing.T) {
	defaults := Defaults{}
	expected := Defaults{
		Type:              "github",
		Delay:             "0s",
		AllowInvalidCerts: new(false),
		DesiredStatusCode: new(uint16(0)),
		MaxTries:          new(uint8(3)),
		SilentFails:       new(false),
	}

	defaults.Default()

	fieldTests := []test.FieldAssertion{
		{Name: "Type", Got: defaults.Type, Want: expected.Type, Mode: test.CompareEqual},
		{Name: "Delay", Got: defaults.Delay, Want: expected.Delay, Mode: test.CompareEqual},
		{Name: "AllowInvalidCerts", Got: test.StringifyPtr(defaults.AllowInvalidCerts), Want: test.StringifyPtr(expected.AllowInvalidCerts), Mode: test.CompareEqual},
		{Name: "DesiredStatusCode", Got: test.StringifyPtr(defaults.DesiredStatusCode), Want: test.StringifyPtr(expected.DesiredStatusCode), Mode: test.CompareEqual},
		{Name: "MaxTries", Got: test.StringifyPtr(defaults.MaxTries), Want: test.StringifyPtr(expected.MaxTries), Mode: test.CompareEqual},
		{Name: "SilentFails", Got: test.StringifyPtr(defaults.SilentFails), Want: test.StringifyPtr(expected.SilentFails), Mode: test.CompareEqual},
	}
	if testErr := test.AssertFields(t, fieldTests, "Defaults"); testErr != nil {
		t.Fatal(testErr)
	}
}
