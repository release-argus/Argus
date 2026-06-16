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

// Package command provides CLI command execution for services.
package command

import (
	"time"

	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/web/metric"
)

// NewController returns a new [Controller].
func NewController(
	serviceStatus *status.Status,
	command Commands,
	shoutrrrNotifiers shoutrrr.Shoutrrrs,
	parentInterval *string,
) *Controller {
	if len(command) == 0 {
		return nil
	}
	var field Controller

	field.ServiceStatus = serviceStatus
	field.Command = command
	field.Failed = &serviceStatus.Fails.Command
	commandCount := len(field.Command)
	if field.Failed.Length() != commandCount {
		field.Failed.Init(commandCount)
	}
	field.nextRunnable = make([]time.Time, commandCount)

	field.ParentInterval = parentInterval

	// Command fail notifiers.
	field.Notifiers = Notifiers{
		Shoutrrr: shoutrrrNotifiers,
	}

	return &field
}

// InitMetrics registers Prometheus counters for all Command success/failure results.
func (c *Controller) InitMetrics() {
	if c == nil {
		return
	}
	serviceID := c.ServiceStatus.ServiceInfo.ID

	// ############
	// # Counters #
	// ############
	for _, cmd := range c.Command {
		name := cmd.String()
		metric.InitPrometheusCounter(
			metric.CommandResultTotal,
			name,
			serviceID,
			"",
			metric.ActionResultSuccess,
		)
		metric.InitPrometheusCounter(
			metric.CommandResultTotal,
			name,
			serviceID,
			"",
			metric.ActionResultFail,
		)
	}
}

// DeleteMetrics removes Prometheus counters for all Command success/failure results.
func (c *Controller) DeleteMetrics() {
	if c == nil {
		return
	}
	serviceID := c.ServiceStatus.ServiceInfo.ID

	// ############
	// # Counters #
	// ############
	for _, cmd := range c.Command {
		name := cmd.String()
		metric.DeletePrometheusCounter(
			metric.CommandResultTotal,
			name,
			serviceID,
			"",
			metric.ActionResultSuccess,
		)
		metric.DeletePrometheusCounter(
			metric.CommandResultTotal,
			name,
			serviceID,
			"",
			metric.ActionResultFail,
		)
	}
}

// IsRunnable returns whether the current time at `index` is before nextRunnable.
// If out of range, it returns false.
func (c *Controller) IsRunnable(index int) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// out of range.
	if index >= len(c.nextRunnable) {
		return false
	}

	return time.Now().UTC().After(c.nextRunnable[index])
}

// NextRunnable returns the next runnable time of the Command at `index`.
// If out of range, it returns a zero time.
func (c *Controller) NextRunnable(index int) time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// out of range.
	if index >= len(c.nextRunnable) {
		return time.Time{}
	}

	return c.nextRunnable[index]
}

// SetNextRunnable records when the command at index may run again.
func (c *Controller) SetNextRunnable(index int, time time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if index < len(c.nextRunnable) {
		c.nextRunnable[index] = time
	}
}

// SetExecuting blocks or extends the next runnable time while a command is running.
func (c *Controller) SetExecuting(index int, executing bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If out of range.
	if index >= len(c.nextRunnable) {
		return
	}

	// Different times depending on pass/fail.
	if !util.DerefOr(c.Failed.Get(index), true) {
		parentInterval, _ := time.ParseDuration(*c.ParentInterval)
		c.nextRunnable[index] = time.Now().UTC().Add(2 * parentInterval)
	} else {
		c.nextRunnable[index] = time.Now().UTC().Add(15 * time.Second)
	}

	// Block reruns whilst running for up to an hour.
	if executing {
		c.nextRunnable[index] = c.nextRunnable[index].Add(time.Hour)
	}
}
