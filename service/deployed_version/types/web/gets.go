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

// Package web provides a web-based lookup type.
package web

import (
	"io"
	"strings"

	"github.com/release-argus/Argus/util"
)

// allowInvalidCerts returns whether invalid HTTPS certs are allowed.
func (l *Lookup) allowInvalidCerts() bool {
	return *util.FirstNonNilPtr(
		l.AllowInvalidCerts,
		l.Defaults.AllowInvalidCerts,
		l.HardDefaults.AllowInvalidCerts)
}

// body returns the Body of the Lookup.
func (l *Lookup) body() io.Reader {
	if l.Body == "" {
		return nil
	}
	return strings.NewReader(l.Body)
}

// method returns the method of the Lookup.
func (l *Lookup) method() string {
	return util.FirstNonDefault(
		l.Method,
		l.Defaults.Method,
		l.HardDefaults.Method)
}

// url returns the URL of the Lookup.
func (l *Lookup) url() string {
	return util.EvalEnvVars(l.URL)
}
