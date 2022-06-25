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

package command

import (
	"encoding/json"
	"fmt"
	"testing"

	api_types "github.com/release-argus/Argus/web/api/types"
)

func TestAnnounceCommandNoChannel(t *testing.T) {
	// GIVEN a Controller with no Announce channel
	controllerName := "TEST"
	commandController := Controller{
		ServiceID: &controllerName,
		Command: &Slice{
			Command{"ls", "-lah"},
		},
		Failed: make(Fails, 1),
	}

	// WHEN AnnounceCommand is run
	commandController.AnnounceCommand(0)

	// THEN the function doesn't hang
}

func testAnnounceCommand() Controller {
	controllerName := "TEST"
	commandController := Controller{
		ServiceID: &controllerName,
		Command: &Slice{
			Command{"ls", "-lah"},
		},
		Failed: make(Fails, 1),
	}
	channel := make(chan []byte)
	commandController.Announce = &channel
	return commandController
}

func TestAnnounceCommandWhenNotRun(t *testing.T) {
	// GIVEN a Controller with an Announce channel
	commandController := testAnnounceCommand()

	// WHEN AnnounceCommand is run with the command not having been run
	commandController.Failed[0] = nil
	go commandController.AnnounceCommand(0)
	result := <-*commandController.Announce
	var parsedResult api_types.WebSocketMessage
	//#nosec G104 -- Disregard
	//nolint:errcheck // ^
	json.Unmarshal(result, &parsedResult)

	// THEN it broadcasts nil to the Announce channel
	if parsedResult.CommandData["ls -lah"].Failed != nil {
		t.Errorf("got %t. expected %s", *parsedResult.CommandData["ls -lah"].Failed, "false")
	}
}

func TestAnnounceCommandWhenRunPassed(t *testing.T) {
	// GIVEN a Controller with an Announce channel
	commandController := testAnnounceCommand()

	// WHEN AnnounceCommand is run with the command not failing
	failed := false
	commandController.Failed[0] = &failed
	go commandController.AnnounceCommand(0)
	result := <-*commandController.Announce
	var parsedResult api_types.WebSocketMessage
	//#nosec G104 -- Disregard
	//nolint:errcheck // ^
	json.Unmarshal(result, &parsedResult)

	// THEN it broadcasts false to the Announce channel
	if *parsedResult.CommandData["ls -lah"].Failed != false {
		got := "nil"
		if parsedResult.CommandData["ls -lah"].Failed != nil {
			got = fmt.Sprint(*parsedResult.CommandData["ls -lah"].Failed)
		}
		t.Errorf("got %s. expected %s", got, "false")
	}
}

func TestAnnounceCommandWhenRunFailed(t *testing.T) {
	// GIVEN a Controller with an Announce channel
	commandController := testAnnounceCommand()

	// WHEN AnnounceCommand is run with the command failing
	failed := true
	commandController.Failed[0] = &failed
	go commandController.AnnounceCommand(0)
	result := <-*commandController.Announce
	var parsedResult api_types.WebSocketMessage
	//#nosec G104 -- Disregard
	//nolint:errcheck // ^
	json.Unmarshal(result, &parsedResult)

	// THEN it broadcasts true to the Announce channel
	if *parsedResult.CommandData["ls -lah"].Failed != true {
		got := "nil"
		if parsedResult.CommandData["ls -lah"].Failed != nil {
			got = fmt.Sprint(parsedResult.CommandData["ls -lah"].Failed)
		}
		t.Errorf("got %s. expected %s", got, "true")
	}
}

func testFindController() Controller {
	controllerName := "TEST"
	commandController := Controller{
		ServiceID: &controllerName,
		Command: &Slice{
			Command{"ls", "-lah"},
			Command{"ls", "-lah", "a"},
			Command{"ls", "-lah", "b"},
		},
		Failed: make(Fails, 3),
	}
	return commandController
}

func TestFindUnknown(t *testing.T) {
	// GIVEN we have a Slice of Commands
	commandController := testFindController()
	var index *int

	// WHEN Find is run for an unknown command
	function := "random function"
	index = commandController.Find(function)

	// THEN nil is returned
	if index != nil {
		t.Errorf("Command %q was found at index %d instead of nil", function, *index)
	}
}

func TestFindKnown(t *testing.T) {
	// GIVEN we have a Slice of Commands
	commandController := testFindController()
	var index *int

	// WHEN Find is run with a known command
	function := "ls -lah a"
	index = commandController.Find(function)

	// THEN it returns the correct index of that command
	if index == nil || *index != 1 {
		got := "nil"
		if index != nil {
			got = fmt.Sprint(index)
		}
		t.Errorf("Command %q was found at index %s instead of 1", function, got)
	}
}

func TestResetFailsNilController(t *testing.T) {
	// GIVEN a nil Controller
	var commandController Controller

	// WHEN ResetFails is run
	commandController.ResetFails()

	// THEN the command doesn't hang
}

func TestResetFailsNonNilController(t *testing.T) {
	// GIVEN a Controller with Commands that have failed
	controllerName := "TEST"
	commandController := Controller{
		ServiceID: &controllerName,
		Command: &Slice{
			Command{"ls", "-lah"},
			Command{"ls", "-lah", "a"},
			Command{"ls", "-lah", "b"},
			Command{"ls", "-lah", "c"},
			Command{"ls", "-lah", "d"},
		},
		Failed: make(Fails, 5),
	}
	failed0 := true
	commandController.Failed[0] = &failed0
	failed1 := true
	commandController.Failed[1] = &failed1
	failed2 := false
	commandController.Failed[2] = &failed2
	failed3 := true
	commandController.Failed[3] = &failed3
	failed4 := true
	commandController.Failed[4] = &failed4

	// WHEN ResetFails is called
	commandController.ResetFails()
	// THEN all the fails become nil
	for _, failed := range commandController.Failed {
		if failed != nil {
			t.Errorf("Reset failed, got %v", commandController.Failed)
		}

	}
}
