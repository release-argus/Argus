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
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/dashboard"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	dvbase "github.com/release-argus/Argus/service/deployed_version/types/base"
	latestver "github.com/release-argus/Argus/service/latest_version"
	lvbase "github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/service/status"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// httpLatestVersionRefreshUncreated creates a latest version lookup and queries it.
//
// Method: GET
//
// Query Parameters:
//
//	semantic_versioning: Optional boolean parameter to override semantic versioning defaults.
//	overrides: Required parameter to provide parameters for the version lookup.
//
// Response:
//
//	On success: JSON object containing the refreshed version and the current UTC datetime.
//	On error: HTTP 400 Bad Request with an error message.
func (api *API) httpLatestVersionRefreshUncreated(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpVersionRefreshUncreated_Latest", Secondary: getIP(r)}

	queryParams := r.URL.Query()

	overrides := queryParams.Get("overrides")
	// Verify overrides are provided.
	if overrides == "" {
		err := errors.New("overrides: <required>")
		logx.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusBadRequest)
		return
	}

	// Status
	svcStatus := status.Status{}
	svcStatus.Init(
		0, 0, 0,
		status.ServiceInfo{
			ID: logFrom.Primary,
		},
		&dashboard.Options{},
	)

	// Options
	options, _ := opt.Decode(
		"json", []byte("{}"),
		opt.DefaultsConfig{
			Soft: api.Config.Defaults.Service.LatestVersion.Options,
			Hard: api.Config.HardDefaults.Service.LatestVersion.Options,
		},
	)
	options.SemanticVersioning = util.StringToBoolPtr(queryParams.Get("semantic_versioning"))

	// Create the LatestVersionLookup.
	lv, err := latestver.Decode(
		"json", []byte(overrides),
		options,
		&svcStatus,
		lvbase.DefaultsConfig{
			Soft: &api.Config.Defaults.Service.LatestVersion,
			Hard: &api.Config.HardDefaults.Service.LatestVersion,
		},
	)
	if err == nil {
		err = lv.CheckValues()
	}
	// Error creating/validating the LatestVersionLookup.
	if err != nil {
		logx.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusBadRequest)
		return
	}

	// Query the latest version lookup.
	version, _, err := latestver.Refresh(
		lv,
		nil, nil, nil,
	)
	if err != nil {
		failRequest(&w, err, http.StatusBadRequest)
		return
	}

	// Return the version found.
	api.writeJSON(
		w,
		apitype.RefreshAPI{
			Version: version,
			Date:    time.Now().UTC(),
		},
		logFrom,
	)
}

// httpDeployedVersionRefreshUncreated creates a deployed version lookup and queries it.
//
// Method: GET
//
// Query Parameters:
//
//	semantic_versioning: Optional boolean parameter to override semantic versioning defaults.
//	overrides: Required parameter to provide parameters for the version lookup.
//
// Response:
//
//	On success: JSON object containing the refreshed version and the current UTC datetime.
//	On error: HTTP 400 Bad Request with an error message.
func (api *API) httpDeployedVersionRefreshUncreated(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpVersionRefreshUncreated_Deployed", Secondary: getIP(r)}

	queryParams := r.URL.Query()
	overrides := queryParams.Get("overrides")

	// Verify overrides are provided.
	if overrides == "" {
		err := errors.New("overrides: <required>")
		logx.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusBadRequest)
		return
	}

	// Status
	svcStatus := status.Status{}
	svcStatus.Init(
		0, 0, 0,
		status.ServiceInfo{
			ID: logFrom.Primary,
		},
		&dashboard.Options{},
	)

	// Options.
	options, _ := opt.Decode(
		"json", []byte("{}"),
		opt.DefaultsConfig{
			Soft: api.Config.Defaults.Service.LatestVersion.Options,
			Hard: api.Config.HardDefaults.Service.LatestVersion.Options,
		},
	)
	options.SemanticVersioning = util.StringToBoolPtr(queryParams.Get("semantic_versioning"))

	// Create the DeployedVersionLookup.
	dvl, err := deployedver.Decode(
		"json", []byte(overrides),
		options,
		&svcStatus,
		dvbase.DefaultsConfig{
			Soft: &api.Config.Defaults.Service.DeployedVersionLookup,
			Hard: &api.Config.HardDefaults.Service.DeployedVersionLookup,
		},
	)
	if err == nil {
		err = dvl.CheckValues()
	}
	// Error creating/validating the DeployedVersionLookup.
	if err != nil {
		logx.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusBadRequest)
		return
	}

	// Query the deployed version lookup.
	version, err := deployedver.Refresh(
		dvl,
		"", nil, nil, nil,
	)
	if err != nil {
		failRequest(&w, err, http.StatusBadRequest)
		return
	}

	// Return the version found.
	api.writeJSON(
		w,
		apitype.RefreshAPI{
			Version: version,
			Date:    time.Now().UTC(),
		},
		logFrom,
	)
}

