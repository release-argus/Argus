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
	"errors"
	"strings"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/util"
)

// CommonDefaults are the Forgejo default values that apply to every instance.
type CommonDefaults struct {
	UsePreRelease *bool `json:"use_prerelease,omitzero" yaml:"use_prerelease,omitzero"` // Whether releases with prerelease tag are considered.
}

// IsZero implements the yaml.IsZeroer interface.
func (d CommonDefaults) IsZero() bool {
	return d.UsePreRelease == nil
}

// HostDefaults are the default values for a single forge instance.
type HostDefaults struct {
	URL               string `json:"url,omitzero" yaml:"url,omitzero"`                                 // Instance to query, e.g. "https://codeberg.org".
	AccessToken       string `json:"access_token,omitzero" yaml:"access_token,omitzero"`               // Access token to use.
	AllowInvalidCerts *bool  `json:"allow_invalid_certs,omitzero" yaml:"allow_invalid_certs,omitzero"` // Allow invalid SSL certificates.
}

// IsZero implements the yaml.IsZeroer interface.
func (d HostDefaults) IsZero() bool {
	return d.URL == "" &&
		d.AccessToken == "" &&
		d.AllowInvalidCerts == nil
}

// inherit fills the receiver's unset fields from entry.
func (d *HostDefaults) inherit(entry HostDefaults) {
	if d.URL == "" {
		d.URL = entry.URL
	}
	if d.AccessToken == "" {
		d.AccessToken = entry.AccessToken
	}
	if d.AllowInvalidCerts == nil {
		d.AllowInvalidCerts = entry.AllowInvalidCerts
	}
}

// Defaults are the Forgejo-specific default values for a Lookup.
//
// Instances are keyed by a name, which only labels them.
type Defaults struct {
	Common CommonDefaults          `json:"common,omitzero" yaml:"common,omitzero"`
	Host   map[string]HostDefaults `json:"host,omitempty" yaml:"host,omitempty"`

	hardDefaults *Defaults
}

// SetDefaults gives the receiver a layer to fall back on.
func (d *Defaults) SetDefaults(dflts *Defaults) {
	if d == nil {
		return
	}

	d.hardDefaults = dflts
}

// resolvedHost returns the defaults `name` resolves to across the receiver and the
// layers it falls back on, and whether any of them names it.
//
// Field by field - one layer may name the instance while another holds the
// credential that reaches it.
func (d *Defaults) resolvedHost(name string) (resolved HostDefaults, named bool) {
	for layer := d; layer != nil; layer = layer.hardDefaults {
		entry, matched := layer.hostEntry(name)
		if !matched {
			continue
		}
		named = true

		resolved.inherit(entry)
	}

	return
}

// IsZero implements the yaml.IsZeroer interface.
func (d Defaults) IsZero() bool {
	return d.Common.IsZero() &&
		len(d.Host) == 0
}

// Default sets the values of the receiver to their default values.
func (d *Defaults) Default() {
	usePreRelease := false
	d.Common.UsePreRelease = &usePreRelease
}

// hostEntry returns the entry `name` addresses, compared case-insensitively.
//
// Matching is never by URL - several entries may address one instance, each
// with its own token, so a URL identifies none of them.
func (d *Defaults) hostEntry(name string) (HostDefaults, bool) {
	if d == nil || name == "" {
		return HostDefaults{}, false
	}

	for key, entry := range d.Host {
		if strings.EqualFold(key, name) {
			return entry, true
		}
	}

	return HostDefaults{}, false
}

// CheckValues validates the fields of the receiver.
//
// Entries may address one instance - a forge can hold several credentials, each
// reaching different repositories.
func (d *Defaults) CheckValues() error {
	if d == nil || len(d.Host) == 0 {
		return nil
	}

	var errs []error
	spellings := make(map[string][]string, len(d.Host))
	for _, name := range util.SortedKeys(d.Host) {
		lower := strings.ToLower(name)
		spellings[lower] = append(spellings[lower], name)
	}

	for _, name := range util.SortedKeys(d.Host) {
		if strings.TrimSpace(name) == "" {
			errs = append(errs, errors.New("an instance must be named"))
			continue
		}
		if name != strings.TrimSpace(name) {
			errs = append(errs,
				&decode.ErrKey{
					Key:         name,
					Description: "surrounded by whitespace",
				})
			continue
		}

		if others := spellings[strings.ToLower(name)]; len(others) > 1 {
			errs = append(errs, &decode.ErrKey{
				Key:         name,
				Description: "already used, names are case-insensitive",
			})
			continue
		}

		// Only the layer that has one to fall back on can tell that no URL is coming.
		resolved, _ := d.resolvedHost(name)
		if resolved.URL == "" {
			if d.hardDefaults != nil {
				errs = append(errs, &decode.ErrKeyField{
					Key: name,
					Err: &decode.ErrField{
						Key:         "url",
						Description: "e.g. https://codeberg.org",
					},
				})
			}

			continue
		}

		if resolved.AccessToken != "" && isPlaintext(util.EvalEnvVars(resolved.URL)) {
			logx.Warn(
				"access_token will be sent unencrypted as the url uses http",
				logx.LogFrom{Primary: "latest_version/forgejo", Secondary: name},
				true,
			)
		}

		ownURL := d.Host[name].URL
		if ownURL == "" {
			continue
		}

		if problem := urlProblem(util.EvalEnvVars(ownURL)); problem != "" {
			errs = append(errs, &decode.ErrKeyField{
				Key: name,
				Err: &decode.ErrField{
					Key:         "url",
					Value:       ownURL,
					Description: problem,
				},
			})
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return &decode.ErrKeyField{
		Key: "host",
		Err: errors.Join(errs...),
	}
}
