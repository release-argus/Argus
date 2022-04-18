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

package slack

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

// Send every Slack in this Slice.
func (s *Slice) Send(message string, serviceInfo *utils.ServiceInfo) (err error) {
	if s == nil {
		return nil
	}
	if serviceInfo == nil {
		serviceInfo = &utils.ServiceInfo{ID: ""}
	}

	errs := make(chan error)
	for index := range *s {
		// Send each Slack message up to MaxTries number of times until they 200.
		go func(slack *Slack) {
			logFrom := utils.LogFrom{Primary: *slack.ID, Secondary: serviceInfo.ID}
			triesLeft := slack.GetMaxTries() // Number of times to send Slack (until 200 received).

			// Delay sending the Slack message by the defined interval.
			msg := fmt.Sprintf("Sleeping for %s before sending the Slack message", slack.GetDelay())
			jLog.Info(msg, logFrom, slack.GetDelay() != "0s")
			time.Sleep(slack.GetDelayDuration())

			payload := slack.GetPayload(message, serviceInfo)
			for {
				err := slack.Send(payload, logFrom)

				// SUCCESS!
				if err == nil {
					metrics.InitPrometheusCounterWithIDAndServiceIDAndResult(metrics.SlackMetric, *slack.ID, serviceInfo.ID, "SUCCESS")
					failed := false
					slack.Failed = &failed
					errs <- nil
					return
				}

				// FAIL
				jLog.Error(err, logFrom, true)
				metrics.InitPrometheusCounterWithIDAndServiceIDAndResult(metrics.SlackMetric, *slack.ID, serviceInfo.ID, "FAIL")
				triesLeft--

				// Give up after MaxTries.
				if triesLeft == 0 {
					err = fmt.Errorf("failed %d times to send a Slack message to %s", slack.GetMaxTries(), *slack.GetURL())
					jLog.Error(err, logFrom, true)
					failed := true
					slack.Failed = &failed
					errs <- err
					return
				}

				// Space out retries.
				time.Sleep(10 * time.Second)
			}
		}((*s)[index])
		// Space out Slack messages.const.
		time.Sleep(3 * time.Second)
	}

	for range *s {
		errFound := <-errs
		if errFound != nil {
			if err == nil {
				err = errFound
			} else {
				err = fmt.Errorf("%s\n%s", err.Error(), errFound.Error())
			}
		}
	}
	return
}

// Send the Slack.
func (s *Slack) Send(payloadData []byte, logFrom utils.LogFrom) error {
	req, err := http.NewRequest(http.MethodPost, *s.GetURL(), bytes.NewReader(payloadData))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	req = req.WithContext(ctx)
	defer cancel()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// If verbose or above, print the error every time
		err = fmt.Errorf("Slack\n%s", err)
		jLog.Verbose(err, logFrom, true)
		return err
	}
	defer resp.Body.Close()

	// SUCCESS (2XX)
	if strconv.Itoa(resp.StatusCode)[:1] == "2" {
		jLog.Info("Slack message sent", logFrom, true)
		return nil
	}

	// FAIL
	return fmt.Errorf("%sSlack message failed to send.\n%s", utils.FormatMessageSource(logFrom), err)
}
