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

	"github.com/release-argus/Argus/utils"
)

// CheckValues of the Service(s) in the Slice.
func (s *Slice) CheckValues(prefix string) error {
	var errs error

	for key := range *s {
		var serviceErrors error
		service := (*s)[key]

		// Check Service
		if err := service.CheckValues(prefix); err != nil {
			serviceErrors = fmt.Errorf("%s%w",
				utils.ErrorToString(serviceErrors), err)
		}

		// URL Commands
		if err := service.LatestVersion.URLCommandsCheckValues(prefix + "  "); err != nil {
			errs = fmt.Errorf("%s%w",
				utils.ErrorToString(errs), err)
		}

		// Check DeployedVersionLookup
		if err := service.DeployedVersionLookup.CheckValues(prefix + "  "); err != nil {
			serviceErrors = fmt.Errorf("%s%w",
				utils.ErrorToString(serviceErrors), err)
		}

		// Check Notify(s)
		if err := service.Notify.CheckValues(prefix + "  "); err != nil {
			serviceErrors = fmt.Errorf("%s%w",
				utils.ErrorToString(serviceErrors), err)
		}

		// Check WebHook(s)
		if err := service.WebHook.CheckValues(prefix + "  "); err != nil {
			serviceErrors = fmt.Errorf("%s%w",
				utils.ErrorToString(serviceErrors), err)
		}

		if serviceErrors != nil {
			errs = fmt.Errorf("%s  %s:\\%w",
				utils.ErrorToString(errs), key, serviceErrors)
		}
	}
	return errs
}

// CheckValues of the Service.
func (s *Service) CheckValues(prefix string) (errs error) {
	if optionsErrs := s.Options.CheckValues(prefix + "  "); optionsErrs != nil {
		errs = fmt.Errorf("%soptions:\\%w",
			prefix, optionsErrs)
	}

	if latestVersionErrs := s.Options.CheckValues(prefix + "  "); latestVersionErrs != nil {
		errs = fmt.Errorf("%slatest_version:\\%w",
			prefix, latestVersionErrs)
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
		(*s)[serviceID].Print(prefix + "  ")
	}
}

// Print will print the Service.
func (s *Service) Print(prefix string) {
	fmt.Printf("%s%s:\n", prefix, s.ID)
	prefix += "  "

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

	// WebHook.
	s.WebHook.Print(prefix)
}
