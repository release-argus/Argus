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
	"strings"
	"testing"

	"github.com/release-argus/Argus/utils"
)

func TestSetLog(t *testing.T) {
	// GIVEN jLog is nil and we have a fresh one
	jLog = nil
	new := utils.NewJLog("DEBUG", false)

	// WHEN jLog is set with SetLog
	SetLog(new)

	// THEN jLog is set to this new one
	if jLog != new {
		t.Errorf("jLog should be %v but is %v", new, jLog)
	}
}

func TestInitMetricsOfNonMaster(t *testing.T) {
	// GIVEN a Shoutrrr with no Main
	shoutrrr := Shoutrrr{}

	// WHEN initMetrics is called on this
	shoutrrr.initMetrics(nil)

	// THEN the function doesn't create the metrics (would have crashed on *nil)
}

func TestInitMetrics(t *testing.T) {
	// GIVEN a Shoutrrr with a Main
	id := ""
	shoutrrr := Shoutrrr{ID: &id, Main: &Shoutrrr{}}

	// WHEN initMetrics is called on this
	shoutrrr.initMetrics(&id)

	// THEN the Prometheus metrics are created
}

func TestInitMapsWithNil(t *testing.T) {
	// GIVEN a nil Shoutrrr
	var shoutrrr *Shoutrrr

	// WHEN InitMaps is called on it
	shoutrrr.InitMaps()

	// THEN the maps will stay nil
	if shoutrrr != nil {
		t.Errorf("The Shoutrrr was nil, but is now %v", *shoutrrr)
	}
}

func TestInitOptions(t *testing.T) {
	// GIVEN a Shoutrrr with capital Options
	shoutrrr := Shoutrrr{
		Options: map[string]string{
			"fOO":   "1",
			"OTHER": "2",
		},
	}

	// WHEN InitOptions is called
	shoutrrr.initOptions()

	// THEN the Option keys are converted to lowercase
	for key := range shoutrrr.Options {
		want := strings.ToLower(key)
		if key != want {
			t.Errorf("Keys were not lowercased, got %s, want %s", key, want)
		}
	}
}

func TestInitURLFields(t *testing.T) {
	// GIVEN a Shoutrrr with capital URLFields
	shoutrrr := Shoutrrr{
		URLFields: map[string]string{
			"fOO":   "1",
			"OTHER": "2",
		},
	}

	// WHEN InitURLFields is called
	shoutrrr.initURLFields()

	// THEN the Option keys are converted to lowercase
	for key := range shoutrrr.URLFields {
		want := strings.ToLower(key)
		if key != want {
			t.Errorf("Keys were not lowercased, got %s, want %s", key, want)
		}
	}
}

func TestInitParams(t *testing.T) {
	// GIVEN a Shoutrrr with capital Params
	shoutrrr := Shoutrrr{
		Params: map[string]string{
			"fOO":   "1",
			"OTHER": "2",
		},
	}

	// WHEN InitParams is called
	shoutrrr.initParams()

	// THEN the Option keys are converted to lowercase
	for key := range shoutrrr.Params {
		want := strings.ToLower(key)
		if key != want {
			t.Errorf("Keys were not lowercased, got %s, want %s", key, want)
		}
	}
}

func TestInitMapsUppercasesOptions(t *testing.T) {
	// GIVEN a nil Shoutrrr
	shoutrrr := Shoutrrr{
		Options: map[string]string{
			"ofOO":   "1",
			"oOTHER": "2",
		},
		URLFields: map[string]string{
			"UfOO":   "1",
			"uOTHER": "2",
		},
		Params: map[string]string{
			"pfOO":   "1",
			"POTHER": "2",
		},
	}

	// WHEN InitMaps is called on it
	shoutrrr.InitMaps()

	// THEN the keys of all maps will be lowercased
	for key := range shoutrrr.Options {
		want := strings.ToLower(key)
		if key != want {
			t.Errorf("Keys were not lowercased, got %s, want %s", key, want)
		}
	}
}

func TestInitMapsUppercasesURLFields(t *testing.T) {
	// GIVEN a nil Shoutrrr
	shoutrrr := Shoutrrr{
		Options: map[string]string{
			"ofOO":   "1",
			"oOTHER": "2",
		},
		URLFields: map[string]string{
			"UfOO":   "1",
			"uOTHER": "2",
		},
		Params: map[string]string{
			"pfOO":   "1",
			"POTHER": "2",
		},
	}

	// WHEN InitMaps is called on it
	shoutrrr.InitMaps()

	// THEN the keys of all maps will be lowercased
	for key := range shoutrrr.URLFields {
		want := strings.ToLower(key)
		if key != want {
			t.Errorf("Keys were not lowercased, got %s, want %s", key, want)
		}
	}
}

func TestInitMapsUppercasesParams(t *testing.T) {
	// GIVEN a nil Shoutrrr
	shoutrrr := Shoutrrr{
		Options: map[string]string{
			"ofOO":   "1",
			"oOTHER": "2",
		},
		URLFields: map[string]string{
			"UfOO":   "1",
			"uOTHER": "2",
		},
		Params: map[string]string{
			"pfOO":   "1",
			"POTHER": "2",
		},
	}

	// WHEN InitMaps is called on it
	shoutrrr.InitMaps()

	// THEN the keys of all maps will be lowercased
	for key := range shoutrrr.Params {
		want := strings.ToLower(key)
		if key != want {
			t.Errorf("Keys were not lowercased, got %s, want %s", key, want)
		}
	}
}

