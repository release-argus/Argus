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

// Package web provides a web-based lookup type.
package web

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/release-argus/Argus/util"
)

// CheckValues validates the fields of the Lookup struct.
func (l *Lookup) CheckValues(prefix string) error {
	var errs []error
	// Method.
	l.Method = strings.ToUpper(l.Method)
	if l.Method == "" {
		l.Method = http.MethodGet
	} else if !util.Contains(supportedMethods, l.Method) {
		errs = append(errs,
			fmt.Errorf("%smethod: %q <invalid> (only [%s] are allowed)",
				prefix, l.Method, strings.Join(supportedMethods, ", ")))
	}
	// Body unused in GET, ensure it is empty.
	if l.Method == http.MethodGet {
		l.Body = ""
	}

	// URL.
	if l.URL == "" && l.Defaults != nil {
		errs = append(errs,
			fmt.Errorf("%surl: <required> (URL to get the deployed_version is required)",
				prefix))
	}

	// JSON.
	if l.JSON != "" {
		_, err := util.ParseKeys(l.JSON)
		if err != nil {
			errs = append(errs,
				fmt.Errorf("%sjson: %q <invalid> - %w",
					prefix, l.JSON, err))
		}
	}

	// RegEx.
	if l.Regex != "" {
		_, err := regexp.Compile(l.Regex)
		if err != nil {
			errs = append(errs,
				fmt.Errorf("%sregex: %q <invalid>",
					prefix, l.Regex))
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
