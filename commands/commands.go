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
	"os/exec"
	"time"

	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	metric "github.com/release-argus/Argus/web/metrics"
)

// Exec will execute all `Command` for the controller and returns all errors encountered
func (c *Controller) Exec(logFrom *util.LogFrom) (errs error) {
	if c == nil || c.Command == nil || len(*c.Command) == 0 {
		return errs
	}

	errChan := make(chan error)
	for index := range *c.Command {
		index := index
		go func(controller *Controller) {
			errChan <- controller.ExecIndex(logFrom, index)
		}(c)

		// Space out Command starts.
		time.Sleep(200 * time.Millisecond)
	}

	for range *c.Command {
		err := <-errChan
		if err != nil {
			errs = fmt.Errorf("%s\n%w",
				util.ErrorToString(errs), err)
		}
	}

	return
}

func (c *Controller) ExecIndex(logFrom *util.LogFrom, index int) (err error) {
	if index >= len(*c.Command) {
		return
	}
	// block reruns whilst running
	c.SetExecuting(index, true)

	// Copy Command and apply Jinja templating
	command := (*c.Command)[index].ApplyTemplate(c.ServiceStatus)

	// Execute
	err = command.Exec(logFrom)

	// Set fail/not
	failed := err != nil
	c.Failed.Set(index, failed)

	// Announce
	c.AnnounceCommand(index)

	metricResult := "SUCCESS"
	if failed {
		metricResult = "FAIL"
		//#nosec G104 -- Errors will be logged to CL
		//nolint:errcheck // ^
		c.Notifiers.Shoutrrr.Send(
			fmt.Sprintf("Command failed for %q", *c.ServiceStatus.ServiceID),
			(*c.Command)[index].String()+"\n"+err.Error(),
			&util.ServiceInfo{ID: *c.ServiceStatus.ServiceID},
			true)
	}
	metric.IncreasePrometheusCounter(metric.CommandMetric,
		(*c.Command)[index].String(),
		*c.ServiceStatus.ServiceID,
		"",
		metricResult)

	return err
}

func (c *Command) Exec(logFrom *util.LogFrom) error {
	jLog.Info(fmt.Sprintf("Executing '%s'", c), *logFrom, true)
	out, err := exec.Command((*c)[0], (*c)[1:]...).Output()

	jLog.Error(util.ErrorToString(err), *logFrom, err != nil)
	jLog.Info(string(out), *logFrom, err == nil && string(out) != "")

	//nolint:wrapcheck
	return err
}

func (c *Command) ApplyTemplate(serviceStatus *svcstatus.Status) (command Command) {
	if serviceStatus == nil {
		return *c
	}

	command = Command(make([]string, len(*c)))
	copy(command, *c)
	serviceInfo := util.ServiceInfo{LatestVersion: serviceStatus.GetLatestVersion()}
	for i := range command {
		command[i] = util.TemplateString(command[i], serviceInfo)
	}
	return
}
