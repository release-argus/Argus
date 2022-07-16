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

package testing

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/utils"
)

func TestGetAllShoutrrrNamesWithBothNotifyAndServiceNotify(t *testing.T) {
	// GIVEN a Config with a Service containing a Shoutrrr and a Main Shoutrrr
	jLog = utils.NewJLog("WARN", false)
	InitJLog(jLog)
	var (
		nID string = "test"
	)
	notifier := shoutrrr.Shoutrrr{
		ID: &nID,
	}
	cfg := config.Config{
		Service: service.Slice{
			nID: &service.Service{
				Notify: &shoutrrr.Slice{
					nID: &notifier,
				},
			},
		},
		Notify: shoutrrr.Slice{
			"discord": &shoutrrr.Shoutrrr{},
		},
	}

	// WHEN getAllShoutrrrNames is called on this config
	got := getAllShoutrrrNames(&cfg)

	// THEN a list of all shoutrrr's will be returned
	want := []string{"discord", nID}
	fail := false
	if len(got) != len(want) {
		fail = true
	} else {
		for i := range got {
			if got[i] != want[i] {
				fail = true
				break
			}
		}
	}
	if fail {
		t.Errorf("Expected a list of all notifiers\ngot:  %v\nwant: %v",
			got, want)
	}
}

func TestGetAllShoutrrrNamesWithOnlyNotify(t *testing.T) {
	// GIVEN a Config with a Service containing a Main Shoutrrr
	jLog = utils.NewJLog("WARN", false)
	InitJLog(jLog)
	cfg := config.Config{
		Notify: shoutrrr.Slice{
			"discord": &shoutrrr.Shoutrrr{},
		},
	}

	// WHEN getAllShoutrrrNames is called on this config
	got := getAllShoutrrrNames(&cfg)

	// THEN a list of all shoutrrr's will be returned
	want := []string{"discord"}
	fail := false
	if len(got) != len(want) {
		fail = true
	} else {
		for i := range got {
			if got[i] != want[i] {
				fail = true
				break
			}
		}
	}
	if fail {
		t.Errorf("Expected a list of all notifiers\ngot:  %v\nwant: %v",
			got, want)
	}
}

func TestGetAllShoutrrrNamesWithOnlyServices(t *testing.T) {
	// GIVEN a Config with a Service containing a Service Shoutrrr
	jLog = utils.NewJLog("WARN", false)
	InitJLog(jLog)
	var (
		nID0 string = "test"
		nID1 string = "other"
	)
	notifier0 := shoutrrr.Shoutrrr{
		ID: &nID0,
	}
	notifier1 := shoutrrr.Shoutrrr{
		ID: &nID1,
	}
	cfg := config.Config{
		Service: service.Slice{
			nID0: &service.Service{
				Notify: &shoutrrr.Slice{
					nID0: &notifier0,
				},
			},
			nID1: &service.Service{
				Notify: &shoutrrr.Slice{
					nID1: &notifier1,
				},
			},
		},
	}

	// WHEN getAllShoutrrrNames is called on this config
	got := getAllShoutrrrNames(&cfg)

	// THEN a list of all shoutrrr's will be returned
	want := []string{nID0, nID1}
	fail := false
	if len(got) != len(want) {
		fail = true
	} else {
		for i := range got {
			if got[i] != want[i] {
				fail = true
				break
			}
		}
	}
	if fail {
		t.Errorf("Expected a list of all notifiers\ngot:  %v\nwant: %v",
			got, want)
	}
}

func TestGetAllShoutrrrNamesWithDuplicates(t *testing.T) {
	// GIVEN a Config with a Service containing a Service Shoutrrr
	jLog = utils.NewJLog("WARN", false)
	InitJLog(jLog)
	var (
		nID0 string = "test"
		nID1 string = "other"
	)
	notifier0 := shoutrrr.Shoutrrr{
		ID: &nID0,
	}
	notifier1 := shoutrrr.Shoutrrr{
		ID: &nID1,
	}
	cfg := config.Config{
		Service: service.Slice{
			nID0: &service.Service{
				Notify: &shoutrrr.Slice{
					nID0: &notifier0,
				},
			},
			nID1: &service.Service{
				Notify: &shoutrrr.Slice{
					nID0: &notifier0,
					nID1: &notifier1,
				},
			},
		},
		Notify: shoutrrr.Slice{
			nID0: &shoutrrr.Shoutrrr{},
		},
	}

	// WHEN getAllShoutrrrNames is called on this config
	got := getAllShoutrrrNames(&cfg)

	// THEN a list of all shoutrrr's will be returned
	want := []string{nID0, nID1}
	fail := false
	if len(got) != len(want) {
		fail = true
	} else {
		for i := range got {
			if got[i] != want[i] {
				fail = true
				break
			}
		}
	}
	if fail {
		t.Errorf("Expected a list of all notifiers\ngot:  %v\nwant: %v",
			got, want)
	}
}

