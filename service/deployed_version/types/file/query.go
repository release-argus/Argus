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
	"regexp"
	"strings"
	"time"

	"github.com/release-argus/Argus/internal/fileaccess"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/util"
)

// Track polls the file and updates the deployed version.
func (l *Lookup) Track() {
	logFrom := logx.LogFrom{Primary: l.GetServiceID()}
	for {
		if l.Status.Deleting() {
			return
		}
		_ = l.Query(true, logFrom) //nolint:errcheck
		time.Sleep(l.Options.GetIntervalDuration())
	}
}

// Query reads the file and updates the deployed version.
func (l *Lookup) Query(metrics bool, logFrom logx.LogFrom) error {
	err := l.query(metrics, logFrom)
	if metrics {
		l.QueryMetrics(l, err)
	}
	return err
}

func (l *Lookup) query(writeToDB bool, logFrom logx.LogFrom) error {
	if err := l.CheckValues(); err != nil {
		logx.Error(err, logFrom, true)
		return err
	}

	content, err := fileaccess.Read(l.Path)
	if err != nil {
		err = fmt.Errorf("read deployed version from %q: %w", l.Path, err)
		logx.Error(err, logFrom, true)
		return err
	}

	version := string(content)
	if l.Regex == "" {
		version = strings.TrimSpace(version)
	}
	if version == "" {
		err := fmt.Errorf("no version found in %q", l.Path)
		logx.Warn(err, logFrom, true)
		return err
	}

	if l.Regex != "" {
		matches := regexp.MustCompile(l.Regex).FindAllStringSubmatch(version, 1)
		if len(matches) == 0 {
			err := fmt.Errorf(
				"regex %q didn't return any matches on %q",
				l.Regex, util.TruncateMessage(version, 100),
			)
			logx.Warn(err, logFrom, true)
			return err
		}
		version = strings.TrimSpace(util.RegexTemplate(matches[0], l.RegexTemplate))
		if version == "" {
			err := fmt.Errorf("regex %q produced an empty version in %q", l.Regex, l.Path)
			logx.Warn(err, logFrom, true)
			return err
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
