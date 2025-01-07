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

// Package config provides the configuration for Argus.
package config

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/release-argus/Argus/util"
	"gopkg.in/yaml.v3"
)

// GetOrder of the Services from `c.File`.
func (c *Config) GetOrder(data []byte) {
	data = util.NormaliseNewlines(data)
	lines := strings.Split(string(data), "\n")
	order := make([]string, 0, len(c.Service))
	afterService := false
	indentation := ""

	// Iterate over the lines to find each service.
	for index := 0; index < len(lines); index++ {
		line := lines[index]
		// Skip empty lines and comments.
		if util.RegexCheck(`^\s*$`, line) {
			continue
		}

		// If we're on the start of a new block.
		if util.RegexCheck(`^[^\s].+:$`, line) {
			// Find `service:` start.
			if line == "service:" {
				afterService = true

				// If this is the last line, break.
				if index == len(lines)-1 {
					break
				}

				// Get indentation on the next line.
				// (Only do this once)
				if indentation == "" {
					index++
					// Search for the next service line if there is an empty line immediately after the service line.
					if lines[index] == "" {
						foundService := false
						serviceRegex := regexp.MustCompile(`^(\s*)[^:]+:$`)
						for i := index + 1; i < len(lines); i++ {
							if matches := serviceRegex.FindStringSubmatch(lines[i]); matches != nil {
								// Start of a new non-Service block!.
								if matches[1] == "" {
									return
								}
								index = i
								foundService = true
								break
							}
						}
						// No potential service found.
						if !foundService {
							return
						}
					}
					line = lines[index]

					indentation = Indentation(lines[index])
					c.Settings.Indentation = uint8(math.Min(float64(len(indentation)), 16))
				}
			} else if afterService {
				break
			}
		}

		if afterService &&
			util.RegexCheck(
				fmt.Sprintf(`^%s[^ ].*:$`, indentation),
				line) {
			// Check whether it is a service and not a setting for a service.
			yamlLine := strings.TrimSpace(strings.TrimRight(line, ":"))
			var serviceName string
			// Unmarshal YAML to handle any special characters.
			_ = yaml.Unmarshal([]byte(yamlLine), &serviceName) // Unmarshal err caught earlier.
			if serviceName != "" && c.Service[serviceName] != nil {
				order = append(order, serviceName)
			}
		}
	}

	c.OrderMutex.Lock()
	c.Order = order
	c.OrderMutex.Unlock()
}

// Indentation returns the indentation (leading spaces) of the line.
func Indentation(line string) (indentation string) {
	for _, v := range line {
		if v == ' ' {
			indentation += " "
		} else {
			return
		}
	}
	return
}
