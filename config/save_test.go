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

//go:build unit

package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/utils"
)

var TIMEOUT time.Duration = 25 * time.Second

func TestSaveHandler(t *testing.T) {
	// GIVEN a message is sent to the SaveHandler
	jLog = utils.NewJLog("WARN", false)
	config := testConfig()
	// Switch Fatal to panic and disable this panic.
	jLog.Testing = true
	defer func() { _ = recover() }()
	go func() {
		*config.SaveChannel <- true
	}()

	// WHEN the SaveHandler is running for a Config with an inaccessible file
	config.SaveHandler()

	// THEN it should have panic'd after TIMEOUT and not reach this
	time.Sleep(TIMEOUT * time.Second)
	t.Errorf("Save should panic'd on inaccessible file location %q",
		config.File)
}

func TestWaitChannelTimeout(t *testing.T) {
	// GIVEN a Config.SaveChannel
	config := testConfig()

	// WHEN the waitChannelTimeout is called
	time.Sleep(time.Second)
	start := time.Now().UTC()
	waitChannelTimeout(config.SaveChannel)

	// THEN after `TIMEOUT`, it would have tried to Save (and failed)
	elapsed := time.Since(start)
	if elapsed < TIMEOUT {
		t.Errorf("waitChannelTimeout should have waited atleast %v, but only waited %v",
			TIMEOUT, elapsed)
	}
}

func TestWaitChannelTimeoutDoesExtend(t *testing.T) {
	// GIVEN a Config.SaveChannel that is in the waitChannelTimeout
	config := testConfig()
	go func() {
		*config.SaveChannel <- true
	}()

	// WHEN another message is sent to the channel mid-way through the wait
	go func() {
		time.Sleep(10 * time.Second)
		*config.SaveChannel <- true
	}()
	time.Sleep(time.Second)
	start := time.Now().UTC()
	waitChannelTimeout(config.SaveChannel)

	// THEN after 2*`TIMEOUT`, it would have tried to Save (and failed)
	elapsed := time.Since(start)
	if elapsed < 2*TIMEOUT ||
		elapsed > 2*TIMEOUT+5*time.Second {
		t.Errorf("waitChannelTimeout should have waited ~%s, but waited %v",
			2*TIMEOUT, elapsed)
	}
}

func TestWaitChannelTimeoutDoesExtendOnce(t *testing.T) {
	// GIVEN a Config.SaveChannel that is in the waitChannelTimeout
	config := testConfig()
	go func() {
		*config.SaveChannel <- true
	}()

	// WHEN two messages are sent to the channel mid-way through the wait
	go func() {
		time.Sleep(10 * time.Second)
		*config.SaveChannel <- true
		time.Sleep(10 * time.Second)
		*config.SaveChannel <- true
	}()
	time.Sleep(time.Second)
	start := time.Now().UTC()
	waitChannelTimeout(config.SaveChannel)

	// THEN after 2*`TIMEOUT`, it would have tried to Save (and failed)
	elapsed := time.Since(start)
	if elapsed < 2*TIMEOUT ||
		elapsed > 2*TIMEOUT+5*time.Second {
		t.Errorf("waitChannelTimeout should have waited ~%s, but waited %v",
			2*TIMEOUT, elapsed)
	}
}

func TestSave(t *testing.T) {
	// GIVEN we have data that wants to be Save'd
	config := Config{File: "../test/config_test.yml"}
	originalData, err := os.ReadFile(config.File)
	had := string(originalData)
	if err != nil {
		t.Fatalf("Failed opening the file for the data we were going to Save\n%s",
			err.Error())
	}
	flags := make(map[string]bool)
	config.Load(config.File, &flags, &utils.JLog{})

	// WHEN we Save it to a new location
	config.File += ".test"
	config.Save()

	// THEN it's the same as the original file
	failed := false
	newData, err := os.ReadFile(config.File)
	had = strings.ReplaceAll(had, "semantic_versioning: n\n", "semantic_versioning: false\n")
	had = strings.ReplaceAll(had, "interval: 123\n", "interval: 123s\n")
	had = strings.ReplaceAll(had, "delay: 2\n", "delay: 2s\n")
	if string(newData) != had {
		failed = true
		t.Errorf("File is different after Save. Got \n%s\nexpecting:\n%s",
			string(newData), had)
	}
	err = os.Remove(config.File)
	if err != nil {
		t.Fatal(err)
	}
	if failed {
		t.Fatal()
	}
}
