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

// ConvertCurrentVersionToDeployedVersion handles the deprecation of `current_version*`,
// Renaming it to `deployed_version*`
func (c *Config) ConvertCurrentVersionToDeployedVersion() {
	for service_id := range c.Service {
		if c.Service[service_id].Status != nil {
			if c.Service[service_id].Status.CurrentVersion != "" {
				if c.Service[service_id].Status.DeployedVersion == "" {
					c.Service[service_id].Status.DeployedVersion = c.Service[service_id].Status.CurrentVersion
				}
				c.Service[service_id].Status.CurrentVersion = ""
			}
			if c.Service[service_id].Status.CurrentVersionTimestamp != "" {
				if c.Service[service_id].Status.DeployedVersionTimestamp == "" {
					c.Service[service_id].Status.DeployedVersionTimestamp = c.Service[service_id].Status.CurrentVersionTimestamp
				}
				c.Service[service_id].Status.CurrentVersionTimestamp = ""
			}
		}
	}
}
