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

package webhook

import (

	//#nosec G505 -- GitHub's X-Hub-Signature uses SHA-1

	"testing"
	"time"

	service_status "github.com/release-argus/Argus/service/status"
)

func TestInit(t *testing.T) {
	{ // GIVEN a nil Slice
		var slice Slice
		// WHEN Init is called
		slice.Init(nil, nil, nil, nil, nil, nil, nil)
		// THEN the function exits without defining any of the vars
		got := len(slice)
		want := 0
		if got != want {
			t.Fatalf("Got len %d, when expecting %d", got, want)
		}
	}
	{ // GIVEN a non-nil Slice with everything else nil
		id0 := "0"
		slice := Slice{
			"0": &WebHook{
				ID: &id0,
			},
		}
		// Won't ever be nil, so don't make it nil in tests
		serviceID := ""
		// WHEN Init is called
		slice.Init(nil, &serviceID, nil, nil, nil, nil, nil)
		// THEN the function exits without defining any of the vars
		got := len(slice)
		want := 1
		if got != want {
			t.Fatalf("Got len %d, when expecting %d", got, want)
		}
	}

	{ // GIVEN a non-nil Slice, matching mains, defaults and hardDefaults
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
		HardDefaults := WebHook{
			Delay: &hardDefaultDelay,
		}
		// WHEN Init is called
		slice.Init(nil, &serviceID, &serviceStatus, &mains, &defaults, &HardDefaults, nil)
		{ // THEN the vars are handed out correctly
			for key := range slice {
				if slice[key].Main != mains[key] {
					t.Fatalf("Main not handed to %s", key)
				}
				if *slice[key].Defaults != defaults {
					t.Fatalf("Defaults not handed to %s", key)
				}
				if *slice[key].HardDefaults != HardDefaults {
					t.Fatalf("HardDefaults not handed to %s", key)
				}
				if slice[key].Notifiers.Shoutrrr != nil {
					t.Fatalf("Notifiers handed to %s (when non should have been)", key)
				}
			}
		}
		{ // WHEN initMetrics is called
			slice["0"].initMetrics("foo")
			// THEN the function runs without error
		}
		{
			// WHEN AllowInvalidCerts is true and GetAllowInvalidCerts is called
			wanted := true
			slice["0"].AllowInvalidCerts = &wanted
			got := slice["0"].GetAllowInvalidCerts()
			// THEN the function returns true
			if got != wanted {
				t.Fatalf("GetAllowInvalidCerts - wanted %t, got %t", wanted, got)
			}
			// WHEN AllowInvalidCerts is false and GetAllowInvalidCerts is called
			wanted = false
			slice["0"].AllowInvalidCerts = &wanted
			got = slice["0"].GetAllowInvalidCerts()
			// THEN the function returns false
			if got != wanted {
				t.Fatalf("GetAllowInvalidCerts - wanted %t, got %t", wanted, got)
			}
		}
		{ // WHEN Delay is "X" and GetDelayDuration is called
			duration := "1s"
			slice["0"].Delay = &duration
			wanted, _ := time.ParseDuration(duration)
			got := slice["0"].GetDelayDuration()
			// THEN the function returns the X as a time.Duration
			if got != wanted {
				t.Fatalf("GetDelayDuration - wanted %s, got %s", wanted, got)
			}
		}
		{ // WHEN DesiredStatusCode is 1 and GetDesiredStatusCode is called
			wanted := 1
			HardDefaults.DesiredStatusCode = &wanted
			got := slice["0"].GetDesiredStatusCode()
			// THEN the function returns the hardDefault
			if got != wanted {
				t.Fatalf("GetDesiredStatusCode - wanted %d, got %d", wanted, got)
			}
		}
		{ // WHEN MaxTries is X and GetMaxTries is called
			wanted := uint(4)
			slice["0"].MaxTries = &wanted
			got := slice["0"].GetMaxTries()
			// THEN the function returns X
			if got != wanted {
				t.Fatalf("GetMaxTries - wanted %d, got %d", wanted, got)
			}
		}
		{ // WHEN Type is "X" and GetType is called
			wanted := "gitlab"
			slice["0"].Type = &wanted
			got := slice["0"].GetType()
			// THEN the function returns "X"
			if got != wanted {
				t.Fatalf("GetType - wanted %s, got %s", wanted, got)
			}
		}
		{ // WHEN Secret is "X" and GetSecret is called
			wanted := "secret"
			slice["0"].Secret = &wanted
			got := slice["0"].GetSecret()
			// THEN the function returns "X"
			if *got != wanted {
				t.Fatalf("GetSecret - wanted %s, got %s", wanted, *got)
			}
		}
		{
			// WHEN SilentFails is true and GetSilentFails is called
			wanted := true
			slice["0"].SilentFails = &wanted
			got := slice["0"].GetSilentFails()
			// THEN the function returns true
			if got != wanted {
				t.Fatalf("GetSilentFails - wanted %t, got %t", wanted, got)
			}
			// WHEN SilentFails is false and GetSilentFails is called
			wanted = false
			slice["0"].SilentFails = &wanted
			got = slice["0"].GetSilentFails()
			// THEN the function returns false
			if got != wanted {
				t.Fatalf("GetSilentFails - wanted %t, got %t", wanted, got)
			}
		}
		{
			// WHEN URL requires no Jinja templating and GetURL is called
			wanted := "https://example.com"
			slice["0"].URL = &wanted
			got := slice["0"].GetURL()
			// THEN the function returns the hardDefault
			if *got != wanted {
				t.Fatalf("GetURL - wanted %s, got %s", wanted, *got)
			}
			// WHEN URL requires Jinja templating and GetURL is called
			template := "https://example.com{% if 'a' == 'a' %}/{{ version }}{% endif %}{% if 'a' == 'b' %}foo{% endif %}"
			slice["0"].URL = &template
			got = slice["0"].GetURL()
			wanted = "https://example.com/status test"
			// THEN the function returns the default
			if *got != wanted {
				t.Fatalf("GetURL - wanted %s, got %s", wanted, *got)
			}
			// AND the template stays intact
			got = slice["0"].URL
			if *got != wanted {
				t.Fatalf("GetURL modified the template - wanted %s, got %s", wanted, *got)
			}
		}
	}
}

