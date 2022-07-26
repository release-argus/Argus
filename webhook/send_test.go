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
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
)

func TestTryWithInvalidURL(t *testing.T) {
	// GIVEN a WebHook with an invalid URL
	jLog = utils.NewJLog("WARN", false)
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

	// WHEN try is called on this WebHook
	err := webhook.try(utils.LogFrom{})

	// THEN err is not nil
	if err == nil {
		t.Errorf("try should have failed with invalid %q URL. err is %s",
			whURL, err.Error())
	}
}

func TestTryWithUnknownHostAndAllowInvalidCerts(t *testing.T) {
	// GIVEN a WebHook which allows invalid certs
	jLog = utils.NewJLog("WARN", false)
	whType := "github"
	whURL := "https://test"
	whSecret := "secret"
	whAllowInvalidCerts := true
	webhook := WebHook{
		Type:              &whType,
		URL:               &whURL,
		Secret:            &whSecret,
		AllowInvalidCerts: &whAllowInvalidCerts,
		Main:              &WebHook{},
		Defaults:          &WebHook{},
		HardDefaults:      &WebHook{},
	}

	// WHEN try is called on this WebHook
	err := webhook.try(utils.LogFrom{})

	// THEN err is about undefined host
	startsWith := "Post \"https://test\": dial tcp: lookup test"
	e := utils.ErrorToString(err)
	if !strings.HasPrefix(e, startsWith) {
		t.Errorf("try with %v should have errored starting %q. Got %q",
			webhook, startsWith, e)
	}
}

func TestTryWithUnknownHostAndRejectInvalidCerts(t *testing.T) {
	// GIVEN a WebHook which doesn't allow invalid certs
	jLog = utils.NewJLog("WARN", false)
	whType := "github"
	whURL := "https://test"
	whSecret := "secret"
	whAllowInvalidCerts := false
	webhook := WebHook{
		Type:              &whType,
		URL:               &whURL,
		Secret:            &whSecret,
		AllowInvalidCerts: &whAllowInvalidCerts,
		Main:              &WebHook{},
		Defaults:          &WebHook{},
		HardDefaults:      &WebHook{},
	}

	// WHEN try is called on this WebHook
	err := webhook.try(utils.LogFrom{})

	// THEN err is about undefined host
	startsWith := "Post \"https://test\": dial tcp: lookup test"
	e := utils.ErrorToString(err)
	if !strings.HasPrefix(e, startsWith) {
		t.Errorf("try with %v should have errored starting %q. Got %q",
			webhook, startsWith, e)
	}
}

func TestTryWithKnownHostAndAllowAnyStatusCode(t *testing.T) {
	// GIVEN a WebHook which accepts 2XX response status code
	jLog = utils.NewJLog("WARN", false)
	webhook := testWebHookSuccessful()
	*webhook.DesiredStatusCode = 0

	// WHEN try is called on this WebHook
	err := webhook.try(utils.LogFrom{})

	// THEN err is nil
	if err != nil {
		t.Errorf("try with %v shouldn't have errored %q",
			webhook, err.Error())
	}
}

func TestTryWithKnownHostAndRejectStatusCode(t *testing.T) {
	// GIVEN a WebHook which returns outside the allowed status code range
	jLog = utils.NewJLog("WARN", false)
	webhook := testWebHookSuccessful()
	*webhook.DesiredStatusCode = 50

	attempts := 3
	for attempts != 0 {
		// WHEN try is called on this WebHook
		err := webhook.try(utils.LogFrom{})

		// THEN err is about the status code not matching
		e := utils.ErrorToString(err)
		if strings.Contains(e, "context deadline exceeded") {
			attempts--
			time.Sleep(2 * time.Second)
			continue
		}
		startsWith := fmt.Sprintf("WebHook didn't %d:", *webhook.DesiredStatusCode)
		if !strings.HasPrefix(e, startsWith) {
			t.Errorf("try with %v should have started with %q, but got %q",
				webhook, startsWith, e)
		}
		return
	}
}

func TestWebHookSendFailWithSilentFails(t *testing.T) {
	// GIVEN a WebHook which will err on Send
	jLog = utils.NewJLog("WARN", false)
	webhook := testWebHookFailing()
	*webhook.SilentFails = true

	attempts := 3
	for attempts != 0 {
		// WHEN Send is called on this WebHook
		err := webhook.Send(utils.ServiceInfo{}, false)

		// THEN err is about the failure
		e := utils.ErrorToString(err)
		if strings.Contains(e, "context deadline exceeded") {
			attempts--
			time.Sleep(2 * time.Second)
			continue
		}
		contains := "WebHook didn't 2XX"
		if !strings.Contains(e, contains) {
			t.Errorf("Send with %v should have errored %q. Got %q",
				webhook, contains, e)
		}
		return
	}
}

func TestWebHookSendFailDoesRetry(t *testing.T) {
	// GIVEN a WebHook which will err on Send
	jLog = utils.NewJLog("WARN", false)
	webhook := testWebHookFailing()
	*webhook.MaxTries = uint(2)

	// WHEN Send is called on this WebHook
	start := time.Now().UTC()
	err := webhook.Send(utils.ServiceInfo{}, false)

	// THEN err is about the failure
	errCount := strings.Count(utils.ErrorToString(err), "WebHook didn't 2XX")
	since := time.Since(start)
	if since < 10*time.Second || errCount != int(*webhook.MaxTries) {
		t.Errorf("LastQueried was %v ago, not recent enough!",
			since)
	}
}

