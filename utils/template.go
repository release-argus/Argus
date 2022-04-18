// Copyright [2022] [Hymenaios]
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

package utils

import (
	"github.com/flosch/pongo2/v5"
)

// TemplateString with pongo2 and `context`.
func TemplateString(tmpl string, context ServiceInfo) string {
	// Compile the template.
	tpl, err := pongo2.FromString(tmpl)
	if err != nil {
		panic(err)
	}

	// Render the template.
	out, err := tpl.Execute(pongo2.Context{
		"service_id":  context.ID,
		"service_url": context.URL,
		"web_url":     context.WebURL,
		"version":     context.LatestVersion,
	})
	if err != nil {
		panic(err)
	}
	return out
}
