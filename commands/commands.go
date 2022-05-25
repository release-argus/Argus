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
	"os/exec"

	"github.com/release-argus/Argus/utils"
)

// Init will set the logger for the package
func Init(log *utils.JLog) {
	jLog = log
}

func (c *Slice) Exec(logFrom *utils.LogFrom) error {
	if c == nil {
		return nil
	}

	for i := range *c {
		if err := (*c)[i].Exec(logFrom); err != nil {
			return err
		}
	}

	return nil
}

func (c *Command) Exec(logFrom *utils.LogFrom) error {
	out, err := exec.Command((*c)[0], (*c)[1:]...).Output()

	if err != nil {
		jLog.Error(utils.ErrorToString(err), *logFrom, true)
		return err
	}

	jLog.Info(string(out), *logFrom, err == nil && string(out) != "")
	return nil
}
