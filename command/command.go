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
	"errors"
	"fmt"
	"math/rand"
	"os/exec"
	"time"

	"github.com/release-argus/Argus/internal/logx"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/web/metric"
)

// ApplyTemplate applies Django templating to the receiver and returns a new [Command].
func (c *Command) ApplyTemplate(serviceInfo serviceinfo.ServiceInfo) Command {
	command := Command(make([]string, len(*c)))
	copy(command, *c)
	for i, cmd := range command {
		command[i] = util.TemplateString(cmd, serviceInfo)
	}
	return command
}

// Exec runs every command in the receiver, staggering starts across goroutines.
func (c *Controller) Exec(logFrom logx.LogFrom) error {
	if c == nil || c.Command == nil || len(c.Command) == 0 {
		return nil
	}

	svcInfo := c.ServiceStatus.GetServiceInfo()
	errChan := make(chan error)
	for index := range c.Command {
		go func(controller *Controller, index int) {
			errChan <- controller.ExecIndex(logFrom, index, svcInfo)
		}(c, index)

		// Space out Command starts.
		//#nosec G404 -- sleep does not need cryptographic security.
		time.Sleep(time.Duration(100+rand.Intn(150)) * time.Millisecond)
	}

	// Collect potential errors from all goroutines.
	var errs []error
	for range c.Command {
		if err := <-errChan; err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// ExecIndex executes the command at index and records metrics and failure state.
func (c *Controller) ExecIndex(
	logFrom logx.LogFrom,
	index int,
	serviceInfo serviceinfo.ServiceInfo,
) error {
	if index >= len(c.Command) {
		return nil
	}
	// block reruns whilst running.
	c.SetExecuting(index, true)

	// Copy Command and apply Jinja templating.
	command := (c.Command)[index].ApplyTemplate(serviceInfo)

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
			(c.Command)[index].String()+"\n"+err.Error(),
			serviceInfo,
			true,
		)
	}
	metric.IncPrometheusCounter(
		metric.CommandResultTotal,
		(c.Command)[index].String(),
		serviceInfo.ID,
		"",
		metricResult,
	)

	return err
}

// Exec executes the command and returns any error encountered.
func (c *Command) Exec(logFrom logx.LogFrom) error {
	logx.Info(
		fmt.Sprintf("Executing '%s'", c),
		logFrom,
		true,
	)
	//#nosec G204 -- Command is user defined.
	out, err := exec.Command((*c)[0], (*c)[1:]...).Output()

	if err != nil {
		logx.Error(err, logFrom, true)
	} else {
		logx.Info(string(out), logFrom, string(out) != "")
	}

	//nolint:wrapcheck
	return err
}
