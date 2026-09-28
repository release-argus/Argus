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

// Package github provides a github-based lookup type.
package github

import (
	"errors"
	"strings"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/service/latest_version/types/forge"
	"github.com/release-argus/Argus/util"
)

// CheckValues validates the fields of the receiver.
func (l *Lookup) CheckValues() error {
	var errs []error

	// Convert full URL to just `owner/repo`.
	if strings.Count(l.URL, "/") > 1 {
		parts := strings.Split(l.URL, "/")
		l.URL = strings.Join(parts[len(parts)-2:], "/")
	}

	if !forge.IsOwnerRepo(util.EvalEnvVars(l.URL)) {
		errs = append(
			errs,
			&decode.ErrField{
				Key:         "url",
				Value:       l.URL,
				Description: "e.g. release-argus/Argus",
			},
		)
	}

	if baseErrs := l.Lookup.CheckValues(); baseErrs != nil {
		errs = append(errs, baseErrs)
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
