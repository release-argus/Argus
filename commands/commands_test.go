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
	"testing"

	"github.com/release-argus/Argus/utils"
)

func TestExec(t *testing.T) {
	Init(utils.NewJLog("ERROR", false))

	{ // GIVEN an empty Controller
		controllerName := "TEST"
		commandController := Controller{
			ServiceID: &controllerName,
			Failed:    make(Fails, 0),
		}
		// WHEN executed
		err := commandController.Exec(&utils.LogFrom{})
		// THEN err is nil
		if err != nil {
			t.Fatalf(`%v command shouldn't have errored as it didn't do anything\n%s`, commandController.Command, err.Error())
		}
	}

	{ // GIVEN a Command that should fail
		controllerName := "TEST"
		commandController := Controller{
			ServiceID: &controllerName,
		}
		commandController.Init(nil, &controllerName, &Slice{Command{"ls", "/root"}}, nil)
		{
			// WHEN it's stringified with .String()
			// THEN it's joined with spaces
			if (*commandController.Command)[0].String() != "ls /root" {
				t.Fatalf(`Command didn't .String() correctly. Expected %q, got %q`, (*commandController.Command)[0].String(), "ls /root")
			}

			// WHEN it's executed
			err := commandController.Exec(&utils.LogFrom{})
			// THEN it returns and error
			if err == nil {
				t.Fatalf(`%v commands should have errored unless you're running as root`, commandController.Command)
			}
		}
	}

	{ // GIVEN a Command that should pass
		controllerName := "TEST"
		commandController := Controller{
			ServiceID: &controllerName,
		}
		commandController.Init(nil, &controllerName, &Slice{Command{"ls"}}, nil)
		{
			// WHEN it's stringified with .String()
			// THEN it's returned as the correct string
			if (*commandController.Command)[0].String() != "ls" {
				t.Fatalf(`Command didn't .String() correctly. Expected %q, got %q`, (*commandController.Command)[0].String(), "ls")
			}

			// WHEN it's executed
			err := commandController.Exec(&utils.LogFrom{})
			// THEN it returns a nil error
			if err != nil {
				t.Fatalf(`%v commands shouldn't have errored as we have access to the current dir\n%s`, commandController.Command, err.Error())
			}
		}
	}
}

func TestExecIndex(t *testing.T) {
	Init(utils.NewJLog("ERROR", false))

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
	{ // WHEN ExecIndex is called on an index that exists
		err := commandController.ExecIndex(&utils.LogFrom{}, 0)
		// THEN err is nil
		if err == nil {
			t.Fatalf(`%q command shouldn't have errored as it was just an\n%s`, (*commandController.Command)[0].String(), err.Error())
		}
		// WHEN ExecIndex is called on an index that exists
		err = commandController.ExecIndex(&utils.LogFrom{}, 1)
		// THEN err is nil
		if err == nil {
			t.Fatalf(`%q command shouldn't have errored as it was just an\n%s`, (*commandController.Command)[1].String(), err.Error())
		}

		// WHEN ExecIndex is called on an index that doesn't exist
		err = commandController.ExecIndex(&utils.LogFrom{}, 2)
		// THEN err is nil
		if err != nil {
			t.Fatalf(`%v command shouldn't have errored as the index was outside the bounds\n%s`, commandController.Command, err.Error())
		}
	}
}
