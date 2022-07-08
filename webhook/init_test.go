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
		t.Errorf("Got len %d, when expecting %d", got, want)
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
		t.Errorf("Slice['0'] shouldn't be %v still", slice["0"])
	}
}

func TestSliceInitWithNonNil(t *testing.T) {
	// GIVEN a non-nil Slice with everything else nil
	id0 := "0"
	slice := Slice{
		"0": &WebHook{
			ID: &id0,
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
		t.Errorf("Got len %d, when expecting %d", got, want)
	}
}

func testInitWithNonNilAndVars() (string, Slice, service_status.Status, Slice, WebHook, WebHook) {
	id0 := "0"
	id1 := "1"
	slice := Slice{
		"0": &WebHook{
			ID: &id0,
		},
		"1": &WebHook{
			ID: &id1,
		},
	}
	serviceID := "id_test"
	serviceStatus := service_status.Status{LatestVersion: "status test"}
	mainSecret0 := "main0"
	mainSecret1 := "main1"
	mains := Slice{
		"0": &WebHook{
			ID:     &id0,
			Secret: &mainSecret0,
		},
		"1": &WebHook{
			ID:     &id1,
			Secret: &mainSecret1,
		},
	}
	defaultURL := "default"
	defaults := WebHook{
		ID:     &id0,
		Secret: &defaultURL,
	}
	hardDefaultDelay := "1s"
	hardDefaults := WebHook{
		Delay: &hardDefaultDelay,
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
			t.Errorf("Main not handed to %s", key)
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
			t.Errorf("Defaults not handed to %s", key)
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
			t.Errorf("HardDefaults not handed to %s", key)
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
			t.Errorf("Notifiers %v weren't handed to %s", notifiers, key)
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

func TestGetAllowInvalidCertsWhenTrue(t *testing.T) {
	// GIVEN a non-nil Slice, matching mains, defaults and hardDefaults
	serviceID, slice, serviceStatus, mains, defaults, hardDefaults := testInitWithNonNilAndVars()
	slice.Init(nil, &serviceID, &serviceStatus, &mains, &defaults, &hardDefaults, nil, nil)

	// WHEN AllowInvalidCerts is true and GetAllowInvalidCerts is called
	wanted := true
	slice["0"].AllowInvalidCerts = &wanted
	got := slice["0"].GetAllowInvalidCerts()

	// THEN the function returns true
	if got != wanted {
		t.Errorf("GetAllowInvalidCerts - wanted %t, got %t", wanted, got)
	}
}

func TestGetAllowInvalidCertsWhenFalse(t *testing.T) {
	// GIVEN a non-nil Slice, matching mains, defaults and hardDefaults
	serviceID, slice, serviceStatus, mains, defaults, hardDefaults := testInitWithNonNilAndVars()
	slice.Init(nil, &serviceID, &serviceStatus, &mains, &defaults, &hardDefaults, nil, nil)

	// WHEN AllowInvalidCerts is false and GetAllowInvalidCerts is called
	wanted := false
	slice["0"].AllowInvalidCerts = &wanted
	got := slice["0"].GetAllowInvalidCerts()

	// THEN the function returns false
	if got != wanted {
		t.Errorf("GetAllowInvalidCerts - wanted %t, got %t", wanted, got)
	}
}

func TestGetDelayDuration(t *testing.T) {
	// GIVEN a non-nil Slice, matching mains, defaults and hardDefaults
	serviceID, slice, serviceStatus, mains, defaults, hardDefaults := testInitWithNonNilAndVars()
	slice.Init(nil, &serviceID, &serviceStatus, &mains, &defaults, &hardDefaults, nil, nil)

	// WHEN Delay is "X" and GetDelayDuration is called
	duration := "1s"
	slice["0"].Delay = &duration
	wanted, _ := time.ParseDuration(duration)
	got := slice["0"].GetDelayDuration()

	// THEN the function returns the X as a time.Duration
	if got != wanted {
		t.Errorf("GetDelayDuration - wanted %s, got %s", wanted, got)
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
		t.Errorf("GetDesiredStatusCode - wanted %d, got %d", wanted, got)
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
		t.Errorf("GetMaxTries - wanted %d, got %d", wanted, got)
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
		t.Errorf("GetType - wanted %s, got %s", wanted, got)
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
		t.Errorf("GetSecret - wanted %s, got %s", wanted, *got)
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
		t.Errorf("GetSilentFails - wanted %t, got %t", wanted, got)
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
		t.Errorf("GetSilentFails - wanted %t, got %t", wanted, got)
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
		t.Errorf("GetURL - wanted %s, got %s", wanted, *got)
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
		t.Errorf("GetURL - wanted %s, got %s", wanted, *got)
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
		t.Errorf("GetURL modified the template - wanted %s, got %s", wanted, *got)
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
		t.Errorf("%s should have been %s, got %s", key, want, req.Header[key][0])
	}
	key = "foo"
	want = "bar"
	if req.Header[key][0] != want {
		t.Errorf("%s should have been %s, got %s", key, want, req.Header[key][0])
	}
	key = "X-Github-Event"
	want = "push"
	if req.Header[key][0] != want {
		t.Errorf("%s should have been %s, got %s", key, want, req.Header[key][0])
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
		t.Errorf("%s should have been %s, got %s", key, want, req.Header[key][0])
	}
	if req.URL.RawQuery == "" {
		t.Error("RawQuery was empty when it should have been encoded data")
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
	whType := "gitlab"
	whURL := "https://test"
	whSecret := "secret"
	failed := false
	serviceInterval := "11m"
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
		Failed:         &failed,
		ServiceStatus:  &service_status.Status{},
		ParentInterval: &serviceInterval,
	}

	// WHEN SetNextRunnable is called on a webhook that ran successfully
	wh.SetNextRunnable()

	// THEN MextRunnable is set to ~2*ParentInterval
	now := time.Now().UTC()
	got := wh.NextRunnable
	parentInterval, _ := time.ParseDuration(*wh.ParentInterval)
	wantMin := now.Add(2 * parentInterval).Add(-1 * time.Second)
	wantMax := now.Add(2 * parentInterval).Add(1 * time.Second)
	if got.Before(wantMin) || got.After(wantMax) {
		t.Errorf("Expected between %s and %s, not %s",
			wantMin, wantMax, got)
	}
}

func TestSetNextRunnableOfFail(t *testing.T) {
	// GIVEN a WebHook that failed
	whType := "gitlab"
	whURL := "https://test"
	whSecret := "secret"
	failed := true
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
		Failed:        &failed,
		ServiceStatus: &service_status.Status{},
	}

	// WHEN SetNextRunnable is called on a command index that failed running
	wh.SetNextRunnable()

	// THEN MextRunnable is set to 15s
	now := time.Now().UTC()
	got := wh.NextRunnable
	wantMin := now.Add(14 * time.Second)
	wantMax := now.Add(16 * time.Second)
	if got.Before(wantMin) || got.After(wantMax) {
		t.Errorf("Expected between %s and %s, not %s",
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
	(*slice["0"]).Failed = &failed0
	failed1 := true
	(*slice["1"]).Failed = &failed1
	failed2 := true
	(*slice["2"]).Failed = &failed2
	failed3 := true
	(*slice["3"]).Failed = &failed3
	failed4 := true
	(*slice["4"]).Failed = &failed4

	// WHEN ResetFails is called
	slice.ResetFails()

	// THEN all the fails become nil and the count stays the same
	for i := range slice {
		if (*slice[i]).Failed != nil {
			t.Errorf("Reset failed, got %t", *(*slice[i]).Failed)
		}
	}
}
