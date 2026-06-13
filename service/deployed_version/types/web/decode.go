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

package web

import (
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/service/deployed_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
)

// DecodeSelf decodes the format-encoded data into the receiver.
func (l *Lookup) DecodeSelf(format string, data []byte) error {
	newL, err := Decode(
		format, data,
		l.Options,
		l.Status,
		base.DefaultsConfig{
			Soft: l.Defaults,
			Hard: l.HardDefaults,
		},
	)
	if err != nil {
		return err
	}
	if newL == nil {
		return nil
	}

	l.Lookup = newL.Lookup
	l.Method = newL.Method
	l.URL = newL.URL
	l.AllowInvalidCerts = newL.AllowInvalidCerts
	l.TargetHeader = newL.TargetHeader
	l.BasicAuth = newL.BasicAuth
	l.Headers = newL.Headers
	l.Body = newL.Body
	l.JSON = newL.JSON
	l.Regex = newL.Regex
	l.RegexTemplate = newL.RegexTemplate

	return nil
}

// Decode creates and returns a new [Lookup] from format-encoded data.
func Decode(
	format string,
	data []byte,
	options *opt.Options,
	status *status.Status,
	cfg base.DefaultsConfig,
) (*Lookup, error) {
	if len(data) == 0 || decode.IsNull(data) {
		return nil, nil
	}

	// Decode Interface.
	var field Lookup

	// Base.
	baseLookup, err := base.Decode(
		format, data,
		options,
		status,
		cfg,
	)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	if baseLookup != nil {
		field.Lookup = *baseLookup
	}

	// Static fields.
	if err := decode.Unmarshal(format, data, &field); err != nil {
		return nil, err //nolint:wrapcheck
	}

	return &field, nil
}

// ApplyOverrides applies format-encoded overrides to the receiver.
func (l *Lookup) ApplyOverrides(format string, data []byte) error {
	if len(data) == 0 {
		return nil
	}

	// Polymorphic fields.
	baseLookup, err := base.ApplyOverrides(
		format, data,
		&l.Lookup,
		l.Options,
		l.Status,
		base.DefaultsConfig{
			Soft: l.Defaults,
			Hard: l.HardDefaults,
		},
	)
	if err != nil {
		return err //nolint:wrapcheck
	}
	if baseLookup != nil {
		l.Lookup = *baseLookup
	}

	// Static fields.
	if err := decode.Unmarshal(format, data, l); err != nil {
		return err //nolint:wrapcheck
	}

	return nil
}
