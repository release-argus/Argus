// Copyright [2022] [Hymenaios]
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

	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/hymenaios-io/Hymenaios/utils"
	metrics "github.com/hymenaios-io/Hymenaios/web/metrics"
)

// Send every WebHook in this Slice with a delay between each webhook.
func (w *Slice) Send(
	serviceInfo utils.ServiceInfo,
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
		time.Sleep(3 * time.Second)
	}

	for range *w {
		err := <-errChan
		if err != nil {
			errs = fmt.Errorf("%s\n%w", utils.ErrorToString(errs), err)
		}
	}
	return
}

// Send the WebHook MaxTries number of times until a success.
func (w *WebHook) Send(
	serviceInfo utils.ServiceInfo,
	useDelay bool,
) (errs error) {
	logFrom := utils.LogFrom{Primary: *w.ID, Secondary: serviceInfo.ID} // For logging
	triesLeft := w.GetMaxTries()                                        // Number of times to send WebHook (until DesiredStatusCode received).

	if useDelay {
		// Delay sending the WebHook message by the defined interval.
		msg := fmt.Sprintf("Sleeping for %s before sending the WebHook", w.GetDelay())
		jLog.Info(msg, logFrom, (w.GetDelay() != "0s"))
		time.Sleep(w.GetDelayDuration())
	}

	for {
		err := w.try(logFrom)

		// SUCCESS!
		if err == nil {
			metrics.IncreasePrometheusCounterWithIDAndServiceIDAndResult(metrics.WebHookMetric, *w.ID, serviceInfo.ID, "SUCCESS")
			failed := false
			w.Failed = &failed
			w.AnnounceSend()
			return nil
		}

		// FAIL!
		jLog.Error(err, logFrom, true)
		metrics.IncreasePrometheusCounterWithIDAndServiceIDAndResult(metrics.WebHookMetric, *w.ID, serviceInfo.ID, "FAIL")
		triesLeft--
		errs = fmt.Errorf("%s\n%w", utils.ErrorToString(errs), err)

		// Give up after MaxTries.
		if triesLeft == 0 {
			err := fmt.Errorf("failed %d times to send the WebHook (%s)", w.GetMaxTries(), *w.ID)
			jLog.Error(err, logFrom, true)
			failed := true
			w.Failed = &failed
			w.AnnounceSend()
			if !w.GetSilentFails() {
				//#nosec G104 -- Errors will be logged to CL
				//nolint:errcheck // Errors will be logged to CL
				w.Notifiers.Gotify.Send("WebHook fail", err.Error(), &serviceInfo)
				//#nosec G104 -- Errors will be logged to CL
				//nolint:errcheck // Errors will be logged to CL
				w.Notifiers.Slack.Send(err.Error(), &serviceInfo)
			}
			return
		}
		// Space out retries.
		time.Sleep(10 * time.Second)
	}
}

// try to send a WebHook to its URL with the body SHA1 and SHA256 encrypted with its Secret.
// It also simulates other GitHub headers and returns when an error is encountered.
func (w *WebHook) try(logFrom utils.LogFrom) (err error) {
	req := w.GetRequest()
	if req == nil {
		err = fmt.Errorf("failed to get *http.request for webhook")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	req = req.WithContext(ctx)
	defer cancel()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// If verbose or above, print the error every time
		err = fmt.Errorf("WebHook:\n%s", err)
		jLog.Error(err, logFrom, (jLog.Level > 2))
		return
	}
	defer resp.Body.Close()

	// SUCCESS
	desiredStatusCode := w.GetDesiredStatusCode()
	if resp.StatusCode == desiredStatusCode || (desiredStatusCode == 0 && (strconv.Itoa(resp.StatusCode)[:1] == "2")) {
		msg := fmt.Sprintf("(%d) WebHook received", resp.StatusCode)
		jLog.Info(msg, logFrom, true)
		return
	}

	// FAIL
	body, _ := ioutil.ReadAll(resp.Body)

	// Pretty desiredStatusCode.
	prettyStatusCode := strconv.Itoa(desiredStatusCode)
	if prettyStatusCode == "0" {
		prettyStatusCode = "2XX"
	}

	return fmt.Errorf(
		"%sWebHook didn't %s:\n%s\n%s", utils.FormatMessageSource(logFrom),
		prettyStatusCode,
		resp.Status,
		body,
	)
}
