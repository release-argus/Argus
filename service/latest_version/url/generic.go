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

package url

import (
	"fmt"

	"github.com/release-argus/Argus/utils"
)

func (l *LatestVersion) URLCommandsCheckValues(prefix string) error {
	return l.URLCommands.CheckValues(prefix)
}

func (l *LatestVersion) CheckValues(prefix string) (errs error) {
	if utils.DefaultIfNil(l.URL) == "" {
		errs = fmt.Errorf("%s%srepo: <required> e.g. 'release-argus/Argus'",
			utils.ErrorToString(errs), prefix)
	}

	if requireErrs := l.Options.CheckValues(prefix + "  "); requireErrs != nil {
		errs = fmt.Errorf("%s  require:\\%w",
			prefix, requireErrs)
	}

	if errs != nil {
		errs = fmt.Errorf("%slatest_version:\\%w",
			prefix, errs)
	}

	return
}
