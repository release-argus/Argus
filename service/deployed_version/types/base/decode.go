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
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
)

// Decode creates and returns a new [Lookup] from format-encoded data.
func Decode(
	format string,
	data []byte,
	options *opt.Options,
	status *status.Status,
	cfg DefaultsConfig,
) (*Lookup, error) {
	if len(data) == 0 || decode.IsNull(data) {
		return nil, nil
	}

	field := Lookup{
		Options:      options,
		Status:       status,
		Defaults:     cfg.Soft,
		HardDefaults: cfg.Hard,
	}

	// Static fields.
	if err := decode.Unmarshal(format, data, &field); err != nil {
		return nil, err //nolint:wrapcheck
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
	status *status.Status,
	cfg DefaultsConfig,
) (*Lookup, error) {
	// Decode new Lookup.
	if target == nil {
		return Decode(
			format, data,
			options, status,
			cfg,
		)
	}

	// Static fields.
	if err := decode.Unmarshal(format, data, target); err != nil {
		return nil, err //nolint:wrapcheck
	}

	return target, nil
}
