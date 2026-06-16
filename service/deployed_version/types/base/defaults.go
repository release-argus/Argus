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

// Package base provides the base struct for deployed_version lookups.
package base

import (
	"errors"
	"net/http"
	"strings"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/service/deployed_version/types/web/constants"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/polymorphic"
)

// #########
// # TYPES #
// #########

// DefaultsConfig pairs soft and hard deployed version defaults.
type DefaultsConfig struct {
	Soft *Defaults
	Hard *Defaults
}

// Defaults are the default values for a Lookup.
type Defaults struct {
	Type              string `json:"type,omitempty" yaml:"type,omitempty"`                               // "manual" | "url".
	AllowInvalidCerts *bool  `json:"allow_invalid_certs,omitempty" yaml:"allow_invalid_certs,omitempty"` // False = Disallows invalid HTTPS certificates.
	Method            string `json:"method,omitempty" yaml:"method,omitempty"`                           // HTTP method.

	Options *opt.Defaults `json:"-" yaml:"-"` // Options for the Lookup.
}

// ############
// # DECODING #
// ############

// DecodeDefaults decodes deployed version defaults from format-encoded data.
func DecodeDefaults(format string, data []byte) (*Defaults, error) {
	var field Defaults

	if err := decode.Unmarshal(format, data, &field); err != nil {
		return nil, &decode.KeyFieldError{
			Key: "deployed_version",
			Err: err,
		}
	}

	return &field, nil
}

// #########
// # STATE #
// #########

// IsZero implements the yaml.IsZeroer interface.
func (d Defaults) IsZero() bool {
	return d.Type == "" &&
		d.AllowInvalidCerts == nil &&
		d.Method == ""
}

// ########
// # INIT #
// ########

// Default sets the values of the receiver to their default values.
func (d *Defaults) Default() {
	// allow_invalid_certs.
	allowInvalidCerts := false
	d.AllowInvalidCerts = &allowInvalidCerts

	// method.
	d.Method = http.MethodGet

	// type.
	d.Type = "url"
}

// ##############
// # VALIDATION #
// ##############

// CheckValues validates the fields of the receiver.
func (d *Defaults) CheckValues() error {
	var errs []error
	// Method.
	d.Method = strings.ToUpper(d.Method)
	if d.Method != "" && !util.Contains(constants.SupportedMethods, d.Method) {
		errs = append(
			errs,
			polymorphic.InvalidTypeError{
				Key:     "method",
				Value:   d.Method,
				Allowed: constants.SupportedMethods,
			},
		)
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
