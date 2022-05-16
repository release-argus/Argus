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

package config

import (
	"strings"

	"github.com/release-argus/Argus/utils"
)

// GetOrder of the Services from `c.File`.
func (c *Config) GetOrder(data []byte) {
	data = utils.NormaliseNewlines(data)
	lines := strings.Split(string(data), "\n")
	var order []string
	afterService := false
	indentation := ""
	for index, line := range lines {
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "#") && len(lines[index]) != 0 {
			// Find `service:` start
			if line == "service:" {
				afterService = true

				// If a service isn't on the next line
				if index == len(lines) || (len(lines[index+1]) != 0 && lines[index+1][0] != ' ') {
					break
				}

				// Get indentation on the next line
				// (Only do this once)
				if indentation == "" {
					for _, v := range lines[index+1] {
						if v == ' ' {
							indentation += " "
						} else {
							break
						}
					}
					c.Settings.Indentation = uint8(len(indentation))
				}
			} else {
				afterService = false
			}
		}
		if afterService && strings.HasPrefix(line, indentation) && !strings.HasPrefix(line, indentation+" ") {
			// Check that it's a service and not a setting for a service.
			order = append(order, strings.TrimSpace(strings.TrimRight(line, ":")))
		}
	}

	if len(c.Order) != 0 {
		// Add Services not in the existing Order.
		for _, serviceID := range order {
			if !utils.Contains(c.Order, serviceID) {
				c.Order = append(c.Order, serviceID)
			}
		}

		// Remove Services in the existing Order that have been removed.
		if len(order) != len(c.Order) {
			deleted := 0
			services := len(c.Order)
			for i := 0; i < services; i++ {
				if !utils.Contains(order, c.Order[i-deleted]) {
					if i == len(c.Order) {
						c.Order = c.Order[:deleted]
					} else {
						c.Order = append(c.Order[:i-deleted], c.Order[i-deleted+1:]...)
					}
					deleted++
				}
			}

		}
	} else {
		c.Order = order
	}
}
