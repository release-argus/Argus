// Copyright [2022] [Argus]
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
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/release-argus/Argus/service"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	api_type "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/web/ui"
	"github.com/vearutop/statigz"
	"github.com/vearutop/statigz/brotli"
)

func (api *API) basicAuth() mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			username, password, ok := r.BasicAuth()
			if ok {
				// Hash purely to prevent ConstantTimeCompare leaking lengths
				usernameHash := sha256.Sum256([]byte(username))
				passwordHash := sha256.Sum256([]byte(password))
				expectedUsernameHash := sha256.Sum256([]byte(api.Config.Settings.Web.BasicAuth.Username))
				expectedPasswordHash := sha256.Sum256([]byte(api.Config.Settings.Web.BasicAuth.Password))

				// Protect from possible timing attacks
				usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
				passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

				if usernameMatch && passwordMatch {
					h.ServeHTTP(w, r)
					return
				}
			}

			w.Header().Set("Connection", "close")
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

// SetupRoutesAPI will setup the HTTP API routes.
func (api *API) SetupRoutesAPI() {
	api.Router.HandleFunc("/api/v1/version", api.httpVersion).Methods("GET")
	// GET, service-edit - refresh on create
	api.Router.HandleFunc("/api/v1/latest_version/refresh", api.httpVersionRefreshUncreated).Methods("GET")
	api.Router.HandleFunc("/api/v1/deployed_version/refresh", api.httpVersionRefreshUncreated).Methods("GET")
	// GET, service-edit - refresh
	api.Router.HandleFunc("/api/v1/latest_version/refresh/{service_name:.+}", api.httpVersionRefresh).Methods("GET")
	api.Router.HandleFunc("/api/v1/deployed_version/refresh/{service_name:.+}", api.httpVersionRefresh).Methods("GET")
	// GET, service-edit - get details
	api.Router.HandleFunc("/api/v1/service/edit", api.httpEditServiceGetOtherDetails).Methods("GET")
	api.Router.HandleFunc("/api/v1/service/edit/{service_name:.+}", api.httpEditServiceGetDetail).Methods("GET")
	// PUT, service-edit - update details
	api.Router.HandleFunc("/api/v1/service/edit/{service_name:.+}", api.httpEditServiceEdit).Methods("PUT")
	// POST, service-edit - new service
	api.Router.HandleFunc("/api/v1/service/new", api.httpEditServiceEdit).Methods("POST")
	// DELETE, service-edit
	api.Router.HandleFunc("/api/v1/service/delete/{service_name:.+}", api.httpEditServiceDelete).Methods("DELETE")
}

// SetupRoutesNodeJS will setup the HTTP routes to the NodeJS files.
func (api *API) SetupRoutesNodeJS() {
	nodeRoutes := []string{
		"/approvals",
		"/config",
		"/flags",
		"/status",
	}
	// Serve the NodeJS files
	for _, route := range nodeRoutes {
		prefix := strings.TrimRight(api.RoutePrefix, "/") + route
		api.Router.Handle(route, http.StripPrefix(prefix, statigz.FileServer(ui.GetFS().(fs.ReadDirFS), brotli.AddEncoding)))
	}

	// Catch-all for JS, CSS, etc...
	api.Router.PathPrefix("/").Handler(http.StripPrefix(api.RoutePrefix, statigz.FileServer(ui.GetFS().(fs.ReadDirFS), brotli.AddEncoding)))
}

