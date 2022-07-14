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

	shoutrrr_types "github.com/containrrr/shoutrrr/pkg/types"
	"github.com/release-argus/Argus/utils"
)

func TestShoutrrr(t *testing.T) {
	// Test one Params
	defaultShoutrr := Shoutrrr{
		Options:   map[string]string{},
		URLFields: map[string]string{},
		Params:    shoutrrr_types.Params{},
	}
	test := Shoutrrr{
		Main:         &defaultShoutrr,
		Defaults:     &defaultShoutrr,
		HardDefaults: &defaultShoutrr,
		Params: shoutrrr_types.Params{
			"BotName": "Test",
		}}
	wantedParams := map[string]string{
		"botname": "Test",
	}

	test.initParams()
	gotParams := test.GetParams(&utils.ServiceInfo{})
	for key := range *gotParams {
		if (*gotParams)[key] != wantedParams[key] {
			t.Errorf(`Shoutrrr, GetParams - Got %v, want match for %q`, (*gotParams)[key], wantedParams[key])
		}
	}
	for key := range wantedParams {
		if (*gotParams)[key] != wantedParams[key] {
			t.Errorf(`Shoutrrr, GetParams - Got %v, want match for %q`, (*gotParams)[key], wantedParams[key])
		}
	}

	// Test multiple Params
	test = Shoutrrr{
		Main:         &defaultShoutrr,
		Defaults:     &defaultShoutrr,
		HardDefaults: &defaultShoutrr,
		Params: shoutrrr_types.Params{
			"BotName": "OtherTest",
			"Icon":    "github",
		}}
	wantedParams = map[string]string{
		"botname": "OtherTest",
		"icon":    "github",
	}
	test.initParams()
	gotParams = test.GetParams(&utils.ServiceInfo{})
	for key := range *gotParams {
		if (*gotParams)[key] != wantedParams[key] {
			t.Errorf(`Shoutrrr, GetParams - Got %v, want match for %q`, (*gotParams)[key], wantedParams[key])
		}
	}
	for key := range wantedParams {
		if (*gotParams)[key] != wantedParams[key] {
			t.Errorf(`Shoutrrr, GetParams - Got %v, want match for %q`, (*gotParams)[key], wantedParams[key])
		}
	}

	// Test GetURL with one Params
	testType := "discord"
	testURLFields := map[string]string{
		"Token":     "bar",
		"WebhookID": "foo",
	}
	test = Shoutrrr{
		Main:         &defaultShoutrr,
		Defaults:     &defaultShoutrr,
		HardDefaults: &defaultShoutrr,
		Type:         testType,
		URLFields:    testURLFields,
		Params:       test.Params,
	}
	wantedURL := "discord://bar@foo"
	test.initURLFields()
	gotURL := test.GetURL()
	if gotURL != wantedURL {
		t.Errorf(`Shoutrrr, GetURL - Got %v, want match for %q`, gotURL, wantedURL)
	}

	// Test GetURL with multiple Params
	testType = "teams"
	testURLFields = map[string]string{
		"Group":      "something",
		"Tenant":     "foo",
		"AltID":      "bar",
		"GroupOwner": "fez",
	}
	testParams := shoutrrr_types.Params{
		"Host":  "mockhost",
		"Title": "test",
	}
	test = Shoutrrr{
		Main:         &defaultShoutrr,
		Defaults:     &defaultShoutrr,
		HardDefaults: &defaultShoutrr,
		Type:         testType,
		URLFields:    testURLFields,
		Params:       testParams,
	}
	wantedURL = "teams://something@foo/bar/fez?host=mockhost"
	test.initParams()
	test.initURLFields()
	gotURL = test.GetURL()
	if gotURL != wantedURL {
		t.Errorf(`Shoutrrr, GetURL - Got %v, want match for %q`, gotURL, wantedURL)
	}

	// Test Defaults
	testType = "teams"
	testServiceURLFields := map[string]string{
		"Group": "something",
	}
	testMainURLFields := map[string]string{
		"Group":  "main",
		"Tenant": "foo",
	}
	testDefaultsURLFields := map[string]string{
		"Group":  "default",
		"Tenant": "default",
		"AltID":  "bar",
	}
	testHardDefaultsURLFields := map[string]string{
		"Group":      "hardDefault",
		"Tenant":     "hardDefault",
		"AltID":      "hardDefault",
		"GroupOwner": "fez",
	}
	test = Shoutrrr{
		Main:         &Shoutrrr{URLFields: testMainURLFields},
		Defaults:     &Shoutrrr{URLFields: testDefaultsURLFields},
		HardDefaults: &Shoutrrr{URLFields: testHardDefaultsURLFields},
		Type:         testType,
		URLFields:    testServiceURLFields,
		Params:       testParams,
	}
	wantedURL = "teams://something@foo/bar/fez?host=mockhost"
	test.initParams()
	test.initURLFields()
	test.Main.initURLFields()
	test.Defaults.initURLFields()
	test.HardDefaults.initURLFields()
	gotURL = test.GetURL()
	if gotURL != wantedURL {
		t.Errorf(`Shoutrrr, GetURL - Got %v, want match for %q`, gotURL, wantedURL)
	}
}

