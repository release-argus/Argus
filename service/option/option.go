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

// Package option provides options for a service.
package option

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Masterminds/semver/v3"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/util"
)

// DefaultsConfig pairs soft and hard service option defaults.
type DefaultsConfig struct {
	Soft *Defaults
	Hard *Defaults
}

// Base is the base struct for Options.
type Base struct {
	Interval           string `json:"interval,omitempty" yaml:"interval,omitempty"`                       // AhBmCs = Sleep A hours, B minutes, and C seconds between queries.
	SemanticVersioning *bool  `json:"semantic_versioning,omitempty" yaml:"semantic_versioning,omitempty"` // Default - true = Version has to follow semantic versioning (https://semver.org/), and be greater than the previous to trigger anything.
}

// IsZero implements the yaml.IsZeroer interface.
func (b Base) IsZero() bool {
	return b.Interval == "" &&
		b.SemanticVersioning == nil
}

// Defaults are the default values for Options.
type Defaults struct {
	Base `json:",inline" yaml:",inline"`
}

// IsZero implements the yaml.IsZeroer interface.
func (o Defaults) IsZero() bool {
	return o.Base.IsZero()
}

// Default sets the values of the receiver to their default values.
func (d *Defaults) Default() {
	// interval.
	d.Interval = "10m"

	// semantic_versioning.
	semanticVersioning := true
	d.SemanticVersioning = &semanticVersioning
}

// Options are the options for a Service, with defaults.
type Options struct {
	Base `json:",inline" yaml:",inline"`

	Active *bool `json:"active,omitempty" yaml:"active,omitempty"` // Disable the service.

	Defaults     *Defaults `json:"-" yaml:"-"` // Defaults.
	HardDefaults *Defaults `json:"-" yaml:"-"` // Hard Defaults.
}

// IsZero implements the yaml.IsZeroer interface.
func (o Options) IsZero() bool {
	return o.Active == nil && o.Base.IsZero()
}

// Copy returns a deep copy of the receiver.
func (o *Options) Copy() *Options {
	if o == nil {
		return nil
	}

	return &Options{
		Base: Base{
			Interval:           o.Interval,
			SemanticVersioning: util.ClonePtr(o.SemanticVersioning),
		},
		Active:       util.ClonePtr(o.Active),
		Defaults:     o.Defaults,
		HardDefaults: o.HardDefaults,
	}
}

// String implements fmt.Stringer and returns a YAML representation.
func (o *Options) String() string {
	if o == nil {
		return ""
	}
	return decode.ToYAMLString(o, "")
}

// GetActive returns whether the service is active
// (If Active is nil, it defaults to true).
func (o *Options) GetActive() bool {
	return util.DerefOr(o.Active, true)
}

// SetDefaults assigns defaults to the receiver.
func (o *Options) SetDefaults(defaults, hardDefaults *Defaults) {
	o.Defaults = defaults
	o.HardDefaults = hardDefaults
}

// GetInterval between queries for the latest/deployed version.
func (o *Options) GetInterval() string {
	return util.FirstNonDefault(
		o.Interval,
		o.Defaults.Interval,
		o.HardDefaults.Interval,
	)
}

// GetSemanticVersioning returns whether the Service uses Semantic Versioning.
func (o *Options) GetSemanticVersioning() bool {
	return *util.FirstNonNilPtr(
		o.SemanticVersioning,
		o.Defaults.SemanticVersioning,
		o.HardDefaults.SemanticVersioning,
	)
}

// VerifySemanticVersioning returns an error if the version is not following Semantic Versioning.
func (o *Options) VerifySemanticVersioning(version string, logFrom logx.LogFrom) (*semver.Version, error) {
	semanticVersion, err := semver.NewVersion(version)
	if err != nil {
		err = fmt.Errorf(
			"failed to convert %q to a semantic version. "+
				"If all versions follow this format, consider adding url_commands to transform the version into the 'MAJOR.MINOR.PATCH' format (https://semver.org/). "+
				"Alternatively, you can disable semantic versioning either globally with defaults.service.semantic_versioning or for this specific service using the options.semantic_versioning variable",
			version,
		)
		logx.Error(err, logFrom, true)
		return nil, err
	}

	return semanticVersion, nil
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

// CheckValues validates the fields of the receiver.
func (b *Base) CheckValues() error {
	// interval.
	if b.Interval != "" {
		// Treat integers as seconds by default.
		if _, err := strconv.Atoi(b.Interval); err == nil {
			b.Interval += "s"
		}
		if _, err := time.ParseDuration(b.Interval); err != nil {
			return &decode.FieldError{
				Key:         "interval",
				Value:       b.Interval,
				Description: "use 'AhBmCs' duration format",
			}
		}
	}

	return nil
}
