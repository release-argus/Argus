// Copyright [2023] [Argus]
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
	"strings"
	"time"

	"github.com/release-argus/Argus/util"
)

// CheckValues of this SliceDefaults.
func (w *SliceDefaults) CheckValues(prefix string) (errs error) {
	if w == nil {
		return
	}

	keys := util.SortedKeys(*w)
	itemPrefix := prefix + "    "
	for _, key := range keys {
		if err := (*w)[key].CheckValues(itemPrefix); err != nil {
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

// CheckValues of this Slice.
func (w *Slice) CheckValues(prefix string) (errs error) {
	if w == nil {
		return
	}

	keys := util.SortedKeys(*w)
	itemPrefix := prefix + "    "
	for _, key := range keys {
		if err := (*w)[key].CheckValues(itemPrefix); err != nil {
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
func (w *WebHookBase) CheckValues(prefix string) (errs error) {
	// type
	if w.Type != "" && !util.Contains(supportedTypes, w.Type) {
		errs = fmt.Errorf("%s%stype: %q <invalid> (supported types = [%s])\\",
			util.ErrorToString(errs), prefix, w.Type, strings.Join(supportedTypes, ","))
	}

	// delay
	if w.Delay != "" {
		// Default to seconds when an integer is provided
		if _, err := strconv.Atoi(w.Delay); err == nil {
			w.Delay += "s"
		}
		if _, err := time.ParseDuration(w.Delay); err != nil {
			errs = fmt.Errorf("%s%sdelay: %q <invalid> (Use 'AhBmCs' duration format)\\",
				util.ErrorToString(errs), prefix, w.Delay)
		}
	}
	// url
	if !util.CheckTemplate(w.URL) {
		errs = fmt.Errorf("%s%surl: %q <invalid> (didn't pass templating)\\",
			util.ErrorToString(errs), prefix, w.URL)
	}
	// custom_headers
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

// CheckValues are valid for this WebHook recipient.
func (w *WebHook) CheckValues(prefix string) (errs error) {
	errs = w.WebHookBase.CheckValues(prefix)

	// type
	whType := w.GetType()
	if whType == "" {
		errs = fmt.Errorf("%s%stype: <required> (supported types = [%s])\\",
			util.ErrorToString(errs), prefix, strings.Join(supportedTypes, ","))
		// Check that the Type doesn't differ in the Main
	} else if w.Main.Type != "" && whType != w.Main.Type {
		errs = fmt.Errorf("%s%stype: %q != %q <invalid> (omit the type, or make it the same as the root webhook.%s.type)\\",
			util.ErrorToString(errs), prefix, whType, w.Main.Type, w.ID)
	}

	// url
	if util.FirstNonDefault(
		w.URL,
		w.Main.URL,
		w.Defaults.URL,
		w.HardDefaults.URL) == "" {
		errs = fmt.Errorf("%s%surl: <required> (here, in the root webhook.%s, or in defaults)\\",
			util.ErrorToString(errs), prefix, w.ID)
	}
	// secret
	if w.GetSecret() == "" {
		errs = fmt.Errorf("%s%ssecret: <required> (here, in the root webhook.%s, or in defaults)\\",
			util.ErrorToString(errs), prefix, w.ID)
	}

	return
}

// Print the SliceDefaults.
func (w *SliceDefaults) Print(prefix string) {
	if w == nil || len(*w) == 0 {
		return
	}

	str := w.String(prefix + "  ")
	fmt.Printf("%swebhook:\n%s", prefix, str)
}
