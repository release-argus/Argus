// Copyright [2023] [Argus]
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
			yamlLine := strings.TrimSpace(strings.TrimRight(line, ":"))
			var serviceName string
			// Unmarshal YAML to handle any special characters
			_ = yaml.Unmarshal([]byte(yamlLine), &serviceName) // Unmarhsal err caught earlier
			order = append(order, serviceName)
		}
	}

	c.OrderMutex.Lock()
	c.Order = order
	c.OrderMutex.Unlock()
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
