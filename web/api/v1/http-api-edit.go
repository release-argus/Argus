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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	api_type "github.com/release-argus/Argus/web/api/types"
)

// httpVersionRefreshUncreated will create the latest/deployed version lookup type and query it.
//
// # GET
//
// params: service params to use
func (api *API) httpVersionRefreshUncreated(w http.ResponseWriter, r *http.Request) {
	logFromPrimary := "httpVersionRefreshUncreated_Latest"
	deployedVersionRefresh := strings.Contains(r.URL.String(), "/deployed_version/refresh?")
	if deployedVersionRefresh {
		logFromPrimary = "httpVersionRefreshUncreated_Deployed"
	}
	logFrom := &util.LogFrom{Primary: logFromPrimary, Secondary: getIP(r)}
	jLog.Verbose("-", logFrom, true)

	// Set headers
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "application/json")
	queryParams := r.URL.Query()

	status := svcstatus.Status{}
	status.Init(
		0, 0, 0,
		&logFromPrimary,
		nil)

	var (
		version string
		err     error
	)
	if deployedVersionRefresh {
		deployedVersionLookup := deployedver.New(
			nil, nil, nil, "",
			opt.New(
				nil, "", nil,
				&api.Config.Defaults.Service.Options,
				&api.Config.HardDefaults.Service.Options),
			"", nil, &status, "",
			&api.Config.Defaults.Service.DeployedVersionLookup,
			&api.Config.HardDefaults.Service.DeployedVersionLookup)
		// Deployed Version
		version, _, err = deployedVersionLookup.Refresh(
			getParam(&queryParams, "allow_invalid_certs"),
			getParam(&queryParams, "basic_auth"),
			getParam(&queryParams, "headers"),
			getParam(&queryParams, "json"),
			getParam(&queryParams, "regex"),
			getParam(&queryParams, "regex_template"),
			getParam(&queryParams, "semantic_versioning"),
			getParam(&queryParams, "url"))
	} else {
		latestVersion := latestver.Lookup{
			Options: &opt.Options{
				Defaults:     &api.Config.Defaults.Service.Options,
				HardDefaults: &api.Config.HardDefaults.Service.Options},
			Status:       &status,
			Defaults:     &api.Config.Defaults.Service.LatestVersion,
			HardDefaults: &api.Config.HardDefaults.Service.LatestVersion}
		// Latest Version
		version, _, err = latestVersion.Refresh(
			getParam(&queryParams, "access_token"),
			getParam(&queryParams, "allow_invalid_certs"),
			getParam(&queryParams, "require"),
			getParam(&queryParams, "semantic_versioning"),
			getParam(&queryParams, "type"),
			getParam(&queryParams, "url"),
			getParam(&queryParams, "url_commands"),
			getParam(&queryParams, "use_prerelease"))
	}

	statusCode := http.StatusOK
	if err != nil {
		statusCode = http.StatusBadRequest
	}
	w.WriteHeader(statusCode)
	err = json.NewEncoder(w).Encode(api_type.RefreshAPI{
		Version: version,
		Error:   util.ErrorToString(err),
		Date:    time.Now().UTC(),
	})
	jLog.Error(err, logFrom, err != nil)
}

// httpVersionRefresh refreshes the latest/deployed version of the target service.
//
// # GET
//
// service_name: service name to refresh
//
// ...params?: service params to override
func (api *API) httpVersionRefresh(w http.ResponseWriter, r *http.Request) {
	// service to refresh
	targetService, _ := url.QueryUnescape(mux.Vars(r)["service_name"])

	logFromPrimary := "httpVersionRefresh_Latest"
	deployedVersionRefresh := strings.Contains(r.URL.String(), "/deployed_version/refresh/")
	if deployedVersionRefresh {
		logFromPrimary = "httpVersionRefresh_Deployed"
	}
	logFrom := &util.LogFrom{Primary: logFromPrimary, Secondary: getIP(r)}
	jLog.Verbose(targetService, logFrom, true)

	// Set headers
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "application/json")

	queryParams := r.URL.Query()

	// Check if service exists
	api.Config.OrderMutex.RLock()
	defer api.Config.OrderMutex.RUnlock()
	if api.Config.Service[targetService] == nil {
		err := fmt.Sprintf("service %q not found", targetService)
		jLog.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusNotFound)
		return
	}

	// Refresh the latest/deployed version lookup type
	var (
		version  string
		err      error
		announce bool
	)
	if deployedVersionRefresh {
		if api.Config.Service[targetService].DeployedVersionLookup == nil {
			api.Config.Service[targetService].DeployedVersionLookup = &deployedver.Lookup{
				Options: &api.Config.Service[targetService].Options,
				Status: &svcstatus.Status{
					ServiceID: &targetService},
				Defaults:     &api.Config.Defaults.Service.DeployedVersionLookup,
				HardDefaults: &api.Config.HardDefaults.Service.DeployedVersionLookup,
			}
		}
		// Deployed Version
		version, announce, err = api.Config.Service[targetService].DeployedVersionLookup.Refresh(
			getParam(&queryParams, "allow_invalid_certs"),
			getParam(&queryParams, "basic_auth"),
			getParam(&queryParams, "headers"),
			getParam(&queryParams, "json"),
			getParam(&queryParams, "regex"),
			getParam(&queryParams, "regex_template"),
			getParam(&queryParams, "semantic_versioning"),
			getParam(&queryParams, "url"),
		)

		if announce {
			api.Config.Service[targetService].DeployedVersionLookup.HandleNewVersion(version, true)
		}
	} else {
		// Latest Version
		version, announce, err = api.Config.Service[targetService].LatestVersion.Refresh(
			getParam(&queryParams, "access_token"),
			getParam(&queryParams, "allow_invalid_certs"),
			getParam(&queryParams, "require"),
			getParam(&queryParams, "semantic_versioning"),
			getParam(&queryParams, "type"),
			getParam(&queryParams, "url"),
			getParam(&queryParams, "url_commands"),
			getParam(&queryParams, "use_prerelease"),
		)

		if announce {
			api.Config.Service[targetService].HandleUpdateActions(true)
		}
	}

	err = json.NewEncoder(w).Encode(api_type.RefreshAPI{
		Version: version,
		Error:   util.ErrorToString(err),
		Date:    time.Now().UTC(),
	})
	jLog.Error(err, logFrom, err != nil)
}

