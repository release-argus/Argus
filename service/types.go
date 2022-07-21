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

	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service/deployed_version"
	"github.com/release-argus/Argus/service/latest_version"
	lv_github "github.com/release-argus/Argus/service/latest_version/github"
	lv_url "github.com/release-argus/Argus/service/latest_version/url"
	"github.com/release-argus/Argus/service/options"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/webhook"
)

var (
	jLog *utils.JLog
)

// Slice is a slice mapping of Service.
type Slice map[string]*Service

// Service is a source to be serviceed and provides everything needed to extract
// the latest version from the URL provided.
type Service struct {
	ID                    string                   `yaml:"-"`                          // service_name
	Type                  string                   `yaml:"-"`                          // service_name
	Comment               *string                  `yaml:"comment,omitempty"`          // Comment on the Service
	Options               options.Options          `yaml:"options,omitempty"`          // Options to give the Service
	LatestVersion         latest_version.Lookup    `yaml:"latest_version,omitempty"`   // Vars to getting the latest version of the Service
	CommandController     *command.Controller      `yaml:"-"`                          // The controller for the OS Commands that tracks fails and has the announce channel
	Command               *command.Slice           `yaml:"command,omitempty"`          // OS Commands to run on new release
	WebHook               *webhook.Slice           `yaml:"webhook,omitempty"`          // Service-specific WebHook vars
	Notify                *shoutrrr.Slice          `yaml:"notify,omitempty"`           // Service-specific Shoutrrr vars
	DeployedVersionLookup *deployed_version.Lookup `yaml:"deployed_version,omitempty"` // Var to scrape the Service's current deployed version
	Dashboard             DashboardOptions         `yaml:"dashboard,omitempty"`        // Options for the dashboard
	Status                *service_status.Status   `yaml:"-"`                          // Track the Status of this source (version and regex misses)
	HardDefaults          *Service                 `yaml:"-"`                          // Hardcoded default values
	Defaults              *Service                 `yaml:"-"`                          // Default values

	// TODO: Deprecate
	OldStatus *service_status.OldStatus `yaml:"status,omitempty"` // For moving version info to argus.db
}

func (s *Service) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	var svc Service
	svc.LatestVersion = lv_github.LatestVersion{}
	err = unmarshal(&svc)
	if err == nil {
		switch *svc.LatestVersion.(lv_github.LatestVersion).Type {
		case "github":
			*s = svc
		case "url":
			svc.LatestVersion = lv_url.LatestVersion{}
			err = unmarshal(&svc)
			if err == nil {
				*s = svc
			}
		default:
			err = fmt.Errorf("latest_version type %q not supported",
				*svc.LatestVersion.(lv_github.LatestVersion).Type)
		}
	}
	return
}
