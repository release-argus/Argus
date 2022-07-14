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

package shoutrrr

import (
	"strconv"
	"testing"
	"time"

	"github.com/release-argus/Argus/utils"
)

func testGet() Shoutrrr {
	sID := "test"
	failed := false

	hardDefaults := Shoutrrr{
		Options: map[string]string{
			"delay": "3s",
		},
		URLFields: map[string]string{
			"token":     "harddefaults token",
			"webhookid": "harddefaults webhookid",
		},
		Params: map[string]string{
			"title": "harddefaults",
		},
	}
	defaults := Shoutrrr{
		Options: map[string]string{
			"delay": "2s",
		},
		URLFields: map[string]string{
			"token":     "harddefaults token",
			"webhookid": "harddefaults webhookid",
		},
		Params: map[string]string{
			"title": "harddefaults",
		},
	}
	main := Shoutrrr{
		Options: map[string]string{
			"delay": "1s",
		},
		URLFields: map[string]string{
			"token":     "main token",
			"webhookid": "main webhookid",
		},
		Params: map[string]string{
			"title": "main",
		},
	}

	return Shoutrrr{
		ID:     &sID,
		Type:   "discord",
		Failed: &failed,
		Options: map[string]string{
			"delay":     "0s",
			"message":   "{% if 'a' == 'a' %}{{ version }}-foo{% endif %}",
			"max_tries": "4",
		},
		URLFields: map[string]string{
			"token":     "master token",
			"webhookid": "master webhookid",
		},
		Params: map[string]string{
			"avatar": "argus.png",
			"title":  "{% if 'a' == 'a' %}{{ version }}-master{% endif %}",
		},
		HardDefaults: &hardDefaults,
		Defaults:     &defaults,
		Main:         &main,
	}
}

func TestGetOption(t *testing.T) {
	// GIVEN a Shoutrrr with the closest Options.delay in Main
	shoutrrr := testGet()
	shoutrrr.Options = map[string]string{}

	// WHEN GetOption is called with "delay"
	got := shoutrrr.GetOption("delay")

	// THEN we get the delay of the master Shoutrrr
	want := shoutrrr.Main.Options["delay"]
	if got != want {
		t.Errorf("Should have got %q delay from the main, not %s",
			want, got)
	}
}

func TestGetSelfOption(t *testing.T) {
	// GIVEN a Shoutrrr with Options.delay
	shoutrrr := testGet()

	// WHEN GetSelfOption is called with "delay"
	got := shoutrrr.GetSelfOption("delay")

	// THEN we get the delay of the master Shoutrrr
	want := shoutrrr.Options["delay"]
	if got != want {
		t.Errorf("Should have got %q delay from the master, not %s",
			want, got)
	}
}

func TestSetOption(t *testing.T) {
	// GIVEN a Shoutrrr with Options.delay
	shoutrrr := testGet()

	// WHEN SetOption is called with "delay" to override the previous value
	want := "new"
	shoutrrr.SetOption("delay", want)

	// THEN we get this value with we GetSelfOption
	got := shoutrrr.GetSelfOption("delay")
	if got != want {
		t.Errorf("Should have got %q delay from the main, not %s",
			want, got)
	}
}

func TestGetURLField(t *testing.T) {
	// GIVEN a Shoutrrr with the closest URLFields.token in Main
	shoutrrr := testGet()
	shoutrrr.URLFields = map[string]string{}

	// WHEN GetURLField is called with "token"
	got := shoutrrr.GetURLField("token")

	// THEN we get the token of the master Shoutrrr
	want := shoutrrr.Main.URLFields["token"]
	if got != want {
		t.Errorf("Should have got %q token from the main, not %s",
			want, got)
	}
}

func TestGetSelfURLField(t *testing.T) {
	// GIVEN a Shoutrrr with URLFields.token
	shoutrrr := testGet()

	// WHEN GetSelfURLField is called with "token"
	got := shoutrrr.GetSelfURLField("token")

	// THEN we get the token of the master Shoutrrr
	want := shoutrrr.URLFields["token"]
	if got != want {
		t.Errorf("Should have got %q token from the master, not %s",
			want, got)
	}
}

func TestSetURLField(t *testing.T) {
	// GIVEN a Shoutrrr with URLFields.token
	shoutrrr := testGet()

	// WHEN SetURLField is called with "token" to override the previous value
	want := "new"
	shoutrrr.SetURLField("token", want)

	// THEN we get this value with we GetSelfURLField
	got := shoutrrr.GetSelfURLField("token")
	if got != want {
		t.Errorf("Should have got %q token from the main, not %s",
			want, got)
	}
}

