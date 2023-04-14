// Copyright [2022] [Argus]
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
	for _, key := range keys {
		if serviceErrs := (*s)[key].CheckValues(prefix); serviceErrs != nil {
			errs = fmt.Errorf("%s%w",
				util.ErrorToString(errs), serviceErrs)
		}
	}
	return
}

// CheckValues of the Service.
func (s *Service) CheckValues(prefix string) (errs error) {
	if optionErrs := s.Options.CheckValues(prefix + "  "); optionErrs != nil {
		errs = fmt.Errorf("%s%w",
			util.ErrorToString(errs), optionErrs)
	}
	if latestVersionErrs := s.LatestVersion.CheckValues(prefix + "  "); latestVersionErrs != nil {
		errs = fmt.Errorf("%s%w",
			util.ErrorToString(errs), latestVersionErrs)
	}
	if deployedVersionErrs := s.DeployedVersionLookup.CheckValues(prefix + "  "); deployedVersionErrs != nil {
		errs = fmt.Errorf("%s%w",
			util.ErrorToString(errs), deployedVersionErrs)
	}
	if notifyErrs := s.Notify.CheckValues(prefix + "  "); notifyErrs != nil {
		errs = fmt.Errorf("%s%w",
			util.ErrorToString(errs), notifyErrs)
	}
	if commandErrs := s.Command.CheckValues(prefix + "  "); commandErrs != nil {
		errs = fmt.Errorf("%s%w",
			util.ErrorToString(errs), commandErrs)
	}
	if webhookErrs := s.WebHook.CheckValues(prefix + "  "); webhookErrs != nil {
		errs = fmt.Errorf("%s%w",
			util.ErrorToString(errs), webhookErrs)
	}

	if errs != nil && s.Defaults != nil {
		errs = fmt.Errorf("  %s:\\%w",
			s.ID, errs)
	}
	return
}

// Print will print the Service's in the Slice.
func (s *Slice) Print(prefix string, order []string) {
	if s == nil {
		return
	}

	fmt.Printf("%sservice:\n", prefix)
	for _, serviceID := range order {
		fmt.Printf("%s  %s:\n", prefix, serviceID)
		(*s)[serviceID].Print(prefix + "    ")
	}
}

// Print will print the Service.
func (s *Service) Print(prefix string) {
	util.PrintlnIfNotDefault(s.Comment,
		fmt.Sprintf("%scomment: %q", prefix, s.Comment))

	// Options
	s.Options.Print(prefix)

	// Latest Version
	s.LatestVersion.Print(prefix)

	// Dashboard
	s.Dashboard.Print(prefix)

	// Deployed Version
	s.DeployedVersionLookup.Print(prefix)

	// Notify.
	s.Notify.Print(prefix)

	// Command.
	s.Command.Print(prefix)

	// WebHook.
	s.WebHook.Print(prefix)
}
