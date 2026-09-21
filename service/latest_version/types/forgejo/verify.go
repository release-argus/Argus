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
				Value:       util.EvalEnvVars(l.Host),
				Description: problem,
			})
	}

	if !isOwnerRepo(l.URL) {
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

// isOwnerRepo reports whether path is an "owner/repo" the API can be addressed
// with.
func isOwnerRepo(path string) bool {
	owner, repo, found := strings.Cut(path, "/")

	return found &&
		owner != "" && repo != "" &&
		!strings.Contains(repo, "/") &&
		path == strings.TrimSpace(path)
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

	parsed, err := l.parseHost()
	if err != nil {
		return "not a valid URL"
	}

	switch {
	case parsed.Scheme != "http" && parsed.Scheme != "https":
		return "scheme must be http or https"
	case parsed.Hostname() == "":
		return "no hostname"
	case strings.HasSuffix(parsed.Path, "/"):
		return "trailing '/'"
	}

	return ""
}
