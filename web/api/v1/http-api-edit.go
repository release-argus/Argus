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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// httpLatestVersionRefreshUncreated will create the latest version lookup type and query it.
//
// # GET
//
// Path Parameters:
//
//	"semantic_versioning": Optional boolean parameter to override semantic versioning defaults.
//	"overrides": Required parameter to provide parameters for the version lookup.
//
// Response:
//
//	On success: JSON object containing the refreshed version and the current UTC datetime.
//	On error: HTTP 400 Bad Request with an error message.
func (api *API) httpLatestVersionRefreshUncreated(w http.ResponseWriter, r *http.Request) {
	logFrom := logutil.LogFrom{Primary: "httpVersionRefreshUncreated_Latest", Secondary: getIP(r)}

	queryParams := r.URL.Query()
	overrides := util.DereferenceOrDefault(getParam(&queryParams, "overrides"))

	// Verify overrides are provided.
	if overrides == "" {
		err := errors.New("overrides: <required>")
		logutil.Log.Error(err, logFrom, true)
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}

	// Status
	svcStatus := status.Status{}
	svcStatus.Init(
		0, 0, 0,
		&logFrom.Primary, nil,
		nil)

	// Options
	options := opt.New(
		nil, "",
		util.StringToBoolPtr(util.PtrValueOrValue(
			getParam(&queryParams, "semantic_versioning"), "",
		)),
		api.Config.Defaults.Service.LatestVersion.Options,
		api.Config.HardDefaults.Service.LatestVersion.Options)

	// Extract the desired lookup type.
	lookupType, err := extractLookupType(overrides, logFrom)
	if err != nil {
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create the LatestVersionLookup.
	lv, err := latestver.New(
		lookupType,
		"json", overrides,
		options,
		&svcStatus,
		&api.Config.Defaults.Service.LatestVersion, &api.Config.HardDefaults.Service.LatestVersion)
	if err == nil {
		err = lv.CheckValues("")
	} else {
		err = errors.New(strings.ReplaceAll(err.Error(), "latestver.Lookup", "latest_version"))
	}
	// Error creating/validating the LatestVersionLookup.
	if err != nil {
		logutil.Log.Error(err, logFrom, true)
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}

	// Query the latest version lookup.
	version, _, err := latestver.Refresh(
		lv,
		nil, nil)
	if err != nil {
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return the version found.
	api.writeJSON(w, apitype.RefreshAPI{
		Version: version,
		Date:    time.Now().UTC(),
	}, logFrom)
}

// httpDeployedVersionRefreshUncreated will create the deployed version lookup type and query it.
//
// # GET
//
// Path Parameters:
//
//	"semantic_versioning": Optional boolean parameter to override semantic versioning defaults.
//	"overrides": Required parameter to provide parameters for the version lookup.
//
// Response:
//
//	On success: JSON object containing the refreshed version and the current UTC datetime.
//	On error: HTTP 400 Bad Request with an error message.
func (api *API) httpDeployedVersionRefreshUncreated(w http.ResponseWriter, r *http.Request) {
	logFrom := logutil.LogFrom{Primary: "httpVersionRefreshUncreated_Deployed", Secondary: getIP(r)}

	queryParams := r.URL.Query()
	overrides := util.DereferenceOrDefault(getParam(&queryParams, "overrides"))

	// Verify overrides are provided.
	if overrides == "" {
		err := errors.New("overrides: <required>")
		logutil.Log.Error(err, logFrom, true)
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}

	// Status
	svcStatus := status.Status{}
	svcStatus.Init(
		0, 0, 0,
		&logFrom.Primary, nil,
		nil)

	// Options
	options := opt.New(
		nil, "",
		util.StringToBoolPtr(util.PtrValueOrValue(
			getParam(&queryParams, "semantic_versioning"), "",
		)),
		api.Config.Defaults.Service.LatestVersion.Options,
		api.Config.HardDefaults.Service.LatestVersion.Options)

	// Extract the desired lookup type.
	lookupType, err := extractLookupType(overrides, logFrom)
	if err != nil {
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create the DeployedVersionLookup.
	dvl, err := deployedver.New(
		lookupType,
		"json", overrides,
		options,
		&svcStatus,
		&api.Config.Defaults.Service.DeployedVersionLookup, &api.Config.HardDefaults.Service.DeployedVersionLookup)
	if err == nil {
		err = dvl.CheckValues("")
	} else {
		err = errors.New(strings.ReplaceAll(err.Error(), "deployedver.Lookup", "deployed_version"))
	}
	// Error creating/validating the DeployedVersionLookup.
	if err != nil {
		logutil.Log.Error(err, logFrom, true)
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}

	// Query the DeployedVersionLookup.
	version, err := deployedver.Refresh(
		dvl,
		"", nil)
	if err != nil {
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return the version found.
	api.writeJSON(w, apitype.RefreshAPI{
		Version: version,
		Date:    time.Now().UTC(),
	}, logFrom)
}

// httpLatestVersionRefresh refreshes the latest version of the target service.
//
// # GET
//
// Path Parameters:
//
//	service_id: The ID of the Service to refresh the LatestVersion of.
//
// Query Parameters:
//
//	"overrides": Optional parameter to provide parameters for the version lookup.
//	"semantic_versioning": Optional boolean parameter to override semantic versioning defaults.
//
// Response:
//
//	On success: JSON object containing the refreshed version and the current UTC datetime.
//	On error: HTTP 400 Bad Request with an error message.
func (api *API) httpLatestVersionRefresh(w http.ResponseWriter, r *http.Request) {
	logFrom := logutil.LogFrom{Primary: "httpVersionRefresh_Latest", Secondary: getIP(r)}
	// Service to refresh.
	targetService, _ := url.QueryUnescape(mux.Vars(r)["service_id"])

	queryParams := r.URL.Query()

	// Check if service exists.
	api.Config.OrderMutex.RLock()
	defer api.Config.OrderMutex.RUnlock()
	if api.Config.Service[targetService] == nil {
		err := fmt.Sprintf("service %q not found", targetService)
		logutil.Log.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusNotFound)
		return
	}

	// Parameters
	var (
		overrides          = getParam(&queryParams, "overrides")
		semanticVersioning = getParam(&queryParams, "semantic_versioning")
	)

	// Query the LatestVersion lookup.
	version, announce, err := latestver.Refresh(
		api.Config.Service[targetService].LatestVersion,
		overrides,
		semanticVersioning)
	if announce {
		api.Config.Service[targetService].HandleUpdateActions(true)
	}
	if err != nil {
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return the version found.
	api.writeJSON(w, apitype.RefreshAPI{
		Version: version,
		Date:    time.Now().UTC(),
	}, logFrom)
}

// httpDeployedVersionRefresh refreshes the latest/deployed version of the target service.
//
// # GET
//
// Path Parameters:
//
//	service_id: The ID of the Service to refresh the DeployedVersion of.
//
// Query Parameters:
//
//	"overrides": Optional parameter to provide parameters for the version lookup.
//	"semantic_versioning": Optional boolean parameter to override semantic versioning defaults.
//
// Response:
//
//	On success: JSON object containing the refreshed version and the current UTC datetime.
//	On error: HTTP 400 Bad Request with an error message.
func (api *API) httpDeployedVersionRefresh(w http.ResponseWriter, r *http.Request) {
	logFrom := logutil.LogFrom{Primary: "httpVersionRefresh_Deployed", Secondary: getIP(r)}
	// Service to refresh.
	targetService, _ := url.QueryUnescape(mux.Vars(r)["service_id"])

	queryParams := r.URL.Query()

	// Check if service exists.
	api.Config.OrderMutex.RLock()
	defer api.Config.OrderMutex.RUnlock()
	if api.Config.Service[targetService] == nil {
		err := fmt.Sprintf("service %q not found", targetService)
		logutil.Log.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusNotFound)
		return
	}

	// Parameters.
	var (
		overrides          = util.DereferenceOrDefault(getParam(&queryParams, "overrides"))
		semanticVersioning = getParam(&queryParams, "semantic_versioning")
	)

	// Existing DeployedVersionLookup?
	dvl := api.Config.Service[targetService].DeployedVersionLookup
	// Must create the DeployedVersionLookup if it doesn't exist.
	if dvl == nil {
		// Extract the desired lookup type.
		lookupType, err := extractLookupType(overrides, logFrom)
		if err != nil {
			failRequest(&w, err.Error(), http.StatusBadRequest)
			return
		}

		// Status
		svcStatus := status.Status{}
		svcStatus.Init(
			0, 0, 0,
			&logFrom.Primary, nil,
			nil)

		dvl, _ = deployedver.New(
			lookupType,
			"json", "{}",
			&api.Config.Service[targetService].Options,
			&svcStatus,
			&api.Config.Service[targetService].Defaults.DeployedVersionLookup,
			&api.Config.Service[targetService].HardDefaults.DeployedVersionLookup) //nolint:errcheck // empty JSON.
	}

	// Query the DeployedVersionLookup.
	version, err := deployedver.Refresh(
		dvl,
		overrides,
		semanticVersioning)
	if err != nil {
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return the version found.
	api.writeJSON(w, apitype.RefreshAPI{
		Version: version,
		Date:    time.Now().UTC(),
	}, logFrom)
}

// extractLookupType extracts the desired `type` from the provided JSON.
func extractLookupType(overrides string, logFrom logutil.LogFrom) (string, error) {
	// Extract the desired lookup type.
	var temp struct {
		Type string `yaml:"type" json:"type"`
	}

	if err := util.UnmarshalConfig(
		"json", overrides,
		&temp); err != nil {
		err = fmt.Errorf("invalid JSON: %w", err)
		logutil.Log.Error(err, logFrom, true)
		return "", err
	}

	return temp.Type, nil
}

// httpServiceDetail handles sending details about a Service.
//
// # GET
//
// Path Parameters:
//
//	service_id: The ID of the Service to get details for.
//
// Response:
//
//	JSON object containing the service details.
func (api *API) httpServiceDetail(w http.ResponseWriter, r *http.Request) {
	logFrom := logutil.LogFrom{Primary: "httpServiceDetail", Secondary: getIP(r)}
	// Service to get details of.
	targetService, _ := url.QueryUnescape(mux.Vars(r)["service_id"])

	// Find the Service.
	api.Config.OrderMutex.RLock()
	svc := api.Config.Service[targetService]
	// Convert to API Type, censoring secrets.
	serviceConfig := convertAndCensorService(svc)
	api.Config.OrderMutex.RUnlock()

	if svc == nil {
		err := fmt.Sprintf("service %q not found", targetService)
		logutil.Log.Error(err, logFrom, true)
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
// # GET
//
// Response:
//
//	JSON object containing the global details.
func (api *API) httpOtherServiceDetails(w http.ResponseWriter, r *http.Request) {
	logFrom := logutil.LogFrom{Primary: "httpOtherServiceDetails", Secondary: getIP(r)}

	// Convert to JSON type that swaps slices for lists.
	api.writeJSON(w,
		apitype.Config{
			HardDefaults: convertAndCensorDefaults(&api.Config.HardDefaults),
			Defaults:     convertAndCensorDefaults(&api.Config.Defaults),
			Notify:       convertAndCensorNotifySliceDefaults(&api.Config.Notify),
			WebHook:      convertAndCensorWebHookSliceDefaults(&api.Config.WebHook),
		},
		logFrom)
}

// httpServiceEdit handles creating/editing a Service.
//
// # PUT
//
// Path Parameters:
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
	logFrom := logutil.LogFrom{Primary: "httpServiceEdit", Secondary: getIP(r)}
	api.Config.OrderMutex.RLock()
	defer api.Config.OrderMutex.RUnlock()

	// Service to modify (empty for create new).
	targetService, _ := url.QueryUnescape(mux.Vars(r)["service_id"])
	reqType := "create"
	if targetService != "" {
		reqType = "edit"
	}

	var oldServiceSummary *apitype.ServiceSummary
	// EDIT the existing service.
	if targetService != "" {
		if api.Config.Service[targetService] == nil {
			failRequest(&w,
				fmt.Sprintf("edit %q failed, service could not be found",
					targetService),
				http.StatusNotFound)
			return
		}
		oldServiceSummary = api.Config.Service[targetService].Summary()
	}

	// Payload.
	payload := http.MaxBytesReader(w, r.Body, 1024_00)

	// Create the new/modified service.
	targetServicePtr := api.Config.Service[targetService]
	newService, err := service.FromPayload(
		targetServicePtr, // nil if creating new.
		&payload,
		&api.Config.Defaults.Service,
		&api.Config.HardDefaults.Service,
		&api.Config.Notify,
		&api.Config.Defaults.Notify,
		&api.Config.HardDefaults.Notify,
		&api.Config.WebHook,
		&api.Config.Defaults.WebHook,
		&api.Config.HardDefaults.WebHook,
		logFrom)
	if err != nil {
		logutil.Log.Error(err, logFrom, true)
		failRequest(&w,
			fmt.Sprintf(`%s %q failed (invalid json)\n%s`,
				reqType, targetService, err),
			http.StatusBadRequest)
		return
	}

	// CREATE a new service, but one with this ID already exists.
	if (targetService == "" && api.Config.Service[newService.ID] != nil) ||
		// CREATE/EDIT, but a service with this name already exists.
		api.Config.ServiceWithNameExists(newService.Name, targetService) {
		failRequest(&w,
			fmt.Sprintf("create %q failed, service with this name already exists",
				newService.ID),
			http.StatusBadRequest)
		return
	}

	// Check the values.
	if err := newService.CheckValues(""); err != nil {
		logutil.Log.Error(err, logFrom, true)

		failRequest(&w,
			fmt.Sprintf(`%s %q failed (invalid values)\n%s`,
				reqType, util.FirstNonDefault(targetService, newService.ID), err),
			http.StatusBadRequest)
		return
	}

	// Ensure LatestVersion and DeployedVersion (if set) can fetch.
	if err := newService.CheckFetches(); err != nil {
		logutil.Log.Error(err, logFrom, true)

		failRequest(&w,
			fmt.Sprintf(`%s %q failed (fetches failed)\n%s`,
				reqType, util.FirstNonDefault(targetService, newService.ID), err),
			http.StatusBadRequest)
		return
	}

	// DeployedVersion is LatestVersion if there is no DeployedVersionLookup.
	if newService.DeployedVersionLookup == nil {
		newService.Status.SetDeployedVersion(
			newService.Status.LatestVersion(), newService.Status.LatestVersionTimestamp(),
			false)
	}

	// Add the new service to the config.
	api.Config.OrderMutex.RUnlock() // Locked above.
	//#nosec G104 -- Fail for duplicate service name handled above.
	//nolint:errcheck // ^
	api.Config.AddService(targetService, newService)
	api.Config.OrderMutex.RLock() // Lock again for the defer.

	newServiceSummary := newService.Summary()
	// Announce the edit.
	api.announceEdit(oldServiceSummary, newServiceSummary)

	msg := "created"
	if targetService != "" {
		msg = "edited"
	}
	api.writeJSON(w, apitype.Response{
		Message: fmt.Sprintf(
			"%s service %q",
			msg, util.ValueOrValue(targetService, newService.ID))},
		logFrom)
}

// httpServiceDelete handles deleting a Service.
//
// # DELETE
//
// Path Parameters:
//
//	service_id: The ID of the Service to delete.
//
// Response:
//
//	On success: HTTP 200 OK
//	On error: HTTP 400 Bad Request with an error message.
func (api *API) httpServiceDelete(w http.ResponseWriter, r *http.Request) {
	logFrom := logutil.LogFrom{Primary: "httpServiceDelete", Secondary: getIP(r)}
	// Service to delete.
	targetService, _ := url.QueryUnescape(mux.Vars(r)["service_id"])

	// If service doesn't exist, return 404.
	if api.Config.Service[targetService] == nil {
		failRequest(&w,
			fmt.Sprintf("delete %q failed, service not found",
				targetService),
			http.StatusNotFound)
		return
	}

	api.Config.DeleteService(targetService)

	// Announce deletion.
	api.announceDelete(targetService)

	api.writeJSON(w, apitype.Response{
		Message: fmt.Sprintf(
			"deleted service %q", targetService)},
		logFrom)
}

// httpNotifyTest handles testing a Notify.
//
// # POST
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
	logFrom := logutil.LogFrom{Primary: "httpNotifyTest", Secondary: getIP(r)}

	// Read payload.
	payload := http.MaxBytesReader(w, r.Body, 1024_00)
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(payload); err != nil {
		logutil.Log.Error(err, logFrom, true)
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}
	// Parse it.
	var parsedPayload shoutrrr.TestPayload
	err := json.Unmarshal(buf.Bytes(), &parsedPayload)
	if err != nil {
		logutil.Log.Error(err, logFrom, true)
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the Notify.
	var serviceNotify *shoutrrr.Shoutrrr
	var latestVersion string
	// From the ServiceIDPrevious.
	if parsedPayload.ServiceIDPrevious != "" {
		api.Config.OrderMutex.RLock()
		defer api.Config.OrderMutex.RUnlock()
		// Check whether service exists.
		if api.Config.Service[parsedPayload.ServiceIDPrevious] != nil {
			// Check whether notifier exists.
			if api.Config.Service[parsedPayload.ServiceIDPrevious].Notify != nil {
				serviceNotify = api.Config.Service[parsedPayload.ServiceIDPrevious].Notify[parsedPayload.NamePrevious]
			}
			latestVersion = api.Config.Service[parsedPayload.ServiceIDPrevious].Status.LatestVersion()
		}
	}

	// Apply any overrides.
	testNotify, serviceURL, err := shoutrrr.FromPayload(
		parsedPayload,
		serviceNotify,
		api.Config.Notify, api.Config.Defaults.Notify, api.Config.HardDefaults.Notify)
	if err != nil {
		logutil.Log.Error(err, logFrom, true)
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}

	// Give the LatestVersion as the message may use it.
	testNotify.ServiceStatus.SetLatestVersion(latestVersion, "", false)

	// Send the message.
	err = testNotify.TestSend(serviceURL)
	if err != nil {
		logutil.Log.Error(err, logFrom, true)
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}

	api.writeJSON(w, apitype.Response{
		Message: "message sent"},
		logFrom)
}
