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
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/release-argus/Argus/util"
	metric "github.com/release-argus/Argus/web/metrics"
)

// Send every WebHook in this Slice with a delay between each webhook.
func (w *Slice) Send(
	serviceInfo *util.ServiceInfo,
	useDelay bool,
) (errs error) {
	if w == nil {
		return
	}

	errChan := make(chan error)
	for index := range *w {
		go func(webhook *WebHook) {
			errChan <- webhook.Send(serviceInfo, useDelay)
		}((*w)[index])

		// Space out WebHook send starts.
		time.Sleep(200 * time.Millisecond)
	}

	for range *w {
		err := <-errChan
		if err != nil {
			errs = fmt.Errorf("%s\n%w",
				util.ErrorToString(errs), err)
		}
	}
	return
}

// Send the WebHook MaxTries number of times until a success.
func (w *WebHook) Send(
	serviceInfo *util.ServiceInfo,
	useDelay bool,
) (errs error) {
	logFrom := util.LogFrom{Primary: w.ID, Secondary: serviceInfo.ID} // For logging
	triesLeft := w.GetMaxTries()                                      // Number of times to send WebHook (until DesiredStatusCode received).

	if useDelay && w.GetDelay() != "0s" {
		// Delay sending the WebHook message by the defined interval.
		msg := fmt.Sprintf("Sleeping for %s before sending the WebHook", w.GetDelay())
		jLog.Info(msg, logFrom, true)
		w.SetExecuting(true, true) // disable sending of auto_approved w/ delay
		time.Sleep(w.GetDelayDuration())
	} else {
		w.SetExecuting(false, true)
	}

	for {
		// Check if we're deleting the Service.
		if w.ServiceStatus.Deleting {
			return
		}

		// Try sending the WebHook.
		err := w.try(&logFrom)

		// SUCCESS!
		if err == nil {
			metric.IncreasePrometheusCounter(metric.WebHookMetric,
				w.ID,
				serviceInfo.ID,
				"",
				"SUCCESS")
			failed := false
			w.Failed.Set(w.ID, &failed)
			w.AnnounceSend()
			return nil
		}

		// FAIL!
		jLog.Error(err, logFrom, true)
		metric.IncreasePrometheusCounter(metric.WebHookMetric,
			w.ID,
			serviceInfo.ID,
			"",
			"FAIL")
		triesLeft--
		errs = fmt.Errorf("%s\n%w",
			util.ErrorToString(errs), err)

		// Give up after MaxTries.
		if triesLeft == 0 {
			err := fmt.Errorf("failed %d times to send the WebHook for %s to %q",
				w.GetMaxTries(), *w.ServiceStatus.ServiceID, w.ID)
			jLog.Error(err, logFrom, true)
			failed := true
			w.Failed.Set(w.ID, &failed)
			w.AnnounceSend()
			if !w.GetSilentFails() {
				//#nosec G104 -- Errors will be logged to CL
				//nolint:errcheck // ^
				w.Notifiers.Send("WebHook fail", err.Error(), serviceInfo)
			}
			return
		}
		// Space out retries.
		time.Sleep(5 * time.Second)
	}
}

// try to send a WebHook to its URL with the body SHA1 and SHA256 encrypted with its Secret.
// It also simulates other GitHub headers and returns when an error is encountered.
func (w *WebHook) try(logFrom *util.LogFrom) (err error) {
	req := w.GetRequest()
	if req == nil {
		err = fmt.Errorf("failed to get *http.request for webhook")
		jLog.Error(err, *logFrom, true)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	req = req.WithContext(ctx)
	defer cancel()

	// HTTPS insecure skip verify.
	customTransport := &http.Transport{}
	if w.GetAllowInvalidCerts() {
		customTransport = http.DefaultTransport.(*http.Transport).Clone()
		//#nosec G402 -- explicitly wanted InsecureSkipVerify
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := &http.Client{Transport: customTransport}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyOkay := checkWebHookBody(string(body))

	// SUCCESS
	desiredStatusCode := w.GetDesiredStatusCode()
	if bodyOkay && (resp.StatusCode == desiredStatusCode || (desiredStatusCode == 0 && (strconv.Itoa(resp.StatusCode)[:1] == "2"))) {
		msg := fmt.Sprintf("(%d) WebHook received", resp.StatusCode)
		jLog.Info(msg, *logFrom, true)
		return
	}

	// FAIL
	// Pretty desiredStatusCode.
	prettyStatusCode := strconv.Itoa(desiredStatusCode)
	if prettyStatusCode == "0" {
		prettyStatusCode = "2XX"
	}

	return fmt.Errorf(
		"%sWebHook gave %d, not %s:\n%s\n%s",
		util.FormatMessageSource(*logFrom),
		resp.StatusCode,
		prettyStatusCode,
		resp.Status,
		string(body),
	)
}

func (n *Notifiers) Send(title string, message string, serviceInfo *util.ServiceInfo) error {
	if n == nil || n.Shoutrrr == nil {
		return nil
	}

	//nolint:wrapcheck
	return (*n.Shoutrrr).Send(title, message, serviceInfo, false)
}

func checkWebHookBody(body string) (okay bool) {
	okay = true
	if body == "" {
		return
	}
	invalidContains := []string{
		"do not have permission",
		"rules were not satisfied",
	}
	for i := range invalidContains {
		if strings.Contains(body, invalidContains[i]) {
			okay = false
			return
		}
	}
	return
}