// httpLatestVersionRefresh refreshes the latest version of the target service.
//
// Method: GET
//
// Query Parameters:
//
//	service_id: The ID of the Service to refresh the LatestVersion of.
//	overrides: Optional parameter to provide parameters for the version lookup.
//	semantic_versioning: Optional boolean parameter to override semantic versioning defaults.
//
// Response:
//
//	On success: JSON object containing the refreshed version and the current UTC datetime.
//	On error: HTTP 400 Bad Request with an error message.
func (api *API) httpLatestVersionRefresh(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpVersionRefresh_Latest", Secondary: getIP(r)}
	// Service to refresh.
	serviceID, ok := requireQueryParam(w, r, "service_id")
	if !ok {
		return
	}

	// Refreshes take the per-service lock shared; reject if an edit/delete holds it.
	op := api.acquireServiceOp(serviceID)
	defer api.releaseServiceOp(serviceID, op)
	if !op.mu.TryRLock() {
		failRequest(
			&w,
			fmt.Errorf("refresh %q failed, another operation is in progress for this service", serviceID),
			http.StatusConflict,
		)
		return
	}
	defer op.mu.RUnlock()

	queryParams := r.URL.Query()

	// Check whether service exists.
	api.Config.OrderMu.RLock()
	svc := api.Config.Service[serviceID]
	api.Config.OrderMu.RUnlock()
	if svc == nil {
		err := fmt.Errorf("service %q not found", serviceID)
		logx.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusNotFound)
		return
	}

	// Parameters
	var (
		overrides          = queryParams.Get("overrides")
		semanticVersioning = getParam(queryParams, "semantic_versioning")
	)

	var overrideBytes []byte
	// SecretRefs.
	var secretRefs *shared.VSecretRef
	if overrides != "" {
		overrideBytes = []byte(overrides)
		if err := decode.Unmarshal("json", overrideBytes, &secretRefs); err != nil {
			logx.Error(err, logFrom, true)
			failRequest(&w, err, http.StatusBadRequest)
			return
		}
	}

	// Query the latest version lookup.
	version, announce, err := latestver.Refresh(
		svc.LatestVersion,
		overrideBytes,
		semanticVersioning,
		secretRefs,
	)
	if announce {
		svc.HandleUpdateActions(true)
	}
	if err != nil {
		failRequest(&w, err, http.StatusBadRequest)
		return
	}

	// Return the version found.
	api.writeJSON(
		w,
		apitype.RefreshAPI{
			Version: version,
			Date:    time.Now().UTC(),
		},
		logFrom,
	)
}

