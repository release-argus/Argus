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
	"errors"
	"strings"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/service/latest_version/types/forge"
	"github.com/release-argus/Argus/util"
)

// CheckValues validates the fields of the receiver.
func (l *Lookup) CheckValues() error {
	var errs []error

	l.Host = strings.TrimRight(l.Host, "/")

	if problem := l.hostProblem(); problem != "" {
		errs = append(errs,
			&decode.ErrField{
				Key:         "host",
				Value:       l.Host,
				Description: problem,
			})
	}

	if l.accessToken() != "" && isPlaintext(util.EvalEnvVars(l.resolveHost())) {
		logx.Warn(
			"access_token will be sent unencrypted, as the host uses http",
			logx.LogFrom{Primary: "latest_version", Secondary: l.GetServiceID()},
			true,
		)
	}

	if !forge.IsOwnerRepo(util.EvalEnvVars(l.URL)) {
		errs = append(errs,
			&decode.ErrField{
				Key:         "url",
				Value:       l.URL,
				Description: "e.g. owner/repo",
			})
	}

	if baseErrs := l.Lookup.CheckValues(); baseErrs != nil {
		errs = append(errs, baseErrs)
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// hostProblem returns what to tell the user about [Lookup.Host], or an empty
// string when it is usable.
func (l *Lookup) hostProblem() string {
	raw := util.EvalEnvVars(l.Host)
	switch {
	case strings.TrimSpace(raw) == "":
		return "e.g. https://codeberg.org"
	case raw != strings.TrimSpace(raw):
		return "surrounded by whitespace"
	}

	// A name only labels an instance, however it is spelled, so an entry without
	// a URL addresses nothing.
	if url, named := l.namedInstance(); named && url == "" {
		return "names an instance with no url"
	}

	return urlProblem(util.EvalEnvVars(l.resolveHost()))
}
