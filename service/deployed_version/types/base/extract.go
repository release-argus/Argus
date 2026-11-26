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

// Package base provides the base struct for deployed_version lookups.
package base

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/util"
)

// ExtractVersion returns the version from text using regex and template.
//
// The first match of regex is formatted with template (see [util.RegexTemplate]).
// An error is returned if regex doesn't match, or yields an empty version.
// regex must already be valid, as CheckValues verifies it compiles.
func ExtractVersion(text, regex, template string) (string, error) {
	matches := regexp.MustCompile(regex).FindAllStringSubmatch(text, 1)
	if len(matches) == 0 {
		return "", fmt.Errorf(
			"regex %q didn't return any matches on %q",
			regex, util.TruncateMessage(text, 100),
		)
	}

	version := util.RegexTemplate(matches[0], template)
	if strings.TrimSpace(version) == "" {
		return "", fmt.Errorf("regex %q produced an empty version", regex)
	}
	return version, nil
}

// QueryWithMetrics runs query, and records the metrics for the result when metrics is true.
//
// parent is the lookup embedding the receiver, used for its type label.
func (l *Lookup) QueryWithMetrics(
	parent Interface,
	query func(metrics bool, logFrom logx.LogFrom) error,
	metrics bool,
	logFrom logx.LogFrom,
) error {
	err := query(metrics, logFrom)
	if metrics {
		l.QueryMetrics(parent, err)
	}
	return err
}

// Track queries l at its configured interval until its service is being deleted.
func Track(l Interface) {
	logFrom := logx.LogFrom{Primary: l.GetServiceID()}

	for {
		if l.GetStatus().Deleting() {
			return
		}

		_ = l.Query(true, logFrom) //nolint:errcheck

		time.Sleep(l.GetOptions().GetIntervalDuration())
	}
}
