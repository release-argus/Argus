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

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/release-argus/Argus/config/decode"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/polymorphic"
)

// ####################
// # REGISTRY | TYPES #
// ####################

// CommonRegistryDefaults holds shared default fields for registry checkers.
type CommonRegistryDefaults struct {
	Auth RegistryAuthDefaults `json:"auth,omitempty" yaml:"auth,omitempty"`
}

// CommonRegistry holds shared fields for a registry checkers.
type CommonRegistry struct {
	Type            string                          `json:"type,omitempty" yaml:"type,omitempty"` // Type of registry to check.
	ContainerDetail `json:",inline" yaml:",inline"` // Image/Tag to check.
	Auth            RegistryAuth                    `json:"auth,omitempty" yaml:"auth,omitempty"` // Auth details.

	// defaults form a fallback chain:
	//
	// instance -> provider defaults -> global defaults
	//
	// Values are resolved from most specific to least specific.
	defaults RegistryDefaults
}

// CommonRegistryDecode is an unmarshal-only helper for [CommonRegistry].
type CommonRegistryDecode struct {
	Type            string          `json:"type,omitempty" yaml:"type,omitempty"`
	ContainerDetail ContainerDetail `json:",inline" yaml:",inline"`
}

// #######################
// # REGISTRY | DECODING #
// #######################

