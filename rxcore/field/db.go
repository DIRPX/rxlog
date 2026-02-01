/*
   Copyright 2025 The DIRPX Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package field

const (
	// DBSystem is the field key that identifies the database system
	// (for example, "postgresql", "mysql", "redis").
	DBSystem = "db_system"

	// DBName is the field key that records the database name.
	DBName = "db_name"

	// DBUser is the field key that records the database user.
	DBUser = "db_user"

	// DBStatement is the field key that records the database statement or
	// query text, possibly redacted or normalized.
	DBStatement = "db_statement"

	// DBOperation is the field key that records the logical database
	// operation (for example, "SELECT", "INSERT", "UPDATE").
	DBOperation = "db_operation"

	// DBDuration is the field key that records the time taken to execute
	// the database operation.
	DBDuration = "db_duration"

	// DBRowsAffected is the field key that records how many rows were
	// affected by the database operation, if known.
	DBRowsAffected = "db_rows_affected"
)
