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

package gotify

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/hymenaios-io/Hymenaios/utils"
	metrics "github.com/hymenaios-io/Hymenaios/web/metrics"
)

// Send every Gotification in this Slice.
func (g *Slice) Send(
	title string,
	message string,
	serviceInfo *utils.ServiceInfo,
) error {
	if g == nil {
		return nil
	}
	if serviceInfo == nil {
		serviceInfo = &utils.ServiceInfo{ID: ""}
	}

	errs := make(chan error)
	for key := range *g {
		// Send each Gotify message up to s.MaxTries number of times until they 200.
		go func(gotify *Gotify) {
			logFrom := utils.LogFrom{Primary: *gotify.ID, Secondary: serviceInfo.ID} // For logging
			triesLeft := gotify.GetMaxTries()                                        // Number of times to send Gotify (until 200 received).

			// Delay sending the Gotify message by the defined interval.
			msg := fmt.Sprintf("%s, Sleeping for %s before sending the Gotify message", *gotify.ID, gotify.GetDelay())
			jLog.Info(msg, logFrom, gotify.GetDelay() != "0s")
			time.Sleep(gotify.GetDelayDuration())

			payload := gotify.GetPayload(title, message, serviceInfo)
			for {
				err := gotify.Send(payload, logFrom)

				// SUCCESS!
				if err == nil {
					metrics.InitPrometheusCounterWithIDAndServiceIDAndResult(metrics.GotifyMetric, *gotify.ID, serviceInfo.ID, "SUCCESS")
					failed := false
					gotify.Failed = &failed
					errs <- nil
					return
				}

				// FAIL
				jLog.Error(err.Error(), logFrom, true)
				metrics.InitPrometheusCounterWithIDAndServiceIDAndResult(metrics.GotifyMetric, *gotify.ID, serviceInfo.ID, "FAIL")
				triesLeft--

				// Give up after MaxTries.
				if triesLeft == 0 {
					msg = fmt.Sprintf("failed %d times to send a Gotify message to %s", gotify.GetMaxTries(), *gotify.GetURL())
					jLog.Error(msg, logFrom, true)
					failed := true
					gotify.Failed = &failed
					errs <- err
					return
				}

				// Space out retries.
				time.Sleep(10 * time.Second)
			}
		}((*g)[key])
		// Space out Gotify messages.const.
		time.Sleep(3 * time.Second)
	}

	var err error
	for range *g {
		errFound := <-errs
		if errFound != nil {
			if err == nil {
				err = errFound
			} else {
				err = fmt.Errorf("%s\n%s", err.Error(), errFound.Error())
			}
		}
	}
	return err
}

// HandleExtras will parse the messaging extras from 'extras' and 'defaults' into the Payload.
func (p *Payload) HandleExtras(
	gotify *Gotify,
	serviceInfo *utils.ServiceInfo,
) {
	// When received on Android and Gotify app is in focus
	androidAction := gotify.GetExtraAndroidAction()
	if androidAction != nil {
		p.Extras["android::action"] = map[string]interface{}{
			"onReceive": map[string]string{
				"intentURL": utils.TemplateString(*androidAction, *serviceInfo),
			},
		}
	}

	// Fomatting (markdown / plain)
	clientDisplay := gotify.GetExtraClientDisplay()
	if clientDisplay != nil {
		p.Extras["client::display"] = map[string]interface{}{
			"click": map[string]string{
				"url": utils.TemplateString(*clientDisplay, *serviceInfo),
			},
		}
	}

	// When the notification is clicked (Android)
	clientNotification := gotify.GetExtraClientNotification()
	if clientNotification != nil {
		p.Extras["client::notification"] = map[string]interface{}{
			"click": map[string]string{
				"url": utils.TemplateString(*clientNotification, *serviceInfo),
			},
		}
	}
}

// Send the Gotification.
func (g *Gotify) Send(
	payload []byte,
	logFrom utils.LogFrom,
) error {
	gotifyURL := fmt.Sprintf("%s/message?token=%s", *g.GetURL(), *g.GetToken())
	req, err := http.NewRequest(http.MethodPost, gotifyURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	req = req.WithContext(ctx)
	defer cancel()

	if err != nil {
		err = fmt.Errorf("Gotify\n%s", err)
		jLog.Verbose(err, logFrom, true)
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// If verbose or above, print the error every time
		err = fmt.Errorf("Gotify\n%s", err)
		jLog.Verbose(err, logFrom, true)
		return err
	}
	defer resp.Body.Close()

	// SUCCESS (2XX)
	if strconv.Itoa(resp.StatusCode)[:1] == "2" {
		msg := "Gotify message sent"
		jLog.Info(msg, logFrom, true)
		return nil
	}

	// FAIL
	return fmt.Errorf("%sGotify message failed to send.\n%s", utils.FormatMessageSource(logFrom), err)
}