func TestNotifyTestWithNoShoutrrrFlag(t *testing.T) {
	// GIVEN a Config with a Service containing a Shoutrrr
	jLog = utils.NewJLog("WARN", false)
	InitJLog(jLog)
	var (
		nID       string = "test"
		nMaxTries string = "1"
	)
	notifier := shoutrrr.Shoutrrr{
		ID: &nID,
		Options: map[string]string{
			"max_tries": nMaxTries,
		},
	}
	cfg := config.Config{
		Service: service.Slice{
			nID: &service.Service{
				Notify: &shoutrrr.Slice{
					nID: &notifier,
				},
			},
		},
	}
	flag := ""
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN NotifyTest is called with an empty (undefined) flag
	NotifyTest(&flag, &cfg, jLog)

	// THEN nothing will be run/printed
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	output := string(out)
	want := ""
	if want != output {
		t.Errorf("NotifyTest with %q flag shouldn't print anything, got\n%s",
			flag, output)
	}
}

func TestFindShoutrrrWithUnknownShoutrrr(t *testing.T) {
	// GIVEN a Config with a Service containing a Shoutrrr and a Main Shoutrrr
	jLog = utils.NewJLog("WARN", false)
	InitJLog(jLog)
	var (
		nID       string = "test"
		nMaxTries string = "1"
	)
	notifier := shoutrrr.Shoutrrr{
		ID: &nID,
		Options: map[string]string{
			"max_tries": nMaxTries,
		},
	}
	cfg := config.Config{
		Service: service.Slice{
			nID: &service.Service{
				Notify: &shoutrrr.Slice{
					nID: &notifier,
				},
			},
		},
		Notify: shoutrrr.Slice{
			"discord": &shoutrrr.Shoutrrr{},
		},
	}
	flag := "other_" + nID
	// Switch Fatal to panic and disable this panic.
	jLog.Testing = true
	defer func() {
		r := recover()
		if !strings.Contains(r.(string), " could not be found ") {
			t.Error(r)
		}
	}()

	// WHEN findShoutrrr is called for a Shoutrrr not in the config
	got := findShoutrrr(flag, &cfg, jLog, utils.LogFrom{})

	// THEN it will be printed that the command couldn't be found
	t.Errorf("Should os.Exit(1) from %q not being in %v, err got %v",
		flag, cfg, got)
}

func TestNotifyTestWithUnknownServiceShoutrrr(t *testing.T) {
	// GIVEN a Config with a Service containing a Shoutrrr
	jLog = utils.NewJLog("WARN", false)
	InitJLog(jLog)
	var (
		nID       string = "test"
		nMaxTries string = "1"
	)
	notifier := shoutrrr.Shoutrrr{
		ID: &nID,
		Options: map[string]string{
			"max_tries": nMaxTries,
		},
	}
	cfg := config.Config{
		Service: service.Slice{
			nID: &service.Service{
				Notify: &shoutrrr.Slice{
					nID: &notifier,
				},
			},
		},
		Notify: shoutrrr.Slice{
			"discord": &shoutrrr.Shoutrrr{},
		},
	}
	flag := "other_" + nID
	// Switch Fatal to panic and disable this panic.
	jLog.Testing = true
	defer func() {
		r := recover()
		if !strings.Contains(r.(string), " could not be found ") {
			t.Error(r)
		}
	}()

	// WHEN NotifyTest is called with a Shoutrrr not in the config
	NotifyTest(&flag, &cfg, jLog)

	// THEN it will be printed that the command couldn't be found
	t.Error("Should os.Exit(1), err")
}

