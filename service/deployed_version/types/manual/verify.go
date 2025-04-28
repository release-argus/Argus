// Copyright [2025] [Argus]
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

// Package manual provides a manually set version lookup.
package manual

import (
	logutil "github.com/release-argus/Argus/util/log"
)

// CheckValues validates the fields of the Lookup struct.
func (l *Lookup) CheckValues(prefix string) error {
	logFrom := logutil.LogFrom{Primary: l.GetServiceID()}
	return l.Query(false, logFrom)
}
