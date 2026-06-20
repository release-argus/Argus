// Copyright [2026] [Argus]
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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/release-argus/Argus/util"
)

// Fails keeps track of the unsent/fail/pass status of the different senders.
type Fails struct {
	Command  FailsCommand  `json:"-" yaml:"-"` // Command unsent/fail/pass.
	Shoutrrr FailsShoutrrr `json:"-" yaml:"-"` // Shoutrrr unsent/fail/pass.
	WebHook  FailsWebHook  `json:"-" yaml:"-"` // WebHook unsent/fail/pass.
}

// failsBase is the base struct for the Fails structs.
type failsBase struct {
	mu    sync.RWMutex     // Mutex for concurrent access.
	fails map[string]*bool // Map of index to fail status.
}

// Init initialises the internal failure map with the given capacity.
func (f *failsBase) Init(length int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.fails = make(map[string]*bool, length)
}

// Copy returns a deep copy of the receiver.
func (f *failsBase) Copy() *failsBase {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return &failsBase{
		fails: util.CopyMap(f.fails),
	}
}

// Get returns the fail status of the given index.
func (f *failsBase) Get(index string) *bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.fails[index]
}

// Set updates the fail state for the given index.
func (f *failsBase) Set(index string, state *bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.fails[index] = state
}

// AllPassed returns whether all the indexes have passed (fail=false).
func (f *failsBase) AllPassed() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for _, fail := range f.fails {
		if util.DerefOr(fail, true) {
			return false
		}
	}

	return true
}

// Reset sets the fail state of all indexes to nil.
func (f *failsBase) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()

	for i := range f.fails {
		f.fails[i] = nil
	}
}

// Length returns the number of tracked entries.
func (f *failsBase) Length() int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return len(f.fails)
}

// String returns a string representation of the fail states.
func (f *failsBase) String(prefix string) string {
	f.mu.RLock()
	defer f.mu.RUnlock()
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
		// '<prefix><key>: <val>\n'
		builder.WriteString(prefix)
		builder.WriteString(key)
		builder.WriteString(": ")
		builder.WriteString(val)
		builder.WriteString("\n")
	}

	return builder.String()
}

// FailsCommand keeps track of the unsent/fail/pass status of the Command sender.
type FailsCommand struct {
	failsBase
	fails []*bool
}

// Init initialises the internal failure slice with the given capacity.
func (f *FailsCommand) Init(length int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.fails = make([]*bool, length)
}

// Get returns the fail status of the Command at the given index.
func (f *FailsCommand) Get(index int) *bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.fails[index]
}

// Set updates the fail status of the Command at the given index.
func (f *FailsCommand) Set(index int, state bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.fails[index] = &state
}

// AllPassed returns whether all the Commands have passed (fail=false).
func (f *FailsCommand) AllPassed() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for _, fail := range f.fails {
		if util.DerefOr(fail, true) {
			return false
		}
	}

	return true
}

// Reset clears all the entries in the fails slice by setting each element to nil.
func (f *FailsCommand) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()

	for i := range f.fails {
		f.fails[i] = nil
	}
}

// Length returns the number of elements in the fails slice.
func (f *FailsCommand) Length() int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return len(f.fails)
}

// String returns a string representation of the fail states.
func (f *FailsCommand) String(prefix string) string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if len(f.fails) == 0 {
		return ""
	}

	var builder strings.Builder

	for i, fail := range f.fails {
		val := "nil"
		if fail != nil {
			val = strconv.FormatBool(*fail)
		}

		// '<prefix>- <i>: <val>\n'
		builder.WriteString(prefix)
		builder.WriteString("- ")
		builder.WriteString(strconv.Itoa(i))
		builder.WriteString(": ")
		builder.WriteString(val)
		builder.WriteString("\n")
	}

	return builder.String()
}

// FailsShoutrrr keeps track of the unsent/fail/pass status of the Shoutrrr sender.
type FailsShoutrrr struct {
	failsBase
}

