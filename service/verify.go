// Copyright [2022] [Hymenaios]
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
	"regexp"
	"strconv"
	"time"

	"github.com/hymenaios-io/Hymenaios/utils"
)

// CheckValues of the Service(s) in the Slice.
func (s *Slice) CheckValues(prefix string) error {
	var errs error

	for key := range *s {
		var serviceErrors error
		service := (*s)[key]

		// Check Service
		if err := service.CheckValues(prefix); err != nil {
			serviceErrors = fmt.Errorf("%s%w", utils.ErrorToString(serviceErrors), err)
		}

		// Check DeployedVersionLookup
		if err := service.DeployedVersionLookup.CheckValues(prefix + "  "); err != nil {
			serviceErrors = fmt.Errorf("%s%w", utils.ErrorToString(serviceErrors), err)
		}

		// Check Gotify(s)
		if err := service.Gotify.CheckValues(prefix + "  "); err != nil {
			serviceErrors = fmt.Errorf("%s%w", utils.ErrorToString(serviceErrors), err)
		}

		// Check Slack(s)
		if err := service.Slack.CheckValues(prefix + "  "); err != nil {
			serviceErrors = fmt.Errorf("%s%w", utils.ErrorToString(serviceErrors), err)
		}

		// Check WebHook(s)
		if err := service.WebHook.CheckValues(prefix + "  "); err != nil {
			serviceErrors = fmt.Errorf("%s%w", utils.ErrorToString(serviceErrors), err)
		}

		if serviceErrors != nil {
			errs = fmt.Errorf("%s  %s:\\%w", utils.ErrorToString(errs), key, serviceErrors)
		}
	}
	return errs
}

// CheckValues of the Service.
func (s *Service) CheckValues(prefix string) (errs error) {
	// Interval
	if s.Interval != nil {
		// Default to seconds when an integer is provided
		if _, err := strconv.Atoi(*s.Interval); err == nil {
			*s.Interval += "s"
		}
		if _, err := time.ParseDuration(*s.Interval); err != nil {
			errs = fmt.Errorf("%s%s  interval: <invalid> %q (Use 'AhBmCs' duration format)\\", utils.ErrorToString(errs), prefix, *s.Interval)
		}
	}

	// Type
	if s.Defaults != nil {
		if s.Type == nil {
			errs = fmt.Errorf("%stype: <missing> (Services require a type)\\", utils.ErrorToString(errs))
		} else if *s.Type != "github" && *s.Type != "url" {
			errs = fmt.Errorf("%s%s  type: <invalid> %q (Should be either 'github' or 'url')\\", utils.ErrorToString(errs), prefix, *s.Type)
		}
	}

	// RegEx
	if s.RegexContent != nil {
		_, err := regexp.Compile(*s.RegexContent)
		if err != nil {
			errs = fmt.Errorf("%s%s  regex_content: <invalid> %q (Invalid RegEx)\\", utils.ErrorToString(errs), prefix, *s.RegexContent)
		}
	}
	if s.RegexVersion != nil {
		_, err := regexp.Compile(*s.RegexVersion)
		if err != nil {
			errs = fmt.Errorf("%s%s  regex_version: <invalid> %q (Invalid RegEx)\\", utils.ErrorToString(errs), prefix, *s.RegexVersion)
		}
	}

	// Status
	if s.Status != nil {
		var statusErrs error
		if s.Status.CurrentVersionTimestamp != nil {
			_, err := time.Parse(time.RFC3339, *s.Status.CurrentVersionTimestamp)
			if err != nil {
				statusErrs = fmt.Errorf("%s%s    current_version_timestamp: <invalid> %q (Failed to convert to RFC3339 format)\\", utils.ErrorToString(errs), prefix, *s.Status.CurrentVersionTimestamp)
			}
		}
		if s.Status.LatestVersionTimestamp != nil {
			_, err := time.Parse(time.RFC3339, *s.Status.LatestVersionTimestamp)
			if err != nil {
				statusErrs = fmt.Errorf("%s%s    latest_version_timestamp: <invalid> %q (Failed to convert to RFC3339 format)\\", utils.ErrorToString(errs), prefix, *s.Status.LatestVersionTimestamp)
			}
		}
		if statusErrs != nil {
			errs = fmt.Errorf("%s%s  status:\\%s", utils.ErrorToString(errs), prefix, statusErrs)
		}
	}

	// URL Commands
	if err := s.URLCommands.CheckValues(prefix + "  "); err != nil {
		errs = fmt.Errorf("%s%w", utils.ErrorToString(errs), err)
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
	// Service.
	utils.PrintlnIfNotNil(s.Type, fmt.Sprintf("%stype: %s", prefix, utils.DefaultIfNil(s.Type)))
	utils.PrintlnIfNotNil(s.URL, fmt.Sprintf("%surl: %s", prefix, utils.DefaultIfNil(s.URL)))
	utils.PrintlnIfNotNil(s.AllowInvalidCerts, fmt.Sprintf("%sallow_invalid_certs: %t", prefix, utils.DefaultIfNil(s.AllowInvalidCerts)))
	utils.PrintlnIfNotNil(s.AccessToken, fmt.Sprintf("%saccess_token: %s", prefix, utils.DefaultIfNil(s.AccessToken)))
	utils.PrintlnIfNotNil(s.SemanticVersioning, fmt.Sprintf("%ssemantic_versioning: %t", prefix, utils.DefaultIfNil(s.SemanticVersioning)))
	utils.PrintlnIfNotNil(s.Interval, fmt.Sprintf("%sinterval: %s", prefix, utils.DefaultIfNil(s.Interval)))
	if s.URLCommands != nil {
		fmt.Printf("%surl_commands:\n", prefix)
		s.URLCommands.Print(prefix)
	}
	utils.PrintlnIfNotNil(s.RegexContent, fmt.Sprintf("%sregex_content: %s", prefix, utils.DefaultIfNil(s.RegexContent)))
	utils.PrintlnIfNotNil(s.RegexVersion, fmt.Sprintf("%sregex_version: %s", prefix, utils.DefaultIfNil(s.RegexVersion)))
	utils.PrintlnIfNotNil(s.UsePreRelease, fmt.Sprintf("%suse_prerelease: %t", prefix, utils.DefaultIfNil(s.UsePreRelease)))
	utils.PrintlnIfNotNil(s.WebURL, fmt.Sprintf("%sweb_url: %s", prefix, utils.DefaultIfNil(s.WebURL)))
	utils.PrintlnIfNotNil(s.AutoApprove, fmt.Sprintf("%sauto_approve: %t", prefix, utils.DefaultIfNil(s.AutoApprove)))
	utils.PrintlnIfNotNil(s.IgnoreMisses, fmt.Sprintf("%signore_misses: %t", prefix, utils.DefaultIfNil(s.IgnoreMisses)))
	utils.PrintlnIfNotNil(s.Icon, fmt.Sprintf("%sicon: %s", prefix, utils.DefaultIfNil(s.Icon)))

	s.DeployedVersionLookup.Print(prefix)

	if s.Status != nil && *s.Status != (Status{}) {
		fmt.Printf("%sstatus:\n", prefix)
		s.Status.Print(prefix + "  ")
	}

	// Gotify.
	s.Gotify.Print(prefix + "  ")

	// Slack.
	s.Slack.Print(prefix + "  ")

	// WebHook.
	s.WebHook.Print(prefix + "  ")
}
