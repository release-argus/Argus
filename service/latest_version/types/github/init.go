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

// Package github provides a github-based lookup type.
package github

import (
	"github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
)

// Init the Lookup, assigning Defaults, and initialising child structs.
func (l *Lookup) Init(
	options *opt.Options,
	status *status.Status,
	defaults, hardDefaults *base.Defaults,
) {
	l.Lookup.Init(
		options,
		status,
		defaults, hardDefaults)

	l.data.SetETag(getEmptyListETag())
}
