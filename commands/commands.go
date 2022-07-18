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

	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
	metrics "github.com/release-argus/Argus/web/metrics"
)

// Init will set the logger for the package
func Init(log *utils.JLog) {
	jLog = log
}

// Exec will execute all `Command` for the controller and returns all errors encountered
func (c *Controller) Exec(logFrom *utils.LogFrom) (errs error) {
	if c == nil || c.Command == nil || len(*c.Command) == 0 {
		return nil
	}

	errChan := make(chan error)
	for index := range *c.Command {
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
				utils.ErrorToString(errs), err)
		}
	}

	return
}

func (c *Controller) ExecIndex(logFrom *utils.LogFrom, index int) error {
	if index >= len(*c.Command) {
		return nil
	}
	// block reruns whilst running
	c.SetNextRunnable(index, true)

	// Copy Command and apply Jinja templating
	command := (*c.Command)[index].ApplyTemplate(c.ServiceStatus)

	// Execute
	err := command.Exec(logFrom)

	// Set fail/not
	failed := err != nil
	c.Failed[index] = &failed

	// Announce
	c.AnnounceCommand(index)

	metricResult := "SUCCESS"
	if failed {
		metricResult = "FAIL"
		//#nosec G104 -- Errors will be logged to CL
		//nolint:errcheck // ^
		c.Notifiers.Shoutrrr.Send(
			"Command failed for "+*c.ServiceID,
			(*c.Command)[index].String()+"\n"+err.Error(),
			nil,
			true)
	}
	metrics.IncreasePrometheusCounterActions(metrics.CommandMetric, (*c.Command)[index].String(), *c.ServiceID, "", metricResult)

	return err
}

func (c *Command) Exec(logFrom *utils.LogFrom) error {
	jLog.Info(fmt.Sprintf("Executing '%s'", c.String()), *logFrom, true)
	out, err := exec.Command((*c)[0], (*c)[1:]...).Output()

	jLog.Error(utils.ErrorToString(err), *logFrom, err != nil)
	jLog.Info(string(out), *logFrom, err == nil && string(out) != "")

	return err
}

func (c *Command) ApplyTemplate(serviceStatus *service_status.Status) Command {
	if serviceStatus == nil {
		return *c
	}

	command := Command(make([]string, len(*c)))
	copy(command, *c)
	for i := range command {
		command[i] = utils.TemplateString(command[i], utils.ServiceInfo{LatestVersion: serviceStatus.LatestVersion})
	}
	return command
}
