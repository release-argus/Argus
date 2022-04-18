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

package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hymenaios-io/Hymenaios/utils"
	"gopkg.in/yaml.v3"
)

// SaveHandler will listen to the `SaveChannel` and save the config (after a delay)
// when it receives a message.
func (c *Config) SaveHandler(filePath *string) {
	for {
		select {
		case _ = <-*c.SaveChannel:
			waitChannelTimeout(c.SaveChannel)
			c.Save()
		}
	}
}

// waitChannelTimeout will remove from `channel` and wait 30 seconds.
//
// Repeat until channel is empty at the end of the 30 seconds.
func waitChannelTimeout(channel *chan bool) {
	for {
		// Clear queue
		for len(*channel) != 0 {
			_ = <-*channel
		}

		// Sleep 30s
		time.Sleep(30 * time.Second)

		// End if channel is still empty
		if len(*channel) == 0 {
			break
		}
	}
}

// Save `c.File`.
func (c *Config) Save() {
	// Write the config to file (unordered slices, but with an order list)
	file, err := os.OpenFile(c.File, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	errMsg := fmt.Sprintf("error opening/creating file: %v", err)
	jLog.Fatal(errMsg, utils.LogFrom{}, err != nil)

	yamlEncoder := yaml.NewEncoder(file)
	yamlEncoder.SetIndent(int(c.Settings.Indentation))

	err = yamlEncoder.Encode(c)
	jLog.Fatal(
		fmt.Sprintf(
			"error encoding %s:\n%v\n",
			c.File,
			err,
		),
		utils.LogFrom{},
		err != nil,
	)
	err = file.Close()
	jLog.Fatal(
		fmt.Sprintf(
			"error opening %s:\n%v\n",
			c.File,
			err,
		),
		utils.LogFrom{},
		err != nil,
	)

	// Read the file.
	data, err := os.ReadFile(c.File)
	msg := fmt.Sprintf("Error reading %q\n%s", c.File, err)
	jLog.Fatal(msg, utils.LogFrom{}, err != nil)
	lines := strings.Split(string(utils.NormaliseNewlines(data)), "\n")

	// Fix the ordering of the read data.
	var (
		changed      bool   = true
		indentation  string = strings.Repeat(" ", int(c.Settings.Indentation))
		serviceCount int    = len(c.Service)

		configType string

		orderIndexStart int
		orderIndexEnd   int
		foundOrder      bool

		currentServiceNumber   int
		currentOrder           []string = make([]string, serviceCount)
		currentOrderIndexStart []int    = make([]int, serviceCount)
		currentOrderIndexEnd   []int    = make([]int, serviceCount)
	)
	for index, line := range lines {
		if !strings.HasPrefix(line, " ") {
			configType = strings.TrimRight(line, ":")

			// Remove ordering var.
			if configType == "order" {
				foundOrder = true
				orderIndexStart = index
			} else if foundOrder {
				orderIndexEnd = index
				lines = append(lines[:orderIndexStart], lines[orderIndexEnd:]...)
				// Only remove once.
				foundOrder = false
			}
			continue
		}

		switch configType {
		case "service":
			if strings.HasPrefix(line, indentation) && !strings.HasPrefix(line, indentation+" ") {
				// Services ID
				currentOrder[currentServiceNumber] = strings.TrimSpace(strings.TrimRight(line, ":"))
				currentOrderIndexStart[currentServiceNumber] = index

				currentOrderIndexEnd[currentServiceNumber] = len(lines)
				for i := range lines {
					if i <= index {
						continue
					}
					if (strings.HasPrefix(lines[i], indentation) && !strings.HasPrefix(lines[i], indentation+" ")) ||
						!strings.HasPrefix(line, " ") {
						currentOrderIndexEnd[currentServiceNumber] = i
						break
					}
				}
				currentServiceNumber++
			}
		}
	}

	// Service Bubble Sort
	for changed == true {
		changed = false
		// nth Pass
		for i := range c.Order {
			if i == 0 {
				continue
			}

			// Check if `i` should be before `i-1`
			swap := false
			for j := range c.Order {
				if c.Order[j] == currentOrder[i] {
					swap = true
					break
				}
				if c.Order[j] == currentOrder[i-1] {
					break
				}
			}

			if swap {
				// currentID needs to be moved before previousID
				// previousIDIndex:currentIDIndexStart:currentIDIndexEnd:
				// start[i-1]:start[i]:end[i]:
				// 0(before) 1(previousID) 2(currentID) 3(after)
				// becomes
				// 0(before) 2(currentID) 1(previousID) 3(after)
				tmp := make([]string, len(lines))
				copy(tmp, lines)
				// tmp = 1,3
				tmp = append(tmp[currentOrderIndexStart[i-1]:currentOrderIndexStart[i]], tmp[currentOrderIndexEnd[i]:]...)

				// lines = 0,2
				lines = append(lines[:currentOrderIndexStart[i-1]], lines[currentOrderIndexStart[i]:currentOrderIndexEnd[i]]...)

				// lines = 0,2,1,3
				lines = append(lines, tmp...)
				changed = true

				// Update the current ordering values
				tmpStr := currentOrder[i-1]
				currentOrder[i-1] = currentOrder[i]
				currentOrder[i] = tmpStr
				lengthCurrent := currentOrderIndexEnd[i] - currentOrderIndexStart[i]
				currentOrderIndexEnd[i-1] = currentOrderIndexStart[i-1] + lengthCurrent
				currentOrderIndexStart[i] = currentOrderIndexEnd[i-1]
			}
		}
	}

	// Open the file.
	file, err = os.OpenFile(c.File, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	errMsg = fmt.Sprintf("error opening/creating file: %v", err)
	jLog.Fatal(errMsg, utils.LogFrom{}, err != nil)

	// Buffered writes to the file.
	writer := bufio.NewWriter(file)
	// -1 to stop ending with two empty lines.
	for i := range lines[:len(lines)-1] {
		fmt.Fprintln(writer, lines[i])
	}

	// Flush the writes.
	err = writer.Flush()
	errMsg = fmt.Sprintf("error writing file: %v", err)
	jLog.Fatal(errMsg, utils.LogFrom{}, err != nil)
	jLog.Info(
		fmt.Sprintf("Saved service updates to %s", c.File),
		utils.LogFrom{},
		true)
}
