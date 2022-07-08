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
	"time"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/utils"
)

func TestInitNils(t *testing.T) {
	// GIVEN nil Controller
	var controller *Controller

	// WHEN Init is called
	controller.Init(
		&utils.JLog{},
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	// THEN the controller stays nil
	if controller != nil {
		t.Error("Init a nil produced a non-nil")
	}
}

func TestInitNonNil(t *testing.T) {
	// GIVEN a non-nil Controller
	controller := &Controller{}

	// WHEN Init is called
	name := "TEST"
	controller.Init(
		utils.NewJLog("DEBUG", false),
		&name,
		nil,
		&Slice{
			Command{"false"},
			Command{"false"},
		},
		&shoutrrr.Slice{},
		nil,
	)

	// THEN Failed is initialised with length 2
	if len(*controller.Command) != len(controller.Failed) {
		t.Errorf("Failed should have been of length %d, but is length %d", len(*controller.Command), len(controller.Failed))
	}
}

func TestInitMetrics(t *testing.T) {
	// GIVEN a Controller with Commands
	name := "TEST"
	controller := Controller{
		ServiceID: &name,
		Command: &Slice{
			Command{"false"},
			Command{"false"},
		},
		Failed: make(Fails, 2),
	}

	// WHEN initMetrics is called on this Controller
	controller.initMetrics()

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

func TestGetNextRunnable(t *testing.T) {
	// GIVEN a Controller with a Command with a NextRunnable time
	name := "TEST"
	controller := Controller{
		ServiceID: &name,
		Command:   nil,
		Failed:    make(Fails, 1),
	}
	controller.Init(
		&utils.JLog{},
		&name,
		nil,
		&Slice{
			Command{"false"},
		},
		nil,
		nil,
	)
	nextRunnable, _ := time.Parse(time.RFC3339, "2022-01-01T01:01:01Z")
	controller.NextRunnable[0] = nextRunnable

	// WHEN GetNextRunnable is called with an index out of bounds
	got := controller.GetNextRunnable(0)

	// THEN we got the default time.Time value
	want := nextRunnable
	if got != want {
		t.Errorf("Expected %s, got %s",
			want, got)
	}
}

func TestGetNextRunnableOutOfRange(t *testing.T) {
	// GIVEN a Controller with 5 Commands with NextRunnable times
	name := "TEST"
	controller := Controller{
		ServiceID: &name,
		Command:   nil,
		Failed:    make(Fails, 1),
	}
	controller.Init(
		&utils.JLog{},
		&name,
		nil,
		&Slice{
			Command{"false"},
		},
		nil,
		nil,
	)
	nextRunnable, _ := time.Parse(time.RFC3339, "2022-01-01T01:01:01Z")
	controller.NextRunnable[0] = nextRunnable

	// WHEN GetNextRunnable is called with an index out of bounds
	got := controller.GetNextRunnable(1)

	// THEN we got the default time.Time value
	var want time.Time
	if got != want {
		t.Errorf("Expected the default time.Time %q, got %q",
			want, got)
	}
}

func TestIsRunnableTrue(t *testing.T) {
	// GIVEN a Command with NextRunnable before the current time
	name := "TEST"
	controller := Controller{
		ServiceID: &name,
		Command:   nil,
		Failed:    make(Fails, 1),
	}
	controller.Init(
		&utils.JLog{},
		&name,
		nil,
		&Slice{
			Command{"false"},
		},
		nil,
		nil,
	)
	controller.NextRunnable[0] = time.Now().UTC().Add(-time.Minute)

	// WHEN IsRunnable is called on it
	ranAt := time.Now().UTC()
	got := controller.IsRunnable(0)

	// THEN true was returned
	want := true
	if got != want {
		t.Fatalf("IsRunnable was ran at\n%s with NextRunnable\n%s. Expected %t, got %t",
			ranAt, controller.NextRunnable[0], want, got)
	}
}

func TestIsRunnableFalse(t *testing.T) {
	// GIVEN a Command with NextRunnable after the current time
	name := "TEST"
	controller := Controller{
		ServiceID: &name,
		Command:   nil,
		Failed:    make(Fails, 1),
	}
	controller.Init(
		&utils.JLog{},
		&name,
		nil,
		&Slice{
			Command{"false"},
		},
		nil,
		nil,
	)
	controller.NextRunnable[0] = time.Now().UTC().Add(time.Minute)

	// WHEN IsRunnable is called on it
	ranAt := time.Now().UTC()
	got := controller.IsRunnable(0)

	// THEN false was returned
	want := false
	if got != want {
		t.Fatalf("IsRunnable was ran at\n%s with NextRunnable\n%s. Expected %t, got %t",
			ranAt, controller.NextRunnable[0], want, got)
	}
}

func TestIsRunnableOutOfRange(t *testing.T) {
	// GIVEN a Command with NextRunnable after the current time
	name := "TEST"
	controller := Controller{
		ServiceID: &name,
		Command:   nil,
		Failed:    make(Fails, 1),
	}
	controller.Init(
		&utils.JLog{},
		&name,
		nil,
		&Slice{
			Command{"false"},
		},
		nil,
		nil,
	)
	controller.NextRunnable[0] = time.Now().UTC().Add(-time.Minute)

	// WHEN IsRunnable is called on it with an index out of range
	ranAt := time.Now().UTC()
	got := controller.IsRunnable(1)

	// THEN false was returned
	want := false
	if got != want {
		t.Fatalf("IsRunnable was ran at\n%s with NextRunnable\n%s. Expected %t, got %t",
			ranAt, controller.NextRunnable[0], want, got)
	}
}

func TestSetNextRunnableOfPass(t *testing.T) {
	// GIVEN a Controller with a Command that passed
	name := "TEST"
	serviceInterval := "5m"
	controller := Controller{
		ServiceID: &name,
		Command:   nil,
		Failed:    make(Fails, 1),
	}
	controller.Init(
		&utils.JLog{},
		&name,
		nil,
		&Slice{
			Command{"false"},
		},
		nil,
		&serviceInterval,
	)
	failed := false
	controller.Failed[0] = &failed

	// WHEN SetNextRunnable is called on a command index that ran successfully
	controller.SetNextRunnable(0)

	// THEN MextRunnable is set to ~2*ParentInterval
	now := time.Now().UTC()
	got := controller.GetNextRunnable(0)
	parentInterval, _ := time.ParseDuration(*controller.ParentInterval)
	wantMin := now.Add(2 * parentInterval).Add(-1 * time.Second)
	wantMax := now.Add(2 * parentInterval).Add(1 * time.Second)
	if got.Before(wantMin) || got.After(wantMax) {
		t.Errorf("Expected between %s and %s, not %s",
			wantMin, wantMax, got)
	}
}

func TestSetNextRunnableOfFail(t *testing.T) {
	// GIVEN a Controller with a Command that failed
	name := "TEST"
	serviceInterval := "5m"
	controller := Controller{
		ServiceID:      &name,
		Command:        nil,
		Failed:         make(Fails, 1),
		ParentInterval: &serviceInterval,
	}
	controller.Init(
		&utils.JLog{},
		&name,
		nil,
		&Slice{
			Command{"false"},
		},
		nil,
		nil,
	)
	failed := true
	controller.Failed[0] = &failed

	// WHEN SetNextRunnable is called on a command index that failed running
	controller.SetNextRunnable(0)

	// THEN MextRunnable is set to 15s
	now := time.Now().UTC()
	got := controller.GetNextRunnable(0)
	wantMin := now.Add(14 * time.Second)
	wantMax := now.Add(16 * time.Second)
	if got.Before(wantMin) || got.After(wantMax) {
		t.Errorf("Expected between %s and %s, not %s",
			wantMin, wantMax, got)
	}
}

func TestSetNextRunnableOutOfRange(t *testing.T) {
	// GIVEN a Controller with 1 Command with NextRunnable times
	name := "TEST"
	controller := Controller{
		ServiceID: &name,
		Command:   nil,
		Failed:    make(Fails, 1),
	}
	controller.Init(
		&utils.JLog{},
		&name,
		nil,
		&Slice{
			Command{"false"},
		},
		nil,
		nil,
	)
	nextRunnable, _ := time.Parse(time.RFC3339, "2022-01-01T01:01:01Z")
	controller.NextRunnable[0] = nextRunnable

	// WHEN SetNextRunnable is called with an index out of bounds
	controller.SetNextRunnable(1)

	// THEN no Controller.NextRunnable has been changed
	for i := range controller.NextRunnable {
		if controller.NextRunnable[i] != nextRunnable {
			t.Errorf("NextRunnable modified from %s, got %s",
				nextRunnable, controller.NextRunnable[i])
		}
	}
}
