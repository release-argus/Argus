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

// Package db provides database functionality for Argus to keep track of versions found/deployed/approved.
package db

import (
	"database/sql"

	"github.com/release-argus/Argus/config"
	logutil "github.com/release-argus/Argus/util/log"
)

type api struct {
	config *config.Config
	db     *sql.DB
}

var (
	logFrom logutil.LogFrom = logutil.LogFrom{Primary: "db"}
)
