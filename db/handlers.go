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

package db

import (
	"fmt"

	db_types "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/utils"
	_ "modernc.org/sqlite"
)

// handler will listen to the DatabaseChannel and act on
// incoming messages.
func (api *api) handler() {
	for message := range *api.config.DatabaseChannel {
		api.updateRow(
			message.ServiceID,
			message.Cells,
		)
	}
}

// updateRow will update the cells of the serviceID row.
func (api *api) updateRow(serviceID string, cells []db_types.Cell) {
	replace := ""
	for i := range cells {
		replace += fmt.Sprintf("'%s' = '%s',",
			cells[i].Column, cells[i].Value)
	}
	sqlStmt := fmt.Sprintf("UPDATE status SET %s WHERE id = '%s'",
		replace[:len(replace)-1], serviceID)
	_, err := api.db.Exec(sqlStmt)
	jLog.Error(
		fmt.Sprintf("updateRow: %s", utils.ErrorToString(err)),
		*logFrom,
		err != nil)
}
