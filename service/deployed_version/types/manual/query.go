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

// Package manual provides a manually set version lookup.
package manual

import (
	"errors"
	"time"

	"github.com/release-argus/Argus/internal/logx"
)

// applyOptions controls how a pending Version is applied.
type applyOptions struct {
	writeToDB bool // Persist the version to the database.
	announce  bool // Broadcast the version to WebSocket clients.
	rateLimit bool // Enforce the minimum gap between updates.
}

// Track applies any version given in the config.
// A manual lookup has nothing to poll, so this runs once and returns.
func (l *Lookup) Track() {
	l.ApplyConfiguredVersion()
}

// ApplyConfiguredVersion applies any version given in the config to the Status,
// persisting it without announcing it.
//
// Callers that announce the change themselves (a service create/edit) rely on this
// being silent; a refresh announces via [Lookup.Query] instead.
func (l *Lookup) ApplyConfiguredVersion() {
	logFrom := logx.LogFrom{Primary: l.GetServiceID()}

	applied, err := l.apply(applyOptions{writeToDB: true}, logFrom)
	if err != nil {
		logx.Error(err, logFrom, true)
		return
	}

	// The Version has been consumed, so save to drop it from the config file.
	if applied {
		l.Status.SendSave()
	}
}

// Query applies the manual version override if set, validating semantic versioning
// if required.
func (l *Lookup) Query(metrics bool, logFrom logx.LogFrom) error {
	_, err := l.apply(
		applyOptions{writeToDB: metrics, announce: true, rateLimit: true},
		logFrom,
	)

	return err
}

// apply moves any pending Version onto the Status, as directed by opts,
// and reports whether a Version was consumed.
func (l *Lookup) apply(opts applyOptions, logFrom logx.LogFrom) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.Version == "" {
		return false, nil
	}
	defer func() { l.Version = "" }()

	if opts.rateLimit {
		lastQueriedAt, _ := time.Parse(time.RFC3339, l.Status.DeployedVersionTimestamp())
		if time.Since(lastQueriedAt) < time.Second {
			return false, errors.New("manual version updates are rate-limited. Please try again in 1 second")
		}
	}

	// If semantic versioning is enabled, check version formatting.
	if l.Options.GetSemanticVersioning() {
		if _, err := l.Options.VerifySemanticVersioning(l.Version, logFrom); err != nil {
			return false, err //nolint:wrapcheck
		}
	}

	l.HandleNewVersion(l.Version, "", opts.writeToDB, opts.announce, logFrom)

	return true, nil
}
