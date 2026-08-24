package storage

const environmentCreateTableSql = `CREATE TABLE IF NOT EXISTS environments (
			id        TEXT PRIMARY KEY,
			name      TEXT NOT NULL,
			type      TEXT NOT NULL,
			path  TEXT NOT NULL DEFAULT ''
)`

const scriptsCreateTableSql = `CREATE TABLE IF NOT EXISTS scripts (
			id               TEXT PRIMARY KEY,
			name             TEXT NOT NULL,
			work_dir         TEXT NOT NULL DEFAULT '',
			runner           TEXT NOT NULL DEFAULT '',
			params           TEXT NOT NULL DEFAULT '[]'
)`
