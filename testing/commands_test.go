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

package testing

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/utils"
)

func TestCommandTestWithNoService(t *testing.T) {
	// GIVEN a Config with a Service containing a Command
	jLog = utils.NewJLog("WARN", false)
	InitJLog(jLog)
	serviceID := "test"
	cfg := config.Config{
		Service: service.Slice{
			serviceID: &service.Service{
				ID: &serviceID,
				Command: &command.Slice{
					command.Command{"true", "0"},
				},
				CommandController: &command.Controller{},
			},
		},
	}
	flag := ""
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN CommandTest is called with an empty (undefined) flag
	CommandTest(&flag, &cfg, jLog)

	// THEN nothing will be run/printed
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	output := string(out)
	want := ""
	if want != output {
		t.Errorf("CommandTest with %q flag shouldn't print anything, got\n%s",
			flag, output)
	}
}

func TestCommandTestWithUnknownService(t *testing.T) {
	// GIVEN a Config with a Service containing a Command
	jLog = utils.NewJLog("WARN", false)
	InitJLog(jLog)
	cfg := config.Config{
		Service: service.Slice{
			"test": &service.Service{
				Command: &command.Slice{
					command.Command{"true", "0"},
				},
			},
		},
	}
	flag := "other_test"
	// Switch Fatal to panic and disable this panic.
	jLog.Testing = true
	defer func() {
		r := recover()
		if !strings.Contains(r.(string), " could not be found ") {
			t.Error(r)
		}
	}()

	// WHEN CommandTest is called with a Service not in the config
	CommandTest(&flag, &cfg, jLog)

	// THEN it will be printed that the command couldn't be found
	t.Error("Should os.Exit(1), err")
}

func TestCommandTestWithKnownService(t *testing.T) {
	// GIVEN a Config with a Service containing a Command
	jLog = utils.NewJLog("INFO", false)
	InitJLog(jLog)
	serviceID := "test"
	interval := "11m"
	cfg := config.Config{
		Service: service.Slice{
			serviceID: &service.Service{
				ID: &serviceID,
				Command: &command.Slice{
					command.Command{"true", "0"},
				},
				CommandController: &command.Controller{},
				Interval:          &interval,
			},
		},
	}
	cfg.Service[serviceID].CommandController.Init(jLog, &serviceID, nil, cfg.Service[serviceID].Command, nil, cfg.Service[serviceID].Interval)
	flag := serviceID
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	// Switch Fatal to panic and disable this panic.
	jLog.Testing = true

	// WHEN CommandTest is called with a Service not in the config
	CommandTest(&flag, &cfg, jLog)

	// THEN it Command will be executed and output
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	output := string(out)
	if !strings.Contains(output, "Executing ") {
		t.Errorf("Expected Command to have been executed, got\n%s",
			output)
	}
}
