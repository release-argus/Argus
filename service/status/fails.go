// Copyright [2025] [Argus]
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

// Package status provides the status functionality to keep track of the approved/deployed/latest versions of a Service.
package status

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/release-argus/Argus/util"
)

// Fails keeps track of the unsent/fail/pass status of the different senders.
type Fails struct {
	Command  FailsCommand  `yaml:"-" json:"-"` // Command unsent/fail/pass.
	Shoutrrr FailsShoutrrr `yaml:"-" json:"-"` // Shoutrrr unsent/fail/pass.
	WebHook  FailsWebHook  `yaml:"-" json:"-"` // WebHook unsent/fail/pass.
}

// failsBase is the base struct for the Fails structs.
type failsBase struct {
	mutex sync.RWMutex     // Mutex for concurrent access.
	fails map[string]*bool // Map of index to fail status.
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

	for _, fail := range f.fails {
		if util.DereferenceOrNilValue(fail, true) {
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
func (f *failsBase) String(prefix string) string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	if len(f.fails) == 0 {
		return ""
	}

	var builder strings.Builder

	keys := util.SortedKeys(f.fails)

	for _, key := range keys {
		val := "nil"
		if f.fails[key] != nil {
			val = strconv.FormatBool(*f.fails[key])
		}
		builder.WriteString(fmt.Sprintf("%s%s: %s\n",
			prefix, key, val))
	}

	return builder.String()
}

// FailsCommand keeps track of the unsent/fail/pass status of the Command sender.
type FailsCommand struct {
	failsBase
	fails []*bool
}

// Init will initialise the slice with the given length.
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

	for _, fail := range f.fails {
		if util.DereferenceOrNilValue(fail, true) {
			return false
		}
	}

	return true
}

// Reset clears all the entries in the fails slice by setting each element to nil.
func (f *FailsCommand) Reset() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	for i := range f.fails {
		f.fails[i] = nil
	}
}

// Length returns the amount of elements in the fails slice.
// in a thread-safe manner.
func (f *FailsCommand) Length() int {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return len(f.fails)
}

// String representation of FailsCommand.
func (f *FailsCommand) String(prefix string) string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	if len(f.fails) == 0 {
		return ""
	}

	var builder strings.Builder

	for i, fail := range f.fails {
		val := "nil"
		if fail != nil {
			val = strconv.FormatBool(*fail)
		}

		builder.WriteString(fmt.Sprintf("%s- %d: %s\n",
			prefix, i, val))
	}

	return builder.String()
}

// FailsShoutrrr keeps track of the unsent/fail/pass status of the Shoutrrr sender.
type FailsShoutrrr struct {
	failsBase
}

// FailsWebHook keeps track of the unsent/fail/pass status of the WebHook sender.
type FailsWebHook struct {
	failsBase
}

// String returns a string representation of the Fails.
func (f *Fails) String(prefix string) string {
	var builder strings.Builder
	itemPrefix := prefix + "  "

	if shoutrrrStr := f.Shoutrrr.String(itemPrefix); shoutrrrStr != "" {
		builder.WriteString(fmt.Sprintf("%sshoutrrr:\n%s",
			prefix, shoutrrrStr))
	}

	if commandStr := f.Command.String(itemPrefix); commandStr != "" {
		builder.WriteString(fmt.Sprintf("%scommand:\n%s",
			prefix, commandStr))
	}

	if webhookStr := f.WebHook.String(itemPrefix); webhookStr != "" {
		builder.WriteString(fmt.Sprintf("%swebhook:\n%s",
			prefix, webhookStr))
	}

	if builder.Len() == 0 {
		return ""
	}

	result := builder.String()
	return result
}

// Copy copies the contents of another Fails into this one.
func (f *Fails) Copy(from *Fails) {
	// Command.
	f.Command.mutex.Lock()
	defer from.Command.mutex.RUnlock()
	from.Command.mutex.RLock()
	defer f.Command.mutex.Unlock()

	f.Command.fails = make([]*bool, len(from.Command.fails))
	copy(f.Command.fails, from.Command.fails)

	// Shoutrrr.
	f.Shoutrrr.mutex.Lock()
	defer f.Shoutrrr.mutex.Unlock()
	from.Shoutrrr.mutex.RLock()
	defer from.Shoutrrr.mutex.RUnlock()

	f.Shoutrrr.fails = make(map[string]*bool, len(from.Shoutrrr.fails))
	for key, fail := range from.Shoutrrr.fails {
		f.Shoutrrr.fails[key] = fail
	}

	// WebHook.
	f.WebHook.mutex.Lock()
	defer f.WebHook.mutex.Unlock()
	from.WebHook.mutex.RLock()
	defer from.WebHook.mutex.RUnlock()

	f.WebHook.fails = make(map[string]*bool, len(from.WebHook.fails))
	for key, fail := range from.WebHook.fails {
		f.WebHook.fails[key] = fail
	}
}

// resetFails resets the state of the Command, Shoutrrr, and WebHook fails.
func (f *Fails) resetFails() {
	f.Command.Reset()
	f.Shoutrrr.Reset()
	f.WebHook.Reset()
}
