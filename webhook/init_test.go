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
	"fmt"
	"testing"
	"time"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	service_status "github.com/release-argus/Argus/service/status"
)

func TestSliceInitWithNilSlice(t *testing.T) {
	// GIVEN a nil Slice
	var slice Slice

	// WHEN Init is called
	slice.Init(nil, nil, nil, nil, nil, nil, nil, nil)

	// THEN the function exits without defining any of the vars
	got := len(slice)
	want := 0
	if got != want {
		t.Errorf("Got len %d, when expecting %d",
			got, want)
	}
}

func TestSliceInitWithNilWebHook(t *testing.T) {
	// GIVEN a non-nil Slice with a nil WebHook
	slice := Slice{
		"0": nil,
	}
	serviceID := "test"

	// WHEN Init is called
	slice.Init(nil, &serviceID, nil, nil, nil, nil, nil, nil)

	// THEN the function initialises the nil WebHook
	if slice["0"] == nil {
		t.Errorf("Slice['0'] shouldn't be %v still",
			slice["0"])
	}
}

func TestSliceInitWithNonNil(t *testing.T) {
	// GIVEN a non-nil Slice with everything else nil
	id0 := "0"
	slice := Slice{
		"0": &WebHook{
			ID: id0,
		},
	}
	// Won't ever be nil, so don't make it nil in tests
	serviceID := ""

	// WHEN Init is called
	slice.Init(nil, &serviceID, nil, nil, nil, nil, nil, nil)

	// THEN the function exits without defining any of the vars
	got := len(slice)
	want := 1
	if got != want {
		t.Errorf("Got len %d, when expecting %d",
			got, want)
	}
}

func testInitWithNonNilAndVars() (string, Slice, service_status.Status, Slice, WebHook, WebHook) {
	id0 := "0"
	id1 := "1"
	slice := Slice{
		id0: &WebHook{
			ID: id0,
		},
		id1: &WebHook{
			ID: id1,
		},
	}
	serviceID := "id_test"
	serviceStatus := service_status.Status{LatestVersion: "status test"}
	mainSecret0 := "main0"
	mainSecret1 := "main1"
	mains := Slice{
		id0: &WebHook{
			ID:     id0,
			Secret: &mainSecret0,
		},
		id1: &WebHook{
			ID:     id1,
			Secret: &mainSecret1,
		},
	}
	defaultURL := "default"
	defaults := WebHook{
		ID:     id0,
		Secret: &defaultURL,
	}
	hardDefaults := WebHook{
		Delay: "1s",
	}

	return serviceID, slice, serviceStatus, mains, defaults, hardDefaults
}

func TestSliceInitMainsHandedOut(t *testing.T) {
	// GIVEN a non-nil Slice, matching mains, defaults and hardDefaults
	serviceID, slice, serviceStatus, mains, defaults, hardDefaults := testInitWithNonNilAndVars()

	// WHEN Init is called
	slice.Init(nil, &serviceID, &serviceStatus, &mains, &defaults, &hardDefaults, nil, nil)

	// THEN the mains are handed out correctly
	for key := range slice {
		if slice[key].Main != mains[key] {
			t.Errorf("Main not handed to %s",
				key)
		}
	}
}

func TestSliceInitDefaultsHandedOut(t *testing.T) {
	// GIVEN a non-nil Slice, matching mains, defaults and hardDefaults
	serviceID, slice, serviceStatus, mains, defaults, hardDefaults := testInitWithNonNilAndVars()

	// WHEN Init is called
	slice.Init(nil, &serviceID, &serviceStatus, &mains, &defaults, &hardDefaults, nil, nil)

	// THEN the defaults are handed out correctly
	for key := range slice {
		if *slice[key].Defaults != defaults {
			t.Errorf("Defaults not handed to %s",
				key)
		}
	}
}

func TestSliceInitHardDefaultsHandedOut(t *testing.T) {
	// GIVEN a non-nil Slice, matching mains, defaults and hardDefaults
	serviceID, slice, serviceStatus, mains, defaults, hardDefaults := testInitWithNonNilAndVars()

	// WHEN Init is called
	slice.Init(nil, &serviceID, &serviceStatus, &mains, &defaults, &hardDefaults, nil, nil)

	// THEN the hard defaults are handed out correctly
	for key := range slice {
		if *slice[key].HardDefaults != hardDefaults {
			t.Errorf("HardDefaults not handed to %s",
				key)
		}
	}
}

