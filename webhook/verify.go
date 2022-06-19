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
	"fmt"
	"strconv"
	"time"

	"github.com/release-argus/Argus/utils"
)

// CheckValues of this Slice.
func (w *Slice) CheckValues(prefix string) (errs error) {
	if w == nil {
		return
	}

	for key := range *w {
		if err := (*w)[key].CheckValues(prefix + "    "); err != nil {
			errs = fmt.Errorf("%s%s  %s:\\%w", utils.ErrorToString(errs), prefix, key, err)
		}
	}

	if errs != nil {
		errs = fmt.Errorf("%swebhook:\\%s", prefix, utils.ErrorToString(errs))
	}
	return
}

// CheckValues are valid for this WebHook recipient.
func (w *WebHook) CheckValues(prefix string) (errs error) {
	// Delay
	if w.Delay != nil {
		// Default to seconds when an integer is provided
		if _, err := strconv.Atoi(*w.Delay); err == nil {
			*w.Delay += "s"
		}
		if _, err := time.ParseDuration(*w.Delay); err != nil {
			errs = fmt.Errorf("%s%sdelay: %q <invalid> (Use 'AhBmCs' duration format)", utils.ErrorToString(errs), prefix, *w.Delay)
		}
	}

	if w.Main != nil {
		if w.GetType() != "github" && w.GetType() != "gitlab" {
			errs = fmt.Errorf("%s%stype: %q <invalid> (the only webhook type is 'github' or 'gitlab' currently)\\", utils.ErrorToString(errs), prefix, w.GetType())
		}
		if w.GetURL() == nil {
			errs = fmt.Errorf("%s%surl: <required> (here, or in webhook.%s)\\", utils.ErrorToString(errs), prefix, *w.ID)
		}
		if w.GetSecret() == nil {
			errs = fmt.Errorf("%s%ssecret: <required> (here, or in webhook.%s)\\", utils.ErrorToString(errs), prefix, *w.ID)
		}
	}
	return
}

// Print the Slice.
func (w *Slice) Print(prefix string) {
	if w == nil || len(*w) == 0 {
		return
	}

	fmt.Printf("%swebhook:\n", prefix)
	for webhookID, webhook := range *w {
		fmt.Printf("%s  %s:\n", prefix, webhookID)
		webhook.Print(prefix + "    ")
	}
}

// Print the WebHook Struct.
func (w *WebHook) Print(prefix string) {
	utils.PrintlnIfNotNil(w.Type, fmt.Sprintf("%stype: %s", prefix, utils.DefaultIfNil(w.Type)))
	utils.PrintlnIfNotNil(w.URL, fmt.Sprintf("%surl: %s", prefix, utils.DefaultIfNil(w.URL)))
	utils.PrintlnIfNotNil(w.AllowInvalidCerts, fmt.Sprintf("%sallow_invalid_certs: %t", prefix, utils.DefaultIfNil(w.AllowInvalidCerts)))
	utils.PrintlnIfNotNil(w.Secret, fmt.Sprintf("%ssecret: %q", prefix, utils.DefaultIfNil(w.Secret)))
	utils.PrintlnIfNotNil(w.DesiredStatusCode, fmt.Sprintf("%sdesired_status_code: %d", prefix, utils.DefaultIfNil(w.DesiredStatusCode)))
	utils.PrintlnIfNotNil(w.Delay, fmt.Sprintf("%sdelay: %s", prefix, utils.DefaultIfNil(w.Delay)))
	utils.PrintlnIfNotNil(w.MaxTries, fmt.Sprintf("%smax_tries: %d", prefix, utils.DefaultIfNil(w.MaxTries)))
	utils.PrintlnIfNotNil(w.SilentFails, fmt.Sprintf("%ssilent_fails: %t", prefix, utils.DefaultIfNil(w.SilentFails)))
}
