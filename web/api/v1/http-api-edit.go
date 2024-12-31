// Copyright [2024] [Argus]
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
	setCommonHeaders(w)

	logFrom := util.LogFrom{Secondary: getIP(r)}
	logFrom.Primary = "httpVersionRefreshUncreated_Latest"
	jLog.Verbose("-", logFrom, true)

	queryParams := r.URL.Query()
	overrides := util.DereferenceOrDefault(getParam(&queryParams, "overrides"))

	// Verify overrides are provided.
	if overrides == "" {
		err := errors.New("overrides: <required>")
		jLog.Error(err, logFrom, true)
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}

	// Status
	svcStatus := status.Status{}
	svcStatus.Init(
		0, 0, 0,
		&logFrom.Primary,
		nil)

	// Options
	options := opt.New(
		nil, "",
		util.StringToBoolPtr(util.PtrValueOrValue(
			getParam(&queryParams, "semantic_versioning"), "",
		)),
		api.Config.Defaults.Service.LatestVersion.Options,
		api.Config.HardDefaults.Service.LatestVersion.Options)

	// Extract the desired lookup type
	var temp struct {
		Type string `yaml:"type" json:"type"`
	}
	if err := util.UnmarshalConfig(
		"json", overrides,
		&temp); err != nil {
		err = fmt.Errorf("invalid JSON: %w", err)
		jLog.Error(err, logFrom, true)
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create the LatestVersionLookup
	lv, err := latestver.New(
		temp.Type,
		"json", util.DereferenceOrDefault(
			getParam(&queryParams, "overrides")),
		options,
		&svcStatus,
		&api.Config.Defaults.Service.LatestVersion, &api.Config.HardDefaults.Service.LatestVersion)
	if err == nil {
		err = lv.CheckValues("")
	} else {
		err = errors.New(strings.ReplaceAll(err.Error(), "latestver.Lookup", "latest_version"))
	}
	// Error creating/validating the LatestVersionLookup
	if err != nil {
		jLog.Error(err, logFrom, true)
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}

	// Query the latest version lookup
	version, _, err := latestver.Refresh(
		lv,
		nil, nil)
	if err != nil {
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return the version found.
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(apitype.RefreshAPI{
		Version: version,
		Date:    time.Now().UTC(),
	})
	jLog.Error(err, logFrom, err != nil)
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
	setCommonHeaders(w)

	logFrom := util.LogFrom{Secondary: getIP(r)}
	logFrom.Primary = "httpVersionRefreshUncreated_Deployed"
	jLog.Verbose("-", logFrom, true)

	queryParams := r.URL.Query()
	overrides := util.DereferenceOrDefault(getParam(&queryParams, "overrides"))

	// Verify overrides are provided
	if overrides == "" {
		err := errors.New("overrides: <required>")
		jLog.Error(err, logFrom, true)
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}

	// Status
	svcStatus := status.Status{}
	svcStatus.Init(
		0, 0, 0,
		&logFrom.Primary,
		nil)

	// Options
	options := opt.New(
		nil, "",
		util.StringToBoolPtr(util.PtrValueOrValue(
			getParam(&queryParams, "semantic_versioning"), "",
		)),
		api.Config.Defaults.Service.LatestVersion.Options,
		api.Config.HardDefaults.Service.LatestVersion.Options)

	// Create the DeployedVersionLookup.
	dvl, err := deployedver.New(
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
		jLog.Error(err, logFrom, true)
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}

	// Query the DeployedVersionLookup.
	version, err := dvl.Refresh(
		&logFrom.Primary,
		nil, nil)
	if err != nil {
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return the version found.
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(apitype.RefreshAPI{
		Version: version,
		Date:    time.Now().UTC(),
	})
	jLog.Error(err, logFrom, err != nil)
}

// httpLatestVersionRefresh refreshes the latest version of the target service.
//
// # GET
//
// Path Parameters:
//
//	service_name: service name to refresh
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
	setCommonHeaders(w)

	// Service to refresh.
	targetService, _ := url.QueryUnescape(mux.Vars(r)["service_name"])

	logFromPrimary := "httpVersionRefresh_Latest"
	logFrom := util.LogFrom{Primary: logFromPrimary, Secondary: getIP(r)}
	jLog.Verbose(targetService, logFrom, true)

	queryParams := r.URL.Query()

	// Check if service exists.
	api.Config.OrderMutex.RLock()
	defer api.Config.OrderMutex.RUnlock()
	if api.Config.Service[targetService] == nil {
		err := fmt.Sprintf("service %q not found", targetService)
		jLog.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusNotFound)
		return
	}

	// Parameters,
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
	err = json.NewEncoder(w).Encode(apitype.RefreshAPI{
		Version: version,
		Date:    time.Now().UTC(),
	})
	jLog.Error(err, logFrom, err != nil)
}