func TestSliceInitNotifiersHandedOut(t *testing.T) {
	// GIVEN a non-nil Slice with notifiers
	serviceID, slice, serviceStatus, mains, defaults, hardDefaults := testInitWithNonNilAndVars()
	notifiers := shoutrrr.Slice{
		"test": &shoutrrr.Shoutrrr{},
	}

	// WHEN Init is called
	slice.Init(nil, &serviceID, &serviceStatus, &mains, &defaults, &hardDefaults, &notifiers, nil)

	// THEN the notifiers are handed out correctly
	for key := range slice {
		if slice[key].Notifiers.Shoutrrr == nil {
			t.Errorf("Notifiers %v weren't handed to %s",
				notifiers, key)
		}
	}
}

func TestSliceInitMetrics(t *testing.T) {
	// GIVEN a non-nil Slice, matching mains, defaults and hardDefaults
	serviceID, slice, serviceStatus, mains, defaults, hardDefaults := testInitWithNonNilAndVars()
	slice.Init(nil, &serviceID, &serviceStatus, &mains, &defaults, &hardDefaults, nil, nil)

	// WHEN initMetrics is called
	slice["0"].initMetrics("foo")

	// THEN the function runs without error
}

func TestGetAllowInvalidCerts(t *testing.T) {
	// GIVEN a non-nil Slice, matching mains, defaults and hardDefaults
	serviceID, slice, serviceStatus, mains, defaults, hardDefaults := testInitWithNonNilAndVars()
	slice.Init(nil, &serviceID, &serviceStatus, &mains, &defaults, &hardDefaults, nil, nil)
	tests := map[string]struct {
		allowInvalidCertsRoot        *bool
		allowInvalidCertsMain        *bool
		allowInvalidCertsDefault     *bool
		allowInvalidCertsHardDefault *bool
		wantBool                     bool
	}{
		"root overrides all": {wantBool: true, allowInvalidCertsRoot: boolPtr(true),
			allowInvalidCertsMain: boolPtr(false), allowInvalidCertsDefault: boolPtr(false), allowInvalidCertsHardDefault: boolPtr(false)},
		"main overrides default+hardDefault": {wantBool: true, allowInvalidCertsRoot: nil,
			allowInvalidCertsMain: boolPtr(true), allowInvalidCertsDefault: boolPtr(false), allowInvalidCertsHardDefault: boolPtr(false)},
		"default overrides hardDefault": {wantBool: true, allowInvalidCertsRoot: nil, allowInvalidCertsMain: nil,
			allowInvalidCertsDefault: boolPtr(false), allowInvalidCertsHardDefault: boolPtr(false)},
		"hardDefault is last resort": {wantBool: true, allowInvalidCertsRoot: nil, allowInvalidCertsMain: nil, allowInvalidCertsDefault: nil,
			allowInvalidCertsHardDefault: boolPtr(true)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			slice["0"].AllowInvalidCerts = tc.allowInvalidCertsRoot
			slice["0"].Main.AllowInvalidCerts = tc.allowInvalidCertsMain
			slice["0"].Defaults.AllowInvalidCerts = tc.allowInvalidCertsDefault
			slice["0"].HardDefaults.AllowInvalidCerts = tc.allowInvalidCertsHardDefault

			// WHEN GetAllowInvalidCerts is called
			got := slice["0"].GetAllowInvalidCerts()

			// THEN the function returns the correct result
			if got != tc.wantBool {
				t.Errorf("%s:\nwant: %t\ngot:  %t",
					name, tc.wantBool, got)
			}
		})
	}
}

func TestGetDelayDuration(t *testing.T) {
	// GIVEN a non-nil Slice, matching mains, defaults and hardDefaults
	serviceID, slice, serviceStatus, mains, defaults, hardDefaults := testInitWithNonNilAndVars()
	slice.Init(nil, &serviceID, &serviceStatus, &mains, &defaults, &hardDefaults, nil, nil)

	// WHEN Delay is "X" and GetDelayDuration is called
	slice["0"].Delay = "1s"
	wanted, _ := time.ParseDuration(slice["0"].Delay)
	got := slice["0"].GetDelayDuration()

	// THEN the function returns the X as a time.Duration
	if got != wanted {
		t.Errorf("GetDelayDuration - wanted %s, got %s",
			wanted, got)
	}
}

