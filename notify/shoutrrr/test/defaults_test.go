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

package test

import (
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func TestPlainConfig(t *testing.T) {
	notifyCfg := PlainConfig()

	fieldTests := []test.FieldAssertion{
		{Name: "Root", Got: notifyCfg.Root, Mode: test.CompareNotNil},
		{Name: "Defaults", Got: notifyCfg.Defaults, Mode: test.CompareNotNil},
		{Name: "HardDefaults", Got: notifyCfg.HardDefaults, Mode: test.CompareNotNil},
	}
	if testErr := test.AssertFields(t, fieldTests, "Config"); testErr != nil {
		t.Error(testErr)
	}
}
