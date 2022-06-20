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

package command

import (
	"encoding/json"
	"fmt"
	"testing"

	api_types "github.com/release-argus/Argus/web/api/types"
)

func TestAnnounceCommand(t *testing.T) {
	{ // GIVEN a Controller with no Announce channel
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

	{ // GIVEN a Controller with an Announce channel
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
		var (
			parsedResult api_types.WebSocketMessage
			result       []byte
			failed       bool
		)

		// WHEN AnnounceCommand is run with the command not having been run
		go commandController.AnnounceCommand(0)
		result = <-channel
		err := json.Unmarshal(result, &parsedResult)
		if err != nil {
			t.Fatalf("Failed unmarshaling the JSON - %s", err.Error())
		}
		// THEN it broadcasts nil to the Announce channel
		if parsedResult.CommandData["ls -lah"].Failed != nil {
			t.Fatalf("got %t. expected %s", *parsedResult.CommandData["ls -lah"].Failed, "false")
		}

		// WHEN AnnounceCommand is run with the command not failing
		failed = false
		commandController.Failed[0] = &failed
		go commandController.AnnounceCommand(0)
		result = <-channel
		//#nosec G104 -- Disregard
		//nolint:errcheck // ^
		json.Unmarshal(result, &parsedResult)
		// THEN it broadcasts false to the Announce channel
		if *parsedResult.CommandData["ls -lah"].Failed != false {
			got := "nil"
			if parsedResult.CommandData["ls -lah"].Failed != nil {
				got = fmt.Sprint(*parsedResult.CommandData["ls -lah"].Failed)
			}
			t.Fatalf("got %s. expected %s", got, "false")
		}

		// WHEN AnnounceCommand is run with the command failing
		failed = true
		commandController.Failed[0] = &failed
		go commandController.AnnounceCommand(0)
		result = <-channel
		//#nosec G104 -- Disregard
		//nolint:errcheck // ^
		json.Unmarshal(result, &parsedResult)
		// THEN it broadcasts true to the Announce channel
		if *parsedResult.CommandData["ls -lah"].Failed != true {
			got := "nil"
			if parsedResult.CommandData["ls -lah"].Failed != nil {
				got = fmt.Sprint(parsedResult.CommandData["ls -lah"].Failed)
			}
			t.Fatalf("got %s. expected %s", got, "true")
		}
	}
}

func TestFind(t *testing.T) {
	{ // GIVEN we have a Slice of Commands
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
		var index *int

		// WHEN Find is run for an unknown command
		function := "random function"
		index = commandController.Find(function)
		// THEN nil is returned
		if index != nil {
			t.Fatalf("Command %q was found at index %d instead of nil", function, *index)
		}

		// WHEN Find is run with a known command
		function = "ls -lah a"
		index = commandController.Find(function)
		// THEN it returns the correct index of that command
		if index != nil && *index != 1 {
			got := "nil"
			if index != nil {
				got = fmt.Sprint(index)
			}
			t.Fatalf("Command %q was found at index %s instead of 1", function, got)
		}
	}
}

func TestResetFails(t *testing.T) {
	{ // GIVEN a nil Controller
		var commandController Controller
		// WHEN ResetFails is run
		// THEN it exits successfully
		commandController.ResetFails()
	}

	{ // GIVEN a Controller with Commands that have failed
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
		failsBefore := len(commandController.Failed)
		commandController.ResetFails()
		// THEN all the fails become nil
		for _, failed := range commandController.Failed {
			if failed != nil {
				t.Fatalf("Reset failed, got %v", commandController.Failed)
			}
		}
		// AND the count stays the same
		failsAfter := len(commandController.Failed)
		if failsBefore != failsAfter {
			t.Fatalf("Reset added/removed elements to the Failed list. Wanted %d, got %d", failsBefore, failsAfter)
		}
	}
}
