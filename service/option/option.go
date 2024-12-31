// Copyright [2024] [Argus]
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

// Package option provides options for a service.
package option

import (
	"fmt"
	"strconv"
	"time"

	"github.com/release-argus/Argus/util"
)

// Base is the base struct for Options.
type Base struct {
	Interval           string `yaml:"interval,omitempty" json:"interval,omitempty"`                       // AhBmCs = Sleep A hours, B minutes, and C seconds between queries.
	SemanticVersioning *bool  `yaml:"semantic_versioning,omitempty" json:"semantic_versioning,omitempty"` // Default - true = Version has to follow semantic versioning (https://semver.org/), and be greater than the previous to trigger anything.
}

// Defaults are the default values for Options.
type Defaults struct {
	Base `yaml:",inline" json:",inline"`
}

// NewDefaults returns a new Defaults.
func NewDefaults(
	interval string,
	semanticVersioning *bool,
) *Defaults {
	return &Defaults{
		Base: Base{
			Interval:           interval,
			SemanticVersioning: semanticVersioning}}
}

// Default sets this Defaults to the default values.
func (od *Defaults) Default() {
	// interval
	od.Interval = "10m"

	// semantic_versioning
	semanticVersioning := true
	od.SemanticVersioning = &semanticVersioning
}

// Options are the options for a Service.
type Options struct {
	Base `yaml:",inline" json:",inline"`

	Active *bool `yaml:"active,omitempty" json:"active,omitempty"` // Disable the service.

	Defaults     *Defaults `yaml:"-" json:"-"` // Defaults.
	HardDefaults *Defaults `yaml:"-" json:"-"` // Hard Defaults.
}

// New Options.
func New(
	active *bool,
	interval string,
	semanticVersioning *bool,
	defaults, hardDefaults *Defaults,
) *Options {
	return &Options{
		Base: Base{
			Interval:           interval,
			SemanticVersioning: semanticVersioning},
		Active:       active,
		Defaults:     defaults,
		HardDefaults: hardDefaults}
}

// Copy the Options.
func (o *Options) Copy() *Options {
	if o == nil {
		return nil
	}

	return &Options{
		Base: Base{
			Interval:           o.Interval,
			SemanticVersioning: util.CopyPointer(o.SemanticVersioning)},
		Active:       util.CopyPointer(o.Active),
		Defaults:     o.Defaults,
		HardDefaults: o.HardDefaults}
}

// String returns a string representation of the Options.
func (o *Options) String() string {
	if o == nil {
		return ""
	}
	return util.ToYAMLString(o, "")
}

// GetActive status of the Service.
func (o *Options) GetActive() bool {
	return util.DereferenceOrNilValue(o.Active, true)
}

// GetInterval between queries for the latest/deployed version.
func (o *Options) GetInterval() string {
	return util.FirstNonDefault(
		o.Interval,
		o.Defaults.Interval,
		o.HardDefaults.Interval)
}

// GetSemanticVersioning returns whether the Service uses Semantic Versioning.
func (o *Options) GetSemanticVersioning() bool {
	return *util.FirstNonNilPtr(
		o.SemanticVersioning,
		o.Defaults.SemanticVersioning,
		o.HardDefaults.SemanticVersioning)
}

// GetIntervalPointer returns a pointer to the interval between queries on latest/deployed version.
func (o *Options) GetIntervalPointer() *string {
	if o.Interval != "" {
		return &o.Interval
	}
	if o.Defaults.Interval != "" {
		return &o.Defaults.Interval
	}
	return &o.HardDefaults.Interval
}

// GetIntervalDuration returns the interval between queries on latest/deployed version.
func (o *Options) GetIntervalDuration() time.Duration {
	d, _ := time.ParseDuration(o.GetInterval())
	return d
}

// CheckValues validates the fields of the Base struct.
func (b *Base) CheckValues(prefix string) error {
	// Interval
	if b.Interval != "" {
		// Treat integers as seconds by default.
		if _, err := strconv.Atoi(b.Interval); err == nil {
			b.Interval += "s"
		}
		if _, err := time.ParseDuration(b.Interval); err != nil {
			return fmt.Errorf("%sinterval: %q <invalid> (Use 'AhBmCs' duration format)",
				prefix, b.Interval)
		}
	}

	return nil
}
