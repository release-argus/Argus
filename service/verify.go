// Copyright [2025] [Argus]
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

	"github.com/release-argus/Argus/util"
)

// Print the Services in the Slice.
func (s *Slice) Print(prefix string, order []string) {
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
			fmt.Printf("%s  %s:%s%s", prefix, serviceID, delim, itemStr)
		}
	}
}

// CheckValues validates the fields of the Defaults struct.
func (d *Defaults) CheckValues(prefix string) error {
	var errs []error

	util.AppendCheckError(&errs, prefix, "options", d.Options.CheckValues(prefix+"  "))
	util.AppendCheckError(&errs, prefix, "latest_version", d.LatestVersion.CheckValues(prefix+"  "))

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// CheckValues validates the fields of each Service in the Slice.
func (s *Slice) CheckValues(prefix string) error {
	if s == nil {
		return nil
	}

	var errs []error
	keys := util.SortedKeys(*s)
	itemPrefix := prefix + "  "
	for _, key := range keys {
		util.AppendCheckError(&errs, prefix, key, (*s)[key].CheckValues(itemPrefix))
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// CheckValues validates the fields of the Service struct.
func (s *Service) CheckValues(prefix string) error {
	if s == nil {
		return nil
	}

	var errs []error
	errPrefix := prefix + "  "
	util.AppendCheckError(&errs, prefix, "options", s.Options.CheckValues(errPrefix))
	if s.LatestVersion != nil {
		util.AppendCheckError(&errs, prefix, "latest_version", s.LatestVersion.CheckValues(errPrefix))
	} else {
		util.AppendCheckError(&errs, prefix, "latest_version", errors.New(errPrefix+"latest_version is nil"))
	}
	if s.DeployedVersionLookup != nil {
		util.AppendCheckError(&errs, prefix, "deployed_version", s.DeployedVersionLookup.CheckValues(errPrefix))
	}
	util.AppendCheckError(&errs, prefix, "notify", s.Notify.CheckValues(errPrefix))
	util.AppendCheckError(&errs, prefix, "command", s.Command.CheckValues(errPrefix))
	util.AppendCheckError(&errs, prefix, "webhook", s.WebHook.CheckValues(errPrefix))
	util.AppendCheckError(&errs, prefix, "dashboard", s.Dashboard.CheckValues(errPrefix))

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
