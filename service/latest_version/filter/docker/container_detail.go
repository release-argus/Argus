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

package docker

import "github.com/release-argus/Argus/util"

// #########
// # TYPES #
// #########

// ContainerDetail holds the image and tag for a Docker registry query.
type ContainerDetail struct {
	Image    string                   `json:"image,omitempty" yaml:"image,omitempty"` // Image of the container.
	Tag      string                   `json:"tag,omitempty" yaml:"tag,omitempty"`     // Tag of the Image.
	Defaults *ContainerDetailDefaults `json:"-" yaml:"-"`                             // Tag default chain.
}

// ContainerDetailDefaults holds default container values.
type ContainerDetailDefaults struct {
	Tag      string                   `json:"tag,omitempty" yaml:"tag,omitempty"` // Default Tag template.
	Defaults *ContainerDetailDefaults `json:"-" yaml:"-"`                         // Next link in the defaults chain.
}

// #########
// # STATE #
// #########

// IsZero implements the yaml.IsZeroer interface.
func (c *ContainerDetail) IsZero() bool {
	return c.Image == "" &&
		c.Tag == ""
}

// IsZero implements the yaml.IsZeroer interface.
func (c *ContainerDetailDefaults) IsZero() bool {
	return c == nil || c.Tag == ""
}

// Copy returns a deep copy of the receiver.
func (c *ContainerDetail) Copy() ContainerDetail {
	return ContainerDetail{
		Image:    c.Image,
		Tag:      c.Tag,
		Defaults: c.Defaults,
	}
}

// ############
// # DEFAULTS #
// ############

// Default sets the values of the receiver to their default values.
func (c *ContainerDetailDefaults) Default() {
	c.Tag = "{{ version }}"
}

// ##########
// # VALUES #
// ##########

// GetImage returns the container image, resolving environment variables.
func (c *ContainerDetail) GetImage() string {
	return util.EvalEnvVars(c.Image)
}

// GetTag returns the container tag, resolving defaults and environment variables.
func (c *ContainerDetail) GetTag() string {
	if c.Tag != "" {
		return util.EvalEnvVars(c.Tag)
	}

	return c.Defaults.GetTag()
}

// GetTag returns the default container tag, resolving defaults and environment variables.
func (c *ContainerDetailDefaults) GetTag() string {
	for detail := c; detail != nil; detail = detail.Defaults {
		if detail.Tag != "" {
			return util.EvalEnvVars(detail.Tag)
		}
	}

	return ""
}
