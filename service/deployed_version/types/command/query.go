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

// Package command provides a command-based version lookup.
package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/util"
)

// Track polls the command at the configured interval, updating the deployed version on each query.
func (l *Lookup) Track() {
	logFrom := logx.LogFrom{Primary: l.GetServiceID()}

	// Track forever.
	for {
		// If we are deleting this Service, stop tracking it.
		if l.Status.Deleting() {
			return
		}

		// Query the deployed version.
		_ = l.Query(true, logFrom) //nolint:errcheck

		// Sleep interval between queries.
		time.Sleep(l.Options.GetIntervalDuration())
	}
}

// Query queries the command for the deployed version, sets Prometheus metrics if requested, and returns any error.
func (l *Lookup) Query(metrics bool, logFrom logx.LogFrom) error {
	err := l.query(metrics, logFrom)

	if metrics {
		l.QueryMetrics(l, err)
	}

	return err
}

// query runs the command and updates DeployedVersion if changed.
func (l *Lookup) query(writeToDB bool, logFrom logx.LogFrom) error {
	version, err := l.getVersion(logFrom)
	if err != nil {
		return err
	}

	// Set the deployed version if it has changed.
	l.HandleNewVersion(version, "", writeToDB, true, logFrom) //nolint:wrapcheck

	return nil
}

// getVersion runs the command, extracts the version from its stdout, and returns it.
func (l *Lookup) getVersion(logFrom logx.LogFrom) (string, error) {
	l.mu.RLock()
	commandSlice := slices.Clone(l.Command)
	regex := l.Regex
	l.mu.RUnlock()

	if len(commandSlice) == 0 {
		err := errors.New("no command specified for deployed_version query")
		logx.Warn(err, logFrom, true)
		return "", err
	}

	// Create a context with a timeout so a hung command cannot block the interval.
	ctx, cancel := context.WithTimeout(context.Background(), defaultCommandTimeout)
	defer cancel()

	// Eval env vars in each arg (e.g. ${INTERVAL}).
	for i, arg := range commandSlice {
		commandSlice[i] = util.EvalEnvVars(arg)
	}

	command := exec.CommandContext(ctx, commandSlice[0], commandSlice[1:]...) //nolint:gosec
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		err = fmt.Errorf(
			"failed running command %q: %w: %s",
			strings.Join(commandSlice, " "), err, strings.TrimSpace(stderr.String()),
		)
		logx.Warn(err, logFrom, true)
		return "", err
	}

	version := strings.TrimSpace(stdout.String())
	if version == "" {
		err := fmt.Errorf(
			"no version found in command %q output",
			strings.Join(commandSlice, " "),
		)
		logx.Warn(err, logFrom, true)
		return "", err
	}

	// If a regex is provided, use it to extract the version.
	if regex != "" {
		re := regexp.MustCompile(regex)
		texts := re.FindAllStringSubmatch(version, 1)

		if len(texts) == 0 {
			err := fmt.Errorf(
				"regex %q didn't return any matches on %q",
				regex, util.TruncateMessage(version, 100),
			)
			logx.Warn(err, logFrom, true)
			return "", err
		}

		regexMatches := texts[0]
		version = util.RegexTemplate(regexMatches, "")
	}

	// If semantic versioning is enabled, check the version is in the correct format.
	if l.Options.GetSemanticVersioning() {
		if _, err := l.Options.VerifySemanticVersioning(version, logFrom); err != nil {
			logx.Warn(err, logFrom, true)
			return "", err //nolint:wrapcheck
		}
	}

	return version, nil
}