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

func TestCommands(t *testing.T) {
	Init(utils.NewJLog("ERROR", false))

	var commandController Controller
	commandController.Failed = make(Fails, 1)
	err := commandController.Exec(&utils.LogFrom{})
	name := "TEST"
	commandController.ServiceID = &name
	if err != nil {
		t.Fatalf(`%v command shouldn't have errored as it didn't do anything\n%s`, commandController.Command, err.Error())
	}

	commandController.Command = &Slice{Command{"ls", "/root"}}
	if (*commandController.Command)[0].String() != "ls /root" {
		t.Fatalf(`Command didn't .String() correctly. Expected %q, got %q`, (*commandController.Command)[0].String(), "ls /root")
	}
	err = commandController.Exec(&utils.LogFrom{})
	if err == nil {
		t.Fatalf(`%v commands should have errored unless you're running as root`, commandController.Command)
	}

	commandController.Command = &Slice{Command{"ls"}}
	if (*commandController.Command)[0].String() != "ls" {
		t.Fatalf(`Command didn't .String() correctly. Expected %q, got %q`, (*commandController.Command)[0].String(), "ls")
	}
	err = commandController.Exec(&utils.LogFrom{})
	if err != nil {
		t.Fatalf(`%v commands shouldn't have errored as we have access to the current dir\n%s`, commandController.Command, err.Error())
	}
}