func TestGetDesiredStatusCode(t *testing.T) {
	// GIVEN a non-nil Slice, matching mains, defaults and hardDefaults
	serviceID, slice, serviceStatus, mains, defaults, hardDefaults := testInitWithNonNilAndVars()
	slice.Init(nil, &serviceID, &serviceStatus, &mains, &defaults, &hardDefaults, nil, nil)

	// WHEN DesiredStatusCode is 1 and GetDesiredStatusCode is called
	wanted := 1
	hardDefaults.DesiredStatusCode = &wanted
	got := slice["0"].GetDesiredStatusCode()

	// THEN the function returns the hardDefault
	if got != wanted {
		t.Errorf("GetDesiredStatusCode - wanted %d, got %d",
			wanted, got)
	}
}

func TestGetFailStatus(t *testing.T) {
	// GIVEN a non-nil Slice, matching mains, defaults and hardDefaults
	serviceID, slice, serviceStatus, mains, defaults, hardDefaults := testInitWithNonNilAndVars()
	slice.Init(nil, &serviceID, &serviceStatus, &mains, &defaults, &hardDefaults, nil, nil)
	tests := map[string]struct {
		status *bool
	}{
		"nil":   {status: nil},
		"false": {status: boolPtr(false)},
		"true":  {status: boolPtr(true)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			(*slice["0"].Failed)["0"] = tc.status

			// WHEN GetFailStatus is called
			got := slice["0"].GetFailStatus()

			// THEN the function returns the correct result
			g := "<nil>"
			if got != nil {
				g = fmt.Sprint(got)
			}
			want := "<nil>"
			if tc.status != nil {
				want = fmt.Sprint(tc.status)
			}
			if g != want {
				t.Errorf("%s:\nwant: %s\ngot:  %s",
					name, want, g)
			}
		})
	}
}

func TestGetMaxTries(t *testing.T) {
	// GIVEN a non-nil Slice, matching mains, defaults and hardDefaults
	serviceID, slice, serviceStatus, mains, defaults, hardDefaults := testInitWithNonNilAndVars()
	slice.Init(nil, &serviceID, &serviceStatus, &mains, &defaults, &hardDefaults, nil, nil)

	// WHEN MaxTries is X and GetMaxTries is called
	wanted := uint(4)
	slice["0"].MaxTries = &wanted
	got := slice["0"].GetMaxTries()

	// THEN the function returns X
	if got != wanted {
		t.Errorf("GetMaxTries - wanted %d, got %d",
			wanted, got)
	}
}

func TestGetType(t *testing.T) {
	// GIVEN a non-nil Slice, matching mains, defaults and hardDefaults
	serviceID, slice, serviceStatus, mains, defaults, hardDefaults := testInitWithNonNilAndVars()
	slice.Init(nil, &serviceID, &serviceStatus, &mains, &defaults, &hardDefaults, nil, nil)

	// WHEN Type is "X" and GetType is called
	wanted := "gitlab"
	slice["0"].Type = &wanted
	got := slice["0"].GetType()

	// THEN the function returns "X"
	if got != wanted {
		t.Errorf("GetType - wanted %s, got %s",
			wanted, got)
	}
}

func TestGetSecret(t *testing.T) {
	// GIVEN a non-nil Slice, matching mains, defaults and hardDefaults
	serviceID, slice, serviceStatus, mains, defaults, hardDefaults := testInitWithNonNilAndVars()
	slice.Init(nil, &serviceID, &serviceStatus, &mains, &defaults, &hardDefaults, nil, nil)

	// WHEN Secret is "X" and GetSecret is called
	wanted := "secret"
	slice["0"].Secret = &wanted
	got := slice["0"].GetSecret()

	// THEN the function returns "X"
	if *got != wanted {
		t.Errorf("GetSecret - wanted %s, got %s",
			wanted, *got)
	}
}

