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

// Package latestver provides the latest_version lookup service to for a service.
package latestver

import (
	lvgithub "github.com/release-argus/Argus/service/latest_version/types/github"
	lvweb "github.com/release-argus/Argus/service/latest_version/types/web"
	"github.com/release-argus/Argus/util/polymorphic"
)

// PossibleTypes for the latest_version Lookup.
var PossibleTypes = []string{
	lvgithub.Type,
	lvweb.Type,
}

// ServiceMap maps a service type to a Lookup constructor.
var ServiceMap = map[string]func() Lookup{
	lvgithub.Type: func() Lookup { return &lvgithub.Lookup{} },
	lvweb.Type:    func() Lookup { return &lvweb.Lookup{} },
	"web":         func() Lookup { return &lvweb.Lookup{} },
}

// ServiceMapInheritable is [ServiceMap] wrapped for polymorphic inheritance decoding.
var ServiceMapInheritable = polymorphic.ToInheritableMap(ServiceMap)
