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

// Package github provides a github-based lookup type.
package github

import (
	"errors"
	"fmt"
	"strings"
)

// CheckValues validates the fields of the Lookup struct.
func (l *Lookup) CheckValues(prefix string) error {
	var errs []error
	if l.URL == "" {
		errs = append(errs,
			fmt.Errorf("%surl: <required> e.g. 'release-argus/Argus'",
				prefix))
		// Convert full URL to just `owner/repo`.
	} else if strings.Count(l.URL, "/") > 1 {
		parts := strings.Split(l.URL, "/")
		l.URL = strings.Join(parts[len(parts)-2:], "/")
	}

	if baseErrs := l.Lookup.CheckValues(prefix); baseErrs != nil {
		errs = append(errs, baseErrs)
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
