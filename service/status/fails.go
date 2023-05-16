// Copyright [2023] [Argus]
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

package svcstatus

import (
	"fmt"
	"sync"

	"github.com/release-argus/Argus/util"
)

// Fails keeps track of the unsent/fail/pass status of the different senders.
type Fails struct {
	Command  FailsCommand  `yaml:"-" json:"-"` // Command unsent/fail/pass.
	Shoutrrr FailsShoutrrr `yaml:"-" json:"-"` // Shoutrrr unsent/fail/pass.
	WebHook  FailsWebHook  `yaml:"-" json:"-"` // WebHook unsent/fail/pass.
}

// failsBase is a base struct for the Fails structs.
type failsBase struct {
	fails map[string]*bool // map of index to fail status.
	mutex sync.RWMutex     // Mutex for concurrent access.
}

// Init the failsBase.
func (f *failsBase) Init(length int) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.fails = make(map[string]*bool, length)
}

// Get the fail status of this index.
func (f *failsBase) Get(index string) *bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return f.fails[index]
}

// Set the fail state of this index.
func (f *failsBase) Set(index string, state *bool) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.fails[index] = state
}

// AllPassed returns whether all the indexes have passed (fail=false).
func (f *failsBase) AllPassed() bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	for i := range f.fails {
		if util.EvalNilPtr(f.fails[i], true) {
			return false
		}
	}

	return true
}

// Reset of the indexes.
func (f *failsBase) Reset() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	for i := range f.fails {
		f.fails[i] = nil
	}
}

// Length of the failsBase.
func (f *failsBase) Length() int {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return len(f.fails)
}

// String representation of failsBase.
func (f *failsBase) String() (str string) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	keys := util.SortedKeys(f.fails)

	for _, key := range keys {
		val := "nil"
		if f.fails[key] != nil {
			val = fmt.Sprint(*f.fails[key])
		}

		str += fmt.Sprintf("%v: %s, ", key, val)
	}
	if len(str) != 0 {
		str = "{" + str[:len(str)-2] + "}"
	}

	return
}

type FailsCommand struct {
	failsBase
	fails []*bool
}

func (f *FailsCommand) Init(length int) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.fails = make([]*bool, length)
}

// Get the fail status of the Command at this index.
func (f *FailsCommand) Get(index int) *bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return f.fails[index]
}

// Set the fail status of the Command at this index.
func (f *FailsCommand) Set(index int, state bool) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.fails[index] = &state
}

// AllPassed returns whether all the Commands have passed (fail=false).
func (f *FailsCommand) AllPassed() bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	for i := range f.fails {
		if util.EvalNilPtr(f.fails[i], true) {
			return false
		}
	}

	return true
}

// Reset of the Command's.
func (f *FailsCommand) Reset() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	for i := range f.fails {
		f.fails[i] = nil
	}
}

// Length of the FailsCommand.
func (f *FailsCommand) Length() int {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return len(f.fails)
}

// String representation of FailsCommand.
func (f *FailsCommand) String() (str string) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	for i := range f.fails {
		val := "nil"
		if f.fails[i] != nil {
			val = fmt.Sprint(*f.fails[i])
		}

		str += fmt.Sprintf("%v: %s, ", i, val)
	}
	if len(str) != 0 {
		str = "[" + str[:len(str)-2] + "]"
	}

	return
}

type FailsShoutrrr struct {
	failsBase
}

type FailsWebHook struct {
	failsBase
}

// String returns a string representation of the Fails.
func (f *Fails) String() (str string) {
	str = ""

	shoutrrrStr := f.Shoutrrr.String()
	if shoutrrrStr != "" {
		str += fmt.Sprintf("shoutrrr: %s, ", shoutrrrStr)
	}

	commandStr := f.Command.String()
	if commandStr != "" {
		str += fmt.Sprintf("command: %s, ", commandStr)
	}

	webhookStr := f.WebHook.String()
	if webhookStr != "" {
		str += fmt.Sprintf("webhook: %s, ", webhookStr)
	}

	if len(str) != 0 {
		// Trim the trailing ', '
		str = str[:len(str)-2]
	}
	return
}

// Reset all the fails (nil them).
func (f *Fails) resetFails() {
	f.Command.Reset()
	f.Shoutrrr.Reset()
	f.WebHook.Reset()
}
