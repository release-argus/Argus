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
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/flosch/pongo2/v6"

	serviceinfo "github.com/release-argus/Argus/service/status/info"
)

var pongoMu = sync.Mutex{}

// errTemplateLoad rejects a template's attempt to read from disk.
var errTemplateLoad = errors.New("loading templates from disk is not supported")

// denyLoader refuses every template load, so tags that pull in another
// template ({% include %}, {% extends %}, {% ssi %}) cannot read local files.
type denyLoader struct{}

func (denyLoader) Abs(_, name string) string { return name }

func (denyLoader) Get(path string) (io.Reader, error) {
	return nil, fmt.Errorf("%w: %q", errTemplateLoad, path)
}

// pongoSet compiles every template, with disk access denied.
var pongoSet = pongo2.NewSet("argus", denyLoader{})

// TemplateString applies Django templating to template using service info,
// returning template unchanged if it cannot be compiled or rendered.
func TemplateString(template string, info serviceinfo.ServiceInfo) (rendered string) {
	// If the string does not represent a Jinja template.
	if !strings.Contains(template, "{") {
		return template
	}
	// pongo2 DATA RACE.
	pongoMu.Lock()
	defer pongoMu.Unlock()

	// A malformed or unrenderable template yields the string unchanged rather
	// than taking the process down.
	defer func() {
		if r := recover(); r != nil {
			rendered = template
		}
	}()

	// Compile the template.
	tpl, err := pongoSet.FromString(template)
	if err != nil {
		return template
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
		return template
	}
	return result
}

// CheckTemplate verifies the validity of the template.
func CheckTemplate(template string) bool {
	// pongo2 DATA RACE.
	pongoMu.Lock()
	defer pongoMu.Unlock()

	_, err := pongoSet.FromString(template)
	return err == nil
}
