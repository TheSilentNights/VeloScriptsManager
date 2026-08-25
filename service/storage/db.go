package storage

import (
	"database/sql"
	"fmt"
	"log"

	"modernc.org/sqlite"
)

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
		{"scripts", "environments", "TEXT NOT NULL DEFAULT '[]'"},
		{"environments", "env", "TEXT NOT NULL DEFAULT '[]'"},
		{"environments", "children", "TEXT NOT NULL DEFAULT '[]'"},
	}
	for _, c := range columns {
		if err := addColumnIfMissing(db, c.table, c.column, c.definition); err != nil {
			return err
		}
	}
	return nil
}

// addColumnIfMissing adds a column to a table when it does not already exist.
// Needed because CREATE TABLE IF NOT EXISTS never modifies an existing table.
func addColumnIfMissing(db *sql.DB, table, column, definition string) error {
	rows, err := db.Query("SELECT name FROM pragma_table_info(?)", table)
	if err != nil {
		return err
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			panic(err)
		}
	}(rows)

	exists := false
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		if name == column {
			exists = true
			break
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if !exists {
		_, err = db.Exec("ALTER table " + table + " ADD COLUMN " + column + " " + definition)
		return err
	}
	return nil
}
