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

// Package config provides the configuration for Argus.
package config

import (
	"fmt"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

// AddService to the config (or replace/rename an existing service).
func (c *Config) AddService(oldServiceID string, newService *service.Service) error {
	c.OrderMutex.Lock()
	defer c.OrderMutex.Unlock()
	logFrom := logutil.LogFrom{Primary: "AddService"}

	// Check a service does not already exist with the new id/name, if the name is changing.
	if oldServiceID != newService.ID &&
		(c.Service[newService.ID] != nil || c.ServiceWithNameExists(newService.ID, oldServiceID)) {
		err := fmt.Errorf("service %q already exists", newService.ID)
		logutil.Log.Error(err, logFrom, true)
		return err
	}

	logFrom.Secondary = newService.ID
	// Whether we need to save the config.
	changedService := oldServiceID != newService.ID ||
		c.Service[oldServiceID].String("") != newService.String("")
	// Whether we need to update the database.
	changedDB := oldServiceID == "" || c.Service[oldServiceID] == nil ||
		!c.Service[oldServiceID].Status.SameVersions(&newService.Status)
	// New service.
	if oldServiceID == "" || c.Service[oldServiceID] == nil {
		logutil.Log.Info("Adding service", logFrom, true)
		c.Order = append(c.Order, newService.ID)
		// Create the service map if it doesn't exist.
		//nolint:typecheck
		if c.Service == nil {
			c.Service = make(map[string]*service.Service)
		}

		// Targeting an existing service.
	} else {
		// Keeping the same ID.
		if oldServiceID == newService.ID {
			logutil.Log.Info("Replacing service", logFrom, true)
			// Delete the old service.
			c.Service[oldServiceID].PrepDelete(false)

			// Old service being given a new ID.
		} else {
			c.RenameService(oldServiceID, newService)
		}
	}

	// Add/Replace the service in the config.
	c.Service[newService.ID] = newService

	// Trigger a save if the Service has changed.
	if changedService {
		*c.HardDefaults.Service.Status.SaveChannel <- true
	}

	// Update the database if the service is new, or the versions changed.
	if changedDB {
		*c.HardDefaults.Service.Status.DatabaseChannel <- dbtype.Message{
			ServiceID: newService.ID,
			Cells: []dbtype.Cell{
				{Column: "latest_version", Value: newService.Status.LatestVersion()},
				{Column: "latest_version_timestamp", Value: newService.Status.LatestVersionTimestamp()},
				{Column: "deployed_version", Value: newService.Status.DeployedVersion()},
				{Column: "deployed_version_timestamp", Value: newService.Status.DeployedVersionTimestamp()},
				{Column: "approved_version", Value: newService.Status.ApprovedVersion()}}}
	}

	// Start tracking the service.
	go c.Service[newService.ID].Track()

	return nil
}

// ServiceWithNameExists checks whether a Service with the given name exists.
func (c *Config) ServiceWithNameExists(name, oldServiceID string) bool {
	// Have no name, so skip the check.
	if name == "" {
		return false
	}

	// Check if the name exists in a service.
	for id, svc := range c.Service {
		// Name exists, and not from the service being modified.
		if svc.Name == name && id != oldServiceID {
			return true
		}
	}
	return false
}

// RenameService in the config from `oldService` to `newService` and remove `oldService`.
func (c *Config) RenameService(oldService string, newService *service.Service) {
	// Check whether the target service doesn't exist,
	// or a rename is not required (name is the same),
	// or a service with this new name already exists.
	if c.Service[oldService] == nil || oldService == newService.ID || c.Service[newService.ID] != nil {
		return
	}

	logutil.Log.Info(
		fmt.Sprintf("%q", newService.ID),
		logutil.LogFrom{Primary: "RenameService", Secondary: oldService},
		true)
	// Replace the service in the order/config.
	c.Order = util.ReplaceElement(c.Order, oldService, newService.ID)
	c.Service[newService.ID] = newService
	// Rename the primary key for this service in the database.
	*c.HardDefaults.Service.Status.DatabaseChannel <- dbtype.Message{
		ServiceID: oldService,
		Cells: []dbtype.Cell{
			{Column: "id", Value: newService.ID}}}
	// Remove the old service.
	c.Service[oldService].PrepDelete(false)
	delete(c.Service, oldService)
}

// DeleteService from the config.
func (c *Config) DeleteService(serviceID string) {
	c.OrderMutex.Lock()
	defer c.OrderMutex.Unlock()

	// Check whether the service exists.
	if c.Service[serviceID] == nil {
		return
	}

	logutil.Log.Info(
		"Deleting service",
		logutil.LogFrom{Primary: "DeleteService", Secondary: serviceID},
		true)
	// Remove the service from the Order.
	c.Order = util.RemoveElement(c.Order, serviceID)

	// nil the channels and set the `deleting` flag.
	c.Service[serviceID].PrepDelete(true)

	// Remove the service from the config.
	delete(c.Service, serviceID)

	// Trigger save.
	*c.HardDefaults.Service.Status.SaveChannel <- true
}
