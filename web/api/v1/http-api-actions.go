// Copyright [2023] [Argus]
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

package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/release-argus/Argus/util"
	api_type "github.com/release-argus/Argus/web/api/types"
)

// httpServiceActions will return all Command(s)/WebHook(s) of a service.
//
// Required params:
//
// service_name - Service ID to get the actions of.
func (api *API) httpServiceGetActions(w http.ResponseWriter, r *http.Request) {
	logFrom := util.LogFrom{Primary: "httpServiceActions", Secondary: getIP(r)}
	targetService, _ := url.QueryUnescape(mux.Vars(r)["service_name"])
	jLog.Verbose(targetService, &logFrom, true)

	api.Config.OrderMutex.RLock()
	svc := api.Config.Service[targetService]
	defer api.Config.OrderMutex.RUnlock()
	if svc == nil {
		err := fmt.Sprintf("service %q not found", targetService)
		jLog.Error(err, &logFrom, true)
		failRequest(&w, err, http.StatusNotFound)
		return
	}

	// Commands
	commandSummary := make(map[string]api_type.CommandSummary, len(svc.Command))
	if svc.CommandController != nil {
		for i := range *svc.CommandController.Command {
			command := (*svc.CommandController.Command)[i].ApplyTemplate(&svc.Status)
			commandSummary[command.String()] = api_type.CommandSummary{
				Failed:       svc.Status.Fails.Command.Get(i),
				NextRunnable: svc.CommandController.NextRunnable(i)}
		}
	}
	// WevHooks
	webhookSummary := make(map[string]api_type.WebHookSummary, len(svc.WebHook))
	for key := range svc.WebHook {
		webhookSummary[key] = api_type.WebHookSummary{
			Failed:       svc.Status.Fails.WebHook.Get(key),
			NextRunnable: svc.WebHook[key].NextRunnable(),
		}
	}

	msg := api_type.ActionSummary{
		Command: commandSummary,
		WebHook: webhookSummary}

	err := json.NewEncoder(w).Encode(msg)
	jLog.Error(err, &logFrom, err != nil)
}

type RunActionsPayload struct {
	Target *string `json:"target"`
}

// httpServiceRunActions handles approvals/rejections of the latest version of a service.
//
// Required params:
//
// service_name - Service ID to target.
//
// target - The action to take. Can be one of:
//   - "ARGUS_ALL" - Approve all actions.
//   - "ARGUS_FAILED" - Approve all failed actions.
//   - "ARGUS_SKIP" - Skip this release.
//   - "webhook_<webhook_id>" - Approve a specific WebHook.
//   - "command_<command_id>" - Approve a specific Command.
func (api *API) httpServiceRunActions(w http.ResponseWriter, r *http.Request) {
	logFrom := &util.LogFrom{Primary: "httpServiceRunActions", Secondary: getIP(r)}
	targetService, _ := url.QueryUnescape(mux.Vars(r)["service_name"])
	jLog.Verbose(targetService, logFrom, true)

	// Check the service exists.
	api.Config.OrderMutex.RLock()
	svc := api.Config.Service[targetService]
	defer api.Config.OrderMutex.RUnlock()
	if svc == nil {
		err := fmt.Sprintf("service %q not found", targetService)
		jLog.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusNotFound)
		return
	}

	// Get target from payload.
	payloadBytes := http.MaxBytesReader(w, r.Body, 102400)
	var payload RunActionsPayload
	err := json.NewDecoder(payloadBytes).Decode(&payload)
	if err != nil {
		jLog.Error(fmt.Sprintf("Invalid payload - %v", err), logFrom, true)
		failRequest(&w, "invalid payload", http.StatusBadRequest)
		return
	}
	if payload.Target == nil {
		errMsg := "invalid payload, target service not provided"
		jLog.Error(errMsg, logFrom, true)
		failRequest(&w, errMsg, http.StatusBadRequest)
		return
	}
	if !svc.Options.GetActive() {
		errMsg := "service is inactive, actions can't be run for it"
		jLog.Error(errMsg, logFrom, true)
		failRequest(&w, errMsg, http.StatusBadRequest)
		return
	}

	// SKIP this release
	if *payload.Target == "ARGUS_SKIP" {
		msg := fmt.Sprintf("%q release skip - %q",
			targetService, svc.Status.LatestVersion())
		jLog.Info(msg, logFrom, true)
		svc.HandleSkip()
		return
	}

	if svc.WebHook == nil && svc.Command == nil {
		jLog.Error(fmt.Sprintf("%q does not have any commands/webhooks to approve", targetService), logFrom, true)
		return
	}

	// Send the WebHook(s).
	msg := fmt.Sprintf("%s %q Release actioned - %q",
		targetService,
		svc.Status.LatestVersion(),
		strings.ReplaceAll(
			strings.ReplaceAll(
				strings.ReplaceAll(*payload.Target,
					"ARGUS_ALL", "ALL"),
				"ARGUS_FAILED", "ALL UNSENT/FAILED"),
			"ARGUS_SKIP", "SKIP"),
	)
	jLog.Info(msg, logFrom, true)
	switch *payload.Target {
	case "ARGUS_ALL", "ARGUS_FAILED":
		go svc.HandleFailedActions()
	default:
		if strings.HasPrefix(*payload.Target, "webhook_") {
			go svc.HandleWebHook(strings.TrimPrefix(*payload.Target, "webhook_"))
		} else {
			go svc.HandleCommand(strings.TrimPrefix(*payload.Target, "command_"))
		}
	}
}
