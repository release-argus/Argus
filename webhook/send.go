// Copyright [2025] [Argus]
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

// Package webhook provides WebHook functionality to services.
package webhook

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
	"github.com/release-argus/Argus/web/metric"
)

// Send sends each WebHook, optionally applying a delay between them.
func (s *WebHooks) Send(serviceInfo serviceinfo.ServiceInfo, useDelay bool) error {
	if s == nil {
		return nil
	}

	errChan := make(chan error, len(*s))
	for _, wh := range *s {
		go func(webhook *WebHook) {
			errChan <- webhook.Send(serviceInfo, useDelay)
		}(wh)

		// Space out WebHook send starts.
		//#nosec G404 —- sleep does not need cryptographic security.
		time.Sleep(time.Duration(100+rand.Intn(150)) * time.Millisecond)
	}

	// Wait for all WebHooks to finish.
	var errs []error
	for range *s {
		if err := <-errChan; err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// Send attempts to send the WebHook up to MaxTries times until successful.
func (w *WebHook) Send(serviceInfo serviceinfo.ServiceInfo, useDelay bool) error {
	logFrom := logutil.LogFrom{Primary: w.ID, Secondary: serviceInfo.ID}

	if useDelay && w.GetDelay() != "0s" {
		// Delay sending the WebHook message by the defined interval.
		msg := fmt.Sprintf("Sleeping for %s before sending the WebHook", w.GetDelay())
		logutil.Log.Info(msg, logFrom, true)
		w.SetExecuting(true, true) // disable sending of auto_approved w/ delay.
		time.Sleep(w.GetDelayDuration())
	} else {
		w.SetExecuting(false, true)
	}

	sendErrs := util.RetryWithBackoff(
		func() error {
			err := w.try(logFrom)
			w.parseTry(err, w.ServiceStatus.ServiceInfo.ID, logFrom)
			if err == nil {
				return nil
			}
			return err
		},
		w.GetMaxTries(),
		1*time.Second,
		30*time.Second,
		w.ServiceStatus.Deleting,
	)
	if sendErrs == nil {
		return nil
	}

	err := fmt.Errorf("failed %d times to send the WebHook for %s to %q",
		w.GetMaxTries(), w.ServiceStatus.ServiceInfo.ID, w.ID)
	logutil.Log.Error(err, logFrom, true)
	failed := true
	w.Failed.Set(w.ID, &failed)
	w.AnnounceSend()
	if !w.GetSilentFails() {
		//#nosec G104 -- Errors are logged to CLI
		//nolint:errcheck // ^
		w.Notifiers.Send("WebHook fail", err.Error(), serviceInfo)
	}
	return errors.Join(sendErrs, err)
}

// try sends a WebHook to its URL with the body hashed using SHA1 and SHA256,
// encrypted with its Secret, and includes simulated GitHub headers.
func (w *WebHook) try(logFrom logutil.LogFrom) error {
	req := w.BuildRequest()
	if req == nil {
		err := errors.New("failed to get *http.request for WebHook")
		logutil.Log.Error(err, logFrom, true)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	req = req.WithContext(ctx)
	defer cancel()

	// HTTPS insecure skip verify.
	customTransport := &http.Transport{}
	if w.GetAllowInvalidCerts() {
		customTransport = http.DefaultTransport.(*http.Transport).Clone()
		//#nosec G402 -- explicitly wanted InsecureSkipVerify.
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := &http.Client{Transport: customTransport}
	resp, err := client.Do(req)
	if err != nil {
		return err //nolint:wrapcheck
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyOkay := checkWebHookBody(string(body))

	// SUCCESS!
	desiredStatusCode := w.GetDesiredStatusCode()
	if bodyOkay && (resp.StatusCode == int(desiredStatusCode) || (desiredStatusCode == 0 && (strconv.Itoa(resp.StatusCode)[:1] == "2"))) {
		msg := fmt.Sprintf("(%d) WebHook received", resp.StatusCode)
		logutil.Log.Info(msg, logFrom, true)
		return nil
	}

	// FAIL!
	// Pretty desiredStatusCode.
	prettyStatusCode := strconv.FormatUint(uint64(desiredStatusCode), 10)
	if prettyStatusCode == "0" {
		prettyStatusCode = "2XX"
	}

	return fmt.Errorf("WebHook gave %d, not %s:\n%s\n%s",
		resp.StatusCode, prettyStatusCode,
		resp.Status, string(body),
	)
}

// Send sends a message to the Notifiers, if any are available.
func (n *Notifiers) Send(title, message string, serviceInfo serviceinfo.ServiceInfo) error {
	if n == nil || n.Shoutrrr == nil {
		return nil
	}

	//nolint:wrapcheck
	return (*n.Shoutrrr).Send(title, message, serviceInfo, false)
}

// checkWebHookBody checks whether the given body is valid for a WebHook.
// It returns false if the body contains certain forbidden phrases.
func checkWebHookBody(body string) (okay bool) {
	okay = true
	if body == "" {
		return
	}
	invalidContains := []string{
		`(?i)do not have permission`,
		`(?i)rules were not satisfied`,
	}
	for _, re := range invalidContains {
		if util.RegexCheck(re, body) {
			okay = false
			return
		}
	}
	return
}

// parseTry checks the result of a WebHook `try`.
// It updates the Prometheus SUCCESS or FAIL counter
// and logs any error from the `try`.
func (w *WebHook) parseTry(err error, serviceID string, logFrom logutil.LogFrom) {
	// SUCCESS!
	if err == nil {
		metric.IncPrometheusCounter(metric.WebHookResultTotal,
			w.ID,
			serviceID,
			"",
			"SUCCESS")
		failed := false
		w.Failed.Set(w.ID, &failed)
		w.AnnounceSend()
		return
	}

	// FAIL!
	logutil.Log.Error(err, logFrom, true)
	metric.IncPrometheusCounter(metric.WebHookResultTotal,
		w.ID,
		serviceID,
		"",
		"FAIL")
}
