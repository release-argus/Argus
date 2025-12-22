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
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

// drainChannel consumes all messages in the provided channel.
func drainChannel[T any](ch <-chan T) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}

var DebounceDuration = 30 * time.Second

// SaveHandler will listen to the `SaveChannel` and save the config (after a delay)
// when it receives a message.
func (c *Config) SaveHandler(ctx context.Context) {
	for {
		select {
		case <-c.SaveChannel:
			drainAndDebounce(ctx, c.SaveChannel, DebounceDuration)
			c.Save()

		case <-ctx.Done():
			// Shutdown requested: flush everything and exit
			drainChannel(c.SaveChannel)
			return
		}
	}
}

// drainAndDebounce will clear the message queue from `channel` and wait `DebounceDuration`.
//
// Repeat until the channel is empty at the end of the debounce.
func drainAndDebounce[T any](ctx context.Context, channel chan T, duration time.Duration) {
	for {
		// Drain queue.
		drainChannel(channel)

		select {
		case <-time.After(duration):
			// debounce elapsed, check again.
		case <-ctx.Done():
			// shutdown: stop waiting immediately.
			return
		}

		// End if channel still empty.
		if len(channel) == 0 {
			return
		}
	}
}

// Save the configuration to `c.File`.
func (c *Config) Save() (ok bool) {
	// Lock the config.
	c.OrderMutex.Lock()
	defer c.OrderMutex.Unlock()

	// Encode to memory (Go-ordered slices, but with an order list for Services).
	var buf bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&buf)
	yamlEncoder.SetIndent(int(c.Settings.Indentation))
	if err := yamlEncoder.Encode(c); err != nil {
		logutil.Log.Fatal(
			fmt.Sprintf("error encoding config: %v", err),
			logutil.LogFrom{})
		return
	}

	// Reorder and clean the YAML in memory.
	lines := strings.Split(string(util.NormaliseNewlines(buf.Bytes())), "\n")
	lines = c.reorderYAML(lines)

	// Open the file.
	file, err := os.OpenFile(c.File, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		logutil.Log.Fatal(
			fmt.Sprintf("error opening %s: %v", c.File, err),
			logutil.LogFrom{})
		return false
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	// Write all lines except the potential trailing empty one from Split.
	for i, line := range lines {
		if i == len(lines)-1 && line == "" {
			continue
		}
		if _, err := fmt.Fprintln(writer, line); err != nil {
			logutil.Log.Fatal(
				fmt.Sprintf("error writing to %s: %v", c.File, err),
				logutil.LogFrom{})
			return
		}
	}

	if err := writer.Flush(); err != nil {
		logutil.Log.Fatal(
			fmt.Sprintf("error flushing %s: %v", c.File, err),
			logutil.LogFrom{})
		return
	}

	logutil.Log.Info(
		"Saved service updates to "+c.File,
		logutil.LogFrom{}, true)
	return true
}

// reorderYAML rearranges the service blocks and cleans empty values.
func (c *Config) reorderYAML(lines []string) []string {
	var (
		parentKey              = make([]string, 10)
		serviceCount           = len(c.Service)
		currentServiceNumber   = -1
		currentOrder           = make([]string, serviceCount)
		currentOrderIndexStart = make([]int, serviceCount)
		currentOrderIndexEnd   = make([]int, serviceCount)
	)

	for index := 0; index < len(lines); index++ {
		line := lines[index]
		indents := indentCount(line, c.Settings.Indentation)
		key := strings.Split(strings.TrimSpace(line), ":")[0]
		parentKey[indents] = key

		// First item in the Service block.
		if indents == 1 && strings.HasSuffix(line, ":") &&
			parentKey[0] == "service" {
			currentServiceNumber++

			// Service data.
			currentOrder[currentServiceNumber] = key
			currentOrderIndexStart[currentServiceNumber] = index

			// Get the index that this service ends on.
			currentOrderIndexEnd[currentServiceNumber] = len(lines) - 1
			for i := index + 1; i < len(lines); i++ {
				// If the line has no/1 indent, it is the end of this service's block.
				if indentCount(lines[i], c.Settings.Indentation) <= 1 {
					currentOrderIndexEnd[currentServiceNumber] = i - 1
					break
				}
			}
			continue
		}

		// Remove empty maps/arrays: "key: {}" | "key: []"
		if strings.HasSuffix(line, ": {}") || strings.HasSuffix(line, ": []") {
			// Ignore empty notify/webhook mappings under a service as they are using defaults.
			if indents == 3 &&
				// service:           || defaults:
				// <>name:            || <>service:
				// <><>VAR:           || <><>VAR:
				// <><><>DISCORD: {}  || <><><>DISCORD: {}
				(parentKey[0] == "service" || parentKey[0] == "defaults") &&
				(parentKey[2] == "notify" || parentKey[2] == "webhook") {
				continue
			}

			util.RemoveIndex(&lines, index)
			index--
			if currentServiceNumber >= 0 {
				currentOrderIndexEnd[currentServiceNumber]--
			}
		}

		// Remove
		for index >= 0 && strings.HasSuffix(lines[index], ":") {
			// If we are at the end of the file, it is an empty map.
			canRemove := index+1 == len(lines) ||
				// Not at the end of the file, so compare the indentations.
				// <><>bish: <- index
				// <><><>bosh: {}  --- just removed
				// <><>bosh:          | <><><>bash:
				// ^ can remove index | ^ can't remove index as bash is its child.
				// If the current line is indented as much, or more than the next, it is an empty map.
				indentCount(lines[index], c.Settings.Indentation) >= indentCount(lines[index+1], c.Settings.Indentation)
			if !canRemove {
				break
			}
			util.RemoveIndex(&lines, index)
			index--
			if currentServiceNumber >= 0 {
				currentOrderIndexEnd[currentServiceNumber]--
			}
		}
	}

	// Clean defaults.
	removeAllServiceDefaults(
		&lines,
		c.Settings.Indentation,
		&c.Service,
		&currentOrder,
		&currentOrderIndexStart, &currentOrderIndexEnd)

	// Bubble sort services based on c.Order.
	changed := true
	for changed {
		changed = false
		// Check if `i` should be before `i-1`.
		for i := 1; i < len(currentOrderIndexStart); i++ {
			currentIndex := i
			currentID := currentOrder[currentIndex]
			previousIndex := i - 1
			previousID := currentOrder[previousIndex]

			swap := false
			// swap = true if currentID before previousID in c.Order.
			for _, id := range c.Order {
				// Found i (current item).
				if id == currentID {
					swap = true
					break
				}
				// Found i-1 (previous item).
				if id == previousID {
					break
				}
			}
			if swap {
				// currentID needs to move before previousID.
				util.Swap(&lines,
					currentOrderIndexStart[previousIndex], currentOrderIndexEnd[previousIndex],
					currentOrderIndexStart[currentIndex], currentOrderIndexEnd[currentIndex])
				changed = true

				// Swap the current ordering values.
				currentOrder[previousIndex], currentOrder[currentIndex] = currentOrder[currentIndex], currentOrder[previousIndex]
				lengthCurrent := currentOrderIndexEnd[currentIndex] - currentOrderIndexStart[currentIndex]
				currentOrderIndexEnd[previousIndex] = currentOrderIndexStart[previousIndex] + lengthCurrent
				currentOrderIndexStart[currentIndex] = currentOrderIndexEnd[previousIndex] + 1
			}
		}
	}
	return lines
}