// httpDeployedVersionRefresh refreshes the deployed version of the target service.
//
// Method: GET
//
// Query Parameters:
//
//	service_id: The ID of the Service to refresh the DeployedVersion of.
//	overrides: Optional parameter to provide parameters for the version lookup.
//	semantic_versioning: Optional boolean parameter to override semantic versioning defaults.
//
// Response:
//
//	On success: JSON object containing the refreshed version and the current UTC datetime.
//	On error: HTTP 400 Bad Request with an error message.
func (api *API) httpDeployedVersionRefresh(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpVersionRefresh_Deployed", Secondary: getIP(r)}
	// Service to refresh.
	serviceID, ok := requireQueryParam(w, r, "service_id")
	if !ok {
		return
	}

	// Refreshes take the per-service lock shared; reject if an edit/delete holds it.
	op := api.acquireServiceOp(serviceID)
	defer api.releaseServiceOp(serviceID, op)
	if !op.mu.TryRLock() {
		failRequest(
			&w,
			fmt.Errorf("refresh %q failed, another operation is in progress for this service", serviceID),
			http.StatusConflict,
		)
		return
	}
	defer op.mu.RUnlock()

	queryParams := r.URL.Query()

	// Check whether service exists.
	api.Config.OrderMu.RLock()
	svc := api.Config.Service[serviceID]
	api.Config.OrderMu.RUnlock()
	if svc == nil {
		err := fmt.Errorf("service %q not found", serviceID)
		logx.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusNotFound)
		return
	}

	// Parameters.
	var (
		overrides          = queryParams.Get("overrides")
		semanticVersioning = getParam(queryParams, "semantic_versioning")
	)

	var overrideBytes []byte
	// SecretRefs.
	var secretRefs *shared.VSecretRef
	if overrides != "" {
		overrideBytes = []byte(overrides)
		if err := decode.Unmarshal("json", overrideBytes, &secretRefs); err != nil {
			logx.Error(err, logFrom, true)
			failRequest(&w, err, http.StatusBadRequest)
			return
		}
	}

	// Existing DeployedVersionLookup?
	var previousType string
	dvl := svc.DeployedVersionLookup
	// Must create the DeployedVersionLookup if it doesn't exist.
	if dvl == nil {
		if overrides == "" {
			err := errors.New("missing required parameter: overrides")
			failRequest(&w, err, http.StatusBadRequest)
			return
		}

		// Status
		svcStatus := status.Status{}
		svcStatus.Init(
			0, 0, 0,
			status.ServiceInfo{
				ID: logFrom.Primary,
			},
			&dashboard.Options{},
		)

		dvl, _ = deployedver.Decode(
			"json", overrideBytes,
			&svc.Options,
			&svcStatus,
			dvbase.DefaultsConfig{
				Soft: &api.Config.Defaults.Service.DeployedVersionLookup,
				Hard: &api.Config.HardDefaults.Service.DeployedVersionLookup,
			},
		)
	} else {
		previousType = dvl.GetType()
	}

	// Query the deployed version lookup.
	version, err := deployedver.Refresh(
		dvl,
		previousType,
		overrideBytes,
		semanticVersioning,
		secretRefs,
	)
	if err != nil {
		failRequest(&w, err, http.StatusBadRequest)
		return
	}

	// Return the version found.
	api.writeJSON(
		w,
		apitype.RefreshAPI{
			Version: version,
			Date:    time.Now().UTC(),
		},
		logFrom,
	)
}

// httpServiceDetail handles sending details about a Service.
//
// Method: GET
//
// Query Parameters:
//
//	service_id: The ID of the Service to get details for.
//
// Response:
//
//	JSON object containing the service details.
func (api *API) httpServiceDetail(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpServiceDetail", Secondary: getIP(r)}
	// Service to get details of.
	serviceID, ok := requireQueryParam(w, r, "service_id")
	if !ok {
		return
	}

	// Find the Service.
	api.Config.OrderMu.RLock()
	svc := api.Config.Service[serviceID]
	// Convert to API Type, censoring secrets.
	serviceConfig := convertAndCensorService(svc)
	api.Config.OrderMu.RUnlock()

	if svc == nil {
		err := fmt.Errorf("service %q not found", serviceID)
		logx.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusNotFound)
		return
	}

	// Convert to JSON type that swaps slices for lists.
	serviceJSON := apitype.ServiceEdit{
		Name:                  serviceConfig.Name,
		Comment:               serviceConfig.Comment,
		Options:               serviceConfig.Options,
		LatestVersion:         serviceConfig.LatestVersion,
		Command:               serviceConfig.Command,
		Notify:                serviceConfig.Notify.Flatten(),
		WebHook:               serviceConfig.WebHook.Flatten(),
		DeployedVersionLookup: serviceConfig.DeployedVersionLookup,
		Dashboard:             serviceConfig.Dashboard,
		Status:                serviceConfig.Status,
	}

	api.writeJSON(w, serviceJSON, logFrom)
}

