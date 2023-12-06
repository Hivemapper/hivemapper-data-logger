package sql

const ErrorLogsCreateTable string = `
	CREATE TABLE IF NOT EXISTS error_logs (
		id INTEGER NOT NULL PRIMARY KEY,
		system_time TIMESTAMP NOT NULL,
		service_name TEXT NOT NULL,
		message TEXT NOT NULL
	);
	create index if not exists error_time_idx on error_logs(system_time);
`

const insertErrorLogsRawQuery string = `INSERT OR IGNORE INTO error_logs VALUES`

const insertErrorLogsRawFields string = `(NULL,?,?,?),`

const errorLogsPurgeQuery string = `
	DELETE FROM error_logs WHERE system_time < ?;
`

func ErrorLogsCreateTableQuery() string {
	return ErrorLogsCreateTable
}

func ErrorLogsPurgeQuery() string {
	return errorLogsPurgeQuery
}