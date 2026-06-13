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

// Package util provides utility functions for the Argus project.
package util

import (
	"strings"
	"sync"

	"github.com/flosch/pongo2/v6"

	serviceinfo "github.com/release-argus/Argus/service/status/info"
)

var pongoMu = sync.Mutex{}

// TemplateString with pongo2 and service info.
func TemplateString(template string, info serviceinfo.ServiceInfo) string {
	// If the string does not represent a Jinja template.
	if !strings.Contains(template, "{") {
		return template
	}
	// pongo2 DATA RACE.
	pongoMu.Lock()
	defer pongoMu.Unlock()

	// Compile the template.
	tpl, err := pongo2.FromString(template)
	if err != nil {
		panic(err)
	}

	// Render the template.
	result, err := tpl.Execute(
		pongo2.Context{
			"service_id":       info.ID,
			"service_name":     info.Name,
			"service_url":      info.URL,
			"icon":             info.Icon,
			"icon_link_to":     info.IconLinkTo,
			"web_url":          info.WebURL,
			"approved_version": info.ApprovedVersion,
			"deployed_version": info.DeployedVersion,
			"version":          info.LatestVersion,
			"latest_version":   info.LatestVersion,
			"tags":             info.Tags,
		},
	)
	if err != nil {
		panic(err)
	}
	return result
}

// CheckTemplate verifies the validity of the template.
func CheckTemplate(template string) bool {
	// pongo2 DATA RACE.
	pongoMu.Lock()
	defer pongoMu.Unlock()

	_, err := pongo2.FromString(template)
	return err == nil
}