// httpOtherServiceDetails handles sending details about the global notify/webhooks, defaults and hard defaults.
//
// Method: GET
//
// Response:
//
//	JSON object containing the global details.
func (api *API) httpOtherServiceDetails(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpOtherServiceDetails", Secondary: getIP(r)}

	// Convert to JSON type that swaps slices for lists.
	api.writeJSON(
		w,
		apitype.Config{
			HardDefaults: convertAndCensorDefaults(&api.Config.HardDefaults),
			Defaults:     convertAndCensorDefaults(&api.Config.Defaults),
			Notify:       convertAndCensorNotifiersDefaults(api.Config.Notify),
			WebHook:      convertAndCensorWebHooksDefaults(api.Config.WebHook),
		},
		logFrom,
	)
}

// httpTemplateParse parses a template with provided or default parameters.
//
// Method: GET
//
// Query Parameters:
//
//	service_id: The ID of the Service to get the default parameters from.
//	template: The template to parse.
//	params: Optional JSON object containing the parameters to override the defaults.
//		id: string
//		name: string
//		url: string
//		icon: string
//		icon_link_to: string
//		web_url: string
//		approved_version: string
//		deployed_version: string
//		latest_version: string
//
// Response:
//
//	JSON object containing the parsed template.
func (api *API) httpTemplateParse(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpTemplateParse", Secondary: getIP(r)}

	// Extract query parameters.
	serviceID, ok := requireQueryParam(w, r, "service_id")
	if !ok {
		return
	}
	template, ok := requireQueryParam(w, r, "template")
	if !ok {
		return
	}
	params := r.URL.Query().Get("params")

	fullParams := serviceinfo.ServiceInfo{}

	// Fetch default parameters from the service configuration.
	api.Config.OrderMu.RLock()
	defer api.Config.OrderMu.RUnlock()
	serviceConfig, exists := api.Config.Service[serviceID]
	// Decode default parameters.
	if exists {
		fullParams = serviceConfig.Status.GetServiceInfo()
	}

	if e := util.CheckTemplate(template); !e {
		failRequest(
			&w,
			errors.New("failed to parse template"),
			http.StatusBadRequest,
		)
		return
	}

	// Override with query parameters.
	if params != "" {
		if err := decode.Unmarshal("json", []byte(params), &fullParams); err != nil {
			failRequest(
				&w,
				fmt.Errorf("invalid 'params' query parameter format - %w", err),
				http.StatusBadRequest,
			)
			return
		}
	}

	// Expand any environment variables in the template.
	template = util.EvalEnvVars(template)

	// Decode the template with the parameters.
	parsed := util.TemplateString(template, fullParams)

	// Respond with the parsed template.
	api.writeJSON(w, map[string]string{"parsed": parsed}, logFrom)
}

