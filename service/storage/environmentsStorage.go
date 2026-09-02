package storage

import (
	"database/sql"
	"encoding/json"
	"log"

	squirrel "github.com/Masterminds/squirrel"
)

type EnvironmentRepo struct {
	db *sql.DB
}

func CreateEnvironmentRepo(db *sql.DB) *EnvironmentRepo {
	return &EnvironmentRepo{db: db}
}

func (er *EnvironmentRepo) List() ([]Environment, error) {
	query, args, err := squirrel.
		Select("id", "name", "paths", "env").
		From("environments").
		OrderBy("name").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := er.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			log.Printf("close rows: %v", err)
		}
	}(rows)

	var out = make([]Environment, 0)
	for rows.Next() {
		e, err := scanEnvironment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *e)
	}

	return out, rows.Err()
}

func (er *EnvironmentRepo) Get(id string) (*Environment, error) {
	query, args, err := squirrel.
		Select("id", "name", "paths", "env").
		From("environments").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := er.db.QueryRow(query, args...)
	return scanEnvironment(row)
}

func (er *EnvironmentRepo) Insert(e Environment) (int64, error) {
	paths, err := json.Marshal(e.Paths)
	if err != nil {
		return 0, err
	}
	env, err := json.Marshal(e.Env)
	if err != nil {
		return 0, err
	}

	query, args, err := squirrel.
		Insert("environments").
		Columns("id", "name", "paths", "env").
		Values(e.ID, e.Name, string(paths), string(env)).
		ToSql()
	if err != nil {
		return 0, err
	}

	result, err := er.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affected, nil
}

func (er *EnvironmentRepo) Update(e Environment) (int64, error) {
	paths, err := json.Marshal(e.Paths)
	if err != nil {
		return 0, err
	}
	env, err := json.Marshal(e.Env)
	if err != nil {
		return 0, err
	}

	query, args, err := squirrel.
		Update("environments").
		Set("name", e.Name).
		Set("paths", string(paths)).
		Set("env", string(env)).
		Where(squirrel.Eq{"id": e.ID}).
		ToSql()
	if err != nil {
		return 0, err
	}

	result, err := er.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, nil
}

func (er *EnvironmentRepo) Delete(id string) (int64, error) {
	query, args, err := squirrel.
		Delete("environments").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return 0, err
	}

	result, err := er.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affected, nil
}

// scanner abstracts *sql.Row and *sql.Rows so a single decode helper works for
// both Get and List.
type scanner interface {
	Scan(dest ...any) error
}

func scanEnvironment(s scanner) (*Environment, error) {
	var e Environment
	var paths string
	var env string
	if err := s.Scan(&e.ID, &e.Name, &paths, &env); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(paths), &e.Paths); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(env), &e.Env); err != nil {
		return nil, err
	}
	return &e, nil
}
