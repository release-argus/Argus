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
	"time"

	api_types "github.com/release-argus/Argus/web/api/types"
)

func TestAnnounceCommandNoChannel(t *testing.T) {
	// GIVEN a Controller with no Announce channel
	controllerName := "TEST"
	parentInterval := "10m"
	controller := Controller{
		ServiceID: &controllerName,
		Command: &Slice{
			Command{"ls", "-lah"},
		},
		Failed:         make(Fails, 1),
		NextRunnable:   make([]time.Time, 1),
		ParentInterval: &parentInterval,
	}

	// WHEN AnnounceCommand is run
	controller.AnnounceCommand(0)

	// THEN the function doesn't hang
}

func testAnnounceCommand() Controller {
	controllerName := "TEST"
	channel := make(chan []byte)
	parentInterval := "5m"
	controller := Controller{
		ServiceID: &controllerName,
		Command: &Slice{
			Command{"ls", "-lah"},
		},
		Failed:         make(Fails, 1),
		NextRunnable:   make([]time.Time, 1),
		Announce:       &channel,
		ParentInterval: &parentInterval,
	}
	return controller
}

func TestAnnounceCommandWhenNotRun(t *testing.T) {
	// GIVEN a Controller with an Announce channel
	controller := testAnnounceCommand()

	// WHEN AnnounceCommand is run with the command not having been run
	controller.Failed[0] = nil
	go controller.AnnounceCommand(0)
	result := <-*controller.Announce
	var parsedResult api_types.WebSocketMessage
	json.Unmarshal(result, &parsedResult)

	// THEN it broadcasts nil to the Announce channel
	if parsedResult.CommandData["ls -lah"].Failed != nil {
		t.Errorf("got %t. expected %s",
			*parsedResult.CommandData["ls -lah"].Failed, "false")
	}
}

func TestAnnounceCommandWhenRunPassed(t *testing.T) {
	// GIVEN a Controller with an Announce channel
	controller := testAnnounceCommand()

	// WHEN AnnounceCommand is run with the command not failing
	failed := false
	controller.Failed[0] = &failed
	go controller.AnnounceCommand(0)
	result := <-*controller.Announce
	var parsedResult api_types.WebSocketMessage
	json.Unmarshal(result, &parsedResult)

	// THEN it broadcasts failed=false to the Announce channel
	if *parsedResult.CommandData["ls -lah"].Failed != false {
		got := "nil"
		if parsedResult.CommandData["ls -lah"].Failed != nil {
			got = fmt.Sprint(*parsedResult.CommandData["ls -lah"].Failed)
		}
		t.Errorf("failed - got %s. expected %s",
			got, "false")
	}
	// and NextRunnable is ~2*ParentInterval
	now := time.Now().UTC()
	parentInterval, _ := time.ParseDuration(*controller.ParentInterval)
	wantAfter := now.Add(2 * parentInterval).Add(-time.Second)
	wantBefore := now.Add(2 * parentInterval).Add(time.Second)
	got := parsedResult.CommandData["ls -lah"].NextRunnable
	if got.Before(wantAfter) ||
		got.After(wantBefore) {
		t.Errorf("next_runnable - got %s. expected between %s and %s",
			got, wantAfter, wantBefore)
	}
}

func TestAnnounceCommandWhenRunFailed(t *testing.T) {
	// GIVEN a Controller with an Announce channel
	controller := testAnnounceCommand()

	// WHEN AnnounceCommand is run with the command failing
	failed := true
	controller.Failed[0] = &failed
	go controller.AnnounceCommand(0)
	result := <-*controller.Announce
	var parsedResult api_types.WebSocketMessage
	json.Unmarshal(result, &parsedResult)

	// THEN it broadcasts failed=true to the Announce channel
	if *parsedResult.CommandData["ls -lah"].Failed != true {
		got := "nil"
		if parsedResult.CommandData["ls -lah"].Failed != nil {
			got = fmt.Sprint(parsedResult.CommandData["ls -lah"].Failed)
		}
		t.Errorf("failed - got %s. expected %s",
			got, "true")
	}
	// and NextRunnable is ~15s
	now := time.Now().UTC()
	wantAfter := now.Add(14 * time.Second)
	wantBefore := now.Add(16 * time.Second)
	got := parsedResult.CommandData["ls -lah"].NextRunnable
	if got.Before(wantAfter) ||
		got.After(wantBefore) {
		t.Errorf("next_runnable - got %s. expected between %s and %s",
			got, wantAfter, wantBefore)
	}
}

func testFindController() Controller {
	controllerName := "TEST"
	controller := Controller{
		ServiceID: &controllerName,
		Command: &Slice{
			Command{"ls", "-lah"},
			Command{"ls", "-lah", "a"},
			Command{"ls", "-lah", "b"},
		},
		Failed: make(Fails, 3),
	}
	return controller
}

func TestFindUnknown(t *testing.T) {
	// GIVEN we have a Slice of Commands
	controller := testFindController()
	var index *int

	// WHEN Find is run for an unknown command
	function := "random function"
	index = controller.Find(function)

	// THEN nil is returned
	if index != nil {
		t.Errorf("Command %q was found at index %d instead of nil",
			function, *index)
	}
}

func TestFindKnown(t *testing.T) {
	// GIVEN we have a Slice of Commands
	controller := testFindController()
	var index *int

	// WHEN Find is run with a known command
	function := "ls -lah a"
	index = controller.Find(function)

	// THEN it returns the correct index of that command
	if index == nil || *index != 1 {
		got := "nil"
		if index != nil {
			got = fmt.Sprint(index)
		}
		t.Errorf("Command %q was found at index %s instead of 1",
			function, got)
	}
}

func TestResetFailsNilController(t *testing.T) {
	// GIVEN a nil Controller
	var controller Controller

	// WHEN ResetFails is run
	controller.ResetFails()

	// THEN the command doesn't hang
}

func TestResetFailsNonNilController(t *testing.T) {
	// GIVEN a Controller with Commands that have failed
	controllerName := "TEST"
	controller := Controller{
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
	controller.Failed[0] = &failed0
	failed1 := true
	controller.Failed[1] = &failed1
	failed2 := false
	controller.Failed[2] = &failed2
	failed3 := true
	controller.Failed[3] = &failed3
	failed4 := true
	controller.Failed[4] = &failed4

	// WHEN ResetFails is called
	controller.ResetFails()
	// THEN all the fails become nil
	for _, failed := range controller.Failed {
		if failed != nil {
			t.Errorf("Reset failed, got %v",
				controller.Failed)
		}

	}
}
