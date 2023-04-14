// Copyright [2022] [Argus]
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

package command

import (
	"fmt"

	"github.com/release-argus/Argus/util"
)

// Print will print the Slice.
func (s *Slice) Print(prefix string) {
	if s == nil || len(*s) == 0 {
		return
	}

	fmt.Printf("%scommand:\n", prefix)

	for i := range *s {
		fmt.Printf("%s  - %s\n", prefix, (*s)[i].FormattedString())
	}
}

func (s *Slice) CheckValues(prefix string) (errs error) {
	if s == nil {
		return
	}

	for i := range *s {
		if err := (*s)[i].CheckValues(); err != nil {
			errs = fmt.Errorf("%s%s  item_%d: %w",
				util.ErrorToString(errs), prefix, i, err)
		}
	}

	if errs != nil {
		errs = fmt.Errorf("%scommand:\\%s",
			prefix, util.ErrorToString(errs))
	}
	return
}

func (c *Command) CheckValues() (err error) {
	if c == nil {
		return
	}

	for i := range *c {
		if !util.CheckTemplate((*c)[i]) {
			err = fmt.Errorf("%s (%q) <invalid> (didn't pass templating)\\",
				c.String(), (*c)[i])
			return
		}
	}
	return
}
