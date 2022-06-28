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

package service

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/webhook"
)

func TestServiceCheckValues(t *testing.T) {
	// GIVEN a service with valid values
	service := testServiceGitHub()

	// WHEN CheckValues is called
	err := service.CheckValues("")

	// THEN err is non-nil
	if err != nil {
		t.Errorf("Expect nil err, not\n%s",
			err.Error())
	}
}

func TestServiceCheckValuesWithInvalidInterval(t *testing.T) {
	// GIVEN a service with an invalid Interval
	service := testServiceGitHub()
	*service.Interval = "5x"

	// WHEN CheckValues is called
	err := service.CheckValues("")

	// THEN err is non-nil
	if err == nil {
		t.Errorf("Expecting err from %q being an invalid time.Duration, not %v",
			*service.Interval, err)
	}
}

func TestServiceCheckValuesWithIntInterval(t *testing.T) {
	// GIVEN a service with an integer Interval
	service := testServiceGitHub()
	*service.Interval = "5"
	want := *service.Interval + "s"

	// WHEN CheckValues is called
	err := service.CheckValues("")

	// THEN err is nil
	got := service.GetInterval()
	if got != want {
		t.Errorf("Want %s, got %s", want, got)
	}
	if err != nil {
		t.Errorf("Expecting %q interval err to be nil, not\n%s",
			*service.Interval, err.Error())
	}
}

func TestServiceCheckValuesWithNoType(t *testing.T) {
	// GIVEN a service with no Type
	service := testServiceGitHub()
	service.Type = nil

	// WHEN CheckValues is called
	err := service.CheckValues("")

	// THEN err is non-nil
	if err == nil {
		t.Errorf("Expecting err from Type being %v",
			service.Type)
	}
}

func TestServiceCheckValuesWithInvalidType(t *testing.T) {
	// GIVEN a service with no Type
	service := testServiceGitHub()
	*service.Type = "something"

	// WHEN CheckValues is called
	err := service.CheckValues("")

	// THEN err is non-nil
	if err == nil {
		t.Errorf("Expecting err from Type being %q",
			*service.Type)
	}
}

func TestServiceCheckValuesWithInvalidRegexContent(t *testing.T) {
	// GIVEN a service with invalid RegexContent
	service := testServiceGitHub()
	invalidRegex := "abc[0-"
	service.RegexContent = &invalidRegex

	// WHEN CheckValues is called
	err := service.CheckValues("")

	// THEN err is non-nil
	if err == nil {
		t.Errorf("Expecting err from RegexContent being %q",
			*service.RegexContent)
	}
}

func TestServiceCheckValuesWithInvalidRegexVersion(t *testing.T) {
	// GIVEN a service with invalid RegexVersion
	service := testServiceGitHub()
	invalidRegex := "abc[0-"
	service.RegexVersion = &invalidRegex

	// WHEN CheckValues is called
	err := service.CheckValues("")

	// THEN err is non-nil
	if err == nil {
		t.Errorf("Expecting err from RegexVersion being %q",
			*service.RegexVersion)
	}
}

func TestServiceCheckValuesWithInvalidDeployedVersionTimestamp(t *testing.T) {
	// GIVEN a Service with an Invalid DeployedVersionTimestamp
	service := testServiceGitHub()
	service.Status.DeployedVersionTimestamp = "abc"

	// WHEN CheckValues is called
	err := service.CheckValues("")

	// THEN err is non-nil
	if err == nil {
		t.Errorf("Expecting err from DeployedVersionTimestamp being %q",
			service.Status.DeployedVersionTimestamp)
	}
}

func TestServiceCheckValuesWithInvalidLatestVersionTimestamp(t *testing.T) {
	// GIVEN a Service with an Invalid LatestVersionTimestamp
	service := testServiceGitHub()
	service.Status.LatestVersionTimestamp = "abc"

	// WHEN CheckValues is called
	err := service.CheckValues("")

	// THEN err is non-nil
	if err == nil {
		t.Errorf("Expecting err from LatestVersionTimestamp being %q",
			service.Status.LatestVersionTimestamp)
	}
}

func TestServiceSliceCheckValuesWithSuccess(t *testing.T) {
	// GIVEN a Service with valid values
	service := testServiceGitHub()
	slice := Slice{
		"test": &service,
	}

	// WHEN CheckValues is called
	err := slice.CheckValues("")

	// THEN err is nil
	if err != nil {
		t.Errorf("%v is valid, unexpected err\n%s",
			slice, err.Error())
	}
}

func TestServiceSliceCheckValuesWithFailingService(t *testing.T) {
	// GIVEN a Service with an valid Service value
	service := testServiceGitHub()
	*service.Type = "foo"
	slice := Slice{
		"test": &service,
	}

	// WHEN CheckValues is called
	err := slice.CheckValues("")

	// THEN err is non-nil
	if err == nil {
		t.Errorf("Expecting err from Service.Type being %q",
			*slice["test"].Type)
	}
}

func TestServiceSliceCheckValuesWithFailingURLCommands(t *testing.T) {
	// GIVEN a Service with an invalid URLCommand value
	service := testServiceGitHub()
	urlCommand := testURLCommandRegex()
	urlCommand.Type = "something"
	service.URLCommands = &URLCommandSlice{urlCommand}
	slice := Slice{
		"test": &service,
	}

	// WHEN CheckValues is called
	err := slice.CheckValues("")

	// THEN err is non-nil
	if err == nil {
		t.Errorf("Expecting err from Service.URLCommands[0].Type being %q",
			(*slice["test"].URLCommands)[0].Type)
	}
}