// httpDeployedVersionRefresh refreshes the latest/deployed version of the target service.
//
// # GET
//
// Path Parameters:
//
//	service_name: service name to refresh
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
	setCommonHeaders(w)

	// Service to refresh.
	targetService, _ := url.QueryUnescape(mux.Vars(r)["service_name"])

	logFromPrimary := "httpVersionRefresh_Deployed"
	logFrom := util.LogFrom{Primary: logFromPrimary, Secondary: getIP(r)}
	jLog.Verbose(targetService, logFrom, true)

	queryParams := r.URL.Query()

	// Check if service exists,
	api.Config.OrderMutex.RLock()
	defer api.Config.OrderMutex.RUnlock()
	if api.Config.Service[targetService] == nil {
		err := fmt.Sprintf("service %q not found", targetService)
		jLog.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusNotFound)
		return
	}

	// Parameters.
	var (
		overrides          = getParam(&queryParams, "overrides")
		semanticVersioning = getParam(&queryParams, "semantic_versioning")
	)

	// Existing DeployedVersionLookup?
	dvl := api.Config.Service[targetService].DeployedVersionLookup
	// Must create the DeployedVersionLookup if it doesn't exist.
	if dvl == nil {
		// Status
		svcStatus := status.Status{}
		svcStatus.Init(
			0, 0, 0,
			&logFrom.Primary,
			nil)

		var err error
		dvl, err = deployedver.New(
			"json", util.DereferenceOrDefault(overrides),
			&api.Config.Service[targetService].Options,
			&svcStatus,
			&api.Config.Service[targetService].Defaults.DeployedVersionLookup, &api.Config.Service[targetService].HardDefaults.DeployedVersionLookup)
		if err == nil {
			err = dvl.CheckValues("")
		}
		if err != nil {
			jLog.Error(err, logFrom, true)
			failRequest(&w, err.Error(), http.StatusBadRequest)
			return
		}
		// nil the overrides so that we don't apply them again in the Refresh.
		overrides = nil
	}

	// Query the DeployedVersionLookup.
	version, err := dvl.Refresh(
		&targetService,
		overrides,
		semanticVersioning)
	if err != nil {
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return the version found.
	err = json.NewEncoder(w).Encode(apitype.RefreshAPI{
		Version: version,
		Date:    time.Now().UTC(),
	})
	jLog.Error(err, logFrom, err != nil)
}

// httpServiceDetail handles sending details about a Service.
//
// # GET
//
// Path Parameters:
//
//	service_name: service to get details for
//
// Response:
//
//	JSON object containing the service details.
func (api *API) httpServiceDetail(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w)

	// Service to get details from (empty for create new).
	targetService, _ := url.QueryUnescape(mux.Vars(r)["service_name"])

	logFrom := util.LogFrom{Primary: "httpServiceDetail", Secondary: getIP(r)}
	jLog.Verbose(targetService, logFrom, true)

	// Find the Service.
	api.Config.OrderMutex.RLock()
	svc := api.Config.Service[targetService]
	// Convert to API Type, censoring secrets.
	serviceConfig := convertAndCensorService(svc)
	api.Config.OrderMutex.RUnlock()

	if svc == nil {
		err := fmt.Sprintf("service %q not found", targetService)
		jLog.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusNotFound)
		return
	}

	// Convert to JSON type that swaps slices for lists.
	serviceJSON := apitype.ServiceEdit{
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

	err := json.NewEncoder(w).Encode(serviceJSON)
	jLog.Error(err, logFrom, err != nil)
}

// httpOtherServiceDetails handles sending details about the global notify/webhooks, defaults and hard defaults.
//
// # GET
//
// Response:
//
//	JSON object containing the global details.
func (api *API) httpOtherServiceDetails(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w)

	logFromPrimary := "httpOtherServiceDetails"
	logFrom := util.LogFrom{Primary: logFromPrimary, Secondary: getIP(r)}
	jLog.Verbose("-", logFrom, true)

	// Convert to JSON type that swaps slices for lists.
	err := json.NewEncoder(w).Encode(apitype.Config{
		HardDefaults: convertAndCensorDefaults(&api.Config.HardDefaults),
		Defaults:     convertAndCensorDefaults(&api.Config.Defaults),
		Notify:       convertAndCensorNotifySliceDefaults(&api.Config.Notify),
		WebHook:      convertAndCensorWebHookSliceDefaults(&api.Config.WebHook),
	})
	jLog.Error(err, logFrom, err != nil)
}

