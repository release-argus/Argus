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

	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
	"github.com/release-argus/Argus/web/metric"
)

// Exec will execute every `Command` for the controller.
func (c *Controller) Exec(logFrom logutil.LogFrom) error {
	if c == nil || c.Command == nil || len(*c.Command) == 0 {
		return nil
	}

	svcInfo := c.ServiceStatus.GetServiceInfo()
	errChan := make(chan error)
	for index := range *c.Command {
		go func(controller *Controller, index int) {
			errChan <- controller.ExecIndex(logFrom, index, svcInfo)
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
func (c *Controller) ExecIndex(
	logFrom logutil.LogFrom,
	index int,
	serviceInfo serviceinfo.ServiceInfo,
) error {
	if index >= len(*c.Command) {
		return nil
	}
	// block reruns whilst running.
	c.SetExecuting(index, true)

	// Copy Command and apply Jinja templating.
	command := (*c.Command)[index].ApplyTemplate(serviceInfo)

	// Execute.
	err := command.Exec(logFrom)

	// Set fail/not.
	failed := err != nil
	c.Failed.Set(index, err != nil)

	// Announce.
	c.AnnounceCommand(index, serviceInfo)

	metricResult := metric.ActionResultSuccess
	if failed {
		metricResult = metric.ActionResultFail
		//#nosec G104 -- Errors are logged to CL
		//nolint:errcheck // ^
		c.Notifiers.Shoutrrr.Send(
			fmt.Sprintf("Command failed for %q", serviceInfo.ID),
			(*c.Command)[index].String()+"\n"+err.Error(),
			serviceInfo,
			true)
	}
	metric.IncPrometheusCounter(metric.CommandResultTotal,
		(*c.Command)[index].String(),
		serviceInfo.ID,
		"",
		metricResult)

	return err
}

// Exec this Command and return any errors encountered.
func (c *Command) Exec(logFrom logutil.LogFrom) error {
	logutil.Log.Info(
		fmt.Sprintf("Executing '%s'", c),
		logFrom, true)
	//#nosec G204 -- Command is user defined.
	out, err := exec.Command((*c)[0], (*c)[1:]...).Output()

	if err != nil {
		logutil.Log.Error(err, logFrom, true)
	} else {
		logutil.Log.Info(string(out), logFrom, string(out) != "")
	}

	//nolint:wrapcheck
	return err
}

// ApplyTemplate applies Django templating to the Command.
func (c *Command) ApplyTemplate(serviceInfo serviceinfo.ServiceInfo) Command {
	command := Command(make([]string, len(*c)))
	copy(command, *c)
	for i, cmd := range command {
		command[i] = util.TemplateString(cmd, serviceInfo)
	}
	return command
}
