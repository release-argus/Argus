// Copyright [2024] [Argus]
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
)

// CheckValues validates the fields of the Lookup struct.
func (l *Lookup) CheckValues(prefix string) error {
	var errs []error
	if l.URL == "" {
		errs = append(errs,
			fmt.Errorf("%surl: <required> e.g. 'https://example.com'",
				prefix))
	}

	if baseErrs := l.Lookup.CheckValues(prefix); baseErrs != nil {
		errs = append(errs, baseErrs)
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
