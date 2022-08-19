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

package filters

import (
	"github.com/release-argus/Argus/utils"
)

// command will run r.Command and return an err if it failed.
func (r *Require) ExecCommand(logFrom *utils.LogFrom) error {
	if r == nil || len(r.Command) == 0 {
		return nil
	}
	cmd := r.Command.ApplyTemplate(r.Status)
	return cmd.Exec(logFrom)
}