func TestGetRequest(t *testing.T) {
	{ // GIVEN type github and an invalid URL
		whType := "github"
		whURL := "invalid://	test"
		whSecret := "secret"
		webhook := WebHook{
			Type:         &whType,
			URL:          &whURL,
			Secret:       &whSecret,
			Main:         &WebHook{},
			Defaults:     &WebHook{},
			HardDefaults: &WebHook{},
		}
		// WHEN GetRequest is called
		req := webhook.GetRequest()
		// THEN req is nil
		if req != nil {
			t.Fatal("Invalid URL produced a non-nil http.Request")
		}
	}
	{ // GIVEN type github and a valid URL
		whType := "github"
		whURL := "https://test"
		whSecret := "secret"
		webhook := WebHook{
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
		req := webhook.GetRequest()
		// THEN req is valid
		if req == nil {
			t.Fatal("Invalid URL produced a non-nil http.Request")
		}
		key := "Content-Type"
		want := "application/json"
		if req.Header[key][0] != want {
			t.Fatalf("%s should have been %s, got %s", key, want, req.Header[key][0])
		}
		key = "foo"
		want = "bar"
		if req.Header[key][0] != want {
			t.Fatalf("%s should have been %s, got %s", key, want, req.Header[key][0])
		}
		key = "X-Github-Event"
		want = "push"
		if req.Header[key][0] != want {
			t.Fatalf("%s should have been %s, got %s", key, want, req.Header[key][0])
		}
	}

	{ // GIVEN type gitlab and an invalid URL
		whType := "gitlab"
		whURL := "invalid://	test"
		whSecret := "secret"
		webhook := WebHook{
			Type:         &whType,
			URL:          &whURL,
			Secret:       &whSecret,
			Main:         &WebHook{},
			Defaults:     &WebHook{},
			HardDefaults: &WebHook{},
		}
		// WHEN GetRequest is called
		req := webhook.GetRequest()
		// THEN req is nil
		if req != nil {
			t.Fatal("Invalid URL produced a non-nil http.Request")
		}
	}
	{ // GIVEN type gitlab and a valid URL
		whType := "gitlab"
		whURL := "https://test"
		whSecret := "secret"
		webhook := WebHook{
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
		req := webhook.GetRequest()
		// THEN req is valid
		if req == nil {
			t.Fatal("Invalid URL produced a non-nil http.Request")
		}
		key := "Content-Type"
		want := "application/x-www-form-urlencoded"
		if req.Header[key][0] != want {
			t.Fatalf("%s should have been %s, got %s", key, want, req.Header[key][0])
		}
		if req.URL.RawQuery == "" {
			t.Fatal("RawQuery was empty when it should have been encoded data")
		}
	}
}

func TestResetFails(t *testing.T) {
	{ // GIVEN a nil Slice
		var slice Slice
		// WHEN ResetFails is run
		// THEN it exits successfully
		slice.ResetFails()
	}

	{ // GIVEN a Slice with WebHooks that have failed
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
				t.Fatalf("Reset failed, got %t", *(*slice[i]).Failed)
			}
		}
	}
}