// httpServiceDetail handles sending details about a Service
//
// # GET
//
// service_name: service to get details for
func (api *API) httpServiceDetail(w http.ResponseWriter, r *http.Request) {
	// service to get details from (empty for create new)
	targetService, _ := url.QueryUnescape(mux.Vars(r)["service_name"])

	logFrom := &util.LogFrom{Primary: "httpServiceDetail", Secondary: getIP(r)}
	jLog.Verbose(targetService, logFrom, true)

	// Set Headers
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "application/json")

	// Find the Service
	api.Config.OrderMutex.RLock()
	svc := api.Config.Service[targetService]
	// Convert to API Type, censoring secrets
	serviceConfig := convertAndCensorService(svc)
	api.Config.OrderMutex.RUnlock()

	if svc == nil {
		err := fmt.Sprintf("service %q not found", targetService)
		jLog.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusNotFound)
		return
	}

	// Convert to JSON type that swaps slices for lists
	serviceJSON := api_type.ServiceEdit{
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

// httpOtherServiceDetails handles sending details about the global notify/webhook's, defaults and hard defaults.
//
// # GET
func (api *API) httpOtherServiceDetails(w http.ResponseWriter, r *http.Request) {
	logFromPrimary := "httpOtherServiceDetails"
	logFrom := &util.LogFrom{Primary: logFromPrimary, Secondary: getIP(r)}
	jLog.Verbose("-", logFrom, true)

	// Set headers
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "application/json")

	// Convert to JSON type that swaps slices for lists
	err := json.NewEncoder(w).Encode(api_type.Config{
		HardDefaults: convertAndCensorDefaults(&api.Config.HardDefaults),
		Defaults:     convertAndCensorDefaults(&api.Config.Defaults),
		Notify:       convertAndCensorNotifySliceDefaults(&api.Config.Notify),
		WebHook:      convertAndCensorWebHookSliceDefaults(&api.Config.WebHook),
	})
	jLog.Error(err, logFrom, err != nil)
}

