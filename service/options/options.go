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

package opt

import (
	"fmt"
	"strconv"
	"time"

	"github.com/release-argus/Argus/util"
	"gopkg.in/yaml.v3"
)

type Options struct {
	Active             *bool    `yaml:"active,omitempty" json:"active,omitempty"`                           // Disable the service.
	Interval           string   `yaml:"interval,omitempty" json:"interval,omitempty"`                       // AhBmCs = Sleep A hours, B minutes and C seconds between queries.
	SemanticVersioning *bool    `yaml:"semantic_versioning,omitempty" json:"semantic_versioning,omitempty"` // default - true = Version has to follow semantic versioning (https://semver.org/) and be greater than the previous to trigger anything.
	Defaults           *Options `yaml:"-" json:"-"`                                                         // Defaults
	HardDefaults       *Options `yaml:"-" json:"-"`                                                         // Hard Defaults
}

// String returns a string representation of the Options.
func (o *Options) String() string {
	if o == nil {
		return "<nil>"
	}
	yamlBytes, _ := yaml.Marshal(o)
	return string(yamlBytes)
}

// GetActive status of the Service.
func (o *Options) GetActive() bool {
	return util.EvalNilPtr(o.Active, true)
}

// GetInterval between queries for this Service's latest version.
func (o *Options) GetInterval() string {
	return util.GetFirstNonDefault(
		o.Interval,
		o.Defaults.Interval,
		o.HardDefaults.Interval)
}

// GetSemanticVersioning will return whether Semantic Versioning should be used for this Service.
func (o *Options) GetSemanticVersioning() bool {
	return *util.GetFirstNonNilPtr(
		o.SemanticVersioning,
		o.Defaults.SemanticVersioning,
		o.HardDefaults.SemanticVersioning)
}

// GetIntervalPointer returns a pointer to the interval between queries on this Service's version.
func (o *Options) GetIntervalPointer() *string {
	if o.Interval != "" {
		return &o.Interval
	}
	if o.Defaults.Interval != "" {
		return &o.Defaults.Interval
	}
	return &o.HardDefaults.Interval
}

// GetIntervalDuration returns the interval between queries on this Service's version.
func (o *Options) GetIntervalDuration() time.Duration {
	d, _ := time.ParseDuration(o.GetInterval())
	return d
}

// CheckValues of the option.
func (o *Options) CheckValues(prefix string) (errs error) {
	// Interval
	if o.Interval != "" {
		// Default to seconds when an integer is provided
		if _, err := strconv.Atoi(o.Interval); err == nil {
			o.Interval += "s"
		}
		if _, err := time.ParseDuration(o.Interval); err != nil {
			errs = fmt.Errorf("%s%s  interval: %q <invalid> (Use 'AhBmCs' duration format)\\",
				util.ErrorToString(errs), prefix, o.Interval)
		}
	}

	if errs != nil {
		errs = fmt.Errorf("%soptions:\\%w",
			prefix, errs)
	}

	return
}

// Print the struct.
func (o *Options) Print(prefix string) {
	if o.Active == nil && o.Interval == "" && o.SemanticVersioning == nil {
		return
	}

	fmt.Printf("%soptions:\n", prefix)
	util.PrintlnIfNotNil(o.Active,
		fmt.Sprintf("%s  active: %t", prefix, util.DefaultIfNil(o.Active)))
	util.PrintlnIfNotDefault(o.Interval,
		fmt.Sprintf("%s  interval: %s", prefix, o.Interval))
	util.PrintlnIfNotNil(o.SemanticVersioning,
		fmt.Sprintf("%s  semantic_versioning: %t", prefix, util.DefaultIfNil(o.SemanticVersioning)))
}
