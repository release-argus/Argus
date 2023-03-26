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

	"github.com/release-argus/Argus/util"
)

// CheckValues of this Slice.
func (w *Slice) CheckValues(prefix string) (errs error) {
	if w == nil {
		return
	}

	keys := util.SortedKeys(*w)
	for _, key := range keys {
		if err := (*w)[key].CheckValues(prefix + "    "); err != nil {
			errs = fmt.Errorf("%s%s  %s:\\%w",
				util.ErrorToString(errs), prefix, key, err)
		}
	}

	if errs != nil {
		errs = fmt.Errorf("%swebhook:\\%s",
			prefix, util.ErrorToString(errs))
	}
	return
}

// CheckValues are valid for this WebHook recipient.
func (w *WebHook) CheckValues(prefix string) (errs error) {
	// Delay
	if w.Delay != "" {
		// Default to seconds when an integer is provided
		if _, err := strconv.Atoi(w.Delay); err == nil {
			w.Delay += "s"
		}
		if _, err := time.ParseDuration(w.Delay); err != nil {
			errs = fmt.Errorf("%s%sdelay: %q <invalid> (Use 'AhBmCs' duration format)",
				util.ErrorToString(errs), prefix, w.Delay)
		}
	}

	if !util.CheckTemplate(w.URL) {
		errs = fmt.Errorf("%s%surl: %q <invalid> (didn't pass templating)\\",
			util.ErrorToString(errs), prefix, w.URL)
	}
	if w.Main != nil {
		types := []string{"github", "gitlab"}
		if !util.Contains(types, w.GetType()) {
			errs = fmt.Errorf("%s%stype: %q <invalid> (supported types = %s)\\",
				util.ErrorToString(errs), prefix, w.GetType(), types)
		}
		if util.GetFirstNonDefault(
			w.URL,
			w.Main.URL,
			w.Defaults.URL) == "" {
			errs = fmt.Errorf("%s%surl: <required> (here, or in webhook.%s)\\",
				util.ErrorToString(errs), prefix, w.ID)
		}
		if w.GetSecret() == "" {
			errs = fmt.Errorf("%s%ssecret: <required> (here, or in webhook.%s)\\",
				util.ErrorToString(errs), prefix, w.ID)
		}
	}
	var headerErrs error
	if w.CustomHeaders != nil {
		for i := range *w.CustomHeaders {
			if !util.CheckTemplate((*w.CustomHeaders)[i].Value) {
				headerErrs = fmt.Errorf("%s%s  %s: %q <invalid> (didn't pass templating)\\",
					util.ErrorToString(headerErrs), prefix, (*w.CustomHeaders)[i].Key, (*w.CustomHeaders)[i].Value)
			}
		}
	}
	if headerErrs != nil {
		errs = fmt.Errorf("%s%scustom_headers:\\%w",
			util.ErrorToString(errs), prefix, headerErrs)
	}

	return
}

// Print the Slice.
func (w *Slice) Print(prefix string) {
	if w == nil || len(*w) == 0 {
		return
	}

	fmt.Printf("%swebhook:\n", prefix)
	keys := util.SortedKeys(*w)
	for _, webhookID := range keys {
		fmt.Printf("%s  %s:\n", prefix, webhookID)
		(*w)[webhookID].Print(prefix + "    ")
	}
}

// Print the WebHook Struct.
func (w *WebHook) Print(prefix string) {
	util.PrintlnIfNotDefault(w.Type,
		fmt.Sprintf("%stype: %s", prefix, w.Type))
	util.PrintlnIfNotDefault(w.URL,
		fmt.Sprintf("%surl: %s", prefix, w.URL))
	util.PrintlnIfNotNil(w.AllowInvalidCerts,
		fmt.Sprintf("%sallow_invalid_certs: %t", prefix, util.DefaultIfNil(w.AllowInvalidCerts)))
	if w.CustomHeaders != nil {
		fmt.Printf("%scustom_headers:\n", prefix)
		for i := range *w.CustomHeaders {
			fmt.Printf("%s  - %s: %s\n", prefix, (*w.CustomHeaders)[i].Key, (*w.CustomHeaders)[i].Value)
		}
	}
	util.PrintlnIfNotDefault(w.Secret,
		fmt.Sprintf("%ssecret: %q", prefix, w.Secret))
	util.PrintlnIfNotNil(w.DesiredStatusCode,
		fmt.Sprintf("%sdesired_status_code: %d", prefix, util.DefaultIfNil(w.DesiredStatusCode)))
	util.PrintlnIfNotDefault(w.Delay,
		fmt.Sprintf("%sdelay: %s", prefix, w.Delay))
	util.PrintlnIfNotNil(w.MaxTries,
		fmt.Sprintf("%smax_tries: %d", prefix, util.DefaultIfNil(w.MaxTries)))
	util.PrintlnIfNotNil(w.SilentFails,
		fmt.Sprintf("%ssilent_fails: %t", prefix, util.DefaultIfNil(w.SilentFails)))
}