func TestGetSilentFailsWhenTrue(t *testing.T) {
	// GIVEN a non-nil Slice, matching mains, defaults and hardDefaults with SilentFails true
	serviceID, slice, serviceStatus, mains, defaults, hardDefaults := testInitWithNonNilAndVars()
	slice.Init(nil, &serviceID, &serviceStatus, &mains, &defaults, &hardDefaults, nil, nil)

	// WHEN GetSilentFails is called
	wanted := true
	slice["0"].SilentFails = &wanted
	got := slice["0"].GetSilentFails()

	// THEN the function returns true
	if got != wanted {
		t.Errorf("GetSilentFails - wanted %t, got %t",
			wanted, got)
	}
}

func TestGetSilentFailsWhenFalse(t *testing.T) {
	// GIVEN a non-nil Slice, matching mains, defaults and hardDefaults with SilentFails false
	serviceID, slice, serviceStatus, mains, defaults, hardDefaults := testInitWithNonNilAndVars()
	slice.Init(nil, &serviceID, &serviceStatus, &mains, &defaults, &hardDefaults, nil, nil)

	// WHEN SilentFails is false and GetSilentFails is called
	wanted := false
	slice["0"].SilentFails = &wanted
	got := slice["0"].GetSilentFails()

	// THEN the function returns false
	if got != wanted {
		t.Errorf("GetSilentFails - wanted %t, got %t",
			wanted, got)
	}
}

func TestGetURLWithNoTemplating(t *testing.T) {
	// GIVEN a non-nil Slice, matching mains, defaults and hardDefaults
	serviceID, slice, serviceStatus, mains, defaults, hardDefaults := testInitWithNonNilAndVars()
	slice.Init(nil, &serviceID, &serviceStatus, &mains, &defaults, &hardDefaults, nil, nil)

	// WHEN URL requires no Jinja templating and GetURL is called
	wanted := "https://example.com"
	slice["0"].URL = &wanted
	got := slice["0"].GetURL()

	// THEN the function returns the hardDefault
	if *got != wanted {
		t.Errorf("GetURL - wanted %s, got %s",
			wanted, *got)
	}
}

func TestGetURLWithTemplatingWorks(t *testing.T) {
	// GIVEN a non-nil Slice, matching mains, defaults and hardDefaults
	serviceID, slice, serviceStatus, mains, defaults, hardDefaults := testInitWithNonNilAndVars()
	slice.Init(nil, &serviceID, &serviceStatus, &mains, &defaults, &hardDefaults, nil, nil)

	// WHEN URL requires Jinja templating and GetURL is called
	template := "https://example.com{% if 'a' == 'a' %}/{{ version }}{% endif %}{% if 'a' == 'b' %}foo{% endif %}"
	slice["0"].URL = &template
	got := slice["0"].GetURL()
	wanted := "https://example.com/status test"

	// THEN the function returns the default
	if *got != wanted {
		t.Errorf("GetURL - wanted %s, got %s",
			wanted, *got)
	}
}
func TestGetURLWithTemplatingUnchangesTemplate(t *testing.T) {
	// GIVEN a non-nil Slice, matching mains, defaults and hardDefaults
	serviceID, slice, serviceStatus, mains, defaults, hardDefaults := testInitWithNonNilAndVars()
	slice.Init(nil, &serviceID, &serviceStatus, &mains, &defaults, &hardDefaults, nil, nil)

	// WHEN URL requires Jinja templating and GetURL is called
	template := "https://example.com{% if 'a' == 'a' %}/{{ version }}{% endif %}{% if 'a' == 'b' %}foo{% endif %}"
	slice["0"].URL = &template
	got := slice["0"].GetURL()
	wanted := "https://example.com/status test"

	// THEN the template stays intact
	if *got != wanted {
		t.Errorf("GetURL modified the template - wanted %s, got %s",
			wanted, *got)
	}
}

func TestGetRequestGitHubInvalidURL(t *testing.T) {
	// GIVEN type github and an invalid URL
	whType := "github"
	whURL := "invalid://	test"
	whSecret := "secret"
	wh := WebHook{
		Type:         &whType,
		URL:          &whURL,
		Secret:       &whSecret,
		Main:         &WebHook{},
		Defaults:     &WebHook{},
		HardDefaults: &WebHook{},
	}

	// WHEN GetRequest is called
	req := wh.GetRequest()

	// THEN req is nil
	if req != nil {
		t.Error("Invalid URL produced a non-nil http.Request")
	}
}

