// Copyright [2026] [Argus]
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
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/polymorphic"
)

// CheckValues validates each entry and reports whether any values were modified.
func (whd *WebHooksDefaults) CheckValues() (error, bool) {
	if whd == nil {
		return nil, false
	}

	var errs []error
	changed := false
	keys := util.SortedKeys(*whd)
	for _, key := range keys {
		d := (*whd)[key]
		err, keyChanged := d.CheckValues()
		if err != nil {
			errs = append(
				errs,
				&decode.KeyFieldError{
					Key: key,
					Err: err,
				},
			)
		}
		changed = changed || keyChanged
	}

	if len(errs) == 0 {
		return nil, changed
	}
	return errors.Join(errs...), false
}

// CheckValues validates the fields of each [WebHook],
// returning errors encountered and whether any values were modified.
func (wh *WebHooks) CheckValues() (error, bool) {
	if wh == nil {
		return nil, false
	}

	var errs []error
	changed := false
	keys := util.SortedKeys(*wh)
	for _, key := range keys {
		err, keyChanged := (*wh)[key].CheckValues()
		if err != nil {
			errs = append(
				errs,
				&decode.KeyFieldError{
					Key: key,
					Err: err,
				},
			)
		}
		changed = changed || keyChanged
	}

	if len(errs) == 0 {
		return nil, changed
	}
	return errors.Join(errs...), false
}

// CheckValues validates the fields of the receiver,
// returning errors encountered and whether any values were modified.
func (wh *WebHook) CheckValues() (error, bool) {
	var errs []error

	// type
	whType := wh.GetType()
	if whType == "" {
		errs = append(
			errs,
			polymorphic.InvalidTypeError{
				Key:     "type",
				Allowed: supportedTypes,
			},
		)
		// Check the Type doesn't differ in the Main.
	} else if wh.Main.Type != "" && whType != wh.Main.Type {
		errs = append(
			errs,
			&decode.FieldError{
				Key:   "type",
				Value: whType,
				Description: fmt.Sprintf(
					"omit 'type', or match the root defaults.%s.type of %q",
					wh.ID, wh.Main.Type,
				),
			},
		)
	}

	baseErrs, changed := wh.Base.CheckValues()
	if baseErrs != nil {
		errs = append(errs, baseErrs)
	}

	// url
	if util.FirstNonDefault(
		wh.URL,
		wh.Main.URL,
		wh.Defaults.URL,
		wh.HardDefaults.URL,
	) == "" {
		errs = append(
			errs,
			&decode.FieldError{
				Key: "url",
				Description: fmt.Sprintf(
					"here, in root.defaults.%s, or in defaults",
					wh.ID,
				),
			},
		)
	}
	// secret
	if wh.GetSecret() == "" {
		errs = append(
			errs,
			&decode.FieldError{
				Key: "secret",
				Description: fmt.Sprintf(
					"here, in root.defaults.%s, or in defaults",
					wh.ID,
				),
			},
		)
	}

	if len(errs) == 0 {
		return nil, changed
	}
	return errors.Join(errs...), false
}

// CheckValues validates the fields of the receiver,
// returning errors encountered and whether any values were modified.
func (b *Base) CheckValues() (error, bool) {
	errs := make([]error, 0, 2)
	changed := false
	// type
	if b.Type != "" && !util.Contains(supportedTypes, b.Type) {
		errs = append(
			errs,
			polymorphic.InvalidTypeError{
				Key:     "type",
				Value:   b.Type,
				Allowed: supportedTypes,
			},
		)
	}
	// url
	if !util.CheckTemplate(b.URL) {
		errs = append(
			errs,
			&decode.FieldError{
				Key:         "url",
				Value:       b.URL,
				Description: "didn't pass templating",
			},
		)
	}
	// Deprecated: custom_header -> headers
	if b.Headers == nil && b.CustomHeaders != nil {
		b.Headers = b.CustomHeaders
		b.CustomHeaders = nil
		changed = true
		logx.Deprecated("Renaming 'webhook.custom_headers' to 'webhook.headers'. If you use any 'ARGUS_*_CUSTOM_HEADERS' environment variables, please update them to 'ARGUS_*_HEADERS' instead.")
	}
	if b.Headers != nil {
		if err := b.checkValuesHeaders(); err != nil {
			errs = append(
				errs,
				&decode.KeyFieldError{
					Key: "headers",
					Err: err,
				},
			)
		}
	}
	// delay
	if b.Delay != "" {
		// Treat integers as seconds by default.
		if _, err := strconv.Atoi(b.Delay); err == nil {
			b.Delay += "s"
		}
		if _, err := time.ParseDuration(b.Delay); err != nil {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "delay",
					Value:       b.Delay,
					Description: "use 'AhBmCs' duration format",
				},
			)
		}
	}

	if len(errs) == 0 {
		return nil, changed
	}
	return errors.Join(errs...), false
}

// Print the WebHooksDefaults.
func (whd *WebHooksDefaults) Print(prefix string) {
	if whd == nil || len(*whd) == 0 {
		return
	}

	str := whd.String(prefix + "  ")
	fmt.Printf(
		"%swebhook:\n%s",
		prefix, str,
	)
}

// checkValuesHeaders validates the fields of the Headers struct,
// returning errors encountered.
func (b *Base) checkValuesHeaders() error {
	var errs []error

	for _, header := range b.Headers {
		if !util.CheckTemplate(header.Value) {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         header.Key,
					Value:       header.Value,
					Description: "didn't pass templating",
				},
			)
		}
	}

	if errs == nil {
		return nil
	}
	return errors.Join(errs...)
}
