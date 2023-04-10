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
	"sync"
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

var resetLock sync.Mutex

func reset() {
	resetLock.Lock()
	defer resetLock.Unlock()
	config = cfg.Config{}
	configFile = stringPtr("")
	configCheckFlag = boolPtr(false)
	testCommandsFlag = stringPtr("")
	testNotifyFlag = stringPtr("")
	testServiceFlag = stringPtr("")
}

func TestTheMain(t *testing.T) {
	// GIVEN different Config's to test
	jLog = *util.NewJLog("WARN", false)
	tests := map[string]struct {
		file           func(path string, t *testing.T)
		outputContains *[]string
		db             string
	}{
		"config with no services": {
			file: testYAML_NoServices,
			db:   "test-no_services.db",
			outputContains: stringListPtr([]string{
				"Found 0 services to monitor",
				"Listening on "})},
		"config with services": {
			file: testYAML_Argus,
			db:   "test-argus.db",
			outputContains: stringListPtr([]string{
				"services to monitor:",
				"release-argus/Argus, Latest Release - ",
				"Listening on "})},
		"config with services and some !active": {
			file: testYAML_Argus_SomeInactive,
			db:   "test-argus-some-inactive.db",
			outputContains: stringListPtr([]string{
				"Found 1 services to monitor:"})},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			file := fmt.Sprintf("%s.yml", name)
			tc.file(file, t)
			defer os.Remove(tc.db)
			reset()
			configFile = &file
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			go func() {
				accessToken := os.Getenv("GITHUB_TOKEN")
				if accessToken == "" {
					return
				}
				for {
					config.OrderMutex.RLock()
					done := false
					if len(config.Order) != 0 {
						config.Defaults.Service.AccessToken = &accessToken
						done = true
					}
					if done {
						return
					}
					config.OrderMutex.RUnlock()
					time.Sleep(100 * time.Millisecond)
				}
			}()

			// WHEN Main is called
			go func() {
				jLog.Testing = true
				main()
			}()
			time.Sleep(3 * time.Second)

			// THEN the program will have printed everything expected
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = stdout
			output := string(out)
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
