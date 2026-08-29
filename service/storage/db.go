package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"modernc.org/sqlite"
)

// ErrNotFound reports that a row matching the requested id does not exist.
var ErrNotFound = errors.New("record not found")

func OpenOrCreate(dbPath string) (*sql.DB, error) {
	connector, err := sqlite.NewConnector(fmt.Sprintf("file:%s", dbPath))
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db := sql.OpenDB(connector)

	errDbPing := db.Ping()

	if errDbPing != nil {
		shutdownDb(db)
		return nil, fmt.Errorf("ping sqlite: %w", errDbPing)
	}

	db.SetMaxOpenConns(1) // sqlite 单文件，单连接足够，避免并发写冲突

	if err := migrate(db); err != nil {
		shutdownDb(db)
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

func shutdownDb(db *sql.DB) {
	err := db.Close()
	if err != nil {
		log.Printf("close db: %v", err)
	}
}

func migrate(db *sql.DB) error {
	stmts := []string{
		scriptsCreateTableSql,
		environmentCreateTableSql,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}

	columns := []struct{ table, column, definition string }{
		{"scripts", "work_dir", "TEXT NOT NULL DEFAULT ''"},
		{"scripts", "command", "TEXT NOT NULL DEFAULT '[]'"},
		{"scripts", "environments", "TEXT NOT NULL DEFAULT '[]'"},
		{"environments", "paths", "TEXT NOT NULL DEFAULT '[]'"},
		{"environments", "env", "TEXT NOT NULL DEFAULT '[]'"},
	}
	for _, c := range columns {
		if err := addColumnIfMissing(db, c.table, c.column, c.definition); err != nil {
			return err
		}
	}

	if err := migrateScriptsCommand(db); err != nil {
		return err
	}

	drops := []struct{ table, column string }{
		{"environments", "type"},
		{"environments", "path"},
		{"environments", "children"},
	}
	for _, c := range drops {
		if err := dropColumnIfExists(db, c.table, c.column); err != nil {
			return err
		}
	}
	return nil
}

func migrateScriptsCommand(db *sql.DB) error {
	runnerExists, err := columnExists(db, "scripts", "runner")
	if err != nil {
		return err
	}
	if !runnerExists {
		return nil
	}

	type scriptRow struct {
		id     string
		runner string
		params string
	}
	scriptRows, err := func() ([]scriptRow, error) {
		rows, err := db.Query("SELECT id, runner, params FROM scripts")
		if err != nil {
			return nil, err
		}
		defer func() { _ = rows.Close() }()

		var out []scriptRow
		for rows.Next() {
			var r scriptRow
			if err := rows.Scan(&r.id, &r.runner, &r.params); err != nil {
				return nil, err
			}
			out = append(out, r)
		}
		return out, rows.Err()
	}()
	if err != nil {
		return err
	}

	for _, r := range scriptRows {
		var params []string
		if err := json.Unmarshal([]byte(r.params), &params); err != nil {
			return err
		}
		command := params
		if r.runner != "" {
			command = append([]string{r.runner}, params...)
		}
		data, err := json.Marshal(command)
		if err != nil {
			return err
		}
		if _, err := db.Exec("UPDATE scripts SET command = ? WHERE id = ?", string(data), r.id); err != nil {
			return err
		}
	}

	if err := dropColumnIfExists(db, "scripts", "runner"); err != nil {
		return err
	}
	return dropColumnIfExists(db, "scripts", "params")
}

// addColumnIfMissing adds a column to a table when it does not already exist.
// Needed because CREATE TABLE IF NOT EXISTS never modifies an existing table.
func addColumnIfMissing(db *sql.DB, table, column, definition string) error {
	exists, err := columnExists(db, table, column)
	if err != nil {
		return err
	}

	if !exists {
		_, err = db.Exec("ALTER table " + table + " ADD COLUMN " + column + " " + definition)
		return err
	}
	return nil
}

func dropColumnIfExists(db *sql.DB, table, column string) error {
	exists, err := columnExists(db, table, column)
	if err != nil {
		return err
	}

	if exists {
		_, err = db.Exec("ALTER table " + table + " DROP COLUMN " + column)
		return err
	}
	return nil
}

func columnExists(db *sql.DB, table, column string) (bool, error) {
	rows, err := db.Query("SELECT name FROM pragma_table_info(?)", table)
	if err != nil {
		return false, err
	}

	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			log.Printf("close rows: %v", err)
		}
	}(rows)

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, err
	}

	return false, nil
}
