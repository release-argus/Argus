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

package config

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TehstSetDefaultsService(t *testing.T) {
	// GIVEN nil defaults
	var defaults Defaults

	// WHEN SetDefaults is called on it
	defaults.SetDefaults()

	// THEN the Service part is initialised to the defined defaults
	want := "10m"
	got := *defaults.Service.Interval
	if got != want {
		t.Errorf("defaults.Service.Interval should have been %s, but got %s",
			want, got)
	}
}

func TestDefaultsSetDefaultsNotify(t *testing.T) {
	// GIVEN nil defaults
	var defaults Defaults

	// WHEN SetDefaults is called on it
	defaults.SetDefaults()

	// THEN the Notify part is initialised to the defined defaults
	want := "Argus"
	got := defaults.Notify["discord"].GetSelfParam("username")
	if got != want {
		t.Errorf("defaults.Notify.discord.Params.username should have been %s, but got %s",
			want, got)
	}
}

func TestDefaultsSetDefaultsWebHook(t *testing.T) {
	// GIVEN nil defaults
	var defaults Defaults

	// WHEN SetDefaults is called on it
	defaults.SetDefaults()

	// THEN the WebHook part is initialised to the defined defaults
	want := "github"
	got := *defaults.WebHook.Type
	if got != want {
		t.Errorf("defaults.WebHook.Type should have been %s, but got %s",
			want, got)
	}
}

func TestDefaultsCheckValuesWithInvalidService(t *testing.T) {
	// GIVEN defaults with an invalid Service.Interval
	var defaults Defaults
	defaults.SetDefaults()
	*defaults.Service.Interval = "10x"

	// WHEN CheckValues is called on it
	err := defaults.CheckValues()

	// THEN err is non-nil
	if err == nil {
		t.Errorf("err shouldn't be %v, Service.Interval was invalid with %q",
			err, *defaults.Service.Interval)
	}
}

func TestDefaultsCheckValuesWithInvalidServiceDeployedVersionLookup(t *testing.T) {
	// GIVEN defaults with an invalid Service.DeployedVersionLookup.Regex
	var defaults Defaults
	defaults.SetDefaults()
	defaults.Service.DeployedVersionLookup.Regex = "^something[0-"

	// WHEN CheckValues is called on it
	err := defaults.CheckValues()

	// THEN err is non-nil
	if err == nil {
		t.Errorf("err shouldn't be %v, Service.DeployedVersionLookup.Regex was invalid with %q",
			err, defaults.Service.DeployedVersionLookup.Regex)
	}
}

func TestDefaultsCheckValuesWithInvalidNotify(t *testing.T) {
	// GIVEN defaults with an invalid Notify.slack.Delay
	var defaults Defaults
	defaults.SetDefaults()
	defaults.Notify["slack"].SetOption("delay", "10x")

	// WHEN CheckValues is called on it
	err := defaults.CheckValues()

	// THEN err is non-nil
	if err == nil {
		t.Errorf("err shouldn't be %v, Notify.slack.Delay was invalid with %q",
			err, defaults.Notify["slack"].GetSelfOption("delay"))
	}
}

func TestDefaultsCheckValuesWithInvalidWebHook(t *testing.T) {
	// GIVEN defaults with an invalid WebHook.Delay
	var defaults Defaults
	defaults.SetDefaults()
	*defaults.WebHook.Delay = "10x"

	// WHEN CheckValues is called on it
	err := defaults.CheckValues()

	// THEN err is non-nil
	if err == nil {
		t.Errorf("err shouldn't be %v, WebHook.Delay was invalid with %q",
			err, *defaults.WebHook.Delay)
	}
}

func TestDefaultsPrint(t *testing.T) {
	// GIVEN unmodified defaults from SetDefaults
	var defaults Defaults
	defaults.SetDefaults()
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Print is called
	defaults.Print()

	// THEN the expected number of lines are printed
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	want := 142
	got := strings.Count(string(out), "\n")
	if got != want {
		t.Errorf("Print should have given %d lines, but gave %d\n%s",
			want, got, out)
	}
}
