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
	"strconv"
	"time"

	"github.com/release-argus/Argus/utils"
)

type Options struct {
	Active                           *bool    `yaml:"active,omitempty"`                 // Disable the service.
	Interval                         *string  `yaml:"interval,omitempty"`               // AhBmCs = Sleep A hours, B minutes and C seconds between queries.
	SemanticVersioning               *bool    `yaml:"semantic_versioning,omitempty"`    // default - true  = Version has to follow semantic versioning (https://semver.org/) and be greater than the previous to trigger anything.
	AutoApprove                      *bool    `yaml:"auto_approve,omitempty"`           // default - true = Requre approval before sending WebHook(s) for new releases
	UsePreRelease                    *bool    `yaml:"use_prerelease,omitempty"`         // Whether the prerelease tag should be used (prereleases are ignored by default)
	LatestVersionAllowInvalidCerts   *bool    `yaml:"lv_allow_invalid_certs,omitempty"` // default - false = Disallows invalid HTTPS certificates in the latest version lookup
	DeployedVersionAllowInvalidCerts *bool    `yaml:"dv_allow_invalid_certs,omitempty"` // default - false = Disallows invalid HTTPS certificates in the deployed version lookup
	Defaults                         *Options `yaml:"-"`                                // Defaults
	HardDefaults                     *Options `yaml:"-"`                                // Hard Defaults
}

// GetInterval between queries for this Service's latest version.
func (o *Options) GetInterval() string {
	return *utils.GetFirstNonDefault(o.Interval, o.Defaults.Interval, o.HardDefaults.Interval)
}

// GetSemanticVersioning will return whether Semantic Versioning should be used for this Service.
func (o *Options) GetSemanticVersioning() bool {
	return *utils.GetFirstNonDefault(o.SemanticVersioning, o.Defaults.SemanticVersioning, o.HardDefaults.SemanticVersioning)
}

// GetAutoApprove will return whether new releases should be auto-approved.
func (o *Options) GetAutoApprove() bool {
	return *utils.GetFirstNonDefault(o.AutoApprove, o.Defaults.AutoApprove, o.HardDefaults.AutoApprove)
}

// Get UsePreRelease will return whether GitHub PreReleases are considered valid for new versions.
func (o *Options) GetUsePreRelease() bool {
	return *utils.GetFirstNonDefault(o.UsePreRelease, o.Defaults.UsePreRelease, o.HardDefaults.UsePreRelease)
}

// GetIntervalPointer returns a pointer to the interval between queries on this Service's version.
func (o *Options) GetIntervalPointer() *string {
	return utils.GetFirstNonNilPtr(o.Interval, o.Defaults.Interval, o.HardDefaults.Interval)
}

// GetIntervalDuration returns the interval between queries on this Service's version.
func (o *Options) GetIntervalDuration() time.Duration {
	d, _ := time.ParseDuration(o.GetInterval())
	return d
}

// CheckValues of the Options.
func (o *Options) CheckValues(prefix string) (errs error) {
	// Interval
	if o.Interval != nil {
		// Default to seconds when an integer is provided
		if _, err := strconv.Atoi(*o.Interval); err == nil {
			*o.Interval += "s"
		}
		if _, err := time.ParseDuration(*o.Interval); err != nil {
			errs = fmt.Errorf("%s%s  interval: %q <invalid> (Use 'AhBmCs' duration format)\\",
				utils.ErrorToString(errs), prefix, *o.Interval)
		}
	}

	return
}

// Print the struct.
func (o *Options) Print(prefix string) {
	fmt.Printf("%soptions:\n", prefix)
	utils.PrintlnIfNotNil(o.Active, fmt.Sprintf("%sactive: %t", prefix, utils.DefaultIfNil(o.Active)))
	utils.PrintlnIfNotNil(o.Interval, fmt.Sprintf("%sinterval: %s", prefix, utils.DefaultIfNil(o.Interval)))
	utils.PrintlnIfNotNil(o.SemanticVersioning, fmt.Sprintf("%ssemantic_versioning: %t", prefix, utils.DefaultIfNil(o.SemanticVersioning)))
	utils.PrintlnIfNotNil(o.AutoApprove, fmt.Sprintf("%sauto_approve: %t", prefix, utils.DefaultIfNil(o.AutoApprove)))
	utils.PrintlnIfNotNil(o.UsePreRelease, fmt.Sprintf("%suse_prerelease: %t", prefix, utils.DefaultIfNil(o.UsePreRelease)))
	utils.PrintlnIfNotNil(o.LatestVersionAllowInvalidCerts, fmt.Sprintf("%slv_allow_invalid_certs: %t", prefix, utils.DefaultIfNil(o.LatestVersionAllowInvalidCerts)))
	utils.PrintlnIfNotNil(o.DeployedVersionAllowInvalidCerts, fmt.Sprintf("%sdv_allow_invalid_certs: %t", prefix, utils.DefaultIfNil(o.DeployedVersionAllowInvalidCerts)))
}
