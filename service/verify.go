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

// Package service provides the service functionality for Argus.
package service

import (
	"errors"
	"fmt"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/util"
)

// Print writes each Service to stdout with the given prefix.
func (s *Services) Print(prefix string, order []string) {
	if s == nil {
		return
	}

	fmt.Printf("%sservice:\n", prefix)
	itemPrefix := prefix + "    "
	for _, serviceID := range order {
		itemStr := (*s)[serviceID].String(itemPrefix)
		if itemStr != "" {
			delim := "\n"
			if itemStr == "{}\n" {
				delim = " "
			}
			fmt.Printf(
				"%s  %s:%s%s",
				prefix,
				serviceID,
				delim,
				itemStr,
			)
		}
	}
}

// CheckValues validates the fields of the receiver.
func (d *Defaults) CheckValues() error {
	var errs []error

	if err := d.Options.CheckValues(); err != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "options",
				Err: err,
			},
		)
	}
	if err := d.LatestVersion.CheckValues(); err != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "latest_version",
				Err: err,
			},
		)
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// CheckValues validates the fields of each [Service].
func (s *Services) CheckValues() (error, bool) {
	if s == nil {
		return nil, false
	}

	var errs []error
	changed := false
	keys := util.SortedKeys(*s)
	for _, key := range keys {
		err, keyChanged := (*s)[key].CheckValues()
		if err != nil {
			errs = append(
				errs,
				&decode.KeyFieldError{
					Key: key,
					Err: err,
				},
			)
		}
		changed = changed || keyChanged
	}

	if len(errs) == 0 {
		return nil, changed
	}
	return errors.Join(errs...), false
}

// CheckValues validates the fields of the receiver.
func (s *Service) CheckValues() (error, bool) {
	if s == nil {
		return nil, false
	}

	var errs []error
	if err := s.Options.CheckValues(); err != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "options",
				Err: err,
			},
		)
	}
	if s.LatestVersion != nil {
		if err := s.LatestVersion.CheckValues(); err != nil {
			errs = append(
				errs,
				&decode.KeyFieldError{
					Key: "latest_version",
					Err: err,
				},
			)
		}
	}
	if s.DeployedVersionLookup != nil {
		if err := s.DeployedVersionLookup.CheckValues(); err != nil {
			errs = append(
				errs,
				&decode.KeyFieldError{
					Key: "deployed_version",
					Err: err,
				},
			)
		}
	}
	if s.LatestVersion == nil && s.DeployedVersionLookup == nil {
		return errors.New("latest_version and/or deployed_version required"), false
	}
	notifyErr, notifyChanged := s.Notify.CheckValues()
	if notifyErr != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "notify",
				Err: notifyErr,
			},
		)
	}
	if err := s.Command.CheckValues(); err != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "command",
				Err: err,
			},
		)
	}
	webhookErr, webhookChanged := s.WebHook.CheckValues()
	if webhookErr != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "webhook",
				Err: webhookErr,
			},
		)
	}
	if err := s.Dashboard.CheckValues(); err != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "dashboard",
				Err: err,
			},
		)
	}

	if len(errs) == 0 {
		return nil, notifyChanged || webhookChanged
	}
	return errors.Join(errs...), false
}
