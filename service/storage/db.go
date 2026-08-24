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
	return nil
}
