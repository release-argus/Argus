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

// Package command provides the cli command functionality for Argus.
package command

import (
	"fmt"
	"strings"
	"time"

	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/web/metric"
)

// LogInit for this package.
func LogInit(log *util.JLog) {
	jLog = log
}

// Init the Command Controller.
func (c *Controller) Init(
	serviceStatus *status.Status,
	command *Slice,
	shoutrrrNotifiers *shoutrrr.Slice,
	parentInterval *string,
) {
	if c == nil || len(*command) == 0 {
		return
	}

	c.ServiceStatus = serviceStatus
	c.Command = command
	c.Failed = &serviceStatus.Fails.Command
	commandCount := len(*c.Command)
	if c.Failed.Length() != commandCount {
		c.Failed.Init(commandCount)
	}
	c.nextRunnable = make([]time.Time, commandCount)

	c.ParentInterval = parentInterval

	// Command fail notifiers.
	c.Notifiers = Notifiers{
		Shoutrrr: shoutrrrNotifiers,
	}
}

// InitMetrics will initialise the Prometheus metrics for each Command in the Controller.
func (c *Controller) InitMetrics() {
	if c == nil {
		return
	}

	// ############
	// # Counters #
	// ############
	for _, cmd := range *c.Command {
		name := cmd.String()
		metric.InitPrometheusCounter(metric.CommandResultTotal,
			name,
			*c.ServiceStatus.ServiceID,
			"",
			"SUCCESS")
		metric.InitPrometheusCounter(metric.CommandResultTotal,
			name,
			*c.ServiceStatus.ServiceID,
			"",
			"FAIL")
	}
}

// DeleteMetrics for this Controller.
func (c *Controller) DeleteMetrics() {
	if c == nil {
		return
	}

	for _, cmd := range *c.Command {
		name := cmd.String()
		metric.DeletePrometheusCounter(metric.CommandResultTotal,
			name,
			*c.ServiceStatus.ServiceID,
			"",
			"SUCCESS")
		metric.DeletePrometheusCounter(metric.CommandResultTotal,
			name,
			*c.ServiceStatus.ServiceID,
			"",
			"FAIL")
	}
}

// FormattedString will convert Command to a string in the format of '[ "arg0", "arg1" ]'.
func (c *Command) FormattedString() string {
	return fmt.Sprintf("[ \"%s\" ]", strings.Join(*c, "\", \""))
}

// IsRunnable will return whether the current time at `index` is before nextRunnable.
// If out of range, it will return false.
func (c *Controller) IsRunnable(index int) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// out of range.
	if index >= len(c.nextRunnable) {
		return false
	}

	return time.Now().UTC().After(c.nextRunnable[index])
}

// NextRunnable returns the nextRunnable of the Command at `index`.
// If out of range, it will return a zero time.
func (c *Controller) NextRunnable(index int) time.Time {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// out of range.
	if index >= len(c.nextRunnable) {
		return time.Time{}
	}

	return c.nextRunnable[index]
}

// SetNextRunnable will set the `time` that the Command at `index` can be re-run.
func (c *Controller) SetNextRunnable(index int, time time.Time) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if index < len(c.nextRunnable) {
		c.nextRunnable[index] = time
	}
}

// SetExecuting will set the time the Command at `index` can be re-run. (longer if `executing`).
func (c *Controller) SetExecuting(index int, executing bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// If out of range.
	if index >= len(c.nextRunnable) {
		return
	}

	// Different times depending on pass/fail.
	if !util.DereferenceOrNilValue(c.Failed.Get(index), true) {
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
