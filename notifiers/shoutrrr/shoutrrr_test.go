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

package shoutrrr

import (
	"testing"

	shoutrrr_types "github.com/containrrr/shoutrrr/pkg/types"
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
	gotParams := test.GetParams()
	for key := range *gotParams {
		if (*gotParams)[key] != wantedParams[key] {
			t.Fatalf(`Shoutrrr, GetParams - Got %v, want match for %q`, (*gotParams)[key], wantedParams[key])
		}
	}
	for key := range wantedParams {
		if (*gotParams)[key] != wantedParams[key] {
			t.Fatalf(`Shoutrrr, GetParams - Got %v, want match for %q`, (*gotParams)[key], wantedParams[key])
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
	gotParams = test.GetParams()
	for key := range *gotParams {
		if (*gotParams)[key] != wantedParams[key] {
			t.Fatalf(`Shoutrrr, GetParams - Got %v, want match for %q`, (*gotParams)[key], wantedParams[key])
		}
	}
	for key := range wantedParams {
		if (*gotParams)[key] != wantedParams[key] {
			t.Fatalf(`Shoutrrr, GetParams - Got %v, want match for %q`, (*gotParams)[key], wantedParams[key])
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
		t.Fatalf(`Shoutrrr, GetURL - Got %v, want match for %q`, gotURL, wantedURL)
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
		t.Fatalf(`Shoutrrr, GetURL - Got %v, want match for %q`, gotURL, wantedURL)
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
		t.Fatalf(`Shoutrrr, GetURL - Got %v, want match for %q`, gotURL, wantedURL)
	}
}
