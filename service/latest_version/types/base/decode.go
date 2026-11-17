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

package base

import (
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util/polymorphic"
)

// Decode creates and returns a new [Lookup] from format-encoded data.
func Decode(
	format string,
	data []byte,
	options *opt.Options,
	svcStatus *status.Status,
	cfg DefaultsConfig,
) (*Lookup, error) {
	if len(data) == 0 || decode.IsNull(data) {
		return nil, nil
	}

	field := Lookup{
		Options:      options,
		Status:       svcStatus,
		Defaults:     cfg.Soft,
		HardDefaults: cfg.Hard,
	}

	// Polymorphic fields.
	//   Require.
	reqRaw, err := polymorphic.Extract(format, data, "require")
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	if reqRaw != nil {
		var err error
		field.Require, err = filter.Decode(
			format, reqRaw,
			svcStatus,
			&cfg.Soft.Require,
		)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}
	}

	// Static fields.
	if err := field.unmarshalStatic(format, data); err != nil {
		return nil, err
	}

	return &field, nil
}

// ApplyOverrides applies format-encoded overrides to target.
// If the target is nil, a new [Lookup] is created.
func ApplyOverrides(
	format string,
	data []byte,
	target *Lookup,
	options *opt.Options,
	svcStatus *status.Status,
	cfg DefaultsConfig,
) (*Lookup, error) {
	// Decode require.
	if target == nil {
		return Decode(
			format, data,
			options, svcStatus,
			cfg,
		)
	}
	if len(data) == 0 {
		return target, nil
	}

	// Polymorphic fields.
	reqRaw, err := polymorphic.Extract(format, data, "require")
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	if reqRaw != nil {
		var err error
		// Require creation/overrides.
		target.Require, err = target.Require.ApplyOverrides(
			format, reqRaw,
			svcStatus,
			&cfg.Soft.Require,
		)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}
	}

	// Static fields.
	if err := target.unmarshalStatic(format, data); err != nil {
		return nil, err
	}

	return target, nil
}

// unmarshalStatic decodes static configuration fields, preserving dynamic fields.
func (l *Lookup) unmarshalStatic(format string, data []byte) error {
	aux := LookupDecode{
		Type:        l.Type,
		URL:         l.URL,
		URLCommands: l.URLCommands,
	}

	if err := decode.Unmarshal(format, data, &aux); err != nil {
		return err //nolint:wrapcheck
	}
	l.Type = aux.Type
	l.URL = aux.URL
	l.URLCommands = aux.URLCommands

	return nil
}

type lookupWithRequire interface {
	GetRequire() *filter.Require
	SetRequire(req *filter.Require)
}

// UnmarshalRequire decodes format-encoded data overrides onto target.
func UnmarshalRequire(
	format string, data []byte,
	target lookupWithRequire,
	svcStatus *status.Status,
	defaults *filter.RequireDefaults,
) error {
	// Extract.
	requireRaw, err := polymorphic.Extract(format, data, "require")
	if err != nil {
		return err //nolint:wrapcheck
	}
	if requireRaw == nil {
		return nil
	}

	// Overrides on existing Require.
	t := target.GetRequire()
	newReq, err := t.ApplyOverrides(
		format, requireRaw,
		svcStatus,
		defaults,
	)
	if err != nil {
		return err //nolint:wrapcheck
	}
	target.SetRequire(newReq)

	return nil
}
