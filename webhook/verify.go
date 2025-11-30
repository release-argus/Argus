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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/release-argus/Argus/util"
)

// CheckValues validates the fields of each Defaults struct.
func (whd *WebHooksDefaults) CheckValues(prefix string) error {
	if whd == nil {
		return nil
	}

	var errs []error
	keys := util.SortedKeys(*whd)
	itemPrefix := prefix + "  "
	for _, key := range keys {
		util.AppendCheckError(&errs, prefix, key,
			(*whd)[key].CheckValues(itemPrefix))
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// CheckValues validates the fields of each WebHook.
func (wh *WebHooks) CheckValues(prefix string) error {
	if wh == nil {
		return nil
	}

	var errs []error
	keys := util.SortedKeys(*wh)
	itemPrefix := prefix + "  "
	for _, key := range keys {
		util.AppendCheckError(&errs, prefix, key,
			(*wh)[key].CheckValues(itemPrefix))
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// CheckValues validates the fields of the Base struct.
func (b *Base) CheckValues(prefix string) error {
	var errs []error
	// type
	if b.Type != "" && !util.Contains(supportedTypes, b.Type) {
		errs = append(errs,
			fmt.Errorf("%stype: %q <invalid> (supported types = ['%s'])",
				prefix, b.Type, strings.Join(supportedTypes, "', '")))
	}
	// url
	if !util.CheckTemplate(b.URL) {
		errs = append(errs,
			fmt.Errorf("%surl: %q <invalid> (didn't pass templating)",
				prefix, b.URL))
	}
	// custom_headers
	if b.CustomHeaders != nil {
		util.AppendCheckError(&errs, prefix, "custom_headers",
			b.checkValuesCustomHeaders(prefix+"  "))
	}
	// delay
	if b.Delay != "" {
		// Treat integers as seconds by default.
		if _, err := strconv.Atoi(b.Delay); err == nil {
			b.Delay += "s"
		}
		if _, err := time.ParseDuration(b.Delay); err != nil {
			errs = append(errs,
				fmt.Errorf("%sdelay: %q <invalid> (Use 'AhBmCs' duration format)",
					prefix, b.Delay))
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

func (b *Base) checkValuesCustomHeaders(prefix string) error {
	var errs []error

	for _, customHeader := range *b.CustomHeaders {
		if !util.CheckTemplate(customHeader.Value) {
			errs = append(errs, fmt.Errorf("%s%s: %q <invalid> (didn't pass templating)",
				prefix, customHeader.Key, customHeader.Value))
		}
	}

	if errs == nil {
		return nil
	}
	return errors.Join(errs...)
}

// CheckValues validates the fields of the WebHook struct.
func (wh *WebHook) CheckValues(prefix string) error {
	var errs []error

	// type
	whType := wh.GetType()
	if whType == "" {
		errs = append(errs, fmt.Errorf("%stype: <required> (supported types = ['%s'])",
			prefix, strings.Join(supportedTypes, "', '")))
		// Check the Type doesn't differ in the Main.
	} else if wh.Main.Type != "" && whType != wh.Main.Type {
		errs = append(errs, fmt.Errorf("%stype: %q != %q <invalid> (omit 'type', or make it match root webhook.%s.type)",
			prefix, whType, wh.Main.Type, wh.ID))
	}

	if baseErrs := wh.Base.CheckValues(prefix); baseErrs != nil {
		errs = append(errs, baseErrs)
	}

	// url
	if util.FirstNonDefault(
		wh.URL,
		wh.Main.URL,
		wh.Defaults.URL,
		wh.HardDefaults.URL) == "" {
		errs = append(errs, fmt.Errorf("%surl: <required> (here, in root webhook.%s, or in defaults)",
			prefix, wh.ID))
	}
	// secret
	if wh.GetSecret() == "" {
		errs = append(errs, fmt.Errorf("%ssecret: <required> (here, in root webhook.%s, or in defaults)",
			prefix, wh.ID))
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// Print the WebHooksDefaults.
func (whd *WebHooksDefaults) Print(prefix string) {
	if whd == nil || len(*whd) == 0 {
		return
	}

	str := whd.String(prefix + "  ")
	fmt.Printf("%swebhook:\n%s",
		prefix, str)
}