func TestWebHookSendFailWithoutSilentFails(t *testing.T) {
	// GIVEN a WebHook which will err on Send
	jLog = utils.NewJLog("WARN", false)
	webhook := testWebHookFailing()
	*webhook.SilentFails = false

	// WHEN Send is called on this WebHook
	err := webhook.Send(utils.ServiceInfo{}, false)

	// THEN err is about the failure
	contains := "WebHook didn't 2XX"
	e := utils.ErrorToString(err)
	if !strings.Contains(e, contains) {
		t.Errorf("Send with %v should have errored %q. Got %q",
			webhook, contains, e)
	}
}

func TestWebHookSendSuccess(t *testing.T) {
	// GIVEN a WebHook which accepts 2XX response status code
	jLog = utils.NewJLog("WARN", false)
	webhook := testWebHookSuccessful()
	*webhook.DesiredStatusCode = 0

	// WHEN try is called on this WebHook
	err := webhook.Send(utils.ServiceInfo{}, false)

	// THEN err is nil
	if err != nil {
		t.Errorf("try with %v shouldn't have errored %q",
			webhook, err.Error())
	}
}

func TestWebHookSendSuccessWithDelay(t *testing.T) {
	// GIVEN a WebHook with 5s Delay
	jLog = utils.NewJLog("WARN", false)
	webhook := testWebHookSuccessful()
	*webhook.Delay = "5s"

	// WHEN Send is called on this WebHook with useDelay
	start := time.Now().UTC()
	webhook.Send(utils.ServiceInfo{}, true)

	// THEN it took >= 5s to return
	elapsed := time.Since(start)
	if elapsed < 5*time.Second {
		t.Errorf("Send with useDelay should have taken atleast %v. Only took %s",
			webhook.Delay, elapsed)
	}
}

func TestSliceSendWithNilSlice(t *testing.T) {
	// GIVEN a nil Slice
	var slice *Slice

	// WHEN Send is called on this Slice
	start := time.Now().UTC()
	slice.Send(utils.ServiceInfo{}, false)

	// THEN it returns almost instantly
	elapsed := time.Since(start)
	if elapsed > time.Second {
		t.Errorf("Send took more than 1s (%fs) with a %v Slice",
			elapsed.Seconds(), slice)
	}
}

func TestSliceSendFail(t *testing.T) {
	// GIVEN a Slice with a WebHook which will err on Send
	jLog = utils.NewJLog("WARN", false)
	webhook := testWebHookFailing()
	slice := Slice{
		"test": &webhook,
	}

	// WHEN Send is called on this Slice
	err := slice.Send(utils.ServiceInfo{}, false)

	// THEN err is about the failure
	contains := "WebHook didn't 2XX"
	e := utils.ErrorToString(err)
	if !strings.Contains(e, contains) {
		t.Errorf("Send with %v should have errored %q. Got %q",
			*slice["test"], contains, e)
	}
}

func TestSliceSendWithMultipleFails(t *testing.T) {
	// GIVEN a Slice with a WebHook which will err on Send
	jLog = utils.NewJLog("WARN", false)
	webhook := testWebHookFailing()
	slice := Slice{
		"test":  &webhook,
		"other": &webhook,
	}

	// WHEN Send is called on this Slice
	err := slice.Send(utils.ServiceInfo{}, false)

	// THEN err is about the failure
	contains := "WebHook didn't 2XX"
	e := utils.ErrorToString(err)
	if strings.Count(e, contains) != len(slice) {
		t.Errorf("Send with %v should have errored %q %d times. Got %q",
			*slice["test"], contains, len(slice), e)
	}
}

func TestSliceSendSuccess(t *testing.T) {
	// GIVEN a Slice with a WebHook which accepts 2XX response status code
	jLog = utils.NewJLog("WARN", false)
	webhook := testWebHookSuccessful()
	slice := Slice{
		"test": &webhook,
	}

	// WHEN try is called on this WebHook
	err := slice.Send(utils.ServiceInfo{}, false)

	// THEN err is nil
	if err != nil {
		t.Errorf("try with %v shouldn't have errored %q",
			*slice["test"], err.Error())
	}
}

func TestNotifiersSendWithNil(t *testing.T) {
	// GIVEN nil Notifiers
	var notifiers *Notifiers

	// WHEN Send is called with them
	err := notifiers.Send("title", "message", &utils.ServiceInfo{})

	// THEN err is nil
	if err != nil {
		t.Errorf("Send on %v Notifiers shouldn't have err'd. Got %s",
			notifiers, err.Error())
	}
}

func TestNotifiersSendWithNotifier(t *testing.T) {
	// GIVEN nil Notifiers
	id := "test"
	notifiers := Notifiers{
		Shoutrrr: &shoutrrr.Slice{
			"test": &shoutrrr.Shoutrrr{
				ID:   &id,
				Type: "gotify",
				URLFields: map[string]string{
					"host":  "example.com",
					"token": "Aqbq-Sk9NUzb.ct",
				},
				Options: map[string]string{
					"max_tries": "1",
					"delay":     "0s",
				},
			},
		},
	}
	jLog = utils.NewJLog("WARN", false)
	(*notifiers.Shoutrrr).Init(jLog, &id, &service_status.Status{}, &shoutrrr.Slice{}, &shoutrrr.Slice{}, &shoutrrr.Slice{})
	(*notifiers.Shoutrrr)["test"].Init(&id, &shoutrrr.Shoutrrr{}, &shoutrrr.Shoutrrr{}, &shoutrrr.Shoutrrr{})

	// WHEN Send is called with them
	err := notifiers.Send("title", "message", &utils.ServiceInfo{})

	// THEN err is a 404
	e := utils.ErrorToString(err)
	if !strings.Contains(e, "HTTP 404") {
		t.Errorf("Send on %v Notifiers should have 404'd. Got \n%s",
			notifiers, e)
	}
}
