package storage

import (
	"database/sql"
	"fmt"
)

func OpenOrCreate(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1) // sqlite 单文件，单连接足够，避免并发写冲突

	if err := migrate(db); err != nil {
		err := db.Close()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	stmts := []string{
		scriptsCreateTableSql,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
