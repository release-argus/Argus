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

//go:build unit

package command

import (
	"fmt"
	"testing"
	"time"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

// #########
// # STATE #
// #########

func TestCommands_Copy(t *testing.T) {
	// GIVEN: a Commands slice.
	tests := []struct {
		name     string
		commands *Commands
		want     Commands
		// Mutations to verify that Copy is independent at the outer slice element level.
		mutateOriginalReassignTo *Command
		wantAfterReassignFirst   Commands
		// Mutations to verify that Copy is independent at the nested slice element level.
		mutateOriginalNested bool
		mutateNestedIndexI   int
		mutateNestedIndexJ   int
		mutateNestedValue    string
	}{
		{
			name:     "nil pointer receiver",
			commands: nil,
		},
		{
			name:     "empty slice receiver",
			commands: test.Ptr(Commands{}),
			want:     Commands{},
		},
		{
			name: "non-empty commands shallow-copied",
			commands: test.Ptr(Commands{
				Command{"ls", "-la"},
				Command{"echo", "hi"},
			}),
			want: Commands{
				{"ls", "-la"},
				{"echo", "hi"},
			},

			mutateOriginalReassignTo: test.Ptr(Command{"changed"}),
			wantAfterReassignFirst: Commands{
				{"ls", "-la"},
				{"echo", "hi"},
			},

			mutateOriginalNested: true,
			mutateNestedIndexI:   1,
			mutateNestedIndexJ:   0,
			mutateNestedValue:    "X",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hadNil := tc.commands == nil

			// WHEN: Copy() is called.
			got := tc.commands.Copy()

			prefix := fmt.Sprintf("%s\nCommands.Copy()", packageName)

			// THEN: nil handling.
			if hadNil {
				if got != nil {
					t.Fatalf(
						"%s mismatch\ngot:  %v\nwant: nil",
						prefix, got,
					)
				}
				return
			}
			if got == nil {
				t.Fatalf("%s mismatch\ngot:  nil\nwant: non-nil", prefix)
			}

			// AND: the lengths match.
			if gotLen, wantLen := len(got), len(tc.want); gotLen != wantLen {
				t.Fatalf(
					"%s length mismatch\ngot:  %d\nwant: %d",
					prefix, gotLen, wantLen,
				)
			}

			// AND: initial values match.
			if testErr := test.AssertSlicesEqualFunc(
				t,
				got,
				tc.want,
				func(a, b Command) bool {
					return util.AreSlicesEqual(a, b)
				},
				fmt.Sprintf("%s mismatch", prefix),
				"input",
			); testErr != nil {
				t.Error(testErr)
			}

			// AND: reassigning an element in the original doesn't affect the copy.
			if tc.mutateOriginalReassignTo != nil {
				(*tc.commands)[0] = *tc.mutateOriginalReassignTo
				if testErr := test.AssertSlicesEqualFunc(
					t,
					got,
					tc.wantAfterReassignFirst,
					func(a, b Command) bool { return util.AreSlicesEqual(a, b) },
					fmt.Sprintf("%s mismatch after reassigning first element", prefix),
					"input",
				); testErr != nil {
					t.Error(testErr)
				}
			}

			// AND: mutating nested Command contents doesn't affect the copy.
			if tc.mutateOriginalNested {
				(*tc.commands)[tc.mutateNestedIndexI][tc.mutateNestedIndexJ] = tc.mutateNestedValue
				if got[tc.mutateNestedIndexI][tc.mutateNestedIndexJ] == tc.mutateNestedValue {
					t.Fatalf(
						"%s nested mutation unexpectedly reflected in copy i=%d j=%d\noriginal: %v\ncopy:     %v",
						prefix,
						tc.mutateNestedIndexI, tc.mutateNestedIndexJ,
						*tc.commands, got,
					)
				}
			}
		})
	}
}