func TestFindShoutrrrWithKnownServiceShoutrrr(t *testing.T) {
	// GIVEN a Config with a Service containing a Shoutrrr
	jLog = utils.NewJLog("WARN", false)
	InitJLog(jLog)
	var (
		nID       string = "test"
		nMaxTries string = "1"
	)
	notifier := shoutrrr.Shoutrrr{
		ID:   &nID,
		Type: "something",
		Options: map[string]string{
			"max_tries": nMaxTries,
		},
		Params: map[string]string{},
		Main: &shoutrrr.Shoutrrr{
			Params: map[string]string{},
		},
		Defaults: &shoutrrr.Shoutrrr{
			Params: map[string]string{},
		},
		HardDefaults: &shoutrrr.Shoutrrr{
			Params: map[string]string{},
		},
	}
	cfg := config.Config{
		Service: service.Slice{
			nID: &service.Service{
				Notify: &shoutrrr.Slice{
					nID: &notifier,
				},
			},
		},
	}
	flag := "test"

	// WHEN NotifyTest is called for a Shoutrrr in the config
	got := findShoutrrr(flag, &cfg, jLog, utils.LogFrom{})

	// THEN it will have returned the correct Shoutrrr
	want := (*cfg.Service[nID].Notify)[flag]
	if got["test"] != want {
		t.Errorf("Expected the %q Notify\ngot:  %v\nwant: %v",
			flag, got, want)
	}
}

func TestNotifyTestWithKnownServiceShoutrrr(t *testing.T) {
	// GIVEN a Config with a Service containing a Shoutrrr
	jLog = utils.NewJLog("WARN", false)
	InitJLog(jLog)
	var (
		nID       string = "test"
		nMaxTries string = "1"
	)
	notifier := shoutrrr.Shoutrrr{
		ID:   &nID,
		Type: "something",
		Options: map[string]string{
			"max_tries": nMaxTries,
		},
		Params: map[string]string{},
		Main: &shoutrrr.Shoutrrr{
			Params: map[string]string{},
		},
		Defaults: &shoutrrr.Shoutrrr{
			Params: map[string]string{},
		},
		HardDefaults: &shoutrrr.Shoutrrr{
			Params: map[string]string{},
		},
	}
	cfg := config.Config{
		Service: service.Slice{
			nID: &service.Service{
				Notify: &shoutrrr.Slice{
					nID: &notifier,
				},
			},
		},
	}
	flag := "test"
	// Switch Fatal to panic and disable this panic.
	jLog.Testing = true
	defer func() {
		r := recover()
		if !strings.Contains(r.(string), " failed to send ") {
			t.Error(r)
		}
	}()

	// WHEN NotifyTest is called with a Shoutrrr in the config
	NotifyTest(&flag, &cfg, jLog)

	// THEN it will be printed that the command couldn't be found
	t.Errorf("Should os.Exit(0), err")
}

func TestFindShoutrrrWithKnownNotifyShoutrrr(t *testing.T) {
	// GIVEN a Config with a Main containing a Shoutrrr
	jLog = utils.NewJLog("WARN", false)
	InitJLog(jLog)
	var (
		nID                   string = "test"
		nType                 string = "smtp"
		optionsMaxTries       string = "1"
		urlFieldHost          string = "smtp.example.com"
		paramFieldFromAddress string = "test@release-argus.io"
		paramFieldtoAddresses string = "someone@you.com"
	)
	notifier := shoutrrr.Shoutrrr{
		ID:   &nID,
		Type: nType,
		Options: map[string]string{
			"max_tries": optionsMaxTries,
		},
		URLFields: map[string]string{
			"host": urlFieldHost,
		},
		Params: map[string]string{
			"fromaddress": paramFieldFromAddress,
			"toaddresses": paramFieldtoAddresses,
		},
	}
	cfg := config.Config{
		Notify: shoutrrr.Slice{
			nID: &notifier,
		},
		Defaults: config.Defaults{
			Notify: shoutrrr.Slice{
				nType: &shoutrrr.Shoutrrr{},
			},
		},
	}
	flag := "test"

	// WHEN NotifyTest is called for a Shoutrrr in the config
	got := findShoutrrr(flag, &cfg, jLog, utils.LogFrom{})

	// THEN it will have returned the correct Shoutrrr
	want := cfg.Notify[flag]
	if got["test"] != want {
		t.Errorf("Expected the %q Notify\ngot:  %v\nwant: %v",
			flag, got["test"], want)
	}
	if len(got["test"].Options) != 1 ||
		got["test"].Options["max_tries"] != optionsMaxTries {
		t.Errorf("options:\ngot:  %v\nwant: %v",
			got["test"], want)
	}
	if len(got["test"].URLFields) != 1 ||
		got["test"].URLFields["host"] != urlFieldHost {
		t.Errorf("url_fields:\ngot:  %v\nwant: %v",
			got["test"].Options, want.Options)
	}
	if len(got["test"].Params) != 2 ||
		got["test"].Params["fromaddress"] != paramFieldFromAddress ||
		got["test"].Params["toaddresses"] != paramFieldtoAddresses {
		t.Errorf("params:\ngot:  %v\nwant: %v",
			got["test"].Params, want.Params)
	}
}

