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
				if index == len(lines)-1 || (len(lines[index+1]) != 0 && lines[index+1][0] != ' ') {
					break
				}

				// Get indentation on the next line
				// (Only do this once)
				if indentation == "" {
					indentation = getIndentation(lines[index+1])
					c.Settings.Indentation = uint8(len(indentation))
				}
			} else if afterService {
				break
			}
		}
		if afterService && strings.HasPrefix(line, indentation) && !strings.HasPrefix(line, indentation+" ") {
			// Check that it's a service and not a setting for a service.
			order = append(order, strings.TrimSpace(strings.TrimRight(line, ":")))
		}
	}

	c.All = order
	c.Order = &c.All

	// Filter out Services that aren't Active
	c.filterInactive()
}

func getIndentation(line string) (indentation string) {
	for _, v := range line {
		if v == ' ' {
			indentation += " "
		} else {
			return
		}
	}
	return
}

func (c *Config) filterInactive() {
	removed := 0
	for index, id := range c.All {
		if !utils.EvalNilPtr(c.Service[id].Active, true) ||
			!utils.EvalNilPtr(c.Service[id].Options.Active, true) {
			if removed == 0 {
				order := make([]string, len(c.All))
				copy(order, c.All)
				c.Order = &order
			}
			utils.RemoveIndex(c.Order, index-removed)
			removed++
		}
	}
}