// httpServiceEdit handles creating/editing a Service.
//
// # POST - create
//
// # PUT - replace
//
// service_name: service to edit (empty for new service)
//
// ...payload: Service they'd like to create/edit with
func (api *API) httpServiceEdit(w http.ResponseWriter, r *http.Request) {
	api.Config.OrderMutex.RLock()
	defer api.Config.OrderMutex.RUnlock()

	// service to modify (empty for create new)
	targetService, _ := url.QueryUnescape(mux.Vars(r)["service_name"])
	reqType := "create"
	if targetService != "" {
		reqType = "edit"
	}

	logFrom := &util.LogFrom{Primary: "httpServiceEdit", Secondary: getIP(r)}
	jLog.Verbose(fmt.Sprintf("%s %s", reqType, targetService),
		logFrom, true)

	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "application/json")

	var oldServiceSummary *api_type.ServiceSummary
	// EDIT the existing service
	if targetService != "" {
		if api.Config.Service[targetService] == nil {
			failRequest(&w, fmt.Sprintf("edit %q failed, service could not be found", targetService))
			return
		}
		oldServiceSummary = api.Config.Service[targetService].Summary()
	}

	// Payload
	payload := http.MaxBytesReader(w, r.Body, 102400)

	// Create the new/edited service
	targetServicePtr := api.Config.Service[targetService]
	newService, err := service.FromPayload(
		targetServicePtr, // nil if creating new
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
		failRequest(&w, fmt.Sprintf(`%s %q failed (invalid json)\%s`,
			reqType, targetService, err.Error()))
		return
	}

	// CREATE a new service, but one with the same name already exists
	if targetService == "" && api.Config.Service[newService.ID] != nil {
		failRequest(&w, fmt.Sprintf("create %q failed, service with this name already exists",
			newService.ID))
		return
	}

	// Check the values
	err = newService.CheckValues("")
	if err != nil {
		jLog.Error(err, logFrom, true)
		// Remove the service name from the error
		err = errors.New(strings.Join(strings.Split(err.Error(), `\`)[1:], `\`))

		failRequest(&w, fmt.Sprintf(`%s %q failed (invalid values)\%s`,
			reqType, util.FirstNonDefault(targetService, newService.ID), err.Error()))
		return
	}

	// Ensure LatestVersion and DeployedVersion (if set) can fetch
	err = newService.CheckFetches()
	if err != nil {
		jLog.Error(err, logFrom, true)

		failRequest(&w, fmt.Sprintf(`%s %q failed (fetches failed)\%s`,
			reqType, util.FirstNonDefault(targetService, newService.ID), err.Error()))
		return
	}

	// Set DeployedVersion to the LatestVersion if there's no DeployedVersionLookup
	if newService.DeployedVersionLookup == nil {
		newService.Status.SetDeployedVersion(newService.Status.LatestVersion(), false)
		newService.Status.SetDeployedVersionTimestamp(newService.Status.LatestVersionTimestamp())
	}

	// Add the new service to the config
	api.Config.OrderMutex.RUnlock() // Locked above
	//nolint:errcheck // Fail for duplicate service name is handled above
	api.Config.AddService(targetService, newService)
	api.Config.OrderMutex.RLock() // Lock again for the defer

	newServiceSummary := newService.Summary()
	// Announce the edit
	api.announceEdit(oldServiceSummary, newServiceSummary)
}

// httpServiceDelete handles deleting a Service.
//
// # DELETE
//
// service_name: service to delete
func (api *API) httpServiceDelete(w http.ResponseWriter, r *http.Request) {
	// service to delete
	targetService, _ := url.QueryUnescape(mux.Vars(r)["service_name"])

	logFromPrimary := "httpServiceDelete"
	logFrom := &util.LogFrom{Primary: logFromPrimary, Secondary: getIP(r)}
	jLog.Verbose(targetService, logFrom, true)

	// If service doesn't exist, return 404.
	if api.Config.Service[targetService] == nil {
		failRequest(&w, fmt.Sprintf("delete %q failed, service not found", targetService))
		return
	}

	api.Config.DeleteService(targetService)

	// Announce deletion
	api.announceDelete(targetService)

	// Return 200
	w.WriteHeader(http.StatusOK)
	//nolint:errcheck
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
//	name_previous?: string (the name of the notify before the current changes)
//	name?: string (required if name_previous not set)
//	type?: string
//	options?: map[string]string
//	url_fields?: map[string]string
//	params?: map[string]string
//	service_url?: string
//	web_url?: string
func (api *API) httpNotifyTest(w http.ResponseWriter, r *http.Request) {
	// Set headers
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "application/json")

	logFrom := &util.LogFrom{Primary: "httpNotifyTest", Secondary: getIP(r)}
	jLog.Verbose("-", logFrom, true)

	// Payload
	payload := http.MaxBytesReader(w, r.Body, 102400)
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(payload); err != nil {
		jLog.Error(err, logFrom, true)
		failRequest(&w, err.Error())
		return
	}
	var parsedPayload shoutrrr.TestPayload
	err := json.Unmarshal(buf.Bytes(), &parsedPayload)
	if err != nil {
		jLog.Error(err, logFrom, true)
		failRequest(&w, err.Error())
		return
	}

	// Get the Notify
	var serviceNotify *shoutrrr.Shoutrrr
	var latestVersion string
	if parsedPayload.ServiceNamePrevious != "" {
		api.Config.OrderMutex.RLock()
		defer api.Config.OrderMutex.RUnlock()
		// Check that the service exists
		if api.Config.Service[parsedPayload.ServiceNamePrevious] != nil {
			// Check that the notify exists
			if api.Config.Service[parsedPayload.ServiceNamePrevious].Notify != nil {
				serviceNotify = api.Config.Service[parsedPayload.ServiceNamePrevious].Notify[parsedPayload.NamePrevious]
			}
			latestVersion = api.Config.Service[parsedPayload.ServiceNamePrevious].Status.LatestVersion()
		}
	}
	testNotify, serviceURL, err := shoutrrr.FromPayload(
		&parsedPayload,
		serviceNotify,
		api.Config.Notify,
		api.Config.Defaults.Notify,
		api.Config.HardDefaults.Notify)
	if err != nil {
		jLog.Error(err, logFrom, true)
		failRequest(&w, err.Error())
		return
	}
	testNotify.ServiceStatus.SetLatestVersion(latestVersion, false)

	// Test the notify
	err = testNotify.TestSend(serviceURL)
	if err != nil {
		jLog.Error(err, logFrom, true)
		failRequest(&w, err.Error())
		return
	}

	// Return 200
	w.WriteHeader(http.StatusOK)
	//nolint:errcheck
	w.Write([]byte(`{}`))
}
