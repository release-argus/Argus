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

package latest_version

import (
	url_command "github.com/release-argus/Argus/service/url_commands"
	api_types "github.com/release-argus/Argus/web/api/types"
)

type Lookup interface {
	GetType() string
	GetFriendlyURL() string
	GetLookupURL() string

	ConvertToAPIType() *api_types.LatestVersion
	GetURLCommands() *url_command.Slice

	Query() (bool, error)

	CheckValues(string) error
	Print(string)
}
