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

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/webhook"
)

func testVerify() Config {
	allOrder := []string{"test"}
	defaults := Defaults{}
	defaults.SetDefaults()
	notify := shoutrrr.Slice{
		"test": defaults.Notify["discord"],
	}
	webhook := webhook.Slice{
		"test": &defaults.WebHook,
	}
	serviceID := "test"
	serviceURL := "release-argus/argus"
	service := service.Slice{
		"test": &service.Service{
			ID:  &serviceID,
			URL: &serviceURL,
		},
	}
	return Config{
		All:      allOrder,
		Defaults: defaults,
		Notify:   notify,
		WebHook:  webhook,
		Service:  service,
	}
}

func TestConfigCheckValuesValid(t *testing.T) {
	// GIVEN a valid Config
	config := testVerify()

	// WHEN CheckValues is called on it
	config.CheckValues()

	// THEN no Fatal panic occured
}

func TestConfigCheckValuesWithInvalidDefaults(t *testing.T) {
	// GIVEN an invalid Defaults Config
	jLog = utils.NewJLog("WARN", false)
	config := testVerify()
	invalid := "0x"
	config.Defaults.Service.Interval = &invalid
	// Switch Fatal to panic and disable this panic.
	jLog.Testing = true
	defer func() { _ = recover() }()

	// WHEN CheckValues is called on it
	config.CheckValues()

	// THEN this call will crash the program
	t.Errorf("Should have panic'd because of Defaults.Service.Interval being invalid (%q)",
		invalid)
}

func TestConfigCheckValuesWithInvalidNotify(t *testing.T) {
	// GIVEN an invalid Defaults Config
	jLog = utils.NewJLog("WARN", false)
	config := testVerify()
	invalid := "0x"
	config.Notify["test"].SetOption("delay", invalid)
	// Switch Fatal to panic and disable this panic.
	jLog.Testing = true
	defer func() { _ = recover() }()

	// WHEN CheckValues is called on it
	config.CheckValues()

	// THEN this call will crash the program
	t.Errorf("Should have panic'd because of Notify.discord.options.delay being invalid (%q)",
		invalid)
}

func TestConfigCheckValuesWithInvalidWebHook(t *testing.T) {
	// GIVEN an invalid Defaults Config
	jLog = utils.NewJLog("WARN", false)
	config := testVerify()
	invalid := "0x"
	config.WebHook["test"].Delay = &invalid
	// Switch Fatal to panic and disable this panic.
	jLog.Testing = true
	defer func() { _ = recover() }()

	// WHEN CheckValues is called on it
	config.CheckValues()

	// THEN this call will crash the program
	t.Errorf("Should have panic'd because of WebHook.test.Delay being invalid (%q)",
		invalid)
}

func TestConfigCheckValuesWithInvalidService(t *testing.T) {
	// GIVEN an invalid Defaults Config
	jLog = utils.NewJLog("WARN", false)
	config := testVerify()
	invalid := "0x"
	config.Service["test"].Interval = &invalid
	// Switch Fatal to panic and disable this panic.
	jLog.Testing = true
	defer func() { _ = recover() }()

	// WHEN CheckValues is called on it
	config.CheckValues()

	// THEN this call will crash the program
	t.Errorf("Should have panic'd because of Service.test.Interval being invalid (%q)",
		invalid)
}

func TestConfigPrintWithFalseFlag(t *testing.T) {
	// GIVEN a Config and the print flag being false
	var flag bool = false
	config := testVerify()
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Print is called
	config.Print(&flag)

	// THEN nothing is printed
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	want := 0
	got := len(string(out))
	if got != want {
		t.Errorf("Print with no flag set printed %d lines. Wanted %d",
			got, want)
	}
}

func TestConfigPrint(t *testing.T) {
	// GIVEN a Config and the print flag being false
	jLog = utils.NewJLog("WARN", false)
	jLog.Testing = true
	var flag bool = true
	config := testVerify()
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Print is called
	config.Print(&flag)

	// THEN nothing is printed
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	want := 165
	got := strings.Count(string(out), "\n")
	if got != want {
		t.Errorf("Print with no flag set printed %d lines. Wanted %d\n%s",
			got, want, string(out))
	}
}
