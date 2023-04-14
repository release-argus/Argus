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
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/release-argus/Argus/util"
	"gopkg.in/yaml.v3"
)

// SaveHandler will listen to the `SaveChannel` and save the config (after a delay)
// when it receives a message.
func (c *Config) SaveHandler() {
	for {
		<-*c.SaveChannel
		waitChannelTimeout(c.SaveChannel)
		c.Save()
	}
}

// waitChannelTimeout will remove from `channel` and wait 30 seconds.
//
// Repeat until channel is empty at the end of the 30 seconds.
func waitChannelTimeout(channel *chan bool) {
	for {
		// Clear queue
		for len(*channel) != 0 {
			<-*channel
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
	// Lock the config
	c.OrderMutex.Lock()
	defer c.OrderMutex.Unlock()

	// Write the config to file (unordered slices, but with an order list)
	file, err := os.OpenFile(c.File, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	errMsg := fmt.Sprintf("error opening/creating file: %v", err)
	jLog.Fatal(errMsg, util.LogFrom{}, err != nil)
	defer file.Close()

	// Create the yaml encoder and set indentation
	yamlEncoder := yaml.NewEncoder(file)
	yamlEncoder.SetIndent(int(c.Settings.Indentation))

	// Write and close the file
	err = yamlEncoder.Encode(c)
	jLog.Fatal(
		fmt.Sprintf("error encoding %s:\n%v\n",
			c.File, err),
		util.LogFrom{},
		err != nil)
	err = file.Close()
	jLog.Fatal(
		fmt.Sprintf("error opening %s:\n%v\n",
			c.File, err),
		util.LogFrom{},
		err != nil)

	// Read the file to find what needs to be re-arranged
	data, err := os.ReadFile(c.File)
	msg := fmt.Sprintf("Error reading %q\n%s", c.File, err)
	jLog.Fatal(msg, util.LogFrom{}, err != nil)
	lines := strings.Split(string(util.NormaliseNewlines(data)), "\n")

	// Fix the ordering of the read data
	var (
		changed      = true
		indentation  = strings.Repeat(" ", int(c.Settings.Indentation))
		serviceCount = len(c.Service)

		configType string // section of the config we're in, e.g. 'service'

		currentServiceNumber   int
		currentOrder           = make([]string, serviceCount)
		currentOrderIndexStart = make([]int, serviceCount+1)
		currentOrderIndexEnd   = make([]int, serviceCount+1)
	)

	// Keep track of the number of lines we've removed and adjust index by it
	linesRemoved := 0
	for index := range lines {
		index -= linesRemoved
		if index < 0 {
			// Say in the first 5 lines we read, we removed 10, index-linesRemoved would be <0, so can't index
			// So we reset the linesRemoved number by subtracking -index and set index to 0
			linesRemoved -= index
			index = 0
		} else if index == len(lines) {
			break
		}
		if !strings.HasPrefix(lines[index], " ") {
			configType = strings.TrimRight(lines[index], ":")
		}

		if configType == "service" &&
			strings.HasSuffix(lines[index], ":") &&
			strings.HasPrefix(lines[index], indentation) &&
			!strings.HasPrefix(lines[index], indentation+" ") {
			// First service will be on 1 because we remove items and decrement
			// currentOrderIndexEnd[currentServiceNumber]. So we want to know when the service
			// has started so that the decrements are direct to the service
			currentServiceNumber++

			// Service's ID
			currentOrder[currentServiceNumber-1] = strings.TrimSpace(strings.TrimRight(lines[index], ":"))
			currentOrderIndexStart[currentServiceNumber] = index

			// Get the index that this service ends on
			currentOrderIndexEnd[currentServiceNumber] = len(lines) - 1
			for i := index + 1; i <= len(lines); i++ {
				// If the line has only 1 indentation or no indentation, it's the end of the Service
				if !strings.HasPrefix(lines[i], indentation+" ") &&
					strings.HasPrefix(lines[i], indentation) ||
					!strings.HasPrefix(lines[i], " ") {
					currentOrderIndexEnd[currentServiceNumber] = i - 1
					break
				}
			}
		}
		// Remove empty key:values
		if strings.HasSuffix(lines[index], ": {}") {
			// Ignore empty notify/webhook mappings under a service as they are using defaults
			// service:
			// <>example:
			// <><>notify:
			// <><><>DISCORD: {}
			// <><><>EMAIL: {}
			// <><>webhook:
			// <><><>WH: {}
			underNotify := false
			TwoIndents := strings.Repeat(" ", 2*int(c.Settings.Indentation))
			ThreeIndents := strings.Repeat(" ", 3*int(c.Settings.Indentation))
			if configType == "service" &&
				!strings.HasPrefix(lines[index], ThreeIndents+" ") &&
				strings.HasPrefix(lines[index], ThreeIndents) {
				// Check that we're under a notify/webhook:
				prevIndex := index
				for prevIndex > 0 {
					prevIndex--
					// line has only 2*indentation
					if !strings.HasPrefix(lines[prevIndex], TwoIndents+" ") &&
						strings.HasPrefix(lines[prevIndex], TwoIndents) {
						underNotify = lines[prevIndex] == TwoIndents+"notify:" ||
							lines[prevIndex] == TwoIndents+"webhook:"
						break
					}
				}
				if underNotify {
					continue
				}
			}

			util.RemoveIndex(&lines, index)
			currentOrderIndexEnd[currentServiceNumber]--
			linesRemoved++

			// Remove level by level
			// Until we don't find an empty map to remove
			// Possible current state:
			// bish
			//   bash:
			//     bosh: {} <- index
			// foo: bar
			removed := true
			parentsRemoved := 0 // shift index by this number (+1 when remove index-1)
			for removed {
				removed = false
				index -= parentsRemoved
				if index == len(lines) {
					continue
				}
				if index < 0 {
					parentsRemoved -= index
					index = 0
				}

				// If it's an empty map
				if strings.HasSuffix(lines[index], ": {}") {
					util.RemoveIndex(&lines, index)
					currentOrderIndexEnd[currentServiceNumber]--
					removed = true
					linesRemoved++
				} else {
					deepestRemovable := util.GetIndentation(lines[index], c.Settings.Indentation)
					if index != 0 &&
						strings.HasSuffix(lines[index-1], ":") &&
						strings.HasPrefix(lines[index-1], deepestRemovable) {

						util.RemoveIndex(&lines, index-1)
						currentOrderIndexEnd[currentServiceNumber]--
						removed = true
						linesRemoved++
						parentsRemoved++
					}
				}
			}
		}
	}

	// Service Bubble Sort
	for changed {
		changed = false
		// nth Pass
		for i := range currentOrderIndexStart {
			// Ignore the first index as it's before the service start
			// Ignore the second as we need something to compare it to
			if i < 2 {
				continue
			}

			// Check if `i-1` should be before `i-2`
			swap := false
			for j := range c.Order {
				// Found i-1 (current item)
				if c.Order[j] == currentOrder[i-1] {
					swap = true
					break
				}
				// Found i-2 (previous item)
				if c.Order[j] == currentOrder[i-2] {
					break
				}
			}

			if swap {
				// currentID needs to be moved before previousID
				util.Swap(
					&lines,
					currentOrderIndexStart[i-1], currentOrderIndexEnd[i-1],
					currentOrderIndexStart[i], currentOrderIndexEnd[i])
				changed = true

				// Update the current ordering values
				currentOrder[i-2], currentOrder[i-1] = currentOrder[i-1], currentOrder[i-2]
				lengthCurrent := currentOrderIndexEnd[i] - currentOrderIndexStart[i]
				currentOrderIndexEnd[i-1] = currentOrderIndexStart[i-1] + lengthCurrent
				currentOrderIndexStart[i] = currentOrderIndexEnd[i-1] + 1
			}
		}
	}

	// Open the file.
	file, err = os.OpenFile(c.File, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	errMsg = fmt.Sprintf("error opening/creating file: %v", err)
	jLog.Fatal(errMsg, util.LogFrom{}, err != nil)

	// Buffered writes to the file.
	writer := bufio.NewWriter(file)
	// -1 to stop ending with two empty lines.
	for i := range lines[:len(lines)-1] {
		fmt.Fprintln(writer, lines[i])
	}

	// Flush the writes.
	err = writer.Flush()
	errMsg = fmt.Sprintf("error writing file: %v", err)
	jLog.Fatal(errMsg, util.LogFrom{}, err != nil)
	jLog.Info(
		fmt.Sprintf("Saved service updates to %s", c.File),
		util.LogFrom{},
		true)
}
