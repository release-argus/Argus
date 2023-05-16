// Copyright [2023] [Argus]
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

package service

import (
	"fmt"

	"github.com/release-argus/Argus/util"
)

// CheckValues of the Service(s) in the Slice.
func (s *Slice) CheckValues(prefix string) (errs error) {
	keys := util.SortedKeys(*s)
	itemPrefix := prefix + "  "
	for _, key := range keys {
		if err := (*s)[key].CheckValues(itemPrefix); err != nil {
			errs = fmt.Errorf("%s%w",
				util.ErrorToString(errs), err)
		}
	}

	if errs != nil {
		errs = fmt.Errorf("%sservice:\\%s",
			prefix, util.ErrorToString(errs))
	}
	return
}

// CheckValues of the Defaults.
func (s *Defaults) CheckValues(prefix string) (errs error) {
	if optionErrs := s.Options.CheckValues(prefix); optionErrs != nil {
		errs = fmt.Errorf("%s%w",
			util.ErrorToString(errs), optionErrs)
	}
	if latestVersionErrs := s.LatestVersion.CheckValues(prefix); latestVersionErrs != nil {
		errs = fmt.Errorf("%s%w",
			util.ErrorToString(errs), latestVersionErrs)
	}

	return
}

// CheckValues of the Service.
func (s *Service) CheckValues(prefix string) (errs error) {
	errPrefix := prefix + "  "
	if optionErrs := s.Options.CheckValues(errPrefix); optionErrs != nil {
		errs = fmt.Errorf("%s%w",
			util.ErrorToString(errs), optionErrs)
	}
	if latestVersionErrs := s.LatestVersion.CheckValues(errPrefix); latestVersionErrs != nil {
		errs = fmt.Errorf("%s%w",
			util.ErrorToString(errs), latestVersionErrs)
	}
	if deployedVersionErrs := s.DeployedVersionLookup.CheckValues(errPrefix); deployedVersionErrs != nil {
		errs = fmt.Errorf("%s%w",
			util.ErrorToString(errs), deployedVersionErrs)
	}
	if notifyErrs := s.Notify.CheckValues(errPrefix); notifyErrs != nil {
		errs = fmt.Errorf("%s%w",
			util.ErrorToString(errs), notifyErrs)
	}
	if commandErrs := s.Command.CheckValues(errPrefix); commandErrs != nil {
		errs = fmt.Errorf("%s%w",
			util.ErrorToString(errs), commandErrs)
	}
	if webhookErrs := s.WebHook.CheckValues(errPrefix); webhookErrs != nil {
		errs = fmt.Errorf("%s%w",
			util.ErrorToString(errs), webhookErrs)
	}

	if errs != nil {
		errs = fmt.Errorf("%s%s:\\%w",
			prefix, s.ID, errs)
	}
	return
}

// Print the Service's in the Slice.
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
