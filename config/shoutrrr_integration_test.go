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

package config

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	statustest "github.com/release-argus/Argus/service/status/test"
)

func TestShoutrrr_Send__EnvTests(t *testing.T) {
	targetEnvPrefix := "ARGUS_SHOUTRRR_INTEGRATION_TEST_"
	newEnvPrefix := "ARGUS_"
	envVars := os.Environ()
	testVars := make(map[string]string, len(shoutrrr.SupportedTypes))
	for _, envVar := range envVars {
		if val, ok := strings.CutPrefix(envVar, targetEnvPrefix); ok {
			val = newEnvPrefix + val
			envKey, envVal, _ := strings.Cut(val, "=")
			testVars[envKey] = envVal
		}
	}
	if len(testVars) == 0 {
		t.Skipf(
			"%s\nNo %q env vars found for Shoutrrr integration tests",
			packageName, targetEnvPrefix,
		)
	}

	test.SetEnv(t, testVars)

	// GIVEN: defaults and regular hardDefaults.
	var defaults, hardDefaults Defaults
	hardDefaults.Default()
	_ = hardDefaults.MapEnvToStruct()
	defaults.SetDefaults(&hardDefaults)

	// GIVEN: a Shoutrrr of each type.
	for _, typ := range shoutrrr.SupportedTypes {
		t.Run(typ, func(t *testing.T) {

			testShoutrrr := shoutrrr.New(
				nil,
				typ,
				typ,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
			)
			svcStatus, _ := statustest.New("yaml", nil)
			testShoutrrr.Init(
				svcStatus,
				nil,
				&shoutrrr.Defaults{},
				hardDefaults.Notify[typ],
			)

			prefix := fmt.Sprintf(
				"%s\nShoutrrr.Send(type=%s)",
				packageName, typ,
			)

			if err, _ := testShoutrrr.CheckValues(); err != nil {
				t.Skipf(
					"%s CheckValues() failed:\n%v",
					prefix, err,
				)
			}

			message := fmt.Sprintf("%s message here", prefix)

			// WHEN: send attempted.
			if err := testShoutrrr.Send(
				"TestShoutrrr_Send__EnvTests",
				message,
				svcStatus.ServiceInfo,
				false,
				false,
			); err != nil {
				t.Errorf(
					"%s failed unexpectedly\n%v",
					prefix, err,
				)
			}
		})
	}
}
