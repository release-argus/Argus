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

// Package db provides database functionality for Argus to keep track of versions found/deployed/approved.
package db

import (
	"context"
	"fmt"
	"strings"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/internal/logx"
)

// writeQuoted will write the string `s` to the builder `b` wrapped in backticks.
func writeQuoted(b *strings.Builder, s string) {
	b.WriteByte('`')
	b.WriteString(s)
	b.WriteByte('`')
}

// Handler processes database update and delete messages until ctx is cancelled.
func (api *api) Handler(ctx context.Context) {
	defer api.db.Close()

	for {
		select {
		case message, ok := <-api.config.DatabaseChannel:
			if !ok {
				logx.Fatal("Database closed", logFrom)
				return
			}

			// If the message is to delete a row.
			if message.Delete {
				api.deleteRow(message.ServiceID)
				// Else, the message is to update a row.
			} else {
				api.updateRow(
					message.ServiceID,
					message.Cells,
				)
			}

		case <-ctx.Done():
			return
		}
	}
}

// updateRow upserts the given cells for serviceID in the status table.
func (api *api) updateRow(serviceID string, cells []dbtype.Cell) {
	if len(cells) == 0 {
		return
	}

	// The columns to update.
	var (
		columnsBuilder strings.Builder
		placeholders   strings.Builder
		updateBuilder  strings.Builder
	)

	params := make([]any, 0, len(cells)+1)

	// Always include id first
	columnsBuilder.WriteString("`id`")
	placeholders.WriteString("?")
	params = append(params, serviceID)

	for i, cell := range cells {
		// columns
		columnsBuilder.WriteString(",")
		writeQuoted(&columnsBuilder, cell.Column)

		// values
		placeholders.WriteString(",?")
		params = append(params, cell.Value)

		// update clause
		if i != 0 {
			updateBuilder.WriteString(",")
		}
		writeQuoted(&updateBuilder, cell.Column)
		updateBuilder.WriteString(" = excluded.")
		writeQuoted(&updateBuilder, cell.Column)
	}

	// The SQL statement.
	//#nosec G201 -- setVarsBuilder is built from trusted sources.
	sqlStmt := fmt.Sprintf(`
			INSERT INTO status (%s)
			VALUES (%s)
			ON CONFLICT(`+"`id`"+`) DO UPDATE SET %s
		`,
		columnsBuilder.String(),
		placeholders.String(),
		updateBuilder.String(),
	)

	if logx.IsLevel("DEBUG") {
		logx.Debug(
			fmt.Sprintf("%s, %v", sqlStmt, params),
			logFrom,
			true,
		)
	}
	if _, err := api.db.Exec(sqlStmt, params...); err != nil {
		logx.Error(
			fmt.Sprintf(
				"updateRow UPSERT: %q %v, %s",
				sqlStmt, params, err,
			),
			logFrom,
			true,
		)
	}
}

// deleteRow removes the status row for serviceID.
func (api *api) deleteRow(serviceID string) {
	// The SQL statement.
	sqlStmt := "DELETE FROM status WHERE id = ?"
	// Log the SQL statement.
	if logx.IsLevel("DEBUG") {
		logx.Debug(
			fmt.Sprintf("%s, %v", sqlStmt, serviceID),
			logFrom,
			true,
		)
	}

	// Execute and log any errors.
	if _, err := api.db.Exec(sqlStmt, serviceID); err != nil {
		logx.Error(
			fmt.Sprintf(
				"deleteRow: %q with %q, %s",
				sqlStmt, serviceID, err,
			),
			logFrom,
			true,
		)
	}
}