func TestShoutrrrInitWithNilShoutrrr(t *testing.T) {
	// GIVEN a nil Shoutrrr and an ID string
	var shoutrrr *Shoutrrr
	id := "test"

	// WHEN Init is called on it
	shoutrrr.Init(&id, &Shoutrrr{}, &Shoutrrr{}, &Shoutrrr{})

	// THEN it is still nil
	if shoutrrr != nil {
		t.Errorf("Shoutrrr should still be nil, not %v after Init", shoutrrr)
	}
}

func TestShoutrrrInitWithNilMain(t *testing.T) {
	// GIVEN a nil Shoutrrr and an ID string
	// vars := testGet()
	var shoutrrr Shoutrrr
	id := "test"

	// WHEN Init is called on it with nil main
	shoutrrr.Init(&id, nil, &Shoutrrr{}, &Shoutrrr{})

	// THEN it is no longer nil
	if shoutrrr.Main == nil {
		t.Errorf("Shoutrrr.Main shouldn't still be %v after Init", shoutrrr)
	}
}

func TestShoutrrrInitGivesMain(t *testing.T) {
	// GIVEN a Shoutrrr with a nil Main
	vars := testGet()
	var shoutrrr Shoutrrr
	id := "test"

	// WHEN Init is called on it with a nil Main
	shoutrrr.Init(&id, vars.Main, &Shoutrrr{}, &Shoutrrr{})

	// THEN it is given vars.Main
	if shoutrrr.Main != vars.Main {
		t.Errorf("Shoutrrr.Main shouldn't be %v after Init", shoutrrr)
	}
}

func TestShoutrrrInitGivesDefaults(t *testing.T) {
	// GIVEN a Shoutrrr with a nil Defaults
	vars := testGet()
	var shoutrrr Shoutrrr
	id := "test"

	// WHEN Init is called on it with a non-nil Defaults
	shoutrrr.Init(&id, &Shoutrrr{}, vars.Defaults, &Shoutrrr{})

	// THEN it is given vars.Defaults
	if shoutrrr.Defaults != vars.Defaults {
		t.Errorf("Shoutrrr.Defaults shouldn't be %v after Init", shoutrrr)
	}
}

func TestShoutrrrInitGivesHardDefaults(t *testing.T) {
	// GIVEN a Shoutrrr with a nil HardDefaults
	vars := testGet()
	var shoutrrr Shoutrrr
	id := "test"

	// WHEN Init is called on it with a non-nil HardDefaults
	shoutrrr.Init(&id, &shoutrrr, &Shoutrrr{}, vars.HardDefaults)

	// THEN it is given vars.HardDefaults
	if shoutrrr.HardDefaults != vars.HardDefaults {
		t.Errorf("Shoutrrr.HardDefaults shouldn't be %v after Init", shoutrrr)
	}
}

func testSlice() Slice {
	id0 := "zero"
	id1 := "one"
	id2 := "two"
	return Slice{
		"0": &Shoutrrr{
			ID:   &id0,
			Type: "discord",
		},
		"1": &Shoutrrr{
			ID:   &id1,
			Type: "telegram",
		},
		"2": &Shoutrrr{
			ID:   &id2,
			Type: "telegram",
		},
	}
}

func TestSliceInitWithNil(t *testing.T) {
	// GIVEN a nil Slice
	var slice *Slice

	// WHEN Init is called on it
	slice.Init(nil, nil, nil, nil, nil)

	// THEN it is still nil
	if slice != nil {
		t.Errorf("Slice should still be nil, not %v", slice)
	}
}

func TestSliceInitWithDefaultMain(t *testing.T) {
	// GIVEN a Slice and a Slice with a Main for "1"
	slice := testSlice()
	serviceID := "test"

	// WHEN Init is called on it with a main for "1"
	slice.Init(nil, &serviceID, &Slice{}, &Slice{}, &Slice{})

	// THEN "1" has an empty main
	if slice["1"].Main.Type != "" ||
		len(slice["1"].Main.Options) != 0 ||
		len(slice["1"].Main.URLFields) != 0 ||
		len(slice["1"].Main.Params) != 0 {
		t.Errorf("Slice['1'] was given %v, not an empty Main (%v)", slice["1"].Main, Shoutrrr{})
	}
}

func TestSliceInitWithMatchingMain(t *testing.T) {
	// GIVEN a Slice and a Slice with a Main for "1"
	slice := testSlice()
	mainID := "main"
	serviceID := "test"
	mains := Slice{
		"1": &Shoutrrr{
			ID: &mainID,
		},
	}

	// WHEN Init is called on it with a main for "1"
	slice.Init(nil, &serviceID, &mains, &Slice{}, &Slice{})

	// THEN "1" has that main
	if slice["1"].Main != mains["1"] {
		t.Errorf("Slice['1'] was given %v, not the matching Main (%v)", slice["1"].Main, mains["1"])
	}
}

