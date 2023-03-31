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

package command

import (
	"fmt"
	"strings"
	"time"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	metric "github.com/release-argus/Argus/web/metrics"
)

// LogInit for this package.
func LogInit(log *util.JLog) {
	jLog = log
}

// Init the Command Controller.
func (c *Controller) Init(
	serviceStatus *svcstatus.Status,
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
	if len(*c.Failed) != len(*c.Command) {
		*c.Failed = make([]*bool, len(*c.Command))
	}
	c.NextRunnable = make([]time.Time, len(*c.Command))

	c.ParentInterval = parentInterval

	// Command fail notifiers
	c.Notifiers = Notifiers{
		Shoutrrr: shoutrrrNotifiers,
	}
}

// initMetrics, giving them all a starting value.
func (c *Controller) InitMetrics() {
	if c == nil {
		return
	}

	// ############
	// # Counters #
	// ############
	for i := range *c.Command {
		name := (*c.Command)[i].String()
		metric.InitPrometheusCounter(metric.CommandMetric,
			name,
			*c.ServiceStatus.ServiceID,
			"",
			"SUCCESS")
		metric.InitPrometheusCounter(metric.CommandMetric,
			name,
			*c.ServiceStatus.ServiceID,
			"",
			"FAIL")
	}

	// ##########
	// # Gauges #
	// ##########
	metric.SetPrometheusGauge(metric.AckWaiting,
		*c.ServiceStatus.ServiceID,
		float64(0))
}

// DeleteMetrics for this Controller.
func (c *Controller) DeleteMetrics() {
	if c == nil {
		return
	}

	for i := range *c.Command {
		name := (*c.Command)[i].String()
		metric.DeletePrometheusCounter(metric.CommandMetric,
			name,
			*c.ServiceStatus.ServiceID,
			"",
			"SUCCESS")
		metric.DeletePrometheusCounter(metric.CommandMetric,
			name,
			*c.ServiceStatus.ServiceID,
			"",
			"FAIL")
	}

	metric.DeletePrometheusGauge(metric.AckWaiting,
		*c.ServiceStatus.ServiceID)
}

// FormattedString will convert Command to a string in the format of '[ "arg0", "arg1" ]'
func (c *Command) FormattedString() string {
	return fmt.Sprintf("[ \"%s\" ]", strings.Join(*c, "\", \""))
}

// GetNextRunnable returns the NextRunnable of this WebHook as time.time.
func (c *Controller) GetNextRunnable(index int) (at time.Time) {
	if index < len(c.NextRunnable) {
		at = c.NextRunnable[index]
	}
	return
}

// IsRunnable will return whether the current time is before NextRunnable
func (c *Controller) IsRunnable(index int) bool {
	// If out of range
	if index >= len(c.NextRunnable) {
		return false
	}

	return time.Now().UTC().After(c.NextRunnable[index])
}

// SetNextRunnable time that the Command at index can be re-run.
func (c *Controller) SetNextRunnable(index int, executing bool) {
	// If out of range
	if index >= len(c.NextRunnable) {
		return
	}

	// Different times depending on pass/fail
	if !util.EvalNilPtr((*c.Failed)[index], true) {
		parentInterval, _ := time.ParseDuration(*c.ParentInterval)
		c.NextRunnable[index] = time.Now().UTC().Add(2 * parentInterval)
	} else {
		c.NextRunnable[index] = time.Now().UTC().Add(15 * time.Second)
	}

	// Block reruns whilst running for up to an hour
	if executing {
		c.NextRunnable[index] = c.NextRunnable[index].Add(time.Hour)
	}
}