// httpVersion serves Argus version JSON over HTTP.
func (api *API) httpVersion(w http.ResponseWriter, r *http.Request) {
	logFrom := util.LogFrom{Primary: "httpVersion", Secondary: getIP(r)}
	api.Log.Verbose("-", logFrom, true)

	// Set headers
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(api_type.VersionAPI{
		Version:   util.Version,
		BuildDate: util.BuildDate,
		GoVersion: util.GoVersion,
	})
	api.Log.Error(err, logFrom, err != nil)
}

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
	logFrom := util.LogFrom{Primary: logFromPrimary, Secondary: getIP(r)}
	api.Log.Verbose("-", logFrom, true)

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
		deployedVersionLookup := deployedver.Lookup{
			Options: &opt.Options{
				Defaults:     &api.Config.Defaults.Service.Options,
				HardDefaults: &api.Config.HardDefaults.Service.Options},
			Status:       &status,
			Defaults:     api.Config.Defaults.Service.DeployedVersionLookup,
			HardDefaults: api.Config.HardDefaults.Service.DeployedVersionLookup,
		}
		// Deployed Version
		version, err = deployedVersionLookup.Refresh(
			getParam(&queryParams, "allow_invalid_certs"),
			getParam(&queryParams, "basic_auth"),
			getParam(&queryParams, "headers"),
			getParam(&queryParams, "json"),
			getParam(&queryParams, "regex"),
			getParam(&queryParams, "semantic_versioning"),
			getParam(&queryParams, "url"),
		)
	} else {
		latestVersion := latestver.Lookup{
			Options: &opt.Options{
				Defaults:     &api.Config.Defaults.Service.Options,
				HardDefaults: &api.Config.HardDefaults.Service.Options,
			},
			Status:       &status,
			Defaults:     &api.Config.Defaults.Service.LatestVersion,
			HardDefaults: &api.Config.HardDefaults.Service.LatestVersion,
		}
		// Latest Version
		version, err = latestVersion.Refresh(
			getParam(&queryParams, "access_token"),
			getParam(&queryParams, "allow_invalid_certs"),
			getParam(&queryParams, "require"),
			getParam(&queryParams, "semantic_versioning"),
			getParam(&queryParams, "type"),
			getParam(&queryParams, "url"),
			getParam(&queryParams, "url_commands"),
			getParam(&queryParams, "use_prerelease"),
		)
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
	api.Log.Error(err, logFrom, err != nil)
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
	logFrom := util.LogFrom{Primary: logFromPrimary, Secondary: getIP(r)}
	api.Log.Verbose(targetService, logFrom, true)

	// Set headers
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "application/json")

	queryParams := r.URL.Query()

	// Check if service exists
	api.Config.OrderMutex.RLock()
	defer api.Config.OrderMutex.RUnlock()
	if api.Config.Service[targetService] == nil {
		err := fmt.Sprintf("service %q not found", targetService)
		api.Log.Verbose(err, logFrom, true)
		failRequest(&w, err, http.StatusNotFound)
		return
	}

	// Refresh the latest/deployed version lookup type
	var (
		version string
		err     error
	)
	if deployedVersionRefresh {
		if api.Config.Service[targetService].DeployedVersionLookup == nil {
			api.Config.Service[targetService].DeployedVersionLookup = &deployedver.Lookup{
				Options: &api.Config.Service[targetService].Options,
				Status: &svcstatus.Status{
					ServiceID: &targetService},
				Defaults:     api.Config.Defaults.Service.DeployedVersionLookup,
				HardDefaults: api.Config.HardDefaults.Service.DeployedVersionLookup,
			}
		}
		// Deployed Version
		version, err = api.Config.Service[targetService].DeployedVersionLookup.Refresh(
			getParam(&queryParams, "allow_invalid_certs"),
			getParam(&queryParams, "basic_auth"),
			getParam(&queryParams, "headers"),
			getParam(&queryParams, "json"),
			getParam(&queryParams, "regex"),
			getParam(&queryParams, "semantic_versioning"),
			getParam(&queryParams, "url"),
		)
	} else {
		// Latest Version
		version, err = api.Config.Service[targetService].LatestVersion.Refresh(
			getParam(&queryParams, "access_token"),
			getParam(&queryParams, "allow_invalid_certs"),
			getParam(&queryParams, "require"),
			getParam(&queryParams, "semantic_versioning"),
			getParam(&queryParams, "type"),
			getParam(&queryParams, "url"),
			getParam(&queryParams, "url_commands"),
			getParam(&queryParams, "use_prerelease"),
		)
	}

	err = json.NewEncoder(w).Encode(api_type.RefreshAPI{
		Version: version,
		Error:   util.ErrorToString(err),
		Date:    time.Now().UTC(),
	})
	api.Log.Error(err, logFrom, err != nil)
}

// httpEditServiceGetDetail handles sending details about a Service
//
// # GET
//
// service_name: service to get details for
func (api *API) httpEditServiceGetDetail(w http.ResponseWriter, r *http.Request) {
	// service to get details from (empty for create new)
	targetService, _ := url.QueryUnescape(mux.Vars(r)["service_name"])

	logFromPrimary := "httpEditServiceGetDetail"
	logFrom := util.LogFrom{Primary: logFromPrimary, Secondary: getIP(r)}
	api.Log.Verbose(targetService, logFrom, true)

	// Set Headers
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "application/json")

	// Find the Service
	api.Config.OrderMutex.RLock()
	addFail := true
	defer func() {
		if addFail {
			api.Config.OrderMutex.RUnlock()
		}
	}()
	if api.Config.Service[targetService] == nil {
		err := fmt.Sprintf("service %q not found", targetService)
		api.Log.Verbose(err, logFrom, true)
		failRequest(&w, err, http.StatusNotFound)
		return
	}
	service := api.Config.Service[targetService]
	addFail = false
	api.Config.OrderMutex.RUnlock()

	// Convert to API Type, censoring secrets
	serviceConfig := convertServiceToAPITypeService(service)
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
	api.Log.Error(err, logFrom, err != nil)
}

// httpEditServiceGetOtherDetails handles sending details about the global notify/webhook's, defaults and hard defaults.
//
// # GET
func (api *API) httpEditServiceGetOtherDetails(w http.ResponseWriter, r *http.Request) {
	logFromPrimary := "httpEditServiceGetOtherDetails"
	logFrom := util.LogFrom{Primary: logFromPrimary, Secondary: getIP(r)}
	api.Log.Verbose("-", logFrom, true)

	// Set headers
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "application/json")

	// Convert to JSON type that swaps slices for lists
	err := json.NewEncoder(w).Encode(api_type.Config{
		HardDefaults: convertAndCensorDefaults(&api.Config.HardDefaults),
		Defaults:     convertAndCensorDefaults(&api.Config.Defaults),
		Notify:       convertAndCensorNotifySlice(&api.Config.Notify),
		WebHook:      convertAndCensorWebHookSlice(&api.Config.WebHook),
	})
	api.Log.Error(err, logFrom, err != nil)
}

