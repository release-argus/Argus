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
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	cfg "github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/utils"
)

func reset() {
	config = cfg.Config{}
	newConfigFile := ""
	configFile = &newConfigFile
	newConfigCheckFlag := false
	configCheckFlag = &newConfigCheckFlag
	newTestCommandsFlag := ""
	testCommandsFlag = &newTestCommandsFlag
	newTestNotifyFlag := ""
	testNotifyFlag = &newTestNotifyFlag
	newTestServiceFlag := ""
	testServiceFlag = &newTestServiceFlag
}

func TestTheMainWithNoServices(t *testing.T) {
	// GIVEN an empty Config (no Services)
	reset()
	jLog = *utils.NewJLog("WARN", false)
	file := "../../test/no_services.yml"
	configFile = &file
	// Switch Fatal to panic and disable this panic.
	jLog.Testing = true
	defer func() {
		r := recover()
		if r == nil || !strings.Contains(r.(string), "No services to monitor") {
			t.Error(r)
		}
	}()

	// WHEN Main is called
	main()

	// THEN the program will exit
	t.Error("This shouldn't be reached since there are 0 Services to monitor")
}

func TestTheMainWithServices(t *testing.T) {
	// GIVEN an empty Config (no Services)
	reset()
	jLog = *utils.NewJLog("INFO", false)
	file := "../../test/argus.yml"
	configFile = &file
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	// Switch Fatal to panic and disable this panic.
	jLog.Testing = true

	// WHEN Main is called
	go main()
	time.Sleep(5 * time.Second)

	// THEN the program will be monitoring Argus for new releases
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	output := string(out)
	if !strings.Contains(output, "services to monitor:") {
		t.Errorf("Couldn't find the count of how many services it's monitoring\n%s",
			output)
	}
	if !strings.Contains(output, "release-argus/Argus, Latest Release - ") {
		t.Errorf("Couldn't find the output of the Latest Release for Argus. Is Track failing?\n%s",
			output)
	}
}
