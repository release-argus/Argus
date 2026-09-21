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

// Package forgejo provides a Forgejo-based lookup type.
package forgejo

import (
	"net/url"

	"github.com/release-argus/Argus/util"
)

// GetType returns the type of the receiver.
func (l *Lookup) GetType() string {
	return Type
}

// ServiceURL returns the repository's URL on the configured instance.
func (l *Lookup) ServiceURL() string {
	host, err := l.parseHost()
	if err != nil || host.Hostname() == "" {
		return l.URL
	}

	return (&url.URL{
		Scheme: host.Scheme,
		Host:   host.Host,
		Path:   host.Path + "/" + l.URL,
	}).String()
}

// resolveHost returns the instance URL [Lookup.Host] addresses.
//
// Host may name a defaults entry or be a URL itself; a name wins, so every
// consumer must resolve through here rather than reading Host directly.
func (l *Lookup) resolveHost() string {
	if url, named := l.namedInstance(); named && url != "" {
		return url
	}

	return l.Host
}

// namedInstance reports whether Host names a defaults entry, and the URL that
// entry gives the instance.
func (l *Lookup) namedInstance() (url string, named bool) {
	for _, defaults := range []*Defaults{l.typeDefaults, l.typeHardDefaults} {
		entry, matched := defaults.hostEntry(l.Host)
		if !matched {
			continue
		}
		named = true
		if entry.URL != "" {
			return entry.URL, true
		}
	}

	return
}

// hostDefaults resolves the defaults for the configured host.
func (l *Lookup) hostDefaults() (resolved HostDefaults) {
	for _, defaults := range []*Defaults{l.typeDefaults, l.typeHardDefaults} {
		entry, matched := defaults.hostEntry(l.Host)
		if !matched {
			continue
		}

		if resolved.URL == "" {
			resolved.URL = entry.URL
		}
		if resolved.AccessToken == "" {
			resolved.AccessToken = entry.AccessToken
		}
		if resolved.AllowInvalidCerts == nil {
			resolved.AllowInvalidCerts = entry.AllowInvalidCerts
		}
	}

	return
}

// accessToken resolves the access token to send to the configured host.
//
// A host matching no entry is never sent a credential.
func (l *Lookup) accessToken() string {
	if l.AccessToken != "" {
		return util.EvalEnvVars(l.AccessToken)
	}

	return util.EvalEnvVars(l.hostDefaults().AccessToken)
}

// allowInvalidCerts resolves whether invalid HTTPS certificates are allowed for
// the configured host.
func (l *Lookup) allowInvalidCerts() bool {
	if l.AllowInvalidCerts != nil {
		return *l.AllowInvalidCerts
	}

	return util.DerefOrZero(l.hostDefaults().AllowInvalidCerts)
}

// usePreRelease resolves whether to consider PreReleases for new versions.
func (l *Lookup) usePreRelease() bool {
	return *util.FirstNonDefault(
		l.UsePreRelease,
		l.typeDefaults.Common.UsePreRelease,
		l.typeHardDefaults.Common.UsePreRelease,
	)
}

// useTagsAPI returns whether the [endpointTags] API may be used as a fallback.
//
// Cannot use tags when:
//   - filtering out pre-releases (tags have no pre-release labeling)
//   - filtering on regex_content (tags have no release assets to match against)
func (l *Lookup) useTagsAPI() bool {
	return !l.usePreRelease() && (l.Require == nil || l.Require.RegexContent == "")
}