// httpServiceEdit handles creating/editing a Service.
//
// # PUT - create/replace
//
// Path Parameters:
//
//	service_name: service to edit (empty for new service)
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
	setCommonHeaders(w)

	api.Config.OrderMutex.RLock()
	defer api.Config.OrderMutex.RUnlock()

	// Service to modify (empty for create new).
	targetService, _ := url.QueryUnescape(mux.Vars(r)["service_name"])
	reqType := "create"
	if targetService != "" {
		reqType = "edit"
	}

	logFrom := util.LogFrom{Primary: "httpServiceEdit", Secondary: getIP(r)}
	jLog.Verbose(
		fmt.Sprintf("%s %s",
			reqType, targetService),
		logFrom, true)

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
		jLog.Error(err, logFrom, true)
		failRequest(&w,
			fmt.Sprintf(`%s %q failed (invalid json)\n%s`,
				reqType, targetService, err),
			http.StatusBadRequest)
		return
	}

	// CREATE a new service, but one with the same name already exists.
	if targetService == "" && api.Config.Service[newService.ID] != nil {
		failRequest(&w,
			fmt.Sprintf("create %q failed, service with this name already exists",
				newService.ID),
			http.StatusBadRequest)
		return
	}

	// Check the values.
	err = newService.CheckValues("")
	if err != nil {
		jLog.Error(err, logFrom, true)

		failRequest(&w,
			fmt.Sprintf(`%s %q failed (invalid values)\n%s`,
				reqType, util.FirstNonDefault(targetService, newService.ID), err),
			http.StatusBadRequest)
		return
	}

	// Ensure LatestVersion and DeployedVersion (if set) can fetch.
	err = newService.CheckFetches()
	if err != nil {
		jLog.Error(err, logFrom, true)

		failRequest(&w,
			fmt.Sprintf(`%s %q failed (fetches failed)\n%s`,
				reqType, util.FirstNonDefault(targetService, newService.ID), err),
			http.StatusBadRequest)
		return
	}

	// DeployedVersion is LatestVersion if there is no DeployedVersionLookup.
	if newService.DeployedVersionLookup == nil {
		newService.Status.SetDeployedVersion(newService.Status.LatestVersion(), newService.Status.LatestVersionTimestamp(), false)
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
}

// httpServiceDelete handles deleting a Service.
//
// # DELETE
//
// Path Parameters:
//
//	service_name: service to delete
//
// Response:
//
//	On success: HTTP 200 OK
//	On error: HTTP 400 Bad Request with an error message.
func (api *API) httpServiceDelete(w http.ResponseWriter, r *http.Request) {
	// Service to delete.
	targetService, _ := url.QueryUnescape(mux.Vars(r)["service_name"])

	logFromPrimary := "httpServiceDelete"
	logFrom := util.LogFrom{Primary: logFromPrimary, Secondary: getIP(r)}
	jLog.Verbose(targetService, logFrom, true)

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

	// Return 200.
	w.WriteHeader(http.StatusOK)
	//#nosec G104 -- Disregard.
	//nolint:errcheck // ^
	w.Write([]byte{})
}

// httpNotifyTest handles testing a Notify.
//
// # POST
//
// Body:
//
//	service_name_previous?: string (the service name before the current changes)
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
	setCommonHeaders(w)

	logFrom := util.LogFrom{Primary: "httpNotifyTest", Secondary: getIP(r)}
	jLog.Verbose("-", logFrom, true)

	// Payload.
	payload := http.MaxBytesReader(w, r.Body, 1024_00)
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(payload); err != nil {
		jLog.Error(err, logFrom, true)
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}
	var parsedPayload shoutrrr.TestPayload
	err := json.Unmarshal(buf.Bytes(), &parsedPayload)
	if err != nil {
		jLog.Error(err, logFrom, true)
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the Notify.
	var serviceNotify *shoutrrr.Shoutrrr
	var latestVersion string
	// From the ServiceNamePrevious.
	if parsedPayload.ServiceNamePrevious != "" {
		api.Config.OrderMutex.RLock()
		defer api.Config.OrderMutex.RUnlock()
		// Check whether service exists.
		if api.Config.Service[parsedPayload.ServiceNamePrevious] != nil {
			// Check whether notifier exists.
			if api.Config.Service[parsedPayload.ServiceNamePrevious].Notify != nil {
				serviceNotify = api.Config.Service[parsedPayload.ServiceNamePrevious].Notify[parsedPayload.NamePrevious]
			}
			latestVersion = api.Config.Service[parsedPayload.ServiceNamePrevious].Status.LatestVersion()
		}
	}
	// Apply any overrides.
	testNotify, serviceURL, err := shoutrrr.FromPayload(
		parsedPayload,
		serviceNotify,
		api.Config.Notify, api.Config.Defaults.Notify, api.Config.HardDefaults.Notify)
	if err != nil {
		jLog.Error(err, logFrom, true)
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}
	// Give the LatestVersion as the message may use it.
	testNotify.ServiceStatus.SetLatestVersion(latestVersion, "", false)

	// Send the message.
	err = testNotify.TestSend(serviceURL)
	if err != nil {
		jLog.Error(err, logFrom, true)
		failRequest(&w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return 200.
	w.WriteHeader(http.StatusOK)
	//#nosec G104 -- Disregard.
	//nolint:errcheck // ^
	w.Write([]byte(`{}`))
}