func TestGetRequestGitHubValidURL(t *testing.T) {
	// GIVEN type github and a valid URL
	whType := "github"
	whURL := "https://test"
	whSecret := "secret"
	wh := WebHook{
		Type:         &whType,
		URL:          &whURL,
		Secret:       &whSecret,
		Main:         &WebHook{},
		Defaults:     &WebHook{},
		HardDefaults: &WebHook{},
		CustomHeaders: &map[string]string{
			"foo": "bar",
		},
		ServiceStatus: &service_status.Status{},
	}

	// WHEN GetRequest is called
	req := wh.GetRequest()

	// THEN req is valid
	if req == nil {
		t.Error("Invalid URL produced a non-nil http.Request")
	}
	key := "Content-Type"
	want := "application/json"
	if req.Header[key][0] != want {
		t.Errorf("%s should have been %s, got %s",
			key, want, req.Header[key][0])
	}
	key = "foo"
	want = "bar"
	if req.Header[key][0] != want {
		t.Errorf("%s should have been %s, got %s",
			key, want, req.Header[key][0])
	}
	key = "X-Github-Event"
	want = "push"
	if req.Header[key][0] != want {
		t.Errorf("%s should have been %s, got %s",
			key, want, req.Header[key][0])
	}
}

func TestGetRequestGitLabInvalidURL(t *testing.T) {
	// GIVEN type gitlab and an invalid URL
	whType := "gitlab"
	whURL := "invalid://	test"
	whSecret := "secret"
	wh := WebHook{
		Type:         &whType,
		URL:          &whURL,
		Secret:       &whSecret,
		Main:         &WebHook{},
		Defaults:     &WebHook{},
		HardDefaults: &WebHook{},
	}

	// WHEN GetRequest is called
	req := wh.GetRequest()

	// THEN req is nil
	if req != nil {
		t.Error("Invalid URL produced a non-nil http.Request")
	}
}

func TestGetRequestGitLabValidURL(t *testing.T) {
	// GIVEN type gitlab and a valid URL
	whType := "gitlab"
	whURL := "https://test"
	whSecret := "secret"
	wh := WebHook{
		Type:         &whType,
		URL:          &whURL,
		Secret:       &whSecret,
		Main:         &WebHook{},
		Defaults:     &WebHook{},
		HardDefaults: &WebHook{},
		CustomHeaders: &map[string]string{
			"foo": "bar",
		},
		ServiceStatus: &service_status.Status{},
	}

	// WHEN GetRequest is called
	req := wh.GetRequest()

	// THEN req is valid
	if req == nil {
		t.Error("Invalid URL produced a non-nil http.Request")
	}
	key := "Content-Type"
	want := "application/x-www-form-urlencoded"
	if req.Header[key][0] != want {
		t.Errorf("%s should have been %s, got %s",
			key, want, req.Header[key][0])
	}
	if req.URL.RawQuery == "" {
		t.Error("RawQuery was empty when it should have been encoded data")
	}
}

func TestSetFailStatus(t *testing.T) {
	// GIVEN a non-nil Slice, matching mains, defaults and hardDefaults
	serviceID, slice, serviceStatus, mains, defaults, hardDefaults := testInitWithNonNilAndVars()
	slice.Init(nil, &serviceID, &serviceStatus, &mains, &defaults, &hardDefaults, nil, nil)
	tests := map[string]struct {
		status *bool
	}{
		"nil":   {status: nil},
		"false": {status: boolPtr(false)},
		"true":  {status: boolPtr(true)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// WHEN SetFailStatus is called
			slice["0"].SetFailStatus(tc.status)

			// THEN the fail status is correctly set for the WebHook
			got := (*slice["0"].Failed)["0"]
			g := "<nil>"
			if got != nil {
				g = fmt.Sprint(got)
			}
			want := "<nil>"
			if tc.status != nil {
				want = fmt.Sprint(tc.status)
			}
			if g != want {
				t.Errorf("%s:\nwant: %s\ngot:  %s",
					name, want, g)
			}
		})
	}
}

