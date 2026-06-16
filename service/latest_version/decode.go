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

package latestver

import (
	"errors"
	"fmt"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/polymorphic"
)

// Decode creates and returns a new [Lookup] from format-encoded data.
func Decode(
	format string,
	data []byte,
	options *opt.Options,
	status *status.Status,
	cfg base.DefaultsConfig,
) (Lookup, error) {
	if len(data) == 0 {
		return nil, nil
	}

	// Polymorphic fields.
	defType := util.FirstNonDefault(cfg.Soft.Type, cfg.Hard.Type)
	// Create.
	fieldInheritable, err := polymorphic.Instantiate(
		format, data,
		nil,
		defType,
		ServiceMapInheritable,
	)
	if err != nil {
		var ite *polymorphic.InvalidTypeError
		// Override constructor type names.
		if errors.As(err, &ite) {
			ite.Allowed = PossibleTypes
		}
		return nil, &decode.KeyFieldError{
			Key: "latest_version",
			Err: err,
		}
	}

	// Assert back to Lookup.
	field, ok := fieldInheritable.(Lookup)
	if !ok {
		return nil, &decode.KeyFieldError{
			Key: "latest_version",
			Err: fmt.Errorf("expected latestver.Lookup, got %T", fieldInheritable),
		}
	}

	field.Init(options, status, cfg)
	if err := field.DecodeSelf(format, data); err != nil {
		return nil, &decode.KeyFieldError{
			Key: "latest_version",
			Err: err,
		}
	}

	return field, nil
}

// ApplyOverrides applies format-encoded overrides to a [Lookup] object.
// If the target is nil, a new [Lookup] is created.
func ApplyOverrides(
	format string,
	data []byte,
	target Lookup,
	options *opt.Options,
	status *status.Status,
	cfg base.DefaultsConfig,
) (Lookup, error) {
	// No overrides.
	if len(data) == 0 {
		return target, nil
	}
	// Remove.
	if decode.IsNull(data) {
		return nil, nil
	}
	// New.
	if target == nil {
		return Decode(
			format, data,
			options,
			status,
			cfg,
		)
	}

	defaultType := util.FirstNonDefault(cfg.Soft.Type, cfg.Hard.Type)
	newType, err := polymorphic.ResolveType(
		format, data,
		target,
		defaultType,
	)
	if err != nil {
		return nil, &decode.KeyFieldError{
			Key: "latest_version",
			Err: err,
		}
	}
	// New lookup type.
	if newType != target.GetType() {
		newLookup, err := Decode(
			format, data,
			options,
			status,
			cfg,
		)
		if err != nil {
			return nil, err
		}
		// Inherit require.
		newReq := newLookup.GetRequire()
		oldReq := target.GetRequire()
		newReq.Inherit(oldReq)
		return newLookup, nil
	}

	target = target.Copy(status)

	if err := target.ApplyOverrides(format, data); err != nil {
		return nil, &decode.KeyFieldError{
			Key: "latest_version",
			Err: err,
		}
	}

	return target, nil
}
