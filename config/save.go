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
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/release-argus/Argus/service"
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
		parentKey    = make([]string, 10)
		serviceCount = len(c.Service)

		currentServiceNumber   int
		currentOrder           = make([]string, serviceCount)
		currentOrderIndexStart = make([]int, serviceCount+1)
		currentOrderIndexEnd   = make([]int, serviceCount+1)
	)

	for index := 0; index < len(lines); index++ {
		line := lines[index]
		indents := indentCount(line, c.Settings.Indentation)
		key := strings.Split(strings.TrimSpace(line), ":")[0]
		parentKey[indents] = key

		// Start of a Service item
		if indents == 1 &&
			strings.HasSuffix(line, ":") &&
			parentKey[0] == "service" {

			// First service will be on 1 because we remove items and decrement
			// currentOrderIndexEnd[currentServiceNumber]. So we want to know when the service
			// has started so that the decrements are direct to the service
			currentServiceNumber++

			// Service's ID
			currentOrder[currentServiceNumber-1] = key
			currentOrderIndexStart[currentServiceNumber] = index

			// Get the index that this service ends on
			currentOrderIndexEnd[currentServiceNumber] = len(lines) - 1
			for i := index + 1; i < len(lines); i++ {
				lIndent := indentCount(lines[i], c.Settings.Indentation)
				// If the line has no/1 indent, it's the end of this Service
				if lIndent <= 1 {
					currentOrderIndexEnd[currentServiceNumber] = i - 1
					break
				}
			}
			continue
		}
		// Remove empty key:values
		if strings.HasSuffix(line, ": {}") {
			// Ignore empty notify/webhook mappings under a service as they are using defaults
			// service:
			// <>example:
			// <><>notify:
			// <><><>DISCORD: {}
			// <><><>EMAIL: {}
			// <><>webhook:
			// <><><>WH: {}
			if indents == 3 {
				// service:           || defaults:
				// <>name:            || <>service:
				// <><>VAR:           || <><>VAR:
				// <><><>DISCORD: {}  || <><><>DISCORD: {}
				if parentKey[0] == "service" || parentKey[0] == "defaults" {
					// Check that we're under a service.X.(notify|webhook): | defaults.service.X.(notify|webhook):
					if parentKey[2] == "notify" || parentKey[2] == "webhook" {
						continue
					}
				}
			}

			util.RemoveIndex(&lines, index)
			index--
			currentOrderIndexEnd[currentServiceNumber]--

			// Remove level by level
			// Until we don't find an empty map to remove
			removed := true
			for removed {
				if index < 0 {
					break
				}
				removed = false

				// This key is a map
				if len(lines) > 0 && strings.HasSuffix(lines[index], ":") {
					canRemove := false
					// If we're at the end of the file, it's an empty map
					if index+1 == len(lines) {
						// <>bish: <- EOF index
						// <><>bosh: {}  --- just removed
						// ^ EOF, can remove index
						canRemove = true
					} else {
						// Not at the end of the file, so compare the indentations
						indexIndentation := indentCount(lines[index], c.Settings.Indentation)
						nextIndentation := indentCount(lines[index+1], c.Settings.Indentation)
						// <><>bish: <- index
						// <><><>bosh: {}  --- just removed
						// <><>bosh:          | <><><>bash:
						// ^ can remove index | ^ can't remove index as bash is its child
						// If the current line is as indented or more than the next, it's an empty map
						canRemove = indexIndentation >= nextIndentation
					}
					if canRemove {
						util.RemoveIndex(&lines, index)
						currentOrderIndexEnd[currentServiceNumber]--
						removed = true
						index--
					}
				}
			}
		}
	}

	// Clean defaults
	removeAllServiceDefaults(
		&lines,
		c.Settings.Indentation,
		&c.Service,
		&currentOrder,
		&currentOrderIndexStart,
		&currentOrderIndexEnd)

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
		util.LogFrom{}, true)
}

// removeAllServiceDefaults removes the written default values from all Services
func removeAllServiceDefaults(
	lines *[]string,
	indentation uint8,
	services *service.Slice,
	currentOrder *[]string,
	currentOrderIndexStart *[]int,
	currentOrderIndexEnd *[]int) {
	linesRemoved := 0
	for i, serviceName := range *currentOrder {
		svc := (*services)[serviceName]
		n, c, w := svc.UsingDefaults()
		// Not using any defaults, skip this Service
		if !n && !c && !w {
			// Update the start/end indices of the next Service
			if i+2 < len(*currentOrderIndexStart) {
				(*currentOrderIndexStart)[i+2] -= linesRemoved
				(*currentOrderIndexEnd)[i+2] -= linesRemoved
			}
			continue
		}

		start := (*currentOrderIndexStart)[i+1]
		end := (*currentOrderIndexEnd)[i+1]
		section := (*lines)[start : end+1]
		size := len(section)
		// Notify
		if n {
			removeSection("notify", &section, indentation, 2)
		}
		// Command
		if c {
			removeSection("command", &section, indentation, 2)
		}
		// WebHook
		if w {
			removeSection("webhook", &section, indentation, 2)
		}

		// Put the section back into the lines
		sectionDecrease := size - len(section)
		newLines := make([]string, len(*lines)-sectionDecrease)
		copy(newLines[:start], (*lines)[:start])
		copy(newLines[start:], section)
		if end+1 < len(*lines) {
			copy(newLines[start+len(section):], (*lines)[end+1:])
		}
		*lines = newLines

		// Update the end index of the current Service
		(*currentOrderIndexEnd)[i+1] -= sectionDecrease
		// Update the start/end indices of the next Service
		linesRemoved += sectionDecrease
		if i+2 < len(*currentOrderIndexStart) {
			(*currentOrderIndexStart)[i+2] -= linesRemoved
			(*currentOrderIndexEnd)[i+2] -= linesRemoved
		}
	}
}

// removeSection removes the given section from the given lines
func removeSection(section string, lines *[]string, indentation uint8, indents int) {
	targetIndentation := (strings.Repeat(" ", int(indentation)*indents))
	outsideIndentation := targetIndentation + " "
	sectionStart := targetIndentation + section + ":"
	var insideSection bool
	for i := 0; i < len(*lines); i++ {
		line := (*lines)[i]
		if !insideSection {
			if strings.HasPrefix(line, sectionStart) {
				insideSection = true
				util.RemoveIndex(lines, i)
				i-- // subtract as we've removed a line
			}
			continue
		}

		// If we're inside the section, and this line isn't indented more than the target,
		// move to the next Service
		if !strings.HasPrefix(line, outsideIndentation) {
			break
		}
		// else, we're still inside, so remove this line
		util.RemoveIndex(lines, i)
		i-- // subtract as we've removed a line
	}
}

// indentCount returns the number of indents in the given line
func indentCount(line string, indentationSize uint8) int {
	// characters of indent / indentationSize = indents
	return len(util.Indentation(line, indentationSize)) / int(indentationSize)
}