func TestIsRunnableTrue(t *testing.T) {
	// GIVEN a WebHook with NextRunnable before the current time
	wh := WebHook{
		NextRunnable: time.Now().UTC().Add(-time.Minute),
	}

	// WHEN IsRunnable is called on it
	ranAt := time.Now().UTC()
	got := wh.IsRunnable()

	// THEN true was returned
	want := true
	if got != want {
		t.Fatalf("IsRunnable was ran at\n%s with NextRunnable\n%s. Expected %t, got %t",
			ranAt, wh.NextRunnable, want, got)
	}
}

func TestIsRunnableFalse(t *testing.T) {
	// GIVEN a WebHook with NextRunnable after the current time
	wh := WebHook{
		NextRunnable: time.Now().UTC().Add(time.Minute),
	}

	// WHEN IsRunnable is called on it
	ranAt := time.Now().UTC()
	got := wh.IsRunnable()

	// THEN false was returned
	want := false
	if got != want {
		t.Fatalf("IsRunnable was ran at\n%s with NextRunnable\n%s. Expected %t, got %t",
			ranAt, wh.NextRunnable, want, got)
	}
}

func TestSetNextRunnableOfPass(t *testing.T) {
	// GIVEN a WebHook that passed
	whID := "test"
	wh := WebHook{
		ID:           whID,
		Type:         stringPtr("gitlab"),
		URL:          stringPtr("https://test"),
		Secret:       stringPtr("secret"),
		Main:         &WebHook{},
		Defaults:     &WebHook{},
		HardDefaults: &WebHook{},
		CustomHeaders: &map[string]string{
			"foo": "bar",
		},
		Failed:         &map[string]*bool{whID: boolPtr(false)},
		ServiceStatus:  &service_status.Status{},
		ParentInterval: stringPtr("11m"),
	}

	// WHEN SetNextRunnable is called on it
	wh.SetNextRunnable(false, false)

	// THEN MextRunnable is set to ~2*ParentInterval
	now := time.Now().UTC()
	got := wh.NextRunnable
	parentInterval, _ := time.ParseDuration(*wh.ParentInterval)
	wantMin := now.Add(2 * parentInterval).Add(-1 * time.Second)
	wantMax := now.Add(2 * parentInterval).Add(1 * time.Second)
	if got.Before(wantMin) || got.After(wantMax) {
		t.Errorf("Expected between\n%s and\n%s, not\n%s",
			wantMin, wantMax, got)
	}
}

func TestSetNextRunnableOfFail(t *testing.T) {
	// GIVEN a WebHook that failed
	whID := "test"
	wh := WebHook{
		ID:           whID,
		Type:         stringPtr("gitlab"),
		URL:          stringPtr("https://test"),
		Secret:       stringPtr("secret"),
		Main:         &WebHook{},
		Defaults:     &WebHook{},
		HardDefaults: &WebHook{},
		CustomHeaders: &map[string]string{
			"foo": "bar",
		},
		Failed:        &map[string]*bool{whID: boolPtr(false)},
		ServiceStatus: &service_status.Status{},
	}

	// WHEN SetNextRunnable is called on it
	wh.SetNextRunnable(false, false)

	// THEN MextRunnable is set to 15s
	now := time.Now().UTC()
	got := wh.NextRunnable
	wantMin := now.Add(14 * time.Second)
	wantMax := now.Add(16 * time.Second)
	if got.Before(wantMin) || got.After(wantMax) {
		t.Errorf("Expected between\n%s and\n%s, not\n%s",
			wantMin, wantMax, got)
	}
}

func TestSetNextRunnableOfNotRun(t *testing.T) {
	// GIVEN a WebHook that hasn't been sent
	whID := "test"
	wh := WebHook{
		ID:           whID,
		Type:         stringPtr("gitlab"),
		URL:          stringPtr("https://test"),
		Secret:       stringPtr("secret"),
		Main:         &WebHook{},
		Defaults:     &WebHook{},
		HardDefaults: &WebHook{},
		CustomHeaders: &map[string]string{
			"foo": "bar",
		},
		Failed:        &map[string]*bool{whID: boolPtr(false)},
		ServiceStatus: &service_status.Status{},
	}

	// WHEN SetNextRunnable is called on it
	wh.SetNextRunnable(false, false)

	// THEN MextRunnable is set to 15s
	now := time.Now().UTC()
	got := wh.NextRunnable
	wantMin := now.Add(14 * time.Second)
	wantMax := now.Add(16 * time.Second)
	if got.Before(wantMin) || got.After(wantMax) {
		t.Errorf("Expected between\n%s and\n%s, not\n%s",
			wantMin, wantMax, got)
	}
}

