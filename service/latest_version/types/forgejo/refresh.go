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

package forgejo

import (
	"github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/util"
)

// InheritSecrets inherits a referenced access token from fromLookup when the
// host is unchanged, then delegates to the base.
func (l *Lookup) InheritSecrets(fromLookup base.BaseInterface, secretRefs *shared.VSecretRef) {
	if l.AccessToken == util.SecretValue {
		l.AccessToken = ""

		// Resolved, so that a name and the URL it names are one instance.
		if oldLookup, ok := fromLookup.(*Lookup); ok &&
			canonicalHost(l.resolveHost()) == canonicalHost(oldLookup.resolveHost()) {
			l.AccessToken = oldLookup.AccessToken
		}
	}

	l.Lookup.InheritSecrets(fromLookup, secretRefs)
}
