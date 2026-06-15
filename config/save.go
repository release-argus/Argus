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

// Package config provides the configuration for Argus.
package config

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/util"
)

// encodeConfigYAML marshals the config to YAML (overridable for tests).
var encodeConfigYAML = func(w io.Writer, indent int, c *Config) error {
	yamlEncoder := decode.NewYAMLEncoder(w, indent)
	defer yamlEncoder.Close()
	return yamlEncoder.Encode(c)
}

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

// DebounceDuration is the delay before persisting config changes after the last edit signal.
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
	c.OrderMu.Lock()
	defer c.OrderMu.Unlock()

	// Encode to memory (Go-ordered slices, but with an order list for Services).
	var buf bytes.Buffer
	if err := encodeConfigYAML(&buf, int(c.Settings.Indentation), c); err != nil {
		logx.Fatal(
			fmt.Sprintf("error encoding config: %v", err),
			logx.LogFrom{},
		)
		return
	}

	// Reorder and clean the YAML in memory.
	lines := strings.Split(string(util.NormaliseNewlines(buf.Bytes())), "\n")
	lines = c.reorderYAML(lines)

	// Open the file.
	file, err := os.OpenFile(c.File, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		logx.Fatal(
			fmt.Sprintf("error opening %s: %v", c.File, err),
			logx.LogFrom{},
		)
		return false
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	// Trim trailing newline.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	// Write all lines.
	for _, line := range lines {
		if _, err := fmt.Fprintln(writer, line); err != nil {
			logx.Fatal(
				fmt.Sprintf("error writing to %s: %v", c.File, err),
				logx.LogFrom{},
			)
			return
		}
	}

	if err := writer.Flush(); err != nil {
		logx.Fatal(
			fmt.Sprintf("error flushing %s: %v", c.File, err),
			logx.LogFrom{},
		)
		return
	}

	logx.Info("Saved service updates to "+c.File, logx.LogFrom{}, true)
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

		// Grow parentKey slice if needed to accommodate deep nesting
		if indents >= len(parentKey) {
			newParentKey := make([]string, indents+10)
			copy(newParentKey, parentKey)
			parentKey = newParentKey
		}
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

			lines = util.RemoveAt(lines, index)
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
			lines = util.RemoveAt(lines, index)
			index--
			if currentServiceNumber >= 0 {
				currentOrderIndexEnd[currentServiceNumber]--
			}
		}
	}

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
				lines = util.SwapRanges(
					lines,
					currentOrderIndexStart[previousIndex], currentOrderIndexEnd[previousIndex],
					currentOrderIndexStart[currentIndex], currentOrderIndexEnd[currentIndex],
				)
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

// indentCount returns the number of indents in the given line.
func indentCount(line string, indentationSize uint8) int {
	// characters of indent / indentationSize = indents.
	return len(util.Indentation(line, indentationSize)) / int(indentationSize)
}
