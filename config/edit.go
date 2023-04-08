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

package config

import (
	"fmt"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/util"
)

// AddService to the config (or replace/rename an existing service).
func (c *Config) AddService(oldServiceID string, newService *service.Service) (err error) {
	c.OrderMutex.Lock()
	defer c.OrderMutex.Unlock()
	logFrom := util.LogFrom{Primary: "AddService"}

	// Check the service doesn't already exist if the name is changing
	if oldServiceID != newService.ID && c.Service[newService.ID] != nil {
		err = fmt.Errorf("service %q already exists", newService.ID)
		jLog.Error(err, logFrom, true)
		return
	}

	logFrom.Secondary = newService.ID
	// New service
	if oldServiceID == "" {
		jLog.Info("Adding service", logFrom, true)
		c.Order = append(c.Order, newService.ID)
		// Create the service map if it doesn't exist
		//nolint:typecheck
		if c.Service == nil {
			c.Service = make(map[string]*service.Service)
		}

		// Targeting an existing service
	} else {
		// Keeping the same ID
		if oldServiceID == newService.ID {
			jLog.Info("Replacing service", logFrom, true)

			// Old service being given a new ID
		} else {
			c.RenameService(oldServiceID, newService)
		}
	}

	// Add/Replace the service in the config
	c.Service[newService.ID] = newService

	// Trigger a save if the Service has changed
	oldService := c.Service[oldServiceID]
	changedService := oldServiceID != newService.ID ||
		oldService.String() != c.Service[newService.ID].String()
	if changedService {
		*c.HardDefaults.Service.Status.SaveChannel <- true
	}

	// Update the database if the Service or versions changed
	newLV := newService.Status.GetLatestVersion()
	newDV := newService.Status.GetDeployedVersion()
	if changedService ||
		newLV != oldService.Status.GetLatestVersion() ||
		newDV != oldService.Status.GetDeployedVersion() {
		*c.HardDefaults.Service.Status.DatabaseChannel <- dbtype.Message{
			ServiceID: newService.ID,
			Cells: []dbtype.Cell{
				{Column: "latest_version", Value: newLV},
				{Column: "latest_version_timestamp", Value: newService.Status.GetLatestVersionTimestamp()},
				{Column: "deployed_version", Value: newDV},
				{Column: "deployed_version_timestamp", Value: newService.Status.GetDeployedVersionTimestamp()},
				{Column: "approved_version", Value: newService.Status.GetApprovedVersion()}}}
	}

	// Start tracking the service
	go c.Service[newService.ID].Track()

	return
}

// RenameService in the config from `oldService` to `newService` and remove `oldService`.
func (c *Config) RenameService(oldService string, newService *service.Service) {
	// Check whether the service being renamed doesn't exist
	// or a rename isn't required (name is the same)
	// or a service with this new name already exists
	if c.Service[oldService] == nil || oldService == newService.ID || c.Service[newService.ID] != nil {
		return
	}

	jLog.Info(
		fmt.Sprintf("%q", newService.ID),
		util.LogFrom{Primary: "RenameService", Secondary: oldService},
		true)
	// Replace the service in the order/config
	c.Order = util.ReplaceElement(c.Order, oldService, newService.ID)
	c.Service[newService.ID] = newService
	// Rename the primary key for this service in the database
	*c.HardDefaults.Service.Status.DatabaseChannel <- dbtype.Message{
		ServiceID: oldService,
		Cells: []dbtype.Cell{
			{Column: "id", Value: newService.ID}}}
	// Remove the old service
	c.Service[oldService].PrepDelete()
	delete(c.Service, oldService)
}

// DeleteService from the config.
func (c *Config) DeleteService(serviceID string) {
	c.OrderMutex.Lock()
	defer c.OrderMutex.Unlock()

	// Check whether the service exists
	if c.Service[serviceID] == nil {
		return
	}

	jLog.Info(
		"Deleting service",
		util.LogFrom{Primary: "DeleteService", Secondary: serviceID},
		true)
	// Remove the service from the Order
	c.Order = util.RemoveElement(c.Order, serviceID)

	// nil the channels and set the `deleting` flag
	c.Service[serviceID].PrepDelete()

	// Remove the service from the config
	delete(c.Service, serviceID)

	// Trigger save
	*c.HardDefaults.Service.Status.SaveChannel <- true
}
