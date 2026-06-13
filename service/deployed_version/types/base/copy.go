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

package base

import "github.com/release-argus/Argus/service/status"

// Clone returns a deep copy of the receiver with the given service status.
func (l *Lookup) Clone(svcStatus *status.Status) *Lookup {
	if l == nil {
		return nil
	}

	return &Lookup{
		Type:         l.Type,
		Options:      l.Options.Copy(),
		Status:       svcStatus,
		Defaults:     l.Defaults,
		HardDefaults: l.HardDefaults,
	}
}

// Copy returns a deep copy of the receiver as a [BaseInterface].
func (l *Lookup) Copy(svcStatus *status.Status) BaseInterface {
	if got := l.Clone(svcStatus); got != nil {
		return got
	}
	return nil
}
