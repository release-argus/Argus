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

// Package base provides the base struct for deployed_version lookups.
package base

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/release-argus/Argus/service/deployed_version/types/web/constants"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/util"
)

// Defaults are the default values for a Lookup.
type Defaults struct {
	AllowInvalidCerts *bool  `json:"allow_invalid_certs,omitempty" yaml:"allow_invalid_certs,omitempty"` // False = Disallows invalid HTTPS certificates.
	Method            string `json:"method,omitempty" yaml:"method,omitempty"`                           // HTTP method.

	Options *opt.Defaults `json:"-" yaml:"-"` // Options for the Lookup.
}

// Default sets these Defaults to the default values.
func (d *Defaults) Default() {
	// allow_invalid_certs.
	allowInvalidCerts := false
	d.AllowInvalidCerts = &allowInvalidCerts

	// method.
	d.Method = http.MethodGet
}

// CheckValues validates the fields of the Defaults struct.
func (d *Defaults) CheckValues(prefix string) error {
	var errs []error
	// Method.
	d.Method = strings.ToUpper(d.Method)
	if d.Method != "" && !util.Contains(constants.SupportedMethods, d.Method) {
		errs = append(errs,
			fmt.Errorf("%smethod: %q <invalid> (supported methods = ['%s'])",
				prefix, d.Method, strings.Join(constants.SupportedMethods, "', '")))
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