func testGetParams() Shoutrrr {
	main := Shoutrrr{
		Params: map[string]string{
			"main": "main-val",
		},
	}
	defaults := Shoutrrr{
		Params: map[string]string{
			"main":    "default-not-main",
			"default": "default-val",
		},
	}
	hardDefaults := Shoutrrr{
		Params: map[string]string{
			"main":        "harddefault-not-main",
			"default":     "harddefault-not-default",
			"harddefault": "harddefault-val",
		},
	}
	return Shoutrrr{
		Params: map[string]string{
			"master":   "master-val",
			"template": "{% if 'a' == 'a' %}{{ version }}-yes{% endif %}",
		},
		Main:         &main,
		Defaults:     &defaults,
		HardDefaults: &hardDefaults,
	}
}

func TestGetParamsMaster(t *testing.T) {
	// GIVEN a Shoutrrr
	shoutrrr := testGetParams()

	// WHEN GetParams is called
	params := shoutrrr.GetParams(&utils.ServiceInfo{})

	// THEN it'll find keys only defined in master correctly
	key := "master"
	got := (*params)[key]
	want := key + "-val"
	if got != want {
		t.Errorf("%s key not looked up correctly from master. Got %s, want %s",
			key, got, want)
	}
}

func TestGetParamsMain(t *testing.T) {
	// GIVEN a Shoutrrr
	shoutrrr := testGetParams()

	// WHEN GetParams is called
	params := shoutrrr.GetParams(&utils.ServiceInfo{})

	// THEN it'll find keys only defined in Main correctly
	key := "main"
	got := (*params)[key]
	want := key + "-val"
	if got != want {
		t.Errorf("%s key not looked up correctly from main. Got %s, want %s",
			key, got, want)
	}
}

func TestGetParamsDefault(t *testing.T) {
	// GIVEN a Shoutrrr
	shoutrrr := testGetParams()

	// WHEN GetParams is called
	params := shoutrrr.GetParams(&utils.ServiceInfo{})

	// THEN it'll find keys only defined in Default correctly
	key := "default"
	got := (*params)[key]
	want := key + "-val"
	if got != want {
		t.Errorf("%s key not looked up correctly from default. Got %s, want %s",
			key, got, want)
	}
}

func TestGetParamsHardDefault(t *testing.T) {
	// GIVEN a Shoutrrr
	shoutrrr := testGetParams()

	// WHEN GetParams is called
	params := shoutrrr.GetParams(&utils.ServiceInfo{})

	// THEN it'll find keys only defined in HardDefault correctly
	key := "harddefault"
	got := (*params)[key]
	want := key + "-val"
	if got != want {
		t.Errorf("%s key not looked up correctly from harddefault. Got %s, want %s",
			key, got, want)
	}
}

func TestGetParamsTemplating(t *testing.T) {
	// GIVEN a Shoutrrr and Service.Info context
	shoutrrr := testGetParams()
	context := utils.ServiceInfo{LatestVersion: "1.2.3"}

	// WHEN GetParams is called
	params := shoutrrr.GetParams(&context)

	// THEN it'll find keys only defined in HardDefault correctly
	got := (*params)["template"]
	want := "1.2.3-yes"
	if got != want {
		t.Errorf("Param templating not applied as expected. Got %s, want %s",
			got, want)
	}
}

