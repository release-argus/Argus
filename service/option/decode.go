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

package option

import (
	"github.com/release-argus/Argus/config/decode"
)

// DecodeDefaults creates and returns new [Defaults] from format-encoded data.
func DecodeDefaults(format string, data []byte) (*Defaults, error) {
	var field Defaults

	if err := decode.Unmarshal(format, data, &field); err != nil {
		return nil, &decode.KeyFieldError{
			Key: "options",
			Err: err,
		}
	}

	return &field, nil
}

// Decode creates and returns new [Options] from format-encoded data.
func Decode(
	format string,
	data []byte,
	cfg DefaultsConfig,
) (*Options, error) {
	field := Options{
		Defaults:     cfg.Soft,
		HardDefaults: cfg.Hard,
	}

	if err := decode.Unmarshal(format, data, &field); err != nil {
		return nil, &decode.KeyFieldError{
			Key: "options",
			Err: err,
		}
	}

	return &field, nil
}