// httpServiceEdit handles creating/editing a Service.
//
// Method: PUT
//
// Query Parameters:
//
//	service_id: The ID of the Service to edit (empty for a new service).
//
// Body:
//
//	JSON object containing the service details.
//
// Response:
//
//	On success: HTTP 200 OK
//	On error: HTTP 400 Bad Request with an error message.
func (api *API) httpServiceEdit(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpServiceEdit", Secondary: getIP(r)}
	// Service to modify (empty for create new).
	serviceID := r.URL.Query().Get("service_id")
	reqType := "create"
	if serviceID != "" {
		reqType = "edit"
	}

	// EDIT: wait out any in-flight operations on this service (a refresh, another
	// edit, or a delete) rather than failing fast, so a background refresh can't
	// bounce a user's save. If the service was deleted or renamed while we waited,
	// return 404.
	if serviceID != "" {
		op := api.acquireServiceOp(serviceID)
		defer api.releaseServiceOp(serviceID, op)
		op.mu.Lock()
		defer op.mu.Unlock()
	}

	api.Config.OrderMu.RLock()

	var oldServiceSummary *apitype.ServiceSummary
	// EDIT the existing service.
	if serviceID != "" {
		if api.Config.Service[serviceID] == nil {
			api.Config.OrderMu.RUnlock()
			failRequest(
				&w,
				fmt.Errorf("edit %q failed, service could not be found", serviceID),
				http.StatusNotFound,
			)
			return
		}
		oldServiceSummary = api.Config.Service[serviceID].Summary()
	}

	// Payload.
	payload := http.MaxBytesReader(w, r.Body, 1024_00)

	svcDefaults, notifyDefaults, webhookDefaults := api.Config.GetDefaults()

	// Create the new/modified service.
	targetServicePtr := api.Config.Service[serviceID]
	newService, err := service.FromPayload(
		targetServicePtr, // nil if creating new.
		&payload,
		svcDefaults, notifyDefaults, webhookDefaults,
		logFrom,
	)
	api.Config.OrderMu.RUnlock()
	if err != nil {
		err = fmt.Errorf(
			`%s %q failed: %w`,
			reqType, serviceID, err,
		)
		logx.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusBadRequest)
		return
	}

	// CREATE/EDIT: service with this ID/Name already exists.
	if (serviceID == "" && api.Config.Service[newService.ID] != nil) ||
		api.Config.ServiceWithNameExists(newService.Name, serviceID) {
		failRequest(
			&w,
			fmt.Errorf("create %q failed, service with this name already exists", newService.ID),
			http.StatusBadRequest,
		)
		return
	}

	// CREATE: new ID is known, reject a concurrent create of the same ID.
	if serviceID == "" {
		op := api.acquireServiceOp(newService.ID)
		defer api.releaseServiceOp(newService.ID, op)
		if !op.mu.TryLock() {
			failRequest(
				&w,
				fmt.Errorf(
					"create %q failed, another operation is in progress for this service",
					newService.ID),
				http.StatusConflict,
			)
			return
		}
		defer op.mu.Unlock()
	}

	// Ensure LatestVersion and DeployedVersion (if set) can fetch.
	if err := newService.CheckFetches(); err != nil {
		err = fmt.Errorf(
			`%s %q failed: %w`,
			reqType, util.FirstNonDefault(serviceID, newService.ID),
			err,
		)
		logx.Error(err, logFrom, true)

		failRequest(&w, err, http.StatusBadRequest)
		return
	}

	// DeployedVersion is LatestVersion if there is no DeployedVersionLookup.
	if newService.DeployedVersionLookup == nil {
		newService.Status.SetDeployedVersion(
			newService.Status.LatestVersion(),
			newService.Status.LatestVersionTimestamp(),
			false,
		)
	}

	// Add the new service to the config.
	if err := api.Config.AddService(serviceID, newService); err != nil {
		err = fmt.Errorf(
			`%s %q failed: %w`,
			reqType, util.FirstNonDefault(serviceID, newService.ID),
			err,
		)
		logx.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusBadRequest)
		return
	}

	newServiceSummary := newService.Summary()
	// Announce the edit.
	api.announceEdit(oldServiceSummary, newServiceSummary)

	msg := "created"
	if serviceID != "" {
		msg = "edited"
	}
	api.writeJSON(
		w,
		apitype.Response{
			Message: fmt.Sprintf(
				"%s service %q",
				msg, util.ValueOr(serviceID, newService.ID),
			),
		},
		logFrom,
	)
}