func TestGetURLForDiscord(t *testing.T) {
	// GIVEN a Discord Shoutrrr
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

	// WHEN GetURL is called on this Shoutrrr
	got := shoutrrr.GetURL()

	// THEN the Shoutrrr URL is correctly formatted
	want := "discord://foo@bar"
	if got != want {
		t.Errorf("Unexpected URL returned. Got %q, want %q",
			got, want)
	}
}

func TestGetURLForSMTP(t *testing.T) {
	// GIVEN a SMTP Shoutrrr
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

	// WHEN GetURL is called on this Shoutrrr
	got := shoutrrr.GetURL()

	// THEN the Shoutrrr URL is correctly formatted
	want := "smtp://bar:foo@example.com:123/?fromaddress=me@me.com&toaddresses=you@you.com"
	if got != want {
		t.Errorf("Unexpected URL returned. Got %q, want %q",
			got, want)
	}
}

func TestGetURLForGotify(t *testing.T) {
	// GIVEN a Gotify Shoutrrr
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

	// WHEN GetURL is called on this Shoutrrr
	got := shoutrrr.GetURL()

	// THEN the Shoutrrr URL is correctly formatted
	want := "gotify://example.com:123/test/foo"
	if got != want {
		t.Errorf("Unexpected URL returned. Got %q, want %q",
			got, want)
	}
}

