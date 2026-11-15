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
			commands: new(Commands{}),
			want:     Commands{},
		},
		{
			name: "non-empty commands deep-copied",
			commands: new(Commands{
				Command{"ls", "-la"},
				Command{"echo", "hi"},
			}),
			want: Commands{
				{"ls", "-la"},
				{"echo", "hi"},
			},

			mutateOriginalReassignTo: new(Command{"changed"}),
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

			got := tc.commands.Copy()

			if hadNil {
				if got != nil {
					t.Fatalf("after Copy: copy mismatch\ngot:  %q\nwant: nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("after Copy: copy mismatch\ngot:  nil\nwant: non-nil")
			}

			if gotLen, wantLen := len(got), len(tc.want); gotLen != wantLen {
				t.Fatalf("after Copy: length mismatch\ngot:  %d\nwant: %d", gotLen, wantLen)
			}

			if testErr := test.AssertSlicesEqualFunc(
				t,
				got,
				tc.want,
				func(a, b Command) bool {
					return util.AreSlicesEqual(a, b)
				},
				"copy",
			); testErr != nil {
				t.Errorf("after Copy: %v", testErr)
			}

			if tc.mutateOriginalReassignTo != nil {
				(*tc.commands)[0] = *tc.mutateOriginalReassignTo
				if testErr := test.AssertSlicesEqualFunc(
					t,
					got,
					tc.wantAfterReassignFirst,
					func(a, b Command) bool { return util.AreSlicesEqual(a, b) },
					"copy",
				); testErr != nil {
					t.Errorf("after reassign: %v", testErr)
				}
			}

			if tc.mutateOriginalNested {
				(*tc.commands)[tc.mutateNestedIndexI][tc.mutateNestedIndexJ] = tc.mutateNestedValue
				if got[tc.mutateNestedIndexI][tc.mutateNestedIndexJ] == tc.mutateNestedValue {
					t.Fatalf(
						"after nested mutation: copy mismatch\ngot:  %q\nwant: NOT %q",
						got[tc.mutateNestedIndexI][tc.mutateNestedIndexJ], tc.mutateNestedValue,
					)
				}
			}
		})
	}
}

func TestCommand_Copy(t *testing.T) {
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

			result := tc.cmd.Copy()

			if tc.cmd == nil {
				if result != nil {
					t.Errorf("copy mismatch\ngot:  %q\nwant: nil", result)
				}
				return
			}

			if !util.AreSlicesEqual(result, *tc.cmd) {
				t.Errorf("copy mismatch\ngot:  %q\nwant: %q", result, *tc.cmd)
			}
		})
	}
}

func TestController_CopyFailsFrom(t *testing.T) {
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
			name:    "doesn't copy if no commands",
			from:    &Controller{},
			to:      &Controller{},
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
				new(true),
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
				new(true),
			},
			toFails: []*bool{
				new(true),
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
				new(true),
				new(false),
			},
			toFails: []*bool{
				new(false),
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

			tc.to.CopyFailsFrom(tc.from)

			if tc.toFails == nil && (tc.to == nil || tc.to.Failed == nil) {
				return
			}
			if gotLen, wantLen := tc.to.Failed.Length(), len(tc.toFails); gotLen != wantLen {
				t.Fatalf("Controller.Failed.Length() mismatch\ngot:  %d\nwant: %d", gotLen, wantLen)
			}

			for i := range tc.toFails {
				got := test.StringifyPtr(tc.to.Failed.Get(i))
				want := test.StringifyPtr(tc.toFails[i])
				if got != want {
					t.Errorf("Controller.Failed.Get(%d) mismatch\ngot:  %s\nwant: %s", i, got, want)
				}
			}

			if testErr := test.AssertSlicesEqualFunc(
				t,
				tc.to.nextRunnable,
				tc.toNextRunnable,
				func(a, b time.Time) bool { return a.Equal(b) },
				"Controller.nextRunnable",
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

			got := tc.cmd.String()

			if got != tc.want {
				t.Errorf("Command.String() mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestCommand_JSON(t *testing.T) {
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

			got := tc.cmd.JSON()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}
