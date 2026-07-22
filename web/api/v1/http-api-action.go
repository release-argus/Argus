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

// Package v1 provides the API for the webserver.
package v1

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/release-argus/Argus/internal/logx"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// httpServiceGetActions handles GET /api/v1/service/actions: returns a list of
// all Commands/WebHooks attached to a service.
//
// Query parameters:
//
//	service_id: the ID of the Service to get the actions of.
//
// Response:
//
//	200 OK: JSON object containing the service's Commands/WebHooks.
//	404 Not Found: on an unknown service.
func (api *API) httpServiceGetActions(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpServiceActions", Secondary: getIP(r)}
	// Service to get actions of.
	serviceID, ok := requireQueryParam(w, r, "service_id")
	if !ok {
		return
	}

	api.Config.OrderMu.RLock()
	svc := api.Config.Service[serviceID]
	defer api.Config.OrderMu.RUnlock()
	if svc == nil {
		err := fmt.Errorf("service %q not found", serviceID)
		logx.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusNotFound)
		return
	}

	// Commands.
	commandSummary := make(map[string]apitype.CommandSummary, len(svc.Command))
	if svc.CommandController != nil {
		svcInfo := svc.Status.GetServiceInfo()
		for i, cmd := range svc.CommandController.Command {
			command := cmd.ApplyTemplate(svcInfo)
			commandSummary[command.String()] = apitype.CommandSummary{
				Failed:       svc.Status.Fails.Command.Get(i),
				NextRunnable: svc.CommandController.NextRunnable(i),
			}
		}
	}
	// WebHooks.
	webhookSummary := make(map[string]apitype.WebHookSummary, len(svc.WebHook))
	for key, wh := range svc.WebHook {
		webhookSummary[key] = apitype.WebHookSummary{
			Failed:       svc.Status.Fails.WebHook.Get(key),
			NextRunnable: wh.NextRunnable(),
		}
	}

	msg := apitype.ActionSummary{
		Command: commandSummary,
		WebHook: webhookSummary,
	}

	api.writeJSON(w, msg, logFrom)
}

// RunActionsPayload holds the target actions to run for a Service.
type RunActionsPayload struct {
	Target string `json:"target"`
}

// Action codes.
const (
	ActionAll    = "ARGUS_ALL"
	ActionSkip   = "ARGUS_SKIP"
	ActionFailed = "ARGUS_FAILED"
)

// httpServiceRunActions handles POST /api/v1/service/actions: approvals/rejections
// of Commands/WebHooks on the latest version of a service.
//
// Query parameters:
//
//	service_id: ID of the Service to target.
//	target: Action to take. Supported values:
//	  ActionAll: approve all actions.
//	  ActionFailed: approve all failed actions.
//	  ActionSkip: skip this release.
//	  webhook_<webhook_id>: approve a specific webhook.
//	  command_<command_id>: approve a specific command.
//
// Response:
//
//	200 OK: on success.
//	400 Bad Request: on a malformed payload, missing target, or inactive service.
//	404 Not Found: on an unknown service.
func (api *API) httpServiceRunActions(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpServiceRunActions", Secondary: getIP(r)}
	// Service to run actions of.
	serviceID, ok := requireQueryParam(w, r, "service_id")
	if !ok {
		return
	}

	// Check the service exists.
	api.Config.OrderMu.RLock()
	svc := api.Config.Service[serviceID]
	defer api.Config.OrderMu.RUnlock()
	if svc == nil {
		err := fmt.Errorf("service %q not found", serviceID)
		logx.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusNotFound)
		return
	}

	// Get target from the payload.
	payloadBytes := http.MaxBytesReader(w, r.Body, 512)
	var payload RunActionsPayload
	if err := json.UnmarshalRead(payloadBytes, &payload); err != nil {
		err = fmt.Errorf("invalid payload: %w", err)
		logx.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusBadRequest)
		return
	}
	if payload.Target == "" {
		err := errors.New("invalid payload, target service not provided")
		logx.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusBadRequest)
		return
	}
	if !svc.Options.GetActive() {
		err := errors.New("service is inactive, actions can't be run for it")
		logx.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusBadRequest)
		return
	}

	// SKIP this release.
	if payload.Target == ActionSkip {
		msg := fmt.Sprintf(
			"%q release skip - %q",
			serviceID, svc.Status.LatestVersion(),
		)
		logx.Info(msg, logFrom, true)
		svc.HandleSkip()
		return
	}

	if svc.WebHook == nil && svc.Command == nil {
		logx.Warn(
			fmt.Sprintf("%q does not have any commands/webhooks to approve", serviceID),
			logFrom,
			true,
		)
		return
	}

	// Send the WebHooks.
	msg := fmt.Sprintf(
		"%s %q Release actioned - %q",
		serviceID,
		svc.Status.LatestVersion(),
		strings.ReplaceAll(
			strings.ReplaceAll(
				strings.ReplaceAll(
					payload.Target, ActionAll, "ALL",
				),
				ActionFailed, "ALL UNSENT/FAILED",
			),
			ActionSkip, "SKIP",
		),
	)
	logx.Info(msg, logFrom, true)
	switch payload.Target {
	case ActionAll, ActionFailed:
		go svc.HandleFailedActions()
	default:
		if after, ok := strings.CutPrefix(payload.Target, "webhook_"); ok {
			go svc.HandleWebHook(after)
		} else {
			go svc.HandleCommand(strings.TrimPrefix(payload.Target, "command_"))
		}
	}

	api.writeJSON(
		w,
		apitype.Response{
			Message: msg,
		},
		logFrom,
	)
}