func TestCommand_Copy(t *testing.T) {
	// GIVEN: a Command.
	tests := []struct {
		name string
		cmd  *Command
	}{
		{
			name: "nil",
			cmd:  nil,
		},
		{
			name: "length 1",
			cmd:  &Command{"ls"},
		},
		{
			name: "length 2",
			cmd:  &Command{"ls", "-l"},
		},
		{
			name: "length 3",
			cmd:  &Command{"ls", "-l", "-a"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Copy() is called on it.
			result := tc.cmd.Copy()

			prefix := fmt.Sprintf("%s\nCommand.Copy()", packageName)

			// THEN: if nil was copied, we got nil
			if tc.cmd == nil {
				if result != nil {
					t.Errorf(
						"%s of nil got %v, want nil",
						prefix, result,
					)
				}
				return
			}

			// AND: the copied length is the same as the source.
			if gotLen, wantLen := len(result), len(*tc.cmd); gotLen != wantLen {
				t.Errorf(
					"%s length mismatch\ngot:  %d\nwant: %d",
					prefix, gotLen, wantLen,
				)
			}

			// AND: the values match.
			for i := range result {
				if result[i] != (*tc.cmd)[i] {
					t.Errorf(
						"%s value mismatch at index %d\ngot:  %v\nwant: %v",
						packageName, i, result[i], (*tc.cmd)[i],
					)
				}
			}
		})
	}
}

