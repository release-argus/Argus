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

func TestNotifyTestWithNoShoutrrr(t *testing.T) {
	// GIVEN a Config with a Service containing a Shoutrrr
	jLog = utils.NewJLog("WARN", false)
	InitJLog(jLog)
	var (
		sID       string = "test"
		nMaxTries string = "1"
	)
	notifier := shoutrrr.Shoutrrr{
		ID: &sID,
		Options: map[string]string{
			"max_tries": nMaxTries,
		},
	}
	cfg := config.Config{
		Service: service.Slice{
			sID: &service.Service{
				Notify: &shoutrrr.Slice{
					sID: &notifier,
				},
			},
		},
	}
	flag := ""
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN NotifyTest is called with an empty (undefined) flag
	NotifyTest(&flag, &cfg)

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

func TestNotifyTestWithUnknownService(t *testing.T) {
	// GIVEN a Config with a Service containing a Shoutrrr
	jLog = utils.NewJLog("WARN", false)
	InitJLog(jLog)
	var (
		sID       string = "test"
		nMaxTries string = "1"
	)
	notifier := shoutrrr.Shoutrrr{
		ID: &sID,
		Options: map[string]string{
			"max_tries": nMaxTries,
		},
	}
	cfg := config.Config{
		Service: service.Slice{
			sID: &service.Service{
				Notify: &shoutrrr.Slice{
					sID: &notifier,
				},
			},
		},
		Notify: shoutrrr.Slice{
			"discord": &shoutrrr.Shoutrrr{},
		},
	}
	flag := "other_" + sID
	// Switch Fatal to panic and disable this panic.
	jLog.Testing = true
	defer func() {
		r := recover()
		if !strings.Contains(r.(string), " could not be found ") {
			t.Error(r)
		}
	}()

	// WHEN NotifyTest is called with a Shoutrrr not in the config
	NotifyTest(&flag, &cfg)

	// THEN it will be printed that the command couldn't be found
	t.Error("Should os.Exit(1), err")
}

func TestNotifyTestWithKnownShoutrrr(t *testing.T) {
	// GIVEN a Config with a Service containing a Shoutrrr
	jLog = utils.NewJLog("WARN", false)
	InitJLog(jLog)
	var (
		sID       string = "test"
		nMaxTries string = "1"
	)
	notifier := shoutrrr.Shoutrrr{
		ID:   &sID,
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
			sID: &service.Service{
				Notify: &shoutrrr.Slice{
					sID: &notifier,
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

	// WHEN NotifyTest is called with a Shoutrrr not in the config
	NotifyTest(&flag, &cfg)

	// THEN it will be printed that the command couldn't be found
	t.Errorf("Should os.Exit(0), err")
}

func TestNotifyTestWithKnownShoutrrrMain(t *testing.T) {
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
	NotifyTest(&flag, &cfg)

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
	NotifyTest(&flag, &cfg)

	// THEN it will be printed that the command couldn't be found
	t.Errorf("Should os.Exit(0), err")
}