func TestSliceInitWithNonMatchingMain(t *testing.T) {
	// GIVEN a Slice and a Slice with a Main for "1"
	slice := testSlice()
	mainID := "main"
	serviceID := "test"
	mains := Slice{
		"1": &Shoutrrr{
			ID: &mainID,
		},
	}

	// WHEN Init is called on it with a main for "1"
	slice.Init(nil, &serviceID, &mains, &Slice{}, &Slice{})

	// THEN "1" has that main
	if slice["0"].Main.ID != nil {
		t.Errorf("Slice['0'] was given id=%q, not a fresh Main (id=%v)", *slice["0"].Main.ID, Shoutrrr{}.ID)
	}
}

func TestSliceInitWithNilElement(t *testing.T) {
	// GIVEN a Slice with a nil Shoutrrr element
	slice := testSlice()
	serviceID := "test"
	slice["1"] = nil

	// WHEN Init is called on it
	slice.Init(nil, &serviceID, &Slice{}, &Slice{}, &Slice{})

	// THEN that element is no longer nil
	if slice["1"] == nil {
		t.Errorf("Slice['1'] should have been created from the Init. Shouldn't be %v", slice["1"])
	}
}

func TestSliceInitWithNilMains(t *testing.T) {
	// GIVEN a Slice with nil Shoutrrr mains
	slice := testSlice()
	serviceID := "test"
	slice["1"] = nil

	// WHEN Init is called on it
	slice.Init(nil, &serviceID, nil, &Slice{}, &Slice{})

	// THEN the elements get non-nil mains
	if slice["1"].Main == nil {
		t.Errorf("Mains were nil in the Init, but non-nil's should. Shouldn't be %v", slice["1"])
	}
}

func TestSliceInitWithDefaults(t *testing.T) {
	// GIVEN a Slice and defaults for "discord"
	slice := testSlice()
	slice["1"] = nil
	serviceID := "test"
	defaults := Slice{
		"discord": &Shoutrrr{
			Params: map[string]string{
				"title": "something",
			},
		},
	}

	// WHEN Init is called on it with nil defaults for the "discord" Shoutrrr type
	slice.Init(nil, &serviceID, &Slice{}, &defaults, &Slice{})

	// THEN all "discord" types have the default
	for i := range slice {
		if slice[i].Type == "discord" && slice[i].Defaults != defaults["discord"] {
			t.Errorf("Slice['%s'] should have been given the %s defaults, %v", i, slice[i].Type, defaults["discord"])
		}
	}
}

func TestSliceInitWithNoDefaults(t *testing.T) {
	// GIVEN a Slice and defaults for "discord"
	slice := testSlice()
	slice["1"] = nil
	serviceID := "test"
	defaults := Slice{
		"discord": &Shoutrrr{
			Params: map[string]string{
				"title": "something",
			},
		},
	}

	// WHEN Init is called on it with nil defaults for the "discord" Shoutrrr type
	slice.Init(nil, &serviceID, &Slice{}, &defaults, &Slice{})

	// THEN no non "discord" types are given this default
	for i := range slice {
		if slice[i].Type != "discord" && slice[i].Defaults == defaults["discord"] {
			t.Errorf("Slice['%s'] shouldn't have been given the discord defaults, %v", i, defaults["discord"])
		}
	}
}

func TestSliceInitWithHardDefaults(t *testing.T) {
	// GIVEN a Slice and hardDefaults for "discord"
	slice := testSlice()
	slice["1"] = nil
	serviceID := "test"
	hardDefaults := Slice{
		"discord": &Shoutrrr{
			Params: map[string]string{
				"title": "something",
			},
		},
	}

	// WHEN Init is called on it with nil hardDefaults for the "discord" Shoutrrr type
	slice.Init(nil, &serviceID, &Slice{}, &Slice{}, &hardDefaults)

	// THEN all "discord" types have the default
	for i := range slice {
		if slice[i].Type == "discord" && slice[i].HardDefaults != hardDefaults["discord"] {
			t.Errorf("Slice['%s'] should have been given the %s hardDefaults, %v", i, slice[i].Type, hardDefaults["discord"])
		}
	}
}

func TestSliceInitWithNoHardDefaults(t *testing.T) {
	// GIVEN a Slice and hardDefaults for "discord"
	slice := testSlice()
	slice["1"] = nil
	serviceID := "test"
	hardDefaults := Slice{
		"discord": &Shoutrrr{
			Params: map[string]string{
				"title": "something",
			},
		},
	}

	// WHEN Init is called on it with nil hardDefaults for the "discord" Shoutrrr type
	slice.Init(nil, &serviceID, &Slice{}, &Slice{}, &hardDefaults)

	// THEN no non "discord" types are given this default
	for i := range slice {
		if slice[i].Type != "discord" && slice[i].HardDefaults == hardDefaults["discord"] {
			t.Errorf("Slice['%s'] shouldn't have been given the discord hardDefaults, %v", i, hardDefaults["discord"])
		}
	}
}