func TestController_CopyFailsFrom(t *testing.T) {
	// GIVEN: a Controller with fails and a Controller to copy them to.
	tests := []struct {
		name                             string
		from, to                         *Controller
		fromFails, toFails               []*bool
		fromNextRunnable, toNextRunnable []time.Time
	}{
		{
			name:    "both nil",
			from:    nil,
			to:      nil,
			toFails: nil,
		},
		{
			name:    "from nil",
			from:    nil,
			to:      &Controller{},
			toFails: nil,
		},
		{
			name:    "to nil",
			from:    &Controller{},
			to:      nil,
			toFails: nil,
		},
		{
			name: "doesn't copy if no commands",
			from: &Controller{},
			to:   &Controller{},
			fromFails: []*bool{
				test.Ptr(true),
				test.Ptr(false),
				nil,
			},
			toFails: nil,
		},
		{
			name: "doesn't copy to new commands",
			from: &Controller{
				Command: Commands{
					{"ls", "-la"},
				},
			},
			to: &Controller{
				Command: Commands{
					{"ls", "-lah"},
				},
			},
			fromFails: []*bool{
				test.Ptr(true),
			},
			toFails: []*bool{
				nil,
			},
			fromNextRunnable: []time.Time{
				time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			toNextRunnable: []time.Time{
				{},
			},
		},
		{
			name: "does copy to retained commands",
			from: &Controller{
				Command: Commands{
					{"ls", "-lah"},
				},
			},
			to: &Controller{
				Command: Commands{
					{"ls", "-lah"},
				},
			},
			fromFails: []*bool{
				test.Ptr(true),
			},
			toFails: []*bool{
				test.Ptr(true),
			},
			fromNextRunnable: []time.Time{
				time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			toNextRunnable: []time.Time{
				time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "does copy to reordered retained commands",
			from: &Controller{
				Command: Commands{
					{"false"},
					{"ls", "-lah"},
				},
			},
			to: &Controller{
				Command: Commands{
					{"ls", "-lah"},
				},
			},
			fromFails: []*bool{
				test.Ptr(true),
				test.Ptr(false),
			},
			toFails: []*bool{
				test.Ptr(false),
			},
			fromNextRunnable: []time.Time{
				time.Date(2022, 2, 2, 0, 0, 0, 0, time.UTC),
				time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			toNextRunnable: []time.Time{
				time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.from != nil && tc.from.Command != nil {
				tc.from = NewController(
					&status.Status{},
					tc.from.Command,
					nil,
					nil,
				)
				for k, v := range tc.fromFails {
					if v != nil {
						tc.from.ServiceStatus.Fails.Command.Set(k, *v)
					}
				}
				copy(tc.from.nextRunnable, tc.fromNextRunnable)
			}
			if tc.to != nil && tc.to.Command != nil {
				tc.to = NewController(
					&status.Status{},
					tc.to.Command,
					nil,
					nil,
				)
			}

			// WHEN: CopyFailsFrom is called.
			tc.to.CopyFailsFrom(tc.from)

			prefix := fmt.Sprintf("%s\nController.CopyFailsFrom()", packageName)

			// THEN: the fails aren't copied to a nil Controller.
			if tc.toFails == nil && (tc.to == nil || tc.to.Failed == nil) {
				return
			} else if tc.to == nil {
				t.Fatalf(
					"%s .Failed mismatch\ngot:  %v\nwant: %v",
					prefix, tc.to, tc.toFails,
				)
			}
			if gotLen, wantLen := tc.to.Failed.Length(), len(tc.toFails); gotLen != wantLen {
				t.Fatalf(
					"%s .Failed length mismatch\ngot:  %d\nwant: %d",
					prefix, gotLen, wantLen,
				)
			}

			// AND: the matching fails are copied to the Controller.
			for i := range tc.toFails {
				got := test.StringifyPtr(tc.to.Failed.Get(i))
				want := test.StringifyPtr(tc.toFails[i])
				if got != want {
					t.Errorf(
						"%s .Failed[%d] mismatch\ngot:  %s\nwant: %s",
						packageName, i,
						got, want,
					)
				}
			}

			// AND: the next_runnable times are copied to the Controller.
			if testErr := test.AssertSlicesEqualFunc(
				t,
				tc.to.nextRunnable,
				tc.toNextRunnable,
				func(a, b time.Time) bool { return a.Equal(b) },
				fmt.Sprintf("%s\n", packageName),
				"nextRunnable",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

// #############
// # STRINGIFY #
// #############

func TestCommand_String(t *testing.T) {
	// GIVEN: a Command.
	tests := []struct {
		name string
		cmd  *Command
		want string
	}{
		{
			name: "empty command",
			cmd:  &Command{},
			want: "",
		},
		{
			name: "nil command",
			want: "",
		},
		{
			name: "command with no args",
			cmd:  &Command{"ls"},
			want: "ls",
		},
		{
			name: "command with one arg",
			cmd:  &Command{"ls", "-lah"},
			want: "ls -lah",
		},
		{
			name: "command with multiple args",
			cmd:  &Command{"ls", "-lah", "/root", "/tmp"},
			want: "ls -lah /root /tmp",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: the command is stringified with String().
			got := tc.cmd.String()

			prefix := fmt.Sprintf(
				"%s\nCommand.String(%+v)",
				packageName, tc.cmd,
			)

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}
		})
	}
}

func TestCommand_JSON(t *testing.T) {
	// GIVEN: a Command.
	tests := []struct {
		name string
		cmd  Command
		want string
	}{
		{
			name: "command with no args",
			cmd:  Command{"ls"},
			want: "[\"ls\"]",
		},
		{
			name: "command with one arg",
			cmd:  Command{"ls", "-lah"},
			want: "[\"ls\",\"-lah\"]",
		},
		{
			name: "command with multiple args",
			cmd:  Command{"ls", "-lah", "/root", "/tmp"},
			want: "[\"ls\",\"-lah\",\"/root\",\"/tmp\"]",
		},
		{
			name: "command with args containing spaces",
			cmd:  Command{"ls -lah", "/root /tmp"},
			want: "[\"ls -lah\",\"/root /tmp\"]",
		},
		{
			name: "invalid UTF-8 falls back to empty array",
			cmd:  Command{"\xff\xfe"},
			want: "[]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: the command is stringified with JSON.
			got := tc.cmd.JSON()

			prefix := fmt.Sprintf(
				"%s\nCommand(%+v).JSON()",
				packageName, tc.cmd,
			)

			// THEN: the result is expected.
			if got != tc.want {
				t.Errorf(
					"%s mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}
		})
	}
}
