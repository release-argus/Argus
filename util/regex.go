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
	"regexp"
	"slices"
	"strings"

	serviceinfo "github.com/release-argus/Argus/service/status/info"
)

// RegexCheck reports whether re matches text.
func RegexCheck(re, text string) bool {
	return regexp.MustCompile(re).MatchString(text)
}

// RegexCheckWithVersion reports whether re matches text with version templated.
func RegexCheckWithVersion(re, text, version string) bool {
	re = TemplateString(re, serviceinfo.ServiceInfo{LatestVersion: version})
	return RegexCheck(re, text)
}

// RegexTemplate substitutes $N placeholders in template with regex capture groups.
func RegexTemplate(regexMatches []string, template string) string {
	result, _ := RegexTemplateWithLimit(regexMatches, template, 0)
	return result
}

// ErrTemplateTooLong is returned by [RegexTemplateWithLimit] when the result would exceed its limit.
var ErrTemplateTooLong = errors.New("templated result too long")

// RegexTemplateWithLimit is [RegexTemplate], returning [ErrTemplateTooLong] rather than building
// a result over limit bytes (0 for no limit).
func RegexTemplateWithLimit(regexMatches []string, template string, limit int) (string, error) {
	// No template, return the text at the last index.
	if template == "" {
		result := regexMatches[len(regexMatches)-1]
		if limit > 0 && len(result) > limit {
			return "", ErrTemplateTooLong
		}
		return result, nil
	}
	if limit > 0 && len(template) > limit {
		return "", ErrTemplateTooLong
	}

	// Replace placeholders with matched groups in reverse order
	// (to prevent replacing '$10' with '$1').
	result := template
	for i, regexMatch := range slices.Backward(regexMatches) {
		placeholder := fmt.Sprintf("$%d", i)
		// Check before replacing, as an inserted match may hold placeholders of its own.
		if limit > 0 {
			count := strings.Count(result, placeholder)
			if count > 0 && len(result)+count*(len(regexMatch)-len(placeholder)) > limit {
				return "", ErrTemplateTooLong
			}
		}
		result = strings.ReplaceAll(result, placeholder, regexMatch)
	}

	return result, nil
}
