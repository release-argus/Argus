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

// CheckValues validates the fields of the SliceDefaults struct.
func (s *SliceDefaults) CheckValues(prefix string) error {
	if s == nil {
		return nil
	}

	var errs []error
	keys := util.SortedKeys(*s)
	itemPrefix := prefix + "  "
	for _, key := range keys {
		util.AppendCheckError(&errs, prefix, key, (*s)[key].CheckValues(itemPrefix))
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// CheckValues validates the fields of each WebHook in the Slice.
func (s *Slice) CheckValues(prefix string) error {
	if s == nil {
		return nil
	}

	var errs []error
	keys := util.SortedKeys(*s)
	itemPrefix := prefix + "  "
	for _, key := range keys {
		util.AppendCheckError(&errs, prefix, key, (*s)[key].CheckValues(itemPrefix))
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
			fmt.Errorf("%stype: %q <invalid> (supported types = [%s])",
				prefix, b.Type, strings.Join(supportedTypes, ",")))
	}
	// url
	if !util.CheckTemplate(b.URL) {
		errs = append(errs,
			fmt.Errorf("%surl: %q <invalid> (didn't pass templating)",
				prefix, b.URL))
	}
	// custom_headers
	if b.CustomHeaders != nil {
		util.AppendCheckError(&errs, prefix, "custom_headers", b.checkValuesCustomHeaders(prefix+"  "))
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
func (w *WebHook) CheckValues(prefix string) error {
	var errs []error

	// type
	whType := w.GetType()
	if whType == "" {
		errs = append(errs, fmt.Errorf("%stype: <required> (supported types = [%s])",
			prefix, strings.Join(supportedTypes, ",")))
		// Check the Type doesn't differ in the Main.
	} else if w.Main.Type != "" && whType != w.Main.Type {
		errs = append(errs, fmt.Errorf("%stype: %q != %q <invalid> (omit 'type', or make it match root webhook.%s.type)",
			prefix, whType, w.Main.Type, w.ID))
	}

	if baseErrs := w.Base.CheckValues(prefix); baseErrs != nil {
		errs = append(errs, baseErrs)
	}

	// url
	if util.FirstNonDefault(
		w.URL,
		w.Main.URL,
		w.Defaults.URL,
		w.HardDefaults.URL) == "" {
		errs = append(errs, fmt.Errorf("%surl: <required> (here, in root webhook.%s, or in defaults)",
			prefix, w.ID))
	}
	// secret
	if w.GetSecret() == "" {
		errs = append(errs, fmt.Errorf("%ssecret: <required> (here, in root webhook.%s, or in defaults)",
			prefix, w.ID))
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// Print the SliceDefaults.
func (s *SliceDefaults) Print(prefix string) {
	if s == nil || len(*s) == 0 {
		return
	}

	str := s.String(prefix + "  ")
	fmt.Printf("%swebhook:\n%s", prefix, str)
}