// Copy returns a deep copy of the receiver.
func (f *FailsShoutrrr) Copy() *FailsShoutrrr {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return &FailsShoutrrr{
		failsBase: *f.failsBase.Copy(),
	}
}

// FailsWebHook keeps track of the unsent/fail/pass status of the WebHook sender.
type FailsWebHook struct {
	failsBase
	nextRunnable map[string]time.Time // Map of index to time at which can next run (for staggering).
}

// Init initialises the internal next runnable map with the given capacity.
func (f *FailsWebHook) Init(length int) {
	f.failsBase.Init(length)
	f.mu.Lock()
	defer f.mu.Unlock()

	f.nextRunnable = make(map[string]time.Time, length)
}

// Copy returns a deep copy of the receiver.
func (f *FailsWebHook) Copy() *FailsWebHook {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return &FailsWebHook{
		failsBase:    *f.failsBase.Copy(),
		nextRunnable: util.CopyMap(f.nextRunnable),
	}
}

// NextRunnable returns the next time at which the index can be re-run.
func (f *FailsWebHook) NextRunnable(index string) time.Time {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.nextRunnable[index]
}

// SetNextRunnable updates the time at which the given index can be re-run.
func (f *FailsWebHook) SetNextRunnable(index string, time time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.nextRunnable[index] = time
}

// String returns a string representation of the receiver.
func (f *Fails) String(prefix string) string {
	var builder strings.Builder
	itemPrefix := prefix + "  "

	if shoutrrrStr := f.Shoutrrr.String(itemPrefix); shoutrrrStr != "" {
		// '<prefix>shoutrrr:\n<shoutrrrStr>'
		builder.WriteString(prefix)
		builder.WriteString("shoutrrr:\n")
		builder.WriteString(shoutrrrStr)
	}

	if commandStr := f.Command.String(itemPrefix); commandStr != "" {
		// '<prefix>command:\n<commandStr>'
		builder.WriteString(prefix)
		builder.WriteString("command:\n")
		builder.WriteString(commandStr)
	}

	if webhookStr := f.WebHook.String(itemPrefix); webhookStr != "" {
		// '<prefix>webhook:\n<commandStr>'
		builder.WriteString(prefix)
		builder.WriteString("webhook:\n")
		builder.WriteString(webhookStr)
	}

	if builder.Len() == 0 {
		return ""
	}

	result := builder.String()
	return result
}

// Copy returns a deep copy of the receiver.
func (f *Fails) Copy(from *Fails) {
	// Command.
	f.Command.mu.Lock()
	defer from.Command.mu.RUnlock()
	from.Command.mu.RLock()
	defer f.Command.mu.Unlock()

	f.Command.fails = make([]*bool, len(from.Command.fails))
	copy(f.Command.fails, from.Command.fails)

	// Shoutrrr.
	f.Shoutrrr.mu.Lock()
	defer f.Shoutrrr.mu.Unlock()
	from.Shoutrrr.mu.RLock()
	defer from.Shoutrrr.mu.RUnlock()

	f.Shoutrrr.fails = make(map[string]*bool, len(from.Shoutrrr.fails))
	for key, fail := range from.Shoutrrr.fails {
		f.Shoutrrr.fails[key] = fail
	}

	// WebHook.
	f.WebHook.mu.Lock()
	defer f.WebHook.mu.Unlock()
	from.WebHook.mu.RLock()
	defer from.WebHook.mu.RUnlock()

	f.WebHook.fails = make(map[string]*bool, len(from.WebHook.fails))
	for key, fail := range from.WebHook.fails {
		f.WebHook.fails[key] = fail
	}
}

// resetFails clears command, shoutrrr, and webhook failure tracking.
func (f *Fails) resetFails() {
	f.Command.Reset()
	f.Shoutrrr.Reset()
	f.WebHook.Reset()
}
