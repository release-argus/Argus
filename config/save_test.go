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

	"github.com/release-argus/Argus/util"
)

var TIMEOUT time.Duration = 30 * time.Second

func TestConfig_SaveHandler(t *testing.T) {
	// GIVEN a message is sent to the SaveHandler
	jLog = util.NewJLog("WARN", false)
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
	// GIVEN a Config.SaveChannel and messages to send/not send
	tests := map[string]struct {
		messages  int
		timeTaken time.Duration
	}{
		"no messages": {
			messages:  0,
			timeTaken: TIMEOUT,
		},
		"one message": {
			messages:  1,
			timeTaken: 2 * TIMEOUT,
		},
		"two messages": {
			messages:  2,
			timeTaken: 2 * TIMEOUT,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			config := testConfig()

			// WHEN those messages are sent to the channel mid-way through the wait
			go func() {
				for tc.messages != 0 {
					time.Sleep(10 * time.Second)
					*config.SaveChannel <- true
					tc.messages--
				}
			}()
			time.Sleep(time.Second)
			start := time.Now().UTC()
			waitChannelTimeout(config.SaveChannel)

			// THEN after `TIMEOUT`, it would have tried to Save
			elapsed := time.Since(start)
			if elapsed < tc.timeTaken-100*time.Millisecond ||
				elapsed > tc.timeTaken+100*time.Millisecond {
				t.Errorf("waitChannelTimeout should have waited atleast %s, but only waited %s",
					tc.timeTaken, elapsed)
			}
		})
	}
}

func TestConfig_Save(t *testing.T) {
	// GIVEN we have a bunch of files that want to be Save'd
	tests := map[string]struct {
		file        func(path string)
		corrections map[string]string
	}{
		"config_test.yml": {file: testYAML_ConfigTest, corrections: map[string]string{
			"listen_port: 0\n":         "listen_port: \"0\"\n",
			"semantic_versioning: n\n": "semantic_versioning: false\n",
			"interval: 123\n":          "interval: 123s\n",
			"delay: 2\n":               "delay: 2s\n",
		}},
		"argus.yml": {file: testYAML_Argus, corrections: map[string]string{
			"listen_port: 0\n": "listen_port: \"0\"\n",
		}},
	}

	for name, tc := range tests {
		name, tc := name, tc
		file := name
		tc.file(file)
		defer os.Remove(file)
		t.Log(file)
		config := Config{File: file}
		originalData, err := os.ReadFile(config.File)
		had := string(originalData)
		if err != nil {
			t.Fatalf("Failed opening the file for the data we were going to Save\n%s",
				err.Error())
		}
		flags := make(map[string]bool)
		config.Load(config.File, &flags, &util.JLog{}) // Global vars could otherwise DATA RACE
		defer os.Remove(*config.Settings.GetDataDatabaseFile())

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN we Save it to a new location
			config.File += ".test"
			config.Save()

			// THEN it's the same as the original file
			failed := false
			newData, err := os.ReadFile(config.File)
			for from := range tc.corrections {
				had = strings.ReplaceAll(had, from, tc.corrections[from])
			}
			if string(newData) != had {
				failed = true
				t.Errorf("%q is different after Save. Got \n%s\nexpecting:\n%s",
					file, string(newData), had)
			}
			err = os.Remove(config.File)
			if err != nil {
				t.Fatal(err)
			}
			if failed {
				t.Fatal()
			}
			time.Sleep(time.Second)
		})
	}
}