// httpServiceDelete handles deleting a Service.
//
// Method: DELETE
//
// Query Parameters:
//
//	service_id: The ID of the Service to delete.
//
// Response:
//
//	On success: HTTP 200 OK
//	On error: HTTP 400 Bad Request with an error message.
func (api *API) httpServiceDelete(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpServiceDelete", Secondary: getIP(r)}
	// Service to delete.
	serviceID, ok := requireQueryParam(w, r, "service_id")
	if !ok {
		return
	}

	// Delete waits out any in-flight operations on this service, then takes the lock exclusively.
	op := api.acquireServiceOp(serviceID)
	defer api.releaseServiceOp(serviceID, op)
	op.mu.Lock()
	defer op.mu.Unlock()

	// If the service no longer exists (e.g. an edit we waited out renamed it), then 404.
	api.Config.OrderMu.RLock()
	exists := api.Config.Service[serviceID] != nil
	api.Config.OrderMu.RUnlock()
	if !exists {
		failRequest(
			&w,
			fmt.Errorf("delete %q failed, service not found", serviceID),
			http.StatusNotFound,
		)
		return
	}

	api.Config.DeleteService(serviceID)

	// Announce deletion.
	api.announceDelete(serviceID)

	api.writeJSON(
		w,
		apitype.Response{
			Message: fmt.Sprintf("deleted service %q", serviceID),
		},
		logFrom,
	)
}

// httpNotifyTest handles testing a Notify.
//
// Method: POST
//
// Body:
//
//	service_id_previous?: string (the service ID before the current changes)
//	service_id: string
//	service_name: string
//	name_previous?: string (the name of the notifier before the current changes)
//	name?: string (required if name_previous not set)
//	type?: string
//	options?: map[string]string
//	url_fields?: map[string]string
//	params?: map[string]string
//	service_url?: string
//	web_url?: string
func (api *API) httpNotifyTest(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpNotifyTest", Secondary: getIP(r)}

	// Read the payload.
	payload := http.MaxBytesReader(w, r.Body, 1024_00)
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(payload); err != nil {
		logx.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusBadRequest)
		return
	}
	if buf.Len() == 0 {
		err := errors.New("body required")
		logx.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusBadRequest)
		return
	}
	// Decode it.
	var parsedPayload shoutrrr.TestPayload
	if err := decode.Unmarshal("json", buf.Bytes(), &parsedPayload); err != nil {
		err = fmt.Errorf("failed to unmarshal payload: %w", err)
		logx.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusBadRequest)
		return
	}

	// Get the Notify.
	var serviceNotify *shoutrrr.Shoutrrr
	var serviceStatus *status.Status
	// From the ServiceIDPrevious.
	if parsedPayload.ServiceIDPrevious != "" {
		api.Config.OrderMu.RLock()
		defer api.Config.OrderMu.RUnlock()
		// Check whether service exists.
		svc := api.Config.Service[parsedPayload.ServiceIDPrevious]
		if svc != nil {
			// Check whether notifier exists.
			if svc.Notify != nil {
				serviceNotify = svc.Notify[parsedPayload.NamePrevious]
			}
			serviceStatus = svc.Status.Copy(false)
		}
	}

	if serviceStatus == nil {
		dash := dashboard.Options{
			OptionsBase: dashboard.OptionsBase{
				WebURL: parsedPayload.WebURL,
			},
		}
		serviceStatus = status.New(
			nil, nil, nil,
			"",
			"", "",
			"", "",
			"",
			&dash,
		)
	}
	serviceStatus.Init(
		0, 1, 0,
		status.ServiceInfo{
			ID:         parsedPayload.ServiceID,
			Name:       parsedPayload.ServiceName,
			ServiceURL: "service_url",
		},
		serviceStatus.Dashboard,
	)

	// Apply any overrides.
	testNotify, err := shoutrrr.FromPayload(
		parsedPayload,
		serviceNotify, serviceStatus,
		shoutrrr.Config{
			Root:         api.Config.Notify,
			Defaults:     api.Config.Defaults.Notify,
			HardDefaults: api.Config.HardDefaults.Notify,
		},
	)
	if err != nil {
		logx.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusBadRequest)
		return
	}

	// Send the message.
	err = testNotify.TestSend(parsedPayload.ServiceURL)
	if err != nil {
		logx.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusBadRequest)
		return
	}

	api.writeJSON(
		w,
		apitype.Response{
			Message: "message sent",
		},
		logFrom,
	)
}
