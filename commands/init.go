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
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
	metrics "github.com/release-argus/Argus/web/metrics"
)

// Init the Command Controller.
func (c *Controller) Init(
	log *utils.JLog,
	serviceID *string,
	serviceStatus *service_status.Status,
	command *Slice,
	shoutrrrNotifiers *shoutrrr.Slice,
	parentInterval *string,
) {
	if log != nil {
		jLog = log
	}
	if c == nil || len(*command) == 0 {
		return
	}

	c.ServiceStatus = serviceStatus
	c.Command = command
	c.Failed = &serviceStatus.Fails.Command
	c.NextRunnable = make([]time.Time, len(*c.Command))

	parentID := *serviceID
	c.ServiceID = &parentID
	c.ParentInterval = parentInterval
	c.initMetrics()

	// Command fail notifiers
	(*c).Notifiers = Notifiers{
		Shoutrrr: shoutrrrNotifiers,
	}
}

// initMetrics, giving them all a starting value.
func (c *Controller) initMetrics() {
	// ############
	// # Counters #
	// ############
	for i := range *c.Command {
		name := (*c.Command)[i].String()
		metrics.InitPrometheusCounterActions(metrics.CommandMetric, name, *c.ServiceID, "", "SUCCESS")
		metrics.InitPrometheusCounterActions(metrics.CommandMetric, name, *c.ServiceID, "", "FAIL")
	}

	// ##########
	// # Gauges #
	// ##########
	metrics.SetPrometheusGaugeWithID(metrics.AckWaiting, *c.ServiceID, float64(0))
}

// FormattedString will convert Command to a string in the format of '[ "arg0", "arg1" ]'
func (c *Command) FormattedString() string {
	return fmt.Sprintf("[ \"%s\" ]", strings.Join(*c, "\", \""))
}

// String will convert Command to a string in the format of 'arg0 arg1'
func (c *Command) String() string {
	return strings.Join(*c, " ")
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
	if !(index < len(c.NextRunnable)) {
		return false
	}

	return time.Now().UTC().After(c.NextRunnable[index])
}

// SetNextRunnable time that the Command at index can be re-run.
func (c *Controller) SetNextRunnable(index int, executing bool) {
	// If out of range
	if !(index < len(c.NextRunnable)) {
		return
	}

	// Different times depending on pass/fail
	if !utils.EvalNilPtr((*c.Failed)[index], true) {
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
