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

package latest_version

import (
	"fmt"
	"strings"

	"github.com/release-argus/Argus/utils"
)

// Print the struct.
func (l *Lookup) Print(prefix string) {
	fmt.Printf("%stype: %s\n", prefix, l.Type)
	utils.PrintlnIfNotDefault(l.URL, fmt.Sprintf("%surl: %s", prefix, l.URL))
	utils.PrintlnIfNotNil(l.AccessToken, fmt.Sprintf("%saccess_token: %s", prefix, utils.DefaultIfNil(l.AccessToken)))
	utils.PrintlnIfNotNil(l.AllowInvalidCerts, fmt.Sprintf("%sallow_invalid_certs: %t", prefix, utils.DefaultIfNil(l.AllowInvalidCerts)))
	utils.PrintlnIfNotNil(l.UsePreRelease, fmt.Sprintf("%suse_prerelease: %t", prefix, utils.DefaultIfNil(l.UsePreRelease)))
	if len(l.URLCommands) != 0 {
		fmt.Printf("%surl_commands:\n", prefix)
		l.URLCommands.Print(prefix)
	}
	if l.Require != nil {
		fmt.Printf("%srequire:\n", prefix)
		l.Require.Print(prefix)
	}
	l.options.Print(prefix)
}

// CheckValues of the Lookup struct
func (l *Lookup) CheckValues(prefix string) (errs error) {
	if l.URL == "" {
		errs = fmt.Errorf("%s%s  url: <required> e.g. github:'release-argus/Argus' or url:'https://example.com'\\",
			utils.ErrorToString(errs), prefix)
	}
	if l.Type == "github" && strings.Count(l.URL, "/") > 1 {
		parts := strings.Split(l.URL, "/")
		l.URL = strings.Join(parts[len(parts)-2:], "/")
	}

	if requireErrs := l.Require.CheckValues(prefix + "  "); requireErrs != nil {
		errs = fmt.Errorf("%s%w",
			utils.ErrorToString(errs), requireErrs)
	}
	if urlCommandErrs := l.URLCommands.CheckValues(prefix + "  "); urlCommandErrs != nil {
		errs = fmt.Errorf("%s%w",
			utils.ErrorToString(errs), urlCommandErrs)
	}

	if errs != nil {
		errs = fmt.Errorf("%slatest_version:\\%w",
			prefix, errs)
	}

	return
}
