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

package file

import (
	"errors"
	"path/filepath"
	"regexp"

	"github.com/release-argus/Argus/config/decode"
)

// CheckValues validates the fields of the receiver.
func (l *Lookup) CheckValues() error {
	var errs []error
	if l.Path == "" {
		errs = append(errs, &decode.ErrField{
			Key:         "path",
			Description: "absolute path to a file allowed by settings.data.file_read_paths",
		})
	} else if !filepath.IsAbs(l.Path) {
		errs = append(errs, &decode.ErrField{
			Key:         "path",
			Value:       l.Path,
			Description: "must be an absolute path",
		})
	}

	if l.Regex == "" {
		errs = append(errs, &decode.ErrField{
			Key:         "regex",
			Description: "regular expression with a capture group containing the version",
		})
	} else if re, err := regexp.Compile(l.Regex); err != nil {
		errs = append(errs, &decode.ErrField{
			Key:         "regex",
			Value:       l.Regex,
			Description: "invalid regular expression",
		})
	} else if re.NumSubexp() == 0 {
		errs = append(errs, &decode.ErrField{
			Key:         "regex",
			Value:       l.Regex,
			Description: "must include a capture group containing the version",
		})
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
