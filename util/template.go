// Copyright [2025] [Argus]
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

var pongoMutex = sync.Mutex{}

// TemplateString with pongo2 and `context`.
func TemplateString(template string, context serviceinfo.ServiceInfo) string {
	// If the string does not represent a Jinja template.
	if !strings.Contains(template, "{") {
		return template
	}
	// pongo2 DATA RACE.
	pongoMutex.Lock()
	defer pongoMutex.Unlock()

	// Compile the template.
	tpl, err := pongo2.FromString(template)
	if err != nil {
		panic(err)
	}

	// Render the template.
	result, err := tpl.Execute(pongo2.Context{
		"service_id":       context.ID,
		"service_name":     context.Name,
		"service_url":      context.URL,
		"icon":             context.Icon,
		"icon_link_to":     context.IconLinkTo,
		"web_url":          context.WebURL,
		"approved_version": context.ApprovedVersion,
		"deployed_version": context.DeployedVersion,
		"version":          context.LatestVersion,
		"latest_version":   context.LatestVersion,
		"tags":             context.Tags,
	})
	if err != nil {
		panic(err)
	}
	return result
}

// CheckTemplate verifies the validity of the template.
func CheckTemplate(template string) bool {
	// pongo2 DATA RACE.
	pongoMutex.Lock()
	defer pongoMutex.Unlock()

	_, err := pongo2.FromString(template)
	return err == nil
}