// httpEditServiceEdit handles creating/editing a Service.
//
// # POST - create
//
// # PUT - replace
//
// service_name: service to edit (empty for new service)
//
// ...payload: Service they'd like to create/edit with
func (api *API) httpEditServiceEdit(w http.ResponseWriter, r *http.Request) {
	api.Config.OrderMutex.RLock()
	defer api.Config.OrderMutex.RUnlock()

	// service to modify (empty for create new)
	targetService, _ := url.QueryUnescape(mux.Vars(r)["service_name"])

	logFromPrimary := "httpEditServiceEdit"
	logFrom := util.LogFrom{Primary: logFromPrimary, Secondary: getIP(r)}
	api.Log.Verbose(targetService, logFrom, true)

	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "application/json")

	var oldServiceSummary *api_type.ServiceSummary
	// EDIT the existing service
	if targetService != "" {
		api.Config.OrderMutex.RLock()
		if api.Config.Service[targetService] == nil {
			api.Config.OrderMutex.RUnlock()
			failRequest(&w, fmt.Sprintf("edit %q failed, service could not be found", targetService))
			return
		}
		oldServiceSummary = api.Config.Service[targetService].Summary()
		api.Config.OrderMutex.RUnlock()
	}

	// Payload
	payload := http.MaxBytesReader(w, r.Body, 104858)

	// Create the new/edited service
	reqType := "create"
	if targetService != "" {
		reqType = "edit"
	}
	api.Config.OrderMutex.RLock()
	targetServicePtr := api.Config.Service[targetService]
	api.Config.OrderMutex.RUnlock()
	newService, err := service.New(
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
		&logFrom)
	if err != nil {
		api.Log.Error(err, logFrom, true)
		failRequest(&w, fmt.Sprintf(`%s %q failed (invalid json)\%s`,
			reqType, targetService, err.Error()))
		return
	}

	// CREATE a new service, but one with the same name already exists
	api.Config.OrderMutex.RLock()
	if targetService == "" && api.Config.Service[newService.ID] != nil {
		api.Config.OrderMutex.RUnlock()
		failRequest(&w, fmt.Sprintf("create %q failed, service with this name already exists",
			newService.ID))
		return
	}
	api.Config.OrderMutex.RUnlock()

	// Check the values
	err = newService.CheckValues("")
	if err != nil {
		api.Log.Error(err, logFrom, true)
		// Remove the service name from the error
		err = errors.New(strings.Join(strings.Split(err.Error(), `\`)[1:], `\`))

		failRequest(&w, fmt.Sprintf(`%s %q failed (invalid values)\%s`, reqType, targetService, err.Error()))
		return
	}

	// Ensure LatestVersion and DeployedVersion (if set) can fetch
	err = newService.CheckFetches()
	if err != nil {
		api.Log.Error(err, logFrom, true)

		failRequest(&w, fmt.Sprintf(`%s %q failed (fetches failed)\%s`, reqType, targetService, err.Error()))
		return
	}

	// Set DeployedVersion to the LatestVersion if there's no DeployedVersionLookup
	if newService.DeployedVersionLookup == nil {
		newService.Status.SetDeployedVersion(newService.Status.GetLatestVersion(), false)
		newService.Status.SetDeployedVersionTimestamp(newService.Status.GetLatestVersionTimestamp())
	}

	// Add the new service to the config
	//nolint:errcheck // Fail for duplicate service name is handled above
	api.Config.AddService(targetService, newService)

	newServiceSummary := newService.Summary()
	// Announce the edit
	api.announceEdit(oldServiceSummary, newServiceSummary)
}

// httpEditServiceDelete handles deleting a Service.
//
// # DELETE
//
// service_name: service to delete
func (api *API) httpEditServiceDelete(w http.ResponseWriter, r *http.Request) {
	// service to delete
	targetService, _ := url.QueryUnescape(mux.Vars(r)["service_name"])

	logFromPrimary := "httpEditServiceDelete"
	logFrom := util.LogFrom{Primary: logFromPrimary, Secondary: getIP(r)}
	api.Log.Verbose(targetService, logFrom, true)

	// If service doesn't exist, return 404.
	if api.Config.Service[targetService] == nil {
		failRequest(&w, fmt.Sprintf("Delete %q failed, service not found", targetService))
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

// failRequest with a JSON response containing a message and status code.
func failRequest(w *http.ResponseWriter, message string, statusCodeOverride ...int) {
	// Default to 400 Bad Request
	statusCode := http.StatusBadRequest
	// Override if provided
	if len(statusCodeOverride) > 0 {
		statusCode = statusCodeOverride[0]
	}

	// Write response
	(*w).WriteHeader(statusCode)
	resp := map[string]string{
		"message": message,
	}
	jsonResp, _ := json.Marshal(resp)
	//nolint:errcheck
	(*w).Write(jsonResp)
}