func TestNotifyTestWithKnownNotifyShoutrrr(t *testing.T) {
	// GIVEN a Config with a Main containing a Shoutrrr
	jLog = utils.NewJLog("WARN", false)
	InitJLog(jLog)
	var (
		nID       string = "test"
		nMaxTries string = "1"
		nType     string = "smtp"
	)
	notifier := shoutrrr.Shoutrrr{
		ID:   &nID,
		Type: nType,
		Options: map[string]string{
			"max_tries": nMaxTries,
		},
		URLFields: map[string]string{
			"host": "smtp.example.com",
		},
		Params: map[string]string{
			"fromaddress": "test@release-argus.io",
			"toaddresses": "someone@you.com",
		},
		Main: &shoutrrr.Shoutrrr{
			Params: map[string]string{},
		},
		Defaults: &shoutrrr.Shoutrrr{
			Params: map[string]string{},
		},
		HardDefaults: &shoutrrr.Shoutrrr{
			Params: map[string]string{},
		},
	}
	cfg := config.Config{
		Notify: shoutrrr.Slice{
			nID: &notifier,
		},
		Defaults: config.Defaults{
			Notify: shoutrrr.Slice{
				nType: &shoutrrr.Shoutrrr{},
			},
		},
	}
	flag := "test"
	// Switch Fatal to panic and disable this panic.
	jLog.Testing = true
	defer func() {
		r := recover()
		if !strings.Contains(r.(string), " failed to send ") {
			t.Error(r)
		}
	}()

	// WHEN NotifyTest is called with a Shoutrrr not in the config
	NotifyTest(&flag, &cfg, jLog)

	// THEN it will be printed that the command couldn't be found
	t.Errorf("Should os.Exit(0), err")
}

func TestNotifyTestWithKnownInvalidShoutrrrMain(t *testing.T) {
	// GIVEN a Config with a Main containing an invalid Shoutrrr
	jLog = utils.NewJLog("WARN", false)
	InitJLog(jLog)
	var (
		nID       string = "test"
		nMaxTries string = "1"
		nType     string = "smtp"
	)
	notifier := shoutrrr.Shoutrrr{
		ID:   &nID,
		Type: nType,
		Options: map[string]string{
			"max_tries": nMaxTries,
		},
		URLFields: map[string]string{},
		Params:    map[string]string{},
		Main: &shoutrrr.Shoutrrr{
			Params: map[string]string{},
		},
		Defaults: &shoutrrr.Shoutrrr{
			Params: map[string]string{},
		},
		HardDefaults: &shoutrrr.Shoutrrr{
			Params: map[string]string{},
		},
	}
	cfg := config.Config{
		Notify: shoutrrr.Slice{
			nID: &notifier,
		},
		Defaults: config.Defaults{
			Notify: shoutrrr.Slice{
				nType: &shoutrrr.Shoutrrr{},
			},
		},
	}
	flag := "test"
	// Switch Fatal to panic and disable this panic.
	jLog.Testing = true
	defer func() {
		r := recover()
		if !strings.Contains(r.(string), "<required>") {
			t.Error(r)
		}
	}()

	// WHEN NotifyTest is called with a Shoutrrr not in the config
	NotifyTest(&flag, &cfg, jLog)

	// THEN it will be printed that the command couldn't be found
	t.Errorf("Should os.Exit(0), err")
}
