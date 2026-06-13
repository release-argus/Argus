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

// Package github provides a github-based lookup type.
package github

import (
	"github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/util"
)

// InheritSecrets will inherit values from `fromLookup` if the values should query the same source.
//
//	Values: githubData, Require.
func (l *Lookup) InheritSecrets(fromLookup base.BaseInterface, secretRefs *shared.VSecretRef) {
	// Check whether inheriting from a GitHub Interface.
	if oldGitHubLookup, ok := fromLookup.(*Lookup); ok {
		// AccessToken
		if l.AccessToken == util.SecretValue {
			l.AccessToken = oldGitHubLookup.AccessToken
		}
		// Querying the same GitHub repo, and the ETag differs.
		if l.URL == oldGitHubLookup.URL && l.data.ETag() != oldGitHubLookup.data.ETag() {
			// Inherit the GitHub data.
			l.data.CopyFrom(&oldGitHubLookup.data)
		}
	}

	l.Lookup.InheritSecrets(fromLookup, secretRefs)
}
