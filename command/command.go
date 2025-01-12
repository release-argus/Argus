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
	"errors"
	"fmt"
	"math/rand"
	"os/exec"
	"time"

	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/web/metric"
)

// Exec will execute every `Command` for the controller.
func (c *Controller) Exec(logFrom util.LogFrom) error {
	if c == nil || c.Command == nil || len(*c.Command) == 0 {
		return nil
	}

	errChan := make(chan error)
	for index := range *c.Command {
		go func(controller *Controller, index int) {
			errChan <- controller.ExecIndex(logFrom, index)
		}(c, index)

		// Space out Command starts.
		//#nosec G404 -- sleep does not need cryptographic security.
		time.Sleep(time.Duration(100+rand.Intn(150)) * time.Millisecond)
	}

	// Collect potential errors from all goroutines.
	var errs []error
	for range *c.Command {
		if err := <-errChan; err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// ExecIndex will execute the `Command` at the given index.
func (c *Controller) ExecIndex(logFrom util.LogFrom, index int) error {
	if index >= len(*c.Command) {
		return nil
	}
	// block reruns whilst running.
	c.SetExecuting(index, true)

	// Copy Command and apply Jinja templating.
	command := (*c.Command)[index].ApplyTemplate(c.ServiceStatus)

	// Execute.
	err := command.Exec(logFrom)

	// Set fail/not.
	failed := err != nil
	c.Failed.Set(index, err != nil)

	// Announce.
	c.AnnounceCommand(index)

	metricResult := "SUCCESS"
	if failed {
		metricResult = "FAIL"
		//#nosec G104 -- Errors are logged to CL
		//nolint:errcheck // ^
		c.Notifiers.Shoutrrr.Send(
			fmt.Sprintf("Command failed for %q", *c.ServiceStatus.ServiceID),
			(*c.Command)[index].String()+"\n"+err.Error(),
			util.ServiceInfo{ID: *c.ServiceStatus.ServiceID},
			true)
	}
	metric.IncPrometheusCounter(metric.CommandResultTotal,
		(*c.Command)[index].String(),
		*c.ServiceStatus.ServiceID,
		"",
		metricResult)

	return err
}

// Exec this Command and return any errors encountered.
func (c *Command) Exec(logFrom util.LogFrom) error {
	jLog.Info(
		fmt.Sprintf("Executing '%s'", c),
		logFrom, true)
	//#nosec G204 -- Command is user defined.
	out, err := exec.Command((*c)[0], (*c)[1:]...).Output()

	if err != nil {
		jLog.Error(err, logFrom, true)
	} else {
		jLog.Info(string(out), logFrom, string(out) != "")
	}

	//nolint:wrapcheck
	return err
}

// ApplyTemplate applies Jinja templating to the Command.
func (c *Command) ApplyTemplate(serviceStatus *status.Status) Command {
	// Can't template without serviceStatus.
	if serviceStatus == nil {
		return *c
	}

	command := Command(make([]string, len(*c)))
	copy(command, *c)
	serviceInfo := util.ServiceInfo{LatestVersion: serviceStatus.LatestVersion()}
	for i, cmd := range command {
		command[i] = util.TemplateString(cmd, serviceInfo)
	}
	return command
}
