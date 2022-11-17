// Copyright [2022] [Argus]
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

package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	cfg "github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/util"
)

func boolPtr(val bool) *bool {
	return &val
}
func stringPtr(val string) *string {
	return &val
}
func stringListPtr(val []string) *[]string {
	return &val
}

func reset() {
	config = cfg.Config{}
	configFile = stringPtr("")
	configCheckFlag = boolPtr(false)
	testCommandsFlag = stringPtr("")
	testNotifyFlag = stringPtr("")
	testServiceFlag = stringPtr("")
}

func TestTheMain(t *testing.T) {
	// GIVEN different Config's to test
	tests := map[string]struct {
		file               string
		panicShouldContain *string
		outputContains     *[]string
		db                 string
	}{
		"config with no services": {file: "../../test/no_services.yml", db: "test-no_services.db", panicShouldContain: stringPtr("No services to monitor")},
		"config with services": {file: "../../test/argus.yml", db: "test-argus.db", outputContains: stringListPtr([]string{
			"services to monitor:",
			"release-argus/Argus, Latest Release - ",
		})},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			reset()
			jLog = *util.NewJLog("WARN", false)
			configFile = &tc.file
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN Main is called
			go func() {
				// Switch Fatal to panic and disable this panic.
				jLog.Testing = true
				defer func() {
					r := recover()
					rStr := fmt.Sprint(r)
					if tc.panicShouldContain != nil {
						if !strings.Contains(rStr, *tc.panicShouldContain) {
							t.Errorf("should have panic'd with:\n%q, not:\n%q",
								*tc.panicShouldContain, r)
						}
					} else if r != nil {
						t.Errorf("wasn't expecting a panic - %q",
							rStr)
					}
				}()
				main()
			}()
			time.Sleep(4 * time.Second)

			// THEN the program will have printed everything expected
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = stdout
			output := string(out)
			os.Remove(tc.db)
			if tc.outputContains != nil {
				for _, text := range *tc.outputContains {
					if !strings.Contains(output, text) {
						t.Errorf("%q couldn't be found in the output:\n%s",
							text, output)
					}
				}
			}
		})
	}
}