// UnmarshalJSON implements the json.Unmarshaler interface.
func (r *CommonRegistryDefaults) UnmarshalJSON(data []byte) error {
	return r.unmarshal("json", data)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (r *CommonRegistryDefaults) UnmarshalYAML(data []byte) error {
	return r.unmarshal("yaml", data)
}

// unmarshal implements the format.Unmarshaler interface.
func (r *CommonRegistryDefaults) unmarshal(format string, data []byte) error {
	// Auth.
	if err := polymorphic.Unmarshal(format, data, "auth", r.Auth); err != nil {
		return err //nolint:wrapcheck
	}

	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (r *CommonRegistry) UnmarshalJSON(data []byte) error {
	return r.unmarshal("json", data)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (r *CommonRegistry) UnmarshalYAML(data []byte) error {
	return r.unmarshal("yaml", data)
}

// unmarshal implements the format.Unmarshaler interface.
func (r *CommonRegistry) unmarshal(format string, data []byte) error {
	aux := CommonRegistryDecode{
		Type: r.Type,
		ContainerDetail: ContainerDetail{
			Image: r.Image,
			Tag:   r.Tag,
		},
	}
	if err := decode.Unmarshal(format, data, &aux); err != nil {
		return err //nolint:wrapcheck
	}
	r.Type = aux.Type
	r.Image = aux.ContainerDetail.Image
	r.Tag = aux.ContainerDetail.Tag

	// Auth.
	if err := polymorphic.Unmarshal(format, data, "auth", r.Auth); err != nil {
		return err //nolint:wrapcheck
	}

	return nil
}

// ####################
// # REGISTRY | STATE #
// ####################

// IsZero implements the yaml.IsZeroer interface.
func (r *CommonRegistry) IsZero() bool {
	if r == nil {
		return true
	}

	return r.Type == "" &&
		r.Image == "" &&
		r.Tag == "" &&
		(r.Auth == nil || r.Auth.IsZero())
}

// Clone returns a deep copy of the receiver.
func (r *CommonRegistry) Clone() *CommonRegistry {
	if r == nil {
		return nil
	}

	field := CommonRegistry{
		Type:            r.Type,
		ContainerDetail: r.ContainerDetail.Copy(), //nolint:staticcheck
		defaults:        r.defaults,
	}
	if r.Auth != nil {
		field.Auth = r.Auth.Copy()
	}

	return &field
}

// ########################
// # REGISTRY | STRINGIFY #
// ########################

// String returns a string representation of the receiver.
func (r *CommonRegistry) String(prefix string) string {
	if r == nil {
		return ""
	}
	return decode.ToYAMLString(r, prefix)
}

// #######################
// # REGISTRY | DEFAULTS #
// #######################

// Defaults returns the defaults for the registry.
func (r *CommonRegistry) Defaults() RegistryDefaults {
	if r.defaults == nil {
		return nil
	}

	return r.defaults
}

// SetDefaults applies the dType registry defaults to the receiver.
func (r *CommonRegistry) SetDefaults(dType string, defaults *Defaults) {
	if defaults == nil {
		return
	}

	rDefaults := getRegistryDefaults(dType, defaults)
	if rDefaults == nil {
		return
	}

	r.defaults = rDefaults
	r.ContainerDetail.Defaults = &defaults.ContainerDetailDefaults
	r.Auth.SetDefaults(r.defaults.GetAuth())
}

// #######################
// # REGISTRY | METADATA #
// #######################

// GetTypeSelf returns the type of the registry.
func (r *CommonRegistry) GetTypeSelf() string {
	return r.Type
}

// GetImageSelf returns the Image to query on.
func (r *CommonRegistry) GetImageSelf() string {
	return r.Image
}

// GetImage returns the Image to query on.
func (r *CommonRegistry) GetImage() string {
	return r.ContainerDetail.GetImage()
}

// GetTagSelf returns the Tag to query for.
func (r *CommonRegistry) GetTagSelf() string {
	return r.Tag
}

// GetTag returns the Tag to query for.
func (r *CommonRegistry) GetTag() string {
	return r.ContainerDetail.GetTag()
}

// #########################
// # REGISTRY | VALIDATION #
// #########################

// CheckValues validates the fields of the receiver.
func (r *CommonRegistry) CheckValues() error {
	var errs []error

	// Image:
	switch image := r.GetImage(); {
	case image == "":
		errs = append(
			errs,
			&decode.FieldError{
				Key:         "image",
				Description: "image to check tags for",
			},
		)
		// 	Invalid.
	case !util.RegexCheck(`^[\w\-\.\/]+$`, image):
		errs = append(
			errs,
			&decode.FieldError{
				Key:         "image",
				Value:       image,
				Description: "ASCII required, input was non-ASCII",
			},
		)
	}

	// Tag
	switch tag := r.GetTag(); {
	case tag == "":
		errs = append(
			errs,
			&decode.FieldError{
				Key:         "tag",
				Description: "tag of image to check for existence",
			},
		)
	case !util.CheckTemplate(tag):
		errs = append(
			errs,
			&decode.FieldError{
				Key:         "tag",
				Value:       tag,
				Description: "didn't pass templating",
			},
		)
	default:
		if _, err := url.Parse("https://example.com/" + tag); err != nil {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "tag",
					Value:       tag,
					Description: "invalid for URL formatting",
				},
			)
		}
	}

	if r.Auth != nil {
		if err := r.Auth.CheckValues(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// #########################
// # REGISTRY | OPERATIONS #
// #########################

// GetTagForVersion returns the tag to search for, templated with version.
func (r *CommonRegistry) GetTagForVersion(version string) string {
	return util.TemplateString(r.GetTag(), serviceinfo.ServiceInfo{LatestVersion: version})
}

// parseBody parses the body of the response.
func (r *CommonRegistry) parseBody(tag string, resp *http.Response) error {
	if resp.StatusCode == http.StatusNotFound {
		return TagNotFoundError{Image: r.GetImage(), Tag: tag}
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf(
			"%s:%s - %s",
			r.GetImage(), tag, string(body),
		)
	}

	return nil
}

// Detail returns the resolved image and tag used for registry queries.
func (r *CommonRegistry) Detail() ContainerDetail {
	return ContainerDetail{
		Image: r.GetImage(),
		Tag:   r.GetTag(),
	}
}

// ##########################
// # REGISTRY | INHERITANCE #
// ##########################

// Inherit will copy the query token and expiry time from the target Registry to this one if their
// auth credentials match.
func (r *CommonRegistry) Inherit(from Registry) {
	if from == nil || r.Auth == nil {
		return
	}

	rAuth := r.GetAuth()
	fromAuth := from.GetAuth()
	if rAuth == nil || fromAuth == nil {
		return
	}

	rAuth.Inherit(fromAuth, r.Detail(), from.Detail())
}

// ###################
// # REGISTRY | AUTH #
// ###################

// GetAuth returns the auth defaults configured on these registry defaults.
func (r *CommonRegistryDefaults) GetAuth() RegistryAuthDefaults {
	return r.Auth
}

// GetAuth returns the auth configured on this registry.
func (r *CommonRegistry) GetAuth() RegistryAuth {
	return r.Auth
}
