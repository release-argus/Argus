// Copyright [2023] [Argus]
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

package deployedver

import (
	"fmt"
	"regexp"

	"github.com/release-argus/Argus/util"
)

// CheckValues of the Lookup.
func (l *Lookup) CheckValues(prefix string) (errs error) {
	if l == nil {
		return
	}

	// URL
	if l.URL == "" && l.Defaults != nil {
		errs = fmt.Errorf("%s%s  url: <required> (URL to get the deployed_version is required)\\",
			util.ErrorToString(errs), prefix)
	}

	// RegEx
	_, err := regexp.Compile(l.Regex)
	if err != nil {
		errs = fmt.Errorf("%s%s  regex: %q <invalid> (Invalid RegEx)\\",
			util.ErrorToString(errs), prefix, l.Regex)
	}

	if errs != nil {
		errs = fmt.Errorf("%sdeployed_version:\\%w",
			prefix, errs)
	}
	return
}
