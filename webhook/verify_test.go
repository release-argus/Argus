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

package webhook

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestSliceCheckValuesWithNil(t *testing.T) {
	// GIVEN a nil Slice
	var slice *Slice

	// WHEN CheckValues is called on it
	err := slice.CheckValues("")

	// THEN err is nil
	if err != nil {
		t.Errorf("CheckValues on %v returned an err - %s",
			slice, err.Error())
	}
}

func TestSliceCheckValuesWithInvalid(t *testing.T) {
	// GIVEN a Slice with an invalid var
	wType := "INVALID"
	url := "example.com"
	secret := "secret"
	slice := Slice{
		"test": &WebHook{
			Type:         &wType,
			URL:          &url,
			Secret:       &secret,
			Main:         &WebHook{},
			Defaults:     &WebHook{},
			HardDefaults: &WebHook{},
		},
	}

	// WHEN CheckValues is called on it
	err := slice.CheckValues("")

	// THEN err is not nil
	if err == nil {
		t.Errorf("CheckValues on %v should have err'd. Got %v",
			slice, err)
	}
}

func TestSliceCheckValuesWithValid(t *testing.T) {
	// GIVEN a Slice with a valid var
	wType := "github"
	slice := Slice{
		"test": &WebHook{Type: &wType},
	}

	// WHEN CheckValues is called on it
	err := slice.CheckValues("")

	// THEN err is nil
	if err != nil {
		t.Errorf("CheckValues on %v shouldn't have err'd. Got %s",
			slice, err.Error())
	}
}

func TestWebHookCheckValuesWithNil(t *testing.T) {
	// GIVEN a nil WebHook
	var webhook WebHook

	// WHEN CheckValues is called on it
	err := webhook.CheckValues("")

	// THEN err is nil
	if err != nil {
		t.Errorf("CheckValues on %v shouldn't have err'd. Got %s",
			webhook, err.Error())
	}
}

func TestWebHookCheckValuesWithNilDelay(t *testing.T) {
	// GIVEN a WebHook with nil Delay
	wType := "github"
	url := "example.com"
	secret := "secret"
	webhook := WebHook{
		Type:         &wType,
		URL:          &url,
		Secret:       &secret,
		Delay:        nil,
		Main:         &WebHook{},
		Defaults:     &WebHook{},
		HardDefaults: &WebHook{},
	}

	// WHEN CheckValues is called on it
	err := webhook.CheckValues("")

	// THEN err is nil
	if err != nil {
		t.Errorf("CheckValues on %v shouldn't have err'd. Got %s",
			webhook, err.Error())
	}
}

func TestWebHookCheckValuesWithIntDelay(t *testing.T) {
	// GIVEN a WebHook with int Delay
	wType := "github"
	url := "example.com"
	secret := "secret"
	declaredDelay := "5"
	delay := declaredDelay
	webhook := WebHook{
		Type:         &wType,
		URL:          &url,
		Secret:       &secret,
		Delay:        &delay,
		Main:         &WebHook{},
		Defaults:     &WebHook{},
		HardDefaults: &WebHook{},
	}

	// WHEN CheckValues is called on it
	webhook.CheckValues("")

	// THEN Delay is converted to seconds
	got := webhook.GetDelay()
	want := declaredDelay + "s"
	if got != want {
		t.Errorf("CheckValues on %v should have converted %s to seconds. Got %s, want %s",
			webhook, delay, got, want)
	}
}

func TestWebHookCheckValuesWithDurationDelay(t *testing.T) {
	// GIVEN a WebHook with duration Delay
	wType := "github"
	url := "example.com"
	secret := "secret"
	declaredDelay := "5s"
	delay := declaredDelay
	webhook := WebHook{
		Type:         &wType,
		URL:          &url,
		Secret:       &secret,
		Delay:        &delay,
		Main:         &WebHook{},
		Defaults:     &WebHook{},
		HardDefaults: &WebHook{},
	}

	// WHEN CheckValues is called on it
	webhook.CheckValues("")

	// THEN Delay is converted to seconds
	got := webhook.GetDelay()
	want := declaredDelay
	if got != want {
		t.Errorf("CheckValues on %v should have converted %s to seconds. Got %s, want %s",
			webhook, delay, got, want)
	}
}

func TestWebHookCheckValuesWithInvalidDurationDelay(t *testing.T) {
	// GIVEN a WebHook with invalid duration Delay
	wType := "github"
	url := "example.com"
	secret := "secret"
	declaredDelay := "5x"
	delay := declaredDelay
	webhook := WebHook{
		Type:         &wType,
		URL:          &url,
		Secret:       &secret,
		Delay:        &delay,
		Main:         &WebHook{},
		Defaults:     &WebHook{},
		HardDefaults: &WebHook{},
	}

	// WHEN CheckValues is called on it
	err := webhook.CheckValues("")

	// THEN Delay is converted to seconds
	if err == nil {
		t.Errorf("CheckValues on %v should have failed parsing %s. Got %s err",
			webhook, declaredDelay, err)
	}
}

func TestWebHookCheckValuesWithInvalidType(t *testing.T) {
	// GIVEN a WebHook with invalid duration Delay
	wType := "INVALID"
	url := "example.com"
	secret := "secret"
	declaredDelay := "5s"
	delay := declaredDelay
	webhook := WebHook{
		Type:         &wType,
		URL:          &url,
		Secret:       &secret,
		Delay:        &delay,
		Main:         &WebHook{},
		Defaults:     &WebHook{},
		HardDefaults: &WebHook{},
	}

	// WHEN CheckValues is called on it
	err := webhook.CheckValues("")

	// THEN Delay is converted to seconds
	if err == nil {
		t.Errorf("CheckValues on %v should have failed parsing %s Type. Got %s err",
			webhook, wType, err)
	}
}

