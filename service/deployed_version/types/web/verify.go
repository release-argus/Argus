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

// Package web provides a web-based lookup type.
package web

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/service/deployed_version/types/web/constants"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/polymorphic"
)

// CheckValues validates the fields of the receiver.
func (l *Lookup) CheckValues() error {
	var errs []error

	// URL.
	if l.URL == "" && l.Defaults != nil {
		errs = append(
			errs,
			&decode.FieldError{
				Key:         "url",
				Description: "URL to get the deployed_version from",
			},
		)
	}

	// Method.
	l.Method = strings.ToUpper(l.Method)
	method := l.method()
	if !util.Contains(constants.SupportedMethods, method) {
		errs = append(
			errs,
			polymorphic.InvalidTypeError{
				Key:     "method",
				Value:   method,
				Allowed: constants.SupportedMethods,
			},
		)
	}
	// Body unused in GET, ensure it is empty.
	if method == http.MethodGet {
		l.Body = ""
	}

	// JSON.
	if l.JSON != "" {
		_, err := decode.ParseKeys(l.JSON)
		if err != nil {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "json",
					Value:       l.JSON,
					Description: "JSON path to the version in the response",
				},
			)
		}
	}

	// RegEx.
	if l.Regex != "" {
		_, err := regexp.Compile(l.Regex)
		if err != nil {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "regex",
					Value:       l.Regex,
					Description: "RegEx to extract the version from the response",
				},
			)
		}
	}
	// Remove the RegExTemplate if no RegEx.
	if l.Regex == "" {
		l.RegexTemplate = ""
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
