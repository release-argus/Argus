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
	"fmt"
	"strings"

	dbtype "github.com/release-argus/Argus/db/types"
	logutil "github.com/release-argus/Argus/util/log"
)

// handler will listen to the DatabaseChannel and act on
// incoming messages to the DatabaseChannel.
func (api *api) handler() {
	defer api.db.Close()
	for message := range *api.config.DatabaseChannel {
		// If the message is to delete a row.
		if message.Delete {
			api.deleteRow(message.ServiceID)
			continue
		}

		// Else, the message is to update a row.
		api.updateRow(
			message.ServiceID,
			message.Cells,
		)
	}
}

// updateRow will update the cells of the serviceID row.
func (api *api) updateRow(serviceID string, cells []dbtype.Cell) {
	if len(cells) == 0 {
		return
	}

	// The columns to update.
	var setVarsBuilder strings.Builder // `column` = ?,`column` = ?,...
	for i, cell := range cells {
		// Add separator before columns after the first.
		if i != 0 {
			setVarsBuilder.WriteString(",")
		}

		// `column` = ?
		setVarsBuilder.WriteString("`")
		setVarsBuilder.WriteString(cell.Column)
		setVarsBuilder.WriteString("` = ?")
	}

	// The SQL statement.
	//#nosec G201 -- setVarsBuilder is built from trusted sources.
	sqlStmt := fmt.Sprintf("UPDATE status SET %s WHERE id = ?",
		setVarsBuilder.String())

	// Get the vars for the SQL statement.
	params := make([]any, len(cells)+1)
	params[len(params)-1] = serviceID
	// The values to update with.
	for i := range cells {
		params[i] = cells[i].Value
	}

	if logutil.Log.IsLevel("DEBUG") {
		logutil.Log.Debug(
			fmt.Sprintf("%s, %v", sqlStmt, params),
			logFrom, true)
	}
	res, err := api.db.Exec(sqlStmt, params...)
	// Query failed.
	if err != nil {
		logutil.Log.Error(
			fmt.Sprintf("updateRow UPDATE: %q %v, %s",
				sqlStmt, params, err),
			logFrom, true)
		return
	}

	count, _ := res.RowsAffected()
	// If the row was updated, return.
	if count != 0 {
		return
	}

	// This ServiceID was not in the DB, insert it.
	// The columns to insert.
	var columnsBuilder strings.Builder           // `column`,`column`,...
	var valuesPlaceholderBuilder strings.Builder // ?,?,...
	for i, cell := range cells {
		// Add separator before entries after the first.
		if i != 0 {
			columnsBuilder.WriteString(",")
			valuesPlaceholderBuilder.WriteString(",")
		}

		// `column`
		columnsBuilder.WriteString("`")
		columnsBuilder.WriteString(cell.Column)
		columnsBuilder.WriteString("`")
		// ?
		valuesPlaceholderBuilder.WriteString("?")
	}

	// The SQL statement.
	sqlStmt = fmt.Sprintf("INSERT INTO status (%s,`id`) VALUES (?,%s)",
		columnsBuilder.String(), valuesPlaceholderBuilder.String())
	// Log the SQL statement.
	if logutil.Log.IsLevel("DEBUG") {
		logutil.Log.Debug(
			fmt.Sprintf("%s, %v", sqlStmt, params),
			logFrom, true)
	}

	// Execute and log any errors.
	if _, err := api.db.Exec(sqlStmt, params...); err != nil {
		logutil.Log.Error(
			fmt.Sprintf("updateRow INSERT: %q %v, %s",
				sqlStmt, params, err),
			logFrom, true)
	}
}

// deleteRow will remove the row of a service from the db.
func (api *api) deleteRow(serviceID string) {
	// The SQL statement.
	sqlStmt := "DELETE FROM status WHERE id = ?"
	// Log the SQL statement.
	if logutil.Log.IsLevel("DEBUG") {
		logutil.Log.Debug(
			fmt.Sprintf("%s, %v", sqlStmt, serviceID),
			logFrom, true)
	}

	// Execute and log any errors.
	if _, err := api.db.Exec(sqlStmt, serviceID); err != nil {
		logutil.Log.Error(
			fmt.Sprintf("deleteRow: %q with %q, %s",
				sqlStmt, serviceID, err),
			logFrom, true)
	}
}
