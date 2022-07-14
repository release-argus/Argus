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
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/release-argus/Argus/utils"
)

func TestSliceCheckValuesOfNil(t *testing.T) {
	// GIVEN a nil Slice
	var slice *Slice

	// WHEN CheckValues is called on this Slice
	err := slice.CheckValues("")

	// THEN err is nil
	if err != nil {
		t.Errorf("CheckValues on %v should have been nil, not %q",
			slice, err.Error())
	}
}

func testShoutrrr() Shoutrrr {
	id := "test"
	return Shoutrrr{
		ID:   &id,
		Type: "discord",
		URLFields: map[string]string{
			"token":     "token",
			"webhookid": "webhookid",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}
}

func TestShoutrrrCheckValues(t *testing.T) {
	// GIVEN a Slice containing a valid Shoutrrr
	shoutrrr := testShoutrrr()
	slice := Slice{"test": &shoutrrr}

	// WHEN CheckValues is called on the Slice
	err := slice.CheckValues("")

	// THEN err will be nil
	if err != nil {
		t.Errorf("%v is valid, CheckValues shouldn't err - %v",
			shoutrrr, err.Error())
	}
}

func TestSliceCheckValuesOfInvalid(t *testing.T) {
	// GIVEN a Slice containing an invalid Shoutrrr
	shoutrrr := testShoutrrr()
	shoutrrr.URLFields = map[string]string{}
	slice := Slice{"test": &shoutrrr}

	// WHEN CheckValues is called on the Slice
	err := slice.CheckValues("")

	// THEN err will be non-nil
	if err == nil {
		t.Errorf("%v is invalid, CheckValues should err not %v",
			shoutrrr, err.Error())
	}
}

func TestCorrectSelfPort(t *testing.T) {
	// GIVEN a Shoutrrr with URLFields.port
	shoutrrr := Shoutrrr{
		URLFields: map[string]string{
			"port": ":123",
		},
	}

	// WHEN correctSelf is called on it
	shoutrrr.correctSelf()

	// THEN leading : is trimmed from port
	got := shoutrrr.GetSelfURLField("port")
	want := "123"
	if got != want {
		t.Errorf("Got %s, want %s",
			got, want)
	}
}

func TestCorrectSelfPath(t *testing.T) {
	// GIVEN a Shoutrrr with URLFields.path
	shoutrrr := Shoutrrr{
		URLFields: map[string]string{
			"path": "/test",
		},
	}

	// WHEN correctSelf is called on it
	shoutrrr.correctSelf()

	// THEN the leading / is trimmed from path
	got := shoutrrr.GetSelfURLField("path")
	want := "test"
	if got != want {
		t.Errorf("Got %s, want %s",
			got, want)
	}
}

func TestCorrectSelfMattermostChannel(t *testing.T) {
	// GIVEN a Mattermost Shoutrrr with channel
	shoutrrr := Shoutrrr{
		Type: "mattermost",
		URLFields: map[string]string{
			"channel": "/test",
		},
	}

	// WHEN correctSelf is called on it
	shoutrrr.correctSelf()

	// THEN leading slashes in channel are removed
	got := shoutrrr.GetSelfURLField("channel")
	want := "test"
	if got != want {
		t.Errorf("Got %s, want %s",
			got, want)
	}
}

func TestCorrectSelfSlackColor(t *testing.T) {
	// GIVEN a Slack Shoutrrr with color
	shoutrrr := Shoutrrr{
		Type: "slack",
		Params: map[string]string{
			"color": "#ffffff",
		},
	}

	// WHEN correctSelf is called on it
	shoutrrr.correctSelf()

	// THEN # in color is replaced by %23
	got := shoutrrr.GetSelfParam("color")
	want := "%23ffffff"
	if got != want {
		t.Errorf("Got %s, want %s",
			got, want)
	}
}

func TestCorrectSelfTeamsAltID(t *testing.T) {
	// GIVEN a teams Shoutrrr with altid
	shoutrrr := Shoutrrr{
		Type: "teams",
		URLFields: map[string]string{
			"altid": "/foo",
		},
	}

	// WHEN correctSelf is called on it
	shoutrrr.correctSelf()

	// THEN the leading / is trimmed from altid
	got := shoutrrr.GetSelfURLField("altid")
	want := "foo"
	if got != want {
		t.Errorf("Got %s, want %s",
			got, want)
	}
}

func TestCorrectSelfTeamsGroupOwner(t *testing.T) {
	// GIVEN a teams Shoutrrr with groupowner
	shoutrrr := Shoutrrr{
		Type: "teams",
		URLFields: map[string]string{
			"groupowner": "/bar",
		},
	}

	// WHEN correctSelf is called on it
	shoutrrr.correctSelf()

	// THEN the leading / is trimmed from groupowner
	got := shoutrrr.GetSelfURLField("groupowner")
	want := "bar"
	if got != want {
		t.Errorf("Got %s, want %s",
			got, want)
	}
}

func TestCorrectSelfZulipChatBotMail(t *testing.T) {
	// GIVEN a zulip_chat Shoutrrr with botmail
	shoutrrr := Shoutrrr{
		Type: "zulip_chat",
		URLFields: map[string]string{
			"botmail": "x@y.com",
		},
	}

	// WHEN correctSelf is called on it
	shoutrrr.correctSelf()

	// THEN the @ is replaced with %40 in botmail
	got := shoutrrr.GetSelfURLField("botmail")
	want := "x%40y.com"
	if got != want {
		t.Errorf("Got %s, want %s",
			got, want)
	}
}

func TestCheckValuesMasterWithValidDiscord(t *testing.T) {
	// GIVEN a valid Discord Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "discord",
		URLFields: map[string]string{
			"token":     "foo",
			"webhookid": "bar",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN no errors were produced
	if errs != nil {
		t.Errorf("%v is valid, so shouldn't err!\nGot %s",
			shoutrrr, errs.Error())
	}
}

func TestCheckValuesMasterWithInvalidDiscord(t *testing.T) {
	// GIVEN an invalid Discord Shoutrrr
	shoutrrr := Shoutrrr{
		Type:         "discord",
		URLFields:    map[string]string{},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN 2 errors were produced
	e := utils.ErrorToString(errs)
	errCount := strings.Count(e, "\\")
	wantCount := 2
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestCheckValuesMasterWithValidSMTP(t *testing.T) {
	// GIVEN a valid SMTP Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "smtp",
		URLFields: map[string]string{
			"username": "bar",
			"password": "foo",
			"host":     "example.com",
			"port":     "123",
		},
		Params: map[string]string{
			"fromaddress": "me@me.com",
			"toaddresses": "you@you.com",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN no errors were produced
	if errs != nil {
		t.Errorf("%v is valid, so shouldn't err!\nGot %s",
			shoutrrr, errs.Error())
	}
}

func TestCheckValuesMasterWithInvalidSMTP(t *testing.T) {
	// GIVEN an invalid SMTP Shoutrrr
	shoutrrr := Shoutrrr{
		Type:         "smtp",
		URLFields:    map[string]string{},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN 3 errors were produced
	e := utils.ErrorToString(errs)
	errCount := strings.Count(e, "\\")
	wantCount := 3
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestCheckValuesMasterWithValidGotify(t *testing.T) {
	// GIVEN a valid Gotify Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "gotify",
		URLFields: map[string]string{
			"token": "foo",
			"host":  "example.com",
			"port":  "123",
			"path":  "test",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN no errors were produced
	if errs != nil {
		t.Errorf("%v is valid, so shouldn't err!\nGot %s",
			shoutrrr, errs.Error())
	}
}

func TestCheckValuesMasterWithInvalidGotify(t *testing.T) {
	// GIVEN an invalid Gotify Shoutrrr
	shoutrrr := Shoutrrr{
		Type:         "gotify",
		URLFields:    map[string]string{},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN 2 errors were produced
	e := utils.ErrorToString(errs)
	errCount := strings.Count(e, "\\")
	wantCount := 2
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestCheckValuesMasterWithValidGoogleChat(t *testing.T) {
	// GIVEN a valid GoogleChat Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "googlechat",
		URLFields: map[string]string{
			"raw": "chat.googleapis.com/v1/spaces/FOO/messages?key=bar&token=baz",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN no errors were produced
	if errs != nil {
		t.Errorf("%v is valid, so shouldn't err!\nGot %s",
			shoutrrr, errs.Error())
	}
}

func TestCheckValuesMasterWithInvalidGoogleChat(t *testing.T) {
	// GIVEN an invalid GoogleChat Shoutrrr
	shoutrrr := Shoutrrr{
		Type:         "googlechat",
		URLFields:    map[string]string{},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN 1 error was produced
	e := utils.ErrorToString(errs)
	errCount := strings.Count(e, "\\")
	wantCount := 1
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestCheckValuesMasterWithValidIFTTT(t *testing.T) {
	// GIVEN a valid IFTTT Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "ifttt",
		URLFields: map[string]string{
			"webhookid": "foo",
		},
		Params: map[string]string{
			"events": "event1,event2",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN no errors were produced
	if errs != nil {
		t.Errorf("%v is valid, so shouldn't err!\nGot %s",
			shoutrrr, errs.Error())
	}
}

func TestCheckValuesMasterWithInvalidIFTTT(t *testing.T) {
	// GIVEN an invalid IFTTT Shoutrrr
	shoutrrr := Shoutrrr{
		Type:         "ifttt",
		URLFields:    map[string]string{},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN 2 errors were produced
	e := utils.ErrorToString(errs)
	errCount := strings.Count(e, "\\")
	wantCount := 2
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestCheckValuesMasterWithValidJoin(t *testing.T) {
	// GIVEN a valid Join Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "join",
		URLFields: map[string]string{
			"apikey": "foo",
		},
		Params: map[string]string{
			"devices": "device1,device2",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN no errors were produced
	if errs != nil {
		t.Errorf("%v is valid, so shouldn't err!\nGot %s",
			shoutrrr, errs.Error())
	}
}

func TestCheckValuesMasterWithInvalidJoin(t *testing.T) {
	// GIVEN an invalid Join Shoutrrr
	shoutrrr := Shoutrrr{
		Type:         "join",
		URLFields:    map[string]string{},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN 2 errors were produced
	e := utils.ErrorToString(errs)
	errCount := strings.Count(e, "\\")
	wantCount := 2
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestCheckValuesMasterWithValidMattermost(t *testing.T) {
	// GIVEN a valid Mattermost Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "mattermost",
		URLFields: map[string]string{
			"username": "bar",
			"host":     "example.com",
			"port":     "123",
			"path":     "test",
			"token":    "bish",
			"channel":  "bash",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN no errors were produced
	if errs != nil {
		t.Errorf("%v is valid, so shouldn't err!\nGot %s",
			shoutrrr, errs.Error())
	}
}

func TestCheckValuesMasterWithInvalidMattermost(t *testing.T) {
	// GIVEN an invalid Mattermost Shoutrrr
	shoutrrr := Shoutrrr{
		Type:         "mattermost",
		URLFields:    map[string]string{},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN 2 errors were produced
	e := utils.ErrorToString(errs)
	errCount := strings.Count(e, "\\")
	wantCount := 2
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestCheckValuesMasterWithValidMatrix(t *testing.T) {
	// GIVEN a valid Matrix Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "matrix",
		URLFields: map[string]string{
			"user":     "foo",
			"password": "bar",
			"host":     "example.com",
			"port":     "123",
			"path":     "test",
		},
		Params: map[string]string{
			"rooms":      "!roomID1,roomAlias2",
			"disabletls": "yes",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN no errors were produced
	if errs != nil {
		t.Errorf("%v is valid, so shouldn't err!\nGot %s",
			shoutrrr, errs.Error())
	}
}

func TestCheckValuesMasterWithInvalidMatrix(t *testing.T) {
	// GIVEN an invalid Matrix Shoutrrr
	shoutrrr := Shoutrrr{
		Type:         "matrix",
		URLFields:    map[string]string{},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN 2 errors were produced
	e := utils.ErrorToString(errs)
	errCount := strings.Count(e, "\\")
	wantCount := 2
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestCheckValuesMasterWithValidOpsGenie(t *testing.T) {
	// GIVEN a valid OpsGenie Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "opsgenie",
		URLFields: map[string]string{
			"host":   "example.com",
			"port":   "123",
			"apikey": "foo",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN no errors were produced
	if errs != nil {
		t.Errorf("%v is valid, so shouldn't err!\nGot %s",
			shoutrrr, errs.Error())
	}
}

func TestCheckValuesMasterWithInvalidOpsGenie(t *testing.T) {
	// GIVEN an invalid OpsGenie Shoutrrr
	shoutrrr := Shoutrrr{
		Type:         "opsgenie",
		URLFields:    map[string]string{},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN 1 error was produced
	e := utils.ErrorToString(errs)
	errCount := strings.Count(e, "\\")
	wantCount := 1
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestCheckValuesMasterWithValidPushbullet(t *testing.T) {
	// GIVEN a valid Pushbullet Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "pushbullet",
		URLFields: map[string]string{
			"token":   "foo",
			"targets": "bar",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN no errors were produced
	if errs != nil {
		t.Errorf("%v is valid, so shouldn't err!\nGot %s",
			shoutrrr, errs.Error())
	}
}

func TestCheckValuesMasterWithInvalidPushbullet(t *testing.T) {
	// GIVEN an invalid Pushbullet Shoutrrr
	shoutrrr := Shoutrrr{
		Type:         "pushbullet",
		URLFields:    map[string]string{},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN 2 errors were produced
	e := utils.ErrorToString(errs)
	errCount := strings.Count(e, "\\")
	wantCount := 2
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestCheckValuesMasterWithValidPushover(t *testing.T) {
	// GIVEN a valid Pushover Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "pushover",
		URLFields: map[string]string{
			"token": "foo",
			"user":  "bar",
		},
		Params: map[string]string{
			"devices": "device1",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN no errors were produced
	if errs != nil {
		t.Errorf("%v is valid, so shouldn't err!\nGot %s",
			shoutrrr, errs.Error())
	}
}

func TestCheckValuesMasterWithInvalidPushover(t *testing.T) {
	// GIVEN an invalid Pushover Shoutrrr
	shoutrrr := Shoutrrr{
		Type:         "pushover",
		URLFields:    map[string]string{},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN 2 errors were produced
	e := utils.ErrorToString(errs)
	errCount := strings.Count(e, "\\")
	wantCount := 2
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestCheckValuesMasterWithValidRocketChat(t *testing.T) {
	// GIVEN a valid RocketChat Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "rocketchat",
		URLFields: map[string]string{
			"username": "foo",
			"host":     "example.com",
			"port":     "123",
			"tokena":   "bish",
			"tokenb":   "bash",
			"channel":  "bosh",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN no errors were produced
	if errs != nil {
		t.Errorf("%v is valid, so shouldn't err!\nGot %s",
			shoutrrr, errs.Error())
	}
}

func TestCheckValuesMasterWithInvalidRocketChat(t *testing.T) {
	// GIVEN an invalid RocketChat Shoutrrr
	shoutrrr := Shoutrrr{
		Type:         "rocketchat",
		URLFields:    map[string]string{},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN 4 errors were produced
	e := utils.ErrorToString(errs)
	errCount := strings.Count(e, "\\")
	wantCount := 4
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestCheckValuesMasterWithValidSlack(t *testing.T) {
	// GIVEN a valid Slack Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "slack",
		URLFields: map[string]string{
			"token":   "hook-foo",
			"channel": "bar",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN no errors were produced
	if errs != nil {
		t.Errorf("%v is valid, so shouldn't err!\nGot %s",
			shoutrrr, errs.Error())
	}
}

func TestCheckValuesMasterWithInvalidSlack(t *testing.T) {
	// GIVEN an invalid Slack Shoutrrr
	shoutrrr := Shoutrrr{
		Type:         "slack",
		URLFields:    map[string]string{},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN 2 errors were produced
	e := utils.ErrorToString(errs)
	errCount := strings.Count(e, "\\")
	wantCount := 2
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestCheckValuesMasterWithValidTeams(t *testing.T) {
	// GIVEN a valid Teams Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "teams",
		URLFields: map[string]string{
			"group":      "foo",
			"tenant":     "bish",
			"altid":      "bash",
			"groupowner": "bosh",
		},
		Params: map[string]string{
			"host": "example.webhook.office.com",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN no errors were produced
	if errs != nil {
		t.Errorf("%v is valid, so shouldn't err!\nGot %s",
			shoutrrr, errs.Error())
	}
}

func TestCheckValuesMasterWithInvalidTeams(t *testing.T) {
	// GIVEN an invalid Teams Shoutrrr
	shoutrrr := Shoutrrr{
		Type:         "teams",
		URLFields:    map[string]string{},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN 5 errors were produced
	e := utils.ErrorToString(errs)
	errCount := strings.Count(e, "\\")
	wantCount := 5
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestCheckValuesMasterWithValidTelegram(t *testing.T) {
	// GIVEN a valid Telegram Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "telegram",
		URLFields: map[string]string{
			"token": "foo",
		},
		Params: map[string]string{
			"chats": "channel1,channel2",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN no errors were produced
	if errs != nil {
		t.Errorf("%v is valid, so shouldn't err!\nGot %s",
			shoutrrr, errs.Error())
	}
}

func TestCheckValuesMasterWithInvalidTelegram(t *testing.T) {
	// GIVEN an invalid Telegram Shoutrrr
	shoutrrr := Shoutrrr{
		Type:         "telegram",
		URLFields:    map[string]string{},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN 2 errors were produced
	e := utils.ErrorToString(errs)
	errCount := strings.Count(e, "\\")
	wantCount := 2
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestCheckValuesMasterWithValidZulipChat(t *testing.T) {
	// GIVEN a valid ZulipChat Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "zulip_chat",
		URLFields: map[string]string{
			"botmail": "bish",
			"botkey":  "bash",
			"host":    "bosh",
			"port":    "123",
			"path":    "test",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN no errors were produced
	if errs != nil {
		t.Errorf("%v is valid, so shouldn't err!\nGot %s",
			shoutrrr, errs.Error())
	}
}

func TestCheckValuesMasterWithInvalidZulipChat(t *testing.T) {
	// GIVEN an invalid ZulipChat Shoutrrr
	shoutrrr := Shoutrrr{
		Type:         "zulip_chat",
		URLFields:    map[string]string{},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN 3 errors were produced
	e := utils.ErrorToString(errs)
	errCount := strings.Count(e, "\\")
	wantCount := 3
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestCheckValuesMasterWithValidShoutrrr(t *testing.T) {
	// GIVEN a valid Raw Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "shoutrrr",
		URLFields: map[string]string{
			"raw": "something",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN no errors were produced
	if errs != nil {
		t.Errorf("%v is valid, so shouldn't err!\nGot %s",
			shoutrrr, errs.Error())
	}
}

func TestCheckValuesMasterWithInvalidShoutrrr(t *testing.T) {
	// GIVEN an invalid Raw Shoutrrr
	shoutrrr := Shoutrrr{
		Type:         "shoutrrr",
		URLFields:    map[string]string{},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN 1 error was produced
	e := utils.ErrorToString(errs)
	errCount := strings.Count(e, "\\")
	wantCount := 1
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestCheckValuesMasterWithInvalidType(t *testing.T) {
	// GIVEN an invalid Raw Shoutrrr
	shoutrrr := Shoutrrr{
		Type:         "something",
		URLFields:    map[string]string{},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN 1 error was produced
	e := utils.ErrorToString(errs)
	errCount := strings.Count(e, "\\")
	wantCount := 1
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestCheckValuesMasterWithNoType(t *testing.T) {
	// GIVEN an invalid Raw Shoutrrr
	shoutrrr := Shoutrrr{
		URLFields:    map[string]string{},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN checkValuesMaster is called on this Shoutrrr
	var errs error
	shoutrrr.checkValuesMaster("", &errs, &errs, &errs, &errs)

	// THEN 1 error was produced
	e := utils.ErrorToString(errs)
	errCount := strings.Count(e, "\\")
	wantCount := 1
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestCheckValuesWithNil(t *testing.T) {
	// GIVEN a nil Shoutrrr
	var shoutrrr *Shoutrrr

	// WHEN CheckValues is called on this Shoutrrr
	err := shoutrrr.CheckValues("")

	// THEN no error was produced
	var want error
	if err != want {
		t.Errorf("CheckValues on %v should be %v errs, not %q",
			shoutrrr, want, err.Error())
	}
}

func TestCheckValuesWithValid(t *testing.T) {
	// GIVEN a valid SMTP Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "smtp",
		Options: map[string]string{
			"delay": "5s",
		},
		URLFields: map[string]string{
			"host": "stmp.example.com",
		},
		Params: map[string]string{
			"fromaddress": "me@me.com",
			"toaddresses": "1@you.com.2@you.com",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// THEN CheckValues is called on this Shoutrrr
	err := shoutrrr.CheckValues("")

	// THEN no error was produced
	if err != nil {
		t.Errorf("Err should not have been found on %v.\nGot %s",
			shoutrrr, err.Error())
	}
}

func TestCheckValuesWithInvalidOptions(t *testing.T) {
	// GIVEN an invalid SMTP Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "smtp",
		Options: map[string]string{
			"delay": "5x",
		},
		URLFields: map[string]string{
			"host": "stmp.example.com",
		},
		Params: map[string]string{
			"fromaddress": "me@me.com",
			"toaddresses": "1@you.com.2@you.com",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// THEN CheckValues is called on this Shoutrrr
	err := shoutrrr.CheckValues("")

	// THEN 1 error is produced
	e := utils.ErrorToString(err)
	errCount := strings.Count(e, "\\")
	wantCount := 2
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestCheckValuesWithIntDelay(t *testing.T) {
	// GIVEN an invalid SMTP Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "smtp",
		Options: map[string]string{
			"delay": "5",
		},
		URLFields: map[string]string{
			"host": "stmp.example.com",
		},
		Params: map[string]string{
			"fromaddress": "me@me.com",
			"toaddresses": "1@you.com.2@you.com",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// THEN CheckValues is called on this Shoutrrr
	err := shoutrrr.CheckValues("")

	// THEN 1 error is produced
	e := utils.ErrorToString(err)
	errCount := strings.Count(e, "\\")
	wantCount := 0
	if errCount != wantCount {
		t.Errorf("%v is valid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestCheckValuesWithInvalidURLFields(t *testing.T) {
	// GIVEN an invalid SMTP Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "smtp",
		Options: map[string]string{
			"delay": "5s",
		},
		URLFields: map[string]string{},
		Params: map[string]string{
			"fromaddress": "me@me.com",
			"toaddresses": "1@you.com.2@you.com",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// THEN CheckValues is called on this Shoutrrr
	err := shoutrrr.CheckValues("")

	// THEN 1 error is produced
	e := utils.ErrorToString(err)
	errCount := strings.Count(e, "\\")
	wantCount := 2
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestCheckValuesWithInvalidParams(t *testing.T) {
	// GIVEN an invalid SMTP Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "smtp",
		Options: map[string]string{
			"delay": "5s",
		},
		URLFields: map[string]string{
			"host": "stmp.example.com",
		},
		Params: map[string]string{
			"fromaddress": "me@me.com",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// THEN CheckValues is called on this Shoutrrr
	err := shoutrrr.CheckValues("")

	// THEN 1 error is produced
	e := utils.ErrorToString(err)
	errCount := strings.Count(e, "\\")
	wantCount := 2
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestCheckValuesWithInvalidLocate(t *testing.T) {
	// GIVEN an invalid Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "shoutrrr",
		URLFields: map[string]string{
			"raw": "teams://foo@bish/bash/bosh?host=example.webhook.office.com",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// THEN CheckValues is called on this Shoutrrr
	err := shoutrrr.CheckValues("")

	// THEN 1 error is produced
	e := utils.ErrorToString(err)
	errCount := strings.Count(e, "\\")
	wantCount := 1
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			shoutrrr, wantCount, errCount, e)
	}
}

func TestSlicePrintNil(t *testing.T) {
	// GIVEN a nil Slice
	var slice *Slice
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Print is called on it
	slice.Print("")

	// THEN no lines are printed
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	got := string(out)
	want := ""
	if got != want {
		t.Errorf("Print on %v should have produced %s, not %q",
			slice, want, got)
	}
}

func testPrintShoutrrr() Shoutrrr {
	return Shoutrrr{
		Type: "zulip_chat",
		Options: map[string]string{
			"message": "release",
		},
		URLFields: map[string]string{
			"botmail": "mail",
		},
		Params: map[string]string{
			"title": "something",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}
}

func TestSlicePrint(t *testing.T) {
	// GIVEN a Slice with 2+ Shoutrrr's
	validShoutrrr := testPrintShoutrrr()
	slice := Slice{
		"first":  &validShoutrrr,
		"second": &validShoutrrr,
	}
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Print is called on it
	slice.Print("")

	// THEN the expected number of lines are printed
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	want := 17
	got := strings.Count(string(out), "\n")
	if got != want {
		t.Errorf("Print should have given %d lines, but gave %d\n%s",
			want, got, out)
	}
}

func TestShoutrrrPrint(t *testing.T) {
	// GIVEN a Shoutrrr
	shoutrrr := testPrintShoutrrr()
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Print is called on it
	shoutrrr.Print("")

	// THEN the expected number of lines are printed
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	want := 7
	got := strings.Count(string(out), "\n")
	if got != want {
		t.Errorf("Print should have given %d lines, but gave %d\n%s",
			want, got, out)
	}
}
