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

// Package v1 provides the API for the webserver.
package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"

	logutil "github.com/release-argus/Argus/util/log"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// httpServiceActions returns all Commands/WebHooks of a service.
//
// Required parameters:
//
//	service_id: the ID of the Service to get the actions of.
func (api *API) httpServiceGetActions(w http.ResponseWriter, r *http.Request) {
	logFrom := logutil.LogFrom{Primary: "httpServiceActions", Secondary: getIP(r)}
	// Service to get actions of.
	targetService, _ := url.QueryUnescape(mux.Vars(r)["service_id"])

	api.Config.OrderMutex.RLock()
	svc := api.Config.Service[targetService]
	defer api.Config.OrderMutex.RUnlock()
	if svc == nil {
		err := fmt.Sprintf("service %q not found", targetService)
		logutil.Log.Error(err, logFrom, true)
		failRequest(&w,
			err,
			http.StatusNotFound)
		return
	}

	// Commands
	commandSummary := make(map[string]apitype.CommandSummary, len(svc.Command))
	if svc.CommandController != nil {
		svcInfo := svc.Status.GetServiceInfo()
		for i, cmd := range *svc.CommandController.Command {
			command := cmd.ApplyTemplate(svcInfo)
			commandSummary[command.String()] = apitype.CommandSummary{
				Failed:       svc.Status.Fails.Command.Get(i),
				NextRunnable: svc.CommandController.NextRunnable(i)}
		}
	}
	// WevHooks
	webhookSummary := make(map[string]apitype.WebHookSummary, len(svc.WebHook))
	for key, wh := range svc.WebHook {
		webhookSummary[key] = apitype.WebHookSummary{
			Failed:       svc.Status.Fails.WebHook.Get(key),
			NextRunnable: wh.NextRunnable(),
		}
	}

	msg := apitype.ActionSummary{
		Command: commandSummary,
		WebHook: webhookSummary}

	api.writeJSON(w, msg, logFrom)
}

// RunActionsPayload holds the target actions to run for a Service.
type RunActionsPayload struct {
	Target string `json:"target"`
}

// httpServiceRunActions handles approvals/rejections of the latest version of a service.
//
// Required parameters:
//
//	service_id: the ID of the Service to target.
//	target: the action to take. One of:
//		"ARGUS_ALL": approve all actions.
//		"ARGUS_FAILED": approve all failed actions.
//		"ARGUS_SKIP": skip this release.
//		"webhook_<webhook_id>": approve a specific webhook.
//		"command_<command_id>": approve a specific command.
func (api *API) httpServiceRunActions(w http.ResponseWriter, r *http.Request) {
	logFrom := logutil.LogFrom{Primary: "httpServiceRunActions", Secondary: getIP(r)}
	// Service to run actions of.
	targetService, _ := url.QueryUnescape(mux.Vars(r)["service_id"])

	// Check the service exists.
	api.Config.OrderMutex.RLock()
	svc := api.Config.Service[targetService]
	defer api.Config.OrderMutex.RUnlock()
	if svc == nil {
		err := fmt.Sprintf("service %q not found", targetService)
		logutil.Log.Error(err, logFrom, true)
		failRequest(&w,
			err,
			http.StatusNotFound)
		return
	}

	// Get target from the payload.
	payloadBytes := http.MaxBytesReader(w, r.Body, 512)
	var payload RunActionsPayload
	err := json.NewDecoder(payloadBytes).Decode(&payload)
	if err != nil {
		logutil.Log.Error(
			"Invalid payload - "+err.Error(),
			logFrom, true)
		failRequest(&w,
			"invalid payload",
			http.StatusBadRequest)
		return
	}
	if payload.Target == "" {
		errMsg := "invalid payload, target service not provided"
		logutil.Log.Error(errMsg, logFrom, true)
		failRequest(&w,
			errMsg,
			http.StatusBadRequest)
		return
	}
	if !svc.Options.GetActive() {
		errMsg := "service is inactive, actions can't be run for it"
		logutil.Log.Error(errMsg, logFrom, true)
		failRequest(&w,
			errMsg,
			http.StatusBadRequest)
		return
	}

	// SKIP this release.
	if payload.Target == "ARGUS_SKIP" {
		msg := fmt.Sprintf("%q release skip - %q",
			targetService, svc.Status.LatestVersion())
		logutil.Log.Info(msg, logFrom, true)
		svc.HandleSkip()
		return
	}

	if svc.WebHook == nil && svc.Command == nil {
		logutil.Log.Error(
			fmt.Sprintf("%q does not have any commands/webhooks to approve", targetService),
			logFrom, true)
		return
	}

	// Send the WebHooks.
	msg := fmt.Sprintf("%s %q Release actioned - %q",
		targetService,
		svc.Status.LatestVersion(),
		strings.ReplaceAll(
			strings.ReplaceAll(
				strings.ReplaceAll(payload.Target,
					"ARGUS_ALL", "ALL"),
				"ARGUS_FAILED", "ALL UNSENT/FAILED"),
			"ARGUS_SKIP", "SKIP"),
	)
	logutil.Log.Info(msg, logFrom, true)
	switch payload.Target {
	case "ARGUS_ALL", "ARGUS_FAILED":
		go svc.HandleFailedActions()
	default:
		if strings.HasPrefix(payload.Target, "webhook_") {
			go svc.HandleWebHook(strings.TrimPrefix(payload.Target, "webhook_"))
		} else {
			go svc.HandleCommand(strings.TrimPrefix(payload.Target, "command_"))
		}
	}

	api.writeJSON(w, apitype.Response{
		Message: msg,
	}, logFrom)
}