func TestGetURLForGoogleChat(t *testing.T) {
	// GIVEN a GoogleChat Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "googlechat",
		URLFields: map[string]string{
			"raw": "chat.googleapis.com/v1/spaces/FOO/messages?key=bar&token=baz",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN GetURL is called on this Shoutrrr
	got := shoutrrr.GetURL()

	// THEN the Shoutrrr URL is correctly formatted
	want := "googlechat://chat.googleapis.com/v1/spaces/FOO/messages?key=bar&token=baz"
	if got != want {
		t.Errorf("Unexpected URL returned. Got %q, want %q",
			got, want)
	}
}

func TestGetURLForIFTTT(t *testing.T) {
	// GIVEN a IFTTT Shoutrrr
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

	// WHEN GetURL is called on this Shoutrrr
	got := shoutrrr.GetURL()

	// THEN the Shoutrrr URL is correctly formatted
	want := "ifttt://foo/?events=event1,event2"
	if got != want {
		t.Errorf("Unexpected URL returned. Got %q, want %q",
			got, want)
	}
}

func TestGetURLForJoin(t *testing.T) {
	// GIVEN a Join Shoutrrr
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

	// WHEN GetURL is called on this Shoutrrr
	got := shoutrrr.GetURL()

	// THEN the Shoutrrr URL is correctly formatted
	want := "join://shoutrrr:foo@join/?devices=device1,device2"
	if got != want {
		t.Errorf("Unexpected URL returned. Got %q, want %q",
			got, want)
	}
}

func TestGetURLForMattermost(t *testing.T) {
	// GIVEN a Mattermost Shoutrrr
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

	// WHEN GetURL is called on this Shoutrrr
	got := shoutrrr.GetURL()

	// THEN the Shoutrrr URL is correctly formatted
	want := "mattermost://bar@example.com:123/test/bish/bash"
	if got != want {
		t.Errorf("Unexpected URL returned. Got %q, want %q",
			got, want)
	}
}

func TestGetURLForMatrix(t *testing.T) {
	// GIVEN a Matrix Shoutrrr
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

	// WHEN GetURL is called on this Shoutrrr
	got := shoutrrr.GetURL()

	// THEN the Shoutrrr URL is correctly formatted
	want := "matrix://foo:bar@example.com:123/test/?rooms=!roomID1,roomAlias2&disableTLS=yes"
	if got != want {
		t.Errorf("Unexpected URL returned. Got %q, want %q",
			got, want)
	}
}

func TestGetURLForMatrixNoRooms(t *testing.T) {
	// GIVEN a Matrix Shoutrrr
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
			"disabletls": "yes",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN GetURL is called on this Shoutrrr
	got := shoutrrr.GetURL()

	// THEN the Shoutrrr URL is correctly formatted
	want := "matrix://foo:bar@example.com:123/test/?disableTLS=yes"
	if got != want {
		t.Errorf("Unexpected URL returned. Got %q, want %q",
			got, want)
	}
}

func TestGetURLForOpsGenie(t *testing.T) {
	// GIVEN a OpsGenie Shoutrrr
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

	// WHEN GetURL is called on this Shoutrrr
	got := shoutrrr.GetURL()

	// THEN the Shoutrrr URL is correctly formatted
	want := "opsgenie://example.com:123/foo"
	if got != want {
		t.Errorf("Unexpected URL returned. Got %q, want %q",
			got, want)
	}
}

func TestGetURLForPushbullet(t *testing.T) {
	// GIVEN a Pushbullet Shoutrrr
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

	// WHEN GetURL is called on this Shoutrrr
	got := shoutrrr.GetURL()

	// THEN the Shoutrrr URL is correctly formatted
	want := "pushbullet://foo/bar"
	if got != want {
		t.Errorf("Unexpected URL returned. Got %q, want %q",
			got, want)
	}
}

func TestGetURLForPushover(t *testing.T) {
	// GIVEN a Pushover Shoutrrr
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

	// WHEN GetURL is called on this Shoutrrr
	got := shoutrrr.GetURL()

	// THEN the Shoutrrr URL is correctly formatted
	want := "pushover://shoutrrr:foo@bar/?devices=device1"
	if got != want {
		t.Errorf("Unexpected URL returned. Got %q, want %q",
			got, want)
	}
}

func TestGetURLForRocketChat(t *testing.T) {
	// GIVEN a RocketChat Shoutrrr
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

	// WHEN GetURL is called on this Shoutrrr
	got := shoutrrr.GetURL()

	// THEN the Shoutrrr URL is correctly formatted
	want := "rocketchat://foo@example.com:123/bish/bish/bosh"
	if got != want {
		t.Errorf("Unexpected URL returned. Got %q, want %q",
			got, want)
	}
}

func TestGetURLForSlack(t *testing.T) {
	// GIVEN a Slack Shoutrrr
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

	// WHEN GetURL is called on this Shoutrrr
	got := shoutrrr.GetURL()

	// THEN the Shoutrrr URL is correctly formatted
	want := "slack://hook-foo@bar"
	if got != want {
		t.Errorf("Unexpected URL returned. Got %q, want %q",
			got, want)
	}
}

func TestGetURLForTeams(t *testing.T) {
	// GIVEN a Teams Shoutrrr
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

	// WHEN GetURL is called on this Shoutrrr
	got := shoutrrr.GetURL()

	// THEN the Shoutrrr URL is correctly formatted
	want := "teams://foo@bish/bash/bosh?host=example.webhook.office.com"
	if got != want {
		t.Errorf("Unexpected URL returned. Got %q, want %q",
			got, want)
	}
}

func TestGetURLForTelegram(t *testing.T) {
	// GIVEN a Telegram Shoutrrr
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

	// WHEN GetURL is called on this Shoutrrr
	got := shoutrrr.GetURL()

	// THEN the Shoutrrr URL is correctly formatted
	want := "telegram://foo@telegram?chats=channel1,channel2"
	if got != want {
		t.Errorf("Unexpected URL returned. Got %q, want %q",
			got, want)
	}
}

func TestGetURLForZulipChat(t *testing.T) {
	// GIVEN a ZulipChat Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "zulipchat",
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

	// WHEN GetURL is called on this Shoutrrr
	got := shoutrrr.GetURL()

	// THEN the Shoutrrr URL is correctly formatted
	want := "zulip://bish:bash@bosh:123/test"
	if got != want {
		t.Errorf("Unexpected URL returned. Got %q, want %q",
			got, want)
	}
}

func TestGetURLForShoutrrr(t *testing.T) {
	// GIVEN an invalid raw Shoutrrr
	shoutrrr := Shoutrrr{
		Type: "shoutrrr",
		URLFields: map[string]string{
			"raw": "something://bish:bash:bosh",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}

	// WHEN GetURL is called on this Shoutrrr
	got := shoutrrr.GetURL()

	// THEN the Shoutrrr URL is correctly formatted
	want := "something://bish:bash:bosh"
	if got != want {
		t.Errorf("Unexpected URL returned. Got %q, want %q",
			got, want)
	}
}

func TestSendWithNil(t *testing.T) {
	// GIVEN a nil slice
	var slice *Slice

	// WHEN Send is called on it
	err := slice.Send("", "", nil)

	// THEN err is nil
	var want error
	if err != want {
		t.Errorf("Send on %v should have produced %v err, not\n%v",
			slice, want, err)
	}
}

func TestSendWithNilServiceInfo(t *testing.T) {
	// GIVEN a Slice with an invalid Shoutrrr
	jLog = utils.NewJLog("WARN", false)
	id := "test"
	slice := Slice{
		"test": &Shoutrrr{
			ID:   &id,
			Type: "shoutrrr",
			Options: map[string]string{
				"max_tries": "1",
				"delay":     "0s",
			},
			URLFields: map[string]string{
				"raw": "something://bish:bash:bosh",
			},
			Main:         &Shoutrrr{},
			Defaults:     &Shoutrrr{},
			HardDefaults: &Shoutrrr{},
		},
	}
	// WHEN Send is called on this Shoutrrr with a nil serviceInfo
	err := slice.Send("title", "message", nil)

	// THEN the Send errors related to the Shoutrrr being invalid
	contain := " invalid port "
	e := utils.ErrorToString(err)
	if !strings.Contains(e, contain) {
		t.Errorf("Send should err about the %q, not\n%v",
			contain, e)
	}
}

func TestSendWithFail(t *testing.T) {
	// GIVEN a Slice with an invalid Shoutrrr
	jLog = utils.NewJLog("WARN", false)
	id := "test"
	slice := Slice{
		"test": &Shoutrrr{
			ID:   &id,
			Type: "slack",
			Options: map[string]string{
				"max_tries": "2",
				"delay":     "1s",
			},
			URLFields: map[string]string{
				"token":   "hook:WNA3PBYV6-F20DUQND3RQ-Webc4MAvoacrpPakR8phF0zi",
				"channel": "bar",
			},
			Main:         &Shoutrrr{},
			Defaults:     &Shoutrrr{},
			HardDefaults: &Shoutrrr{},
		},
	}
	// WHEN Send is called on this Shoutrrr with a nil serviceInfo
	err := slice.Send("title", "message", nil)

	// THEN the Send errors related to the Shoutrrr being invalid
	contain := "failed to send slack notification:"
	e := utils.ErrorToString(err)
	if !strings.Contains(e, contain) {
		t.Errorf("Send should err about %q, not\n%v",
			contain, e)
	}
}

func TestSendWithMultipleFails(t *testing.T) {
	// GIVEN a Slice with an invalid Shoutrrr
	jLog = utils.NewJLog("WARN", false)
	id := "test"
	failingShoutrrr := &Shoutrrr{
		ID:   &id,
		Type: "slack",
		Options: map[string]string{
			"max_tries": "2",
			"delay":     "1s",
		},
		URLFields: map[string]string{
			"token":   "hook:WNA3PBYV6-F20DUQND3RQ-Webc4MAvoacrpPakR8phF0zi",
			"channel": "bar",
		},
		Main:         &Shoutrrr{},
		Defaults:     &Shoutrrr{},
		HardDefaults: &Shoutrrr{},
	}
	slice := Slice{
		"test":  failingShoutrrr,
		"other": failingShoutrrr,
	}
	// WHEN Send is called on this Shoutrrr with a nil serviceInfo
	err := slice.Send("title", "message", nil)

	// THEN the Send errors related to the Shoutrrr being invalid
	contain := "failed to send slack notification:"
	e := utils.ErrorToString(err)
	if strings.Count(e, contain) != len(slice) {
		t.Errorf("Send should err about %q %d times, not\n%v",
			contain, len(slice), e)
	}
}
