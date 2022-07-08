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

	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
)

func testExecController(commands *Slice) Controller {
	num := 0
	if commands != nil {
		num = len(*commands)
	}
	controllerName := "TEST"
	controller := Controller{
		ServiceID: &controllerName,
		Failed:    make(Fails, num),
		Command:   commands,
	}
	return controller
}

func TestExecEmptyController(t *testing.T) {
	Init(utils.NewJLog("ERROR", false))

	// GIVEN an empty Controller
	controller := testExecController(nil)

	// WHEN executed
	err := controller.Exec(&utils.LogFrom{})

	// THEN err is nil
	if err != nil {
		t.Errorf(`%v command shouldn't have errored as it didn't do anything\n%s`, controller.Command, err.Error())
	}
}

func TestExecThatErrors(t *testing.T) {
	Init(utils.NewJLog("ERROR", false))

	// GIVEN a Command that should fail
	serviceName := "test"
	serviceInterval := "11m"
	controller := testExecController(&Slice{Command{"ls", "/root"}})
	controller.Init(jLog, &serviceName, nil, nil, nil, &serviceInterval)

	// WHEN it's executed
	err := controller.Exec(&utils.LogFrom{})

	// THEN it returns an error
	if err == nil {
		t.Errorf(`%v commands should have errored unless you're running as root`, controller.Command)
	}
}

func TestExecThatDoesntError(t *testing.T) {
	Init(utils.NewJLog("ERROR", false))

	// GIVEN a Command that should pass
	name := "test"
	interval := "11m"
	controller := testExecController(&Slice{Command{"ls"}})
	controller.Init(jLog, &name, nil, nil, nil, &interval)

	// WHEN it's executed
	err := controller.Exec(&utils.LogFrom{})

	// THEN it returns a nil error
	if err != nil {
		t.Errorf(`%v commands shouldn't have errored as we have access to the current dir\n%s`, controller.Command, err.Error())
	}
}

func testExecIndexController() Controller {
	name := "TEST"
	serviceInterval := "11m"

	controller := Controller{
		ServiceID: &name,
		Command: &Slice{
			Command{"true", "0"},
			Command{"true", "1"},
		},
		Failed: make(Fails, 2),
	}
	controller.Init(jLog, &name, nil, nil, nil, &serviceInterval)
	return controller
}

func TestExecIndexInRange(t *testing.T) {
	Init(utils.NewJLog("ERROR", false))

	// GIVEN a Controller with Commands
	controller := testExecIndexController()

	// WHEN ExecIndex is called on an index that exists
	index := 1
	errController := controller.ExecIndex(&utils.LogFrom{}, index)
	errCommand := (*controller.Command)[index].Exec(&utils.LogFrom{})

	// THEN err is the same as on the direct Exec
	if errController != errCommand {
		t.Errorf(`%q != %q`, errController.Error(), errCommand.Error())
	}
}

func TestExecIndexOutOfRange(t *testing.T) {
	Init(utils.NewJLog("ERROR", false))

	// GIVEN a Controller with Commands
	controller := testExecIndexController()

	// WHEN ExecIndex is called on an index that doesn't exist
	err := controller.ExecIndex(&utils.LogFrom{}, 2)
	// THEN err is nil
	if err != nil {
		t.Errorf(`%v command shouldn't have errored as the index was outside the bounds of the commands\n%s`, controller.Command, err.Error())
	}
}

func TestApplyTemplateWithNilServiceStatus(t *testing.T) {
	Init(utils.NewJLog("ERROR", false))

	// GIVEN a Controller with nil ServiceStatus
	controllerName := "TEST"
	controller := Controller{
		ServiceID: &controllerName,
		Command: &Slice{
			Command{"false", "{{ version }}"},
		},
		Failed: make(Fails, 1),
	}
	// WHEN ApplyTemplate is called with that nil Status
	command := (*controller.Command)[0].ApplyTemplate(controller.ServiceStatus)

	// THEN the {{ version }} var is not evaluated
	got := command.String()
	want := "false {{ version }}"
	if got != want {
		t.Errorf(`Failed with nil Status. Got %q, wanted %q`, got, want)
	}
}

func TestApplyTemplateWithServiceStatus(t *testing.T) {
	Init(utils.NewJLog("ERROR", false))

	// GIVEN a Controller with a non-nil ServiceStatus
	controllerName := "TEST"
	controller := Controller{
		ServiceID: &controllerName,
		Command: &Slice{
			Command{"false", "{{ version }}"},
		},
		Failed:        make(Fails, 1),
		ServiceStatus: &service_status.Status{LatestVersion: "1.2.3"},
	}
	// WHEN ApplyTemplate is called
	command := (*controller.Command)[0].ApplyTemplate(controller.ServiceStatus)

	// THEN the {{ version }} var is evaluated
	got := command.String()
	want := "false 1.2.3"
	if got != want {
		t.Errorf(`Failed with non-nil Status. Got %q, wanted %q`, got, want)
	}
}
