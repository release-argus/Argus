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

package opt

import (
	"fmt"
	"strconv"
	"time"

	"github.com/release-argus/Argus/util"
)

// OptionsBase is the base struct for Options.
type OptionsBase struct {
	Interval           string `yaml:"interval,omitempty" json:"interval,omitempty"`                       // AhBmCs = Sleep A hours, B minutes and C seconds between queries.
	SemanticVersioning *bool  `yaml:"semantic_versioning,omitempty" json:"semantic_versioning,omitempty"` // default - true = Version has to follow semantic versioning (https://semver.org/) and be greater than the previous to trigger anything.
}

// OptionsDefaults are the default values for Options.
type OptionsDefaults struct {
	OptionsBase `yaml:",inline" json:",inline"`
}

// NewDefaults returns a new OptionsDefaults.
func NewDefaults(
	interval string,
	semanticVersioning *bool,
) *OptionsDefaults {
	return &OptionsDefaults{
		OptionsBase: OptionsBase{
			Interval:           interval,
			SemanticVersioning: semanticVersioning}}
}

type Options struct {
	OptionsBase `yaml:",inline" json:",inline"`

	Active *bool `yaml:"active,omitempty" json:"active,omitempty"` // Disable the service.

	Defaults     *OptionsDefaults `yaml:"-" json:"-"` // Defaults
	HardDefaults *OptionsDefaults `yaml:"-" json:"-"` // Hard Defaults
}

// New Options.
func New(
	active *bool,
	interval string,
	semanticVersioning *bool,
	defaults, hardDefaults *OptionsDefaults,
) *Options {
	return &Options{
		OptionsBase: OptionsBase{
			Interval:           interval,
			SemanticVersioning: semanticVersioning},
		Active:       active,
		Defaults:     defaults,
		HardDefaults: hardDefaults}
}

// String returns a string representation of the Options.
func (o *Options) String() (str string) {
	if o != nil {
		str = util.ToYAMLString(o, "")
	}
	return
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
func (o *OptionsBase) CheckValues(prefix string) (errs error) {
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