func TestServiceSliceCheckValuesWithFailingDeployedVersionLookup(t *testing.T) {
	// GIVEN a Service with an invalid DeployedVersionLookup value
	service := testServiceGitHub()
	dvl := testDeployedVersion()
	dvl.URL = ""
	service.DeployedVersionLookup = &dvl
	slice := Slice{
		"test": &service,
	}

	// WHEN CheckValues is called
	err := slice.CheckValues("")

	// THEN err is non-nil
	if err == nil {
		t.Errorf("Expecting err from Service.DeployedVersionLookup.URL being %q",
			slice["test"].DeployedVersionLookup.URL)
	}
}

func TestServiceSliceCheckValuesWithFailingNotify(t *testing.T) {
	// GIVEN a Service with an invalid Notify value
	service := testServiceGitHub()
	notify := shoutrrr.Slice{
		"test": &shoutrrr.Shoutrrr{
			Type:         "something",
			Main:         &shoutrrr.Shoutrrr{},
			Defaults:     &shoutrrr.Shoutrrr{},
			HardDefaults: &shoutrrr.Shoutrrr{},
		},
	}
	service.Notify = &notify
	slice := Slice{
		"test": &service,
	}

	// WHEN CheckValues is called
	err := slice.CheckValues("")

	// THEN err is non-nil
	if err == nil {
		t.Errorf("Expecting err from Service.Notify['test'].Type being %q",
			(*slice["test"].Notify)["test"].Type)
	}
}

func TestServiceSliceCheckValuesWithFailingWebHook(t *testing.T) {
	// GIVEN a Service with an invalid WebHook value
	service := testServiceGitHub()
	whType := "something"
	whID := "test"
	wh := webhook.Slice{
		whID: &webhook.WebHook{
			ID:           &whID,
			Type:         &whType,
			Main:         &webhook.WebHook{},
			Defaults:     &webhook.WebHook{},
			HardDefaults: &webhook.WebHook{},
		},
	}
	service.WebHook = &wh
	slice := Slice{
		"test": &service,
	}

	// WHEN CheckValues is called
	err := slice.CheckValues("")

	// THEN err is non-nil
	if err == nil {
		t.Errorf("Expecting err from Service.WebHook['test'].Type being %q",
			*(*slice["test"].WebHook)["test"].Type)
	}
}

func TestServicePrintWithService(t *testing.T) {
	// GIVEN a Service
	service := testServiceGitHub()
	service.Status = nil
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN CheckValues is called
	service.Print("")

	// THEN the expected number of lines are printed
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	want := 13
	got := strings.Count(string(out), "\n")
	if got != want {
		t.Errorf("Print should have given %d lines, but gave %d\n%s", want, got, out)
	}
}

func TestServicePrintWithFullService(t *testing.T) {
	// GIVEN a Service with every var defined
	service := testServiceGitHub()
	urlCommand := testURLCommandRegex()
	service.URLCommands = &URLCommandSlice{urlCommand}
	dvl := testDeployedVersion()
	service.DeployedVersionLookup = &dvl
	notify := shoutrrr.Shoutrrr{
		Type:         "something",
		Main:         &shoutrrr.Shoutrrr{},
		Defaults:     &shoutrrr.Shoutrrr{},
		HardDefaults: &shoutrrr.Shoutrrr{},
	}
	service.Notify = &shoutrrr.Slice{"test": &notify}
	whID := "test"
	whURL := "example.com"
	wh := webhook.WebHook{
		ID:           &whID,
		URL:          &whURL,
		Main:         &webhook.WebHook{},
		Defaults:     &webhook.WebHook{},
		HardDefaults: &webhook.WebHook{},
	}
	service.WebHook = &webhook.Slice{"test": &wh}
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN CheckValues is called
	service.Print("")

	// THEN the expected number of lines are printed
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	want := 37
	got := strings.Count(string(out), "\n")
	if got != want {
		t.Errorf("Print should have given %d lines, but gave %d\n%s", want, got, out)
	}
}

func TestSlicePrintWithTwoServices(t *testing.T) {
	// GIVEN a Slice with two Service's
	service := testServiceGitHub()
	service.Status = nil
	slice := Slice{
		"one": &service,
		"two": &service,
	}
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN CheckValues is called
	slice.Print("", []string{"one", "two"})

	// THEN the expected number of lines are printed
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	want := 29
	got := strings.Count(string(out), "\n")
	if got != want {
		t.Errorf("Print should have given %d lines, but gave %d\n%s", want, got, out)
	}
}

func TestSlicePrintWithNil(t *testing.T) {
	// GIVEN a nil Slice
	var slice *Slice
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN CheckValues is called
	slice.Print("", []string{})

	// THEN the expected number of lines are printed
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	want := 0
	got := strings.Count(string(out), "\n")
	if got != want {
		t.Errorf("Print with nil should have given %d lines, but gave %d\n%s", want, got, out)
	}
}

func TestSlicePrintWithTwoServicesOrdered(t *testing.T) {
	// GIVEN a Slice with two Service's
	service := testServiceGitHub()
	service.Status = nil
	slice := Slice{
		"one": &service,
		"two": &service,
	}

	// WHEN CheckValues is called with two different orderings
	// 1
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	wantFirst := "one"
	wantSecond := "two"
	slice.Print("", []string{wantFirst, wantSecond})
	w.Close()
	gotOne, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	// 2
	stdout = os.Stdout
	r, w, _ = os.Pipe()
	os.Stdout = w
	slice.Print("", []string{wantSecond, wantFirst})
	w.Close()
	gotTwo, _ := ioutil.ReadAll(r)
	os.Stdout = stdout

	// THEN the Services are printed in two different orderings
	if string(gotOne) == string(gotTwo) {
		t.Errorf("Print should have used ordering\n%s", gotOne)
	}
}
