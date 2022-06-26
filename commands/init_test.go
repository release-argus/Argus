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
	"testing"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/utils"
)

func TestInitNils(t *testing.T) {
	// GIVEN nil Controller
	var commandController *Controller

	// WHEN Init is called
	commandController.Init(
		&utils.JLog{},
		nil,
		nil,
		nil,
		nil,
	)

	// THEN the controller stays nil
	if commandController != nil {
		t.Error("Init a nil produced a non-nil")
	}
}

func TestInitNonNil(t *testing.T) {
	// GIVEN a non-nil Controller
	commandController := &Controller{}

	// WHEN Init is called
	controllerName := "TEST"
	commandController.Init(
		utils.NewJLog("DEBUG", false),
		&controllerName,
		nil,
		&Slice{
			Command{"false"},
			Command{"false"},
		},
		&shoutrrr.Slice{},
	)

	// THEN Failed is initialised with length 2
	if len(*commandController.Command) != len(commandController.Failed) {
		t.Errorf("Failed should have been of length %d, but is length %d", len(*commandController.Command), len(commandController.Failed))
	}
}

func TestInitMetrics(t *testing.T) {
	// GIVEN a Controller with Commands
	controllerName := "TEST"
	commandController := Controller{
		ServiceID: &controllerName,
		Command: &Slice{
			Command{"false"},
			Command{"false"},
		},
		Failed: make(Fails, 2),
	}

	// WHEN initMetrics is called on this Controller
	commandController.initMetrics()

	// THEN the function doesn't hang
}

func TestFormattedStringMultiArg(t *testing.T) {
	// GIVEN a multi-arg Command
	command := Command{"ls", "-lah", "/root"}
	// WHEN FormattedString is called on it
	got := command.FormattedString()
	want := `[ "ls", "-lah", "/root" ]`
	// THEN it is returned in this format
	if got != want {
		t.Errorf("FormattedString, got %q, wanted %q", got, want)
	}
}

func TestFormattedStringSingleArg(t *testing.T) {
	// GIVEN a no-arg Command
	command := Command{"ls"}
	// WHEN FormattedString is called on it
	got := command.FormattedString()
	want := `[ "ls" ]`
	// THEN it is returned in this format
	if got != want {
		t.Errorf("FormattedString, got %q, wanted %q", got, want)
	}
}

func TestStringMultiArg(t *testing.T) {
	// GIVEN a multi-arg Command
	command := Command{"ls", "-lah", "/root"}
	// WHEN String is called on it
	got := command.String()
	want := "ls -lah /root"
	// THEN it is returned in this format
	if got != want {
		t.Errorf("String, got %q, wanted %q", got, want)
	}
}

func TestStringSingleArg(t *testing.T) {
	// GIVEN a no-arg Command
	command := Command{"ls"}
	// WHEN String is called on it
	got := command.String()
	want := "ls"
	// THEN it is returned in this format
	if got != want {
		t.Errorf("String, got %q, wanted %q", got, want)
	}
}