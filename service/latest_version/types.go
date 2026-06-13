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

// Package latestver provides the latest_version lookup service to for a service.
package latestver

import (
	"github.com/release-argus/Argus/service/latest_version/types/base"
)

// Lookup provides methods for retrieving the latest version of a service.
type Lookup = base.Interface

// IsEqual returns whether `this` lookup is the same as `other` (excluding status).
func IsEqual(this, other Lookup) bool {
	if other == nil || this == nil {
		// Equal if both are nil.
		return other == nil && this == nil
	}
	return this.GetOptions().String() == other.GetOptions().String() &&
		this.String("") == other.String("")
}