// removeAllServiceDefaults removes the written default values from all Services.
func removeAllServiceDefaults(
	lines *[]string,
	indentation uint8,
	services *service.Services,
	currentOrder *[]string,
	currentOrderIndexStart *[]int,
	currentOrderIndexEnd *[]int) {
	linesRemoved := 0
	for i, serviceName := range *currentOrder {
		svc := (*services)[serviceName]
		n, c, w := svc.UsingDefaults()
		// Not using any defaults, skip this Service.
		if !n && !c && !w {
			// Update the start/end indices of the next Service.
			if i+1 < len(*currentOrderIndexStart) {
				(*currentOrderIndexStart)[i+1] -= linesRemoved
				(*currentOrderIndexEnd)[i+1] -= linesRemoved
			}
			continue
		}

		start := (*currentOrderIndexStart)[i]
		end := (*currentOrderIndexEnd)[i]
		section := (*lines)[start : end+1]
		size := len(section)
		// Notify.
		if n {
			removeSection("notify", &section, indentation, 2)
		}
		// Command.
		if c {
			removeSection("command", &section, indentation, 2)
		}
		// WebHook.
		if w {
			removeSection("webhook", &section, indentation, 2)
		}

		// Put the section back into the lines.
		sectionDecrease := size - len(section)
		newLines := make([]string, len(*lines)-sectionDecrease)
		copy(newLines[:start], (*lines)[:start])
		copy(newLines[start:], section)
		if end+1 < len(*lines) {
			copy(newLines[start+len(section):], (*lines)[end+1:])
		}
		*lines = newLines

		// Update the end index of the current Service.
		(*currentOrderIndexEnd)[i] -= sectionDecrease
		// Update the start/end indices of the next Service.
		linesRemoved += sectionDecrease
		if i+1 < len(*currentOrderIndexStart) {
			(*currentOrderIndexStart)[i+1] -= linesRemoved
			(*currentOrderIndexEnd)[i+1] -= linesRemoved
		}
	}
}

// removeSection removes the given section from the given lines.
func removeSection(section string, lines *[]string, indentation uint8, indents int) {
	targetIndentation := strings.Repeat(" ", int(indentation)*indents)
	outsideIndentation := targetIndentation + " "
	sectionStart := targetIndentation + section + ":"
	var insideSection bool
	for i := 0; i < len(*lines); i++ {
		line := (*lines)[i]
		if !insideSection {
			if strings.HasPrefix(line, sectionStart) {
				insideSection = true
				util.RemoveIndex(lines, i)
				i-- // subtract as we've removed a line.
			}
			continue
		}

		// If we are inside the section, and this line is not indented more than the target,
		// move to the next Service.
		if !strings.HasPrefix(line, outsideIndentation) {
			break
		}
		// else, we are still inside, so remove this line.
		util.RemoveIndex(lines, i)
		i-- // subtract as we've removed a line.
	}
}

// indentCount returns the amount of indents in the given line.
func indentCount(line string, indentationSize uint8) int {
	// characters of indent / indentationSize = indents.
	return len(util.Indentation(line, indentationSize)) / int(indentationSize)
}
