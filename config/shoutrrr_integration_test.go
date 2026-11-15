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

package config

import (
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
		t.Skipf("No %q env vars found for Shoutrrr integration tests", targetEnvPrefix)
	}

	test.IsolateEnv(t, testVars)

	var defaults, hardDefaults Defaults
	hardDefaults.Default()
	if !hardDefaults.MapEnvToStruct() {
		t.Fatal("Defaults.MapEnvToStruct() failed")
	}
	defaults.SetDefaults(&hardDefaults)

	for _, typ := range shoutrrr.SupportedTypes {
		t.Run(typ, func(t *testing.T) {
			typePrefix := "NOTIFY_" + strings.ToUpper(typ) + "_"
			configured := false
			for k := range testVars {
				if strings.HasPrefix(k, newEnvPrefix+typePrefix) {
					configured = true
					break
				}
			}
			if !configured {
				t.Skipf("No %q env vars found", targetEnvPrefix+typePrefix)
			}

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

			if err, _ := testShoutrrr.CheckValues(); err != nil {
				t.Fatalf("Shoutrrr.CheckValues() failed: %v", err)
			}

			err := testShoutrrr.Send(
				"TestShoutrrr_Send__EnvTests",
				"message here",
				svcStatus.ServiceInfo,
				false,
				false,
			)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