func TestSetNextRunnableWithDelay(t *testing.T) {
	// GIVEN a WebHook that hasn't been sent
	whID := "test"
	wh := WebHook{
		ID:           whID,
		Type:         stringPtr("gitlab"),
		URL:          stringPtr("https://test"),
		Secret:       stringPtr("secret"),
		Main:         &WebHook{},
		Defaults:     &WebHook{},
		HardDefaults: &WebHook{},
		CustomHeaders: &map[string]string{
			"foo": "bar",
		},
		Delay:          "5m",
		Failed:         &map[string]*bool{whID: boolPtr(false)},
		ServiceStatus:  &service_status.Status{},
		ParentInterval: stringPtr("1h"),
	}

	// WHEN SetNextRunnable is called on a webhook that hasn't been run
	wh.SetNextRunnable(true, false)

	// THEN MextRunnable is set to 15s
	now := time.Now().UTC()
	got := wh.NextRunnable
	wantDelay, _ := time.ParseDuration(wh.Delay)
	wantMin := now.Add(14 * time.Second).Add(wantDelay)
	wantMax := now.Add(16 * time.Second).Add(wantDelay)
	if got.Before(wantMin) || got.After(wantMax) {
		t.Errorf("Expected between\n%s and\n%s, not\n%s",
			wantMin, wantMax, got)
	}
}

func TestSetNextRunnableOfSending(t *testing.T) {
	// GIVEN a WebHook that hasn't been sent
	whMaxTries := uint(2)
	whID := "test"
	wh := WebHook{
		ID:           whID,
		Type:         stringPtr("gitlab"),
		URL:          stringPtr("https://test"),
		Secret:       stringPtr("secret"),
		Main:         &WebHook{},
		Defaults:     &WebHook{},
		HardDefaults: &WebHook{},
		CustomHeaders: &map[string]string{
			"foo": "bar",
		},
		Failed:         &map[string]*bool{whID: boolPtr(false)},
		MaxTries:       &whMaxTries,
		ServiceStatus:  &service_status.Status{},
		ParentInterval: stringPtr("11m"),
	}

	// WHEN SetNextRunnable is called on a webhook that's sending
	wh.SetNextRunnable(false, true)

	// THEN MextRunnable is set to 15s+3*MaxTries
	now := time.Now().UTC()
	got := wh.NextRunnable
	maxTriesIncrement := uint(3)
	wantDelay := time.Duration(wh.GetMaxTries()*maxTriesIncrement) * time.Second
	wantMin := now.Add(14 * time.Second).Add(wantDelay)
	wantMax := now.Add(16 * time.Second).Add(wantDelay)
	if got.Before(wantMin) || got.After(wantMax) {
		t.Errorf("Expected between\n%s and\n%s, not\n%s",
			wantMin, wantMax, got)
	}
}

func TestResetFailsWithNilSlice(t *testing.T) {
	// GIVEN a nil Slice
	var slice *Slice
	// WHEN ResetFails is run
	// THEN it exits successfully
	slice.ResetFails()
}

func TestResetFailsWithSlice(t *testing.T) {
	// GIVEN a Slice with WebHooks that have failed
	slice := Slice{
		"0": &WebHook{},
		"1": &WebHook{},
		"2": &WebHook{},
		"3": &WebHook{},
		"4": &WebHook{},
	}
	failed0 := true
	(*(*slice["0"]).Failed)["0"] = &failed0
	failed1 := true
	(*(*slice["1"]).Failed)["1"] = &failed1
	failed2 := true
	(*(*slice["2"]).Failed)["2"] = &failed2
	failed3 := true
	(*(*slice["3"]).Failed)["3"] = &failed3
	failed4 := true
	(*(*slice["4"]).Failed)["4"] = &failed4

	// WHEN ResetFails is called
	slice.ResetFails()

	// THEN all the fails become nil and the count stays the same
	for i := range slice {
		if (*(*slice[i]).Failed)[i] != nil {
			t.Errorf("Reset failed, got %t",
				*(*(*slice[i]).Failed)[i])
		}
	}
}