func TestGetParam(t *testing.T) {
	// GIVEN a Shoutrrr with the closest Params.avatar in Main
	shoutrrr := testGet()
	shoutrrr.Params = map[string]string{}

	// WHEN GetParam is called with "avatar"
	got := shoutrrr.GetParam("avatar")

	// THEN we get the avatar of the master Shoutrrr
	want := shoutrrr.Main.Params["avatar"]
	if got != want {
		t.Errorf("Should have got %q avatar from the main, not %s",
			want, got)
	}
}

func TestGetSelfParam(t *testing.T) {
	// GIVEN a Shoutrrr with Params.avatar
	shoutrrr := testGet()

	// WHEN GetSelfParam is called with "avatar"
	got := shoutrrr.GetSelfParam("avatar")

	// THEN we get the avatar of the master Shoutrrr
	want := shoutrrr.Params["avatar"]
	if got != want {
		t.Errorf("Should have got %q avatar from the master, not %s",
			want, got)
	}
}

func TestSetParam(t *testing.T) {
	// GIVEN a Shoutrrr with Params.avatar
	shoutrrr := testGet()

	// WHEN SetParam is called with "avatar" to override the previous value
	want := "new"
	shoutrrr.SetParam("avatar", want)

	// THEN we get this value with we GetSelfParam
	got := shoutrrr.GetSelfParam("avatar")
	if got != want {
		t.Errorf("Should have got %q avatar from the main, not %s",
			want, got)
	}
}
func TestGetDelay(t *testing.T) {
	// GIVEN a shoutrrr with no Options.delay
	shoutrrr := testGet()

	// WHEN GetDelay is called
	got := shoutrrr.GetDelay()

	// THEN the function returns the closest Options.delay as a time.Duration
	want := shoutrrr.GetSelfOption("delay")
	if got != want {
		t.Errorf("Want %s, got %s",
			want, got)
	}
}

func TestGetDelayWithNoDelaySet(t *testing.T) {
	// GIVEN a shoutrrr with no Options.delay
	shoutrrr := testGet()
	shoutrrr.Options = map[string]string{}
	shoutrrr.Main.Options = map[string]string{}
	shoutrrr.Defaults.Options = map[string]string{}
	shoutrrr.HardDefaults.Options = map[string]string{}

	// WHEN GetDelay is called
	got := shoutrrr.GetDelay()

	// THEN the function returns the closest Options.delay as a time.Duration
	want := "0s"
	if got != want {
		t.Errorf("Want %s, got %s",
			want, got)
	}
}

func TestGetDelayDuration(t *testing.T) {
	// GIVEN a shoutrrr with Options.delay
	shoutrrr := testGet()

	// WHEN GetDelayDuration is called
	got := shoutrrr.GetDelayDuration()

	// THEN the function returns the closest Options.delay as a time.Duration
	want, _ := time.ParseDuration(shoutrrr.GetOption("delay"))
	if got != want {
		t.Errorf("Want %s, got %s",
			want, got)
	}
}

func TestGetMaxTries(t *testing.T) {
	// GIVEN a shoutrrr with Options.max_tries
	shoutrrr := testGet()

	// WHEN GetDelayDuration is called
	got := shoutrrr.GetMaxTries()

	// THEN the function returns the closest Options.delay as a time.Duration
	want, _ := strconv.ParseUint(shoutrrr.GetSelfOption("max_tries"), 10, 32)
	if got != uint(want) {
		t.Errorf("Want %d, got %d",
			want, got)
	}
}

func TestGetMessage(t *testing.T) {
	// GIVEN a shoutrrr with Options.message and ServiceInfo context
	shoutrrr := testGet()
	context := utils.ServiceInfo{LatestVersion: "1.2.3"}

	// WHEN GetMessage is called
	got := shoutrrr.GetMessage(&context)

	// THEN the message is formatted with that context
	want := "1.2.3-foo"
	if got != want {
		t.Errorf("Not templated correctly. Want %q, got %q",
			want, got)
	}
}

func TestGetTitle(t *testing.T) {
	// GIVEN a shoutrrr with Params.Title and ServiceInfo context
	shoutrrr := testGet()
	context := utils.ServiceInfo{LatestVersion: "1.2.3"}

	// WHEN GetTitle is called
	got := shoutrrr.GetTitle(&context)

	// THEN the Title is formatted with that context
	want := "1.2.3-master"
	if got != want {
		t.Errorf("Not templated correctly. Want %q, got %q",
			want, got)
	}
}

func TestGetType(t *testing.T) {
	// GIVEN a shoutrrr with Type and ServiceInfo context
	shoutrrr := testGet()

	// WHEN GetType is called
	got := shoutrrr.GetType()

	// THEN the Type is formatted with that context
	want := shoutrrr.Type
	if got != want {
		t.Errorf("Want %q, got %q",
			want, got)
	}
}
