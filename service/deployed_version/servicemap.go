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

// Package deployedver provides the deployed_version lookup service to for a service.
package deployedver

import (
	"github.com/release-argus/Argus/service/deployed_version/types/base"
	"github.com/release-argus/Argus/service/deployed_version/types/manual"
	"github.com/release-argus/Argus/service/deployed_version/types/web"
)

// PossibleTypes for the deployed_version Lookup.
var PossibleTypes = []string{
	"url",
	"manual",
}

// ServiceMap maps a service type to a Lookup constructor.
var ServiceMap = map[string]func() base.Interface{
	"url":    func() base.Interface { return &web.Lookup{} },
	"web":    func() base.Interface { return &web.Lookup{} },
	"manual": func() base.Interface { return &manual.Lookup{} },
}
