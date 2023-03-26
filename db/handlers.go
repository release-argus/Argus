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
	"strings"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/util"
)

// handler will listen to the DatabaseChannel and act on
// incoming messages.
func (api *api) handler() {
	for message := range *api.config.DatabaseChannel {
		// If the message is to delete the row
		if message.Delete {
			api.deleteRow(message.ServiceID)
			continue
		}

		// Else, the message is to update the row
		api.updateRow(
			message.ServiceID,
			message.Cells,
		)
	}
}

// updateRow will update the cells of the serviceID row.
func (api *api) updateRow(serviceID string, cells []dbtype.Cell) {
	// The columns to update
	setVars := ""
	for i := range cells {
		setVars += fmt.Sprintf("`%s` = ?,",
			cells[i].Column)
	}
	// Trim the trailing ,
	setVars = setVars[:len(setVars)-1]

	// Get the vars for the SQL statement
	params := make([]interface{}, len(cells)+1)
	params[len(params)-1] = serviceID
	// The values to update with
	for i := range cells {
		params[i] = cells[i].Value
	}

	// The SQL statement
	sqlStmt := fmt.Sprintf("UPDATE status SET %s WHERE id = ?",
		setVars)

	jLog.Debug(fmt.Sprintf("%s, %v", sqlStmt, params), *logFrom, true)
	res, err := api.db.Exec(sqlStmt, params...)
	// Query failed
	if err != nil {
		jLog.Error(
			fmt.Sprintf("updateRow UPDATE: %q %v, %s", sqlStmt, params, util.ErrorToString(err)),
			*logFrom,
			true)
		return
	}

	count, _ := res.RowsAffected()
	// If this ServiceID wasn't in the DB
	if count == 0 {
		// The columns to insert
		columns := ""
		for i := range cells {
			columns += fmt.Sprintf("'%s',", cells[i].Column)
		}
		// The values to insert
		values := strings.Repeat("?,", len(cells))
		// Trim the trailing ,'s
		values = values[:len(values)-1]
		columns = columns[:len(columns)-1]

		// The SQL statement
		sqlStmt = fmt.Sprintf("INSERT INTO status ('id', %s) VALUES (?,%s)",
			columns, values)

		// Get the vars for the SQL statement
		params := make([]interface{}, len(cells)+1)
		params[0] = serviceID
		for i := range cells {
			params[i+1] = cells[i].Value
		}

		jLog.Debug(fmt.Sprintf("%s, %v", sqlStmt, params), *logFrom, true)
		_, err = api.db.Exec(sqlStmt, params...)
		jLog.Error(
			fmt.Sprintf("updateRow INSERT: %q %v, %s", sqlStmt, params, util.ErrorToString(err)),
			*logFrom,
			err != nil)
	}
}

// deleteRow will remove the row of a service from the db.
func (api *api) deleteRow(serviceID string) {
	// The SQL statement
	sqlStmt := "DELETE FROM status WHERE id = ?"

	jLog.Debug(fmt.Sprintf("%s, %v", sqlStmt, serviceID), *logFrom, true)
	_, err := api.db.Exec(sqlStmt, serviceID)
	jLog.Error(
		fmt.Sprintf("deleteRow: %q with %q, %s", sqlStmt, serviceID, util.ErrorToString(err)),
		*logFrom,
		err != nil)
}
