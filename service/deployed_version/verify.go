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

package deployedver

import (
	"fmt"
	"regexp"

	"github.com/release-argus/Argus/util"
)

// Print will print the Lookup.
func (l *Lookup) Print(prefix string) {
	if l == nil {
		return
	}
	fmt.Printf("%sdeployed_version:\n", prefix)
	prefix += "  "

	util.PrintlnIfNotDefault(l.URL,
		fmt.Sprintf("%surl: %s", prefix, l.URL))
	util.PrintlnIfNotNil(l.AllowInvalidCerts,
		fmt.Sprintf("%sallow_invalid_certs: %t", prefix, util.DefaultIfNil(l.AllowInvalidCerts)))
	if l.BasicAuth != nil {
		fmt.Printf("%sbasic_auth:\n", prefix)
		fmt.Printf("%s  username: %s\n", prefix, l.BasicAuth.Username)
		fmt.Printf("%s  password: <secret>\n", prefix)
	}
	if l.Headers != nil {
		fmt.Printf("%sheaders:\n", prefix)
		for _, header := range l.Headers {
			fmt.Printf("%s  - key: %s\n", prefix, header.Key)
			fmt.Printf("%s    value: <secret>\n", prefix)
		}
	}
	util.PrintlnIfNotDefault(l.JSON,
		fmt.Sprintf("%sjson: %q", prefix, l.JSON))
	util.PrintlnIfNotDefault(l.Regex,
		fmt.Sprintf("%sregex: %q", prefix, l.Regex))
}

// CheckValues of the Lookup.
func (l *Lookup) CheckValues(prefix string) (errs error) {
	if l == nil {
		return
	}

	// URL
	if l.URL == "" && l.Defaults != nil {
		errs = fmt.Errorf("%s%s  url: <missing> (URL to get the deployed_version is required)\\",
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
