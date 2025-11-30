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

// Package command provides the cli command functionality for Argus.
package command

import (
	"errors"
	"fmt"

	"github.com/release-argus/Argus/util"
)

// CheckValues validates that each Command passes templating.
func (s *Commands) CheckValues(prefix string) error {
	if s == nil {
		return nil
	}

	var errs []error
	for i, cmd := range *s {
		if err := cmd.CheckValues(); err != nil {
			errs = append(errs,
				fmt.Errorf("%sitem_%d: %w",
					prefix, i, err))
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// CheckValues validates that the Command passes templating.
func (c *Command) CheckValues() error {
	if c == nil {
		return nil
	}

	for _, arg := range *c {
		if !util.CheckTemplate(arg) {
			return fmt.Errorf("%s (%q) <invalid> (didn't pass templating)",
				c.String(), arg)
		}
	}
	return nil
}
