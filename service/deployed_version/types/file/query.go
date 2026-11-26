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

package file

import (
	"fmt"
	"strings"

	"github.com/release-argus/Argus/internal/fileaccess"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/service/deployed_version/types/base"
)

// Track polls the file and updates the deployed version.
func (l *Lookup) Track() {
	base.Track(l)
}

// Query reads the file and updates the deployed version.
func (l *Lookup) Query(metrics bool, logFrom logx.LogFrom) error {
	return l.QueryWithMetrics(l, l.query, metrics, logFrom) //nolint:wrapcheck
}

func (l *Lookup) query(writeToDB bool, logFrom logx.LogFrom) error {
	path := l.path()
	content, err := fileaccess.Read(path)
	if err != nil {
		err = fmt.Errorf("read deployed version from %q: %w", path, err)
		logx.Error(err, logFrom, true)
		return err
	}

	version := string(content)
	if l.Regex == "" {
		version = strings.TrimSpace(version)
	}
	if version == "" {
		err := fmt.Errorf("no version found in %q", path)
		logx.Warn(err, logFrom, true)
		return err
	}

	if l.Regex != "" {
		version, err = base.ExtractVersion(version, l.Regex, l.RegexTemplate)
		if err != nil {
			logx.Warn(err, logFrom, true)
			return err //nolint:wrapcheck
		}
	}

	if l.Options.GetSemanticVersioning() {
		if _, err := l.Options.VerifySemanticVersioning(version, logFrom); err != nil {
			logx.Warn(err, logFrom, true)
			return err //nolint:wrapcheck
		}
	}

	l.HandleNewVersion(version, "", writeToDB, true, logFrom)
	return nil
}