func TestWebHookCheckValuesWithNilURL(t *testing.T) {
	// GIVEN a WebHook with invalid duration Delay
	id := "test"
	wType := "github"
	secret := "secret"
	declaredDelay := "5s"
	delay := declaredDelay
	webhook := WebHook{
		ID:           id,
		Type:         &wType,
		URL:          nil,
		Secret:       &secret,
		Delay:        &delay,
		Main:         &WebHook{},
		Defaults:     &WebHook{},
		HardDefaults: &WebHook{},
	}

	// WHEN CheckValues is called on it
	err := webhook.CheckValues("")

	// THEN Delay is converted to seconds
	if err == nil {
		t.Errorf("CheckValues on %v should have failed parsing nil URL. Got %s err",
			webhook, err)
	}
}

func TestWebHookCheckValuesWithNilSecret(t *testing.T) {
	// GIVEN a WebHook with invalid duration Delay
	id := "test"
	wType := "github"
	url := "example.com"
	declaredDelay := "5s"
	delay := declaredDelay
	webhook := WebHook{
		ID:           id,
		Type:         &wType,
		URL:          &url,
		Secret:       nil,
		Delay:        &delay,
		Main:         &WebHook{},
		Defaults:     &WebHook{},
		HardDefaults: &WebHook{},
	}

	// WHEN CheckValues is called on it
	err := webhook.CheckValues("")

	// THEN Delay is converted to seconds
	if err == nil {
		t.Errorf("CheckValues on %v should have failed parsing nil Secret. Got %s err",
			webhook, err)
	}
}

func TestWebHookCheckValuesWithNilMain(t *testing.T) {
	// GIVEN a WebHook with nil Main
	wType := "github"
	url := "example.com"
	secret := "secret"
	delay := "0s"
	webhook := WebHook{
		Type:   &wType,
		URL:    &url,
		Secret: &secret,
		Delay:  &delay,
	}

	// WHEN CheckValues is called on it
	err := webhook.CheckValues("")

	// THEN err is nil
	if err != nil {
		t.Errorf("CheckValues on %v shouldn't have err'd. Got %s",
			webhook, err.Error())
	}
}

func TestWebHookCheckValuesWithValid(t *testing.T) {
	// GIVEN a Slice with a valid var
	wType := "github"
	slice := Slice{
		"test": &WebHook{Type: &wType},
	}

	// WHEN CheckValues is called on it
	err := slice.CheckValues("")

	// THEN the program returns an err
	if err != nil {
		t.Errorf("CheckValues on %v shouldn't have err'd. Got %s",
			slice, err.Error())
	}
}

func TestSlicePrintWithFreshWebHook(t *testing.T) {
	// GIVEN a fresh WebHook
	var webhook WebHook
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Print is called on it
	webhook.Print("")

	// THEN nothing was logged
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	if string(out) != "" {
		t.Errorf("Print had output %q with %v WebHook",
			string(out), webhook)
	}
}

func TestSlicePrintWithFullWebHook(t *testing.T) {
	// GIVEN a WebHook that has every var filled
	id := "test"
	wType := "github"
	url := "example.com"
	allowInvalidCerts := false
	secret := "secret"
	desiredStatusCode := 201
	declaredDelay := "5s"
	delay := declaredDelay
	maxTries := uint(1)
	silentFails := false
	webhook := WebHook{
		ID:                id,
		Type:              &wType,
		URL:               &url,
		AllowInvalidCerts: &allowInvalidCerts,
		Secret:            &secret,
		DesiredStatusCode: &desiredStatusCode,
		Delay:             &delay,
		MaxTries:          &maxTries,
		SilentFails:       &silentFails,
		Main:              &WebHook{},
		Defaults:          &WebHook{},
		HardDefaults:      &WebHook{},
	}
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Print is called on it
	webhook.Print("")

	// THEN the WebHook was logged
	want := "type: github\nurl: example.com\nallow_invalid_certs: false\nsecret: \"secret\"\ndesired_status_code: 201\ndelay: 5s\nmax_tries: 1\nsilent_fails: false\n"
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	if string(out) != want {
		t.Errorf("Print had output %q with %v Slice. Wanted %q",
			string(out), webhook, want)
	}
}

func TestSlicePrintWithNilSlice(t *testing.T) {
	// GIVEN a nil Slice
	var slice *Slice
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Print is called on it
	slice.Print("")

	// THEN nothing was logged
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	if string(out) != "" {
		t.Errorf("Print had output %q with %v Slice",
			string(out), slice)
	}
}

func TestSlicePrintWithNonNilSlice(t *testing.T) {
	// GIVEN a Slice with a WebHook
	id := "test"
	wType0 := "github"
	url0 := "example.com"
	wType1 := "gitlab"
	url1 := "other.com"
	slice := Slice{
		"test": &WebHook{
			ID:   id,
			Type: &wType0,
			URL:  &url0,
		},
		"other": &WebHook{
			ID:   id,
			Type: &wType1,
			URL:  &url1,
		},
	}
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Print is called on it
	slice.Print("")

	// THEN the WebHook was logged
	want := "webhook:\n  other:\n    type: gitlab\n    url: other.com\n  test:\n    type: github\n    url: example.com\n"
	wantOther := "webhook:\n  test:\n    type: github\n    url: example.com\n  other:\n    type: gitlab\n    url: other.com\n"
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	if string(out) != want && string(out) != wantOther {
		t.Errorf("Print had output %q with %v Slice. Wanted %q",
			string(out), slice, want)
	}
}
