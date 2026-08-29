package storage

const environmentCreateTableSql = `CREATE TABLE IF NOT EXISTS environments (
			id        TEXT PRIMARY KEY,
			name      TEXT NOT NULL,
			paths     TEXT NOT NULL DEFAULT '[]',
			env       TEXT NOT NULL DEFAULT '[]'
)`

const scriptsCreateTableSql = `CREATE TABLE IF NOT EXISTS scripts (
			id               TEXT PRIMARY KEY,
			name             TEXT NOT NULL,
			work_dir         TEXT NOT NULL DEFAULT '',
			command          TEXT NOT NULL DEFAULT '[]',
			environments     TEXT NOT NULL DEFAULT '[]'
)`
