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

package latestver

import (
	"fmt"
	"strings"

	"github.com/release-argus/Argus/util"
)

// Print the struct.
func (l *Lookup) Print(prefix string) {
	fmt.Printf("%slatest_version:\n", prefix)
	prefix += "  "
	util.PrintlnIfNotDefault(l.Type,
		fmt.Sprintf("%stype: %s", prefix, l.Type))
	util.PrintlnIfNotDefault(l.URL,
		fmt.Sprintf("%surl: %s", prefix, l.URL))
	util.PrintlnIfNotNil(l.AccessToken,
		fmt.Sprintf("%saccess_token: %q", prefix, util.DefaultIfNil(l.AccessToken)))
	util.PrintlnIfNotNil(l.AllowInvalidCerts,
		fmt.Sprintf("%sallow_invalid_certs: %t", prefix, util.DefaultIfNil(l.AllowInvalidCerts)))
	util.PrintlnIfNotNil(l.UsePreRelease,
		fmt.Sprintf("%suse_prerelease: %t", prefix, util.DefaultIfNil(l.UsePreRelease)))
	l.URLCommands.Print(prefix)
	l.Require.Print(prefix)
}

// CheckValues of the Lookup struct
func (l *Lookup) CheckValues(prefix string) (errs error) {
	if l.URL == "" {
		if l.Defaults != nil {
			errs = fmt.Errorf("%s%s  url: <required> e.g. github:'release-argus/Argus' or url:'https://example.com'\\",
				util.ErrorToString(errs), prefix)
		}
	} else if l.Type != "url" && l.Type != "github" {
		errType := "<required>"
		if l.Type != "" {
			errType = fmt.Sprintf("%q <invalid>", l.Type)
		}
		errs = fmt.Errorf("%s%s  type: %s e.g. github or url\\",
			util.ErrorToString(errs), prefix, errType)
	}
	if l.Type == "github" && strings.Count(l.URL, "/") > 1 {
		parts := strings.Split(l.URL, "/")
		l.URL = strings.Join(parts[len(parts)-2:], "/")
	}

	if requireErrs := l.Require.CheckValues(prefix + "  "); requireErrs != nil {
		errs = fmt.Errorf("%s%w",
			util.ErrorToString(errs), requireErrs)
	}
	if urlCommandErrs := l.URLCommands.CheckValues(prefix + "  "); urlCommandErrs != nil {
		errs = fmt.Errorf("%s%w",
			util.ErrorToString(errs), urlCommandErrs)
	}

	if errs != nil {
		errs = fmt.Errorf("%slatest_version:\\%w",
			prefix, errs)
	}

	return
}
