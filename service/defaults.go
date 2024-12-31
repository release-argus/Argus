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

// Package service provides the service functionality for Argus.
package service

// Default sets this Defaults to the default values.
func (d *Defaults) Default() {
	// Service.Options
	d.Options.Default()

	// Service.LatestVersion
	d.LatestVersion.Default()

	// Service.DeployedVersionLookup
	d.DeployedVersionLookup.Default()

	// Service.Dashboard
	serviceAutoApprove := false
	d.Dashboard.AutoApprove = &serviceAutoApprove

	d.Init()
}

// Init will hand out the appropriate Defaults.X pointer(s) between structs.
func (d *Defaults) Init() {
	d.LatestVersion.Options = &d.Options
	d.DeployedVersionLookup.Options = &d.Options
}
